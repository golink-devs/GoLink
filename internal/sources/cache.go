package sources

import (
	"sync"
	"time"
)

type entry struct {
	result    *LoadResult
	expiresAt time.Time
}

type TrackCache struct {
	entries map[string]*entry
	ttl     time.Duration
	mu      sync.RWMutex
}

func NewTrackCache(ttl time.Duration) *TrackCache {
	return &TrackCache{
		entries: make(map[string]*entry),
		ttl:     ttl,
	}
}

func (c *TrackCache) Get(identifier string) (*LoadResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[identifier]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.result, true
}

func (c *TrackCache) Set(identifier string, result *LoadResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[identifier] = &entry{
		result:    result,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Cleanup removes expired entries. Should be called periodically.
func (c *TrackCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.entries {
		if now.After(v.expiresAt) {
			delete(c.entries, k)
		}
	}
}
