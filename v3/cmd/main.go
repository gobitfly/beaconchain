package main

import (
	"os"

	app_external "github.com/gobitfly/beaconchain-api/internal/app/external_api"
	app_internal "github.com/gobitfly/beaconchain-api/internal/app/internal_api"
	"github.com/gobitfly/beaconchain-api/internal/common/config"
	"github.com/gobitfly/beaconchain-api/internal/dataaccess/db"
	dataaccess "github.com/gobitfly/beaconchain-api/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-api/internal/log"
)

/**
 * Initializes and kicks off the service.
 */
func main() {
	serviceConfig := config.LoadServiceConfig()

	// Initializes dependencies which are required for the service to operate
	roConnectionAdminDb := db.InitDB(&serviceConfig.ReaderAdminDatabase, db.Postgres)
	rwConnectionAdminDb := db.InitDB(&serviceConfig.WriterAdminDatabase, db.Postgres)
	//roConnectionClickhouse := db.InitDB(&serviceConfig.ReaderClickhouse, db.Clickhouse)
	//rwConnectionClickhouse := db.InitDB(&serviceConfig.WriterClickhouse, db.Clickhouse)

	userRepo := dataaccess.NewDBUserRepository(roConnectionAdminDb, rwConnectionAdminDb)
	valDashboardRepo := dataaccess.NewInMemoryValidatorDashboardRepository()

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
