package main

import (
	"flag"

	"github.com/gobitfly/beaconchain/commons/utils"
	"github.com/gobitfly/beaconchain/exporter/clnode"
	"github.com/gobitfly/beaconchain/exporter/modules"
)

type Config struct {
	DBDSN  string
	CLNode string
}

var conf Config

func main() {
	flag.StringVar(&conf.DBDSN, "db.dsn", "postgres://user:pass@host:port/dbnames", "data-source-name of db, if it starts with projects/ it will use gcp-secretmanager")
	flag.StringVar(&conf.CLNode, "cl.endpoint", "http://localhost:4000", "cl node endpoint")
	flag.Parse()

	// err := db.InitWithDSN(conf.DBDSN)

	go startModules()

	// Keep the program alive until Ctrl+C is pressed
	select {}
}

func startModules() {
	cl, err := clnode.NewNodeDataRetriever(conf.CLNode, nil)
	if err != nil {
		utils.LogFatal(err, "error initializing clnode", 0)
		return
	}

	moduleContext := modules.ModuleContext{
		CL: cl,
		// TODO: EL, DB
	}

	registeredModules := []modules.ModuleInterfaceEpoch{
		modules.NewDashboardDataModule(moduleContext),
		// todo: add more modules here
	}

	for _, module := range registeredModules {
		module.Start(27889)
	}
}
