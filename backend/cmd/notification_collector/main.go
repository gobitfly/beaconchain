package notification_collector

import (
	"flag"
	"fmt"
	"os"
	"time"

	"net/http"
	"sync"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/notification"

	//nolint:gosec
	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	configPath := fs.String("config", "config.yml", "path to config")
	versionFlag := fs.Bool("version", false, "print version and exit")
	_ = fs.Parse(os.Args[2:])

	if *versionFlag {
		log.Info(version.Version)
		log.Info(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg
	log.InfoWithFields(log.Fields{
		"config":    *configPath,
		"version":   version.Version,
		"chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

	if utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 || utils.Config.Chain.ClConfig.SecondsPerSlot == 0 {
		log.Fatal(err, "invalid chain configuration specified, you must specify the slots per epoch, seconds per slot and genesis timestamp in the config file", 0)
	}

	if utils.Config.Metrics.Enabled {
		go func() {
			log.Infof("serving metrics on %v", utils.Config.Metrics.Address)
			if err := metrics.Serve(utils.Config.Metrics.Address, utils.Config.Metrics.Pprof, utils.Config.Metrics.PprofExtra); err != nil {
				log.Fatal(err, "error serving metrics", 0)
			}
		}()
	}

	if utils.Config.Pprof.Enabled {
		go func() {
			log.Infof("starting pprof http server on port %s", utils.Config.Pprof.Port)
			server := &http.Server{
				Addr:         fmt.Sprintf("localhost:%s", utils.Config.Pprof.Port),
				Handler:      nil,
				ReadTimeout:  60 * time.Second,
				WriteTimeout: 60 * time.Second,
			}
			err := server.ListenAndServe()

			if err != nil {
				log.Error(err, "error during ListenAndServe for pprof http server", 0)
			}
		}()
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.WriterDatabase.Username,
			Password:     cfg.WriterDatabase.Password,
			Name:         cfg.WriterDatabase.Name,
			Host:         cfg.WriterDatabase.Host,
			Port:         cfg.WriterDatabase.Port,
			MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
			SSL:          cfg.WriterDatabase.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.ReaderDatabase.Username,
			Password:     cfg.ReaderDatabase.Password,
			Name:         cfg.ReaderDatabase.Name,
			Host:         cfg.ReaderDatabase.Host,
			Port:         cfg.ReaderDatabase.Port,
			MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
			SSL:          cfg.ReaderDatabase.SSL,
		}, "pgx", "postgres")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.AlloyWriter, db.AlloyReader = db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.AlloyWriter.Username,
			Password:     cfg.AlloyWriter.Password,
			Name:         cfg.AlloyWriter.Name,
			Host:         cfg.AlloyWriter.Host,
			Port:         cfg.AlloyWriter.Port,
			MaxOpenConns: cfg.AlloyWriter.MaxOpenConns,
			MaxIdleConns: cfg.AlloyWriter.MaxIdleConns,
			SSL:          cfg.AlloyWriter.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.AlloyReader.Username,
			Password:     cfg.AlloyReader.Password,
			Name:         cfg.AlloyReader.Name,
			Host:         cfg.AlloyReader.Host,
			Port:         cfg.AlloyReader.Port,
			MaxOpenConns: cfg.AlloyReader.MaxOpenConns,
			MaxIdleConns: cfg.AlloyReader.MaxIdleConns,
			SSL:          cfg.AlloyReader.SSL,
		}, "pgx", "postgres")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.FrontendWriterDB, db.FrontendReaderDB = db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.Frontend.WriterDatabase.Username,
			Password:     cfg.Frontend.WriterDatabase.Password,
			Name:         cfg.Frontend.WriterDatabase.Name,
			Host:         cfg.Frontend.WriterDatabase.Host,
			Port:         cfg.Frontend.WriterDatabase.Port,
			MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
		}, &types.DatabaseConfig{
			Username:     cfg.Frontend.ReaderDatabase.Username,
			Password:     cfg.Frontend.ReaderDatabase.Password,
			Name:         cfg.Frontend.ReaderDatabase.Name,
			Host:         cfg.Frontend.ReaderDatabase.Host,
			Port:         cfg.Frontend.ReaderDatabase.Port,
			MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
		}, "pgx", "postgres")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// clickhouse
		db.ClickHouseWriter, db.ClickHouseReader = db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.ClickHouse.WriterDatabase.Username,
			Password:     cfg.ClickHouse.WriterDatabase.Password,
			Name:         cfg.ClickHouse.WriterDatabase.Name,
			Host:         cfg.ClickHouse.WriterDatabase.Host,
			Port:         cfg.ClickHouse.WriterDatabase.Port,
			MaxOpenConns: cfg.ClickHouse.WriterDatabase.MaxOpenConns,
			SSL:          true,
			MaxIdleConns: cfg.ClickHouse.WriterDatabase.MaxIdleConns,
		}, &types.DatabaseConfig{
			Username:     cfg.ClickHouse.ReaderDatabase.Username,
			Password:     cfg.ClickHouse.ReaderDatabase.Password,
			Name:         cfg.ClickHouse.ReaderDatabase.Name,
			Host:         cfg.ClickHouse.ReaderDatabase.Host,
			Port:         cfg.ClickHouse.ReaderDatabase.Port,
			MaxOpenConns: cfg.ClickHouse.ReaderDatabase.MaxOpenConns,
			SSL:          true,
			MaxIdleConns: cfg.ClickHouse.ReaderDatabase.MaxIdleConns,
		}, "clickhouse", "clickhouse")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
		if err != nil {
			log.Fatal(err, "error connecting to bigtable", 0)
		}
		db.BigtableClient = bt
	}()

	if utils.Config.TieredCacheProvider != "redis" {
		log.Fatal(nil, "no cache provider set, please set TierdCacheProvider (redis)", 0)
	}
	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
			log.Infof("tiered Cache initialized, latest finalized epoch: %v", cache.LatestFinalizedEpoch.Get())
		}()
	}

	log.Infof("initializing prices...")
	price.Init(utils.Config.Chain.ClConfig.DepositChainID, utils.Config.Eth1ErigonEndpoint, utils.Config.Frontend.ClCurrency, utils.Config.Frontend.ElCurrency)
	log.Infof("...prices initialized")

	wg.Wait()

	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()
	defer db.AlloyReader.Close()
	defer db.AlloyWriter.Close()
	defer db.ClickHouseReader.Close()
	defer db.ClickHouseWriter.Close()
	defer db.BigtableClient.Close()

	log.Infof("database connection established")

	notification.InitNotificationCollector(utils.Config.Notifications.PubkeyCachePath)

	utils.WaitForCtrlC()

	log.Infof("exiting...")
}
