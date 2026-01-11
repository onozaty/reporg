package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestValidateRepoRoot_ValidRepository(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Test ValidateRepoRoot - should succeed
	err := ValidateRepoRoot(tmpDir)
	if err != nil {
		t.Errorf("ValidateRepoRoot() error = %v, want nil", err)
	}
}

func TestValidateRepoRoot_Subdirectory(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

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

	// Initialize Git repository
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create initial commit (required to have a branch)
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "add", "test.txt")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set user.email: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set user.name: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

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

	// Initialize Git repository
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create initial commit (required to have a branch)
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "add", "test.txt")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set user.email: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set user.name: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create and checkout custom branch
	cmd = exec.Command("git", "-C", tmpDir, "checkout", "-b", "develop")
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

	// Initialize Git repository
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

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

	// Initialize Git repository
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

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

	// Initialize Git repositories
	cmd := exec.Command("git", "-C", tmpDir1, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo 1: %v", err)
	}

	cmd = exec.Command("git", "-C", tmpDir2, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo 2: %v", err)
	}

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
	cmd := exec.Command("git", "-C", validRepo, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create one invalid directory
	invalidRepo := t.TempDir()

	// Test DeduplicateRepoPaths - should fail on invalid repo
	paths := []string{validRepo, invalidRepo}
	_, err := DeduplicateRepoPaths(paths)
	if err == nil {
		t.Error("DeduplicateRepoPaths() expected error for invalid repository, got nil")
	}
}
