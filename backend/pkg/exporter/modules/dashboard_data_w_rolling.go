package modules

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

/**
This file handles the logic for rolling aggregation for 24h, 7d, 31d and 90d but not total, see dashboard_data_w_2_epoch_total.go for that.
The way this works is by adding new epochs to the rolling table and removing the old epochs at the end so that the time duration of rolling stays constant.

If the rolling tables fall out of sync due to long offline time or initial sync, the tables are bootstrapped. This bootstrap method must be provided,
7d, 31d, 90d use a bootstrap from the utc_days table to get started and 24h the hourly table.
*/

type RollingAggregator struct {
	RollingAggregatorInt
	log ModuleLog
}

type RollingAggregatorInt interface {
	bootstrap(tx *sqlx.Tx, days int, tableName string) error

	// return the number of epochs the current head is ahead of the bootstrap table
	// this is the same number the tail of the bootstrap table is below the target tail end
	// this is useful to know how many recent tail epochs you need to fetch
	bootstrapTableToHeadOffset(currentHead uint64) (int64, error)

	// get the threshold on how many epochs you can be behind without bootstrap or at which distance there will be a bootstrap
	getBootstrapOnEpochsBehind() uint64

	// gets the aggegate bounds for a given epoch in the bootstrap table. Is useful if you want to know what aggregate an epoch is part of
	getBootstrapBounds(epoch uint64, days uint64) (uint64, uint64)
}

// Returns the epoch range of a current exported rolling table
// Ideally the epoch range has an exact with of 24h, 7d, 31d or 90d BUT it can be more after bootstrap or less if there are less epochs on the network than the rolling width
func (d *RollingAggregator) getCurrentRollingBounds(tx *sqlx.Tx, tableName string) (edb.EpochBounds, error) {
	var bounds edb.EpochBounds
	var err error
	if tx == nil {
		err = db.AlloyWriter.Get(&bounds, fmt.Sprintf(`SELECT epoch_start as epoch_start, epoch_end as epoch_end FROM %s LIMIT 1`, tableName))
	} else {
		err = tx.Get(&bounds, fmt.Sprintf(`SELECT epoch_start as epoch_start, epoch_end as epoch_end FROM %s LIMIT 1`, tableName))
	}
	return bounds, err
}

// returns the tail epochs (those must be removed from rolling) for a given intendedHeadEpoch for a given rolling table
// fE a tail epoch for rolling 1 day aggregation (225 epochs) for boundsStart 0 (start epoch of last rolling export) and intendedHeadEpoch 227 on ethereum would correspond to a tail range of 0 - 1
// meaning epoch [0,1] must be removed from the rolling table if you want to add epoch 227
// arguments returned are inclusive
func (d *RollingAggregator) getTailBoundsXDays(days int, boundsStart uint64, intendedHeadEpoch uint64, offset int64) (int64, int64) {
	aggTailEpochStart := int64(boundsStart)
	aggTailEpochEnd := int64(intendedHeadEpoch - utils.EpochsPerDay()*uint64(days))
	d.log.Infof("tail bounds for %dd: %d - %d", days, aggTailEpochStart, aggTailEpochEnd)

	// limit to last offset epochs as the rest will not be relevant after bootstrapping
	if aggTailEpochEnd-aggTailEpochStart > offset { //int64(getHourAggregateWidth()) {
		aggTailEpochStart = aggTailEpochEnd - offset
	}
	return aggTailEpochStart, aggTailEpochEnd
}

