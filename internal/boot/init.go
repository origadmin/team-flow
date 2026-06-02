package boot

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	skillfs "github.com/origadmin/team-flow"
	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/flow"
	"github.com/origadmin/team-flow/internal/ide"
	"github.com/origadmin/team-flow/internal/skill"
	"github.com/origadmin/team-flow/internal/toolchain"
	"github.com/spf13/cobra"
)

var (
	useV1     bool
	useV2     bool
	useV3     bool
	force     bool
	autoYes   bool
	initFlow  string
	initTeam  string
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project with team-flow framework",
	Long: `Initialize project with team-flow skill and toolchain.

By default installs v3 (flow-driven). Use --v1 for v1 (task-pool). Use --v2 for v2 (beads-native).

Steps:
  1. Check & install Python 3.10+ and pip (if missing)
  2. Check & install code-review-graph (if missing)
  3. Check & install task database (if missing, v2/v3)
  4. Create .team/ directory and install skill rules
  5. Build code graph (v2)
  6. Initialize task database (v2/v3)
  7. Configure MCP for AI IDE
  8. Verification`,
	RunE: runInit,
}

func init() {
	Cmd.Flags().BoolVar(&useV1, "v1", false, "Install v1 rules (task-pool workflow)")
	Cmd.Flags().BoolVar(&useV2, "v2", false, "Install v2 rules (beads-native workflow)")
	Cmd.Flags().BoolVar(&useV3, "v3", true, "Install v3 rules (flow-driven workflow, default)")
	Cmd.Flags().BoolVar(&force, "force", false, "Force overwrite existing files")
	Cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Auto-confirm all prompts")
	Cmd.Flags().StringVar(&initFlow, "flow", "", "Default flow to bind (v3 only, e.g., dev-flow, novel-flow)")
	Cmd.Flags().StringVar(&initTeam, "team", "", "Team template to install (v3 only, e.g., dev-team, content-team)")
}

