package core

import (
	"context"
	"sync"
	"time"
)

// Cache stores short lived credentials (access tokens and JS tickets) so the
// platform clients do not hit WeChat for every request. Implementations should
// be thread-safe; the in-process memory cache provided by NewMemoryCache is
// suitable for single-process services, and distributed deployments typically
// back this interface with Redis.
type Cache interface {
	// Get returns the cached value for key. exist is false when the key is not
	// cached (or already expired). Implementations should return nil error
	// instead of "not found" style errors.
	Get(ctx context.Context, key string) (value string, exist bool, err error)
	// SetWithTTL stores value under key for ttl.
	SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error
}

// memoryCache is a tiny in-process cache used when the caller does not provide
// one. It is guarded by a mutex: concurrent clients commonly share one client
// (and therefore one cache) across goroutines, and an unguarded map would race.
type memoryCache struct {
	mu sync.RWMutex
	m  map[string]memoryCacheEntry
}

type memoryCacheEntry struct {
	value   string
	expires time.Time
}

// NewMemoryCache returns a goroutine-safe in-process Cache with per-entry TTLs.
// Prefer a shared/distributed cache in multi-instance deployments so token
// refresh requests do not multiply across replicas.
func NewMemoryCache() Cache {
	return &memoryCache{m: make(map[string]memoryCacheEntry)}
}

func (c *memoryCache) Get(_ context.Context, key string) (string, bool, error) {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expires) {
		return "", false, nil
	}
	return e.value, true, nil
}

func (c *memoryCache) SetWithTTL(_ context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		// Guard against servers reporting an expires_in that is too small to be
		// cached meaningfully.
		ttl = time.Minute
	}
	c.mu.Lock()
	c.m[key] = memoryCacheEntry{value: value, expires: time.Now().Add(ttl)}
	c.mu.Unlock()
	return nil
}
