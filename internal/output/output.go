package output

import (
	"fmt"
	"io"
	"os"

	"github.com/onozaty/treecat/internal/scanner"
	"github.com/onozaty/treecat/internal/tree"
)

// Formatter formats and writes the output.
type Formatter struct {
	writer io.Writer
}

// NewFormatter creates a new Formatter.
func NewFormatter(writer io.Writer) *Formatter {
	return &Formatter{
		writer: writer,
	}
}

// Format writes the complete output (tree + file contents).
func (f *Formatter) Format(treeRoot *tree.Node, entries []scanner.FileEntry) error {
	// Write tree section
	treeOutput := tree.Render(treeRoot)
	if _, err := f.writer.Write([]byte(treeOutput)); err != nil {
		return fmt.Errorf("failed to write tree output: %w", err)
	}

	// Add blank line separator between tree and file contents
	if _, err := f.writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Write file contents section
	for _, entry := range entries {
		// Skip directories (only output files)
		if entry.IsDir {
			continue
		}

		// Write file separator with relative path
		separator := fmt.Sprintf("=== %s ===\n", entry.RelPath)
		if _, err := f.writer.Write([]byte(separator)); err != nil {
			return fmt.Errorf("failed to write file separator: %w", err)
		}

		// Read and write file contents
		content, err := os.ReadFile(entry.Path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", entry.RelPath, err)
		}

		// Write file content
		if _, err := f.writer.Write(content); err != nil {
			return fmt.Errorf("failed to write file content for %s: %w", entry.RelPath, err)
		}

		// Add blank line after file content
		if _, err := f.writer.Write([]byte("\n")); err != nil {
			return fmt.Errorf("failed to write blank line: %w", err)
		}
	}

	return nil
}