func runInit(cmd *cobra.Command, args []string) error {
	version := "v3"
	if useV1 && !useV2 && !useV3 {
		version = "v1"
	}
	if useV2 && !useV1 && !useV3 {
		version = "v2"
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
	if version == "v3" {
		ensureFlowBinary(projectPath)
	}

	fmt.Println("\n━━━ Step 2: Project Files ━━━")
	teamDir := filepath.Join(projectPath, ".team")
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		return fmt.Errorf("create .team dir: %w", err)
	}
	toolsYamlPath := filepath.Join(teamDir, "tools.yaml")
	toolsYamlContent := `beads:
  name: "beads"
  description: "Task management with Dolt git-native storage"
  binary:
    path: "bd"
    auto_download:
      enabled: true
      url: "https://github.com/steveyegge/beads/releases/latest"
      method: "github"
  install:
    enabled: true
    url: "https://github.com/steveyegge/beads/releases/latest"
    method: "github"
  commands:
    init:
      name: "init"
      description: "Initialize beads database"
      args: ["init"]
    create:
      name: "create"
      description: "Create a new task"
      args: ["create"]
    list:
      name: "list"
      description: "List all tasks"
      args: ["list"]
    show:
      name: "show"
      description: "Show task details"
      args: ["show"]
    update:
      name: "update"
      description: "Update task"
      args: ["update"]
    close:
      name: "close"
      description: "Close a task"
      args: ["close"]
    ready:
      name: "ready"
      description: "Show tasks ready for pickup"
      args: ["ready"]
    stats:
      name: "stats"
      description: "Show statistics"
      args: ["stats"]
    dolt:
      name: "dolt"
      description: "Dolt git operations"
      args: ["dolt"]
    dep:
      name: "dep"
      description: "Manage dependencies"
      args: ["dep"]
    config:
      name: "config"
      description: "Configuration management"
      args: ["config"]
    children:
      name: "children"
      description: "List child issues"
      args: ["children"]
`
	if _, err := os.Stat(toolsYamlPath); err == nil && !force {
		fmt.Println("  .team/tools.yaml already exists (use --force to overwrite)")
	} else {
		if err := os.WriteFile(toolsYamlPath, []byte(toolsYamlContent), 0644); err != nil {
			return fmt.Errorf("create tools.yaml: %w", err)
		}
		fmt.Println("  ✓ .team/tools.yaml created")
	}

	projectName := filepath.Base(projectPath)

	projectYamlContent := fmt.Sprintf(`name: %s
version: %s
default_flow: %s

paths:
  docs_internal: _docs/%s/
  docs_external: docs/

toolchain:
  backend:
    language: go
    pipeline: go test ./... | go build -o bin/app
  frontend:
    language: typescript
    pipeline: bun run test | bun run build
    framework: react
    ui_library: shadcn
    css_framework: tailwind
    package_manager: bun
`, projectName, version, defaultFlowValue(version), projectName)

	projectYamlPath := filepath.Join(teamDir, "project.yaml")
	if _, err := os.Stat(projectYamlPath); err == nil && !force {
		fmt.Println("  .team/project.yaml already exists (use --force to overwrite)")
	} else {
		if err := os.WriteFile(projectYamlPath, []byte(projectYamlContent), 0644); err != nil {
			return fmt.Errorf("create project.yaml: %w", err)
		}
		fmt.Println("  ✓ .team/project.yaml created")
	}

	constraintsMd := `# Constraints

- No Chinese comments in code
- TDD required
`
	constraintsPath := filepath.Join(teamDir, "constraints.md")
	if _, err := os.Stat(constraintsPath); err == nil && !force {
		fmt.Println("  .team/constraints.md already exists (use --force to overwrite)")
	} else {
		if err := os.WriteFile(constraintsPath, []byte(constraintsMd), 0644); err != nil {
			return fmt.Errorf("create constraints.md: %w", err)
		}
		fmt.Println("  ✓ .team/constraints.md created")
	}

	docsFallbackDir := filepath.Join(teamDir, "docs")
	os.MkdirAll(filepath.Join(docsFallbackDir, "lessons"), 0755)
	os.MkdirAll(filepath.Join(docsFallbackDir, "sessions"), 0755)
	os.MkdirAll(filepath.Join(docsFallbackDir, "conventions"), 0755)

	projectMdPath := filepath.Join(teamDir, "project.md")
	if _, err := os.Stat(projectMdPath); err == nil && !force {
		fmt.Println("  .team/project.md already exists (use --force to overwrite)")
	} else {
		projectMd := fmt.Sprintf(`## Project Configuration

- **Project Name**: %s
- **Team Version**: %s
- **Create Date**: TBD

## Paths

docs_internal: _docs/%s/
docs_external: docs/
%s
## Toolchain

### Backend
pipeline: go test ./... | go build -o bin/app

### Frontend
pipeline: bun run test | bun run build

## Constraints
- No Chinese comments in code
- TDD required
`, projectName, version, projectName, defaultFlowLine(version))

		if err := os.WriteFile(projectMdPath, []byte(projectMd), 0644); err != nil {
			return fmt.Errorf("create project.md: %w", err)
		}
		fmt.Println("  ✓ .team/project.md created (legacy, prefer project.yaml)")
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
	generateDevMD(projectPath, skillPath, version)

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

		fmt.Println("  Initializing task database...")
		beadsDir := filepath.Join(projectPath, ".beads")
		if _, err := os.Stat(beadsDir); err == nil {
			fmt.Println("  ✓ Task database already exists")
		} else {
			if bd.IsAvailable() {
				if _, err := bd.Run("init"); err != nil {
					fmt.Printf("  ⚠ Task database init failed: %v\n", err)
					fmt.Println("  You can init later: flow task init")
				} else {
					fmt.Println("  ✓ Task database initialized")
				}
			} else {
				fmt.Println("  ⏭ Task database not available, skip init")
				fmt.Println("    Restart your terminal and run: flow task init")
			}
		}

		fmt.Println("  Configuring MCP...")
		installMCPConfig(projectPath)
	}

	if version == "v3" {
		fmt.Println("\n━━━ Step 3: v3 Team Setup ━━━")

		teamID := initTeam
		var teamInfo *teamMeta

		if teamID != "" {
			teamInfo = loadTeamFromFS(skillfs.FS, teamID)
			if teamInfo == nil {
				fmt.Printf("  ⚠ Team '%s' not found in preset teams\n", teamID)
			}
		}

		if teamInfo == nil && initFlow != "" {
			teamInfo = findTeamByFlow(skillfs.FS, initFlow)
			if teamInfo != nil {
				fmt.Printf("  Flow '%s' found in team '%s'\n", initFlow, teamInfo.ID)
			}
		}

		if teamInfo == nil {
			teamInfo = selectTeam(skillfs.FS)
		}

		if teamInfo != nil {
			fmt.Printf("  Selected team: %s (%s)\n", teamInfo.Name, teamInfo.ID)

			fmt.Println("\n  Installing team definition...")
			if err := installTeamDefinition(skillfs.FS, teamInfo.ID, teamDir, force); err != nil {
				fmt.Printf("  ⚠ Failed to install team definition: %v\n", err)
			}

			flowName := initFlow
			if flowName == "" {
				flowName = teamInfo.DefaultFlow
			}

			flowFound := false
			for _, f := range teamInfo.Flows {
				if f.ID == flowName {
					flowFound = true
					break
				}
			}
			if !flowFound && len(teamInfo.Flows) > 0 {
				flowName = teamInfo.DefaultFlow
			}

			projectMdPath := filepath.Join(teamDir, "project.md")
			if data, err := os.ReadFile(projectMdPath); err == nil {
				content := string(data)
				if !strings.Contains(content, "default_flow:") {
					content += "\ndefault_flow: " + flowName + "\n"
				} else {
					re := regexp.MustCompile(`default_flow:\s*\S+`)
					content = re.ReplaceAllString(content, "default_flow: "+flowName)
				}
				os.WriteFile(projectMdPath, []byte(content), 0644)
			}

			projectYamlPath := filepath.Join(teamDir, "project.yaml")
			if cfg, err := config.LoadProjectConfig(projectPath); err == nil {
				cfg.DefaultFlow = flowName
				found := false
				for _, f := range cfg.Flows {
					if f.ID == flowName {
						found = true
						break
					}
				}
				if !found {
					cfg.Flows = append(cfg.Flows, config.ProjectFlow{
						ID:     flowName,
						Source: "preset",
					})
				}
				config.SaveProjectConfig(projectPath, cfg)
			} else {
				yamlContent := fmt.Sprintf("name: %s\nversion: v3\ndefault_flow: %s\n", filepath.Base(projectPath), flowName)
				os.WriteFile(projectYamlPath, []byte(yamlContent), 0644)
			}
			fmt.Printf("  ✓ default_flow set to %s\n", flowName)

			fmt.Println("\n  Installing team flows...")
			projectFlowsDir := filepath.Join(projectPath, ".team", "flows")
			os.MkdirAll(projectFlowsDir, 0755)
			copied := installTeamFlows(skillfs.FS, teamInfo.ID, projectFlowsDir, force)
			if copied > 0 {
				fmt.Printf("  ✓ Copied %d flows from team '%s' to .team/flows/\n", copied, teamInfo.ID)
			}

			fmt.Println("\n  Resolving skills...")
			teamDef := loadTeamDefinition(skillfs.FS, teamInfo.ID)
			if teamDef != nil && len(teamDef.SkillTags) > 0 {
				resolved := skill.ResolveSkills(teamDef.SkillTags)
				sf := skill.NewSkillsFile(teamInfo.ID, teamDef.SkillTags, resolved)
				skillsYamlPath := filepath.Join(teamDir, "skills.yaml")
				if _, err := os.Stat(skillsYamlPath); err != nil || force {
					if err := skill.SaveSkillsFile(skillsYamlPath, sf); err != nil {
						fmt.Printf("  ⚠ Failed to write .team/skills.yaml: %v\n", err)
					} else {
						fmt.Printf("  ✓ .team/skills.yaml created (%d skills from %d tags)\n", len(resolved), len(teamDef.SkillTags))
					}
				} else {
					fmt.Println("  .team/skills.yaml already exists (use --force to overwrite)")
				}
			} else {
				fmt.Println("  No skill_tags found in team definition, skipping skill resolution")
			}
		} else {
			fmt.Println("  No team selected. Set one later with:")
			fmt.Println("    flow init --v3 --team dev-team")
			fmt.Println("    Or create a new team: AI loads team-flow-v3-create skill")
		}

		fmt.Println("  Initializing task database...")
		beadsDir := filepath.Join(projectPath, ".beads")
		if _, err := os.Stat(beadsDir); err == nil {
			fmt.Println("  ✓ Task database already exists")
		} else if bd.IsAvailable() {
			if _, err := bd.Run("init"); err != nil {
				fmt.Printf("  ⚠ Task database init failed: %v\n", err)
			} else {
				fmt.Println("  ✓ Task database initialized")
			}
		} else {
			fmt.Println("  ⏭ Task database not available, skip init")
		}
	}

	fmt.Println("\n━━━ Step 4: Verification ━━━")
	printVerify(projectPath, version)

	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║          Initialization Complete!        ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println("\nNext steps:")
	if version == "v3" {
		fmt.Println("  1. flow proc list          — See available flows")
		fmt.Println("  2. flow proc run            — Start the flow engine")
		fmt.Println("  3. AI reads v3 SKILL.md     — Follows exec skill protocol")
		fmt.Println("  4. Switch team              — flow init --v3 --team <team-id>")
		fmt.Println("  5. Create custom team       — AI loads team-flow-v3-create skill")
	} else {
		fmt.Println("  1. Edit .team/project.md")
		fmt.Printf("  2. AI reads %s/SKILL.md\n", skillPath)
		fmt.Println("  3. flow doctor (check anytime)")
	}

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
	if bd.IsAvailable() {
		version, _ := exec.Command(bd.FindPath(), "--version").CombinedOutput()
		fmt.Printf("  ✓ Task database found: %s (%s)\n", bd.FindPath(), strings.TrimSpace(string(version)))

		if err := bd.EnsureOnPath(); err != nil {
			fmt.Printf("  ⚠ Could not add to persistent PATH. Current session updated.\n")
			fmt.Printf("    Manual: Add %s to your PATH\n", filepath.Dir(bd.FindPath()))
		} else {
			fmt.Printf("  ✓ Added %s to PATH\n", filepath.Dir(bd.FindPath()))
		}
		return
	}

	fmt.Println("  Task database not found.")
	if autoYes || confirm("  Install task database?") {
		if err := bd.Install(); err != nil {
			fmt.Println("  ⚠ Auto install failed.")
			fmt.Println("    Manual: https://github.com/steveyegge/beads")
		} else {
			fmt.Println("  ✓ Task database installed")
			if err := bd.EnsureOnPath(); err != nil {
				fmt.Println("  ⚠ Installed but could not be located.")
				fmt.Println("    Restart your terminal and run: flow init")
			} else {
				fmt.Printf("  ✓ Added %s to PATH\n", filepath.Dir(bd.FindPath()))
			}
		}
	} else {
		fmt.Println("  ⏭ Skipped. Task tracking will use file-based fallback.")
	}
}

