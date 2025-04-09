package db2

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

type LastBlockSource interface {
	GetLastBlockInDataTable(chainID string) (uint64, error)
	GetLastBlockInBlocksTable(chainID string) (uint64, error)
}

type CachedLastBlocks struct {
	cache  database.RemoteCache
	source LastBlockSource
}

func NewCachedLastBlocks(cache database.RemoteCache, source LastBlockSource) CachedLastBlocks {
	return CachedLastBlocks{cache: cache, source: source}
}

func (c CachedLastBlocks) SetInBlocksTable(chainID string, number uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":lastBlockInBlocksTable"
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(number).Bytes(), 0)
}

func (c CachedLastBlocks) SetInDataTable(chainID string, number uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":lastBlockInDataTable"
	return c.cache.Set(ctx, key, new(big.Int).SetUint64(number).Bytes(), 0)
}

func (c CachedLastBlocks) GetInDataTable(chainID string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":lastBlockInDataTable"
	res, err := c.cache.Get(ctx, key)
	if err != nil {
		// key is not yet set, get data from bigtable and store the key in redis
		if errors.Is(err, database.ErrNotFound) {
			lastBlock, err := c.source.GetLastBlockInDataTable(chainID)
			if err != nil {
				return 0, err
			}
			return lastBlock, c.SetInDataTable(chainID, lastBlock)
		}
		return 0, err
	}
	lastBlock := new(big.Int).SetBytes(res)
	return lastBlock.Uint64(), nil
}

func (c CachedLastBlocks) GetInBlocksTable(chainID string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	key := chainID + ":lastBlockInBlocksTable"
	res, err := c.cache.Get(ctx, key)
	if err != nil {
		// key is not yet set, get data from bigtable and store the key in redis
		if errors.Is(err, database.ErrNotFound) {
			lastBlock, err := c.source.GetLastBlockInBlocksTable(chainID)
			if err != nil {
				return 0, err
			}
			return lastBlock, c.SetInBlocksTable(chainID, lastBlock)
		}
		return 0, err
	}
	lastBlock := new(big.Int).SetBytes(res)
	return lastBlock.Uint64(), nil
}

type CachedBalanceUpdates struct {
	database.RemoteCache
}

// Add returns true if the key has been added and false if it was already present
func (c CachedBalanceUpdates) Add(chainID string, address []byte, token []byte) bool {
	key := fmt.Sprintf("%s:%s:%x:%x", chainID, balanceKey, address, token)
	if _, err := c.Get(context.Background(), key); err == nil {
		// already present in cache
		return true
	}
	_ = c.Set(context.Background(), key, []byte{0x1}, utils.Day*2)
	return false
}

func (c CachedBalanceUpdates) Clear(chainID string) {
	prefix := fmt.Sprintf("%s:%s", chainID, balanceKey)
	_ = c.RemoteCache.Clear(context.Background(), prefix)
}
