package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/exporter/modules"
	"github.com/gobitfly/beaconchain/pkg/exporter/services"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		log.Infof(version.Version)
		log.Infof(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg

	log.InfoWithFields(log.Fields{"config": *configPath, "version": version.Version, "commit": version.GitCommit, "chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

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
	if !cfg.JustV2 {
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
			bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
			if err != nil {
				log.Fatal(err, "error connecting to bigtable", 0)
			}
			db.BigtableClient = bt
		}()
		if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
				log.Infof("tiered Cache initialized, latest finalized epoch: %v", cache.LatestFinalizedEpoch.Get())
			}()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			// Initialize the persistent redis client
			rdc := redis.NewClient(&redis.Options{
				Addr:        utils.Config.RedisSessionStoreEndpoint,
				ReadTimeout: time.Second * 20,
			})

			if err := rdc.Ping(context.Background()).Err(); err != nil {
				log.Fatal(err, "error connecting to persistent redis store", 0)
			}
			db.PersistentRedisDbClient = rdc
		}()

		wg.Wait()
		if utils.Config.TieredCacheProvider != "redis" {
			log.Fatal(fmt.Errorf("no cache provider set, please set TierdCacheProvider (example redis)"), "", 0)
		}

	} else {
		log.Warnf("------- EXPORTER RUNNING IN V2 ONLY MODE ------")
	}

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
	//native
	db.ClickHouseNativeWriter = db.MustInitClickhouseNative(&types.DatabaseConfig{
		Username:     cfg.ClickHouse.WriterDatabase.Username,
		Password:     cfg.ClickHouse.WriterDatabase.Password,
		Name:         cfg.ClickHouse.WriterDatabase.Name,
		Host:         cfg.ClickHouse.WriterDatabase.Host,
		Port:         cfg.ClickHouse.WriterDatabase.Port,
		MaxOpenConns: cfg.ClickHouse.WriterDatabase.MaxOpenConns,
		SSL:          true,
		MaxIdleConns: cfg.ClickHouse.WriterDatabase.MaxIdleConns,
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		rpc.CurrentErigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
		if err != nil {
			log.Fatal(err, "error initializing erigon client", 0)
		}

		erigonChainId, err := rpc.CurrentErigonClient.GetNativeClient().ChainID(ctx)
		if err != nil {
			log.Fatal(err, "error retrieving erigon chain id", 0)
		}

		rpc.CurrentGethClient, err = rpc.NewGethClient(utils.Config.Eth1GethEndpoint)
		if err != nil {
			log.Fatal(err, "error initializing geth client", 0)
		}

		gethChainId, err := rpc.CurrentGethClient.GetNativeClient().ChainID(ctx)
		if err != nil {
			log.Fatal(err, "error retrieving geth chain id", 0)
		}

		if !(erigonChainId.String() == gethChainId.String() && erigonChainId.String() == fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID)) {
			log.Fatal(fmt.Errorf("chain id mismatch: erigon chain id %v, geth chain id %v, requested chain id %v", erigonChainId.String(), gethChainId.String(), fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID)), "", 0)
		}
	}()

	if !cfg.JustV2 {
		defer db.ReaderDb.Close()
		defer db.WriterDb.Close()
		defer db.AlloyReader.Close()
		defer db.AlloyWriter.Close()
		defer db.BigtableClient.Close()
	}
	defer db.ClickHouseReader.Close()
	defer db.ClickHouseWriter.Close()
	defer db.ClickHouseNativeWriter.Close()

	context, err := modules.GetModuleContext()
	if err != nil {
		log.Fatal(err, "error getting module context", 0)
	}

	if !cfg.JustV2 {
		go services.StartHistoricPriceService()
	}

	go modules.StartAll(context)

	if utils.Config.Metrics.Enabled {
		go func(addr string) {
			log.Infof("serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				log.Error(err, "error serving metrics", 0)
			}
		}(utils.Config.Metrics.Address)
	}

	// Keep the program alive until Ctrl+C is pressed
	utils.WaitForCtrlC()
}
