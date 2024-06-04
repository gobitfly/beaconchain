package modules

import (
	"bytes"
	"database/sql"
	"fmt"
	"text/template"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

/**
This file handles the logic for rolling aggregation for 24h, 7d, 31d and 90d. Total also relies on the AddRollingCustom method as well as the day and hour aggregate.
The way this works is by adding new epochs to the rolling table (adding to head) and removing the old epochs at the end (removing tail) so that the time duration of rolling stays constant.

If the rolling tables fall out of sync due to long offline time or initial sync, the tables are bootstrapped. This bootstrap method must be provided,
7d, 31d, 90d use a bootstrap from the utc_days table to get started and 24h the hourly table.
*/

/*
	In case exporter needs to be reset and re-export from an older epoch again, the rolling tables must be truncated.
	Total must be bootstrapped by hand, for example via daily utc table. Other rolling tables will be bootstrapped automatically.
*/

type RollingAggregator struct {
	RollingAggregatorInt
	log ModuleLog
}

type RollingAggregatorInt interface {
	bootstrap(tx *sqlx.Tx, days int, tableName string) error

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
		err = db.AlloyWriter.Get(&bounds, fmt.Sprintf(`SELECT max(epoch_start) as epoch_start, epoch_end as epoch_end FROM %s LIMIT 1`, tableName))
	} else {
		err = tx.Get(&bounds, fmt.Sprintf(`SELECT max(epoch_start) as epoch_start, epoch_end as epoch_end FROM %s LIMIT 1`, tableName))
	}
	return bounds, err
}

// returns the tail epochs (those must be removed from rolling) for a given intendedHeadEpoch for a given rolling table
// fE a tail epoch for rolling 1 day aggregation (225 epochs) for boundsStart 0 (start epoch of last rolling export) and intendedHeadEpoch 227 on ethereum would correspond to a tail range of 0 - 1
// meaning epoch [0,1] must be removed from the rolling table if you want to add epoch 227
// first argument (start bound) is inclusive, second arg (end bound) is exclusive
func (d *RollingAggregator) getTailBoundsXDays(days int, boundsStart uint64, intendedHeadEpoch uint64) (int64, int64) {
	aggTailEpochStart := int64(boundsStart) // current bounds start must be removed
	aggTailEpochEnd := int64(intendedHeadEpoch-utils.EpochsPerDay()*uint64(days)) + 1
	d.log.Debugf("tail bounds for %dd: %d - %d | intendedHead: %v | boundsStart: %v", days, aggTailEpochStart, aggTailEpochEnd, intendedHeadEpoch, boundsStart)

	return aggTailEpochStart, aggTailEpochEnd
}

// Note that currentEpochHead is the current exported epoch in the db
func (d *RollingAggregator) Aggregate(days int, tableName string, currentEpochHead uint64) error {
	return d.aggregateInternal(days, tableName, currentEpochHead, false)
}

