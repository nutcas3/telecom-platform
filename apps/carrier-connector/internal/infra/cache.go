package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheStrategy defines caching strategies
type CacheStrategy string

const (
	CacheStrategyLRU        CacheStrategy = "lru"
	CacheStrategyTTL        CacheStrategy = "ttl"
	CacheStrategyWriteThru  CacheStrategy = "write_through"
	CacheStrategyWriteBack  CacheStrategy = "write_back"
	CacheStrategyReadThru   CacheStrategy = "read_through"
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Key       string
	Value     []byte
	ExpiresAt time.Time
	CreatedAt time.Time
	HitCount  int64
}

// Cache provides in-memory caching with multiple strategies
type Cache struct {
	entries    map[string]*CacheEntry
	mu         sync.RWMutex
	maxSize    int
	defaultTTL time.Duration
	strategy   CacheStrategy
	stats      CacheStats
}

// CacheConfig configures the cache
type CacheConfig struct {
	MaxSize    int
	DefaultTTL time.Duration
	Strategy   CacheStrategy
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxSize:    10000,
		DefaultTTL: 5 * time.Minute,
		Strategy:   CacheStrategyTTL,
	}
}

// NewCache creates a new cache instance
func NewCache(config CacheConfig) *Cache {
	c := &Cache{
		entries:    make(map[string]*CacheEntry),
		maxSize:    config.MaxSize,
		defaultTTL: config.DefaultTTL,
		strategy:   config.Strategy,
	}
	go c.cleanupLoop()
	return c
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		c.stats.Misses++
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.Delete(ctx, key)
		c.stats.Misses++
		return nil, false
	}

	c.mu.Lock()
	entry.HitCount++
	c.stats.Hits++
	c.mu.Unlock()

	return entry.Value, true
}

// Set stores a value in cache
func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evict()
	}

	c.entries[key] = &CacheEntry{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}
	c.stats.Sets++

	return nil
}

// SetJSON stores a JSON-serializable value
func (c *Cache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.Set(ctx, key, data, ttl)
}

// GetJSON retrieves and unmarshals a JSON value
func (c *Cache) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	data, exists := c.Get(ctx, key)
	if !exists {
		return false, nil
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal value: %w", err)
	}
	return true, nil
}

// Delete removes a value from cache
func (c *Cache) Delete(_ context.Context, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
	c.stats.Deletes++
}

// Clear removes all entries
func (c *Cache) Clear(_ context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// evict removes the least recently used or oldest entry
func (c *Cache) evict() {
	var oldest *CacheEntry
	var oldestKey string

	for key, entry := range c.entries {
		if oldest == nil || entry.CreatedAt.Before(oldest.CreatedAt) {
			oldest = entry
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.stats.Evictions++
	}
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
			c.stats.Evictions++
		}
	}
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	stats := c.stats
	stats.Size = len(c.entries)
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
	}
	return stats
}

// CacheStats contains cache statistics
type CacheStats struct {
	Size      int     `json:"size"`
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Sets      int64   `json:"sets"`
	Deletes   int64   `json:"deletes"`
	Evictions int64   `json:"evictions"`
	HitRate   float64 `json:"hit_rate_pct"`
}
