package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTSV_SingleResult(t *testing.T) {
	results := []SearchResult{
		{
			Repository:  "owner/repo",
			LocalPath:   "main.go:10",
			MatchedLine: "package main",
			GitHubURL:   "https://github.com/owner/repo/blob/main/main.go#L10",
		},
	}

	var buf bytes.Buffer
	err := WriteTSV(results, &buf)
	if err != nil {
		t.Fatalf("WriteTSV() error = %v, want nil", err)
	}

	want := "owner/repo\tmain.go:10\tpackage main\thttps://github.com/owner/repo/blob/main/main.go#L10\n"
	got := buf.String()

	if got != want {
		t.Errorf("WriteTSV() = %q, want %q", got, want)
	}
}

func TestWriteTSV_MultipleResults(t *testing.T) {
	results := []SearchResult{
		{
			Repository:  "owner/repo",
			LocalPath:   "main.go:10",
			MatchedLine: "package main",
			GitHubURL:   "https://github.com/owner/repo/blob/main/main.go#L10",
		},
		{
			Repository:  "owner/repo",
			LocalPath:   "cmd/root.go:25",
			MatchedLine: "func Execute() error {",
			GitHubURL:   "https://github.com/owner/repo/blob/main/cmd/root.go#L25",
		},
	}

	var buf bytes.Buffer
	err := WriteTSV(results, &buf)
	if err != nil {
		t.Fatalf("WriteTSV() error = %v, want nil", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("WriteTSV() wrote %d lines, want 2", len(lines))
	}

	// Check first line
	want1 := "owner/repo\tmain.go:10\tpackage main\thttps://github.com/owner/repo/blob/main/main.go#L10"
	if lines[0] != want1 {
		t.Errorf("WriteTSV() line 1 = %q, want %q", lines[0], want1)
	}

	// Check second line
	want2 := "owner/repo\tcmd/root.go:25\tfunc Execute() error {\thttps://github.com/owner/repo/blob/main/cmd/root.go#L25"
	if lines[1] != want2 {
		t.Errorf("WriteTSV() line 2 = %q, want %q", lines[1], want2)
	}
}

func TestWriteTSV_EmptyResults(t *testing.T) {
	results := []SearchResult{}

	var buf bytes.Buffer
	err := WriteTSV(results, &buf)
	if err != nil {
		t.Fatalf("WriteTSV() error = %v, want nil", err)
	}

	if buf.Len() != 0 {
		t.Errorf("WriteTSV() wrote %d bytes, want 0", buf.Len())
	}
}

func TestWriteTSV_TabsInMatchedLine(t *testing.T) {
	results := []SearchResult{
		{
			Repository:  "owner/repo",
			LocalPath:   "test.go:5",
			MatchedLine: "key\tvalue\tdata",
			GitHubURL:   "https://github.com/owner/repo/blob/main/test.go#L5",
		},
	}

	var buf bytes.Buffer
	err := WriteTSV(results, &buf)
	if err != nil {
		t.Fatalf("WriteTSV() error = %v, want nil", err)
	}

	// Tabs in matched line should be replaced with spaces
	want := "owner/repo\ttest.go:5\tkey value data\thttps://github.com/owner/repo/blob/main/test.go#L5\n"
	got := buf.String()

	if got != want {
		t.Errorf("WriteTSV() = %q, want %q", got, want)
	}
}

func TestWriteTSV_NewlinesInMatchedLine(t *testing.T) {
	results := []SearchResult{
		{
			Repository:  "owner/repo",
			LocalPath:   "test.go:5",
			MatchedLine: "line1\nline2\rline3",
			GitHubURL:   "https://github.com/owner/repo/blob/main/test.go#L5",
		},
	}

	var buf bytes.Buffer
	err := WriteTSV(results, &buf)
	if err != nil {
		t.Fatalf("WriteTSV() error = %v, want nil", err)
	}

	// Newlines in matched line should be replaced with spaces
	want := "owner/repo\ttest.go:5\tline1 line2 line3\thttps://github.com/owner/repo/blob/main/test.go#L5\n"
	got := buf.String()

	if got != want {
		t.Errorf("WriteTSV() = %q, want %q", got, want)
	}
}

func TestWriteTSV_LeadingTrailingWhitespace(t *testing.T) {
	results := []SearchResult{
		{
			Repository:  "owner/repo",
			LocalPath:   "test.go:5",
			MatchedLine: "  \t  content with spaces  \t  ",
			GitHubURL:   "https://github.com/owner/repo/blob/main/test.go#L5",
		},
	}

	var buf bytes.Buffer
	err := WriteTSV(results, &buf)
	if err != nil {
		t.Fatalf("WriteTSV() error = %v, want nil", err)
	}

	// Leading/trailing whitespace should be trimmed, internal tabs replaced
	want := "owner/repo\ttest.go:5\tcontent with spaces\thttps://github.com/owner/repo/blob/main/test.go#L5\n"
	got := buf.String()

	if got != want {
		t.Errorf("WriteTSV() = %q, want %q", got, want)
	}
}

func TestSanitizeLine_Tabs(t *testing.T) {
	input := "hello\tworld\ttab"
	want := "hello world tab"
	got := sanitizeLine(input)

	if got != want {
		t.Errorf("sanitizeLine(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeLine_Newlines(t *testing.T) {
	input := "line1\nline2\rline3\r\nline4"
	want := "line1 line2 line3  line4"
	got := sanitizeLine(input)

	if got != want {
		t.Errorf("sanitizeLine(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeLine_Whitespace(t *testing.T) {
	input := "  \t  content  \t  "
	want := "content"
	got := sanitizeLine(input)

	if got != want {
		t.Errorf("sanitizeLine(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeLine_Empty(t *testing.T) {
	input := ""
	want := ""
	got := sanitizeLine(input)

	if got != want {
		t.Errorf("sanitizeLine(%q) = %q, want %q", input, got, want)
	}
}

func TestSanitizeLine_OnlyWhitespace(t *testing.T) {
	input := "  \t\n\r  "
	want := ""
	got := sanitizeLine(input)

	if got != want {
		t.Errorf("sanitizeLine(%q) = %q, want %q", input, got, want)
	}
}
