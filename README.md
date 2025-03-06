# Codeclip

Codeclip is a command-line utility that helps developers quickly capture and share code snippets across their codebase. It provides powerful search capabilities and formats code for easy pasting into documentation, chat applications, or AI assistants.

## Features

- **Search-based extraction**: Find and extract code based on patterns or text matches
- **Glob pattern matching**: Select files using glob patterns to include in your snippet collection
- **Context-aware**: Include surrounding code lines for better understanding
- **Function extraction**: Capture entire functions or methods that contain your search terms
- **Token estimation**: Estimate token counts for AI assistants like GPT models
- **Multiple output targets**: Copy to clipboard, print to stdout, or save to a file
- **Code syntax highlighting**: Formats output with language-specific syntax highlighting markers
- **Fuzzy search**: Flexible matching for when you can't remember exact text
- **Template processing**: Embed code snippets directly within a markdown/template file using special tags

## Installation

### Using Go

```bash
go install github.com/grant-wade/codeclip@latest
```

### From Source

```bash
git clone https://github.com/grant-wade/codeclip.git
cd codeclip
go install .
```

## Usage

Codeclip offers three main commands: `search`, `glob`, and `template`.

### Search Command

Search for specific code patterns across your codebase:

```bash
codeclip search "func GetUser" --function
```

This command searches for the pattern "func GetUser" and extracts the entire function containing this text.

Options:
- `--context, -c`: Number of context lines to include (default: 3)
- `--function, -f`: Include entire function/method containing matches
- `--fuzzy, -z`: Enable fuzzy matching for search terms
- `--output, -o`: Output destination (clipboard, stdout, or file path)
- `--estimate, -e`: Estimate token count in output (default: true)
- `--max-tokens, -m`: Maximum tokens to copy (0 for unlimited)
- `--path, -p`: Path to search in (default: current directory)

### Glob Command

Select files using a glob pattern and extract their contents:

```bash
codeclip glob "**/*.go"
```

This command finds all Go files in the current directory and subdirectories.

Options are the same as for the search command.

### Template Command

Process a template file with embedded content tags to automatically include file or glob content into your documentation.

#### Template Syntax

Within your template file, you can include tags in the following formats:
- **Direct file reference**:  
  `{{filename.txt}}`  
  This tag will be replaced with the contents of `filename.txt` formatted as a code block.
  
- **Glob pattern**:  
  `{{**/*.go}}`  
  This tag will be replaced with the contents of all Go files matching the glob pattern, each formatted as a separate code block.

#### Example Template

Create a file named `documentation-tempalte.md` with the following content:

```markdown
# Project Overview

{{README.md}}

## Requirements
{{go.mod}}

## Code
{{**/*.go}}
```

#### Running the Template Command

To process the template and output the result to a file or the clipboard:

```bash
codeclip template documentation-template.md --output documentation.md
```

This command processes `documentation-template.md`, replaces the tags with the corresponding file contents, and writes the output to `documentation.md`.

## Use Cases

- **Sharing code with teammates**: Quickly copy relevant sections of your codebase.
- **AI Assistant Interaction**: Format code snippets properly for AI assistants like ChatGPT.
- **Documentation**: Extract and embed code examples directly into your technical documentation using the template functionality.
- **Code Review**: Share specific portions of code during review discussions.

## Token Estimation

When working with AI models like GPT-4, token limits are important. Codeclip provides an estimation of the number of tokens in the copied content, helping you stay within model limits.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[BSD-3-Clause](LICENSE)