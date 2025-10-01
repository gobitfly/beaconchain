package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// InterceptorDeps is the tiny API the interceptor needs.
type InterceptorDeps interface {
	Inc(method, code string)
}

// MetricsUnaryInterceptor records one cumulative count per request
// with labels (method, service_name via writer, code).
func MetricsUnaryInterceptor(w InterceptorDeps) grpc.UnaryServerInterceptor {
	if w == nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (interface{}, error) {
			return h(ctx, req)
		}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (interface{}, error) {
		resp, err := h(ctx, req)
		w.Inc(info.FullMethod, status.Code(err).String())
		return resp, err
	}
}
