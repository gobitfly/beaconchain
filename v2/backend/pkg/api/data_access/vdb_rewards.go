package dataaccess

import (
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
	result := make([]t.VDBRewardsTableRow, 0)
	var paging t.Paging

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
	sortSearchDirection := ">"
	sortSearchOrder := " ASC"
	if (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse()) {
		sortSearchDirection = "<"
		sortSearchOrder = " DESC"
	}

	// Analyze the search term
	indexSearch := int64(-1)
	epochSearch := int64(-1)
	if search != "" {
		if utils.IsHash(search) {
			// Ensure that we have a "0x" prefix for the search term
			if !strings.HasPrefix(search, "0x") {
				search = "0x" + search
			}
			search = strings.ToLower(search)

			// Get the current validator state to convert pubkey to index
			validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
			defer releaseLock()
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
	const epochLookBack = 10

	queryResult := []struct {
		Epoch                 uint64          `db:"epoch"`
		GroupId               int64           `db:"group_id"`
		ClRewards             int64           `db:"cl_rewards"`
		ElRewards             decimal.Decimal `db:"el_rewards"`
		AttestationsScheduled uint64          `db:"attestations_scheduled"`
		AttestationsExecuted  uint64          `db:"attestations_executed"`
		BlocksScheduled       uint64          `db:"blocks_scheduled"`
		BlocksProposed        uint64          `db:"blocks_proposed"`
		SyncScheduled         uint64          `db:"sync_scheduled"`
		SyncExecuted          uint64          `db:"sync_executed"`
		Slashed               uint64          `db:"slashed"`
	}{}

	queryParams := []interface{}{}
	rewardsQuery := ""

	groupIdSearchMap := make(map[uint64]bool, 0)

	// TODO: El rewards data (blocks_el_reward) will be provided at a later point
	rewardsDataQuery := `
		SUM(COALESCE(e.attestations_reward, 0) + COALESCE(e.blocks_cl_reward, 0) +
		COALESCE(e.sync_rewards, 0) + COALESCE(e.slasher_reward, 0)) AS cl_rewards,
		SUM(COALESCE(e.blocks_el_reward, 0)) AS el_rewards,		
		SUM(COALESCE(e.attestations_scheduled, 0)) AS attestations_scheduled,
		SUM(COALESCE(e.attestations_executed, 0)) AS attestations_executed,
		SUM(COALESCE(e.blocks_scheduled, 0)) AS blocks_scheduled,
		SUM(COALESCE(e.blocks_proposed, 0)) AS blocks_proposed,
		SUM(COALESCE(e.sync_scheduled, 0)) AS sync_scheduled,
		SUM(COALESCE(e.sync_executed, 0)) AS sync_executed,
		SUM(CASE WHEN e.slashed THEN 1 ELSE 0 END) AS slashed
		`

	if dashboardId.Validators == nil {
		queryParams = append(queryParams, dashboardId.Id, latestFinalizedEpoch-epochLookBack)
		whereQuery := fmt.Sprintf("WHERE v.dashboard_id = $%d AND e.epoch > $%d", len(queryParams)-1, len(queryParams))
		if currentCursor.IsValid() {
			if currentCursor.IsReverse() {
				if currentCursor.GroupId == t.AllGroups {
					// The cursor is on the total rewards so get the data for all groups excluding the cursor epoch
					queryParams = append(queryParams, currentCursor.Epoch)
					whereQuery += fmt.Sprintf(" AND e.epoch%[1]s$%[2]d", sortSearchDirection, len(queryParams))
				} else {
					// The cursor is on a specific group, get the data for the whole epoch since we could need it for the total rewards
					queryParams = append(queryParams, currentCursor.Epoch)
					whereQuery += fmt.Sprintf(" AND e.epoch%[1]s=$%[2]d", sortSearchDirection, len(queryParams))
				}
			} else {
				if currentCursor.GroupId == t.AllGroups {
					// The cursor is on the total rewards so get the data for all groups including the cursor epoch
					queryParams = append(queryParams, currentCursor.Epoch)
					whereQuery += fmt.Sprintf(" AND e.epoch%[1]s=$%[2]d", sortSearchDirection, len(queryParams))
				} else {
					// The cursor is on a specific group so get the data for groups before/after it
					queryParams = append(queryParams, currentCursor.Epoch, currentCursor.GroupId)
					whereQuery += fmt.Sprintf(" AND (e.epoch%[1]s$%[2]d OR (e.epoch=$%[2]d AND v.group_id%[1]s$%[3]d))", sortSearchDirection, len(queryParams)-1, len(queryParams))
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
					d.alloyReader.Get(&found, `
						SELECT EXISTS(
								SELECT 1 
								FROM users_val_dashboards_validators
								WHERE dashboard_id = $1 AND validator_index = $2)
						`, dashboardId.Id, indexSearch)
				}
				if !found && epochSearch != -1 {
					queryParams = append(queryParams, epochSearch)
					whereQuery += fmt.Sprintf(" AND e.epoch = $%d", len(queryParams))
				}
			} else {
				// Create a secondary query to get the group ids that match the search term
				// We cannot do everything in one query because we need to know the "epoch total" even for groups we do not search for
				groupIdQueryParams := []interface{}{}

				indexSearchQuery := ""
				if indexSearch != -1 {
					groupIdQueryParams = append(groupIdQueryParams, indexSearch)
					indexSearchQuery = fmt.Sprintf(" OR v.validator_index = $%d", len(groupIdQueryParams))
				}

				groupIdQueryParams = append(groupIdQueryParams, dashboardId.Id, search)
				groupIdQuery := fmt.Sprintf(`
					SELECT
						DISTINCT(group_id)
					FROM users_val_dashboards_validators v
					INNER JOIN users_val_dashboards_groups g ON v.group_id = g.id AND v.dashboard_id = g.dashboard_id
					WHERE v.dashboard_id = $%d AND (g.name ILIKE ($%d||'%%') %s)
					`, len(groupIdQueryParams)-1, len(groupIdQueryParams), indexSearchQuery)

				var groupIdSearch []uint64
				err = d.alloyReader.Select(&groupIdSearch, groupIdQuery, groupIdQueryParams...)
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
						queryParams = append(queryParams, epochSearch)
						whereQuery += fmt.Sprintf(" AND e.epoch = $%d", len(queryParams))
					} else {
						// No search for goup or epoch possible, return empty results
						return result, &paging, nil
					}
				}
			}
		}

		groupIdQuery := "v.group_id,"
		groupByQuery := "GROUP BY e.epoch, v.group_id"
		orderQuery := fmt.Sprintf("ORDER BY e.epoch %[1]s, v.group_id %[1]s", sortSearchOrder)
		if dashboardId.AggregateGroups {
			queryParams = append(queryParams, t.AggregatedGroupId)
			groupIdQuery = fmt.Sprintf("$%d::smallint AS group_id,", len(queryParams))
			groupByQuery = "GROUP BY e.epoch"
			orderQuery = fmt.Sprintf("ORDER BY e.epoch %s", sortSearchOrder)
		}

		rewardsQuery = fmt.Sprintf(`
			SELECT
				e.epoch,
				%s
				%s
			FROM validator_dashboard_data_epoch e
			INNER JOIN users_val_dashboards_validators v ON e.validator_index = v.validator_index
			%s
			%s
			%s`, groupIdQuery, rewardsDataQuery, whereQuery, groupByQuery, orderQuery)
	} else {
		// In case a list of validators is provided set the group to the default id
		queryParams = append(queryParams, pq.Array(dashboardId.Validators), latestFinalizedEpoch-epochLookBack)
		whereQuery := fmt.Sprintf("WHERE e.validator_index = ANY($%d) AND e.epoch > $%d", len(queryParams)-1, len(queryParams))
		if currentCursor.IsValid() {
			queryParams = append(queryParams, currentCursor.Epoch)
			whereQuery += fmt.Sprintf(" AND e.epoch%s$%d", sortSearchDirection, len(queryParams))
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
				queryParams = append(queryParams, epochSearch)
				whereQuery += fmt.Sprintf(" AND e.epoch = $%d", len(queryParams))
			}
		}

		orderQuery := fmt.Sprintf("ORDER BY e.epoch %s", sortSearchOrder)

		queryParams = append(queryParams, t.DefaultGroupId)
		rewardsQuery = fmt.Sprintf(`
			SELECT
				e.epoch,
				$%d::smallint AS group_id,
				%s
			FROM validator_dashboard_data_epoch e
			%s
			GROUP BY e.epoch
			%s`, len(queryParams), rewardsDataQuery, whereQuery, orderQuery)
	}

	err = d.alloyReader.Select(&queryResult, rewardsQuery, queryParams...)
	if err != nil {
		return nil, nil, err
	}

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
		Slashed               uint64
	}
	totalEpochInfo := make(map[uint64]*TotalEpochInfo, 0)

	for _, res := range queryResult {
		duty := t.VDBRewardesTableDuty{}
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
		// TODO: Slashing data is not yet available in the db
		slashingInfo := res.Slashed /*+ "Validators slashed"*/
		if slashingInfo > 0 {
			duty.Slashing = &slashingInfo
		}

		if duty.Attestation != nil || duty.Proposal != nil || duty.Sync != nil || duty.Slashing != nil {
			// Only add groups that had some duty or got slashed
			resultWoTotal = append(resultWoTotal, t.VDBRewardsTableRow{
				Epoch:   res.Epoch,
				Duty:    duty,
				GroupId: res.GroupId,
				Reward: t.ClElValue[decimal.Decimal]{
					El: res.ElRewards,
					Cl: utils.GWeiToWei(big.NewInt(res.ClRewards)),
				},
			})

			// Add it to the total epoch info
			if _, ok := totalEpochInfo[res.Epoch]; !ok {
				totalEpochInfo[res.Epoch] = &TotalEpochInfo{}
			}
			totalEpochInfo[res.Epoch].Groups = append(totalEpochInfo[res.Epoch].Groups, uint64(res.GroupId))
			totalEpochInfo[res.Epoch].ClRewards += res.ClRewards
			totalEpochInfo[res.Epoch].ElRewards = totalEpochInfo[res.Epoch].ElRewards.Add(res.ElRewards)
			totalEpochInfo[res.Epoch].AttestationsScheduled += res.AttestationsScheduled
			totalEpochInfo[res.Epoch].AttestationsExecuted += res.AttestationsExecuted
			totalEpochInfo[res.Epoch].BlocksScheduled += res.BlocksScheduled
			totalEpochInfo[res.Epoch].BlocksProposed += res.BlocksProposed
			totalEpochInfo[res.Epoch].SyncScheduled += res.SyncScheduled
			totalEpochInfo[res.Epoch].SyncExecuted += res.SyncExecuted
			totalEpochInfo[res.Epoch].Slashed += res.Slashed
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

		duty := t.VDBRewardesTableDuty{}
		if totalInfo.AttestationsScheduled > 0 {
			attestationPercentage := (float64(totalInfo.AttestationsExecuted) / float64(totalInfo.AttestationsScheduled)) * 100.0
			duty.Attestation = &attestationPercentage
		}
		if totalInfo.BlocksScheduled > 0 {
			ProposalPercentage := (float64(totalInfo.BlocksProposed) / float64(totalInfo.BlocksScheduled)) * 100.0
			duty.Proposal = &ProposalPercentage
		}
		if totalInfo.SyncScheduled > 0 {
			SyncPercentage := (float64(totalInfo.SyncExecuted) / float64(totalInfo.SyncScheduled)) * 100.0
			duty.Sync = &SyncPercentage
		}
		// TODO: Slashing data is not yet available in the db
		slashingInfo := totalInfo.Slashed /*+ "Validators slashed"*/
		if slashingInfo > 0 {
			duty.Slashing = &slashingInfo
		}

		totalRewards[epoch] = t.VDBRewardsTableRow{
			Epoch:   epoch,
			Duty:    duty,
			GroupId: t.AggregatedGroupId,
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

func (d *DataAccessService) GetValidatorDashboardGroupRewards(dashboardId t.VDBId, groupId int64, epoch uint64) (*t.VDBGroupRewardsData, error) {
	ret := &t.VDBGroupRewardsData{}

	type queryResult struct {
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
		BlocksElReward  decimal.Decimal `db:"blocks_el_reward"`

		SyncScheduled uint32          `db:"sync_scheduled"`
		SyncExecuted  uint32          `db:"sync_executed"`
		SyncRewards   decimal.Decimal `db:"sync_rewards"`

		SlasherRewards decimal.Decimal `db:"slasher_reward"`

		BlocksClAttestationsReward decimal.Decimal `db:"blocks_cl_attestations_reward"`
		BlockClSyncAggregateReward decimal.Decimal `db:"blocks_cl_sync_aggregate_reward"`
	}

	query := `select
			COALESCE(validator_dashboard_data_epoch.attestations_source_reward, 0) as attestations_source_reward,
			COALESCE(validator_dashboard_data_epoch.attestations_target_reward, 0) as attestations_target_reward,
			COALESCE(validator_dashboard_data_epoch.attestations_head_reward, 0) as attestations_head_reward,
			COALESCE(validator_dashboard_data_epoch.attestations_inactivity_reward, 0) as attestations_inactivity_reward,
			COALESCE(validator_dashboard_data_epoch.attestations_inclusion_reward, 0) as attestations_inclusion_reward,
			COALESCE(validator_dashboard_data_epoch.attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(validator_dashboard_data_epoch.attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(validator_dashboard_data_epoch.attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(validator_dashboard_data_epoch.attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(validator_dashboard_data_epoch.blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(validator_dashboard_data_epoch.blocks_proposed, 0) as blocks_proposed,
			COALESCE(validator_dashboard_data_epoch.blocks_cl_reward, 0) as blocks_cl_reward,
			COALESCE(validator_dashboard_data_epoch.blocks_el_reward, 0) as blocks_el_reward,
			COALESCE(validator_dashboard_data_epoch.sync_scheduled, 0) as sync_scheduled,
			COALESCE(validator_dashboard_data_epoch.sync_executed, 0) as sync_executed,
			COALESCE(validator_dashboard_data_epoch.sync_rewards, 0) as sync_rewards,
			COALESCE(validator_dashboard_data_epoch.slasher_reward, 0) as slasher_reward,
			COALESCE(validator_dashboard_data_epoch.blocks_cl_attestations_reward, 0) as blocks_cl_attestations_reward,
			COALESCE(validator_dashboard_data_epoch.blocks_cl_sync_aggregate_reward, 0) as blocks_cl_sync_aggregate_reward
		`

	var rows []*queryResult

	// handle the case when we have a list of validators
	if len(dashboardId.Validators) > 0 {
		whereClause := "from validator_dashboard_data_epoch where validator_index = any($1) and epoch = $2"
		query = fmt.Sprintf("%s %s", query, whereClause)
		err := d.alloyReader.Select(&rows, query, pq.Array(dashboardId.Validators), epoch)
		if err != nil {
			log.Error(err, "error while getting validator dashboard group rewards", 0)
			return nil, err
		}
	} else { // handle the case when we have a dashboard id and an optional group id
		joinAndWhereClause := `from users_val_dashboards_validators inner join validator_dashboard_data_epoch on validator_dashboard_data_epoch.validator_index = users_val_dashboards_validators.validator_index
			where (dashboard_id = $1 and (group_id = $2 or $2 = -1) and epoch = $3)`
		query = fmt.Sprintf("%s %s", query, joinAndWhereClause)
		err := d.alloyReader.Select(&rows, query, dashboardId.Id, groupId, epoch)
		if err != nil {
			log.Error(err, "error while getting validator dashboard group rewards", 0)
			return nil, err
		}
	}

	gWei := decimal.NewFromInt(1e9)

	for _, row := range rows {
		ret.AttestationsHead.Income = ret.AttestationsHead.Income.Add(row.AttestationHeadReward.Mul(gWei))
		ret.AttestationsHead.StatusCount.Success += uint64(row.AttestationHeadExecuted)
		ret.AttestationsHead.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationHeadExecuted)

		ret.AttestationsSource.Income = ret.AttestationsSource.Income.Add(row.AttestationSourceReward.Mul(gWei))
		ret.AttestationsSource.StatusCount.Success += uint64(row.AttestationSourceExecuted)
		ret.AttestationsSource.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationSourceExecuted)

		ret.AttestationsTarget.Income = ret.AttestationsTarget.Income.Add(row.AttestationTargetReward.Mul(gWei))
		ret.AttestationsTarget.StatusCount.Success += uint64(row.AttestationTargetExecuted)
		ret.AttestationsTarget.StatusCount.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationTargetExecuted)

		ret.Inactivity.Income = ret.Inactivity.Income.Add(row.AttestationInactivitytReward.Mul(gWei))
		if row.AttestationInactivitytReward.LessThan(decimal.Zero) {
			ret.Inactivity.StatusCount.Failed++
		} else {
			ret.Inactivity.StatusCount.Success++
		}

		ret.Proposal.Income = ret.Proposal.Income.Add(row.BlocksClReward.Mul(gWei))
		ret.Proposal.StatusCount.Success += uint64(row.BlocksProposed)
		ret.Proposal.StatusCount.Failed += uint64(row.BlocksScheduled) - uint64(row.BlocksProposed)

		ret.Sync.Income = ret.Sync.Income.Add(row.SyncRewards.Mul(gWei))
		ret.Sync.StatusCount.Success += uint64(row.SyncExecuted)
		ret.Sync.StatusCount.Failed += uint64(row.SyncScheduled) - uint64(row.SyncExecuted)

		ret.ProposalClAttIncReward = ret.ProposalClAttIncReward.Add(row.BlocksClAttestationsReward.Mul(gWei))
		ret.ProposalClSyncIncReward = ret.ProposalClSyncIncReward.Add(row.BlockClSyncAggregateReward.Mul(gWei))
		ret.ProposalClSlashingIncReward = ret.ProposalClSlashingIncReward.Add(row.SlasherRewards.Mul(gWei))

		// TODO: Add slashing info once available
	}

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error) {
	// bar chart for the CL and EL rewards for each group for each epoch.
	// NO series for all groups combined except if AggregateGroups is true.
	// series id is group id, series property is 'cl' or 'el'

	latestFinalizedEpoch := cache.LatestFinalizedEpoch.Get()
	const epochLookBack = 10

	queryResult := []struct {
		Epoch     uint64          `db:"epoch"`
		GroupId   int64           `db:"group_id"`
		ElRewards decimal.Decimal `db:"el_rewards"`
		ClRewards int64           `db:"cl_rewards"`
	}{}

	queryParams := []interface{}{}
	rewardsQuery := ""

	// TODO: El rewards data (blocks_el_reward) will be provided at a later point
	rewardsDataQuery := `
		SUM(COALESCE(e.blocks_el_reward, 0)) AS el_rewards,
		SUM(COALESCE(e.attestations_reward, 0) + COALESCE(e.blocks_cl_reward, 0) +
		COALESCE(e.sync_rewards, 0) + COALESCE(e.slasher_reward, 0)) AS cl_rewards
		`

	if dashboardId.Validators == nil {
		groupIdQuery := "v.group_id,"
		groupByQuery := "GROUP BY e.epoch, v.group_id"
		orderQuery := "ORDER BY e.epoch, v.group_id"
		if dashboardId.AggregateGroups {
			queryParams = append(queryParams, t.AggregatedGroupId)
			groupIdQuery = fmt.Sprintf("$%d::smallint AS group_id,", len(queryParams))
			groupByQuery = "GROUP BY e.epoch"
			orderQuery = "ORDER BY e.epoch"
		}

		queryParams = append(queryParams, dashboardId.Id, latestFinalizedEpoch-epochLookBack)
		rewardsQuery = fmt.Sprintf(`
			SELECT
				e.epoch,
				%s
				%s
			FROM validator_dashboard_data_epoch e
			INNER JOIN users_val_dashboards_validators v ON e.validator_index = v.validator_index
			WHERE v.dashboard_id = $%d AND e.epoch > $%d
			%s
			%s`, groupIdQuery, rewardsDataQuery, len(queryParams)-1, len(queryParams), groupByQuery, orderQuery)
	} else {
		// In case a list of validators is provided set the group to the default id
		queryParams = append(queryParams, t.DefaultGroupId, pq.Array(dashboardId.Validators), latestFinalizedEpoch-epochLookBack)
		rewardsQuery = fmt.Sprintf(`
			SELECT
				e.epoch,
				$%d::smallint AS group_id,
				%s
			FROM validator_dashboard_data_epoch e
			WHERE e.validator_index = ANY($%d) AND e.epoch > $%d
			GROUP BY e.epoch
			ORDER BY e.epoch`, len(queryParams)-2, rewardsDataQuery, len(queryParams)-1, len(queryParams))
	}

	err := d.alloyReader.Select(&queryResult, rewardsQuery, queryParams...)
	if err != nil {
		return nil, err
	}

	// Create a map structure to store the data
	epochData := make(map[uint64]map[int64]t.ClElValue[decimal.Decimal])
	epochList := make([]uint64, 0)

	for _, res := range queryResult {
		if _, ok := epochData[res.Epoch]; !ok {
			epochData[res.Epoch] = make(map[int64]t.ClElValue[decimal.Decimal])
			epochList = append(epochList, res.Epoch)
		}

		epochData[res.Epoch][res.GroupId] = t.ClElValue[decimal.Decimal]{
			El: res.ElRewards,
			Cl: utils.GWeiToWei(big.NewInt(res.ClRewards)),
		}
	}

	// Get the list of groups
	// It should be identical for all epochs
	var groupList []int64
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
				result.Series[idx].Data = append(result.Series[idx].Data, epochData[epoch][int64(series.Id)].El)
			} else if series.Property == "cl" {
				result.Series[idx].Data = append(result.Series[idx].Data, epochData[epoch][int64(series.Id)].Cl)
			} else {
				return nil, fmt.Errorf("unknown series property: %s", series.Property)
			}
		}
	}

	return &result, nil
}

