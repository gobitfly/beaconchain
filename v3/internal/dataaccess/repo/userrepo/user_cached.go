package userrepo

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

type CachedRepository struct {
	redis *redis.Client
	Repository
}

func (r *CachedRepository) Initialize(redisClient *redis.Client, userRepo Repository) {
	r.redis = redisClient
	r.Repository = userRepo
}

func (r *CachedRepository) Get(ctx context.Context, userID uint64) (domain.User, error) {
	user, err := r.getUserByIdCache(ctx, userID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return user, err
	}

	user, err = r.Repository.Get(ctx, userID)
	if err != nil {
		return user, err
	}

	err = r.setUserCache(ctx, user)
	if err != nil {
		log.Warnf("failed to cache user %v: %v", user.ID, err)
	}

	return user, nil
}

func (r *CachedRepository) setUserCache(ctx context.Context, user domain.User) error {
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

func (r *CachedRepository) getUserByIdCache(ctx context.Context, id uint64) (domain.User, error) {
	key := fmt.Sprintf("%s%d", cacheUserPrefix, id)
	val, err := r.redis.Get(ctx, key).Bytes() // use .Bytes() instead of .Result() if storing binary
	if err == redis.Nil {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to get user from redis: %w", err)
	}

	var protoUser model.SerializableUser
	if err := proto.Unmarshal(val, &protoUser); err != nil {
		return domain.User{}, fmt.Errorf("failed to unmarshal user proto: %w", err)
	}

	return domain.User{
		ID:               protoUser.Id,
		SubscriptionTier: domain.Tier(protoUser.Tier),
	}, nil
}