func ensureFlowBinary(projectPath string) {
	scriptsDir := filepath.Join(projectPath, "scripts")
	flowExeName := "flow"
	if runtime.GOOS == "windows" {
		flowExeName = "flow.exe"
	}
	targetPath := filepath.Join(scriptsDir, flowExeName)

	if _, err := os.Stat(targetPath); err == nil {
		fmt.Printf("  ✓ Flow binary found: %s\n", targetPath)
		return
	}

	if exePath, err := exec.LookPath("flow"); err == nil {
		if err := os.MkdirAll(scriptsDir, 0755); err != nil {
			fmt.Printf("  ⚠ Could not create scripts/ directory: %v\n", err)
			return
		}
		src, err := os.Open(exePath)
		if err != nil {
			fmt.Printf("  ⚠ Could not read flow binary: %v\n", err)
			return
		}
		defer src.Close()

		dst, err := os.Create(targetPath)
		if err != nil {
			fmt.Printf("  ⚠ Could not create %s: %v\n", targetPath, err)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			fmt.Printf("  ⚠ Could not copy flow binary: %v\n", err)
			return
		}

		fmt.Printf("  ✓ Flow binary copied to: %s\n", targetPath)
		return
	}

	fmt.Println("  ⚠ Flow binary not found in PATH.")
	fmt.Println("    Build from source: cd projects/team-flow && go build -o ../../scripts/flow.exe ./cmd/flow/")
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
		}{".task-db/", filepath.Join(projectPath, ".beads")},
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
	if bd.IsAvailable() {
		fmt.Println("  ✓ flow task (task database)")
	} else {
		fmt.Println("  ✗ flow task (task database)")
	}
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func defaultFlowLine(version string) string {
	name := defaultFlowValue(version)
	if name == "" {
		return ""
	}
	return fmt.Sprintf("default_flow: %s\n", name)
}

