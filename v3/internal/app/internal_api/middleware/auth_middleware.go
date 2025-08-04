package middleware

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthUserInjectorMiddleware is a dummy placeholder for now but establishes the pattern
func AuthUserInjectorMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		ctx = auth.SetUserInContext(metadata.NewIncomingContext(ctx, md), &domain.User{
			ID: 1337,
		})
		return handler(ctx, req)
	}
}
