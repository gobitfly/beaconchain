package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
)

func genesisDepositsExporter(client rpc.Client, dbs db.ConsensusDBI) {
	for {
		shouldSleep, err := processGenesisDeposits(client, dbs)
		if err == nil {
			return // successfully processed genesis deposits
		}

		log.Error(err, "error processing genesis deposits, retrying...", 0)
		if shouldSleep {
			time.Sleep(time.Minute)
		}
	}
}

func processGenesisDeposits(client rpc.ValidatorClient, dbs db.ConsensusDBI) (bool, error) {
	latestEpoch, err := getLatestEpoch(dbs)
	if err != nil {
		return true, fmt.Errorf("error getting latest epoch: %v", err)
	}

	if latestEpoch == 0 {
		return true, fmt.Errorf("beacon chain not started")
	}

	genesisDepositCount, err := getGenesisDepositCount(dbs)
	if err != nil {
		return true, fmt.Errorf("error getting genesis deposit count: %v", err)
	}

	if genesisDepositCount > 0 {
		return false, nil
	}

	genesisValidators, err := client.GetValidatorState(0)
	if err != nil {
		log.Error(err, "error retrieving genesis validator data for genesis-epoch when exporting genesis-deposits: %v", 0)
		return true, fmt.Errorf("error retrieving genesis validator data: %w", err)
	}

	err = exportGenesisDeposits(genesisValidators, dbs)
	if err != nil {
		log.Error(err, "error exporting genesis-deposits: %v", 0)
		return true, fmt.Errorf("error exporting genesis deposits: %w", err)
	}

	log.Infof("exported genesis-deposits for %v genesis-validators", len(genesisValidators.Data))

	return false, nil
}

func getLatestEpoch(db db.ConsensusDBI) (uint64, error) {
	latestEpoch, err := db.GetLatestEpoch()
	if err != nil {
		log.Error(err, "error retrieving latest epoch from the database", 0)
		time.Sleep(time.Second * 10)
		return 0, err
	}

	return latestEpoch, nil
}
func getGenesisDepositCount(db db.ConsensusDBI) (uint64, error) {
	genesisDepositsCount, err := db.GetDepositsCountForBlockSlot()
	if err != nil {
		log.Error(err, "error retrieving genesis-deposits-count when exporting genesis-deposits", 0)
		time.Sleep(time.Minute)
		return 0, err
	}
	return genesisDepositsCount, nil
}

func exportGenesisDeposits(genesisValidators *types.StandardValidatorsResponse, db db.ConsensusDBI) error {
	log.Infof("exporting deposit data for %v genesis validators", len(genesisValidators.Data))

	for i, validator := range genesisValidators.Data {
		if i%1000 == 0 {
			log.Infof("exporting deposit data for genesis validator %v (%v/%v)", validator.Index, i, len(genesisValidators.Data))
		}
		err := db.SaveBlockDeposits(validator.Index, validator.Validator.Pubkey, validator.Validator.WithdrawalCredentials, validator.Balance)
		if err != nil {
			return fmt.Errorf("error exporting genesis-deposits: %v", err)
		}
	}

	err := db.UpdateBlockDepositsSignature()
	if err != nil {
		return fmt.Errorf("error hydrating eth1 data into genesis-deposits: %v", err)
	}

	err = db.UpdateBlockDepositCount(len(genesisValidators.Data))
	if err != nil {
		return fmt.Errorf("error updating deposit count for the genesis slot: %v", err)
	}

	return nil
}
