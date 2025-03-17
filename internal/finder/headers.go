package finder

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// HeaderType represents the type of header element found
type HeaderType string

const (
	Function  HeaderType = "Function"
	Method    HeaderType = "Method"
	Class     HeaderType = "Class"
	Interface HeaderType = "Interface"
	Variable  HeaderType = "Variable"
	Constant  HeaderType = "Constant"
	Import    HeaderType = "Import"
	Field     HeaderType = "Field"
	Enum      HeaderType = "Enum"
	Struct    HeaderType = "Struct"
	Package   HeaderType = "Package"
	Namespace HeaderType = "Namespace"
	Module    HeaderType = "Module"
	Define    HeaderType = "Define"
)

// HeaderElement represents a structural element in a file
type HeaderElement struct {
	Type        HeaderType
	Name        string
	LineNum     int
	Signature   string
	Scope       string          // public, private, protected, etc.
	Parent      string          // for methods, the class/struct they belong to
	Children    []HeaderElement // For nested elements like struct fields or imports
	EndLine     int             // To track where blocks end
	Parameters  []ParameterInfo // For functions/methods, parameter list with types
	ReturnTypes []string        // For functions/methods, return type(s)
	ValueType   string          // For constants/variables, their type
	Value       string          // For constants, their assigned value
}

// ParameterInfo stores detailed information about a parameter
type ParameterInfo struct {
	Name string
	Type string
}

// LanguagePattern defines regex patterns for identifying structural elements
type LanguagePattern struct {
	ElementType    HeaderType
	Pattern        *regexp.Regexp
	NameGroup      int // Which regex group contains the element name
	SignatureGroup int // Which regex group contains the full signature (if any)
	ScopeGroup     int // Which regex group contains the scope (if any)
	ParentGroup    int // Which regex group contains the parent name (if any)
	ParamsGroup    int // Which regex group contains the parameters (if any)
	ReturnsGroup   int // Which regex group contains the return type (if any)
}

