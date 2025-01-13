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
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type DataAccessor interface {
	ValidatorDashboardRepository
	SearchRepository
	NetworkRepository
	ClientRepository
	UserRepository
	AppRepository
	NotificationsRepository
	AdminRepository
	BlockRepository
	ArchiverRepository
	ProtocolRepository
	RatelimitRepository
	HealthzRepository
	MachineRepository

	Close()

	GetLatestFinalizedEpoch(ctx context.Context) (uint64, error)
	GetLatestSlot(ctx context.Context) (uint64, error)
	GetLatestBlock(ctx context.Context) (uint64, error)
	GetLatestExchangeRates(ctx context.Context) ([]t.EthConversionRate, error)

	GetProductSummary(ctx context.Context) (*t.ProductSummary, error)
	GetFreeTierPerks(ctx context.Context) (*t.PremiumPerks, error)

	GetValidatorsFromSlices(ctx context.Context, indices []uint64, publicKeys []string) ([]t.VDBValidator, error)
}

type DataAccessService struct {
	dummy *DummyService

	readerDb                *sqlx.DB
	writerDb                *sqlx.DB
	alloyReader             *sqlx.DB
	alloyWriter             *sqlx.DB
	clickhouseReader        *sqlx.DB
	userReader              *sqlx.DB
	userWriter              *sqlx.DB
	bigtable                *db.Bigtable
	persistentRedisDbClient *redis.Client

	services *services.Services
	config   *types.Config
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
	db.UserReader = das.userWriter
	db.UserWriter = das.userReader
	db.AlloyReader = das.alloyReader
	db.AlloyWriter = das.alloyWriter
	db.ClickHouseReader = das.clickhouseReader
	db.BigtableClient = das.bigtable
	db.PersistentRedisDbClient = das.persistentRedisDbClient

	return das
}

func createDataAccessService(cfg *types.Config) *DataAccessService {
	dataAccessService := DataAccessService{
		dummy:  NewDummyService(),
		config: cfg,
	}

	// Initialize the database
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.writerDb, dataAccessService.readerDb = db.MustInitDB(&cfg.WriterDatabase, &cfg.ReaderDatabase, "pgx", "postgres")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.alloyWriter, dataAccessService.alloyReader = db.MustInitDB(&cfg.AlloyWriter, &cfg.AlloyReader, "pgx", "postgres")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// lets just reuse reader to be extra safe
		dataAccessService.clickhouseReader, _ = db.MustInitDB(&cfg.ClickHouse.ReaderDatabase, &cfg.ClickHouse.ReaderDatabase, "clickhouse", "clickhouse")
	}()

	// Initialize the user database
	wg.Add(1)
	go func() {
		defer wg.Done()
		dataAccessService.userWriter, dataAccessService.userReader = db.MustInitDB(&cfg.Frontend.WriterDatabase, &cfg.Frontend.ReaderDatabase, "pgx", "postgres")
	}()

	// Initialize the bigtable
	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(cfg.Bigtable.Project, cfg.Bigtable.Instance, fmt.Sprintf("%d", cfg.Chain.ClConfig.DepositChainID), cfg.RedisCacheEndpoint)
		if err != nil {
			log.Fatal(err, "error connecting to bigtable", 0)
		}
		dataAccessService.bigtable = bt
	}()

	// Initialize the tiered cache (redis)
	if cfg.TieredCacheProvider == "redis" || len(cfg.RedisCacheEndpoint) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MustInitTieredCache(cfg.RedisCacheEndpoint)
			log.Infof("tiered Cache initialized, latest finalized epoch: %v", cache.LatestFinalizedEpoch.Get())
		}()
	}

	// Initialize the persistent redis client
	wg.Add(1)
	go func() {
		defer wg.Done()
		rdc := redis.NewClient(&redis.Options{
			Addr:        cfg.RedisSessionStoreEndpoint,
			ReadTimeout: time.Second * 60,
		})

		if err := rdc.Ping(context.Background()).Err(); err != nil {
			log.Fatal(err, "error connecting to persistent redis store", 0)
		}
		dataAccessService.persistentRedisDbClient = rdc
	}()

	wg.Wait()

	if cfg.TieredCacheProvider != "redis" {
		log.Fatal(fmt.Errorf("no cache provider set, please set TierdCacheProvider (example redis)"), "", 0)
	}

	// Return the result
	return &dataAccessService
}

func (d *DataAccessService) StartDataAccessServices() {
	// Create the services
	d.services = services.NewServices(d.readerDb, d.writerDb, d.alloyReader, d.alloyWriter, d.clickhouseReader, d.bigtable, d.persistentRedisDbClient)

	// Initialize repositories
	d.registerNotificationInterfaceTypes()
	// Initialize the services

	if d.config.SkipDataAccessServiceInitWait {
		go d.services.InitServices()
	} else {
		d.services.InitServices()
	}
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
	if d.clickhouseReader != nil {
		d.clickhouseReader.Close()
	}
	if d.bigtable != nil {
		d.bigtable.Close()
	}
}

var ErrNotFound = errors.New("not found")
