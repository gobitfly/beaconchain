package modules

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
	"github.com/jmoiron/sqlx"
)

func networkLivenessUpdater(client rpc.Client) {
	prevHeadEpoch, err := getPreviousHeadEpoch(db.WriterDb)
	if err != nil {
		log.Fatal(err, "getting previous head epoch from db error", 0)
	}

	epochDuration := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	slotDuration := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot)

	for {
		r := services.NewStatusReport(constants.Event_ExporterLegacyNetworkLiveness, constants.Default, slotDuration)
		r(constants.Running, nil)

		head, err := client.GetChainHead()
		if err != nil {
			log.Error(err, "error getting chainhead when exporting networkliveness", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
			time.Sleep(slotDuration)
			continue
		}

		if prevHeadEpoch == head.HeadEpoch {
			r(constants.Success, nil)
			time.Sleep(slotDuration)
			continue
		}

		if !isNodeSynced(head.HeadEpoch, epochDuration) {
			r(constants.Failure, map[string]string{"error": "node not synced"})
			time.Sleep(slotDuration)
			continue
		}

		if err := db.SaveNetworkLivenessData(head); err != nil {
			log.Error(err, "error saving networkliveness in db", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
		} else {
			log.Infof("updated networkliveness for epoch %v", head.HeadEpoch)
			prevHeadEpoch = head.HeadEpoch
		}

		if err := updateCache(head); err != nil {
			log.Error(err, "error updating cache", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
		}

		r(constants.Success, nil)
		time.Sleep(slotDuration)
	}
}

func getPreviousHeadEpoch(db *sqlx.DB) (uint64, error) {
	var prevHeadEpoch uint64
	err := db.Get(&prevHeadEpoch, "SELECT COALESCE(MAX(headepoch), 0) FROM network_liveness")
	if err != nil {
		return 0, err
	}
	return prevHeadEpoch, nil
}

func isNodeSynced(headEpoch uint64, epochDuration time.Duration) bool {
	return !time.Now().Add(-epochDuration).After(utils.EpochToTime(headEpoch))
}

func updateCache(head *types.ChainHead) error {
	if err := cache.LatestNodeEpoch.Set(head.HeadEpoch); err != nil {
		return err
	}
	if err := cache.LatestNodeFinalizedEpoch.Set(head.FinalizedEpoch); err != nil {
		return err
	}
	return nil
}
