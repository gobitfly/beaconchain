package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardBlocks(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	var err error
	var currentCursor t.BlocksCursor

	// TODO @LuccaBitfly move validation to handler?
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.BlocksCursor](cursor); err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as BlocksCursor: %w", err)
		}
	}

	// regexes taken from api handler common.go
	searchPubkey := regexp.MustCompile(`^0x[0-9a-fA-F]{96}$`).MatchString(search)
	searchGroup := regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]+$`).MatchString(search)
	searchIndex := regexp.MustCompile(`^[0-9]+$`).MatchString(search)

	validatorMap := make(map[t.VDBValidator]bool)
	params := []interface{}{}
	filteredValidatorsQuery := ""
	validatorMapping, releaseValMapLock, err := d.services.GetCurrentValidatorMapping()
	defer releaseValMapLock()
	if err != nil {
		return nil, nil, err
	}

	// determine validators of interest first
	if dashboardId.Validators == nil {
		// could also optimize this for the average and/or the whale case; will go with some middle-ground, needs testing
		// (query validators twice: once without search applied (fast) to pre-filter scheduled proposals (which are sent to db, want to minimize),
		// again for blocks query with search applied to not having to send potentially huge validator-list)
		startTime := time.Now()
		valis, err := d.getDashboardValidators(ctx, dashboardId, nil)
		log.Debugf("=== getting validators took %s", time.Since(startTime))
		if err != nil {
			return nil, nil, err
		}
		for _, v := range valis {
			validatorMap[v] = true
		}

		// create a subquery to get the (potentially filtered) validators and their groups for later
		params = append(params, dashboardId.Id)
		selectStr := `SELECT validator_index, group_id `
		from := `FROM users_val_dashboards_validators validators `
		where := `WHERE validators.dashboard_id = $1`
		extraConds := make([]string, 0, 3)
		if searchIndex {
			params = append(params, search)
			extraConds = append(extraConds, fmt.Sprintf(`validator_index = $%d`, len(params)))
		}
		if searchGroup {
			from += `INNER JOIN users_val_dashboards_groups groups ON validators.dashboard_id = groups.dashboard_id AND validators.group_id = groups.id `
			// escape the psql single character wildcard "_"; apply prefix-search
			params = append(params, strings.Replace(search, "_", "\\_", -1)+"%")
			extraConds = append(extraConds, fmt.Sprintf(`LOWER(name) LIKE LOWER($%d)`, len(params)))
		}
		if searchPubkey {
			index, ok := validatorMapping.ValidatorIndices[search]
			if !ok && len(extraConds) == 0 {
				// don't even need to query
				return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
			}
			params = append(params, index)
			extraConds = append(extraConds, fmt.Sprintf(`validator_index = $%d`, len(params)))
		}
		if len(extraConds) > 0 {
			where += ` AND (` + strings.Join(extraConds, ` OR `) + `)`
		}

		filteredValidatorsQuery = selectStr + from + where
	} else {
		validators := make([]t.VDBValidator, 0, len(dashboardId.Validators))
		for _, validator := range dashboardId.Validators {
			if searchIndex && fmt.Sprint(validator) != search ||
				searchPubkey && validator != validatorMapping.ValidatorIndices[search] {
				continue
			}
			validatorMap[validator] = true
			validators = append(validators, validator)
			if searchIndex || searchPubkey {
				break
			}
		}
		if len(validators) == 0 {
			return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
		}
		params = append(params, validators)
	}

	var proposals []struct {
		Proposer     t.VDBValidator      `db:"proposer"`
		Group        uint64              `db:"group_id"`
		Epoch        uint64              `db:"epoch"`
		Slot         uint64              `db:"slot"`
		Status       uint64              `db:"status"`
		Block        sql.NullInt64       `db:"block"`
		FeeRecipient []byte              `db:"fee_recipient"`
		ElReward     decimal.NullDecimal `db:"el_reward"`
		ClReward     decimal.NullDecimal `db:"cl_reward"`
		GraffitiText string              `db:"graffiti_text"`

		// for cursor only
		Reward decimal.Decimal
	}

	// handle sorting
	where := ``
	orderBy := `ORDER BY `
	sortOrder := ` ASC`
	if colSort.Desc {
		sortOrder = ` DESC`
	}
	val := t.VDBValidator(0)
	sortColName := `slot`
	switch colSort.Column {
	case enums.VDBBlockProposer:
		sortColName = `proposer`
		val = currentCursor.Proposer
	case enums.VDBBlockStatus:
		sortColName = `status`
		val = currentCursor.Status
	case enums.VDBBlockProposerReward:
		sortColName = `reward`
		val = currentCursor.Reward.BigInt().Uint64()
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
			if sortColName == "status" {
				// explicit cast to int because type of 'status' column is text for some reason
				sortColName += "::int"
			}
			where += fmt.Sprintf(`(slot`+secSign+`$%d AND `+sortColName+` = $%d) OR `+sortColName+sign+`$%d`, len(params)-1, len(params), len(params))
		}
		where += `) `
	}
	if sortOrder == ` ASC` {
		sortOrder += ` NULLS FIRST`
	} else {
		sortOrder += ` NULLS LAST`
	}
	orderBy += sortColName + sortOrder
	if !onlyPrimarySort {
		secSort := `DESC`
		if currentCursor.IsReverse() {
			secSort = `ASC`
		}
		orderBy += `, slot ` + secSort
	}

	// Get scheduled blocks. They aren't written to blocks table, get from duties
	// Will just pass scheduled proposals to query and let db do the sorting etc
	var scheduledProposers []t.VDBValidator
	var scheduledEpochs []uint64
	var scheduledSlots []uint64
	// don't need to query if requested slots are in the past
	latestSlot := cache.LatestSlot.Get()
	if !onlyPrimarySort || !currentCursor.IsValid() ||
		currentCursor.Slot > latestSlot+1 && currentCursor.Reverse != colSort.Desc ||
		currentCursor.Slot < latestSlot+1 && currentCursor.Reverse == colSort.Desc {
		dutiesInfo, releaseLock, err := d.services.GetCurrentDutiesInfo()
		defer releaseLock()
		if err == nil {
			for slot, vali := range dutiesInfo.PropAssignmentsForSlot {
				// only gather scheduled slots
				if _, ok := dutiesInfo.SlotStatus[slot]; ok {
					continue
				}
				// only gather slots scheduled for our validators
				if _, ok := validatorMap[vali]; !ok {
					continue
				}
				scheduledProposers = append(scheduledProposers, dutiesInfo.PropAssignmentsForSlot[slot])
				scheduledEpochs = append(scheduledEpochs, slot/utils.Config.Chain.ClConfig.SlotsPerEpoch)
				scheduledSlots = append(scheduledSlots, slot)
			}
		} else {
			log.Debugf("duties info not available, skipping scheduled slots: %s", err)
		}
		if len(scheduledProposers) > 0 {
			// make sure the distinct clause filters out the correct duplicated row (e.g. block=nil)
			orderBy += `, block`
		}
	}

	groupIdCol := "group_id"
	// this is actually just used for sorting for "reward".. will not consider EL rewards of unfinalized blocks atm
	reward := "reward"
	if dashboardId.Validators != nil {
		groupIdCol = fmt.Sprintf("%d AS %s", t.DefaultGroupId, groupIdCol)
		reward = "coalesce(rb.value / 1e18, ep.fee_recipient_reward) AS " + reward
	}
	selectFields := fmt.Sprintf(`
		r.proposer,
		%s,
		r.epoch,
		r.slot,
		r.status,
		block,
		COALESCE(rb.proposer_fee_recipient, blocks.exec_fee_recipient) AS fee_recipient,
		COALESCE(rb.value / 1e18, ep.fee_recipient_reward) AS el_reward,
		cp.cl_attestations_reward / 1e9 + cp.cl_sync_aggregate_reward / 1e9 + cp.cl_slashing_inclusion_reward / 1e9 as cl_reward,
		r.graffiti_text`, groupIdCol)
	query := fmt.Sprintf(`SELECT
			%s
		FROM ( SELECT * FROM (`, selectFields)
	// supply scheduled proposals, if any
	if len(scheduledProposers) > 0 {
		// distinct to filter out duplicates in an edge case (if dutiesInfo didn't update yet after a block was proposed, but the blocks table was)
		// might be possible to remove this once the TODO in service_slot_viz.go:startSlotVizDataService is resolved
		distinct := "slot"
		if !onlyPrimarySort {
			distinct = sortColName + ", " + distinct
		}
		params = append(params, scheduledProposers)
		params = append(params, scheduledEpochs)
		params = append(params, scheduledSlots)
		query = fmt.Sprintf(`SELECT distinct on (%s) 
			%s
		FROM ( SELECT * FROM (WITH scheduled_proposals (
			proposer,
			epoch,
			slot,
			status,
			block,
			reward,
			graffiti_text
		) AS (SELECT 
			*,
			'0',
			null::int,
			null::int,
			''
			FROM unnest($%d::int[], $%d::int[], $%d::int[]))
		SELECT * FROM scheduled_proposals
		UNION
		(`, distinct, selectFields, len(params)-2, len(params)-1, len(params))
	}
	query += fmt.Sprintf(`
	SELECT
		proposer,
		epoch,
		blocks.slot,
		status,
		exec_block_number AS block,
		%s,
		graffiti_text
	FROM blocks
	`, reward)

	if dashboardId.Validators == nil {
		query += `
		LEFT JOIN cached_proposal_rewards ON cached_proposal_rewards.dashboard_id = $1 AND blocks.slot = cached_proposal_rewards.slot
		`
	} else {
		query += `
		LEFT JOIN execution_payloads ep ON ep.block_hash = blocks.exec_block_hash
		LEFT JOIN relays_blocks rb ON rb.exec_block_hash = blocks.exec_block_hash
		`
	}

	// shrink selection to our filtered validators
	if len(scheduledProposers) > 0 {
		query += `)`
	}
	query += `) as u `
	if dashboardId.Validators == nil {
		query += fmt.Sprintf(`
		INNER JOIN (%s) validators ON validators.validator_index = proposer
		`, filteredValidatorsQuery)
	} else {
		query += `WHERE proposer = ANY($1) `
	}

	params = append(params, limit+1)
	limitStr := fmt.Sprintf(`
		LIMIT $%d
	`, len(params))
	rewardsStr := `) r
	LEFT JOIN consensus_payloads cp on r.slot = cp.slot
	LEFT JOIN blocks on r.slot = blocks.slot
	LEFT JOIN execution_payloads ep ON ep.block_hash = blocks.exec_block_hash
	LEFT JOIN relays_blocks rb ON rb.exec_block_hash = blocks.exec_block_hash
	`

	startTime := time.Now()
	err = d.alloyReader.Select(&proposals, query+where+orderBy+limitStr+rewardsStr+orderBy, params...)
	log.Debugf("=== getting past blocks took %s", time.Since(startTime))
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

	data := make([]t.VDBBlocksTableRow, len(proposals))
	ensMapping := make(map[string]string)
	for i, proposal := range proposals {
		data[i].GroupId = proposal.Group
		if dashboardId.AggregateGroups {
			data[i].GroupId = t.DefaultGroupId
		}
		data[i].Proposer = proposal.Proposer
		data[i].Epoch = proposal.Epoch
		data[i].Slot = proposal.Slot
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
		graffiti := proposal.GraffitiText
		data[i].Graffiti = &graffiti
		if proposal.Status == 3 {
			continue
		}
		block := uint64(proposal.Block.Int64)
		data[i].Block = &block
		var reward t.ClElValue[decimal.Decimal]
		if proposal.ElReward.Valid {
			rewardRecp := t.Address{
				Hash: t.Hash(hexutil.Encode(proposal.FeeRecipient)),
			}
			data[i].RewardRecipient = &rewardRecp
			ensMapping[hexutil.Encode(proposal.FeeRecipient)] = ""
			reward.El = proposal.ElReward.Decimal.Mul(decimal.NewFromInt(1e18))
		}
		if proposal.ClReward.Valid {
			reward.Cl = proposal.ClReward.Decimal.Mul(decimal.NewFromInt(1e18))
		}
		proposals[i].Reward = proposal.ElReward.Decimal.Add(proposal.ClReward.Decimal)
		data[i].Reward = &reward
	}
	// determine reward recipient ENS names
	startTime = time.Now()
	if err := db.GetEnsNamesForAddresses(ensMapping); err != nil {
		return nil, nil, err
	}
	log.Debugf("=== getting ens names took %s", time.Since(startTime))
	for i := range data {
		if data[i].RewardRecipient != nil {
			data[i].RewardRecipient.Ens = ensMapping[string(data[i].RewardRecipient.Hash)]
		}
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
