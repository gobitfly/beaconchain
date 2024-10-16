package db2

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/coocood/freecache"
	"math/big"
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
	db           RawStoreReader
	cache        *freecache.Cache
	hashToNumber sync.Map
}

func WithCache(reader RawStoreReader) *CachedRawStore {
	return &CachedRawStore{
		db:    reader,
		cache: freecache.NewCache(8000 * 1024 * 1024), // 5000 MB limit
	}
}

func (c *CachedRawStore) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockRawData, error) {
	key := blockKey(chainID, number)
	v, err := c.cache.Get([]byte(key))
	if err == nil {
		var block FullBlockRawData
		_ = gob.NewDecoder(bytes.NewReader(v)).Decode(&block)
		return &block, nil
	}

	block, err := c.db.ReadBlockByNumber(chainID, number)
	if block != nil {
		var buf bytes.Buffer
		_ = gob.NewEncoder(&buf).Encode(block)

		if err := c.cache.Set([]byte(key), buf.Bytes(), rawStoreTTL); err != nil {
			return nil, fmt.Errorf("cannot save in cache: %w", err)
		}

		// retrieve the block hash for caching purpose
		var mini MinimalBlock
		_ = json.Unmarshal(block.Block, &mini)
		_ = c.cache.Set([]byte(mini.Result.Hash), big.NewInt(number).Bytes(), rawStoreTTL)
	}
	return block, err
}

func (c *CachedRawStore) ReadBlockByHash(chainID uint64, hash string) (*FullBlockRawData, error) {
	v, err := c.cache.Get([]byte(hash))
	if err != nil {
		return c.db.ReadBlockByHash(chainID, hash)
	}
	number := new(big.Int).SetBytes(v).Int64()

	v, err = c.cache.Get([]byte(blockKey(chainID, number)))
	if err != nil {
		return c.db.ReadBlockByHash(chainID, hash)
	}

	var block FullBlockRawData
	_ = gob.NewDecoder(bytes.NewReader(v)).Decode(&block)
	return &block, nil
}

type CachedRawStore2 struct {
	db    RawStoreReader
	cache sync.Map
}

func WithCache2(reader RawStoreReader) *CachedRawStore2 {
	return &CachedRawStore2{
		db: reader,
	}
}

func (c *CachedRawStore2) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockRawData, error) {
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

func (c *CachedRawStore2) ReadBlockByHash(chainID uint64, hash string) (*FullBlockRawData, error) {
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
