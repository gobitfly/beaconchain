package db

import "github.com/jmoiron/sqlx"

type ConsensusDB struct {
	WriterDb *sqlx.DB
	ReaderDb *sqlx.DB
}

type ConsensusDBI interface {
	GetSyncCommitteesCountPerValidator() (uint64, error)
	GetTotalPeriodSyncCommitteesCountPerValidator() (uint64, error)
	GetCountSoFarSyncCommitteesCountPerValidator(period uint64) (float64, error)
	SaveSyncCommitteesCount(period uint64, count float64) error
	GetEpochValidatorsCount(epoch uint64) (uint64, error)
	GetLatestFinalizedEpoch() (uint64, error)
}
