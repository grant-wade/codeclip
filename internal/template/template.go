package template

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/grant-wade/codeclip/internal/finder"
	"github.com/grant-wade/codeclip/internal/output"
)

// Parser represents a template parser
type Parser struct {
	BasePath string
}

// NewParser creates a new template parser
func NewParser(basePath string) *Parser {
	return &Parser{
		BasePath: basePath,
	}
}

// Process processes the template file and returns the processed content
func (p *Parser) Process(templatePath string) (string, error) {
	templateFile, err := os.Open(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to open template file: %w", err)
	}
	defer templateFile.Close()

	return p.ProcessReader(templateFile)
}

// ProcessReader processes a template from a reader
func (p *Parser) ProcessReader(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	var result strings.Builder

	// Define the regex pattern to match {{...}} expressions
	tagPattern := regexp.MustCompile(`{{([^}]+)}}`)

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()

		// Look for tags in the line
		matches := tagPattern.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			// No tags, just add the line as is
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Process line with tags
		processedLine := line
		for _, match := range matches {
			fullTag := match[0] // The full {{...}} tag
			content := match[1] // Just the part between {{ and }}
			content = strings.TrimSpace(content)

			// Replace the tag with its resolved content
			replacement, err := p.resolveTag(content)
			if err != nil {
				// Don't fail on resolution errors, just leave a comment
				replacement = fmt.Sprintf("<!-- Failed to resolve %s: %v -->", content, err)
			}

			// If we're replacing the whole line with content, format accordingly
			if fullTag == line {
				processedLine = replacement
			} else {
				processedLine = strings.Replace(processedLine, fullTag, replacement, 1)
			}
		}

		result.WriteString(processedLine)
		if !strings.HasSuffix(processedLine, "\n") {
			result.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading template: %w", err)
	}

	return result.String(), nil
}

// resolveTag resolves a single tag to its content
func (p *Parser) resolveTag(tag string) (string, error) {
	// Check if it's a glob pattern
	if strings.Contains(tag, "*") {
		return p.resolveGlob(tag)
	}

	// Otherwise treat as a direct file reference
	return p.resolveFile(tag)
}

// resolveFile resolves a direct file reference
func (p *Parser) resolveFile(filePath string) (string, error) {
	fullPath := filePath
	if !strings.HasPrefix(filePath, "/") {
		fullPath = fmt.Sprintf("%s/%s", p.BasePath, filePath)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File not found, return empty string without error
			return "", nil
		}
		return "", err
	}

	language := finder.DetectLanguage(filePath)
	return fmt.Sprintf("```%s\n%s\n```", language, string(content)), nil
}

// resolveGlob resolves a glob pattern to matching files
func (p *Parser) resolveGlob(pattern string) (string, error) {
	files, err := finder.FindFilesByGlob(p.BasePath, pattern)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		// No files found, return empty string without error
		return "", nil
	}

	// Read all files and format them
	fileContents, err := finder.ReadFiles(files)
	if err != nil {
		return "", err
	}

	return output.FormatFiles(fileContents), nil
}
