package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/exporter/modules"
	"github.com/gobitfly/beaconchain/pkg/exporter/services"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
		return
	}

	logrus.WithField("config", *configPath).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	logrus.WithFields(logrus.Fields{
		"config":    *configPath,
		"version":   version.Version,
		"commit":    version.GitCommit,
		"chainName": utils.Config.Chain.ClConfig.ConfigName}).Printf("starting")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.WriterDatabase.Username,
			Password:     cfg.WriterDatabase.Password,
			Name:         cfg.WriterDatabase.Name,
			Host:         cfg.WriterDatabase.Host,
			Port:         cfg.WriterDatabase.Port,
			MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
		}, &types.DatabaseConfig{
			Username:     cfg.ReaderDatabase.Username,
			Password:     cfg.ReaderDatabase.Password,
			Name:         cfg.ReaderDatabase.Name,
			Host:         cfg.ReaderDatabase.Host,
			Port:         cfg.ReaderDatabase.Port,
			MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		rpc.CurrentErigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
		if err != nil {
			logrus.Fatalf("error initializing erigon client: %v", err)
		}

		erigonChainId, err := rpc.CurrentErigonClient.GetNativeClient().ChainID(ctx)
		if err != nil {
			logrus.Fatalf("error retrieving erigon chain id: %v", err)
		}

		rpc.CurrentGethClient, err = rpc.NewGethClient(utils.Config.Eth1GethEndpoint)
		if err != nil {
			logrus.Fatalf("error initializing geth client: %v", err)
		}

		gethChainId, err := rpc.CurrentGethClient.GetNativeClient().ChainID(ctx)
		if err != nil {
			logrus.Fatalf("error retrieving geth chain id: %v", err)
		}

		if !(erigonChainId.String() == gethChainId.String() && erigonChainId.String() == fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID)) {
			logrus.Fatalf("chain id mismatch: erigon chain id %v, geth chain id %v, requested chain id %v", erigonChainId.String(), gethChainId.String(), fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
		if err != nil {
			logrus.Fatalf("error connecting to bigtable: %v", err)
		}
		db.BigtableClient = bt
	}()

	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
			logrus.Infof("tiered Cache initialized, latest finalized epoch: %v", cache.LatestFinalizedEpoch.Get())
		}()
	}

	wg.Wait()

	if utils.Config.TieredCacheProvider != "redis" {
		logrus.Fatalf("no cache provider set, please set TierdCacheProvider (example redis)")
	}

	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()
	defer db.BigtableClient.Close()

	context, err := modules.GetModuleContext()
	if err != nil {
		utils.LogFatal(err, "error getting module context", 0)
	}

	go services.StartHistoricPriceService()
	go modules.StartAll(context)

	// Keep the program alive until Ctrl+C is pressed
	utils.WaitForCtrlC()
}
