package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	model "github.com/gobitfly/beaconchain-api/api/gen"
	"github.com/gobitfly/beaconchain-api/internal/auth"
	"github.com/gobitfly/beaconchain-api/internal/common/config"
	dataaccess "github.com/gobitfly/beaconchain-api/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-api/internal/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type ApiService struct {
	model.UnimplementedExternalServiceServer
	userRepository      dataaccess.UserRepository
	dashboardRepository dataaccess.ValidatorDashboardRepository
}

/**
 * Initializes the state and dependencies of the service
 */
func InitWithInMemory() (*ApiService, error) {
	return &ApiService{
		//		userRepository:      dataaccess.NewInMemoryUserRepository(),
		dashboardRepository: dataaccess.NewInMemoryValidatorDashboardRepository(),
	}, nil
}

/**
 * Initialize the repositories with proper databases
 */
func InitDependencies(
	config config.ServiceConfig,
	userRepository dataaccess.UserRepository,
	dashboardRepository dataaccess.ValidatorDashboardRepository) (*ApiService, error) {
	return &ApiService{
		userRepository:      userRepository,
		dashboardRepository: dashboardRepository,
	}, nil
}

/**
 * Takes as input a ServiceExecution configuration, and launches a gRPC reverse-proxied HTTP service.
 *
 */
func Run(
	config config.ServiceConfig,
	userRepo dataaccess.UserRepository,
	dashboardRepo dataaccess.ValidatorDashboardRepository,
) {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Info("Starting server...")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GrpcPort))
	if err != nil {
		log.Infof("failed to listen: %v", err)
	}

	s := grpc.NewServer() // Unsecured
	apiService, _ := InitDependencies(config, userRepo, dashboardRepo)
	model.RegisterExternalServiceServer(s, apiService)

	if config.ExposeSchema {
		reflection.Register(s)
	}
	go s.Serve(lis)
	log.Infof("gRPC server listening at %v", lis.Addr())

	// Establish a connection to the gRPC server above
	conn, err := grpc.NewClient(fmt.Sprintf(":%s", config.GrpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Infof("fail to dial: %v", err)
	}
	defer conn.Close()

	// create an HTTP router which sends proxies HTTP requests to the gRPC server.
	// Register both the API and APIv1 Service. We can serve requests for both services from the same endpoint this way
	rmux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(HeaderMatcher))

	client := model.NewExternalServiceClient(conn)
	err = model.RegisterExternalServiceHandlerClient(ctx, rmux, client)
	if err != nil {
		log.Info(err)
	}

	// create a standard HTTP router
	mux := http.NewServeMux()

	serveSwaggerStatics(mux)

	// mount the gRPC HTTP gateway to the root
	mux.Handle("/", rmux)

	// start a standard HTTP server with the router
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", config.HttpPort))
	if err != nil {
		log.Infof("failed to listen: %v", err)
	}
	log.Infof("HTTP server listening and serving at :%s", config.HttpPort)

	err = http.Serve(l, mux)
	if err != nil {
		log.Info(err)
	}

	fmt.Println("To close connection CTRL+C :-)")
}

/**
 * GRPC expects headers in a different format. This function simply passes along all HTTP-specified headers to GRPC.
 * Headers in GRPC will be available via `metadata.FromIncomingContext(ctx)`
 */
func HeaderMatcher(key string) (string, bool) {
	switch key {
	case string(auth.ApiKeyHeader):
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

/**
 * Abstract this later to make it easier to add additional ones.
 *
 */
func serveSwaggerStatics(mux *http.ServeMux) {
	// mount a path to expose the generated OpenAPI specification on disk
	// http://localhost:8080/swagger-ui/#/BeaconchainApiService
	mux.HandleFunc("/swagger-ui/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/gen/internal.swagger.json")
	})

	// mount the Swagger UI that uses the OpenAPI specification path above
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./web/swagger-ui"))))
}
