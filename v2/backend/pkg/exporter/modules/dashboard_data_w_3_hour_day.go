package modules

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/pkg/errors"
)

type hourToDayAggregator struct {
	*dashboardData
	mutex *sync.Mutex
}

const DayAggregateWidth = 225 // todo gnosis
const PartitionDayWidth = 14

func newHourToDayAggregator(d *dashboardData) *hourToDayAggregator {
	return &hourToDayAggregator{
		dashboardData: d,
		mutex:         &sync.Mutex{},
	}
}

func (d *hourToDayAggregator) dayAggregate() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.rolling24hAggregate()
	d.utcDayAggregate()
}

func (d *hourToDayAggregator) rolling24hAggregate() {
	startTime := time.Now()
	defer func() {
		d.log.Infof("rolling 24h aggregate took %v", time.Since(startTime))
	}()

	latestHourlyEpochBounds, err := edb.GetLastExportedHour()
	if err != nil && err != sql.ErrNoRows {
		d.log.Error(err, "failed to get latest dashboard epoch", 0)
		return
	}
	latestHourlyEpoch := latestHourlyEpochBounds.EpochStart

	dayOldHourlyEpoch, err := edb.Get24hOldHourlyEpoch()
	if err != nil {
		d.log.Error(err, "failed to get 24h old dashboard epoch", 0)
		return
	}

	d.log.Infof("latestHourlyEpoch: %d, dayOldHourlyEpoch: %d", latestHourlyEpoch, dayOldHourlyEpoch)

	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		d.log.Error(err, "failed to start transaction", 0)
		return
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`TRUNCATE validator_dashboard_data_rolling_daily`)
	if err != nil {
		d.log.Error(err, "failed to delete old rolling 24h aggregate", 0)
		return
	}

	_, err = tx.Exec(`
		WITH
			balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_hourly WHERE epoch_start = $1
			),
			balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_hourly WHERE epoch_start = $2
			),
			aggregate as (
				SELECT 
					validator_index,
					COALESCE(SUM(COALESCE(attestations_source_reward, 0)),0) as attestations_source_reward,
					COALESCE(SUM(COALESCE(attestations_target_reward, 0)),0) as attestations_target_reward,
					COALESCE(SUM(COALESCE(attestations_head_reward, 0)),0) as attestations_head_reward,
					COALESCE(SUM(COALESCE(attestations_inactivity_reward, 0)),0) as attestations_inactivity_reward,
					COALESCE(SUM(COALESCE(attestations_inclusion_reward, 0)),0) as attestations_inclusion_reward,
					COALESCE(SUM(COALESCE(attestations_reward, 0)),0) as attestations_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_source_reward, 0)),0) as attestations_ideal_source_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_target_reward, 0)),0) as attestations_ideal_target_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_head_reward, 0)),0) as attestations_ideal_head_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_inactivity_reward, 0)),0) as attestations_ideal_inactivity_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_inclusion_reward, 0)),0) as attestations_ideal_inclusion_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_reward, 0)),0) as attestations_ideal_reward,
					COALESCE(SUM(COALESCE(blocks_scheduled, 0)),0) as blocks_scheduled,
					COALESCE(SUM(COALESCE(blocks_proposed, 0)),0) as blocks_proposed,
					COALESCE(SUM(COALESCE(blocks_cl_reward, 0)),0) as blocks_cl_reward,
					COALESCE(SUM(COALESCE(blocks_el_reward, 0)),0) as blocks_el_reward,
					COALESCE(SUM(COALESCE(sync_scheduled, 0)),0) as sync_scheduled,
					COALESCE(SUM(COALESCE(sync_executed, 0)),0) as sync_executed,
					COALESCE(SUM(COALESCE(sync_rewards, 0)),0) as sync_rewards,
					bool_or(slashed) as slashed,
					COALESCE(SUM(COALESCE(deposits_count, 0)),0) as deposits_count,
					COALESCE(SUM(COALESCE(deposits_amount, 0)),0) as deposits_amount,
					COALESCE(SUM(COALESCE(withdrawals_count, 0)),0) as withdrawals_count,
					COALESCE(SUM(COALESCE(withdrawals_amount, 0)),0) as withdrawals_amount,
					COALESCE(SUM(COALESCE(inclusion_delay_sum, 0)),0) as inclusion_delay_sum,
					COALESCE(SUM(COALESCE(sync_chance, 0)),0) as sync_chance,
					COALESCE(SUM(COALESCE(block_chance, 0)),0) as block_chance,
					COALESCE(SUM(COALESCE(attestations_scheduled, 0)),0) as attestations_scheduled,
					COALESCE(SUM(COALESCE(attestations_executed, 0)),0) as attestations_executed,
					COALESCE(SUM(COALESCE(attestation_head_executed, 0)),0) as attestation_head_executed,
					COALESCE(SUM(COALESCE(attestation_source_executed, 0)),0) as attestation_source_executed,
					COALESCE(SUM(COALESCE(attestation_target_executed, 0)),0) as attestation_target_executed,
					COALESCE(SUM(COALESCE(optimal_inclusion_delay_sum, 0)),0) as optimal_inclusion_delay_sum
				FROM validator_dashboard_data_hourly
				WHERE epoch_start >= $1 AND epoch_start <= $2
				GROUP BY validator_index
			)
			INSERT INTO validator_dashboard_data_rolling_daily (
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
				blocks_el_reward,
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
				optimal_inclusion_delay_sum
			)
			SELECT 
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
				blocks_el_reward,
				sync_scheduled,
				sync_executed,
				sync_rewards,
				slashed,
				COALESCE(balance_start, 0),
				COALESCE(balance_end,0),
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
				optimal_inclusion_delay_sum
			FROM aggregate
			LEFT JOIN balance_starts ON aggregate.validator_index = balance_starts.validator_index
			LEFT JOIN balance_ends ON aggregate.validator_index = balance_ends.validator_index
	`, dayOldHourlyEpoch, latestHourlyEpoch)

	if err != nil {
		d.log.Error(err, "failed to insert rolling 24h aggregate", 0)
		return
	}

	err = tx.Commit()
	if err != nil {
		d.log.Error(err, "failed to commit transaction", 0)
		return
	}

	minHourlyDBEpoch, err := edb.GetMinOldHourlyEpoch()
	if err != nil {
		d.log.Error(err, "failed to get min old hourly epoch", 0)
		return
	}

	//Clear old partitions
	var delEpoch uint64
	for i := uint64(0); ; i += HourAggregateWidth {
		delEpoch = dayOldHourlyEpoch - d.epochToHour.getHourRetentionDurationEpochs() - i

		startOfPartition, endOfPartition := d.epochToHour.GetHourPartitionRange(delEpoch)
		err := d.epochToHour.deleteHourlyPartition(startOfPartition, endOfPartition)
		if err != nil {
			d.log.Error(err, "failed to delete old hourly partition", 0)
			return
		}
		d.log.Infof("deleted old hourly partition %d_%d", startOfPartition, endOfPartition)

		if delEpoch < minHourlyDBEpoch {
			break
		}
	}
}

