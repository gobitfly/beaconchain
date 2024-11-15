package raw

import (
	"encoding/json"
	"sync"
	"time"
)

const (
	oneBlockTTL = 1 * time.Second
	blocksTTL   = 30 * time.Second // default ttl, if read it will be deleted sooner
)

type MinimalBlock struct {
	Result struct {
		Hash string `json:"hash"`
	} `json:"result"`
}

type CachedStore struct {
	store StoreReader
	// sync.Map with manual delete have better perf than freecache because we can handle this way a ttl < 1s
	cache sync.Map

	locks   map[string]*sync.RWMutex
	mapLock sync.Mutex // to make the map safe concurrently
}

func WithCache(reader StoreReader) *CachedStore {
	return &CachedStore{
		store: reader,
		locks: make(map[string]*sync.RWMutex),
	}
}

func (c *CachedStore) lockBy(key string) func() {
	c.mapLock.Lock()
	defer c.mapLock.Unlock()

	lock, found := c.locks[key]
	if !found {
		lock = &sync.RWMutex{}
		c.locks[key] = lock
		lock.Lock()
		return lock.Unlock
	}
	lock.RLock()
	return lock.RUnlock
}

func (c *CachedStore) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockData, error) {
	key := blockKey(chainID, number)

	unlock := c.lockBy(key)
	defer unlock()

	v, ok := c.cache.Load(key)
	if ok {
		// once read ensure to delete it from the cache
		go c.unCacheBlockAfter(key, "", oneBlockTTL)
		return v.(*FullBlockData), nil
	}
	// TODO make warning not found in cache
	block, err := c.store.ReadBlockByNumber(chainID, number)
	if block != nil {
		c.cacheBlock(block, oneBlockTTL)
	}
	return block, err
}

func (c *CachedStore) cacheBlock(block *FullBlockData, ttl time.Duration) {
	key := blockKey(block.ChainID, block.BlockNumber)
	c.cache.Store(key, block)

	var mini MinimalBlock
	if len(block.Uncles) != 0 {
		// retrieve the block hash for caching but only if the block has uncle(s)
		_ = json.Unmarshal(block.Block, &mini)
		c.cache.Store(mini.Result.Hash, block.BlockNumber)
	}

	go c.unCacheBlockAfter(key, mini.Result.Hash, ttl)
}

func (c *CachedStore) unCacheBlockAfter(key, hash string, ttl time.Duration) {
	time.Sleep(ttl)
	c.cache.Delete(key)
	c.mapLock.Lock()
	if hash != "" {
		c.cache.Delete(hash)
	}
	defer c.mapLock.Unlock()
	delete(c.locks, key)
}

func (c *CachedStore) ReadBlockByHash(chainID uint64, hash string) (*FullBlockData, error) {
	v, ok := c.cache.Load(hash)
	if !ok {
		return c.store.ReadBlockByHash(chainID, hash)
	}

	v, ok = c.cache.Load(blockKey(chainID, v.(int64)))
	if !ok {
		return c.store.ReadBlockByHash(chainID, hash)
	}

	return v.(*FullBlockData), nil
}

func (c *CachedStore) ReadBlocksByNumber(chainID uint64, start, end int64) ([]*FullBlockData, error) {
	blocks, err := c.store.ReadBlocksByNumber(chainID, start, end)
	if err != nil {
		return nil, err
	}
	for _, block := range blocks {
		c.cacheBlock(block, blocksTTL)
	}
	return blocks, nil
}
