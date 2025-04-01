package modules

import (
	"context"
	"fmt"
	"time"

	db2 "github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
)

type ValidatorClient interface {
	GetValidatorState(epoch uint64) (*types.StandardValidatorsResponse, error)
}

type genesisDepositsExporter struct {
	client ValidatorClient
	db     db2.ConsensusRepository

	delay time.Duration
	ctx   context.Context
}

func newGenesisDepositsExporter(ctx context.Context, client rpc.Client, db db2.ConsensusRepository) genesisDepositsExporter {
	return genesisDepositsExporter{
		client: client,
		db:     db,
		delay:  time.Minute,
		ctx:    ctx,
	}
}

func (e *genesisDepositsExporter) Export() {
	for {
		select {
		case <-e.ctx.Done():
			log.Info("genesis deposit export loop cancelled")
			return
		default:
			// check if the beaconchain has started
			latestEpoch, err := e.db.GetLatestEpoch()
			if err != nil {
				log.Error(err, "error retrieving latest epoch from the database", 0)
				time.Sleep(time.Second * 10)
				continue
			}

			if latestEpoch == 0 {
				time.Sleep(e.delay)
				continue
			}

			// check if genesis-deposits have already been exported
			genesisDepositCount, err := e.db.GetDepositsCountForBlockSlot()
			if err != nil {
				log.Error(err, "error retrieving genesis-deposits-count when exporting genesis-deposits", 0)
				time.Sleep(e.delay)
				continue
			}

			// if genesis-deposits have already been exported exit this go-routine
			if genesisDepositCount > 0 {
				return
			}

			genesisValidators, err := e.client.GetValidatorState(0)
			if err != nil {
				log.Error(err, "error retrieving genesis validator data for genesis-epoch when exporting genesis-deposits: %v", 0)
				time.Sleep(e.delay)
				continue
			}

			err = e.exportGenesisDeposits(genesisValidators)
			if err != nil {
				log.Error(err, "error exporting genesis-deposits: %v", 0)
				time.Sleep(e.delay)
				continue
			}

			log.Infof("exported genesis-deposits for %v genesis-validators", len(genesisValidators.Data))
			return
		}
	}
}

func (e *genesisDepositsExporter) exportGenesisDeposits(genesisValidators *types.StandardValidatorsResponse) error {
	log.Infof("exporting deposit data for %v genesis validators", len(genesisValidators.Data))

	for i, validator := range genesisValidators.Data {
		if i%1000 == 0 {
			log.Infof("exporting deposit data for genesis validator %v (%v/%v)", validator.Index, i, len(genesisValidators.Data))
		}
		err := e.db.SaveBlockDeposits(validator.Index, validator.Validator.Pubkey, validator.Validator.WithdrawalCredentials, validator.Balance)
		if err != nil {
			return fmt.Errorf("error exporting genesis-deposits: %v", err)
		}
	}

	err := e.db.UpdateBlockDepositsSignature()
	if err != nil {
		return fmt.Errorf("error hydrating eth1 data into genesis-deposits: %v", err)
	}

	err = e.db.UpdateBlockDepositCount(len(genesisValidators.Data))
	if err != nil {
		return fmt.Errorf("error updating deposit count for the genesis slot: %v", err)
	}

	return nil
}
