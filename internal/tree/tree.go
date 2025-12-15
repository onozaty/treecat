package tree

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/onozaty/treecat/internal/scanner"
)

// Node represents a node in the directory tree.
type Node struct {
	Name     string  // File or directory name
	Path     string  // Relative path from root
	IsDir    bool    // Whether this is a directory
	Children []*Node // Child nodes (for directories)
}

// Build builds a tree structure from a flat list of file entries.
func Build(entries []scanner.FileEntry) *Node {
	root := &Node{
		Name:  "",
		Path:  "",
		IsDir: true,
	}

	for _, entry := range entries {
		if entry.RelPath == "" || entry.RelPath == "." {
			continue
		}

		// Split path into components
		parts := strings.Split(filepath.ToSlash(entry.RelPath), "/")

		// Navigate/create the tree structure
		current := root
		for i, part := range parts {
			isLast := i == len(parts)-1

			// Find existing child or create new one
			child := findChild(current, part)
			if child == nil {
				child = &Node{
					Name:  part,
					Path:  strings.Join(parts[:i+1], "/"),
					IsDir: !isLast || entry.IsDir,
				}
				current.Children = append(current.Children, child)
			}

			current = child
		}
	}

	// Sort all nodes recursively
	sortNode(root)

	// Prune empty directories (directories without any descendant files)
	pruneEmptyDirectories(root)

	return root
}

// findChild finds a child node by name.
func findChild(node *Node, name string) *Node {
	for _, child := range node.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// sortNode sorts children recursively (directories first, then alphabetically).
func sortNode(node *Node) {
	sort.Slice(node.Children, func(i, j int) bool {
		// Directories come before files
		if node.Children[i].IsDir != node.Children[j].IsDir {
			return node.Children[i].IsDir
		}
		// Otherwise, sort alphabetically
		return node.Children[i].Name < node.Children[j].Name
	})

	// Recursively sort children
	for _, child := range node.Children {
		if child.IsDir {
			sortNode(child)
		}
	}
}

// hasDescendantFiles returns true if a directory node has any descendant files.
// A directory has descendant files if:
// - It directly contains at least one file (non-directory child)
// - At least one of its child directories has descendant files
func hasDescendantFiles(node *Node) bool {
	if !node.IsDir {
		return true // Files always count
	}

	for _, child := range node.Children {
		if !child.IsDir {
			return true // Found a file child
		}
		if hasDescendantFiles(child) {
			return true // Found descendant files in child directory
		}
	}

	return false // No files found
}

// pruneEmptyDirectories removes directory nodes that don't contain any descendant files.
// It recursively processes the tree bottom-up and removes childless directories.
func pruneEmptyDirectories(node *Node) {
	if !node.IsDir {
		return
	}

	// Process children first (bottom-up traversal)
	var keptChildren []*Node
	for _, child := range node.Children {
		if child.IsDir {
			pruneEmptyDirectories(child) // Recursive prune
			// Keep directory only if it has descendant files
			if hasDescendantFiles(child) {
				keptChildren = append(keptChildren, child)
			}
		} else {
			// Always keep files
			keptChildren = append(keptChildren, child)
		}
	}

	node.Children = keptChildren
}

// Render renders the tree structure as a string with box-drawing characters.
func Render(root *Node) string {
	var builder strings.Builder
	renderNode(root, "", true, &builder)
	return builder.String()
}

// renderNode recursively renders a node and its children.
func renderNode(node *Node, prefix string, isLast bool, builder *strings.Builder) {
	// Skip root node name
	if node.Name != "" {
		// Determine the connector
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// Write the current node
		builder.WriteString(prefix)
		builder.WriteString(connector)
		builder.WriteString(node.Name)
		if node.IsDir {
			builder.WriteString("/")
		}
		builder.WriteString("\n")
	}

	// Render children
	childPrefix := prefix
	if node.Name != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		renderNode(child, childPrefix, isLastChild, builder)
	}
}
