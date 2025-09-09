package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/sessionstorerepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestInternalAuthUserInjectorInterceptor(t *testing.T) {
	validUser := domain.User{ID: 4242}

	tests := []struct {
		name            string
		fullMethod      string
		metadata        metadata.MD
		setupMocks      func(sessionRepo *sessionstorerepo.MockRepository)
		expectedCode    codes.Code
		expectUserInCtx bool
	}{
		{
			name:       "successfully injects user and strips session_id",
			fullMethod: "/api_service.v1.InternalService/GetAPIKeys",
			metadata:   metadata.Pairs("session_id", "sess-abc"),
			setupMocks: func(sessionRepo *sessionstorerepo.MockRepository) {
				sessionRepo.On("GetUserFromSessionID", mock.Anything, "sess-abc").Return(validUser, nil)
			},
			expectedCode:    codes.OK,
			expectUserInCtx: true,
		},
		{
			name:         "missing metadata",
			fullMethod:   "/api_service.v1.InternalService/GetAPIKeys",
			metadata:     nil,
			setupMocks:   func(_ *sessionstorerepo.MockRepository) {},
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "missing session_id",
			fullMethod:   "/api_service.v1.InternalService/GetAPIKeys",
			metadata:     metadata.Pairs("other", "value"),
			setupMocks:   func(_ *sessionstorerepo.MockRepository) {},
			expectedCode: codes.Unauthenticated,
		},
		{
			name:       "invalid session id returns unauthenticated",
			fullMethod: "/api_service.v1.InternalService/GetAPIKeys",
			metadata:   metadata.Pairs("session_id", "sess-bad"),
			setupMocks: func(sessionRepo *sessionstorerepo.MockRepository) {
				sessionRepo.On("GetUserFromSessionID", mock.Anything, "sess-bad").Return(domain.User{}, errors.New("not found"))
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "health check bypasses auth",
			fullMethod:   "/grpc.health.v1.Health/Check",
			metadata:     nil,
			setupMocks:   func(_ *sessionstorerepo.MockRepository) {},
			expectedCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionRepo := new(sessionstorerepo.MockRepository)
			if tt.setupMocks != nil {
				tt := tt // avoid loop variable capture in closures
				tt.setupMocks(mockSessionRepo)
			}

			interceptor := AuthUserInjectorInterceptor(mockSessionRepo)

			ctx := context.Background()
			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadata)
			}

			var userFromCtx domain.User

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				if tt.expectUserInCtx {
					// verify user injected
					user, ok := auth.UserFromContext(ctx)
					require.True(t, ok)
					userFromCtx = user

					// ensure session_id is stripped
					md, ok := metadata.FromIncomingContext(ctx)
					require.True(t, ok)
					assert.Empty(t, md.Get("session_id"))
				} else {
					// should not have user in ctx for negative cases or health check
					u, ok := auth.UserFromContext(ctx)
					if tt.fullMethod == "/grpc.health.v1.Health/Check" {
						// health check: bypass, no user expected and repo must not be called
						mockSessionRepo.AssertNotCalled(t, "GetUserFromSessionID", mock.Anything, mock.Anything)
					}
					assert.False(t, ok)
					assert.Equal(t, domain.User{}, u)
				}
				return "ok", nil
			}

			_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: tt.fullMethod}, handler)

			if tt.expectedCode == codes.OK {
				require.NoError(t, err)
				if tt.expectUserInCtx {
					require.NotNil(t, userFromCtx)
					assert.Equal(t, validUser.ID, userFromCtx.ID)
				}
			} else {
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}

			mockSessionRepo.AssertExpectations(t)
		})
	}
}
