package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coocood/freecache"
	"github.com/go-redis/redis/v8"
)

var (
	_ RemoteCache = &MemCache{}
	_ RemoteCache = Redis{}
	_ RemoteCache = FreeCache{}
)

type RemoteCache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

type MemCache struct {
	values map[string][]byte
}

func (m *MemCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if m.values == nil {
		m.values = make(map[string][]byte)
	}
	m.values[key] = value
	if expiration != time.Duration(0) {
		time.AfterFunc(expiration, func() {
			delete(m.values, key)
		})
	}
	return nil
}

func (m *MemCache) Get(ctx context.Context, key string) ([]byte, error) {
	if m.values == nil {
		return nil, ErrNotFound
	}
	val, exist := m.values[key]
	if !exist {
		return nil, ErrNotFound
	}
	return val, nil
}

type Redis struct {
	Client *redis.Client
}

func (r Redis) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

func (r Redis) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return val, nil
}

type FreeCache struct {
	Cache *freecache.Cache
}

func (freeCache FreeCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return freeCache.Cache.Set([]byte(key), value, int(expiration.Seconds()))
}

func (freeCache FreeCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := freeCache.Cache.Get([]byte(key))
	if err != nil {
		if errors.Is(err, freecache.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return val, nil
}

type NoopCache struct{}

func (n NoopCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return nil
}

func (n NoopCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, fmt.Errorf("noop cache")
}
