package db

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

type ExporterCache struct {
	cache database.RemoteCache
}

func NewExporterCache(cache database.RemoteCache) ExporterCache {
	return ExporterCache{
		cache: cache,
	}
}

func (c *ExporterCache) SetLatestEpoch(epoch uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *ExporterCache) SetLatestFinalizedEpoch(epoch uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *ExporterCache) SetLatestSlot(slot uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *ExporterCache) SetLatestProposedSlot(slot uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *ExporterCache) GetLatestEpoch() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.ClConfig.DepositChainID)

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	lastEpoch := new(big.Int).SetBytes(res)
	return lastEpoch.Uint64(), nil
}

func (c *ExporterCache) GetLatestFinalizedEpoch() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	lastEpoch := new(big.Int).SetBytes(res)
	return lastEpoch.Uint64(), nil
}

func (c *ExporterCache) GetLatestSlot() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.ClConfig.DepositChainID)

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	lastEpoch := new(big.Int).SetBytes(res)
	return lastEpoch.Uint64(), nil
}

func (c *ExporterCache) GetLatestProposedSlot() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.ClConfig.DepositChainID)

	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	lastEpoch := new(big.Int).SetBytes(res)
	return lastEpoch.Uint64(), nil
}
