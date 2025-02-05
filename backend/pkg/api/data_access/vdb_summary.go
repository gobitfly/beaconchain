package dataaccess

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/juliangruber/go-intersect"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// for summary charts: series id is group id, no stack

func (d *DataAccessService) GetValidatorDashboardSummaryChart(ctx context.Context, dashboardId t.VDBId, groupIds []int64, efficiency enums.VDBSummaryChartEfficiencyType, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.ChartData[int, float64], error) {
	ret := &t.ChartData[int, float64]{}

	if len(groupIds) == 0 { // short circuit if no groups are selected
		return ret, nil
	}

	var queryResults []*t.VDBValidatorSummaryChartRow

	containsGroups := false
	requestedGroupsMap := make(map[int64]bool)
	for _, groupId := range groupIds {
		requestedGroupsMap[groupId] = true
		if !containsGroups && groupId >= 0 {
			containsGroups = true
		}
	}

	// need default or all groups for anon dashboards and shared dashboards without group sharing
	// TODO could move this to API layer & generalize for all methods
	if (dashboardId.Validators != nil && !requestedGroupsMap[t.AllGroups] && !requestedGroupsMap[t.DefaultGroupId]) ||
		(dashboardId.AggregateGroups && !requestedGroupsMap[t.AllGroups] && !requestedGroupsMap[t.DefaultGroupId]) {
		return ret, nil
	}

	dataTable, dateColumn, err := d.getTableAndDateColumn(aggregation)
	if err != nil {
		return nil, err
	}

	totalLineRequested := requestedGroupsMap[t.AllGroups] || dashboardId.AggregateGroups
	averageNetworkLineRequested := requestedGroupsMap[t.NetworkAverage]

	if dashboardId.Validators != nil {
		query := fmt.Sprintf(`
			SELECT
				%[2]s as ts,
				0 AS group_id, 
				COALESCE(SUM(d.attestations_reward), 0) AS attestations_reward,
				COALESCE(SUM(d.attestations_ideal_reward), 0) AS attestations_ideal_reward,
				COALESCE(SUM(d.blocks_proposed), 0) AS blocks_proposed,
				COALESCE(SUM(d.blocks_scheduled), 0) AS blocks_scheduled,
				COALESCE(SUM(d.sync_executed), 0) AS sync_executed,
				COALESCE(SUM(d.sync_scheduled), 0) AS sync_scheduled
			FROM %[1]s d
			WHERE %[2]s >= fromUnixTimestamp($1) AND %[2]s <= fromUnixTimestamp($2) AND validator_index IN ($3)
			GROUP BY %[2]s;
		`, dataTable, dateColumn)
		err := d.clickhouseReader.SelectContext(ctx, &queryResults, query, afterTs, beforeTs, dashboardId.Validators)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table %s: %w", dataTable, err)
		}
	} else {
		query := fmt.Sprintf(`
		WITH validators AS (
			SELECT validator_index as validator_index, group_id FROM users_val_dashboards_validators WHERE dashboard_id = $3 AND (group_id IN ($4) OR $5)
		)		
		SELECT
			%[2]s as ts,
			v.group_id,
			COALESCE(SUM(d.attestations_reward), 0) AS attestations_reward,
			COALESCE(SUM(d.attestations_ideal_reward), 0) AS attestations_ideal_reward,
			COALESCE(SUM(d.blocks_proposed), 0) AS blocks_proposed,
			COALESCE(SUM(d.blocks_scheduled), 0) AS blocks_scheduled,
			COALESCE(SUM(d.sync_executed), 0) AS sync_executed,
			COALESCE(SUM(d.sync_scheduled), 0) AS sync_scheduled
		FROM %[1]s d
		INNER JOIN validators v ON d.validator_index = v.validator_index
		WHERE %[2]s >= fromUnixTimestamp($1) AND %[2]s <= fromUnixTimestamp($2) AND validator_index in (select validator_index from validators)
		GROUP BY 1, 2;`, dataTable, dateColumn)

		err := d.clickhouseReader.SelectContext(ctx, &queryResults, query, afterTs, beforeTs, dashboardId.Id, groupIds, totalLineRequested)
		if err != nil {
			return nil, fmt.Errorf("error retrieving data from table %s: %w", dataTable, err)
		}
	}

	// convert the returned data to the expected return type (not pretty)
	tsMap := make(map[time.Time]bool)
	data := make(map[time.Time]map[int64]float64)
	groupMap := make(map[int64]bool)

	totalEfficiencyMap := make(map[time.Time]*t.VDBValidatorSummaryChartRow)
	for _, row := range queryResults {
		tsMap[row.Timestamp] = true

		if data[row.Timestamp] == nil {
			data[row.Timestamp] = make(map[int64]float64)
		}

		if !dashboardId.AggregateGroups && requestedGroupsMap[row.GroupId] {
			groupEfficiency, err := d.calculateChartEfficiency(efficiency, row)
			if err != nil {
				return nil, err
			}

			data[row.Timestamp][row.GroupId] = groupEfficiency
			groupMap[row.GroupId] = true
		}

		if totalLineRequested {
			if totalEfficiencyMap[row.Timestamp] == nil {
				totalEfficiencyMap[row.Timestamp] = &t.VDBValidatorSummaryChartRow{
					Timestamp: row.Timestamp,
				}
			}
			totalEfficiencyMap[row.Timestamp].AttestationReward += row.AttestationReward
			totalEfficiencyMap[row.Timestamp].AttestationIdealReward += row.AttestationIdealReward
			totalEfficiencyMap[row.Timestamp].BlocksProposed += row.BlocksProposed
			totalEfficiencyMap[row.Timestamp].BlocksScheduled += row.BlocksScheduled
			totalEfficiencyMap[row.Timestamp].SyncExecuted += row.SyncExecuted
			totalEfficiencyMap[row.Timestamp].SyncScheduled += row.SyncScheduled
		}
	}

	if averageNetworkLineRequested {
		// Get the average network efficiency
		efficiency, err := d.services.GetCurrentEfficiencyInfo()
		if err != nil {
			return nil, err
		}
		averageNetworkEfficiency := utils.CalculateTotalEfficiency(
			efficiency.AttestationEfficiency[enums.Last24h], efficiency.ProposalEfficiency[enums.Last24h], efficiency.SyncEfficiency[enums.Last24h])

		for ts := range tsMap {
			data[ts][int64(t.NetworkAverage)] = averageNetworkEfficiency
		}
		groupMap[t.NetworkAverage] = true
	}

	if totalLineRequested {
		totalLineGroupId := int64(t.AllGroups)
		if dashboardId.AggregateGroups {
			totalLineGroupId = t.DefaultGroupId
		}
		for _, row := range totalEfficiencyMap {
			totalEfficiency, err := d.calculateChartEfficiency(efficiency, row)
			if err != nil {
				return nil, err
			}

			data[row.Timestamp][totalLineGroupId] = totalEfficiency
		}
		groupMap[totalLineGroupId] = true
	}

	tsArray := make([]time.Time, 0, len(tsMap))
	for ts := range tsMap {
		tsArray = append(tsArray, ts)
	}
	sort.Slice(tsArray, func(i, j int) bool {
		return tsArray[i].Before(tsArray[j])
	})

	groupsArray := slices.Collect(maps.Keys(groupMap))
	slices.Sort(groupsArray)

	ret.Categories = make([]uint64, 0, len(tsArray))
	for _, ts := range tsArray {
		ret.Categories = append(ret.Categories, uint64(ts.Unix()))
	}
	ret.Series = make([]t.ChartSeries[int, float64], 0, len(groupsArray))

	seriesMap := make(map[int64]*t.ChartSeries[int, float64])
	for _, group := range groupsArray {
		series := t.ChartSeries[int, float64]{
			Id:   int(group),
			Data: make([]float64, 0, len(tsMap)),
		}
		seriesMap[group] = &series
	}

	for _, ts := range tsArray {
		for _, group := range groupsArray {
			seriesMap[group].Data = append(seriesMap[group].Data, data[ts][group])
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

func (d *DataAccessService) GetLatestExportedChartTs(ctx context.Context, aggregation enums.ChartAggregation) (uint64, error) {
	table, dateColumn, err := d.getViewAndDateColumn(aggregation)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`SELECT max(%s) FROM %s`, dateColumn, table)
	var ts time.Time
	err = d.clickhouseReader.GetContext(ctx, &ts, query)
	if err != nil {
		return 0, fmt.Errorf("error retrieving latest exported chart timestamp: %w", err)
	}

	return uint64(ts.Unix()), nil
}

func (d *DataAccessService) GetValidatorDashboardSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64) (*t.VDBGeneralSummaryValidators, error) {
	result := &t.VDBGeneralSummaryValidators{}

	// Get the validator indices
	var groupIds []uint64
	if !dashboardId.AggregateGroups && groupId != t.AllGroups {
		groupIds = append(groupIds, uint64(groupId))
	}

	validatorIndices, err := d.getDashboardValidators(ctx, dashboardId, groupIds)
	if err != nil {
		return nil, err
	}

	latestEpoch := cache.LatestFinalizedEpoch.Get()
	latestStats := cache.LatestStats.Get()
	var activationChurnRate uint64

	if latestStats.ValidatorActivationChurnLimit == nil {
		activationChurnRate = 4
		log.Warnf("Activation Churn rate not set in config using 4 as default")
	} else {
		activationChurnRate = *latestStats.ValidatorActivationChurnLimit
	}

	stats := cache.LatestStats.Get()
	if stats == nil || stats.LatestValidatorWithdrawalIndex == nil {
		return nil, errors.New("stats not available")
	}

	// Get the current validator state
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}

	// Fill the data
	for _, validatorIndex := range validatorIndices {
		metadata := validatorMapping.ValidatorMetadata[validatorIndex]

		switch constypes.ValidatorDbStatus(metadata.Status) {
		case constypes.DbDeposited:
			result.Deposited = append(result.Deposited, validatorIndex)
		case constypes.DbPending:
			validatorInfo := t.IndexTimestamp{
				Index: validatorIndex,
			}
			if metadata.ActivationEpoch.Valid {
				validatorInfo.Timestamp = uint64(utils.EpochToTime(uint64(metadata.ActivationEpoch.Int64)).Unix())
			} else if metadata.Queues.ActivationIndex.Valid {
				queuePosition := uint64(metadata.Queues.ActivationIndex.Int64)
				epochsToWait := (queuePosition - 1) / activationChurnRate
				// calculate dequeue epoch
				estimatedActivationEpoch := latestEpoch + epochsToWait + 1
				// add activation offset
				estimatedActivationEpoch += utils.Config.Chain.ClConfig.MaxSeedLookahead + 1
				validatorInfo.Timestamp = uint64(utils.EpochToTime(estimatedActivationEpoch).Unix())
			}
			result.Pending = append(result.Pending, validatorInfo)
		case constypes.DbActiveOnline:
			result.Online = append(result.Online, validatorIndex)
		case constypes.DbActiveOffline:
			result.Offline = append(result.Offline, validatorIndex)
		case constypes.DbSlashingOnline, constypes.DbSlashingOffline:
			result.Slashing = append(result.Slashing, validatorIndex)
			if constypes.ValidatorDbStatus(metadata.Status) == constypes.DbSlashingOffline {
				result.Offline = append(result.Offline, validatorIndex)
			} else {
				result.Online = append(result.Online, validatorIndex)
			}
		case constypes.DbExitingOnline, constypes.DbExitingOffline:
			result.Exiting = append(result.Exiting, t.IndexTimestamp{
				Index:     validatorIndex,
				Timestamp: uint64(utils.EpochToTime(uint64(metadata.ExitEpoch.Int64)).Unix()),
			})
			if constypes.ValidatorDbStatus(metadata.Status) == constypes.DbExitingOffline {
				result.Offline = append(result.Offline, validatorIndex)
			} else {
				result.Online = append(result.Online, validatorIndex)
			}
		case constypes.DbExited, constypes.DbSlashed:
			if constypes.ValidatorDbStatus(metadata.Status) == constypes.DbSlashed {
				result.Slashed = append(result.Slashed, validatorIndex)
			} else {
				result.Exited = append(result.Exited, validatorIndex)
			}

			if metadata.WithdrawableEpoch.Valid && metadata.WithdrawableEpoch.Int64 <= int64(latestEpoch) {
				if metadata.Balance != 0 {
					validatorInfo := t.IndexTimestamp{
						Index: validatorIndex,
					}

					if utils.IsValidWithdrawalCredentialsAddress(fmt.Sprintf("%x", metadata.WithdrawalCredentials)) {
						distance, err := d.getWithdrawableCountFromCursor(validatorIndex, *stats.LatestValidatorWithdrawalIndex)
						if err != nil {
							return nil, err
						}

						timeToWithdrawal := d.getTimeToNextWithdrawal(distance)
						validatorInfo.Timestamp = uint64(timeToWithdrawal.Unix())
					}

					result.Withdrawing = append(result.Withdrawing, validatorInfo)
				} else {
					result.Withdrawn = append(result.Withdrawn, validatorIndex)
				}
			}
		}
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardSyncSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSyncSummaryValidators, error) {
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	result := &t.VDBSyncSummaryValidators{}
	var resultMutex = &sync.RWMutex{}
	wg := errgroup.Group{}

	// Get the validator indices
	var groupIds []uint64
	if !dashboardId.AggregateGroups && groupId != t.AllGroups {
		groupIds = append(groupIds, uint64(groupId))
	}
	validatorIndices, err := d.getDashboardValidators(ctx, dashboardId, groupIds)
	if err != nil {
		return nil, err
	}

	// Get the current and next sync committee validators
	wg.Go(func() error {
		currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, err := d.getCurrentAndUpcomingSyncCommittees(ctx, cache.LatestEpoch.Get())
		if err != nil {
			return err
		}

		resultMutex.Lock()
		for _, validatorIndex := range validatorIndices {
			if currentSyncCommitteeValidators[validatorIndex] {
				result.Current = append(result.Current, validatorIndex)
			}
			if upcomingSyncCommitteeValidators[validatorIndex] {
				result.Upcoming = append(result.Upcoming, validatorIndex)
			}
		}
		resultMutex.Unlock()

		return nil
	})

	// Get the past sync committee validators
	wg.Go(func() error {
		epochStart, err := d.getEpochStart(ctx, period)
		if err != nil {
			return err
		}

		validatorCountMap, err := d.getPastSyncCommittees(ctx, validatorIndices, epochStart, cache.LatestEpoch.Get())
		if err != nil {
			return err
		}

		resultMutex.Lock()
		for validatorIndex, count := range validatorCountMap {
			result.Past = append(result.Past, t.VDBValidatorSyncPast{
				Index: validatorIndex,
				Count: count,
			})
		}
		resultMutex.Unlock()

		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardSlashingsSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBSlashingsSummaryValidators, error) {
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	result := &t.VDBSlashingsSummaryValidators{}

	// Get the table names based on the period
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	var queryResult []struct {
		EpochStart     uint64 `db:"epoch_start"`
		EpochEnd       uint64 `db:"epoch_end"`
		ValidatorIndex uint64 `db:"validator_index"`
		Slashed        bool   `db:"slashed"`
		SlashedAmount  uint32 `db:"slashed_amount"`
	}

	// Build the query
	ds := goqu.Dialect("postgres").
		From(goqu.L(fmt.Sprintf("%s AS r FINAL", clickhouseTable))).
		With("validators", goqu.L("(SELECT group_id, validator_index FROM users_val_dashboards_validators WHERE dashboard_id = ?)", dashboardId.Id)).
		Select(
			goqu.L("r.epoch_start"),
			goqu.L("r.epoch_end"),
			goqu.L("r.validator_index"),
			goqu.L("r.slashed"),
			goqu.L("COALESCE(r.blocks_slashing_count, 0) AS slashed_amount")).
		Where(goqu.L("(r.slashed OR r.blocks_slashing_count > 0)"))

	// handle the case when we have a list of validators
	if len(dashboardId.Validators) > 0 {
		ds = ds.
			Where(goqu.L("r.validator_index IN ?", dashboardId.Validators))
	} else {
		ds = ds.
			InnerJoin(goqu.L("validators v"), goqu.On(goqu.L("r.validator_index = v.validator_index"))).
			Where(goqu.L("r.validator_index IN (SELECT validator_index FROM validators)"))

		if groupId != t.AllGroups {
			ds = ds.Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	err = d.clickhouseReader.SelectContext(ctx, &queryResult, query, args...)
	if err != nil {
		log.Error(err, "error while getting validator dashboard slashed validators list", 0)
		return nil, err
	}

	// Process the data and get the slashing validators
	var slashingValidators []uint64
	var slashedValidators []uint64
	for _, queryEntry := range queryResult {
		if queryEntry.SlashedAmount > 0 {
			slashingValidators = append(slashingValidators, queryEntry.ValidatorIndex)
		}

		if queryEntry.Slashed {
			slashedValidators = append(slashedValidators, queryEntry.ValidatorIndex)
		}
	}

	if len(slashingValidators) == 0 && len(slashedValidators) == 0 {
		// We don't have any slashing or slashed validators so we can return early
		return result, nil
	}

	slashingValidatorsMap := utils.SliceToMap(slashingValidators)
	slashedValidatorsMap := utils.SliceToMap(slashedValidators)

	// If we have slashing validators then get the validators that got slashed
	proposalSlashings := make(map[uint64][]uint64)
	proposalSlashed := make(map[uint64]uint64)
	attestationSlashings := make(map[uint64][]uint64)
	attestationSlashed := make(map[uint64]uint64)

	slotStart := queryResult[0].EpochStart * utils.Config.Chain.ClConfig.SlotsPerEpoch
	slotEnd := (queryResult[0].EpochEnd+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1

	wg := errgroup.Group{}

	// Get the proposal slashings
	wg.Go(func() error {
		var queryResult []struct {
			ProposerSlashing uint64 `db:"proposer"`
			ProposerSlashed  uint64 `db:"proposerindex"`
		}

		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("b.proposer"),
				goqu.L("bps.proposerindex")).
			From(goqu.L("blocks_proposerslashings bps")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("b.slot = bps.block_slot"))).
			Where(goqu.L("bps.block_slot >= ? AND bps.block_slot <= ?", slotStart, slotEnd)).
			Where(goqu.L("(b.proposer = ANY(?) OR bps.proposerindex = ANY(?))", pq.Array(slashingValidators), pq.Array(slashedValidators)))

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.alloyReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table blocks_proposerslashings: %w", err)
		}

		for _, queryEntry := range queryResult {
			if _, ok := slashingValidatorsMap[queryEntry.ProposerSlashing]; ok {
				if _, ok := proposalSlashings[queryEntry.ProposerSlashing]; !ok {
					proposalSlashings[queryEntry.ProposerSlashing] = make([]uint64, 0)
				}
				proposalSlashings[queryEntry.ProposerSlashing] = append(proposalSlashings[queryEntry.ProposerSlashing], queryEntry.ProposerSlashed)
			}
			if _, ok := slashedValidatorsMap[queryEntry.ProposerSlashed]; ok {
				proposalSlashed[queryEntry.ProposerSlashed] = queryEntry.ProposerSlashing
			}
		}
		return nil
	})

	// Get the attestation slashings
	wg.Go(func() error {
		var queryResult []struct {
			Proposer               uint64        `db:"proposer"`
			Attestestation1Indices pq.Int64Array `db:"attestation1_indices"`
			Attestestation2Indices pq.Int64Array `db:"attestation2_indices"`
		}

		ds := goqu.Dialect("postgres").
			Select(
				goqu.L("b.proposer"),
				goqu.L("bas.attestation1_indices"),
				goqu.L("bas.attestation2_indices")).
			From(goqu.L("blocks_attesterslashings bas")).
			LeftJoin(goqu.L("blocks b"), goqu.On(goqu.L("b.slot = bas.block_slot"))).
			Where(goqu.L("bas.block_slot >= ? AND bas.block_slot <= ?", slotStart, slotEnd))

		if len(slashedValidators) == 0 {
			// If we don't have any slashed validators then we can just get the slashing validators
			ds = ds.
				Where(goqu.L("b.proposer = ANY(?)", pq.Array(slashingValidators)))
		}

		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = d.alloyReader.SelectContext(ctx, &queryResult, query, args...)
		if err != nil {
			return fmt.Errorf("error retrieving data from table blocks_attesterslashings: %w", err)
		}

		for _, queryEntry := range queryResult {
			inter := intersect.Simple(queryEntry.Attestestation1Indices, queryEntry.Attestestation2Indices)
			if len(inter) == 0 {
				log.WarnWithStackTrace(nil, "No intersection found for attestation violation", 0)
			}
			for _, v := range inter {
				if _, ok := slashingValidatorsMap[queryEntry.Proposer]; ok {
					if _, ok := attestationSlashings[queryEntry.Proposer]; !ok {
						attestationSlashings[queryEntry.Proposer] = make([]uint64, 0)
					}
					attestationSlashings[queryEntry.Proposer] = append(attestationSlashings[queryEntry.Proposer], uint64(v.(int64)))
				}
				if _, ok := slashedValidatorsMap[uint64(v.(int64))]; ok {
					attestationSlashed[uint64(v.(int64))] = queryEntry.Proposer
				}
			}
		}
		return nil
	})

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	// Combine the proposal and attestation slashings
	slashings := make(map[uint64][]uint64)
	for slashingIdx, slashedIdxs := range proposalSlashings {
		if _, ok := slashings[slashingIdx]; !ok {
			slashings[slashingIdx] = make([]uint64, 0)
		}
		slashings[slashingIdx] = append(slashings[slashingIdx], slashedIdxs...)
	}
	for slashingIdx, slashedIdxs := range attestationSlashings {
		if _, ok := slashings[slashingIdx]; !ok {
			slashings[slashingIdx] = make([]uint64, 0)
		}
		slashings[slashingIdx] = append(slashings[slashingIdx], slashedIdxs...)
	}

	// Process the data
	for slashingIdx, slashedIdxs := range slashings {
		result.HasSlashed = append(result.HasSlashed, t.VDBValidatorHasSlashed{
			Index:          slashingIdx,
			SlashedIndices: slashedIdxs,
		})
	}

	// Fill the slashed validators
	for slashedIdx, slashingIdx := range proposalSlashed {
		result.GotSlashed = append(result.GotSlashed, t.VDBValidatorGotSlashed{
			Index:     slashedIdx,
			SlashedBy: slashingIdx,
		})
	}
	for slashedIdx, slashingIdx := range attestationSlashed {
		result.GotSlashed = append(result.GotSlashed, t.VDBValidatorGotSlashed{
			Index:     slashedIdx,
			SlashedBy: slashingIdx,
		})
	}

	return result, nil
}

func (d *DataAccessService) GetValidatorDashboardProposalSummaryValidators(ctx context.Context, dashboardId t.VDBId, groupId int64, period enums.TimePeriod) (*t.VDBProposalSummaryValidators, error) {
	// possible periods are: all_time, last_30d, last_7d, last_24h, last_1h
	result := &t.VDBProposalSummaryValidators{}

	if dashboardId.AggregateGroups {
		// If we are aggregating groups then ignore the group id and sum up everything
		groupId = t.AllGroups
	}

	// Get the table name based on the period
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return nil, err
	}

	var epochQueryResult struct {
		EpochStart uint64 `db:"epoch_start"`
		EpochEnd   uint64 `db:"epoch_end"`
	}

	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("epoch_start"),
			goqu.L("epoch_end")).
		From(goqu.L(fmt.Sprintf("%s FINAL", clickhouseTable))).
		Order(goqu.L("epoch_start").Asc()).
		Limit(1)

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	err = d.clickhouseReader.GetContext(ctx, &epochQueryResult, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving epoch info for proposals: %w", err)
	}

	// Build the query and get the data
	var queryResult []struct {
		Slot           uint64 `db:"slot"`
		Status         string `db:"status"`
		ValidatorIndex uint64 `db:"proposer"`
	}

	ds = goqu.Dialect("postgres").
		Select(
			goqu.L("b.slot"),
			goqu.L("b.status"),
			goqu.L("b.proposer")).
		From(goqu.L("blocks b")).
		Where(goqu.L("b.epoch >= ? AND b.epoch <= ?", epochQueryResult.EpochStart, epochQueryResult.EpochEnd))

	if len(dashboardId.Validators) > 0 {
		ds = ds.
			Where(goqu.L("b.proposer = ANY(?)", pq.Array(dashboardId.Validators)))
	} else {
		ds = ds.
			InnerJoin(goqu.L("users_val_dashboards_validators v"), goqu.On(goqu.L("b.proposer = v.validator_index"))).
			Where(goqu.L("v.dashboard_id = ?", dashboardId.Id))

		if groupId != t.AllGroups {
			ds = ds.Where(goqu.L("v.group_id = ?", groupId))
		}
	}

	query, args, err = ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	err = d.alloyReader.SelectContext(ctx, &queryResult, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data from table blocks: %w", err)
	}

	// Process the data
	proposedValidatorMap := make(map[uint64][]uint64)
	missedValidatorMap := make(map[uint64][]uint64)
	for _, row := range queryResult {
		if row.Status == "1" {
			if _, ok := proposedValidatorMap[row.ValidatorIndex]; !ok {
				proposedValidatorMap[row.ValidatorIndex] = make([]uint64, 0)
			}
			proposedValidatorMap[row.ValidatorIndex] = append(proposedValidatorMap[row.ValidatorIndex], row.Slot)
		} else {
			if _, ok := missedValidatorMap[row.ValidatorIndex]; !ok {
				missedValidatorMap[row.ValidatorIndex] = make([]uint64, 0)
			}
			missedValidatorMap[row.ValidatorIndex] = append(missedValidatorMap[row.ValidatorIndex], row.Slot)
		}
	}

	for validatorIndex, slotNumbers := range proposedValidatorMap {
		result.Proposed = append(result.Proposed, t.IndexSlots{
			Index: validatorIndex,
			Slots: slotNumbers,
		})
	}
	for validatorIndex, slotNumbers := range missedValidatorMap {
		result.Missed = append(result.Missed, t.IndexSlots{
			Index: validatorIndex,
			Slots: slotNumbers,
		})
	}

	return result, nil
}

