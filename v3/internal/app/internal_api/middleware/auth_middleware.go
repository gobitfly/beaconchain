package middleware

import (
	"context"
	"strings"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/sessionstorerepo"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthUserInjectorInterceptor verifies that a valid user session exists when accessing an authenticated
// endpoint
func AuthUserInjectorInterceptor(sessionStoreRepo sessionstorerepo.Repository) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Bypass health check methods
		if io.IsHealthCheckEndpoint(info.FullMethod) {
			return handler(ctx, req)
		}

		mustAuthenticate := isEndpointAuthenticationRequired(info.FullMethod)
		if !mustAuthenticate {
			return handler(ctx, req)
		}

		md, _ := metadata.FromIncomingContext(ctx)
		sessionID := md.Get("session_id")
		if len(sessionID) == 0 {
			return nil, common.NewInternalUserFacingError(codes.Unauthenticated, "missing metadata")
		}

		user, err := sessionStoreRepo.GetUserFromSessionID(ctx, sessionID[0])
		if err != nil {
			return nil, common.NewInternalUserFacingError(codes.Unauthenticated, "invalid session ID")
		}

		newMD := md.Copy()
		newMD.Delete("session_id") // strip authorization header from metadata

		ctx = metadata.NewIncomingContext(ctx, newMD)
		ctx = auth.SetUserInContext(ctx, user)
		return handler(ctx, req)
	}
}

// getEndpointAuthenticationRequirements retrieves the authentication requirement for a specific method from the protobuf definition for the InternalService.
// returns true if an endpoint requires auth.
// Defaults to true if no auth requirement has been configured for the specific service
func isEndpointAuthenticationRequired(fullMethod string) bool {
	service := model.File_api_service_v1_internal_proto.Services().ByName("InternalService")
	methodName := strings.TrimPrefix(fullMethod, "/"+string(service.FullName())+"/")
	method := service.Methods().ByName(protoreflect.Name(methodName))

	if method == nil {
		return true
	}
	requiresAuth, ok := proto.GetExtension(method.Options(), model.E_RequireAuthentication).(bool)
	if !ok {
		return true
	}

	return requiresAuth
}
