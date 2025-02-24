package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

func UpdatePubkeyTag() {
	log.Infof("Started Pubkey Tags Updater")
	for {
		err := updatePubkeyTagOnce()
		if err != nil {
			log.Error(err, "error updating pubkey tags", 0)
		}
		time.Sleep(constants.Duration10Mins)
	}
}

func updatePubkeyTagOnce() error {
	startTime := time.Now()
	statusReport := createPubkeyTagStatusReport()
	statusReport(constants.Running, nil)

	err := db.UpdatePubkeyTags()
	if err != nil {
		handlePubkeyTagsError(err, statusReport)
		return err
	}

	log.Infof("Updating Pubkey Tags took %v sec.", time.Since(startTime).Seconds())
	handlePubkeyTagSuccess(startTime, statusReport)
	metrics.TaskDuration.WithLabelValues("validator_pubkey_tag_updater").Observe(time.Since(startTime).Seconds())

	return nil
}

func createPubkeyTagStatusReport() func(status constants.StatusType, metadata map[string]string) {
	return services.NewStatusReport(constants.Event_ExporterLegacyPubkeyTags, constants.Duration10Mins, time.Second*12, utils.Config.DeploymentType)
}

func handlePubkeyTagsError(err error, statusReport func(status constants.StatusType, metadata map[string]string)) {
	log.Error(err, "error updating validator_tags", 0)
	statusReport(constants.Failure, map[string]string{"error": err.Error()})
}

func handlePubkeyTagSuccess(startTime time.Time, statusReport func(status constants.StatusType, metadata map[string]string)) {
	statusReport(constants.Success, map[string]string{
		"took":     time.Since(startTime).String(),
		"took_raw": fmt.Sprintf("%v", time.Since(startTime).Milliseconds()),
	})
}
