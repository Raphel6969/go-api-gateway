package cache

import (
	"sync"
	"time"
)

type entry struct {
	body       []byte
	statusCode int
	expiration time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]entry
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]entry),
	}

	go c.cleanupLoop()
	return c
}

func (c *MemoryCache) Set(key string, statusCode int, body []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = entry{
		body:       body,
		statusCode: statusCode,
		expiration: time.Now().Add(ttl),
	}
}

func (c *MemoryCache) Get(key string) ([]byte, int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, 0, false
	}

	if time.Now().After(item.expiration) {
		return nil, 0, false
	}

	return item.body, item.statusCode, true
}

func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
