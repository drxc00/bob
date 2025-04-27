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
	"time"
)

type CachedData struct {
	Path string `json:"path"`
	Data any    `json:"data"`
}

type Cache[T any] struct {
	Type     string       `json:"type"`     // Type of data, e.g., "node_modules", "git"
	Validity int64        `json:"validity"` // Cache expiration timestamp (Unix time)
	Data     map[string]T `json:"data"`     // Map to hold the cached data, key is the identifier (e.g., path)
}

func NewCache[T any](t string) *Cache[T] {
	return &Cache[T]{
		Type:     t,
		Validity: time.Now().Add(time.Hour * 24).Unix(), // Expire after 24 hours by default
		Data:     make(map[string]T),
	}
}

func (c *Cache[T]) SetValidity(validity int64) {
	c.Validity = validity
}

func (c *Cache[T]) GetAll() []T {
	var data []T
	for _, value := range c.Data {
		data = append(data, value)
	}

	return data
}

func (c *Cache[T]) Get(identifier string) (T, bool) {
	var data T
	if _, ok := c.Data[identifier]; ok {
		data = c.Data[identifier]
	}
	return data, false
}

func (c *Cache[T]) Set(identifier string, data T) {
	c.Data[identifier] = data
}

func (c *Cache[T]) Delete(identifier string) {
	delete(c.Data, identifier)
}

func (c *Cache[T]) IsExpired() bool {
	return time.Now().Unix() > c.Validity
}

func (c *Cache[T]) Save() error {
	// Save the cache data to the JSON file
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	filename := getFileName(c.Type)

	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		// 0644 is the default file permissions for a new file
		if err := os.WriteFile(filename, b, 0644); err != nil {
			return err
		}
	}

	return os.WriteFile(filename, b, 0644)
}

func (c *Cache[T]) Load() (bool, error) {
	filename := getFileName(c.Type)

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

func getFileName(t string) string {
	return "/.bob-" + t + "-cache.json"
}
