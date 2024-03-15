package modules

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
)

type dayToWeeklyAggregator struct {
	*dashboardData
	mutex *sync.Mutex
}

func newDayToWeeklyAggregator(d *dashboardData) *dayToWeeklyAggregator {
	return &dayToWeeklyAggregator{
		dashboardData: d,
		mutex:         &sync.Mutex{},
	}
}

func (d *dayToWeeklyAggregator) rolling7dAggregate() {
	d.rollingXdAggregate(7, "validator_dashboard_data_rolling_weekly")
}

func (d *dayToWeeklyAggregator) rolling31dAggregate() {
	d.rollingXdAggregate(31, "validator_dashboard_data_rolling_monthly")
}

func (d *dayToWeeklyAggregator) rollingXdAggregate(days int, tableName string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	startTime := time.Now()
	defer func() {
		d.log.Infof("rolling %vd aggregate took %v", days, time.Since(startTime))
	}()

	latestDayBounds, err := edb.GetLastExportedDay()
	if err != nil && err != sql.ErrNoRows {
		d.log.Error(err, "failed to get latest dashboard epoch", 0)
		return
	}
	latestDay := latestDayBounds.Day

	weekOldDay, err := edb.GetXDayOldDay(days)
	if err != nil {
		d.log.Error(err, fmt.Sprintf("failed to get %dd old dashboard epoch", days), 0)
		return
	}

	d.log.Infof("latestDay: %v, oldDay: %v", latestDay, weekOldDay)

	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		d.log.Error(err, "failed to start transaction", 0)
		return
	}
	defer utils.Rollback(tx)

	_, err = tx.Exec(fmt.Sprintf(`TRUNCATE %s`, tableName))
	if err != nil {
		d.log.Error(err, fmt.Sprintf("failed to delete old rolling %dd aggregate", days), 0)
		return
	}

	_, err = tx.Exec(fmt.Sprintf(`
		WITH
			balance_starts as (
				SELECT validator_index, balance_start FROM validator_dashboard_data_daily WHERE day = $1
			),
			balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_daily WHERE day = $2
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
				FROM validator_dashboard_data_daily
				WHERE day >= $1 AND day <= $2
				GROUP BY validator_index
			)
			INSERT INTO %s (
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
	`, tableName), weekOldDay, latestDay)

	if err != nil {
		d.log.Error(err, fmt.Sprintf("failed to insert rolling %dd aggregate", days), 0)
		return
	}

	err = tx.Commit()
	if err != nil {
		d.log.Error(err, "failed to commit transaction", 0)
		return
	}
}
