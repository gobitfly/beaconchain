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
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

type EpochClient interface {
	GetChainHead() (*types.ChainHead, error)
	GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error)
	GetEpochData(epoch uint64, skipHistoricBalances bool) (*types.EpochData, error)
	GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error)
}

type networkLivenessUpdater struct {
	client EpochClient
	db     db2.ConsensusRepository
	ctx    context.Context
	cache  *cache.TieredCacheBase
}

func newNetworkLivenessUpdater(ctx context.Context, client rpc.Client, db db2.ConsensusRepository) networkLivenessUpdater {
	if cache.TieredCache == nil {
		log.Fatal(nil, "TieredCache is not initialised", 0)
	}

	return networkLivenessUpdater{
		client: client,
		db:     db,
		ctx:    ctx,
		cache:  cache.TieredCache,
	}
}

func (n *networkLivenessUpdater) Export() {
	prevHeadEpoch, err := n.db.GetNetworkLivenessPreviousHeadEpoch()
	if err != nil {
		log.Fatal(err, "getting previous head epoch from db error", 0)
	}

	epochDuration := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	slotDuration := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot)

	for {
		select {
		case <-n.ctx.Done():
			log.Info("network liveness export loop cancelled")
			return
		default:
			statusReport := services.NewStatusReport(constants.Event_ExporterLegacyNetworkLiveness, utils.Config.DeploymentType, constants.Default, slotDuration)
			statusReport(constants.Running, nil)

			head, err := n.client.GetChainHead()
			if err != nil {
				log.Error(err, "error getting chainhead when exporting network liveness", 0)
				statusReport(constants.Failure, map[string]string{"error": err.Error()})
				time.Sleep(slotDuration)
				continue
			}

			if prevHeadEpoch == head.HeadEpoch {
				statusReport(constants.Success, nil)
				time.Sleep(slotDuration)
				continue
			}

			// wait for node to be synced
			if nodeNotSynced(head.HeadEpoch, epochDuration) {
				statusReport(constants.Failure, map[string]string{"error": "node not synced"})
				time.Sleep(slotDuration)
				continue
			}

			err = n.db.SaveNetworkLivenessData(head)
			if err != nil {
				log.Error(err, "error saving network liveness in db", 0)
				statusReport(constants.Failure, map[string]string{"error": err.Error()})
			} else {
				log.Infof("updated network liveness for epoch %v", head.HeadEpoch)
				prevHeadEpoch = head.HeadEpoch
			}

			err = n.updateCache(head)
			if err != nil {
				log.Error(err, "error updating cache", 0)
				statusReport(constants.Failure, map[string]string{"error": err.Error()})
			}

			statusReport(constants.Success, nil)

			time.Sleep(slotDuration)
		}
	}
}

func nodeNotSynced(headEpoch uint64, epochDuration time.Duration) bool {
	return time.Now().Add(-epochDuration).After(utils.EpochToTime(headEpoch))
}

func (n *networkLivenessUpdater) updateCache(head *types.ChainHead) error {
	latestNodeEpochKey := fmt.Sprintf("%d:frontend:latestNodeFinalizedEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	latestNodeFinalizedEpochKey := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)

	if err := n.cache.SetUint64(latestNodeEpochKey, head.HeadEpoch, utils.Day); err != nil {
		return fmt.Errorf("error setting latestNodeEpoch in cache: %w", err)
	}
	if err := n.cache.SetUint64(latestNodeFinalizedEpochKey, head.FinalizedEpoch, utils.Day); err != nil {
		return fmt.Errorf("error setting latestNodeFinalizedEpoch in cache: %w", err)
	}

	return nil
}
