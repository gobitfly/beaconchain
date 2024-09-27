package archiver

import (
	"flag"
	"os"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"

	"github.com/gobitfly/beaconchain/pkg/archiver"
)

func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := fs.Bool("version", false, "Show version and exit")
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

	log.InfoWithFields(log.Fields{"config": *configPath, "version": version.Version, "commit": version.GitCommit, "chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

	dataAccessor := dataaccess.NewDataAccessService(cfg)
	defer dataAccessor.Close()

	archiver, err := archiver.NewArchiver(dataAccessor)
	if err != nil {
		log.Fatal(err, "error initializing archiving service", 0)
	}
	go archiver.Start()
	utils.WaitForCtrlC()
}
