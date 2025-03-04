package db

import "github.com/jmoiron/sqlx"

type ConsensusDB struct {
	WriterDb *sqlx.DB
	ReaderDb *sqlx.DB
}

type ConsensusDBI interface {
	GetAllSlots(tx *sqlx.Tx) ([]uint64, error)
	GetLastSlot(tx *sqlx.Tx) (uint64, error)
	SetSlotFinalizationAndStatus(slot uint64, finalized bool, status string, tx *sqlx.Tx) error
}
