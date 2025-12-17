package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/onozaty/treecat/internal/encoding"
	"github.com/onozaty/treecat/internal/scanner"
	"github.com/onozaty/treecat/internal/tree"
)

// Formatter formats and writes the output.
type Formatter struct {
	writer      io.Writer
	converter   encoding.Converter               // DEPRECATED: for backward compat during transition
	encodingMap map[string]encoding.Converter    // extension to converter map
}

// NewFormatter creates a new Formatter.
func NewFormatter(writer io.Writer) *Formatter {
	return &Formatter{
		writer:    writer,
		converter: nil,
	}
}

// NewFormatterWithEncoding creates a new Formatter with encoding conversion support.
func NewFormatterWithEncoding(writer io.Writer, converter encoding.Converter) *Formatter {
	return &Formatter{
		writer:    writer,
		converter: converter,
	}
}

// NewFormatterWithEncodingMap creates a Formatter with per-extension encoding support.
func NewFormatterWithEncodingMap(writer io.Writer, encodingMap map[string]encoding.Converter) *Formatter {
	return &Formatter{
		writer:      writer,
		encodingMap: encodingMap,
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

		// Select converter based on extension (for encodingMap) or use single converter
		var converter encoding.Converter
		if f.encodingMap != nil {
			ext := filepath.Ext(entry.Path)
			if ext != "" {
				normalizedExt := encoding.NormalizeExtension(ext)
				converter = f.encodingMap[normalizedExt]
			}
		} else if f.converter != nil {
			converter = f.converter
		}

		// Convert encoding if converter found
		if converter != nil {
			content, err = converter.ConvertToUTF8(content)
			if err != nil {
				return fmt.Errorf("failed to convert encoding for %s: %w", entry.RelPath, err)
			}
		}

		// Remove BOM from all files (not just converted ones)
		content, _ = encoding.RemoveBOM(content)

		// Normalize line endings for all files (not just converted ones)
		content = encoding.NormalizeNewlines(content)

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
