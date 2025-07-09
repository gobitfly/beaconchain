package main

import (
	"context"
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

	// Initializes dependencies which are required for the service to operate
	roConnectionAdminDb := data_sources.InitDB(&serviceConfig.ReaderAdminDatabase, data_sources.Postgres)
	rwConnectionAdminDb := data_sources.InitDB(&serviceConfig.WriterAdminDatabase, data_sources.Postgres)

	roConnectionChainDb := data_sources.InitDB(&serviceConfig.ReaderChainDatabase, data_sources.Postgres)
	rwConnectionChainDb := data_sources.InitDB(&serviceConfig.WriterChainDatabase, data_sources.Postgres)

	redisConnection, err := data_sources.InitRedisCache(context.Background(), &serviceConfig.Redis)
	if err != nil {
		log.Infof("Failed to initialize Redis cache: %v\n", err)
		os.Exit(1)
	}
	roConnectionClickhouse := data_sources.InitDB(&serviceConfig.ReaderClickhouse, data_sources.Clickhouse)
	rwConnectionClickhouse := data_sources.InitDB(&serviceConfig.WriterClickhouse, data_sources.Clickhouse)

	bigtableConnection, err := data_sources.InitBigtable(context.Background(), &serviceConfig.Bigtable)
	if err != nil {
		log.Infof("Failed to initialize Bigtable: %v\n", err)
		os.Exit(1)
	}

	userRepo := dataaccess.NewDBUserRepository(roConnectionAdminDb, rwConnectionAdminDb)
	valDashboardRepo := dataaccess.NewDBValidatorDashboardRepository(roConnectionChainDb, rwConnectionChainDb, roConnectionClickhouse, rwConnectionClickhouse, redisConnection, bigtableConnection)

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
