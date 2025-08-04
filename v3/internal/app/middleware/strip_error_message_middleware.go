package middleware

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/common"
	"google.golang.org/grpc"
)

func StripErrorMessageMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		return nil, common.SanitizeErrorMessage(err)
	}
}
