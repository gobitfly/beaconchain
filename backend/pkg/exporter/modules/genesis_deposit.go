package modules

import (
	"context"
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
)

type genesisDepositsExporter struct {
	client rpc.ValidatorClient
	db     db.ConsensusDBI

	offset time.Duration
	ctx    context.Context
}

func newGenesisDepositsExporter(client rpc.Client, db db.ConsensusDBI) genesisDepositsExporter {
	return genesisDepositsExporter{
		client: client,
		db:     db,
		offset: time.Minute,
		ctx:    context.Background(),
	}
}

func (e genesisDepositsExporter) Export() {
	for {
		select {
		case <-e.ctx.Done():
			log.Info("export loop cancelled", 0)
			return
		default:
			latestEpoch, err := e.getLatestEpoch(e.db)
			if err != nil {
				continue
			}

			if latestEpoch == 0 {
				time.Sleep(e.offset)
				continue
			}

			genesisDepositCount, err := e.getGenesisDepositCount(e.db)
			if err != nil {
				continue
			}

			if genesisDepositCount > 0 {
				return
			}

			genesisValidators, err := e.client.GetValidatorState(0)
			if err != nil {
				log.Error(err, "error retrieving genesis validator data for genesis-epoch when exporting genesis-deposits: %v", 0)
				time.Sleep(e.offset)
				continue
			}

			err = e.exportGenesisDeposits(genesisValidators, e.db)
			if err != nil {
				log.Error(err, "error exporting genesis-deposits: %v", 0)
				time.Sleep(e.offset)
				continue
			}

			log.Infof("exported genesis-deposits for %v genesis-validators", len(genesisValidators.Data))
			return
		}
	}
}

func (e genesisDepositsExporter) getLatestEpoch(db db.ConsensusDBI) (uint64, error) {
	latestEpoch, err := db.GetLatestEpoch()
	if err != nil {
		log.Error(err, "error retrieving latest epoch from the database", 0)
		time.Sleep(time.Second * 10)
		return 0, err
	}

	return latestEpoch, nil
}
func (e genesisDepositsExporter) getGenesisDepositCount(db db.ConsensusDBI) (uint64, error) {
	genesisDepositsCount, err := db.GetDepositsCountForBlockSlot()
	if err != nil {
		log.Error(err, "error retrieving genesis-deposits-count when exporting genesis-deposits", 0)
		time.Sleep(e.offset)
		return 0, err
	}
	return genesisDepositsCount, nil
}

func (e genesisDepositsExporter) exportGenesisDeposits(genesisValidators *types.StandardValidatorsResponse, db db.ConsensusDBI) error {
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
