package db

import (
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

type ValidatorRepository interface {
	SaveValidators(epoch uint64, validators []*types.Validator, client rpc.Client, batchSize int, tx *sqlx.Tx) error
	SaveValidatorQueue(validators *types.ValidatorQueue, tx *sqlx.Tx) error
}

type validatorRepository struct{}

func NewValidatorRepository() ValidatorRepository {
	return &validatorRepository{}
}

func (r *validatorRepository) SaveValidators(epoch uint64, validators []*types.Validator, client rpc.Client, batchSize int, tx *sqlx.Tx) error {
	return SaveValidators(epoch, validators, client, batchSize, tx)
}

func (r *validatorRepository) SaveValidatorQueue(validators *types.ValidatorQueue, tx *sqlx.Tx) error {
	return SaveValidatorQueue(validators, tx)
}
