package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type ValidatorDashboardSummaryRow struct {
	GroupId                int64           `db:"result_group_id"`
	GroupName              string          `db:"group_name"`
	ValidatorIndices       []uint64        `db:"validator_indices"`
	ClRewards              int64           `db:"cl_rewards"`
	AttestationReward      decimal.Decimal `db:"attestations_reward"`
	AttestationIdealReward decimal.Decimal `db:"attestations_ideal_reward"`
	AttestationsObserved   uint64          `db:"attestations_observed"`
	AttestationsScheduled  uint64          `db:"attestations_scheduled"`
	BlocksProposed         uint64          `db:"blocks_proposed"`
	BlocksScheduled        uint64          `db:"blocks_scheduled"`
	SyncExecuted           uint64          `db:"sync_executed"`
	SyncScheduled          uint64          `db:"sync_scheduled"`
	MinEpochStart          int64           `db:"min_epoch_start"`
	MaxEpochEnd            int64           `db:"max_epoch_end"`
}

type ValidatorDashboardSummaryResult []ValidatorDashboardSummaryRow

func (d *DataAccessService) addValidatorsToQuery(ds *goqu.SelectDataset, dashboardId t.VDBId, validators []t.VDBValidator) *goqu.SelectDataset {
	if len(validators) > 0 {
		// If validators are provided, use the default group ID and filter by the validators
		ds = ds.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("b.proposer = ANY(?)", pq.Array(validators)))
	} else {
		// If no validators are provided, handle based on whether groups are aggregated
		if dashboardId.AggregateGroups {
			// Use the default group ID if groups are aggregated
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))
		} else {
			// Use the group ID from the validators table if groups are not aggregated
			ds = ds.
				SelectAppend(goqu.L("v.group_id AS result_group_id"))
		}

		// Join with the validators table and filter by dashboard ID
		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("b.proposer = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))
	}

	return ds
}

func (d *DataAccessService) GetValidatorDashboardSummary(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64, protocolModes t.VDBProtocolModes) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	query, args, err := d.buildValidatorDashboardSummaryQuery(ctx, dashboardId, period, colSort, search, protocolModes)
	if err != nil {
		return nil, nil, err
	}

	queryResult, err := d.executeValidatorDashboardSummaryQuery(ctx, query, args)
	if err != nil {
		return nil, nil, err
	}

	result, paging, err := d.processValidatorDashboardSummaryResult(ctx, queryResult, dashboardId, period, colSort, search, protocolModes)
	if err != nil {
		return nil, nil, err
	}

	return result, paging, nil
}

func (d *DataAccessService) getSearchValidator(ctx context.Context, search string, groupNameSearchEnabled bool) (int, error) {
	searchValidator := -1

	if search == "" {
		return searchValidator, nil
	}

	// Handle search term that starts with "0x" (validator public key)
	if strings.HasPrefix(search, "0x") && utils.IsHash(search) {
		search = strings.ToLower(search)

		// Fetch the current validator mapping
		validatorMapping, err := d.services.GetCurrentValidatorMapping()
		if err != nil {
			return -1, fmt.Errorf("error fetching validator mapping: %w", err)
		}

		// Check if the search term matches a validator public key
		if index, ok := validatorMapping.ValidatorIndices[search]; ok {
			searchValidator = int(index)
		} else {
			// No validator index found for the public key
			return -1, nil
		}
	} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
		// Handle search term as a validator index (number)
		searchValidator = int(number)
	} else if !groupNameSearchEnabled {
		// If group name search is not enabled and the search term is not a number, return no results
		return -1, nil
	}

	return searchValidator, nil
}

