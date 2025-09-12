package app

import (
	"context"
	"fmt"
	"net"

	"buf.build/go/protovalidate"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc/health/grpc_health_v1"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/app/internal_api/middleware"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/sessionstorerepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ApiService struct {
	model.UnimplementedInternalServiceServer
	userRepository   userrepo.Repository
	apiKeyRepository apikeyrepo.Repository
	limiter          *limits.Limiter
}

// InitDependencies
// Initialize the repositories with proper databases
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

// Run
// Takes as input a ServiceExecution configuration, and launches a gRPC reverse-proxied HTTP service.
func Run(
	config config.ServiceConfig,
) {

	log.Info("Starting server...")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GrpcPort))
	if err != nil {
		log.Infof("failed to listen: %v", err)
	}

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("failed to create validator: %v", err)
	}

	var (
		userRepoI         userrepo.Repository
		apikeyRepoI       apikeyrepo.Repository
		sessionStoreRepoI sessionstorerepo.Repository
	)
	if config.IsCloudDeployment {
		// TODO remove & use actual db repositories
		userRepoI = &userrepo.MockRepository{}
		apikeyRepoI = &apikeyrepo.MockRepository{}
		sessionStoreRepoI = &sessionstorerepo.MockRepository{}
	} else {
		userDbRepo := &userrepo.DBRepository{}
		apikeyRepo := &apikeyrepo.CachedRepository{}
		sessionStoreRepo := &sessionstorerepo.DBRepository{}
		userRepoI = userDbRepo
		apikeyRepoI = apikeyRepo
		sessionStoreRepoI = sessionStoreRepo

		dbAPIKeyRepo := &apikeyrepo.DBRepository{}

		// init async
		go func() {
			dataSources := data_sources.ApiDataSources{}
			dataSources.InitApiConnections(&config)
			userDbRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
			dbAPIKeyRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
			apikeyRepo.Initialize(dataSources.Redis, dbAPIKeyRepo)
			sessionStoreRepo.Initialize(dataSources.Redis)
		}()
	}
	apiService, _ := InitDependencies(userRepoI, apikeyRepoI)

	var unaryInterceptors []grpc.UnaryServerInterceptor
	unaryInterceptors = append(unaryInterceptors, protovalidate_middleware.UnaryServerInterceptor(validator))
	unaryInterceptors = append(unaryInterceptors, middleware.StripErrorMessageMiddleware())
	unaryInterceptors = append(unaryInterceptors, middleware.RecoveryMiddleware())
	unaryInterceptors = append(unaryInterceptors, middleware.AuthUserInjectorInterceptor(sessionStoreRepoI))

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
	)

	// Unsecured on application level, authenticated via gcp access management
	if config.ExposeSchema {
		reflection.Register(grpcServer)
	}

	model.RegisterInternalServiceServer(grpcServer, apiService)
	grpc_health_v1.RegisterHealthServer(grpcServer, apiService)

	go func() {
		log.Infof("gRPC server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	log.Infof("To close connection CTRL+C :-)")
	select {} // block
}

func (s *ApiService) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	resp := grpc_health_v1.HealthCheckResponse_SERVING

	if s.userRepository.Ping() != nil {
		resp = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: resp,
	}, nil
}

func (s *ApiService) List(ctx context.Context, req *grpc_health_v1.HealthListRequest) (*grpc_health_v1.HealthListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "List not implemented")
}

func (s *ApiService) Watch(*grpc_health_v1.HealthCheckRequest, grpc.ServerStreamingServer[grpc_health_v1.HealthCheckResponse]) error {
	return status.Errorf(codes.Unimplemented, "Watch not implemented")
}
