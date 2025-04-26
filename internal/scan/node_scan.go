package scan

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ScannedNodeModule struct {
	Path         string
	Staleness    int64 // In days
	Size         int64
	LastModified time.Time
}

func NodeScan(path string, staleness int64) ([]ScannedNodeModule, error) {
	var scannedNodeModules []ScannedNodeModule = []ScannedNodeModule{}

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
			return err
		}

		// If the path is a directory and it's named "node_modules"
		if info.IsDir() && info.Name() == "node_modules" {
			wg.Add(1)

			go func(nodeModulePath string) {
				defer wg.Done()

				// Get the last modified and accessed times of the directory containing the node_modules directory
				// We do this so that we can know if the project has been updated since the last time we scanned it
				// If we based it on the node_modules folder alone, it will not be accurate if the project does not have any new dependencies
				parentDirectory := filepath.Dir(nodeModulePath)
				parentDirectoryInfo, err := os.Stat(parentDirectory)
				if err != nil {
					// TODO: Handle error
				}

				// Calculate the staleness of the node_modules directory
				parentDirAge := currentTime.Sub(parentDirectoryInfo.ModTime()).Hours() / 24
				nodeModuleStaleness := int64(parentDirAge)

				if staleness != 0 && nodeModuleStaleness < staleness {
					// We skip the node_modules directory if the staleness is less than the specified staleness
					return
				}

				// Get the size of the node_modules directory
				dirSize, err := DirSize(nodeModulePath)
				if err != nil {
					// TODO: Handle error
				}

				// Create and populate a ScannedNodeModule struct
				scannedNodeModule := ScannedNodeModule{
					Path:         nodeModulePath,
					Size:         dirSize,
					LastModified: parentDirectoryInfo.ModTime(),
					Staleness:    nodeModuleStaleness,
				}

				// Make sure that other goroutines don't modify the slice at the same time
				mutex.Lock()
				scannedNodeModules = append(scannedNodeModules, scannedNodeModule)
				mutex.Unlock()
			}(path)

			// If a node_modules directory is found, stop walking the directory tree
			return filepath.SkipDir
		}

		return nil
	})

	// Wait for all goroutines to finish
	wg.Wait()

	if err != nil {
		log.Fatal(err)
		return []ScannedNodeModule{}, err
	}

	return scannedNodeModules, nil
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return size, nil
}
