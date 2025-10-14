package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/app/apputils"
	"github.com/gobitfly/beaconchain-backend/internal/app/external_api/middleware"
	httpmiddleware "github.com/gobitfly/beaconchain-backend/internal/app/middleware"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/ethereumnetworkrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

type ApiService struct {
	userRepository      userrepo.Repository
	limiter             *limits.Limiter
	ethereumNetworkRepo ethereumnetworkrepo.Repository
}

// InitDependencies
// Initialize the repositories with proper databases
func InitDependencies(
	userRepository userrepo.Repository,
	ethereumNetworkRepo ethereumnetworkrepo.Repository,
) (*ApiService, error) {
	return &ApiService{
		userRepository:      userRepository,
		limiter:             limits.NewLimiter(),
		ethereumNetworkRepo: ethereumNetworkRepo,
	}, nil
}

var _ model.StrictServerInterface = (*ApiService)(nil)

// Run
// Takes as input a ServiceExecution configuration, and launches a gRPC reverse-proxied HTTP service.
func Run(
	config config.ServiceConfig,
) {
	log.Info("Starting server...")

	dataSources := data_sources.ApiDataSources{}

	dbUserRepo := &userrepo.DBRepository{}
	dbAPIKeyRepo := &apikeyrepo.DBRepository{}
	userRepo := &userrepo.CachedRepository{}
	apiKeyRepo := &apikeyrepo.CachedRepository{}
	ethereumNetworkRepo := &ethereumnetworkrepo.DBRepository{}

	dataSources.InitApiConnections(&config) // initialize blocking as middlewares depend on it

	// init async
	go func() {
		dbUserRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
		dbAPIKeyRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
		userRepo.Initialize(dataSources.Redis, dbUserRepo)
		apiKeyRepo.Initialize(dataSources.Redis, dbAPIKeyRepo)
		ethereumNetworkRepo.Initialize(dataSources.RoChainDb)
	}()

	apiService, _ := InitDependencies(userRepo, ethereumNetworkRepo)

	mux := http.NewServeMux()

	hStrict := model.NewStrictHandlerWithOptions(apiService, nil, model.StrictHTTPServerOptions{
		ResponseErrorHandlerFunc: apputils.ErrorHandler,
		RequestErrorHandlerFunc:  apputils.ErrorHandler,
	})

	h := model.HandlerWithOptions(hStrict, model.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: apputils.ErrorHandler,
		Middlewares: []model.MiddlewareFunc{
			middleware.AuthUserInjectorMiddleware(userRepo, apiKeyRepo, apputils.ErrorHandler),
			httpmiddleware.RecoveryMiddleware(),
		},
	})

	// configure validation middleware
	spec, err := model.GetSwagger()
	if err != nil {
		log.Fatalf("Failed to load OpenAPI spec: %v", err)
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

// serveSwaggerStatics
// Abstract this later to make it easier to add additional ones.
// func serveSwaggerStatics(mux *http.ServeMux) {
// 	// mount a path to expose the generated OpenAPI specification on disk
// 	// http://localhost:8080/swagger-ui/#/BeaconchainApiService
// 	mux.HandleFunc("/swagger-ui/swagger.json", func(w http.ResponseWriter, r *http.Request) {
// 		http.ServeFile(w, r, "./api/gen/api_service/v1/external.swagger.json")
// 	})

// 	// mount the Swagger UI that uses the OpenAPI specification path above
// 	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./web/swagger-ui"))))
// }

// getEndpointRatelimit retrieves the rate limit for a specific endpoint and tier from the protobuf definition for the ExternalService.
// func getEndpointRatelimit(fullMethod string, tier domain.Tier) (*model.RateLimitSettings, error) {
// 	service := model.File_api_service_v1_external_proto.Services().ByName("ExternalService")
// 	methodName := strings.TrimPrefix(fullMethod, "/"+string(service.FullName())+"/")
// 	method := service.Methods().ByName(protoreflect.Name(methodName))
// 	if fullMethod == "/grpc.health.v1.Health/Check" || fullMethod == "/grpc.health.v1.Health/List" || fullMethod == "/grpc.health.v1.Health/Watch" {
// 		return nil, nil
// 	}
// 	if method == nil {
// 		return nil, fmt.Errorf("method not found: %s", methodName)
// 	}
// 	ratelimitOpts := proto.GetExtension(method.Options(), model.E_RateLimitsPerTier).(*model.RateLimitsPerTier)
// 	if ratelimitOpts == nil {
// 		return nil, fmt.Errorf("rate limit options not found for method: %s", methodName)
// 	}
// 	switch tier {
// 	case domain.TierFree:
// 		return ratelimitOpts.Free, nil
// 	case domain.TierHobbyist:
// 		return ratelimitOpts.Hobbyist, nil
// 	case domain.TierBusiness:
// 		return ratelimitOpts.Business, nil
// 	case domain.TierScale:
// 		return ratelimitOpts.Scale, nil
// 	}
// 	return nil, fmt.Errorf("unknown subscription tier: %s", tier)
// }
