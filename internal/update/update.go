package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/origadmin/team-flow/internal/updater"
	"github.com/origadmin/team-flow/internal/version"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "update",
	Short: "Update flow CLI to latest version",
	Long: `Check for updates and download/install the latest version of flow CLI.

This command will:
  1. Check GitHub for latest release
  2. Download the appropriate binary for your OS/arch
  3. Replace the current executable

Usage:
  flow update          Update to latest version
  flow update --check  Only check for updates, don't install`,
	RunE: runUpdate,
}

var (
	checkOnly bool
)

func init() {
	Cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         flow CLI Updater                 ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("  Current: %s\n\n", version.Info())

	fmt.Println("  Checking for updates...")
	updateCheck, err := updater.CheckForUpdate(version.Version)
	if err != nil {
		return fmt.Errorf("check update: %w", err)
	}

	if !updateCheck.HasUpdate {
		fmt.Println("  ✓ Already running latest version!")
		return nil
	}

	fmt.Printf("  ⬆ Update available: %s → %s\n", updateCheck.CurrentVersion, updateCheck.LatestVersion)

	if checkOnly {
		if updateCheck.DownloadURL != "" {
			fmt.Printf("    Download: %s\n", updateCheck.DownloadURL)
		}
		return nil
	}

	if updateCheck.DownloadURL == "" {
		return fmt.Errorf("no matching download found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("  Downloading: %s\n", updateCheck.DownloadURL)
	exePath, err := updater.GetCurrentExecutablePath()
	if err != nil {
		return fmt.Errorf("get exe path: %w", err)
	}
	fmt.Printf("  Target: %s\n", exePath)

	// Download new version
	if err := downloadAndReplace(updateCheck.DownloadURL, exePath); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("  ✓ Update completed successfully!")
	fmt.Printf("  Run 'flow --version' to confirm\n")
	return nil
}

func downloadAndReplace(url, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: %s", resp.Status)
	}

	tempPath := targetPath + ".tmp"
	out, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("write temp: %w", err)
	}

	backupPath := targetPath + ".bak"
	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Rename(targetPath, backupPath); err != nil {
			os.Remove(tempPath)
			return fmt.Errorf("backup old: %w", err)
		}
	}

	if err := os.Rename(tempPath, targetPath); err != nil {
		os.Rename(backupPath, targetPath)
		os.Remove(tempPath)
		return fmt.Errorf("replace: %w", err)
	}

	os.Remove(backupPath)
	return nil
}
