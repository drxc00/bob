package internal

import (
	"sync"
)

var (
	globalCache   *Cache[any]
	globalCacheMu sync.Once
)

// GetGlobalCache returns the singleton instance of the cache
func GetGlobalCache() *Cache[any] {
	globalCacheMu.Do(func() {
		globalCache = NewCache[any]()
	})
	return globalCache
}
