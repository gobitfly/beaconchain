package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

type networkLivenessUpdater struct {
	client rpc.EpochClient
	db     db.ConsensusDBI
	ctx    context.Context
}

func newNetworkLivenessUpdater(client rpc.Client, db db.ConsensusDBI) networkLivenessUpdater {
	return networkLivenessUpdater{
		client: client,
		db:     db,
		ctx:    context.Background(),
	}
}

func (n networkLivenessUpdater) Export() {
	select {
	case <-n.ctx.Done():
		log.Info("export loop cancelled", 0)
		return
	default:
		prevHeadEpoch, err := n.db.GetNetworkLivenessPreviousHeadEpoch()
		if err != nil {
			log.Fatal(err, "getting previous head epoch from db error", 0)
		}

		epochDuration := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot*utils.Config.Chain.ClConfig.SlotsPerEpoch)
		slotDuration := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot)

		for {
			deployment := utils.Config.DeploymentType
			statusReport := services.NewStatusReport(constants.Event_ExporterLegacyNetworkLiveness, constants.Default, slotDuration, deployment)
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

			err = updateCache(head)
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

func updateCache(head *types.ChainHead) error {
	if err := cache.LatestNodeEpoch.Set(head.HeadEpoch); err != nil {
		return fmt.Errorf("error setting latestNodeEpoch in cache")
	}
	if err := cache.LatestNodeFinalizedEpoch.Set(head.FinalizedEpoch); err != nil {
		return fmt.Errorf("error setting latestNodeFinalizedEpoch in cache")
	}
	return nil
}