func defaultFlowValue(version string) string {
	if version != "v3" {
		return ""
	}
	flowName := initFlow
	if flowName == "" {
		if initTeam != "" {
			if team := loadTeamFromFS(skillfs.FS, initTeam); team != nil {
				flowName = team.DefaultFlow
			}
		}
		if flowName == "" {
			flowName = "dev-flow"
		}
	}
	return flowName
}

type teamMeta struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	NameZh       string `json:"name_zh"`
	Description  string `json:"description"`
	DescriptionZh string `json:"description_zh"`
	DefaultFlow  string `json:"default_flow"`
	Flows        []struct {
		ID          string `json:"id"`
		File        string `json:"file"`
		Description string `json:"description"`
		Default     bool   `json:"default"`
	} `json:"flows"`
	Tags []string `json:"tags"`
}

func loadTeamFromFS(fsys embed.FS, teamID string) *teamMeta {
	data, err := fsys.ReadFile(skillfs.OrgsRoot + "/" + teamID + "/team.json")
	if err != nil {
		return nil
	}
	var team teamMeta
	if json.Unmarshal(data, &team) != nil {
		return nil
	}
	return &team
}

func loadTeamDefinition(fsys embed.FS, teamID string) *flow.TeamDefinition {
	data, err := fsys.ReadFile(skillfs.OrgsRoot + "/" + teamID + "/team.json")
	if err != nil {
		return nil
	}
	var team flow.TeamDefinition
	if json.Unmarshal(data, &team) != nil {
		return nil
	}
	return &team
}

