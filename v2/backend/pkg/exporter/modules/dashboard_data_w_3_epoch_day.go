package modules

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type epochToDayAggregator struct {
	*dashboardData
	mutex             *sync.Mutex
	rollingAggregator RollingAggregator
}

const PartitionDayWidth = 6

func newEpochToDayAggregator(d *dashboardData) *epochToDayAggregator {
	return &epochToDayAggregator{
		dashboardData: d,
		mutex:         &sync.Mutex{},
		rollingAggregator: RollingAggregator{
			log: d.log,
			RollingAggregatorInt: &DayRollingAggregatorImpl{
				log: d.log,
			},
		},
	}
}

func GetDayAggregateWidth() uint64 {
	return utils.EpochsPerDay()
}

func (d *epochToDayAggregator) dayAggregate(currentExportedEpoch uint64) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	err := d.utcDayAggregate(currentExportedEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to utc day aggregate")
	}

	return nil
}

// used to retrieve missing historic epochs in database for rolling 24h aggregation
// intentedHeadEpoch is the head you currently want to export
func (d *epochToDayAggregator) getMissingRolling24TailEpochs(intendedHeadEpoch uint64) ([]uint64, error) {
	return d.rollingAggregator.getMissingRollingTailEpochs(1, intendedHeadEpoch, "validator_dashboard_data_rolling_daily")
}

func (d *epochToDayAggregator) rolling24hAggregate(currentEpochHead uint64) error {
	return d.rollingAggregator.Aggregate(1, "validator_dashboard_data_rolling_daily", currentEpochHead)
}

// Returns the epoch_start and epoch_end (the epoch bounds of a UTC day) for a given epoch.
// epoch_start is inclusive, epoch_end is exclusive.
func getDayAggregateBounds(epoch uint64) (uint64, uint64) {
	offset := utils.GetEpochOffsetGenesis()
	epoch += offset                                                             // offset to utc
	startOfPartition := epoch / GetDayAggregateWidth() * GetDayAggregateWidth() // inclusive
	endOfPartition := startOfPartition + GetDayAggregateWidth()                 // exclusive
	if startOfPartition < offset {
		startOfPartition = offset
	}
	return startOfPartition - offset, endOfPartition - offset
}

func (d *epochToDayAggregator) utcDayAggregate(currentExportedEpoch uint64) error {
	startTime := time.Now()
	defer func() {
		d.log.Infof("utc day aggregate took %v", time.Since(startTime))
		metrics.TaskDuration.WithLabelValues("exporter_v2dash_agg_utc_day").Observe(time.Since(startTime).Seconds())
	}()

	latestExportedDay, err := edb.GetLastExportedDay()
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to get latest daily epoch")
	}

	gaps, err := edb.GetDashboardEpochGapsBetween(currentExportedEpoch, int64(latestExportedDay.EpochEnd))
	if err != nil {
		return errors.Wrap(err, "failed to get dashboard epoch gaps")
	}

	if len(gaps) > 0 {
		return fmt.Errorf("gaps in dashboard epoch for utc day agg, skipping for now: %v (%v-%v)", gaps, latestExportedDay.EpochEnd, currentExportedEpoch) // sanity, this should never happen
	}

	_, currentEndBound := getDayAggregateBounds(currentExportedEpoch)

	for epoch := latestExportedDay.EpochStart; epoch <= currentEndBound; epoch += GetDayAggregateWidth() {
		boundsStart, boundsEnd := getDayAggregateBounds(epoch)
		if latestExportedDay.EpochEnd == boundsEnd { // no need to update last hour entry if it is complete
			d.log.Infof("skipping updating last day entry since it is complete")
			continue
		}

		// no need to aggregate epoch data that hasn't been exported yet
		if boundsEnd > currentEndBound {
			continue
		}

		// define start bounds as lastHourExported.EpochEnd for first iteration
		if epoch == latestExportedDay.EpochStart {
			boundsStart = latestExportedDay.EpochEnd
		}

		// scope down to max currentExportedEpoch (since epoch data is inclusive, add 1)
		if currentExportedEpoch+1 >= boundsStart && currentExportedEpoch+1 < boundsEnd {
			boundsEnd = currentExportedEpoch + 1
		}

		err = d.aggregateUtcDayWithBounds(boundsStart, boundsEnd)
		if err != nil {
			d.log.Error(err, "failed to aggregate utc day with bounds", 0)
			return errors.Wrap(err, "failed to aggregate utc day  with bounds")
		}
	}

	return nil
}