func (d *DataAccessService) buildValidatorDashboardSummaryQuery(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, colSort t.Sort[enums.VDBSummaryColumn], search string, protocolModes t.VDBProtocolModes) (string, []interface{}, error) {
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return "", nil, err
	}

	groupNameSearchEnabled := !dashboardId.AggregateGroups && dashboardId.Validators == nil
	searchValidator, err := d.getSearchValidator(ctx, search, groupNameSearchEnabled)
	if err != nil {
		return "", nil, err
	}
	if searchValidator == -1 {
		// No validator found for the search term
		return "", nil, nil
	}

	validators := make([]t.VDBValidator, 0)
	if dashboardId.Validators != nil {
		validatorFound := false
		for _, validator := range dashboardId.Validators {
			if searchValidator != -1 && int(validator) == searchValidator {
				validatorFound = true
			}
			validators = append(validators, validator)
		}
		if searchValidator != -1 && !validatorFound {
			return "", nil, nil
		}
	}

	ds := goqu.Dialect("postgres").
		From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, clickhouseTable))).
		With("validators", goqu.L("(SELECT dashboard_id, group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
		Select(
			goqu.L("ARRAY_AGG(r.validator_index) AS validator_indices"),
			goqu.L(`
				(
					SUM(COALESCE(finalizeAggregation(r.balance_end), 0)) +
					SUM(COALESCE(r.withdrawals_amount, 0)) -
					SUM(COALESCE(r.deposits_amount, 0)) -
					SUM(COALESCE(finalizeAggregation(r.balance_start), 0))
				) AS cl_rewards
			`),
			goqu.L("COALESCE(SUM(r.attestations_reward)::decimal, 0) AS attestations_reward"),
			goqu.L("COALESCE(SUM(r.attestations_ideal_reward)::decimal, 0) AS attestations_ideal_reward"),
			goqu.L("COALESCE(SUM(r.attestations_observed), 0) AS attestations_observed"),
			goqu.L("COALESCE(SUM(r.attestations_scheduled), 0) AS attestations_scheduled"),
			goqu.L("COALESCE(SUM(r.blocks_proposed), 0) AS blocks_proposed"),
			goqu.L("COALESCE(SUM(r.blocks_scheduled), 0) AS blocks_scheduled"),
			goqu.L("COALESCE(SUM(r.sync_executed), 0) AS sync_executed"),
			goqu.L("COALESCE(SUM(r.sync_scheduled), 0) AS sync_scheduled"),
			goqu.L("COALESCE(MIN(r.epoch_start), 0) AS min_epoch_start"),
			goqu.L("COALESCE(MAX(r.epoch_end), 0) AS max_epoch_end")).
		GroupBy(goqu.L("result_group_id"))

	if len(validators) > 0 {
		ds = ds.
			SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
			Where(goqu.L("r.validator_index IN ?", validators))
	} else {
		if dashboardId.AggregateGroups {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))
		} else {
			ds = ds.
				SelectAppend(goqu.L("v.group_id AS result_group_id"))
		}

		ds = ds.
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("r.validator_index IN (SELECT validator_index FROM validators)"))

		if groupNameSearchEnabled && (search != "" || colSort.Column == enums.VDBSummaryColumns.Group) {
			ds = ds.
				SelectAppend(goqu.L("g.name AS group_name")).
				InnerJoin(goqu.L("users_val_dashboards_groups g"), goqu.On(goqu.L("v.group_id = g.id AND v.dashboard_id = g.dashboard_id"))).
				GroupByAppend(goqu.L("group_name"))
		}
	}

	return ds.Prepared(true).ToSQL()
}

func (d *DataAccessService) executeValidatorDashboardSummaryQuery(ctx context.Context, query string, args []interface{}) (ValidatorDashboardSummaryResult, error) {
	var queryResult ValidatorDashboardSummaryResult

	err := d.clickhouseReader.SelectContext(ctx, &queryResult, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from table: %w", err)
	}

	return queryResult, nil
}

