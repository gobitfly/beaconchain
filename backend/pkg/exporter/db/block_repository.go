package db

import (
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

type BlockRepository interface {
	SaveBlock(block *types.Block, isHeadEpoch bool, tx *sqlx.Tx) error
}

type blockRepository struct{}

func NewBlockRepository() BlockRepository {
	return &blockRepository{}
}

func (r *blockRepository) SaveBlock(block *types.Block, isHeadEpoch bool, tx *sqlx.Tx) error {
	return SaveBlock(block, isHeadEpoch, tx)
}