func findTeamByFlow(fsys embed.FS, flowID string) *teamMeta {
	entries, err := fs.ReadDir(fsys, skillfs.OrgsRoot)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, readErr := fsys.ReadFile(skillfs.OrgsRoot + "/" + entry.Name() + "/team.json")
		if readErr != nil {
			continue
		}
		var team teamMeta
		if json.Unmarshal(data, &team) != nil {
			continue
		}
		for _, f := range team.Flows {
			if f.ID == flowID {
				return &team
			}
		}
	}
	return nil
}

func selectTeam(fsys embed.FS) *teamMeta {
	entries, err := fs.ReadDir(fsys, skillfs.OrgsRoot)
	if err != nil || len(entries) == 0 {
		fmt.Println("  No preset teams found.")
		return nil
	}

	var teams []teamMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, readErr := fsys.ReadFile(skillfs.OrgsRoot + "/" + entry.Name() + "/team.json")
		if readErr != nil {
			continue
		}
		var team teamMeta
		if json.Unmarshal(data, &team) != nil {
			continue
		}
		teams = append(teams, team)
	}

	if len(teams) == 0 {
		fmt.Println("  No preset teams found.")
		return nil
	}

	fmt.Println("\n  Available teams:")
	fmt.Println()
	for i, t := range teams {
		desc := t.Description
		if t.DescriptionZh != "" {
			desc = t.DescriptionZh
		}
		fmt.Printf("    %2d. %-20s — %s (%d flows)\n", i+1, t.ID, desc, len(t.Flows))
	}
	fmt.Printf("     0. Create new team (via AI)\n")
	fmt.Println()

	if autoYes {
		fmt.Printf("  Auto-selected: %s (--yes mode)\n", teams[0].ID)
		return &teams[0]
	}

	fmt.Print("  Select team [1-0]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "0" {
		fmt.Println("  → AI will help you create a new team via team-flow-v3-create skill")
		return nil
	}

	idx := 0
	fmt.Sscanf(input, "%d", &idx)
	if idx >= 1 && idx <= len(teams) {
		return &teams[idx-1]
	}

	for i := range teams {
		if teams[i].ID == input {
			return &teams[i]
		}
	}

	fmt.Printf("  Using first preset: %s\n", teams[0].ID)
	return &teams[0]
}

