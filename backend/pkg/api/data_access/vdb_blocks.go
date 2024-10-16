package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type table string

// Stringer interface
func (t table) String() string {
	return string(t)
}

//func (t table) C(column string) exp.IdentifierExpression {
//	return goqu.I(string(t) + "." + column)
//}

func (t table) C(column string) string {
	return string(t) + "." + column
}

func (d *DataAccessService) GetValidatorDashboardBlocks(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBBlocksColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBBlocksTableRow, *t.Paging, error) {
	// @DATA-ACCESS incorporate protocolModes

	// -------------------------------------
	// Setup
	var err error
	var currentCursor t.BlocksCursor
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, nil, err
	}
	validators := table("validators")
	groups := table("goups")

	// TODO @LuccaBitfly move validation to handler?
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.BlocksCursor](cursor); err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as BlocksCursor: %w", err)
		}
	}

	searchPubkey := regexp.MustCompile(`^0x[0-9a-fA-F]{96}$`).MatchString(search)
	searchGroup := regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]+$`).MatchString(search)
	searchIndex := regexp.MustCompile(`^[0-9]+$`).MatchString(search)

	// -------------------------------------
	// Goqu Query: Determine validators filtered by search
	type validatorGroup struct {
		Validator t.VDBValidator `db:"validator_index"`
		Group     uint64         `db:"group_id"`
	}
	var filteredValidators []validatorGroup
	validatorsDs := goqu.Dialect("postgres").
		From(
			goqu.T("users_val_dashboards_validators").As(validators),
		).
		Select(
			validators.C("validator_index"),
		)
	if dashboardId.Validators == nil {
		validatorsDs = validatorsDs.
			Select(
				// TODO mustn't be here, can be done further down
				validators.C("group_id"),
			).
			Where(goqu.Ex{validators.C("dashboard_id"): dashboardId.Id})

		// apply search filters
		if searchIndex {
			validatorsDs = validatorsDs.Where(goqu.Ex{validators.C("validator_index"): search})
		}
		if searchGroup {
			validatorsDs = validatorsDs.
				InnerJoin(goqu.T("users_val_dashboards_groups").As(groups), goqu.On(
					goqu.Ex{validators.C("dashboard_id"): groups.C("dashboard_id")},
					goqu.Ex{validators.C("group_id"): groups.C("id")},
				)).
				Where(
					goqu.L("LOWER(?)", groups.C("name")).Like(strings.Replace(search, "_", "\\_", -1) + "%"),
				)
		}
		if searchPubkey {
			index, ok := validatorMapping.ValidatorIndices[search]
			if !ok && !searchGroup && !searchIndex {
				// searched pubkey doesn't exist, don't even need to query anything
				return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
			}

			validatorsDs = validatorsDs.
				Where(goqu.Ex{validators.C("validator_index"): index})
		}
	} else {
		for _, validator := range dashboardId.Validators {
			if searchIndex && fmt.Sprint(validator) != search ||
				searchPubkey && validator != validatorMapping.ValidatorIndices[search] {
				continue
			}
			filteredValidators = append(filteredValidators, validatorGroup{
				Validator: validator,
				Group:     t.DefaultGroupId,
			})
			if searchIndex || searchPubkey {
				break
			}
		}
		validatorsDs = validatorsDs.
			Where(goqu.L(
				validators.C("validator_index")+" = ANY(?)", pq.Array(filteredValidators)),
			)
	}

	if dashboardId.Validators == nil {
		validatorsQuery, validatorsArgs, err := validatorsDs.Prepared(true).ToSQL()
		if err != nil {
			return nil, nil, err
		}
		if err = d.alloyReader.SelectContext(ctx, &filteredValidators, validatorsQuery, validatorsArgs...); err != nil {
			return nil, nil, err
		}
	}
	if len(filteredValidators) == 0 {
		return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
	}

	// -------------------------------------
	// Gather scheduled blocks
	// found in dutiesInfo; pass results to final query later and let db do the sorting etc
	validatorSet := make(map[t.VDBValidator]bool)
	for _, v := range filteredValidators {
		validatorSet[v.Validator] = true
	}
	var scheduledProposers []t.VDBValidator
	var scheduledEpochs []uint64
	var scheduledSlots []uint64
	// don't need if requested slots are in the past
	latestSlot := cache.LatestSlot.Get()
	onlyPrimarySort := colSort.Column == enums.VDBBlockSlot || colSort.Column == enums.VDBBlockBlock
	if !onlyPrimarySort || !currentCursor.IsValid() ||
		currentCursor.Slot > latestSlot+1 && currentCursor.Reverse != colSort.Desc ||
		currentCursor.Slot < latestSlot+1 && currentCursor.Reverse == colSort.Desc {
		dutiesInfo, err := d.services.GetCurrentDutiesInfo()
		if err == nil {
			for slot, vali := range dutiesInfo.PropAssignmentsForSlot {
				// only gather scheduled slots
				if _, ok := dutiesInfo.SlotStatus[slot]; ok {
					continue
				}
				// only gather slots scheduled for our validators
				if _, ok := validatorSet[vali]; !ok {
					continue
				}
				scheduledProposers = append(scheduledProposers, dutiesInfo.PropAssignmentsForSlot[slot])
				scheduledEpochs = append(scheduledEpochs, slot/utils.Config.Chain.ClConfig.SlotsPerEpoch)
				scheduledSlots = append(scheduledSlots, slot)
			}
		} else {
			log.Debugf("duties info not available, skipping scheduled slots: %s", err)
		}
	}

	// WIP

	var proposals []struct {
		Proposer     t.VDBValidator      `db:"proposer"`
		Group        uint64              `db:"group_id"`
		Epoch        uint64              `db:"epoch"`
		Slot         uint64              `db:"slot"`
		Status       uint64              `db:"status"`
		Block        sql.NullInt64       `db:"exec_block_number"`
		FeeRecipient []byte              `db:"fee_recipient"`
		ElReward     decimal.NullDecimal `db:"el_reward"`
		ClReward     decimal.NullDecimal `db:"cl_reward"`
		GraffitiText string              `db:"graffiti_text"`

		// for cursor only
		Reward decimal.Decimal
	}

	// handle sorting
	params := make([]any, 0)
	where := ``
	orderBy := `ORDER BY `
	sortOrder := ` ASC`
	if colSort.Desc {
		sortOrder = ` DESC`
	}
	var offset any
	sortColName := `slot`
	switch colSort.Column {
	case enums.VDBBlockProposer:
		sortColName = `proposer`
		offset = currentCursor.Proposer
	case enums.VDBBlockStatus:
		sortColName = `status`
		offset = currentCursor.Status
	case enums.VDBBlockProposerReward:
		sortColName = `el_reward + cl_reward`
		offset = currentCursor.Reward
	}
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
			params = append(params, offset)
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
	secSort := `DESC`
	if !onlyPrimarySort {
		if currentCursor.IsReverse() {
			secSort = `ASC`
		}
		orderBy += `, slot ` + secSort
	}

	groupIdCol := "group_id"
	if dashboardId.Validators != nil {
		groupIdCol = fmt.Sprintf("%d AS %s", t.DefaultGroupId, groupIdCol)
	}
	selectFields := fmt.Sprintf(`
		blocks.proposer,
		blocks.epoch,
		blocks.slot,
		%s,
		blocks.status,
		exec_block_number,
		COALESCE(rb.proposer_fee_recipient, blocks.exec_fee_recipient) AS fee_recipient,
		COALESCE(rb.value / 1e18, ep.fee_recipient_reward) AS el_reward,
		cp.cl_attestations_reward / 1e9 + cp.cl_sync_aggregate_reward / 1e9 + cp.cl_slashing_inclusion_reward / 1e9 as cl_reward,
		blocks.graffiti_text`, groupIdCol)
	cte := fmt.Sprintf(`WITH past_blocks AS (SELECT
			%s
		FROM blocks
		`, selectFields)
	/*if dashboardId.Validators == nil {
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
	query += `) as u `*/
	if dashboardId.Validators == nil {
		//cte += fmt.Sprintf(`
		//INNER JOIN (%s) validators ON validators.validator_index = proposer`, filteredValidatorsQuery)
	} else {
		if len(where) == 0 {
			where += `WHERE `
		} else {
			where += `AND `
		}
		where += `proposer = ANY($1) `
	}

	params = append(params, limit+1)
	limitStr := fmt.Sprintf(`
		LIMIT $%d
	`, len(params))
	// relay bribe deduplication; select most likely (=max) relay bribe value for each block
	cte += `
	LEFT JOIN consensus_payloads cp on blocks.slot = cp.slot
	LEFT JOIN execution_payloads ep ON ep.block_hash = blocks.exec_block_hash
	LEFT JOIN LATERAL (SELECT exec_block_hash, proposer_fee_recipient, max(value) as value
        FROM relays_blocks
        WHERE relays_blocks.exec_block_hash = blocks.exec_block_hash
		GROUP BY exec_block_hash, proposer_fee_recipient
    ) rb ON rb.exec_block_hash = blocks.exec_block_hash
	)
	`

	distinct := ""
	if !onlyPrimarySort {
		distinct = sortColName
	}
	from := `past_blocks `
	selectStr := `SELECT * FROM ` + from
	if len(distinct) > 0 {
		selectStr = `SELECT DISTINCT ON (` + distinct + `) * FROM ` + from
	}

	query := selectStr + from + where + orderBy + limitStr
	// supply scheduled proposals, if any
	if len(scheduledProposers) > 0 {
		// distinct to filter out duplicates in an edge case (if dutiesInfo didn't update yet after a block was proposed, but the blocks table was)
		// might be possible to remove this once the TODO in service_slot_viz.go:startSlotVizDataService is resolved
		params = append(params, scheduledProposers)
		params = append(params, scheduledEpochs)
		params = append(params, scheduledSlots)
		cte += fmt.Sprintf(`,
		scheduled_blocks as (
			SELECT
			prov.proposer,
			prov.epoch,
			prov.slot,
			%s,
			'0'::text AS status,
			NULL::int AS exec_block_number,
			''::bytea AS fee_recipient,
			NULL::float AS el_reward,
			NULL::float AS cl_reward,
			''::text AS graffiti_text
		FROM unnest($%d::int[], $%d::int[], $%d::int[]) AS prov(proposer, epoch, slot)
		`, groupIdCol, len(params)-2, len(params)-1, len(params))
		if dashboardId.Validators == nil {
			// add group id
			cte += fmt.Sprintf(`INNER JOIN users_val_dashboards_validators validators 
			ON validators.dashboard_id = $1 
			AND validators.validator_index = ANY($%d::int[])
			`, len(params)-2)
		}
		cte += `) `
		if len(distinct) != 0 {
			distinct += ", "
		}
		// keep all ordering, sorting etc
		distinct += "slot"
		selectStr = `SELECT DISTINCT ON (` + distinct + `) * FROM `
		// encapsulate past blocks query to ensure performance
		from = `(
			( ` + query + ` )
			UNION ALL
			SELECT * FROM scheduled_blocks
		) as combined
		`
		// make sure the distinct clause filters out the correct duplicated row (e.g. block=nil)
		orderBy += `, exec_block_number NULLS LAST`
		query = selectStr + from + where + orderBy + limitStr
	}

	startTime := time.Now()
	err = d.alloyReader.SelectContext(ctx, &proposals, cte+query, params...)
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
	addressMapping := make(map[string]*t.Address)
	contractStatusRequests := make([]db.ContractInteractionAtRequest, 0, len(proposals))
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
			addressMapping[hexutil.Encode(proposal.FeeRecipient)] = nil
			contractStatusRequests = append(contractStatusRequests, db.ContractInteractionAtRequest{
				Address:  fmt.Sprintf("%x", proposal.FeeRecipient),
				Block:    proposal.Block.Int64,
				TxIdx:    -1,
				TraceIdx: -1,
			})
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
	// determine ens/names
	if err := d.GetNamesAndEnsForAddresses(ctx, addressMapping); err != nil {
		return nil, nil, err
	}
	log.Debugf("=== getting ens + labels names took %s", time.Since(startTime))
	// determine contract statuses
	contractStatuses, err := d.bigtable.GetAddressContractInteractionsAt(contractStatusRequests)
	if err != nil {
		return nil, nil, err
	}
	var contractIdx int
	for i := range data {
		if data[i].RewardRecipient != nil {
			data[i].RewardRecipient = addressMapping[string(data[i].RewardRecipient.Hash)]
			data[i].RewardRecipient.IsContract = contractStatuses[contractIdx] == types.CONTRACT_CREATION || contractStatuses[contractIdx] == types.CONTRACT_PRESENT
			contractIdx += 1
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
