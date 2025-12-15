package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/onozaty/treecat/internal/filter"
	"github.com/onozaty/treecat/internal/output"
	"github.com/onozaty/treecat/internal/scanner"
	"github.com/onozaty/treecat/internal/tree"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "treecat [directory]",
		Short: "Combine multiple files into one with tree structure for LLM consumption",
		Long: `treecat is a CLI tool that combines multiple files from a directory into a single output.
It displays a directory tree structure at the top, followed by file contents separated by markers.
Perfect for providing codebase context to LLMs.`,
		Args: cobra.MaximumNArgs(1),
		RunE: run,
	}

	cmd.Flags().StringSliceP("exclude", "e", []string{}, "Exclude patterns (comma-separated glob patterns)")
	cmd.Flags().StringSliceP("include", "i", []string{}, "Include patterns (comma-separated glob patterns)")
	cmd.Flags().Bool("no-gitignore", false, "Ignore .gitignore file")

	return cmd
}

var rootCmd = newRootCmd()

func run(cmd *cobra.Command, args []string) error {
	// Get flag values
	excludePatterns, _ := cmd.Flags().GetStringSlice("exclude")
	includePatterns, _ := cmd.Flags().GetStringSlice("include")
	noGitignore, _ := cmd.Flags().GetBool("no-gitignore")
	// Get target directory (default to current directory)
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create filters
	var filters []filter.Filter

	// Add gitignore filter (unless disabled)
	if !noGitignore {
		gitignoreFilter, err := filter.NewGitignoreFilter(absPath)
		if err != nil {
			return fmt.Errorf("failed to create gitignore filter: %w", err)
		}
		filters = append(filters, gitignoreFilter)
	}

	// Add pattern filter (if include/exclude patterns are specified)
	if len(includePatterns) > 0 || len(excludePatterns) > 0 {
		patternFilter := filter.NewPatternFilter(absPath, includePatterns, excludePatterns)
		filters = append(filters, patternFilter)
	}

	// Create composite filter
	compositeFilter := filter.NewCompositeFilter(absPath, filters...)

	// Create scanner
	scan, err := scanner.NewScanner(absPath, compositeFilter)
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	// Scan directory
	entries, err := scan.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Build tree
	treeRoot := tree.Build(entries)

	// Create formatter and output
	formatter := output.NewFormatter(os.Stdout)
	if err := formatter.Format(treeRoot, entries); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}