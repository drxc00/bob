package scan

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charlievieth/fastwalk"
	"github.com/drxc00/sweepy/internal/cache"
	"github.com/drxc00/sweepy/types"
	"github.com/drxc00/sweepy/utils"
)

func NodeScan(ctx types.ScanContext, ch chan<- string) ([]types.ScannedNodeModule, types.ScanInfo, error) {
	// We apply Mutual Exclusion to the goroutines to prevent race conditions
	var mutex sync.Mutex  // Mutex for concurrent access to scannedNodeModules
	var wg sync.WaitGroup // Wait group for parallel scanning
	var scannedNodeModules []types.ScannedNodeModule = []types.ScannedNodeModule{}
	var totalSize int64 = 0
	var totalStaleness float64 = 0

	// Scan Time
	startTime := time.Now()

	// Cache handler
	cache := cache.GetGlobalCache()
	cacheLoaded := false

	if !ctx.NoCache {
		ok, loadErr := cache.Load()
		if ok {
			cacheLoaded = true
		} else {
			ch <- fmt.Sprintf("Error loading cache: %v", loadErr)
		}
	}

	if cacheLoaded && !cache.IsExpired() && !ctx.NoCache && !ctx.ResetCache {

		// Filter cached entries based on staleness criteria
		for p, module := range cache.Data {
			// Check if the path contains the current path
			if !strings.Contains(p, ctx.Path) {
				continue
			}

			if ctx.Staleness != 0 && module.Staleness < ctx.Staleness {
				continue
			}

			// Send to channel
			ch <- fmt.Sprintf("Found %s in cache", module.Path)

			// Add the module to the slice of scannedNodeModules
			mutex.Lock()
			totalSize += module.Size
			totalStaleness += float64(module.Staleness)
			scannedNodeModules = append(scannedNodeModules, module)
			mutex.Unlock()
		}

		// If we have entries from cache and don't need a full rescan, return early
		if len(scannedNodeModules) > 0 && !ctx.NoCache {
			scanDuration := time.Since(startTime)
			return scannedNodeModules, types.ScanInfo{TotalSize: totalSize, AvgStaleness: totalStaleness, ScanDuration: scanDuration}, nil
		}
	}

	// Fastwalk is a faster alternative to filepath.Walk
	// Wraps our walk function to ignore permission errors
	walkFn := fastwalk.IgnorePermissionErrors(func(p string, d fs.DirEntry, err error) error {
		// Check if the walk function encountered an error
		if err != nil {
			// For other errors, log but continue walking
			if ctx.Verbose {
				ch <- fmt.Sprintf("Error accessing path %s: %v", p, err)
			}
			return fastwalk.SkipDir
		}

		if d == nil {
			if ctx.Verbose {
				ch <- fmt.Sprintf("Skipping nil directory entry at %s", p)
			}
			return nil
		}

		// If the path is a directory and it's named "node_modules"
		if d.IsDir() && d.Name() == "node_modules" {

			if ctx.Verbose {
				ch <- fmt.Sprintf("Scanning %s", p)
			}

			wg.Add(1) // Add to the wait group

			go func(nodeModulePath string) {
				defer wg.Done()

				// Get the last modified and accessed times of the directory containing the node_modules directory
				// We do this so that we can know if the project has been updated since the last time we scanned it
				// If we based it on the node_modules folder alone, it will not be accurate.
				parentDir := filepath.Dir(nodeModulePath)
				lastModified, lerr := GetLastModified(parentDir)

				if lerr != nil {
					utils.Log("Error when determining last modified: %v\n", lerr)
					if ctx.Verbose {
						ch <- fmt.Sprintf("Error when determining last modified: %v\n", lerr)
					}

				}

				daysSinceModified := int64(startTime.Sub(lastModified).Hours() / 24)

				if ctx.Staleness != 0 && daysSinceModified < ctx.Staleness {
					// We skip the node_modules directory if the staleness is less than the specified staleness
					return
				}

				// Get the size of the node_modules directory
				dirSize, err := DirSizeFastWalk(nodeModulePath)
				if err != nil {
					ch <- fmt.Sprintf("Error when calculating dir size: %v\n", err)
					return
				}

				// Add for stats
				// Make sure that other goroutines don't modify the slice at the same time
				mutex.Lock()
				totalSize += dirSize
				totalStaleness += float64(daysSinceModified)
				// Create and populate a ScannedNodeModule struct
				scannedNodeModule := types.ScannedNodeModule{
					Path:         nodeModulePath,
					Size:         dirSize,
					LastModified: lastModified,
					Staleness:    daysSinceModified,
				}
				scannedNodeModules = append(scannedNodeModules, scannedNodeModule)
				cache.Set(nodeModulePath, scannedNodeModule) // Save to cache handler
				mutex.Unlock()
			}(p)

			// If a node_modules directory is found, stop walking the directory tree
			return fastwalk.SkipDir
		}

		return nil
	})

	err := fastwalk.Walk(&fastwalk.DefaultConfig, ctx.Path, walkFn)

	// Wait for all goroutines to finish
	// If this is not added, the program will simply exit without any output
	wg.Wait()

	// Close the channel
	close(ch)

	// Calculate the scan duration
	scanDuration := time.Since(startTime)

	if err != nil {
		utils.Log("Error after scanning: %v\n", err)
		log.Print(err)
		return []types.ScannedNodeModule{}, types.ScanInfo{}, err
	}

	// We only save the cache if we are not using the --no-cache flag
	// Or if we are using the --reset-cache flag.
	// This will override the cache and save the new data to the cache.
	if !ctx.NoCache || ctx.ResetCache || cache.IsExpired() {
		if cache.IsExpired() {
			cacheValidity := startTime.Add(time.Hour * 24).Unix()
			cache.SetValidity(cacheValidity)
		}
		saveErr := cache.Save()
		if saveErr != nil {
			utils.Log("Error saving cache: %v\n", saveErr)
		}
	}

	// Prevent division by zero
	var avgStaleness float64 = 0
	if len(scannedNodeModules) == 0 {
		avgStaleness = 0
	} else {
		avgStaleness = totalStaleness / float64(len(scannedNodeModules))
	}

	return scannedNodeModules, types.ScanInfo{TotalSize: totalSize, AvgStaleness: avgStaleness, ScanDuration: scanDuration}, nil
}
