package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache is a simple in-memory cache
type MemoryCache struct {
	mu       sync.RWMutex
	items    map[string]cacheItem
	ttl      time.Duration
	stopChan chan struct{}
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache with the specified TTL
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	mc := &MemoryCache{
		items:    make(map[string]cacheItem),
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go mc.startCleanup(context.Background())

	return mc
}

// startCleanup periodically removes expired items from the cache
func (mc *MemoryCache) startCleanup(ctx context.Context) {
	ticker := time.NewTicker(mc.ttl * 2) // Clean up twice as often as TTL
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mc.stopChan:
			return
		case <-ticker.C:
			mc.mu.Lock()
			now := time.Now()

			for key, item := range mc.items {
				if item.expiration.Before(now) {
					delete(mc.items, key)
				}
			}
			mc.mu.Unlock()
		}
	}
}

// Stop stops the cleanup goroutine
func (mc *MemoryCache) Stop() {
	close(mc.stopChan)
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(key string) (interface{}, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, found := mc.items[key]
	if !found {
		return nil, false
	}

	// Check if the item has expired
	if item.expiration.Before(time.Now()) {
		return nil, false
	}

	return item.value, true
}

// Set stores a value in the cache with the default TTL
func (mc *MemoryCache) Set(key string, value interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(mc.ttl),
	}
}

// Delete removes a key from the cache
func (mc *MemoryCache) Delete(key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.items, key)
}

// Clear removes all items from the cache
func (mc *MemoryCache) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]cacheItem)
}
