package data_sources

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
)

func InitRedisCache(ctx context.Context, redisConfig *config.RedisConfig) (*redis.Client, error) {
	rdc := redis.NewClient(&redis.Options{
		Addr:        redisConfig.Endpoint,
		ReadTimeout: time.Second * 3,
	})

	if err := rdc.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdc, nil
}
