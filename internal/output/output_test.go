package output

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/onozaty/treecat/internal/scanner"
	"github.com/onozaty/treecat/internal/tree"
)

func TestFormatter_EmptyOutput(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter(&buf)

	root := &tree.Node{Name: "", IsDir: true}
	entries := []scanner.FileEntry{}

	err := formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()
	// Should only have the blank line separator
	if result != "\n" {
		t.Errorf("Expected only newline, got %q", result)
	}
}

func TestFormatter_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewFormatter(&buf)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "test.txt", Path: "test.txt", IsDir: false},
		},
	}

	entries := []scanner.FileEntry{
		{Path: filePath, RelPath: "test.txt", IsDir: false},
	}

	err := formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()

	// Check tree section
	if !strings.Contains(result, "└── test.txt") {
		t.Error("Expected tree output to contain file")
	}

	// Check file separator
	if !strings.Contains(result, "=== test.txt ===") {
		t.Error("Expected file separator")
	}

	// Check file content
	if !strings.Contains(result, content) {
		t.Error("Expected file content")
	}
}

func TestFormatter_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("Content 1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("Content 2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewFormatter(&buf)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "file1.txt", Path: "file1.txt", IsDir: false},
			{Name: "file2.txt", Path: "file2.txt", IsDir: false},
		},
	}

	entries := []scanner.FileEntry{
		{Path: file1, RelPath: "file1.txt", IsDir: false},
		{Path: file2, RelPath: "file2.txt", IsDir: false},
	}

	err := formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()

	// Check both files are present
	if !strings.Contains(result, "=== file1.txt ===") {
		t.Error("Expected separator for file1")
	}
	if !strings.Contains(result, "=== file2.txt ===") {
		t.Error("Expected separator for file2")
	}
	if !strings.Contains(result, "Content 1") {
		t.Error("Expected content of file1")
	}
	if !strings.Contains(result, "Content 2") {
		t.Error("Expected content of file2")
	}
}

func TestFormatter_WithDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory and file
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	filePath := filepath.Join(subdir, "file.txt")
	if err := os.WriteFile(filePath, []byte("Content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewFormatter(&buf)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "subdir",
				Path:  "subdir",
				IsDir: true,
				Children: []*tree.Node{
					{Name: "file.txt", Path: "subdir/file.txt", IsDir: false},
				},
			},
		},
	}

	entries := []scanner.FileEntry{
		{Path: subdir, RelPath: "subdir", IsDir: true},
		{Path: filePath, RelPath: filepath.Join("subdir", "file.txt"), IsDir: false},
	}

	err := formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()

	// Check tree shows directory
	if !strings.Contains(result, "subdir/") {
		t.Error("Expected directory in tree")
	}

	// Check file content is output (not directory)
	separatorPath := filepath.ToSlash(filepath.Join("subdir", "file.txt"))
	expectedSeparator := "=== " + separatorPath + " ==="
	if !strings.Contains(result, expectedSeparator) {
		t.Errorf("Expected separator %q", expectedSeparator)
	}
	if !strings.Contains(result, "Content") {
		t.Error("Expected file content")
	}

	// Directory itself should not have content section
	if strings.Contains(result, "=== subdir ===") {
		t.Error("Directory should not have content section")
	}
}

func TestFormatter_FileReadError(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer
	formatter := NewFormatter(&buf)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "nonexistent.txt", Path: "nonexistent.txt", IsDir: false},
		},
	}

	// Reference non-existent file
	nonexistentPath := filepath.Join(tmpDir, "nonexistent.txt")
	entries := []scanner.FileEntry{
		{Path: nonexistentPath, RelPath: "nonexistent.txt", IsDir: false},
	}

	err := formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format should not fail on file read error: %v", err)
	}

	result := buf.String()

	// Should have separator
	if !strings.Contains(result, "=== nonexistent.txt ===") {
		t.Error("Expected file separator even for unreadable file")
	}

	// Should have error marker
	if !strings.Contains(result, "[Error reading file:") {
		t.Error("Expected error marker in output")
	}
}

func TestFormatter_CompleteOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("Content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("Content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewFormatter(&buf)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "file1.txt", Path: "file1.txt", IsDir: false},
			{Name: "file2.txt", Path: "file2.txt", IsDir: false},
		},
	}

	entries := []scanner.FileEntry{
		{Path: file1, RelPath: "file1.txt", IsDir: false},
		{Path: file2, RelPath: "file2.txt", IsDir: false},
	}

	err := formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()

	// Expected output format:
	expected := `├── file1.txt
└── file2.txt

=== file1.txt ===
Content1
=== file2.txt ===
Content2
`

	if result != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestFormatter_CompleteOutputWithDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subdir := filepath.Join(tmpDir, "dir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	file1 := filepath.Join(tmpDir, "root.txt")
	file2 := filepath.Join(subdir, "nested.txt")
	if err := os.WriteFile(file1, []byte("Root content"), 0644); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("Nested content"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewFormatter(&buf)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "dir",
				Path:  "dir",
				IsDir: true,
				Children: []*tree.Node{
					{Name: "nested.txt", Path: "dir/nested.txt", IsDir: false},
				},
			},
			{Name: "root.txt", Path: "root.txt", IsDir: false},
		},
	}

	entries := []scanner.FileEntry{
		{Path: subdir, RelPath: "dir", IsDir: true},
		{Path: file2, RelPath: filepath.Join("dir", "nested.txt"), IsDir: false},
		{Path: file1, RelPath: "root.txt", IsDir: false},
	}

	err := formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()

	// Expected output format:
	expectedSeparator1 := "=== " + filepath.ToSlash(filepath.Join("dir", "nested.txt")) + " ==="
	expected := `├── dir/
│   └── nested.txt
└── root.txt

` + expectedSeparator1 + `
Nested content
=== root.txt ===
Root content
`

	if result != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, result)
	}
}
