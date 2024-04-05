package modules

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type hourToDayAggregator struct {
	*dashboardData
	mutex             *sync.Mutex
	rollingAggregator RollingAggregator
}

const PartitionDayWidth = 6

func newHourToDayAggregator(d *dashboardData) *hourToDayAggregator {
	return &hourToDayAggregator{
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

func (d *hourToDayAggregator) dayAggregate(workingOnHead bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	errGroup := &errgroup.Group{}
	errGroup.SetLimit(databaseAggregationParallelism)

	if workingOnHead {
		errGroup.Go(func() error {
			err := d.rolling24hAggregate()
			if err != nil {
				return errors.Wrap(err, "failed to rolling 24h aggregate")
			}
			d.log.Infof("finished dayAggregate rolling 24h")
			return nil
		})
	}

	errGroup.Go(func() error {
		err := d.utcDayAggregate()
		if err != nil {
			return errors.Wrap(err, "failed to utc day aggregate")
		}
		d.log.Infof("finished dayAggregate errgroup utc")
		return nil
	})

	err := errGroup.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to wait for day aggregation")
	}

	d.log.Infof("finished dayAggregate all finished")

	return nil
}

// used to retrieve missing historic epochs in database for rolling 24h aggregation
// intentedHeadEpoch is the head you currently want to export
func (d *hourToDayAggregator) getMissingRolling24TailEpochs(intendedHeadEpoch uint64) ([]uint64, error) {
	return d.rollingAggregator.getMissingRollingTailEpochs(1, intendedHeadEpoch, "validator_dashboard_data_rolling_daily")
}

func (d *hourToDayAggregator) rolling24hAggregate() error {
	return d.rollingAggregator.Aggregate(1, "validator_dashboard_data_rolling_daily")
}

func (d *hourToDayAggregator) getDayAggregateBounds(epoch uint64) (uint64, uint64) {
	offset := utils.GetEpochOffsetGenesis()
	epoch += offset                                                             // offset to utc
	startOfPartition := epoch / GetDayAggregateWidth() * GetDayAggregateWidth() // inclusive
	endOfPartition := startOfPartition + GetDayAggregateWidth()                 // exclusive
	if startOfPartition < offset {
		startOfPartition = offset
	}
	return startOfPartition - offset, endOfPartition - offset
}

func (d *hourToDayAggregator) utcDayAggregate() error {
	startTime := time.Now()
	defer func() {
		d.log.Infof("utc day aggregate took %v", time.Since(startTime))
	}()

	latestExportedDay, err := edb.GetLastExportedDay()
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to get latest daily epoch")
	}

	latestExportedHour, err := edb.GetLastExportedHour()
	if err != nil {
		return errors.Wrap(err, "failed to get latest hourly epoch")
	}

	_, currentEndBound := d.getDayAggregateBounds(latestExportedHour.EpochStart)

	for epoch := latestExportedDay.EpochStart; epoch <= currentEndBound; epoch += GetDayAggregateWidth() {
		boundsStart, boundsEnd := d.getDayAggregateBounds(epoch)
		if latestExportedDay.EpochEnd == boundsEnd { // no need to update last hour entry if it is complete
			d.log.Infof("skipping updating last day entry since it is complete")
			continue
		}

		err = d.aggregateUtcDaySpecific(boundsStart, boundsEnd)
		if err != nil {
			d.log.Error(err, "failed to aggregate utc day specific", 0)
			return errors.Wrap(err, "failed to aggregate utc day specific")
		}
	}

	return nil
}

