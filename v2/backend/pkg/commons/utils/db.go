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

func NullInt16(i int16) sql.NullInt16 {
	return sql.NullInt16{Int16: i, Valid: true}
}

func NullInt32(i int32) sql.NullInt32 {
	return sql.NullInt32{Int32: i, Valid: true}
}

func NullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}
