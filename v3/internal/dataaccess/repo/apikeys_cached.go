package dataaccess

import (
	"context"
	"fmt"
	"strconv"
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

const (
	// Cache settings
	cacheAPIKeyPrefix = "apikey:"
	cacheAPIKeyTTL    = time.Minute

	// Metadata settings
	cacheAPIKeyMetaPrefix      = "apikey_meta:"
	cacheLastUsedFlushInterval = time.Minute   // Note that a higher value decreases db load but also increases inaccuracy of last used time by that value
	cacheMetaTTL               = 1 * time.Hour // Ongoing usage will keep accuracy alive during this duration, we drop the accuracy to flush interval after
	cacheMetaLastUsedField     = "lu"
	cacheMetaLastUsedFlushed   = "luf"
)

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

func (r *CachedAPIKeyRepository) UpdateLastUsedAt(
	ctx context.Context,
	key apikey.HashedKeyCredential,
) error {
	now := time.Now()

	_, lastFlushed, err := r.getAPIKeyLastUsedMeta(ctx, key)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	if lastFlushed == nil || lastFlushed.Before(time.Now().Add(-cacheLastUsedFlushInterval)) {
		// Flush to DB
		if err := r.apikeyRepo.UpdateLastUsedAt(ctx, key); err != nil {
			return err
		}
		return r.updateAPIKeyLastUsedMeta(ctx, key, &now, &now) // also update last flushed time
	}

	return r.updateAPIKeyLastUsedMeta(ctx, key, &now, nil) // only update last used time
}

func (r *CachedAPIKeyRepository) updateAPIKeyLastUsedMeta(
	ctx context.Context,
	key apikey.HashedKeyCredential,
	lastUsed *time.Time,
	lastUsedFlushed *time.Time,
) error {
	redisKey := r.getMetaRedisKey(key)
	updates := make(map[string]interface{})

	if lastUsed != nil {
		updates[cacheMetaLastUsedField] = lastUsed.Unix()
	}
	if lastUsedFlushed != nil {
		updates[cacheMetaLastUsedFlushed] = lastUsedFlushed.Unix()
	}

	if len(updates) == 0 {
		return nil
	}

	pipe := r.redis.TxPipeline()
	pipe.HSet(ctx, redisKey, updates)
	pipe.Expire(ctx, redisKey, cacheMetaTTL) // extend lifetime of metadata
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to update api key metadata in redis: %w", err)
	}
	return nil
}

func (r *CachedAPIKeyRepository) getAPIKeyLastUsedMeta(
	ctx context.Context,
	key apikey.HashedKeyCredential,
) (*time.Time, *time.Time, error) {
	redisKey := r.getMetaRedisKey(key)
	vals, err := r.redis.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to HGETALL api key metadata: %w", err)
	}
	if len(vals) == 0 {
		return nil, nil, domain.ErrNotFound
	}

	var lastUsed, lastUsedFlushed *time.Time

	if v, ok := vals[cacheMetaLastUsedField]; ok {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			t := time.Unix(ts, 0)
			lastUsed = &t
		}
	}
	if v, ok := vals[cacheMetaLastUsedFlushed]; ok {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			t := time.Unix(ts, 0)
			lastUsedFlushed = &t
		}
	}

	return lastUsed, lastUsedFlushed, nil
}

func (r *CachedAPIKeyRepository) getMetaRedisKey(key apikey.HashedKeyCredential) string {
	return fmt.Sprintf("%s%s", cacheAPIKeyMetaPrefix, key.String())
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
	keys, err := r.apikeyRepo.GetAPIKeys(ctx, userID, keyName)
	if err != nil {
		return nil, err
	}

	// fetch last used times from cache
	for i := range keys {
		lastUsed, _, err := r.getAPIKeyLastUsedMeta(ctx, apikey.HashedKeyCredential(keys[i].Value))
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			log.Warnf("failed to get api key %s metadata: %v", keys[i].Value, err)
			continue
		}
		keys[i].LastUsedAt = lastUsed
	}

	return keys, nil
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

	redisKey := r.getRedisKey(key)
	if err := r.redis.Set(ctx, redisKey, data, cacheAPIKeyTTL).Err(); err != nil {
		return fmt.Errorf("failed to set api key in redis: %w", err)
	}
	return nil
}

func (r *CachedAPIKeyRepository) deleteAPIKeyCache(ctx context.Context, key apikey.HashedKeyCredential) error {
	redisKey := r.getRedisKey(key)
	if err := r.redis.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("failed to delete api key from cache: %w", err)
	}
	return nil
}

func (r *CachedAPIKeyRepository) getRedisKey(key apikey.HashedKeyCredential) string {
	return fmt.Sprintf("%s%s", cacheAPIKeyPrefix, key.String())
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
