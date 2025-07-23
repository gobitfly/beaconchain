package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StripErrorMessageMiddleware replaces gRPC error messages with generic ones
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

		st, ok := status.FromError(err)
		if !ok {
			return nil, status.Error(codes.Internal, "internal server error. please try again later.")
		}

		return nil, status.Error(st.Code(), genericMessageForCode(st.Code()))
	}
}

func genericMessageForCode(code codes.Code) string {
	switch code {
	case codes.NotFound:
		return "resource not found"
	case codes.PermissionDenied:
		return "permission denied"
	case codes.AlreadyExists:
		return "resource already exists"
	case codes.InvalidArgument:
		return "invalid request parameters"
	default:
		return "internal server error. please try again later."
	}
}
