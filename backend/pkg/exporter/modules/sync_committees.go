package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	db2 "github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

type SyncCommitteeClient interface {
	GetSyncCommittee(stateID string, epoch uint64) (*constypes.StandardSyncCommittee, error)
}

type syncCommitteesExporter struct {
	client SyncCommitteeClient
	db     db2.ConsensusRepository

	delay time.Duration
	ctx   context.Context
	cache *cache.TieredCacheBase
}

func NewSyncCommitteesExporter(ctx context.Context, client rpc.Client, db db2.ConsensusRepository) syncCommitteesExporter {
	if cache.TieredCache == nil {
		log.Fatal(nil, "TieredCache is not initialised", 0)
	}

	return syncCommitteesExporter{
		client: client,
		db:     db,
		delay:  time.Second * 12,
		ctx:    ctx,
		cache:  cache.TieredCache,
	}
}

func (s syncCommitteesExporter) Export() {
	for {
		select {
		case <-s.ctx.Done():
			log.Info("sync committees export loop cancelled")
			return
		default:
			startTime := time.Now()
			statusReporter := services.NewStatusReporter(constants.Event_ExporterLegacySyncCommittees, constants.Default, time.Second*12)
			statusReporter.Report(constants.Running, nil)

			err := s.exportSyncCommittees()
			if err != nil {
				log.Error(err, "error exporting sync_committees", 0, map[string]interface{}{"duration": time.Since(startTime)})
				statusReporter.Report(constants.Failure, map[string]string{"error": err.Error()})
			}
			statusReporter.Report(constants.Success, map[string]string{
				"took":     time.Since(startTime).String(),
				"took_raw": fmt.Sprintf("%v", time.Since(startTime).Milliseconds()),
			})

			time.Sleep(s.delay)
		}
	}
}

func (s *syncCommitteesExporter) exportSyncCommittees() error {
	dbPeriods, err := s.db.GetSyncCommitteesPeriods()
	if err != nil {
		return err
	}

	dbPeriodsMap := make(map[uint64]bool, len(dbPeriods))
	for _, period := range dbPeriods {
		dbPeriodsMap[period] = true
	}

	latestNodeFinalizedEpochKey := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)
	currEpoch, err := s.cache.GetUint64WithLocalTimeout(latestNodeFinalizedEpochKey, time.Second*5)
	if err != nil {
		return fmt.Errorf("error getting latestNodeEpoch from cache: %w", err)
	}
	if currEpoch > 0 { // guard against underflows
		currEpoch = currEpoch - 1
	}

	lastPeriod := utils.SyncPeriodOfEpoch(currEpoch) + 1 // we can look into the future
	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.ClConfig.AltairForkEpoch)

	for period := firstPeriod; period <= lastPeriod; period++ {
		if _, exists := dbPeriodsMap[period]; exists {
			continue
		}
		if err := s.ExportSyncCommitteeData(period); err != nil {
			return fmt.Errorf("error exporting sync-committee at period %v: %w", period, err)
		}
	}

	return nil
}

func (s *syncCommitteesExporter) ExportSyncCommitteeData(period uint64) error {
	startTime := time.Now()
	defer func(startTime time.Time) {
		log.InfoWithFields(log.Fields{
			"period":   period,
			"epoch":    utils.FirstEpochOfSyncPeriod(period),
			"duration": time.Since(startTime),
		}, "exported sync_committee")
	}(startTime)

	data, err := s.GetSyncCommitteAtPeriod(period)
	if err != nil {
		return err
	}

	err = s.db.SaveSyncCommitteeData(data)
	if err != nil {
		return err
	}

	return nil
}

func (s *syncCommitteesExporter) GetSyncCommitteAtPeriod(period uint64) ([]types.SyncCommittee, error) {
	stateID, epoch := calculateStateIDAndEpoch(period)
	firstEpoch := utils.FirstEpochOfSyncPeriod(period)
	lastEpoch := firstEpoch + utils.Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod - 1

	log.Infof("exporting sync committee assignments for period %v (epoch %v to %v)", period, firstEpoch, lastEpoch)

	// Note that the order we receive the validators from the node in is crucial
	// and determines which bit reflects them in the block sync aggregate bits
	syncCommittee, err := s.client.GetSyncCommittee(fmt.Sprintf("%d", stateID), epoch)
	if err != nil {
		return nil, err
	}

	syncCommitteeResults := []types.SyncCommittee{}
	for i, idxStr := range syncCommittee.Validators {
		syncCommitteeResults = append(syncCommitteeResults, types.SyncCommittee{
			Period:         period,
			ValidatorIndex: uint64(idxStr),
			CommitteeIndex: uint64(i),
		})
	}

	return syncCommitteeResults, nil
}

func calculateStateIDAndEpoch(period uint64) (uint64, uint64) {
	stateID := uint64(0)
	if period > 0 {
		stateID = utils.FirstEpochOfSyncPeriod(period-1) * utils.Config.Chain.ClConfig.SlotsPerEpoch
	}
	epoch := utils.FirstEpochOfSyncPeriod(period)

	if stateID/utils.Config.Chain.ClConfig.SlotsPerEpoch <= utils.Config.Chain.ClConfig.AltairForkEpoch {
		stateID = utils.Config.Chain.ClConfig.AltairForkEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
		epoch = utils.Config.Chain.ClConfig.AltairForkEpoch
	}

	return stateID, epoch
}
