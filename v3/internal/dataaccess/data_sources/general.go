package data_sources

import (
	"context"
	"fmt"
	"os"

	"github.com/gobitfly/beaconchain-api/internal/common/config"
	"github.com/jmoiron/sqlx"
)

type ChainRoConnection *sqlx.DB
type ChainRwConnection *sqlx.DB

type AdminRoConnection *sqlx.DB
type AdminRwConnection *sqlx.DB

type ClickhouseRoConnection *sqlx.DB
type ClickhouseRwConnection *sqlx.DB

type ApiDataSources struct {
	Bigtable *Bigtable
	Redis    *RedisCache

	RoChainDb ChainRoConnection
	RwChainDb ChainRwConnection

	RoAdminDb AdminRoConnection
	RwAdminDb AdminRwConnection

	RoChDb ClickhouseRoConnection
	RwChDb ClickhouseRwConnection
}

func InitApiConnections(config *config.ServiceConfig) *ApiDataSources {
	dataSources := &ApiDataSources{
		RoChainDb: InitDB(&config.ReaderChainDatabase, Postgres),
		RwChainDb: InitDB(&config.WriterChainDatabase, Postgres),

		RoAdminDb: InitDB(&config.ReaderAdminDatabase, Postgres),
		RwAdminDb: InitDB(&config.WriterAdminDatabase, Postgres),

		RoChDb: InitDB(&config.ReaderClickhouse, Clickhouse),
		RwChDb: InitDB(&config.WriterClickhouse, Clickhouse),
	}

	redis, err := InitRedisCache(context.Background(), &config.Redis)
	if err != nil {
		fmt.Printf("Failed to initialize Redis cache: %v\n", err)
		os.Exit(1)
	}
	dataSources.Redis = redis

	bigtable, err := InitBigtable(context.Background(), &config.Bigtable)
	if err != nil {
		fmt.Printf("Failed to initialize Bigtable: %v\n", err)
		os.Exit(1)
	}
	dataSources.Bigtable = bigtable

	return dataSources
}
