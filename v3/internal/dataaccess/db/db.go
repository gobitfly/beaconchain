package db

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "database/sql/driver"

	"github.com/gobitfly/beaconchain-api/internal/common/config"
	"github.com/gobitfly/beaconchain-api/internal/log"
	"github.com/jmoiron/sqlx"

	// This brings in the driver
	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	_ "github.com/jackc/pgx/v5/stdlib"
)

/*
var DBPGX *pgxpool.Conn
*/
type DatabaseType string

const (
	Postgres   DatabaseType = "postgres"   // Set up and managed by the developer
	Clickhouse DatabaseType = "clickhouse" // Deployed and managed for internal use only
)

/**
 * Initializes a database connection, panicing if a connection was not able to be established.
 */
func InitDB(dbConfig *config.DatabaseConfig, databaseType DatabaseType) *sqlx.DB {
	extraParams := []string{}

	sslParam := databaseType.getSSLParam(dbConfig.SSL)
	extraParams = append(extraParams, sslParam)

	// TODO: Also include the `connection_open_strategy` stuff for handling multiple failover hosts.
	/*
		var hosts string
		hosts = net.JoinHostPort(writer.Host, writer.Port)
		if len(writer.Failovers) > 0 {
			for _, failover := range writer.Failovers {
				hosts += "," + net.JoinHostPort(failover.Host, failover.Port)
			}
			extraParams += "&connection_open_strategy=in_order"
		}*/
	db := sqlx.MustConnect(databaseType.getDriverName(), createDbConnectionString(databaseType, *dbConfig, extraParams))

	if dbConfig.MaxOpenConns == 0 {
		dbConfig.MaxOpenConns = 50
		log.Infof("MaxOpenConns for database %s found to be 0. Setting it to default of %d", dbConfig.DbName, dbConfig.MaxOpenConns)
	}
	if dbConfig.MaxIdleConns == 0 {
		dbConfig.MaxIdleConns = 10
		log.Infof("MaxIdleConns for database %s found to be 0. Setting it to default of %d", dbConfig.DbName, dbConfig.MaxIdleConns)
	}
	if dbConfig.MaxOpenConns < dbConfig.MaxIdleConns {
		log.Infof("MaxOpenConns (%d) for database %s found to be less than MaxIdleConns (%d). Setting MaxIdleConns to MaxOpenConns", dbConfig.MaxOpenConns, dbConfig.DbName, dbConfig.MaxIdleConns)
		dbConfig.MaxIdleConns = dbConfig.MaxOpenConns
	}

	db.SetConnMaxIdleTime(time.Second * 30)
	db.SetConnMaxLifetime(time.Minute)
	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)

	return db
}

func createDbConnectionString(databaseType DatabaseType, dbConfig config.DatabaseConfig, extraParams []string) string {
	return fmt.Sprintf("%s://%s:%s@%s/%s?%s", string(databaseType), dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.DbName, strings.Join(extraParams, "&"))
}

func (databaseType DatabaseType) getSSLParam(shouldUseSSL bool) string {
	switch databaseType {
	case Postgres:
		// Defensively assume SSL is enabled by default
		sslValue := "require"
		if !shouldUseSSL {
			sslValue = "disable"
		}
		return fmt.Sprintf("sslmode=%s", sslValue)
	case Clickhouse:
		return fmt.Sprintf("secure=%s", strconv.FormatBool(shouldUseSSL))
	default:
		log.Fatalf("Unknown databaseType: %s", string(databaseType))
	}

	return ""
}

func (dbType DatabaseType) getDriverName() string {
	switch dbType {
	case Postgres:
		return "pgx"
	case Clickhouse:
		return "clickhouse"
	}
	return ""
}

var ErrNotFound = errors.New("not found")
