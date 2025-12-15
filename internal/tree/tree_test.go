package tree

import (
	"strings"
	"testing"

	"github.com/onozaty/treecat/internal/scanner"
)

func TestBuild_EmptyEntries(t *testing.T) {
	entries := []scanner.FileEntry{}
	root := Build(entries)

	if root.Name != "" {
		t.Errorf("Expected root name to be empty, got %q", root.Name)
	}
	if !root.IsDir {
		t.Error("Expected root to be a directory")
	}
	if len(root.Children) != 0 {
		t.Errorf("Expected no children, got %d", len(root.Children))
	}
}

func TestBuild_SingleFile(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "file.txt", IsDir: false},
	}
	root := Build(entries)

	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}

	child := root.Children[0]
	if child.Name != "file.txt" {
		t.Errorf("Expected name 'file.txt', got %q", child.Name)
	}
	if child.IsDir {
		t.Error("Expected file, not directory")
	}
}

func TestBuild_NestedStructure(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "dir1", IsDir: true},
		{RelPath: "dir1/file1.txt", IsDir: false},
		{RelPath: "dir1/file2.txt", IsDir: false},
		{RelPath: "dir2", IsDir: true},
		{RelPath: "dir2/subdir", IsDir: true},
		{RelPath: "dir2/subdir/file3.txt", IsDir: false},
		{RelPath: "file.txt", IsDir: false},
	}
	root := Build(entries)

	// Check root has 3 children: dir1, dir2, file.txt
	if len(root.Children) != 3 {
		t.Fatalf("Expected 3 children at root, got %d", len(root.Children))
	}

	// Check sorting: directories first (dir1, dir2), then files (file.txt)
	if root.Children[0].Name != "dir1" || !root.Children[0].IsDir {
		t.Errorf("Expected first child to be dir1 (dir), got %s (isDir=%v)", root.Children[0].Name, root.Children[0].IsDir)
	}
	if root.Children[1].Name != "dir2" || !root.Children[1].IsDir {
		t.Errorf("Expected second child to be dir2 (dir), got %s (isDir=%v)", root.Children[1].Name, root.Children[1].IsDir)
	}
	if root.Children[2].Name != "file.txt" || root.Children[2].IsDir {
		t.Errorf("Expected third child to be file.txt (file), got %s (isDir=%v)", root.Children[2].Name, root.Children[2].IsDir)
	}

	// Check dir1 has 2 files
	dir1 := root.Children[0]
	if len(dir1.Children) != 2 {
		t.Errorf("Expected dir1 to have 2 children, got %d", len(dir1.Children))
	}

	// Check dir2/subdir structure
	dir2 := root.Children[1]
	if len(dir2.Children) != 1 {
		t.Errorf("Expected dir2 to have 1 child, got %d", len(dir2.Children))
	}
	subdir := dir2.Children[0]
	if subdir.Name != "subdir" || !subdir.IsDir {
		t.Errorf("Expected subdir (dir), got %s (isDir=%v)", subdir.Name, subdir.IsDir)
	}
	if len(subdir.Children) != 1 {
		t.Errorf("Expected subdir to have 1 child, got %d", len(subdir.Children))
	}
}

