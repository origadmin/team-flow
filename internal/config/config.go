package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/proc"
	"github.com/spf13/cobra"
)

var pathsJSON bool

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `Manage team-flow configuration: inspect paths, variables, and settings.`,
}

var pathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show all v3 path variables",
	Long: `Show all resolved v3 path variables for the current project.

Reads .team/project.md and .team/version to resolve paths.
Use --json for machine-readable output.`,
	RunE: runPaths,
}

func init() {
	Cmd.AddCommand(pathsCmd)
	pathsCmd.Flags().BoolVar(&pathsJSON, "json", false, "Output in JSON format")
}

type PathVars struct {
	Version       string `json:"version"`
	Workspace     string `json:"workspace"`
	TEAM_PATH     string `json:"TEAM_PATH"`
	DOCS_INTERNAL string `json:"DOCS_INTERNAL"`
	DOCS_EXTERNAL string `json:"DOCS_EXTERNAL"`
	BEADS_DB      string `json:"BEADS_DB"`
	SKILL_PATH    string `json:"SKILL_PATH"`
	FLOW_DIR      string `json:"FLOW_DIR"`
	DEFAULT_FLOW  string `json:"DEFAULT_FLOW"`
}

func resolvePaths(root string) PathVars {
	vars := PathVars{}

	versionFile := filepath.Join(root, ".team", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		vars.Version = strings.TrimSpace(string(data))
	}

	vars.Workspace = root
	vars.TEAM_PATH = filepath.Join(root, ".team")
	vars.BEADS_DB = filepath.Join(root, ".beads")
	vars.FLOW_DIR = filepath.Join(root, ".team", "flows")

	vars.DOCS_INTERNAL = proc.ResolveDocsPath(root)
	if vars.DOCS_INTERNAL != "" && !filepath.IsAbs(vars.DOCS_INTERNAL) {
		vars.DOCS_INTERNAL = filepath.Join(root, vars.DOCS_INTERNAL)
	}

	vars.DOCS_EXTERNAL = proc.ResolveDocsExternalPath(root)
	if vars.DOCS_EXTERNAL != "" && !filepath.IsAbs(vars.DOCS_EXTERNAL) {
		vars.DOCS_EXTERNAL = filepath.Join(root, vars.DOCS_EXTERNAL)
	}

	vars.SKILL_PATH = findSkillPath(root)

	defaultFlow, _ := proc.ResolveDefaultFlowName(root)
	vars.DEFAULT_FLOW = defaultFlow

	return vars
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
		if _, err := os.Stat(skillMd); err == nil {
			return c
		}
	}
	return ""
}

func runPaths(cmd *cobra.Command, args []string) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	vars := resolvePaths(root)

	if pathsJSON {
		data, err := json.MarshalIndent(vars, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal paths: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	rows := []struct {
		name  string
		value string
	}{
		{"version", vars.Version},
		{"workspace", vars.Workspace},
		{"TEAM_PATH", vars.TEAM_PATH},
		{"DOCS_INTERNAL", vars.DOCS_INTERNAL},
		{"DOCS_EXTERNAL", vars.DOCS_EXTERNAL},
		{"BEADS_DB", vars.BEADS_DB},
		{"SKILL_PATH", vars.SKILL_PATH},
		{"FLOW_DIR", vars.FLOW_DIR},
		{"DEFAULT_FLOW", vars.DEFAULT_FLOW},
	}

	maxName := 0
	for _, r := range rows {
		if len(r.name) > maxName {
			maxName = len(r.name)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%-*s  VALUE\n", maxName, "VARIABLE")
	fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %s\n", maxName, strings.Repeat("-", maxName), strings.Repeat("-", 40))
	for _, r := range rows {
		fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %s\n", maxName, r.name, r.value)
	}

	return nil
}
