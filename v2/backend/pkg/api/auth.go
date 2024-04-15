package api

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
)

func NewSessionManager(redisEndpoint string) *scs.SessionManager {
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
	scs.Cookie.SameSite = http.SameSiteLaxMode
	scs.Cookie.Secure = true

	scs.Store = redisstore.New(pool)

	return scs
}
