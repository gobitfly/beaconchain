package dataaccess

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (d *DataAccessService) GetValidatorDashboardRewards(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	result := make([]t.VDBRewardsTableRow, 0)
	var paging t.Paging

	wg := errgroup.Group{}

	// Initialize the cursor
	var currentCursor t.RewardsCursor
	var err error
	if cursor != "" {
		currentCursor, err = utils.StringToCursor[t.RewardsCursor](cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse passed cursor as RewardsCursor: %w", err)
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

			validatorMapping, err := d.services.GetCurrentValidatorMapping()
			if err != nil {
				return nil, nil, err
			}

			if index, ok := validatorMapping.ValidatorIndices[search]; ok {
				indexSearch = int64(index)
			} else {
				// No validator index for pubkey found, return empty results
				return result, &paging, nil
			}
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			indexSearch = int64(number)
			epochSearch = int64(number)
		}
	}

	latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()
	const epochLookBack = 9
	startEpoch := uint64(0)
	if latestFinalizedEpoch > epochLookBack {
		startEpoch = latestFinalizedEpoch - epochLookBack
	}

	groupIdSearchMap := make(map[uint64]bool, 0)

	// ------------------------------------------------------------------------------------------------------------------
	// Get rocketpool minipool infos if needed
	var rpInfos *t.RPInfo
	if protocolModes.RocketPool {
		rpInfos, err = d.getRocketPoolInfos(ctx, dashboardId, t.AllGroups)
		if err != nil {
			return nil, nil, err
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main (CL) and EL rewards queries
	rewardsDs := goqu.Dialect("postgres").
		From(goqu.L("validator_dashboard_data_epoch e")).
		With("validators", goqu.L("(SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
		Select(
			goqu.L("e.epoch"),
			goqu.L("e.validator_index"),
			goqu.L(`(e.attestations_reward + e.blocks_cl_reward + e.sync_rewards) AS cl_rewards`),
			goqu.L("e.attestations_scheduled"),
			goqu.L("e.attestations_executed"),
			goqu.L("e.blocks_scheduled"),
			goqu.L("e.blocks_proposed"),
			goqu.L("e.sync_scheduled"),
			goqu.L("e.sync_executed"),
			goqu.L("e.slashed"),
			goqu.L("e.blocks_slashing_count")).
		Where(goqu.L("e.epoch_timestamp >= fromUnixTimestamp(?)", utils.EpochToTime(startEpoch).Unix()))

	elDs := goqu.Dialect("postgres").
		Select(
			goqu.L("b.epoch"),
			goqu.L("b.proposer"),
			goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS el_rewards")).
		From(goqu.L("users_val_dashboards_validators v")).
		Where(goqu.L("b.epoch >= ?", startEpoch)).
		LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("v.validator_index = b.proposer AND b.status = '1'"))).
		LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
		LeftJoin(
			goqu.Lateral(goqu.Dialect("postgres").
				From("relays_blocks").
				Select(
					goqu.L("exec_block_hash"),
					goqu.L("proposer_fee_recipient"),
					goqu.MAX("value").As("value")).
				Where(goqu.L("relays_blocks.exec_block_hash = b.exec_block_hash")).
				GroupBy("exec_block_hash", "proposer_fee_recipient")).As("rb"),
			goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")),
		).
		GroupBy(goqu.L("b.epoch"), goqu.L("b.proposer"))

	if rpInfos != nil && protocolModes.RocketPool {
		// Exclude rewards that went to the smoothing pool
		elDs = elDs.
			Where(goqu.L("(b.exec_fee_recipient != ? OR (rb.proposer_fee_recipient IS NOT NULL AND rb.proposer_fee_recipient != ?))", rpInfos.SmoothingPoolAddress, rpInfos.SmoothingPoolAddress))
	}

	if dashboardId.Validators == nil {
		rewardsDs = rewardsDs.
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("e.validator_index = v.validator_index"))).
			Where(goqu.L("e.validator_index IN (SELECT validator_index FROM validators)"))
		elDs = elDs.
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))
		if currentCursor.IsValid() {
			if currentCursor.IsReverse() {
				if currentCursor.GroupId == t.AllGroups {
					// The cursor is on the total rewards so get the data for all groups excluding the cursor epoch
					rewardsDs = rewardsDs.Where(goqu.L(fmt.Sprintf("e.epoch_timestamp %s fromUnixTimestamp(?)", sortSearchDirection), utils.EpochToTime(currentCursor.Epoch).Unix()))
					elDs = elDs.Where(goqu.L(fmt.Sprintf("b.epoch %s ?", sortSearchDirection), currentCursor.Epoch))
				} else {
					// The cursor is on a specific group, get the data for the whole epoch since we could need it for the total rewards
					rewardsDs = rewardsDs.Where(goqu.L(fmt.Sprintf("e.epoch_timestamp %s= fromUnixTimestamp(?)", sortSearchDirection), utils.EpochToTime(currentCursor.Epoch).Unix()))
					elDs = elDs.Where(goqu.L(fmt.Sprintf("b.epoch %s= ?", sortSearchDirection), currentCursor.Epoch))
				}
			} else {
				if currentCursor.GroupId == t.AllGroups {
					// The cursor is on the total rewards so get the data for all groups including the cursor epoch
					rewardsDs = rewardsDs.Where(goqu.L(fmt.Sprintf("e.epoch_timestamp %s= fromUnixTimestamp(?)", sortSearchDirection), utils.EpochToTime(currentCursor.Epoch).Unix()))
					elDs = elDs.Where(goqu.L(fmt.Sprintf("b.epoch %s= ?", sortSearchDirection), currentCursor.Epoch))
				} else {
					// The cursor is on a specific group so get the data for groups before/after it
					rewardsDs = rewardsDs.Where(goqu.L(fmt.Sprintf("(e.epoch_timestamp %[1]s fromUnixTimestamp(?) OR (e.epoch_timestamp = fromUnixTimestamp(?) AND v.group_id %[1]s ?))", sortSearchDirection),
						utils.EpochToTime(currentCursor.Epoch).Unix(), utils.EpochToTime(currentCursor.Epoch).Unix(), currentCursor.GroupId))
					elDs = elDs.Where(goqu.L(fmt.Sprintf("(b.epoch %[1]s ? OR (b.epoch = ? AND v.group_id %[1]s ?))", sortSearchDirection),
						currentCursor.Epoch, currentCursor.Epoch, currentCursor.GroupId))
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
					err = d.readerDb.GetContext(ctx, &found, `
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
					rewardsDs = rewardsDs.Where(goqu.L("e.epoch_timestamp = fromUnixTimestamp(?)", utils.EpochToTime(uint64(epochSearch)).Unix()))
					elDs = elDs.Where(goqu.L("b.epoch = ?", epochSearch))
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
				err = d.readerDb.SelectContext(ctx, &groupIdSearch, groupIdQuery, groupIdArgs...)
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
						rewardsDs = rewardsDs.Where(goqu.L("e.epoch_timestamp = fromUnixTimestamp(?)", utils.EpochToTime(uint64(epochSearch)).Unix()))
						elDs = elDs.Where(goqu.L("b.epoch = ?", epochSearch))
					} else {
						// No search for goup or epoch possible, return empty results
						return result, &paging, nil
					}
				}
			}
		}

		if dashboardId.AggregateGroups {
			rewardsDs = rewardsDs.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))

			elDs = elDs.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))

			if isReverseDirection {
				rewardsDs = rewardsDs.Order(goqu.L("e.epoch").Desc())
				elDs = elDs.Order(goqu.L("b.epoch").Desc())
			} else {
				rewardsDs = rewardsDs.Order(goqu.L("e.epoch").Asc())
				elDs = elDs.Order(goqu.L("b.epoch").Asc())
			}
		} else {
			rewardsDs = rewardsDs.
				SelectAppend(goqu.L("v.group_id AS result_group_id"))
			elDs = elDs.
				SelectAppend(goqu.L("v.group_id AS result_group_id")).
				GroupByAppend(goqu.L("result_group_id"))

			if isReverseDirection {
				rewardsDs = rewardsDs.Order(goqu.L("e.epoch").Desc(), goqu.L("result_group_id").Desc())
				elDs = elDs.Order(goqu.L("b.epoch").Desc(), goqu.L("result_group_id").Desc())
			} else {
				rewardsDs = rewardsDs.Order(goqu.L("e.epoch").Asc(), goqu.L("result_group_id").Asc())
				elDs = elDs.Order(goqu.L("b.epoch").Asc(), goqu.L("result_group_id").Asc())
			}
		}
	} else {
		// In case a list of validators is provided set the group to the default id
		rewardsDs = rewardsDs.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("e.validator_index IN ?", dashboardId.Validators))
		elDs = elDs.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("b.proposer = ANY(?)", pq.Array(dashboardId.Validators)))

		if currentCursor.IsValid() {
			rewardsDs = rewardsDs.Where(goqu.L(fmt.Sprintf("e.epoch_timestamp %s fromUnixTimestamp(?)", sortSearchDirection), utils.EpochToTime(currentCursor.Epoch).Unix()))
			elDs = elDs.Where(goqu.L(fmt.Sprintf("b.epoch %s ?", sortSearchDirection), currentCursor.Epoch))
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
				rewardsDs = rewardsDs.Where(goqu.L("e.epoch_timestamp = fromUnixTimestamp(?)", utils.EpochToTime(uint64(epochSearch)).Unix()))
				elDs = elDs.Where(goqu.L("b.epoch = ?", epochSearch))
			}
		}

		if isReverseDirection {
			rewardsDs = rewardsDs.Order(goqu.L("e.epoch").Desc())
			elDs = elDs.Order(goqu.L("b.epoch").Desc())
		} else {
			rewardsDs = rewardsDs.Order(goqu.L("e.epoch").Asc())
			elDs = elDs.Order(goqu.L("b.epoch").Asc())
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main query and get the data

	type QueryResultSum struct {
		Epoch                 uint64
		GroupId               int64
		ClRewards             decimal.Decimal
		AttestationsScheduled uint64
		AttestationsExecuted  uint64
		BlocksScheduled       uint64
		BlocksProposed        uint64
		SyncScheduled         uint64
		SyncExecuted          uint64
		Slashed               uint64
		BlocksSlashingCount   uint64
	}
	var queryResultSum []QueryResultSum
	smoothingPoolRewards := make(map[uint64]map[int64]decimal.Decimal, 0) // epoch -> group -> reward

	wg.Go(func() error {
		type QueryResult struct {
			Epoch                 uint64 `db:"epoch"`
			GroupId               int64  `db:"result_group_id"`
			ValidatorIndex        uint64 `db:"validator_index"`
			ClRewards             int64  `db:"cl_rewards"`
			AttestationsScheduled uint64 `db:"attestations_scheduled"`
			AttestationsExecuted  uint64 `db:"attestations_executed"`
			BlocksScheduled       uint64 `db:"blocks_scheduled"`
			BlocksProposed        uint64 `db:"blocks_proposed"`
			SyncScheduled         uint64 `db:"sync_scheduled"`
			SyncExecuted          uint64 `db:"sync_executed"`
			Slashed               bool   `db:"slashed"`
			BlocksSlashingCount   uint64 `db:"blocks_slashing_count"`
		}
		var queryResult []QueryResult

		query, args, err := rewardsDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.clickhouseReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving rewards data: %w", err)
		}

		validatorGroupMap := make(map[uint64]int64)
		for _, row := range queryResult {
			if len(queryResultSum) == 0 ||
				queryResultSum[len(queryResultSum)-1].Epoch != row.Epoch ||
				queryResultSum[len(queryResultSum)-1].GroupId != row.GroupId {
				queryResultSum = append(queryResultSum, QueryResultSum{
					Epoch:   row.Epoch,
					GroupId: row.GroupId,
				})
			}

			validatorGroupMap[row.ValidatorIndex] = row.GroupId

			current := &queryResultSum[len(queryResultSum)-1]

			current.AttestationsScheduled += row.AttestationsScheduled
			current.AttestationsExecuted += row.AttestationsExecuted
			current.BlocksScheduled += row.BlocksScheduled
			current.BlocksProposed += row.BlocksProposed
			current.SyncScheduled += row.SyncScheduled
			current.SyncExecuted += row.SyncExecuted
			if row.Slashed {
				current.Slashed++
			}
			current.BlocksSlashingCount += row.BlocksSlashingCount

			reward := utils.GWeiToWei(big.NewInt(row.ClRewards))
			if rpInfos != nil && protocolModes.RocketPool {
				if rpValidator, ok := rpInfos.Minipool[row.ValidatorIndex]; ok {
					reward = reward.Mul(d.getRocketPoolOperatorFactor(rpValidator))
				}
			}
			current.ClRewards = current.ClRewards.Add(reward)
		}

		// Calculate smoothing pool rewards
		// Has to be done here in the cl and not el part because here we have the list of all relevant validators
		if rpInfos != nil && protocolModes.RocketPool {
			for validatorIndex, groupId := range validatorGroupMap {
				for epoch, reward := range rpInfos.Minipool[validatorIndex].SmoothingPoolRewards {
					if _, ok := smoothingPoolRewards[epoch]; !ok {
						smoothingPoolRewards[epoch] = make(map[int64]decimal.Decimal)
					}
					smoothingPoolRewards[epoch][groupId] = smoothingPoolRewards[epoch][groupId].Add(reward)
				}
			}
		}

		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[uint64]map[int64]decimal.Decimal)
	wg.Go(func() error {
		elQueryResult := []struct {
			Proposer  uint64          `db:"proposer"`
			Epoch     uint64          `db:"epoch"`
			GroupId   int64           `db:"result_group_id"`
			ElRewards decimal.Decimal `db:"el_rewards"`
		}{}

		query, args, err := elDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.readerDb.SelectContext(ctx, &elQueryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving el rewards data for rewards: %w", err)
		}

		for _, entry := range elQueryResult {
			if _, ok := elRewards[entry.Epoch]; !ok {
				elRewards[entry.Epoch] = make(map[int64]decimal.Decimal)
			}

			reward := entry.ElRewards
			if rpInfos != nil && protocolModes.RocketPool {
				if rpValidator, ok := rpInfos.Minipool[entry.Proposer]; ok {
					reward = reward.Mul(d.getRocketPoolOperatorFactor(rpValidator))
				}
			}
			elRewards[entry.Epoch][entry.GroupId] = elRewards[entry.Epoch][entry.GroupId].Add(reward)
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator dashboard rewards data: %w", err)
	}

	// Add smoothing pool rewards to el rewards
	if rpInfos != nil && protocolModes.RocketPool {
		for epoch, groupRewards := range smoothingPoolRewards {
			for groupId, reward := range groupRewards {
				if _, ok := elRewards[epoch]; !ok {
					elRewards[epoch] = make(map[int64]decimal.Decimal)
				}
				elRewards[epoch][groupId] = elRewards[epoch][groupId].Add(reward)
			}
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Create the result without the total rewards first
	resultWoTotal := make([]t.VDBRewardsTableRow, 0)

	type TotalEpochInfo struct {
		Groups                []uint64
		ClRewards             decimal.Decimal
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
	for _, res := range queryResultSum {
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

		slashings := res.Slashed + res.BlocksSlashingCount
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
					Cl: res.ClRewards,
				},
			})

			// Add it to the total epoch info
			if _, ok := totalEpochInfo[res.Epoch]; !ok {
				totalEpochInfo[res.Epoch] = &TotalEpochInfo{}
			}
			totalEpochInfo[res.Epoch].Groups = append(totalEpochInfo[res.Epoch].Groups, uint64(res.GroupId))
			totalEpochInfo[res.Epoch].ClRewards = totalEpochInfo[res.Epoch].ClRewards.Add(res.ClRewards)
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
				Cl: totalInfo.ClRewards,
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

func (d *DataAccessService) GetValidatorDashboardGroupRewards(ctx context.Context, dashboardId t.VDBId, groupId int64, epoch uint64, protocolModes t.VDBProtocolModes) (*t.VDBGroupRewardsData, error) {
	ret := &t.VDBGroupRewardsData{}

	wg := errgroup.Group{}
	var err error

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Get rocketpool minipool infos if needed
	var rpInfos *t.RPInfo
	if protocolModes.RocketPool {
		rpInfos, err = d.getRocketPoolInfos(ctx, dashboardId, groupId)
		if err != nil {
			return nil, err
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main and EL rewards queries
	rewardsDs := goqu.Dialect("postgres").
		From(goqu.L("validator_dashboard_data_epoch e")).
		With("validators", goqu.L("(SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
		Select(
			goqu.L("e.validator_index"),
			goqu.L("e.attestations_source_reward"),
			goqu.L("e.attestations_target_reward"),
			goqu.L("e.attestations_head_reward"),
			goqu.L("e.attestations_inactivity_reward"),
			goqu.L("e.attestations_inclusion_reward"),
			goqu.L("e.attestations_scheduled"),
			goqu.L("e.attestation_head_executed"),
			goqu.L("e.attestation_source_executed"),
			goqu.L("e.attestation_target_executed"),
			goqu.L("e.blocks_scheduled"),
			goqu.L("e.blocks_proposed"),
			goqu.L("e.blocks_cl_reward"),
			goqu.L("e.sync_scheduled"),
			goqu.L("e.sync_executed"),
			goqu.L("e.sync_rewards"),
			goqu.L("e.slashed"),
			goqu.L("e.blocks_slashing_count"),
			goqu.L("e.blocks_cl_slasher_reward"),
			goqu.L("e.blocks_cl_attestations_reward"),
			goqu.L("e.blocks_cl_sync_aggregate_reward")).
		Where(goqu.L("e.epoch_timestamp = fromUnixTimestamp(?)", utils.EpochToTime(epoch).Unix()))

	elDs := goqu.Dialect("postgres").
		Select(
			goqu.L("b.proposer"),
			goqu.L("COALESCE(SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)), 0) AS blocks_el_reward")).
		From(goqu.L("users_val_dashboards_validators v")).
		LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("v.validator_index = b.proposer AND b.status = '1'"))).
		LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
		LeftJoin(
			goqu.Lateral(goqu.Dialect("postgres").
				From("relays_blocks").
				Select(
					goqu.L("exec_block_hash"),
					goqu.L("proposer_fee_recipient"),
					goqu.MAX("value").As("value")).
				Where(goqu.L("relays_blocks.exec_block_hash = b.exec_block_hash")).
				GroupBy("exec_block_hash", "proposer_fee_recipient")).As("rb"),
			goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")),
		).
		Where(goqu.L("b.epoch = ?", epoch)).
		GroupBy(goqu.L("b.proposer"))

	if rpInfos != nil && protocolModes.RocketPool {
		// Exclude rewards that went to the smoothing pool
		elDs = elDs.
			Where(goqu.L("(b.exec_fee_recipient != ? OR (rb.proposer_fee_recipient IS NOT NULL AND rb.proposer_fee_recipient != ?))", rpInfos.SmoothingPoolAddress, rpInfos.SmoothingPoolAddress))
	}

	if dashboardId.Validators == nil {
		// handle the case when we have a dashboard id and an optional group id
		rewardsDs = rewardsDs.
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("e.validator_index = v.validator_index"))).
			Where(goqu.L("e.validator_index IN (SELECT validator_index FROM validators)"))
		elDs = elDs.
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))
		if groupId != t.AllGroups {
			rewardsDs = rewardsDs.Where(goqu.L("v.group_id = ?", groupId))
			elDs = elDs.Where(goqu.L("v.group_id = ?", groupId))
		}
	} else {
		// handle the case when we have a list of validators
		rewardsDs = rewardsDs.
			Where(goqu.L("e.validator_index IN ?", dashboardId.Validators))
		elDs = elDs.
			Where(goqu.L("b.proposer = ANY(?)", pq.Array(dashboardId.Validators)))
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main query and get the data
	queryResult := []struct {
		ValidatorIndex uint64 `db:"validator_index"`

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

		Slashed             bool   `db:"slashed"`
		BlocksSlashingCount uint32 `db:"blocks_slashing_count"`

		BlocksClSlasherReward      decimal.Decimal `db:"blocks_cl_slasher_reward"`
		BlocksClAttestationsReward decimal.Decimal `db:"blocks_cl_attestations_reward"`
		BlockClSyncAggregateReward decimal.Decimal `db:"blocks_cl_sync_aggregate_reward"`
	}{}

	wg.Go(func() error {
		query, args, err := rewardsDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.clickhouseReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving group rewards data: %w", err)
		}
		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[uint64]decimal.Decimal)
	wg.Go(func() error {
		elQueryResult := []struct {
			Proposer  uint64          `db:"proposer"`
			ElRewards decimal.Decimal `db:"blocks_el_reward"`
		}{}

		query, args, err := elDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.readerDb.SelectContext(ctx, &elQueryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving el rewards data for group rewards: %w", err)
		}

		for _, entry := range elQueryResult {
			elRewards[entry.Proposer] = entry.ElRewards
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group rewards data: %w", err)
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Create the result
	gWei := decimal.NewFromInt(1e9)

	for _, entry := range queryResult {
		rpFactor := decimal.NewFromInt(1)
		if rpInfos != nil && protocolModes.RocketPool {
			if rpValidator, ok := rpInfos.Minipool[entry.ValidatorIndex]; ok {
				rpFactor = d.getRocketPoolOperatorFactor(rpValidator)
			}
		}

		ret.AttestationsHead.Income = ret.AttestationsHead.Income.Add(entry.AttestationHeadReward.Mul(gWei).Mul(rpFactor))
		ret.AttestationsHead.StatusCount.Success += uint64(entry.AttestationHeadExecuted)
		ret.AttestationsHead.StatusCount.Failed += uint64(entry.AttestationsScheduled) - uint64(entry.AttestationHeadExecuted)

		ret.AttestationsSource.Income = ret.AttestationsSource.Income.Add(entry.AttestationSourceReward.Mul(gWei).Mul(rpFactor))
		ret.AttestationsSource.StatusCount.Success += uint64(entry.AttestationSourceExecuted)
		ret.AttestationsSource.StatusCount.Failed += uint64(entry.AttestationsScheduled) - uint64(entry.AttestationSourceExecuted)

		ret.AttestationsTarget.Income = ret.AttestationsTarget.Income.Add(entry.AttestationTargetReward.Mul(gWei).Mul(rpFactor))
		ret.AttestationsTarget.StatusCount.Success += uint64(entry.AttestationTargetExecuted)
		ret.AttestationsTarget.StatusCount.Failed += uint64(entry.AttestationsScheduled) - uint64(entry.AttestationTargetExecuted)

		ret.Inactivity.Income = ret.Inactivity.Income.Add(entry.AttestationInactivitytReward.Mul(gWei).Mul(rpFactor))
		if entry.AttestationInactivitytReward.LessThan(decimal.Zero) {
			ret.Inactivity.StatusCount.Failed++
		} else {
			ret.Inactivity.StatusCount.Success++
		}

		elReward := elRewards[entry.ValidatorIndex].Mul(rpFactor)
		if rpInfos != nil && protocolModes.RocketPool {
			if _, ok := rpInfos.Minipool[entry.ValidatorIndex]; ok {
				elReward = elReward.Add(rpInfos.Minipool[entry.ValidatorIndex].SmoothingPoolRewards[epoch])
			}
		}

		ret.Proposal.Income = ret.Proposal.Income.Add(entry.BlocksClReward.Mul(gWei).Mul(rpFactor).Add(elReward))
		ret.ProposalElReward = ret.ProposalElReward.Add(elReward)
		ret.Proposal.StatusCount.Success += uint64(entry.BlocksProposed)
		ret.Proposal.StatusCount.Failed += uint64(entry.BlocksScheduled) - uint64(entry.BlocksProposed)

		ret.Sync.Income = ret.Sync.Income.Add(entry.SyncRewards.Mul(gWei).Mul(rpFactor))
		ret.Sync.StatusCount.Success += uint64(entry.SyncExecuted)
		ret.Sync.StatusCount.Failed += uint64(entry.SyncScheduled) - uint64(entry.SyncExecuted)

		ret.Slashing.StatusCount.Success += uint64(entry.BlocksSlashingCount)
		if entry.Slashed {
			ret.Slashing.StatusCount.Failed++
		}

		ret.ProposalClAttIncReward = ret.ProposalClAttIncReward.Add(entry.BlocksClAttestationsReward.Mul(gWei).Mul(rpFactor))
		ret.ProposalClSyncIncReward = ret.ProposalClSyncIncReward.Add(entry.BlockClSyncAggregateReward.Mul(gWei).Mul(rpFactor))
		ret.ProposalClSlashingIncReward = ret.ProposalClSlashingIncReward.Add(entry.BlocksClSlasherReward.Mul(gWei).Mul(rpFactor))
	}

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardRewardsChart(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes) (*t.ChartData[int, decimal.Decimal], error) {
	// bar chart for the CL and EL rewards for each group for each epoch.
	// NO series for all groups combined except if AggregateGroups is true.
	// series id is group id, series property is 'cl' or 'el'

	wg := errgroup.Group{}
	var err error

	latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()
	const epochLookBack = 224
	startEpoch := uint64(0)
	if latestFinalizedEpoch > epochLookBack {
		startEpoch = latestFinalizedEpoch - epochLookBack
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Get rocketpool minipool infos if needed
	var rpInfos *t.RPInfo
	if protocolModes.RocketPool {
		rpInfos, err = d.getRocketPoolInfos(ctx, dashboardId, t.AllGroups)
		if err != nil {
			return nil, err
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main and EL rewards queries
	rewardsDs := goqu.Dialect("postgres").
		Select(
			goqu.L("e.validator_index"),
			goqu.L("e.epoch"),
			goqu.L(`(e.attestations_reward + e.blocks_cl_reward + e.sync_rewards) AS cl_rewards`)).
		From(goqu.L("validator_dashboard_data_epoch e")).
		With("validators", goqu.L("(SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
		Where(goqu.L("e.epoch_timestamp >= fromUnixTimestamp(?)", utils.EpochToTime(startEpoch).Unix()))

	elDs := goqu.Dialect("postgres").
		Select(
			goqu.L("b.proposer"),
			goqu.L("b.epoch"),
			goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS el_rewards")).
		From(goqu.L("users_val_dashboards_validators v")).
		LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("v.validator_index = b.proposer AND b.status = '1'"))).
		LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
		LeftJoin(
			goqu.Lateral(goqu.Dialect("postgres").
				From("relays_blocks").
				Select(
					goqu.L("exec_block_hash"),
					goqu.L("proposer_fee_recipient"),
					goqu.MAX("value").As("value")).
				Where(goqu.L("relays_blocks.exec_block_hash = b.exec_block_hash")).
				GroupBy("exec_block_hash", "proposer_fee_recipient")).As("rb"),
			goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")),
		).
		Where(goqu.L("b.epoch >= ?", startEpoch)).
		GroupBy(goqu.L("b.epoch"), goqu.L("b.proposer"))

	if rpInfos != nil && protocolModes.RocketPool {
		// Exclude rewards that went to the smoothing pool
		elDs = elDs.
			Where(goqu.L("(b.exec_fee_recipient != ? OR (rb.proposer_fee_recipient IS NOT NULL AND rb.proposer_fee_recipient != ?))", rpInfos.SmoothingPoolAddress, rpInfos.SmoothingPoolAddress))
	}

	if dashboardId.Validators == nil {
		rewardsDs = rewardsDs.
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("e.validator_index = v.validator_index"))).
			Where(goqu.L("e.validator_index IN (SELECT validator_index FROM validators)"))
		elDs = elDs.
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if dashboardId.AggregateGroups {
			rewardsDs = rewardsDs.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
				Order(goqu.L("e.epoch").Asc())
			elDs = elDs.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
				Order(goqu.L("b.epoch").Asc())
		} else {
			rewardsDs = rewardsDs.
				SelectAppend(goqu.L("v.group_id AS result_group_id")).
				Order(goqu.L("e.epoch").Asc(), goqu.L("result_group_id").Asc())
			elDs = elDs.
				SelectAppend(goqu.L("v.group_id AS result_group_id")).
				GroupByAppend(goqu.L("result_group_id")).
				Order(goqu.L("b.epoch").Asc(), goqu.L("result_group_id").Asc())
		}
	} else {
		// In case a list of validators is provided set the group to the default id
		rewardsDs = rewardsDs.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("e.validator_index IN ?", dashboardId.Validators)).
			Order(goqu.L("e.epoch").Asc())
		elDs = elDs.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("b.proposer = ANY(?)", pq.Array(dashboardId.Validators))).
			Order(goqu.L("b.epoch").Asc())
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main query and get the data
	type QueryResultSum struct {
		Epoch     uint64
		GroupId   uint64
		ClRewards decimal.Decimal
	}
	var queryResultSum []QueryResultSum
	smoothingPoolRewards := make(map[uint64]map[uint64]decimal.Decimal, 0) // epoch -> group -> reward

	wg.Go(func() error {
		queryResult := []struct {
			ValidatorIndex uint64 `db:"validator_index"`
			Epoch          uint64 `db:"epoch"`
			GroupId        uint64 `db:"result_group_id"`
			ClRewards      int64  `db:"cl_rewards"`
		}{}

		query, args, err := rewardsDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.clickhouseReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving rewards chart data: %w", err)
		}

		validatorGroupMap := make(map[uint64]uint64)
		for _, entry := range queryResult {
			if len(queryResultSum) == 0 ||
				queryResultSum[len(queryResultSum)-1].Epoch != entry.Epoch ||
				queryResultSum[len(queryResultSum)-1].GroupId != entry.GroupId {
				queryResultSum = append(queryResultSum, QueryResultSum{
					Epoch:   entry.Epoch,
					GroupId: entry.GroupId,
				})
			}

			validatorGroupMap[entry.ValidatorIndex] = entry.GroupId

			current := &queryResultSum[len(queryResultSum)-1]
			reward := utils.GWeiToWei(big.NewInt(entry.ClRewards))
			if rpInfos != nil && protocolModes.RocketPool {
				if rpValidator, ok := rpInfos.Minipool[entry.ValidatorIndex]; ok {
					reward = reward.Mul(d.getRocketPoolOperatorFactor(rpValidator))
				}
			}
			current.ClRewards = current.ClRewards.Add(reward)
		}

		// Calculate smoothing pool rewards
		// Has to be done here in the cl and not el part because here we have the list of all relevant validators
		if rpInfos != nil && protocolModes.RocketPool {
			for validatorIndex, groupId := range validatorGroupMap {
				for epoch, reward := range rpInfos.Minipool[validatorIndex].SmoothingPoolRewards {
					if _, ok := smoothingPoolRewards[epoch]; !ok {
						smoothingPoolRewards[epoch] = make(map[uint64]decimal.Decimal)
					}
					smoothingPoolRewards[epoch][groupId] = smoothingPoolRewards[epoch][groupId].Add(reward)
				}
			}
		}

		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[uint64]map[uint64]decimal.Decimal)
	wg.Go(func() error {
		elQueryResult := []struct {
			Proposer  uint64          `db:"proposer"`
			Epoch     uint64          `db:"epoch"`
			GroupId   uint64          `db:"result_group_id"`
			ElRewards decimal.Decimal `db:"el_rewards"`
		}{}

		query, args, err := elDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.readerDb.SelectContext(ctx, &elQueryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving el rewards data for rewards chart: %w", err)
		}

		for _, entry := range elQueryResult {
			if _, ok := elRewards[entry.Epoch]; !ok {
				elRewards[entry.Epoch] = make(map[uint64]decimal.Decimal)
			}

			reward := entry.ElRewards
			if rpInfos != nil && protocolModes.RocketPool {
				if rpValidator, ok := rpInfos.Minipool[entry.Proposer]; ok {
					reward = reward.Mul(d.getRocketPoolOperatorFactor(rpValidator))
				}
			}
			elRewards[entry.Epoch][entry.GroupId] = elRewards[entry.Epoch][entry.GroupId].Add(reward)
		}

		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard rewards chart data: %w", err)
	}

	// Add smoothing pool rewards to el rewards
	if rpInfos != nil && protocolModes.RocketPool {
		for epoch, groupRewards := range smoothingPoolRewards {
			for groupId, reward := range groupRewards {
				if _, ok := elRewards[epoch]; !ok {
					elRewards[epoch] = make(map[uint64]decimal.Decimal)
				}
				elRewards[epoch][groupId] = elRewards[epoch][groupId].Add(reward)
			}
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Create a map structure to store the data
	epochData := make(map[uint64]map[uint64]t.ClElValue[decimal.Decimal])
	epochList := make([]uint64, 0)

	for _, res := range queryResultSum {
		if _, ok := epochData[res.Epoch]; !ok {
			epochData[res.Epoch] = make(map[uint64]t.ClElValue[decimal.Decimal])
			epochList = append(epochList, res.Epoch)
		}

		epochData[res.Epoch][res.GroupId] = t.ClElValue[decimal.Decimal]{
			El: elRewards[res.Epoch][res.GroupId],
			Cl: res.ClRewards,
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

func (d *DataAccessService) GetValidatorDashboardDuties(ctx context.Context, dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	result := make([]t.VDBEpochDutiesTableRow, 0)
	var paging t.Paging

	wg := errgroup.Group{}

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
			return nil, nil, fmt.Errorf("failed to parse passed cursor as ValidatorDutiesCursor: %w", err)
		}
	}

	// Prepare the sorting
	isReverseDirection := (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse())

	// Analyze the search term
	indexSearch := int64(-1)
	if search != "" {
		if strings.HasPrefix(search, "0x") && utils.IsHash(search) {
			search = strings.ToLower(search)

			// Get the current validator state to convert pubkey to index
			validatorMapping, err := d.services.GetCurrentValidatorMapping()
			if err != nil {
				return nil, nil, err
			}
			if index, ok := validatorMapping.ValidatorIndices[search]; ok {
				indexSearch = int64(index)
			} else {
				// No validator index for pubkey found, return empty results
				return result, &paging, nil
			}
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			indexSearch = int64(number)
		} else {
			// No valid search term found, return empty results
			return result, &paging, nil
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Get rocketpool minipool infos if needed
	var rpInfos *t.RPInfo
	if protocolModes.RocketPool {
		rpInfos, err = d.getRocketPoolInfos(ctx, dashboardId, groupId)
		if err != nil {
			return nil, nil, err
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main and EL rewards queries
	rewardsDs := goqu.Dialect("postgres").
		Select(
			goqu.L("e.validator_index"),
			goqu.L("e.attestations_scheduled"),
			goqu.L("e.attestation_source_executed"),
			goqu.L("e.attestations_source_reward"),
			goqu.L("e.attestation_target_executed"),
			goqu.L("e.attestations_target_reward"),
			goqu.L("e.attestation_head_executed"),
			goqu.L("e.attestations_head_reward"),
			goqu.L("e.sync_scheduled"),
			goqu.L("e.sync_executed"),
			goqu.L("e.sync_rewards"),
			goqu.L("e.slashed"),
			goqu.L("e.blocks_slashing_count"),
			goqu.L("e.blocks_cl_slasher_reward"),
			goqu.L("e.blocks_scheduled"),
			goqu.L("e.blocks_proposed"),
			goqu.L("e.blocks_cl_attestations_reward"),
			goqu.L("e.blocks_cl_sync_aggregate_reward")).
		From(goqu.L("validator_dashboard_data_epoch e")).
		Where(goqu.L("e.epoch_timestamp = fromUnixTimestamp(?)", utils.EpochToTime(epoch).Unix())).
		Where(goqu.L(`
			(e.attestations_scheduled +
			e.sync_scheduled +
			e.blocks_scheduled +
			CASE WHEN e.slashed THEN 1 ELSE 0 END +
			e.blocks_slashing_count) > 0`))

	elDs := goqu.Dialect("postgres").
		Select(
			goqu.L("b.proposer"),
			goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS el_rewards")).
		From(goqu.L("blocks b")).
		LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
		LeftJoin(
			goqu.Lateral(goqu.Dialect("postgres").
				From("relays_blocks").
				Select(
					goqu.L("exec_block_hash"),
					goqu.L("proposer_fee_recipient"),
					goqu.MAX("value").As("value")).
				Where(goqu.L("relays_blocks.exec_block_hash = b.exec_block_hash")).
				GroupBy("exec_block_hash", "proposer_fee_recipient")).As("rb"),
			goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash")),
		).
		Where(goqu.L("b.epoch = ?", epoch)).
		Where(goqu.L("b.status = '1'")).
		GroupBy(goqu.L("b.proposer"))

	if rpInfos != nil && protocolModes.RocketPool {
		// Exclude rewards that went to the smoothing pool
		elDs = elDs.
			Where(goqu.L("(b.exec_fee_recipient != ? OR (rb.proposer_fee_recipient IS NOT NULL AND rb.proposer_fee_recipient != ?))", rpInfos.SmoothingPoolAddress, rpInfos.SmoothingPoolAddress))
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Add further conditions
	if dashboardId.Validators == nil {
		rewardsDs = rewardsDs.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("e.validator_index = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))
		elDs = elDs.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("b.proposer = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if groupId != t.AllGroups {
			rewardsDs = rewardsDs.Where(goqu.L("v.group_id = ?", groupId))
			elDs = elDs.Where(goqu.L("v.group_id = ?", groupId))
		}

		if indexSearch != -1 {
			rewardsDs = rewardsDs.Where(goqu.L("e.validator_index = ?", indexSearch))
			elDs = elDs.Where(goqu.L("b.proposer = ?", indexSearch))
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

		rewardsDs = rewardsDs.Where(goqu.L("e.validator_index IN ?", validators))
		elDs = elDs.Where(goqu.L("b.proposer = ANY(?)", pq.Array(validators)))
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Get the main data
	type QueryResultBase struct {
		ValidatorIndex             uint64 `db:"validator_index"`
		AttestationsScheduled      uint64 `db:"attestations_scheduled"`
		AttestationsSourceExecuted uint64 `db:"attestation_source_executed"`
		AttestationsTargetExecuted uint64 `db:"attestation_target_executed"`
		AttestationsHeadExecuted   uint64 `db:"attestation_head_executed"`
		SyncScheduled              uint64 `db:"sync_scheduled"`
		SyncExecuted               uint64 `db:"sync_executed"`
		Slashed                    bool   `db:"slashed"`
		BlocksSlashingCount        uint64 `db:"blocks_slashing_count"`
		BlocksScheduled            uint64 `db:"blocks_scheduled"`
		BlocksProposed             uint64 `db:"blocks_proposed"`
	}

	type QueryResult struct {
		QueryResultBase
		AttestationsSourceReward    int64 `db:"attestations_source_reward"`
		AttestationsTargetReward    int64 `db:"attestations_target_reward"`
		AttestationsHeadReward      int64 `db:"attestations_head_reward"`
		SyncRewards                 int64 `db:"sync_rewards"`
		BlocksClSlasherReward       int64 `db:"blocks_cl_slasher_reward"`
		BlocksClAttestationsReward  int64 `db:"blocks_cl_attestations_reward"`
		BlocksClSyncAggregateReward int64 `db:"blocks_cl_sync_aggregate_reward"`
	}

	type QueryResultAdjusted struct {
		QueryResultBase
		AttestationsSourceReward    decimal.Decimal
		AttestationsTargetReward    decimal.Decimal
		AttestationsHeadReward      decimal.Decimal
		SyncRewards                 decimal.Decimal
		BlocksClSlasherReward       decimal.Decimal
		BlocksClAttestationsReward  decimal.Decimal
		BlocksClSyncAggregateReward decimal.Decimal
	}

	var queryResultAdjusted []QueryResultAdjusted
	wg.Go(func() error {
		var queryResult []QueryResult

		query, args, err := rewardsDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.clickhouseReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving validator rewards data: %w", err)
		}

		for _, entry := range queryResult {
			rpFactor := decimal.NewFromInt(1)
			if rpInfos != nil && protocolModes.RocketPool {
				if rpValidator, ok := rpInfos.Minipool[entry.ValidatorIndex]; ok {
					rpFactor = d.getRocketPoolOperatorFactor(rpValidator)
				}
			}

			current := QueryResultAdjusted{QueryResultBase: entry.QueryResultBase}

			current.AttestationsSourceReward = utils.GWeiToWei(big.NewInt(entry.AttestationsSourceReward)).Mul(rpFactor)
			current.AttestationsTargetReward = utils.GWeiToWei(big.NewInt(entry.AttestationsTargetReward)).Mul(rpFactor)
			current.AttestationsHeadReward = utils.GWeiToWei(big.NewInt(entry.AttestationsHeadReward)).Mul(rpFactor)
			current.SyncRewards = utils.GWeiToWei(big.NewInt(entry.SyncRewards)).Mul(rpFactor)
			current.BlocksClSlasherReward = utils.GWeiToWei(big.NewInt(entry.BlocksClSlasherReward)).Mul(rpFactor)
			current.BlocksClAttestationsReward = utils.GWeiToWei(big.NewInt(entry.BlocksClAttestationsReward)).Mul(rpFactor)
			current.BlocksClSyncAggregateReward = utils.GWeiToWei(big.NewInt(entry.BlocksClSyncAggregateReward)).Mul(rpFactor)

			queryResultAdjusted = append(queryResultAdjusted, current)
		}

		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[uint64]decimal.Decimal)
	wg.Go(func() error {
		elQueryResult := []struct {
			Proposer  uint64          `db:"proposer"`
			ElRewards decimal.Decimal `db:"el_rewards"`
		}{}

		query, args, err := elDs.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.readerDb.SelectContext(ctx, &elQueryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving validator el rewards data for rewards: %w", err)
		}

		for _, entry := range elQueryResult {
			reward := entry.ElRewards
			if rpInfos != nil && protocolModes.RocketPool {
				if rpValidator, ok := rpInfos.Minipool[entry.Proposer]; ok {
					reward = reward.Mul(d.getRocketPoolOperatorFactor(rpValidator))
				}
			}
			elRewards[entry.Proposer] = reward
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator dashboard rewards data: %w", err)
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Create the result
	cursorData := make([]t.ValidatorDutiesCursor, 0)
	for _, res := range queryResultAdjusted {
		clReward := res.AttestationsHeadReward.Add(res.AttestationsSourceReward).Add(res.AttestationsTargetReward).
			Add(res.SyncRewards).
			Add(res.BlocksClAttestationsReward).Add(res.BlocksClSyncAggregateReward).Add(res.BlocksClSlasherReward)
		totalReward := clReward.Add(elRewards[res.ValidatorIndex])

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
		if res.Slashed || res.BlocksSlashingCount > 0 {
			slashedEvent := t.ValidatorHistoryEvent{
				Income: res.BlocksClSlasherReward,
			}
			if res.Slashed {
				if res.BlocksSlashingCount > 0 {
					slashedEvent.Status = "partial"
				} else {
					slashedEvent.Status = "failed"
				}
			} else if res.BlocksSlashingCount > 0 {
				slashedEvent.Status = "success"
			}
			row.Duties.Slashing = &slashedEvent
		}

		// Get proposal data
		if res.BlocksScheduled > 0 {
			proposalEvent := t.ValidatorHistoryProposal{
				ElIncome:                     elRewards[res.ValidatorIndex],
				ClAttestationInclusionIncome: res.BlocksClAttestationsReward,
				ClSyncInclusionIncome:        res.BlocksClSyncAggregateReward,
				ClSlashingInclusionIncome:    res.BlocksClSlasherReward,
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
			Reward: totalReward,
		})
	}

	// Sort the result
	totalReward := func(resultEntry t.VDBEpochDutiesTableRow) decimal.Decimal {
		totalReward := decimal.Zero
		if resultEntry.Duties.AttestationSource != nil {
			totalReward = totalReward.Add(resultEntry.Duties.AttestationHead.Income).Add(resultEntry.Duties.AttestationSource.Income).Add(resultEntry.Duties.AttestationTarget.Income)
		}
		if resultEntry.Duties.Sync != nil {
			totalReward = totalReward.Add(resultEntry.Duties.Sync.Income)
		}
		if resultEntry.Duties.Proposal != nil {
			totalReward = totalReward.Add(resultEntry.Duties.Proposal.ElIncome).Add(resultEntry.Duties.Proposal.ClAttestationInclusionIncome).Add(resultEntry.Duties.Proposal.ClSyncInclusionIncome).Add(resultEntry.Duties.Proposal.ClSlashingInclusionIncome)
		}
		return totalReward
	}

	sort.Slice(result, func(i, j int) bool {
		switch colSort.Column {
		case enums.VDBDutiesColumns.Validator:
			if isReverseDirection {
				return result[i].Validator > result[j].Validator
			}
			return result[i].Validator < result[j].Validator
		case enums.VDBDutiesColumns.Reward:
			// It is possible that rewards are equal, in that case we sort by validator index secondarily
			if isReverseDirection {
				if totalReward(result[i]).Equal(totalReward(result[j])) {
					return result[i].Validator > result[j].Validator
				}
				return totalReward(result[i]).GreaterThan(totalReward(result[j]))
			}
			if totalReward(result[i]).Equal(totalReward(result[j])) {
				return result[i].Validator < result[j].Validator
			}
			return totalReward(result[i]).LessThan(totalReward(result[j]))
		default:
			return false
		}
	})

	sort.Slice(cursorData, func(i, j int) bool {
		switch colSort.Column {
		case enums.VDBDutiesColumns.Validator:
			if isReverseDirection {
				return cursorData[i].Index > cursorData[j].Index
			}
			return cursorData[i].Index < cursorData[j].Index
		case enums.VDBDutiesColumns.Reward:
			// It is possible that rewards are equal, in that case we sort by validator index secondarily
			if isReverseDirection {
				if cursorData[i].Reward.Equal(cursorData[j].Reward) {
					return cursorData[i].Index > cursorData[j].Index
				}
				return cursorData[i].Reward.GreaterThan(cursorData[j].Reward)
			}
			if cursorData[i].Reward.Equal(cursorData[j].Reward) {
				return cursorData[i].Index < cursorData[j].Index
			}
			return cursorData[i].Reward.LessThan(cursorData[j].Reward)
		default:
			return false
		}
	})

	// Remove data before the cursor
	if currentCursor.IsValid() {
		cursorIndex := -1

		for idx, cursorEntry := range cursorData {
			if cursorEntry.Index == currentCursor.Index && cursorEntry.Reward.Equal(currentCursor.Reward) {
				cursorIndex = idx
				break
			}
		}
		if cursorIndex == -1 {
			return nil, nil, fmt.Errorf("cursor not found in data")
		}

		result = result[cursorIndex+1:]
		cursorData = cursorData[cursorIndex+1:]
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

func (d *DataAccessService) getValidatorHistoryEvent(income decimal.Decimal, scheduledEvents, executedEvents uint64) *t.ValidatorHistoryEvent {
	if scheduledEvents > 0 {
		validatorHistoryEvent := t.ValidatorHistoryEvent{
			Income: income,
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
