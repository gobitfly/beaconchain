package dataaccess

import (
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
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (d *DataAccessService) GetValidatorDashboardSummary(dashboardId t.VDBId, period enums.TimePeriod, cursor string, colSort t.Sort[enums.VDBSummaryColumn], search string, limit uint64) ([]t.VDBSummaryTableRow, *t.Paging, error) {
	result := make([]t.VDBSummaryTableRow, 0)
	var paging t.Paging

	wg := errgroup.Group{}

	// Get the table name based on the period
	// TODO: validator_dashboard_data_rolling_hourly does not exist yet
	tableName := ""
	switch period {
	case enums.TimePeriods.AllTime:
		tableName = "validator_dashboard_data_rolling_total"
	case enums.TimePeriods.Last1h:
		fallthrough
	case enums.TimePeriods.Last24h:
		tableName = "validator_dashboard_data_rolling_daily"
	case enums.TimePeriods.Last7d:
		tableName = "validator_dashboard_data_rolling_weekly"
	case enums.TimePeriods.Last30d:
		tableName = "validator_dashboard_data_rolling_monthly"
	}

	// Searching for a group name is not supported when aggregating groups or for guest dashboards
	groupNameSearchEnabled := !dashboardId.AggregateGroups && dashboardId.Validators == nil

	// Analyze the search term
	searchValidator := -1
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
				searchValidator = int(index)
			} else {
				// No validator index for pubkey found, return empty results
				releaseLock()
				return result, &paging, nil
			}
			releaseLock()
		} else if number, err := strconv.ParseUint(search, 10, 64); err == nil {
			searchValidator = int(number)
		} else if !groupNameSearchEnabled {
			return result, &paging, nil
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Fill the validators list if we have a guest dashboard
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
			// The searched validator is not part of the dashboard
			return result, &paging, nil
		}
	}

	// ------------------------------------------------------------------------------------------------------------------
	// Get the average network efficiency
	efficiency, releaseEfficiencyLock, err := d.services.GetCurrentEfficiencyInfo()
	if err != nil {
		releaseEfficiencyLock()
		return nil, nil, err
	}
	averageNetworkEfficiency := d.calculateTotalEfficiency(
		efficiency.AttestationEfficiency[period], efficiency.ProposalEfficiency[period], efficiency.SyncEfficiency[period])
	releaseEfficiencyLock()

	// ------------------------------------------------------------------------------------------------------------------
	// Build the main query and get the data
	var queryResult []struct {
		GroupId                int64           `db:"result_group_id"`
		GroupName              string          `db:"group_name"`
		ValidatorIndices       pq.Int64Array   `db:"validator_indices"`
		ClRewards              int64           `db:"cl_rewards"`
		AttestationReward      decimal.Decimal `db:"attestations_reward"`
		AttestationIdealReward decimal.Decimal `db:"attestations_ideal_reward"`
		AttestationsExecuted   uint64          `db:"attestations_executed"`
		AttestationsScheduled  uint64          `db:"attestations_scheduled"`
		BlocksProposed         uint64          `db:"blocks_proposed"`
		BlocksScheduled        uint64          `db:"blocks_scheduled"`
		SyncExecuted           uint64          `db:"sync_executed"`
		SyncScheduled          uint64          `db:"sync_scheduled"`
	}

	wg.Go(func() error {
		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("ARRAY_AGG(r.validator_index) AS validator_indices"),
				goqu.L("SUM(COALESCE(r.attestations_reward, 0) + COALESCE(r.blocks_cl_reward, 0) + COALESCE(r.sync_rewards, 0) + COALESCE(r.slasher_reward, 0)) AS cl_rewards"),
				goqu.L("COALESCE(SUM(r.attestations_reward)::decimal, 0) AS attestations_reward"),
				goqu.L("COALESCE(SUM(r.attestations_ideal_reward)::decimal, 0) AS attestations_ideal_reward"),
				goqu.L("COALESCE(SUM(r.attestations_executed), 0) AS attestations_executed"),
				goqu.L("COALESCE(SUM(r.attestations_scheduled), 0) AS attestations_scheduled"),
				goqu.L("COALESCE(SUM(r.blocks_proposed), 0) AS blocks_proposed"),
				goqu.L("COALESCE(SUM(r.blocks_scheduled), 0) AS blocks_scheduled"),
				goqu.L("COALESCE(SUM(r.sync_executed), 0) AS sync_executed"),
				goqu.L("COALESCE(SUM(r.sync_scheduled), 0) AS sync_scheduled")).
			From(goqu.T(tableName).As("r")).
			GroupBy(goqu.L("result_group_id"))

		if len(validators) > 0 {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
				Where(goqu.L("r.validator_index = ANY(?)", pq.Array(validators)))
		} else {
			if dashboardId.AggregateGroups {
				ds = ds.
					SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))
			} else {
				ds = ds.
					SelectAppend(goqu.L("v.group_id AS result_group_id"))
			}

			ds = ds.
				InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
				Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

			if groupNameSearchEnabled && (search != "" || colSort.Column == enums.VDBSummaryColumns.Group) {
				// Get the group names since we can filter and/or sort for them
				ds = ds.
					SelectAppend(goqu.L("g.name AS group_name")).
					InnerJoin(goqu.L("users_val_dashboards_groups g"), goqu.On(goqu.L("v.group_id = g.id AND v.dashboard_id = g.dashboard_id"))).
					GroupByAppend(goqu.L("group_name"))
			}
		}

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
		}
		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the EL rewards
	elRewards := make(map[int64]decimal.Decimal)
	wg.Go(func() error {
		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("SUM(COALESCE(rb.value, ep.fee_recipient_reward * 1e18, 0)) AS el_rewards")).
			From(goqu.T(tableName).As("r")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("b.epoch >= r.epoch_start AND b.epoch <= r.epoch_end AND r.validator_index = b.proposer AND b.status = '1'"))).
			LeftJoin(goqu.L("execution_payloads ep"), goqu.On(goqu.L("ep.block_hash = b.exec_block_hash"))).
			LeftJoin(goqu.L("relays_blocks rb"), goqu.On(goqu.L("rb.exec_block_hash = b.exec_block_hash"))).
			GroupBy(goqu.L("result_group_id"))

		if len(validators) > 0 {
			ds = ds.
				SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId)).
				Where(goqu.L("r.validator_index = ANY(?)", pq.Array(validators)))
		} else {
			if dashboardId.AggregateGroups {
				ds = ds.
					SelectAppend(goqu.L("?::smallint AS result_group_id", t.DefaultGroupId))
			} else {
				ds = ds.
					SelectAppend(goqu.L("v.group_id AS result_group_id"))
			}

			ds = ds.
				InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
				Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))
		}

		var queryResult []struct {
			GroupId   int64           `db:"result_group_id"`
			ElRewards decimal.Decimal `db:"el_rewards"`
		}

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.alloyReader.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
		}

		for _, entry := range queryResult {
			elRewards[entry.GroupId] = entry.ElRewards
		}
		return nil
	})

	// ------------------------------------------------------------------------------------------------------------------
	// Get the current and next sync committee validators
	latestEpoch := cache.LatestEpoch.Get()
	currentSyncCommitteeValidators := make(map[uint64]bool)
	upcomingSyncCommitteeValidators := make(map[uint64]bool)
	wg.Go(func() error {
		currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)
		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("validatorindex"),
				goqu.L("period")).
			From("sync_committees").
			Where(goqu.L("period IN (?, ?)", currentSyncPeriod, currentSyncPeriod+1))

		var queryResult []struct {
			ValidatorIndex uint64 `db:"validatorindex"`
			Period         uint64 `db:"period"`
		}

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %v", err)
		}

		err = d.readerDb.Select(&queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table %s: %v", tableName, err)
		}

		for _, queryEntry := range queryResult {
			if queryEntry.Period == currentSyncPeriod {
				currentSyncCommitteeValidators[queryEntry.ValidatorIndex] = true
			} else {
				upcomingSyncCommitteeValidators[queryEntry.ValidatorIndex] = true
			}
		}

		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving validator dashboard summary data: %v", err)
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
		AttestationsExecuted   uint64
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
		uiValidatorIndices := make([]uint64, len(queryEntry.ValidatorIndices))
		for i, validatorIndex := range queryEntry.ValidatorIndices {
			uiValidatorIndices[i] = uint64(validatorIndex)
		}

		resultEntry := t.VDBSummaryTableRow{
			GroupId:                  queryEntry.GroupId,
			AverageNetworkEfficiency: averageNetworkEfficiency,
		}

		// Status
		for _, validatorIndex := range uiValidatorIndices {
			if currentSyncCommitteeValidators[validatorIndex] {
				resultEntry.Status.CurrentSyncCount++
			}
			if upcomingSyncCommitteeValidators[validatorIndex] {
				resultEntry.Status.UpcomingSyncCount++
			}
		}

		// Validator statuses
		validatorStatuses, err := d.getValidatorStatuses(uiValidatorIndices)
		if err != nil {
			return nil, nil, err
		}
		for _, status := range validatorStatuses {
			if status == enums.ValidatorStatuses.Online {
				resultEntry.Validators.Online++
			} else if status == enums.ValidatorStatuses.Offline {
				resultEntry.Validators.Offline++
			} else {
				if status == enums.ValidatorStatuses.Slashed {
					resultEntry.Status.SlashedCount++
				}
				resultEntry.Validators.Exited++
			}
		}
		total.Validators.Online += resultEntry.Validators.Online
		total.Validators.Offline += resultEntry.Validators.Offline
		total.Validators.Exited += resultEntry.Validators.Exited
		total.Status.SlashedCount += resultEntry.Status.SlashedCount

		// Attestations
		resultEntry.Attestations.Success = queryEntry.AttestationsExecuted
		resultEntry.Attestations.Failed = queryEntry.AttestationsScheduled - queryEntry.AttestationsExecuted

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
		resultEntry.Efficiency = d.calculateTotalEfficiency(attestationEfficiency, proposerEfficiency, syncEfficiency)

		// Add the duties info to the total
		total.AttestationReward = total.AttestationReward.Add(queryEntry.AttestationReward)
		total.AttestationIdealReward = total.AttestationIdealReward.Add(queryEntry.AttestationIdealReward)
		total.AttestationsExecuted += queryEntry.AttestationsExecuted
		total.AttestationsScheduled += queryEntry.AttestationsScheduled
		total.BlocksProposed += queryEntry.BlocksProposed
		total.BlocksScheduled += queryEntry.BlocksScheduled
		total.SyncExecuted += queryEntry.SyncExecuted
		total.SyncScheduled += queryEntry.SyncScheduled

		// If the search permits it add the entry to the result
		if search != "" {
			prefixSearch := strings.ToLower(search)
			for _, validatorIndex := range uiValidatorIndices {
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
			Validators:               total.Validators,
			AverageNetworkEfficiency: averageNetworkEfficiency,
			Reward:                   total.Reward,
		}

		// Attestations
		totalEntry.Attestations.Success = total.AttestationsExecuted
		totalEntry.Attestations.Failed = total.AttestationsScheduled - total.AttestationsExecuted

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
		totalEntry.Efficiency = d.calculateTotalEfficiency(totalAttestationEfficiency, totalProposerEfficiency, totalSyncEfficiency)

		result = append([]t.VDBSummaryTableRow{totalEntry}, result...)
	}

	paging.TotalCount = uint64(len(result))

	return result, &paging, nil
}

