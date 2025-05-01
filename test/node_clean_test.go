package test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/drxc00/sweepy/internal"
	"github.com/drxc00/sweepy/types"
	"github.com/drxc00/sweepy/utils"

	"github.com/drxc00/sweepy/internal/clean"
	"github.com/drxc00/sweepy/internal/scan"
)

func TestNodeClean(t *testing.T) {

	tests := []struct {
		name          string
		targetModules []string
		expectError   bool
		expectedCount int
	}{
		{
			name: "Clean single module",
			targetModules: []string{
				"project1/node_modules",
			},
			expectError:   false,
			expectedCount: 2, // Should have 2 remaining modules
		},
		{
			name: "Clean multiple modules",
			targetModules: []string{
				"project1/node_modules",
				"project2/node_modules",
			},
			expectError:   false,
			expectedCount: 1, // Should have 1 remaining module
		},
		{
			name: "Clean non-existent module",
			targetModules: []string{
				"non-existent/node_modules",
			},
			expectError:   true,
			expectedCount: 3, // Should still have all modules
		},
		{
			name: "Clean nested module",
			targetModules: []string{
				"project3/subfolder/node_modules",
			},
			expectError:   false,
			expectedCount: 2, // Should have 2 remaining modules
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testDir, projectPaths, cleanup := utils.SetupTestDirectory(t)
			defer cleanup()

			// Create a new cache
			cache := internal.GetGlobalCache()
			// Clear cache
			cache.Clear()

			// Save paths
			for _, p := range projectPaths {
				cache.Set(p, types.ScannedNodeModule{
					Path:         p,
					Size:         100,
					LastModified: time.Now(),
					Staleness:    0,
				})
			}

			// Save cache
			err := cache.Save()
			if err != nil {
				t.Fatalf("Failed to save cache: %v", err)
			}

			t_ctx := types.NewScanContext(testDir, "0", true, false, false)

			// Convert relative paths to absolute
			var absolutePaths []string
			for _, path := range tt.targetModules {
				absolutePaths = append(absolutePaths, filepath.Join(testDir, path))
			}

			// Clean the modules
			for _, ap := range absolutePaths {
				c_err := clean.CleanNodeModule(ap)
				if c_err != nil && !tt.expectError {
					t.Fatal("Did not expect error but got one")
				} else if c_err == nil && tt.expectError {
					t.Fatal("Expected error but got none", c_err)
				}
			}

			// Test ch
			ch := make(chan string, 100)
			// Verify remaining modules
			modules, _, err := scan.NodeScan(t_ctx, ch)
			// Load cache
			_, err = cache.Load()
			if err != nil {
				t.Fatalf("Failed to load cache: %v", err)
			}

			if err != nil {
				t.Fatalf("Failed to scan modules after cleaning: %v", err)
			}

			if len(modules) != tt.expectedCount {
				t.Errorf("Expected %d remaining modules, got %d", tt.expectedCount, len(modules))
			}

			if len(cache.Data) != tt.expectedCount {
				t.Errorf("Expected %d remaining modules in cache, got %d", tt.expectedCount, len(cache.Data))
			}
		})
	}
}