func (d *hourToDayAggregator) getDayAggregateBounds(epoch uint64) (uint64, uint64) {
	offset := utils.GetEpochOffsetGenesis()
	epoch += offset                                                   // offset to utc
	startOfPartition := epoch / DayAggregateWidth * DayAggregateWidth // inclusive
	endOfPartition := startOfPartition + DayAggregateWidth            // exclusive
	return startOfPartition - offset, endOfPartition - offset
}

func (d *hourToDayAggregator) utcDayAggregate() {
	startTime := time.Now()
	defer func() {
		d.log.Infof("utc day aggregate took %v", time.Since(startTime))
	}()

	latestDayBounds, err := edb.GetLastExportedDay()
	if err != nil && err != sql.ErrNoRows {
		d.log.Error(err, "failed to get latest daily epoch", 0)
		return
	}

	latestHourlyBounds, err := edb.GetLastExportedHour()
	if err != nil {
		d.log.Error(err, "failed to get latest hourly epoch", 0)
		return
	}

	if latestDayBounds.EpochStart == 0 {
		latestDayBounds.EpochStart = latestHourlyBounds.EpochStart
	}

	_, currentEndBound := d.getDayAggregateBounds(latestHourlyBounds.EpochStart)

	for epoch := latestDayBounds.EpochStart; epoch <= currentEndBound; epoch += DayAggregateWidth {
		boundsStart, boundsEnd := d.getDayAggregateBounds(epoch)
		if latestDayBounds.EpochEnd == boundsEnd { // no need to update last hour entry if it is complete
			d.log.Infof("skipping updating last day entry since it is complete")
			continue
		}

		err = d.aggregateUtcDaySpecific(boundsStart, boundsEnd)
		if err != nil {
			d.log.Error(err, "failed to aggregate utc day specific", 0)
			return
		}
	}
}

