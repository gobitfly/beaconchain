package modules

import (
	"database/sql"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/pkg/errors"
)

type epochToTotalAggregator struct {
	*dashboardData
	mutex *sync.Mutex
}

func newEpochToTotalAggregator(d *dashboardData) *epochToTotalAggregator {
	return &epochToTotalAggregator{
		dashboardData: d,
		mutex:         &sync.Mutex{},
	}
}

// Assumes no gaps in epochs
func (d *epochToTotalAggregator) aggregateTotal() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	startTime := time.Now()
	defer func() {
		d.log.Infof("aggregate total took %v", time.Since(startTime))
	}()

	lastTotalExported, err := edb.GetLastExportedTotalEpoch()
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to get last exported total epoch")
	}

	currentEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get latest dashboard epoch")
	}

	if currentEpoch <= lastTotalExported.EpochEnd {
		return errors.Wrap(err, "total export nothing to do, currentEpoch <= lastTotalExported.EpochEnd")
	}

	err = d.aggregateAndAddToTotal(lastTotalExported.EpochEnd+1, currentEpoch)
	if err != nil {
		return errors.Wrap(err, "failed to aggregate total")
	}

	return nil
}

// both inclusive
func (d *epochToTotalAggregator) aggregateAndAddToTotal(epochStart, epochEnd uint64) error {
	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	d.log.Infof("aggregating totalendEpoch: %d", epochEnd)

	_, err = tx.Exec(`
		WITH
			end_epoch as (
				SELECT max(epoch) as epoch FROM validator_dashboard_data_epoch where epoch <= $2 AND epoch >= $1
			),
			balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_epoch WHERE epoch = $1
			),
			balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_epoch WHERE epoch = (SELECT epoch FROM end_epoch)
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
					SUM(blocks_el_reward) as blocks_el_reward,
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
					SUM(optimal_inclusion_delay_sum) as optimal_inclusion_delay_sum
				FROM validator_dashboard_data_epoch
				WHERE epoch >= $1 AND epoch <= $2
				GROUP BY validator_index
			)
			INSERT INTO validator_dashboard_data_rolling_total (
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
				$1,
				(SELECT epoch FROM end_epoch) as epoch,
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
			FROM aggregate
			LEFT JOIN balance_starts ON aggregate.validator_index = balance_starts.validator_index
			LEFT JOIN balance_ends ON aggregate.validator_index = balance_ends.validator_index
			ON CONFLICT (validator_index) DO UPDATE SET
				attestations_source_reward = validator_dashboard_data_rolling_total.attestations_source_reward + EXCLUDED.attestations_source_reward,
				attestations_target_reward = validator_dashboard_data_rolling_total.attestations_target_reward + EXCLUDED.attestations_target_reward,
				attestations_head_reward = validator_dashboard_data_rolling_total.attestations_head_reward + EXCLUDED.attestations_head_reward,
				attestations_inactivity_reward = validator_dashboard_data_rolling_total.attestations_inactivity_reward + EXCLUDED.attestations_inactivity_reward,
				attestations_inclusion_reward = validator_dashboard_data_rolling_total.attestations_inclusion_reward + EXCLUDED.attestations_inclusion_reward,
				attestations_reward = validator_dashboard_data_rolling_total.attestations_reward + EXCLUDED.attestations_reward,
				attestations_ideal_source_reward = validator_dashboard_data_rolling_total.attestations_ideal_source_reward + EXCLUDED.attestations_ideal_source_reward,
				attestations_ideal_target_reward = validator_dashboard_data_rolling_total.attestations_ideal_target_reward + EXCLUDED.attestations_ideal_target_reward,
				attestations_ideal_head_reward = validator_dashboard_data_rolling_total.attestations_ideal_head_reward + EXCLUDED.attestations_ideal_head_reward,
				attestations_ideal_inactivity_reward = validator_dashboard_data_rolling_total.attestations_ideal_inactivity_reward + EXCLUDED.attestations_ideal_inactivity_reward,
				attestations_ideal_inclusion_reward = validator_dashboard_data_rolling_total.attestations_ideal_inclusion_reward + EXCLUDED.attestations_ideal_inclusion_reward,
				attestations_ideal_reward = validator_dashboard_data_rolling_total.attestations_ideal_reward + EXCLUDED.attestations_ideal_reward,
				blocks_scheduled = validator_dashboard_data_rolling_total.blocks_scheduled + EXCLUDED.blocks_scheduled,
				blocks_proposed = validator_dashboard_data_rolling_total.blocks_proposed + EXCLUDED.blocks_proposed,
				blocks_cl_reward = validator_dashboard_data_rolling_total.blocks_cl_reward + EXCLUDED.blocks_cl_reward,
				blocks_el_reward = validator_dashboard_data_rolling_total.blocks_el_reward + EXCLUDED.blocks_el_reward,
				sync_scheduled = validator_dashboard_data_rolling_total.sync_scheduled + EXCLUDED.sync_scheduled,
				sync_executed = validator_dashboard_data_rolling_total.sync_executed + EXCLUDED.sync_executed,
				sync_rewards = validator_dashboard_data_rolling_total.sync_rewards + EXCLUDED.sync_rewards,
				slashed = EXCLUDED.slashed,
				balance_end = EXCLUDED.balance_end,
				deposits_count = validator_dashboard_data_rolling_total.deposits_count + EXCLUDED.deposits_count,
				deposits_amount = validator_dashboard_data_rolling_total.deposits_amount + EXCLUDED.deposits_amount,
				withdrawals_count = validator_dashboard_data_rolling_total.withdrawals_count + EXCLUDED.withdrawals_count,
				withdrawals_amount = validator_dashboard_data_rolling_total.withdrawals_amount + EXCLUDED.withdrawals_amount,
				inclusion_delay_sum = validator_dashboard_data_rolling_total.inclusion_delay_sum + EXCLUDED.inclusion_delay_sum,
				sync_chance = validator_dashboard_data_rolling_total.sync_chance + EXCLUDED.sync_chance,
				block_chance = validator_dashboard_data_rolling_total.block_chance + EXCLUDED.block_chance,
				attestations_scheduled = validator_dashboard_data_rolling_total.attestations_scheduled + EXCLUDED.attestations_scheduled,
				attestations_executed = validator_dashboard_data_rolling_total.attestations_executed + EXCLUDED.attestations_executed,
				attestation_head_executed = validator_dashboard_data_rolling_total.attestation_head_executed + EXCLUDED.attestation_head_executed,
				attestation_source_executed = validator_dashboard_data_rolling_total.attestation_source_executed + EXCLUDED.attestation_source_executed,
				attestation_target_executed = validator_dashboard_data_rolling_total.attestation_target_executed + EXCLUDED.attestation_target_executed,
				optimal_inclusion_delay_sum = validator_dashboard_data_rolling_total.optimal_inclusion_delay_sum + EXCLUDED.optimal_inclusion_delay_sum,
				epoch_end = EXCLUDED.epoch_end
	`, epochStart, epochEnd)

	if err != nil {
		return err
	}

	return tx.Commit()
}
