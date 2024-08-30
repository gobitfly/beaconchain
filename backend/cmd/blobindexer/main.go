package blobindexer

import (
	"flag"
	"os"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"

	"github.com/gobitfly/beaconchain/pkg/blobindexer"
)

func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)

	configFlag := fs.String("config", "config.yml", "path to config")
	versionFlag := fs.Bool("version", false, "print version and exit")
	_ = fs.Parse(os.Args[2:])
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
