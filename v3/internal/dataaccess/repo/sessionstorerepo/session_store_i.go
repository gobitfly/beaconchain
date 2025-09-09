package sessionstorerepo

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

type Repository interface {
	Initialize(redisClient *redis.Client)
	GetUserFromSessionID(ctx context.Context, sessionId string) (domain.User, error)
}
