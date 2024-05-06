package modules

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type dayUpAggregator struct {
	*dashboardData
	setupMutex        *sync.Mutex
	mutexes           map[string]*sync.Mutex
	rollingAggregator RollingAggregator
}

func newDayUpAggregator(d *dashboardData) *dayUpAggregator {
	return &dayUpAggregator{
		dashboardData: d,
		mutexes:       make(map[string]*sync.Mutex),
		setupMutex:    &sync.Mutex{},
		rollingAggregator: RollingAggregator{
			log: d.log,
			RollingAggregatorInt: &MultipleDaysRollingAggregatorImpl{
				log: d.log,
			},
		},
	}
}

func (d *dayUpAggregator) rolling7dAggregate(currentEpochHead uint64) error {
	return d.aggregateRollingXDays(7, "validator_dashboard_data_rolling_weekly", currentEpochHead)
}

func (d *dayUpAggregator) rolling30dAggregate(currentEpochHead uint64) error {
	return d.aggregateRollingXDays(30, "validator_dashboard_data_rolling_monthly", currentEpochHead)
}

func (d *dayUpAggregator) rolling90dAggregate(currentEpochHead uint64) error {
	return d.aggregateRollingXDays(90, "validator_dashboard_data_rolling_90d", currentEpochHead)
}

// for a given epoch intendedHeadEpoch, what epochs are needed for removal from the rolling tables
func (d *dayUpAggregator) getMissingRollingDayTailEpochs(intendedHeadEpoch uint64) ([]uint64, error) {
	week, err := d.getMissingRollingXDaysTailEpochs(7, intendedHeadEpoch, "validator_dashboard_data_rolling_weekly")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get missing 7d tail epochs")
	}
	month, err := d.getMissingRollingXDaysTailEpochs(30, intendedHeadEpoch, "validator_dashboard_data_rolling_monthly")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get missing 30d tail epochs")
	}
	ninety, err := d.getMissingRollingXDaysTailEpochs(90, intendedHeadEpoch, "validator_dashboard_data_rolling_90d")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get missing 90d tail epochs")
	}

	d.log.Infof("missing tail 7d: %v", week)
	d.log.Infof("missing tail 30d: %v", month)
	d.log.Infof("missing tail 90d: %v", ninety)

	return utils.Deduplicate(append(append(week, month...), ninety...)), nil
}

// for a given epoch intendedHeadEpoch, what epochs are needed to add to the rolling table (excluding the current epoch)
func (d *dayUpAggregator) getMissingRollingDayHeadEpochs(intendedHeadEpoch uint64) ([]uint64, error) {
	week, err := d.getMissingRollingXDaysHeadEpochs(7, intendedHeadEpoch-1, "validator_dashboard_data_rolling_weekly")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get missing 7d head epochs")
	}
	month, err := d.getMissingRollingXDaysHeadEpochs(30, intendedHeadEpoch-1, "validator_dashboard_data_rolling_monthly")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get missing 30d head epochs")
	}
	ninety, err := d.getMissingRollingXDaysHeadEpochs(90, intendedHeadEpoch-1, "validator_dashboard_data_rolling_90d")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get missing 90d head epochs")
	}

	d.log.Infof("missing head 7d: %v", week)
	d.log.Infof("missing head 30d: %v", month)
	d.log.Infof("missing head 90d: %v", ninety)

	return utils.Deduplicate(append(append(week, month...), ninety...)), nil
}

func (d *dayUpAggregator) getMissingRollingXDaysTailEpochs(days int, intendedHeadEpoch uint64, tableName string) ([]uint64, error) {
	return d.rollingAggregator.getMissingRollingTailEpochs(days, intendedHeadEpoch, tableName)
}

func (d *dayUpAggregator) getMissingRollingXDaysHeadEpochs(days int, intendedHeadEpoch uint64, tableName string) ([]uint64, error) {
	bounds, err := d.rollingAggregator.getCurrentRollingBounds(nil, tableName)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to get latest exported rolling %dd bounds", days))
		}
	}

	if bounds.EpochEnd <= 0 || intendedHeadEpoch-bounds.EpochEnd < d.epochWriter.getRetentionEpochDuration() {
		return nil, nil
	}
	return edb.GetMissingEpochsBetween(int64(bounds.EpochEnd), int64(intendedHeadEpoch)+1)
}

