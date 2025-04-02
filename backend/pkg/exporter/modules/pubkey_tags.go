package modules

import (
	"context"
	"fmt"
	"time"

	db2 "github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

type pubkeyTagsUpdater struct {
	db    db2.ConsensusRepository
	delay time.Duration
	ctx   context.Context
}

func newPubkeyTagsUpdater(ctx context.Context, db db2.ConsensusRepository) pubkeyTagsUpdater {
	return pubkeyTagsUpdater{
		db:    db,
		delay: time.Minute * 10,
		ctx:   ctx,
	}
}

func (p *pubkeyTagsUpdater) Update() {
	log.Infof("Started Pubkey Tags Updater")
	for {
		select {
		case <-p.ctx.Done():
			log.Info("update loop cancelled")
			return
		default:
			startTime := time.Now()
			statusReport := services.StatusReporter.NewStatusReport(constants.Event_ExporterLegacyPubkeyTags, p.delay, time.Second*12)
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
}
