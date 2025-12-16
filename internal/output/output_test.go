package output

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/onozaty/treecat/internal/encoding"
	"github.com/onozaty/treecat/internal/scanner"
	"github.com/onozaty/treecat/internal/tree"
	"golang.org/x/text/encoding/japanese"
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
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}

	// Should have error message
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected 'failed to read file' in error message, got: %v", err)
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

func TestFormatter_WithEncodingConversion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with Shift_JIS content
	filePath := filepath.Join(tmpDir, "sjis.txt")
	originalText := "こんにちは世界"
	encoder := japanese.ShiftJIS.NewEncoder()
	shiftJISBytes, err := encoder.Bytes([]byte(originalText))
	if err != nil {
		t.Fatalf("Failed to encode to Shift_JIS: %v", err)
	}
	if err := os.WriteFile(filePath, shiftJISBytes, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create converter
	converter, err := encoding.NewConverter("shift_jis")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewFormatterWithEncoding(&buf, converter)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "sjis.txt", Path: "sjis.txt", IsDir: false},
		},
	}

	entries := []scanner.FileEntry{
		{Path: filePath, RelPath: "sjis.txt", IsDir: false},
	}

	err = formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()

	// Check that the content was converted to UTF-8
	if !strings.Contains(result, originalText) {
		t.Errorf("Expected UTF-8 content %q in output", originalText)
	}

	// Verify tree section
	if !strings.Contains(result, "└── sjis.txt") {
		t.Error("Expected tree output to contain file")
	}

	// Verify file separator
	if !strings.Contains(result, "=== sjis.txt ===") {
		t.Error("Expected file separator")
	}
}

func TestFormatter_WithoutConverter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with UTF-8 content
	filePath := filepath.Join(tmpDir, "utf8.txt")
	content := "Hello, 世界"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var buf bytes.Buffer
	// Create formatter without converter (nil converter)
	formatter := NewFormatterWithEncoding(&buf, nil)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "utf8.txt", Path: "utf8.txt", IsDir: false},
		},
	}

	entries := []scanner.FileEntry{
		{Path: filePath, RelPath: "utf8.txt", IsDir: false},
	}

	err := formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()

	// Check that the content is output as-is (no conversion)
	if !strings.Contains(result, content) {
		t.Errorf("Expected content %q in output", content)
	}
}

func TestFormatter_EncodingConversionError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with invalid Shift_JIS content (binary data)
	filePath := filepath.Join(tmpDir, "binary.dat")
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	if err := os.WriteFile(filePath, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create converter
	converter, err := encoding.NewConverter("shift_jis")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewFormatterWithEncoding(&buf, converter)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "binary.dat", Path: "binary.dat", IsDir: false},
		},
	}

	entries := []scanner.FileEntry{
		{Path: filePath, RelPath: "binary.dat", IsDir: false},
	}

	// Note: transform.Reader doesn't return errors for invalid sequences,
	// it uses replacement characters instead. So this should succeed.
	err = formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format should not fail with invalid bytes (replacement chars used): %v", err)
	}

	result := buf.String()
	// Result should contain something (with replacement characters)
	if len(result) == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestFormatter_MultipleFilesWithEncoding(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files with Shift_JIS content
	file1Path := filepath.Join(tmpDir, "file1.txt")
	file2Path := filepath.Join(tmpDir, "file2.txt")

	text1 := "ファイル１"
	text2 := "ファイル２"

	encoder := japanese.ShiftJIS.NewEncoder()
	sjisBytes1, _ := encoder.Bytes([]byte(text1))
	sjisBytes2, _ := encoder.Bytes([]byte(text2))

	if err := os.WriteFile(file1Path, sjisBytes1, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2Path, sjisBytes2, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Create converter
	converter, err := encoding.NewConverter("shift_jis")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewFormatterWithEncoding(&buf, converter)

	root := &tree.Node{
		Name:  "",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "file1.txt", Path: "file1.txt", IsDir: false},
			{Name: "file2.txt", Path: "file2.txt", IsDir: false},
		},
	}

	entries := []scanner.FileEntry{
		{Path: file1Path, RelPath: "file1.txt", IsDir: false},
		{Path: file2Path, RelPath: "file2.txt", IsDir: false},
	}

	err = formatter.Format(root, entries)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	result := buf.String()

	// Check that both files were converted to UTF-8
	if !strings.Contains(result, text1) {
		t.Errorf("Expected UTF-8 content %q in output", text1)
	}
	if !strings.Contains(result, text2) {
		t.Errorf("Expected UTF-8 content %q in output", text2)
	}
}
