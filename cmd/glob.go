package cmd

import (
	"github.com/grant-wade/codeclip/internal/finder"
	"github.com/grant-wade/codeclip/internal/output"
	"github.com/spf13/cobra"
)

var globCmd = &cobra.Command{
	Use:   "glob [pattern]",
	Short: "Select files using glob pattern and copy to clipboard",
	Long: `Select files matching the provided glob pattern and copy their contents to clipboard.
Examples:
  codeclip glob "**/*.go"
  codeclip glob "src/**/*.{js,ts}" --output file.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		
		files, err := finder.FindFilesByGlob(inputPath, pattern)
		if err != nil {
			return err
		}
		
		result, err := finder.ReadFiles(files)
		if err != nil {
			return err
		}
		
		formatted := output.FormatFiles(result)
		stats := output.CalculateStats(formatted)
		
		err = output.CopyToTarget(formatted, outputTarget)
		if err != nil {
			return err
		}
		
		output.PrintSummary(stats, result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(globCmd)
}