func (d *DataAccessService) GetValidatorDashboardGroupSummary(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBGroupSummaryData, error) {
	// TODO: implement data retrieval for the following new field
	// Fetch validator list for user dashboard from the dashboard table when querying the past sync committees as the rolling table might miss exited validators
	// TotalMissedRewards

	var err error
	ret := &t.VDBGroupSummaryData{}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	// retrieve the members of the current, previous & upcoming sync committees
	currentSyncCommitteeMembers := map[uint32]bool{}
	upcomingSyncCommitteeMembers := map[uint32]bool{}

	type syncCommitteeQueryResultType struct {
		Period         int    `db:"period"`
		ValidatorIndex uint32 `db:"validatorindex"`
	}

	latestEpoch := cache.LatestEpoch.Get()
	currentSyncPeriod := int(utils.SyncPeriodOfEpoch(latestEpoch))
	upcomingSyncPeriod := currentSyncPeriod + 1

	var syncCommitteeQueryResult []syncCommitteeQueryResultType
	err = d.readerDb.Select(&syncCommitteeQueryResult, `SELECT period, validatorindex FROM sync_committees WHERE period >= $1 AND period <= $2`, currentSyncPeriod, upcomingSyncPeriod)
	if err != nil {
		return nil, fmt.Errorf("error retrieving sync committee current and next period data: %v", err)
	}

	for _, row := range syncCommitteeQueryResult {
		switch row.Period {
		case currentSyncPeriod:
			currentSyncCommitteeMembers[row.ValidatorIndex] = true
		case upcomingSyncPeriod:
			upcomingSyncCommitteeMembers[row.ValidatorIndex] = true
		}
	}

	table := ""
	slashedByCountTable := ""
	days := 0

	switch period {
	case enums.TimePeriods.Last24h:
		table = "validator_dashboard_data_rolling_daily"
		slashedByCountTable = "validator_dashboard_data_rolling_daily_slashedby_count"
		days = 1
	case enums.TimePeriods.Last7d:
		table = "validator_dashboard_data_rolling_weekly"
		slashedByCountTable = "validator_dashboard_data_rolling_weekly_slashedby_count"
		days = 7
	case enums.TimePeriods.Last30d:
		table = "validator_dashboard_data_rolling_monthly"
		slashedByCountTable = "validator_dashboard_data_rolling_monthly_slashedby_count"
		days = 30
	case enums.TimePeriods.AllTime:
		table = "validator_dashboard_data_rolling_total"
		slashedByCountTable = "validator_dashboard_data_rolling_total_slashedby_count"
		days = -1
	default:
		return nil, fmt.Errorf("not-implemented time period: %v", period)
	}

	query := `select
			users_val_dashboards_validators.validator_index,
			epoch_start,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			%[1]s.slashed_by IS NOT NULL AS slashed_in_period,
			COALESCE(%[2]s.slashed_amount, 0) AS slashed_amount,
			COALESCE(blocks_expected, 0) as blocks_expected,
			COALESCE(inclusion_delay_sum, 0) as inclusion_delay_sum,
			COALESCE(sync_committees_expected, 0) as sync_committees_expected
		from users_val_dashboards_validators
		inner join %[1]s on %[1]s.validator_index = users_val_dashboards_validators.validator_index
		left join %[2]s on %[2]s.slashed_by = users_val_dashboards_validators.validator_index
		where (dashboard_id = $1 and (group_id = $2 OR $2 = -1))
		`

	if dashboardId.Validators != nil {
		query = `select
			validator_index,
			epoch_start,
			COALESCE(attestations_reward, 0) as attestations_reward,
			COALESCE(attestations_ideal_reward, 0) as attestations_ideal_reward,
			COALESCE(attestations_scheduled, 0) as attestations_scheduled,
			COALESCE(attestation_head_executed, 0) as attestation_head_executed,
			COALESCE(attestation_source_executed, 0) as attestation_source_executed,
			COALESCE(attestation_target_executed, 0) as attestation_target_executed,
			COALESCE(blocks_scheduled, 0) as blocks_scheduled,
			COALESCE(blocks_proposed, 0) as blocks_proposed,
			COALESCE(sync_scheduled, 0) as sync_scheduled,
			COALESCE(sync_executed, 0) as sync_executed,
			%[1]s.slashed_by IS NOT NULL AS slashed_in_period,
			COALESCE(%[2]s.slashed_amount, 0) AS slashed_amount,
			COALESCE(blocks_expected, 0) as blocks_expected,
			COALESCE(inclusion_delay_sum, 0) as inclusion_delay_sum,
			COALESCE(sync_committees_expected, 0) as sync_committees_expected
		from %[1]s
		left join %[2]s on %[2]s.slashed_by = %[1]s.validator_index
		where %[1]s.validator_index = ANY($1)
	`
	}

	validators := make([]t.VDBValidator, 0)
	if dashboardId.Validators != nil {
		validators = dashboardId.Validators
	}

	type queryResult struct {
		ValidatorIndex         uint32 `db:"validator_index"`
		EpochStart             int    `db:"epoch_start"`
		AttestationReward      int64  `db:"attestations_reward"`
		AttestationIdealReward int64  `db:"attestations_ideal_reward"`

		AttestationsScheduled     int64 `db:"attestations_scheduled"`
		AttestationHeadExecuted   int64 `db:"attestation_head_executed"`
		AttestationSourceExecuted int64 `db:"attestation_source_executed"`
		AttestationTargetExecuted int64 `db:"attestation_target_executed"`

		BlocksScheduled uint32 `db:"blocks_scheduled"`
		BlocksProposed  uint32 `db:"blocks_proposed"`

		SyncScheduled uint32 `db:"sync_scheduled"`
		SyncExecuted  uint32 `db:"sync_executed"`

		SlashedInPeriod bool   `db:"slashed_in_period"`
		SlashedAmount   uint32 `db:"slashed_amount"`

		BlockChance            float64 `db:"blocks_expected"`
		SyncCommitteesExpected float64 `db:"sync_committees_expected"`

		InclusionDelaySum int64 `db:"inclusion_delay_sum"`
	}

	var rows []*queryResult

	if len(validators) > 0 {
		err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table, slashedByCountTable), validators)
	} else {
		err = d.alloyReader.Select(&rows, fmt.Sprintf(query, table, slashedByCountTable), dashboardId.Id, groupId)
	}

	if err != nil {
		return nil, fmt.Errorf("error retrieving validator dashboard group summary data: %v", err)
	}

	totalAttestationRewards := int64(0)
	totalIdealAttestationRewards := int64(0)
	totalBlockChance := float64(0)
	totalInclusionDelaySum := int64(0)
	totalInclusionDelayDivisor := int64(0)
	totalSyncExpected := float64(0)
	totalProposals := uint32(0)

	validatorArr := make([]t.VDBValidator, 0)
	startEpoch := math.MaxUint32
	for _, row := range rows {
		if row.EpochStart < startEpoch { // set the start epoch for querying the EL APR
			startEpoch = row.EpochStart
		}
		validatorArr = append(validatorArr, t.VDBValidator(row.ValidatorIndex))
		totalAttestationRewards += row.AttestationReward
		totalIdealAttestationRewards += row.AttestationIdealReward

		ret.AttestationsHead.Success += uint64(row.AttestationHeadExecuted)
		ret.AttestationsHead.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationHeadExecuted)

		ret.AttestationsSource.Success += uint64(row.AttestationSourceExecuted)
		ret.AttestationsSource.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationSourceExecuted)

		ret.AttestationsTarget.Success += uint64(row.AttestationTargetExecuted)
		ret.AttestationsTarget.Failed += uint64(row.AttestationsScheduled) - uint64(row.AttestationTargetExecuted)

		if row.ValidatorIndex == 0 && row.BlocksProposed > 0 && row.BlocksProposed != row.BlocksScheduled {
			row.BlocksProposed-- // subtract the genesis block from validator 0 (TODO: remove when fixed in the dashoard data exporter)
		}

		totalProposals += row.BlocksScheduled
		if row.BlocksScheduled > 0 {
			if ret.ProposalValidators == nil {
				ret.ProposalValidators = make([]t.VDBValidator, 0, 10)
			}
			ret.ProposalValidators = append(ret.ProposalValidators, t.VDBValidator(row.ValidatorIndex))
		}

		ret.SyncCommittee.StatusCount.Success += uint64(row.SyncExecuted)
		ret.SyncCommittee.StatusCount.Failed += uint64(row.SyncScheduled) - uint64(row.SyncExecuted)

		if row.SyncScheduled > 0 {
			if ret.SyncCommittee.Validators == nil {
				ret.SyncCommittee.Validators = make([]t.VDBValidator, 0, 10)
			}
			ret.SyncCommittee.Validators = append(ret.SyncCommittee.Validators, t.VDBValidator(row.ValidatorIndex))

			if currentSyncCommitteeMembers[row.ValidatorIndex] {
				ret.SyncCommitteeCount.CurrentValidators++
			}
			if upcomingSyncCommitteeMembers[row.ValidatorIndex] {
				ret.SyncCommitteeCount.UpcomingValidators++
			}
		}

		if row.SlashedInPeriod {
			ret.Slashings.StatusCount.Failed++
			ret.Slashings.Validators = append(ret.Slashings.Validators, t.VDBValidator(row.ValidatorIndex))
		}
		if row.SlashedAmount > 0 {
			ret.Slashings.StatusCount.Success += uint64(row.SlashedAmount)
			ret.Slashings.Validators = append(ret.Slashings.Validators, t.VDBValidator(row.ValidatorIndex))
		}

		totalBlockChance += row.BlockChance
		totalInclusionDelaySum += row.InclusionDelaySum
		totalSyncExpected += row.SyncCommitteesExpected

		if row.InclusionDelaySum > 0 {
			totalInclusionDelayDivisor += row.AttestationsScheduled
		}
	}

	_, ret.Apr.El, _, ret.Apr.Cl, err = d.internal_getElClAPR(validatorArr, days)
	if err != nil {
		return nil, err
	}

	if len(validators) > 0 {
		validatorArr = validators
	}
	err = d.readerDb.Get(&ret.SyncCommitteeCount.PastPeriods, `SELECT COUNT(*) FROM sync_committees WHERE period < $1 AND validatorindex = ANY($2)`, currentSyncPeriod, validatorArr)
	if err != nil {
		return nil, fmt.Errorf("error retrieving past sync committee count: %v", err)
	}

	ret.AttestationEfficiency = float64(totalAttestationRewards) / float64(totalIdealAttestationRewards) * 100
	if ret.AttestationEfficiency < 0 || math.IsNaN(ret.AttestationEfficiency) {
		ret.AttestationEfficiency = 0
	}

	luckDays := float64(days)
	if days == -1 {
		luckDays = time.Since(time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)).Hours() / 24
		if luckDays == 0 {
			luckDays = 1
		}
	}

	if totalBlockChance > 0 {
		ret.Luck.Proposal.Percent = (float64(totalProposals)) / totalBlockChance * 100

		// calculate the average time it takes for the set of validators to propose a single block on average
		ret.Luck.Proposal.Average = time.Duration((luckDays / totalBlockChance) * float64(utils.Day))
	} else {
		ret.Luck.Proposal.Percent = 0
	}

	if totalSyncExpected == 0 {
		ret.Luck.Sync.Percent = 0
	} else {
		totalSyncSlotDuties := float64(ret.SyncCommittee.StatusCount.Failed) + float64(ret.SyncCommittee.StatusCount.Success)
		slotDutiesPerSyncCommittee := float64(utils.SlotsPerSyncCommittee())
		syncCommittees := math.Ceil(totalSyncSlotDuties / slotDutiesPerSyncCommittee) // gets the number of sync committees
		ret.Luck.Sync.Percent = syncCommittees / totalSyncExpected * 100

		// calculate the average time it takes for the set of validators to be elected into a sync committee on average
		ret.Luck.Sync.Average = time.Duration((luckDays / totalSyncExpected) * float64(utils.Day))
	}

	if totalInclusionDelayDivisor > 0 {
		ret.AttestationAvgInclDist = 1.0 + float64(totalInclusionDelaySum)/float64(totalInclusionDelayDivisor)
	} else {
		ret.AttestationAvgInclDist = 0
	}

	return ret, nil
}

