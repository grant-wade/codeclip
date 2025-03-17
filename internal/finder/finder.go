package finder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// FileContent represents the content of a file
type FileContent struct {
	Path     string
	Content  string
	Language string
}

// FindFilesByGlob locates files matching the given glob pattern
func FindFilesByGlob(basePath, pattern string) ([]string, error) {
	var matches []string

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}

		match, err := doublestar.Match(pattern, relPath)
		if err != nil {
			return err
		}

		if match && !info.IsDir() {
			matches = append(matches, path)
		}
		return nil
	})

	return matches, err
}

// FindAllCodeFiles finds all code files in the given directory
func FindAllCodeFiles(basePath string) ([]string, error) {
	return FindFilesByGlob(basePath, "**/*.{go,js,ts,py,java,c,cpp,h,hpp,cs,rb,php,pl,sh,html,css,md,json,yaml,yml}")
}

// ReadFiles reads the content of the provided files
func ReadFiles(paths []string) ([]FileContent, error) {
	var results []FileContent

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		language := DetectLanguage(path)
		results = append(results, FileContent{
			Path:     path,
			Content:  string(data),
			Language: language,
		})
	}

	return results, nil
}

// detectLanguage determines the language of a file based on its extension
func DetectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".pl":
		return "perl"
	case ".sh":
		return "bash"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "plaintext"
	}
}
