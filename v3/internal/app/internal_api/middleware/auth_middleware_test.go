package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/sessionstorerepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type expectUserKeyType string

const expectUserKey = expectUserKeyType("expect_user")

// Helper to create a request with a specific cookie
func newRequestWithCookie(path, sessionID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if sessionID != "" {
		req.AddCookie(&http.Cookie{Name: SessionIDCookieName, Value: sessionID})
	}
	return req
}

// Helper to capture the error passed to the errorHandler
type ErrorCapture struct {
	Err error
}

func (ec *ErrorCapture) Handler(w http.ResponseWriter, r *http.Request, err error) {
	ec.Err = err
	// Write a 401/400 response so the test can check the ResponseWriter status
	var status int
	var msg string
	var apiErr *common.APIError
	if errors.As(err, &apiErr) {
		status = apiErr.Status
		if apiErr.Visibility == common.ErrorVisibilityUser {
			msg = apiErr.Message
		}
	} else {
		status = http.StatusInternalServerError
	}
	w.WriteHeader(status)
	if msg != "" {
		_, err := w.Write([]byte(msg))
		if err != nil {
			ec.Err = err
		}
	}
}

// --- TestInternalAuthUserInjectorInterceptor (HTTP Version) ---
func TestAuthUserInjectorMiddleware(t *testing.T) {
	validUser := domain.User{ID: 4242}

	// This is the Mock next handler that runs *after* the middleware.
	// We use it to verify the context and stripped headers.
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		// Assert that the cookie header is stripped in the request received by the next handler
		assert.Empty(t, r.Header.Get("Cookie"), "The 'Cookie' header must be stripped from the request")

		// Verify user injection status based on a flag set in the test case
		if r.Context().Value(expectUserKey) != nil {
			// Verify user is injected
			user, ok := auth.UserFromContext(r.Context())
			require.True(t, ok)
			require.Equal(t, validUser.ID, user.ID)
		} else {
			// Verify user is NOT injected
			_, ok := auth.UserFromContext(r.Context())
			assert.False(t, ok)
		}
	})

	// Helper function to set the expectation flag in the context for the nextHandler
	injectExpectation := func(r *http.Request, expectUser bool) *http.Request {
		if expectUser {
			return r.WithContext(context.WithValue(r.Context(), expectUserKey, true))
		}
		return r
	}

	tests := []struct {
		name                 string
		path                 string
		sessionID            string
		setupMocks           func(sessionRepo *sessionstorerepo.MockRepository)
		expectedHTTPStatus   int // The status written by the errorHandler or nextHandler
		expectedErrorCapture bool
		expectedErrorMsg     string // Message if expectedErrorCapture is true
		expectedErrorCode    int
		expectUserInCtx      bool
	}{
		{
			name:      "successfully injects user and strips cookie",
			path:      "/api-keys",
			sessionID: "sess-abc",
			setupMocks: func(sessionRepo *sessionstorerepo.MockRepository) {
				sessionRepo.On("GetUserFromSessionID", mock.Anything, "sess-abc").Return(validUser, nil)
			},
			expectedHTTPStatus: http.StatusOK,
			expectUserInCtx:    true,
		},
		{
			name:                 "missing cookie returns 401",
			path:                 "/api-keys",
			sessionID:            "", // No cookie set
			setupMocks:           func(_ *sessionstorerepo.MockRepository) {},
			expectedHTTPStatus:   http.StatusUnauthorized,
			expectedErrorCapture: true,
			expectedErrorMsg:     "Missing session cookie",
		},
		{
			name:      "invalid session id returns unauthenticated internal error",
			path:      "/api-keys",
			sessionID: "sess-bad",
			setupMocks: func(sessionRepo *sessionstorerepo.MockRepository) {
				sessionRepo.On("GetUserFromSessionID", mock.Anything, "sess-bad").Return(domain.User{}, errors.New("not found"))
			},
			expectedHTTPStatus:   http.StatusUnauthorized, // Default for unhandled/internal errors
			expectedErrorCapture: true,
			expectedErrorMsg:     "Invalid session ID",
		},
		{
			name:               "health check bypasses auth",
			path:               "/healthz", // Path that is exempt from auth
			sessionID:          "",
			setupMocks:         func(_ *sessionstorerepo.MockRepository) {},
			expectedHTTPStatus: http.StatusOK,
			expectUserInCtx:    false,
		},
		{
			name:               "ready check bypasses auth",
			path:               "/readyz", // Path that is exempt from auth
			sessionID:          "",
			setupMocks:         func(_ *sessionstorerepo.MockRepository) {},
			expectedHTTPStatus: http.StatusOK,
			expectUserInCtx:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionRepo := new(sessionstorerepo.MockRepository)
			if tt.setupMocks != nil {
				tt.setupMocks(mockSessionRepo)
			}

			errorCapture := &ErrorCapture{}
			middleware := AuthUserInjectorMiddleware(mockSessionRepo, errorCapture.Handler)

			req := newRequestWithCookie(tt.path, tt.sessionID)
			// Inject the flag for the nextHandler to check expectations
			req = injectExpectation(req, tt.expectUserInCtx)

			rr := httptest.NewRecorder()

			// Wrap the next handler with the middleware
			handlerToServe := middleware(nextHandler)
			handlerToServe.ServeHTTP(rr, req)

			// --- Assertions ---
			assert.Equal(t, tt.expectedHTTPStatus, rr.Code, "HTTP status code mismatch")

			if tt.path == "/health" {
				// Ensure repo was not called for health check
				mockSessionRepo.AssertNotCalled(t, "GetUserFromSessionID", mock.Anything, mock.Anything)
			} else if tt.sessionID != "" {
				mockSessionRepo.AssertCalled(t, "GetUserFromSessionID", mock.Anything, tt.sessionID)
			}

			if tt.expectedErrorCapture {
				require.NotNil(t, errorCapture.Err)
				// Check for user-facing error message and status
				if userErr, ok := errorCapture.Err.(*common.APIError); ok {
					assert.Equal(t, tt.expectedErrorMsg, userErr.Message)
					assert.Equal(t, tt.expectedHTTPStatus, userErr.Status)
				} else {
					t.Fatalf("Captured error was not a recognized type: %v", errorCapture.Err)
				}
			} else {
				assert.Nil(t, errorCapture.Err, "Expected no error to be captured by the error handler")
			}

			mockSessionRepo.AssertExpectations(t)
		})
	}
}
