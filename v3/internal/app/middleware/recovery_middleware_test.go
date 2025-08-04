package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecoveryInterceptor(t *testing.T) {
	interceptor := RecoveryMiddleware()

	t.Run("normal handler does not panic", func(t *testing.T) {
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		}

		resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}, handler)

		require.NoError(t, err)
		require.Equal(t, "ok", resp)
	})

	t.Run("panic handler is recovered", func(t *testing.T) {
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("something bad happened")
		}

		resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/Method",
		}, handler)

		// Expect no panic propagated
		require.Nil(t, resp)

		// Expect gRPC Internal error
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.Internal, st.Code())
	})
}
