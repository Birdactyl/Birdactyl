package services

import (
	"sync"
	"time"
)

type cacheItem struct {
	value      interface{}
	expiration int64
}

type MemoryCache struct {
	items map[string]cacheItem
	mu    sync.RWMutex
}

var Cache = &MemoryCache{
	items: make(map[string]cacheItem),
}

func init() {
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			Cache.Cleanup()
		}
	}()
}

func (c *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(duration).UnixNano(),
	}
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}
	if time.Now().UnixNano() > item.expiration {
		return nil, false
	}
	return item.value, true
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *MemoryCache) DeletePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.items, k)
		}
	}
}

func (c *MemoryCache) Cleanup() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if now > v.expiration {
			delete(c.items, k)
		}
	}
}
