package cache

import (
	"sync"

	"github.com/drxc00/sweepy/types"
)

var (
	globalCache   *Cache[types.ScannedNodeModule]
	globalCacheMu sync.Once
)

// GetGlobalCache returns the singleton instance of the cache
func GetGlobalCache() *Cache[types.ScannedNodeModule] {
	globalCacheMu.Do(func() {
		globalCache = NewCache[types.ScannedNodeModule]()
	})
	return globalCache
}