func (d *DataAccessService) getCurrentAndUpcomingSyncCommittees(ctx context.Context, latestEpoch uint64) (map[uint64]bool, map[uint64]bool, error) {
	currentSyncCommitteeValidators := make(map[uint64]bool)
	upcomingSyncCommitteeValidators := make(map[uint64]bool)

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
		return nil, nil, fmt.Errorf("error preparing query: %w", err)
	}

	err = d.readerDb.SelectContext(ctx, &queryResult, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving sync committee current and next period data: %w", err)
	}

	for _, queryEntry := range queryResult {
		if queryEntry.Period == currentSyncPeriod {
			currentSyncCommitteeValidators[queryEntry.ValidatorIndex] = true
		} else {
			upcomingSyncCommitteeValidators[queryEntry.ValidatorIndex] = true
		}
	}

	return currentSyncCommitteeValidators, upcomingSyncCommitteeValidators, nil
}

func (d *DataAccessService) getEpochStart(ctx context.Context, period enums.TimePeriod) (uint64, error) {
	clickhouseTable, _, err := d.getTablesForPeriod(period)
	if err != nil {
		return 0, err
	}

	ds := goqu.Dialect("postgres").
		Select(goqu.L("epoch_start")).
		From(goqu.L(fmt.Sprintf("%s FINAL", clickhouseTable))).
		Order(goqu.L("epoch_start").Asc()).
		Limit(1)

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return 0, fmt.Errorf("error preparing query: %w", err)
	}

	var epochStart uint64
	err = d.clickhouseReader.GetContext(ctx, &epochStart, query, args...)
	if err != nil {
		return 0, fmt.Errorf("error retrieving cutoff epoch for past sync committees: %w", err)
	}

	return epochStart, nil
}

