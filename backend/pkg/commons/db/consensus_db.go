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
	SaveValidatorTags(data []types.SSVExporterData) error
	DeleteValidatorTags() error
	DeleteInvalidTags() error
}
