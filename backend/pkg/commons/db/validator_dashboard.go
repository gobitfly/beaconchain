package db

import (
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

var VDBReaderDB *sqlx.DB
var VDBWriterDB *sqlx.DB

func MustInitVDBDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig) {
	VDBReaderDB, VDBWriterDB = mustInitDB(writer, reader)
}
