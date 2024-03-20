package utils

import (
	"database/sql"
	"errors"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
)

var duplicateEntryError = &pgconn.PgError{Code: "23505"}

func Rollback(tx *sqlx.Tx) {
	err := tx.Rollback()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.Error(err, "error rolling back transaction", 0)
	}
}

func IsDuplicatedKeyError(err error) bool {
	return errors.Is(err, duplicateEntryError)
}
