package services

import (
	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/jmoiron/sqlx"
)

type Services struct {
	readerDb                *sqlx.DB
	writerDb                *sqlx.DB
	alloyReader             *sqlx.DB
	alloyWriter             *sqlx.DB
	bigtable                *db.Bigtable
	persistentRedisDbClient *redis.Client
}

func NewServices(readerDb, writerDb, alloyReader, alloyWriter *sqlx.DB, bigtable *db.Bigtable, persistentRedisDbClient *redis.Client) *Services {
	return &Services{
		readerDb:                readerDb,
		writerDb:                writerDb,
		alloyReader:             alloyReader,
		alloyWriter:             alloyWriter,
		bigtable:                bigtable,
		persistentRedisDbClient: persistentRedisDbClient,
	}
}

func (s *Services) InitServices() {
	// go s.startSlotVizDataService()
	// go s.startIndexMappingService()
}
