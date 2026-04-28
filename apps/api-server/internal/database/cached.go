package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/cache"
)

// CachedDatabase wraps a database with caching capabilities
type CachedDatabase struct {
	*Database
	cache cache.Cache
}

// NewCachedDatabase creates a new cached database wrapper
func NewCachedDatabase(db *Database) *CachedDatabase {
	return &CachedDatabase{
		Database: db,
		cache:    cache.GetDefaultCache(),
	}
}

// CachedFind performs a cached database find operation
func (cd *CachedDatabase) CachedFind(dest any, query string, args ...any) error {
	cacheKey := cache.CacheKey("find", fmt.Sprintf("%s:%v", query, args))

	// Try to get from cache first
	if cached, found := cd.cache.Get(cacheKey); found {
		if data, ok := cached.([]byte); ok {
			return json.Unmarshal(data, dest)
		}
	}

	// Cache miss - query database
	if err := cd.DB.Find(dest, args...).Error; err != nil {
		return err
	}

	// Cache the result
	if data, err := json.Marshal(dest); err == nil {
		cd.cache.Set(cacheKey, data, 5*time.Minute)
	}

	return nil
}

// CachedFirst performs a cached database first operation
func (cd *CachedDatabase) CachedFirst(dest any, query string, args ...any) error {
	cacheKey := cache.CacheKey("first", fmt.Sprintf("%s:%v", query, args))

	if cached, found := cd.cache.Get(cacheKey); found {
		if data, ok := cached.([]byte); ok {
			return json.Unmarshal(data, dest)
		}
	}

	if err := cd.DB.First(dest, args...).Error; err != nil {
		return err
	}

	if data, err := json.Marshal(dest); err == nil {
		cd.cache.Set(cacheKey, data, 5*time.Minute)
	}

	return nil
}

// InvalidateCache invalidates cache entries
func (cd *CachedDatabase) InvalidateCache(pattern string) {
	cd.cache.Clear()
}

// InvalidateCacheKey invalidates a specific cache key
func (cd *CachedDatabase) InvalidateCacheKey(key string) {
	cd.cache.Delete(key)
}

// CacheStats returns cache statistics
func (cd *CachedDatabase) CacheStats() cache.CacheStats {
	return cd.cache.Stats()
}
