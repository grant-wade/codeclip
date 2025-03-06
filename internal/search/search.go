package search

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/grant-wade/codeclip/internal/finder"
)

// Options defines search configuration options
type Options struct {
	ContextLines   int
	EntireFunction bool
	FuzzySearch    bool
}

// SearchResult represents a search match with context
type SearchResult struct {
	Files []SearchFile
}

// SearchFile represents a file with search matches
type SearchFile struct {
	Path     string
	Language string
	Snippets []CodeSnippet
}

// CodeSnippet represents a matched code snippet
type CodeSnippet struct {
	StartLine int
	EndLine   int
	Content   string
	MatchInfo string
}

// SearchInFiles searches for pattern in the given files
func SearchInFiles(files []string, pattern string, opts Options) (SearchResult, error) {
	result := SearchResult{}

	for _, file := range files {
		matches, err := searchInFile(file, pattern, opts)
		if err != nil {
			return result, err
		}

		if len(matches.Snippets) > 0 {
			// Merge overlapping snippets before adding to result
			matches.Snippets = mergeOverlappingSnippets(matches.Snippets)
			result.Files = append(result.Files, matches)
		}
	}

	return result, nil
}

// mergeOverlappingSnippets combines snippets that overlap or are adjacent
func mergeOverlappingSnippets(snippets []CodeSnippet) []CodeSnippet {
	if len(snippets) <= 1 {
		return snippets
	}

	// Sort snippets by start line
	sort.Slice(snippets, func(i, j int) bool {
		return snippets[i].StartLine < snippets[j].StartLine
	})

	// Merge overlapping or adjacent snippets
	result := []CodeSnippet{snippets[0]}
	for i := 1; i < len(snippets); i++ {
		current := snippets[i]
		previous := &result[len(result)-1]

		// Check if current snippet overlaps or is adjacent to the previous one
		// Using a small buffer (e.g., 3 lines) to merge closely located snippets
		const adjacencyBuffer = 3
		if current.StartLine <= previous.EndLine+adjacencyBuffer {
			// Merge the snippets
			if current.EndLine > previous.EndLine {
				// Combine content by using the lines from the combined range
				lines := strings.Split(previous.Content, "\n")
				currentLines := strings.Split(current.Content, "\n")

				// Calculate how many new lines to add from the current snippet
				linesToAdd := current.EndLine - previous.EndLine
				if linesToAdd > 0 && len(currentLines) >= linesToAdd {
					lines = append(lines, currentLines[len(currentLines)-linesToAdd:]...)
				}

				previous.Content = strings.Join(lines, "\n")
				previous.EndLine = current.EndLine
			}

			// Update match info to reflect the merge
			if previous.MatchInfo != current.MatchInfo {
				previous.MatchInfo = "Multiple matches between lines " +
					strconv.Itoa(previous.StartLine) + "-" + strconv.Itoa(previous.EndLine)
			}
		} else {
			// No overlap, add as a new snippet
			result = append(result, current)
		}
	}

	return result
}

// searchInFile searches for pattern in a single file
func searchInFile(path, pattern string, opts Options) (SearchFile, error) {
	fileObj := SearchFile{
		Path:     path,
		Language: finder.DetectLanguage(path),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fileObj, err
	}

	var regex *regexp.Regexp
	if opts.FuzzySearch {
		// Create a fuzzy pattern
		fuzzyPattern := strings.Join(strings.Split(pattern, ""), ".*")
		regex, err = regexp.Compile("(?i)" + fuzzyPattern)
	} else {
		regex, err = regexp.Compile(pattern)
	}
	if err != nil {
		return fileObj, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for i, line := range lines {
		if regex.MatchString(line) {
			snippet := extractSnippet(lines, i, opts)
			fileObj.Snippets = append(fileObj.Snippets, snippet)
		}
	}

	return fileObj, nil
}

// extractSnippet extracts code around the match based on options
func extractSnippet(lines []string, matchLine int, opts Options) CodeSnippet {
	// If entire function mode is on, try to extract the function
	if opts.EntireFunction {
		start, end := findFunctionBounds(lines, matchLine)
		if start != -1 && end != -1 {
			return CodeSnippet{
				StartLine: start + 1, // 1-indexed for display
				EndLine:   end + 1,
				Content:   strings.Join(lines[start:end+1], "\n"),
				MatchInfo: "Function containing match at line " + strconv.Itoa(matchLine+1),
			}
		}
	}

	// Fall back to context lines
	start := max(0, matchLine-opts.ContextLines)
	end := min(len(lines)-1, matchLine+opts.ContextLines)

	return CodeSnippet{
		StartLine: start + 1, // 1-indexed for display
		EndLine:   end + 1,
		Content:   strings.Join(lines[start:end+1], "\n"),
		MatchInfo: "Match at line " + strconv.Itoa(matchLine+1),
	}
}

// findFunctionBounds tries to determine function start/end around the match line
func findFunctionBounds(lines []string, matchLine int) (start, end int) {
	// Simple heuristic - look for opening/closing braces or indentation changes
	// This would need to be language-specific for a robust implementation

	// For now, this is a simplistic implementation that works with C-like languages
	start = matchLine
	for start > 0 {
		if strings.Contains(lines[start], "func ") || strings.Contains(lines[start], "function ") {
			break
		}
		start--
	}

	bracketCount := 0
	for i := start; i < len(lines); i++ {
		line := lines[i]
		bracketCount += strings.Count(line, "{") - strings.Count(line, "}")

		if bracketCount <= 0 && i > matchLine {
			end = i
			break
		}
	}

	if end == 0 {
		end = min(len(lines)-1, start+20) // Fallback to 20 lines if we can't find the end
	}

	return start, end
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
