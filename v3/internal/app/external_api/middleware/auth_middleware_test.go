package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthUserInjectorMiddleware(t *testing.T) {
	validKeyBase62 := "TDbUML7MRn7PbYBbfLRJebfFWOtYad2CcEt6UIBc4km"
	validKey, err := apikey.FromBase62(validKeyBase62)
	require.NoError(t, err)

	validUser := domain.User{ID: 1337}
	validStoredKey := apikey.APIKey{
		UserID: validUser.ID,
		Name:   "test-key",
	}

	tests := []struct {
		name           string
		authHeader     string
		setupMocks     func(userRepo *userrepo.MockRepository, apiKeyRepo *apikeyrepo.MockRepository)
		expectedStatus int
		expectUser     bool
	}{
		{
			name:       "successfully injects user",
			authHeader: fmt.Sprintf("Bearer %s", validKeyBase62),
			setupMocks: func(userRepo *userrepo.MockRepository, apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("Get", mock.Anything, validKey).Return(validStoredKey, nil)
				userRepo.On("Get", mock.Anything, validStoredKey.UserID).Return(validUser, nil)
				apiKeyRepo.On("UpdateLastUsedAt", mock.Anything, validKey).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			setupMocks:     func(userRepo *userrepo.MockRepository, apiKeyRepo *apikeyrepo.MockRepository) {},
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
		{
			name:           "invalid format",
			authHeader:     "Token not-a-real-key",
			setupMocks:     func(userRepo *userrepo.MockRepository, apiKeyRepo *apikeyrepo.MockRepository) {},
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
		{
			name:           "invalid base62 key",
			authHeader:     "Bearer invalid$$",
			setupMocks:     func(userRepo *userrepo.MockRepository, apiKeyRepo *apikeyrepo.MockRepository) {},
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
		{
			name:       "API key not found",
			authHeader: fmt.Sprintf("Bearer %s", validKeyBase62),
			setupMocks: func(userRepo *userrepo.MockRepository, apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("Get", mock.Anything, validKey).Return(apikey.APIKey{}, errors.New("not found"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
		{
			name:       "user not found",
			authHeader: fmt.Sprintf("Bearer %s", validKeyBase62),
			setupMocks: func(userRepo *userrepo.MockRepository, apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("Get", mock.Anything, validKey).Return(validStoredKey, nil)
				userRepo.On("Get", mock.Anything, validStoredKey.UserID).Return(domain.User{}, domain.ErrNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(userrepo.MockRepository)
			mockAPIKeyRepo := new(apikeyrepo.MockRepository)
			tt.setupMocks(mockUserRepo, mockAPIKeyRepo)

			// capture errors via a simple ErrorHandler
			errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
				apiErr, ok := err.(*common.APIError)
				require.True(t, ok)
				w.WriteHeader(apiErr.Status)
				err = json.NewEncoder(w).Encode(map[string]string{"error": apiErr.Message})
				require.NoError(t, err)
			}

			// Create middleware
			middleware := AuthUserInjectorMiddleware(mockUserRepo, mockAPIKeyRepo, errorHandler)

			// Wrap a dummy handler to check context
			var capturedUser domain.User
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectUser {
					u, ok := auth.UserFromContext(r.Context())
					require.True(t, ok)
					capturedUser = u
					assert.Equal(t, validUser.ID, capturedUser.ID)

					assert.Equal(t, r.Header.Get("Authorization"), "")
				}
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte("ok"))
				require.NoError(t, err)
			}))

			req := httptest.NewRequest("GET", "/somepath", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if rec.Code >= 400 {
				var body map[string]string
				err := json.NewDecoder(rec.Body).Decode(&body)
				require.NoError(t, err)
				assert.Contains(t, body, "error")
			}

			mockUserRepo.AssertExpectations(t)
			mockAPIKeyRepo.AssertExpectations(t)
		})
	}
}
