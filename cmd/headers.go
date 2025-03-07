package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/grant-wade/codeclip/internal/finder"
	"github.com/grant-wade/codeclip/internal/output"
	"github.com/spf13/cobra"
)

var headersCmd = &cobra.Command{
	Use:   "headers [file or glob pattern]",
	Short: "Extract headers (functions, classes, etc.) from code files",
	Long: `Extract and display structural elements (headers) from code files.
This command identifies functions, classes, methods, and other important structures
based on the programming language of each file.

Examples:
  codeclip headers main.go
  codeclip headers "**/*.go"
  codeclip headers --output headers.md "src/**/*.{js,ts}"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]

		// Check if the input is a single file or a glob pattern
		files := []string{}
		if isSingleFile(pattern) {
			// Add single file directly
			files = append(files, pattern)
		} else {
			// Find files matching the glob pattern
			globFiles, err := finder.FindFilesByGlob(inputPath, pattern)
			if err != nil {
				return fmt.Errorf("failed to find files: %w", err)
			}
			files = globFiles
		}

		if len(files) == 0 {
			return fmt.Errorf("no files found matching pattern: %s", pattern)
		}

		var allHeaders strings.Builder
		allHeaders.WriteString("# Code Structure Headers\n\n")

		// Process each file
		for _, filePath := range files {
			headers, err := finder.CollectHeaders(filePath)
			if err != nil {
				return fmt.Errorf("failed to collect headers from %s: %w", filePath, err)
			}

			if len(headers) > 0 {
				allHeaders.WriteString(fmt.Sprintf("## %s\n\n", filePath))
				allHeaders.WriteString("```\n")
				allHeaders.WriteString(finder.FormatHeaders(headers))
				allHeaders.WriteString("```\n\n")
			}
		}

		formatted := allHeaders.String()
		stats := output.CalculateStats(formatted)

		// Copy to target (clipboard, stdout, or file)
		err := output.CopyToTarget(formatted, outputTarget)
		if err != nil {
			return fmt.Errorf("failed to copy output: %w", err)
		}

		// Print summary
		output.PrintSummary(stats, files)
		return nil
	},
}

// isSingleFile checks if the pattern appears to be a single file rather than a glob pattern
func isSingleFile(pattern string) bool {
	return !strings.ContainsAny(pattern, "*?[]{}") && fileExists(pattern)
}

// fileExists checks if the specified file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	rootCmd.AddCommand(headersCmd)

	// Add specific flags for this command if needed
	// For now, we're using the flags already defined in root.go
}
