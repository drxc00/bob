package scan

import (
	"errors"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/charlievieth/fastwalk"
)

func DirSizeFastWalk(path string) (int64, error) {
	var totalSize int64

	err := fastwalk.Walk(&fastwalk.DefaultConfig, path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip permission errors silently
			if errors.Is(err, fs.ErrPermission) {
				return fastwalk.SkipDir
			}
			// Skip other errors but continue walking
			return nil
		}

		if d != nil && !d.IsDir() {
			// Get file info for all non-directory entries
			info, err := d.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
		return nil
	})

	return totalSize, err
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
			if err == nil {
				totalSize += info.Size()
			}
		}
		return nil
	})
	return totalSize, err
}

func GetLastModified(p string) (time.Time, error) {
	// Accepts a directory path `p` as input.
	// This directory path is assumed as the parent directory of the node_modules directory.
	var lastModified time.Time
	err := fastwalk.Walk(&fastwalk.DefaultConfig, p, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip problematic files or directories
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

	return lastModified, err
}
