package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

func syncCommitteesExporter(rpcClient rpc.Client) {
	for {
		startTime := time.Now()
		statusReport := createStatusReport()
		statusReport(constants.Running, nil)

		err := exportSyncCommittees(rpcClient)
		if err != nil {
			handleExportError(err, startTime, statusReport)
		} else {
			handleExportSuccess(startTime, statusReport)
		}

		time.Sleep(time.Second * 12)
	}
}

func createStatusReport() func(status constants.StatusType, metadata map[string]string) {
	return services.NewStatusReport(constants.Event_ExporterLegacySyncCommittees, constants.Default, time.Second*12, utils.Config.DeploymentType)
}

func handleExportError(err error, startTime time.Time, statusReport func(status constants.StatusType, metadata map[string]string)) {
	log.Error(err, "error exporting sync_committees", 0, map[string]interface{}{"duration": time.Since(startTime)})
	statusReport(constants.Failure, map[string]string{"error": err.Error()})
}

func handleExportSuccess(startTime time.Time, statusReport func(status constants.StatusType, metadata map[string]string)) {
	statusReport(constants.Success, map[string]string{
		"took":     time.Since(startTime).String(),
		"took_raw": fmt.Sprintf("%v", time.Since(startTime).Milliseconds()),
	})
}

func exportSyncCommittees(rpcClient rpc.Client) error {
	dbPeriods, err := db.GetSyncCommitteesPeriods()
	if err != nil {
		return err
	}

	dbPeriodsMap := createDBPeriodsMap(dbPeriods)
	firstPeriod, lastPeriod := calculateSyncPeriodRange()

	for period := firstPeriod; period <= lastPeriod; period++ {
		_, exists := dbPeriodsMap[period]
		if !exists {
			if err := ExportSyncCommitteeData(rpcClient, period); err != nil {
				return fmt.Errorf("error exporting sync-committee at period %v: %w", period, err)
			}
		}
	}

	return nil
}

func createDBPeriodsMap(dbPeriods []uint64) map[uint64]bool {
	dbPeriodsMap := make(map[uint64]bool, len(dbPeriods))
	for _, period := range dbPeriods {
		dbPeriodsMap[period] = true
	}
	return dbPeriodsMap
}

func calculateSyncPeriodRange() (uint64, uint64) {
	currEpoch := cache.LatestFinalizedEpoch.Get()
	if currEpoch > 0 { // guard against underflows
		currEpoch = currEpoch - 1
	}
	lastPeriod := utils.SyncPeriodOfEpoch(currEpoch) + 1 // we can look into the future
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.ClConfig.AltairForkEpoch)
	return firstPeriod, lastPeriod
}

func ExportSyncCommitteeData(rpcClient rpc.Client, period uint64) error {
	startTime := time.Now()
	data, err := GetSyncCommitteAtPeriod(rpcClient, period)
	if err != nil {
		return err
	}

	args, ids := parseSyncArgsAndIDs(data)

	err = db.SaveSyncCommitteeData(args, ids)
	if err != nil {
		return err
	}

	log.InfoWithFields(log.Fields{
		"period":   period,
		"epoch":    utils.FirstEpochOfSyncPeriod(period),
		"duration": time.Since(startTime),
	}, "exported sync_committee")

	return nil
}

func parseSyncArgsAndIDs(data []SyncCommittee) ([]interface{}, []string) {
	nArgs := 3
	args := make([]interface{}, len(data)*nArgs)
	ids := make([]string, len(data))
	for i, entry := range data {
		args[i*nArgs+0] = entry.Period
		args[i*nArgs+1] = entry.ValidatorIndex
		args[i*nArgs+2] = entry.CommitteeIndex
		ids[i] = fmt.Sprintf("($%d,$%d,$%d)", i*nArgs+1, i*nArgs+2, i*nArgs+3)
	}

	return args, ids
}

func GetSyncCommitteAtPeriod(rpcClient rpc.Client, period uint64) ([]SyncCommittee, error) {
	stateID, epoch := calculateStateIDAndEpoch(period)
	firstEpoch := utils.FirstEpochOfSyncPeriod(period)
	lastEpoch := firstEpoch + utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod - 1

	log.Infof("exporting sync committee assignments for period %v (epoch %v to %v)", period, firstEpoch, lastEpoch)

	// Note that the order we receive the validators from the node in is crucial
	// and determines which bit reflects them in the block sync aggregate bits
	syncCommittee, err := rpcClient.GetSyncCommittee(fmt.Sprintf("%d", stateID), epoch)
	if err != nil {
		return nil, err
	}

	return parseSyncCommitteeResult(syncCommittee, period), nil
}

func calculateStateIDAndEpoch(period uint64) (uint64, uint64) {
	stateID := calculateStateID(period)
	epoch := utils.FirstEpochOfSyncPeriod(period)

	if stateID/utils.Config.Chain.ClConfig.SlotsPerEpoch <= utils.Config.Chain.ClConfig.AltairForkEpoch {
		stateID = utils.Config.Chain.ClConfig.AltairForkEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		epoch = utils.Config.Chain.ClConfig.AltairForkEpoch
	}

	return stateID, epoch
}

func calculateStateID(period uint64) uint64 {
	if period > 0 {
		return utils.FirstEpochOfSyncPeriod(period-1) * utils.Config.Chain.ClConfig.SlotsPerEpoch
	}
	return 0
}

func parseSyncCommitteeResult(c *types.StandardSyncCommittee, period uint64) []SyncCommittee {
	result := []SyncCommittee{}
	for i, idxStr := range c.Validators {
		result = append(result, SyncCommittee{
			Period:         period,
			ValidatorIndex: uint64(idxStr),
			CommitteeIndex: uint64(i),
		})
	}
	return result
}

type SyncCommittee struct {
	Period         uint64 `json:"period"`
	ValidatorIndex uint64 `json:"validatorindex"`
	CommitteeIndex uint64 `json:"committeeindex"`
}
