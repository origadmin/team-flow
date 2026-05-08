package boot

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	skillfs "github.com/origadmin/team-flow"
	"github.com/origadmin/team-flow/internal/toolchain"
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
	Short: "Initialize project with team-flow framework",
	Long: `Initialize project with team-flow skill and toolchain.

By default installs v2 (beads-native). Use --v1 for v1 (task-pool).

Steps:
  1. Check & install Python 3.10+ and pip (if missing)
  2. Check & install code-review-graph (if missing)
  3. Check & install beads bd CLI (if missing, v2 only)
  4. Create .team/ directory and install skill rules
  5. Build code graph (v2)
  6. Initialize beads database (v2)
  7. Configure MCP for AI IDE
  8. Verification`,
	RunE: runInit,
}

func init() {
	Cmd.Flags().BoolVar(&useV1, "v1", false, "Install v1 rules (task-pool workflow)")
	Cmd.Flags().BoolVar(&useV2, "v2", true, "Install v2 rules (beads-native workflow, default)")
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

## Paths

docs_path: _docs/%s/

## Toolchain

### Backend
pipeline: go test ./... | go build -o bin/app

### Frontend
pipeline: bun run test | bun run build

## Constraints
- No Chinese comments in code
- TDD required
`, filepath.Base(projectPath), version, filepath.Base(projectPath))

	projectMdPath := filepath.Join(teamDir, "project.md")
	if _, err := os.Stat(projectMdPath); err == nil && !force {
		fmt.Println("  .team/project.md already exists (use --force to overwrite)")
	} else {
		if err := os.WriteFile(projectMdPath, []byte(projectMd), 0644); err != nil {
			return fmt.Errorf("create project.md: %w", err)
		}
		fmt.Println("  ✓ .team/project.md created")
	}

	versionPath := filepath.Join(teamDir, "version")
	if _, err := os.Stat(versionPath); err == nil && !force {
		fmt.Println("  .team/version already exists (use --force to overwrite)")
	} else {
		if err := os.WriteFile(versionPath, []byte(version), 0644); err != nil {
			return fmt.Errorf("create version: %w", err)
		}
		fmt.Printf("  ✓ .team/version created (%s)\n", version)
	}

	installSkill(projectPath, skillfs.FS, version, force)

	skillPath := detectSkillPath(projectPath)
	generateDevMD(projectPath, skillPath)

	if version == "v2" {
		fmt.Println("\n━━━ Step 3: v2 Tools Init ━━━")

		pyPath = toolchain.FindPythonPath()

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
			bdPath := toolchain.FindBdPath()
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
		installMCPConfig(projectPath)
	}

	fmt.Println("\n━━━ Step 4: Verification ━━━")
	printVerify(projectPath, version)

	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║          Initialization Complete!        ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit .team/project.md")
	fmt.Printf("  2. AI reads %s/SKILL.md\n", skillPath)
	fmt.Println("  3. flow doctor (check anytime)")

	return nil
}

func ensurePython() string {
	pyPath := toolchain.FindPythonPath()
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

	return toolchain.FindPythonPath()
}

func ensurePip(pyPath string) {
	pipPath := toolchain.FindPipPath()
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
	if err == nil && strings.Contains(strings.ToLower(string(out)), "nodes") {
		fmt.Println("  ✓ code-review-graph installed")
		return
	}

	fmt.Println("  code-review-graph not found.")
	if autoYes || confirm("  Install code-review-graph via pip?") {
		pipPath := toolchain.FindPipPath()
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
	bdPath := toolchain.FindBdPath()
	if bdPath != "" {
		version, _ := exec.Command(bdPath, "--version").CombinedOutput()
		fmt.Printf("  ✓ beads found: %s (%s)\n", bdPath, strings.TrimSpace(string(version)))

		if !toolchain.IsBdOnPath() {
			fmt.Println("  bd is not on PATH. Adding...")
			if toolchain.AddBdToPath(bdPath) {
				fmt.Printf("  ✓ Added %s to PATH (current session + persistent)\n", filepath.Dir(bdPath))
			} else {
				fmt.Printf("  ⚠ Could not add to persistent PATH. Current session updated.\n")
				fmt.Printf("    Manual: Add %s to your PATH\n", filepath.Dir(bdPath))
			}
		}
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
				bdPath = toolchain.FindBdPath()
				if bdPath != "" {
					toolchain.AddBdToPath(bdPath)
					fmt.Printf("  ✓ Added %s to PATH\n", filepath.Dir(bdPath))
				} else {
					fmt.Println("  ⚠ bd installed but could not be located.")
					fmt.Println("    Restart your terminal and run: flow init")
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



func printVerify(projectPath, version string) {
	checks := []struct {
		label string
		path  string
	}{
		{".team/", filepath.Join(projectPath, ".team")},
		{".team/version", filepath.Join(projectPath, ".team", "version")},
	}

	ides := detectIDEs(projectPath)
	for _, ide := range ides {
		if ide.Detected {
			label := fmt.Sprintf("%s skills/team-flow/", ide.Name)
			checks = append(checks, struct {
				label string
				path  string
			}{label, filepath.Join(ide.SkillDir, "team-flow")})
		}
	}

	if len(ides) == 0 || !func() bool {
		for _, ide := range ides {
			if ide.Detected {
				return true
			}
		}
		return false
	}() {
		checks = append(checks, struct {
			label string
			path  string
		}{".team-flow/skill/", filepath.Join(projectPath, ".team-flow", "skill")})
	}

	checks = append(checks,
		struct {
			label string
			path  string
		}{".beads/", filepath.Join(projectPath, ".beads")},
		struct {
			label string
			path  string
		}{".code-review-graph/", filepath.Join(projectPath, ".code-review-graph")},
	)

	for _, ide := range ides {
		if ide.Detected {
			mcpPath := filepath.Join(ide.ConfigDir, "mcp.json")
			label := fmt.Sprintf("%s mcp.json", ide.Name)
			checks = append(checks, struct {
				label string
				path  string
			}{label, mcpPath})
		}
	}

	for _, c := range checks {
		if _, err := os.Stat(c.path); err == nil {
			fmt.Printf("  ✓ %-25s\n", c.label)
		} else {
			fmt.Printf("  ✗ %-25s\n", c.label)
		}
	}

	if toolchain.FindPythonPath() != "" {
		fmt.Println("  ✓ python")
	} else {
		fmt.Println("  ✗ python")
	}
	if toolchain.FindBdPath() != "" {
		fmt.Println("  ✓ bd (beads)")
	} else {
		fmt.Println("  ✗ bd (beads)")
	}
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func copyFromFS(fsys embed.FS, srcDir, dst string, overwrite bool, excludeDirs []string) (copied, skipped int, err error) {
	err = fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		var relPath string
		if srcDir == "." {
			relPath = path
			if relPath == "." {
				return nil
			}
		} else {
			if path == srcDir {
				return nil
			}
			relPath = strings.TrimPrefix(path, srcDir+"/")
		}

		parts := strings.Split(relPath, "/")
		for _, part := range parts {
			for _, exc := range excludeDirs {
				if part == exc {
					if d.IsDir() {
						return fs.SkipDir
					}
					return nil
				}
			}
		}

		dstPath := filepath.Join(dst, filepath.FromSlash(relPath))

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		if !overwrite {
			if _, err := os.Stat(dstPath); err == nil {
				skipped++
				return nil
			}
		}

		data, readErr := fsys.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		if mkdirErr := os.MkdirAll(filepath.Dir(dstPath), 0755); mkdirErr != nil {
			return mkdirErr
		}

		if writeErr := os.WriteFile(dstPath, data, 0644); writeErr != nil {
			return writeErr
		}
		copied++
		return nil
	})
	return
}

func detectSkillPath(projectPath string) string {
	ides := detectIDEs(projectPath)
	for _, ide := range ides {
		if ide.Detected {
			rel, err := filepath.Rel(projectPath, filepath.Join(ide.SkillDir, "team-flow"))
			if err == nil {
				return strings.ReplaceAll(rel, `\`, "/")
			}
		}
	}
	return ".team-flow/skill"
}

func generateDevMD(projectPath, skillPath string) {
	fmt.Println("  Generating IDE bridge files...")

	ides := detectIDEs(projectPath)
	if len(ides) == 0 {
		ides = append(ides, IDEInfo{
			Name:       "Trae",
			ConfigDir:  filepath.Join(projectPath, ".trae"),
			BridgePath: ".trae/rules/dev.md",
			BridgeFmt:  "trae",
			Detected:   true,
		})
	}

	for _, ide := range ides {
		if !ide.Detected {
			continue
		}

		content := generateBridgeFileContent(ide.BridgeFmt, skillPath)
		if content == "" {
			continue
		}

		bridgeAbsPath := filepath.Join(projectPath, ide.BridgePath)

		if _, err := os.Stat(bridgeAbsPath); err == nil && !force {
			fmt.Printf("  %s %s already exists (use --force to overwrite)\n", ide.Name, ide.BridgePath)
			continue
		}

		os.MkdirAll(filepath.Dir(bridgeAbsPath), 0755)
		if err := os.WriteFile(bridgeAbsPath, []byte(content), 0644); err != nil {
			fmt.Printf("  ⚠ %s %s write error: %v\n", ide.Name, ide.BridgePath, err)
		} else {
			fmt.Printf("  ✓ %s %s created\n", ide.Name, ide.BridgePath)
		}
	}
}

func generateBridgeFileContent(format, skillPath string) string {
	statusLine := "## Status Line (MANDATORY)\n\n" +
		"Every response MUST start with:\n\n" +
		"```\n" +
		"[Role: {role} | TaskPool: {status} | Phase: {phase} | Asset: {asset}]\n" +
		"```\n\n" +
		"- Role: Triage/TechLead/Dev/QA/PM/DevOps/Analysis/UIDesigner\n" +
		"- TaskPool: beads issue ID (e.g., `team-flow-7bd`) or ❌unread\n" +
		"- Phase: ready/analyze/design/implement/verify/review/-(N/A)\n" +
		"- Asset: directory path or -(N/A)\n\n" +
		"**When TaskPool = ❌unread, you MUST execute `bd ready` FIRST before any other action.**\n" +
		"This reads the task pool and assigns you a task. Never start working without a task ID.\n" +
		"If no tasks exist, create one with `bd create` before starting work.\n\n"

	skillRef := "Read `.team/version` for active version (v1 or v2).\n" +
		"Load and follow ALL rules in `" + skillPath + "/SKILL.md` as the entry point.\n"

	switch format {
	case "trae":
		return "# team-flow Rules\n\n" +
			statusLine +
			skillRef +
			"The SKILL.md contains complete collaboration rules including:\n" +
			"- Dispatch Guard (4 iron rules)\n" +
			"- Regression Guard (10 iron rules)\n" +
			"- Task Management (beads v2 / task-pool v1)\n" +
			"- Role Mapping and Sub-agent orchestration\n" +
			"- Three-Layer Gates\n" +
			"- Core Principle: Skill ≠ Project\n"
	case "cursor":
		return "---\n" +
			"description: team-flow AI collaboration rules\n" +
			"globs:\n" +
			"  - \"**/*\"\n" +
			"---\n\n" +
			"# team-flow Rules\n\n" +
			statusLine +
			skillRef
	case "claude":
		return "# team-flow Rules\n\n" +
			statusLine +
			skillRef
	case "openclaw":
		return "# team-flow Rules\n\n" +
			statusLine +
			skillRef
	default:
		return ""
	}
}

