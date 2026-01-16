package search

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// collectMatches is a helper function that collects all matches into a slice
func collectMatches(pattern, dir string, opts SearchOptions) ([]Match, error) {
	var matches []Match
	err := SearchRepo(pattern, dir, opts, func(match Match) error {
		matches = append(matches, match)
		return nil
	})
	return matches, err
}

func TestSearchRepo_WithMatches(t *testing.T) {
	// Create temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	if err := os.WriteFile(testFile1, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	testFile2 := filepath.Join(tmpDir, "test2.txt")
	if err := os.WriteFile(testFile2, []byte("package search\nfunc Search() {}\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test SearchRepo with pattern "package"
	var matches []Match
	err := SearchRepo("package", tmpDir, SearchOptions{}, func(match Match) error {
		matches = append(matches, match)
		return nil
	})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	// Should find 2 matches
	if len(matches) != 2 {
		t.Errorf("SearchRepo() returned %d matches, want 2", len(matches))
	}

	// Verify matches contain expected data
	for _, match := range matches {
		if match.LineNumber != 1 {
			t.Errorf("Match line number = %d, want 1", match.LineNumber)
		}
		if match.LineText == "" {
			t.Error("Match line text is empty")
		}
		if match.RelPath == "" {
			t.Error("Match relative path is empty")
		}
	}
}

func TestSearchRepo_NoMatches(t *testing.T) {
	// Create temporary directory with test file
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test SearchRepo with pattern that won't match
	matches, err := collectMatches("nonexistent_pattern_xyz", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	// Should find 0 matches
	if len(matches) != 0 {
		t.Errorf("SearchRepo() returned %d matches, want 0", len(matches))
	}
}

func TestSearchRepo_EmptyDirectory(t *testing.T) {
	// Create empty temporary directory
	tmpDir := t.TempDir()

	// Test SearchRepo on empty directory
	matches, err := collectMatches("package", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	// Should find 0 matches
	if len(matches) != 0 {
		t.Errorf("SearchRepo() returned %d matches, want 0", len(matches))
	}
}

func TestSearchRepo_MultipleMatchesInSingleFile(t *testing.T) {
	// Create temporary directory with test file
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

import "fmt"

func main() {
	fmt.Println("package")
	// package comment
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test SearchRepo with pattern "package"
	matches, err := collectMatches("package", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	// Should find 3 matches (line 1, 6, 7)
	if len(matches) != 3 {
		t.Errorf("SearchRepo() returned %d matches, want 3", len(matches))
	}

	// Verify line numbers
	expectedLines := []int{1, 6, 7}
	for i, match := range matches {
		if match.LineNumber != expectedLines[i] {
			t.Errorf("Match %d line number = %d, want %d", i, match.LineNumber, expectedLines[i])
		}
	}
}

func TestSearchRepo_RelativePath(t *testing.T) {
	// Create temporary directory with nested structure
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	testFile := filepath.Join(subDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test SearchRepo
	matches, err := collectMatches("package", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	if len(matches) != 1 {
		t.Fatalf("SearchRepo() returned %d matches, want 1", len(matches))
	}

	// Verify relative path is correct
	expectedPath := filepath.Join("subdir", "test.go")
	if matches[0].RelPath != expectedPath {
		t.Errorf("Match relative path = %v, want %v", matches[0].RelPath, expectedPath)
	}
}

func TestSearchRepo_UTF8Content(t *testing.T) {
	// Create temporary directory with UTF-8 content
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "日本語のテスト\nUTF-8 content\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test SearchRepo with Japanese pattern
	matches, err := collectMatches("日本語", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	if len(matches) != 1 {
		t.Fatalf("SearchRepo() returned %d matches, want 1", len(matches))
	}

	// Verify UTF-8 content is preserved
	// Note: trailing newline is stripped
	expected := "日本語のテスト"
	if matches[0].LineText != expected {
		t.Errorf("Match line text = %q, want %q", matches[0].LineText, expected)
	}
}

func TestSearchRepo_InvalidPattern(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test SearchRepo with invalid regex pattern
	_, err := collectMatches("[invalid", tmpDir, SearchOptions{})
	if err == nil {
		t.Error("SearchRepo() expected error for invalid pattern, got nil")
	}
}

func TestSearchRepo_NonexistentDirectory(t *testing.T) {
	// Test SearchRepo with non-existent directory
	_, err := collectMatches("package", "/nonexistent/directory/path", SearchOptions{})
	if err == nil {
		t.Error("SearchRepo() expected error for non-existent directory, got nil")
	}
}

func TestSearchRepo_RipgrepInstalled(t *testing.T) {
	// Verify ripgrep is installed
	_, err := exec.LookPath("rg")
	if err != nil {
		t.Skip("ripgrep not installed, skipping tests")
	}
}

func TestRipgrepMessage_MatchType(t *testing.T) {
	// Sample JSON from ripgrep --json output (match type)
	jsonData := `{
		"type": "match",
		"data": {
			"path": {"text": "main.go"},
			"lines": {"text": "package main"},
			"line_number": 1,
			"absolute_offset": 0,
			"submatches": []
		}
	}`

	var msg RipgrepMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if msg.Type != "match" {
		t.Errorf("Type = %v, want 'match'", msg.Type)
	}

	if msg.Data == nil {
		t.Error("Data is nil, want non-nil")
	}
}

func TestRipgrepMessage_BeginType(t *testing.T) {
	// Sample JSON from ripgrep --json output (begin type)
	jsonData := `{
		"type": "begin",
		"data": {
			"path": {"text": "main.go"}
		}
	}`

	var msg RipgrepMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if msg.Type != "begin" {
		t.Errorf("Type = %v, want 'begin'", msg.Type)
	}
}

func TestMatchData_ValidMatch(t *testing.T) {
	// Sample match data JSON
	jsonData := `{
		"path": {"text": "internal/git/repo.go"},
		"lines": {"text": "func ValidateRepoRoot(path string) error {"},
		"line_number": 10,
		"absolute_offset": 250,
		"submatches": []
	}`

	var matchData MatchData
	err := json.Unmarshal([]byte(jsonData), &matchData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check path
	if matchData.Path.Text == nil {
		t.Fatal("Path.Text is nil, want non-nil")
	}
	if *matchData.Path.Text != "internal/git/repo.go" {
		t.Errorf("Path.Text = %v, want 'internal/git/repo.go'", *matchData.Path.Text)
	}

	// Check lines
	if matchData.Lines.Text == nil {
		t.Fatal("Lines.Text is nil, want non-nil")
	}
	if *matchData.Lines.Text != "func ValidateRepoRoot(path string) error {" {
		t.Errorf("Lines.Text = %v, want 'func ValidateRepoRoot(path string) error {'", *matchData.Lines.Text)
	}

	// Check line number
	if matchData.LineNumber != 10 {
		t.Errorf("LineNumber = %v, want 10", matchData.LineNumber)
	}
}

func TestMatchData_WithUTF8Text(t *testing.T) {
	// Sample match data with UTF-8 text
	jsonData := `{
		"path": {"text": "test.go"},
		"lines": {"text": "// 日本語コメント"},
		"line_number": 5
	}`

	var matchData MatchData
	err := json.Unmarshal([]byte(jsonData), &matchData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if matchData.Lines.Text == nil {
		t.Fatal("Lines.Text is nil, want non-nil")
	}
	if *matchData.Lines.Text != "// 日本語コメント" {
		t.Errorf("Lines.Text = %v, want '// 日本語コメント'", *matchData.Lines.Text)
	}
}

func TestMatchData_EmptyPath(t *testing.T) {
	// Match data without path text (should be null)
	jsonData := `{
		"path": {},
		"lines": {"text": "some content"},
		"line_number": 1
	}`

	var matchData MatchData
	err := json.Unmarshal([]byte(jsonData), &matchData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if matchData.Path.Text != nil {
		t.Errorf("Path.Text = %v, want nil", *matchData.Path.Text)
	}
}

func TestMatchData_EmptyLines(t *testing.T) {
	// Match data without lines text (should be null)
	jsonData := `{
		"path": {"text": "test.go"},
		"lines": {},
		"line_number": 1
	}`

	var matchData MatchData
	err := json.Unmarshal([]byte(jsonData), &matchData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if matchData.Lines.Text != nil {
		t.Errorf("Lines.Text = %v, want nil", *matchData.Lines.Text)
	}
}

func TestPathData_WithText(t *testing.T) {
	jsonData := `{"text": "src/main.go"}`

	var pathData PathData
	err := json.Unmarshal([]byte(jsonData), &pathData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if pathData.Text == nil {
		t.Fatal("Text is nil, want non-nil")
	}
	if *pathData.Text != "src/main.go" {
		t.Errorf("Text = %v, want 'src/main.go'", *pathData.Text)
	}
}

func TestPathData_Empty(t *testing.T) {
	jsonData := `{}`

	var pathData PathData
	err := json.Unmarshal([]byte(jsonData), &pathData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if pathData.Text != nil {
		t.Errorf("Text = %v, want nil", pathData.Text)
	}
}

func TestTextData_WithText(t *testing.T) {
	jsonData := `{"text": "package main"}`

	var textData TextData
	err := json.Unmarshal([]byte(jsonData), &textData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if textData.Text == nil {
		t.Fatal("Text is nil, want non-nil")
	}
	if *textData.Text != "package main" {
		t.Errorf("Text = %v, want 'package main'", *textData.Text)
	}
}

func TestTextData_Empty(t *testing.T) {
	jsonData := `{}`

	var textData TextData
	err := json.Unmarshal([]byte(jsonData), &textData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if textData.Text != nil {
		t.Errorf("Text = %v, want nil", textData.Text)
	}
}

func TestFullRipgrepJSONParsing(t *testing.T) {
	// Complete example of ripgrep JSON output with match type
	jsonData := `{"type":"match","data":{"path":{"text":"cmd/root.go"},"lines":{"text":"package cmd"},"line_number":1,"absolute_offset":0,"submatches":[{"match":{"text":"package"},"start":0,"end":7}]}}`

	var msg RipgrepMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if msg.Type != "match" {
		t.Errorf("Type = %v, want 'match'", msg.Type)
	}

	var matchData MatchData
	err = json.Unmarshal(msg.Data, &matchData)
	if err != nil {
		t.Fatalf("Failed to unmarshal match data: %v", err)
	}

	// Verify all fields
	if matchData.Path.Text == nil || *matchData.Path.Text != "cmd/root.go" {
		t.Errorf("Path = %v, want 'cmd/root.go'", matchData.Path.Text)
	}

	if matchData.Lines.Text == nil || *matchData.Lines.Text != "package cmd" {
		t.Errorf("Lines = %v, want 'package cmd'", matchData.Lines.Text)
	}

	if matchData.LineNumber != 1 {
		t.Errorf("LineNumber = %v, want 1", matchData.LineNumber)
	}
}

func TestSearchRepo_IgnoreCase(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("PACKAGE main\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test case-sensitive (should not match)
	matches, err := collectMatches("package", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	if len(matches) != 0 {
		t.Errorf("Case-sensitive search found %d matches, want 0", len(matches))
	}

	// Test case-insensitive (should match)
	matches, err = collectMatches("package", tmpDir, SearchOptions{IgnoreCase: true})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	if len(matches) != 1 {
		t.Errorf("Case-insensitive search found %d matches, want 1", len(matches))
	}
}

func TestSearchRepo_Glob_Include(t *testing.T) {
	tmpDir := t.TempDir()

	goFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(goFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to write .go file: %v", err)
	}

	txtFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("package test\n"), 0644); err != nil {
		t.Fatalf("Failed to write .txt file: %v", err)
	}

	matches, err := collectMatches("package", tmpDir, SearchOptions{
		Globs: []string{"*.go"},
	})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	if len(matches) != 1 {
		t.Errorf("Glob search found %d matches, want 1", len(matches))
	}

	if len(matches) > 0 && !strings.HasSuffix(matches[0].RelPath, ".go") {
		t.Errorf("Expected .go file, got %s", matches[0].RelPath)
	}
}

func TestSearchRepo_Glob_Exclude(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	testFileExcluded := filepath.Join(tmpDir, "test_test.go")
	if err := os.WriteFile(testFileExcluded, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	matches, err := collectMatches("package", tmpDir, SearchOptions{
		Globs: []string{"*.go", "!*_test.go"},
	})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	if len(matches) != 1 {
		t.Errorf("Glob exclude search found %d matches, want 1", len(matches))
	}

	if len(matches) > 0 && strings.Contains(matches[0].RelPath, "_test.go") {
		t.Errorf("Expected to exclude test files, but got %s", matches[0].RelPath)
	}
}

func TestSearchRepo_Hidden(t *testing.T) {
	tmpDir := t.TempDir()

	hiddenFile := filepath.Join(tmpDir, ".hidden")
	if err := os.WriteFile(hiddenFile, []byte("secret data\n"), 0644); err != nil {
		t.Fatalf("Failed to write hidden file: %v", err)
	}

	// Without --hidden
	matches, err := collectMatches("secret", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	if len(matches) != 0 {
		t.Errorf("Search without --hidden found %d matches, want 0", len(matches))
	}

	// With --hidden
	matches, err = collectMatches("secret", tmpDir, SearchOptions{Hidden: true})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	if len(matches) != 1 {
		t.Errorf("Search with --hidden found %d matches, want 1", len(matches))
	}
}

func TestSearchRepo_FixedStrings(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("func main() {}\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Without -F, "main()" is regex (would match)
	matches, err := collectMatches("main()", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	if len(matches) != 1 {
		t.Errorf("Regex search found %d matches, want 1", len(matches))
	}

	// With -F, "main()" is literal (would match)
	matches, err = collectMatches("main()", tmpDir, SearchOptions{FixedStrings: true})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	if len(matches) != 1 {
		t.Errorf("Fixed-strings search found %d matches, want 1", len(matches))
	}

	// Test that regex metacharacters are treated literally with -F
	testFile2 := filepath.Join(tmpDir, "regex.txt")
	if err := os.WriteFile(testFile2, []byte(".*pattern\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Without -F, ".*" would be regex (match any chars)
	matches, err = collectMatches(".*pattern", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	// Should match both files (as regex)
	if len(matches) < 1 {
		t.Errorf("Regex search found %d matches, want at least 1", len(matches))
	}

	// With -F, ".*pattern" must match literally
	matches, err = collectMatches(".*pattern", tmpDir, SearchOptions{FixedStrings: true})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	// Should only match the file with literal ".*pattern"
	if len(matches) != 1 {
		t.Errorf("Fixed-strings search found %d matches, want 1", len(matches))
	}
}

func TestSearchRepo_CombinedOptions(t *testing.T) {
	tmpDir := t.TempDir()

	hiddenFile := filepath.Join(tmpDir, ".hidden.go")
	if err := os.WriteFile(hiddenFile, []byte("PACKAGE main\n"), 0644); err != nil {
		t.Fatalf("Failed to write hidden file: %v", err)
	}

	matches, err := collectMatches("package", tmpDir, SearchOptions{
		IgnoreCase: true,
		Globs:      []string{"*.go"},
		Hidden:     true,
	})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	if len(matches) != 1 {
		t.Errorf("Combined options search found %d matches, want 1", len(matches))
	}
}

func TestSearchRepo_VeryLongLine(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with a very long line (100KB, exceeds default 64KB buffer)
	// This tests that the increased buffer size (10MB) can handle large lines
	longLine := strings.Repeat("a", 100*1024) + " test pattern in long line"
	testFile := filepath.Join(tmpDir, "long.txt")
	content := longLine + "\nanother test line\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Search for "test" pattern - should find matches in both the long line and short line
	matches, err := collectMatches("test", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	// Should find 2 matches (one in the long line, one in the short line)
	if len(matches) != 2 {
		t.Errorf("Search found %d matches, want 2", len(matches))
	}
}

func TestSearchRepo_MaxLineLength(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with a long line
	longLine := strings.Repeat("x", 200) + " test pattern here"
	testFile := filepath.Join(tmpDir, "test.txt")
	content := longLine + "\nshort test line\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test without MaxLineLength (should return full line)
	matches, err := collectMatches("test", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	if len(matches) != 2 {
		t.Fatalf("Search found %d matches, want 2", len(matches))
	}
	// First match should have full line content (newline is stripped)
	if len(matches[0].LineText) != len(longLine) {
		t.Errorf("Without MaxLineLength, line length = %d, want %d", len(matches[0].LineText), len(longLine))
	}

	// Test with MaxLineLength = 50 (should truncate to 50 chars + "...")
	matches, err = collectMatches("test", tmpDir, SearchOptions{MaxLineLength: 50})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}
	if len(matches) != 2 {
		t.Fatalf("Search found %d matches, want 2", len(matches))
	}
	// First match should be truncated to 50 + "..." = 53 chars
	if len(matches[0].LineText) != 53 {
		t.Errorf("With MaxLineLength=50, line length = %d, want 53 (50 + '...')", len(matches[0].LineText))
	}
	if !strings.HasSuffix(matches[0].LineText, "...") {
		t.Errorf("Truncated line should end with '...', got: %s", matches[0].LineText)
	}
	// Second match (short line) should not be truncated (newline is stripped)
	expectedShortLineLen := len("short test line")
	if len(matches[1].LineText) != expectedShortLineLen {
		t.Errorf("Short line should not be truncated, length = %d, want %d", len(matches[1].LineText), expectedShortLineLen)
	}
}

func TestSearchRepo_Encoding(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file with Shift-JIS encoding
	// 「テスト」in Shift-JIS: 0x83 0x65 0x83 0x58 0x83 0x67
	testFileSJIS := filepath.Join(tmpDir, "test_sjis.txt")
	sjisContent := []byte{0x83, 0x65, 0x83, 0x58, 0x83, 0x67, 0x0a} // テスト\n in Shift-JIS
	if err := os.WriteFile(testFileSJIS, sjisContent, 0644); err != nil {
		t.Fatalf("Failed to write Shift-JIS test file: %v", err)
	}

	// Test with default/auto encoding - should NOT find Shift-JIS file
	// (auto encoding only detects UTF-8/UTF-16 BOM)
	matches, err := collectMatches("テスト", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() with default encoding error = %v, want nil", err)
	}
	if len(matches) != 0 {
		t.Errorf("Search with default encoding should not find Shift-JIS file, found %d matches", len(matches))
	}

	// Test with Shift-JIS encoding specified - should find the file
	matches, err = collectMatches("テスト", tmpDir, SearchOptions{Encoding: "shift_jis"})
	if err != nil {
		t.Fatalf("SearchRepo() with shift_jis encoding error = %v, want nil", err)
	}
	if len(matches) != 1 {
		t.Errorf("Search with shift_jis encoding found %d matches, want 1", len(matches))
	}
	if len(matches) > 0 && matches[0].LineText != "テスト" {
		t.Errorf("Match line text = %q, want %q", matches[0].LineText, "テスト")
	}

	// Create a UTF-8 test file
	testFileUTF8 := filepath.Join(tmpDir, "test_utf8.txt")
	if err := os.WriteFile(testFileUTF8, []byte("UTF-8テスト\n"), 0644); err != nil {
		t.Fatalf("Failed to write UTF-8 test file: %v", err)
	}

	// Test with UTF-8 encoding specified
	matches, err = collectMatches("UTF-8", tmpDir, SearchOptions{Encoding: "utf-8"})
	if err != nil {
		t.Fatalf("SearchRepo() with utf-8 encoding error = %v, want nil", err)
	}
	if len(matches) != 1 {
		t.Errorf("Search with utf-8 encoding found %d matches, want 1", len(matches))
	}

	// Test with auto encoding (default) - should find UTF-8 file
	matches, err = collectMatches("UTF-8", tmpDir, SearchOptions{Encoding: "auto"})
	if err != nil {
		t.Fatalf("SearchRepo() with auto encoding error = %v, want nil", err)
	}
	if len(matches) != 1 {
		t.Errorf("Search with auto encoding found %d matches, want 1 (UTF-8 file only)", len(matches))
	}
}

func TestSearchRepo_OnMatchCallback(t *testing.T) {
	// Create temporary directory with test files
	tmpDir := t.TempDir()

	testFile1 := filepath.Join(tmpDir, "test1.txt")
	if err := os.WriteFile(testFile1, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	testFile2 := filepath.Join(tmpDir, "test2.txt")
	if err := os.WriteFile(testFile2, []byte("package search\nfunc Search() {}\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test with callback
	var callbackMatches []Match
	err := SearchRepo("package", tmpDir, SearchOptions{}, func(match Match) error {
		callbackMatches = append(callbackMatches, match)
		return nil
	})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	// Should have received 2 matches via callback
	if len(callbackMatches) != 2 {
		t.Errorf("Callback received %d matches, want 2", len(callbackMatches))
	}

	// Verify callback matches contain expected data
	for _, match := range callbackMatches {
		if match.LineNumber != 1 {
			t.Errorf("Match line number = %d, want 1", match.LineNumber)
		}
		if match.LineText == "" {
			t.Error("Match line text is empty")
		}
	}
}

func TestSearchRepo_OnMatchCallback_Error(t *testing.T) {
	// Create temporary directory with test file
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test with callback that returns an error
	callbackErr := fmt.Errorf("callback error")
	err := SearchRepo("package", tmpDir, SearchOptions{}, func(match Match) error {
		return callbackErr
	})

	// Should return the callback error
	if err == nil {
		t.Error("SearchRepo() expected error from callback, got nil")
	}
	if !strings.Contains(err.Error(), "callback error") {
		t.Errorf("Error message = %v, want to contain 'callback error'", err)
	}
}

func TestSearchRepo_BytesFieldDecoding(t *testing.T) {
	// This test verifies that ripgrep's "bytes" field (base64-encoded) is properly decoded
	// When ripgrep encounters non-UTF-8 content without --encoding flag, it returns base64-encoded bytes
	tmpDir := t.TempDir()

	// Create a file with mixed ASCII and Shift-JIS content
	// "ABC" + "テスト" (Shift-JIS: 0x83 0x65 0x83 0x58 0x83 0x67) + "XYZ\n"
	testFile := filepath.Join(tmpDir, "mixed.txt")
	mixedContent := []byte{0x41, 0x42, 0x43, 0x83, 0x65, 0x83, 0x58, 0x83, 0x67, 0x58, 0x59, 0x5a, 0x0a}
	if err := os.WriteFile(testFile, mixedContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Search for "ABC" without --encoding flag
	// ripgrep will detect non-UTF-8 content and return bytes field instead of text field
	matches, err := collectMatches("ABC", tmpDir, SearchOptions{})
	if err != nil {
		t.Fatalf("SearchRepo() error = %v, want nil", err)
	}

	if len(matches) != 1 {
		t.Fatalf("SearchRepo() returned %d matches, want 1", len(matches))
	}

	// Verify the line text was decoded from base64 bytes field
	// The line should contain "ABC", the Shift-JIS bytes (as raw bytes), and "XYZ"
	if !strings.Contains(matches[0].LineText, "ABC") {
		t.Errorf("Match line text = %q, should contain 'ABC'", matches[0].LineText)
	}
	if !strings.Contains(matches[0].LineText, "XYZ") {
		t.Errorf("Match line text = %q, should contain 'XYZ'", matches[0].LineText)
	}

	// The full decoded bytes should match the original content (minus newline)
	expectedBytes := []byte{0x41, 0x42, 0x43, 0x83, 0x65, 0x83, 0x58, 0x83, 0x67, 0x58, 0x59, 0x5a}
	if string(expectedBytes) != matches[0].LineText {
		t.Errorf("Match line text bytes mismatch.\nGot:  %v\nWant: %v",
			[]byte(matches[0].LineText), expectedBytes)
	}

	// Verify that the Shift-JIS part is NOT correctly decoded (i.e., it's garbled)
	// If it were correctly decoded, it would be "テスト", but since we didn't use --encoding,
	// the bytes 0x83 0x65 0x83 0x58 0x83 0x67 remain as invalid UTF-8 and appear garbled
	if strings.Contains(matches[0].LineText, "テスト") {
		t.Errorf("Match line text should NOT contain correctly decoded 'テスト', got %q", matches[0].LineText)
	}

	// The middle bytes (Shift-JIS "テスト") should be present as raw invalid UTF-8 bytes
	// When Go tries to interpret these as UTF-8, they will be replaced with replacement characters
	// or remain as invalid sequences. We verify this by checking the byte sequence is present.
	middleBytes := []byte{0x83, 0x65, 0x83, 0x58, 0x83, 0x67}
	if !strings.Contains(matches[0].LineText, string(middleBytes)) {
		t.Errorf("Match line text should contain raw Shift-JIS bytes as garbled text")
	}
}
