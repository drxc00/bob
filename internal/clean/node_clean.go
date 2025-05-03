package clean

import (
	"errors"
	"os"

	"github.com/drxc00/sweepy/internal/cache"
)

func CleanNodeModule(p string) error {
	// Check if node_modules exists
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return errors.New("node_modules does not exist or has been deleted")
	}
	if err != nil {
		return err
	}

	// Additional check to ensure it's a directory
	if !info.IsDir() {
		return errors.New("path is not a directory")
	}

	// Remove node_modules
	// also remove it from the cache
	if err := os.RemoveAll(p); err != nil {
		return err
	}

	// Only load cache after successful removal
	cache := cache.GetGlobalCache()
	ok, c_err := cache.Load()

	if c_err != nil {
		return c_err // Return the cache error, not the previous err
	}

	if !ok {
		return nil // no cache, nothing to do, no error, just return, no need to show error to
	}

	cache.Delete(p)
	return cache.Save() // Return any error from Save directly
}
