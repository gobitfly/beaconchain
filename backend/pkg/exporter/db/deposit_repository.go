package db

import "github.com/jmoiron/sqlx"

type DepositRepository interface {
	UpdateQueueDeposits(tx *sqlx.Tx) error
	CacheBlockDepositLookup() error
}

type depositRepository struct{}

func NewDepositRepository() DepositRepository {
	return &depositRepository{}
}

func (r *depositRepository) UpdateQueueDeposits(tx *sqlx.Tx) error {
	return UpdateQueueDeposits(tx)
}

func (r *depositRepository) CacheBlockDepositLookup() error {
	return CacheBlockDepositLookup()
}
