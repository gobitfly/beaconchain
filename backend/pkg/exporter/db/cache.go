package db

import (
	"context"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
)

type ExporterCacher struct {
	cache database.RemoteCache
}

func NewCachedLastBlocks(cache database.RemoteCache) ExporterCacher {
	return ExporterCacher{
		cache: cache,
	}
}

func (c *ExporterCacher) SetLatestEpoch(chainID string, epoch uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":frontend:latestEpoch"
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *ExporterCacher) SetLatestFinalizedEpoch(chainID string, epoch uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":frontend:latestFinalized"
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *ExporterCacher) SetLatestSlot(chainID string, slot uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":frontend:slot"
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *ExporterCacher) SetLatestProposedSlot(chainID string, slot uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":frontend:latestProposedSlot"
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *ExporterCacher) GetLatestEpoch(chainID string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":frontend:latestEpoch"

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	lastEpoch := new(big.Int).SetBytes(res)
	return lastEpoch.Uint64(), nil
}

func (c *ExporterCacher) GetLatestFinalizedEpoch(chainID string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":frontend:latestFinalized"

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	lastEpoch := new(big.Int).SetBytes(res)
	return lastEpoch.Uint64(), nil
}

func (c *ExporterCacher) GetLatestSlot(chainID string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":frontend:slot"

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	lastEpoch := new(big.Int).SetBytes(res)
	return lastEpoch.Uint64(), nil
}

func (c *ExporterCacher) GetLatestProposedSlot(chainID string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":frontend:latestProposedSlot"

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	lastEpoch := new(big.Int).SetBytes(res)
	return lastEpoch.Uint64(), nil
}
