package main

import (
	"flag"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/exporter/modules"
	"github.com/gobitfly/beaconchain/pkg/exporter/services"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DBDSN  string
	CLNode string
}

var conf Config

func main() {
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&conf.DBDSN, "db.dsn", "postgres://user:pass@host:port/dbnames", "data-source-name of db, if it starts with projects/ it will use gcp-secretmanager")
	flag.StringVar(&conf.CLNode, "cl.endpoint", "http://localhost:4000", "cl node endpoint")
	flag.Parse()

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

	// err := db.InitWithDSN(conf.DBDSN)

	context, err := modules.GetModuleContext()
	if err != nil {
		utils.LogFatal(err, "error getting module context", 0)
	}

	go services.StartHistoricPriceService()
	go modules.StartAll(context)

	// Keep the program alive until Ctrl+C is pressed
	utils.WaitForCtrlC()
}
