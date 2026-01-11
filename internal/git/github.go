package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Regex patterns for GitHub URLs
	httpsPattern = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
	sshPattern   = regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
)

// GetGitHubRemoteURL returns the origin remote URL for the repository.
func GetGitHubRemoteURL(repoRoot string) (string, error) {
	// Execute: git -C <repoRoot> remote get-url origin
	cmd := exec.Command("git", "-C", repoRoot, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	return remoteURL, nil
}

// ParseGitHubURL parses a GitHub remote URL and extracts the owner and repository name.
// Supports both HTTPS and SSH formats.
func ParseGitHubURL(remoteURL string) (owner, repo string, err error) {
	// Try HTTPS pattern first
	if matches := httpsPattern.FindStringSubmatch(remoteURL); matches != nil {
		return matches[1], matches[2], nil
	}

	// Try SSH pattern
	if matches := sshPattern.FindStringSubmatch(remoteURL); matches != nil {
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("not a valid GitHub URL: %s", remoteURL)
}

// BuildGitHubFileURL constructs a GitHub blob URL for a specific file and line number.
func BuildGitHubFileURL(owner, repo, branch, relPath string, lineNum int) string {
	// Ensure forward slashes in path (cross-platform compatibility)
	relPath = filepath.ToSlash(relPath)

	// Construct URL: https://github.com/{owner}/{repo}/blob/{branch}/{path}#L{line}
	// Note: GitHub handles special characters in paths without URL encoding
	return fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s#L%d",
		owner, repo, branch, relPath, lineNum)
}
