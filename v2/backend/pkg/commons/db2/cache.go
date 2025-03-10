package db2

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
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
