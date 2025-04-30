package scan

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/drxc00/sweepy/internal"
	"github.com/drxc00/sweepy/types"
	"github.com/drxc00/sweepy/utils"
)

type ScannedNodeModule struct {
	Path         string
	Staleness    int64 // In days
	Size         int64
	LastModified time.Time
}

type ScanInfo struct {
	TotalSize    int64
	AvgStaleness float64
}

func sortScannedNodeModules(modules *[]ScannedNodeModule) {
	sort.Slice(*modules, func(i, j int) bool {
		return (*modules)[i].Staleness > (*modules)[j].Staleness
	})
}

func NodeScan(ctx types.ScanContext, ch chan<- string) ([]ScannedNodeModule, ScanInfo, error) {
	// We apply Mutual Exclusion to the goroutines to prevent race conditions
	var mutex sync.Mutex  // Mutex for concurrent access to scannedNodeModules
	var wg sync.WaitGroup // Wait group for parallel scanning
	var scannedNodeModules []ScannedNodeModule = []ScannedNodeModule{}
	var totalSize int64 = 0
	var totalStaleness float64 = 0

	// Cache handler
	cache := internal.NewCache[ScannedNodeModule]()
	cacheLoaded := false

	// Time for calculating staleness
	currentTime := time.Now()

	if !ctx.NoCache {
		ok, loadErr := cache.Load()
		if ok && loadErr == nil {
			cacheLoaded = true
		}
	}

	if cacheLoaded && !cache.IsExpired() && !ctx.NoCache && !ctx.ResetCache {
		// Get all cached entries
		// cachedEntries := cache.GetAll()

		// Filter cached entries based on staleness criteria
		for p, module := range cache.Data {
			// Check if the path contains the current path
			if !strings.Contains(p, ctx.Path) {
				continue
			}

			// Send to channel
			ch <- fmt.Sprintf("Found %s in cache", module.Path)

			// Stats handler
			mutex.Lock()
			totalSize += module.Size
			totalStaleness += float64(module.Staleness)
			mutex.Unlock()

			daysSinceModified := int64(currentTime.Sub(module.LastModified).Hours() / 24)

			// If the module meets our staleness criteria, add it directly without scanning
			if ctx.Staleness == 0 || daysSinceModified >= ctx.Staleness {
				mutex.Lock()
				scannedNodeModules = append(scannedNodeModules, module)
				mutex.Unlock()
			}
		}

		// If we have entries from cache and don't need a full rescan, return early
		if len(scannedNodeModules) > 0 && !ctx.NoCache {
			/*
				We could add a flag here to decide if we want to skip the scan completely
				For now, we'll continue to scan for any new directories not in cache
				TODO implement if needed
			*/
		}
	}

	// Set of paths we've already processed from cache
	processedPaths := make(map[string]bool)
	for _, module := range scannedNodeModules {
		processedPaths[module.Path] = true
	}

	err := filepath.WalkDir(ctx.Path, func(p string, d fs.DirEntry, err error) error {
		// Check if the walk function encountered an error
		if err != nil {
			// Check if the error is a permission error
			if errors.Is(err, fs.ErrPermission) {
				// We don't want to stop the walk function if we encounter a permission error
				// log.Print(err)
				if ctx.Verbose {
					ch <- fmt.Sprintf("Permission denied: %v", err)
				}
				return filepath.SkipDir
			}
			return err
		}

		// Check if the current path is in the cache
		if _, ok := cache.Get(p); ok && !cache.IsExpired() && !ctx.NoCache && !ctx.ResetCache {
			// If the path is in the cache, add it to the slice of scannedNodeModules
			c, ok := cache.Get(p)
			if !ok {
				// If the path is in the cache but the data is not found, something went wrong
				utils.Log("Error when scanning: %v\n", err)
				// so we want to continue scanning without caching
			} else {
				if ctx.Verbose {
					ch <- fmt.Sprintf("Found %s in cache", c.Path)
				}
				mutex.Lock()
				scannedNodeModules = append(scannedNodeModules, c)
				mutex.Unlock()
				return filepath.SkipDir // Short circuit the walk function
			}
		}

		// If the path is a directory and it's named "node_modules"
		if d.IsDir() && d.Name() == "node_modules" {

			// Immediate return if the path has already been processed
			if processedPaths[p] {
				return filepath.SkipDir
			}

			if ctx.Verbose {
				ch <- fmt.Sprintf("Scanning %s", p)
			}

			wg.Add(1) // Add to the wait group

			go func(nodeModulePath string) {
				defer wg.Done()

				/*
					Get the last modified and accessed times of the directory containing the node_modules directory
					We do this so that we can know if the project has been updated since the last time we scanned it
					If we based it on the node_modules folder alone, it will not be accurate.
				*/
				parentDir := filepath.Dir(nodeModulePath)
				parentDirInfo, err := os.Stat(parentDir)

				if err != nil {
					utils.Log("Error when scanning: %v\n", err)
					if ctx.Verbose {
						ch <- fmt.Sprintf("Error when scanning: %v\n", err)
					}
					return
				}

				// Calculate the staleness of the `node_modules` directory
				// Calculate staleness in days
				lastModified, lerr := GetLastModified(parentDir)

				if lerr != nil {
					utils.Log("Error when scanning: %v\n", lerr)
					if ctx.Verbose {
						ch <- fmt.Sprintf("Error when scanning: %v\n", lerr)
					}
					return
				}

				daysSinceModified := int64(currentTime.Sub(lastModified).Hours() / 24)

				if ctx.Staleness != 0 && daysSinceModified < ctx.Staleness {
					// We skip the node_modules directory if the staleness is less than the specified staleness
					return
				}

				// Get the size of the node_modules directory
				dirSize, err := DirSize(nodeModulePath)
				if err != nil {
					utils.Log("Error when scanning: %v\n", err)
					return
				}

				// Add for stats
				mutex.Lock()
				totalSize += dirSize
				totalStaleness += float64(daysSinceModified)
				mutex.Unlock()

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
				cache.Set(nodeModulePath, scannedNodeModule) // Save to cache handler
				mutex.Unlock()
			}(p)

			// If a node_modules directory is found, stop walking the directory tree
			return filepath.SkipDir
		}

		return nil
	})

	// Wait for all goroutines to finish
	// If this is not added, the program will simply exit without any output
	wg.Wait()

	// Close the channel
	close(ch)

	if err != nil {
		utils.Log("Error when scanning: %v\n", err)
		return []ScannedNodeModule{}, ScanInfo{}, err
	}

	// Sort the scannedNodeModules by staleness
	sortScannedNodeModules(&scannedNodeModules)

	/*
		We only save the cache if we are not using the --no-cache flag
		Or if we are using the --reset-cache flag.
		This will override the cache and save the new data to the cache.
	*/
	if !ctx.NoCache || ctx.ResetCache {
		saveErr := cache.Save()
		if saveErr != nil {
			utils.Log("Error when scanning: %v\n", saveErr)
		}
	}

	// Prevent division by zero
	var avgStaleness float64 = 0
	if len(scannedNodeModules) == 0 {
		avgStaleness = 0
	} else {
		avgStaleness = totalStaleness / float64(len(scannedNodeModules))
	}

	return scannedNodeModules, ScanInfo{TotalSize: totalSize, AvgStaleness: avgStaleness}, nil
}

func DirSize(path string) (int64, error) {
	var totalSize int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip problematic files or directories
			return nil
		}
		if !d.IsDir() {
			// Use Info() only if necessary (costs a syscall)
			info, err := d.Info()
			if err != nil {
				// If we can't get file info, skip it
				return nil
			}
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalSize, nil
}

func GetLastModified(p string) (time.Time, error) {
	/*
		Accepts a directory path `p` as input.
		This directory path is assumed as the parent directory of the node_modules directory.
	*/

	var lastModified time.Time
	err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// Check the modification time of the file
			if inf, err := d.Info(); err == nil {
				if inf.ModTime().After(lastModified) {
					lastModified = inf.ModTime()
				}
			}
		}
		return nil // Continue walking
	})

	if err != nil {
		panic(err)
	}

	return lastModified, nil
}
