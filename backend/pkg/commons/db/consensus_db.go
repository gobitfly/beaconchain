package db

import "github.com/jmoiron/sqlx"

type ConsensusDB struct {
	WriterDb *sqlx.DB
	ReaderDb *sqlx.DB
}

type ConsensusDBI interface {
	SaveValidatorTags(valueStrings []string, valueArgs [][]byte) error
	DeleteValidatorTags() error
	DeleteInvalidTags() error
}
