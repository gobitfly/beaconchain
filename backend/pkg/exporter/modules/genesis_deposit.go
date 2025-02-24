package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
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
	latestEpoch, err := db.GetLatestEpoch()
	if err != nil {
		log.Error(err, "error retrieving latest epoch from the database", 0)
		time.Sleep(time.Second * 10)
		return false
	}
	return latestEpoch > 0
}

func areGenesisDepositsExported() bool {
	genesisDepositsCount, err := db.GetDepositsCountForBlockSlot()
	if err != nil {
		log.Error(err, "error retrieving genesis-deposits-count when exporting genesis-deposits", 0)
		time.Sleep(time.Minute)
		return false
	}
	return genesisDepositsCount > 0
}

func exportGenesisDeposits(genesisValidators *types.StandardValidatorsResponse) error {
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
