package db

import (
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

type ConsensusDB struct {
	WriterDb *sqlx.DB
	ReaderDb *sqlx.DB
}

type ConsensusDBI interface {
	SaveNetworkLivenessData(head *types.ChainHead) error
	GetNetworkLivenessPreviousHeadEpoch() (uint64, error)
}