func installTeamFlows(fsys embed.FS, teamID, destDir string, overwrite bool) int {
	entries, err := fs.ReadDir(fsys, skillfs.OrgsRoot+"/"+teamID+"/flows")
	if err != nil {
		return 0
	}

	copied := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		dst := filepath.Join(destDir, entry.Name())
		if !overwrite {
			if _, err := os.Stat(dst); err == nil {
				continue
			}
		}
		data, readErr := fsys.ReadFile(skillfs.OrgsRoot + "/" + teamID + "/flows/" + entry.Name())
		if readErr != nil {
			continue
		}
		if writeErr := os.WriteFile(dst, data, 0644); writeErr != nil {
			continue
		}
		copied++
	}
	return copied
}

func installTeamDefinition(fsys embed.FS, teamID, teamDir string, overwrite bool) error {
	data, err := fsys.ReadFile(skillfs.OrgsRoot + "/" + teamID + "/team.json")
	if err != nil {
		return fmt.Errorf("team '%s' definition not found: %w", teamID, err)
	}

	dst := filepath.Join(teamDir, "team.json")
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			fmt.Println("  .team/team.json already exists (use --force to overwrite)")
			return nil
		}
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("write .team/team.json: %w", err)
	}

	fmt.Println("  ✓ .team/team.json created")
	return nil
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

