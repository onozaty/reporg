package search

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
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
	Text *string `json:"text,omitempty"`
}

// SearchRepo executes ripgrep search on the given repository and returns all matches.
func SearchRepo(pattern, repoRoot string) ([]Match, error) {
	// Check if ripgrep is installed
	if _, err := exec.LookPath("rg"); err != nil {
		return nil, fmt.Errorf("ripgrep not found: please install ripgrep")
	}

	// Execute: rg --json <pattern> <repoRoot>
	cmd := exec.Command("rg", "--json", pattern, repoRoot)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ripgrep: %w", err)
	}

	var matches []Match
	scanner := bufio.NewScanner(stdout)

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

		// Extract line text (skip if not UTF-8 text)
		if matchData.Lines.Text == nil {
			continue
		}
		lineText := *matchData.Lines.Text

		matches = append(matches, Match{
			RelPath:    relPath,
			LineNumber: matchData.LineNumber,
			LineText:   lineText,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ripgrep output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		// Exit code 1 means no matches found, which is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return matches, nil // Return empty matches
		}
		return nil, fmt.Errorf("ripgrep failed: %w", err)
	}

	return matches, nil
}