// Note that currentEpochHead is the current exported epoch in the db
func (d *RollingAggregator) aggregateInternal(days int, tableName string, currentEpochHead uint64, forceBootstrap bool) error {
	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer utils.Rollback(tx)

	bootstrap := false
	if forceBootstrap {
		bootstrap = true
		d.log.Infof("force bootstrap rolling %dd", days)
	}

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

	// if current stored rolling table is far behind, bootstrap again
	// in this case far means more than what we aggregate in the hour table, meaning a bootstrap
	// will get faster to head then re-exporting amount of getBootstrapOnEpochsBehind() old epochs
	if currentEpochHead+1-bounds.EpochEnd >= d.getBootstrapOnEpochsBehind() { // EpochEnd is excl so +1 to get the inclusive epoch number
		d.log.Infof("currentEpochHead: %d, bounds.EpochEnd: %d, getBootstrapOnEpochsBehind(): %d, leftsum: %d", currentEpochHead, bounds.EpochEnd, d.getBootstrapOnEpochsBehind(), currentEpochHead+1-bounds.EpochEnd)
		bootstrap = true
	}

	if bootstrap {
		metrics.Errors.WithLabelValues(fmt.Sprintf("exporter_v2dash_agg_bootstrap_%dd", days)).Inc()
		d.log.Infof("rolling %dd bootstraping starting", days)

		err = d.bootstrap(tx, days, tableName)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to bootstrap rolling %dd aggregate", days))
		}

		bounds, err = d.getCurrentRollingBounds(tx, tableName)
		if err != nil {
			return errors.Wrap(err, "failed to get current rolling bounds")
		}

		d.log.Infof("rolling %dd bootstraping finished, currentHead: %v | bounds: %v | Epochs Per Day: %v", days, currentEpochHead, bounds, utils.EpochsPerDay())

		// if rolling bounds are exactly what they should be, we are done here
		if currentEpochHead == bounds.EpochEnd-1 && bounds.EpochEnd-uint64(days)*utils.EpochsPerDay() == bounds.EpochStart {
			log.Infof("rolling %dd is up to date, nothing to do", days) // perfect bounds after bootstrap, lucky day, done here
			err = tx.Commit()
			if err != nil {
				return errors.Wrap(err, "failed to commit transaction")
			}
			return nil
		}
	}

	if !bootstrap && bounds.EpochEnd-bounds.EpochStart != utils.EpochsPerDay()*uint64(days) {
		log.Warnf("rolling %dd boundaries are out of bounds (%d-%d, %d), this is expected after bootstrap, but not after that. Keep an eye on it", days, bounds.EpochStart, bounds.EpochEnd, bounds.EpochEnd-bounds.EpochStart)
	}

	// bounds for what to aggregate and add to the head of the rolling table
	aggHeadEpochStart := bounds.EpochEnd
	aggHeadEpochEnd := currentEpochHead + 1

	// bounds for what to aggregate and remove from the tail of the rolling table
	aggTailEpochStart, aggTailEpochEnd := d.getTailBoundsXDays(days, bounds.EpochStart, currentEpochHead)
	d.log.Infof("rolling %dd epochs: %d - %d, %d - %d", days, aggHeadEpochStart, aggHeadEpochEnd, aggTailEpochStart, aggTailEpochEnd)

	// sanity check if all tail epochs are present in db
	missing, err := edb.GetMissingEpochsBetween(aggTailEpochStart, aggTailEpochEnd)
	if err != nil {
		return errors.Wrap(err, "failed to get missing tail epochs")
	}
	if len(missing) > 0 {
		// If exporter falls back on head around the bootstrap bound end, it will start backfill
		// and will trigger a rolling aggregate on bound end. But it will not bootstrap since recent data is
		// maybe just a few epochs old. So it will try to add to head and remove from tail but since we are in
		// bootstrap mode tail epochs won't be there.
		// Now when this occurs and epochs are missing AND the current head is a bootstrap end, we force a bootstrap to solve this.

		// a bool helper to indicate whether currentEpochHead is the end of a bootstrap bound
		_, utcEndBound := getDayAggregateBounds(currentEpochHead)
		isCurrentEpochHeadBootstrapBound := utcEndBound-1 == currentEpochHead

		if isCurrentEpochHeadBootstrapBound {
			return d.aggregateInternal(days, tableName, currentEpochHead, true)
		}
		return errors.New(fmt.Sprintf("missing epochs in db for rolling %dd tail: %v", days, missing))
	}

	// sanity check if all head epochs are present in db
	missingHead, err := edb.GetMissingEpochsBetween(int64(aggHeadEpochStart), int64(aggHeadEpochEnd))
	if err != nil {
		return errors.Wrap(err, "failed to get missing head epochs")
	}
	if len(missingHead) > 0 {
		return errors.New(fmt.Sprintf("missing epochs in db for rolling %dd head: %v", days, missingHead))
	}

	// add head and fix/remove from tail
	err = d.aggregateRolling(tx, tableName, aggHeadEpochStart, aggHeadEpochEnd, aggTailEpochStart, aggTailEpochEnd)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to aggregate rolling %dd", days))
	}

	// Sanity check, get bounds again after rolling aggregate
	sanityBounds, err := d.getCurrentRollingBounds(tx, tableName)
	if err != nil {
		return errors.Wrap(err, "failed to get current rolling bounds for sanity check")
	}

	// skip sanity check
	if sanityBounds.EpochEnd-sanityBounds.EpochStart != utils.EpochsPerDay()*uint64(days) {
		// only do sanity check if tail bounds are not negative (we store them as uint in the db so it will never be utils.EpochsPerDay()*uint64(days) in this case
		if aggTailEpochStart >= 0 && aggTailEpochEnd >= 0 {
			return errors.New(fmt.Sprintf("sanity check failed, rolling boundaries are out of bounds for %vd agg (%d-%d, %d)", days, sanityBounds.EpochStart, sanityBounds.EpochEnd, sanityBounds.EpochEnd-sanityBounds.EpochStart))
		}
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

	needsBootstrap := int64(intendedHeadEpoch-bounds.EpochEnd) >= int64(d.getBootstrapOnEpochsBehind())

	d.log.Infof("%dd needs bootstrap: %v", days, needsBootstrap)
	// if rolling table is empty / not bootstrapped yet or needs a bootstrap assume bounds of what the would be after a bootstrap
	if (bounds.EpochEnd == 0 && bounds.EpochStart == 0) || needsBootstrap {
		// assume bounds after bootstrap
		startBound, endBounds := d.getBootstrapBounds(intendedHeadEpoch, uint64(days))
		bounds.EpochStart = startBound
		bounds.EpochEnd = endBounds
		d.log.Infof("bootstrap bounds for rolling %dd: %d - %d | current Head (excl): %v", days, bounds.EpochStart, bounds.EpochEnd, intendedHeadEpoch+1)
	}

	aggTailEpochStart, aggTailEpochEnd := d.getTailBoundsXDays(days, bounds.EpochStart, intendedHeadEpoch)

	return edb.GetMissingEpochsBetween(aggTailEpochStart, aggTailEpochEnd)
}

