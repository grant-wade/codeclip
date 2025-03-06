package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grant-wade/codeclip/internal/finder"
	"github.com/grant-wade/codeclip/internal/search"
)

// FormatFiles formats file contents with code backticks
func FormatFiles(files []finder.FileContent) string {
	var builder strings.Builder

	for _, file := range files {
		builder.WriteString(fmt.Sprintf("```%s filename=%s\n", file.Language, file.Path))
		builder.WriteString(file.Content)
		builder.WriteString("\n```\n\n")
	}

	return builder.String()
}

// FormatSearchResults formats search results with code backticks
func FormatSearchResults(results search.SearchResult) string {
	var builder strings.Builder

	// Collect all snippets by file path for further processing
	fileSnippets := make(map[string][]formattingSnippet)

	for _, file := range results.Files {
		for _, snippet := range file.Snippets {
			if _, exists := fileSnippets[file.Path]; !exists {
				fileSnippets[file.Path] = []formattingSnippet{}
			}

			fileSnippets[file.Path] = append(fileSnippets[file.Path], formattingSnippet{
				StartLine: snippet.StartLine,
				EndLine:   snippet.EndLine,
				Content:   snippet.Content,
				Language:  file.Language,
				Path:      file.Path,
			})
		}
	}

	// Process each file's snippets separately
	for _, snippets := range fileSnippets {
		// Deduplicate and merge snippets for this file
		mergedSnippets := mergeFormattingSnippets(snippets)

		// Add the merged snippets to the output
		for _, snippet := range mergedSnippets {
			builder.WriteString(fmt.Sprintf("```%s filename=%s (lines %d-%d)\n",
				snippet.Language, snippet.Path, snippet.StartLine, snippet.EndLine))
			builder.WriteString(snippet.Content)
			builder.WriteString("\n```\n\n")
		}
	}

	return builder.String()
}

// formattingSnippet holds snippet data for formatting purposes
type formattingSnippet struct {
	StartLine int
	EndLine   int
	Content   string
	Language  string
	Path      string
}

// mergeFormattingSnippets eliminates duplicate snippets and merges overlapping ones
func mergeFormattingSnippets(snippets []formattingSnippet) []formattingSnippet {
	if len(snippets) <= 1 {
		return snippets
	}

	// Sort by start line
	sort.Slice(snippets, func(i, j int) bool {
		return snippets[i].StartLine < snippets[j].StartLine
	})

	// Merge overlapping snippets
	result := []formattingSnippet{snippets[0]}
	for i := 1; i < len(snippets); i++ {
		current := snippets[i]
		previous := &result[len(result)-1]

		// Check for significant overlap (>50% of the smaller snippet)
		overlap := min(previous.EndLine, current.EndLine) - max(previous.StartLine, current.StartLine)
		smallerRange := min(previous.EndLine-previous.StartLine, current.EndLine-current.StartLine)

		if overlap > 0 && float64(overlap) >= float64(smallerRange)*0.5 {
			// Merge snippets - take the union of line ranges
			newStartLine := min(previous.StartLine, current.StartLine)
			newEndLine := max(previous.EndLine, current.EndLine)

			// Get the full content for the merged range by using the snippet with larger range
			// or the current one if they have equal range
			var newContent string
			if current.EndLine-current.StartLine >= previous.EndLine-previous.StartLine {
				newContent = current.Content
			} else {
				newContent = previous.Content
			}

			// Update previous with merged data
			previous.StartLine = newStartLine
			previous.EndLine = newEndLine
			previous.Content = newContent
		} else if current.StartLine > previous.EndLine {
			// No overlap, add as a new snippet
			result = append(result, current)
		} else if current.EndLine > previous.EndLine {
			previous.EndLine = current.EndLine
			// We need to combine contents - this is a simplified approach
			// In a real implementation you might need to read the file again to get the proper content
			linesPrevious := strings.Split(previous.Content, "\n")
			linesCurrent := strings.Split(current.Content, "\n")

			// Calculate how many lines to take from current
			linesToAdd := current.EndLine - previous.EndLine
			if len(linesCurrent) >= linesToAdd {
				linesPrevious = append(linesPrevious, linesCurrent[len(linesCurrent)-linesToAdd:]...)
				previous.Content = strings.Join(linesPrevious, "\n")
			}
		}
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
