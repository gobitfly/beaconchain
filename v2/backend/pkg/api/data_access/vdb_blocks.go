package dataaccess

import (
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardBlocks(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	var err error
	var currentCursor t.BlocksCursor

	// TODO @LuccaBitfly move validation to handler
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.BlocksCursor](cursor); err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as BlocksCursor: %w", err)
		}
	}

	// regexes taken from api handler common.go
	searchPubkey := regexp.MustCompile(`^0x[0-9a-fA-F]{96}$`).MatchString(search)
	searchGroup := regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]+$`).MatchString(search)
	searchIndex := regexp.MustCompile(`^[0-9]+$`).MatchString(search)

	validatorGroupMap := make(map[uint64]uint64)
	params := []interface{}{}
	if dashboardId.Validators == nil {
		// Get the validators and their groups in case a dashboard id is provided
		queryResult := []struct {
			ValidatorIndex uint64 `db:"validator_index"`
			GroupId        uint64 `db:"group_id"`
		}{}

		paramsGroups := []interface{}{dashboardId.Id}
		selectStr := `SELECT validator_index, group_id `
		from := `FROM users_val_dashboards_validators validators `
		where := `WHERE validators.dashboard_id = $1`
		extraConds := make([]string, 0, 3)
		if searchIndex {
			paramsGroups = append(paramsGroups, search)
			extraConds = append(extraConds, fmt.Sprintf(`validator_index = $%d`, len(paramsGroups)))
		}
		if searchGroup {
			from += `INNER JOIN users_val_dashboards_groups groups ON validators.dashboard_id = groups.dashboard_id AND validators.group_id = groups.id `
			// escape the psql single character wildcard "_"; apply prefix-search
			paramsGroups = append(paramsGroups, strings.Replace(search, "_", "\\_", -1)+"%")
			extraConds = append(extraConds, fmt.Sprintf(`LOWER(name) LIKE LOWER($%d)`, len(paramsGroups)))
		}
		if searchPubkey {
			validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
			defer releaseValMapLock()
			if err != nil {
				return nil, nil, err
			}
			index, ok := validatorMapping.ValidatorIndices[search]
			if !ok && len(extraConds) == 0 {
				// don't even need to query
				return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
			}
			paramsGroups = append(paramsGroups, index)
			extraConds = append(extraConds, fmt.Sprintf(`validator_index = $%d`, len(paramsGroups)))
		}
		if len(extraConds) > 0 {
			where += ` AND (` + strings.Join(extraConds, ` OR `) + `)`
		}

		startTime := time.Now()
		log.Infof(selectStr + from + where)
		err := d.alloyReader.Select(&queryResult, selectStr+from+where, paramsGroups...)
		log.Infof("=== getting (filtered) validators took %s", time.Since(startTime))
		if err != nil {
			return nil, nil, err
		}
		if len(queryResult) == 0 {
			return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
		}
		validators := []uint64{}
		for _, res := range queryResult {
			validatorGroupMap[res.ValidatorIndex] = res.GroupId

			validators = append(validators, res.ValidatorIndex)
			// groups = append(groups, res.GroupId)
		}
		params = append(params, validators)
		// params = append(params, groups)
		// TODO once the blocks table has been migrated, this can be replaced with a simple join
		// withGroups = fmt.Sprintf(`WITH validator_groups (validator_index, group_id) AS (SELECT * FROM unnest($%d::int[], $%d::int[]))`, len(params)-1, len(params))
	} else {
		// In case a list of validators is provided, set the group to default 0
		validatorGroupQryValues := ""
		validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
		defer releaseValMapLock()
		if err != nil {
			return nil, nil, err
		}
		for _, validator := range dashboardId.Validators {
			if searchIndex && fmt.Sprint(validator.Index) != search ||
				searchPubkey && (validatorMapping.ValidatorIndices[search] == nil || validator.Index != *validatorMapping.ValidatorIndices[search]) {
				continue
			}
			validatorGroupMap[validator.Index] = t.DefaultGroupId
			params = append(params, validator.Index)
			validatorGroupQryValues = validatorGroupQryValues + fmt.Sprintf(`($%d::int, %d::int), `, len(params), t.DefaultGroupId)
			if searchIndex || searchPubkey {
				break
			}
		}
		if len(params) == 0 {
			return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
		}
		// using where on the (filtered) list of validator indices later down would obv be better for this, but let's not make this even less readable
		// withGroups = fmt.Sprintf(`WITH validator_groups (validator_index, group_id) AS (VALUES %s)`, validatorGroupQryValues[:len(validatorGroupQryValues)-2])
	}

	type proposal struct {
		Proposer     int64 `db:"proposer"`
		Group        int64
		Epoch        uint64        `db:"epoch"`
		Slot         int64         `db:"slot"`
		Status       int64         `db:"status"`
		Block        sql.NullInt64 `db:"exec_block_number"`
		FeeRecipient []byte        `db:"exec_fee_recipient"`
		Mev          sql.NullInt64 `db:"mev_reward"`
		MevRecipient []byte        `db:"proposer_fee_recipient"`
		GraffitiText string        `db:"graffiti_text"`
		// TODO fill this properly, used for cursor atm
		Reward int64
	}
	proposals := make([]proposal, 0)

	// scheduled blocks aren't written to blocks table, get from duties
	dutiesInfo, releaseLock, err := d.services.GetCurrentDutiesInfo()
	// just pass scheduled proposals to query and let db do the sorting etc
	var props []uint64
	var epochs []uint64
	var slots []uint64
	defer releaseLock()
	if err == nil {
		for slot, vali := range dutiesInfo.PropAssignmentsForSlot {
			// only gather scheduled slots
			if _, ok := dutiesInfo.SlotStatus[slot]; ok {
				continue
			}
			// only gather slots scheduled for our validators
			if _, ok := validatorGroupMap[vali]; !ok {
				continue
			}
			props = append(props, dutiesInfo.PropAssignmentsForSlot[slot])
			epochs = append(epochs, slot/utils.Config.Chain.ClConfig.SlotsPerEpoch)
			slots = append(slots, slot)
		}
		params = append(params, props)
		params = append(params, epochs)
		params = append(params, slots)
	} else {
		log.Debugf("duties info not available, skipping scheduled slots: %s", err)
	}

	where := ``
	orderBy := `ORDER BY `
	sortOrder := ` ASC`
	if colSort.Desc {
		sortOrder = ` DESC`
	}
	val := int64(-1)
	sortColName := `slot`
	switch colSort.Column {
	case enums.VDBBlockProposer:
		sortColName = `proposer`
		val = currentCursor.Proposer
	case enums.VDBBlockStatus:
		sortColName = `status`
		val = currentCursor.Status
	case enums.VDBBlockProposerReward:
		// TODO need to sum up reward data; CL rewards missing, EL only in BT
		sortColName = `mev_reward`
		val = currentCursor.Reward
	}
	onlyPrimarySort := sortColName == `slot`
	if currentCursor.IsValid() {
		sign := ` > `
		if colSort.Desc && !currentCursor.IsReverse() || !colSort.Desc && currentCursor.IsReverse() {
			sign = ` < `
		}
		if currentCursor.IsReverse() {
			if sortOrder == ` ASC` {
				sortOrder = ` DESC`
			} else {
				sortOrder = ` ASC`
			}
		}
		params = append(params, currentCursor.Slot)
		where += `WHERE (`
		if onlyPrimarySort {
			where += `slot` + sign + fmt.Sprintf(`$%d`, len(params))
		} else {
			params = append(params, val)
			secSign := ` < `
			if currentCursor.IsReverse() {
				secSign = ` > `
			}
			// explicit cast to int because type of 'status' column is text for some reason
			where += fmt.Sprintf(`(slot`+secSign+`$%d AND `+sortColName+`::int = $%d) OR `+sortColName+`::int`+sign+`$%d`, len(params)-1, len(params), len(params))
		}
		where += `) `
	}
	orderBy += sortColName + sortOrder
	if !onlyPrimarySort {
		secSort := `DESC`
		if currentCursor.IsReverse() {
			secSort = `ASC`
		}
		orderBy += `, slot ` + secSort
	}
	if len(props) > 0 {
		// make sure the distinct clause filters out the correct row (e.g. block=nil)
		orderBy += `, exec_block_number`
	}

	query := ""
	if len(props) > 0 {
		// distinct to filter out duplicates in an edge case (if dutiesInfo didn't update yet after a block was proposed, but the blocks table was)
		// might be possible to remove this once the TODO in service_slot_viz.go:startSlotVizDataService is resolved
		distinct := "slot"
		if !onlyPrimarySort {
			distinct = sortColName + ", " + distinct
		}
		query = fmt.Sprintf(`SELECT distinct on (%s) * FROM (WITH scheduled_proposals (
			proposer,
			epoch,
			slot,
			status,
			exec_block_number,
			exec_fee_recipient,
			mev_reward,
			proposer_fee_recipient,
			graffiti_text
		) AS (SELECT 
			*,
			'0',
			null::int,
			null::bytea,
			null::int,
			null::bytea,
			''
			FROM unnest($2::int[], $3::int[], $4::int[]))
		SELECT * FROM scheduled_proposals
		UNION
		(`, distinct)
	}
	query += `
	SELECT
		proposer,
		epoch,
		slot,
		status,
		exec_block_number,
		exec_fee_recipient,
		relays_blocks.value AS mev_reward,
		COALESCE(proposer_fee_recipient, '') AS proposer_fee_recipient,
		graffiti_text
	FROM blocks
	LEFT JOIN relays_blocks ON blocks.exec_block_hash = relays_blocks.exec_block_hash
	WHERE proposer = ANY($1)
	`
	if len(props) > 0 {
		query += `)) as u `
	}

	params = append(params, limit+1)
	limitStr := fmt.Sprintf(`
		LIMIT $%d
	`, len(params))

	// fmt.Println(query + where + orderBy + limitStr)
	startTime := time.Now()
	err = d.readerDb.Select(&proposals, query+where+orderBy+limitStr, params...)
	log.Infof("=== getting past blocks took %s", time.Since(startTime))
	if err != nil {
		return nil, nil, err
	}
	if len(proposals) == 0 {
		return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
	}
	moreDataFlag := len(proposals) > int(limit)
	if moreDataFlag {
		proposals = proposals[:len(proposals)-1]
	}
	if currentCursor.IsReverse() {
		slices.Reverse(proposals)
	}

	blocksNoRelay := make([]uint64, 0, len(proposals))
	data := make([]t.VDBBlocksTableRow, len(proposals))
	for _, proposal := range proposals {
		if proposal.Status == 1 && !proposal.Mev.Valid {
			blocksNoRelay = append(blocksNoRelay, uint64(proposal.Block.Int64))
		}
	}
	// (non-mev) tx reward is in bt
	startTime = time.Now()
	indexedBlocksNoRelay, err := d.bigtable.GetBlocksIndexedMultiple(blocksNoRelay, uint64(len(blocksNoRelay)))
	log.Infof("=== bigtable took %s", time.Since(startTime))
	if err != nil {
		return nil, nil, err
	}
	idxNoRelayBlocks := len(indexedBlocksNoRelay) - 1
	ensMapping := make(map[string]string)
	for i, proposal := range proposals {
		data[i].GroupId = validatorGroupMap[uint64(proposal.Proposer)]
		data[i].Proposer = uint64(proposal.Proposer)
		data[i].Epoch = proposal.Epoch
		data[i].Slot = uint64(proposal.Slot)
		switch proposal.Status {
		case 0:
			data[i].Status = "scheduled"
		case 1:
			data[i].Status = "success"
		case 2:
			data[i].Status = "missed"
		case 3:
			data[i].Status = "orphaned"
		default:
			// invalid
		}
		if proposal.Status == 0 || proposal.Status == 2 {
			continue
		}
		data[i].Graffiti = proposal.GraffitiText
		if proposal.Status == 3 {
			continue
		}
		data[i].Block = uint64(proposal.Block.Int64)
		rewardRecipient := ""
		if proposal.Mev.Valid {
			data[i].Reward.El = decimal.NewFromInt(proposal.Mev.Int64)
			rewardRecipient = hexutil.Encode(proposal.MevRecipient)
		} else {
			// bt returns blocks sorted desc
			for ; idxNoRelayBlocks >= 0; idxNoRelayBlocks-- {
				if indexedBlocksNoRelay[idxNoRelayBlocks].Number == uint64(proposal.Block.Int64) {
					data[i].Reward.El = decimal.NewFromBigInt(utils.Eth1TotalReward(indexedBlocksNoRelay[idxNoRelayBlocks]), 0)
					break
				}
			}
			rewardRecipient = hexutil.Encode(proposal.FeeRecipient)
		}
		data[i].RewardRecipient.Hash = t.Hash(rewardRecipient)
		ensMapping[rewardRecipient] = ""
		// TODO CL rewards - depends on BIDS-3036
	}
	// determine reward recipient ENS names
	if err := db.GetEnsNamesForAddresses(ensMapping); err != nil {
		return nil, nil, err
	}
	for i := range data {
		data[i].RewardRecipient.Ens = ensMapping[string(data[i].RewardRecipient.Hash)]
	}
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return data, &t.Paging{}, nil
	}
	p, err := utils.GetPagingFromData(proposals, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, err
	}
	return data, p, nil
}
