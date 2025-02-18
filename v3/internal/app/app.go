package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	model "github.com/gobitfly/beaconchain-api/api/gen"
	"github.com/gobitfly/beaconchain-api/internal/dataaccess"
	"github.com/gobitfly/beaconchain-api/internal/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
"context"
"fmt"
"net"
"net/http"

model "github.com/gobitfly/beaconchain-api/api/gen"
log "github.com/gobitfly/beaconchain-api/internal/log"
"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
"google.golang.org/grpc"
"google.golang.org/grpc/credentials/insecure"*/

var (
	port = 9090
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
 * Takes as input a ServiceExecution configuration, and launches a gRPC reverse-proxied HTTP service.
 *
 */
func (service *ApiService) Run() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Info("Starting server...")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Infof("failed to listen: %v", err)
	}

	s := grpc.NewServer() // Unsecured
	apiService, _ := InitWithInMemory()
	model.RegisterBeaconchainApiServiceServer(s, apiService)

	go s.Serve(lis)
	log.Infof("gRPC server listening at %v", lis.Addr())

	// Establish a connection to the gRPC server above
	conn, err := grpc.NewClient(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	// create a standard HTTP router
	mux := http.NewServeMux()

	// mount a path to expose the generated OpenAPI specification on disk
	mux.HandleFunc("/swagger-ui/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/gen/beaconchain_api.swagger.json")
	})

	// mount the Swagger UI that uses the OpenAPI specification path above
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./web/swagger-ui"))))

	// mount the gRPC HTTP gateway to the root
	mux.Handle("/", rmux)

	// start a standard HTTP server with the router
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Info(err)
	}
	log.Infof("HTTP server listening and serving at :8080")

	fmt.Println("To close connection CTRL+C :-)")
}

func (service *ApiService) ExecutionBlock(ctx context.Context, in *model.ExecutionBlockRequest) (*model.ExecutionBlockResponse, error) {
	data := model.BlockSummary{BlockNumber: "123", BlockHash: "fab"}
	return &model.ExecutionBlockResponse{Status: "ok", Data: &data}, nil
}