func (d *RollingAggregator) Aggregate(days int, tableName string) error {
	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer utils.Rollback(tx)

	bootstrap := false

	// get epoch boundaries for current stored rolling 24h
	bounds, err := d.getCurrentRollingBounds(tx, tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			bootstrap = true
			log.Infof("bootstraping rolling %dd due to empty table", days)
		} else {
			return errors.Wrap(err, "failed to get current rolling bounds")
		}
	}

	// get current stored epoch table head
	currentEpochHead, err := edb.GetLatestDashboardEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get last exported epoch")
	}

	// if current stored rolling 24 is far behind, bootstrap again
	// in this case far means more than what we aggregate in the hour table, meaning a bootstrap
	// will get faster to head then re-exporting amount of getHourAggregateWidth() old epochs
	if currentEpochHead+1-bounds.EpochEnd >= d.getBootstrapOnEpochsBehind() { // EpochEnd is excl so +1 to get the inclusive epoch number
		d.log.Infof("currentEpochHead: %d, bounds.EpochEnd: %d, getBootstrapOnEpochsBehind(): %d, leftsum: %d", currentEpochHead, bounds.EpochEnd, d.getBootstrapOnEpochsBehind(), currentEpochHead+1-bounds.EpochEnd)
		bootstrap = true
	}

	if bootstrap {
		d.log.Infof("rolling %dd bootstraping starting", days)

		err = d.bootstrap(tx, days, tableName)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to bootstrap rolling %dd aggregate", days))
		}
		bounds, err = d.getCurrentRollingBounds(tx, tableName)
		if err != nil {
			return errors.Wrap(err, "failed to get current rolling bounds")
		}

		d.log.Infof("rolling %dd bootstraping finished", days)
	}

	if currentEpochHead == bounds.EpochEnd-1 && bounds.EpochEnd-utils.EpochsPerDay() == bounds.EpochStart { // todo check bounds ok?
		log.Infof("rolling %dd is up to date, nothing to do", days)
		err = tx.Commit()
		if err != nil {
			return errors.Wrap(err, "failed to commit transaction")
		}
		return nil
	}

	if !bootstrap && bounds.EpochEnd-bounds.EpochStart != utils.EpochsPerDay()*uint64(days) { // todo need EpochEnd is excl so -1 to get the inclusive epoch number?
		log.Warnf("rolling %dd boundaries are out of bounds (%d-%d, %d), this is expected after bootstrap, but not after that. Keep an eye on it", days, bounds.EpochStart, bounds.EpochEnd, bounds.EpochEnd-bounds.EpochStart-1)
	}

	// how many epochs will the epochs table be ahead of the aggregated table
	bootstrapOffset, err := d.bootstrapTableToHeadOffset(currentEpochHead)
	if err != nil {
		return errors.Wrap(err, "failed to get bootstrap offset")
	}
	d.log.Infof("bootstrap Offset for rolling %dd: %d", days, bootstrapOffset)

	// bounds for what to aggregate and add to the head of the rolling 24h
	aggHeadEpochStart := bounds.EpochEnd
	aggHeadEpochEnd := currentEpochHead

	// bounds for what to aggregate and remove from the tail of the rolling 24h
	aggTailEpochStart, aggTailEpochEnd := d.getTailBoundsXDays(days, bounds.EpochStart, currentEpochHead, bootstrapOffset)

	d.log.Infof("rolling %dd epochs: %d - %d, %d - %d", days, aggHeadEpochStart, aggHeadEpochEnd, aggTailEpochStart, aggTailEpochEnd)
	// sanity check if all tail epochs are present in db
	missing, err := getMissingEpochsBetween(aggTailEpochStart, aggTailEpochEnd)
	if err != nil {
		return errors.Wrap(err, "failed to get missing tail epochs")
	}
	if len(missing) > 0 {
		return errors.New(fmt.Sprintf("missing epochs in db for rolling %dd tail: %v", days, missing))
	}

	// add head and fix/remove from tail
	err = d.aggregateRolling(tx, tableName, aggHeadEpochStart, aggHeadEpochEnd, aggTailEpochStart, aggTailEpochEnd)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to aggregate rolling %dd", days))
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (d *RollingAggregator) getMissingRollingTailEpochs(days int, intendedHeadEpoch uint64, tableName string) ([]uint64, error) {
	bounds, err := d.getCurrentRollingBounds(nil, tableName)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to get latest exported rolling %dd bounds", days))
		}
	}

	offset, err := d.bootstrapTableToHeadOffset(intendedHeadEpoch)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bootstrap offset")
	}

	needsBootstrap := int64(intendedHeadEpoch-bounds.EpochEnd) >= int64(d.getBootstrapOnEpochsBehind())

	d.log.Infof("bootstrap Offset for rolling %dd: %d. Needs bootstrap: %v", days, offset, needsBootstrap)
	// if rolling table is empty / not bootstrapped yet or needs a bootstrap assume bounds of what the would be after a bootstrap
	if (bounds.EpochEnd == 0 && bounds.EpochStart == 0) || needsBootstrap {
		startBound, endBounds := d.getBootstrapBounds(intendedHeadEpoch, uint64(days))
		bounds.EpochStart = startBound
		bounds.EpochEnd = endBounds
		d.log.Infof("bootstrap bounds for rolling %dd: %d - %d | current Head (excl): %v", days, bounds.EpochStart, bounds.EpochEnd, intendedHeadEpoch+1)
	}

	aggTailEpochStart, aggTailEpochEnd := d.getTailBoundsXDays(days, bounds.EpochStart, intendedHeadEpoch, offset)

	return getMissingEpochsBetween(aggTailEpochStart, aggTailEpochEnd)
}

