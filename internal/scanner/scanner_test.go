package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onozaty/treecat/internal/filter"
)

func TestNewScanner_InvalidPath(t *testing.T) {
	_, err := NewScanner("/nonexistent/path", nil)
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestNewScanner_FileInsteadOfDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := NewScanner(filePath, nil)
	if err == nil {
		t.Error("Expected error when path is a file, not directory")
	}
}

func TestScanner_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	scanner, err := NewScanner(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries in empty directory, got %d", len(entries))
	}
}

func TestScanner_BasicFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		"file1.txt",
		"file2.go",
		"subdir/file3.md",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	scanner, err := NewScanner(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should include both files and directories
	// Expected: file1.txt, file2.go, subdir, subdir/file3.md
	if len(entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(entries))
		for i, e := range entries {
			t.Logf("Entry %d: %s (isDir: %v)", i, e.RelPath, e.IsDir)
		}
	}

	// Check sorting (lexicographic order)
	expected := []string{"file1.txt", "file2.go", "subdir", filepath.Join("subdir", "file3.md")}
	for i, e := range entries {
		if e.RelPath != expected[i] {
			t.Errorf("Entry %d: expected %s, got %s", i, expected[i], e.RelPath)
		}
	}
}

func TestScanner_WithFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		"file1.go",
		"file2.txt",
		"file3.go",
		"test.log",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	// Create scanner first to get the absolute path it uses
	scanner, err := NewScanner(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	// Create filter using the same absolute path as scanner
	f := filter.NewPatternFilter(scanner.Root, []string{"*.go"}, nil)
	scanner.Filter = f

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only include .go files
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
		for _, e := range entries {
			t.Logf("Entry: %s", e.RelPath)
		}
	}

	for _, e := range entries {
		if filepath.Ext(e.RelPath) != ".go" {
			t.Errorf("Expected only .go files, got %s", e.RelPath)
		}
	}
}

func TestScanner_GitDirectoryExcluded(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory with files
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}
	gitFile := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitFile, []byte("git config"), 0644); err != nil {
		t.Fatalf("Failed to create git config: %v", err)
	}

	// Create normal file
	normalFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(normalFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create normal file: %v", err)
	}

	// Use CompositeFilter which excludes .git
	f := filter.NewCompositeFilter(tmpDir)
	scanner, err := NewScanner(tmpDir, f)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only include file.txt, not .git or anything inside it
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
		for _, e := range entries {
			t.Logf("Entry: %s", e.RelPath)
		}
	}

	if len(entries) > 0 && entries[0].RelPath != "file.txt" {
		t.Errorf("Expected file.txt, got %s", entries[0].RelPath)
	}
}

func TestScanner_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	paths := []string{
		"a/b/c/file.txt",
		"a/file1.txt",
		"x/y/file2.txt",
	}

	for _, p := range paths {
		fullPath := filepath.Join(tmpDir, p)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	scanner, err := NewScanner(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	entries, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Check that entries are sorted lexicographically
	for i := 1; i < len(entries); i++ {
		if entries[i-1].RelPath >= entries[i].RelPath {
			t.Errorf("Entries not sorted: %s >= %s", entries[i-1].RelPath, entries[i].RelPath)
		}
	}
}
