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
	"github.com/gobitfly/beaconchain/pkg/commons/log"
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
				if err != nil {
					if !errors.Is(err, sql.ErrNoRows) {
						return nil, nil, err
					}
				} else {
					// Index in the dashboard, add the group to the search
					queryParams = append(queryParams, groupIdSearch)
					groupIdSearchQuery = fmt.Sprintf("OR v.group_id = $%d", len(queryParams))
				}
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
	result := make([]t.VDBRewardsTableRow, 0)
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
		validators := make([]uint64, 0)
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}

		whereClause := "from validator_dashboard_data_epoch where validator_index = any($1) and epoch = $2"
		query = fmt.Sprintf("%s %s", query, whereClause)
		err := d.alloyReader.Select(&rows, query, pq.Array(validators), epoch)
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
	// bar chart for the CL and EL rewards for each group for each epoch. NO series for all groups combined
	// series id is group id, series property is 'cl' or 'el'

	queryResult := []struct {
		Epoch     uint64          `db:"epoch"`
		GroupId   uint64          `db:"group_id"`
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
		queryParams = append(queryParams, dashboardId.Id)
		rewardsQuery = fmt.Sprintf(`
			SELECT
				e.epoch,
				v.group_id,
				%s
			FROM validator_dashboard_data_epoch e
			INNER JOIN users_val_dashboards_validators v ON e.validator_index = v.validator_index
			WHERE v.dashboard_id = $%d
			GROUP BY e.epoch, v.group_id
			ORDER BY e.epoch, v.group_id`, rewardsDataQuery, len(queryParams))
	} else {
		// In case a list of validators is provided set the group to the default id
		validators := make([]uint64, 0)
		for _, validator := range dashboardId.Validators {
			validators = append(validators, validator.Index)
		}

		queryParams = append(queryParams, t.DefaultGroupId, pq.Array(validators))
		rewardsQuery = fmt.Sprintf(`
			SELECT
				e.epoch,
				$%d::smallint AS group_id,
				%s
			FROM validator_dashboard_data_epoch e
			WHERE e.validator_index = ANY($%d)
			GROUP BY e.epoch
			ORDER BY e.epoch`, len(queryParams)-1, rewardsDataQuery, len(queryParams))
	}

	err := d.alloyReader.Select(&queryResult, rewardsQuery, queryParams...)
	if err != nil {
		return nil, err
	}

	// Create a map structure to store the data
	epochData := make(map[uint64]map[uint64]t.ClElValue[decimal.Decimal])
	epochList := make([]uint64, 0)

	for _, res := range queryResult {
		if _, ok := epochData[res.Epoch]; !ok {
			epochData[res.Epoch] = make(map[uint64]t.ClElValue[decimal.Decimal])
			epochList = append(epochList, res.Epoch)
		}

		epochData[res.Epoch][res.GroupId] = t.ClElValue[decimal.Decimal]{
			El: res.ElRewards,
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

func (d *DataAccessService) GetValidatorDashboardDuties(dashboardId t.VDBId, epoch uint64, groupId int64, cursor string, colSort t.Sort[enums.VDBDutiesColumn], search string, limit uint64) ([]t.VDBEpochDutiesTableRow, *t.Paging, error) {
	// WORKING spletka
	return d.dummy.GetValidatorDashboardDuties(dashboardId, epoch, groupId, cursor, colSort, search, limit)
}