func (d *DataAccessService) internal_getElClAPR(validators []t.VDBValidator, days int) (elIncome decimal.Decimal, elAPR float64, clIncome decimal.Decimal, clAPR float64, err error) {
	var reward sql.NullInt64
	table := ""

	switch days {
	case 1:
		table = "validator_dashboard_data_rolling_daily"
	case 7:
		table = "validator_dashboard_data_rolling_weekly"
	case 30:
		table = "validator_dashboard_data_rolling_monthly"
	case -1:
		table = "validator_dashboard_data_rolling_90d"
	default:
		return decimal.Zero, 0, decimal.Zero, 0, fmt.Errorf("invalid days value: %v", days)
	}

	query := `select (SUM(COALESCE(balance_end,0)) + SUM(COALESCE(withdrawals_amount,0)) - SUM(COALESCE(deposits_amount,0)) - SUM(COALESCE(balance_start,0))) reward FROM %s WHERE validator_index = ANY($1)`

	err = db.AlloyReader.Get(&reward, fmt.Sprintf(query, table), validators)
	if err != nil || !reward.Valid {
		return decimal.Zero, 0, decimal.Zero, 0, err
	}

	aprDivisor := days
	if days == -1 { // for all time APR
		aprDivisor = 90
	}
	clAPR = ((float64(reward.Int64) / float64(aprDivisor)) / (float64(32e9) * float64(len(validators)))) * 365.0 * 100.0
	if math.IsNaN(clAPR) {
		clAPR = 0
	}
	if days == -1 {
		err = db.AlloyReader.Get(&reward, fmt.Sprintf(query, "validator_dashboard_data_rolling_total"), validators)
		if err != nil || !reward.Valid {
			return decimal.Zero, 0, decimal.Zero, 0, err
		}
	}
	clIncome = decimal.NewFromInt(reward.Int64).Mul(decimal.NewFromInt(1e9))

	query = `
	SELECT 
		COALESCE(SUM(COALESCE(rb.value / 1e18, fee_recipient_reward)), 0)
	FROM blocks 
	LEFT JOIN execution_payloads ON blocks.exec_block_hash = execution_payloads.block_hash
	LEFT JOIN relays_blocks rb ON blocks.exec_block_hash = rb.exec_block_hash
	WHERE proposer = ANY($1) AND status = '1' AND slot >= (SELECT MIN(epoch_start) * $2 FROM %s WHERE validator_index = ANY($1));`
	err = db.AlloyReader.Get(&elIncome, fmt.Sprintf(query, table), validators, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		return decimal.Zero, 0, decimal.Zero, 0, err
	}
	elIncomeFloat, _ := elIncome.Float64()
	elAPR = ((elIncomeFloat / float64(aprDivisor)) / (float64(32e18) * float64(len(validators)))) * 365.0 * 100.0

	if days == -1 {
		err = db.AlloyReader.Get(&elIncome, fmt.Sprintf(query, "validator_dashboard_data_rolling_total"), validators, utils.Config.Chain.ClConfig.SlotsPerEpoch)
		if err != nil {
			return decimal.Zero, 0, decimal.Zero, 0, err
		}
	}
	elIncome = elIncome.Mul(decimal.NewFromInt(1e18))

	return elIncome, elAPR, clIncome, clAPR, nil
}

