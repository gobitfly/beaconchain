package main

import (
	"os"

	app_external "github.com/gobitfly/beaconchain-api/internal/app/external_api"
	app_internal "github.com/gobitfly/beaconchain-api/internal/app/internal_api"
	"github.com/gobitfly/beaconchain-api/internal/common/config"
	"github.com/gobitfly/beaconchain-api/internal/dataaccess/data_sources"
	dataaccess "github.com/gobitfly/beaconchain-api/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-api/internal/log"
)

/**
 * Initializes and kicks off the service.
 */
func main() {
	serviceConfig := config.LoadServiceConfig()
	dataSources := data_sources.InitApiConnections(serviceConfig)

	userRepo := dataaccess.NewDBUserRepository(dataSources.RoAdminDb, dataSources.RwAdminDb)
	valDashboardRepo := dataaccess.NewDBValidatorDashboardRepository(dataSources.RoChainDb, dataSources.RwChainDb, dataSources.RoChDb, dataSources.RwChDb, dataSources.Redis, dataSources.Bigtable)

	// Pass in initialized dependencies to service, and start the service
	switch serviceConfig.Type {
	case "external":
		app_external.Run(*serviceConfig, userRepo, valDashboardRepo)
	case "internal":
		app_internal.Run(*serviceConfig, userRepo, valDashboardRepo)
	default:
		log.Infof("Unknown API type: %s\n", serviceConfig.Type)
		os.Exit(1)
	}
}
