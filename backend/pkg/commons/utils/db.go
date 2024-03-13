package utils

import (
	"database/sql"
	"errors"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/jmoiron/sqlx"
)

func Rollback(tx *sqlx.Tx) {
	err := tx.Rollback()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.Error(err, "error rolling back transaction", 0)
	}
}
