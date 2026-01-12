package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a test Git repository with a GitHub remote and returns the directory path
func setupTestRepo(t *testing.T, remoteURL string) string {
	tmpDir := t.TempDir()

	// Initialize Git repository
	exec.Command("git", "-C", tmpDir, "init").Run()
	exec.Command("git", "-C", tmpDir, "remote", "add", "origin", remoteURL).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit to establish the repository
	readmePath := filepath.Join(tmpDir, "README.md")
	os.WriteFile(readmePath, []byte("# Test Repository\n"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "README.md").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit").Run()

	return tmpDir
}

// commitFile creates a file and commits it to the repository
func commitFile(t *testing.T, repoDir, filename, content string) {
	filePath := filepath.Join(repoDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	exec.Command("git", "-C", repoDir, "add", filename).Run()
	exec.Command("git", "-C", repoDir, "commit", "-m", "Add "+filename).Run()
}

func TestRun_BasicSearch(t *testing.T) {
	// Setup test repository
	tmpDir := setupTestRepo(t, "https://github.com/test/repo.git")
	commitFile(t, tmpDir, "test.go", "package main\n")

	// Output to file
	outputFile := filepath.Join(tmpDir, "output.tsv")

	// Execute command
	cmd := newRootCmd()
	cmd.SetArgs([]string{"package", tmpDir, "-o", outputFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if output == "" {
		t.Error("Expected non-empty output")
	}

	if !strings.Contains(output, "\t") {
		t.Error("Output should be in TSV format (contain tabs)")
	}

	if !strings.Contains(output, "test/repo") {
		t.Error("Output should contain repository name 'test/repo'")
	}

	if !strings.Contains(output, "test.go") {
		t.Error("Output should contain filename 'test.go'")
	}

	if !strings.Contains(output, "package main") {
		t.Error("Output should contain matched line 'package main'")
	}

	if !strings.Contains(output, "https://github.com/test/repo/blob/") {
		t.Error("Output should contain GitHub URL")
	}
}

func TestRun_OutputToFile(t *testing.T) {
	// Setup test repository
	tmpDir := setupTestRepo(t, "git@github.com:owner/repo.git")
	commitFile(t, tmpDir, "test.txt", "search pattern here\n")

	// Output file path
	outputFile := filepath.Join(tmpDir, "result.tsv")

	// Execute command
	cmd := newRootCmd()
	cmd.SetArgs([]string{"pattern", tmpDir, "-o", outputFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Read and verify output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if output == "" {
		t.Error("Output file should not be empty")
	}

	if !strings.Contains(output, "test.txt") {
		t.Error("Output should contain filename 'test.txt'")
	}
}

func TestRun_InvalidRepository(t *testing.T) {
	// Create temporary directory (not a Git repository)
	tmpDir := t.TempDir()

	// Execute command
	cmd := newRootCmd()
	cmd.SetArgs([]string{"pattern", tmpDir})

	// Execute command - should fail
	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for non-git directory, got nil")
	}
}

func TestRun_NoMatches(t *testing.T) {
	// Setup test repository
	tmpDir := setupTestRepo(t, "https://github.com/test/repo.git")
	commitFile(t, tmpDir, "test.txt", "no match here\n")

	// Output to file
	outputFile := filepath.Join(tmpDir, "output.tsv")

	// Execute command
	cmd := newRootCmd()
	cmd.SetArgs([]string{"nonexistent_pattern_xyz", tmpDir, "-o", outputFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	// Verify output file is empty (no matches)
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) != 0 {
		t.Errorf("Expected empty output for no matches, got: %s", string(content))
	}
}

func TestRun_NotGitHubRepository(t *testing.T) {
	// Setup test repository with GitLab remote
	tmpDir := setupTestRepo(t, "https://gitlab.com/owner/repo.git")
	commitFile(t, tmpDir, "test.txt", "pattern\n")

	// Execute command
	cmd := newRootCmd()
	cmd.SetArgs([]string{"pattern", tmpDir})

	// Execute command - should fail (not a GitHub repo)
	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for non-GitHub repository, got nil")
	}
}

func TestRun_InsufficientArguments(t *testing.T) {
	// Execute command with insufficient args
	cmd := newRootCmd()
	cmd.SetArgs([]string{"pattern"})

	// Execute command - should fail
	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for insufficient arguments, got nil")
	}
}

func TestRun_MultipleRepositories(t *testing.T) {
	// Setup two test repositories
	tmpDir1 := setupTestRepo(t, "https://github.com/test/repo1.git")
	commitFile(t, tmpDir1, "file1.txt", "test pattern\n")

	tmpDir2 := setupTestRepo(t, "https://github.com/test/repo2.git")
	commitFile(t, tmpDir2, "file2.txt", "test pattern\n")

	// Output to file
	outputFile := filepath.Join(t.TempDir(), "output.tsv")

	// Execute command
	cmd := newRootCmd()
	cmd.SetArgs([]string{"pattern", tmpDir1, tmpDir2, "-o", outputFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 2 {
		t.Errorf("Expected at least 2 results (one from each repo), got %d", len(lines))
	}

	// Verify repository names appear in output
	if !strings.Contains(output, "test/repo1") {
		t.Error("Output should contain repository name 'test/repo1'")
	}

	if !strings.Contains(output, "test/repo2") {
		t.Error("Output should contain repository name 'test/repo2'")
	}
}

func TestRun_IgnoreCaseFlag(t *testing.T) {
	tmpDir := setupTestRepo(t, "https://github.com/test/repo.git")
	commitFile(t, tmpDir, "test.txt", "PATTERN here\n")

	outputFile := filepath.Join(tmpDir, "output.tsv")

	// Without -i
	cmd := newRootCmd()
	cmd.SetArgs([]string{"pattern", tmpDir, "-o", outputFile})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content, _ := os.ReadFile(outputFile)
	if len(content) != 0 {
		t.Error("Expected no matches without -i flag")
	}

	// With -i
	os.Remove(outputFile)
	cmd = newRootCmd()
	cmd.SetArgs([]string{"-i", "pattern", tmpDir, "-o", outputFile})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content, _ = os.ReadFile(outputFile)
	if len(content) == 0 {
		t.Error("Expected matches with -i flag")
	}
}

func TestRun_FixedStringsFlag(t *testing.T) {
	tmpDir := setupTestRepo(t, "https://github.com/test/repo.git")
	commitFile(t, tmpDir, "test.go", "func main() {}\n")

	outputFile := filepath.Join(tmpDir, "output.tsv")

	// With -F flag
	cmd := newRootCmd()
	cmd.SetArgs([]string{"-F", "main()", tmpDir, "-o", outputFile})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content, _ := os.ReadFile(outputFile)
	if len(content) == 0 {
		t.Error("Expected matches with -F flag")
	}
}

func TestRun_GlobFlag(t *testing.T) {
	tmpDir := setupTestRepo(t, "https://github.com/test/repo.git")
	commitFile(t, tmpDir, "test.go", "package main\n")
	commitFile(t, tmpDir, "test.txt", "package test\n")

	outputFile := filepath.Join(tmpDir, "output.tsv")

	// Test with --glob to only search .go files
	cmd := newRootCmd()
	cmd.SetArgs([]string{"package", tmpDir, "-g", "*.go", "-o", outputFile})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content, _ := os.ReadFile(outputFile)
	output := string(content)

	if !strings.Contains(output, "test.go") {
		t.Error("Expected test.go in results")
	}

	if strings.Contains(output, "test.txt") {
		t.Error("Did not expect test.txt in results")
	}
}

func TestRun_HiddenFlag(t *testing.T) {
	tmpDir := setupTestRepo(t, "https://github.com/test/repo.git")

	// Create hidden file
	hiddenPath := filepath.Join(tmpDir, ".hidden")
	os.WriteFile(hiddenPath, []byte("secret\n"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "-f", ".hidden").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "Add hidden file").Run()

	outputFile := filepath.Join(tmpDir, "output.tsv")

	// Test without --hidden (should not find)
	cmd := newRootCmd()
	cmd.SetArgs([]string{"secret", tmpDir, "-o", outputFile})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content, _ := os.ReadFile(outputFile)
	if len(content) != 0 {
		t.Error("Expected no matches without --hidden flag")
	}

	// Test with --hidden (should find)
	os.Remove(outputFile)
	cmd = newRootCmd()
	cmd.SetArgs([]string{"secret", tmpDir, "--hidden", "-o", outputFile})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content, _ = os.ReadFile(outputFile)
	if len(content) == 0 {
		t.Error("Expected matches with --hidden flag")
	}
}

func TestRun_MultipleGlobFlags(t *testing.T) {
	tmpDir := setupTestRepo(t, "https://github.com/test/repo.git")
	commitFile(t, tmpDir, "main.go", "package main\n")
	commitFile(t, tmpDir, "main_test.go", "package main\n")

	outputFile := filepath.Join(tmpDir, "output.tsv")

	// Test with multiple --glob flags
	cmd := newRootCmd()
	cmd.SetArgs([]string{"package", tmpDir, "-g", "*.go", "-g", "!*_test.go", "-o", outputFile})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	content, _ := os.ReadFile(outputFile)
	output := string(content)

	if !strings.Contains(output, "main.go") {
		t.Error("Expected main.go in results")
	}

	if strings.Contains(output, "main_test.go") {
		t.Error("Did not expect main_test.go in results")
	}
}
