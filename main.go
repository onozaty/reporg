package main

import (
	"fmt"
	"os"

	"github.com/onozaty/reporg/internal/git"
	"github.com/onozaty/reporg/internal/output"
	"github.com/onozaty/reporg/internal/search"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "dev"
)

// RepoContext contains information about a Git repository needed for generating URLs.
type RepoContext struct {
	Root   string // Absolute path to repository root
	Owner  string // GitHub owner
	Repo   string // Repository name
	Branch string // Branch name for URLs
}

var rootCmd = newRootCmd()

func newRootCmd() *cobra.Command {
	versionInfo := Version
	if Commit != "dev" {
		versionInfo = fmt.Sprintf("%s (commit: %s)", Version, Commit)
	}

	cmd := &cobra.Command{
		Use:   "reporg <pattern> <repoRoot1> [repoRoot2...]",
		Short: "Search git repositories with ripgrep and generate shareable references",
		Long: `reporg searches Git repositories using ripgrep and outputs results in TSV format.
Each result includes the local file path, matched line content, and GitHub URL reference.`,
		Version: versionInfo,
		Args:    cobra.MinimumNArgs(2),
		RunE:    run,
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")

	return cmd
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	pattern := args[0]
	repoPaths := args[1:]

	// Get flags
	outputFile, _ := cmd.Flags().GetString("output")

	// Validate and deduplicate repository paths
	uniqueRepos, err := git.DeduplicateRepoPaths(repoPaths)
	if err != nil {
		return fmt.Errorf("repository validation failed: %w", err)
	}

	var allResults []output.SearchResult

	// Process each repository
	for _, repoRoot := range uniqueRepos {
		// Get repository context
		repoCtx, err := getRepoContext(repoRoot)
		if err != nil {
			return fmt.Errorf("failed to get repository context for %s: %w", repoRoot, err)
		}

		// Execute search
		matches, err := search.SearchRepo(pattern, repoRoot)
		if err != nil {
			return fmt.Errorf("search failed in %s: %w", repoRoot, err)
		}

		// Convert matches to search results with GitHub URLs
		repository := fmt.Sprintf("%s/%s", repoCtx.Owner, repoCtx.Repo)
		for _, match := range matches {
			localPath := fmt.Sprintf("%s:%d", match.RelPath, match.LineNumber)
			githubURL := git.BuildGitHubFileURL(
				repoCtx.Owner,
				repoCtx.Repo,
				repoCtx.Branch,
				match.RelPath,
				match.LineNumber,
			)

			allResults = append(allResults, output.SearchResult{
				Repository:  repository,
				LocalPath:   localPath,
				MatchedLine: match.LineText,
				GitHubURL:   githubURL,
			})
		}
	}

	// Determine output destination
	writer := os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	// Write TSV output
	if err := output.WriteTSV(allResults, writer); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// getRepoContext retrieves repository context information needed for GitHub URL generation.
func getRepoContext(repoRoot string) (*RepoContext, error) {
	// Get GitHub remote URL
	remoteURL, err := git.GetGitHubRemoteURL(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Parse GitHub URL
	owner, repo, err := git.ParseGitHubURL(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("not a GitHub repository: %w", err)
	}

	// Determine branch name
	// Try to get current branch
	branch, err := git.GetCurrentBranch(repoRoot)
	if err != nil || branch == "" {
		// Fallback to "main"
		branch = "main"
	}

	return &RepoContext{
		Root:   repoRoot,
		Owner:  owner,
		Repo:   repo,
		Branch: branch,
	}, nil
}
