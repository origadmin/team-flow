package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	syncSource    string
	syncRepo      string
	syncState     string
	syncLabel     string
	syncDryRun    bool
	syncOverwrite bool
)

var Cmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync external issues to local tasks",
	Long: `Sync issues from external sources (GitHub, etc.) to local task database.

Supports v1 (task-pool.md), v2 (beads), and v3 (beads + flow) backends.
Auto-detects project version from .team/version.

Examples:
  flow task sync --source github --repo owner/repo
  flow task sync --source github --repo owner/repo --state open
  flow task sync --source github --repo owner/repo --label bug
  flow task sync --source github --repo owner/repo --dry-run
  flow task sync --source github --repo owner/repo --overwrite`,
	RunE: runSync,
}

func init() {
	Cmd.Flags().StringVar(&syncSource, "source", "github", "Issue source (github)")
	Cmd.Flags().StringVar(&syncRepo, "repo", "", "Repository (owner/repo format, auto-detected from git remote)")
	Cmd.Flags().StringVar(&syncState, "state", "open", "Issue state filter (open/closed/all)")
	Cmd.Flags().StringVar(&syncLabel, "label", "", "Filter by label")
	Cmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Preview changes without writing")
	Cmd.Flags().BoolVar(&syncOverwrite, "overwrite", false, "Update existing synced tasks")
}

func runSync(cmd *cobra.Command, args []string) error {
	if syncRepo == "" {
		syncRepo = detectRepoFromGit()
		if syncRepo == "" {
			return fmt.Errorf("--repo is required (or run from a git repository with GitHub remote)")
		}
		fmt.Printf("  Auto-detected repo: %s\n", syncRepo)
	}

	if syncSource != "github" {
		return fmt.Errorf("unsupported source: %s (currently only 'github' is supported)", syncSource)
	}

	if !IsGhAvailable() {
		return fmt.Errorf("gh CLI not found. Install: https://cli.github.com")
	}

	version := detectVersion()
	if version == "" {
		return fmt.Errorf(".team/version not found. Run: flow init")
	}

	fmt.Printf("Syncing GitHub issues from %s (version: %s)\n", syncRepo, version)
	fmt.Println()

	opts := SyncOptions{
		Repo:      syncRepo,
		State:     syncState,
		Label:     syncLabel,
		Version:   version,
		Overwrite: syncOverwrite,
		DryRun:    syncDryRun,
	}

	if syncDryRun {
		fmt.Println("  [DRY RUN] No changes will be written")
		fmt.Println()
	}

	result, err := SyncFromGithub(opts)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("  Result: %s\n", result.Summary())

	if len(result.Created) > 0 {
		fmt.Println()
		fmt.Println("  Created:")
		for _, c := range result.Created {
			fmt.Printf("    ✓ %s\n", c)
		}
	}

	if len(result.Updated) > 0 {
		fmt.Println()
		fmt.Println("  Updated:")
		for _, u := range result.Updated {
			fmt.Printf("    ↻ %s\n", u)
		}
	}

	if len(result.Skipped) > 0 {
		fmt.Println()
		fmt.Println("  Skipped (already synced):")
		for _, s := range result.Skipped {
			fmt.Printf("    - %s\n", s)
		}
	}

	if len(result.Failed) > 0 {
		fmt.Println()
		fmt.Println("  Failed:")
		for _, f := range result.Failed {
			fmt.Printf("    ✗ %s\n", f)
		}
	}

	fmt.Println()

	if !syncDryRun && len(result.Created) > 0 {
		fmt.Println("  Next: Run 'flow task list' to see synced tasks")
		fmt.Println("        Triage will analyze and classify them on next session")
	}

	return nil
}

func detectVersion() string {
	cwd, _ := os.Getwd()
	versionFile := filepath.Join(cwd, ".team", "version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func detectRepoFromGit() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	gitDir := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return ""
	}

	remotes, err := readGitRemotes(cwd)
	if err != nil {
		return ""
	}

	for _, remote := range remotes {
		if repo := parseGithubRepo(remote); repo != "" {
			return repo
		}
	}

	return ""
}

func readGitRemotes(dir string) ([]string, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git not found")
	}

	cmd := exec.Command(gitPath, "-C", dir, "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var remotes []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "(fetch)") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				remotes = append(remotes, fields[1])
			}
		}
	}
	return remotes, nil
}

func parseGithubRepo(remoteURL string) string {
	remoteURL = strings.TrimSpace(remoteURL)

	if strings.HasPrefix(remoteURL, "git@github.com:") {
		repo := strings.TrimPrefix(remoteURL, "git@github.com:")
		repo = strings.TrimSuffix(repo, ".git")
		return repo
	}

	if strings.Contains(remoteURL, "github.com") {
		parts := strings.SplitN(remoteURL, "github.com", 2)
		if len(parts) == 2 {
			repo := strings.TrimPrefix(parts[1], ":")
			repo = strings.TrimPrefix(repo, "/")
			repo = strings.TrimSuffix(repo, ".git")
			return repo
		}
	}

	return ""
}
