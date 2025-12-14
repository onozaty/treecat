package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	excludePatterns []string
	includePatterns []string
	noGitignore     bool
)

var rootCmd = &cobra.Command{
	Use:   "treecat [directory]",
	Short: "Combine multiple files into one with tree structure for LLM consumption",
	Long: `treecat is a CLI tool that combines multiple files from a directory into a single output.
It displays a directory tree structure at the top, followed by file contents separated by markers.
Perfect for providing codebase context to LLMs.`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Exclude patterns (comma-separated glob patterns)")
	rootCmd.Flags().StringSliceVarP(&includePatterns, "include", "i", []string{}, "Include patterns (comma-separated glob patterns)")
	rootCmd.Flags().BoolVar(&noGitignore, "no-gitignore", false, "Ignore .gitignore file")
}

func run(cmd *cobra.Command, args []string) error {
	// Get target directory (default to current directory)
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	// TODO: Implement the main logic
	fmt.Fprintf(os.Stderr, "Target directory: %s\n", targetDir)
	fmt.Fprintf(os.Stderr, "Exclude patterns: %v\n", excludePatterns)
	fmt.Fprintf(os.Stderr, "Include patterns: %v\n", includePatterns)
	fmt.Fprintf(os.Stderr, "No gitignore: %v\n", noGitignore)

	fmt.Println("treecat implementation coming soon...")

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}