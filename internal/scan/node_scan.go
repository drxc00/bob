package scan

import (
	"os"
	"path/filepath"
)

// FindNodeModules tries to find the first node_modules folder in a given path.
// This function will recursively search for node_modules folders in the given path.
// If a node_modules folder is found, it will return the path to that folder.
// If no node_modules folder is found, it will return an empty string.
func FindNodeModules(path string) ([]string, error) {
	var nodeModulesPaths []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		// Check if the walk function encountered an error
		if err != nil {
			return err
		}

		// If the path is a directory and it's named "node_modules"
		if info.IsDir() && info.Name() == "node_modules" {
			nodeModulesPaths = append(nodeModulesPaths, path)
			return filepath.SkipDir // Stop traversing this directory and its subdirectories
		}

		return nil
	})

	if err != nil {
		return []string{}, err
	}

	return nodeModulesPaths, nil
}
