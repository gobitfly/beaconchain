package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (d *DataAccessService) GetValidatorDashboardRewards(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	result := make([]t.VDBRewardsTableRow, 0)
	var paging t.Paging

	wg := errgroup.Group{}

	// Initialize the cursor
	var currentCursor t.RewardsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.RewardsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as WithdrawalsCursor: %w", err)
		}
	}

	// Prepare the sorting
	isReverseDirection := (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse())
	sortSearchDirection := ">"
	if isReverseDirection {
		sortSearchDirection = "<"
	}

	// Analyze the search term
	indexSearch := int64(-1)
	epochSearch := int64(-1)
	if search != "" {
		if strings.HasPrefix(search, "0x") && utils.IsHash(search) {
			search = strings.ToLower(search)

			// Get the current validator state to convert pubkey to index
			validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
			if err != nil {
				releaseLock()
				return nil, nil, err
			}
			if index, ok := validatorMapping.ValidatorIndices[search]; ok {
				indexSearch = int64(index)
			} else {
				// No validator index for pubkey found, return empty results
				releaseLock()
				return result, &paging, nil
			}
			releaseLock()
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			indexSearch = int64(number)
			epochSearch = int64(number)
		}
	}

	latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()
	const epochLookBack = 10

	groupIdSearchMap := make(map[uint64]bool, 0)

	// ------------------------------------------------------------------------------------------------------------------
	// Build the query that serves as base for both the main and EL rewards queries
	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("e.epoch")).
		From(goqu.L("validator_dashboard_data_epoch e")).
		Where(goqu.L("e.epoch > ?", latestFinalizedEpoch-epochLookBack))

	if dashboardId.Validators == nil {
		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("e.validator_index = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if currentCursor.IsValid() {
			condition := ""
			if currentCursor.IsReverse() {
				if currentCursor.GroupId == t.AllGroups {
					// The cursor is on the total rewards so get the data for all groups excluding the cursor epoch
					condition = fmt.Sprintf("e.epoch %s ?", sortSearchDirection)
				} else {
					// The cursor is on a specific group, get the data for the whole epoch since we could need it for the total rewards
					condition = fmt.Sprintf("e.epoch %s= ?", sortSearchDirection)
				}
				ds = ds.Where(goqu.L(condition, currentCursor.Epoch))
			} else {
				if currentCursor.GroupId == t.AllGroups {
					// The cursor is on the total rewards so get the data for all groups including the cursor epoch
					condition = fmt.Sprintf("e.epoch %s= ?", sortSearchDirection)
					ds = ds.Where(goqu.L(condition, currentCursor.Epoch))
				} else {
					// The cursor is on a specific group so get the data for groups before/after it
					condition = fmt.Sprintf("e.epoch %[1]s ? OR (e.epoch = ? AND v.group_id %[1]s ?)", sortSearchDirection)
					ds = ds.Where(goqu.L(condition, currentCursor.Epoch, currentCursor.Epoch, currentCursor.GroupId))
				}
			}
		}

		if search != "" {
			if dashboardId.AggregateGroups {
				if epochSearch == -1 && indexSearch == -1 {
					// If we have a search term but no epoch or index search then we can return empty results
					return result, &paging, nil
				}

				found := false
				if indexSearch != -1 {
					// Find whether the index is in the dashboard
					// If it is then show all the data
					err = d.alloyReader.Get(&found, `
						SELECT EXISTS(
							SELECT 1 
							FROM users_val_dashboards_validators
							WHERE dashboard_id = $1 AND validator_index = $2)
						`, dashboardId.Id, indexSearch)
					if err != nil {
						return nil, nil, err
					}
				}
				if !found && epochSearch != -1 {
					ds = ds.Where(goqu.L("e.epoch = ?", epochSearch))
				}
			} else {
				// Create a secondary query to get the group ids that match the search term
				// We cannot do everything in one query because we need to know the "epoch total" even for groups we do not search for
				groupIdDs := goqu.Dialect("postgres").
					Select(goqu.L("DISTINCT(group_id)")).
					From(goqu.L("users_val_dashboards_validators v")).
					InnerJoin(goqu.L("users_val_dashboards_groups g"), goqu.On(goqu.L("v.group_id = g.id AND v.dashboard_id = g.dashboard_id"))).
					Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

				if indexSearch != -1 {
					groupIdDs = groupIdDs.
						Where(goqu.L("(g.name ILIKE (?||'%') OR v.validator_index = ?)", search, indexSearch))
				} else {
					groupIdDs = groupIdDs.
						Where(goqu.L("g.name ILIKE (?||'%')", search))
				}

				groupIdQuery, groupIdArgs, err := groupIdDs.Prepared(true).ToSQL()
				if err != nil {
					return nil, nil, err
				}

				var groupIdSearch []uint64
				err = d.alloyReader.Select(&groupIdSearch, groupIdQuery, groupIdArgs...)
				if err != nil {
					return nil, nil, err
				}

				// Convert to a map for an easy check later
				for _, groupId := range groupIdSearch {
					groupIdSearchMap[groupId] = true
				}

				if len(groupIdSearchMap) == 0 {
					if epochSearch != -1 {
						// If we have an epoch search but no group search then we can restrict the query to the epoch
						ds = ds.Where(goqu.L("e.epoch = ?", epochSearch))
					} else {
						// No search for goup or epoch possible, return empty results
						return result, &paging, nil
					}
				}
			}
		}

		if dashboardId.AggregateGroups {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
				GroupBy(goqu.L("e.epoch"))

			if isReverseDirection {
				ds = ds.Order(goqu.L("e.epoch").Desc())
			} else {
				ds = ds.Order(goqu.L("e.epoch").Asc())
			}
		} else {
			ds = ds.
				SelectAppend(goqu.L("v.group_id AS result_group_id")).
				GroupBy(goqu.L("e.epoch"), goqu.L("result_group_id"))

			if isReverseDirection {
				ds = ds.Order(goqu.L("e.epoch").Desc(), goqu.L("result_group_id").Desc())
			} else {
				ds = ds.Order(goqu.L("e.epoch").Asc(), goqu.L("result_group_id").Asc())
			}
		}
	} else {
		// In case a list of validators is provided set the group to the default id
		ds = ds.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("e.validator_index = ANY(?)", pq.Array(dashboardId.Validators))).
			GroupBy(goqu.L("e.epoch"))

		if currentCursor.IsValid() {
			condition := fmt.Sprintf("e.epoch %s ?", sortSearchDirection)
			ds = ds.Where(goqu.L(condition, currentCursor.Epoch))
		}
		if search != "" {
			if epochSearch == -1 && indexSearch == -1 {
				// If we have a search term but no epoch or index search then we can return empty results
				return result, &paging, nil
			}

			found := false
			if indexSearch != -1 {
				// Find whether the index is in the list of validators
				// If it is then show all the data
				found = utils.ElementExists(dashboardId.Validators, t.VDBValidator(indexSearch))
			}
			if !found && epochSearch != -1 {
				ds = ds.Where(goqu.L("e.epoch = ?", epochSearch))
			}
		}

		if isReverseDirection {
			ds = ds.Order(goqu.L("e.epoch").Desc())
		} else {
			ds = ds.Order(goqu.L("e.epoch").Asc())
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main query and get the data
	queryResult := []struct {
		Epoch                 uint64 `db:"epoch"`
		GroupId               int64  `db:"result_group_id"`
		ClRewards             int64  `db:"cl_rewards"`
		AttestationsScheduled uint64 `db:"attestations_scheduled"`
		AttestationsExecuted  uint64 `db:"attestations_executed"`
		BlocksScheduled       uint64 `db:"blocks_scheduled"`
		BlocksProposed        uint64 `db:"blocks_proposed"`
		SyncScheduled         uint64 `db:"sync_scheduled"`
		SyncExecuted          uint64 `db:"sync_executed"`
		SlashedInEpoch        uint64 `db:"slashed_in_epoch"`
		SlashedAmount         uint64 `db:"slashed_amount"`
	}{}

	wg.Go(func() error {
		rewardsDs := ds.
			SelectAppend(
				goqu.L(`SUM(COALESCE(e.attestations_reward, 0) + COALESCE(e.blocks_cl_reward, 0) +
				COALESCE(e.sync_rewards, 0) + COALESCE(e.slasher_reward, 0)) AS cl_rewards`),
				goqu.L("SUM(COALESCE(e.attestations_scheduled, 0)) AS attestations_scheduled"),
				goqu.L("SUM(COALESCE(e.attestations_executed, 0)) AS attestations_executed"),
				goqu.L("SUM(COALESCE(e.blocks_scheduled, 0)) AS blocks_scheduled"),
				goqu.L("SUM(COALESCE(e.blocks_proposed, 0)) AS blocks_proposed"),
				goqu.L("SUM(COALESCE(e.sync_scheduled, 0)) AS sync_scheduled"),
				goqu.L("SUM(COALESCE(e.sync_executed, 0)) AS sync_executed"),
				goqu.L("SUM(CASE WHEN e.slashed_by IS NOT NULL THEN 1 ELSE 0 END) AS slashed_in_epoch"),
				goqu.L("SUM(COALESCE(s.slashed_amount, 0)) AS slashed_amount")).
			LeftJoin(goqu.L("validator_dashboard_data_epoch_slashedby_count s"), goqu.On(goqu.L("e.epoch = s.epoch AND e.validator_index = s.slashed_by")))

		query, args, err := rewardsDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving rewards data: %v", err)
		}
		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[uint64]map[int64]decimal.Decimal)
	wg.Go(func() error {
		elDs := ds.
			SelectAppend(
				goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS el_rewards")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("e.epoch = b.epoch AND e.validator_index = b.proposer AND b.status = '1'"))).
			LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
			LeftJoin(goqu.L("relays_blocks rb"), goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")))

		elQueryResult := []struct {
			Epoch     uint64          `db:"epoch"`
			GroupId   int64           `db:"result_group_id"`
			ElRewards decimal.Decimal `db:"el_rewards"`
		}{}

		query, args, err := elDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&elQueryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving el rewards data for rewards: %v", err)
		}

		for _, entry := range elQueryResult {
			if _, ok := elRewards[entry.Epoch]; !ok {
				elRewards[entry.Epoch] = make(map[int64]decimal.Decimal)
			}
			elRewards[entry.Epoch][entry.GroupId] = entry.ElRewards
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator dashboard rewards data: %v", err)
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Create the result without the total rewards first
	resultWoTotal := make([]t.VDBRewardsTableRow, 0)

	type TotalEpochInfo struct {
		Groups                []uint64
		ClRewards             int64
		ElRewards             decimal.Decimal
		AttestationsScheduled uint64
		AttestationsExecuted  uint64
		BlocksScheduled       uint64
		BlocksProposed        uint64
		SyncScheduled         uint64
		SyncExecuted          uint64
		Slashings             uint64
	}
	totalEpochInfo := make(map[uint64]*TotalEpochInfo, 0)

	for _, res := range queryResult {
		duty := t.VDBRewardsTableDuty{}
		if res.AttestationsScheduled > 0 {
			attestationPercentage := (float64(res.AttestationsExecuted) / float64(res.AttestationsScheduled)) * 100.0
			duty.Attestation = &attestationPercentage
		}
		if res.BlocksScheduled > 0 {
			ProposalPercentage := (float64(res.BlocksProposed) / float64(res.BlocksScheduled)) * 100.0
			duty.Proposal = &ProposalPercentage
		}
		if res.SyncScheduled > 0 {
			SyncPercentage := (float64(res.SyncExecuted) / float64(res.SyncScheduled)) * 100.0
			duty.Sync = &SyncPercentage
		}

		slashings := res.SlashedInEpoch + res.SlashedAmount
		if slashings > 0 {
			duty.Slashing = &slashings
		}

		if duty.Attestation != nil || duty.Proposal != nil || duty.Sync != nil || duty.Slashing != nil {
			// Only add groups that had some duty or got slashed
			resultWoTotal = append(resultWoTotal, t.VDBRewardsTableRow{
				Epoch:   res.Epoch,
				Duty:    duty,
				GroupId: res.GroupId,
				Reward: t.ClElValue[decimal.Decimal]{
					El: elRewards[res.Epoch][res.GroupId],
					Cl: utils.GWeiToWei(big.NewInt(res.ClRewards)),
				},
			})

			// Add it to the total epoch info
			if _, ok := totalEpochInfo[res.Epoch]; !ok {
				totalEpochInfo[res.Epoch] = &TotalEpochInfo{}
			}
			totalEpochInfo[res.Epoch].Groups = append(totalEpochInfo[res.Epoch].Groups, uint64(res.GroupId))
			totalEpochInfo[res.Epoch].ClRewards += res.ClRewards
			totalEpochInfo[res.Epoch].ElRewards = totalEpochInfo[res.Epoch].ElRewards.Add(elRewards[res.Epoch][res.GroupId])
			totalEpochInfo[res.Epoch].AttestationsScheduled += res.AttestationsScheduled
			totalEpochInfo[res.Epoch].AttestationsExecuted += res.AttestationsExecuted
			totalEpochInfo[res.Epoch].BlocksScheduled += res.BlocksScheduled
			totalEpochInfo[res.Epoch].BlocksProposed += res.BlocksProposed
			totalEpochInfo[res.Epoch].SyncScheduled += res.SyncScheduled
			totalEpochInfo[res.Epoch].SyncExecuted += res.SyncExecuted
			totalEpochInfo[res.Epoch].Slashings += slashings
		}
	}

	// Get the total rewards for the epoch if there is more than one group
	totalRewards := make(map[uint64]t.VDBRewardsTableRow, 0)
	for epoch, totalInfo := range totalEpochInfo {
		// We show the "epoch total" row if:
		// 1. There is more than one group which had duties
		// 2. If only one group had duties but we are searching for groups and the group that had duties is not in the search
		if len(totalInfo.Groups) == 1 && (len(groupIdSearchMap) == 0 || groupIdSearchMap[totalInfo.Groups[0]]) {
			continue
		}

		duty := t.VDBRewardsTableDuty{}
		if totalInfo.AttestationsScheduled > 0 {
			attestationPercentage := (float64(totalInfo.AttestationsExecuted) / float64(totalInfo.AttestationsScheduled)) * 100.0
			duty.Attestation = &attestationPercentage
		}
		if totalInfo.BlocksScheduled > 0 {
			proposalPercentage := (float64(totalInfo.BlocksProposed) / float64(totalInfo.BlocksScheduled)) * 100.0
			duty.Proposal = &proposalPercentage
		}
		if totalInfo.SyncScheduled > 0 {
			SyncPercentage := (float64(totalInfo.SyncExecuted) / float64(totalInfo.SyncScheduled)) * 100.0
			duty.Sync = &SyncPercentage
		}

		if totalInfo.Slashings > 0 {
			slashings := totalInfo.Slashings
			duty.Slashing = &slashings
		}

		totalRewards[epoch] = t.VDBRewardsTableRow{
			Epoch:   epoch,
			Duty:    duty,
			GroupId: t.AllGroups,
			Reward: t.ClElValue[decimal.Decimal]{
				El: totalInfo.ElRewards,
				Cl: utils.GWeiToWei(big.NewInt(totalInfo.ClRewards)),
			},
		}
	}

	// Reverse the data if the cursor is reversed to correct it to the requested direction
	if currentCursor.IsReverse() {
		slices.Reverse(resultWoTotal)
	}

	// Place the total rewards in the result data at the correct position and ignore group data that is not searched for
	// Ascending or descending order makes no difference but the cursor direction does
	previousEpoch := int64(-1)
	if currentCursor.IsValid() && !currentCursor.IsReverse() {
		previousEpoch = int64(currentCursor.Epoch)
	}
	for _, res := range resultWoTotal {
		if previousEpoch != int64(res.Epoch) {
			if totalReward, ok := totalRewards[res.Epoch]; ok {
				result = append(result, totalReward)
			}
		}

		// If we reach a specific group cursor which should only happen if the cursor is reversed don't include it and stop
		if currentCursor.IsReverse() && currentCursor.Epoch == res.Epoch && currentCursor.GroupId == res.GroupId {
			break
		}
		// If we don't search for specific groups or the group is in the search add the row
		if len(groupIdSearchMap) == 0 || groupIdSearchMap[uint64(res.GroupId)] {
			result = append(result, res)
		}
		previousEpoch = int64(res.Epoch)
	}

	// Flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &paging, nil
	}

	// Remove the last entries from data
	if moreDataFlag {
		if currentCursor.IsReverse() {
			result = result[len(result)-int(limit):]
		} else {
			result = result[:limit]
		}
	}

	p, err := utils.GetPagingFromData(result, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupRewards(ctx context.Context, dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error) {
	ret := &t.VDBGroupRewardsData{}

	wg := errgroup.Group{}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the query that serves as base for both the main and EL rewards queries
	ds := goqu.Dialect("postgres").
		From(goqu.L("validator_dashboard_data_epoch e")).
		Where(goqu.L("e.epoch = ?", epoch))

	// handle the case when we have a list of validators
	if len(dashboardId.Validators) > 0 {
		ds = ds.
			Where(goqu.L("e.validator_index = ANY(?)", pq.Array(dashboardId.Validators)))
	} else { // handle the case when we have a dashboard id and an optional group id
		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators AS v"), goqu.On(goqu.L("e.validator_index = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if groupId != t.AllGroups {
			ds = ds.Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main query and get the data
	queryResult := []struct {
		AttestationSourceReward      decimal.Decimal `db:"attestations_source_reward"`
		AttestationTargetReward      decimal.Decimal `db:"attestations_target_reward"`
		AttestationHeadReward        decimal.Decimal `db:"attestations_head_reward"`
		AttestationInactivitytReward decimal.Decimal `db:"attestations_inactivity_reward"`
		AttestationInclusionReward   decimal.Decimal `db:"attestations_inclusion_reward"`

		AttestationsScheduled     int64 `db:"attestations_scheduled"`
		AttestationHeadExecuted   int64 `db:"attestation_head_executed"`
		AttestationSourceExecuted int64 `db:"attestation_source_executed"`
		AttestationTargetExecuted int64 `db:"attestation_target_executed"`

		BlocksScheduled uint32          `db:"blocks_scheduled"`
		BlocksProposed  uint32          `db:"blocks_proposed"`
		BlocksClReward  decimal.Decimal `db:"blocks_cl_reward"`

		SyncScheduled uint32          `db:"sync_scheduled"`
		SyncExecuted  uint32          `db:"sync_executed"`
		SyncRewards   decimal.Decimal `db:"sync_rewards"`

		SlashedInEpoch bool            `db:"slashed_in_epoch"`
		SlashedAmount  uint32          `db:"slashed_amount"`
		SlasherRewards decimal.Decimal `db:"slasher_reward"`

		BlocksClAttestationsReward decimal.Decimal `db:"blocks_cl_attestations_reward"`
		BlockClSyncAggregateReward decimal.Decimal `db:"blocks_cl_sync_aggregate_reward"`
	}{}

	wg.Go(func() error {
		rewardsDs := ds.
			Select(
				goqu.L("COALESCE(e.attestations_source_reward, 0) AS attestations_source_reward"),
				goqu.L("COALESCE(e.attestations_target_reward, 0) AS attestations_target_reward"),
				goqu.L("COALESCE(e.attestations_head_reward, 0) AS attestations_head_reward"),
				goqu.L("COALESCE(e.attestations_inactivity_reward, 0) AS attestations_inactivity_reward"),
				goqu.L("COALESCE(e.attestations_inclusion_reward, 0) AS attestations_inclusion_reward"),
				goqu.L("COALESCE(e.attestations_scheduled, 0) AS attestations_scheduled"),
				goqu.L("COALESCE(e.attestation_head_executed, 0) AS attestation_head_executed"),
				goqu.L("COALESCE(e.attestation_source_executed, 0) AS attestation_source_executed"),
				goqu.L("COALESCE(e.attestation_target_executed, 0) AS attestation_target_executed"),
				goqu.L("COALESCE(e.blocks_scheduled, 0) AS blocks_scheduled"),
				goqu.L("COALESCE(e.blocks_proposed, 0) AS blocks_proposed"),
				goqu.L("COALESCE(e.blocks_cl_reward, 0) AS blocks_cl_reward"),
				goqu.L("COALESCE(e.sync_scheduled, 0) AS sync_scheduled"),
				goqu.L("COALESCE(e.sync_executed, 0) AS sync_executed"),
				goqu.L("COALESCE(e.sync_rewards, 0) AS sync_rewards"),
				goqu.L("e.slashed_by IS NOT NULL AS slashed_in_epoch"),
				goqu.L("COALESCE(s.slashed_amount, 0) AS slashed_amount"),
				goqu.L("COALESCE(e.slasher_reward, 0) AS slasher_reward"),
				goqu.L("COALESCE(e.blocks_cl_attestations_reward, 0) AS blocks_cl_attestations_reward"),
				goqu.L("COALESCE(e.blocks_cl_sync_aggregate_reward, 0) AS blocks_cl_sync_aggregate_reward")).
			LeftJoin(goqu.L("validator_dashboard_data_epoch_slashedby_count AS s"), goqu.On(goqu.L("e.epoch = s.epoch AND e.validator_index = s.slashed_by")))

		query, args, err := rewardsDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving group rewards data: %v", err)
		}
		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	var elRewards decimal.Decimal
	wg.Go(func() error {
		elDs := ds.
			Select(
				goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS blocks_el_reward")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("e.epoch = b.epoch AND e.validator_index = b.proposer AND b.status = '1'"))).
			LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
			LeftJoin(goqu.L("relays_blocks rb"), goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")))

		query, args, err := elDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Get(&elRewards, query, args...)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error retrieving el rewards data for group rewards: %v", err)
		}
		return nil
	})

	err := wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group rewards data: %v", err)
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Create the result
	gWei := decimal.NewFromInt(1e9)

	for _, entry := range queryResult {
		ret.AttestationsHead.Income = ret.AttestationsHead.Income.Add(entry.AttestationHeadReward.Mul(gWei))
		ret.AttestationsHead.StatusCount.Success += uint64(entry.AttestationHeadExecuted)
		ret.AttestationsHead.StatusCount.Failed += uint64(entry.AttestationsScheduled) - uint64(entry.AttestationHeadExecuted)

		ret.AttestationsSource.Income = ret.AttestationsSource.Income.Add(entry.AttestationSourceReward.Mul(gWei))
		ret.AttestationsSource.StatusCount.Success += uint64(entry.AttestationSourceExecuted)
		ret.AttestationsSource.StatusCount.Failed += uint64(entry.AttestationsScheduled) - uint64(entry.AttestationSourceExecuted)

		ret.AttestationsTarget.Income = ret.AttestationsTarget.Income.Add(entry.AttestationTargetReward.Mul(gWei))
		ret.AttestationsTarget.StatusCount.Success += uint64(entry.AttestationTargetExecuted)
		ret.AttestationsTarget.StatusCount.Failed += uint64(entry.AttestationsScheduled) - uint64(entry.AttestationTargetExecuted)

		ret.Inactivity.Income = ret.Inactivity.Income.Add(entry.AttestationInactivitytReward.Mul(gWei))
		if entry.AttestationInactivitytReward.LessThan(decimal.Zero) {
			ret.Inactivity.StatusCount.Failed++
		} else {
			ret.Inactivity.StatusCount.Success++
		}

		ret.Proposal.Income = ret.Proposal.Income.Add(entry.BlocksClReward.Mul(gWei))
		ret.Proposal.StatusCount.Success += uint64(entry.BlocksProposed)
		ret.Proposal.StatusCount.Failed += uint64(entry.BlocksScheduled) - uint64(entry.BlocksProposed)

		ret.Sync.Income = ret.Sync.Income.Add(entry.SyncRewards.Mul(gWei))
		ret.Sync.StatusCount.Success += uint64(entry.SyncExecuted)
		ret.Sync.StatusCount.Failed += uint64(entry.SyncScheduled) - uint64(entry.SyncExecuted)

		ret.Slashing.Income = ret.Slashing.Income.Add(entry.SlasherRewards.Mul(gWei))
		ret.Slashing.StatusCount.Success += uint64(entry.SlashedAmount)
		if entry.SlashedInEpoch {
			ret.Slashing.StatusCount.Failed++
		}

		ret.ProposalClAttIncReward = ret.ProposalClAttIncReward.Add(entry.BlocksClAttestationsReward.Mul(gWei))
		ret.ProposalClSyncIncReward = ret.ProposalClSyncIncReward.Add(entry.BlockClSyncAggregateReward.Mul(gWei))
		ret.ProposalClSlashingIncReward = ret.ProposalClSlashingIncReward.Add(entry.SlasherRewards.Mul(gWei))
	}

	ret.Proposal.Income = ret.Proposal.Income.Add(elRewards)
	ret.ProposalElReward = elRewards

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardRewardsChart(ctx context.Context, dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error) {
	// bar chart for the CL and EL rewards for each group for each epoch.
	// NO series for all groups combined except if AggregateGroups is true.
	// series id is group id, series property is 'cl' or 'el'

	wg := errgroup.Group{}

	latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()
	const epochLookBack = 10

	// ------------------------------------------------------------------------------------------------------------------
	// Build the query that serves as base for both the main and EL rewards queries
	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("e.epoch")).
		From(goqu.L("validator_dashboard_data_epoch e")).
		Where(goqu.L("e.epoch > ?", latestFinalizedEpoch-epochLookBack))

	if dashboardId.Validators == nil {
		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("e.validator_index = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if dashboardId.AggregateGroups {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
				GroupBy(goqu.L("e.epoch")).
				Order(goqu.L("e.epoch").Asc())
		} else {
			ds = ds.
				SelectAppend(goqu.L("v.group_id AS result_group_id")).
				GroupBy(goqu.L("e.epoch"), goqu.L("result_group_id")).
				Order(goqu.L("e.epoch").Asc(), goqu.L("result_group_id").Asc())
		}
	} else {
		// In case a list of validators is provided set the group to the default id
		ds = ds.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("e.validator_index = ANY(?)", pq.Array(dashboardId.Validators))).
			GroupBy(goqu.L("e.epoch")).
			Order(goqu.L("e.epoch").Asc())
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main query and get the data
	queryResult := []struct {
		Epoch     uint64 `db:"epoch"`
		GroupId   uint64 `db:"result_group_id"`
		ClRewards int64  `db:"cl_rewards"`
	}{}

	wg.Go(func() error {
		rewardsDs := ds.
			SelectAppend(
				goqu.L(`SUM(COALESCE(e.attestations_reward, 0) + COALESCE(e.blocks_cl_reward, 0) +
				COALESCE(e.sync_rewards, 0) + COALESCE(e.slasher_reward, 0)) AS cl_rewards`))
		query, args, err := rewardsDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving rewards chart data: %v", err)
		}
		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[uint64]map[uint64]decimal.Decimal)
	wg.Go(func() error {
		elDs := ds.
			SelectAppend(
				goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS el_rewards")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("e.epoch = b.epoch AND e.validator_index = b.proposer AND b.status = '1'"))).
			LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
			LeftJoin(goqu.L("relays_blocks rb"), goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")))

		elQueryResult := []struct {
			Epoch     uint64          `db:"epoch"`
			GroupId   uint64          `db:"result_group_id"`
			ElRewards decimal.Decimal `db:"el_rewards"`
		}{}

		query, args, err := elDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&elQueryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving el rewards data for rewards chart: %v", err)
		}

		for _, entry := range elQueryResult {
			if _, ok := elRewards[entry.Epoch]; !ok {
				elRewards[entry.Epoch] = make(map[uint64]decimal.Decimal)
			}
			elRewards[entry.Epoch][entry.GroupId] = entry.ElRewards
		}
		return nil
	})

	err := wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard rewards chart data: %v", err)
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Create a map structure to store the data
	epochData := make(map[uint64]map[uint64]t.ClElValue[decimal.Decimal])
	epochList := make([]uint64, 0)

	for _, res := range queryResult {
		if _, ok := epochData[res.Epoch]; !ok {
			epochData[res.Epoch] = make(map[uint64]t.ClElValue[decimal.Decimal])
			epochList = append(epochList, res.Epoch)
		}

		epochData[res.Epoch][res.GroupId] = t.ClElValue[decimal.Decimal]{
			El: elRewards[res.Epoch][res.GroupId],
			Cl: utils.GWeiToWei(big.NewInt(res.ClRewards)),
		}
	}

	// Get the list of groups
	// It should be identical for all epochs
	var groupList []uint64
	for _, groupData := range epochData {
		for groupId := range groupData {
			groupList = append(groupList, groupId)
		}
		break
	}
	slices.Sort(groupList)

	// Create the result
	var result t.ChartData[int, decimal.Decimal]

	// Create the series structure
	propertyNames := []string{"el", "cl"}
	for _, groupId := range groupList {
		for _, propertyName := range propertyNames {
			result.Series = append(result.Series, t.ChartSeries[int, decimal.Decimal]{
				Id:       int(groupId),
				Property: propertyName,
			})
		}
	}

	// Fill the epoch data
	for _, epoch := range epochList {
		result.Categories = append(result.Categories, epoch)
		for idx, series := range result.Series {
			if series.Property == "el" {
				result.Series[idx].Data = append(result.Series[idx].Data, epochData[epoch][uint64(series.Id)].El)
			} else if series.Property == "cl" {
				result.Series[idx].Data = append(result.Series[idx].Data, epochData[epoch][uint64(series.Id)].Cl)
			} else {
				return nil, fmt.Errorf("unknown series property: %s", series.Property)
			}
		}
	}

	return &result, nil
}

func (d *DataAccessService) GetValidatorDashboardDuties(ctx context.Context, dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	result := make([]t.VDBEpochDutiesTableRow, 0)
	var paging t.Paging

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	// Initialize the cursor
	var currentCursor t.ValidatorDutiesCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.ValidatorDutiesCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as WithdrawalsCursor: %w", err)
		}
	}

	// Prepare the sorting
	isReverseDirection := (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse())
	sortSearchDirection := ">"
	if isReverseDirection {
		sortSearchDirection = "<"
	}

	// Analyze the search term
	indexSearch := int64(-1)
	if search != "" {
		if strings.HasPrefix(search, "0x") && utils.IsHash(search) {
			search = strings.ToLower(search)

			// Get the current validator state to convert pubkey to index
			validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
			if err != nil {
				releaseLock()
				return nil, nil, err
			}
			if index, ok := validatorMapping.ValidatorIndices[search]; ok {
				indexSearch = int64(index)
			} else {
				// No validator index for pubkey found, return empty results
				releaseLock()
				return result, &paging, nil
			}
			releaseLock()
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			indexSearch = int64(number)
		} else {
			// No valid search term found, return empty results
			return result, &paging, nil
		}
	}

	// Use a subquery to allow access to total_reward in the where clause
	// rewardsQuery := fmt.Sprintf(`
	// 	SELECT *
	// 		FROM (
	// 		WITH rewards AS(
	// 			SELECT
	// 				e.validator_index,
	// 				(CAST((
	// 					COALESCE(e.attestations_reward, 0) +
	// 					COALESCE(e.blocks_cl_reward, 0) +
	// 					COALESCE(e.sync_rewards, 0) +
	// 					COALESCE(e.slasher_reward, 0)
	// 				) AS numeric) * 10^9) AS blocks_cl_reward,
	// 				COALESCE(e.attestations_scheduled, 0) AS attestations_scheduled,
	// 				COALESCE(e.attestation_source_executed, 0) AS attestation_source_executed,
	// 				COALESCE(e.attestations_source_reward, 0) AS attestations_source_reward,
	// 				COALESCE(e.attestation_target_executed, 0) AS attestation_target_executed,
	// 				COALESCE(e.attestations_target_reward, 0) AS attestations_target_reward,
	// 				COALESCE(e.attestation_head_executed, 0) AS attestation_head_executed,
	// 				COALESCE(e.attestations_head_reward, 0) AS attestations_head_reward,
	// 				COALESCE(e.sync_scheduled, 0) AS sync_scheduled,
	// 				COALESCE(e.sync_executed, 0) AS sync_executed,
	// 				COALESCE(e.sync_rewards, 0) AS sync_rewards,
	// 				e.slashed_by IS NOT NULL AS slashed_in_epoch,
	// 				COALESCE(s.slashed_amount, 0) AS slashed_amount,
	// 				COALESCE(e.slasher_reward, 0) AS slasher_reward,
	// 				COALESCE(e.blocks_scheduled, 0) AS blocks_scheduled,
	// 				COALESCE(e.blocks_proposed, 0) AS blocks_proposed,
	// 				COALESCE(r.value, ep.fee_recipient_reward * 1e18, 0) AS blocks_el_reward,
	// 				COALESCE(e.blocks_cl_attestations_reward, 0) AS blocks_cl_attestations_reward,
	// 				COALESCE(e.blocks_cl_sync_aggregate_reward, 0) AS blocks_cl_sync_aggregate_reward
	// 			FROM validator_dashboard_data_epoch e
	// 			%[1]s
	// 			%[2]s
	// 		),
	// 		el_rewards AS (
	// 			SELECT
	// 			    e.validator_index,
	// 			    SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS blocks_el_reward
	// 			FROM validator_dashboard_data_epoch e
	// 			%[1]s
	// 			LEFT JOIN blocks b ON e.epoch = b.epoch AND e.validator_index = b.proposer AND b.status = '1'
	// 			LEFT JOIN execution_payloads ep ON ep.block_hash = b.exec_block_hash
	// 			LEFT JOIN relays_blocks rb ON rb.exec_block_hash = b.exec_block_hash
	// 			%[2]s
	// 			GROUP BY e.validator_index
	// 		)
	// 		SELECT
	// 			r.validator_index,
	// 			(r.blocks_cl_reward + elr.blocks_el_reward) AS total_reward,
	// 			r.attestations_scheduled,
	// 			r.attestation_source_executed,
	// 			r.attestation_source_reward,
	// 			r.attestation_target_executed,
	// 			r.attestations_target_reward,
	// 			r.attestation_head_executed,
	// 			r.attestations_head_reward,
	// 			r.sync_scheduled,
	// 			r.sync_executed,
	// 			r.sync_rewards,
	// 			r.slashed_in_epoch,
	// 			r.slashed_amount,
	// 			r.slasher_reward,
	// 			r.blocks_scheduled,
	// 			r.blocks_proposed,
	// 			elr.blocks_el_reward,
	// 			r.blocks_cl_attestations_reward,
	// 			r.blocks_cl_sync_aggregate_reward
	// 		FROM rewards r
	// 		LEFT JOIN el_rewards elr ON r.validator_index = elr.validator_index
	// 	) AS subquery
	//  	%[3]s
	// 	%[4]s
	// 	%[5]s`, joinSubquery, whereSubquery, whereQuery, orderQuery, limitQuery)

	// ------------------------------------------------------------------------------------------------------------------
	// Build the subquery that serves as base for both the main and EL rewards subqueries
	subDs := goqu.
		Select(
			goqu.L("e.validator_index")).
		From(goqu.L("validator_dashboard_data_epoch e")).
		Where(goqu.L("e.epoch = ?", epoch))

	if dashboardId.Validators == nil {
		subDs = subDs.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("e.validator_index = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if groupId != t.AllGroups {
			subDs = subDs.Where(goqu.L("v.group_id = ?", groupId))
		}

		if indexSearch != -1 {
			subDs = subDs.Where(goqu.L("e.validator_index = ?", indexSearch))
		}
	} else {
		// In case a list of validators is provided set the group to the default id
		validators := make([]t.VDBValidator, 0)
		for _, validator := range dashboardId.Validators {
			if indexSearch == -1 || validator == t.VDBValidator(indexSearch) {
				validators = append(validators, validator)
			}
		}
		if len(validators) == 0 {
			// No validators to search for
			return result, &paging, nil
		}

		subDs = subDs.
			Where(goqu.L("e.validator_index = ANY(?)", pq.Array(validators)))
	}

	if colSort.Column == enums.VDBDutiesColumns.Validator {
		if currentCursor.IsValid() {
			// If we have a valid cursor only check the results before/after it
			condition := fmt.Sprintf("e.validator_index %s ?", sortSearchDirection)
			subDs = subDs.Where(goqu.L(condition, currentCursor.Index))
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the rewards subquery
	rewardsSubDs := subDs.
		SelectAppend(
			goqu.L(`(CAST((
						COALESCE(e.attestations_reward, 0) +
						COALESCE(e.blocks_cl_reward, 0) +
						COALESCE(e.sync_rewards, 0) +
						COALESCE(e.slasher_reward, 0)
					) AS numeric) * 10^9) AS blocks_cl_reward`),
			goqu.L("COALESCE(e.attestations_scheduled, 0) AS attestations_scheduled"),
			goqu.L("COALESCE(e.attestation_source_executed, 0) AS attestation_source_executed"),
			goqu.L("COALESCE(e.attestations_source_reward, 0) AS attestations_source_reward"),
			goqu.L("COALESCE(e.attestation_target_executed, 0) AS attestation_target_executed"),
			goqu.L("COALESCE(e.attestations_target_reward, 0) AS attestations_target_reward"),
			goqu.L("COALESCE(e.attestation_head_executed, 0) AS attestation_head_executed"),
			goqu.L("COALESCE(e.attestations_head_reward, 0) AS attestations_head_reward"),
			goqu.L("COALESCE(e.sync_scheduled, 0) AS sync_scheduled"),
			goqu.L("COALESCE(e.sync_executed, 0) AS sync_executed"),
			goqu.L("COALESCE(e.sync_rewards, 0) AS sync_rewards"),
			goqu.L("e.slashed_by IS NOT NULL AS slashed_in_epoch"),
			goqu.L("COALESCE(s.slashed_amount, 0) AS slashed_amount"),
			goqu.L("COALESCE(e.slasher_reward, 0) AS slasher_reward"),
			goqu.L("COALESCE(e.blocks_scheduled, 0) AS blocks_scheduled"),
			goqu.L("COALESCE(e.blocks_proposed, 0) AS blocks_proposed"),
			goqu.L("COALESCE(r.value, ep.fee_recipient_reward * 1e18, 0) AS blocks_el_reward"),
			goqu.L("COALESCE(e.blocks_cl_attestations_reward, 0) AS blocks_cl_attestations_reward"),
			goqu.L("COALESCE(e.blocks_cl_sync_aggregate_reward, 0) AS blocks_cl_sync_aggregate_reward")).
		Where(goqu.L(`(
			COALESCE(e.attestations_scheduled, 0) +
			COALESCE(e.sync_scheduled,0) +
			COALESCE(e.blocks_scheduled,0) +
			CASE WHEN e.slashed_by IS NOT NULL THEN 1 ELSE 0 END +
			COALESCE(s.slashed_amount, 0)
		) > 0)`)).
		LeftJoin(goqu.L("validator_dashboard_data_epoch_slashedby_count AS s"), goqu.On(goqu.L("e.epoch = s.epoch AND e.validator_index = s.slashed_by"))).
		As("rewards")

	// ------------------------------------------------------------------------------------------------------------------
	// Build the EL rewards subquery
	elRewardsSubDs := subDs.
		SelectAppend(goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS blocks_el_reward")).
		LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("e.epoch = b.epoch AND e.validator_index = b.proposer AND b.status = '1'"))).
		LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
		LeftJoin(goqu.L("relays_blocks rb"), goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash"))).
		GroupBy(goqu.L("e.validator_index")).
		As("el_rewards")

	// ------------------------------------------------------------------------------------------------------------------
	// Build the full subquery based on the two
	fullSubDs := goqu.
		Select(
			goqu.L("r.validator_index"),
			goqu.L("(r.blocks_cl_reward + elr.blocks_el_reward) AS total_reward"),
			goqu.L("r.attestations_scheduled"),
			goqu.L("r.attestation_source_executed"),
			goqu.L("r.attestation_source_reward"),
			goqu.L("r.attestation_target_executed"),
			goqu.L("r.attestations_target_reward"),
			goqu.L("r.attestation_head_executed"),
			goqu.L("r.attestations_head_reward"),
			goqu.L("r.sync_scheduled"),
			goqu.L("r.sync_executed"),
			goqu.L("r.sync_rewards"),
			goqu.L("r.slashed_in_epoch"),
			goqu.L("r.slashed_amount"),
			goqu.L("r.slasher_reward"),
			goqu.L("r.blocks_scheduled"),
			goqu.L("r.blocks_proposed"),
			goqu.L("elr.blocks_el_reward"),
			goqu.L("r.blocks_cl_attestations_reward"),
			goqu.L("r.blocks_cl_sync_aggregate_reward")).
		From(rewardsSubDs.As("r")).
		LeftJoin(elRewardsSubDs.As("elr"), goqu.On(goqu.L("r.validator_index = elr.validator_index")))

	// ------------------------------------------------------------------------------------------------------------------
	// Build the full query
	ds := goqu.Dialect("postgres").
		From(fullSubDs).
		Limit(uint(limit + 1))

	if colSort.Column == enums.VDBDutiesColumns.Validator {
		if isReverseDirection {
			ds = ds.Order(goqu.L("r.validator_index").Desc())
		} else {
			ds = ds.Order(goqu.L("r.validator_index").Asc())
		}
	} else if colSort.Column == enums.VDBDutiesColumns.Reward {
		if currentCursor.IsValid() {
			// If we have a valid cursor only check the results before/after it
			condition := fmt.Sprintf("(total_reward %[1]s ? OR (total_reward = ? AND validator_index %[1]s ?))", sortSearchDirection)
			ds = ds.Where(goqu.L(condition, currentCursor.Reward, currentCursor.Reward, currentCursor.Index))
		}

		if isReverseDirection {
			ds = ds.Order(goqu.L("total_reward").Desc(), goqu.L("r.validator_index").Desc())
		} else {
			ds = ds.Order(goqu.L("total_reward").Asc(), goqu.L("r.validator_index").Asc())
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Get the data
	queryResult := []struct {
		ValidatorIndex              uint64          `db:"validator_index"`
		TotalReward                 decimal.Decimal `db:"total_reward"`
		AttestationsScheduled       uint64          `db:"attestations_scheduled"`
		AttestationsSourceExecuted  uint64          `db:"attestation_source_executed"`
		AttestationsSourceReward    int64           `db:"attestations_source_reward"`
		AttestationsTargetExecuted  uint64          `db:"attestation_target_executed"`
		AttestationsTargetReward    int64           `db:"attestations_target_reward"`
		AttestationsHeadExecuted    uint64          `db:"attestation_head_executed"`
		AttestationsHeadReward      int64           `db:"attestations_head_reward"`
		SyncScheduled               uint64          `db:"sync_scheduled"`
		SyncExecuted                uint64          `db:"sync_executed"`
		SyncRewards                 int64           `db:"sync_rewards"`
		SlashedInEpoch              bool            `db:"slashed_in_epoch"`
		SlashedAmount               uint64          `db:"slashed_amount"`
		SlasherReward               int64           `db:"slasher_reward"`
		BlocksScheduled             uint64          `db:"blocks_scheduled"`
		BlocksProposed              uint64          `db:"blocks_proposed"`
		BlocksElReward              decimal.Decimal `db:"blocks_el_reward"`
		BlocksClAttestationsReward  int64           `db:"blocks_cl_attestations_reward"`
		BlocksClSyncAggregateReward int64           `db:"blocks_cl_sync_aggregate_reward"`
	}{}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing query: %v", err)
	}

	err = d.alloyReader.Select(&queryResult, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator rewards data: %v", err)
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Create the result
	cursorData := make([]t.ValidatorDutiesCursor, 0)
	for _, res := range queryResult {
		row := t.VDBEpochDutiesTableRow{
			Validator: res.ValidatorIndex,
		}

		// Get attestation data
		row.Duties.AttestationSource = d.getValidatorHistoryEvent(res.AttestationsSourceReward, res.AttestationsScheduled, res.AttestationsSourceExecuted)
		row.Duties.AttestationTarget = d.getValidatorHistoryEvent(res.AttestationsTargetReward, res.AttestationsScheduled, res.AttestationsTargetExecuted)
		row.Duties.AttestationHead = d.getValidatorHistoryEvent(res.AttestationsHeadReward, res.AttestationsScheduled, res.AttestationsHeadExecuted)

		// Get sync data
		row.Duties.Sync = d.getValidatorHistoryEvent(res.SyncRewards, res.SyncScheduled, res.SyncExecuted)
		row.Duties.SyncCount = res.SyncExecuted

		// Get slashing data
		if res.SlashedInEpoch || res.SlashedAmount > 0 {
			slashedEvent := t.ValidatorHistoryEvent{
				Income: utils.GWeiToWei(big.NewInt(res.SlasherReward)),
			}
			if res.SlashedInEpoch {
				if res.SlashedAmount > 0 {
					slashedEvent.Status = "partial"
				}
				slashedEvent.Status = "failed"
			} else if res.SlashedAmount > 0 {
				slashedEvent.Status = "success"
			}
			row.Duties.Slashing = &slashedEvent
		}

		// Get proposal data
		if res.BlocksScheduled > 0 {
			proposalEvent := t.ValidatorHistoryProposal{
				ElIncome:                     res.BlocksElReward,
				ClAttestationInclusionIncome: utils.GWeiToWei(big.NewInt(res.BlocksClAttestationsReward)),
				ClSyncInclusionIncome:        utils.GWeiToWei(big.NewInt(res.BlocksClSyncAggregateReward)),
				ClSlashingInclusionIncome:    utils.GWeiToWei(big.NewInt(res.SlasherReward)),
			}

			if res.BlocksProposed == 0 {
				proposalEvent.Status = "failed"
			} else if res.BlocksProposed == res.BlocksScheduled {
				proposalEvent.Status = "success"
			} else {
				proposalEvent.Status = "partial"
			}
			row.Duties.Proposal = &proposalEvent
		}

		result = append(result, row)
		cursorData = append(cursorData, t.ValidatorDutiesCursor{
			Index:  res.ValidatorIndex,
			Reward: res.TotalReward,
		})
	}

	// Flag if above limit
	moreDataFlag := len(result) > int(limit)
	if !moreDataFlag && !currentCursor.IsValid() {
		// No paging required
		return result, &paging, nil
	}

	// Remove the last entries from data
	if moreDataFlag {
		result = result[:limit]
		cursorData = cursorData[:limit]
	}

	// Reverse the data if the cursor is reversed to correct it to the requested direction
	if currentCursor.IsReverse() {
		slices.Reverse(result)
		slices.Reverse(cursorData)
	}

	p, err := utils.GetPagingFromData(cursorData, currentCursor, moreDataFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get paging: %w", err)
	}

	return result, p, nil
}

func (d *DataAccessService) getValidatorHistoryEvent(income int64, scheduledEvents, executedEvents uint64) *t.ValidatorHistoryEvent {
	if scheduledEvents > 0 {
		validatorHistoryEvent := t.ValidatorHistoryEvent{
			Income: utils.GWeiToWei(big.NewInt(income)),
		}
		if executedEvents == 0 {
			validatorHistoryEvent.Status = "failed"
		} else if executedEvents < scheduledEvents {
			validatorHistoryEvent.Status = "partial"
		} else {
			validatorHistoryEvent.Status = "success"
		}
		return &validatorHistoryEvent
	}
	return nil
}
