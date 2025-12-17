# treecat

[![GitHub license](https://img.shields.io/github/license/onozaty/treecat)](https://github.com/onozaty/treecat/blob/main/LICENSE)
[![Test](https://github.com/onozaty/treecat/actions/workflows/test.yaml/badge.svg)](https://github.com/onozaty/treecat/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/onozaty/treecat/graph/badge.svg?token=V114XLWDBR)](https://codecov.io/gh/onozaty/treecat)


> Combine multiple files into a single output with directory tree visualization, optimized for LLM consumption

**treecat** is a command-line tool that aggregates multiple files from a directory into a single output, displaying a directory tree structure at the top followed by the contents of each file. It's specifically designed to provide comprehensive codebase context to Large Language Models (LLMs) like Claude Code, ChatGPT, and GitHub Copilot.

Perfect for feeding entire project structures to AI assistants, code reviews, documentation generation, and any scenario where you need a unified view of multiple files with their directory hierarchy.

## Features

- **Directory tree visualization with file content aggregation** - Single output combining tree structure and file contents
- **Optimized for LLM context** - Ideal for Claude Code, ChatGPT, GitHub Copilot, and other AI coding assistants
- **Automatic .gitignore pattern application** - Respects your project's .gitignore rules by default
- **Flexible glob pattern filtering** - Include/exclude files with powerful `--include` and `--exclude` patterns
- **Character encoding conversion** - Convert non-UTF-8 encodings (Shift_JIS, EUC-JP, GB2312, etc.) to UTF-8
- **UTF-8 BOM removal and line ending normalization** - Ensures consistent output (CRLF → LF)
- **Empty directory pruning** - Automatically excludes empty directories after filtering

## Installation

Download the latest binary from [GitHub Releases](https://github.com/onozaty/treecat/releases/latest).

## Quick Start

```bash
# Output current directory to file
treecat . > output.txt

# Output specific directory
treecat /path/to/project > output.txt

# Filter by file patterns
treecat . --include "**/*.go,**/*.md" > output.txt

# Exclude patterns
treecat . --exclude "*.png,*.jpg,node_modules/**" > output.txt
```

## Usage

### Basic Usage

```bash
treecat [directory] [flags]
```

If no directory is specified, the current directory (`.`) is used.

### Command-line Options

#### Filtering Options

**`-i, --include <patterns>`**

Include only files matching the specified glob patterns (comma-separated). When specified, only files matching these patterns will be included in the output.

```bash
treecat . --include "**/*.go,**/*.md"
```

**`-e, --exclude <patterns>`**

Exclude files matching the specified glob patterns (comma-separated). These patterns are applied after .gitignore rules.

```bash
treecat . --exclude "*.log,*.tmp,dist/**"
```

**`--no-gitignore`**

Ignore `.gitignore` file patterns. By default, treecat automatically applies `.gitignore` rules found in the target directory.

```bash
treecat . --no-gitignore
```

#### Encoding Options

**`--encoding-map <extension:encoding,...>`**

Specify character encoding per file extension. Files matching the specified extensions are converted from their original encoding to UTF-8 in the output.

**Supported encodings** (IANA standard names):
- **Japanese**: `shift_jis`, `euc-jp`, `iso-2022-jp`
- **Western European**: `windows-1252`, `iso-8859-1`, `iso-8859-15`
- **Chinese**: `gb2312`, `gbk`, `big5`
- **Korean**: `euc-kr`
- And more via `golang.org/x/text/encoding/htmlindex`

**Notes:**
- Encoding names are case-insensitive (`Shift_JIS`, `shift-jis`, `SHIFT_JIS` are all equivalent)
- Underscores and hyphens are interchangeable (`shift_jis` = `shift-jis`)
- Extensions are specified without leading dots and are case-insensitive (`txt`, `TXT`, `.txt` all work)

```bash
# Single extension
treecat . --encoding-map "txt:shift_jis"

# Multiple extensions with different encodings
treecat legacy/ --encoding-map "txt:shift_jis,log:euc-jp,csv:windows-1252"
```

**Encoding behavior:**
- **BOM (Byte Order Mark) removal**: UTF-8 BOM is automatically removed from all files
- **Line ending normalization**: All line endings (CRLF, CR, LF) are converted to LF (`\n`)
- **Unmapped files**: Files not in the encoding map are treated as UTF-8

### Filtering Priority

Filters are applied in the following order:

1. **`.git/` directory exclusion** - Always applied (highest priority)
2. **`.gitignore` patterns** - Applied unless `--no-gitignore` is specified
3. **`--exclude` patterns** - User-specified exclusion patterns
4. **`--include` patterns** - If specified, only matching files are included (lowest priority)

### Examples

#### Filter by file type

```bash
# Only Go source files
treecat . --include "**/*.go" > go-code.txt

# Source code and documentation
treecat . --include "**/*.go,**/*.md,**/*.txt" > project.txt
```

#### Exclude files and directories

```bash
# Exclude image files
treecat . --exclude "*.png,*.jpg,*.gif,*.svg" > output.txt

# Exclude common build/dependency directories
treecat . --exclude "vendor/**,node_modules/**,dist/**,.next/**" > output.txt
```

#### Combine include and exclude

```bash
# Go files excluding tests
treecat . --include "**/*.go" --exclude "**/*_test.go" > main-code.txt

# Python source excluding virtual environments
treecat . --include "**/*.py" --exclude "venv/**,**/__pycache__/**" > python-code.txt
```

#### Character encoding conversion

```bash
# Convert Shift_JIS encoded .txt files to UTF-8
treecat legacy-docs/ --encoding-map "txt:shift_jis" > output.txt

# Multiple encodings for different file types
treecat old-project/ --encoding-map "txt:shift_jis,log:euc-jp,csv:windows-1252" > output.txt

# Japanese legacy codebase with mixed encodings
treecat src/ --encoding-map "txt:shift_jis,md:euc-jp" > codebase.txt
```

## Output Format

The output consists of two main sections:

```
project-root/
├── file1.go
├── README.md
└── src/
    └── main.go

=== file1.go ===
package main

func main() {
    // ...
}

=== README.md ===
# Project

Documentation here...

=== src/main.go ===
package main

// Additional code...
```

**1. Directory tree** (top section):
- Shows the hierarchical structure with the root directory name
- Uses Unicode box-drawing characters (`├──`, `└──`, `│`)
- Empty directories (after filtering) are automatically excluded
- Directories are displayed before files at the same level
- Lexicographic sorting within each level

**2. File contents** (below tree):
- Each file is separated by `=== filepath ===` markers
- File paths are relative to the specified root directory
- Files appear in lexicographic order by relative path

## License

MIT License

See [LICENSE](LICENSE) file for details.

## Author

[onozaty](https://github.com/onozaty)
