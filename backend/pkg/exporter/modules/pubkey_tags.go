package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

type pubkeyUpdater struct {
	db    db.ConsensusDBI
	delay time.Duration
	ctx   context.Context
}

func newPubkeyUpdater(ctx context.Context, db db.ConsensusDBI) pubkeyUpdater {
	return pubkeyUpdater{
		db:    db,
		delay: time.Minute * 10,
		ctx:   ctx,
	}
}

func (p *pubkeyUpdater) Update() {
	log.Infof("Started Pubkey Tags Updater")
	for {
		startTime := time.Now()
		statusReport := services.NewStatusReport(constants.Event_ExporterLegacyPubkeyTags, utils.Config.DeploymentType, p.delay, time.Second*12)
		statusReport(constants.Running, nil)

		err := p.db.UpdatePubkeyTags()
		if err != nil {
			log.Error(err, "error updating validator_tags", 0)
			statusReport(constants.Failure, map[string]string{"error": err.Error()})
		}

		log.Infof("Updating Pubkey Tags took %v sec.", time.Since(startTime).Seconds())

		statusReport(constants.Success, map[string]string{
			"took":     time.Since(startTime).String(),
			"took_raw": fmt.Sprintf("%v", time.Since(startTime).Milliseconds()),
		})

		metrics.TaskDuration.WithLabelValues("validator_pubkey_tag_updater").Observe(time.Since(startTime).Seconds())

		time.Sleep(p.delay)
	}
}
