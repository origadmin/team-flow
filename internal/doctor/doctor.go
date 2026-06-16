package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/origadmin/team-flow/internal/toolchain"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose toolchain issues",
	Long: `Diagnose team-flow toolchain issues and provide fix suggestions.

Checks:
  - Python version and availability
  - pip availability
  - code-review-graph installation and graph status
  - beads (bd CLI) installation and database status
  - MCP configuration
  - team-flow framework files completeness
  - PATH and environment issues`,
	RunE: runDoctor,
}

type Diagnosis struct {
	Name    string
	Status  string
	Detail  string
	Fix     string
	Healthy bool
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║       team-flow Toolchain Doctor         ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("  OS: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	diagnoses := []Diagnosis{}

	diagnoses = append(diagnoses, diagnosePython())
	diagnoses = append(diagnoses, diagnosePip())
	diagnoses = append(diagnoses, diagnoseCodeReviewGraph())
	diagnoses = append(diagnoses, diagnoseBeads())
	diagnoses = append(diagnoses, diagnoseMCP())
	diagnoses = append(diagnoses, diagnoseTeamFramework())

	healthy := 0
	unhealthy := 0

	for _, d := range diagnoses {
		icon := "✓"
		if !d.Healthy {
			icon = "✗"
			unhealthy++
		} else {
			healthy++
		}

		fmt.Printf("  %s %-25s %s\n", icon, d.Name, d.Detail)
		if !d.Healthy && d.Fix != "" {
			fmt.Printf("    → Fix: %s\n", d.Fix)
		}
	}

	fmt.Printf("\n  Result: %d healthy, %d issues\n", healthy, unhealthy)

	if unhealthy > 0 {
		fmt.Println("\n  Quick fix: flow setup")
		fmt.Println("  Or fix individual issues as suggested above.")
	} else {
		fmt.Println("\n  All checks passed! Toolchain is ready.")
	}

	return nil
}

func diagnosePython() Diagnosis {
	d := Diagnosis{Name: "Python 3.10+"}

	pyPath := toolchain.FindPythonPath()
	if pyPath == "" {
		d.Status = "missing"
		d.Detail = "not found"
		d.Fix = "Install Python 3.10+: winget install Python.Python.3.12 (Windows) / brew install python@3.12 (macOS)"
		return d
	}

	out, err := exec.Command(pyPath, "--version").CombinedOutput()
	if err != nil {
		d.Status = "error"
		d.Detail = fmt.Sprintf("found at %s but --version failed", pyPath)
		d.Fix = "Reinstall Python"
		return d
	}

	version := strings.TrimSpace(string(out))
	d.Detail = fmt.Sprintf("%s (%s)", version, pyPath)

	out2, err := exec.Command(pyPath, "-c", "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')").CombinedOutput()
	if err == nil {
		v := strings.TrimSpace(string(out2))
		parts := strings.Split(v, ".")
		major, minor := 0, 0
		if len(parts) >= 2 {
			fmt.Sscanf(parts[0], "%d", &major)
			fmt.Sscanf(parts[1], "%d", &minor)
		}
		if major < 3 || (major == 3 && minor < 10) {
			d.Status = "outdated"
			d.Detail = fmt.Sprintf("%s (need 3.10+)", version)
			d.Fix = "Upgrade Python to 3.10+"
			return d
		}
	}

	d.Healthy = true
	d.Status = "ok"
	return d
}

func diagnosePip() Diagnosis {
	d := Diagnosis{Name: "pip"}

	for _, name := range []string{"pip", "pip3"} {
		if _, err := exec.LookPath(name); err == nil {
			out, _ := exec.Command(name, "--version").CombinedOutput()
			d.Detail = strings.TrimSpace(string(out))
			d.Healthy = true
			d.Status = "ok"
			return d
		}
	}

	pyPath := toolchain.FindPythonPath()
	if pyPath != "" {
		out, _ := exec.Command(pyPath, "-m", "pip", "--version").CombinedOutput()
		if strings.Contains(string(out), "pip") {
			d.Detail = "available via python -m pip"
			d.Healthy = true
			d.Status = "ok"
			return d
		}
	}

	d.Status = "missing"
	d.Detail = "not found"
	d.Fix = "python -m ensurepip --upgrade"
	return d
}

func diagnoseCodeReviewGraph() Diagnosis {
	d := Diagnosis{Name: "code-review-graph"}

	pyPath := toolchain.FindPythonPath()
	if pyPath == "" {
		d.Status = "blocked"
		d.Detail = "Python not available"
		d.Fix = "Install Python first"
		return d
	}

	out, err := exec.Command(pyPath, "-m", "code_review_graph", "status").CombinedOutput()
	if err != nil {
		d.Status = "missing"
		d.Detail = "not installed"
		d.Fix = "pip install code-review-graph"
		return d
	}

	output := string(out)
	if strings.Contains(strings.ToLower(output), "nodes") {
		d.Detail = "installed, graph built"
		d.Healthy = true
		d.Status = "ok"
		return d
	}

	if strings.Contains(output, "Empty graph") || strings.Contains(output, "0 nodes") {
		d.Detail = "installed, graph NOT built"
		d.Fix = "cd <project> && python -m code_review_graph build"
		d.Status = "incomplete"
		return d
	}

	d.Detail = "installed"
	d.Healthy = true
	d.Status = "ok"
	return d
}

func diagnoseBeads() Diagnosis {
	d := Diagnosis{Name: "task manager (flow task)"}
	d.Status = "ok"
	d.Detail = "using local .team/tasks/ directory"
	d.Healthy = true
	return d
}

func diagnoseMCP() Diagnosis {
	d := Diagnosis{Name: "MCP Configuration"}

	projectPath, _ := os.Getwd()
	traeMCP := filepath.Join(projectPath, ".trae", "mcp.json")

	if _, err := os.Stat(traeMCP); err == nil {
		d.Detail = ".trae/mcp.json exists"
		d.Healthy = true
		d.Status = "ok"
		return d
	}

	cursorMCP := filepath.Join(projectPath, ".cursor", "mcp.json")
	if _, err := os.Stat(cursorMCP); err == nil {
		d.Detail = ".cursor/mcp.json exists"
		d.Healthy = true
		d.Status = "ok"
		return d
	}

	d.Status = "missing"
	d.Detail = "no MCP config found"
	d.Fix = "flow setup (Step 7) or manually create .trae/mcp.json"
	return d
}

func diagnoseTeamFramework() Diagnosis {
	d := Diagnosis{Name: "team-flow Framework"}

	projectPath, _ := os.Getwd()

	candidates := []string{
		filepath.Join(projectPath, ".trae", "skills", "team-flow", "SKILL.md"),
		filepath.Join(projectPath, ".cursor", "skills", "team-flow", "SKILL.md"),
		filepath.Join(projectPath, ".claude", "skills", "team-flow", "SKILL.md"),
		filepath.Join(projectPath, ".openclaw", "skills", "team-flow", "SKILL.md"),
		filepath.Join(projectPath, ".team-flow", "skill", "SKILL.md"),
		filepath.Join(projectPath, ".team", "SKILL.md"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			teamDir := filepath.Dir(c)
			v1Skill := filepath.Join(teamDir, "SKILL.md")
			v2Skill := filepath.Join(teamDir, "v2", "SKILL.md")

			hasV1 := fileExists(v1Skill)
			hasV2 := fileExists(v2Skill)

			if hasV1 && hasV2 {
				d.Detail = fmt.Sprintf("v1+v2 at %s", teamDir)
			} else if hasV2 {
				d.Detail = fmt.Sprintf("v2 at %s", teamDir)
			} else if hasV1 {
				d.Detail = fmt.Sprintf("v1 at %s", teamDir)
			}

			d.Healthy = true
			d.Status = "ok"
			return d
		}
	}

	d.Status = "missing"
	d.Detail = "SKILL.md not found"
	d.Fix = "flow init"
	return d
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}


