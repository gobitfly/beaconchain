package dataaccess

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	model "github.com/gobitfly/beaconchain-backend/domain/gen"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/log"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const cacheUserPrefix = "user:"

type CachedUserRepository struct {
	redis    *redis.Client
	userRepo UserAuthRepository
}

func (r *CachedUserRepository) Initialize(redisClient *redis.Client, userRepo UserAuthRepository) {
	r.redis = redisClient
	r.userRepo = userRepo
}

func (r *CachedUserRepository) GetUserById(ctx context.Context, userID uint64) (*domain.User, error) {
	user, err := r.getUserByIdCache(ctx, userID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	user, err = r.userRepo.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = r.setUserCache(ctx, user)
	if err != nil {
		log.Warnf("failed to cache user %v: %v", user.ID, err)
	}

	return user, nil
}

func (r *CachedUserRepository) setUserCache(ctx context.Context, user *domain.User) error {
	protoUser := &model.SerializableUser{
		Id:   user.ID,
		Tier: string(user.SubscriptionTier),
	}

	data, err := proto.Marshal(protoUser)
	if err != nil {
		return fmt.Errorf("failed to marshal user proto: %w", err)
	}

	key := fmt.Sprintf("%s%d", cacheUserPrefix, user.ID)
	if err := r.redis.Set(ctx, key, data, time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set user in redis: %w", err)
	}
	return nil
}

func (r *CachedUserRepository) getUserByIdCache(ctx context.Context, id uint64) (*domain.User, error) {
	key := fmt.Sprintf("%s%d", cacheUserPrefix, id)
	val, err := r.redis.Get(ctx, key).Bytes() // use .Bytes() instead of .Result() if storing binary
	if err == redis.Nil {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user from redis: %w", err)
	}

	var protoUser model.SerializableUser
	if err := proto.Unmarshal(val, &protoUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user proto: %w", err)
	}

	return &domain.User{
		ID:               protoUser.Id,
		SubscriptionTier: domain.Tier(protoUser.Tier),
	}, nil
}
