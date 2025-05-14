package api

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gomodule/redigo/redis"
)

var day time.Duration = time.Hour * 24
var sessionDuration time.Duration = day * 365

func newSessionManager(cfg *types.Config) *scs.SessionManager {
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.RedisSessionStoreEndpoint)
		},
	}

	scs := scs.New()
	scs.Lifetime = sessionDuration
	scs.Cookie.Name = "session_id"
	scs.Cookie.HttpOnly = true
	scs.Cookie.Persist = true
	scs.Cookie.Domain = cfg.Frontend.SessionCookieDomain
	sameSite := http.SameSiteLaxMode
	secure := !cfg.Frontend.Debug
	if cfg.Frontend.SessionSameSiteNone {
		sameSite = http.SameSiteNoneMode
		secure = true
	}
	scs.Cookie.Secure = secure
	scs.Cookie.SameSite = sameSite

	scs.Store = redisstore.New(pool)

	return scs
}

// returns a middleware that extends the session expiration if the session is older than 1 day
func getSlidingSessionExpirationMiddleware(scs *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deadline := scs.Deadline(r.Context()) // unauthenticated requests have deadline set to now+sessionDuration
			if time.Until(deadline) < sessionDuration-day {
				scs.SetDeadline(r.Context(), time.Now().Add(sessionDuration).UTC()) // setting to utc because library also does that internally
			}
			next.ServeHTTP(w, r)
		})
	}
}

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if body is not empty, check if content type is set to json
		if (r.Method == http.MethodPost || r.Method == http.MethodPut) && r.ContentLength > 0 && r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "bad request: Content-Type header must be application/json", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
