package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/log"
)

const (
	AuthHeader = "Authorization"
)

// AuthUserInjectorMiddleware authenticates via API key and injects the user into the request context.
func AuthUserInjectorMiddleware(userRepo userrepo.AuthRepository, apiKeyRepo apikeyrepo.AuthRepository, errorHandler func(w http.ResponseWriter, r *http.Request, err error)) model.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(AuthHeader)
			if authHeader == "" {
				err := common.NewAPIUserFacingError(http.StatusUnauthorized, "missing authorization header")
				errorHandler(w, r, err)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				err := common.NewAPIUserFacingError(http.StatusUnauthorized, "invalid authorization format")
				errorHandler(w, r, err)
				return
			}

			base62Key := strings.TrimPrefix(authHeader, "Bearer ")
			apiKey, err := apikey.FromBase62(base62Key)
			if err != nil {
				err := common.NewAPIUserFacingError(http.StatusUnauthorized, "invalid authorization format")
				errorHandler(w, r, err)
				return
			}

			key, err := apiKeyRepo.Get(r.Context(), apiKey)
			if err != nil {
				err := common.NewAPIUserFacingError(http.StatusUnauthorized, "invalid API key")
				errorHandler(w, r, err)
				return
			}

			user, err := userRepo.Get(r.Context(), key.UserID)
			if err != nil {
				if !errors.Is(err, domain.ErrNotFound) {
					log.Error(fmt.Errorf("failed to get user by API key: %v", err))
				}
				err := common.NewAPIUserFacingError(http.StatusUnauthorized, "invalid API key")
				errorHandler(w, r, err)
				return
			}

			// Optional: update last-used timestamp for the API key
			if err := apiKeyRepo.UpdateLastUsedAt(r.Context(), apiKey); err != nil {
				log.Warn(fmt.Errorf("failed to update last used time for API key: %v", err))
			}

			// Inject user and API key into context
			ctx := r.Context()
			ctx = auth.SetUserInContext(ctx, user)
			ctx = auth.SetAPIKeyInContext(ctx, apiKey.String())

			// Remove Authorization header before passing to next handler
			rStripped := r.Clone(ctx)
			rStripped.Header.Del(AuthHeader)

			next.ServeHTTP(w, rStripped)
		})
	}
}
