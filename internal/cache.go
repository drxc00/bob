/*
	This package holds the mechanism for caching the results of the scans.
	The cache is stored in specific file names, .bob-node-modules-cache.json and .bob-git-cache.json
	The file names are hardcoded and cannot be changed.
	The cache is stored in the user's home directory.
*/

package internal

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Cache[T any] struct {
	Validity int64        `json:"validity"` // Cache expiration timestamp (Unix time)
	Data     map[string]T `json:"data"`     // Map to hold the cached data, key is the identifier (e.g., path)
	mu       sync.RWMutex // Mutex to protect concurrent access
}

func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		Validity: time.Now().Add(time.Hour * 24).Unix(), // Expire after 24 hours by default
		Data:     make(map[string]T),
		mu:       sync.RWMutex{},
	}
}

func (c *Cache[T]) SetValidity(validity int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Validity = validity
}

func (c *Cache[T]) GetAll() map[string]T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Data
}

func (c *Cache[T]) Get(identifier string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var data T
	if _, ok := c.Data[identifier]; ok {
		data = c.Data[identifier]
	}
	return data, false
}

func (c *Cache[T]) Set(identifier string, data T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Data[identifier] = data
}

func (c *Cache[T]) Delete(identifier string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.Data, identifier)
}

func (c *Cache[T]) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return time.Now().Unix() > c.Validity
}

func (c *Cache[T]) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Save the cache data to the JSON file
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	filename := "sweepy.cache.json"

	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		// 0644 is the default file permissions for a new file
		if err := os.WriteFile(filename, b, 0644); err != nil {
			return err
		}
		return nil
	}

	return os.WriteFile(filename, b, 0644)
}

func (c *Cache[T]) Load() (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	filename := "sweepy.cache.json"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, err // File doesn't exist, nothing to load
	}

	// Read the cache file
	data, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}

	// Unmarshal the data into the cache structure
	err = json.Unmarshal(data, c)
	if err != nil {
		return false, err
	}

	return true, nil
}
