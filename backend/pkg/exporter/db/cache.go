package db

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

type SlotExporterCacheRepository interface {
	SetLatestEpoch(epoch uint64) error
	SetLatestFinalizedEpoch(epoch uint64) error
	SetLatestSlot(slot uint64) error
	SetLatestProposedSlot(slot uint64) error
	SetEpochAssignments(epoch uint64, value []byte, expiration time.Duration) error
	SetValidatorMapping(value []byte, expiration time.Duration) error
	GetLatestEpoch() (uint64, error)
	GetLatestFinalizedEpoch() (uint64, error)
	GetLatestSlot() (uint64, error)
	GetLatestProposedSlot() (uint64, error)
}

type SlotExporterCache struct {
	cache database.RemoteCache
}

func NewSlotExporterCache(cache database.RemoteCache) *SlotExporterCache {
	return &SlotExporterCache{
		cache: cache,
	}
}

func (c *SlotExporterCache) setValue(key string, value []byte, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	return c.cache.Set(ctx, key, value, expiration)
}

func (c *SlotExporterCache) getValue(key string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	res, err := c.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return new(big.Int).SetBytes(res).Uint64(), nil
}

func (c *SlotExporterCache) SetLatestEpoch(epoch uint64) error {
	key := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *SlotExporterCache) SetLatestFinalizedEpoch(epoch uint64) error {
	key := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *SlotExporterCache) SetLatestSlot(slot uint64) error {
	key := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *SlotExporterCache) SetLatestProposedSlot(slot uint64) error {
	key := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *SlotExporterCache) SetEpochAssignments(epoch uint64, value []byte, expiration time.Duration) error {
	key := fmt.Sprintf("%d:ea:%d", utils.Config.Chain.ClConfig.DepositChainID, epoch)
	return c.setValue(key, value, expiration)
}

func (c *SlotExporterCache) SetValidatorMapping(value []byte, expiration time.Duration) error {
	key := fmt.Sprintf("%d:vm", utils.Config.Chain.ClConfig.DepositChainID)
	return c.setValue(key, value, expiration)
}

func (c *SlotExporterCache) GetLatestEpoch() (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	return c.getValue(key)
}

func (c *SlotExporterCache) GetLatestFinalizedEpoch() (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)
	return c.getValue(key)
}

func (c *SlotExporterCache) GetLatestSlot() (uint64, error) {
	key := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.getValue(key)
}

func (c *SlotExporterCache) GetLatestProposedSlot() (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.ClConfig.DepositChainID)
	return c.getValue(key)
}
