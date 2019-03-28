package cache

import (
	"github.com/freeznet/tomato/config"
	"github.com/freeznet/tomato/dependencies/lru"
)

type lruCacheAdapter struct {
	ttl int64
	cache *lru.Cache
}

func newLRUCacheAdapter(ttl int64, maxSize int) *lruCacheAdapter {
	if ttl == 0 {
		ttl = config.TConfig.CacheTTL
	}

	if maxSize ==0 {
		maxSize = config.TConfig.CacheMaxSize
	}

	lru := &lruCacheAdapter{
		ttl: ttl,
		cache: lru.New(maxSize),
	}

	return lru
}

func (lru *lruCacheAdapter) get(key string) interface{} {
	if record, ok := lru.cache.Get(key); ok {
		return record
	}
	return nil
}

func (lru *lruCacheAdapter) put(key string, value interface{}, ttl int64)  {
	lru.ttl = ttl
	lru.cache.Add(key, value)
}

func (lru *lruCacheAdapter) del(key string)  {
	lru.cache.Remove(key)
}

func (lru *lruCacheAdapter) clear()  {
	lru.cache.Clear()
}