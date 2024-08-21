package main

import (
	"flag"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"

	"github.com/gobitfly/beaconchain/pkg/blobindexer"
)

func main() {
	configFlag := flag.String("config", "config.yml", "path to config")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		log.Info(version.Version)
		return
	}
	utils.Config = &types.Config{}
	err := utils.ReadConfig(utils.Config, *configFlag)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	if utils.Config.Metrics.Enabled {
		go func() {
			log.Infof("serving metrics on %v", utils.Config.Metrics.Address)
			if err := metrics.Serve(utils.Config.Metrics.Address, utils.Config.Metrics.Pprof); err != nil {
				log.Fatal(err, "error serving metrics", 0)
			}
		}()
	}
	blobIndexer, err := blobindexer.NewBlobIndexer()
	if err != nil {
		log.Fatal(err, "error initializing blob indexer", 0)
	}
	go blobIndexer.Start()
	utils.WaitForCtrlC()
}
