package cmd

import (
	"fmt"
	"os"

	"github.com/grant-wade/codeclip/internal/output"
	"github.com/grant-wade/codeclip/internal/template"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template [template-file]",
	Short: "Process a template file with embedded content tags",
	Long: `Process a template file containing content tags in the format {{path/to/file}} or {{glob/pattern}}.
The content of these files will be embedded in the output at the position of the tags.

Examples:
  codeclip template project-overview.md
  codeclip template docs-template.md --output documentation.md

Template syntax:
  {{filename.txt}}          - Include content of a specific file
  {{**/*.go}}               - Include all .go files matching the glob pattern
  {{dir/*.{js,ts}}}         - Include all .js and .ts files in dir/

If a file isn't found, the tag will be replaced with an empty string.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templatePath := args[0]

		// Check if template exists
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return fmt.Errorf("template file not found: %s", templatePath)
		}

		// Process template
		parser := template.NewParser(inputPath)
		processed, err := parser.Process(templatePath)
		if err != nil {
			return fmt.Errorf("template processing failed: %w", err)
		}

		// Calculate stats
		stats := output.CalculateStats(processed)

		// Check token limit if specified
		if maxTokens > 0 && stats.EstimatedTokens > maxTokens {
			return fmt.Errorf("output exceeds token limit: %d > %d", stats.EstimatedTokens, maxTokens)
		}

		// Send to output target
		err = output.CopyToTarget(processed, outputTarget)
		if err != nil {
			return err
		}

		// Print summary
		output.PrintSummary(stats, []string{templatePath})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