func (d *dayUpAggregator) aggregateRollingXDays(days int, tableName string, currentEpochHead uint64) error {
	d.setupMutex.Lock()
	if _, ok := d.mutexes[tableName]; !ok {
		d.mutexes[tableName] = &sync.Mutex{}
	}
	d.setupMutex.Unlock()

	d.mutexes[tableName].Lock()
	defer d.mutexes[tableName].Unlock()

	return d.rollingAggregator.Aggregate(days, tableName, currentEpochHead)
}

// -- rolling aggregate --

type MultipleDaysRollingAggregatorImpl struct {
	log ModuleLog
}

// returns both start_epochs
// the epoch_start from the the bootstrap tail
// and the epoch_start from the bootstrap head (epoch_start of epoch)
func (d *MultipleDaysRollingAggregatorImpl) getBootstrapBounds(epoch uint64, days uint64) (uint64, uint64) {
	currentStartBounds, _ := getDayAggregateBounds(epoch)
	xDayOldEpoch := int64(epoch - days*utils.EpochsPerDay())
	if xDayOldEpoch < 0 {
		xDayOldEpoch = 0
	}
	dayOldBoundsStart, _ := getDayAggregateBounds(uint64(xDayOldEpoch))
	return dayOldBoundsStart, currentStartBounds
}

// how many epochs can the rolling table be behind without bootstrapping
func (d *MultipleDaysRollingAggregatorImpl) getBootstrapOnEpochsBehind() uint64 {
	return utils.EpochsPerDay()
}

// bootstrap rolling 7d, 30d, 90d table from utc day table
func (d *MultipleDaysRollingAggregatorImpl) bootstrap(tx *sqlx.Tx, days int, tableName string) error {
	startTime := time.Now()
	defer func() {
		d.log.Infof("bootstrap rolling %vd took %v", days, time.Since(startTime))
	}()

	latestDayBounds, err := edb.GetLastExportedDay()
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to get latest exported day")
	}
	latestDay := latestDayBounds.Day

	tailStart, _ := d.getBootstrapBounds(latestDayBounds.EpochStart, uint64(days))
	xDayOldDay := utils.EpochToTime(tailStart).Format("2006-01-02")

	d.log.Infof("agg %dd | latestDay: %v, oldDay: %v", days, latestDay, xDayOldDay)

	_, err = tx.Exec(fmt.Sprintf(`TRUNCATE %s`, tableName))
	if err != nil {
		return errors.Wrap(err, "failed to delete old rolling aggregate")
	}

	_, err = tx.Exec(fmt.Sprintf(`
		WITH
			epoch_starts as (
				SELECT min(epoch_start) as epoch_start FROM validator_dashboard_data_daily WHERE day = $1 LIMIT 1
			),
			epoch_ends as (
				SELECT max(epoch_end) as epoch_end FROM validator_dashboard_data_daily WHERE day = $2 LIMIT 1
			),
			balance_starts as (
				SELECT validator_index, balance_start, epoch_start FROM validator_dashboard_data_daily WHERE day = $1
			),
			balance_ends as (
				SELECT validator_index, balance_end, epoch_end FROM validator_dashboard_data_daily WHERE day = $2
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
					SUM(block_chance) as block_chance,
					SUM(sync_chance) as sync_chance,
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
				FROM validator_dashboard_data_daily
				WHERE day >= $1 AND day <= $2
				GROUP BY validator_index
			)
			INSERT INTO %s (
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
				block_chance,
				sync_chance,
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
				(SELECT epoch_start FROM epoch_starts),
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
				block_chance,
				sync_chance,
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
	`, tableName), xDayOldDay, latestDay)

	if err != nil {
		return errors.Wrap(err, "failed to insert rolling aggregate")
	}

	return nil
}
