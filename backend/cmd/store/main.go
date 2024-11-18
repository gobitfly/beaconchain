package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func main() {
	args := os.Args[1:]
	project := args[0]
	instance := args[1]
	table := args[2]

	bt, err := database.NewBigTable(project, instance, nil)
	if err != nil {
		panic(err)
	}
	remote := database.NewRemote(database.Wrap(bt, table))
	go func() {
		log.Info("starting remote raw store on port 8087")
		if err := http.ListenAndServe("0.0.0.0:8087", remote.Routes()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	utils.WaitForCtrlC()
}
