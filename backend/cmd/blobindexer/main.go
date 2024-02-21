package main

import (
	"flag"
	"fmt"

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
		fmt.Println(version.Version)
		return
	}
	utils.Config = &types.Config{}
	err := utils.ReadConfig(utils.Config, *configFlag)
	if err != nil {
		utils.LogFatal(err, "error reading config file", 0)
	}
	blobIndexer, err := blobindexer.NewBlobIndexer()
	if err != nil {
		utils.LogFatal(err, "error initializing blob indexer", 0)
	}
	go blobIndexer.Start()
	utils.WaitForCtrlC()
}
