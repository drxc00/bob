package scan

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/drxc00/bob/utils"
)

type ScannedNodeModule struct {
	Path         string
	Staleness    int64 // In days
	Size         int64
	LastModified time.Time
}

func NodeScan(path string, staleness int64) ([]ScannedNodeModule, error) {
	var scannedNodeModules []ScannedNodeModule = []ScannedNodeModule{}
	cache := utils.NewCache[ScannedNodeModule]("node_modules")
	ok, loadErr := cache.Load()

	if !ok && loadErr != nil {
		log.Fatal(loadErr)
		// Still continue scanning without caching
	}

	// We apply Mutual Exclusion to the goroutines to prevent race conditions
	// Since we want to append to the slice of scannedNodeModules, we need to make sure that
	// other goroutines don't modify the slice at the same time
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Current Time for calculating staleness
	currentTime := time.Now()

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		// Check if the walk function encountered an error
		if err != nil {
			// Check if the error is a permission error
			if errors.Is(err, fs.ErrPermission) {
				// We don't want to stop the walk function if we encounter a permission error
				// We simply want to skip the directory
				log.Printf("Skipping directory %s", err)
				return filepath.SkipDir
			}
			return err
		}

		// Check if the current path is in the cache
		if _, ok := cache.Get(path); ok && !cache.IsExpired() {
			// If the path is in the cache, add it to the slice of scannedNodeModules
			c, ok := cache.Get(path)
			if !ok {
				// If the path is in the cache but the data is not found, something went wrong
				log.Fatal(err)
				// so we want to continue scanning without caching
			} else {
				scannedNodeModules = append(scannedNodeModules, c)
				return filepath.SkipDir // Short circuit the walk function
			}
		}

		// If the path is a directory and it's named "node_modules"
		if info.IsDir() && info.Name() == "node_modules" {
			wg.Add(1)

			go func(nodeModulePath string) {
				defer wg.Done()

				// Get the last modified and accessed times of the directory containing the node_modules directory
				// We do this so that we can know if the project has been updated since the last time we scanned it
				// If we based it on the node_modules folder alone, it will not be accurate if the project does not have any new dependencies
				parentDir := filepath.Dir(nodeModulePath)
				parentDirInfo, err := os.Stat(parentDir)

				if err != nil {
					log.Fatal(err)
					return
				}

				// Calculate the staleness of the node_modules directory
				// Calculate staleness in days
				daysSinceModified := int64(currentTime.Sub(parentDirInfo.ModTime()).Hours() / 24)

				if staleness != 0 && daysSinceModified < staleness {
					// We skip the node_modules directory if the staleness is less than the specified staleness
					return
				}

				// Get the size of the node_modules directory
				dirSize, err := DirSize(nodeModulePath)
				if err != nil {
					log.Fatal(err)
					return
				}

				// Create and populate a ScannedNodeModule struct
				scannedNodeModule := ScannedNodeModule{
					Path:         nodeModulePath,
					Size:         dirSize,
					LastModified: parentDirInfo.ModTime(),
					Staleness:    daysSinceModified,
				}

				// Make sure that other goroutines don't modify the slice at the same time
				mutex.Lock()
				scannedNodeModules = append(scannedNodeModules, scannedNodeModule)
				cache.Set(nodeModulePath, scannedNodeModule)
				mutex.Unlock()
			}(path)

			// If a node_modules directory is found, stop walking the directory tree
			return filepath.SkipDir
		}

		return nil
	})

	// Wait for all goroutines to finish
	// If this is not added, the program will simply exit without any output
	wg.Wait()

	// Save the cache
	saveErr := cache.Save()

	if saveErr != nil {
		log.Fatal(saveErr)
		return []ScannedNodeModule{}, saveErr
	}

	if err != nil {
		log.Fatal(err)
		return []ScannedNodeModule{}, err
	}

	return scannedNodeModules, nil
}

func DirSize(path string) (int64, error) {
	var totalSize int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalSize, nil
}
