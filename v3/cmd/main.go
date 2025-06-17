package main

import (
	app "github.com/gobitfly/beaconchain-api/internal/app/external_api"
	"github.com/gobitfly/beaconchain-api/internal/common/config"
	"github.com/gobitfly/beaconchain-api/internal/dataaccess/db"
	dataaccess "github.com/gobitfly/beaconchain-api/internal/dataaccess/repo"
)

/**
 * Initializes and kicks off the service.
 */
func main() {
	//flag.String("environment", "development", "The environment (and thus the config set) for which the service should be run (development, staging, production)")

	//serviceConfig := config.LoadServiceConfig(config.Development)
	serviceConfig := config.LoadServiceConfig()

	// Initializes dependencies which are required for the service to operate
	roConnectionAdminDb := db.InitDB(&serviceConfig.ReaderAdminDatabase, db.Postgres)
	rwConnectionAdminDb := db.InitDB(&serviceConfig.WriterAdminDatabase, db.Postgres)
	//roConnectionClickhouse := db.InitDB(&serviceConfig.ReaderClickhouse, db.Clickhouse)
	//rwConnectionClickhouse := db.InitDB(&serviceConfig.WriterClickhouse, db.Clickhouse)

	userRepo := dataaccess.NewDBUserRepository(roConnectionAdminDb, rwConnectionAdminDb)
	valDashboardRepo := dataaccess.NewInMemoryValidatorDashboardRepository()

	// Pass in initialized dependencies to service, and start the service

	app.Run(*serviceConfig, userRepo, valDashboardRepo)
}
