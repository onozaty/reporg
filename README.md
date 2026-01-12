# reporg

[![Test](https://github.com/onozaty/reporg/actions/workflows/test.yaml/badge.svg)](https://github.com/onozaty/reporg/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/onozaty/reporg/branch/main/graph/badge.svg)](https://codecov.io/gh/onozaty/reporg)
[![GitHub release](https://img.shields.io/github/release/onozaty/reporg.svg)](https://github.com/onozaty/reporg/releases/latest)
[![License](https://img.shields.io/github/license/onozaty/reporg.svg)](LICENSE)

English | [日本語](README.ja.md)

**reporg** is a CLI tool that searches Git repositories using [ripgrep](https://github.com/BurntSushi/ripgrep) and outputs search results in an easy-to-share format (TSV).

## Features

- **Local Search**: Fast full-text search using ripgrep
- **GitHub URL Generation**: Automatically generates GitHub URLs for each search result line
- **TSV Format Output**: Easy to process with spreadsheets and other tools
- **Multiple Repository Support**: Search multiple repositories at once
- **Rich Search Options**: Case-insensitive search, glob patterns, hidden file search, fixed string search, and more

## Installation

### Go install

```bash
go install github.com/onozaty/reporg@latest
```

### Binary Download

Download the binary for your platform from the [Releases](https://github.com/onozaty/reporg/releases) page.

## Prerequisites

To use reporg, you need to have [ripgrep](https://github.com/BurntSushi/ripgrep) installed.

```bash
# macOS
brew install ripgrep

# Ubuntu/Debian
sudo apt-get install ripgrep

# Windows (Chocolatey)
choco install ripgrep
```

For more details, see the [ripgrep installation guide](https://github.com/BurntSushi/ripgrep#installation).

## Search Behavior

reporg uses ripgrep for searching. The following files and directories are automatically skipped:

- Files and directories listed in `.gitignore`
- Files and directories listed in `.ignore` or `.rgignore`
- Hidden files and directories (those starting with `.`) - can be included with the `--hidden` option

## Usage

### Basic Usage

```bash
reporg <pattern> <repoRoot1> [repoRoot2...]
```

- `pattern`: Search pattern (regular expression)
- `repoRoot`: Git repository root directory (multiple can be specified)

**Examples:**

```bash
# Search for "TODO" in current directory
reporg "TODO" .

# Search multiple repositories
reporg "TODO" /path/to/repo1 /path/to/repo2
```

### Output Format

Search results are output in TSV (tab-separated values) format:

```
owner/repo	src/main.go:12	// TODO: refactor	https://github.com/owner/repo/blob/main/src/main.go#L12
owner/repo	src/utils.go:25	// TODO: optimize	https://github.com/owner/repo/blob/main/src/utils.go#L25
```

**Columns (tab-separated):**

1. `repository`: Repository identifier in `owner/repo` format
2. `local_path`: File path and line number (`path/to/file:LINE` format)
3. `matched_line`: Content of the matched line
4. `github_url`: GitHub URL to the corresponding line

### Options

#### Output Destination

```bash
# Output to file
reporg "TODO" . -o result.tsv
```

#### Search Options

**Case-insensitive search:**

```bash
# Matches "TODO", "todo", "Todo", etc.
reporg -i "todo" /path/to/repo
```

**Filter files with glob patterns:**

```bash
# Search only Go files
reporg "package" /repo -g "*.go"

# Search Go files but exclude test files
reporg "package" /repo -g "*.go" -g "!*_test.go"

# Search only specific directory
reporg "TODO" /repo -g "src/**"
```

**Search hidden files:**

```bash
# Include hidden files in search
reporg "secret" /repo --hidden

# Search hidden files but exclude .git directory
reporg "config" /repo --hidden -g "!.git/**"
```

**Fixed string search (not regex):**

```bash
# Search "main()" as a literal string, not regex
reporg -F "main()" /repo

# Search for patterns containing parentheses
reporg -F "if (x > 0) {" /repo
```

**Combining options:**

```bash
# Case-insensitive search for "TODO" in Go files, save to file
reporg -i "todo" /repo -g "*.go" -o results.tsv

# Fixed string search including hidden files
reporg -F "config.value" /repo --hidden -o config-usage.tsv

# Search multiple repositories with multiple conditions
reporg -i "error" /repo1 /repo2 -g "*.go" -g "!vendor/**" --hidden
```

### All Options

```
  -o, --output string         Output file path (default: stdout)
  -i, --ignore-case           Case-insensitive search
  -g, --glob pattern          Filter files by glob pattern (can be specified multiple times)
      --hidden                Include hidden files and directories in search
  -F, --fixed-strings         Treat pattern as literal string, not regex
  -h, --help                  Show help
  -v, --version               Show version information
```

## Limitations

- **GitHub only**: Currently only GitHub repositories are supported
- **Git repository root required**: Specified paths must be Git repository root directories (subdirectories will cause an error)

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Author

[onozaty](https://github.com/onozaty)
