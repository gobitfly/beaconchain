package database

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/coocood/freecache"
	"github.com/go-redis/redis/v8"
)

var (
	_ RemoteCache = &MemCache{}
	_ RemoteCache = Redis{}
	_ RemoteCache = FreeCache{}
	_ RemoteCache = NoopCache{}
)

type RemoteCache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Clear(ctx context.Context, prefix string) error
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

func (m *MemCache) Clear(ctx context.Context, prefix string) error {
	for v := range maps.Keys(m.values) {
		if strings.HasPrefix(v, prefix) {
			delete(m.values, v)
		}
	}
	return nil
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

func (r Redis) Clear(ctx context.Context, prefix string) error {
	iter := r.Client.Scan(ctx, 0, prefix, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.Client.Unlink(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("redis unlink %s: %w", iter.Val(), err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis clear cache: %w", err)
	}
	return nil
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

func (freeCache FreeCache) Clear(ctx context.Context, prefix string) error {
	b := []byte(prefix)
	iter := freeCache.Cache.NewIterator()
	for {
		entry := iter.Next()
		if entry == nil {
			break
		}
		if bytes.HasPrefix(entry.Key, b) {
			freeCache.Cache.Del(entry.Key)
		}
	}
	return nil
}

type NoopCache struct{}

func (n NoopCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return nil
}

func (n NoopCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, fmt.Errorf("noop cache")
}

func (n NoopCache) Clear(ctx context.Context, prefix string) error {
	return nil
}
