package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	model "github.com/gobitfly/beaconchain-api/api/gen"
	"github.com/gobitfly/beaconchain-api/internal/common/config"
	dataaccess "github.com/gobitfly/beaconchain-api/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-api/internal/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ApiService struct {
	model.UnimplementedBeaconchainApiServiceServer
	userRepository      dataaccess.UserRepository
	dashboardRepository dataaccess.ValidatorDashboardRepository
}

/**
 * Initializes the state and dependencies of the service
 */
func InitWithInMemory() (*ApiService, error) {
	return &ApiService{
		userRepository:      dataaccess.NewInMemoryUserRepository(),
		dashboardRepository: dataaccess.NewInMemoryValidatorDashboardRepository(),
	}, nil
}

/**
 * Initialize the repositories with proper databases
 */
func Init(
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
	apiService, _ := Init(config, userRepo, dashboardRepo)
	model.RegisterBeaconchainApiServiceServer(s, apiService)

	go s.Serve(lis)
	log.Infof("gRPC server listening at %v", lis.Addr())

	// Establish a connection to the gRPC server above
	conn, err := grpc.NewClient(fmt.Sprintf(":%s", config.GrpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Infof("fail to dial: %v", err)
	}
	defer conn.Close()

	// create an HTTP router which sends proxies HTTP requests to the gRPC server.
	rmux := runtime.NewServeMux()
	client := model.NewBeaconchainApiServiceClient(conn)
	err = model.RegisterBeaconchainApiServiceHandlerClient(ctx, rmux, client)
	if err != nil {
		log.Info(err)
	}

	clientv1 := model.NewBeaconchainApiV1ServiceClient(conn)
	err = model.RegisterBeaconchainApiV1ServiceHandlerClient(ctx, rmux, clientv1)
	if err != nil {
		log.Info(err)
	}

	// create a standard HTTP router
	mux := http.NewServeMux()

	// mount a path to expose the generated OpenAPI specification on disk
	// http://localhost:8080/swagger-ui/#/BeaconchainApiService
	mux.HandleFunc("/swagger-ui/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/gen/beaconchain_api.swagger.json")
	})

	// mount a path to expose the generated OpenAPI specification on disk
	// http://localhost:8080/swagger-ui/v1/#/BeaconchainApiV1Service
	mux.HandleFunc("/swagger-ui/v1/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/gen/beaconchain_api_v1.swagger.json")
	})

	// mount the Swagger UI that uses the OpenAPI specification path above
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./web/swagger-ui"))))
	mux.Handle("/swagger-ui/v1/", http.StripPrefix("/swagger-ui/v1/", http.FileServer(http.Dir("./web/swagger-ui"))))

	// mount the gRPC HTTP gateway to the root
	mux.Handle("/", rmux)

	// start a standard HTTP server with the router
	err = http.ListenAndServe(fmt.Sprintf(":%s", config.HttpPort), mux)
	if err != nil {
		log.Info(err)
	}
	log.Infof("HTTP server listening and serving at :%s", config.HttpPort)

	fmt.Println("To close connection CTRL+C :-)")
}
