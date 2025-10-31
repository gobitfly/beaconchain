package apikeyrepo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain-backend/domain/gen/model/v1"
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

	// Last Used settings
	cacheAPIKeyLastUsedPrefix  = "apikey_lu:"  // #nosec G101
	cacheLastUsedFlushInterval = time.Minute   // Note that a higher value decreases db load but also increases inaccuracy of last used time by that value
	cacheLastUsedTTL           = 1 * time.Hour // Ongoing usage will keep accuracy alive during this duration, we drop the accuracy to flush interval after
	cacheLastUsedField         = "lu"
	cacheLastUsedFlushed       = "luf"
)

type CachedRepository struct {
	redis *redis.Client
	Repository
}

func (r *CachedRepository) Initialize(redisClient *redis.Client, apikeyRepo Repository) {
	r.redis = redisClient
	r.Repository = apikeyRepo
}

func (r *CachedRepository) Get(ctx context.Context, key apikey.HashedKeyCredential) (apikey.APIKey, error) {
	apiKey, err := r.getAPIKeyCache(ctx, key)
	if err == nil {
		return apiKey, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return apikey.APIKey{}, err
	}

	apiKey, err = r.Repository.Get(ctx, key)
	if err != nil {
		return apikey.APIKey{}, err
	}

	err = r.setAPIKeyCache(ctx, key, apiKey)
	if err != nil {
		log.Warnf("failed to cache API key %v: %v", key.String(), err)
	}

	return apiKey, nil
}

// UpdateLastUsedAt updates the last used timestamp of the API key.
// To reduce database load, we use a caching strategy where we only flush to the database
// if there is no unflushed last used time in cache or the last flushed time is older than a set interval.
func (r *CachedRepository) UpdateLastUsedAt(
	ctx context.Context,
	key apikey.HashedKeyCredential,
) error {
	now := time.Now()

	// check if there is an unflushed last used time in cache
	_, lastFlushed, err := r.getAPIKeyLastUsedCache(ctx, key)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	if shouldFlush(lastFlushed, cacheLastUsedFlushInterval) {
		if err := r.Repository.UpdateLastUsedAt(ctx, key); err != nil { // flush to DB
			return err
		}
		return r.updateAPIKeyLastUsedCache(ctx, key, now, &now) // also update last flushed time in cache
	}

	// otherwise only update the unflushed last used time in cache
	return r.updateAPIKeyLastUsedCache(ctx, key, now, nil)
}

func shouldFlush(lastFlushed time.Time, interval time.Duration) bool {
	// This also implicitly handles the zero-value case (never flushed):
	// time.Since(time.Time{}) is effectively "forever ago", so it will trigger a flush.
	return time.Since(lastFlushed) > interval
}

// Stores last used and last flushed times in a redis hash separate from the API key cache
func (r *CachedRepository) updateAPIKeyLastUsedCache(
	ctx context.Context,
	key apikey.HashedKeyCredential,
	lastUsed time.Time,
	lastUsedFlushed *time.Time,
) error {
	redisKey := r.formatLastUsedRedisKey(key)
	updates := make(map[string]interface{})

	updates[cacheLastUsedField] = lastUsed.UnixMilli()

	if lastUsedFlushed != nil {
		updates[cacheLastUsedFlushed] = lastUsedFlushed.UnixMilli()
	}

	pipe := r.redis.TxPipeline()
	pipe.HSet(ctx, redisKey, updates)
	pipe.Expire(ctx, redisKey, cacheLastUsedTTL) // extend lifetime of metadata
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to update api key metadata in redis: %w", err)
	}
	return nil
}

// Retrieves last used and last flushed times from redis hash
func (r *CachedRepository) getAPIKeyLastUsedCache(
	ctx context.Context,
	key apikey.HashedKeyCredential,
) (time.Time, time.Time, error) {
	redisKey := r.formatLastUsedRedisKey(key)
	vals, err := r.redis.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to HGETALL api key metadata: %w", err)
	}
	if len(vals) == 0 {
		return time.Time{}, time.Time{}, domain.ErrNotFound
	}

	var lastUsed, lastUsedFlushed time.Time

	if v, ok := vals[cacheLastUsedField]; ok {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			lastUsed = time.UnixMilli(ts)
		}
	}
	if v, ok := vals[cacheLastUsedFlushed]; ok {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			lastUsedFlushed = time.UnixMilli(ts)
		}
	}

	return lastUsed, lastUsedFlushed, nil
}

func (r *CachedRepository) formatLastUsedRedisKey(key apikey.HashedKeyCredential) string {
	return fmt.Sprintf("%s%s", cacheAPIKeyLastUsedPrefix, key.String())
}

func (r *CachedRepository) Delete(ctx context.Context, userID uint64, name string) error {
	key, err := r.Repository.GetAll(ctx, userID, &name)
	if err != nil {
		return err
	}

	if len(key) == 0 {
		return domain.ErrNotFound
	}

	err = r.Repository.Delete(ctx, userID, name)
	if err != nil {
		return err
	}

	return r.deleteAPIKeyCache(ctx, apikey.HashedKeyCredential(key[0].Value))
}

func (r *CachedRepository) Disable(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	key, err := r.Repository.Disable(ctx, userID, name)
	if err != nil {
		return apikey.APIKey{}, err
	}

	return key, r.deleteAPIKeyCache(ctx, apikey.HashedKeyCredential(key.Value))
}

func (r *CachedRepository) GetAll(ctx context.Context, userID uint64, keyName *string) ([]apikey.APIKey, error) {
	keys, err := r.Repository.GetAll(ctx, userID, keyName)
	if err != nil {
		return nil, err
	}

	// keys contain last_used_time from db but
	// we check redis cache if there is a more recent last_used time that has not been flushed yet and use that
	for i := range keys {
		lastUsed, _, err := r.getAPIKeyLastUsedCache(ctx, apikey.HashedKeyCredential(keys[i].Value))
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			log.Warnf("failed to get api key %s metadata: %v", keys[i].Value, err)
			continue
		}
		keys[i].LastUsedAt = &lastUsed
	}

	return keys, nil
}

func (r *CachedRepository) setAPIKeyCache(ctx context.Context, key apikey.HashedKeyCredential, apiKey apikey.APIKey) error {
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

	redisKey := r.formatRedisKey(key)
	if err := r.redis.Set(ctx, redisKey, data, cacheAPIKeyTTL).Err(); err != nil {
		return fmt.Errorf("failed to set api key in redis: %w", err)
	}
	return nil
}

func (r *CachedRepository) deleteAPIKeyCache(ctx context.Context, key apikey.HashedKeyCredential) error {
	redisKey := r.formatRedisKey(key)
	if err := r.redis.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("failed to delete api key from cache: %w", err)
	}
	return nil
}

func (r *CachedRepository) formatRedisKey(key apikey.HashedKeyCredential) string {
	return fmt.Sprintf("%s%s", cacheAPIKeyPrefix, key.String())
}

func (r *CachedRepository) getAPIKeyCache(ctx context.Context, key apikey.HashedKeyCredential) (apikey.APIKey, error) {
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
