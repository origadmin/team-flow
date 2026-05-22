package ver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "ver",
	Short: "Show project version and status",
	Long: `Show current project version and skill status.
Auto-detects if project is on a lower version and suggests upgrade.

Usage:
  flow ver          Show current version and status`,
	Args: cobra.NoArgs,
	RunE: runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	versionFile := filepath.Join(projectPath, ".team", "version")
	currentVersion := "unknown"
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion = strings.TrimSpace(string(data))
	}

	return showStatus(projectPath, currentVersion)
}

func showStatus(projectPath, currentVersion string) error {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        team-flow Version Status          ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	if currentVersion == "unknown" {
		fmt.Println("  ⛔ .team/version not found")
		fmt.Println("     Run: flow init")
		return nil
	}

	fmt.Printf("  Active version: %s\n\n", currentVersion)

	switch currentVersion {
	case "v1":
		fmt.Println("  Task management: task-pool.md")
		tpPath := filepath.Join(projectPath, ".team", "task-pool.md")
		if fileExists(tpPath) {
			fmt.Println("  ✓ task-pool.md exists")
		} else {
			fmt.Println("  ⚠ task-pool.md not found")
		}
		fmt.Println()
		fmt.Println("  ⬆ Upgrade available: v1 → v2")
		fmt.Println("    Run: flow migrate")
	case "v2":
		fmt.Println("  Task management: beads (bd CLI)")
		beadsDir := filepath.Join(projectPath, ".beads")
		if dirExists(beadsDir) {
			fmt.Println("  ✓ .beads/ exists")
		} else {
			fmt.Println("  ⚠ .beads/ not found (run: bd init)")
		}
		fmt.Println()
		fmt.Println("  ⬆ Upgrade available: v2 → v3")
		fmt.Println("    Run: flow migrate v3")
	case "v3":
		fmt.Println("  Execution engine: flow-engine (v3)")
		flowsDir := filepath.Join(projectPath, "v3", "flows")
		if dirExists(flowsDir) {
			fmt.Println("  ✓ v3/flows/ exists")
		} else {
			altFlowsDir := filepath.Join(projectPath, ".team", "flows")
			if dirExists(altFlowsDir) {
				fmt.Println("  ✓ .team/flows/ exists")
			} else {
				fmt.Println("  ⚠ No flows directory found")
			}
		}
		beadsDir := filepath.Join(projectPath, ".beads")
		if dirExists(beadsDir) {
			fmt.Println("  ✓ .beads/ exists (beads is version-agnostic)")
		}
		fmt.Println()
		fmt.Println("  Commands:")
		fmt.Println("    flow proc run       Start/resume flow execution")
		fmt.Println("    flow proc list      List available flows")
	default:
		fmt.Printf("  ⚠ Unknown version: %s\n", currentVersion)
		fmt.Println("    Run: flow init --v2")
	}

	skillPath := findSkillPath(projectPath)
	if skillPath != "" {
		fmt.Printf("\n  Skill location: %s\n", skillPath)
	} else {
		fmt.Println("\n  ⚠ Skill not installed in any detected path")
		fmt.Println("     Run: flow init")
	}

	fmt.Println()
	fmt.Println("  Commands:")
	fmt.Println("    flow ver        Show this status")
	fmt.Println("    flow migrate    Upgrade version (v1→v2, v2→v3)")

	return nil
}

func findSkillPath(projectPath string) string {
	home, _ := os.UserHomeDir()

	candidates := []string{
		filepath.Join(home, ".agents", "skills", "team-flow"),
		filepath.Join(home, ".trae-cn", "skills", "team-flow"),
		filepath.Join(home, ".claude", "skills", "team-flow"),
		filepath.Join(home, ".cursor", "skills", "team-flow"),
		filepath.Join(projectPath, ".agents", "skills", "team-flow"),
		filepath.Join(projectPath, ".trae", "skills", "team-flow"),
		filepath.Join(projectPath, ".claude", "skills", "team-flow"),
	}

	for _, c := range candidates {
		skillMd := filepath.Join(c, "SKILL.md")
		if fileExists(skillMd) {
			return c
		}
	}
	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}