func (d *hourToDayAggregator) aggregateUtcDaySpecific(firstEpochOfDay, lastEpochOfDay uint64) error {
	d.log.Infof("aggregating day of epoch %d", firstEpochOfDay)
	partitionStartRange, partitionEndRange := d.GetDayPartitionRange(lastEpochOfDay)

	err := d.createDayPartition(partitionStartRange, partitionEndRange)
	if err != nil {
		return errors.Wrap(err, "failed to create day partition")
	}

	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`
		WITH
			end_epoch as (
				SELECT max(epoch_start) as epoch, max(epoch_end) as epoch_end FROM validator_dashboard_data_hourly where epoch_start >= $1 AND epoch_start < $2
			),
			balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_hourly WHERE epoch_start = $1
			),
			balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_hourly WHERE epoch_start = (SELECT epoch FROM end_epoch)
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
					SUM(sync_scheduled) as sync_scheduled,
					SUM(sync_executed) as sync_executed,
					SUM(sync_rewards) as sync_rewards,
					bool_or(slashed) as slashed,
					SUM(deposits_count) as deposits_count,
					SUM(deposits_amount) as deposits_amount,
					SUM(withdrawals_count) as withdrawals_count,
					SUM(withdrawals_amount) as withdrawals_amount,
					SUM(inclusion_delay_sum) as inclusion_delay_sum,
					SUM(sync_chance) as sync_chance,
					SUM(block_chance) as block_chance,
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
				WHERE epoch_start >= $1 AND epoch_start < $2
				GROUP BY validator_index
			)
			INSERT INTO validator_dashboard_data_daily (
				day,
				epoch_start,
				epoch_end,
				validator_index,
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
				sync_chance,
				block_chance,
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
				$3,
				$1,
				(SELECT epoch_end FROM end_epoch), -- exclusive, hence use epoch_end
				aggregate.validator_index,
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
				sync_chance,
				block_chance,
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
			ON CONFLICT (day, validator_index) DO UPDATE SET
				attestations_source_reward = EXCLUDED.attestations_source_reward,
				attestations_target_reward = EXCLUDED.attestations_target_reward,
				attestations_head_reward = EXCLUDED.attestations_head_reward,
				attestations_inactivity_reward = EXCLUDED.attestations_inactivity_reward,
				attestations_inclusion_reward = EXCLUDED.attestations_inclusion_reward,
				attestations_reward = EXCLUDED.attestations_reward,
				attestations_ideal_source_reward = EXCLUDED.attestations_ideal_source_reward,
				attestations_ideal_target_reward = EXCLUDED.attestations_ideal_target_reward,
				attestations_ideal_head_reward = EXCLUDED.attestations_ideal_head_reward,
				attestations_ideal_inactivity_reward = EXCLUDED.attestations_ideal_inactivity_reward,
				attestations_ideal_inclusion_reward = EXCLUDED.attestations_ideal_inclusion_reward,
				attestations_ideal_reward = EXCLUDED.attestations_ideal_reward,
				blocks_scheduled = EXCLUDED.blocks_scheduled,
				blocks_proposed = EXCLUDED.blocks_proposed,
				blocks_cl_reward = EXCLUDED.blocks_cl_reward,
				sync_scheduled = EXCLUDED.sync_scheduled,
				sync_executed = EXCLUDED.sync_executed,
				sync_rewards = EXCLUDED.sync_rewards,
				slashed = EXCLUDED.slashed,
				balance_start = EXCLUDED.balance_start,
				balance_end = EXCLUDED.balance_end,
				deposits_count = EXCLUDED.deposits_count,
				deposits_amount = EXCLUDED.deposits_amount,
				withdrawals_count = EXCLUDED.withdrawals_count,
				withdrawals_amount = EXCLUDED.withdrawals_amount,
				inclusion_delay_sum = EXCLUDED.inclusion_delay_sum,
				sync_chance = EXCLUDED.sync_chance,
				block_chance = EXCLUDED.block_chance,
				attestations_scheduled = EXCLUDED.attestations_scheduled,
				attestations_executed = EXCLUDED.attestations_executed,
				attestation_head_executed = EXCLUDED.attestation_head_executed,
				attestation_source_executed = EXCLUDED.attestation_source_executed,
				attestation_target_executed = EXCLUDED.attestation_target_executed,
				optimal_inclusion_delay_sum = EXCLUDED.optimal_inclusion_delay_sum,
				epoch_start = EXCLUDED.epoch_start,
				epoch_end = EXCLUDED.epoch_end,
				slasher_reward = EXCLUDED.slasher_reward,
				slashed_by = EXCLUDED.slashed_by,
				slashed_violation = EXCLUDED.slashed_violation,
				last_executed_duty_epoch = EXCLUDED.last_executed_duty_epoch
	`, firstEpochOfDay, lastEpochOfDay, utils.EpochToTime(firstEpochOfDay))

	if err != nil {
		return errors.Wrap(err, "failed to insert daily aggregate")
	}

	return tx.Commit()
}

func (d *hourToDayAggregator) GetDayPartitionRange(epoch uint64) (time.Time, time.Time) {
	startOfPartition := epoch / (PartitionDayWidth * GetDayAggregateWidth()) * PartitionDayWidth * GetDayAggregateWidth() // inclusive
	endOfPartition := startOfPartition + PartitionDayWidth*GetDayAggregateWidth()                                         // exclusive
	return utils.EpochToTime(startOfPartition), utils.EpochToTime(endOfPartition)
}

func (d *hourToDayAggregator) createDayPartition(dayFrom, dayTo time.Time) error {
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

func (d *DayRollingAggregatorImpl) getBootstrapOnEpochsBehind() uint64 {
	return getHourAggregateWidth()
}

func (d *DayRollingAggregatorImpl) bootstrapTableToHeadOffset(currentHead uint64) (int64, error) {
	lastExportedHour, err := edb.GetLastExportedHour()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get latest hourly epoch")
	}

	// modulo in case the current epoch triggers a new aggregation, so offset of getHourAggregateWidth() is actually offset 0
	return (int64(currentHead) - (int64(lastExportedHour.EpochStart) - 1)) % int64(getHourAggregateWidth()), nil
}

