package db

import "github.com/jmoiron/sqlx"

type ConsensusDB struct {
	WriterDb *sqlx.DB
	ReaderDb *sqlx.DB
}

type ConsensusDBI interface {
	GetLatestEpoch() (uint64, error)
	GetDepositsCountForBlockSlot() (uint64, error)
	SaveBlockDeposits(validatorIndex uint64, pubkey []byte, withdrawalCredentials []byte, balance uint64) error
	UpdateBlockDepositsSignature() error
	UpdateBlockDepositCount(count int) error
}
