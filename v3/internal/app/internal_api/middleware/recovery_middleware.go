package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/gobitfly/beaconchain-backend/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryMiddleware catches panics in unary RPCs and converts them into gRPC errors.
func RecoveryMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Errorf("panic recovered: %v\n%s", r, debug.Stack()))

				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
