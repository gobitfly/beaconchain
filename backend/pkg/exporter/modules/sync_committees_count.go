package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

type syncCommitteesCountExporter struct {
	db db2.ConsensusRepository

	delay time.Duration
	ctx   context.Context
}

func newSyncCommitteesCountExporter(ctx context.Context, db db2.ConsensusRepository) syncCommitteesCountExporter {
	return syncCommitteesCountExporter{
		db:    db,
		delay: time.Second * 12,
		ctx:   ctx,
	}
}

func (sc syncCommitteesCountExporter) Export() {
	for {
		select {
		case <-sc.ctx.Done():
			log.Info("sync committees count export loop cancelled")
			return
		default:
			startTime := time.Now()
			statusReporter := services.NewStatusReporter(constants.Event_ExporterLegacySyncCommitteesCount, constants.Default, time.Second*12)
			statusReporter.Report(constants.Running, nil)

			err := sc.processSyncCommitteesCount()
			if err != nil {
				log.Error(err, "error exporting sync_committees_count_per_validator", 0)
				statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
			} else {
				statusReporter.Report(constants.Success, map[string]string{
					"took":     time.Since(startTime).String(),
					"took_raw": fmt.Sprintf("%v", time.Since(startTime).Milliseconds()),
				})
			}

			time.Sleep(sc.delay)
		}
	}
}

func (sc *syncCommitteesCountExporter) processSyncCommitteesCount() error {
	rowCount, err := sc.db.GetSyncCommitteesCountPerValidator()
	if err != nil {
		return err
	}

	latestFinalizedEpoch, err := sc.db.GetLatestFinalizedEpoch()
	if err != nil {
		log.Error(err, "error retrieving latest exported finalized epoch from the database", 0)
	}

	currentPeriod := utils.SyncPeriodOfEpoch(latestFinalizedEpoch)
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.ClConfig.AltairForkEpoch)

	firstPeriod, countSoFar, err := sc.getEpochPeriodAndCountFromDB(rowCount, firstPeriod)
	if err != nil {
		return err
	}

	return sc.exportSyncCommitteesCount(firstPeriod, currentPeriod, countSoFar)
}

func (sc *syncCommitteesCountExporter) getEpochPeriodAndCountFromDB(rowCount uint64, firstPeriod uint64) (uint64, float64, error) {
	countSoFar := float64(0)
	if rowCount > 0 {
		dbPeriod, err := sc.db.GetTotalPeriodSyncCommitteesCountPerValidator()
		if err != nil {
			return 0, 0, err
		}

		if firstPeriod <= dbPeriod {
			// continue where we left off last time
			firstPeriod = dbPeriod + 1
		}

		countSoFar, err = sc.db.GetCountSoFarSyncCommitteesCountPerValidator(dbPeriod)
		if err != nil {
			return 0, 0, err
		}
	}

	return firstPeriod, countSoFar, nil
}

func (sc *syncCommitteesCountExporter) exportSyncCommitteesCount(firstPeriod, currentPeriod uint64, countSoFar float64) error {
	for period := firstPeriod; period <= currentPeriod; period++ {
		startTime := time.Now()

		count, err := sc.calculateCountForPeriod(period, countSoFar)
		if err != nil {
			return err
		}

		log.Infof("exporting sync committee count for period %v", period)

		err = sc.db.SaveSyncCommitteesCount(period, count)
		if err != nil {
			return fmt.Errorf("error exporting sync-committee count at period %v: %w", period, err)
		}

		log.InfoWithFields(log.Fields{
			"period":   period,
			"duration": time.Since(startTime),
		}, "exported sync_committees_count_per_validator")
	}

	return nil
}

func (sc *syncCommitteesCountExporter) calculateCountForPeriod(period uint64, countSoFar float64) (float64, error) {
	if period == 0 {
		return 0.0, nil
	}

	epoch := utils.FirstEpochOfSyncPeriod(period - 1)
	totalValidatorsCount, err := sc.db.GetEpochValidatorsCount(epoch)
	if err != nil {
		return 0, fmt.Errorf("error retrieving validators count for epoch %v: %v", epoch, err)
	}

	return countSoFar + (float64(utils.Config.Chain.ClConfig.SyncCommitteeSize) / float64(totalValidatorsCount)), nil
}