func (d *DayRollingAggregatorImpl) bootstrap(tx *sqlx.Tx, days int, tableName string) error {
	startTime := time.Now()
	defer func() {
		d.log.Infof("rolling 24h aggregate took %v", time.Since(startTime))
	}()

	latestHourlyEpochBounds, err := edb.GetLastExportedHour()
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to get latest dashboard epoch")
	}
	latestHourlyEpoch := latestHourlyEpochBounds.EpochStart

	dayOldHourlyEpoch, err := edb.Get24hOldHourlyEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get 24h old dashboard epoch")
	}

	d.log.Infof("latestHourlyEpoch: %d, dayOldHourlyEpoch: %d", latestHourlyEpoch, dayOldHourlyEpoch)

	_, err = tx.Exec(`TRUNCATE validator_dashboard_data_rolling_daily`)
	if err != nil {
		return errors.Wrap(err, "failed to delete old rolling 24h aggregate")
	}

	_, err = tx.Exec(`
		WITH
			epoch_ends as (
				SELECT epoch_end FROM validator_dashboard_data_hourly WHERE epoch_start = $2 LIMIT 1
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
					SUM(sync_scheduled) as sync_scheduled,
					SUM(sync_executed) as sync_executed,
					SUM(sync_rewards) as sync_rewards,
					bool_or(slashed) as slashed,
					SUM(deposits_count) as deposits_count,
					SUM(deposits_amount) as deposits_amount,
					SUM(withdrawals_count) as withdrawals_count,
					SUM(withdrawals_amount) as withdrawals_amount,
					SUM(inclusion_delay_sum) as inclusion_delay_sum,
					SUM(sync_chance) as sync_chance,
					SUM(block_chance) as block_chance,
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
				sync_chance,
				block_chance,
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
				sync_chance,
				block_chance,
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
	`, dayOldHourlyEpoch, latestHourlyEpoch)

	if err != nil {
		return errors.Wrap(err, "failed to insert rolling 24h aggregate")
	}

	return nil
}

// all inclusive
// func (d *hourToDayAggregator) aggregateUpdateRolling24h(tx *sqlx.Tx, headEpochStart, headEpochEnd uint64, tailEpochStart, tailEpochEnd int64) error {
// 	startTime := time.Now()
// 	d.log.Infof("aggregating rolling24h head: %d - %d | footer: %d - %d", headEpochStart, headEpochEnd, tailEpochStart, tailEpochEnd)
// 	defer func() {
// 		d.log.Infof("aggregating rolling24h took %v", time.Since(startTime))
// 	}()

// 	if tailEpochEnd < 0 {
// 		// if selected timeframe (24) is more than epochs exists we log an info
// 		d.log.Infof("rolling 24h tail epoch is negative, no end cutting")
// 	}

