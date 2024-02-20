package modules

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func genesisDepositsExporter(client rpc.Client) {
	for {
		// check if the beaconchain has started
		var latestEpoch uint64
		err := db.WriterDb.Get(&latestEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
		if err != nil {
			logger.Errorf("error retrieving latest epoch from the database: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		if latestEpoch == 0 {
			time.Sleep(time.Minute)
			continue
		}

		// check if genesis-deposits have already been exported
		var genesisDepositsCount uint64
		err = db.WriterDb.Get(&genesisDepositsCount, "SELECT COUNT(*) FROM blocks_deposits WHERE block_slot=0")
		if err != nil {
			logger.Errorf("error retrieving genesis-deposits-count when exporting genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		// if genesis-deposits have already been exported exit this go-routine
		if genesisDepositsCount > 0 {
			return
		}

		genesisValidators, err := client.GetValidatorState(0)
		if err != nil {
			logger.Errorf("error retrieving genesis validator data for genesis-epoch when exporting genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		tx, err := db.WriterDb.Beginx()
		if err != nil {
			logger.Errorf("error beginning db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		logger.Infof("exporting deposit data for %v genesis validators", len(genesisValidators.Data))
		for i, validator := range genesisValidators.Data {
			if i%1000 == 0 {
				logger.Infof("exporting deposit data for genesis validator %v (%v/%v)", validator.Index, i, len(genesisValidators.Data))
			}
			_, err = tx.Exec(`INSERT INTO blocks_deposits (block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature)
			VALUES (0, '\x01', $1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
				validator.Index, utils.MustParseHex(validator.Validator.Pubkey), utils.MustParseHex(validator.Validator.WithdrawalCredentials), validator.Balance, []byte{0x0},
			)
			if err != nil {
				err := tx.Rollback()
				if err != nil {
					utils.LogError(err, "error rolling back transaction", 0)
				}
				logger.Errorf("error exporting genesis-deposits: %v", err)
				time.Sleep(time.Minute)
				continue
			}
		}

		// hydrate the eth1 deposit signature for all genesis validators that have a corresponding eth1 deposit
		_, err = tx.Exec(`
			UPDATE blocks_deposits 
			SET signature = a.signature 
			FROM (
				SELECT DISTINCT ON(publickey) publickey, signature 
				FROM eth1_deposits 
				WHERE valid_signature = true) AS a 
			WHERE block_slot = 0 AND blocks_deposits.publickey = a.publickey AND blocks_deposits.signature = '\x'`)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				utils.LogError(err, "error rolling back transaction", 0)
			}
			logger.Errorf("error hydrating eth1 data into genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		// update deposits-count
		_, err = tx.Exec("UPDATE blocks SET depositscount = $1 WHERE slot = 0", len(genesisValidators.Data))
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				utils.LogError(err, "error rolling back transaction", 0)
			}
			logger.Errorf("error updating deposit count for the genesis slot: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		err = tx.Commit()
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				utils.LogError(err, "error rolling back transaction", 0)
			}
			logger.Errorf("error committing db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		logger.Infof("exported genesis-deposits for %v genesis-validators", len(genesisValidators.Data))
		return
	}
}
