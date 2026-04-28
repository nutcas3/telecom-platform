package repository

import (
	"context"
	"sync"
	"time"
)

// CachedProfileStore wraps a ProfileRepository with an in-memory cache.
type CachedProfileStore struct {
	repo      ProfileRepository
	cache     map[string]*cacheEntry
	cacheMu   sync.RWMutex
	ttl       time.Duration
	cleanup   chan struct{}
	closeOnce sync.Once
}

type cacheEntry struct {
	profile   *Profile
	expiresAt time.Time
}

// NewCachedProfileStore creates a new cached profile store.
func NewCachedProfileStore(repo ProfileRepository, ttl time.Duration) *CachedProfileStore {
	c := &CachedProfileStore{
		repo:    repo,
		cache:   make(map[string]*cacheEntry),
		ttl:     ttl,
		cleanup: make(chan struct{}),
	}
	go c.cleanupLoop()
	return c
}

// Get retrieves a profile from cache or the underlying repository.
func (c *CachedProfileStore) Get(ctx context.Context, iccid string) (*Profile, error) {
	// Try cache first
	c.cacheMu.RLock()
	entry, ok := c.cache[iccid]
	c.cacheMu.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return entry.profile, nil
	}

	// Cache miss or expired, fetch from repository
	profile, err := c.repo.Get(ctx, iccid)
	if err != nil {
		return nil, err
	}

	// Update cache
	c.cacheMu.Lock()
	c.cache[iccid] = &cacheEntry{
		profile:   profile,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.cacheMu.Unlock()

	return profile, nil
}

// Create creates a profile and invalidates the cache.
func (c *CachedProfileStore) Create(ctx context.Context, p *Profile) error {
	if err := c.repo.Create(ctx, p); err != nil {
		return err
	}

	// Invalidate cache for this profile
	c.cacheMu.Lock()
	delete(c.cache, p.ICCID)
	c.cacheMu.Unlock()

	return nil
}

// List returns profiles from the repository (not cached).
func (c *CachedProfileStore) List(ctx context.Context, f ListFilter) ([]*Profile, int, error) {
	return c.repo.List(ctx, f)
}

// UpdateState updates a profile state and invalidates the cache.
func (c *CachedProfileStore) UpdateState(ctx context.Context, iccid, state string) (*Profile, error) {
	profile, err := c.repo.UpdateState(ctx, iccid, state)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for this profile
	c.cacheMu.Lock()
	delete(c.cache, iccid)
	c.cacheMu.Unlock()

	return profile, nil
}

// Delete deletes a profile and invalidates the cache.
func (c *CachedProfileStore) Delete(ctx context.Context, iccid string) error {
	if err := c.repo.Delete(ctx, iccid); err != nil {
		return err
	}

	// Invalidate cache for this profile
	c.cacheMu.Lock()
	delete(c.cache, iccid)
	c.cacheMu.Unlock()

	return nil
}

// Ping checks the underlying repository health.
func (c *CachedProfileStore) Ping() error {
	return c.repo.Ping()
}

// Close stops the cleanup goroutine.
func (c *CachedProfileStore) Close() error {
	c.closeOnce.Do(func() {
		close(c.cleanup)
	})
	return nil
}

// cleanupLoop periodically removes expired cache entries.
func (c *CachedProfileStore) cleanupLoop() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.cleanup:
			return
		}
	}
}

// cleanupExpired removes expired entries from the cache.
func (c *CachedProfileStore) cleanupExpired() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	now := time.Now()
	for iccid, entry := range c.cache {
		if now.After(entry.expiresAt) {
			delete(c.cache, iccid)
		}
	}
}

// Invalidate manually clears the cache for a specific profile.
func (c *CachedProfileStore) Invalidate(iccid string) {
	c.cacheMu.Lock()
	delete(c.cache, iccid)
	c.cacheMu.Unlock()
}

// InvalidateAll clears the entire cache.
func (c *CachedProfileStore) InvalidateAll() {
	c.cacheMu.Lock()
	c.cache = make(map[string]*cacheEntry)
	c.cacheMu.Unlock()
}
