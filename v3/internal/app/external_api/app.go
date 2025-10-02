package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/app/apputils"
	"github.com/gobitfly/beaconchain-backend/internal/app/external_api/middleware"
	httpmiddleware "github.com/gobitfly/beaconchain-backend/internal/app/middleware"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/ethereumnetworkrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/validatorrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"github.com/gobitfly/beaconchain-backend/internal/ratelimit"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

type ApiService struct {
	chainConfigs        config.ChainConfigs
	userRepository      userrepo.Repository
	limiter             *limits.Limiter
	ethereumNetworkRepo ethereumnetworkrepo.Repository
	validatorRepository validatorrepo.Repository
}

// InitDependencies
// Initialize the repositories with proper databases
func InitDependencies(
	chainConfigs config.ChainConfigs,
	userRepository userrepo.Repository,
	ethereumNetworkRepo ethereumnetworkrepo.Repository,
	validatorRepository validatorrepo.Repository,
) (*ApiService, error) {
	return &ApiService{
		chainConfigs:        chainConfigs,
		userRepository:      userRepository,
		limiter:             limits.NewLimiter(),
		ethereumNetworkRepo: ethereumNetworkRepo,
		validatorRepository: validatorRepository,
	}, nil
}

var _ model.StrictServerInterface = (*ApiService)(nil)

// Run
// Takes as input a ServiceExecution configuration, and launches a gRPC reverse-proxied HTTP service.
func Run(
	config config.ServiceConfig,
	chainConfigs config.ChainConfigs,
) {

	log.Infof("chainconfigs %+v", chainConfigs)
	log.Info("Starting server...")

	dataSources := data_sources.ApiDataSources{}

	dbUserRepo := &userrepo.DBRepository{}
	dbAPIKeyRepo := &apikeyrepo.DBRepository{}
	userRepo := &userrepo.CachedRepository{}
	apiKeyRepo := &apikeyrepo.CachedRepository{}
	ethereumNetworkRepo := &ethereumnetworkrepo.DBRepository{}
	validatorRepo := &validatorrepo.DBRepository{}

	dataSources.InitApiConnections(&config) // initialize blocking as middlewares depend on it

	// init async
	go func() {
		dbUserRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
		dbAPIKeyRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
		userRepo.Initialize(dataSources.Redis, dbUserRepo)
		apiKeyRepo.Initialize(dataSources.Redis, dbAPIKeyRepo)
		ethereumNetworkRepo.Initialize(dataSources.RoChainDb, dataSources.RoChDb)
		validatorRepo.Initialize(dataSources.RoChainDb, dataSources.RoChDb)
	}()

	apiService, _ := InitDependencies(chainConfigs, userRepo, ethereumNetworkRepo, validatorRepo)

	mux := http.NewServeMux()

	hStrict := model.NewStrictHandlerWithOptions(apiService, nil, model.StrictHTTPServerOptions{
		ResponseErrorHandlerFunc: apputils.ErrorHandler,
		RequestErrorHandlerFunc:  apputils.ErrorHandler,
	})

	// configure validation middleware
	spec, err := model.GetSwagger()
	if err != nil {
		log.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	rateLimits, err := buildEndpointRateLimitsMap(spec)
	if err != nil {
		log.Fatalf("failed to build endpoint rate limit map: %v", err)
	}

	h := model.HandlerWithOptions(hStrict, model.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: apputils.ErrorHandler,
		Middlewares: []model.MiddlewareFunc{
			ratelimit.Middleware(dataSources.Redis, getEndpointRateLimit(spec, rateLimits), apputils.ErrorHandler),
			middleware.AuthUserInjectorMiddleware(userRepo, apiKeyRepo, apputils.ErrorHandler),
			httpmiddleware.RecoveryMiddleware(),
		},
	})

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

type endpointRateLimitKey struct {
	operationID string
	tier        domain.Tier
}

// buildEndpointRateLimitsMap builds a map of endpoint rate limits from the OpenAPI spec.
func buildEndpointRateLimitsMap(spec *openapi3.T) (map[endpointRateLimitKey]domain.RateLimitSettings, error) {
	endpointRateLimits := make(map[endpointRateLimitKey]domain.RateLimitSettings)
	for _, pathItem := range spec.Paths.Map() {
		for _, operation := range pathItem.Operations() {
			// put all existing tier settings into the map
			operationID := operation.OperationID
			if operationID == "" {
				return nil, fmt.Errorf("operationID is empty for endpoint with rate limit settings")
			}
			extensions, ok := operation.Extensions["x-ratelimits"].(map[string]any)
			if !ok {
				continue // no rate limit defined
			}
			for tierName, tierInfo := range extensions {
				steadyRate, ok := tierInfo.(map[string]any)["steady_rate"].(float64)
				if !ok {
					return nil, fmt.Errorf("invalid steady_rate for tier %s in endpoint %s", tierName, operation.OperationID)
				}
				bucketCapacity, ok := tierInfo.(map[string]any)["bucket_capacity"].(float64)
				if !ok {
					return nil, fmt.Errorf("invalid bucket_capacity for tier %s in endpoint %s", tierName, operation.OperationID)
				}

				endpointRateLimits[endpointRateLimitKey{
					operationID: operationID,
					tier:        domain.Tier(strings.ToUpper(tierName)),
				}] = domain.RateLimitSettings{
					SteadyRate:     float32(steadyRate),
					BucketCapacity: int(bucketCapacity),
				}
			}
		}
	}
	return endpointRateLimits, nil
}

// getEndpointRateLimit returns a func that retrieves the rate limit for a specific endpoint and tier from the OpenAPI spec.
func getEndpointRateLimit(spec *openapi3.T, rateLimits map[endpointRateLimitKey]domain.RateLimitSettings) ratelimit.GetEndpointSettingsFunc {
	// router finds the operationID from the request
	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		log.Fatalf("failed to create router for rate limit middleware: %v", err)
	}
	return func(r *http.Request, tier domain.Tier) (string, *domain.RateLimitSettings) {
		route, _, _ := router.FindRoute(r)
		operationID := route.Operation.OperationID
		settings, ok := rateLimits[endpointRateLimitKey{
			operationID: operationID,
			tier:        tier,
		}]
		if !ok {
			return operationID, nil // no rate limit defined, middleware will apply global default
		}
		return operationID, &settings
	}
}
