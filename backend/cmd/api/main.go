package main

import (
	"net/http"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

// TODO load these from config
const (
	host     = "0.0.0.0"
	port     = "8081"
	dummyApi = false
)

func main() {
	// TODO load config

	var dai dataaccess.DataAccessInterface
	if dummyApi {
		dai = dataaccess.NewDummyService()
	} else {
		dai = dataaccess.NewDataAccessService()
	}

	router := api.NewApiRouter(dai)
	srv := &http.Server{
		Handler:      router,
		Addr:         host + ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Infof("Serving on %s:%s", host, port)
	log.Fatal(srv.ListenAndServe(), "Error while serving", 0)
}
