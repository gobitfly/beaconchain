package main

import (
	"github.com/gobitfly/beaconchain-api/internal/app"
)

func main() {
	server := &app.ApiService{}
	server.Run()
}
