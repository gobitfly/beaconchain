package modules

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

func networkLivenessUpdater(client rpc.Client) {
	var prevHeadEpoch uint64
	err := db.WriterDb.Get(&prevHeadEpoch, "SELECT COALESCE(MAX(headepoch), 0) FROM network_liveness")
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

		// wait for node to be synced
		if time.Now().Add(-epochDuration).After(utils.EpochToTime(head.HeadEpoch)) {
			r(constants.Failure, map[string]string{"error": "node not synced"})
			time.Sleep(slotDuration)
			continue
		}

		_, err = db.WriterDb.Exec(`
			INSERT INTO network_liveness (ts, headepoch, finalizedepoch, justifiedepoch, previousjustifiedepoch)
			VALUES (NOW(), $1, $2, $3, $4)`,
			head.HeadEpoch, head.FinalizedEpoch, head.JustifiedEpoch, head.PreviousJustifiedEpoch)
		if err != nil {
			log.Error(err, "error saving networkliveness", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
		} else {
			log.Infof("updated networkliveness for epoch %v", head.HeadEpoch)
			prevHeadEpoch = head.HeadEpoch
		}

		err = cache.LatestNodeEpoch.Set(head.HeadEpoch)
		if err != nil {
			log.Error(err, "error setting latestNodeEpoch in cache", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
		}

		err = cache.LatestNodeFinalizedEpoch.Set(head.FinalizedEpoch)
		if err != nil {
			log.Error(err, "error setting latestNodeFinalizedEpoch in cache", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
		}
		r(constants.Success, nil)

		time.Sleep(slotDuration)
	}
}
