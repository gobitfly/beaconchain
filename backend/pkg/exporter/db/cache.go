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

func (c *ExporterCache) setValue(key string, value []byte, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	return c.cache.Set(ctx, key, value, expiration)
}

func (c *ExporterCache) getValue(key string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return new(big.Int).SetBytes(res).Uint64(), nil
}

func (c *ExporterCache) SetLatestEpoch(epoch uint64) error {
	key := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *ExporterCache) SetLatestFinalizedEpoch(epoch uint64) error {
	key := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *ExporterCache) SetLatestSlot(slot uint64) error {
	key := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *ExporterCache) SetLatestProposedSlot(slot uint64) error {
	key := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *ExporterCache) SetEpochAssignments(epoch uint64, value []byte, expiration time.Duration) error {
	key := fmt.Sprintf("%d:ea:%d", utils.Config.Chain.ClConfig.DepositChainID, epoch)
	return c.setValue(key, value, expiration)
}

func (c *ExporterCache) SetValidatorMapping(value []byte, expiration time.Duration) error {
	key := fmt.Sprintf("%d:vm", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, value, expiration)
}

func (c *ExporterCache) GetLatestEpoch() (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	return c.getValue(key)
}

func (c *ExporterCache) GetLatestFinalizedEpoch() (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)
	return c.getValue(key)
}

func (c *ExporterCache) GetLatestSlot() (uint64, error) {
	key := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.getValue(key)
}

func (c *ExporterCache) GetLatestProposedSlot() (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.getValue(key)
}
