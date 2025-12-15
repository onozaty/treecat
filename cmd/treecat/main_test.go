package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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

	expected := `├── file1.txt
└── file2.txt

=== file1.txt ===
Content 1
=== file2.txt ===
Content 2
`

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

	expected := `└── main.go

=== main.go ===
package main
`

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

	expected := `└── main.go

=== main.go ===
package main
`

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

	expected := `├── .gitignore
└── main.go

=== .gitignore ===
*.log

=== main.go ===
package main
`

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

	expected := `├── .gitignore
├── app.log
└── main.go

=== .gitignore ===
*.log

=== app.log ===
log content
=== main.go ===
package main
`

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

	expected := `└── main.go

=== main.go ===
package main
`

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
	expected := `├── subdir/
│   └── nested.txt
└── root.txt

=== root.txt ===
root content
` + expectedSeparator + `
nested content
`

	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}
