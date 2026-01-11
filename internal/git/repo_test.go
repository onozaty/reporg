package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initTestRepo initializes a Git repository with an initial commit
func initTestRepo(t *testing.T, dir string) {
	t.Helper()

	exec.Command("git", "-C", dir, "init").Run()
	exec.Command("git", "-C", dir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", dir, "config", "user.name", "Test User").Run()

	// Create initial commit (required for git rev-parse on macOS)
	readmePath := filepath.Join(dir, "README.md")
	os.WriteFile(readmePath, []byte("test\n"), 0644)
	exec.Command("git", "-C", dir, "add", "README.md").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "Initial commit").Run()
}

func TestValidateRepoRoot_ValidRepository(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository with initial commit
	initTestRepo(t, tmpDir)

	// Test ValidateRepoRoot - should succeed
	err := ValidateRepoRoot(tmpDir)
	if err != nil {
		t.Errorf("ValidateRepoRoot() error = %v, want nil", err)
	}
}

func TestValidateRepoRoot_Subdirectory(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository with initial commit
	initTestRepo(t, tmpDir)

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Test ValidateRepoRoot with subdirectory - should fail
	err := ValidateRepoRoot(subDir)
	if err == nil {
		t.Error("ValidateRepoRoot() expected error for subdirectory, got nil")
	}
}

func TestValidateRepoRoot_NotGitRepository(t *testing.T) {
	// Create temporary directory without Git
	tmpDir := t.TempDir()

	// Test ValidateRepoRoot - should fail
	err := ValidateRepoRoot(tmpDir)
	if err == nil {
		t.Error("ValidateRepoRoot() expected error for non-git directory, got nil")
	}
}

func TestGetCurrentBranch_RepositoryWithBranch(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository with initial commit
	initTestRepo(t, tmpDir)

	// Test GetCurrentBranch
	branch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Errorf("GetCurrentBranch() error = %v, want nil", err)
	}

	// Branch name should be "main" or "master" depending on Git version
	if branch != "main" && branch != "master" {
		t.Errorf("GetCurrentBranch() = %v, want 'main' or 'master'", branch)
	}
}

func TestGetCurrentBranch_CustomBranch(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository with initial commit
	initTestRepo(t, tmpDir)

	// Create and checkout custom branch
	cmd := exec.Command("git", "-C", tmpDir, "checkout", "-b", "develop")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create branch: %v", err)
	}

	// Test GetCurrentBranch
	branch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Errorf("GetCurrentBranch() error = %v, want nil", err)
	}

	if branch != "develop" {
		t.Errorf("GetCurrentBranch() = %v, want 'develop'", branch)
	}
}

func TestDeduplicateRepoPaths_SingleRepository(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository with initial commit
	initTestRepo(t, tmpDir)

	// Test DeduplicateRepoPaths
	paths := []string{tmpDir}
	unique, err := DeduplicateRepoPaths(paths)
	if err != nil {
		t.Fatalf("DeduplicateRepoPaths() error = %v, want nil", err)
	}

	if len(unique) != 1 {
		t.Errorf("DeduplicateRepoPaths() returned %d paths, want 1", len(unique))
	}
}

func TestDeduplicateRepoPaths_Duplicate(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository with initial commit
	initTestRepo(t, tmpDir)

	// Test DeduplicateRepoPaths with duplicates
	paths := []string{tmpDir, tmpDir, tmpDir}
	unique, err := DeduplicateRepoPaths(paths)
	if err != nil {
		t.Fatalf("DeduplicateRepoPaths() error = %v, want nil", err)
	}

	if len(unique) != 1 {
		t.Errorf("DeduplicateRepoPaths() returned %d paths, want 1", len(unique))
	}
}

func TestDeduplicateRepoPaths_MultipleDifferent(t *testing.T) {
	// Create two temporary directories for Git repositories
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	// Initialize Git repositories with initial commits
	initTestRepo(t, tmpDir1)
	initTestRepo(t, tmpDir2)

	// Test DeduplicateRepoPaths with different repos
	paths := []string{tmpDir1, tmpDir2}
	unique, err := DeduplicateRepoPaths(paths)
	if err != nil {
		t.Fatalf("DeduplicateRepoPaths() error = %v, want nil", err)
	}

	if len(unique) != 2 {
		t.Errorf("DeduplicateRepoPaths() returned %d paths, want 2", len(unique))
	}
}

func TestDeduplicateRepoPaths_InvalidRepository(t *testing.T) {
	// Create temporary directory without Git
	tmpDir := t.TempDir()

	// Test DeduplicateRepoPaths - should fail
	paths := []string{tmpDir}
	_, err := DeduplicateRepoPaths(paths)
	if err == nil {
		t.Error("DeduplicateRepoPaths() expected error for non-git directory, got nil")
	}
}

func TestDeduplicateRepoPaths_MixedValidInvalid(t *testing.T) {
	// Create one valid Git repository
	validRepo := t.TempDir()
	initTestRepo(t, validRepo)

	// Create one invalid directory
	invalidRepo := t.TempDir()

	// Test DeduplicateRepoPaths - should fail on invalid repo
	paths := []string{validRepo, invalidRepo}
	_, err := DeduplicateRepoPaths(paths)
	if err == nil {
		t.Error("DeduplicateRepoPaths() expected error for invalid repository, got nil")
	}
}
