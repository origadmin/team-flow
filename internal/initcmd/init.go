package initcmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	useV1   bool
	useV2   bool
	force   bool
	autoYes bool
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project with team-flow framework (includes toolchain setup)",
	Long: `Initialize project and setup toolchain from scratch.

Steps:
  1. Check & install Python 3.10+ and pip (if missing)
  2. Check & install code-review-graph (if missing)
  3. Check & install beads bd CLI (if missing)
  4. Create .team/ directory and copy framework files
  5. Build code graph (v2)
  6. Initialize beads database (v2)
  7. Configure MCP for AI IDE (v2)
  8. Verification

Supports v1 (task-pool) and v2 (beads-native) modes.`,
	RunE: runInit,
}

func init() {
	Cmd.Flags().BoolVar(&useV1, "v1", false, "Use v1 (task-pool)")
	Cmd.Flags().BoolVar(&useV2, "v2", true, "Use v2 (beads-native)")
	Cmd.Flags().BoolVar(&force, "force", false, "Force overwrite existing files")
	Cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Auto-confirm all prompts")
}

func runInit(cmd *cobra.Command, args []string) error {
	version := "v2"
	if useV1 && !useV2 {
		version = "v1"
	}

	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║     team-flow Project Initialization     ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("  Project: %s\n", projectPath)
	fmt.Printf("  Version: %s\n", version)
	fmt.Printf("  OS:      %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	fmt.Println("━━━ Step 1: Toolchain Check ━━━")
	pyPath := ensurePython()
	ensurePip(pyPath)
	if version == "v2" {
		ensureCodeReviewGraph(pyPath)
		ensureBeads()
	}

	fmt.Println("\n━━━ Step 2: Project Files ━━━")
	teamDir := filepath.Join(projectPath, ".team")
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		return fmt.Errorf("create .team dir: %w", err)
	}

	projectMd := fmt.Sprintf(`## Project Configuration

- **Project Name**: %s
- **Team Version**: %s
- **Create Date**: TBD

## Toolchain

### Backend
pipeline: go test ./... | go build -o bin/app

### Frontend
pipeline: bun run test | bun run build

## Constraints
- No Chinese comments in code
- TDD required
`, filepath.Base(projectPath), version)

	projectMdPath := filepath.Join(teamDir, "project.md")
	if _, err := os.Stat(projectMdPath); err == nil && !force {
		fmt.Println("  .team/project.md already exists (use --force to overwrite)")
	} else {
		if err := os.WriteFile(projectMdPath, []byte(projectMd), 0644); err != nil {
			return fmt.Errorf("create project.md: %w", err)
		}
		fmt.Println("  ✓ .team/project.md created")
	}

	teamSrc, err := findTeamSource()
	if err != nil {
		fmt.Printf("  ⚠ Cannot find team/ source: %v\n", err)
		fmt.Println("  Please manually copy team/ framework files to your project")
	} else {
		frameworkDir := filepath.Join(projectPath, "framework", "_team")
		var excludeDirs []string
		if version == "v1" {
			excludeDirs = []string{"v2"}
		}

		copied, skipped, err := copyDir(teamSrc, frameworkDir, force, excludeDirs)
		if err != nil {
			fmt.Printf("  ⚠ Copy error: %v\n", err)
		} else {
			fmt.Printf("  ✓ Copied %d files to %s", copied, frameworkDir)
			if skipped > 0 {
				fmt.Printf(" (%d skipped, use --force to overwrite)", skipped)
			}
			fmt.Println()
		}
	}

	if version == "v2" {
		fmt.Println("\n━━━ Step 3: v2 Tools Init ━━━")

		pyPath = findPythonPath()

		fmt.Println("  Building code graph...")
		graphDB := filepath.Join(projectPath, ".code-review-graph")
		if _, err := os.Stat(graphDB); err == nil {
			fmt.Println("  ✓ Code graph already built")
		} else if pyPath != "" {
			cmd := exec.Command(pyPath, "-m", "code_review_graph", "build")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("  ⚠ Graph build failed: %v\n", err)
				fmt.Println("  You can build later: python -m code_review_graph build")
			} else {
				fmt.Println("  ✓ Code graph built")
			}
		} else {
			fmt.Println("  ⏭ Graph build skipped (Python not available)")
		}

		fmt.Println("  Initializing beads database...")
		beadsDir := filepath.Join(projectPath, ".beads")
		if _, err := os.Stat(beadsDir); err == nil {
			fmt.Println("  ✓ .beads already exists")
		} else {
			bdPath := findBdPath()
			if bdPath != "" {
				cmd := exec.Command(bdPath, "init")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Printf("  ⚠ bd init failed: %v\n", err)
					fmt.Println("  You can init later: bd init")
				} else {
					fmt.Println("  ✓ beads database initialized")
				}
			} else {
				fmt.Println("  ⏭ beads not installed, skip bd init")
				fmt.Println("    Restart your terminal and run: bd init")
			}
		}

		fmt.Println("  Configuring MCP...")
		traeMCP := filepath.Join(projectPath, ".trae", "mcp.json")
		if _, err := os.Stat(traeMCP); err == nil {
			fmt.Println("  ✓ .trae/mcp.json already exists")
		} else if autoYes || confirm("  Create .trae/mcp.json?") {
			mcpPythonCmd := "python"
			if pyPath != "" {
				mcpPythonCmd = pyPath
			}
			mcpBdCmd := "bd"
			if bdPath := findBdPath(); bdPath != "" {
				mcpBdCmd = bdPath
			}
			mcpConfig := fmt.Sprintf(`{
  "mcpServers": {
    "code-review-graph": {
      "command": "%s",
      "args": ["-m", "code_review_graph", "serve"],
      "env": {}
    },
    "beads": {
      "command": "%s",
      "args": ["mcp"],
      "env": {}
    }
  }
}`, strings.ReplaceAll(mcpPythonCmd, `\`, `\\`), strings.ReplaceAll(mcpBdCmd, `\`, `\\`))
			os.MkdirAll(filepath.Join(projectPath, ".trae"), 0755)
			if err := os.WriteFile(traeMCP, []byte(mcpConfig), 0644); err == nil {
				fmt.Println("  ✓ .trae/mcp.json created")
				fmt.Println("    Note: Restart AI IDE to load MCP servers")
			}
		}
	}

	fmt.Println("\n━━━ Step 4: Verification ━━━")
	printVerify(projectPath, version)

	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║          Initialization Complete!        ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println("\nNext steps:")
	if version == "v2" {
		fmt.Println("  1. Edit .team/project.md")
		fmt.Println("  2. AI reads framework/_team/v2/SKILL.md")
		fmt.Println("  3. flow doctor (check anytime)")
	} else {
		fmt.Println("  1. Edit .team/project.md")
		fmt.Println("  2. AI reads framework/_team/SKILL.md")
	}

	return nil
}

func ensurePython() string {
	pyPath := findPythonPath()
	if pyPath != "" {
		out, err := exec.Command(pyPath, "-c", "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')").CombinedOutput()
		if err == nil {
			v := strings.TrimSpace(string(out))
			parts := strings.Split(v, ".")
			major, minor := 0, 0
			if len(parts) >= 2 {
				fmt.Sscanf(parts[0], "%d", &major)
				fmt.Sscanf(parts[1], "%d", &minor)
			}
			if major >= 3 && minor >= 10 {
				fmt.Printf("  ✓ Python %s found: %s\n", v, pyPath)
				return pyPath
			}
			fmt.Printf("  ⚠ Python %s found but need 3.10+\n", v)
		}
		return pyPath
	}

	fmt.Println("  Python 3.10+ not found.")
	if autoYes || confirm("  Install Python?") {
		switch runtime.GOOS {
		case "windows":
			if _, err := exec.LookPath("winget"); err == nil {
				cmd := exec.Command("winget", "install", "Python.Python.3.12", "--accept-package-agreements", "--accept-source-agreements")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Println("  ⚠ winget install failed. Please install manually:")
					fmt.Println("    https://www.python.org/downloads/")
					fmt.Println("    ⚠ Check 'Add Python to PATH' during installation!")
				} else {
					fmt.Println("  ✓ Python installed. Restart terminal to apply PATH.")
				}
			} else {
				fmt.Println("  winget not found. Please install Python manually:")
				fmt.Println("    https://www.python.org/downloads/")
				fmt.Println("    ⚠ Check 'Add Python to PATH' during installation!")
			}
		case "darwin":
			if _, err := exec.LookPath("brew"); err == nil {
				cmd := exec.Command("brew", "install", "python@3.12")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			} else {
				fmt.Println("  Please install: brew install python@3.12")
			}
		default:
			cmd := exec.Command("sudo", "apt-get", "install", "-y", "python3", "python3-pip")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	} else {
		fmt.Println("  ⏭ Skipped. Some features will not work without Python.")
	}

	return findPythonPath()
}

func ensurePip(pyPath string) {
	pipPath := findPipPath()
	if pipPath != "" {
		fmt.Printf("  ✓ pip found: %s\n", pipPath)
		return
	}

	if pyPath != "" {
		out, _ := exec.Command(pyPath, "-m", "pip", "--version").CombinedOutput()
		if strings.Contains(string(out), "pip") {
			fmt.Println("  ✓ pip available via python -m pip")
			return
		}

		fmt.Println("  pip not found. Installing...")
		cmd := exec.Command(pyPath, "-m", "ensurepip", "--upgrade")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("  ⚠ ensurepip failed. Try: python -m ensurepip --upgrade")
		} else {
			fmt.Println("  ✓ pip installed")
		}
	}
}

func ensureCodeReviewGraph(pyPath string) {
	if pyPath == "" {
		fmt.Println("  ⏭ code-review-graph: skipped (Python not available)")
		return
	}

	out, err := exec.Command(pyPath, "-m", "code_review_graph", "status").CombinedOutput()
	if err == nil && strings.Contains(string(out), "nodes") {
		fmt.Println("  ✓ code-review-graph installed")
		return
	}

	fmt.Println("  code-review-graph not found.")
	if autoYes || confirm("  Install code-review-graph via pip?") {
		pipPath := findPipPath()
		installCmd := "pip"
		if pipPath != "" {
			installCmd = pipPath
		} else if pyPath != "" {
			installCmd = pyPath
		}

		var cmd *exec.Cmd
		if installCmd == pyPath {
			cmd = exec.Command(pyPath, "-m", "pip", "install", "code-review-graph")
		} else {
			cmd = exec.Command(installCmd, "install", "code-review-graph")
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("  ⚠ Install failed: %v\n", err)
			fmt.Println("    Try: pip install code-review-graph")
		} else {
			fmt.Println("  ✓ code-review-graph installed")
		}
	} else {
		fmt.Println("  ⏭ Skipped. flow graph commands will not work.")
	}
}

func ensureBeads() {
	bdPath := findBdPath()
	if bdPath != "" {
		version, _ := exec.Command(bdPath, "--version").CombinedOutput()
		fmt.Printf("  ✓ beads found: %s (%s)\n", bdPath, strings.TrimSpace(string(version)))
		return
	}

	fmt.Println("  beads (bd CLI) not found.")
	if autoYes || confirm("  Install beads?") {
		switch runtime.GOOS {
		case "windows":
			psCmd := `irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex`
			cmd := exec.Command("powershell", "-Command", psCmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("  ⚠ Auto install failed. Manual: irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 | iex")
			} else {
				fmt.Println("  ✓ beads installed")
				bdPath = findBdPath()
				if bdPath == "" {
					fmt.Println("  ⚠ bd installed but not on PATH yet.")
					fmt.Println("    Restart your terminal, or add to PATH:")
					fmt.Printf("    %s\\Programs\\bd\n", os.Getenv("LOCALAPPDATA"))
				}
			}
		case "darwin":
			cmd := exec.Command("brew", "install", "beads")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		default:
			fmt.Println("  Manual: curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/install.sh | bash")
		}
	} else {
		fmt.Println("  ⏭ Skipped. Task tracking will use task-pool.md fallback.")
	}
}

func findPythonPath() string {
	for _, name := range []string{"python", "python3"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	home, _ := os.UserHomeDir()
	localAppData := os.Getenv("LOCALAPPDATA")
	programFiles := os.Getenv("ProgramFiles")

	candidates := []string{}
	switch runtime.GOOS {
	case "windows":
		if localAppData != "" {
			candidates = append(candidates,
				filepath.Join(localAppData, "Programs", "Python", "Python311", "python.exe"),
				filepath.Join(localAppData, "Programs", "Python", "Python312", "python.exe"),
				filepath.Join(localAppData, "Programs", "Python", "Python313", "python.exe"),
				filepath.Join(localAppData, "Programs", "Python", "Python310", "python.exe"),
			)
		}
		if programFiles != "" {
			candidates = append(candidates,
				filepath.Join(programFiles, "Python311", "python.exe"),
				filepath.Join(programFiles, "Python312", "python.exe"),
				filepath.Join(programFiles, "Python313", "python.exe"),
			)
		}
		candidates = append(candidates,
			`C:\Python311\python.exe`,
			`C:\Python312\python.exe`,
			`C:\Python310\python.exe`,
		)
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, "AppData", "Local", "Programs", "Python", "Python311", "python.exe"),
				filepath.Join(home, "AppData", "Local", "Programs", "Python", "Python312", "python.exe"),
				filepath.Join(home, "AppData", "Local", "Programs", "Python", "Python313", "python.exe"),
			)
		}
	case "darwin":
		candidates = append(candidates,
			"/usr/bin/python3",
			"/usr/local/bin/python3",
			"/opt/homebrew/bin/python3",
		)
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "python3"),
			)
		}
	default:
		candidates = append(candidates,
			"/usr/bin/python3",
			"/usr/local/bin/python3",
		)
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "python3"),
			)
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

func findPipPath() string {
	for _, name := range []string{"pip", "pip3"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

func findBdPath() string {
	if path, err := exec.LookPath("bd"); err == nil {
		return path
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
				filepath.Join("/usr/local/bin", "bd"),
				filepath.Join("/opt/homebrew/bin", "bd"),
			)
		}
	default:
		if home != "" {
			candidates = append(candidates,
				filepath.Join(home, ".local", "bin", "bd"),
				filepath.Join("/usr/local/bin", "bd"),
			)
		}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

func printVerify(projectPath, version string) {
	checks := []struct {
		label string
		path  string
	}{
		{".team/", filepath.Join(projectPath, ".team")},
		{"framework/_team/", filepath.Join(projectPath, "framework", "_team")},
	}

	if version == "v2" {
		checks = append(checks,
			struct {
				label string
				path  string
			}{".beads/", filepath.Join(projectPath, ".beads")},
			struct {
				label string
				path  string
			}{".code-review-graph/", filepath.Join(projectPath, ".code-review-graph")},
			struct {
				label string
				path  string
			}{".trae/mcp.json", filepath.Join(projectPath, ".trae", "mcp.json")},
		)
	}

	for _, c := range checks {
		if _, err := os.Stat(c.path); err == nil {
			fmt.Printf("  ✓ %-25s\n", c.label)
		} else {
			fmt.Printf("  ✗ %-25s\n", c.label)
		}
	}

	if findPythonPath() != "" {
		fmt.Println("  ✓ python")
	} else {
		fmt.Println("  ✗ python")
	}
	if version == "v2" {
		if findBdPath() != "" {
			fmt.Println("  ✓ bd (beads)")
		} else {
			fmt.Println("  ✗ bd (beads)")
		}
	}
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func findTeamSource() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)
	wd, _ := os.Getwd()

	candidates := []string{
		filepath.Join(exeDir, "team"),
		filepath.Join(exeDir, "..", "team"),
		filepath.Join(exeDir, "..", "..", "team"),
		filepath.Join(wd, "team"),
		filepath.Join(wd, "..", "team"),
		filepath.Join(wd, "..", "..", "team"),
	}

	goModPath := findGoModuleRoot(wd)
	if goModPath != "" {
		candidates = append(candidates,
			filepath.Join(goModPath, "team"),
			filepath.Join(goModPath, "..", "team"),
		)

		entries, _ := os.ReadDir(goModPath)
		for _, e := range entries {
			if e.IsDir() && (strings.HasPrefix(e.Name(), "projects") || strings.HasPrefix(e.Name(), "team")) {
				candidates = append(candidates, filepath.Join(goModPath, e.Name(), "team-flow", "team"))
			}
		}
	}

	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if _, err := os.Stat(filepath.Join(abs, "SKILL.md")); err == nil {
			return abs, nil
		}
	}

	return "", fmt.Errorf("team/ directory with SKILL.md not found")
}

func findGoModuleRoot(startDir string) string {
	dir := startDir
	for i := 0; i < 20; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func copyDir(src, dst string, overwrite bool, excludeDirs []string) (copied, skipped int, err error) {
	err = filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, _ := filepath.Rel(src, path)
		if relPath == "." {
			return nil
		}

		parts := strings.Split(relPath, string(filepath.Separator))
		for _, part := range parts {
			for _, exc := range excludeDirs {
				if part == exc {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		if !overwrite {
			if _, err := os.Stat(dstPath); err == nil {
				skipped++
				return nil
			}
		}

		if err := copyFile(path, dstPath); err != nil {
			return err
		}
		copied++
		return nil
	})
	return
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	info, _ := os.Stat(src)
	if info != nil {
		os.Chmod(dst, info.Mode())
	}
	return nil
}