func (d *DataAccessService) processValidatorDashboardSummaryResult(ctx context.Context, queryResult ValidatorDashboardSummaryResult, dashboardId t.VDBId, period enums.TimePeriod, colSort t.Sort[enums.VDBSummaryColumn], search string, protocolModes t.VDBProtocolModes) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	result := make([]t.VDBSummaryTableRow, 0)
	var paging t.Paging
	if len(queryResult) == 0 {
		return result, &paging, nil
	}

	groupNameSearchEnabled := !dashboardId.AggregateGroups && dashboardId.Validators == nil
	searchValidator, err := d.getSearchValidator(ctx, search, groupNameSearchEnabled)
	if err != nil {
		return nil, nil, err
	}
	if searchValidator == -1 {
		// No validator found for the search term
		return nil, nil, nil
	}

	efficiency, err := d.services.GetCurrentEfficiencyInfo()
	if err != nil {
		return nil, nil, err
	}
	averageNetworkEfficiency := utils.CalculateTotalEfficiency(
		efficiency.AttestationEfficiency[period], efficiency.ProposalEfficiency[period], efficiency.SyncEfficiency[period])

	epochMin := int64(math.MaxInt32)
	epochMax := int64(0)

	for _, row := range queryResult {
		if row.MinEpochStart < epochMin {
			epochMin = row.MinEpochStart
		}
		if row.MaxEpochEnd > epochMax {
			epochMax = row.MaxEpochEnd
		}
	}

	elRewards, err := d.getElRewards(ctx, epochMin, epochMax, dashboardId, queryResult[0].ValidatorIndices)
	if err != nil {
		return nil, nil, err
	}

	currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err := d.getCurrentAndUpcomingSyncCommittees(ctx, cache.LatestEpoch.Get())
	if err != nil {
		return nil, nil, err
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Sort by group name, after this the name is no longer relevant
	if groupNameSearchEnabled && colSort.Column == enums.VDBSummaryColumns.Group {
		sort.Slice(queryResult, func(i, j int) bool {
			if colSort.Desc {
				return queryResult[i].GroupName > queryResult[j].GroupName
			} else {
				return queryResult[i].GroupName < queryResult[j].GroupName
			}
		})
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Calculate the result
	total := struct {
		GroupId                int64
		Status                 t.VDBSummaryStatus
		Validators             t.VDBSummaryValidators
		AttestationReward      decimal.Decimal
		AttestationIdealReward decimal.Decimal
		AttestationsObserved   uint64
		AttestationsScheduled  uint64
		BlocksProposed         uint64
		BlocksScheduled        uint64
		SyncExecuted           uint64
		SyncScheduled          uint64
		Reward                 t.ClElValue[decimal.Decimal]
	}{
		GroupId: t.AllGroups,
	}

	for _, queryEntry := range queryResult {
		resultEntry := t.VDBSummaryTableRow{
			GroupId:                  queryEntry.GroupId,
			AverageNetworkEfficiency: averageNetworkEfficiency,
		}

		// Status
		for _, validatorIndex := range queryEntry.ValidatorIndices {
			if currentSyncCommitteeValidators[validatorIndex] {
				resultEntry.Status.CurrentSyncCount++
			}
			if upcomingSyncCommitteeValidators[validatorIndex] {
				resultEntry.Status.UpcomingSyncCount++
			}
		}
		total.Status.CurrentSyncCount += resultEntry.Status.CurrentSyncCount
		total.Status.UpcomingSyncCount += resultEntry.Status.UpcomingSyncCount

		// Validator statuses
		validatorMapping, err := d.services.GetCurrentValidatorMapping()
		if err != nil {
			return nil, nil, err
		}

		for _, validator := range queryEntry.ValidatorIndices {
			metadata := validatorMapping.ValidatorMetadata[validator]

			// As deposited and pending validators are neither online nor offline they are counted as the third state (exited)
			switch constypes.ValidatorDbStatus(metadata.Status) {
			case constypes.DbDeposited:
				resultEntry.Validators.Exited++
			case constypes.DbPending:
				resultEntry.Validators.Exited++
			case constypes.DbActiveOnline, constypes.DbExitingOnline, constypes.DbSlashingOnline:
				resultEntry.Validators.Online++
			case constypes.DbActiveOffline, constypes.DbExitingOffline, constypes.DbSlashingOffline:
				resultEntry.Validators.Offline++
			case constypes.DbSlashed:
				resultEntry.Validators.Exited++
				resultEntry.Status.SlashedCount++
			case constypes.DbExited:
				resultEntry.Validators.Exited++
			}
		}
		total.Validators.Online += resultEntry.Validators.Online
		total.Validators.Offline += resultEntry.Validators.Offline
		total.Validators.Exited += resultEntry.Validators.Exited
		total.Status.SlashedCount += resultEntry.Status.SlashedCount

		// Attestations
		resultEntry.Attestations.Success = queryEntry.AttestationsObserved
		resultEntry.Attestations.Failed = queryEntry.AttestationsScheduled - queryEntry.AttestationsObserved

		// Proposals
		resultEntry.Proposals.Success = queryEntry.BlocksProposed
		resultEntry.Proposals.Failed = queryEntry.BlocksScheduled - queryEntry.BlocksProposed

		// Rewards
		resultEntry.Reward.Cl = utils.GWeiToWei(big.NewInt(queryEntry.ClRewards))
		if _, ok := elRewards[queryEntry.GroupId]; ok {
			resultEntry.Reward.El = elRewards[queryEntry.GroupId]
		}
		total.Reward.Cl = total.Reward.Cl.Add(resultEntry.Reward.Cl)
		total.Reward.El = total.Reward.El.Add(resultEntry.Reward.El)

		// Efficiency
		var attestationEfficiency, proposerEfficiency, syncEfficiency sql.NullFloat64
		if !queryEntry.AttestationIdealReward.IsZero() {
			attestationEfficiency.Float64 = queryEntry.AttestationReward.Div(queryEntry.AttestationIdealReward).InexactFloat64()
			attestationEfficiency.Valid = true
		}
		if queryEntry.BlocksScheduled > 0 {
			proposerEfficiency.Float64 = float64(queryEntry.BlocksProposed) / float64(queryEntry.BlocksScheduled)
			proposerEfficiency.Valid = true
		}
		if queryEntry.SyncScheduled > 0 {
			syncEfficiency.Float64 = float64(queryEntry.SyncExecuted) / float64(queryEntry.SyncScheduled)
			syncEfficiency.Valid = true
		}
		resultEntry.Efficiency = utils.CalculateTotalEfficiency(attestationEfficiency, proposerEfficiency, syncEfficiency)

		// Add the duties info to the total
		total.AttestationReward = total.AttestationReward.Add(queryEntry.AttestationReward)
		total.AttestationIdealReward = total.AttestationIdealReward.Add(queryEntry.AttestationIdealReward)
		total.AttestationsObserved += queryEntry.AttestationsObserved
		total.AttestationsScheduled += queryEntry.AttestationsScheduled
		total.BlocksProposed += queryEntry.BlocksProposed
		total.BlocksScheduled += queryEntry.BlocksScheduled
		total.SyncExecuted += queryEntry.SyncExecuted
		total.SyncScheduled += queryEntry.SyncScheduled

		// If the search permits it add the entry to the result
		if search != "" {
			prefixSearch := strings.ToLower(search)
			for _, validatorIndex := range queryEntry.ValidatorIndices {
				if searchValidator != -1 && validatorIndex == uint64(searchValidator) ||
					(groupNameSearchEnabled && strings.HasPrefix(strings.ToLower(queryEntry.GroupName), prefixSearch)) {
					result = append(result, resultEntry)
					break
				}
			}
		} else {
			result = append(result, resultEntry)
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Sort the result
	// For sorting consider a 0/0 => 0% as lower than a 0/5 => 0%
	var sortParam func(resultEntry t.VDBSummaryTableRow) float64
	switch colSort.Column {
	case enums.VDBSummaryColumns.Validators:
		sortParam = func(resultEntry t.VDBSummaryTableRow) float64 {
			divisor := float64(resultEntry.Validators.Online + resultEntry.Validators.Offline)
			if divisor == 0 {
				return -1
			}
			return float64(resultEntry.Validators.Online) / divisor
		}
	case enums.VDBSummaryColumns.Efficiency:
		sortParam = func(resultEntry t.VDBSummaryTableRow) float64 {
			return resultEntry.Efficiency
		}
	case enums.VDBSummaryColumns.Attestations:
		sortParam = func(resultEntry t.VDBSummaryTableRow) float64 {
			divisor := float64(resultEntry.Attestations.Success + resultEntry.Attestations.Failed)
			if divisor == 0 {
				return -1
			}
			return float64(resultEntry.Attestations.Success) / divisor
		}
	case enums.VDBSummaryColumns.Proposals:
		sortParam = func(resultEntry t.VDBSummaryTableRow) float64 {
			divisor := float64(resultEntry.Proposals.Success + resultEntry.Proposals.Failed)
			if divisor == 0 {
				return -1
			}
			return float64(resultEntry.Proposals.Success) / divisor
		}
	case enums.VDBSummaryColumns.Reward:
		rewardSortParam := func(resultEntry t.VDBSummaryTableRow) decimal.Decimal {
			return resultEntry.Reward.Cl.Add(resultEntry.Reward.El)
		}
		sort.Slice(result, func(i, j int) bool {
			if colSort.Desc {
				return rewardSortParam(result[i]).GreaterThan(rewardSortParam(result[j]))
			} else {
				return rewardSortParam(result[i]).LessThan(rewardSortParam(result[j]))
			}
		})
	case enums.VDBSummaryColumns.Group:
	default:
		return nil, nil, fmt.Errorf("error sorting data: unexpected sorting type")
	}

	if sortParam != nil {
		sort.Slice(result, func(i, j int) bool {
			if colSort.Desc {
				return sortParam(result[i]) > sortParam(result[j])
			} else {
				return sortParam(result[i]) < sortParam(result[j])
			}
		})
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Calculate the total
	if len(queryResult) > 1 && len(result) > 0 {
		// We have more than one group and at least one group remains after the filtering so we need to show the total row
		totalEntry := t.VDBSummaryTableRow{
			GroupId:                  total.GroupId,
			Status:                   total.Status,
			Validators:               total.Validators,
			AverageNetworkEfficiency: averageNetworkEfficiency,
			Reward:                   total.Reward,
		}

		// Attestations
		totalEntry.Attestations.Success = total.AttestationsObserved
		totalEntry.Attestations.Failed = total.AttestationsScheduled - total.AttestationsObserved

		// Proposals
		totalEntry.Proposals.Success = total.BlocksProposed
		totalEntry.Proposals.Failed = total.BlocksScheduled - total.BlocksProposed

		// Efficiency
		var totalAttestationEfficiency, totalProposerEfficiency, totalSyncEfficiency sql.NullFloat64
		if !total.AttestationIdealReward.IsZero() {
			totalAttestationEfficiency.Float64 = total.AttestationReward.Div(total.AttestationIdealReward).InexactFloat64()
			totalAttestationEfficiency.Valid = true
		}
		if total.BlocksScheduled > 0 {
			totalProposerEfficiency.Float64 = float64(total.BlocksProposed) / float64(total.BlocksScheduled)
			totalProposerEfficiency.Valid = true
		}
		if total.SyncScheduled > 0 {
			totalSyncEfficiency.Float64 = float64(total.SyncExecuted) / float64(total.SyncScheduled)
			totalSyncEfficiency.Valid = true
		}
		totalEntry.Efficiency = utils.CalculateTotalEfficiency(totalAttestationEfficiency, totalProposerEfficiency, totalSyncEfficiency)

		result = append([]t.VDBSummaryTableRow{totalEntry}, result...)
	}

	paging.TotalCount = uint64(len(result))

	return result, &paging, nil
}

type ElRewardsQueryResult struct {
	GroupId   int64           `db:"result_group_id"`
	ElRewards decimal.Decimal `db:"el_rewards"`
}

func (d *DataAccessService) getElRewards(ctx context.Context, epochMin, epochMax int64, dashboardId t.VDBId, validators []t.VDBValidator) (map[int64]decimal.Decimal, error) {
	query, args, err := d.buildElRewardsQuery(epochMin, epochMax, dashboardId, validators)
	if err != nil {
		return nil, fmt.Errorf("error building EL rewards query: %w", err)
	}

	elRewardsQueryResult, err := d.executeElRewardsQuery(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("error executing EL rewards query: %w", err)
	}

	elRewards := d.processElRewardsQueryResult(elRewardsQueryResult)
	return elRewards, nil
}

func (d *DataAccessService) buildElRewardsQuery(epochMin, epochMax int64, dashboardId t.VDBId, validators []t.VDBValidator) (string, []interface{}, error) {
	ds := goqu.Dialect("postgres").
		Select(
			goqu.COALESCE(goqu.SUM(goqu.I("value")), 0).As("el_rewards")).
		From(goqu.I("execution_rewards_finalized").As("b")).
		Where(goqu.L("b.epoch >= ? AND b.epoch <= ?", epochMin, epochMax)).
		GroupBy(goqu.L("result_group_id"))
	ds = d.addValidatorsToQuery(ds, dashboardId, validators)
	return ds.Prepared(true).ToSQL()
}

func (d *DataAccessService) executeElRewardsQuery(ctx context.Context, query string, args []interface{}) ([]ElRewardsQueryResult, error) {
	var elRewardsQueryResult []ElRewardsQueryResult

	err := d.alloyReader.SelectContext(ctx, &elRewardsQueryResult, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from table blocks: %w", err)
	}

	return elRewardsQueryResult, nil
}

func (d *DataAccessService) processElRewardsQueryResult(elRewardsQueryResult []ElRewardsQueryResult) map[int64]decimal.Decimal {
	elRewards := make(map[int64]decimal.Decimal)
	for _, entry := range elRewardsQueryResult {
		elRewards[entry.GroupId] = entry.ElRewards
	}
	return elRewards
}

type LastScheduledBlockAndSyncResult struct {
	LastScheduledBlockEpoch *int64 `db:"last_scheduled_block_epoch"`
	LastSyncEpoch           *int64 `db:"last_scheduled_sync_epoch"`
}

func (d *DataAccessService) getLastScheduledBlockAndSyncDate(ctx context.Context, dashboardId t.VDBId, groupId int64) (time.Time, time.Time, error) {
	ds, err := d.buildLastScheduledBlockAndSyncQuery(dashboardId, groupId)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error building query: %w", err)
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error preparing query: %w", err)
	}

	row, err := d.executeLastScheduledBlockAndSyncQuery(ctx, query, args)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error executing query: %w", err)
	}

	lastScheduledBlockTime, lastSyncTime, err := d.processLastScheduledBlockAndSyncResult(row)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error processing result: %w", err)
	}

	return lastScheduledBlockTime, lastSyncTime, nil
}

func (d *DataAccessService) buildLastScheduledBlockAndSyncQuery(dashboardId t.VDBId, groupId int64) (*goqu.SelectDataset, error) {
	clickhouseTotalTable, _, err := d.getTablesForPeriod(enums.AllTime)
	if err != nil {
		return nil, fmt.Errorf("error getting table for period: %w", err)
	}

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("MAX(last_scheduled_block_epoch) as last_scheduled_block_epoch"),
			goqu.L("MAX(last_scheduled_sync_epoch) as last_scheduled_sync_epoch")).
		From(goqu.L(fmt.Sprintf(`%s AS r FINAL`, clickhouseTotalTable)))

	if dashboardId.Validators == nil {
		ds = ds.
			With("validators", goqu.L("(SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = ? AND (group_id = ? OR ?::smallint = -1))", dashboardId.Id, groupId, groupId)).
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("validator_index IN (SELECT validator_index FROM validators)"))
	} else {
		ds = ds.
			Where(goqu.L("validator_index IN ?", dashboardId.Validators))
	}

	return ds, nil
}

func (d *DataAccessService) executeLastScheduledBlockAndSyncQuery(ctx context.Context, query string, args []interface{}) (LastScheduledBlockAndSyncResult, error) {
	var row LastScheduledBlockAndSyncResult
	err := d.clickhouseReader.GetContext(ctx, &row, query, args...)
	if err != nil {
		return LastScheduledBlockAndSyncResult{}, fmt.Errorf("error executing query: %w", err)
	}
	return row, nil
}

func (d *DataAccessService) processLastScheduledBlockAndSyncResult(row LastScheduledBlockAndSyncResult) (time.Time, time.Time, error) {
	if row.LastScheduledBlockEpoch == nil || row.LastSyncEpoch == nil {
		return time.Time{}, time.Time{}, nil
	}

	lastScheduledBlockTime := utils.EpochToTime(uint64(*row.LastScheduledBlockEpoch))
	lastSyncTime := utils.EpochToTime(uint64(*row.LastSyncEpoch))

	return lastScheduledBlockTime, lastSyncTime, nil
}
