package services

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var ErrWaiting error = errors.New("waiting for service to be initialized")

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
	wg := &sync.WaitGroup{}
	log.Infof("initializing services...")
	wg.Add(4)
	go s.startSlotVizDataService(wg)
	go s.startIndexMappingService(wg)
	go s.startEfficiencyDataService(wg)
	go s.startEmailSenderService(wg)

	log.Infof("initializing prices...")
	price.Init(utils.Config.Chain.ClConfig.DepositChainID, utils.Config.Eth1ErigonEndpoint, utils.Config.Frontend.ClCurrency, utils.Config.Frontend.ElCurrency)
	log.Infof("...prices initialized")

	wg.Wait()
	log.Infof("...services initialized")
}
