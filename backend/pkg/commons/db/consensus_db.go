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
	SaveSyncCommitteeData(data []types.SyncCommittee) error
	GetSyncCommitteesPeriods() ([]uint64, error)
}
