package git

import (
	"os/exec"
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "HTTPS URL with .git",
			remoteURL: "https://github.com/onozaty/reporg.git",
			wantOwner: "onozaty",
			wantRepo:  "reporg",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL without .git",
			remoteURL: "https://github.com/onozaty/reporg",
			wantOwner: "onozaty",
			wantRepo:  "reporg",
			wantErr:   false,
		},
		{
			name:      "SSH URL with .git",
			remoteURL: "git@github.com:onozaty/reporg.git",
			wantOwner: "onozaty",
			wantRepo:  "reporg",
			wantErr:   false,
		},
		{
			name:      "SSH URL without .git",
			remoteURL: "git@github.com:onozaty/reporg",
			wantOwner: "onozaty",
			wantRepo:  "reporg",
			wantErr:   false,
		},
		{
			name:      "Invalid URL - not GitHub",
			remoteURL: "https://gitlab.com/owner/repo.git",
			wantOwner: "",
			wantRepo:  "",
			wantErr:   true,
		},
		{
			name:      "Invalid URL - malformed",
			remoteURL: "not-a-url",
			wantOwner: "",
			wantRepo:  "",
			wantErr:   true,
		},
		{
			name:      "HTTPS URL with hyphen and underscore",
			remoteURL: "https://github.com/my-org/my_repo.git",
			wantOwner: "my-org",
			wantRepo:  "my_repo",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepo, err := ParseGitHubURL(tt.remoteURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotOwner != tt.wantOwner {
				t.Errorf("ParseGitHubURL() owner = %v, want %v", gotOwner, tt.wantOwner)
			}

			if gotRepo != tt.wantRepo {
				t.Errorf("ParseGitHubURL() repo = %v, want %v", gotRepo, tt.wantRepo)
			}
		})
	}
}

func TestBuildGitHubFileURL(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		repo     string
		branch   string
		relPath  string
		lineNum  int
		wantURL  string
	}{
		{
			name:    "Simple file path",
			owner:   "onozaty",
			repo:    "reporg",
			branch:  "main",
			relPath: "main.go",
			lineNum: 10,
			wantURL: "https://github.com/onozaty/reporg/blob/main/main.go#L10",
		},
		{
			name:    "Nested file path",
			owner:   "onozaty",
			repo:    "reporg",
			branch:  "main",
			relPath: "internal/git/remote.go",
			lineNum: 42,
			wantURL: "https://github.com/onozaty/reporg/blob/main/internal/git/remote.go#L42",
		},
		{
			name:    "Different branch",
			owner:   "onozaty",
			repo:    "reporg",
			branch:  "develop",
			relPath: "cmd/root.go",
			lineNum: 1,
			wantURL: "https://github.com/onozaty/reporg/blob/develop/cmd/root.go#L1",
		},
		{
			name:    "File with space in path",
			owner:   "onozaty",
			repo:    "reporg",
			branch:  "main",
			relPath: "path with spaces/file.go",
			lineNum: 5,
			wantURL: "https://github.com/onozaty/reporg/blob/main/path with spaces/file.go#L5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL := BuildGitHubFileURL(tt.owner, tt.repo, tt.branch, tt.relPath, tt.lineNum)

			if gotURL != tt.wantURL {
				t.Errorf("BuildGitHubFileURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestGetGitHubRemoteURL(t *testing.T) {
	tests := []struct {
		name       string
		remoteURL  string
		wantURL    string
		wantErr    bool
	}{
		{
			name:      "HTTPS remote URL",
			remoteURL: "https://github.com/test/repo.git",
			wantURL:   "https://github.com/test/repo.git",
			wantErr:   false,
		},
		{
			name:      "SSH remote URL",
			remoteURL: "git@github.com:test/repo.git",
			wantURL:   "git@github.com:test/repo.git",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for Git repository
			tmpDir := t.TempDir()

			// Initialize Git repository
			cmd := exec.Command("git", "-C", tmpDir, "init")
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to init git repo: %v", err)
			}

			// Add remote
			cmd = exec.Command("git", "-C", tmpDir, "remote", "add", "origin", tt.remoteURL)
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to add remote: %v", err)
			}

			// Test GetGitHubRemoteURL
			gotURL, err := GetGitHubRemoteURL(tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetGitHubRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotURL != tt.wantURL {
				t.Errorf("GetGitHubRemoteURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestGetGitHubRemoteURL_NoOrigin(t *testing.T) {
	// Create temporary directory for Git repository
	tmpDir := t.TempDir()

	// Initialize Git repository without remote
	cmd := exec.Command("git", "-C", tmpDir, "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Test GetGitHubRemoteURL - should fail with no origin
	_, err := GetGitHubRemoteURL(tmpDir)
	if err == nil {
		t.Error("GetGitHubRemoteURL() expected error for repo without origin, got nil")
	}
}
