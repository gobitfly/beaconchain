package db

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
)

type SlotExporterCacheRepository interface {
	SetLatestEpoch(chainID, epoch uint64) error
	SetLatestFinalizedEpoch(chainID, epoch uint64) error
	SetLatestSlot(chainID, slot uint64) error
	SetLatestProposedSlot(chainID, slot uint64) error
	SetEpochAssignments(chainID, epoch uint64, value []byte, expiration time.Duration) error
	SetValidatorMapping(chainID uint64, value []byte, expiration time.Duration) error
	GetLatestEpoch(chainID uint64) (uint64, error)
	GetLatestFinalizedEpoch(chainID uint64) (uint64, error)
	GetLatestSlot(chainID uint64) (uint64, error)
	GetLatestProposedSlot(chainID uint64) (uint64, error)
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

func (c *SlotExporterCache) SetLatestEpoch(chainID, epoch uint64) error {
	key := fmt.Sprintf("%d:frontend:latestEpoch", chainID)
	return c.setValue(key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *SlotExporterCache) SetLatestFinalizedEpoch(chainID, epoch uint64) error {
	key := fmt.Sprintf("%d:frontend:latestFinalized", chainID)
	return c.setValue(key, new(big.Int).SetUint64(epoch).Bytes(), 0)
}

func (c *SlotExporterCache) SetLatestSlot(chainID, slot uint64) error {
	key := fmt.Sprintf("%d:frontend:slot", chainID)
	return c.setValue(key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *SlotExporterCache) SetLatestProposedSlot(chainID, slot uint64) error {
	key := fmt.Sprintf("%d:frontend:latestProposedSlot", chainID)
	return c.setValue(key, new(big.Int).SetUint64(slot).Bytes(), 0)
}

func (c *SlotExporterCache) SetEpochAssignments(chainID, epoch uint64, value []byte, expiration time.Duration) error {
	key := fmt.Sprintf("%d:ea:%d", chainID, epoch)
	return c.setValue(key, value, expiration)
}

func (c *SlotExporterCache) SetValidatorMapping(chainID uint64, value []byte, expiration time.Duration) error {
	key := fmt.Sprintf("%d:vm", chainID)
	return c.setValue(key, value, expiration)
}

func (c *SlotExporterCache) GetLatestEpoch(chainID uint64) (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestEpoch", chainID)
	return c.getValue(key)
}

func (c *SlotExporterCache) GetLatestFinalizedEpoch(chainID uint64) (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestFinalized", chainID)
	return c.getValue(key)
}

func (c *SlotExporterCache) GetLatestSlot(chainID uint64) (uint64, error) {
	key := fmt.Sprintf("%d:frontend:slot", chainID)
	return c.getValue(key)
}

func (c *SlotExporterCache) GetLatestProposedSlot(chainID uint64) (uint64, error) {
	key := fmt.Sprintf("%d:frontend:latestProposedSlot", chainID)
	return c.getValue(key)
}
