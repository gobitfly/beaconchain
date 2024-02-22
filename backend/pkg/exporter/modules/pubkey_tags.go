package modules

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
)

func UpdatePubkeyTag() {
	log.Infof("Started Pubkey Tags Updater")
	for {
		start := time.Now()

		tx, err := db.WriterDb.Beginx()
		if err != nil {
			log.Error(err, "Error connecting to DB", 0)
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
			// return err
		}

		err = tx.Commit()
		if err != nil {
			log.Error(err, "error committing transaction", 0)
		}
		_ = tx.Rollback()

		log.Infof("Updating Pubkey Tags took %v sec.", time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("validator_pubkey_tag_updater").Observe(time.Since(start).Seconds())

		time.Sleep(time.Minute * 10)
	}
}