// languagePatternRegistry contains the regex patterns for each language
var languagePatternRegistry = map[string][]LanguagePattern{
	"go": {
		{
			ElementType:  Function,
			Pattern:      regexp.MustCompile(`^func\s+([A-Za-z0-9_]+)\s*\((.*?)\)(?:\s+\(?([^{]*)\)?)?`),
			NameGroup:    1,
			ParamsGroup:  2,
			ReturnsGroup: 3,
		},
		{
			ElementType:  Method,
			Pattern:      regexp.MustCompile(`^func\s+\(\w+\s+\*?([A-Za-z0-9_]+)\)\s+([A-Za-z0-9_]+)\s*\((.*?)\)(?:\s+\(?([^{]*)\)?)?`),
			NameGroup:    2,
			ParentGroup:  1,
			ParamsGroup:  3,
			ReturnsGroup: 4,
		},
		{
			ElementType: Struct,
			Pattern:     regexp.MustCompile(`^type\s+([A-Za-z0-9_]+)\s+struct\s+{`),
			NameGroup:   1,
		},
		{
			ElementType: Interface,
			Pattern:     regexp.MustCompile(`^type\s+([A-Za-z0-9_]+)\s+interface\s+{`),
			NameGroup:   1,
		},
		{
			ElementType: Constant,
			Pattern:     regexp.MustCompile(`^const\s+([A-Za-z0-9_]+)(?:\s+([A-Za-z0-9_\[\]<>]+))?(?:\s*=\s*(.+))?`),
			NameGroup:   1,
		},
		{
			ElementType: Variable,
			Pattern:     regexp.MustCompile(`^var\s+([A-Za-z0-9_]+)(?:\s+([A-Za-z0-9_\[\]<>]+))?(?:\s*=\s*(.+))?`),
			NameGroup:   1,
		},
		{
			ElementType: Import,
			Pattern:     regexp.MustCompile(`^import\s+\(`),
			NameGroup:   0,
		},
	},
	"python": {
		{
			ElementType:  Function,
			Pattern:      regexp.MustCompile(`^def\s+([A-Za-z0-9_]+)\s*\((.*?)\)(?:\s*->\s*([A-Za-z0-9_\[\],\s\.]+))?:`),
			NameGroup:    1,
			ParamsGroup:  2,
			ReturnsGroup: 3,
		},
		{
			ElementType:  Method,
			Pattern:      regexp.MustCompile(`^\s+def\s+([A-Za-z0-9_]+)\s*\((self|cls)(?:,\s*(.*?))?\)(?:\s*->\s*([A-Za-z0-9_\[\],\s\.]+))?:`),
			NameGroup:    1,
			ParamsGroup:  3, // Group 3 contains parameters after self/cls
			ReturnsGroup: 4,
		},
		{
			ElementType: Class,
			Pattern:     regexp.MustCompile(`^class\s+([A-Za-z0-9_]+)(?:\((.*?)\))?:`),
			NameGroup:   1,
			ParamsGroup: 2, // For parent classes
		},
		{
			ElementType: Import,
			Pattern:     regexp.MustCompile(`^(?:from\s+([A-Za-z0-9_\.]+)\s+)?import\s+(.+)`),
			NameGroup:   2, // Import names
			ParentGroup: 1, // From module
		},
		{
			ElementType:    Variable,
			Pattern:        regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)\s*=\s*(.+)`),
			NameGroup:      1,
			SignatureGroup: 2, // Value
		},
	},
	"java": {
		{
			ElementType: Class,
			Pattern:     regexp.MustCompile(`\b(public|private|protected)?\s+class\s+([A-Za-z0-9_]+)`),
			NameGroup:   2,
			ScopeGroup:  1,
		},
		{
			ElementType: Method,
			Pattern:     regexp.MustCompile(`\b(public|private|protected)?\s+[A-Za-z0-9_<>]+\s+([A-Za-z0-9_]+)\s*\(`),
			NameGroup:   2,
			ScopeGroup:  1,
		},
		{
			ElementType: Interface,
			Pattern:     regexp.MustCompile(`\b(public|private|protected)?\s+interface\s+([A-Za-z0-9_]+)`),
			NameGroup:   2,
			ScopeGroup:  1,
		},
	},
	"javascript": {
		{
			ElementType: Function,
			Pattern:     regexp.MustCompile(`function\s+([A-Za-z0-9_]+)\s*\(`),
			NameGroup:   1,
		},
		{
			ElementType: Variable,
			Pattern:     regexp.MustCompile(`(const|let|var)\s+([A-Za-z0-9_]+)\s*=`),
			NameGroup:   2,
		},
		{
			ElementType: Class,
			Pattern:     regexp.MustCompile(`class\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
		{
			ElementType: Method,
			Pattern:     regexp.MustCompile(`\b([A-Za-z0-9_]+)\s*\(\)\s*{`),
			NameGroup:   1,
		},
	},
	"typescript": {
		{
			ElementType: Function,
			Pattern:     regexp.MustCompile(`function\s+([A-Za-z0-9_]+)\s*\(`),
			NameGroup:   1,
		},
		{
			ElementType: Variable,
			Pattern:     regexp.MustCompile(`(const|let|var)\s+([A-Za-z0-9_]+)\s*:`),
			NameGroup:   2,
		},
		{
			ElementType: Class,
			Pattern:     regexp.MustCompile(`class\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
		{
			ElementType: Interface,
			Pattern:     regexp.MustCompile(`interface\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
	},
	"c": {
		{
			ElementType: Function,
			Pattern:     regexp.MustCompile(`^[A-Za-z0-9_]+\s+([A-Za-z0-9_]+)\s*\(`),
			NameGroup:   1,
		},
		{
			ElementType: Struct,
			Pattern:     regexp.MustCompile(`struct\s+([A-Za-z0-9_]+)\s*{`),
			NameGroup:   1,
		},
		{
			ElementType: Define,
			Pattern:     regexp.MustCompile(`#define\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
	},
	"cpp": {
		{
			ElementType: Function,
			Pattern:     regexp.MustCompile(`^[A-Za-z0-9_:<>]+\s+([A-Za-z0-9_]+)\s*\(`),
			NameGroup:   1,
		},
		{
			ElementType: Class,
			Pattern:     regexp.MustCompile(`class\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
		{
			ElementType: Struct,
			Pattern:     regexp.MustCompile(`struct\s+([A-Za-z0-9_]+)\s*{`),
			NameGroup:   1,
		},
		{
			ElementType: Namespace,
			Pattern:     regexp.MustCompile(`namespace\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
	},
	"csharp": {
		{
			ElementType: Class,
			Pattern:     regexp.MustCompile(`\b(public|private|protected|internal)?\s+class\s+([A-Za-z0-9_]+)`),
			NameGroup:   2,
			ScopeGroup:  1,
		},
		{
			ElementType: Method,
			Pattern:     regexp.MustCompile(`\b(public|private|protected|internal)?\s+[A-Za-z0-9_<>]+\s+([A-Za-z0-9_]+)\s*\(`),
			NameGroup:   2,
			ScopeGroup:  1,
		},
		{
			ElementType: Interface,
			Pattern:     regexp.MustCompile(`\b(public|private|protected|internal)?\s+interface\s+([A-Za-z0-9_]+)`),
			NameGroup:   2,
			ScopeGroup:  1,
		},
		{
			ElementType: Namespace,
			Pattern:     regexp.MustCompile(`namespace\s+([A-Za-z0-9_\.]+)`),
			NameGroup:   1,
		},
	},
	"ruby": {
		{
			ElementType: Class,
			Pattern:     regexp.MustCompile(`^class\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
		{
			ElementType: Function,
			Pattern:     regexp.MustCompile(`^\s*def\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
		{
			ElementType: Module,
			Pattern:     regexp.MustCompile(`^module\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
	},
	"php": {
		{
			ElementType: Class,
			Pattern:     regexp.MustCompile(`class\s+([A-Za-z0-9_]+)`),
			NameGroup:   1,
		},
		{
			ElementType: Function,
			Pattern:     regexp.MustCompile(`function\s+([A-Za-z0-9_]+)\s*\(`),
			NameGroup:   1,
		},
		{
			ElementType: Method,
			Pattern:     regexp.MustCompile(`\b(public|private|protected)?\s+function\s+([A-Za-z0-9_]+)\s*\(`),
			NameGroup:   2,
			ScopeGroup:  1,
		},
	},
}

// defaultPatterns contains generic patterns that might work across languages
var defaultPatterns = []LanguagePattern{
	{
		ElementType: Function,
		Pattern:     regexp.MustCompile(`\b(function|def|func)\s+([A-Za-z0-9_]+)\s*\(`),
		NameGroup:   2,
	},
	{
		ElementType: Class,
		Pattern:     regexp.MustCompile(`\b(class)\s+([A-Za-z0-9_]+)`),
		NameGroup:   2,
	},
}

// CollectHeaders extracts headers (functions, classes, etc.) from a file
func CollectHeaders(path string) ([]HeaderElement, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	language := DetectLanguage(path)

	// Get patterns for this language
	patterns, exists := languagePatternRegistry[language]
	if !exists {
		// Fall back to default patterns
		patterns = defaultPatterns
	}

	var headers []HeaderElement
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Variables to track block elements (like imports, structs)
	var currentBlock *HeaderElement
	var inImportBlock bool
	var braceCount int
	var currentClass string

	// Reread the file contents for block processing
	fileContents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file for block processing: %w", err)
	}
	lines := strings.Split(string(fileContents), "\n")

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Skip empty lines and common comment prefixes
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" ||
			strings.HasPrefix(trimmedLine, "//") ||
			strings.HasPrefix(trimmedLine, "#") ||
			strings.HasPrefix(trimmedLine, "/*") {
			continue
		}

		// Track indentation level for Python
		if language == "python" {
			leadingSpaces := len(line) - len(strings.TrimLeft(line, " \t"))
			if leadingSpaces == 0 && trimmedLine != "" {
				// Reset class context when going back to top level
				currentClass = ""
			}
		}

		// Track braces for block elements
		braceCount += strings.Count(trimmedLine, "{") - strings.Count(trimmedLine, "}")

		// Check if we're in an import block
		if inImportBlock {
			if trimmedLine == ")" {
				inImportBlock = false
				if currentBlock != nil {
					currentBlock.EndLine = lineNum
					headers = append(headers, *currentBlock)
					currentBlock = nil
				}
			} else if !strings.HasPrefix(trimmedLine, "//") && trimmedLine != "" {
				// Extract package name from import line
				importName := strings.Trim(trimmedLine, " \t\"")
				if currentBlock != nil {
					currentBlock.Children = append(currentBlock.Children, HeaderElement{
						Type:    Import,
						Name:    importName,
						LineNum: lineNum,
					})
				}
			}
			continue
		}

		// Try each pattern for this language
		for _, pattern := range patterns {
			matches := pattern.Pattern.FindStringSubmatch(trimmedLine)
			if len(matches) > pattern.NameGroup && pattern.NameGroup > 0 {
				header := HeaderElement{
					Type:    pattern.ElementType,
					Name:    matches[pattern.NameGroup],
					LineNum: lineNum,
				}

				// Add signature if available
				if pattern.SignatureGroup > 0 && len(matches) > pattern.SignatureGroup {
					header.Signature = matches[pattern.SignatureGroup]
				} else {
					// Use the entire line as a fallback signature
					header.Signature = trimmedLine
				}

				// Add scope if available
				if pattern.ScopeGroup > 0 && len(matches) > pattern.ScopeGroup {
					header.Scope = matches[pattern.ScopeGroup]
				}

				// Add parent if available
				if pattern.ParentGroup > 0 && len(matches) > pattern.ParentGroup {
					header.Parent = matches[pattern.ParentGroup]
				}

				// For Python methods, set the current class as parent if not specified
				if language == "python" && pattern.ElementType == Method && header.Parent == "" && currentClass != "" {
					header.Parent = currentClass
				}

				// Add parameters if available
				if pattern.ParamsGroup > 0 && len(matches) > pattern.ParamsGroup {
					params := parseParametersWithTypes(matches[pattern.ParamsGroup], language)
					header.Parameters = params
				}

				// Add return types if available
				if pattern.ReturnsGroup > 0 && len(matches) > pattern.ReturnsGroup {
					returns := parseReturnTypes(matches[pattern.ReturnsGroup], language)
					header.ReturnTypes = returns
				}

				// Track current class for Python
				if language == "python" && pattern.ElementType == Class {
					currentClass = header.Name
				}

				// Handle Python imports specifically
				if language == "python" && pattern.ElementType == Import {
					// Special handling for Python imports
					if header.Parent != "" { // "from X import Y" case
						moduleFrom := header.Parent
						importItems := strings.Split(header.Name, ",")

						for _, item := range importItems {
							item = strings.TrimSpace(item)
							if item == "*" {
								header.Children = append(header.Children, HeaderElement{
									Type:    Import,
									Name:    "*",
									Parent:  moduleFrom,
									LineNum: lineNum,
								})
							} else if strings.Contains(item, " as ") {
								parts := strings.Split(item, " as ")
								origName := strings.TrimSpace(parts[0])
								aliasName := strings.TrimSpace(parts[1])
								header.Children = append(header.Children, HeaderElement{
									Type:      Import,
									Name:      origName,
									Parent:    moduleFrom,
									Signature: aliasName, // Use Signature to store alias
									LineNum:   lineNum,
								})
							} else if item != "" {
								header.Children = append(header.Children, HeaderElement{
									Type:    Import,
									Name:    item,
									Parent:  moduleFrom,
									LineNum: lineNum,
								})
							}
						}
					} else { // Direct "import X" case
						importItems := strings.Split(header.Name, ",")
						for _, item := range importItems {
							item = strings.TrimSpace(item)
							if strings.Contains(item, " as ") {
								parts := strings.Split(item, " as ")
								origName := strings.TrimSpace(parts[0])
								aliasName := strings.TrimSpace(parts[1])
								header.Children = append(header.Children, HeaderElement{
									Type:      Import,
									Name:      origName,
									Signature: aliasName, // Use Signature to store alias
									LineNum:   lineNum,
								})
							} else if item != "" {
								header.Children = append(header.Children, HeaderElement{
									Type:    Import,
									Name:    item,
									LineNum: lineNum,
								})
							}
						}
					}
				}

				// Handle constants and variables with type and value info
				if header.Type == Constant || header.Type == Variable {
					if len(matches) > 2 && matches[2] != "" {
						header.ValueType = matches[2]
					}
					if len(matches) > 3 && matches[3] != "" {
						header.Value = strings.TrimSpace(matches[3])
						// Remove trailing comments from value
						if idx := strings.Index(header.Value, "//"); idx >= 0 {
							header.Value = strings.TrimSpace(header.Value[:idx])
						}
					}
				}

				// Special handling for imports
				if header.Type == Import && matches[0] == "import (" {
					inImportBlock = true
					currentBlock = &header
					continue
				}

				// For struct and class types, extract fields
				if header.Type == Struct || header.Type == Class {
					extractMembers(&header, lineNum, lines)
				}

				headers = append(headers, header)
				break // Stop after first matching pattern
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	return headers, nil
}

// parseParametersWithTypes splits a parameter string into individual parameters with type information
func parseParametersWithTypes(paramStr string, language string) []ParameterInfo {
	if strings.TrimSpace(paramStr) == "" {
		return nil
	}

	// Different parsing based on language
	switch language {
	case "go":
		// Handle Go's "name1, name2 type" format
		var params []ParameterInfo
		paramGroups := splitParamsRespectingBrackets(paramStr)

		for _, group := range paramGroups {
			group = strings.TrimSpace(group)
			if group == "" {
				continue
			}

			// Handle Go's "name1, name2 type" format
			parts := strings.Fields(group)
			if len(parts) >= 2 {
				// Last part is the type
				paramType := parts[len(parts)-1]

				// If there are more than 2 parts or the first part contains commas,
				// we have multiple parameters of the same type
				if len(parts) > 2 || strings.Contains(parts[0], ",") {
					nameList := strings.Join(parts[:len(parts)-1], " ")
					names := strings.Split(nameList, ",")
					for _, name := range names {
						name = strings.TrimSpace(name)
						if name != "" {
							params = append(params, ParameterInfo{
								Name: name,
								Type: paramType,
							})
						}
					}
				} else {
					// Simple case: one parameter with its type
					params = append(params, ParameterInfo{
						Name: parts[0],
						Type: paramType,
					})
				}
			} else if len(parts) == 1 {
				// Handle unnamed parameter (type only)
				params = append(params, ParameterInfo{
					Name: "",
					Type: parts[0],
				})
			}
		}
		return params

	case "python":
		// Handle Python parameter syntax including type hints
		var params []ParameterInfo
		paramGroups := splitParamsRespectingBrackets(paramStr)

		for _, group := range paramGroups {
			group = strings.TrimSpace(group)
			if group == "" {
				continue
			}

			// Check for type annotations (param: type)
			if strings.Contains(group, ":") {
				parts := strings.SplitN(group, ":", 2)
				paramName := strings.TrimSpace(parts[0])
				paramType := ""
				if len(parts) > 1 {
					paramType = strings.TrimSpace(parts[1])
				}

				// Handle default values
				if strings.Contains(paramName, "=") {
					nameParts := strings.SplitN(paramName, "=", 2)
					paramName = strings.TrimSpace(nameParts[0])
				}

				params = append(params, ParameterInfo{
					Name: paramName,
					Type: paramType,
				})
			} else {
				// Parameter without type annotation
				paramName := group

				// Handle default values
				if strings.Contains(paramName, "=") {
					nameParts := strings.SplitN(paramName, "=", 2)
					paramName = strings.TrimSpace(nameParts[0])
				}

				params = append(params, ParameterInfo{
					Name: paramName,
					Type: "",
				})
			}
		}
		return params

	case "javascript", "typescript":
		// Handle JS/TS parameter syntax
		var params []ParameterInfo
		paramGroups := splitParamsRespectingBrackets(paramStr)

		for _, group := range paramGroups {
			group = strings.TrimSpace(group)
			if group == "" {
				continue
			}

			// Check for TypeScript type annotations (param: type)
			if strings.Contains(group, ":") {
				parts := strings.SplitN(group, ":", 2)
				paramName := strings.TrimSpace(parts[0])
				paramType := ""
				if len(parts) > 1 {
					paramType = strings.TrimSpace(parts[1])
				}
				params = append(params, ParameterInfo{
					Name: paramName,
					Type: paramType,
				})
			} else {
				// JavaScript without type annotation
				params = append(params, ParameterInfo{
					Name: group,
					Type: "",
				})
			}
		}
		return params

	default:
		// Generic handling for other languages
		var params []ParameterInfo
		paramList := strings.Split(paramStr, ",")

		for _, p := range paramList {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			// Try to separate name and type based on common patterns
			// This is a simplistic approach, would need language-specific handling for best results
			parts := strings.Fields(p)
			if len(parts) >= 2 {
				// Assume last part is type, rest is name (not perfect but common pattern)
				paramName := strings.Join(parts[:len(parts)-1], " ")
				paramType := parts[len(parts)-1]
				params = append(params, ParameterInfo{
					Name: paramName,
					Type: paramType,
				})
			} else if len(parts) == 1 {
				// Just a name without explicit type
				params = append(params, ParameterInfo{
					Name: parts[0],
					Type: "",
				})
			}
		}
		return params
	}
}

// splitParamsRespectingBrackets splits parameters by comma while respecting brackets
func splitParamsRespectingBrackets(paramStr string) []string {
	var params []string
	depth := 0
	start := 0

	for i, char := range paramStr {
		switch char {
		case '<', '(', '[', '{':
			depth++
		case '>', ')', ']', '}':
			depth--
		case ',':
			if depth == 0 {
				params = append(params, paramStr[start:i])
				start = i + 1
			}
		}
	}

	// Add the last parameter
	if start < len(paramStr) {
		params = append(params, paramStr[start:])
	}

	return params
}

// parseReturnTypes extracts return type information
func parseReturnTypes(returnStr string, language string) []string {
	returnStr = strings.TrimSpace(returnStr)
	if returnStr == "" {
		return nil
	}

	// Different parsing based on language
	switch language {
	case "go":
		// Handle Go's multiple return values
		// If it starts with (, it's multiple return values
		if strings.HasPrefix(returnStr, "(") && strings.HasSuffix(returnStr, ")") {
			// Remove parentheses and split by comma
			returnStr = returnStr[1 : len(returnStr)-1]
			returns := strings.Split(returnStr, ",")
			for i := range returns {
				returns[i] = strings.TrimSpace(returns[i])
			}
			return returns
		}
		// Otherwise it's a single return value
		return []string{returnStr}
	default:
		// For most languages, just return the trimmed string
		return []string{returnStr}
	}
}

// extractMembers finds the members/fields of a block element (struct, class, etc.)
func extractMembers(header *HeaderElement, startLine int, lines []string) {
	if startLine >= len(lines) {
		return
	}

	// Find opening brace
	braceCount := 0
	insideBlock := false
	for i := startLine - 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if !insideBlock && strings.Contains(line, "{") {
			insideBlock = true
		}

		if insideBlock {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			// Extract field
			if strings.Contains(line, ":") || strings.Contains(line, " ") {
				field := extractField(line)
				if field != "" {
					header.Children = append(header.Children, HeaderElement{
						Type:    Field,
						Name:    field,
						LineNum: i + 1,
					})
				}
			}

			if braceCount == 0 {
				header.EndLine = i + 1
				break
			}
		}
	}
}

// extractField attempts to extract a field name from a struct field line
func extractField(line string) string {
	// Skip comments and empty lines
	if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.TrimSpace(line) == "" {
		return ""
	}

	// Remove trailing comments
	if idx := strings.Index(line, "//"); idx >= 0 {
		line = line[:idx]
	}

	line = strings.TrimSpace(line)

	// Try to extract field name
	parts := strings.Fields(line)
	if len(parts) > 0 {
		// Remove any punctuation/symbols
		fieldName := strings.Trim(parts[0], ":,")
		// Skip if it's just a bracket or common keyword
		if fieldName != "{" && fieldName != "}" && fieldName != "struct" && fieldName != "class" {
			return fieldName
		}
	}

	return ""
}

// FormatHeaders formats the extracted headers into a readable string
func FormatHeaders(headers []HeaderElement) string {
	var builder strings.Builder

	for _, header := range headers {
		// Format based on the header type
		switch header.Type {
		case Function, Method:
			// Format function/method with parameters and return type
			var signature strings.Builder
			if header.Parent != "" {
				signature.WriteString(fmt.Sprintf("%s.%s(", header.Parent, header.Name))
			} else {
				signature.WriteString(fmt.Sprintf("%s(", header.Name))
			}

			// Add parameters with types
			if len(header.Parameters) > 0 {
				paramStrs := make([]string, len(header.Parameters))
				for i, p := range header.Parameters {
					if p.Type != "" {
						if p.Name != "" {
							paramStrs[i] = fmt.Sprintf("%s %s", p.Name, p.Type)
						} else {
							paramStrs[i] = p.Type
						}
					} else {
						paramStrs[i] = p.Name
					}
				}
				signature.WriteString(strings.Join(paramStrs, ", "))
			}
			signature.WriteString(")")

			// Add return types
			if len(header.ReturnTypes) > 0 {
				if len(header.ReturnTypes) == 1 {
					signature.WriteString(" " + header.ReturnTypes[0])
				} else {
					signature.WriteString(" (")
					signature.WriteString(strings.Join(header.ReturnTypes, ", "))
					signature.WriteString(")")
				}
			}

			if header.Scope != "" {
				builder.WriteString(fmt.Sprintf("Line %d: %s %s: %s\n",
					header.LineNum, header.Scope, header.Type, signature.String()))
			} else {
				builder.WriteString(fmt.Sprintf("Line %d: %s: %s\n",
					header.LineNum, header.Type, signature.String()))
			}

		case Constant, Variable:
			// Include type and value information for constants/variables
			var valueInfo string
			if header.ValueType != "" && header.Value != "" {
				valueInfo = fmt.Sprintf("%s = %s", header.ValueType, header.Value)
			} else if header.ValueType != "" {
				valueInfo = header.ValueType
			} else if header.Value != "" {
				valueInfo = fmt.Sprintf("= %s", header.Value)
			} else {
				valueInfo = header.Name
			}

			if header.Scope != "" {
				builder.WriteString(fmt.Sprintf("Line %d: %s %s: %s %s\n",
					header.LineNum, header.Scope, header.Type, header.Name, valueInfo))
			} else {
				builder.WriteString(fmt.Sprintf("Line %d: %s: %s %s\n",
					header.LineNum, header.Type, header.Name, valueInfo))
			}

		case Import:
			builder.WriteString(fmt.Sprintf("Line %d: %s:\n", header.LineNum, header.Type))
			if len(header.Children) > 0 {
				for _, child := range header.Children {
					if child.Parent != "" {
						if child.Signature != "" {
							builder.WriteString(fmt.Sprintf("    from %s import %s as %s\n",
								child.Parent, child.Name, child.Signature))
						} else {
							builder.WriteString(fmt.Sprintf("    from %s import %s\n",
								child.Parent, child.Name))
						}
					} else {
						if child.Signature != "" {
							builder.WriteString(fmt.Sprintf("    import %s as %s\n",
								child.Name, child.Signature))
						} else {
							builder.WriteString(fmt.Sprintf("    import %s\n", child.Name))
						}
					}
				}
			} else {
				builder.WriteString(fmt.Sprintf("    %s\n", header.Name))
			}

		case Struct, Class:
			if header.Scope != "" {
				builder.WriteString(fmt.Sprintf("Line %d: %s %s: %s\n", header.LineNum, header.Scope, header.Type, header.Name))
			} else {
				builder.WriteString(fmt.Sprintf("Line %d: %s: %s\n", header.LineNum, header.Type, header.Name))
			}
			for _, child := range header.Children {
				builder.WriteString(fmt.Sprintf("    %s\n", child.Name))
			}

		default:
			if header.Scope != "" {
				builder.WriteString(fmt.Sprintf("Line %d: %s %s: %s\n", header.LineNum, header.Scope, header.Type, header.Name))
			} else {
				builder.WriteString(fmt.Sprintf("Line %d: %s: %s\n", header.LineNum, header.Type, header.Name))
			}
		}
	}

	return builder.String()
}

// CollectHeadersAsString is a convenience function to get formatted headers directly
func CollectHeadersAsString(path string) (string, error) {
	headers, err := CollectHeaders(path)
	if err != nil {
		return "", err
	}

	if len(headers) == 0 {
		return "No headers found.", nil
	}

	return FormatHeaders(headers), nil
}
