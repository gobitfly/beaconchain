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

	currentExportedEpoch, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get latest dashboard epoch")
	}

	if currentExportedEpoch < lastTotalExported.EpochEnd {
		return errors.Wrap(err, "total export nothing to do, currentEpoch <= lastTotalExported.EpochEnd")
	}

	gaps, err := edb.GetDashboardEpochGapsBetween(currentExportedEpoch, int64(lastTotalExported.EpochEnd))
	if err != nil {
		return errors.Wrap(err, "failed to get dashboard epoch gaps")
	}

	if len(gaps) > 0 {
		return fmt.Errorf("gaps in dashboard epoch, skipping total for now: %v", gaps) // sanity, this should never happen
	}

	err = d.aggregateAndAddToTotal(lastTotalExported.EpochEnd, currentExportedEpoch)
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

	d.log.Infof("aggregating total (from: %d) up to %d", epochStart, epochEnd)

	_, err = tx.Exec(`
		WITH
			end_epoch as (
				SELECT max(epoch) as epoch FROM validator_dashboard_data_epoch where epoch <= $2 AND epoch >= $1
			),
			--balance_starts as ( -- we dont need this for updating, only for bootstraping
			--	SELECT validator_index, balance_start FROM validator_dashboard_data_epoch WHERE epoch = $1
			--),
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
				$1,
				(SELECT epoch + 1 FROM end_epoch) as epoch, -- exclusive
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
				32e9, --balance_start,
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
			--LEFT JOIN balance_starts ON aggregate.validator_index = balance_starts.validator_index
			LEFT JOIN balance_ends ON aggregate.validator_index = balance_ends.validator_index
			ON CONFLICT (validator_index) DO UPDATE SET
				attestations_source_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_source_reward, 0) + COALESCE(EXCLUDED.attestations_source_reward, 0),
				attestations_target_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_target_reward, 0) + COALESCE(EXCLUDED.attestations_target_reward, 0),
				attestations_head_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_head_reward, 0) + COALESCE(EXCLUDED.attestations_head_reward, 0),
				attestations_inactivity_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_inactivity_reward, 0) + COALESCE(EXCLUDED.attestations_inactivity_reward, 0),
				attestations_inclusion_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_inclusion_reward, 0) + COALESCE(EXCLUDED.attestations_inclusion_reward, 0),
				attestations_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_reward, 0) + COALESCE(EXCLUDED.attestations_reward, 0),
				attestations_ideal_source_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_ideal_source_reward, 0) + COALESCE(EXCLUDED.attestations_ideal_source_reward, 0),
				attestations_ideal_target_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_ideal_target_reward, 0) + COALESCE(EXCLUDED.attestations_ideal_target_reward, 0),
				attestations_ideal_head_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_ideal_head_reward, 0) + COALESCE(EXCLUDED.attestations_ideal_head_reward, 0),
				attestations_ideal_inactivity_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_ideal_inactivity_reward, 0) + COALESCE(EXCLUDED.attestations_ideal_inactivity_reward, 0),
				attestations_ideal_inclusion_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_ideal_inclusion_reward, 0) + COALESCE(EXCLUDED.attestations_ideal_inclusion_reward, 0),
				attestations_ideal_reward = COALESCE(validator_dashboard_data_rolling_total.attestations_ideal_reward, 0) + COALESCE(EXCLUDED.attestations_ideal_reward, 0),
				blocks_scheduled = COALESCE(validator_dashboard_data_rolling_total.blocks_scheduled, 0) + COALESCE(EXCLUDED.blocks_scheduled, 0),
				blocks_proposed = COALESCE(validator_dashboard_data_rolling_total.blocks_proposed, 0) + COALESCE(EXCLUDED.blocks_proposed, 0),
				blocks_cl_reward = COALESCE(validator_dashboard_data_rolling_total.blocks_cl_reward, 0) + COALESCE(EXCLUDED.blocks_cl_reward, 0),
				sync_scheduled = COALESCE(validator_dashboard_data_rolling_total.sync_scheduled, 0) + COALESCE(EXCLUDED.sync_scheduled, 0),
				sync_executed = COALESCE(validator_dashboard_data_rolling_total.sync_executed, 0) + COALESCE(EXCLUDED.sync_executed, 0),
				sync_rewards = COALESCE(validator_dashboard_data_rolling_total.sync_rewards, 0) + COALESCE(EXCLUDED.sync_rewards, 0),
				slashed = EXCLUDED.slashed,
				balance_end = EXCLUDED.balance_end,
				deposits_count = COALESCE(validator_dashboard_data_rolling_total.deposits_count, 0) + COALESCE(EXCLUDED.deposits_count, 0),
				deposits_amount = COALESCE(validator_dashboard_data_rolling_total.deposits_amount, 0) + COALESCE(EXCLUDED.deposits_amount, 0),
				withdrawals_count = COALESCE(validator_dashboard_data_rolling_total.withdrawals_count, 0) + COALESCE(EXCLUDED.withdrawals_count, 0),
				withdrawals_amount = COALESCE(validator_dashboard_data_rolling_total.withdrawals_amount, 0) + COALESCE(EXCLUDED.withdrawals_amount, 0),
				inclusion_delay_sum = COALESCE(validator_dashboard_data_rolling_total.inclusion_delay_sum, 0) + COALESCE(EXCLUDED.inclusion_delay_sum, 0),
				sync_chance = COALESCE(validator_dashboard_data_rolling_total.sync_chance, 0) + COALESCE(EXCLUDED.sync_chance, 0),
				block_chance = COALESCE(validator_dashboard_data_rolling_total.block_chance, 0) + COALESCE(EXCLUDED.block_chance, 0),
				attestations_scheduled = COALESCE(validator_dashboard_data_rolling_total.attestations_scheduled, 0) + COALESCE(EXCLUDED.attestations_scheduled, 0),
				attestations_executed = COALESCE(validator_dashboard_data_rolling_total.attestations_executed, 0) + COALESCE(EXCLUDED.attestations_executed, 0),
				attestation_head_executed = COALESCE(validator_dashboard_data_rolling_total.attestation_head_executed, 0) + COALESCE(EXCLUDED.attestation_head_executed, 0),
				attestation_source_executed = COALESCE(validator_dashboard_data_rolling_total.attestation_source_executed, 0) + COALESCE(EXCLUDED.attestation_source_executed, 0),
				attestation_target_executed = COALESCE(validator_dashboard_data_rolling_total.attestation_target_executed, 0) + COALESCE(EXCLUDED.attestation_target_executed, 0),
				optimal_inclusion_delay_sum = COALESCE(validator_dashboard_data_rolling_total.optimal_inclusion_delay_sum, 0) + COALESCE(EXCLUDED.optimal_inclusion_delay_sum, 0),
				slasher_reward = COALESCE(validator_dashboard_data_rolling_total.slasher_reward, 0) + COALESCE(EXCLUDED.slasher_reward, 0),
				slashed_by = EXCLUDED.slashed_by,
				slashed_violation = EXCLUDED.slashed_violation,
				last_executed_duty_epoch = EXCLUDED.last_executed_duty_epoch,
				epoch_end = EXCLUDED.epoch_end
	`, epochStart, epochEnd)

	if err != nil {
		return err
	}

	return tx.Commit()
}
