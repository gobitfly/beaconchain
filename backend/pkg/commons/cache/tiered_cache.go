package cache

import (
	"context"
	"encoding/json"

	"fmt"
	"strconv"
	"time"

	"github.com/coocood/freecache"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

// Tiered cache is a cache implementation combining a
type tieredCache struct {
	localGoCache *freecache.Cache
	remoteCache  RemoteCache
}

type RemoteCache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	SetString(ctx context.Context, key, value string, expiration time.Duration) error
	SetUint64(ctx context.Context, key string, value uint64, expiration time.Duration) error
	SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error

	Get(ctx context.Context, key string, returnValue any) (any, error)
	GetString(ctx context.Context, key string) (string, error)
	GetUint64(ctx context.Context, key string) (uint64, error)
	GetBool(ctx context.Context, key string) (bool, error)
}

var TieredCache *tieredCache

func MustInitTieredCache(redisAddress string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	remoteCache, err := InitRedisCache(ctx, redisAddress)
	if err != nil {
		log.Fatal(err, "error initializing remote redis cache", 0, map[string]interface{}{"address": redisAddress})
	}

	TieredCache = &tieredCache{
		remoteCache:  remoteCache,
		localGoCache: freecache.NewCache(100 * 1024 * 1024), // 100 MB
	}
}

func (cache *tieredCache) SetString(key, value string, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	err := cache.localGoCache.Set([]byte(key), []byte(value), int(expiration.Seconds()))
	if err != nil {
		return err
	}
	return cache.remoteCache.SetString(ctx, key, value, expiration)
}

func (cache *tieredCache) GetStringWithLocalTimeout(key string, localExpiration time.Duration) (string, error) {
	// try to retrieve the key from the local cache
	wanted, err := cache.localGoCache.Get([]byte(key))
	if err == nil {
		return string(wanted), nil
	}

	// retrieve the key from the remote cache
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	value, err := cache.remoteCache.GetString(ctx, key)
	if err != nil {
		return "", err
	}

	err = cache.localGoCache.Set([]byte(key), []byte(value), int(localExpiration.Seconds()))
	if err != nil {
		return "", err
	}
	return value, nil
}

func (cache *tieredCache) SetUint64(key string, value uint64, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	err := cache.localGoCache.Set([]byte(key), []byte(fmt.Sprintf("%d", value)), int(expiration.Seconds()))
	if err != nil {
		return err
	}
	return cache.remoteCache.SetUint64(ctx, key, value, expiration)
}

func (cache *tieredCache) GetUint64WithLocalTimeout(key string, localExpiration time.Duration) (uint64, error) {
	// try to retrieve the key from the local cache
	wanted, err := cache.localGoCache.Get([]byte(key))
	if err == nil {
		returnValue, err := strconv.ParseUint(string(wanted), 10, 64)
		if err != nil {
			return 0, err
		}
		return returnValue, nil
	}

	// retrieve the key from the remote cache
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	value, err := cache.remoteCache.GetUint64(ctx, key)
	if err != nil {
		return 0, err
	}

	err = cache.localGoCache.Set([]byte(key), []byte(fmt.Sprintf("%d", value)), int(localExpiration.Seconds()))
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (cache *tieredCache) SetBool(key string, value bool, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	err := cache.localGoCache.Set([]byte(key), []byte(fmt.Sprintf("%t", value)), int(expiration.Seconds()))
	if err != nil {
		return err
	}
	return cache.remoteCache.SetBool(ctx, key, value, expiration)
}

func (cache *tieredCache) GetBoolWithLocalTimeout(key string, localExpiration time.Duration) (bool, error) {
	// try to retrieve the key from the local cache
	wanted, err := cache.localGoCache.Get([]byte(key))
	if err == nil {
		returnValue, err := strconv.ParseBool(string(wanted))
		if err != nil {
			return false, err
		}
		return returnValue, nil
	}

	// retrieve the key from the remote cache
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	value, err := cache.remoteCache.GetBool(ctx, key)
	if err != nil {
		return false, err
	}

	err = cache.localGoCache.Set([]byte(key), []byte(fmt.Sprintf("%t", value)), int(localExpiration.Seconds()))
	if err != nil {
		return false, err
	}
	return value, nil
}

func (cache *tieredCache) Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	valueMarshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = cache.localGoCache.Set([]byte(key), valueMarshal, int(expiration.Seconds()))
	if err != nil {
		return err
	}
	return cache.remoteCache.Set(ctx, key, value, expiration)
}

func (cache *tieredCache) GetWithLocalTimeout(key string, localExpiration time.Duration, returnValue interface{}) (interface{}, error) {
	// try to retrieve the key from the local cache
	wanted, err := cache.localGoCache.Get([]byte(key))
	if err == nil {
		err = json.Unmarshal(wanted, returnValue)
		if err != nil {
			log.Error(err, "error unmarshalling data for key", 0, map[string]interface{}{"key": key})
			return nil, err
		}
		return returnValue, nil
	}

	// retrieve the key from the remote cache
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	value, err := cache.remoteCache.Get(ctx, key, returnValue)
	if err != nil {
		return nil, err
	}

	valueMarshal, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	err = cache.localGoCache.Set([]byte(key), valueMarshal, int(localExpiration.Seconds()))
	if err != nil {
		return nil, err
	}
	return value, nil
}
