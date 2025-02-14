package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

func syncCommitteesCountExporter() {
	for {
		startTime := time.Now()
		statusReport := createCommitteesCountStatusReport()
		statusReport(constants.Running, nil)

		err := processSyncCommitteesCount()
		if err != nil {
			handleCommitteesCountError(err, statusReport)
		} else {
			handleCommitteesCountSuccess(startTime, statusReport)
		}

		time.Sleep(time.Second * 12)
	}
}

func createCommitteesCountStatusReport() func(status constants.StatusType, metadata map[string]string) {
	return services.NewStatusReport(constants.Event_ExporterLegacySyncCommitteesCount, constants.Default, time.Second*12)
}

func handleCommitteesCountError(err error, statusReport func(status constants.StatusType, metadata map[string]string)) {
	log.Error(err, "error exporting sync_committees_count_per_validator", 0)
	statusReport(constants.Failure, map[string]string{"error": err.Error()})
}

func handleCommitteesCountSuccess(startTime time.Time, statusReport func(status constants.StatusType, metadata map[string]string)) {
	statusReport(constants.Success, map[string]string{
		"took":     time.Since(startTime).String(),
		"took_raw": fmt.Sprintf("%v", time.Since(startTime).Milliseconds()),
	})
}

func processSyncCommitteesCount() error {
	rowCount, err := db.GetSyncCommitteesCountPerValidator()
	if err != nil {
		return err
	}

	latestFinalizedEpoch, err := db.GetLatestFinalizedEpoch()
	if err != nil {
		log.Error(err, "error retrieving latest exported finalized epoch from the database", 0)
	}

	currentPeriod := utils.SyncPeriodOfEpoch(latestFinalizedEpoch)
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.ClConfig.AltairForkEpoch)

	firstPeriod, countSoFar, err := getEpochPeriodAndCountFromDB(rowCount, firstPeriod)
	if err != nil {
		return err
	}

	return exportSyncCommitteesCount(firstPeriod, currentPeriod, countSoFar)
}

func getEpochPeriodAndCountFromDB(rowCount uint64, firstPeriod uint64) (uint64, float64, error) {
	countSoFar := float64(0)
	if rowCount > 0 {
		dbPeriod, err := db.GetTotalPeriodSyncCommitteesCountPerValidator()
		if err != nil {
			return 0, 0, err
		}

		if firstPeriod <= dbPeriod {
			// continue where we left off last time
			firstPeriod = dbPeriod + 1
		}

		countSoFar, err = db.GetCountSoFarSyncCommitteesCountPerValidator(dbPeriod)
		if err != nil {
			return 0, 0, err
		}
	}

	return firstPeriod, countSoFar, nil
}

func exportSyncCommitteesCount(firstPeriod, currentPeriod uint64, countSoFar float64) error {
	for period := firstPeriod; period <= currentPeriod; period++ {
		timeStart := time.Now()

		count, err := calculateCountForPeriod(period, countSoFar)
		if err != nil {
			return err
		}

		log.Infof("exporting sync committee count for period %v", period)

		err = db.SaveSyncCommitteesCount(period, count)
		if err != nil {
			return fmt.Errorf("error exporting sync-committee count at period %v: %w", period, err)
		}

		log.InfoWithFields(log.Fields{
			"period":   period,
			"duration": time.Since(timeStart),
		}, "exported sync_committees_count_per_validator")
	}

	return nil
}

func calculateCountForPeriod(period uint64, countSoFar float64) (float64, error) {
	if period == 0 {
		return 0.0, nil
	}

	epoch := utils.FirstEpochOfSyncPeriod(period - 1)
	totalValidatorsCount, err := db.GetEpochValidatorsCount(epoch)
	if err != nil {
		return 0, fmt.Errorf("error retrieving validators count for epoch %v: %v", epoch, err)
	}

	return countSoFar + (float64(utils.Config.Chain.ClConfig.SyncCommitteeSize) / float64(totalValidatorsCount)), nil
}
