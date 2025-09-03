package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	dataaccess "github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/metadata"
)

// AuthUserInjectorInterceptor returns a gRPC interceptor that authenticates via API key
// and injects the user into the context.
func AuthUserInjectorInterceptor(userRepo dataaccess.UserAuthRepository, apiKeyRepo dataaccess.APIKeyAuthRepository) grpc.UnaryServerInterceptor {
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

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, common.NewExternalError(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, common.NewExternalError(codes.Unauthenticated, "missing authorization header")
		}

		authHeader := authHeaders[0]
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return nil, common.NewExternalError(codes.Unauthenticated, "invalid authorization format")
		}

		base62Key := strings.TrimPrefix(authHeader, "Bearer ")
		apiKey, err := apikey.FromBase62(base62Key)
		if err != nil {
			return nil, common.NewExternalError(codes.Unauthenticated, "invalid authorization format")
		}

		key, err := apiKeyRepo.GetAPIKey(ctx, apiKey)
		if err != nil {
			return nil, common.NewExternalError(codes.Unauthenticated, "invalid API key")
		}

		user, err := userRepo.GetUserById(ctx, key.UserID)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				log.Error(fmt.Errorf("failed to get user by API key: %v", err))
			}
			return nil, common.NewExternalError(codes.Unauthenticated, "invalid API key")
		}

		err = apiKeyRepo.UpdateLastUsedAt(ctx, apiKey)
		if err != nil {
			log.Warn(fmt.Errorf("failed to update last used time for API key: %v", err))
		}

		// Security: Strip the authorization header from metadata before continuing
		newMD := md.Copy()
		newMD.Delete("authorization")

		ctx = metadata.NewIncomingContext(ctx, newMD)
		ctx = auth.SetUserInContext(ctx, user)
		ctx = auth.SetAPIKeyInContext(ctx, apiKey.String())

		return handler(ctx, req)
	}
}