// 	_, err := tx.Exec(`
// 		WITH
// 			head_balance_ends as (
// 				SELECT validator_index, balance_end FROM validator_dashboard_data_epoch WHERE epoch = $2
// 			),
// 			footer_balance_starts as (
// 				SELECT validator_index, balance_end as balance_start FROM validator_dashboard_data_epoch WHERE epoch = $4 -- since $4 will be removed cause function is incluside, end balance of $4 = start balance of $4 + 1
// 			),
// 			aggregate_head as (
// 				SELECT
// 					validator_index,
// 					SUM(attestations_source_reward) as attestations_source_reward,
// 					SUM(attestations_target_reward) as attestations_target_reward,
// 					SUM(attestations_head_reward) as attestations_head_reward,
// 					SUM(attestations_inactivity_reward) as attestations_inactivity_reward,
// 					SUM(attestations_inclusion_reward) as attestations_inclusion_reward,
// 					SUM(attestations_reward) as attestations_reward,
// 					SUM(attestations_ideal_source_reward) as attestations_ideal_source_reward,
// 					SUM(attestations_ideal_target_reward) as attestations_ideal_target_reward,
// 					SUM(attestations_ideal_head_reward) as attestations_ideal_head_reward,
// 					SUM(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
// 					SUM(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
// 					SUM(attestations_ideal_reward) as attestations_ideal_reward,
// 					SUM(blocks_scheduled) as blocks_scheduled,
// 					SUM(blocks_proposed) as blocks_proposed,
// 					SUM(blocks_cl_reward) as blocks_cl_reward,
// 					SUM(blocks_el_reward) as blocks_el_reward,
// 					SUM(sync_scheduled) as sync_scheduled,
// 					SUM(sync_executed) as sync_executed,
// 					SUM(sync_rewards) as sync_rewards,
// 					bool_or(slashed) as slashed,
// 					SUM(deposits_count) as deposits_count,
// 					SUM(deposits_amount) as deposits_amount,
// 					SUM(withdrawals_count) as withdrawals_count,
// 					SUM(withdrawals_amount) as withdrawals_amount,
// 					SUM(inclusion_delay_sum) as inclusion_delay_sum,
// 					SUM(sync_chance) as sync_chance,
// 					SUM(block_chance) as block_chance,
// 					SUM(attestations_scheduled) as attestations_scheduled,
// 					SUM(attestations_executed) as attestations_executed,
// 					SUM(attestation_head_executed) as attestation_head_executed,
// 					SUM(attestation_source_executed) as attestation_source_executed,
// 					SUM(attestation_target_executed) as attestation_target_executed,
// 					SUM(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum
// 				FROM validator_dashboard_data_epoch
// 				WHERE epoch >= $1 AND epoch <= $2
// 				GROUP BY validator_index
// 			),
// 			aggregate_tail as (
// 				SELECT
// 					validator_index,
// 					SUM(attestations_source_reward) as attestations_source_reward,
// 					SUM(attestations_target_reward) as attestations_target_reward,
// 					SUM(attestations_head_reward) as attestations_head_reward,
// 					SUM(attestations_inactivity_reward) as attestations_inactivity_reward,
// 					SUM(attestations_inclusion_reward) as attestations_inclusion_reward,
// 					SUM(attestations_reward) as attestations_reward,
// 					SUM(attestations_ideal_source_reward) as attestations_ideal_source_reward,
// 					SUM(attestations_ideal_target_reward) as attestations_ideal_target_reward,
// 					SUM(attestations_ideal_head_reward) as attestations_ideal_head_reward,
// 					SUM(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
// 					SUM(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
// 					SUM(attestations_ideal_reward) as attestations_ideal_reward,
// 					SUM(blocks_scheduled) as blocks_scheduled,
// 					SUM(blocks_proposed) as blocks_proposed,
// 					SUM(blocks_cl_reward) as blocks_cl_reward,
// 					SUM(blocks_el_reward) as blocks_el_reward,
// 					SUM(sync_scheduled) as sync_scheduled,
// 					SUM(sync_executed) as sync_executed,
// 					SUM(sync_rewards) as sync_rewards,
// 					SUM(deposits_count) as deposits_count,
// 					SUM(deposits_amount) as deposits_amount,
// 					SUM(withdrawals_count) as withdrawals_count,
// 					SUM(withdrawals_amount) as withdrawals_amount,
// 					SUM(inclusion_delay_sum) as inclusion_delay_sum,
// 					SUM(sync_chance) as sync_chance,
// 					SUM(block_chance) as block_chance,
// 					SUM(attestations_scheduled) as attestations_scheduled,
// 					SUM(attestations_executed) as attestations_executed,
// 					SUM(attestation_head_executed) as attestation_head_executed,
// 					SUM(attestation_source_executed) as attestation_source_executed,
// 					SUM(attestation_target_executed) as attestation_target_executed,
// 					SUM(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum
// 				FROM validator_dashboard_data_epoch
// 				WHERE epoch >= $3 AND epoch <= $4
// 				GROUP BY validator_index
// 			),
// 			result as (
// 				SELECT
// 					$4 + 1 as epoch_start, --since its inclusive in the func $4 will be deducted hence +1
// 					$2 + 1 as epoch_end, -- exclusive
// 					aggregate_head.validator_index,
// 					COALESCE(aggregate_head.attestations_source_reward, 0) - COALESCE(aggregate_tail.attestations_source_reward, 0) as attestations_source_reward,
// 					COALESCE(aggregate_head.attestations_target_reward, 0) - COALESCE(aggregate_tail.attestations_target_reward, 0) as attestations_target_reward,
// 					COALESCE(aggregate_head.attestations_head_reward, 0) - COALESCE(aggregate_tail.attestations_head_reward, 0) as attestations_head_reward,
// 					COALESCE(aggregate_head.attestations_inactivity_reward, 0) - COALESCE(aggregate_tail.attestations_inactivity_reward, 0) as attestations_inactivity_reward,
// 					COALESCE(aggregate_head.attestations_inclusion_reward, 0) - COALESCE(aggregate_tail.attestations_inclusion_reward, 0) as attestations_inclusion_reward,
// 					COALESCE(aggregate_head.attestations_reward, 0) - COALESCE(aggregate_tail.attestations_reward, 0) as attestations_reward,
// 					COALESCE(aggregate_head.attestations_ideal_source_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
// 					COALESCE(aggregate_head.attestations_ideal_target_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
// 					COALESCE(aggregate_head.attestations_ideal_head_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
// 					COALESCE(aggregate_head.attestations_ideal_inactivity_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
// 					COALESCE(aggregate_head.attestations_ideal_inclusion_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
// 					COALESCE(aggregate_head.attestations_ideal_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_reward, 0) as attestations_ideal_reward,
// 					COALESCE(aggregate_head.blocks_scheduled, 0) - COALESCE(aggregate_tail.blocks_scheduled, 0) as blocks_scheduled,
// 					COALESCE(aggregate_head.blocks_proposed, 0) - COALESCE(aggregate_tail.blocks_proposed, 0) as blocks_proposed,
// 					COALESCE(aggregate_head.blocks_cl_reward, 0) - COALESCE(aggregate_tail.blocks_cl_reward, 0) as blocks_cl_reward,
// 					COALESCE(aggregate_head.blocks_el_reward, 0) - COALESCE(aggregate_tail.blocks_el_reward, 0) as blocks_el_reward,
// 					COALESCE(aggregate_head.sync_scheduled, 0) - COALESCE(aggregate_tail.sync_scheduled, 0) as sync_scheduled,
// 					COALESCE(aggregate_head.sync_executed, 0) - COALESCE(aggregate_tail.sync_executed, 0) as sync_executed,
// 					COALESCE(aggregate_head.sync_rewards, 0) - COALESCE(aggregate_tail.sync_rewards, 0) as sync_rewards,
// 					aggregate_head.slashed,
// 					balance_start,
// 					balance_end,
// 					COALESCE(aggregate_head.deposits_count, 0) - COALESCE(aggregate_tail.deposits_count, 0) as deposits_count,
// 					COALESCE(aggregate_head.deposits_amount, 0) - COALESCE(aggregate_tail.deposits_amount, 0) as deposits_amount,
// 					COALESCE(aggregate_head.withdrawals_count, 0) - COALESCE(aggregate_tail.withdrawals_count, 0) as withdrawals_count,
// 					COALESCE(aggregate_head.withdrawals_amount, 0) - COALESCE(aggregate_tail.withdrawals_amount, 0) as withdrawals_amount,
// 					COALESCE(aggregate_head.inclusion_delay_sum, 0) - COALESCE(aggregate_tail.inclusion_delay_sum, 0) as inclusion_delay_sum,
// 					COALESCE(aggregate_head.sync_chance, 0) - COALESCE(aggregate_tail.sync_chance, 0) as sync_chance,
// 					COALESCE(aggregate_head.block_chance, 0) - COALESCE(aggregate_tail.block_chance, 0) as block_chance,
// 					COALESCE(aggregate_head.attestations_scheduled, 0) - COALESCE(aggregate_tail.attestations_scheduled, 0) as attestations_scheduled,
// 					COALESCE(aggregate_head.attestations_executed, 0) - COALESCE(aggregate_tail.attestations_executed, 0) as attestations_executed,
// 					COALESCE(aggregate_head.attestation_head_executed, 0) - COALESCE(aggregate_tail.attestation_head_executed, 0) as attestation_head_executed,
// 					COALESCE(aggregate_head.attestation_source_executed, 0) - COALESCE(aggregate_tail.attestation_source_executed, 0) as attestation_source_executed,
// 					COALESCE(aggregate_head.attestation_target_executed, 0) - COALESCE(aggregate_tail.attestation_target_executed, 0) as attestation_target_executed,
// 					COALESCE(aggregate_head.optimal_inclusion_delay_sum, 0) - COALESCE(aggregate_tail.optimal_inclusion_delay_sum, 0) as optimal_inclusion_delay_sum
// 				FROM aggregate_head
// 				LEFT JOIN aggregate_tail ON aggregate_head.validator_index = aggregate_tail.validator_index
// 				LEFT JOIN footer_balance_starts ON aggregate_head.validator_index = footer_balance_starts.validator_index
// 				LEFT JOIN head_balance_ends ON aggregate_head.validator_index = head_balance_ends.validator_index
// 			)
// 			UPDATE validator_dashboard_data_rolling_daily SET
// 					attestations_source_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_source_reward, 0) + result.attestations_source_reward,
// 					attestations_target_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_target_reward, 0) + result.attestations_target_reward,
// 					attestations_head_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_head_reward, 0) + result.attestations_head_reward,
// 					attestations_inactivity_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_inactivity_reward, 0) + result.attestations_inactivity_reward,
// 					attestations_inclusion_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_inclusion_reward, 0) + result.attestations_inclusion_reward,
// 					attestations_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_reward, 0) + result.attestations_reward,
// 					attestations_ideal_source_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_source_reward, 0) + result.attestations_ideal_source_reward,
// 					attestations_ideal_target_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_target_reward, 0) + result.attestations_ideal_target_reward,
// 					attestations_ideal_head_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_head_reward, 0) + result.attestations_ideal_head_reward,
// 					attestations_ideal_inactivity_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_inactivity_reward, 0) + result.attestations_ideal_inactivity_reward,
// 					attestations_ideal_inclusion_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_inclusion_reward, 0) + result.attestations_ideal_inclusion_reward,
// 					attestations_ideal_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_reward, 0) + result.attestations_ideal_reward,
// 					blocks_scheduled = COALESCE(validator_dashboard_data_rolling_daily.blocks_scheduled, 0) + result.blocks_scheduled,
// 					blocks_proposed = COALESCE(validator_dashboard_data_rolling_daily.blocks_proposed, 0) + result.blocks_proposed,
// 					blocks_cl_reward = COALESCE(validator_dashboard_data_rolling_daily.blocks_cl_reward, 0) + result.blocks_cl_reward,
// 					blocks_el_reward = COALESCE(validator_dashboard_data_rolling_daily.blocks_el_reward, 0) + result.blocks_el_reward,
// 					sync_scheduled = COALESCE(validator_dashboard_data_rolling_daily.sync_scheduled, 0) + result.sync_scheduled,
// 					sync_executed = COALESCE(validator_dashboard_data_rolling_daily.sync_executed, 0) + result.sync_executed,
// 					sync_rewards = COALESCE(validator_dashboard_data_rolling_daily.sync_rewards, 0) + result.sync_rewards,
// 					slashed = result.slashed,
// 					balance_end = result.balance_end,
// 					deposits_count = COALESCE(validator_dashboard_data_rolling_daily.deposits_count, 0) + result.deposits_count,
// 					deposits_amount = COALESCE(validator_dashboard_data_rolling_daily.deposits_amount, 0) + result.deposits_amount,
// 					withdrawals_count = COALESCE(validator_dashboard_data_rolling_daily.withdrawals_count, 0) + result.withdrawals_count,
// 					withdrawals_amount = COALESCE(validator_dashboard_data_rolling_daily.withdrawals_amount, 0) + result.withdrawals_amount,
// 					inclusion_delay_sum = COALESCE(validator_dashboard_data_rolling_daily.inclusion_delay_sum, 0) + result.inclusion_delay_sum,
// 					sync_chance = COALESCE(validator_dashboard_data_rolling_daily.sync_chance, 0) + result.sync_chance,
// 					block_chance = COALESCE(validator_dashboard_data_rolling_daily.block_chance, 0) + result.block_chance,
// 					attestations_scheduled = COALESCE(validator_dashboard_data_rolling_daily.attestations_scheduled, 0) + result.attestations_scheduled,
// 					attestations_executed = COALESCE(validator_dashboard_data_rolling_daily.attestations_executed, 0) + result.attestations_executed,
// 					attestation_head_executed = COALESCE(validator_dashboard_data_rolling_daily.attestation_head_executed, 0) + result.attestation_head_executed,
// 					attestation_source_executed = COALESCE(validator_dashboard_data_rolling_daily.attestation_source_executed, 0) + result.attestation_source_executed,
// 					attestation_target_executed = COALESCE(validator_dashboard_data_rolling_daily.attestation_target_executed, 0) + result.attestation_target_executed,
// 					optimal_inclusion_delay_sum = COALESCE(validator_dashboard_data_rolling_daily.optimal_inclusion_delay_sum, 0) + result.optimal_inclusion_delay_sum,
// 					epoch_end = result.epoch_end,
// 					epoch_start = result.epoch_start
// 				FROM result
// 				WHERE validator_dashboard_data_rolling_daily.validator_index = result.validator_index;

