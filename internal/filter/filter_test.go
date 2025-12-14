package filter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitignoreFilter_NoGitignore(t *testing.T) {
	// Create temp directory without .gitignore
	tmpDir := t.TempDir()

	filter, err := NewGitignoreFilter(tmpDir)
	if err != nil {
		t.Fatalf("NewGitignoreFilter failed: %v", err)
	}

	// Should include everything when no .gitignore exists
	testPath := filepath.Join(tmpDir, "test.txt")
	if !filter.ShouldInclude(testPath, false) {
		t.Error("Expected file to be included when no .gitignore exists")
	}
}

func TestGitignoreFilter_WithGitignore(t *testing.T) {
	// Create temp directory with .gitignore
	tmpDir := t.TempDir()

	gitignoreContent := `*.log
node_modules/
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	filter, err := NewGitignoreFilter(tmpDir)
	if err != nil {
		t.Fatalf("NewGitignoreFilter failed: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{"normal file", filepath.Join(tmpDir, "test.txt"), false, true},
		{"log file", filepath.Join(tmpDir, "app.log"), false, false},
		{"node_modules dir", filepath.Join(tmpDir, "node_modules"), true, false},
		{"file in node_modules", filepath.Join(tmpDir, "node_modules", "package.json"), false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.ShouldInclude(tt.path, tt.isDir)
			if got != tt.expected {
				t.Errorf("ShouldInclude(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.expected)
			}
		})
	}
}

func TestPatternFilter_Exclude(t *testing.T) {
	tmpDir := t.TempDir()

	filter := NewPatternFilter(tmpDir, nil, []string{"*.log", "temp/*"})

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"normal file", filepath.Join(tmpDir, "test.txt"), true},
		{"log file", filepath.Join(tmpDir, "app.log"), false},
		{"file in temp", filepath.Join(tmpDir, "temp", "file.txt"), false},
		{"go file", filepath.Join(tmpDir, "main.go"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.ShouldInclude(tt.path, false)
			if got != tt.expected {
				t.Errorf("ShouldInclude(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestPatternFilter_Include(t *testing.T) {
	tmpDir := t.TempDir()

	filter := NewPatternFilter(tmpDir, []string{"*.go", "*.md"}, nil)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"go file", filepath.Join(tmpDir, "main.go"), true},
		{"md file", filepath.Join(tmpDir, "README.md"), true},
		{"txt file", filepath.Join(tmpDir, "notes.txt"), false},
		{"log file", filepath.Join(tmpDir, "app.log"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.ShouldInclude(tt.path, false)
			if got != tt.expected {
				t.Errorf("ShouldInclude(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestPatternFilter_IncludeAndExclude(t *testing.T) {
	tmpDir := t.TempDir()

	// Include all .go files, but exclude test files
	filter := NewPatternFilter(tmpDir, []string{"*.go"}, []string{"*_test.go"})

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"main.go", filepath.Join(tmpDir, "main.go"), true},
		{"test file", filepath.Join(tmpDir, "main_test.go"), false},
		{"txt file", filepath.Join(tmpDir, "readme.txt"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.ShouldInclude(tt.path, false)
			if got != tt.expected {
				t.Errorf("ShouldInclude(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestCompositeFilter_GitDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	filter := NewCompositeFilter(tmpDir)

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{".git dir", filepath.Join(tmpDir, ".git"), true, false},
		{"file in .git", filepath.Join(tmpDir, ".git", "config"), false, false},
		{"normal file", filepath.Join(tmpDir, "test.txt"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.ShouldInclude(tt.path, tt.isDir)
			if got != tt.expected {
				t.Errorf("ShouldInclude(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.expected)
			}
		})
	}
}

func TestPatternFilter_IncludeWithDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Include only .go files (with ** for nested files)
	filter := NewPatternFilter(tmpDir, []string{"**/*.go"}, nil)

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
	}{
		{"go file in root", filepath.Join(tmpDir, "main.go"), false, true},
		{"txt file", filepath.Join(tmpDir, "readme.txt"), false, false},
		{"directory", filepath.Join(tmpDir, "subdir"), true, true}, // Directories should always be included with include patterns
		{"nested go file", filepath.Join(tmpDir, "subdir", "file.go"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filter.ShouldInclude(tt.path, tt.isDir)
			if got != tt.expected {
				t.Errorf("ShouldInclude(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.expected)
			}
		})
	}
}

func TestCompositeFilter_MultipleFilters(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .gitignore
	gitignoreContent := `*.log
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	gitignoreFilter, err := NewGitignoreFilter(tmpDir)
	if err != nil {
		t.Fatalf("NewGitignoreFilter failed: %v", err)
	}

	patternFilter := NewPatternFilter(tmpDir, []string{"*.go", "*.md"}, nil)

	composite := NewCompositeFilter(tmpDir, gitignoreFilter, patternFilter)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"go file", filepath.Join(tmpDir, "main.go"), true},
		{"md file", filepath.Join(tmpDir, "README.md"), true},
		{"log file (gitignore)", filepath.Join(tmpDir, "app.log"), false},
		{"txt file (not in include)", filepath.Join(tmpDir, "notes.txt"), false},
		{".git dir", filepath.Join(tmpDir, ".git"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := composite.ShouldInclude(tt.path, false)
			if got != tt.expected {
				t.Errorf("ShouldInclude(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}