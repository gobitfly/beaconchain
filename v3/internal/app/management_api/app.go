package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"

	model "github.com/gobitfly/beaconchain-api/api/gen"
	"github.com/gobitfly/beaconchain-api/internal/common/config"

	"github.com/gobitfly/beaconchain-api/internal/auth"
	dataaccess "github.com/gobitfly/beaconchain-api/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-api/internal/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type ApiService struct {
	model.UnimplementedBeaconchainApiServiceServer
	userRepository      dataaccess.UserRepository
	dashboardRepository dataaccess.ValidatorDashboardRepository
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

	s := grpc.NewServer(
		grpc.UnaryInterceptor(AuthInterceptor(userRepo)),
	) // Unsecured
	apiService, _ := InitDependencies(config, userRepo, dashboardRepo)
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
	rmux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(HeaderMatcher))
	client := model.NewBeaconchainApiServiceClient(conn)
	err = model.RegisterBeaconchainApiServiceHandlerClient(ctx, rmux, client)
	if err != nil {
		log.Info(err)
	}

	// create a standard HTTP router
	mux := http.NewServeMux()

	serveSwaggerStatics(mux)

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

func AuthInterceptor(userRepository dataaccess.UserRepository) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		log.Infof("GRPC Metadata: %v", md)

		vals := metadata.ValueFromIncomingContext(ctx, string(auth.ApiKeyHeader))
		if len(vals) == 0 {
			return ctx, status.Errorf(codes.Unauthenticated, "No %s header", string(auth.ApiKeyHeader))
		}
		var apikey = vals[0] // Just check the first, if there are for some reason multiple
		/*
			token, err := auth.AuthFromMD(ctx, "bearer")
			if err != nil {
				return nil, err
			}

			tokenInfo, err := parseToken(token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
			}*/

		//ctx = logging.InjectFields(ctx, logging.Fields{"auth.sub", userClaimFromToken(tokenInfo)})

		// Using the APIKey, look it up in the database to get the userId
		user, err := userRepository.GetUserByApiKey(ctx, vals[0])
		if err != nil {
			log.Info("Request called with no API Key")
		}
		// TODO: Remove this, we dont want to print keys
		log.Debugf("Request called with API Key %s", apikey)

		// Now use the userId to get the User struct, and put it in the context

		return context.WithValue(ctx, auth.CtxUserKey, user), nil
	}
}

/**
 * GRPC expects headers in a format which includes . This function simply passes along all HTTP-specified headers to GRPC.
 * Headers in GRPC will be available via `metadata.FromIncomingContext(ctx)`
 */
func HeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
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
		http.ServeFile(w, r, "./api/gen/beaconchain_api.swagger.json")
	})

	// mount the Swagger UI that uses the OpenAPI specification path above
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./web/swagger-ui"))))
}
