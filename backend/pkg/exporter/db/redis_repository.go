package db

import (
	"context"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
)

type RedisRepository interface {
	SetValue(ctx context.Context, key string, data []byte, expirationDuration time.Duration) error
}

type redisRepository struct{}

func NewRedisRepository() RedisRepository {
	return &redisRepository{}
}

func (r *redisRepository) SetValue(ctx context.Context, key string, data []byte, expirationDuration time.Duration) error {
	return db.PersistentRedisDbClient.Set(ctx, key, data, expirationDuration).Err()
}