// Adds the new epochs (headEpochStart to headEpochEnd) to the rolling table and removes the old ones (tailEpochStart to tailEpochEnd)
// all arguments are inclusive
func (d *RollingAggregator) aggregateRolling(tx *sqlx.Tx, tableName string, headEpochStart, headEpochEnd uint64, tailEpochStart, tailEpochEnd int64) error {
	d.log.Infof("aggregating rolling %s epochs: %d - %d, %d - %d", tableName, headEpochStart, headEpochEnd, tailEpochStart, tailEpochEnd)
	defer d.log.Infof("aggregated rolling %s epochs: %d - %d, %d - %d", tableName, headEpochStart, headEpochEnd, tailEpochStart, tailEpochEnd)

	if headEpochEnd >= headEpochStart {
		err := d.addToRolling(tx, tableName, headEpochStart, headEpochEnd)
		if err != nil {
			return errors.Wrap(err, "failed to add epochs to rolling")
		}
	}
	if tailEpochEnd >= tailEpochStart {
		err := d.removeFromRolling(tx, tableName, tailEpochStart, tailEpochEnd)
		if err != nil {
			return errors.Wrap(err, "failed to remove epochs from rolling")
		}
	}
	return nil
}

func (d *RollingAggregator) addToRolling(tx *sqlx.Tx, tableName string, startEpoch, endEpoch uint64) error {
	startTime := time.Now()
	d.log.Infof("add to rolling %s epochs: %d - %d", tableName, startEpoch, endEpoch)
	defer func() {
		d.log.Infof("added to rolling %s took %v", tableName, time.Since(startTime))
	}()

	result, err := tx.Exec(fmt.Sprintf(`
		WITH
			head_balance_ends as (
				SELECT validator_index, balance_end FROM validator_dashboard_data_epoch WHERE epoch = $2
			),
			aggregate_head as (
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
			),
			result as (
				SELECT
					$2 + 1 as epoch_end, -- exclusive
					aggregate_head.validator_index as validator_index,
					COALESCE(aggregate_head.attestations_source_reward, 0) as attestations_source_reward,
					COALESCE(aggregate_head.attestations_target_reward, 0) as attestations_target_reward,
					COALESCE(aggregate_head.attestations_head_reward, 0) as attestations_head_reward,
					COALESCE(aggregate_head.attestations_inactivity_reward, 0) as attestations_inactivity_reward,
					COALESCE(aggregate_head.attestations_inclusion_reward, 0) as attestations_inclusion_reward,
					COALESCE(aggregate_head.attestations_reward, 0) as attestations_reward,
					COALESCE(aggregate_head.attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
					COALESCE(aggregate_head.attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
					COALESCE(aggregate_head.attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
					COALESCE(aggregate_head.attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
					COALESCE(aggregate_head.attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
					COALESCE(aggregate_head.attestations_ideal_reward, 0) as attestations_ideal_reward,
					COALESCE(aggregate_head.blocks_scheduled, 0) as blocks_scheduled,
					COALESCE(aggregate_head.blocks_proposed, 0) as blocks_proposed,
					COALESCE(aggregate_head.blocks_cl_reward, 0) as blocks_cl_reward,
					COALESCE(aggregate_head.sync_scheduled, 0) as sync_scheduled,
					COALESCE(aggregate_head.sync_executed, 0) as sync_executed,
					COALESCE(aggregate_head.sync_rewards, 0) as sync_rewards,
					aggregate_head.slashed,
					balance_end,
					COALESCE(aggregate_head.deposits_count, 0) as deposits_count,
					COALESCE(aggregate_head.deposits_amount, 0) as deposits_amount,
					COALESCE(aggregate_head.withdrawals_count, 0) as withdrawals_count,
					COALESCE(aggregate_head.withdrawals_amount, 0) as withdrawals_amount,
					COALESCE(aggregate_head.inclusion_delay_sum, 0) as inclusion_delay_sum,
					COALESCE(aggregate_head.sync_chance, 0) as sync_chance,
					COALESCE(aggregate_head.block_chance, 0) as block_chance,
					COALESCE(aggregate_head.attestations_scheduled, 0) as attestations_scheduled,
					COALESCE(aggregate_head.attestations_executed, 0) as attestations_executed,
					COALESCE(aggregate_head.attestation_head_executed, 0) as attestation_head_executed,
					COALESCE(aggregate_head.attestation_source_executed, 0) as attestation_source_executed,
					COALESCE(aggregate_head.attestation_target_executed, 0) as attestation_target_executed,
					COALESCE(aggregate_head.optimal_inclusion_delay_sum, 0) as optimal_inclusion_delay_sum,
					COALESCE(aggregate_head.slasher_reward, 0) as slasher_reward,
					aggregate_head.slashed_by,
					aggregate_head.slashed_violation,
					aggregate_head.last_executed_duty_epoch
				FROM aggregate_head  
				LEFT JOIN head_balance_ends ON aggregate_head.validator_index = head_balance_ends.validator_index
			)
			UPDATE %s AS v SET
					attestations_source_reward = COALESCE(v.attestations_source_reward, 0) + result.attestations_source_reward,
					attestations_target_reward = COALESCE(v.attestations_target_reward, 0) + result.attestations_target_reward,
					attestations_head_reward = COALESCE(v.attestations_head_reward, 0) + result.attestations_head_reward,
					attestations_inactivity_reward = COALESCE(v.attestations_inactivity_reward, 0) + result.attestations_inactivity_reward,
					attestations_inclusion_reward = COALESCE(v.attestations_inclusion_reward, 0) + result.attestations_inclusion_reward,
					attestations_reward = COALESCE(v.attestations_reward, 0) + result.attestations_reward,
					attestations_ideal_source_reward = COALESCE(v.attestations_ideal_source_reward, 0) + result.attestations_ideal_source_reward,
					attestations_ideal_target_reward = COALESCE(v.attestations_ideal_target_reward, 0) + result.attestations_ideal_target_reward,
					attestations_ideal_head_reward = COALESCE(v.attestations_ideal_head_reward, 0) + result.attestations_ideal_head_reward,
					attestations_ideal_inactivity_reward = COALESCE(v.attestations_ideal_inactivity_reward, 0) + result.attestations_ideal_inactivity_reward,
					attestations_ideal_inclusion_reward = COALESCE(v.attestations_ideal_inclusion_reward, 0) + result.attestations_ideal_inclusion_reward,
					attestations_ideal_reward = COALESCE(v.attestations_ideal_reward, 0) + result.attestations_ideal_reward,
					blocks_scheduled = COALESCE(v.blocks_scheduled, 0) + result.blocks_scheduled,
					blocks_proposed = COALESCE(v.blocks_proposed, 0) + result.blocks_proposed,
					blocks_cl_reward = COALESCE(v.blocks_cl_reward, 0) + result.blocks_cl_reward,
					sync_scheduled = COALESCE(v.sync_scheduled, 0) + result.sync_scheduled,
					sync_executed = COALESCE(v.sync_executed, 0) + result.sync_executed,
					sync_rewards = COALESCE(v.sync_rewards, 0) + result.sync_rewards,
					slashed = COALESCE(result.slashed, v.slashed),
					balance_end = COALESCE(result.balance_end, v.balance_end),
					deposits_count = COALESCE(v.deposits_count, 0) + result.deposits_count,
					deposits_amount = COALESCE(v.deposits_amount, 0) + result.deposits_amount,
					withdrawals_count = COALESCE(v.withdrawals_count, 0) + result.withdrawals_count,
					withdrawals_amount = COALESCE(v.withdrawals_amount, 0) + result.withdrawals_amount,
					inclusion_delay_sum = COALESCE(v.inclusion_delay_sum, 0) + result.inclusion_delay_sum,
					sync_chance = COALESCE(v.sync_chance, 0) + result.sync_chance,
					block_chance = COALESCE(v.block_chance, 0) + result.block_chance,
					attestations_scheduled = COALESCE(v.attestations_scheduled, 0) + result.attestations_scheduled,
					attestations_executed = COALESCE(v.attestations_executed, 0) + result.attestations_executed,
					attestation_head_executed = COALESCE(v.attestation_head_executed, 0) + result.attestation_head_executed,
					attestation_source_executed = COALESCE(v.attestation_source_executed, 0) + result.attestation_source_executed,
					attestation_target_executed = COALESCE(v.attestation_target_executed, 0) + result.attestation_target_executed,
					optimal_inclusion_delay_sum = COALESCE(v.optimal_inclusion_delay_sum, 0) + result.optimal_inclusion_delay_sum,
					epoch_end = result.epoch_end,
					epoch_start = result.epoch_start,
					slasher_reward = COALESCE(v.slasher_reward, 0) + result.slasher_reward,
					slashed_by = COALESCE(result.slashed_by, v.slashed_by),
					slashed_violation = COALESCE(result.slashed_violation, v.slashed_violation),
					last_executed_duty_epoch =  COALESCE(result.last_executed_duty_epoch, v.last_executed_duty_epoch)
				FROM result
				WHERE v.validator_index = result.validator_index;
			
	`, tableName), startEpoch, endEpoch)

	if err != nil {
		return errors.Wrap(err, "failed to update rolling table")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	d.log.Infof("updated %s, affected %d rows", tableName, rowsAffected)
	if rowsAffected == 0 {
		d.log.Infof("no rows affected, nothing to update for %s", tableName)
	}

	return err
}

// args inclusive
func (d *RollingAggregator) removeFromRolling(tx *sqlx.Tx, tableName string, startEpoch, endEpoch int64) error {
	startTime := time.Now()
	d.log.Infof("remove from rolling %s epochs: %d - %d ", tableName, startEpoch, endEpoch)
	defer func() {
		d.log.Infof("removed from rolling %s took %v", tableName, time.Since(startTime))
	}()

	if endEpoch < 0 {
		// if selected time frame is more than epochs exists we log an info
		d.log.Infof("rolling %sd tail epoch is negative, no end cutting", tableName)
		endEpoch = -1 // since its inclusive make it -1 so it stored 0 in table
	}

	result, err := tx.Exec(fmt.Sprintf(`
		WITH
			footer_balance_starts as (
				SELECT validator_index, balance_end as balance_start FROM validator_dashboard_data_epoch WHERE epoch = $2 -- end balance of epoch we want to remove = start epoch of epoch we start from
			),
			aggregate_tail as (
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
			),
			result as (
				SELECT
					$2 + 1 as epoch_start, --since its inclusive in the func $2 will be deducted hence +1
					aggregate_tail.validator_index as validator_index,
					COALESCE(aggregate_tail.attestations_source_reward, 0) as attestations_source_reward,
					COALESCE(aggregate_tail.attestations_target_reward, 0) as attestations_target_reward,
					COALESCE(aggregate_tail.attestations_head_reward, 0) as attestations_head_reward,
					COALESCE(aggregate_tail.attestations_inactivity_reward, 0) as attestations_inactivity_reward,
					COALESCE(aggregate_tail.attestations_inclusion_reward, 0) as attestations_inclusion_reward,
					COALESCE(aggregate_tail.attestations_reward, 0) as attestations_reward,
					COALESCE(aggregate_tail.attestations_ideal_source_reward, 0) as attestations_ideal_source_reward,
					COALESCE(aggregate_tail.attestations_ideal_target_reward, 0) as attestations_ideal_target_reward,
					COALESCE(aggregate_tail.attestations_ideal_head_reward, 0) as attestations_ideal_head_reward,
					COALESCE(aggregate_tail.attestations_ideal_inactivity_reward, 0) as attestations_ideal_inactivity_reward,
					COALESCE(aggregate_tail.attestations_ideal_inclusion_reward, 0) as attestations_ideal_inclusion_reward,
					COALESCE(aggregate_tail.attestations_ideal_reward, 0) as attestations_ideal_reward,
					COALESCE(aggregate_tail.blocks_scheduled, 0) as blocks_scheduled,
					COALESCE(aggregate_tail.blocks_proposed, 0) as blocks_proposed,
					COALESCE(aggregate_tail.blocks_cl_reward, 0) as blocks_cl_reward,
					COALESCE(aggregate_tail.sync_scheduled, 0) as sync_scheduled,
					COALESCE(aggregate_tail.sync_executed, 0) as sync_executed,
					COALESCE(aggregate_tail.sync_rewards, 0) as sync_rewards,
					balance_start,
					COALESCE(aggregate_tail.deposits_count, 0) as deposits_count,
					COALESCE(aggregate_tail.deposits_amount, 0) as deposits_amount,
					COALESCE(aggregate_tail.withdrawals_count, 0) as withdrawals_count,
					COALESCE(aggregate_tail.withdrawals_amount, 0) as withdrawals_amount,
					COALESCE(aggregate_tail.inclusion_delay_sum, 0) as inclusion_delay_sum,
					COALESCE(aggregate_tail.sync_chance, 0) as sync_chance,
					COALESCE(aggregate_tail.block_chance, 0) as block_chance,
					COALESCE(aggregate_tail.attestations_scheduled, 0) as attestations_scheduled,
					COALESCE(aggregate_tail.attestations_executed, 0) as attestations_executed,
					COALESCE(aggregate_tail.attestation_head_executed, 0) as attestation_head_executed,
					COALESCE(aggregate_tail.attestation_source_executed, 0) as attestation_source_executed,
					COALESCE(aggregate_tail.attestation_target_executed, 0) as attestation_target_executed,
					COALESCE(aggregate_tail.optimal_inclusion_delay_sum, 0) as optimal_inclusion_delay_sum,
					COALESCE(aggregate_tail.slasher_reward, 0) as slasher_reward,
					aggregate_tail.last_executed_duty_epoch
				FROM aggregate_tail  
				LEFT JOIN footer_balance_starts ON aggregate_tail.validator_index = footer_balance_starts.validator_index
			)
			UPDATE %s AS v SET
					attestations_source_reward = COALESCE(v.attestations_source_reward, 0) - result.attestations_source_reward,
					attestations_target_reward = COALESCE(v.attestations_target_reward, 0) - result.attestations_target_reward,
					attestations_head_reward = COALESCE(v.attestations_head_reward, 0) - result.attestations_head_reward,
					attestations_inactivity_reward = COALESCE(v.attestations_inactivity_reward, 0) - result.attestations_inactivity_reward,
					attestations_inclusion_reward = COALESCE(v.attestations_inclusion_reward, 0) - result.attestations_inclusion_reward,
					attestations_reward = COALESCE(v.attestations_reward, 0) - result.attestations_reward,
					attestations_ideal_source_reward = COALESCE(v.attestations_ideal_source_reward, 0) - result.attestations_ideal_source_reward,
					attestations_ideal_target_reward = COALESCE(v.attestations_ideal_target_reward, 0) - result.attestations_ideal_target_reward,
					attestations_ideal_head_reward = COALESCE(v.attestations_ideal_head_reward, 0) - result.attestations_ideal_head_reward,
					attestations_ideal_inactivity_reward = COALESCE(v.attestations_ideal_inactivity_reward, 0) - result.attestations_ideal_inactivity_reward,
					attestations_ideal_inclusion_reward = COALESCE(v.attestations_ideal_inclusion_reward, 0) - result.attestations_ideal_inclusion_reward,
					attestations_ideal_reward = COALESCE(v.attestations_ideal_reward, 0) - result.attestations_ideal_reward,
					blocks_scheduled = COALESCE(v.blocks_scheduled, 0) - result.blocks_scheduled,
					blocks_proposed = COALESCE(v.blocks_proposed, 0) - result.blocks_proposed,
					blocks_cl_reward = COALESCE(v.blocks_cl_reward, 0) - result.blocks_cl_reward,
					sync_scheduled = COALESCE(v.sync_scheduled, 0) - result.sync_scheduled,
					sync_executed = COALESCE(v.sync_executed, 0) - result.sync_executed,
					sync_rewards = COALESCE(v.sync_rewards, 0) - result.sync_rewards,
					balance_start = COALESCE(result.balance_start, v.balance_start),
					deposits_count = COALESCE(v.deposits_count, 0) - result.deposits_count,
					deposits_amount = COALESCE(v.deposits_amount, 0) - result.deposits_amount,
					withdrawals_count = COALESCE(v.withdrawals_count, 0) - result.withdrawals_count,
					withdrawals_amount = COALESCE(v.withdrawals_amount, 0) - result.withdrawals_amount,
					inclusion_delay_sum = COALESCE(v.inclusion_delay_sum, 0) - result.inclusion_delay_sum,
					sync_chance = COALESCE(v.sync_chance, 0) - result.sync_chance,
					block_chance = COALESCE(v.block_chance, 0) - result.block_chance,
					attestations_scheduled = COALESCE(v.attestations_scheduled, 0) - result.attestations_scheduled,
					attestations_executed = COALESCE(v.attestations_executed, 0) - result.attestations_executed,
					attestation_head_executed = COALESCE(v.attestation_head_executed, 0) - result.attestation_head_executed,
					attestation_source_executed = COALESCE(v.attestation_source_executed, 0) - result.attestation_source_executed,
					attestation_target_executed = COALESCE(v.attestation_target_executed, 0) - result.attestation_target_executed,
					optimal_inclusion_delay_sum = COALESCE(v.optimal_inclusion_delay_sum, 0) - result.optimal_inclusion_delay_sum,
					epoch_start = result.epoch_start,
					slasher_reward = COALESCE(v.slasher_reward, 0) - result.slasher_reward,
					last_executed_duty_epoch =  COALESCE(result.last_executed_duty_epoch, v.last_executed_duty_epoch)
				FROM result
				WHERE v.validator_index = result.validator_index;
			
	`, tableName), startEpoch, endEpoch)

	if err != nil {
		return errors.Wrap(err, "failed to update rolling table")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	d.log.Infof("updated %s, affected %d rows", tableName, rowsAffected)
	if rowsAffected == 0 {
		d.log.Infof("no rows affected, nothing to update for %s", tableName)
	}

	return err
}
