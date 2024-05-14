package dataaccess

import (
	"database/sql"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (d *DataAccessService) GetValidatorDashboardRewards(dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRewardsColumn], search string, limit uint64) ([]t.VDBRewardsTableRow, *t.Paging, error) {
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
			if utils.IsHash(search) {
				// Get the current validator state to convert pubkey to index
				validatorMapping, releaseLock, err := d.services.GetCurrentValidatorMapping()
				defer releaseLock()
				if err != nil {
					return nil, nil, err
				}
				if index, ok := validatorMapping.ValidatorIndices[search]; ok {
					indexSearch = int64(*index)
				} else {
					// No validator index for pubkey found, return empty results
					return nil, &paging, nil
				}
			}
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			indexSearch = int64(number)
			epochSearch = int64(number)
		}
	}

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
		queryParams = append(queryParams, dashboardId.Id)
		whereQuery := fmt.Sprintf("WHERE v.dashboard_id = $%d", len(queryParams))
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

		joinQuery := ""
		if search != "" {
			// Join the groups table to get the group name which we search for
			joinQuery = "INNER JOIN users_val_dashboards_groups g ON v.group_id = g.id AND v.dashboard_id = g.dashboard_id"

			epochSearchQuery := ""
			if epochSearch != -1 {
				queryParams = append(queryParams, epochSearch)
				epochSearchQuery = fmt.Sprintf("OR e.epoch = $%d", len(queryParams))
			}
			groupIdSearchQuery := ""
			if indexSearch != -1 {
				// Check in what group if any the searched for index is
				var groupIdSearch uint64
				err = d.alloyReader.Get(&groupIdSearch, `
					SELECT
						group_id
					FROM users_val_dashboards_validators
					WHERE dashboard_id = $1 AND validator_index = $2
					`, dashboardId.Id, indexSearch)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					// Index not in the dashboard
					return nil, nil, err
				}

				queryParams = append(queryParams, groupIdSearch)
				groupIdSearchQuery = fmt.Sprintf("OR v.group_id = $%d", len(queryParams))
			}
			queryParams = append(queryParams, search)
			whereQuery += fmt.Sprintf(` AND (g.name ILIKE ($%d||'%%') %s %s)`, len(queryParams), epochSearchQuery, groupIdSearchQuery)
		}

		orderQuery := fmt.Sprintf("ORDER BY e.epoch %[1]s, v.group_id %[1]s", sortSearchOrder)

		rewardsQuery = fmt.Sprintf(`
			SELECT
				e.epoch,
				v.group_id,
				%s
			FROM validator_dashboard_data_epoch e
			INNER JOIN users_val_dashboards_validators v ON e.validator_index = v.validator_index
			%s
			%s
			GROUP BY e.epoch, v.group_id
			%s`, rewardsDataQuery, joinQuery, whereQuery, orderQuery)
	} else {
		// In case a list of validators is provided set the group to the default id
		validators := make([]uint64, 0)
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}

		queryParams = append(queryParams, pq.Array(validators))
		whereQuery := fmt.Sprintf("WHERE e.validator_index = ANY($%d)", len(queryParams))
		if currentCursor.IsValid() {
			queryParams = append(queryParams, currentCursor.Epoch)
			whereQuery += fmt.Sprintf(" AND e.epoch%s$%d", sortSearchDirection, len(queryParams))
		}
		if search != "" {
			if epochSearch != -1 || indexSearch != -1 {
				found := false
				if indexSearch != -1 {
					// Find whether the index is in the list of validators
					// If it is then show all the data
					for _, validator := range dashboardId.Validators {
						if validator.Index == uint64(indexSearch) {
							found = true
							break
						}
					}
				}
				if !found && epochSearch != -1 {
					queryParams = append(queryParams, epochSearch)
					whereQuery += fmt.Sprintf(" AND e.epoch = $%d", len(queryParams))
				}
			} else {
				return nil, &paging, nil
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

	// Create the result
	resultWoTotal := make([]t.VDBRewardsTableRow, 0)
	result := make([]t.VDBRewardsTableRow, 0)

	type TotalEpochInfo struct {
		GroupCount            uint8
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
			totalEpochInfo[res.Epoch].GroupCount++
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
		if totalInfo.GroupCount < 2 {
			// Only one group, no need to show the data
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

	// Place the total rewards in the result data at the correct position
	// Ascending or descending order makes no difference but the cursor direction does
	if len(totalRewards) > 0 {
		previousEpoch := int64(-1)
		if currentCursor.IsValid() && !currentCursor.IsReverse() {
			previousEpoch = int64(currentCursor.Epoch)
		}
		for _, res := range resultWoTotal {
			// If we reach the cursor which should only happen if the cursor is reversed don't include it and stop
			if currentCursor.Epoch == res.Epoch &&
				(currentCursor.GroupId == res.GroupId || currentCursor.GroupId == t.AllGroups) {
				break
			}

			if previousEpoch != int64(res.Epoch) {
				if totalReward, ok := totalRewards[res.Epoch]; ok {
					result = append(result, totalReward)
				}
			}
			result = append(result, res)
			previousEpoch = int64(res.Epoch)
		}
	} else {
		result = resultWoTotal
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
	// WORKING Peter
	return d.dummy.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
}

func (d *DataAccessService) GetValidatorDashboardRewardsChart(dashboardId t.VDBId) (*t.ChartData[int, decimal.Decimal], error) {
	// WORKING spletka
	// bar chart for the CL and EL rewards for each group for each epoch. NO series for all groups combined
	// series id is group id, series property is 'cl' or 'el'
	return d.dummy.GetValidatorDashboardRewardsChart(dashboardId)
}

func (d *DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	// WORKING spletka
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, groupId, cursor, colSort, search, limit)
}
