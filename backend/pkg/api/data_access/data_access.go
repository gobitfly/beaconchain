package dataaccess

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/api/services"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type DataAccessor interface {
	ValidatorDashboardRepository
	SearchRepository
	NetworkRepository
	UserRepository

	Close()

	GetLatestSlot() (uint64, error)
	GetLatestExchangeRates() ([]t.EthConversionRate, error)

	GetProductSummary() (*t.ProductSummary, error)

	GetValidatorsFromSlices(indices []uint64, publicKeys []string) ([]t.VDBValidator, error)
}

type DataAccessService struct {
	dummy *DummyService

	readerDb                *sqlx.DB
	writerDb                *sqlx.DB
	alloyReader             *sqlx.DB
	alloyWriter             *sqlx.DB
	userReader              *sqlx.DB
	userWriter              *sqlx.DB
	bigtable                *db.Bigtable
	persistentRedisDbClient *redis.Client

	services *services.Services
}

// ensure DataAccessService pointer implements DataAccessor
var _ DataAccessor = (*DataAccessService)(nil)

func NewDataAccessService(cfg *types.Config) *DataAccessService {
	// Create the data access service
	das := createDataAccessService(cfg)

	// TODO: We set the global db connections here to have access to the functions in the db package
	// which use them without having to rewrite every single one.
	// This should be removed and the db functions should become methods of a struct that contains the db pointers.
	db.ReaderDb = das.readerDb
	db.WriterDb = das.writerDb
	db.AlloyReader = das.alloyReader
	db.AlloyWriter = das.alloyWriter
	db.BigtableClient = das.bigtable
	db.PersistentRedisDbClient = das.persistentRedisDbClient

	// Create the services
	das.services = services.NewServices(das.readerDb, das.writerDb, das.alloyReader, das.alloyWriter, das.bigtable, das.persistentRedisDbClient)

	// Initialize the services
	das.services.InitServices()

	return das
}

func createDataAccessService(cfg *types.Config) *DataAccessService {
	dataAccessService := DataAccessService{
		dummy: NewDummyService()}

	// Initialize the database
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.writerDb, dataAccessService.readerDb = db.MustInitDB(
			&types.DatabaseConfig{
				Username:     cfg.WriterDatabase.Username,
				Password:     cfg.WriterDatabase.Password,
				Name:         cfg.WriterDatabase.Name,
				Host:         cfg.WriterDatabase.Host,
				Port:         cfg.WriterDatabase.Port,
				MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
				MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
			},
			&types.DatabaseConfig{
				Username:     cfg.ReaderDatabase.Username,
				Password:     cfg.ReaderDatabase.Password,
				Name:         cfg.ReaderDatabase.Name,
				Host:         cfg.ReaderDatabase.Host,
				Port:         cfg.ReaderDatabase.Port,
				MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
				MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
			},
		)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.alloyWriter, dataAccessService.alloyReader = db.MustInitDB(
			&types.DatabaseConfig{
				Username:     cfg.AlloyWriter.Username,
				Password:     cfg.AlloyWriter.Password,
				Name:         cfg.AlloyWriter.Name,
				Host:         cfg.AlloyWriter.Host,
				Port:         cfg.AlloyWriter.Port,
				MaxOpenConns: cfg.AlloyWriter.MaxOpenConns,
				MaxIdleConns: cfg.AlloyWriter.MaxIdleConns,
				SSL:          cfg.AlloyWriter.SSL,
			},
			&types.DatabaseConfig{
				Username:     cfg.AlloyReader.Username,
				Password:     cfg.AlloyReader.Password,
				Name:         cfg.AlloyReader.Name,
				Host:         cfg.AlloyReader.Host,
				Port:         cfg.AlloyReader.Port,
				MaxOpenConns: cfg.AlloyReader.MaxOpenConns,
				MaxIdleConns: cfg.AlloyReader.MaxIdleConns,
				SSL:          cfg.AlloyReader.SSL,
			},
		)
	}()

	// Initialize the user database
	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.userWriter, dataAccessService.userReader = db.MustInitDB(
			&types.DatabaseConfig{
				Username:     cfg.Frontend.WriterDatabase.Username,
				Password:     cfg.Frontend.WriterDatabase.Password,
				Name:         cfg.Frontend.WriterDatabase.Name,
				Host:         cfg.Frontend.WriterDatabase.Host,
				Port:         cfg.Frontend.WriterDatabase.Port,
				MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
				MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
			},
			&types.DatabaseConfig{
				Username:     cfg.Frontend.ReaderDatabase.Username,
				Password:     cfg.Frontend.ReaderDatabase.Password,
				Name:         cfg.Frontend.ReaderDatabase.Name,
				Host:         cfg.Frontend.ReaderDatabase.Host,
				Port:         cfg.Frontend.ReaderDatabase.Port,
				MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
				MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
			},
		)
	}()

	// Initialize the bigtable
	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
		if err != nil {
			log.Fatal(err, "error connecting to bigtable", 0)
		}
		dataAccessService.bigtable = bt
	}()

	// Initialize the tiered cache (redis)
	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
			log.Infof("tiered Cache initialized, latest finalized epoch: %v", cache.LatestFinalizedEpoch.Get())
		}()
	}

	// Initialize the persistent redis client
	wg.Add(1)
	go func() {
		defer wg.Done()
		rdc := redis.NewClient(&redis.Options{
			Addr:        utils.Config.RedisSessionStoreEndpoint,
			ReadTimeout: time.Second * 60,
		})

		if err := rdc.Ping(context.Background()).Err(); err != nil {
			log.Fatal(err, "error connecting to persistent redis store", 0)
		}
		dataAccessService.persistentRedisDbClient = rdc
	}()

	wg.Wait()

	if utils.Config.TieredCacheProvider != "redis" {
		log.Fatal(fmt.Errorf("no cache provider set, please set TierdCacheProvider (example redis)"), "", 0)
	}

	// Return the result
	return &dataAccessService
}

func (d *DataAccessService) Close() {
	if d.readerDb != nil {
		d.readerDb.Close()
	}
	if d.writerDb != nil {
		d.writerDb.Close()
	}
	if d.alloyReader != nil {
		d.alloyReader.Close()
	}
	if d.alloyWriter != nil {
		d.alloyWriter.Close()
	}
	if d.bigtable != nil {
		d.bigtable.Close()
	}
}

var ErrNotFound = errors.New("not found")