// for summary charts: series id is group id, no stack

func (d *DataAccessService) GetValidatorDashboardSummaryChart(dashboardId t.VDBId) (*t.ChartData[int, float64], error) {
	ret := &t.ChartData[int, float64]{}

	type queryResult struct {
		StartEpoch            uint64          `db:"epoch_start"`
		GroupId               uint64          `db:"group_id"`
		AttestationEfficiency sql.NullFloat64 `db:"attestation_efficiency"`
		ProposerEfficiency    sql.NullFloat64 `db:"proposer_efficiency"`
		SyncEfficiency        sql.NullFloat64 `db:"sync_efficiency"`
	}

	var queryResults []queryResult

	cutOffDate := time.Date(2023, 9, 27, 23, 59, 59, 0, time.UTC).Add(time.Hour*24*30).AddDate(0, 0, -30)

	if dashboardId.Validators != nil {
		query := `select epoch_start, 0 AS group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
			select
				epoch_start,
				COALESCE(SUM(attestations_reward), 0)::decimal / NULLIF(SUM(attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				COALESCE(SUM(blocks_proposed), 0)::decimal / NULLIF(SUM(blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				COALESCE(SUM(sync_executed), 0)::decimal / NULLIF(SUM(sync_scheduled)::decimal, 0) AS sync_efficiency
			from  validator_dashboard_data_daily
			WHERE day > $1 AND validator_index = ANY($2)
			group by 1
		) as a ORDER BY epoch_start;`
		err := d.alloyReader.Select(&queryResults, query, cutOffDate, dashboardId.Validators)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %v", err)
		}
	} else {
		queryParams := []interface{}{cutOffDate, dashboardId.Id}

		groupIdQuery := "v.group_id,"
		groupByQuery := "GROUP BY 1, 2"
		orderQuery := "ORDER BY epoch_start, group_id"
		if dashboardId.AggregateGroups {
			queryParams = append(queryParams, t.DefaultGroupId)
			groupIdQuery = "$3::smallint AS group_id,"
			groupByQuery = "GROUP BY 1"
			orderQuery = "ORDER BY epoch_start"
		}
		query := fmt.Sprintf(`SELECT epoch_start, group_id, attestation_efficiency, proposer_efficiency, sync_efficiency FROM (
			SELECT
				d.epoch_start, 
				%s
				COALESCE(SUM(d.attestations_reward), 0)::decimal / NULLIF(SUM(d.attestations_ideal_reward)::decimal, 0) AS attestation_efficiency,
				COALESCE(SUM(d.blocks_proposed), 0)::decimal / NULLIF(SUM(d.blocks_scheduled)::decimal, 0) AS proposer_efficiency,
				COALESCE(SUM(d.sync_executed), 0)::decimal / NULLIF(SUM(d.sync_scheduled)::decimal, 0) AS sync_efficiency
			FROM users_val_dashboards_validators v
			INNER JOIN validator_dashboard_data_daily d ON d.validator_index = v.validator_index
			WHERE day > $1 AND dashboard_id = $2
			%s
		) as a %s;`, groupIdQuery, groupByQuery, orderQuery)
		err := d.alloyReader.Select(&queryResults, query, queryParams...)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table validator_dashboard_data_daily: %v", err)
		}
	}

	// convert the returned data to the expected return type (not pretty)
	epochsMap := make(map[uint64]bool)
	groups := make(map[uint64]bool)
	data := make(map[uint64]map[uint64]float64)
	for _, row := range queryResults {
		epochsMap[row.StartEpoch] = true
		groups[row.GroupId] = true

		if data[row.StartEpoch] == nil {
			data[row.StartEpoch] = make(map[uint64]float64)
		}
		data[row.StartEpoch][row.GroupId] = d.calculateTotalEfficiency(row.AttestationEfficiency, row.ProposerEfficiency, row.SyncEfficiency)
	}

	epochsArray := make([]uint64, 0, len(epochsMap))
	for epoch := range epochsMap {
		epochsArray = append(epochsArray, epoch)
	}
	sort.Slice(epochsArray, func(i, j int) bool {
		return epochsArray[i] < epochsArray[j]
	})

	groupsArray := make([]uint64, 0, len(groups))
	for group := range groups {
		groupsArray = append(groupsArray, group)
	}
	sort.Slice(groupsArray, func(i, j int) bool {
		return groupsArray[i] < groupsArray[j]
	})

	ret.Categories = epochsArray
	ret.Series = make([]t.ChartSeries[int, float64], 0, len(groupsArray))

	seriesMap := make(map[uint64]*t.ChartSeries[int, float64])
	for group := range groups {
		series := t.ChartSeries[int, float64]{
			Id:   int(group),
			Data: make([]float64, 0, len(epochsMap)),
		}
		seriesMap[group] = &series
	}

	for _, epoch := range epochsArray {
		for _, group := range groupsArray {
			seriesMap[group].Data = append(seriesMap[group].Data, data[epoch][group])
		}
	}

	for _, series := range seriesMap {
		ret.Series = append(ret.Series, *series)
	}

	sort.Slice(ret.Series, func(i, j int) bool {
		return ret.Series[i].Id < ret.Series[j].Id
	})

	return ret, nil
}

func (d *DataAccessService) GetValidatorDashboardSummaryValidators(dashboardId t.VDBId, groupId int64) (*t.VDBGeneralSummaryValidators, error) {
	// TODO @DATA-ACCESS
	return d.dummy.GetValidatorDashboardSummaryValidators(dashboardId, groupId)
}
func (d *DataAccessService) GetValidatorDashboardSyncSummaryValidators(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSyncSummaryValidators, error) {
	// TODO @DATA-ACCESS
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	return d.dummy.GetValidatorDashboardSyncSummaryValidators(dashboardId, groupId, period)
}
func (d *DataAccessService) GetValidatorDashboardSlashingsSummaryValidators(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSlashingsSummaryValidators, error) {
	// TODO @DATA-ACCESS
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	return d.dummy.GetValidatorDashboardSlashingsSummaryValidators(dashboardId, groupId, period)
}
func (d *DataAccessService) GetValidatorDashboardProposalSummaryValidators(dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBProposalSummaryValidators, error) {
	// TODO @DATA-ACCESS
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	return d.dummy.GetValidatorDashboardProposalSummaryValidators(dashboardId, groupId, period)
}
