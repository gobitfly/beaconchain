package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// we wrap all our errors in this codebase

func (d *dashboardData) rollingTask() {
	for {
		// loop to complete incomplete epochs
		err := d.handleRollings()
		if err != nil {
			d.log.Error(err, "failed to handle rollings", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		time.Sleep(10 * time.Second)
	}
}

// max rollings to do at the same time
var RollingsAtOnce = 3

// max sub parts to do in a single rolling at the same time
var RollingPartsInParallel = 3

func (d *dashboardData) handleRollings() error {
	// fork for every rolling we have to do
	rollings := []edb.Rollings{
		edb.Rolling1h,
		edb.Rolling24h,
		edb.Rolling7d,
		edb.Rolling30d,
		edb.Rolling90d,
		edb.RollingTotal,
	}
	// but lets limit to x rollings
	eg := errgroup.Group{}
	eg.SetLimit(RollingsAtOnce)
	for _, rolling := range rollings {
		rolling := rolling
		eg.Go(func() error {
			return d.doRollingCheck(rolling)
		})
	}
	err := eg.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to do rollings")
	}
	return nil
}

func (d *dashboardData) doRollingCheck(rolling edb.Rollings) error {
	finishedEpoch, err := edb.GetLatestFinishedEpoch()
	if err != nil {
		return errors.Wrap(err, "failed to get latest finished epoch")
	}
	if finishedEpoch < 0 {
		d.log.Infof("no finished epoch yet")
		return nil
	}
	metrics.State.WithLabelValues("dashboard_data_exporter_finished_epoch").Set(float64(finishedEpoch))
	// if finishedEpoch is not the same as the safeepoch we skip updating the rolling so resyncing after falling back is fast
	if safeEpoch := d.latestSafeEpoch.Load(); finishedEpoch != safeEpoch {
		d.log.Infof("skipping rolling %s update, finished epoch %d, safe epoch %d", rolling, finishedEpoch, safeEpoch)
		return nil
	}
	rollingEpoch, err := edb.GetRollingLastEpoch(rolling)
	if err != nil {
		return errors.Wrap(err, "failed to get rolling last epoch")
	}
	metrics.State.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_rolling_%s_epoch", rolling)).Set(float64(rollingEpoch))
	if rollingEpoch >= finishedEpoch {
		d.log.Debugf("rolling %s is up to date", rolling)
		return nil
	}
	d.log.Infof("rolling %s is outdated, latest epoch %d, latest finished epoch %d", rolling, rollingEpoch, finishedEpoch)
	// update metric after run
	defer func() {
		rollingEpoch, err := edb.GetRollingLastEpoch(rolling)
		if err != nil {
			d.log.Error(err, "failed to get rolling last epoch", 0)
			return
		}
		metrics.State.WithLabelValues(fmt.Sprintf("dashboard_data_exporter_rolling_%s_epoch", rolling)).Set(float64(rollingEpoch))
	}()

	// next, nuke the unsafe rolling tables to prepare them for us to fill them
	err = edb.NukeUnsafeRollingTable(rolling)
	if err != nil {
		return errors.Wrap(err, "failed to nuke unsafe rolling table")
	}
	// now we fetch the start & end for each pre-aggregated table we use

	minTs := utils.EpochToTime(uint64(finishedEpoch)).Add(-rolling.GetDuration())
	d.log.Infof("rolling %s, min ts %s", rolling, minTs)
	tables := []edb.RollingSources{
		edb.RollingSourceMonthly,
		edb.RollingSourceDaily,
		edb.RollingSourceHourly,
		edb.RollingSourceEpochly,
	}
	minMaxMap := make(map[edb.RollingSources]*edb.MinMax)
	var lowestSeenTs *time.Time
	for _, table := range tables {
		minmax, err := edb.GetMinMaxForRollingSource(table, minTs, lowestSeenTs)
		if err != nil {
			return errors.Wrap(err, "failed to get min max for rolling source")
		}
		if minmax == nil {
			//d.log.Debug("rolling %s, source %s, no data", rolling, table)
			continue
		}
		minMaxMap[table] = minmax
		d.log.Infof("rolling %s, source %s, min %s, max %s", rolling, table, minmax.Min, minmax.Max)
		lowestSeenTs = minmax.Min
	}
	// now the transfer logic for each source
	eg := errgroup.Group{}
	eg.SetLimit(RollingPartsInParallel)
	for source, minmax := range minMaxMap {
		if minmax == nil {
			continue
		}
		source := source
		minmax := minmax
		eg.Go(func() error {
			d.log.Infof("transferring rolling source %s to rolling %s", source, rolling)
			start := time.Now()
			err := edb.TransferRollingSourceToRolling(rolling, source, *minmax)
			if err != nil {
				return errors.Wrap(err, "failed to transfer rolling source to rolling")
			}
			d.log.Infof("transferred rolling source %s to rolling %s in %s", source, rolling, time.Since(start))
			return nil
		})
	}
	err = eg.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to transfer all rolling sources")
	}
	// now we swap the tables
	err = edb.SwapRollingTables(rolling)
	if err != nil {
		return errors.Wrap(err, "failed to swap rolling tables")
	}
	return nil
}

/*

	return nil // not yet implemented
}
*/