// firstEpochOfDay incl, lastEpochOfDay excl
func (d *epochToDayAggregator) aggregateUtcDayWithBounds(firstEpochOfDay, lastEpochOfDay uint64) error {
	d.log.Infof("aggregating day of epoch %d", firstEpochOfDay)
	partitionStartRange, partitionEndRange := d.GetDayPartitionRange(lastEpochOfDay)

	err := d.createDayPartition(partitionStartRange, partitionEndRange)
	if err != nil {
		return errors.Wrap(err, "failed to create day partition")
	}

	boundsStart, _ := getDayAggregateBounds(firstEpochOfDay)

	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer utils.Rollback(tx)

	err = AddToRollingCustom(tx, CustomRolling{
		StartEpoch:           firstEpochOfDay,
		EndEpoch:             lastEpochOfDay,
		StartBoundEpoch:      int64(boundsStart),
		TableFrom:            "validator_dashboard_data_epoch",
		TableTo:              "validator_dashboard_data_daily",
		TableFromEpochColumn: "epoch",
		Log:                  d.log,
		TailBalancesQuery: `balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_epoch WHERE epoch = $3
		),`,
		TailBalancesJoinQuery:         `LEFT JOIN balance_starts ON aggregate_head.validator_index = balance_starts.validator_index`,
		TailBalancesInsertColumnQuery: `balance_start,`,
		TableDayColum:                 "day,",
		TableDayValue:                 fmt.Sprintf("'%s' as day,", utils.EpochToTime(boundsStart).Format("2006-01-02")),
		TableConflict:                 "(day, validator_index)",
	})

	if err != nil {
		return errors.Wrap(err, "failed to insert daily aggregate")
	}

	return tx.Commit()
}

func (d *epochToDayAggregator) GetDayPartitionRange(epoch uint64) (time.Time, time.Time) {
	_, boundsEnd := getDayAggregateBounds(epoch)
	startOfPartition := boundsEnd / (PartitionDayWidth * GetDayAggregateWidth()) * PartitionDayWidth * GetDayAggregateWidth() // inclusive
	endOfPartition := startOfPartition + PartitionDayWidth*GetDayAggregateWidth()                                             // exclusive
	return utils.EpochToTime(startOfPartition), utils.EpochToTime(endOfPartition)
}

