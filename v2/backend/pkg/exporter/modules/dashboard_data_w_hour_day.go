package modules

import (
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
)

type hourToDayAggregator struct {
	*dashboardData
	mutex *sync.Mutex
}

func newHourToDayAggregator(d *dashboardData) *hourToDayAggregator {
	return &hourToDayAggregator{
		dashboardData: d,
		mutex:         &sync.Mutex{},
	}
}

func (d *hourToDayAggregator) rolling24hAggregate() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	startTime := time.Now()
	defer func() {
		d.log.Infof("rolling 24h aggregate took %v", time.Since(startTime))
	}()

	latestHourlyEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		d.log.Error(err, "failed to get latest dashboard epoch", 0)
		return
	}

	dayOldHourlyEpoch, err := edb.Get24hOldHourlyEpoch()
	if err != nil {
		d.log.Error(err, "failed to get 24h old dashboard epoch", 0)
		return
	}

	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		d.log.Error(err, "failed to start transaction", 0)
		return
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(`DELETE FROM validator_dashboard_data_rolling_daily`)
	if err != nil {
		d.log.Error(err, "failed to delete old rolling 24h aggregate", 0)
	}

	/*
			attestations_scheduled smallint,
		    attestations_executed smallint,
		    attestation_head_executed smallint,
		    attestation_source_executed smallint,
		    attestation_target_executed smallint,
	*/

	_, err = tx.Exec(`
		WITH
			balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_hourly WHERE epoch_start = $2
			),
			balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_hourly WHERE epoch_start = $1
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
					COALESCE(SUM(COALESCE(attestation_target_executed, 0)),0) as attestation_target_executed
				FROM validator_dashboard_data_hourly
				WHERE epoch_start >= $2 AND epoch_start <= $1
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
				attestation_target_executed
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
				attestation_target_executed
			FROM aggregate
			LEFT JOIN balance_starts ON aggregate.validator_index = balance_starts.validator_index
			LEFT JOIN balance_ends ON aggregate.validator_index = balance_ends.validator_index
	`, latestHourlyEpoch, dayOldHourlyEpoch)

	if err != nil {
		d.log.Error(err, "failed to insert rolling 24h aggregate", 0)
	}

	err = tx.Commit()
	if err != nil {
		d.log.Error(err, "failed to commit transaction", 0)
	}

	minHourlyDBEpoch, err := edb.GetMinOldHourlyEpoch()
	if err != nil {
		d.log.Error(err, "failed to get min old hourly epoch", 0)
		return
	}

	// TODO clear hourly
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

/*
WITH
			first_entry_of_day as (
				SELECT min(ts) as ts FROM validator_dashboard_data_hourly WHERE DATE(ts) = $1
			),
			last_entry_of_day as (
				SELECT max(ts) as ts FROM validator_dashboard_data_hourly WHERE DATE(ts) = $1
			),
			balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_hourly WHERE ts = (SELECT ts FROM first_entry_of_day)
			),
			balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_epoch WHERE ts = (SELECT ts FROM last_entry_of_day)
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
					COALESCE(SUM(COALESCE(withdrawals_amount, 0)),0) as withdrawals_amount
				FROM validator_dashboard_data_hourly
				WHERE ts >= (SELECT ts FROM first_entry_of_day) AND ts <= (SELECT ts FROM last_entry_of_day)
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
				withdrawals_amount
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
				withdrawals_amount
			FROM aggregate
			LEFT JOIN balance_starts ON aggregate.validator_index = balance_starts.validator_index
			LEFT JOIN balance_ends ON aggregate.validator_index = balance_ends.validator_index
*/