func (d *DataAccessService) getPastSyncCommittees(ctx context.Context, indicies []uint64, epochStart uint64, latestEpoch uint64) (map[uint64]uint64, error) {
	pastSyncPeriodCutoff := utils.SyncPeriodOfEpoch(epochStart)
	currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)

	// Get the past sync committee validators
	ds := goqu.Dialect("postgres").
		Select(
			goqu.L("sc.validatorindex")).
		From(goqu.L("sync_committees sc")).
		Where(goqu.L("period >= ? AND period < ? AND validatorindex = ANY(?)", pastSyncPeriodCutoff, currentSyncPeriod, pq.Array(indicies)))

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	var validatorIndices []uint64
	err = d.alloyReader.SelectContext(ctx, &validatorIndices, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving data for past sync committees: %w", err)
	}

	validatorCountMap := make(map[uint64]uint64)
	for _, validatorIndex := range validatorIndices {
		validatorCountMap[validatorIndex]++
	}

	return validatorCountMap, nil
}

func (d *DataAccessService) getTablesForPeriod(period enums.TimePeriod) (string, int, error) {
	table := ""
	hours := 0

	switch period {
	case enums.TimePeriods.Last1h:
		table = "validator_dashboard_data_rolling_1h"
		hours = 1
	case enums.TimePeriods.Last24h:
		table = "validator_dashboard_data_rolling_24h"
		hours = 24
	case enums.TimePeriods.Last7d:
		table = "validator_dashboard_data_rolling_7d"
		hours = 7 * 24
	case enums.TimePeriods.Last30d:
		table = "validator_dashboard_data_rolling_30d"
		hours = 30 * 24
	case enums.TimePeriods.AllTime:
		table = "validator_dashboard_data_rolling_total"
		hours = -1
	default:
		return "", 0, fmt.Errorf("not-implemented time period: %v", period)
	}

	return table, hours, nil
}

