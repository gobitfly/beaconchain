package data_sources

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/log"
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
	Redis    *redis.Client

	RoChainDb ChainRoConnection
	RwChainDb ChainRwConnection

	RoAdminDb AdminRoConnection
	RwAdminDb AdminRwConnection

	RoChDb ClickhouseRoConnection
	RwChDb ClickhouseRwConnection
}

func (dataSources *ApiDataSources) InitApiConnections(config *config.ServiceConfig) *ApiDataSources {
	// TODO hoodi, gnosis
	dataSources.RoChainDb = InitDB(&config.ReaderChainDatabaseMainnet, Postgres)
	dataSources.RwChainDb = InitDB(&config.WriterChainDatabaseMainnet, Postgres)

	dataSources.RoAdminDb = InitDB(&config.ReaderAdminDatabase, Postgres)
	dataSources.RwAdminDb = InitDB(&config.WriterAdminDatabase, Postgres)

	// TODO hoodi, gnosis
	dataSources.RoChDb = InitDB(&config.ReaderClickhouseMainnet, Clickhouse)
	dataSources.RwChDb = InitDB(&config.WriterClickhouseMainnet, Clickhouse)

	redis, err := InitRedisCache(context.Background(), &config.Redis)
	if err != nil {
		log.Infof("Failed to initialize Redis cache: %v\n", err)
		os.Exit(1)
	}
	dataSources.Redis = redis

	bigtable, err := InitBigtable(context.Background(), &config.Bigtable)
	if err != nil {
		log.Infof("Failed to initialize Bigtable: %v\n", err)
		os.Exit(1)
	}
	dataSources.Bigtable = bigtable

	return dataSources
}
