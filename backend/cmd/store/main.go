package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func main() {
	args := os.Args[1:]
	project := args[0]
	instance := args[1]
	table := args[2]
	port, err := strconv.Atoi(args[3])
	if err != nil {
		panic(err)
	}

	bt, err := database.NewBigTable(project, instance, nil)
	if err != nil {
		panic(err)
	}
	remote := database.NewRemote(database.Wrap(bt, table))
	go func() {
		log.Infof("starting remote raw store on port %d", port)
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), remote.Routes()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	utils.WaitForCtrlC()
}
