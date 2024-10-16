package db2

import (
	"encoding/json"
	"sync"
	"time"
)

var ttl2 = 200 * time.Millisecond

type MinimalBlock struct {
	Result struct {
		Hash string `json:"hash"`
	} `json:"result"`
}

type CachedRawStore struct {
	db RawStoreReader
	// sync.Map with manual delete have better perf than freecache because we can handle this way a ttl < 1s
	cache sync.Map
}

func WithCache(reader RawStoreReader) *CachedRawStore {
	return &CachedRawStore{
		db: reader,
	}
}

func (c *CachedRawStore) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockRawData, error) {
	key := blockKey(chainID, number)
	v, ok := c.cache.Load(key)
	if ok {
		return v.(*FullBlockRawData), nil
	}

	block, err := c.db.ReadBlockByNumber(chainID, number)
	if block != nil {
		c.cache.Store(key, block)

		// retrieve the block hash for caching purpose
		var mini MinimalBlock
		_ = json.Unmarshal(block.Block, &mini)
		c.cache.Store(mini.Result.Hash, number)
		go func() {
			time.Sleep(ttl2)
			c.cache.Delete(key)
			c.cache.Delete(mini.Result.Hash)
		}()
	}
	return block, err
}

func (c *CachedRawStore) ReadBlockByHash(chainID uint64, hash string) (*FullBlockRawData, error) {
	v, ok := c.cache.Load(hash)
	if !ok {
		return c.db.ReadBlockByHash(chainID, hash)
	}

	v, ok = c.cache.Load(blockKey(chainID, v.(int64)))
	if !ok {
		return c.db.ReadBlockByHash(chainID, hash)
	}

	return v.(*FullBlockRawData), nil
}
