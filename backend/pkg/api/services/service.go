package services

import (
	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
)

type Services struct {
	readerDb                *sqlx.DB
	writerDb                *sqlx.DB
	alloyReader             *sqlx.DB
	alloyWriter             *sqlx.DB
	clickhouseReader        *sqlx.DB
	bigtable                *db.Bigtable
	persistentRedisDbClient *redis.Client
}

func NewServices(readerDb, writerDb, alloyReader, alloyWriter, clickhouseReader *sqlx.DB, bigtable *db.Bigtable, persistentRedisDbClient *redis.Client) *Services {
	return &Services{
		readerDb:                readerDb,
		writerDb:                writerDb,
		alloyReader:             alloyReader,
		alloyWriter:             alloyWriter,
		clickhouseReader:        clickhouseReader,
		bigtable:                bigtable,
		persistentRedisDbClient: persistentRedisDbClient,
	}
}

func (s *Services) InitServices() {
	go s.startSlotVizDataService()
	go s.startIndexMappingService()
	go s.startEfficiencyDataService()

	log.Infof("initializing prices...")
	price.Init(utils.Config.Chain.ClConfig.DepositChainID, utils.Config.Eth1ErigonEndpoint, utils.Config.Frontend.ClCurrency, utils.Config.Frontend.ElCurrency)
	log.Infof("...prices initialized")
}
