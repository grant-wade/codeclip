package cmd

import (
	"fmt"

	"github.com/grant-wade/codeclip/internal/finder"
	"github.com/grant-wade/codeclip/internal/output"
	"github.com/grant-wade/codeclip/internal/search"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [pattern]",
	Short: "Search for code matching pattern and copy to clipboard",
	Long: `Search for code matching the provided pattern and copy results to clipboard with context.
Examples:
  codeclip search "func GetUser" --function
  codeclip search "api.call" --context 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		searchPattern := args[0]

		files, err := finder.FindAllCodeFiles(inputPath)
		if err != nil {
			return err
		}

		searchResults, err := search.SearchInFiles(files, searchPattern, search.Options{
			ContextLines:   contextLines,
			EntireFunction: entireFunction,
			FuzzySearch:    fuzzySearch,
		})
		if err != nil {
			return err
		}

		formatted := output.FormatSearchResults(searchResults)
		stats := output.CalculateStats(formatted)

		if maxTokens > 0 && stats.EstimatedTokens > maxTokens {
			return fmt.Errorf("output exceeds token limit: %d > %d", stats.EstimatedTokens, maxTokens)
		}

		err = output.CopyToTarget(formatted, outputTarget)
		if err != nil {
			return err
		}

		output.PrintSummary(stats, searchResults)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
