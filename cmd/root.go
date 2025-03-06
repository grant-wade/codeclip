package cmd

import (
	"github.com/spf13/cobra"
)

var (
	contextLines    int
	entireFunction  bool
	fuzzySearch     bool
	outputTarget    string
	estimateTokens  bool
	maxTokens       int
	inputPath       string
)

var rootCmd = &cobra.Command{
	Use:   "codeclip",
	Short: "Copy code to clipboard with advanced filtering",
	Long: `A CLI tool for quickly copying codebases to your clipboard with support
for advanced filtering, search, and formatting options.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&contextLines, "context", "c", 3, "Number of context lines before and after matches")
	rootCmd.PersistentFlags().BoolVarP(&entireFunction, "function", "f", false, "Include entire function/method containing matches")
	rootCmd.PersistentFlags().BoolVarP(&fuzzySearch, "fuzzy", "z", false, "Enable fuzzy matching for search terms")
	rootCmd.PersistentFlags().StringVarP(&outputTarget, "output", "o", "clipboard", "Output destination (clipboard, stdout, or file path)")
	rootCmd.PersistentFlags().BoolVarP(&estimateTokens, "estimate", "e", true, "Estimate token count in output")
	rootCmd.PersistentFlags().IntVarP(&maxTokens, "max-tokens", "m", 0, "Maximum tokens to copy (0 for unlimited)")
	rootCmd.PersistentFlags().StringVarP(&inputPath, "path", "p", ".", "Path to search in")
}