// 	`, headEpochStart, headEpochEnd, tailEpochStart, tailEpochEnd)
// 	/*

// 	   INSERT INTO validator_dashboard_data_rolling_daily (
// 	   				epoch_start,
// 	   				epoch_end,
// 	   				validator_index,
// 	   				attestations_source_reward,
// 	   				attestations_target_reward,
// 	   				attestations_head_reward,
// 	   				attestations_inactivity_reward,
// 	   				attestations_inclusion_reward,
// 	   				attestations_reward,
// 	   				attestations_ideal_source_reward,
// 	   				attestations_ideal_target_reward,
// 	   				attestations_ideal_head_reward,
// 	   				attestations_ideal_inactivity_reward,
// 	   				attestations_ideal_inclusion_reward,
// 	   				attestations_ideal_reward,
// 	   				blocks_scheduled,
// 	   				blocks_proposed,
// 	   				blocks_cl_reward,
// 	   				blocks_el_reward,
// 	   				sync_scheduled,
// 	   				sync_executed,
// 	   				sync_rewards,
// 	   				slashed,
// 	   				balance_start,
// 	   				balance_end,
// 	   				deposits_count,
// 	   				deposits_amount,
// 	   				withdrawals_count,
// 	   				withdrawals_amount,
// 	   				inclusion_delay_sum,
// 	   				sync_chance,
// 	   				block_chance,
// 	   				attestations_scheduled,
// 	   				attestations_executed,
// 	   				attestation_head_executed,
// 	   				attestation_source_executed,
// 	   				attestation_target_executed,
// 	   				optimal_inclusion_delay_sum
// 	   			)
// 	   			SELECT
// 	   				$4 + 1, --since its inclusive in the func $4 will be duducted hence +1
// 	   				$2 + 1, -- exclusive
// 	   				aggregate_head.validator_index,
// 	   				COALESCE(aggregate_head.attestations_source_reward, 0) - COALESCE(aggregate_tail.attestations_source_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_target_reward, 0) - COALESCE(aggregate_tail.attestations_target_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_head_reward, 0) - COALESCE(aggregate_tail.attestations_head_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_inactivity_reward, 0) - COALESCE(aggregate_tail.attestations_inactivity_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_inclusion_reward, 0) - COALESCE(aggregate_tail.attestations_inclusion_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_reward, 0) - COALESCE(aggregate_tail.attestations_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_ideal_source_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_source_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_ideal_target_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_target_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_ideal_head_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_head_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_ideal_inactivity_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_inactivity_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_ideal_inclusion_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_inclusion_reward, 0),
// 	   				COALESCE(aggregate_head.attestations_ideal_reward, 0) - COALESCE(aggregate_tail.attestations_ideal_reward, 0),
// 	   				COALESCE(aggregate_head.blocks_scheduled, 0) - COALESCE(aggregate_tail.blocks_scheduled, 0),
// 	   				COALESCE(aggregate_head.blocks_proposed, 0) - COALESCE(aggregate_tail.blocks_proposed, 0),
// 	   				COALESCE(aggregate_head.blocks_cl_reward, 0) - COALESCE(aggregate_tail.blocks_cl_reward, 0),
// 	   				COALESCE(aggregate_head.blocks_el_reward, 0) - COALESCE(aggregate_tail.blocks_el_reward, 0),
// 	   				COALESCE(aggregate_head.sync_scheduled, 0) - COALESCE(aggregate_tail.sync_scheduled, 0),
// 	   				COALESCE(aggregate_head.sync_executed, 0) - COALESCE(aggregate_tail.sync_executed, 0),
// 	   				COALESCE(aggregate_head.sync_rewards, 0) - COALESCE(aggregate_tail.sync_rewards, 0),
// 	   				aggregate_head.slashed,
// 	   				balance_start,
// 	   				balance_end,
// 	   				COALESCE(aggregate_head.deposits_count, 0) - COALESCE(aggregate_tail.deposits_count, 0),
// 	   				COALESCE(aggregate_head.deposits_amount, 0) - COALESCE(aggregate_tail.deposits_amount, 0),
// 	   				COALESCE(aggregate_head.withdrawals_count, 0) - COALESCE(aggregate_tail.withdrawals_count, 0),
// 	   				COALESCE(aggregate_head.withdrawals_amount, 0) - COALESCE(aggregate_tail.withdrawals_amount, 0),
// 	   				COALESCE(aggregate_head.inclusion_delay_sum, 0) - COALESCE(aggregate_tail.inclusion_delay_sum, 0),
// 	   				COALESCE(aggregate_head.sync_chance, 0) - COALESCE(aggregate_tail.sync_chance, 0),
// 	   				COALESCE(aggregate_head.block_chance, 0) - COALESCE(aggregate_tail.block_chance, 0),
// 	   				COALESCE(aggregate_head.attestations_scheduled, 0) - COALESCE(aggregate_tail.attestations_scheduled, 0),
// 	   				COALESCE(aggregate_head.attestations_executed, 0) - COALESCE(aggregate_tail.attestations_executed, 0),
// 	   				COALESCE(aggregate_head.attestation_head_executed, 0) - COALESCE(aggregate_tail.attestation_head_executed, 0),
// 	   				COALESCE(aggregate_head.attestation_source_executed, 0) - COALESCE(aggregate_tail.attestation_source_executed, 0),
// 	   				COALESCE(aggregate_head.attestation_target_executed, 0) - COALESCE(aggregate_tail.attestation_target_executed, 0),
// 	   				COALESCE(aggregate_head.optimal_inclusion_delay_sum, 0) - COALESCE(aggregate_tail.optimal_inclusion_delay_sum, 0)
// 	   			FROM aggregate_head
// 	   			LEFT JOIN aggregate_tail ON aggregate_head.validator_index = aggregate_tail.validator_index
// 	   			LEFT JOIN footer_balance_starts ON aggregate_head.validator_index = footer_balance_starts.validator_index
// 	   			LEFT JOIN head_balance_ends ON aggregate_head.validator_index = head_balance_ends.validator_index
// 	   			ON CONFLICT (validator_index) DO UPDATE SET
// 	   				attestations_source_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_source_reward, 0) + EXCLUDED.attestations_source_reward,
// 	   				attestations_target_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_target_reward, 0) + EXCLUDED.attestations_target_reward,
// 	   				attestations_head_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_head_reward, 0) + EXCLUDED.attestations_head_reward,
// 	   				attestations_inactivity_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_inactivity_reward, 0) + EXCLUDED.attestations_inactivity_reward,
// 	   				attestations_inclusion_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_inclusion_reward, 0) + EXCLUDED.attestations_inclusion_reward,
// 	   				attestations_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_reward, 0) + EXCLUDED.attestations_reward,
// 	   				attestations_ideal_source_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_source_reward, 0) + EXCLUDED.attestations_ideal_source_reward,
// 	   				attestations_ideal_target_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_target_reward, 0) + EXCLUDED.attestations_ideal_target_reward,
// 	   				attestations_ideal_head_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_head_reward, 0) + EXCLUDED.attestations_ideal_head_reward,
// 	   				attestations_ideal_inactivity_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_inactivity_reward, 0) + EXCLUDED.attestations_ideal_inactivity_reward,
// 	   				attestations_ideal_inclusion_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_inclusion_reward, 0) + EXCLUDED.attestations_ideal_inclusion_reward,
// 	   				attestations_ideal_reward = COALESCE(validator_dashboard_data_rolling_daily.attestations_ideal_reward, 0) + EXCLUDED.attestations_ideal_reward,
// 	   				blocks_scheduled = COALESCE(validator_dashboard_data_rolling_daily.blocks_scheduled, 0) + EXCLUDED.blocks_scheduled,
// 	   				blocks_proposed = COALESCE(validator_dashboard_data_rolling_daily.blocks_proposed, 0) + EXCLUDED.blocks_proposed,
// 	   				blocks_cl_reward = COALESCE(validator_dashboard_data_rolling_daily.blocks_cl_reward, 0) + EXCLUDED.blocks_cl_reward,
// 	   				blocks_el_reward = COALESCE(validator_dashboard_data_rolling_daily.blocks_el_reward, 0) + EXCLUDED.blocks_el_reward,
// 	   				sync_scheduled = COALESCE(validator_dashboard_data_rolling_daily.sync_scheduled, 0) + EXCLUDED.sync_scheduled,
// 	   				sync_executed = COALESCE(validator_dashboard_data_rolling_daily.sync_executed, 0) + EXCLUDED.sync_executed,
// 	   				sync_rewards = COALESCE(validator_dashboard_data_rolling_daily.sync_rewards, 0) + EXCLUDED.sync_rewards,
// 	   				slashed = EXCLUDED.slashed,
// 	   				balance_end = EXCLUDED.balance_end,
// 	   				deposits_count = COALESCE(validator_dashboard_data_rolling_daily.deposits_count, 0) + EXCLUDED.deposits_count,
// 	   				deposits_amount = COALESCE(validator_dashboard_data_rolling_daily.deposits_amount, 0) + EXCLUDED.deposits_amount,
// 	   				withdrawals_count = COALESCE(validator_dashboard_data_rolling_daily.withdrawals_count, 0) + EXCLUDED.withdrawals_count,
// 	   				withdrawals_amount = COALESCE(validator_dashboard_data_rolling_daily.withdrawals_amount, 0) + EXCLUDED.withdrawals_amount,
// 	   				inclusion_delay_sum = COALESCE(validator_dashboard_data_rolling_daily.inclusion_delay_sum, 0) + EXCLUDED.inclusion_delay_sum,
// 	   				sync_chance = COALESCE(validator_dashboard_data_rolling_daily.sync_chance, 0) + EXCLUDED.sync_chance,
// 	   				block_chance = COALESCE(validator_dashboard_data_rolling_daily.block_chance, 0) + EXCLUDED.block_chance,
// 	   				attestations_scheduled = COALESCE(validator_dashboard_data_rolling_daily.attestations_scheduled, 0) + EXCLUDED.attestations_scheduled,
// 	   				attestations_executed = COALESCE(validator_dashboard_data_rolling_daily.attestations_executed, 0) + EXCLUDED.attestations_executed,
// 	   				attestation_head_executed = COALESCE(validator_dashboard_data_rolling_daily.attestation_head_executed, 0) + EXCLUDED.attestation_head_executed,
// 	   				attestation_source_executed = COALESCE(validator_dashboard_data_rolling_daily.attestation_source_executed, 0) + EXCLUDED.attestation_source_executed,
// 	   				attestation_target_executed = COALESCE(validator_dashboard_data_rolling_daily.attestation_target_executed, 0) + EXCLUDED.attestation_target_executed,
// 	   				optimal_inclusion_delay_sum = COALESCE(validator_dashboard_data_rolling_daily.optimal_inclusion_delay_sum, 0) + EXCLUDED.optimal_inclusion_delay_sum,
// 	   				epoch_end = EXCLUDED.epoch_end,
// 	   				epoch_start = EXCLUDED.epoch_start
// 	*/
// 	return err
// }