// Adds the new epochs (headEpochStart to headEpochEnd) to the rolling table and removes the old ones (tailEpochStart to tailEpochEnd)
// start params are inclusive, end params are exclusive
func (d *RollingAggregator) aggregateRolling(tx *sqlx.Tx, tableName string, headEpochStart, headEpochEnd uint64, tailEpochStart, tailEpochEnd int64) error {
	d.log.Infof("aggregating rolling %s epochs: %d - %d, %d - %d", tableName, headEpochStart, headEpochEnd, tailEpochStart, tailEpochEnd)
	defer d.log.Infof("aggregated rolling %s epochs: %d - %d, %d - %d", tableName, headEpochStart, headEpochEnd, tailEpochStart, tailEpochEnd)

	// Important to remove first since head could contain validators that were not present in tail, would interfere with boundaries
	if tailEpochEnd >= tailEpochStart {
		err := d.removeFromRolling(tx, tableName, tailEpochStart, tailEpochEnd)
		if err != nil {
			return errors.Wrap(err, "failed to remove epochs from rolling")
		}
	}
	if headEpochEnd >= headEpochStart {
		err := d.addToRolling(tx, tableName, headEpochStart, headEpochEnd, tailEpochEnd)
		if err != nil {
			return errors.Wrap(err, "failed to add epochs to rolling")
		}
	}

	return nil
}

// Inserts new validators or updated existing ones into the rolling table
// startEpoch incl | endEpoch, tailEnd excl
func (d *RollingAggregator) addToRolling(tx *sqlx.Tx, tableName string, startEpoch, endEpoch uint64, tailEnd int64) error {
	startTime := time.Now()
	d.log.Infof("add to rolling %s epochs: %d - %d", tableName, startEpoch, endEpoch)
	defer func() {
		d.log.Infof("added to rolling %s took %v", tableName, time.Since(startTime))
	}()

	if tailEnd < 0 {
		tailEnd = 0
	}

	return AddToRollingCustom(tx, CustomRolling{
		StartEpoch:           startEpoch,
		EndEpoch:             endEpoch,
		StartBoundEpoch:      tailEnd,
		TableFrom:            edb.EpochWriterTableName,
		TableTo:              tableName,
		TableFromEpochColumn: "epoch",
		Log:                  d.log,
		TableConflict:        "(validator_index)",
	})
}

type CustomRolling struct {
	Log                  ModuleLog // for logging, must provide
	StartEpoch           uint64    // incl, must be provided
	EndEpoch             uint64    // excl, must be provided
	StartBoundEpoch      int64     // incl, must be provided
	TableFrom            string    // must provide
	TableTo              string    // must provide
	TableFromEpochColumn string    // must provide
	TableConflict        string    // must provide

	TailBalancesQuery             string // optional
	TailBalancesJoinQuery         string // optional
	TailBalancesInsertColumnQuery string // optional
	TableDayColum                 string // optional
	TableDayValue                 string // optional

	Agg TableAGG // do not provide, will be overwritten
}

