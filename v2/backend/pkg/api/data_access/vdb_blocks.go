package dataaccess

import (
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
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

	validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
	defer releaseValMapLock()
	if err != nil {
		return nil, nil, err
	}

	validatorGroupMap := make(map[uint64]uint64)
	var validators []uint64
	validatorGroupQryValues := ""
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
		if searchIndex {
			where += ` AND validator_index = $2`
			paramsGroups = append(paramsGroups, search)
		} else if searchGroup {
			from += `INNER JOIN users_val_dashboards_groups groups ON validators.dashboard_id = groups.dashboard_id AND validators.group_id = groups.id `
			where += ` AND LOWER(name) LIKE LOWER($2)`
			paramsGroups = append(paramsGroups, strings.Replace(search, "_", "\\_", -1)+"%")
		} else if searchPubkey {
			index, ok := validatorMapping.ValidatorIndices[search]
			if !ok {
				return nil, nil, fmt.Errorf("unknown pubkey: %s", search)
			}
			where += ` AND validator_index = $2`
			paramsGroups = append(paramsGroups, index)
		}

		err := d.alloyReader.Select(&queryResult, selectStr+from+where, paramsGroups...)
		if err != nil {
			return nil, nil, err
		}
		for i, res := range queryResult {
			validatorGroupMap[res.ValidatorIndex] = res.GroupId
			validators = append(validators, res.ValidatorIndex)

			validatorGroupQryValues = validatorGroupQryValues + fmt.Sprintf(`($%d::int, $%d::int), `, i*2+1, (i+1)*2)
			params = append(params, res.ValidatorIndex)
			params = append(params, res.GroupId)
		}
	} else {
		// In case a list of validators is provided, set the group to default 0
		validators = make([]uint64, 0, len(dashboardId.Validators))
		for _, validator := range dashboardId.Validators {
			if searchIndex && fmt.Sprint(validator.Index) != search ||
				searchPubkey && (validatorMapping.ValidatorIndices[search] == nil || validator.Index != *validatorMapping.ValidatorIndices[search]) {
				continue
			}
			validatorGroupMap[validator.Index] = t.DefaultGroupId
			validators = append(validators, validator.Index)
			if searchIndex || searchPubkey {
				break
			}
		}
	}
	if len(validators) == 0 {
		return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
	}

	// TODO once the blocks table has been migrated, this can be replaced with a simple join
	with := fmt.Sprintf(`WITH validator_groups (validator_index, group_id) AS (VALUES %s)`, validatorGroupQryValues[:len(validatorGroupQryValues)-2])

	type proposal struct {
		Proposer     int64         `db:"proposer"`
		Group        int64         `db:"group_id"`
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
	defer releaseLock()
	if err != nil {
		return nil, nil, err
	}
	// just pass scheduled proposals to query and let db do the sorting etc
	scheduledPropsQryValues := ""
	for slot, vali := range dutiesInfo.PropAssignmentsForSlot {
		// only gather scheduled slots
		if _, ok := dutiesInfo.SlotStatus[slot]; ok {
			continue
		}
		// only gather slots scheduled for our validators
		if _, ok := validatorGroupMap[vali]; !ok {
			continue
		}
		params = append(params, dutiesInfo.PropAssignmentsForSlot[slot])
		params = append(params, validatorGroupMap[dutiesInfo.PropAssignmentsForSlot[slot]])
		params = append(params, slot/utils.Config.Chain.ClConfig.SlotsPerEpoch)
		params = append(params, slot)
		scheduledPropsQryValues = scheduledPropsQryValues + fmt.Sprintf(`($%d::int, $%d::int, $%d::int, $%d::int, '0', null::int, null::bytea, null::int, null::bytea, ''), `, len(params)-3, len(params)-2, len(params)-1, len(params))
	}

	query := ""
	if len(scheduledPropsQryValues) > 0 {
		query = fmt.Sprintf(`SELECT * FROM (SELECT * FROM (WITH scheduled_proposals (
			proposer,
			group_id,
			epoch,
			slot,
			status,
			exec_block_number,
			exec_fee_recipient,
			mev_reward,
			proposer_fee_recipient,
			graffiti_text
		) AS (VALUES %s)
		SELECT * FROM scheduled_proposals) proposals
		UNION
		(`, scheduledPropsQryValues[:len(scheduledPropsQryValues)-2])
	}

	query += with + `
	SELECT
		proposer,
		group_id,
		epoch,
		slot,
		status,
		exec_block_number,
		exec_fee_recipient,
		relays_blocks.value AS mev_reward,
		COALESCE(proposer_fee_recipient, '') AS proposer_fee_recipient,
		graffiti_text
	FROM validator_groups
	LEFT JOIN blocks ON blocks.proposer = validator_groups.validator_index
	LEFT JOIN relays_blocks ON blocks.exec_block_hash = relays_blocks.exec_block_hash
	`
	if len(scheduledPropsQryValues) > 0 {
		query += `)) as u `
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
	case enums.VDBBlockGroup:
		sortColName = `group_id`
		val = currentCursor.Group
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
		if colSort.Desc && !currentCursor.Reverse || !colSort.Desc && currentCursor.Reverse {
			sign = ` < `
		}
		if currentCursor.Reverse {
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
			if currentCursor.Reverse {
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
		if currentCursor.Reverse {
			secSort = `ASC`
		}
		orderBy += `, slot ` + secSort
	}
	params = append(params, limit+1)
	limitStr := fmt.Sprintf(`
		LIMIT $%d
	`, len(params))

	err = d.readerDb.Select(&proposals, query+where+orderBy+limitStr, params...)
	if err != nil {
		return nil, nil, err
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
	indexedBlocksNoRelay, err := d.bigtable.GetBlocksIndexedMultiple(blocksNoRelay, uint64(len(blocksNoRelay)))
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
