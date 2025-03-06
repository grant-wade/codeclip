package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/grant-wade/codeclip/internal/finder"
	"github.com/grant-wade/codeclip/internal/search"
)

// Stats represents statistics about the copied content
type Stats struct {
	FileCount       int
	SnippetCount    int
	CharCount       int
	LineCount       int
	EstimatedTokens int
}

// CalculateStats computes statistics for the copied content
func CalculateStats(content string) Stats {
	stats := Stats{
		CharCount:    len(content),
		LineCount:    strings.Count(content, "\n") + 1,
		SnippetCount: strings.Count(content, "```") / 2,
	}
	
	// Simple token estimation - adjust based on your needs
	// Most LLMs use ~4 chars per token on average
	stats.EstimatedTokens = (stats.CharCount + 3) / 4
	
	return stats
}

// PrintSummary displays a summary of the copied content
func PrintSummary(stats Stats, files interface{}) {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)
	blue := color.New(color.FgBlue)

	fileCount := 0
	switch v := files.(type) {
	case []finder.FileContent:
		fileCount = len(v)
	case search.SearchResult:
		fileCount = len(v.Files)
	}

	bold.Println("\n📋 Codeclip Summary:")
	fmt.Printf("  Files: %d\n", fileCount)
	fmt.Printf("  Snippets: %d\n", stats.SnippetCount)
	fmt.Printf("  Lines: %d\n", stats.LineCount)
	fmt.Printf("  Characters: %d\n", stats.CharCount)
	blue.Printf("  Est. Tokens: %d\n\n", stats.EstimatedTokens)

	if stats.SnippetCount > 0 {
		green.Println("✓ Code successfully copied!")
	}
}