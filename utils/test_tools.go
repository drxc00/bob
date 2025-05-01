package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func SetupTestDirectory(t *testing.T) (string, []string, func()) {
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

	return tempDir, projectPaths, cleanup
}
