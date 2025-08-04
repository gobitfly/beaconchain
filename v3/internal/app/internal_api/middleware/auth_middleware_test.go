package middleware

import (
	"context"
	"testing"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Just test the contract for the dummy implementation for now
func TestAuthUserInjectorMiddleware_AttachesUserToContext(t *testing.T) {
	interceptor := AuthUserInjectorMiddleware()

	md := metadata.Pairs("dummy", "value")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var userFromCtx *domain.User

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		user, ok := auth.UserFromContext(ctx)
		require.True(t, ok, "user should be injected into context")
		require.Equal(t, uint64(1337), user.ID)

		userFromCtx = user
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Fake",
	}, handler)

	require.NoError(t, err)
	require.Equal(t, "ok", resp)
	require.NotNil(t, userFromCtx, "user should not be nil")
}
