package middleware

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	dataaccess "github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthUserInjectorInterceptor(t *testing.T) {
	validKeyBase62 := "TDbUML7MRn7PbYBbfLRJebfFWOtYad2CcEt6UIBc4km"
	validKey, err := apikey.FromBase62(validKeyBase62)
	require.NoError(t, err)

	validUser := domain.User{ID: 1337}
	validStoredKey := apikey.APIKey{
		UserID: validUser.ID,
		Name:   "test-key",
	}

	tests := []struct {
		name            string
		metadata        metadata.MD
		setupMocks      func(userRepo *dataaccess.MockUserRepository, authRepo *dataaccess.MockAPIKeyRepository)
		expectedCode    codes.Code
		expectUserInCtx bool
	}{
		{
			name:     "successfully injects user",
			metadata: metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", validKeyBase62)),
			setupMocks: func(userRepo *dataaccess.MockUserRepository, authRepo *dataaccess.MockAPIKeyRepository) {
				authRepo.On("GetAPIKey", mock.Anything, validKey).Return(validStoredKey, nil)
				userRepo.On("GetUserById", mock.Anything, validStoredKey.UserID).Return(validUser, nil)
				authRepo.On("UpdateLastUsedAt", mock.Anything, validKey).Return(nil)
			},
			expectedCode:    codes.OK,
			expectUserInCtx: true,
		},
		{
			name:            "missing metadata",
			metadata:        nil,
			setupMocks:      func(_ *dataaccess.MockUserRepository, _ *dataaccess.MockAPIKeyRepository) {},
			expectedCode:    codes.Unauthenticated,
			expectUserInCtx: false,
		},
		{
			name:            "missing authorization header",
			metadata:        metadata.Pairs("something", "else"),
			setupMocks:      func(_ *dataaccess.MockUserRepository, _ *dataaccess.MockAPIKeyRepository) {},
			expectedCode:    codes.Unauthenticated,
			expectUserInCtx: false,
		},
		{
			name:            "malformed authorization header",
			metadata:        metadata.Pairs("authorization", "Token not-a-real-key"),
			setupMocks:      func(_ *dataaccess.MockUserRepository, _ *dataaccess.MockAPIKeyRepository) {},
			expectedCode:    codes.Unauthenticated,
			expectUserInCtx: false,
		},
		{
			name:            "invalid base62 key",
			metadata:        metadata.Pairs("authorization", "Bearer invalid$$"),
			setupMocks:      func(_ *dataaccess.MockUserRepository, _ *dataaccess.MockAPIKeyRepository) {},
			expectedCode:    codes.Unauthenticated,
			expectUserInCtx: false,
		},
		{
			name:     "API key not found",
			metadata: metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", validKeyBase62)),
			setupMocks: func(_ *dataaccess.MockUserRepository, authRepo *dataaccess.MockAPIKeyRepository) {
				authRepo.On("GetAPIKey", mock.Anything, validKey).Return(apikey.APIKey{}, errors.New("not found"))
			},
			expectedCode:    codes.Unauthenticated,
			expectUserInCtx: false,
		},
		{
			name:     "user not found",
			metadata: metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", validKeyBase62)),
			setupMocks: func(userRepo *dataaccess.MockUserRepository, authRepo *dataaccess.MockAPIKeyRepository) {
				authRepo.On("GetAPIKey", mock.Anything, validKey).Return(validStoredKey, nil)
				userRepo.On("GetUserById", mock.Anything, validStoredKey.UserID).Return(domain.User{}, domain.ErrNotFound)
			},
			expectedCode:    codes.Unauthenticated,
			expectUserInCtx: false,
		},
		{
			name:     "update last used fails",
			metadata: metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", validKeyBase62)),
			setupMocks: func(userRepo *dataaccess.MockUserRepository, authRepo *dataaccess.MockAPIKeyRepository) {
				authRepo.On("GetAPIKey", mock.Anything, validKey).Return(validStoredKey, nil)
				userRepo.On("GetUserById", mock.Anything, validStoredKey.UserID).Return(validUser, nil)
				authRepo.On("UpdateLastUsedAt", mock.Anything, validKey).Return(errors.New("db error"))
			},
			expectedCode:    codes.OK,
			expectUserInCtx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(dataaccess.MockUserRepository)
			mockAuthRepo := new(dataaccess.MockAPIKeyRepository)

			tt.setupMocks(mockUserRepo, mockAuthRepo)

			interceptor := AuthUserInjectorInterceptor(mockUserRepo, mockAuthRepo)

			ctx := context.Background()
			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadata)
			}

			var userFromCtx domain.User

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				if tt.expectUserInCtx {
					userFromCtx = assertUserInjectedIntoContext(t, ctx, validUser, validKey.String())

					// Ensure "authorization" is stripped
					md, ok := metadata.FromIncomingContext(ctx)
					require.True(t, ok)
					assert.Empty(t, md.Get("authorization"), "authorization header should be stripped")
				} else {
					assert.Nil(t, auth.MustUserFromContext(ctx))
					assert.Nil(t, auth.MustAPIKeyFromContext(ctx))
				}
				return "ok", nil
			}

			_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Test"}, handler)

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

			mockUserRepo.AssertExpectations(t)
			mockAuthRepo.AssertExpectations(t)
		})
	}
}

func assertUserInjectedIntoContext(t *testing.T, ctx context.Context, expected domain.User, expectedHashedKey string) domain.User {
	t.Helper()

	// Check user context value
	user, ok := auth.UserFromContext(ctx)
	require.True(t, ok, "user should be injected into context")
	require.NotNil(t, user, "user in context should not be nil")
	assert.Equal(t, expected.ID, user.ID)

	// Check hashed API key context value
	hashedKey, ok := auth.APIKeyFromContext(ctx)
	assert.True(t, ok, "hashed API key should be injected into context")
	assert.Equal(t, expectedHashedKey, hashedKey)

	return user
}
