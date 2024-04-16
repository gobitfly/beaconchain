package api

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
)

func NewSessionManager(redisEndpoint string) *scs.SessionManager {
	// TODO: replace redis with user db down the line (or replace sessions with oauth2)
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisEndpoint)
		},
	}

	scs := scs.New()
	// TODO: change to 1 week before merging
	scs.Lifetime = time.Minute * 10
	scs.Cookie.Name = "session_id"
	scs.Cookie.HttpOnly = true
	scs.Cookie.Persist = true
	scs.Cookie.SameSite = http.SameSiteLaxMode
	// TODO: change to true before merging
	scs.Cookie.Secure = false

	scs.Store = redisstore.New(pool)

	return scs
}