func (d *DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	result := make([]t.VDBEpochDutiesTableRow, 0)
	var paging t.Paging

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
	sortSearchDirection := ">"
	sortSearchOrder := " ASC"
	if (colSort.Desc && !currentCursor.IsReverse()) || (!colSort.Desc && currentCursor.IsReverse()) {
		sortSearchDirection = "<"
		sortSearchOrder = " DESC"
	}

	// Analyze the search term
	indexSearch := int64(-1)
	if search != "" {
		if utils.IsHash(search) {
			// Ensure that we have a "0x" prefix for the search term
			if !strings.HasPrefix(search, "0x") {
				search = "0x" + search
			}
			search = strings.ToLower(search)

			// Get the current validator state to convert pubkey to index
			validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
			defer releaseLock()
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
		}
	}

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
		Slashed                     bool            `db:"slashed"`
		SlasherReward               int64           `db:"slasher_reward"`
		BlocksScheduled             uint64          `db:"blocks_scheduled"`
		BlocksProposed              uint64          `db:"blocks_proposed"`
		BlocksElReward              decimal.Decimal `db:"blocks_el_reward"`
		BlocksClAttestationsReward  int64           `db:"blocks_cl_attestations_reward"`
		BlocksClSyncAggregateReward int64           `db:"blocks_cl_sync_aggregate_reward"`
	}{}

	queryParams := []interface{}{}

	joinSubquery := ""

	queryParams = append(queryParams, epoch)
	whereSubquery := fmt.Sprintf(`WHERE epoch = $%d
		AND (
			COALESCE(e.attestations_scheduled, 0) +
			COALESCE(e.sync_scheduled,0) +
			COALESCE(e.blocks_scheduled,0) +
			CASE WHEN e.slashed THEN 1 ELSE 0 END
		) > 0
	`, len(queryParams))
	whereQuery := ""

	orderQuery := ""

	queryParams = append(queryParams, limit+1)
	limitQuery := fmt.Sprintf(`LIMIT $%d
	`, len(queryParams))

	if dashboardId.Validators == nil {
		joinSubquery = `INNER JOIN users_val_dashboards_validators v ON e.validator_index = v.validator_index
		`

		queryParams = append(queryParams, dashboardId.Id)
		whereSubquery += fmt.Sprintf(`AND v.dashboard_id = $%d
		`, len(queryParams))

		if groupId != t.AllGroups {
			queryParams = append(queryParams, groupId)
			whereSubquery += fmt.Sprintf(`AND v.group_id = $%d
			`, len(queryParams))
		}

		if indexSearch != -1 {
			queryParams = append(queryParams, indexSearch)
			whereSubquery += fmt.Sprintf(`AND e.validator_index = $%d
			`, len(queryParams))
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

		queryParams = append(queryParams, pq.Array(validators))
		whereSubquery += fmt.Sprintf(`AND e.validator_index = ANY($%d)
		`, len(queryParams))
	}

	if colSort.Column == enums.VDBDutiesColumns.Validator {
		if currentCursor.IsValid() {
			// If we have a valid cursor only check the results before/after it
			queryParams = append(queryParams, currentCursor.Index)
			whereSubquery += fmt.Sprintf(`AND e.validator_index%s$%d
			`, sortSearchDirection, len(queryParams))
		}
		orderQuery = fmt.Sprintf(`ORDER BY validator_index %s
		`, sortSearchOrder)
	} else if colSort.Column == enums.VDBDutiesColumns.Reward {
		if currentCursor.IsValid() {
			// If we have a valid cursor only check the results before/after it
			queryParams = append(queryParams, currentCursor.Reward, currentCursor.Index)

			whereQuery += fmt.Sprintf(`WHERE (total_reward%[1]s$%[2]d OR (total_reward=$%[2]d AND validator_index%[1]s$%[3]d))
			`, sortSearchDirection, len(queryParams)-1, len(queryParams))
		}
		orderQuery = fmt.Sprintf(`ORDER BY total_reward %[1]s, validator_index %[1]s
		`, sortSearchOrder)
	}

	// Use a subquery to allow access to total_reward in the where clause
	// TODO: El rewards data (blocks_el_reward) will be provided at a later point
	rewardsQuery := fmt.Sprintf(`
		SELECT *
		FROM (
			SELECT
				e.validator_index,
				(
					COALESCE(e.blocks_el_reward, 0) +
					CAST((
						COALESCE(e.attestations_reward, 0) +
						COALESCE(e.blocks_cl_reward, 0) +
						COALESCE(e.sync_rewards, 0) +
						COALESCE(e.slasher_reward, 0)
					) AS numeric) * 10^9
				) AS total_reward,	
				COALESCE(e.attestations_scheduled, 0) AS attestations_scheduled,
				COALESCE(e.attestation_source_executed, 0) AS attestation_source_executed,
				COALESCE(e.attestations_source_reward, 0) AS attestations_source_reward,
				COALESCE(e.attestation_target_executed, 0) AS attestation_target_executed,
				COALESCE(e.attestations_target_reward, 0) AS attestations_target_reward,
				COALESCE(e.attestation_head_executed, 0) AS attestation_head_executed,
				COALESCE(e.attestations_head_reward, 0) AS attestations_head_reward,
				COALESCE(e.sync_scheduled, 0) AS sync_scheduled,
				COALESCE(e.sync_executed, 0) AS sync_executed,
				COALESCE(e.sync_rewards, 0) AS sync_rewards,
				COALESCE(e.slashed, false) AS slashed,
				COALESCE(e.slasher_reward, 0) AS slasher_reward,
				COALESCE(e.blocks_scheduled, 0) AS blocks_scheduled,
				COALESCE(e.blocks_proposed, 0) AS blocks_proposed,
				COALESCE(e.blocks_el_reward, 0) AS blocks_el_reward,
				COALESCE(e.blocks_cl_attestations_reward, 0) AS blocks_cl_attestations_reward,
				COALESCE(e.blocks_cl_sync_aggregate_reward, 0) AS blocks_cl_sync_aggregate_reward
			FROM validator_dashboard_data_epoch e
			%s
			%s
		) AS subquery
		%s
		%s
		%s`, joinSubquery, whereSubquery, whereQuery, orderQuery, limitQuery)

	err = d.alloyReader.Select(&queryResult, rewardsQuery, queryParams...)
	if err != nil {
		return nil, nil, err
	}

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
		// TODO: Slashing data is not yet available in the db
		if res.Slashed /*|| "Validators slashed" > 0*/ {
			slashedEvent := t.ValidatorHistoryEvent{
				Income: utils.GWeiToWei(big.NewInt(res.SlasherReward)),
			}
			if res.Slashed {
				// if "Validators slashed" > 0 {
				// 	slashedEvent.Status = "partial"
				// }
				slashedEvent.Status = "failed"
			} /*else if "Validators slashed" > 0 {
				slashedEvent.Status = "success"
			}*/
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
