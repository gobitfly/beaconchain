package db

import "github.com/jmoiron/sqlx"

type ConsensusDB struct {
	WriterDb *sqlx.DB
	ReaderDb *sqlx.DB
}

type ConsensusDBI interface {
	UpdatePubkeyTags() error
}
