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
	"github.com/doug-martin/goqu/v9/exp"
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

	// TODO @LuccaBitfly move validation to handler?
	if cursor != "" {
		if currentCursor, err = utils.StringToCursor[t.BlocksCursor](cursor); err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as BlocksCursor: %w", err)
		}
	}

	searchPubkey := regexp.MustCompile(`^0x[0-9a-fA-F]{96}$`).MatchString(search)
	searchGroup := regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]+$`).MatchString(search)
	searchIndex := regexp.MustCompile(`^[0-9]+$`).MatchString(search)

	validators := goqu.T("validators") // could adapt data type to make handling as table/alias less confusing
	blocks := goqu.T("blocks")
	groups := goqu.T("groups")

	type validatorGroup struct {
		Validator t.VDBValidator `db:"validator_index"`
		Group     uint64         `db:"group_id"`
	}

	// -------------------------------------
	// Goqu Query to determine validators filtered by search
	var filteredValidatorsDs *goqu.SelectDataset
	var filteredValidators []validatorGroup

	filteredValidatorsDs = goqu.Dialect("postgres").
		Select(
			"validator_index",
		)
	if dashboardId.Validators == nil {
		filteredValidatorsDs = filteredValidatorsDs.
			From(goqu.T("users_val_dashboards_validators").As(validators.GetTable())).
			Where(validators.Col("dashboard_id").Eq(dashboardId.Id))
		// apply search filters
		searches := []exp.Expression{}
		if searchIndex {
			searches = append(searches, validators.Col("validator_index").Eq(search))
		}
		if searchGroup {
			filteredValidatorsDs = filteredValidatorsDs.
				InnerJoin(goqu.T("users_val_dashboards_groups").As(groups), goqu.On(
					validators.Col("group_id").Eq(groups.Col("id")),
					validators.Col("dashboard_id").Eq(groups.Col("dashboard_id")),
				))
			searches = append(searches,
				goqu.L("LOWER(?)", groups.Col("name")).Like(strings.Replace(strings.ToLower(search), "_", "\\_", -1)+"%"),
			)
		}
		if searchPubkey {
			index, ok := validatorMapping.ValidatorIndices[search]
			if !ok && !searchGroup && !searchIndex {
				// searched pubkey doesn't exist, don't even need to query anything
				return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
			}
			searches = append(searches,
				validators.Col("validator_index").Eq(index),
			)
		}
		if len(searches) > 0 {
			filteredValidatorsDs = filteredValidatorsDs.Where(goqu.Or(searches...))
		}
	} else {
		validatorList := make([]t.VDBValidator, 0, len(dashboardId.Validators))
		for _, validator := range dashboardId.Validators {
			if searchIndex && fmt.Sprint(validator) != search ||
				searchPubkey && validator != validatorMapping.ValidatorIndices[search] {
				continue
			}
			filteredValidators = append(filteredValidators, validatorGroup{
				Validator: validator,
				Group:     t.DefaultGroupId,
			})
			validatorList = append(validatorList, validator)
			if searchIndex || searchPubkey {
				break
			}
		}
		filteredValidatorsDs = filteredValidatorsDs.
			From(
				goqu.Dialect("postgres").
					From(
						goqu.L("unnest(?::int[])", pq.Array(validatorList)).As("validator_index"),
					).
					As(validators.GetTable()),
			)
	}

	// -------------------------------------
	// Constuct final query
	var blocksDs *goqu.SelectDataset

	// 1. Tables
	blocksDs = filteredValidatorsDs.
		InnerJoin(blocks, goqu.On(
			blocks.Col("proposer").Eq(validators.Col("validator_index")),
		)).
		LeftJoin(goqu.T("consensus_payloads").As("cp"), goqu.On(
			blocks.Col("slot").Eq(goqu.I("cp.slot")),
		)).
		LeftJoin(goqu.T("execution_payloads").As("ep"), goqu.On(
			blocks.Col("exec_block_hash").Eq(goqu.I("ep.block_hash")),
		)).
		LeftJoin(
			// relay bribe deduplication; select most likely (=max) relay bribe value for each block
			goqu.Lateral(goqu.Dialect("postgres").
				From(goqu.T("relays_blocks")).
				Select(
					goqu.I("relays_blocks.exec_block_hash"),
					goqu.I("relays_blocks.proposer_fee_recipient"),
					goqu.MAX(goqu.I("relays_blocks.value")).As("value")).
				GroupBy(
					"exec_block_hash",
					"proposer_fee_recipient",
				)).As("rb"),
			goqu.On(
				goqu.I("rb.exec_block_hash").Eq(blocks.Col("exec_block_hash")),
			),
		)

	// 2. Selects
	groupIdQ := goqu.C("group_id").(exp.Aliaseable)
	if dashboardId.Validators != nil {
		groupIdQ = exp.NewLiteralExpression("?::int", t.DefaultGroupId)
	}
	groupId := groupIdQ.As("group_id")

	blocksDs = blocksDs.
		SelectAppend(
			blocks.Col("epoch"),
			blocks.Col("slot"),
			groupId,
			blocks.Col("status"),
			blocks.Col("exec_block_number"),
			blocks.Col("graffiti_text"),
			goqu.COALESCE(goqu.I("rb.proposer_fee_recipient"), blocks.Col("exec_fee_recipient")).As("fee_recipient"),
			goqu.COALESCE(goqu.L("rb.value / 1e18"), goqu.I("ep.fee_recipient_reward")).As("el_reward"),
			goqu.L("cp.cl_attestations_reward / 1e9 + cp.cl_sync_aggregate_reward / 1e9 + cp.cl_slashing_inclusion_reward / 1e9").As("cl_reward"),
		)

	// 3. Sorting and pagination
	defaultColumns := []t.SortColumn{
		{Column: enums.VDBBlocksColumns.Slot.ToExpr(), Desc: true, Offset: currentCursor.Slot},
	}
	var offset any
	switch colSort.Column {
	case enums.VDBBlocksColumns.Proposer:
		offset = currentCursor.Proposer
	case enums.VDBBlocksColumns.Block:
		offset = currentCursor.Block
		if !currentCursor.Block.Valid {
			offset = nil
		}
	case enums.VDBBlocksColumns.Status:
		offset = fmt.Sprintf("%d", currentCursor.Status) // type of 'status' column is text for some reason
	case enums.VDBBlocksColumns.ProposerReward:
		offset = currentCursor.Reward
	}

	order, directions := applySortAndPagination(defaultColumns, t.SortColumn{Column: colSort.Column.ToExpr(), Desc: colSort.Desc, Offset: offset}, currentCursor.GenericCursor)
	blocksDs = goqu.Dialect("postgres").From(goqu.T("past_blocks_cte")).
		With("past_blocks_cte", blocksDs). // encapsulate so we can use selected fields
		Order(order...)
	if directions != nil {
		blocksDs = blocksDs.Where(directions)
	}

	// 4. Limit
	blocksDs = blocksDs.Limit(uint(limit + 1))

	// 5. Gather and supply scheduled blocks to let db do the sorting etc
	latestSlot := cache.LatestSlot.Get()
	onlyPrimarySort := colSort.Column == enums.VDBBlockSlot
	if !(onlyPrimarySort || colSort.Column == enums.VDBBlockBlock) ||
		!currentCursor.IsValid() ||
		currentCursor.Slot > latestSlot+1 ||
		colSort.Desc == currentCursor.Reverse {
		dutiesInfo, err := d.services.GetCurrentDutiesInfo()
		if err == nil {
			if dashboardId.Validators == nil {
				// fetch filtered validators if not done yet
				filteredValidatorsDs = filteredValidatorsDs.
					SelectAppend(groupIdQ)
				validatorsQuery, validatorsArgs, err := filteredValidatorsDs.Prepared(true).ToSQL()
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

			validatorSet := make(map[t.VDBValidator]uint64)
			for _, v := range filteredValidators {
				validatorSet[v.Validator] = v.Group
			}
			var scheduledProposers []t.VDBValidator
			var scheduledGroups []uint64
			var scheduledEpochs []uint64
			var scheduledSlots []uint64
			// don't need if requested slots are in the past
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
				scheduledGroups = append(scheduledGroups, validatorSet[vali])
				scheduledEpochs = append(scheduledEpochs, slot/utils.Config.Chain.ClConfig.SlotsPerEpoch)
				scheduledSlots = append(scheduledSlots, slot)
			}

			scheduledDs := goqu.Dialect("postgres").
				From(
					goqu.L("unnest(?::int[], ?::int[], ?::int[], ?::int[]) AS prov(validator_index, group_id, epoch, slot)", pq.Array(scheduledProposers), pq.Array(scheduledGroups), pq.Array(scheduledEpochs), pq.Array(scheduledSlots)),
				).
				Select(
					goqu.C("validator_index"),
					goqu.C("epoch"),
					goqu.C("slot"),
					groupId,
					goqu.V("0").As("status"),
					goqu.V(nil).As("exec_block_number"),
					goqu.V(nil).As("graffiti_text"),
					goqu.V(nil).As("fee_recipient"),
					goqu.V(nil).As("el_reward"),
					goqu.V(nil).As("cl_reward"),
				).
				As("scheduled_blocks")

			// Supply to result query
			// distinct + block number ordering to filter out duplicates in an edge case (if dutiesInfo didn't update yet after a block was proposed, but the blocks table was)
			// might be possible to remove this once the TODO in service_slot_viz.go:startSlotVizDataService is resolved
			blocksDs = goqu.Dialect("Postgres").
				From(blocksDs.Union(scheduledDs)). // wrap union to apply order
				Order(order...).
				OrderAppend(goqu.C("exec_block_number").Desc().NullsLast()).
				Limit(uint(limit + 1)).
				Distinct(enums.VDBBlocksColumns.Slot.ToExpr())
			if directions != nil {
				blocksDs = blocksDs.Where(directions)
			}
			if !onlyPrimarySort {
				blocksDs = blocksDs.
					Distinct(colSort.Column.ToExpr(), enums.VDBBlocksColumns.Slot.ToExpr())
			}
		} else {
			log.Warnf("Error getting scheduled proposals, DutiesInfo not available in Redis: %s", err)
		}
	}

	// -------------------------------------
	// Execute query
	var proposals []struct {
		Proposer     t.VDBValidator      `db:"validator_index"`
		Group        uint64              `db:"group_id"`
		Epoch        uint64              `db:"epoch"`
		Slot         uint64              `db:"slot"`
		Status       uint64              `db:"status"`
		Block        sql.NullInt64       `db:"exec_block_number"`
		FeeRecipient []byte              `db:"fee_recipient"`
		ElReward     decimal.NullDecimal `db:"el_reward"`
		ClReward     decimal.NullDecimal `db:"cl_reward"`
		GraffitiText sql.NullString      `db:"graffiti_text"`

		// for cursor only
		Reward decimal.Decimal
	}
	startTime := time.Now()
	query, args, err := blocksDs.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, err
	}
	err = d.alloyReader.SelectContext(ctx, &proposals, query, args...)
	log.Debugf("=== getting past blocks took %s", time.Since(startTime))
	if err != nil {
		return nil, nil, err
	}
	if len(proposals) == 0 {
		return make([]t.VDBBlocksTableRow, 0), &t.Paging{}, nil
	}

	// -------------------------------------
	// Prepare result
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
		if proposal.GraffitiText.Valid {
			graffiti := proposal.GraffitiText.String
			data[i].Graffiti = &graffiti
		}
		if proposal.Block.Valid {
			block := uint64(proposal.Block.Int64)
			data[i].Block = &block
		}
		if proposal.Status == 3 {
			continue
		}
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