func TestRender_EmptyTree(t *testing.T) {
	root := &Node{Name: "", IsDir: true}
	result := Render(root)

	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestRender_SingleFile(t *testing.T) {
	root := &Node{
		Name:  "",
		IsDir: true,
		Children: []*Node{
			{Name: "file.txt", IsDir: false},
		},
	}
	result := Render(root)

	expected := "└── file.txt\n"
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestRender_MultipleFiles(t *testing.T) {
	root := &Node{
		Name:  "",
		IsDir: true,
		Children: []*Node{
			{Name: "file1.txt", IsDir: false},
			{Name: "file2.txt", IsDir: false},
		},
	}
	result := Render(root)

	expected := "├── file1.txt\n└── file2.txt\n"
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestRender_DirectoriesAndFiles(t *testing.T) {
	root := &Node{
		Name:  "",
		IsDir: true,
		Children: []*Node{
			{
				Name:  "dir1",
				IsDir: true,
				Children: []*Node{
					{Name: "file1.txt", IsDir: false},
				},
			},
			{Name: "file2.txt", IsDir: false},
		},
	}
	result := Render(root)

	expected := `├── dir1/
│   └── file1.txt
└── file2.txt
`
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestRender_ComplexStructure(t *testing.T) {
	root := &Node{
		Name:  "",
		IsDir: true,
		Children: []*Node{
			{
				Name:  "dir1",
				IsDir: true,
				Children: []*Node{
					{Name: "file1.txt", IsDir: false},
					{Name: "file2.txt", IsDir: false},
				},
			},
			{
				Name:  "dir2",
				IsDir: true,
				Children: []*Node{
					{
						Name:  "subdir",
						IsDir: true,
						Children: []*Node{
							{Name: "file3.txt", IsDir: false},
						},
					},
				},
			},
			{Name: "file.txt", IsDir: false},
		},
	}
	result := Render(root)

	expected := `├── dir1/
│   ├── file1.txt
│   └── file2.txt
├── dir2/
│   └── subdir/
│       └── file3.txt
└── file.txt
`
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestBuildAndRender_Integration(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "README.md", IsDir: false},
		{RelPath: "cmd", IsDir: true},
		{RelPath: "cmd/main.go", IsDir: false},
		{RelPath: "internal", IsDir: true},
		{RelPath: "internal/filter", IsDir: true},
		{RelPath: "internal/filter/filter.go", IsDir: false},
		{RelPath: "internal/scanner", IsDir: true},
		{RelPath: "internal/scanner/scanner.go", IsDir: false},
	}

	root := Build(entries)
	result := Render(root)

	// Check that the output contains expected patterns
	if !strings.Contains(result, "cmd/") {
		t.Error("Expected output to contain 'cmd/'")
	}
	if !strings.Contains(result, "main.go") {
		t.Error("Expected output to contain 'main.go'")
	}
	if !strings.Contains(result, "internal/") {
		t.Error("Expected output to contain 'internal/'")
	}
	if !strings.Contains(result, "filter/") {
		t.Error("Expected output to contain 'filter/'")
	}

	// Verify tree structure with box-drawing characters
	lines := strings.Split(result, "\n")
	if len(lines) < 8 {
		t.Errorf("Expected at least 8 lines, got %d", len(lines))
	}

	// Directories should come before files
	var foundCmd, foundInternal, foundReadme bool
	for _, line := range lines {
		if strings.Contains(line, "cmd/") {
			foundCmd = true
		}
		if strings.Contains(line, "internal/") {
			foundInternal = true
		}
		if strings.Contains(line, "README.md") {
			foundReadme = true
			// README should come after directories
			if !foundCmd || !foundInternal {
				t.Error("Expected README.md to come after directories")
			}
		}
	}

	if !foundCmd || !foundInternal || !foundReadme {
		t.Error("Expected to find cmd/, internal/, and README.md in output")
	}
}

func TestPruneEmptyDirectories_LeafDirectory(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "empty", IsDir: true},
		{RelPath: "file.txt", IsDir: false},
	}
	root := Build(entries)

	// Should only have the file, empty directory should be pruned
	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}
	if root.Children[0].Name != "file.txt" || root.Children[0].IsDir {
		t.Errorf("Expected file.txt (file), got %s (isDir=%v)", root.Children[0].Name, root.Children[0].IsDir)
	}
}

