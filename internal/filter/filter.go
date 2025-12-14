package filter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Filter determines whether a file or directory should be included.
type Filter interface {
	ShouldInclude(path string, isDir bool) bool
}

// GitignoreFilter filters files based on .gitignore patterns.
type GitignoreFilter struct {
	matcher gitignore.Matcher
	rootDir string
}

// NewGitignoreFilter creates a new GitignoreFilter.
// If .gitignore doesn't exist, returns a filter that includes everything.
func NewGitignoreFilter(rootDir string) (*GitignoreFilter, error) {
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// No .gitignore file, return nil matcher (include everything)
		return &GitignoreFilter{
			matcher: nil,
			rootDir: rootDir,
		}, nil
	}

	// Read .gitignore patterns
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return nil, err
	}

	var patterns []gitignore.Pattern
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}

	matcher := gitignore.NewMatcher(patterns)

	return &GitignoreFilter{
		matcher: matcher,
		rootDir: rootDir,
	}, nil
}

// ShouldInclude returns true if the file should be included.
func (f *GitignoreFilter) ShouldInclude(path string, isDir bool) bool {
	if f.matcher == nil {
		return true
	}

	// Get relative path from root
	relPath, err := filepath.Rel(f.rootDir, path)
	if err != nil {
		// If we can't get relative path, include it
		return true
	}

	// Convert to forward slashes for gitignore matching
	relPath = filepath.ToSlash(relPath)

	// Split path into components for gitignore matching
	parts := strings.Split(relPath, "/")

	// Check if any part of the path is ignored
	matched := f.matcher.Match(parts, isDir)

	// If matched, it means it should be ignored
	return !matched
}

// PatternFilter filters files based on glob patterns.
type PatternFilter struct {
	includePatterns []string
	excludePatterns []string
	rootDir         string
}

// NewPatternFilter creates a new PatternFilter.
func NewPatternFilter(rootDir string, includePatterns, excludePatterns []string) *PatternFilter {
	return &PatternFilter{
		includePatterns: includePatterns,
		excludePatterns: excludePatterns,
		rootDir:         rootDir,
	}
}

// ShouldInclude returns true if the file should be included based on patterns.
func (f *PatternFilter) ShouldInclude(path string, isDir bool) bool {
	// Get relative path from root
	relPath, err := filepath.Rel(f.rootDir, path)
	if err != nil {
		return true
	}

	// Check exclude patterns first
	for _, pattern := range f.excludePatterns {
		matched, err := doublestar.Match(pattern, relPath)
		if err == nil && matched {
			return false
		}
	}

	// If include patterns are specified, only include files that match
	if len(f.includePatterns) > 0 {
		for _, pattern := range f.includePatterns {
			matched, err := doublestar.Match(pattern, relPath)
			if err == nil && matched {
				return true
			}
		}
		// If include patterns exist but nothing matched, exclude
		return false
	}

	// No include patterns, and didn't match exclude patterns
	return true
}

// CompositeFilter combines multiple filters.
type CompositeFilter struct {
	filters []Filter
	rootDir string
}

// NewCompositeFilter creates a new CompositeFilter.
func NewCompositeFilter(rootDir string, filters ...Filter) *CompositeFilter {
	return &CompositeFilter{
		filters: filters,
		rootDir: rootDir,
	}
}

// ShouldInclude returns true if all filters agree the file should be included.
func (f *CompositeFilter) ShouldInclude(path string, isDir bool) bool {
	// Always exclude .git directory
	if strings.Contains(path, string(filepath.Separator)+".git"+string(filepath.Separator)) ||
		strings.HasSuffix(path, string(filepath.Separator)+".git") ||
		filepath.Base(path) == ".git" {
		return false
	}

	// Check all filters
	for _, filter := range f.filters {
		if !filter.ShouldInclude(path, isDir) {
			return false
		}
	}

	return true
}
