package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/onozaty/treecat/internal/filter"
)

// FileEntry represents a file or directory entry.
type FileEntry struct {
	Path    string // Absolute path
	RelPath string // Relative to scan root
	IsDir   bool   // Whether it's a directory
	Size    int64  // File size in bytes
}

// Scanner scans a directory and collects files.
type Scanner struct {
	Root   string         // Root directory (absolute path)
	Filter filter.Filter  // Filter to apply when scanning
}

// NewScanner creates a new Scanner.
func NewScanner(root string, filter filter.Filter) (*Scanner, error) {
	// Convert to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", absRoot)
	}

	return &Scanner{
		Root:   absRoot,
		Filter: filter,
	}, nil
}

// Scan walks the directory and returns a list of files.
func (s *Scanner) Scan() ([]FileEntry, error) {
	var entries []FileEntry

	err := filepath.WalkDir(s.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Return error immediately (don't skip)
			return fmt.Errorf("cannot access %s: %w", path, err)
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("cannot get info for %s: %w", path, err)
		}

		isDir := d.IsDir()

		// Apply filter
		if s.Filter != nil && !s.Filter.ShouldInclude(path, isDir) {
			if isDir {
				// Skip this directory entirely
				return filepath.SkipDir
			}
			// Skip this file
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.Root, path)
		if err != nil {
			return fmt.Errorf("cannot get relative path for %s: %w", path, err)
		}

		// Skip root directory itself
		if relPath == "." {
			return nil
		}

		// Add to entries list (include both files and directories for tree building)
		entries = append(entries, FileEntry{
			Path:    path,
			RelPath: relPath,
			IsDir:   isDir,
			Size:    info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort by relative path (lexicographic order)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RelPath < entries[j].RelPath
	})

	return entries, nil
}