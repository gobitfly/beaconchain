package api

import (
	"encoding/hex"
	"net/http"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/csrf"
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

// returns goriila/csrf middleware with the given config settings
func getCsrfProtectionMiddleware(cfg *types.Config) func(http.Handler) http.Handler {
	csrfBytes, err := hex.DecodeString(cfg.Frontend.CsrfAuthKey)
	if err != nil {
		log.Fatal(err, "error decoding cfg.Frontend.CsrfAuthKey, set it to a valid hex string", 0)
	}
	if len(csrfBytes) == 0 {
		log.Warn("CSRF auth key is empty, unsafe requests will not work! Set cfg.Frontend.Debug to true to disable CSRF protection or cfg.Frontend.CsrfAuthKey.")
	}
	sameSite := csrf.SameSiteStrictMode
	if cfg.Frontend.SessionSameSiteNone {
		sameSite = csrf.SameSiteNoneMode
	}

	return csrf.Protect(
		csrfBytes,
		csrf.Secure(!cfg.Frontend.CsrfInsecure),
		csrf.Path("/"),
		csrf.Domain(cfg.Frontend.SessionCookieDomain),
		csrf.SameSite(sameSite),
	)
}

// returns a middleware that injects the CSRF token into the response headers
func csrfInjecterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CSRF-Token", csrf.Token(r))
		next.ServeHTTP(w, r)
	})
}
