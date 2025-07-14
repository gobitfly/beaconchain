package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	dataaccess "github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ApiService struct {
	model.UnimplementedInternalServiceServer
	userRepository      dataaccess.UserRepository
	dashboardRepository dataaccess.ValidatorDashboardRepository
}

// InitDependencies
// Initialize the repositories with proper databases
func InitDependencies(
	userRepository dataaccess.UserRepository,
	dashboardRepository dataaccess.ValidatorDashboardRepository) (*ApiService, error) {
	return &ApiService{
		userRepository:      userRepository,
		dashboardRepository: dashboardRepository,
	}, nil
}

// Run
// Takes as input a ServiceExecution configuration, and launches a gRPC reverse-proxied HTTP service.
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
	apiService, _ := InitDependencies(userRepo, dashboardRepo)
	model.RegisterInternalServiceServer(s, apiService)

	if config.ExposeSchema {
		reflection.Register(s)
	}
	go func() {
		err := s.Serve(lis)
		if err != nil {
			log.Infof("failed to serve: %v", err)
		}
	}()
	log.Infof("gRPC server listening at %v", lis.Addr())

	// Establish a connection to the gRPC server above
	conn, err := grpc.NewClient(fmt.Sprintf(":%s", config.GrpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Infof("fail to dial: %v", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Infof("failed to close connection: %v", err)
		}
	}()

	// create an HTTP router which sends proxies HTTP requests to the gRPC server.
	rmux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(HeaderMatcher))
	client := model.NewInternalServiceClient(conn)
	err = model.RegisterInternalServiceHandlerClient(ctx, rmux, client)
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

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: time.Second,
	}
	err = server.Serve(l)
	if err != nil {
		log.Info(err)
	}

	log.Infof("To close connection CTRL+C :-)")
}

func AuthInterceptor(userRepository dataaccess.UserRepository) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}
		log.Infof("GRPC Metadata: %v", md)

		vals := metadata.ValueFromIncomingContext(ctx, string(auth.ApiKeyHeader))
		if len(vals) == 0 {
			return ctx, status.Errorf(codes.Unauthenticated, "No %s header", string(auth.ApiKeyHeader))
		}
		var apikey = vals[0] // Just check the first, if there are for some reason multiple

		// token, err := auth.AuthFromMD(ctx, "bearer")
		// if err != nil {
		//	return nil, err
		// }
		//
		// tokenInfo, err := parseToken(token)
		// if err != nil {
		//	return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		// }

		// ctx = logging.InjectFields(ctx, logging.Fields{"auth.sub", userClaimFromToken(tokenInfo)})

		// Using the APIKey, look it up in the database to get the userId
		user, err := userRepository.GetUserByApiKey(ctx, vals[0])
		if err != nil {
			log.Info("Request called with no API Key")
		}
		// TODO: Remove this, we dont want to print keys
		log.Debugf("Request called with API Key %s", apikey)

		// Now use the userId to get the User struct, and put it in the context

		return handler(context.WithValue(ctx, auth.CtxUserKey, user), req)
	}
}

// HeaderMatcher
// GRPC expects headers in a format which includes . This function simply passes along all HTTP-specified headers to GRPC.
// Headers in GRPC will be available via `metadata.FromIncomingContext(ctx)`
func HeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case string(auth.ApiKeyHeader):
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// serveSwaggerStatics
// Abstract this later to make it easier to add additional ones.
func serveSwaggerStatics(mux *http.ServeMux) {
	// mount a path to expose the generated OpenAPI specification on disk
	// http://localhost:8080/swagger-ui/#/InternalService
	mux.HandleFunc("/swagger-ui/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/gen/api_service/v1/internal.swagger.json")
	})

	// mount the Swagger UI that uses the OpenAPI specification path above
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./web/swagger-ui"))))
}
