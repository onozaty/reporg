package output

import (
	"bytes"
	"errors"
	"testing"
)

// errorWriter is a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestTSVWriter_Write_SingleResult(t *testing.T) {
	var buf bytes.Buffer
	writer := NewTSVWriter(&buf)

	result := SearchResult{
		Repository:  "owner/repo",
		LocalPath:   "main.go:10",
		MatchedLine: "package main",
		GitHubURL:   "https://github.com/owner/repo/blob/main/main.go#L10",
	}

	err := writer.Write(result)
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	want := "owner/repo\tmain.go:10\tpackage main\thttps://github.com/owner/repo/blob/main/main.go#L10\n"
	got := buf.String()

	if got != want {
		t.Errorf("Write() = %q, want %q", got, want)
	}
}

func TestTSVWriter_Write_MultipleResults(t *testing.T) {
	var buf bytes.Buffer
	writer := NewTSVWriter(&buf)

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

	for _, result := range results {
		err := writer.Write(result)
		if err != nil {
			t.Fatalf("Write() error = %v, want nil", err)
		}
	}

	want := "owner/repo\tmain.go:10\tpackage main\thttps://github.com/owner/repo/blob/main/main.go#L10\n" +
		"owner/repo\tcmd/root.go:25\tfunc Execute() error {\thttps://github.com/owner/repo/blob/main/cmd/root.go#L25\n"
	got := buf.String()

	if got != want {
		t.Errorf("Write() = %q, want %q", got, want)
	}
}

func TestTSVWriter_Write_TabsInMatchedLine(t *testing.T) {
	var buf bytes.Buffer
	writer := NewTSVWriter(&buf)

	result := SearchResult{
		Repository:  "owner/repo",
		LocalPath:   "test.go:5",
		MatchedLine: "key\tvalue\tdata",
		GitHubURL:   "https://github.com/owner/repo/blob/main/test.go#L5",
	}

	err := writer.Write(result)
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Tabs in matched line should be replaced with spaces
	want := "owner/repo\ttest.go:5\tkey value data\thttps://github.com/owner/repo/blob/main/test.go#L5\n"
	got := buf.String()

	if got != want {
		t.Errorf("Write() = %q, want %q", got, want)
	}
}

func TestTSVWriter_Write_NewlinesInMatchedLine(t *testing.T) {
	var buf bytes.Buffer
	writer := NewTSVWriter(&buf)

	result := SearchResult{
		Repository:  "owner/repo",
		LocalPath:   "test.go:5",
		MatchedLine: "line1\nline2\rline3",
		GitHubURL:   "https://github.com/owner/repo/blob/main/test.go#L5",
	}

	err := writer.Write(result)
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Newlines in matched line should be replaced with spaces
	want := "owner/repo\ttest.go:5\tline1 line2 line3\thttps://github.com/owner/repo/blob/main/test.go#L5\n"
	got := buf.String()

	if got != want {
		t.Errorf("Write() = %q, want %q", got, want)
	}
}

func TestTSVWriter_Write_Error(t *testing.T) {
	writer := NewTSVWriter(&errorWriter{})

	result := SearchResult{
		Repository:  "owner/repo",
		LocalPath:   "test.go:5",
		MatchedLine: "test",
		GitHubURL:   "https://github.com/owner/repo/blob/main/test.go#L5",
	}

	err := writer.Write(result)
	if err == nil {
		t.Error("Write() expected error, got nil")
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
