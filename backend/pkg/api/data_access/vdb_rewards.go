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
				indexSearch = int64(*index)
			} else {
				// No validator index for pubkey found, return empty results
				return result, &paging, nil
			}
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			indexSearch = int64(number)
			epochSearch = int64(number)
		}
	}

	queryResult := []struct {
		Epoch                 uint64          `db:"epoch"`
		GroupId               uint64          `db:"group_id"`
		ClRewards             int64           `db:"cl_rewards"`
		ElRewards             decimal.Decimal `db:"el_rewards"`
		AttestationsScheduled uint64          `db:"attestations_scheduled"`
		AttestationsExecuted  uint64          `db:"attestations_executed"`
		BlocksScheduled       uint64          `db:"blocks_scheduled"`
		BlocksProposed        uint64          `db:"blocks_proposed"`
		SyncScheduled         uint64          `db:"sync_scheduled"`
		SyncExecuted          uint64          `db:"sync_executed"`
		SlashedViolation      uint64          `db:"slashed_violation"`
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
		SUM(COALESCE(e.slashed_violation, 0)) AS slashed_violation
		`

	if dashboardId.Validators == nil {
		queryParams = append(queryParams, dashboardId.Id)
		whereQuery := fmt.Sprintf("WHERE v.dashboard_id = $%d", len(queryParams))
		if currentCursor.IsValid() {
			queryParams = append(queryParams, currentCursor.Epoch, currentCursor.GroupId)
			whereQuery += fmt.Sprintf(" AND (e.epoch%[1]s$%[2]d OR (e.epoch=$%[2]d AND v.group_id%[1]s$%[3]d))", sortSearchDirection, len(queryParams)-1, len(queryParams))
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
				return result, &paging, nil
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
	for _, res := range queryResult {
		duty := t.VDBRewardesTableDuty{}
		if res.AttestationsScheduled > 0 {
			attestationPercentage := float64(res.AttestationsExecuted) / float64(res.AttestationsScheduled)
			duty.Attestation = &attestationPercentage
		}
		if res.BlocksScheduled > 0 {
			ProposalPercentage := float64(res.BlocksProposed) / float64(res.BlocksScheduled)
			duty.Proposal = &ProposalPercentage
		}
		if res.SyncScheduled > 0 {
			SyncPercentage := float64(res.SyncExecuted) / float64(res.SyncScheduled)
			duty.Sync = &SyncPercentage
		}
		if res.SlashedViolation > 0 {
			slashedViolation := res.SlashedViolation
			duty.Slashing = &slashedViolation
		}
		reward := t.ClElValue[decimal.Decimal]{
			El: res.ElRewards,
			Cl: utils.GWeiToWei(big.NewInt(res.ClRewards)),
		}
		if duty.Attestation != nil || duty.Proposal != nil || duty.Sync != nil || duty.Slashing != nil {
			// Only add groups that had some duty
			result = append(result, t.VDBRewardsTableRow{
				Epoch:   res.Epoch,
				Duty:    duty,
				GroupId: res.GroupId,
				Reward:  reward,
			})
		}
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
	}

	// Reverse the data if the cursor is reversed to correct it to the requested direction
	if currentCursor.IsReverse() {
		slices.Reverse(result)
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
				indexSearch = int64(*index)
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
		validators := make([]uint64, 0)
		for _, validator := range dashboardId.Validators {
			if indexSearch == -1 || validator.Index == uint64(indexSearch) {
				validators = append(validators, validator.Index)
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
