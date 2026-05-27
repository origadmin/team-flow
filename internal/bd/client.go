package bd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FindWorkspaceRoot finds the workspace root directory by looking for .beads or .git
func FindWorkspaceRoot(startDir string) (string, error) {
	dir := startDir
	for {
		// Check for .beads directory first
		if _, err := os.Stat(filepath.Join(dir, ".beads")); err == nil {
			return dir, nil
		}
		// Check for .git directory as fallback
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		// Move up to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding
			return "", fmt.Errorf("no workspace root found. Run from a directory with .beads/ or .git/")
		}
		dir = parent
	}
}

// RunInWorkspace runs a bd command in the workspace root
func RunInWorkspace(startDir string, args ...string) (string, error) {
	workspaceRoot, err := FindWorkspaceRoot(startDir)
	if err != nil {
		return "", err
	}
	
	path := MustPath()
	
	cmd := exec.Command(path, args...)
	cmd.Dir = workspaceRoot // Set the working directory to workspace root
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		if len(output) > 0 {
			return string(output), fmt.Errorf("task-db %s failed: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("task-db %s failed: %w", strings.Join(args, " "), err)
	}
	
	return string(output), nil
}

// RunQuietInWorkspace runs a bd command quietly in the workspace root
func RunQuietInWorkspace(startDir string, args ...string) (string, error) {
	workspaceRoot, err := FindWorkspaceRoot(startDir)
	if err != nil {
		return "", err
	}
	
	path := MustPath()
	
	cmd := exec.Command(path, args...)
	cmd.Dir = workspaceRoot
	output, err := cmd.Output()
	
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(exitErr.Stderr), fmt.Errorf("task-db %s failed: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("task-db %s failed: %w", strings.Join(args, " "), err)
	}
	
	return string(output), nil
}

var (
	bdPath     string
	bdPathOnce bool
)

func FindPath() string {
	if bdPath != "" {
		return bdPath
	}

	if path, err := exec.LookPath("bd"); err == nil {
		bdPath = path
		return bdPath
	}

	home, _ := os.UserHomeDir()
	localAppData := os.Getenv("LOCALAPPDATA")

	candidates := []string{}
	switch runtime.GOOS {
	case "windows":
		if localAppData != "" {
			candidates = append(candidates,
				filepath.Join(localAppData, "Programs", "bd", "bd.exe"),
				filepath.Join(localAppData, "bd", "bd.exe"),
			)
		}
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, "AppData", "Local", "Programs", "bd", "bd.exe"),
				filepath.Join(home, ".local", "bin", "bd.exe"),
			)
		}
	case "darwin":
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "bd"),
				"/usr/local/bin/bd",
				"/opt/homebrew/bin/bd",
			)
		}
	default:
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "bd"),
				"/usr/local/bin/bd",
			)
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			bdPath = c
			return bdPath
		}
	}

	return ""
}

func IsAvailable() bool {
	return FindPath() != ""
}

func EnsureOnPath() error {
	path := FindPath()
	if path == "" {
		return fmt.Errorf("bd CLI not found")
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	currentPath := os.Getenv("PATH")
	if strings.Contains(currentPath, filepath.Dir(path)) {
		return nil
	}

	newPath := currentPath + separator + filepath.Dir(path)
	os.Setenv("PATH", newPath)

	if runtime.GOOS == "windows" {
		bdDir := filepath.Dir(path)
		psScript := fmt.Sprintf(
			`$current = [Environment]::GetEnvironmentVariable('Path', 'User'); if ($current -notlike '*%s*') { [Environment]::SetEnvironmentVariable('Path', $current + ';%s', 'User') }`,
			bdDir, bdDir,
		)
		cmd := exec.Command("powershell", "-Command", psScript)
		cmd.Run()
	}

	return nil
}

func MustPath() string {
	path := FindPath()
	if path == "" {
		fmt.Println("\n⚠ beads (bd CLI) not found.")
		fmt.Println("  Install: https://github.com/steveyegge/beads")
		os.Exit(1)
	}
	return path
}

func Run(args ...string) (string, error) {
	path := MustPath()

	cmd := exec.Command(path, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) > 0 {
			return string(output), fmt.Errorf("task-db %s failed: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("task-db %s failed: %w", strings.Join(args, " "), err)
	}

	return string(output), nil
}

func RunQuiet(args ...string) (string, error) {
	path := MustPath()

	cmd := exec.Command(path, args...)
	output, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(exitErr.Stderr), fmt.Errorf("task-db %s failed: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("task-db %s failed: %w", strings.Join(args, " "), err)
	}

	return string(output), nil
}

func Command(args ...string) *exec.Cmd {
	path := MustPath()
	return exec.Command(path, args...)
}

func Install() error {
	fmt.Println("Installing beads...")

	var installCmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		installCmd = exec.Command("powershell", "-Command", "irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex")
	default:
		installCmd = exec.Command("sh", "-c", "curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/install.sh | bash")
	}

	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	return nil
}

func InstallQuiet() error {
	var installCmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		installCmd = exec.Command("powershell", "-Command", "irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex")
	default:
		installCmd = exec.Command("sh", "-c", "curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/install.sh | bash")
	}

	return installCmd.Run()
}
