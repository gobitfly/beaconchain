package main

import (
	"net/http"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api"
	"github.com/sirupsen/logrus"
)

const (
	HOST = "0.0.0.0"
	PORT = "8081"
)

func main() {
	// TODO load config

	// TODO init db/cache

	router := api.GetApiRouter()
	srv := &http.Server{
		Handler:      router,
		Addr:         HOST + ":" + PORT,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	logrus.Infof("Serving on %s:%s", HOST, PORT)
	logrus.Fatal(srv.ListenAndServe())
}
