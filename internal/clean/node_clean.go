package clean

import (
	"errors"
	"os"

	"github.com/drxc00/sweepy/internal"
)

func CleanNodeModule(p string) error {
	// Check if node_modules exists
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return errors.New("node_modules does not exist or has been deleted.")
	}

	// Remove node_modules
	// also remove it from the cache
	// Comment for testing
	err := os.RemoveAll(p)
	if err != nil {
		return err
	}

	// Remove node_modules from the cache
	cache := internal.GetGlobalCache()
	ok, c_err := cache.Load()

	if c_err != nil {
		return err
	}

	if !ok {
		return nil // no cache, nothing to do, no error, just return, no need to show error to
	}

	cache.Delete(p)
	cache.Save()

	return nil
}
