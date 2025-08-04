package app

import (
	"context"
	"fmt"
	"net"
	"strings"

	"buf.build/go/protovalidate"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc/health/grpc_health_v1"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/app/internal_api/middleware"
	globalmiddleware "github.com/gobitfly/beaconchain-backend/internal/app/middleware"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	dataaccess "github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ApiService struct {
	model.UnimplementedInternalServiceServer
	userRepository      dataaccess.UserRepository
	dashboardRepository dataaccess.ValidatorDashboardRepository
	authRepository      dataaccess.APIKeyRepository
}

// InitDependencies
// Initialize the repositories with proper databases
func InitDependencies(
	userRepository dataaccess.UserRepository,
	dashboardRepository dataaccess.ValidatorDashboardRepository,
	authRepository dataaccess.APIKeyRepository) (*ApiService, error) {
	return &ApiService{
		userRepository:      userRepository,
		dashboardRepository: dashboardRepository,
		authRepository:      authRepository,
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

	var userRepoI dataaccess.UserRepository
	var vdbRepoI dataaccess.ValidatorDashboardRepository
	var apikeyRepoI dataaccess.APIKeyRepository
	if config.IsCloudDeployment {
		// TODO remove & use actual db repositories
		userRepoI = &dataaccess.MockUserRepository{}
		vdbRepoI = &dataaccess.DummyValidatorDashboardRepository{}
		apikeyRepoI = &dataaccess.MockAPIKeyRepository{}
	} else {
		userDbRepo := &dataaccess.DBUserRepository{}
		vbdDbRepo := &dataaccess.DBValidatorDashboardRepository{}
		apikeyRepo := &dataaccess.CachedAPIKeyRepository{}
		userRepoI = userDbRepo
		vdbRepoI = vbdDbRepo
		apikeyRepoI = apikeyRepo

		dbAPIKeyRepo := &dataaccess.DBAPIKeyRepository{}

		// init async
		go func() {
			dataSources := data_sources.ApiDataSources{}
			dataSources.InitApiConnections(&config)
			userDbRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
			vbdDbRepo.Initialize(dataSources.RoChainDb, dataSources.RwChainDb, dataSources.RoChDb, dataSources.RwChDb, dataSources.Redis, dataSources.Bigtable)
			dbAPIKeyRepo.Initialize(dataSources.RoAdminDb, dataSources.RwAdminDb)
			apikeyRepo.Initialize(dataSources.Redis, dbAPIKeyRepo)
		}()
	}

	apiService, _ := InitDependencies(userRepoI, vdbRepoI, apikeyRepoI)

	var unaryInterceptors []grpc.UnaryServerInterceptor
	unaryInterceptors = append(unaryInterceptors, protovalidate_middleware.UnaryServerInterceptor(validator))
	unaryInterceptors = append(unaryInterceptors, globalmiddleware.StripErrorMessageMiddleware())
	unaryInterceptors = append(unaryInterceptors, globalmiddleware.RecoveryMiddleware())
	unaryInterceptors = append(unaryInterceptors, middleware.AuthUserInjectorMiddleware())

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

// HeaderMatcher
// GRPC expects headers in a format which includes . This function simply passes along all HTTP-specified headers to GRPC.
// Headers in GRPC will be available via `metadata.FromIncomingContext(ctx)`
func HeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case string(auth.APIKeyHeader):
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func (s *ApiService) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	resp := grpc_health_v1.HealthCheckResponse_SERVING

	if s.userRepository.Ping() != nil || s.dashboardRepository.Ping() != nil {
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
