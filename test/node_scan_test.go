package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/drxc00/sweepy/internal/scan"
	"github.com/drxc00/sweepy/types"
)

func setupTestDirectory(t *testing.T) (string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sweepy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test project structure
	projectPaths := []string{
		filepath.Join(tempDir, "project1", "node_modules"),
		filepath.Join(tempDir, "project2", "node_modules"),
		filepath.Join(tempDir, "project3", "subfolder", "node_modules"),
	}

	for _, path := range projectPaths {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Create some dummy files in node_modules
		dummyFile := filepath.Join(path, "dummy.js")
		if err := os.WriteFile(dummyFile, []byte("dummy content"), 0644); err != nil {
			t.Fatalf("Failed to create dummy file: %v", err)
		}
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestNodeScan(t *testing.T) {
	// Setup test directory
	testDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	tests := []struct {
		name          string
		ctx           types.ScanContext
		expectedCount int
		expectError   bool
	}{
		{
			name: "Basic scan",
			ctx: types.ScanContext{
				Path:      testDir,
				Staleness: 0,
				NoCache:   true,
			},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name: "Scan with staleness",
			ctx: types.ScanContext{
				Path:      testDir,
				Staleness: 365, // Set high staleness to exclude all test directories
				NoCache:   true,
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create channel for progress updates
			ch := make(chan string)

			// Run scan in goroutine
			go func() {
				for range ch {
					// Consume progress messages
				}
			}()

			modules, info, err := scan.NodeScan(tt.ctx, ch)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(modules) != tt.expectedCount {
				t.Errorf("Expected %d modules, got %d", tt.expectedCount, len(modules))
			}
			if info.TotalSize <= 0 && tt.expectedCount > 0 {
				t.Error("Expected non-zero total size")
			}
		})
	}
}

func TestDirSize(t *testing.T) {
	// Setup test directory
	testDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	// Create a test file with known size
	testFile := filepath.Join(testDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	size, err := scan.DirSize(testDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedSize := int64(len(testContent))
	if size < expectedSize {
		t.Errorf("Expected size >= %d, got %d", expectedSize, size)
	}
}

func TestGetLastModified(t *testing.T) {
	// Setup test directory
	testDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	// Create a new file with current timestamp
	testFile := filepath.Join(testDir, "recent.txt")
	if err := os.WriteFile(testFile, []byte("recent"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get last modified time
	lastMod, err := scan.GetLastModified(testDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if last modified time is recent
	if time.Since(lastMod) > time.Minute {
		t.Error("Last modified time is older than expected")
	}
}
