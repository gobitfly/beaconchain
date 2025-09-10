package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	LegacySessionPrefix     = "scs:session:"
	UserIdSessionKey        = "user_id"
	AuthenticatedSessionKey = "authenticated"
	SubscriptionSessionKey  = "subscription"
)

func main() {
	var host = "localhost"
	var port = 6479
	var userID = 1337
	var premTier = "diamond"

	fmt.Println("This utility will create a test user session in Redis.")
	fmt.Println("You can then use the session ID to authenticate requests to the API.")
	fmt.Println("Press Enter to accept the default values in parentheses.")

	fmt.Print("Enter Redis host (default: localhost): ")
	fmt.Scanln(&host)
	if host == "" {
		host = "localhost"
	}

	fmt.Print("Enter Redis port (default: 6479): ")
	fmt.Scanln(&port)
	if port == 0 {
		port = 6479
	}

	fmt.Print("Enter User ID (default: 1337): ")
	fmt.Scanln(&userID)
	if userID == 0 {
		userID = 1337
	}

	fmt.Print("Enter Subscription Premium Tier (default: diamond): ")
	fmt.Scanln(&premTier)
	if premTier == "" {
		premTier = "diamond"
	}

	fmt.Printf("Creating session for User ID %d with package '%s' on Redis %s:%d\n", userID, premTier, host, port)

	ctx := context.Background()

	sessionID := "testsession123"
	redisKey := LegacySessionPrefix + sessionID

	aux := &struct {
		Deadline time.Time
		Values   map[string]interface{}
	}{
		Deadline: time.Now().Add(24 * time.Hour), // valid for 24h
		Values: map[string]interface{}{
			UserIdSessionKey:        userID,
			AuthenticatedSessionKey: true,
			SubscriptionSessionKey:  premTier,
		},
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(aux); err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})

	if err := rdb.Set(ctx, redisKey, buf.Bytes(), 0).Err(); err != nil {
		panic(err)
	}

	fmt.Printf("✅ Session written to Redis key: %s\n", redisKey)
}
