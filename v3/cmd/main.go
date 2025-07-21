package main

import (
	"os"

	app_external "github.com/gobitfly/beaconchain-backend/internal/app/external_api"
	app_internal "github.com/gobitfly/beaconchain-backend/internal/app/internal_api"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/log"
)

/**
 * Initializes and kicks off the service.
 */
func main() {
	serviceConfig := config.LoadServiceConfig()
	switch serviceConfig.Type {
	case "external":
		app_external.Run(*serviceConfig)
	case "internal":
		app_internal.Run(*serviceConfig)
	default:
		log.Infof("Unknown API type: %s\n", serviceConfig.Type)
		os.Exit(1)
	}
}
