package main

import (
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
)

// TODO load these from config
const (
	host     = "0.0.0.0"
	port     = "8080"
	dummyApi = false
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

	var dai dataaccess.DataAccessInterface
	if dummyApi {
		dai = dataaccess.NewDummyService()
	} else {
		dai = dataaccess.NewDataAccessService(cfg)
	}
	defer dai.CloseDataAccessService()

	router := api.NewApiRouter(dai)
	srv := &http.Server{
		Handler:      router,
		Addr:         net.JoinHostPort(cfg.Frontend.Server.Host, cfg.Frontend.Server.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Infof("Serving on %s:%s", host, port)
	log.Fatal(srv.ListenAndServe(), "Error while serving", 0)
}