func installMCPConfig(projectPath string) {
	pyPath := toolchain.FindPythonPath()
	bdPath := toolchain.FindBdPath()

	mcpPythonCmd := "python"
	if pyPath != "" {
		mcpPythonCmd = pyPath
	}
	mcpBdCmd := "bd"
	if bdPath != "" {
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

	ides := detectIDEs(projectPath)
	installed := false

	for _, ide := range ides {
		if !ide.Detected {
			continue
		}

		mcpPath := filepath.Join(ide.ConfigDir, "mcp.json")
		if _, err := os.Stat(mcpPath); err == nil {
			fmt.Printf("  ✓ %s mcp.json already exists\n", ide.Name)
			installed = true
			continue
		}

		if autoYes || confirm(fmt.Sprintf("  Create %s mcp.json?", ide.Name)) {
			os.MkdirAll(ide.ConfigDir, 0755)
			if err := os.WriteFile(mcpPath, []byte(mcpConfig), 0644); err == nil {
				fmt.Printf("  ✓ %s mcp.json created\n", ide.Name)
				installed = true
			}
		}
	}

	if !installed {
		fmt.Println("  No IDE detected. MCP config not created.")
		fmt.Println("  You can create it manually for your IDE.")
	}

	fmt.Println("    Note: Restart AI IDE to load MCP servers")
}

type IDEInfo struct {
	Name       string
	SkillDir   string
	ConfigDir  string
	BridgePath string
	BridgeFmt  string
	Detected   bool
}

func detectIDEs(projectPath string) []IDEInfo {
	ides := []IDEInfo{
		{
			Name:       "Trae",
			SkillDir:   filepath.Join(projectPath, ".trae", "skills"),
			ConfigDir:  filepath.Join(projectPath, ".trae"),
			BridgePath: ".trae/rules/team-flow.md",
			BridgeFmt:  "trae",
		},
		{
			Name:       "Cursor",
			SkillDir:   filepath.Join(projectPath, ".cursor", "skills"),
			ConfigDir:  filepath.Join(projectPath, ".cursor"),
			BridgePath: ".cursor/rules/team-flow.mdc",
			BridgeFmt:  "cursor",
		},
		{
			Name:       "Claude",
			SkillDir:   filepath.Join(projectPath, ".claude", "skills"),
			ConfigDir:  filepath.Join(projectPath, ".claude"),
			BridgePath: ".claude/rules/team-flow.md",
			BridgeFmt:  "claude",
		},
		{
			Name:       "OpenClaw",
			SkillDir:   filepath.Join(projectPath, ".openclaw", "skills"),
			ConfigDir:  filepath.Join(projectPath, ".openclaw"),
			BridgePath: ".openclaw/rules/team-flow.md",
			BridgeFmt:  "openclaw",
		},
	}

	for i := range ides {
		if _, err := os.Stat(ides[i].ConfigDir); err == nil {
			ides[i].Detected = true
		}
	}

	return ides
}

func installSkill(projectPath string, fsys embed.FS, version string, overwrite bool) {
	fmt.Println("\n  Installing skill...")

	if npxPath, err := exec.LookPath("npx"); err == nil {
		fmt.Printf("  npx found: %s\n", npxPath)
		if autoYes || confirm("  Install skill via npx skills add? (recommended)") {
			ides := detectIDEs(projectPath)
			detectedIDE := ""
			for _, ide := range ides {
				if ide.Detected {
					detectedIDE = ide.Name
					break
				}
			}

			agentFlag := ""
			switch strings.ToLower(detectedIDE) {
			case "trae":
				agentFlag = "--agent trae"
			case "cursor":
				agentFlag = "--agent cursor"
			case "claude":
				agentFlag = "--agent claude"
			}

			npxCmd := fmt.Sprintf("npx skills add origadmin/team-flow %s", agentFlag)
			fmt.Printf("  Running: %s\n", npxCmd)

			parts := []string{"skills", "add", "origadmin/team-flow"}
			if agentFlag != "" {
				parts = append(parts, strings.Fields(agentFlag)...)
			}

			cmd := exec.Command(npxPath, parts...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("  ⚠ npx skills add failed: %v\n", err)
				fmt.Println("  Falling back to embedded copy...")
				installSkillFromFS(projectPath, fsys, version, overwrite)
			} else {
				fmt.Println("  ✓ Skill installed via npx")
				return
			}
		}
	}

	installSkillFromFS(projectPath, fsys, version, overwrite)
}

func installSkillFromFS(projectPath string, fsys embed.FS, version string, overwrite bool) {
	var srcDir string
	var excludeDirs []string

	if version == "v2" {
		srcDir = "team/v2"
	} else {
		srcDir = "team/v1"
	}

	sharedDirs := []string{"team/config", "team/templates", "team/lessons", "team/standards", "team/scripts"}

	ides := detectIDEs(projectPath)
	installed := false

	for _, ide := range ides {
		if !ide.Detected {
			continue
		}

		skillDir := filepath.Join(ide.SkillDir, "team-flow")

		skillEntry := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillEntry); err != nil || overwrite {
			entryData, readErr := fsys.ReadFile("team/SKILL.md")
			if readErr != nil {
				fmt.Printf("  ⚠ Read SKILL.md error: %v\n", readErr)
			} else {
				os.MkdirAll(skillDir, 0755)
				if writeErr := os.WriteFile(skillEntry, entryData, 0644); writeErr != nil {
					fmt.Printf("  ⚠ Write SKILL.md error: %v\n", writeErr)
				}
			}
		}

		if version == "v2" {
			for _, sharedDir := range sharedDirs {
				c, _, err := copyFromFS(fsys, sharedDir, skillDir, overwrite, excludeDirs)
				if err == nil && c > 0 {
					fmt.Printf("  ✓ %s shared: %d files from %s\n", ide.Name, c, sharedDir)
				}
			}
		}

		copied, skipped, err := copyFromFS(fsys, srcDir, skillDir, overwrite, excludeDirs)
		if err != nil {
			fmt.Printf("  ⚠ %s skill copy error: %v\n", ide.Name, err)
			continue
		}

		fmt.Printf("  ✓ %s (%s): Copied %d files", ide.Name, version, copied)
		if skipped > 0 {
			fmt.Printf(" (%d skipped)", skipped)
		}
		fmt.Println()
		installed = true
	}

	if !installed {
		fmt.Println("  No IDE configuration detected. Installing to .agents/skills/team-flow/")
		skillDir := filepath.Join(projectPath, ".agents", "skills", "team-flow")

		skillEntry := filepath.Join(skillDir, "SKILL.md")
		entryData, readErr := fsys.ReadFile("team/SKILL.md")
		if readErr == nil {
			os.MkdirAll(skillDir, 0755)
			os.WriteFile(skillEntry, entryData, 0644)
		}

		if version == "v2" {
			for _, sharedDir := range sharedDirs {
				copyFromFS(fsys, sharedDir, skillDir, overwrite, excludeDirs)
			}
		}

		copied, skipped, err := copyFromFS(fsys, srcDir, skillDir, overwrite, excludeDirs)
		if err != nil {
			fmt.Printf("  ⚠ Skill copy error: %v\n", err)
		} else {
			fmt.Printf("  ✓ (%s): Copied %d files", version, copied)
			if skipped > 0 {
				fmt.Printf(" (%d skipped)", skipped)
			}
			fmt.Println()
		}
	}
}
