package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/encoding/japanese"
)

func TestIntegration_BasicDirectory(t *testing.T) {
	// Create temporary test directory
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command (create new instance for test isolation)
	cmd := newRootCmd()
	cmd.SetArgs([]string{tmpDir})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedTree := `├── file1.txt
└── file2.txt

=== file1.txt ===
Content 1
=== file2.txt ===
Content 2
`

	// Expected output should include tmpDir as first line
	expected := tmpDir + "\n" + expectedTree

	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}

func TestIntegration_WithIncludePattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	goFile := filepath.Join(tmpDir, "main.go")
	txtFile := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(goFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create go file: %v", err)
	}
	if err := os.WriteFile(txtFile, []byte("readme"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command with include pattern (create new instance for test isolation)
	cmd := newRootCmd()
	cmd.SetArgs([]string{tmpDir, "--include", "**/*.go"})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedTree := `└── main.go

=== main.go ===
package main
`

	expected := tmpDir + "\n" + expectedTree

	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}

func TestIntegration_WithExcludePattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	goFile := filepath.Join(tmpDir, "main.go")
	logFile := filepath.Join(tmpDir, "app.log")
	if err := os.WriteFile(goFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create go file: %v", err)
	}
	if err := os.WriteFile(logFile, []byte("log content"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command with exclude pattern (create new instance for test isolation)
	cmd := newRootCmd()
	cmd.SetArgs([]string{tmpDir, "--exclude", "**/*.log"})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedTree := `└── main.go

=== main.go ===
package main
`

	expected := tmpDir + "\n" + expectedTree

	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}

func TestIntegration_WithGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .gitignore
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*.log\n"), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create test files
	goFile := filepath.Join(tmpDir, "main.go")
	logFile := filepath.Join(tmpDir, "app.log")
	if err := os.WriteFile(goFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create go file: %v", err)
	}
	if err := os.WriteFile(logFile, []byte("log content"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command (gitignore should be respected by default)
	rootCmd.SetArgs([]string{tmpDir})
	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedTree := `├── .gitignore
└── main.go

=== .gitignore ===
*.log

=== main.go ===
package main
`

	expected := tmpDir + "\n" + expectedTree

	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}

func TestIntegration_WithNoGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .gitignore
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*.log\n"), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create test files
	goFile := filepath.Join(tmpDir, "main.go")
	logFile := filepath.Join(tmpDir, "app.log")
	if err := os.WriteFile(goFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create go file: %v", err)
	}
	if err := os.WriteFile(logFile, []byte("log content"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command with --no-gitignore flag (create new instance for test isolation)
	cmd := newRootCmd()
	cmd.SetArgs([]string{tmpDir, "--no-gitignore"})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedTree := `├── .gitignore
├── app.log
└── main.go

=== .gitignore ===
*.log

=== app.log ===
log content
=== main.go ===
package main
`

	expected := tmpDir + "\n" + expectedTree

	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}

func TestIntegration_GitDirectoryExcluded(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}
	gitConfig := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitConfig, []byte("git config"), 0644); err != nil {
		t.Fatalf("Failed to create git config: %v", err)
	}

	// Create normal file
	normalFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(normalFile, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create normal file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command (create new instance for test isolation)
	cmd := newRootCmd()
	cmd.SetArgs([]string{tmpDir})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedTree := `└── main.go

=== main.go ===
package main
`

	expected := tmpDir + "\n" + expectedTree

	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}

func TestIntegration_WithNestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create files
	rootFile := filepath.Join(tmpDir, "root.txt")
	nestedFile := filepath.Join(subdir, "nested.txt")
	if err := os.WriteFile(rootFile, []byte("root content"), 0644); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}
	if err := os.WriteFile(nestedFile, []byte("nested content"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command (create new instance for test isolation)
	cmd := newRootCmd()
	cmd.SetArgs([]string{tmpDir})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedSeparator := "=== " + filepath.ToSlash(filepath.Join("subdir", "nested.txt")) + " ==="
	expectedTree := `├── subdir/
│   └── nested.txt
└── root.txt

=== root.txt ===
root content
` + expectedSeparator + `
nested content
`

	expected := tmpDir + "\n" + expectedTree

	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}

func TestIntegration_PruneEmptyDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure with empty directories
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "docs"), 0755) // Empty directory
	os.MkdirAll(filepath.Join(tmpDir, "tests", "unit"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "build"), 0755) // Empty directory

	// Create .go files
	os.WriteFile(filepath.Join(tmpDir, "src", "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "tests", "unit", "test.go"), []byte("package test\n"), 0644)

	// Create non-.go file in docs (should be filtered out)
	os.WriteFile(filepath.Join(tmpDir, "docs", "README.md"), []byte("# Docs\n"), 0644)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command with --include "**/*.go" filter
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--include", "**/*.go", tmpDir})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Expected output should NOT include docs/ or build/ directories
	// since they have no .go files
	expectedSrcSeparator := "=== " + filepath.ToSlash(filepath.Join("src", "main.go")) + " ==="
	expectedTestSeparator := "=== " + filepath.ToSlash(filepath.Join("tests", "unit", "test.go")) + " ==="

	expectedTree := `├── src/
│   └── main.go
└── tests/
    └── unit/
        └── test.go

` + expectedSrcSeparator + `
package main

` + expectedTestSeparator + `
package test

`

	expected := tmpDir + "\n" + expectedTree

	if output != expected {
		t.Errorf("Output mismatch.\nExpected (%d bytes):\n%q\n\nGot (%d bytes):\n%q", len(expected), expected, len(output), output)
	}

	// Verify docs/ and build/ are NOT in the output
	if strings.Contains(output, "docs/") {
		t.Error("Output should not contain 'docs/' directory (it has no .go files)")
	}
	if strings.Contains(output, "build/") {
		t.Error("Output should not contain 'build/' directory (it's empty)")
	}
	if strings.Contains(output, "README.md") {
		t.Error("Output should not contain 'README.md' (filtered out by include pattern)")
	}
}

func TestIntegration_EncodingConversion(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a Shift_JIS encoded file
	sjisFile := filepath.Join(tmpDir, "japanese.txt")
	originalText := "こんにちは世界"
	encoder := japanese.ShiftJIS.NewEncoder()
	sjisBytes, err := encoder.Bytes([]byte(originalText))
	if err != nil {
		t.Fatalf("Failed to encode to Shift_JIS: %v", err)
	}
	if err := os.WriteFile(sjisFile, sjisBytes, 0644); err != nil {
		t.Fatalf("Failed to create Shift_JIS file: %v", err)
	}

	// Create a UTF-8 file for comparison
	utf8File := filepath.Join(tmpDir, "english.txt")
	utf8Content := "Hello, World!"
	if err := os.WriteFile(utf8File, []byte(utf8Content), 0644); err != nil {
		t.Fatalf("Failed to create UTF-8 file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command with --encoding flag
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--encoding", "shift_jis", tmpDir})
	err = cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	result := buf.String()

	// Verify the Shift_JIS file was converted to UTF-8
	if !strings.Contains(result, originalText) {
		t.Errorf("Expected UTF-8 content %q in output, but not found", originalText)
	}

	// Verify the UTF-8 file is output correctly
	if !strings.Contains(result, utf8Content) {
		t.Errorf("Expected UTF-8 content %q in output, but not found", utf8Content)
	}

	// Verify tree structure
	if !strings.Contains(result, "japanese.txt") {
		t.Error("Expected 'japanese.txt' in tree output")
	}
	if !strings.Contains(result, "english.txt") {
		t.Error("Expected 'english.txt' in tree output")
	}

	// Verify file separators
	if !strings.Contains(result, "=== japanese.txt ===") {
		t.Error("Expected separator for japanese.txt")
	}
	if !strings.Contains(result, "=== english.txt ===") {
		t.Error("Expected separator for english.txt")
	}
}

func TestIntegration_InvalidEncoding(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a simple file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run command with invalid encoding (no need to capture stdout for error test)
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--encoding", "invalid-encoding", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid encoding, got nil")
	}

	// Verify error message mentions encoding
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "encoding") {
		t.Errorf("Expected 'encoding' in error message, got: %s", errorMsg)
	}
}

func TestIntegration_EncodingWithMultipleFiles(t *testing.T) {
	// Create temporary test directory with subdirectory
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create multiple Shift_JIS files
	files := map[string]string{
		filepath.Join(tmpDir, "file1.txt"):    "ファイル１",
		filepath.Join(subdir, "file2.txt"):    "ファイル２",
		filepath.Join(tmpDir, "file3.txt"):    "ファイル３",
	}

	encoder := japanese.ShiftJIS.NewEncoder()
	for path, text := range files {
		sjisBytes, err := encoder.Bytes([]byte(text))
		if err != nil {
			t.Fatalf("Failed to encode %s: %v", path, err)
		}
		if err := os.WriteFile(path, sjisBytes, 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command with --encoding flag
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--encoding", "shift_jis", tmpDir})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	result := buf.String()

	// Verify all files were converted correctly
	for _, text := range files {
		if !strings.Contains(result, text) {
			t.Errorf("Expected text %q in output, but not found", text)
		}
	}

	// Verify tree structure includes subdirectory
	if !strings.Contains(result, "subdir/") {
		t.Error("Expected 'subdir/' in tree output")
	}
}

func TestIntegration_NoEncodingFlag(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a UTF-8 file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, 世界"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command without --encoding flag (should work as before)
	cmd := newRootCmd()
	cmd.SetArgs([]string{tmpDir})
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	result := buf.String()

	// Verify content is output (no conversion, raw bytes)
	if !strings.Contains(result, content) {
		t.Errorf("Expected content %q in output", content)
	}
}
