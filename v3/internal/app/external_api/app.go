package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	api "github.com/gobitfly/beaconchain-backend/api/external/model"
	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/app/external_api/middleware"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/ethereumnetworkrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/gobitfly/beaconchain-backend/internal/log"
)

type ApiService struct {
	model.UnimplementedExternalServiceServer
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

var _ api.StrictServerInterface = (*ApiService)(nil)

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

	hStrict := api.NewStrictHandlerWithOptions(apiService, nil, api.StrictHTTPServerOptions{
		ResponseErrorHandlerFunc: errorHandler,
		RequestErrorHandlerFunc:  errorHandler,
	})

	h := api.HandlerWithOptions(hStrict, api.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: errorHandler,
		Middlewares: []api.MiddlewareFunc{
			middleware.AuthUserInjectorMiddleware(userRepo, apiKeyRepo, errorHandler),
			middleware.RecoveryMiddleware(),
		},
	})

	healthHandler := initHealthHandler(&dataSources)
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

func initHealthHandler(dataSources *data_sources.ApiDataSources) *HealthHandler {
	timeout := 10 * time.Second

	return NewHealthHandler(map[string]Pingable{
		"Redis":     WithTimeout(func(ctx context.Context) error { return dataSources.Redis.Ping(ctx).Err() }, timeout),
		"RoAdminDb": WithTimeout(dataSources.RoAdminDb.DB.PingContext, timeout),
		"RwAdminDb": WithTimeout(dataSources.RwAdminDb.DB.PingContext, timeout),
		"RoChDb":    WithTimeout(dataSources.RoChDb.DB.PingContext, timeout),
		"RwChDb":    WithTimeout(dataSources.RwChDb.DB.PingContext, timeout),
	})
}

func WithTimeout(p Pingable, d time.Duration) Pingable {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		return p.PingContext(ctx)
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Default response
	status := http.StatusInternalServerError
	msg := common.GenericErrMsg

	// Handle our custom API errors (user facing and internal)
	var apiErr *common.APIError
	if errors.As(err, &apiErr) {
		status = apiErr.Status
		if apiErr.Visibility == common.ErrorVisibilityUser {
			msg = apiErr.Message
		}
		log.WarnWithFields(apiErr.Extras, fmt.Sprintf("API error: %s | Status: %d | Path: %s ", apiErr.Message, apiErr.Status, r.URL.Path))
	} else {
		// Bad request errors as thrown by generated server
		switch e := err.(type) {
		case *api.InvalidParamFormatError, *api.RequiredParamError:
			status = http.StatusBadRequest
			msg = e.Error()
		default:
			log.Error(fmt.Sprintf("Internal error: %v | Path: %s", err, r.URL.Path))
		}
	}

	w.WriteHeader(status)
	if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": msg}); encodeErr != nil {
		log.Error(fmt.Sprintf("failed to encode error response: %v", encodeErr))
	}
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
