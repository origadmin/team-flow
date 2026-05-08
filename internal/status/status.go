package status

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/toolchain"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status",
	Long: `Show current project team-flow version and status.

Displays:
  - v1 or v2 mode detection
  - Framework file completeness
  - beads status (if v2)
  - code-review-graph status (if v2)
  - Missing files/directories`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Println("=== Project Status ===")
	fmt.Printf("  Path: %s\n\n", projectPath)

	teamPath := findTeamDir(projectPath)

	hasTeamDir := teamPath != ""

	fmt.Println("--- Directories ---")
	fmt.Printf("  team-flow/  %s\n", func() string {
		if hasTeamDir {
			return "✓ " + teamPath
		}
		return "✗ not found"
	}())

	if !hasTeamDir {
		fmt.Println("\n  ⚠ No team-flow configuration found. Run 'flow init' first.")
		return nil
	}

	checkPath := teamPath

	fmt.Println("\n--- v1 Core Files ---")
	v1Files := []struct {
		name string
		path string
	}{
		{"SKILL.md", filepath.Join(checkPath, "SKILL.md")},
		{"BOUNDARY.md", filepath.Join(checkPath, "BOUNDARY.md")},
		{"task-pool.md", filepath.Join(checkPath, "task-pool.md")},
		{"config/team-config.json", filepath.Join(checkPath, "config", "team-config.json")},
		{"prompts/triage.md", filepath.Join(checkPath, "prompts", "triage.md")},
		{"workflows/shared.md", filepath.Join(checkPath, "workflows", "shared.md")},
		{"workflows/framework-workflow.md", filepath.Join(checkPath, "workflows", "framework-workflow.md")},
		{"templates/task-pool-template.md", filepath.Join(checkPath, "templates", "task-pool-template.md")},
		{"templates/feature-test-template.md", filepath.Join(checkPath, "templates", "feature-test-template.md")},
		{"templates/bug-test-template.md", filepath.Join(checkPath, "templates", "bug-test-template.md")},
		{"scripts/merge-task-pool.ps1", filepath.Join(checkPath, "scripts", "merge-task-pool.ps1")},
	}
	v1Complete := true
	for _, f := range v1Files {
		exists := fileExists(f.path)
		if !exists {
			v1Complete = false
		}
		fmt.Printf("  %-30s %s\n", f.name, boolMark(exists))
	}

	fmt.Println("\n--- v2 Core Files ---")
	v2Files := []struct {
		name string
		path string
	}{
		{"v2/SKILL.md", filepath.Join(checkPath, "v2", "SKILL.md")},
		{"v2/BOUNDARY.md", filepath.Join(checkPath, "v2", "BOUNDARY.md")},
		{"v2/prompts/triage.md", filepath.Join(checkPath, "v2", "prompts", "triage.md")},
		{"v2/workflows/shared.md", filepath.Join(checkPath, "v2", "workflows", "shared.md")},
		{"v2/scripts/migrate-tasks.ps1", filepath.Join(checkPath, "v2", "scripts", "migrate-tasks.ps1")},
		{"beads/SKILL.md", filepath.Join(checkPath, "beads", "SKILL.md")},
	}
	v2Complete := true
	for _, f := range v2Files {
		exists := fileExists(f.path)
		if !exists {
			v2Complete = false
		}
		fmt.Printf("  %-30s %s\n", f.name, boolMark(exists))
	}

	fmt.Println("\n--- Version Detection ---")
	v1Skill := filepath.Join(checkPath, "SKILL.md")
	v2Skill := filepath.Join(checkPath, "v2", "SKILL.md")

	if fileExists(v2Skill) {
		fmt.Println("  Detected: v2 (beads-native + code-review-graph)")
	} else if fileExists(v1Skill) {
		fmt.Println("  Detected: v1 (task-pool)")
	} else {
		fmt.Println("  Detected: unknown")
	}

	if v1Complete {
		fmt.Println("  v1 Complete: ✓")
	} else {
		fmt.Println("  v1 Complete: ✗ (missing files)")
	}
	if v2Complete {
		fmt.Println("  v2 Complete: ✓")
	} else {
		fmt.Println("  v2 Complete: ✗ (missing files)")
	}

	fmt.Println("\n--- External Tools ---")
	bdPath := toolchain.FindBdPath()
	if bdPath != "" {
		fmt.Printf("  bd (beads):  ✓ %s\n", bdPath)
	} else {
		fmt.Println("  bd (beads):  ✗ not installed")
	}

	pyPath := toolchain.FindPythonPath()
	if pyPath != "" {
		out, err := exec.Command(pyPath, "-m", "code_review_graph", "status").CombinedOutput()
		if err == nil && strings.Contains(string(out), "nodes") {
			fmt.Println("  code-review-graph: ✓ installed")
		} else {
			fmt.Println("  code-review-graph: ✗ not installed (pip install code-review-graph)")
		}
	} else {
		fmt.Println("  code-review-graph: ✗ python not found")
	}

	beadsDir := filepath.Join(projectPath, ".beads")
	if _, err := os.Stat(beadsDir); err == nil {
		fmt.Printf("  .beads/:     ✓ initialized\n")
	} else {
		fmt.Println("  .beads/:     ✗ not initialized (run 'bd init')")
	}

	graphDB := filepath.Join(projectPath, ".code-review-graph")
	if _, err := os.Stat(graphDB); err == nil {
		fmt.Println("  .code-review-graph/: ✓ built")
	} else {
		fmt.Println("  .code-review-graph/: ✗ not built (run 'flow graph build')")
	}

	return nil
}

func boolMark(b bool) string {
	if b {
		return "✓"
	}
	return "✗"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func findTeamDir(projectPath string) string {
	candidates := []string{
		filepath.Join(projectPath, ".trae", "skills", "team-flow"),
		filepath.Join(projectPath, ".cursor", "skills", "team-flow"),
		filepath.Join(projectPath, ".claude", "skills", "team-flow"),
		filepath.Join(projectPath, ".openclaw", "skills", "team-flow"),
		filepath.Join(projectPath, ".team-flow", "skill"),
		filepath.Join(projectPath, ".team"),
	}

	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if fileExists(filepath.Join(abs, "SKILL.md")) {
			return abs
		}
	}
	return ""
}
