package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ValidateRepoRoot validates that the given path is a Git repository root.
// It returns an error if the path is not a Git repository or is a subdirectory.
func ValidateRepoRoot(path string) error {
	// Get absolute path for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Execute: git -C <path> rev-parse --show-toplevel
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("not a git repository: %s", path)
	}

	// Get the repository root from git output
	repoRoot := strings.TrimSpace(string(output))

	// Compare with the input path (must be exact match)
	absRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for repo root: %w", err)
	}

	if absPath != absRepoRoot {
		return fmt.Errorf("path is not a repository root (subdirectory detected): %s", path)
	}

	return nil
}

// GetCurrentBranch returns the current branch name of the repository.
// Returns an empty string if the repository is in detached HEAD state.
func GetCurrentBranch(repoRoot string) (string, error) {
	// Execute: git -C <repoRoot> branch --show-current
	cmd := exec.Command("git", "-C", repoRoot, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// DeduplicateRepoPaths takes a list of repository paths and returns unique repository roots.
// It validates each path and removes duplicates based on canonical paths.
func DeduplicateRepoPaths(paths []string) ([]string, error) {
	seen := make(map[string]bool)
	var unique []string

	for _, path := range paths {
		// Validate that it's a repository root
		if err := ValidateRepoRoot(path); err != nil {
			return nil, err
		}

		// Get canonical absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Add to unique list if not seen
		if !seen[absPath] {
			seen[absPath] = true
			unique = append(unique, absPath)
		}
	}

	return unique, nil
}
