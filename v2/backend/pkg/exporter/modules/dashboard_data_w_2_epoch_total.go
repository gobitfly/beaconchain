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
func (d *epochToTotalAggregator) aggregateTotal(currentExportedEpoch uint64) error {
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
	err = d.dayUp.rollingAggregator.addToRolling(tx, "validator_dashboard_data_rolling_total", epochStart, epochEnd, 0)
	if err != nil {
		return errors.Wrap(err, "failed to add to rolling total")
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}