func generateDevMD(projectPath, skillPath, version string) {
	fmt.Println("  Generating IDE bridge files...")

	ides := detectIDEs(projectPath)
	if len(ides) == 0 {
		ides = append(ides, ide.IDEInfo{
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

		content := generateBridgeFileContent(ide.BridgeFmt, skillPath, version)
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

func generateBridgeFileContent(format, skillPath, version string) string {
	if version == "v3" {
		v3Content := `# team-flow v3

> Load ` + "`" + skillPath + "/SKILL.md" + "`" + ` and follow ALL rules defined there.
> This file is a bridge — the actual rules are in the SKILL.md and its references.

## ⛔ MANDATORY: Session Startup Protocol (EVERY SESSION)

Before responding to ANY user request, execute:

` + "```" + `
Step 1: flow project detect    → Lock the project
Step 2: flow proc run          → Get work instructions (role, rules, tools, docs)
Step 3: Adopt principal role   → Read alias, persona, traits from output
` + "```" + `

Only AFTER Step 3, start working on user requests.

## ⛔ CRITICAL: Flow-First Enforcement

**When .team/ directory exists in the project, ALL work MUST go through the flow.**

` + "```" + `
⛔ FORBIDDEN (when flow is active):
- Starting to code/implement/modify files directly
- Skipping flow proc run and working independently
- Ignoring the principal role's dispatch-only rule
- Creating files without flow proc run specifying them as deliverables

✅ REQUIRED:
- Run flow proc run FIRST to get your role and instructions
- Follow the role's rules (especially Dispatch Guard for principal)
- Pass through gates: flow gate pass <node> --condition <type>
- Advance nodes: flow proc run <next-node-id>
` + "```" + `

**If you find yourself about to write code or modify files without having run ` + "`flow proc run`" + ` first — STOP. Run ` + "`flow proc run`" + ` and follow the output.**

## Status Line (MANDATORY — Every Response)

Format: ` + "`[{alias} | {node_name}({node_id}:{flow}) | {ref} | {phase}]`" + `

Data from ` + "`flow proc run`" + ` output only — never hardcode.

## Entry Point

1. Load ` + "`" + skillPath + "/SKILL.md" + "`" + ` for full v3 protocol
2. Run ` + "`flow proc run`" + ` to start the flow engine
3. Follow engine output — it provides everything
`
		switch format {
		case "cursor":
			return "---\n" +
				"description: team-flow v3 — load SKILL.md and follow Session Startup Protocol before any work\n" +
				"globs:\n" +
				"  - \"**/*\"\n" +
				"---\n\n" +
				v3Content
		case "trae", "claude", "openclaw", "gemini":
			return v3Content
		default:
			return ""
		}
	}

	skillRef := "# team-flow\n\n" +
		"Read `.team/version` for active version.\n" +
		"Follow ALL rules in `" + skillPath + "/SKILL.md` as the entry point.\n" +
		"Run `flow proc run` to start the flow engine.\n"

	switch format {
	case "cursor":
		return "---\n" +
			"description: team-flow AI collaboration rules\n" +
			"globs:\n" +
			"  - \"**/*\"\n" +
			"---\n\n" +
			skillRef
	case "trae", "claude", "openclaw":
		return skillRef
	default:
		return ""
	}
}

func installMCPConfig(projectPath string) {
	pyPath := toolchain.FindPythonPath()
	bdPath := bd.FindPath()

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
    "flow-task": {
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

func detectIDEs(projectPath string) []ide.IDEInfo {
	return ide.DetectIDEs(projectPath, skillfs.FS)
}

func isInteractive() bool {
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func installSkill(projectPath string, fsys embed.FS, version string, overwrite bool) {
	fmt.Println("\n  Installing skill...")

	isNonInteractive := !isInteractive()

	if isNonInteractive {
		fmt.Println("  ✓ Non-interactive mode detected")
		fmt.Println("  ✓ Using embedded skill files (no network required)")
		installSkillFromFS(projectPath, fsys, version, overwrite)
		return
	}

	if npxPath, err := exec.LookPath("npx"); err == nil {
		fmt.Printf("  npx found: %s\n", npxPath)
		if autoYes || confirm("  Install skill via npx skills add? (recommended)") {
			ides := detectIDEs(projectPath)
			detectedIDEInfo := ide.IDEInfo{}
			for _, info := range ides {
				if info.Detected {
					detectedIDEInfo = info
					break
				}
			}

			agentFlag := ""
			if detectedIDEInfo.NpxAgent != "" {
				agentFlag = "--agent " + detectedIDEInfo.NpxAgent
			}

			npxCmd := fmt.Sprintf("npx skills add origadmin/team-flow %s", agentFlag)
			fmt.Printf("  Running: %s\n", npxCmd)

			parts := []string{"skills", "add", "origadmin/team-flow"}
			if agentFlag != "" {
				parts = append(parts, strings.Fields(agentFlag)...)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			
			cmd := exec.CommandContext(ctx, npxPath, parts...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			fmt.Println("  (timeout after 2 minutes)")
			
			if err := cmd.Run(); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					fmt.Println("  ⚠ npx skills add timed out after 2 minutes")
				} else {
					fmt.Printf("  ⚠ npx skills add failed: %v\n", err)
				}
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

	switch version {
	case "v3":
		srcDir = skillfs.SkillRoot + "/v3"
	case "v2":
		srcDir = skillfs.SkillRoot + "/v2"
	default:
		srcDir = skillfs.SkillRoot + "/v1"
	}

	sharedDirs := []string{skillfs.SkillRoot + "/v1/config", skillfs.SkillRoot + "/v1/templates", skillfs.SkillRoot + "/v1/lessons", skillfs.SkillRoot + "/v1/scripts"}

	installIDESkills := func(ideEntry ide.IDEInfo) {
		skillDir := filepath.Join(projectPath, ideEntry.Subdir, "skills", "team-flow")

		skillEntry := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillEntry); err != nil || overwrite {
			entryPath := skillfs.SkillRoot + "/v1/SKILL.md"
			if version == "v2" {
				entryPath = skillfs.SkillRoot + "/v2/SKILL.md"
			} else if version == "v3" {
				entryPath = skillfs.SkillRoot + "/v3/SKILL.md"
			}
			entryData, readErr := fsys.ReadFile(entryPath)
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
					fmt.Printf("  ✓ %s shared: %d files from %s\n", ideEntry.Name, c, sharedDir)
				}
			}
		}

		copied, skipped, err := copyFromFS(fsys, srcDir, skillDir, overwrite, excludeDirs)
		if err != nil {
			fmt.Printf("  ⚠ %s skill copy error: %v\n", ideEntry.Name, err)
			return
		}

		fmt.Printf("  ✓ %s (%s): Copied %d files", ideEntry.Name, version, copied)
		if skipped > 0 {
			fmt.Printf(" (%d skipped)", skipped)
		}
		fmt.Println()

		if version == "v3" {
			installFallbackVersion(fsys, skillDir, "v2", ideEntry.Name, overwrite)
		} else if version == "v2" {
			installFallbackVersion(fsys, skillDir, "v1", ideEntry.Name, overwrite)
		}
	}

	registry, regErr := ide.LoadRegistry(fsys)
	if regErr != nil {
		fmt.Printf("  ⚠ Failed to load IDE registry: %v\n", regErr)
		fmt.Println("  Using built-in defaults")
	}

	installed := false
	if registry != nil {
		for _, entry := range registry.IDEs {
			configDir := filepath.Join(projectPath, entry.Subdir)
			shouldInstall := false
			if _, err := os.Stat(configDir); err == nil {
				shouldInstall = true
			} else if version == "v3" && entry.AlwaysInstall {
				shouldInstall = true
			}
			if shouldInstall {
				installIDESkills(ide.IDEInfo{Name: entry.Name, Subdir: entry.Subdir})
				installed = true
			}
		}
	}

	if !installed {
		fmt.Println("  No IDE configuration detected. Installing to .agents/skills/team-flow/")
		skillDir := filepath.Join(projectPath, ".agents", "skills", "team-flow")

		skillEntry := filepath.Join(skillDir, "SKILL.md")
		entryPath := skillfs.SkillRoot + "/v1/SKILL.md"
	if version == "v2" {
		entryPath = skillfs.SkillRoot + "/v2/SKILL.md"
	} else if version == "v3" {
		entryPath = skillfs.SkillRoot + "/v3/SKILL.md"
	}
	entryData, readErr := fsys.ReadFile(entryPath)
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

		if version == "v3" {
			installFallbackVersion(fsys, skillDir, "v2", "local", overwrite)
		} else if version == "v2" {
			installFallbackVersion(fsys, skillDir, "v1", "local", overwrite)
		}
	}
}

func installFallbackVersion(fsys embed.FS, skillDir, fallbackVersion, ideName string, overwrite bool) {
	fallbackSrcDir := skillfs.SkillRoot + "/" + fallbackVersion
	fallbackDstDir := filepath.Join(skillDir, fallbackVersion)

	fallbackSkillEntry := filepath.Join(fallbackDstDir, "SKILL.md")
	if _, err := os.Stat(fallbackSkillEntry); err == nil && !overwrite {
		fmt.Printf("  ✓ %s: %s fallback already exists\n", ideName, fallbackVersion)
		return
	}

	copied, _, err := copyFromFS(fsys, fallbackSrcDir, fallbackDstDir, overwrite, nil)
	if err != nil {
		fmt.Printf("  ⚠ %s: %s fallback copy error: %v\n", ideName, fallbackVersion, err)
		return
	}

	fmt.Printf("  ✓ %s: %s fallback preserved (%d files)\n", ideName, fallbackVersion, copied)
}
