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
func (d *epochToTotalAggregator) aggregateTotal(currentExportedEpoch uint64) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	startTime := time.Now()
	defer func() {
		d.log.Infof("aggregate total took %v", time.Since(startTime))
		metrics.TaskDuration.WithLabelValues("exporter_v2dash_agg_total").Observe(time.Since(startTime).Seconds())
	}()

	lastTotalExported, err := edb.GetLastExportedTotalEpoch()
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to get last exported total epoch")
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

	err = d.aggregateAndAddToTotal(lastTotalExported.EpochEnd, currentExportedEpoch+1)
	if err != nil {
		return errors.Wrap(err, "failed to aggregate total")
	}

	return nil
}

// epochStart incl, epochEnd excl
func (d *epochToTotalAggregator) aggregateAndAddToTotal(epochStart, epochEnd uint64) error {
	tx, err := db.AlloyWriter.Beginx()
	if err != nil {
		return err
	}
	defer utils.Rollback(tx)

	d.log.Infof("aggregating total (from: %d) up to %d", epochStart, epochEnd)

	err = AddToRollingCustom(tx, CustomRolling{
		StartEpoch:                    epochStart,
		EndEpoch:                      epochEnd,
		StartBoundEpoch:               0,
		TableFrom:                     "validator_dashboard_data_epoch",
		TableTo:                       "validator_dashboard_data_rolling_total",
		TailBalancesInsertColumnQuery: "0 as balance_start,", // Since all validators start with a 0 balance until deposit is voted in, we can just set it to 0. Genesis validators will be set to 0 to unify the data access approach
		TableFromEpochColumn:          "epoch",
		Log:                           d.log,
		TableConflict:                 "(validator_index)",

		// This may come in handy at some point so leaving it there if you need the first value in an epoch range for a given validator

		// TailBalancesQuery: `
		// 	balance_start_epochs as (
		// 		SELECT validator_index, MIN(epoch) as epoch FROM validator_dashboard_data_epoch WHERE epoch >= $1 AND epoch <= $2 AND balance_start IS NOT NULL
		// 		GROUP BY validator_index
		// 	),
		// 	balance_starts as (
		// 			SELECT validator_index, balance_start FROM balance_start_epochs LEFT JOIN validator_dashboard_data_epoch USING (validator_index, epoch)
		// 	),`,
		// TailBalancesJoinQuery:         `LEFT JOIN balance_starts ON aggregate_head.validator_index = balance_starts.validator_index`,
		//TailBalancesInsertColumnQuery: "balance_start,",
	})

	if err != nil {
		return err
	}

	return tx.Commit()
}
