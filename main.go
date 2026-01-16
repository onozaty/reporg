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
	cmd.Flags().BoolP("ignore-case", "i", false, "Case-insensitive search")
	cmd.Flags().StringSliceP("glob", "g", nil, "Include or exclude files matching glob pattern (can be specified multiple times)")
	cmd.Flags().Bool("hidden", false, "Search hidden files and directories")
	cmd.Flags().BoolP("fixed-strings", "F", false, "Treat pattern as literal string, not regex")
	cmd.Flags().IntP("max-line-length", "m", 0, "Maximum line length in output (0 = no limit). Lines longer than this will be truncated with '...'")
	cmd.Flags().StringP("encoding", "E", "auto", "Text encoding to use for reading files (e.g., utf-8, shift_jis, euc-jp, iso-2022-jp). Default: auto (UTF-8/UTF-16 BOM detection)")

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
	ignoreCase, _ := cmd.Flags().GetBool("ignore-case")
	globs, _ := cmd.Flags().GetStringSlice("glob")
	hidden, _ := cmd.Flags().GetBool("hidden")
	fixedStrings, _ := cmd.Flags().GetBool("fixed-strings")
	maxLineLength, _ := cmd.Flags().GetInt("max-line-length")
	encoding, _ := cmd.Flags().GetString("encoding")

	// Validate and deduplicate repository paths
	uniqueRepos, err := git.DeduplicateRepoPaths(repoPaths)
	if err != nil {
		return fmt.Errorf("repository validation failed: %w", err)
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

	// Create TSV writer
	tsvWriter := output.NewTSVWriter(writer)

	// Process each repository
	for _, repoRoot := range uniqueRepos {
		// Get repository context
		repoCtx, err := getRepoContext(repoRoot)
		if err != nil {
			return fmt.Errorf("failed to get repository context for %s: %w", repoRoot, err)
		}

		repository := fmt.Sprintf("%s/%s", repoCtx.Owner, repoCtx.Repo)

		// Create search options
		searchOpts := search.SearchOptions{
			IgnoreCase:    ignoreCase,
			Globs:         globs,
			Hidden:        hidden,
			FixedStrings:  fixedStrings,
			MaxLineLength: maxLineLength,
			Encoding:      encoding,
		}

		// Execute search with callback for real-time output
		err = search.SearchRepo(pattern, repoRoot, searchOpts, func(match search.Match) error {
			// Convert match to search result and write immediately
			localPath := fmt.Sprintf("%s:%d", match.RelPath, match.LineNumber)
			githubURL := git.BuildGitHubFileURL(
				repoCtx.Owner,
				repoCtx.Repo,
				repoCtx.Branch,
				match.RelPath,
				match.LineNumber,
			)

			result := output.SearchResult{
				Repository:  repository,
				LocalPath:   localPath,
				MatchedLine: match.LineText,
				GitHubURL:   githubURL,
			}

			return tsvWriter.Write(result)
		})
		if err != nil {
			return fmt.Errorf("search failed in %s: %w", repoRoot, err)
		}
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
