package output

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// SearchResult represents a single search match with all required information for output.
type SearchResult struct {
	Repository  string // "owner/repo" format
	LocalPath   string // e.g., "src/main.go:12"
	MatchedLine string // The matched line content
	GitHubURL   string // Full GitHub URL with line number
}

// WriteTSV writes search results in TSV format to the given writer.
// Format: {Repository}\t{LocalPath}\t{MatchedLine}\t{GitHubURL}\n
func WriteTSV(results []SearchResult, writer io.Writer) error {
	w := bufio.NewWriter(writer)
	defer w.Flush()

	for _, result := range results {
		// Sanitize matched line: replace tabs and newlines with spaces
		sanitized := sanitizeLine(result.MatchedLine)

		// Write TSV line
		line := fmt.Sprintf("%s\t%s\t%s\t%s\n",
			result.Repository,
			result.LocalPath,
			sanitized,
			result.GitHubURL)

		if _, err := w.WriteString(line); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

// sanitizeLine replaces tabs and newlines with spaces to preserve TSV structure.
func sanitizeLine(text string) string {
	// Replace tabs with spaces
	text = strings.ReplaceAll(text, "\t", " ")
	// Replace newlines with spaces
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)
	return text
}
