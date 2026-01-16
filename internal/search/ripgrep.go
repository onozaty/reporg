package search

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Match represents a single search match result.
type Match struct {
	RelPath    string // Relative path from repository root
	LineNumber int    // Line number (1-indexed)
	LineText   string // The matched line content
}

// RipgrepMessage represents a single JSON message from ripgrep's --json output.
type RipgrepMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// MatchData represents the data field of a "match" type message from ripgrep.
type MatchData struct {
	Path       PathData `json:"path"`
	Lines      TextData `json:"lines"`
	LineNumber int      `json:"line_number"`
}

// PathData represents the path information in ripgrep JSON output.
type PathData struct {
	Text *string `json:"text,omitempty"`
}

// TextData represents text data that can be either UTF-8 text or base64 bytes.
type TextData struct {
	Text  *string `json:"text,omitempty"`
	Bytes *string `json:"bytes,omitempty"` // Base64-encoded bytes for non-UTF-8 content
}

// SearchOptions contains optional parameters for ripgrep search.
type SearchOptions struct {
	IgnoreCase    bool     // Enable case-insensitive search (-i)
	Globs         []string // Glob patterns to filter files (--glob)
	Hidden        bool     // Search hidden files and directories (--hidden)
	FixedStrings  bool     // Treat pattern as literal string, not regex (-F)
	MaxLineLength int      // Maximum length of line text in output (0 = no limit)
	Encoding      string   // Text encoding to use (--encoding, default: auto)
}

// SearchRepo executes ripgrep search on the given repository.
// The onMatch callback is called for each match found.
func SearchRepo(pattern, repoRoot string, opts SearchOptions, onMatch func(Match) error) error {
	// Check if ripgrep is installed
	if _, err := exec.LookPath("rg"); err != nil {
		return fmt.Errorf("ripgrep not found: please install ripgrep from https://github.com/BurntSushi/ripgrep#installation")
	}

	// Build ripgrep arguments
	args := []string{"--json"}

	// Add case-insensitive flag if requested
	if opts.IgnoreCase {
		args = append(args, "-i")
	}

	// Add glob patterns if specified
	for _, glob := range opts.Globs {
		args = append(args, "--glob", glob)
	}

	// Add hidden flag if requested
	if opts.Hidden {
		args = append(args, "--hidden")
	}

	// Add fixed-strings flag if requested
	if opts.FixedStrings {
		args = append(args, "-F")
	}

	// Add encoding flag if specified
	if opts.Encoding != "" {
		args = append(args, "--encoding", opts.Encoding)
	}

	// Add pattern and path
	args = append(args, pattern, repoRoot)

	// Execute: rg --json [options] <pattern> <repoRoot>
	cmd := exec.Command("rg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ripgrep: %w", err)
	}

	scanner := bufio.NewScanner(stdout)

	// Increase buffer size to handle large JSON lines (default is 64KB, set to 10MB)
	// This allows processing of very long lines (e.g., minified JavaScript) without errors
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	// Process each line of JSON output
	for scanner.Scan() {
		line := scanner.Bytes()

		var msg RipgrepMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue // Skip invalid JSON lines
		}

		// Only process "match" type messages
		if msg.Type != "match" {
			continue
		}

		var matchData MatchData
		if err := json.Unmarshal(msg.Data, &matchData); err != nil {
			continue // Skip if we can't parse match data
		}

		// Extract path (skip if not UTF-8 text)
		if matchData.Path.Text == nil {
			continue
		}
		absPath := *matchData.Path.Text

		// Convert to relative path from repository root
		relPath, err := filepath.Rel(repoRoot, absPath)
		if err != nil {
			relPath = absPath // Fall back to absolute path if conversion fails
		}

		// Extract line text
		// ripgrep uses "text" field for UTF-8 content and "bytes" field for non-UTF-8 content.
		// When using --encoding with non-UTF-8 encodings, ripgrep may output base64-encoded bytes.
		lineText := ""
		if matchData.Lines.Text != nil {
			// UTF-8 text content
			lineText = *matchData.Lines.Text
		} else if matchData.Lines.Bytes != nil {
			// Base64-encoded bytes (for non-UTF-8 content)
			// Decode base64 to get the original bytes, then convert to string
			if decoded, err := base64.StdEncoding.DecodeString(*matchData.Lines.Bytes); err == nil {
				lineText = string(decoded)
			}
		}

		// Remove trailing newline characters (LF, CRLF, CR)
		lineText = strings.TrimRight(lineText, "\r\n")

		// Truncate line text if MaxLineLength is specified and line exceeds the limit
		if opts.MaxLineLength > 0 && len(lineText) > opts.MaxLineLength {
			lineText = lineText[:opts.MaxLineLength] + "..."
		}

		match := Match{
			RelPath:    relPath,
			LineNumber: matchData.LineNumber,
			LineText:   lineText,
		}

		// Call the callback
		if err := onMatch(match); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading ripgrep output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		// Exit code 1 means no matches found, which is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("ripgrep failed: %w", err)
	}

	return nil
}
