package modules

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/jmoiron/sqlx"
)

func genesisDepositsExporter(client rpc.Client) {
	for {
		if !isBeaconChainStarted() {
			time.Sleep(time.Minute)
			continue
		}

		if areGenesisDepositsExported() {
			return
		}

		genesisValidators, err := client.GetValidatorState(0)
		if err != nil {
			log.Error(err, "error retrieving genesis validator data for genesis-epoch when exporting genesis-deposits: %v", 0)
			time.Sleep(time.Minute)
			continue
		}

		err = exportGenesisDeposits(genesisValidators)
		if err != nil {
			log.Error(err, "error exporting genesis-deposits: %v", 0)
			time.Sleep(time.Minute)
			continue
		}

		log.Infof("exported genesis-deposits for %v genesis-validators", len(genesisValidators.Data))
		return
	}
}

func isBeaconChainStarted() bool {
	var latestEpoch uint64
	err := db.WriterDb.Get(&latestEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
	if err != nil {
		log.Error(err, "error retrieving latest epoch from the database", 0)
		time.Sleep(time.Second * 10)
		return false
	}
	return latestEpoch > 0
}

func areGenesisDepositsExported() bool {
	var genesisDepositsCount uint64
	err := db.WriterDb.Get(&genesisDepositsCount, "SELECT COUNT(*) FROM blocks_deposits WHERE block_slot=0")
	if err != nil {
		log.Error(err, "error retrieving genesis-deposits-count when exporting genesis-deposits", 0)
		time.Sleep(time.Minute)
		return false
	}
	return genesisDepositsCount > 0
}

func exportGenesisDeposits(genesisValidators *types.StandardValidatorsResponse) error {
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		log.Error(err, "error beginning db-tx when exporting genesis-deposits", 0)
		return err
	}

	log.Infof("exporting deposit data for %v genesis validators", len(genesisValidators.Data))
	for i, validator := range genesisValidators.Data {
		if i%1000 == 0 {
			log.Infof("exporting deposit data for genesis validator %v (%v/%v)", validator.Index, i, len(genesisValidators.Data))
		}
		_, err = tx.Exec(`INSERT INTO blocks_deposits (block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature)
		VALUES (0, '\x01', $1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
			validator.Index, validator.Validator.Pubkey, validator.Validator.WithdrawalCredentials, validator.Balance, []byte{0x0},
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error exporting genesis-deposits: %v", err)
		}
	}

	if err := hydrateEth1DepositSignatures(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := updateDepositsCount(tx, len(genesisValidators.Data)); err != nil {
		tx.Rollback()
		return err
	}

	return commitOrRollbackTx(tx)
}

func hydrateEth1DepositSignatures(tx *sqlx.Tx) error {
	_, err := tx.Exec(`
		UPDATE blocks_deposits
		SET signature = a.signature
		FROM (
			SELECT DISTINCT ON(publickey) publickey, signature
			FROM eth1_deposits
			WHERE valid_signature = true) AS a
		WHERE block_slot = 0 AND blocks_deposits.publickey = a.publickey AND blocks_deposits.signature = '\x'`)
	if err != nil {
		return fmt.Errorf("error hydrating eth1 data into genesis-deposits: %v", err)
	}
	return nil
}

func updateDepositsCount(tx *sqlx.Tx, count int) error {
	_, err := tx.Exec("UPDATE blocks SET depositscount = $1 WHERE slot = 0", count)
	if err != nil {
		return fmt.Errorf("error updating deposit count for the genesis slot: %v", err)
	}
	return nil
}

func commitOrRollbackTx(tx *sqlx.Tx) error {
	err := tx.Commit()
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			log.Error(rollbackErr, "error rolling back transaction", 0)
		}
		log.Error(err, "error committing db-tx when exporting genesis-deposits", 0)
		return err
	}
	return nil
}
