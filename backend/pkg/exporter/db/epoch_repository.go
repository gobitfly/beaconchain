package db

import (
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

type EpochRepository interface {
	SaveEpoch(epoch uint64, validators []*types.Validator, client rpc.Client, tx *sqlx.Tx) error
	UpdateEpochStatus(epochParticipationStats *types.ValidatorParticipation, tx *sqlx.Tx) error
}

type epochRepository struct{}

func NewEpochRepository() EpochRepository {
	return &epochRepository{}
}

func (r *epochRepository) SaveEpoch(epoch uint64, validators []*types.Validator, client rpc.Client, tx *sqlx.Tx) error {
	return SaveEpoch(epoch, validators, client, tx)
}

func (r *epochRepository) UpdateEpochStatus(epochParticipationStats *types.ValidatorParticipation, tx *sqlx.Tx) error {
	return UpdateEpochStatus(epochParticipationStats, tx)
}
