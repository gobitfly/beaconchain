package main

import (
	"github.com/gobitfly/beaconchain-api/internal/app/management_api"
	"github.com/gobitfly/beaconchain-api/internal/common/config"
	"github.com/gobitfly/beaconchain-api/internal/dataaccess/db"
	"github.com/gobitfly/beaconchain-api/internal/dataaccess/repo"
)

/**
 * Initializes and kicks off the service.
 */
func main() {
	serviceConfig := config.LoadServiceConfig(config.Development)

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