type TableAGG struct {
	SUM                              string
	BOOL_OR                          string
	MAX                              string
	AGG_END                          string
	WHERE_AND_GROUP                  string
	BALANCE_END_Q                    string
	BALANCE_END_COLUMN               string
	HEAD_BALANCE_QUERY               string
	HEAD_BALANCE_JOIN                string
	HEAD_BALANCE_SINGLE_EPOCH_COLUMN string
}

// This method is the bread and butter of all aggregation. It is used by rolling window aggregation to add to head,
// it is used by total to add to head, it is used by utc day and hour aggregation to add to head
func AddToRollingCustom(tx *sqlx.Tx, custom CustomRolling) error {
	if custom.TailBalancesInsertColumnQuery == "" {
		custom.TailBalancesInsertColumnQuery = "null,"
	}

	if custom.StartEpoch > custom.EndEpoch-1 {
		custom.Log.Infof("nothing to do, start epoch is greater than end epoch (%d > %d)", custom.StartEpoch, custom.EndEpoch-1)
		return nil
	}

	// When aggregating multiple epochs (aggregate mode)
	custom.Agg = TableAGG{
		SUM:     "SUM(",
		BOOL_OR: "bool_or(",
		MAX:     "MAX(",
		AGG_END: ")",

		// What epochs to select for aggregate
		WHERE_AND_GROUP: fmt.Sprintf(`
			WHERE %[1]s >= $1 AND %[1]s < $2 
			GROUP BY validator_index
		`, custom.TableFromEpochColumn),

		// Head balance from the most recent epoch so select and join the agg set
		HEAD_BALANCE_QUERY: fmt.Sprintf(`
			head_balance_ends as (
				SELECT validator_index, balance_end FROM %[1]s WHERE %[2]s = $2 -1
			),
		`, custom.TableFrom, custom.TableFromEpochColumn),
		HEAD_BALANCE_JOIN: `
			LEFT JOIN head_balance_ends ON aggregate_head.validator_index = head_balance_ends.validator_index
		`,
		HEAD_BALANCE_SINGLE_EPOCH_COLUMN: "", // since data is provided by join we dont need to select it
	}

	// I know how this looks but the query planner of postgres has difficulty with the agg group by for just one epoch
	// So here some optimization for when we only aggregate one epoch
	// just select, no aggregate, grouping or join needed
	if custom.StartEpoch == custom.EndEpoch-1 {
		custom.Agg = TableAGG{
			SUM:             "",
			BOOL_OR:         "",
			MAX:             "",
			AGG_END:         "",
			WHERE_AND_GROUP: fmt.Sprintf(`WHERE %s = $1`, custom.TableFromEpochColumn),

			// For head balance with just one epoch just use the column, forget about the join
			HEAD_BALANCE_SINGLE_EPOCH_COLUMN: "balance_end,",
		}

		// If start bound is the same as start and end, we can ommit that join altogether
		// postgres really doesnt like joining on the same epoch
		if custom.StartBoundEpoch == int64(custom.StartEpoch) && custom.TailBalancesQuery != "" {
			custom.TailBalancesQuery = ""
			custom.TailBalancesJoinQuery = ""
			custom.Agg.HEAD_BALANCE_SINGLE_EPOCH_COLUMN += custom.TailBalancesInsertColumnQuery
		}
	}

	// Use NULLIF(x,0) to save storage space
	tmpl := `
		-- Agg Add, From: {{ .TableFrom }}, To: {{ .TableTo }}
		WITH
			{{ .Agg.HEAD_BALANCE_QUERY }}
			{{ .TailBalancesQuery }} -- balance start query
			aggregate_head as (
				SELECT 
					validator_index,
					{{ .Agg.HEAD_BALANCE_SINGLE_EPOCH_COLUMN }}
					{{ .Agg.SUM }}attestations_source_reward{{ .Agg.AGG_END }} as attestations_source_reward,
					{{ .Agg.SUM }}attestations_target_reward{{ .Agg.AGG_END }} as attestations_target_reward,
					{{ .Agg.SUM }}attestations_head_reward{{ .Agg.AGG_END }} as attestations_head_reward,
					{{ .Agg.SUM }}attestations_inactivity_reward{{ .Agg.AGG_END }} as attestations_inactivity_reward,
					{{ .Agg.SUM }}attestations_inclusion_reward{{ .Agg.AGG_END }} as attestations_inclusion_reward,
					{{ .Agg.SUM }}attestations_reward{{ .Agg.AGG_END }} as attestations_reward,
					{{ .Agg.SUM }}attestations_ideal_source_reward{{ .Agg.AGG_END }} as attestations_ideal_source_reward,
					{{ .Agg.SUM }}attestations_ideal_target_reward{{ .Agg.AGG_END }} as attestations_ideal_target_reward,
					{{ .Agg.SUM }}attestations_ideal_head_reward{{ .Agg.AGG_END }} as attestations_ideal_head_reward,
					{{ .Agg.SUM }}attestations_ideal_inactivity_reward{{ .Agg.AGG_END }} as attestations_ideal_inactivity_reward,
					{{ .Agg.SUM }}attestations_ideal_inclusion_reward{{ .Agg.AGG_END }} as attestations_ideal_inclusion_reward,
					{{ .Agg.SUM }}attestations_ideal_reward{{ .Agg.AGG_END }} as attestations_ideal_reward,
					{{ .Agg.SUM }}blocks_scheduled{{ .Agg.AGG_END }} as blocks_scheduled,
					{{ .Agg.SUM }}blocks_proposed{{ .Agg.AGG_END }} as blocks_proposed,
					{{ .Agg.SUM }}blocks_cl_reward{{ .Agg.AGG_END }} as blocks_cl_reward,
					{{ .Agg.SUM }}blocks_cl_attestations_reward{{ .Agg.AGG_END }} as blocks_cl_attestations_reward,
					{{ .Agg.SUM }}blocks_cl_sync_aggregate_reward{{ .Agg.AGG_END }} as blocks_cl_sync_aggregate_reward,
					{{ .Agg.SUM }}sync_scheduled{{ .Agg.AGG_END }} as sync_scheduled,
					{{ .Agg.SUM }}sync_executed{{ .Agg.AGG_END }} as sync_executed,
					{{ .Agg.SUM }}sync_rewards{{ .Agg.AGG_END }} as sync_rewards,
					{{ .Agg.BOOL_OR }}slashed{{ .Agg.AGG_END }} as slashed,
					{{ .Agg.SUM }}deposits_count{{ .Agg.AGG_END }} as deposits_count,
					{{ .Agg.SUM }}deposits_amount{{ .Agg.AGG_END }} as deposits_amount,
					{{ .Agg.SUM }}withdrawals_count{{ .Agg.AGG_END }} as withdrawals_count,
					{{ .Agg.SUM }}withdrawals_amount{{ .Agg.AGG_END }} as withdrawals_amount,
					{{ .Agg.SUM }}inclusion_delay_sum{{ .Agg.AGG_END }} as inclusion_delay_sum,
					{{ .Agg.SUM }}blocks_expected{{ .Agg.AGG_END }} as blocks_expected,
					{{ .Agg.SUM }}sync_committees_expected{{ .Agg.AGG_END }} as sync_committees_expected, 
					{{ .Agg.SUM }}attestations_scheduled{{ .Agg.AGG_END }} as attestations_scheduled,
					{{ .Agg.SUM }}attestations_executed{{ .Agg.AGG_END }} as attestations_executed,
					{{ .Agg.SUM }}attestation_head_executed{{ .Agg.AGG_END }} as attestation_head_executed,
					{{ .Agg.SUM }}attestation_source_executed{{ .Agg.AGG_END }} as attestation_source_executed,
					{{ .Agg.SUM }}attestation_target_executed{{ .Agg.AGG_END }} as attestation_target_executed,
					{{ .Agg.SUM }}optimal_inclusion_delay_sum{{ .Agg.AGG_END }} as optimal_inclusion_delay_sum,
					{{ .Agg.SUM }}slasher_reward{{ .Agg.AGG_END }} as slasher_reward,
					{{ .Agg.MAX }}slashed_by{{ .Agg.AGG_END }} as slashed_by,
					{{ .Agg.MAX }}slashed_violation{{ .Agg.AGG_END }} as slashed_violation,
					{{ .Agg.MAX }}last_executed_duty_epoch{{ .Agg.AGG_END }} as last_executed_duty_epoch		
				FROM {{ .TableFrom }}
				{{ .Agg.WHERE_AND_GROUP }} 
			)
			INSERT INTO {{ .TableTo }} (
				{{ .TableDayColum }}
				epoch_end,
				epoch_start,
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
				blocks_cl_attestations_reward,
				blocks_cl_sync_aggregate_reward,
				sync_scheduled,
				sync_executed,
				sync_rewards,
				slashed,
				balance_end,
				balance_start,
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
				{{ .TableDayValue }}
				$2 as epoch_end, -- exclusive
				$3 as epoch_start, -- inclusive, only write on insert - do not update in UPDATE part! Use tail start epoch
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
				COALESCE(aggregate_head.blocks_cl_attestations_reward, 0) as blocks_cl_attestations_reward,
				COALESCE(aggregate_head.blocks_cl_sync_aggregate_reward, 0) as blocks_cl_sync_aggregate_reward,
				COALESCE(aggregate_head.sync_scheduled, 0) as sync_scheduled,
				COALESCE(aggregate_head.sync_executed, 0) as sync_executed,
				COALESCE(aggregate_head.sync_rewards, 0) as sync_rewards,
				aggregate_head.slashed,
				balance_end,
				{{ .TailBalancesInsertColumnQuery }} -- balance_start
				COALESCE(aggregate_head.deposits_count, 0) as deposits_count,
				COALESCE(aggregate_head.deposits_amount, 0) as deposits_amount,
				COALESCE(aggregate_head.withdrawals_count, 0) as withdrawals_count,
				COALESCE(aggregate_head.withdrawals_amount, 0) as withdrawals_amount,
				COALESCE(aggregate_head.inclusion_delay_sum, 0) as inclusion_delay_sum,
				COALESCE(aggregate_head.blocks_expected, 0) as blocks_expected,
				COALESCE(aggregate_head.sync_committees_expected, 0) as sync_committees_expected,
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
			{{ .TailBalancesJoinQuery }} -- balance start join
			{{ .Agg.HEAD_BALANCE_JOIN }}
			ON CONFLICT {{ .TableConflict }} DO UPDATE SET
					attestations_source_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_source_reward, 0) + EXCLUDED.attestations_source_reward, 0),
					attestations_target_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_target_reward, 0) + EXCLUDED.attestations_target_reward, 0),
					attestations_head_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_head_reward, 0) + EXCLUDED.attestations_head_reward, 0),
					attestations_inactivity_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_inactivity_reward, 0) + EXCLUDED.attestations_inactivity_reward, 0),
					attestations_inclusion_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_inclusion_reward, 0) + EXCLUDED.attestations_inclusion_reward, 0),
					attestations_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_reward, 0) + EXCLUDED.attestations_reward, 0),
					attestations_ideal_source_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_ideal_source_reward, 0) + EXCLUDED.attestations_ideal_source_reward, 0),
					attestations_ideal_target_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_ideal_target_reward, 0) + EXCLUDED.attestations_ideal_target_reward, 0),
					attestations_ideal_head_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_ideal_head_reward, 0) + EXCLUDED.attestations_ideal_head_reward, 0),
					attestations_ideal_inactivity_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_ideal_inactivity_reward, 0) + EXCLUDED.attestations_ideal_inactivity_reward, 0),
					attestations_ideal_inclusion_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_ideal_inclusion_reward, 0) + EXCLUDED.attestations_ideal_inclusion_reward, 0),
					attestations_ideal_reward = NULLIF(COALESCE({{ .TableTo }}.attestations_ideal_reward, 0) + EXCLUDED.attestations_ideal_reward, 0),
					blocks_scheduled = NULLIF(COALESCE({{ .TableTo }}.blocks_scheduled, 0) + EXCLUDED.blocks_scheduled, 0),
					blocks_proposed = NULLIF(COALESCE({{ .TableTo }}.blocks_proposed, 0) + EXCLUDED.blocks_proposed, 0),
					blocks_cl_reward = NULLIF(COALESCE({{ .TableTo }}.blocks_cl_reward, 0) + EXCLUDED.blocks_cl_reward, 0),
					blocks_cl_attestations_reward = NULLIF(COALESCE({{ .TableTo }}.blocks_cl_attestations_reward, 0) + EXCLUDED.blocks_cl_attestations_reward, 0),
					blocks_cl_sync_aggregate_reward = NULLIF(COALESCE({{ .TableTo }}.blocks_cl_sync_aggregate_reward, 0) + EXCLUDED.blocks_cl_sync_aggregate_reward, 0),
					sync_scheduled = NULLIF(COALESCE({{ .TableTo }}.sync_scheduled, 0) + EXCLUDED.sync_scheduled, 0),
					sync_executed = NULLIF(COALESCE({{ .TableTo }}.sync_executed, 0) + EXCLUDED.sync_executed, 0),
					sync_rewards = NULLIF(COALESCE({{ .TableTo }}.sync_rewards, 0) + EXCLUDED.sync_rewards, 0),
					slashed = EXCLUDED.slashed OR {{ .TableTo }}.slashed,
					balance_end = COALESCE(EXCLUDED.balance_end, {{ .TableTo }}.balance_end),
					deposits_count = NULLIF(COALESCE({{ .TableTo }}.deposits_count, 0) + EXCLUDED.deposits_count, 0),
					deposits_amount = NULLIF(COALESCE({{ .TableTo }}.deposits_amount, 0) + EXCLUDED.deposits_amount, 0),
					withdrawals_count = NULLIF(COALESCE({{ .TableTo }}.withdrawals_count, 0) + EXCLUDED.withdrawals_count, 0),
					withdrawals_amount = NULLIF(COALESCE({{ .TableTo }}.withdrawals_amount, 0) + EXCLUDED.withdrawals_amount, 0),
					inclusion_delay_sum = NULLIF(COALESCE({{ .TableTo }}.inclusion_delay_sum, 0) + EXCLUDED.inclusion_delay_sum, 0),
					blocks_expected = NULLIF(COALESCE({{ .TableTo }}.blocks_expected, 0) + EXCLUDED.blocks_expected, 0),
					sync_committees_expected = NULLIF(COALESCE({{ .TableTo }}.sync_committees_expected, 0) + EXCLUDED.sync_committees_expected, 0),
					attestations_scheduled = NULLIF(COALESCE({{ .TableTo }}.attestations_scheduled, 0) + EXCLUDED.attestations_scheduled, 0),
					attestations_executed = NULLIF(COALESCE({{ .TableTo }}.attestations_executed, 0) + EXCLUDED.attestations_executed, 0),
					attestation_head_executed = NULLIF(COALESCE({{ .TableTo }}.attestation_head_executed, 0) + EXCLUDED.attestation_head_executed, 0),
					attestation_source_executed = NULLIF(COALESCE({{ .TableTo }}.attestation_source_executed, 0) + EXCLUDED.attestation_source_executed, 0),
					attestation_target_executed = NULLIF(COALESCE({{ .TableTo }}.attestation_target_executed, 0) + EXCLUDED.attestation_target_executed, 0),
					optimal_inclusion_delay_sum = NULLIF(COALESCE({{ .TableTo }}.optimal_inclusion_delay_sum, 0) + EXCLUDED.optimal_inclusion_delay_sum, 0),
					epoch_end = EXCLUDED.epoch_end,
					slasher_reward = NULLIF(COALESCE({{ .TableTo }}.slasher_reward, 0) + EXCLUDED.slasher_reward, 0),
					slashed_by = COALESCE(EXCLUDED.slashed_by, {{ .TableTo }}.slashed_by),
					slashed_violation = COALESCE(EXCLUDED.slashed_violation, {{ .TableTo }}.slashed_violation),
					last_executed_duty_epoch =  COALESCE(EXCLUDED.last_executed_duty_epoch, {{ .TableTo }}.last_executed_duty_epoch)`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	var queryBuffer bytes.Buffer
	if err := t.Execute(&queryBuffer, custom); err != nil {
		return errors.Wrap(err, "failed to execute template")
	}

	custom.Log.Debugf("TableTo: %v | TableFrom: %v | StartEpoch: %v | EndEpoch: %v | StartBoundEpoch: %v", custom.TableTo, custom.TableFrom, custom.StartEpoch, custom.EndEpoch, custom.StartBoundEpoch)

	result, err := tx.Exec(queryBuffer.String(),
		custom.StartEpoch, custom.EndEpoch, custom.StartBoundEpoch,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update rolling table")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	custom.Log.Infof("updated %s, affected %d rows", custom.TableTo, rowsAffected)
	if rowsAffected == 0 {
		custom.Log.Infof("no rows affected, nothing to update for %s", custom.TableTo)
	}

	return err
}

// startEpoch incl, endEpoch excl
func (d *RollingAggregator) removeFromRolling(tx *sqlx.Tx, tableName string, startEpoch, endEpoch int64) error {
	startTime := time.Now()
	d.log.Infof("remove from rolling %s epochs: %d - %d ", tableName, startEpoch, endEpoch)
	defer func() {
		d.log.Infof("removed from rolling %s took %v", tableName, time.Since(startTime))
	}()

	if endEpoch-1 < 0 {
		// if selected time frame is more than epochs exists we log an info
		d.log.Infof("rolling %sd tail epoch is negative, no end cutting", tableName)
		endEpoch = 0 // since its inclusive make it -1 so it stored 0 in table
	}

	result, err := tx.Exec(fmt.Sprintf(`
		WITH
			footer_balance_starts as (
				SELECT validator_index, balance_end as balance_start FROM %[2]s WHERE epoch = $2 -1 -- end balance of epoch we want to remove = start epoch of epoch we start from
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
					SUM(blocks_cl_attestations_reward) as blocks_cl_attestations_reward,
					SUM(blocks_cl_sync_aggregate_reward) as blocks_cl_sync_aggregate_reward,
					SUM(sync_scheduled) as sync_scheduled,
					SUM(sync_executed) as sync_executed,
					SUM(sync_rewards) as sync_rewards,
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
					MAX(last_executed_duty_epoch) as last_executed_duty_epoch
				FROM %[2]s
				WHERE epoch >= $1 AND epoch < $2
				GROUP BY validator_index
			),
			result as (
				SELECT
					$2 as epoch_start, --since its inclusive in the func $2 will be deducted hence +1
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
					COALESCE(aggregate_tail.blocks_cl_attestations_reward, 0) as blocks_cl_attestations_reward,
					COALESCE(aggregate_tail.blocks_cl_sync_aggregate_reward, 0) as blocks_cl_sync_aggregate_reward,
					COALESCE(aggregate_tail.sync_scheduled, 0) as sync_scheduled,
					COALESCE(aggregate_tail.sync_executed, 0) as sync_executed,
					COALESCE(aggregate_tail.sync_rewards, 0) as sync_rewards,
					balance_start,
					COALESCE(aggregate_tail.deposits_count, 0) as deposits_count,
					COALESCE(aggregate_tail.deposits_amount, 0) as deposits_amount,
					COALESCE(aggregate_tail.withdrawals_count, 0) as withdrawals_count,
					COALESCE(aggregate_tail.withdrawals_amount, 0) as withdrawals_amount,
					COALESCE(aggregate_tail.inclusion_delay_sum, 0) as inclusion_delay_sum,
					COALESCE(aggregate_tail.blocks_expected, 0) as blocks_expected,
					COALESCE(aggregate_tail.sync_committees_expected, 0) as sync_committees_expected,
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
			UPDATE %[1]s AS v SET
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
					blocks_cl_attestations_reward = COALESCE(v.blocks_cl_attestations_reward, 0) - result.blocks_cl_attestations_reward,
					blocks_cl_sync_aggregate_reward = COALESCE(v.blocks_cl_sync_aggregate_reward, 0) - result.blocks_cl_sync_aggregate_reward,
					sync_scheduled = COALESCE(v.sync_scheduled, 0) - result.sync_scheduled,
					sync_executed = COALESCE(v.sync_executed, 0) - result.sync_executed,
					sync_rewards = COALESCE(v.sync_rewards, 0) - result.sync_rewards,
					balance_start = COALESCE(result.balance_start, v.balance_start),
					deposits_count = COALESCE(v.deposits_count, 0) - result.deposits_count,
					deposits_amount = COALESCE(v.deposits_amount, 0) - result.deposits_amount,
					withdrawals_count = COALESCE(v.withdrawals_count, 0) - result.withdrawals_count,
					withdrawals_amount = COALESCE(v.withdrawals_amount, 0) - result.withdrawals_amount,
					inclusion_delay_sum = COALESCE(v.inclusion_delay_sum, 0) - result.inclusion_delay_sum,
					blocks_expected = COALESCE(v.blocks_expected, 0) - result.blocks_expected,
					sync_committees_expected = COALESCE(v.sync_committees_expected, 0) - result.sync_committees_expected,
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
			
	`, tableName, edb.EpochWriterTableName), startEpoch, endEpoch)

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
