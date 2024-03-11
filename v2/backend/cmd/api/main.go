package main

import (
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/rs/cors"

	"github.com/gobitfly/beaconchain/pkg/api/services"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
)

// TODO load these from config
const (
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

	var dai dataaccess.DataAccessor
	if dummyApi {
		dai = dataaccess.NewDummyService()
	} else {
		dai = InitServices(dataaccess.NewDataAccessService(cfg))
	}
	defer dai.CloseDataAccessService()

	router := api.NewApiRouter(dai)
	handler := cors.AllowAll().Handler(router)

	srv := &http.Server{
		Handler:      handler,
		Addr:         net.JoinHostPort(cfg.Frontend.Server.Host, cfg.Frontend.Server.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Infof("Serving on %s:%s", cfg.Frontend.Server.Host, cfg.Frontend.Server.Port)
	log.Fatal(srv.ListenAndServe(), "Error while serving", 0)
}

func InitServices(das dataaccess.DataAccessService) dataaccess.DataAccessor {
	db.ReaderDb = das.ReaderDb
	db.WriterDb = das.WriterDb
	db.PersistentRedisDbClient = das.PersistentRedisDbClient

	go services.StartSlotVizDataService()

	return das
}