func (d *DataAccessService) getTableAndDateColumn(aggregation enums.ChartAggregation) (string, string, error) {
	var table, dateColumn string

	switch aggregation {
	case enums.IntervalEpoch:
		table = "validator_dashboard_data_epoch"
		dateColumn = "epoch_timestamp"
	case enums.IntervalHourly:
		table = "validator_dashboard_data_hourly"
		dateColumn = "t"
	case enums.IntervalDaily:
		table = "validator_dashboard_data_daily"
		dateColumn = "t"
	case enums.IntervalWeekly:
		table = "validator_dashboard_data_weekly"
		dateColumn = "t"
	default:
		return "", "", fmt.Errorf("unexpected aggregation type: %v", aggregation)
	}

	return table, dateColumn, nil
}

func (d *DataAccessService) getViewAndDateColumn(aggregation enums.ChartAggregation) (string, string, error) {
	var view, dateColumn string

	switch aggregation {
	case enums.IntervalEpoch:
		view = "view_validator_dashboard_data_epoch_max_ts"
		dateColumn = "t"
	case enums.IntervalHourly:
		view = "view_validator_dashboard_data_hourly_max_ts"
		dateColumn = "t"
	case enums.IntervalDaily:
		view = "view_validator_dashboard_data_daily_max_ts"
		dateColumn = "t"
	case enums.IntervalWeekly:
		view = "view_validator_dashboard_data_weekly_max_ts"
		dateColumn = "t"
	default:
		return "", "", fmt.Errorf("unexpected aggregation type: %v", aggregation)
	}

	return view, dateColumn, nil
}
