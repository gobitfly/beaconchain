package dataaccess

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	model "github.com/gobitfly/beaconchain-backend/domain/gen"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"github.com/google/uuid"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const cacheAPIKeyPrefix = "apikey:"

type CachedAPIKeyRepository struct {
	redis      *redis.Client
	apikeyRepo APIKeyRepository
}

func (r *CachedAPIKeyRepository) Initialize(redisClient *redis.Client, apikeyRepo APIKeyRepository) {
	r.redis = redisClient
	r.apikeyRepo = apikeyRepo
}

func (r *CachedAPIKeyRepository) GetAPIKey(ctx context.Context, key apikey.HashedKeyCredential) (apikey.APIKey, error) {
	apiKey, err := r.getAPIKeyCache(ctx, key)
	if err == nil {
		return apiKey, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return apikey.APIKey{}, err
	}

	apiKey, err = r.apikeyRepo.GetAPIKey(ctx, key)
	if err != nil {
		return apikey.APIKey{}, err
	}

	err = r.setAPIKeyCache(ctx, key, apiKey)
	if err != nil {
		log.Warnf("failed to cache API key %v: %v", key.String(), err)
	}

	return apiKey, nil
}

func (r *CachedAPIKeyRepository) UpdateLastUsedAt(ctx context.Context, key apikey.HashedKeyCredential) error {
	return r.apikeyRepo.UpdateLastUsedAt(ctx, key) // todo caching not part of mvp
}

func (r *CachedAPIKeyRepository) CreateAPIKey(ctx context.Context, userID uint64, key apikey.APIKey) (apikey.APIKey, error) {
	return r.apikeyRepo.CreateAPIKey(ctx, userID, key)
}

func (r *CachedAPIKeyRepository) DeleteAPIKey(ctx context.Context, userID uint64, name string) error {
	key, err := r.apikeyRepo.GetAPIKeys(ctx, userID, &name)
	if err != nil {
		return err
	}

	if len(key) == 0 {
		return domain.ErrNotFound
	}

	err = r.apikeyRepo.DeleteAPIKey(ctx, userID, name)
	if err != nil {
		return err
	}

	return r.deleteAPIKeyCache(ctx, apikey.HashedKeyCredential(key[0].Value))
}

func (r *CachedAPIKeyRepository) DisableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	key, err := r.apikeyRepo.DisableAPIKey(ctx, userID, name)
	if err != nil {
		return apikey.APIKey{}, err
	}

	return key, r.deleteAPIKeyCache(ctx, apikey.HashedKeyCredential(key.Value))
}

func (r *CachedAPIKeyRepository) EnableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	return r.apikeyRepo.EnableAPIKey(ctx, userID, name)
}

func (r *CachedAPIKeyRepository) GetAPIKeys(ctx context.Context, userID uint64, keyName *string) ([]apikey.APIKey, error) {
	return r.apikeyRepo.GetAPIKeys(ctx, userID, keyName)
}

func (r *CachedAPIKeyRepository) setAPIKeyCache(ctx context.Context, key apikey.HashedKeyCredential, apiKey apikey.APIKey) error {
	protoAPIKey := &model.SerializableAPIKey{
		UserId: apiKey.UserID,
		ApiKeyId: &model.UUID{
			Value: apiKey.ID[:],
		},
	}

	data, err := proto.Marshal(protoAPIKey)
	if err != nil {
		return fmt.Errorf("failed to marshal api key proto: %w", err)
	}

	redisKey := fmt.Sprintf("%s%s", cacheAPIKeyPrefix, key.String())
	if err := r.redis.Set(ctx, redisKey, data, time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set api key in redis: %w", err)
	}
	return nil
}

func (r *CachedAPIKeyRepository) deleteAPIKeyCache(ctx context.Context, key apikey.HashedKeyCredential) error {
	redisKey := fmt.Sprintf("%s%s", cacheAPIKeyPrefix, key.String())
	if err := r.redis.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("failed to delete api key from cache: %w", err)
	}
	return nil
}

func (r *CachedAPIKeyRepository) getAPIKeyCache(ctx context.Context, key apikey.HashedKeyCredential) (apikey.APIKey, error) {
	redisKey := fmt.Sprintf("%s%s", cacheAPIKeyPrefix, key.String())
	val, err := r.redis.Get(ctx, redisKey).Bytes()
	if err == redis.Nil {
		return apikey.APIKey{}, domain.ErrNotFound
	}
	if err != nil {
		return apikey.APIKey{}, fmt.Errorf("failed to get api key from redis: %w", err)
	}

	var protoAPIKey model.SerializableAPIKey
	if err := proto.Unmarshal(val, &protoAPIKey); err != nil {
		return apikey.APIKey{}, fmt.Errorf("failed to unmarshal user proto: %w", err)
	}

	uuid, err := uuid.FromBytes(protoAPIKey.ApiKeyId.Value)
	if err != nil {
		return apikey.APIKey{}, fmt.Errorf("failed to parse UUID from bytes: %w", err)
	}

	return apikey.APIKey{
		ID:     &uuid,
		UserID: protoAPIKey.UserId,
	}, nil
}