func (d *epochToDayAggregator) createDayPartition(dayFrom, dayTo time.Time) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS validator_dashboard_data_daily_%s_%s
		PARTITION OF validator_dashboard_data_daily
			FOR VALUES FROM ('%s') TO ('%s')
		`,
		dayToYYMMDDLabel(dayFrom), dayToYYMMDDLabel(dayTo), dayToDDMMYY(dayFrom), dayToDDMMYY(dayTo),
	))
	return err
}

func dayToYYMMDDLabel(day time.Time) string {
	return day.Format("20060102")
}

func dayToDDMMYY(day time.Time) string {
	return day.Format("02-January-2006")
}

// -- rolling aggregate --

type DayRollingAggregatorImpl struct {
	log ModuleLog
}

// returns both start_epochs
// the epoch_start from the the bootstrap tail
// and the epoch_start from the bootstrap head (epoch_start of latestExportedHourEpoch)
func (d *DayRollingAggregatorImpl) getBootstrapBounds(latestExportedHourEpoch uint64, _ uint64) (uint64, uint64) {
	currentStartBounds, _ := getHourAggregateBounds(latestExportedHourEpoch)

	dayOldEpoch := int64(latestExportedHourEpoch - utils.EpochsPerDay())
	if dayOldEpoch < 0 {
		dayOldEpoch = 0
	}
	dayOldBoundsStart, _ := getHourAggregateBounds(uint64(dayOldEpoch))
	return dayOldBoundsStart, currentStartBounds
}

// How many epochs can you be behind in the rolling table without bootstrap
func (d *DayRollingAggregatorImpl) getBootstrapOnEpochsBehind() uint64 {
	return getHourAggregateWidth()
}

// Bootstrap rolling 24h table from hourly table
func (d *DayRollingAggregatorImpl) bootstrap(tx *sqlx.Tx, days int, tableName string) error {
	startTime := time.Now()
	defer func() {
		d.log.Infof("bootstrap 24h aggregate took %v", time.Since(startTime))
	}()

	latestHourlyEpochBounds, err := edb.GetLastExportedHour()
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to get latest dashboard epoch")
	}

	dayOldBoundsStart, latestHourlyEpoch := d.getBootstrapBounds(latestHourlyEpochBounds.EpochEnd, 1)

	var found bool
	err = db.AlloyWriter.Get(&found, `
		SELECT true FROM validator_dashboard_data_hourly WHERE epoch_start = $1 LIMIT 1 
	`, dayOldBoundsStart)
	if err != nil || !found {
		return errors.Wrap(err, fmt.Sprintf("failed to check if tail validator_dashboard_data_hourly epoch_start %v exists", dayOldBoundsStart))
	}

	latestHourlyStartBound, _ := getHourAggregateBounds(latestHourlyEpoch - 1) // excl
	d.log.Infof("latestHourlyEpoch (excl): %d, dayOldHourlyEpoch: %d, latestHourlyStartBound (incl): %d", latestHourlyEpoch, dayOldBoundsStart, latestHourlyStartBound)

	_, err = tx.Exec(`TRUNCATE validator_dashboard_data_rolling_daily`)
	if err != nil {
		return errors.Wrap(err, "failed to delete old rolling 24h aggregate")
	}

	_, err = tx.Exec(`
		WITH
			epoch_ends as (
				SELECT max(epoch_end) as epoch_end FROM validator_dashboard_data_hourly WHERE epoch_start = $2 LIMIT 1
			),
			balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_hourly WHERE epoch_start = $1
			),
			balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_hourly WHERE epoch_start = $2
			),
			aggregate as (
				SELECT 
					validator_index,
					SUM(attestations_source_reward) as attestations_source_reward,
					SUM(attestations_target_reward) as attestations_target_reward,
					SUM(attestations_head_reward) as attestations_head_reward,
					SUM(attestations_inactivity_reward) as attestations_inactivity_reward,
					SUM(attestations_inclusion_reward) as attestations_inclusion_reward,
					SUM(attestations_reward) as attestations_reward,
					SUM(attestations_ideal_source_reward) as attestations_ideal_source_reward,
					SUM(attestations_ideal_target_reward) as attestations_ideal_target_reward,
					SUM(attestations_ideal_head_reward) as attestations_ideal_head_reward,
					SUM(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
					SUM(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
					SUM(attestations_ideal_reward) as attestations_ideal_reward,
					SUM(blocks_scheduled) as blocks_scheduled,
					SUM(blocks_proposed) as blocks_proposed,
					SUM(blocks_cl_reward) as blocks_cl_reward,
					SUM(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
					SUM(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
					SUM(sync_scheduled) as sync_scheduled,
					SUM(sync_executed) as sync_executed,
					SUM(sync_rewards) as sync_rewards,
					bool_or(slashed) as slashed,
					SUM(deposits_count) as deposits_count,
					SUM(deposits_amount) as deposits_amount,
					SUM(withdrawals_count) as withdrawals_count,
					SUM(withdrawals_amount) as withdrawals_amount,
					SUM(inclusion_delay_sum) as inclusion_delay_sum,
					SUM(blocks_expected) as blocks_expected,
					SUM(sync_committees_expected) as sync_committees_expected,
					SUM(attestations_scheduled) as attestations_scheduled,
					SUM(attestations_executed) as attestations_executed,
					SUM(attestation_head_executed) as attestation_head_executed,
					SUM(attestation_source_executed) as attestation_source_executed,
					SUM(attestation_target_executed) as attestation_target_executed,
					SUM(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum,
					SUM(slasher_reward) as slasher_reward,
					MAX(slashed_by) as slashed_by,
					MAX(slashed_violation) as slashed_violation,
					MAX(last_executed_duty_epoch) as last_executed_duty_epoch		
				FROM validator_dashboard_data_hourly
				WHERE epoch_start >= $1 AND epoch_start <= $2
				GROUP BY validator_index
			)
			INSERT INTO validator_dashboard_data_rolling_daily (
				validator_index,
				epoch_start,
				epoch_end,
				attestations_source_reward,
				attestations_target_reward,
				attestations_head_reward,
				attestations_inactivity_reward,
				attestations_inclusion_reward,
				attestations_reward,
				attestations_ideal_source_reward,
				attestations_ideal_target_reward,
				attestations_ideal_head_reward,
				attestations_ideal_inactivity_reward,
				attestations_ideal_inclusion_reward,
				attestations_ideal_reward,
				blocks_scheduled,
				blocks_proposed,
				blocks_cl_reward,
				blocks_cl_attestations_reward,
				blocks_cl_sync_aggregate_reward,
				sync_scheduled,
				sync_executed,
				sync_rewards,
				slashed,
				balance_start,
				balance_end,
				deposits_count,
				deposits_amount,
				withdrawals_count,
				withdrawals_amount,
				inclusion_delay_sum,
				blocks_expected,
				sync_committees_expected,
				attestations_scheduled,
				attestations_executed,
				attestation_head_executed,
				attestation_source_executed,
				attestation_target_executed,
				optimal_inclusion_delay_sum,
				slasher_reward,
				slashed_by,
				slashed_violation,
				last_executed_duty_epoch
			)
			SELECT 
				aggregate.validator_index,
				$1,
				(SELECT epoch_end FROM epoch_ends), 
				attestations_source_reward,
				attestations_target_reward,
				attestations_head_reward,
				attestations_inactivity_reward,
				attestations_inclusion_reward,
				attestations_reward,
				attestations_ideal_source_reward,
				attestations_ideal_target_reward,
				attestations_ideal_head_reward,
				attestations_ideal_inactivity_reward,
				attestations_ideal_inclusion_reward,
				attestations_ideal_reward,
				blocks_scheduled,
				blocks_proposed,
				blocks_cl_reward,
				blocks_cl_attestations_reward,
				blocks_cl_sync_aggregate_reward,
				sync_scheduled,
				sync_executed,
				sync_rewards,
				slashed,
				balance_start,
				balance_end,
				deposits_count,
				deposits_amount,
				withdrawals_count,
				withdrawals_amount,
				inclusion_delay_sum,
				blocks_expected,
				sync_committees_expected,
				attestations_scheduled,
				attestations_executed,
				attestation_head_executed,
				attestation_source_executed,
				attestation_target_executed,
				optimal_inclusion_delay_sum,
				slasher_reward,
				slashed_by,
				slashed_violation,
				last_executed_duty_epoch
			FROM aggregate
			LEFT JOIN balance_starts ON aggregate.validator_index = balance_starts.validator_index
			LEFT JOIN balance_ends ON aggregate.validator_index = balance_ends.validator_index
	`, dayOldBoundsStart, latestHourlyStartBound)

	if err != nil {
		return errors.Wrap(err, "failed to insert rolling 24h aggregate")
	}

	return nil
}
