package database

import (
	"context"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
)

type RemoteCacheWithMetrics struct {
	cache   RemoteCache
	metrics metrics.MetricsRepository
}

func NewRemoteCacheWithMetrics(cache RemoteCache, metrics metrics.MetricsRepository) RemoteCacheWithMetrics {
	return RemoteCacheWithMetrics{
		cache:   cache,
		metrics: metrics,
	}
}

func (c *RemoteCacheWithMetrics) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("cache", "set", time.Since(timeStart))
	}(timeStart)

	err := c.cache.Set(ctx, key, value, expiration)
	if err != nil {
		c.metrics.Error("cache_set")
		return err
	}
	return nil
}

func (c *RemoteCacheWithMetrics) Get(ctx context.Context, key string) ([]byte, error) {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		c.metrics.ObserveClientCallDuration("cache", "get", time.Since(timeStart))
	}(timeStart)

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		c.metrics.Error("cache_get")
		return nil, err
	}
	return res, nil
}
