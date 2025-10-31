package middleware

import (
	"net/http"
	"strings"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/sessionstorerepo"
)

const SessionIDCookieName = "session_id"

// AuthUserInjectorMiddleware verifies a valid user session exists and injects the user into the context.
func AuthUserInjectorMiddleware(sessionStoreRepo sessionstorerepo.Repository, errorHandler func(w http.ResponseWriter, r *http.Request, err error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mustAuthenticate := isEndpointAuthenticationRequiredHTTP(r.URL.Path)
			if !mustAuthenticate {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(SessionIDCookieName)
			if err != nil {
				authErr := common.NewAPIUserFacingError(http.StatusUnauthorized, "Missing session cookie")
				errorHandler(w, r, authErr)
				return
			}
			sessionID := cookie.Value

			user, err := sessionStoreRepo.GetUserFromSessionID(r.Context(), sessionID)
			if err != nil {
				internalErr := common.NewAPIUserFacingError(http.StatusUnauthorized, "Invalid session ID")
				errorHandler(w, r, internalErr)
				return
			}

			ctx := auth.SetUserInContext(r.Context(), user)
			r = r.WithContext(ctx)

			rStripped := r.Clone(ctx)
			rStripped.Header.Del("Cookie")

			next.ServeHTTP(w, rStripped)
		})
	}
}

func isEndpointAuthenticationRequiredHTTP(path string) bool {
	if strings.HasPrefix(path, "/healthz") || strings.HasPrefix(path, "/readyz") {
		return false
	}
	return true
}
