package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	model "github.com/gobitfly/beaconchain-backend/api/inhouse/model"
	"github.com/gobitfly/beaconchain-backend/internal/app/apputils"
	"github.com/gobitfly/beaconchain-backend/internal/app/internal_api/middleware"
	httpmiddleware "github.com/gobitfly/beaconchain-backend/internal/app/middleware"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/sessionstorerepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

type ApiService struct {
	userRepository   userrepo.Repository
	apiKeyRepository apikeyrepo.Repository
	limiter          *limits.Limiter
}

func InitDependencies(
	userRepository userrepo.Repository,
	apiKeyRepository apikeyrepo.Repository,
) (*ApiService, error) {
	return &ApiService{
		userRepository:   userRepository,
		apiKeyRepository: apiKeyRepository,
		limiter:          limits.NewLimiter(),
	}, nil
}

var _ model.StrictServerInterface = (*ApiService)(nil)

// Run
// Takes as input a ServiceExecution configuration, and launches a gRPC reverse-proxied HTTP service.
func Run(
	config config.ServiceConfig,
) {
	log.Info("Starting server...")

	// OpenTelemetry + Google Cloud exporter (handles batching/retries/flush).
	otelCounter, err := httpmiddleware.NewOtelCounter(
		context.Background(),
		config.Metrics.ProjectID,
		os.Getenv("K_SERVICE"),
		os.Getenv("K_REVISION"),
	)
	if err != nil {
		log.Warnf("failed to init otel metrics: %v", err)
	}
	defer func() {
		if otelCounter != nil {
			_ = otelCounter.Close(context.Background())
		}
	}()

	userDbRepo := &userrepo.DBRepository{}
	apikeyRepo := &apikeyrepo.CachedRepository{}
	sessionStoreRepo := &sessionstorerepo.DBRepository{}
	dbAPIKeyRepo := &apikeyrepo.DBRepository{}
	dataSources := data_sources.ApiDataSources{}
	dataSources.InitApiConnections(&config)

	// init async
	go func() {
		userDbRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
		dbAPIKeyRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
		apikeyRepo.Initialize(dataSources.Redis, dbAPIKeyRepo)
		sessionStoreRepo.Initialize(dataSources.Redis)
	}()
	apiService, _ := InitDependencies(userDbRepo, apikeyRepo)

	mux := http.NewServeMux()

	hStrict := model.NewStrictHandlerWithOptions(apiService, nil, model.StrictHTTPServerOptions{
		ResponseErrorHandlerFunc: apputils.ErrorHandler,
		RequestErrorHandlerFunc:  apputils.ErrorHandler,
	})

	h := model.HandlerWithOptions(hStrict, model.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: apputils.ErrorHandler,
		Middlewares: []model.MiddlewareFunc{
			middleware.AuthUserInjectorMiddleware(sessionStoreRepo, apputils.ErrorHandler),
			httpmiddleware.RecoveryMiddleware(),
			httpmiddleware.MetricsHTTPMiddleware(otelCounter),
		},
	})

	// configure validation middleware
	spec, err := model.GetSwagger()
	if err != nil {
		defer func() {
			log.Fatalf("Failed to load OpenAPI spec: %v", err)
		}()
		return
	}
	validationMiddleware := nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
		ErrorHandlerWithOpts: apputils.ValidationErrorHandler,
		Options: openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc, // we do our own auth in a separate middleware
		},
	})
	// validation MUST wrap the server handler, so it runs before the server reads the request query
	// since the middleware sets the default values directly on the request var
	h = validationMiddleware(h)

	healthHandler := apputils.InitHealthHandler(&dataSources)
	mux.HandleFunc("/healthz", healthHandler.ServeHealth)
	mux.HandleFunc("/readyz", healthHandler.ServeReady)
	// serveSwaggerStatics(mux)

	go func() {
		s := &http.Server{
			Handler:           h,
			Addr:              fmt.Sprintf(":%s", config.HttpPort),
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       120 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
		}

		log.Infof("HTTP server listening on %s", s.Addr)
		log.Fatal(s.ListenAndServe())
	}()

	log.Infof("To close connection CTRL+C :-)")
	select {} // block
}
