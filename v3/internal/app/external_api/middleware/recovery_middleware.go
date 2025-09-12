package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/log"
)

// RecoveryMiddleware recovers from panics in HTTP handlers and returns 500.
func RecoveryMiddleware() model.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered in HTTP handler", "error", rec, "stack", string(debug.Stack()))

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
