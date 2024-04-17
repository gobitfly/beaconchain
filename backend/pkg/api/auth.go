package api

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
)

func NewSessionManager(redisEndpoint string, secure bool) *scs.SessionManager {
	// TODO: replace redis with user db down the line (or replace sessions with oauth2)
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisEndpoint)
		},
	}

	scs := scs.New()
	scs.Lifetime = time.Hour * 24 * 7
	scs.Cookie.Name = "session_id"
	scs.Cookie.HttpOnly = true
	scs.Cookie.Persist = true
	scs.Cookie.SameSite = http.SameSiteNoneMode
	scs.Cookie.Secure = secure

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