func TestPruneEmptyDirectories_IntermediateEmpty(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "dir1", IsDir: true},
		{RelPath: "dir1/dir2", IsDir: true},
		{RelPath: "dir1/dir2/file.txt", IsDir: false},
		{RelPath: "empty", IsDir: true},
	}
	root := Build(entries)

	// Should have only dir1, empty should be pruned
	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.Children))
	}
	if root.Children[0].Name != "dir1" {
		t.Errorf("Expected dir1, got %s", root.Children[0].Name)
	}

	// Check dir1 has dir2
	dir1 := root.Children[0]
	if len(dir1.Children) != 1 {
		t.Fatalf("Expected dir1 to have 1 child, got %d", len(dir1.Children))
	}
	if dir1.Children[0].Name != "dir2" {
		t.Errorf("Expected dir2, got %s", dir1.Children[0].Name)
	}

	// Check dir2 has file.txt
	dir2 := dir1.Children[0]
	if len(dir2.Children) != 1 {
		t.Fatalf("Expected dir2 to have 1 child, got %d", len(dir2.Children))
	}
	if dir2.Children[0].Name != "file.txt" {
		t.Errorf("Expected file.txt, got %s", dir2.Children[0].Name)
	}
}

func TestPruneEmptyDirectories_NestedEmpty(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "dir1", IsDir: true},
		{RelPath: "dir1/dir2", IsDir: true},
		{RelPath: "dir1/dir2/dir3", IsDir: true},
	}
	root := Build(entries)

	// All directories are empty, should have no children
	if len(root.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(root.Children))
	}
}

func TestPruneEmptyDirectories_MixedStructure(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "dir1", IsDir: true},
		{RelPath: "dir1/file.txt", IsDir: false},
		{RelPath: "dir2", IsDir: true},
		{RelPath: "dir3", IsDir: true},
		{RelPath: "dir3/sub", IsDir: true},
		{RelPath: "dir3/sub/file.txt", IsDir: false},
		{RelPath: "dir4", IsDir: true},
	}
	root := Build(entries)

	// Should have dir1 and dir3, dir2 and dir4 should be pruned
	if len(root.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(root.Children))
	}

	// Check dir1 with file.txt
	if root.Children[0].Name != "dir1" {
		t.Errorf("Expected first child to be dir1, got %s", root.Children[0].Name)
	}
	if len(root.Children[0].Children) != 1 {
		t.Errorf("Expected dir1 to have 1 child, got %d", len(root.Children[0].Children))
	}

	// Check dir3 with sub/file.txt
	if root.Children[1].Name != "dir3" {
		t.Errorf("Expected second child to be dir3, got %s", root.Children[1].Name)
	}
	if len(root.Children[1].Children) != 1 {
		t.Errorf("Expected dir3 to have 1 child, got %d", len(root.Children[1].Children))
	}
	if root.Children[1].Children[0].Name != "sub" {
		t.Errorf("Expected dir3 child to be sub, got %s", root.Children[1].Children[0].Name)
	}
}

func TestPruneEmptyDirectories_AllEmpty(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "dir1", IsDir: true},
		{RelPath: "dir2", IsDir: true},
		{RelPath: "dir3", IsDir: true},
	}
	root := Build(entries)

	// All directories are empty, should have no children
	if len(root.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(root.Children))
	}
}

func TestPruneEmptyDirectories_OnlyFiles(t *testing.T) {
	entries := []scanner.FileEntry{
		{RelPath: "file1.txt", IsDir: false},
		{RelPath: "file2.txt", IsDir: false},
	}
	root := Build(entries)

	// Should have both files unchanged
	if len(root.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(root.Children))
	}
	if root.Children[0].Name != "file1.txt" || root.Children[0].IsDir {
		t.Errorf("Expected file1.txt (file), got %s (isDir=%v)", root.Children[0].Name, root.Children[0].IsDir)
	}
	if root.Children[1].Name != "file2.txt" || root.Children[1].IsDir {
		t.Errorf("Expected file2.txt (file), got %s (isDir=%v)", root.Children[1].Name, root.Children[1].IsDir)
	}
}