func (d *hourToDayAggregator) aggregateUtcDaySpecific(firstEpochOfDay, lastEpochOfDay uint64) error {
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
				SELECT max(epoch_start) as epoch FROM validator_dashboard_data_hourly where epoch_start >= $1 AND epoch_start < $2
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
					COALESCE(SUM(COALESCE(attestations_source_reward, 0)),0) as attestations_source_reward,
					COALESCE(SUM(COALESCE(attestations_target_reward, 0)),0) as attestations_target_reward,
					COALESCE(SUM(COALESCE(attestations_head_reward, 0)),0) as attestations_head_reward,
					COALESCE(SUM(COALESCE(attestations_inactivity_reward, 0)),0) as attestations_inactivity_reward,
					COALESCE(SUM(COALESCE(attestations_inclusion_reward, 0)),0) as attestations_inclusion_reward,
					COALESCE(SUM(COALESCE(attestations_reward, 0)),0) as attestations_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_source_reward, 0)),0) as attestations_ideal_source_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_target_reward, 0)),0) as attestations_ideal_target_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_head_reward, 0)),0) as attestations_ideal_head_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_inactivity_reward, 0)),0) as attestations_ideal_inactivity_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_inclusion_reward, 0)),0) as attestations_ideal_inclusion_reward,
					COALESCE(SUM(COALESCE(attestations_ideal_reward, 0)),0) as attestations_ideal_reward,
					COALESCE(SUM(COALESCE(blocks_scheduled, 0)),0) as blocks_scheduled,
					COALESCE(SUM(COALESCE(blocks_proposed, 0)),0) as blocks_proposed,
					COALESCE(SUM(COALESCE(blocks_cl_reward, 0)),0) as blocks_cl_reward,
					COALESCE(SUM(COALESCE(blocks_el_reward, 0)),0) as blocks_el_reward,
					COALESCE(SUM(COALESCE(sync_scheduled, 0)),0) as sync_scheduled,
					COALESCE(SUM(COALESCE(sync_executed, 0)),0) as sync_executed,
					COALESCE(SUM(COALESCE(sync_rewards, 0)),0) as sync_rewards,
					bool_or(slashed) as slashed,
					COALESCE(SUM(COALESCE(deposits_count, 0)),0) as deposits_count,
					COALESCE(SUM(COALESCE(deposits_amount, 0)),0) as deposits_amount,
					COALESCE(SUM(COALESCE(withdrawals_count, 0)),0) as withdrawals_count,
					COALESCE(SUM(COALESCE(withdrawals_amount, 0)),0) as withdrawals_amount,
					COALESCE(SUM(COALESCE(inclusion_delay_sum, 0)),0) as inclusion_delay_sum,
					COALESCE(SUM(COALESCE(sync_chance, 0)),0) as sync_chance,
					COALESCE(SUM(COALESCE(block_chance, 0)),0) as block_chance,
					COALESCE(SUM(COALESCE(attestations_scheduled, 0)),0) as attestations_scheduled,
					COALESCE(SUM(COALESCE(attestations_executed, 0)),0) as attestations_executed,
					COALESCE(SUM(COALESCE(attestation_head_executed, 0)),0) as attestation_head_executed,
					COALESCE(SUM(COALESCE(attestation_source_executed, 0)),0) as attestation_source_executed,
					COALESCE(SUM(COALESCE(attestation_target_executed, 0)),0) as attestation_target_executed,
					COALESCE(SUM(COALESCE(optimal_inclusion_delay_sum, 0)),0) as optimal_inclusion_delay_sum
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
				blocks_el_reward,
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
				optimal_inclusion_delay_sum
			)
			SELECT 
				$3,
				$1,
				(SELECT epoch FROM end_epoch),
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
				blocks_el_reward,
				sync_scheduled,
				sync_executed,
				sync_rewards,
				slashed,
				COALESCE(balance_start, 0),
				COALESCE(balance_end,0),
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
				optimal_inclusion_delay_sum
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
				blocks_el_reward = EXCLUDED.blocks_el_reward,
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
				epoch_end = EXCLUDED.epoch_end
	`, firstEpochOfDay, lastEpochOfDay, utils.EpochToTime(firstEpochOfDay))

	if err != nil {
		return errors.Wrap(err, "failed to insert daily aggregate")
	}

	return tx.Commit()
}

func (d *hourToDayAggregator) GetDayPartitionRange(epoch uint64) (time.Time, time.Time) {
	startOfPartition := epoch / (PartitionDayWidth * DayAggregateWidth) * PartitionDayWidth * DayAggregateWidth // inclusive
	endOfPartition := startOfPartition + PartitionDayWidth*DayAggregateWidth                                    // exclusive
	return utils.EpochToTime(startOfPartition), utils.EpochToTime(endOfPartition)
}

func (d *hourToDayAggregator) createDayPartition(dayFrom, dayTo time.Time) error {
	_, err := db.AlloyWriter.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS validator_dashboard_data_daily_%s_%s
		PARTITION OF validator_dashboard_data_daily
			FOR VALUES FROM ('%s') TO ('%s')
		`,
		dayToDDMMYYLabel(dayFrom), dayToDDMMYYLabel(dayTo), dayToDDMMYY(dayFrom), dayToDDMMYY(dayTo),
	))
	return err
}

func dayToDDMMYYLabel(day time.Time) string {
	return day.Format("020106")
}

func dayToDDMMYY(day time.Time) string {
	return day.Format("02-January-2006")
}
