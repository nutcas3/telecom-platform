package cache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem represents an item in the cache
type CacheItem struct {
	Key         string    `json:"key"`
	Value       any       `json:"value"`
	Expiration  time.Time `json:"expiration"`
	CreatedAt   time.Time `json:"created_at"`
	AccessCount int64     `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
}

// IsExpired checks if the cache item has expired
func (ci *CacheItem) IsExpired() bool {
	return time.Now().After(ci.Expiration)
}

// Cache interface defines the cache operations
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration)
	Delete(key string)
	Clear()
	Keys() []string
	Size() int
	Stats() CacheStats
}

// CacheStats provides cache statistics
type CacheStats struct {
	Hits       int64   `json:"hits"`
	Misses     int64   `json:"misses"`
	HitRate    float64 `json:"hit_rate"`
	Size       int     `json:"size"`
	Evictions  int64   `json:"evictions"`
	TotalItems int64   `json:"total_items"`
}

// MemoryCache is an in-memory cache implementation
type MemoryCache struct {
	items       map[string]*CacheItem
	mu          sync.RWMutex
	stats       CacheStats
	maxSize     int
	cleanupTick *time.Ticker
	stopCleanup chan bool
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache(maxSize int, cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items:       make(map[string]*CacheItem),
		maxSize:     maxSize,
		cleanupTick: time.NewTicker(cleanupInterval),
		stopCleanup: make(chan bool),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// cleanup removes expired items from cache
func (mc *MemoryCache) cleanup() {
	for {
		select {
		case <-mc.cleanupTick.C:
			mc.removeExpired()
		case <-mc.stopCleanup:
			return
		}
	}
}

// removeExpired removes expired items from cache
func (mc *MemoryCache) removeExpired() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	for key, item := range mc.items {
		if now.After(item.Expiration) {
			delete(mc.items, key)
			mc.stats.Evictions++
		}
	}
}

// Get retrieves an item from cache
func (mc *MemoryCache) Get(key string) (any, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, exists := mc.items[key]
	if !exists {
		mc.stats.Misses++
		return nil, false
	}

	if item.IsExpired() {
		delete(mc.items, key)
		mc.stats.Misses++
		mc.stats.Evictions++
		return nil, false
	}

	// Update access statistics
	item.AccessCount++
	item.LastAccess = time.Now()
	mc.stats.Hits++

	return item.Value, true
}

// Set stores an item in cache
func (mc *MemoryCache) Set(key string, value any, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if we need to evict items to make space
	if len(mc.items) >= mc.maxSize {
		mc.evictLRU()
	}

	expiration := time.Now().Add(ttl)
	item := &CacheItem{
		Key:         key,
		Value:       value,
		Expiration:  expiration,
		CreatedAt:   time.Now(),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}

	mc.items[key] = item
	mc.stats.TotalItems++
}

// evictLRU removes the least recently used item
func (mc *MemoryCache) evictLRU() {
	var lruKey string
	var lruTime time.Time

	for key, item := range mc.items {
		if lruKey == "" || item.LastAccess.Before(lruTime) {
			lruKey = key
			lruTime = item.LastAccess
		}
	}

	if lruKey != "" {
		delete(mc.items, lruKey)
		mc.stats.Evictions++
	}
}

// Delete removes an item from cache
func (mc *MemoryCache) Delete(key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.items, key)
}

// Clear removes all items from cache
func (mc *MemoryCache) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]*CacheItem)
	mc.stats.Evictions += int64(len(mc.items))
}

// Keys returns all keys in cache
func (mc *MemoryCache) Keys() []string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	keys := make([]string, 0, len(mc.items))
	for key := range mc.items {
		keys = append(keys, key)
	}
	return keys
}

// Size returns the current cache size
func (mc *MemoryCache) Size() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return len(mc.items)
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	total := mc.stats.Hits + mc.stats.Misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(mc.stats.Hits) / float64(total) * 100
	}

	return CacheStats{
		Hits:       mc.stats.Hits,
		Misses:     mc.stats.Misses,
		HitRate:    hitRate,
		Size:       len(mc.items),
		Evictions:  mc.stats.Evictions,
		TotalItems: mc.stats.TotalItems,
	}
}

// Close stops the cache cleanup goroutine
func (mc *MemoryCache) Close() {
	mc.cleanupTick.Stop()
	close(mc.stopCleanup)
}

// RedisCache is a Redis-based cache implementation (placeholder)
type RedisCache struct {
	// Redis client would be here
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(addr string, password string, db int) *RedisCache {
	// Implementation would connect to Redis
	return &RedisCache{}
}

// Get retrieves an item from Redis cache
func (rc *RedisCache) Get(key string) (any, bool) {
	// Redis GET implementation
	return nil, false
}

// Set stores an item in Redis cache
func (rc *RedisCache) Set(key string, value any, ttl time.Duration) {
	// Redis SET implementation
}

// Delete removes an item from Redis cache
func (rc *RedisCache) Delete(key string) {
	// Redis DEL implementation
}

// Clear removes all items from Redis cache
func (rc *RedisCache) Clear() {
	// Redis FLUSH implementation
}

// Keys returns all keys in Redis cache
func (rc *RedisCache) Keys() []string {
	// Redis KEYS implementation
	return nil
}

// Size returns the current cache size
func (rc *RedisCache) Size() int {
	// Redis DBSIZE implementation
	return 0
}

// Stats returns cache statistics
func (rc *RedisCache) Stats() CacheStats {
	// Redis INFO implementation
	return CacheStats{}
}

// CacheManager manages multiple cache instances
type CacheManager struct {
	caches map[string]Cache
	mu     sync.RWMutex
}

// NewCacheManager creates a new cache manager
func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches: make(map[string]Cache),
	}
}

// AddCache adds a cache instance
func (cm *CacheManager) AddCache(name string, cache Cache) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.caches[name] = cache
}

// GetCache retrieves a cache instance
func (cm *CacheManager) GetCache(name string) (Cache, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	cache, exists := cm.caches[name]
	return cache, exists
}

// RemoveCache removes a cache instance
func (cm *CacheManager) RemoveCache(name string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.caches, name)
}

// GetAllStats returns statistics for all caches
func (cm *CacheManager) GetAllStats() map[string]CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := make(map[string]CacheStats)
	for name, cache := range cm.caches {
		stats[name] = cache.Stats()
	}
	return stats
}

// ClearAll clears all caches
func (cm *CacheManager) ClearAll() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, cache := range cm.caches {
		cache.Clear()
	}
}

// Global cache manager instance
var globalCacheManager *CacheManager

// InitializeCacheManager initializes the global cache manager
func InitializeCacheManager() {
	globalCacheManager = NewCacheManager()

	// Add default memory cache
	memoryCache := NewMemoryCache(1000, 5*time.Minute)
	globalCacheManager.AddCache("default", memoryCache)

	// Add specialized caches
	userCache := NewMemoryCache(500, 10*time.Minute)
	globalCacheManager.AddCache("users", userCache)

	serviceCache := NewMemoryCache(200, 2*time.Minute)
	globalCacheManager.AddCache("services", serviceCache)

	metricsCache := NewMemoryCache(1000, 1*time.Minute)
	globalCacheManager.AddCache("metrics", metricsCache)
}

// GetCacheManager returns the global cache manager
func GetCacheManager() *CacheManager {
	return globalCacheManager
}

// GetDefaultCache returns the default cache
func GetDefaultCache() Cache {
	if globalCacheManager == nil {
		InitializeCacheManager()
	}
	cache, _ := globalCacheManager.GetCache("default")
	return cache
}

// GetUserCache returns the user cache
func GetUserCache() Cache {
	if globalCacheManager == nil {
		InitializeCacheManager()
	}
	cache, _ := globalCacheManager.GetCache("users")
	return cache
}

// GetServiceCache returns the service cache
func GetServiceCache() Cache {
	if globalCacheManager == nil {
		InitializeCacheManager()
	}
	cache, _ := globalCacheManager.GetCache("services")
	return cache
}

// GetMetricsCache returns the metrics cache
func GetMetricsCache() Cache {
	if globalCacheManager == nil {
		InitializeCacheManager()
	}
	cache, _ := globalCacheManager.GetCache("metrics")
	return cache
}

// CacheKey generates a cache key with namespace
func CacheKey(namespace, key string) string {
	return fmt.Sprintf("%s:%s", namespace, key)
}

// SerializeValue serializes a value for caching
func SerializeValue(value any) ([]byte, error) {
	return json.Marshal(value)
}

// DeserializeValue deserializes a value from cache
func DeserializeValue(data []byte, target any) error {
	return json.Unmarshal(data, target)
}
