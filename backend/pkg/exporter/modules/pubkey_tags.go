package modules

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	monitoringServices "github.com/gobitfly/beaconchain/pkg/monitoring/services"
)

func UpdatePubkeyTag() {
	log.Infof("Started Pubkey Tags Updater")
	delay := time.Minute * 10
	for {
		start := time.Now()
		r := monitoringServices.NewStatusReport(constants.Event_ExporterLegacyPubkeyTags, delay, time.Second*12)
		r(constants.Running, nil)
		tx, err := db.WriterDb.Beginx()
		if err != nil {
			log.Error(err, "Error connecting to DB", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
			// return err
		}
		_, err = tx.Exec(`INSERT INTO validator_tags (publickey, tag)
		SELECT publickey, FORMAT('pool:%s', sps.name) tag
		FROM eth1_deposits
		inner join stake_pools_stats as sps on ENCODE(from_address::bytea, 'hex')=sps.address
		WHERE sps.name NOT LIKE '%Rocketpool -%'
		ON CONFLICT (publickey, tag) DO NOTHING;`)
		if err != nil {
			log.Error(err, "error updating validator_tags", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
			// return err
		}

		err = tx.Commit()
		if err != nil {
			log.Error(err, "error committing transaction", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
		}
		_ = tx.Rollback()

		log.Infof("Updating Pubkey Tags took %v sec.", time.Since(start).Seconds())
		r(constants.Success, map[string]string{"took": time.Since(start).String(), "took_raw": time.Since(start).String()})
		metrics.TaskDuration.WithLabelValues("validator_pubkey_tag_updater").Observe(time.Since(start).Seconds())

		time.Sleep(delay)
	}
}
