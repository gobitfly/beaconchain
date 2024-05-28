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

func newSessionManager(cfg *types.Config) *scs.SessionManager {
	// TODO: replace redis with user db down the line (or replace sessions with oauth2)
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.RedisSessionStoreEndpoint)
		},
	}

	scs := scs.New()
	scs.Lifetime = time.Hour * 24 * 7
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

func GetAuthMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			query := r.URL.Query().Get("api_key")

			if header != "Bearer "+apiKey && query != apiKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

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
	return csrf.Protect(csrfBytes, csrf.Secure(!cfg.Frontend.CsrfInsecure), csrf.Path("/"), csrf.Domain(cfg.Frontend.SessionCookieDomain), csrf.SameSite(sameSite))
}

func csrfInjecterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CSRF-Token", csrf.Token(r))
		next.ServeHTTP(w, r)
	})
}
