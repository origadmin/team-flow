package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	cfg, cfgErr := LoadProjectConfig(root)
	if cfgErr == nil {
		vars.DOCS_INTERNAL = cfg.Paths.DocsInternal
		vars.DOCS_EXTERNAL = cfg.Paths.DocsExternal
		vars.DEFAULT_FLOW = cfg.DefaultFlow
	} else {
		vars.DOCS_INTERNAL = resolveDocsPathFromMD(root)
		vars.DOCS_EXTERNAL = resolveDocsExternalPathFromMD(root)
		vars.DEFAULT_FLOW, _ = resolveDefaultFlowFromMD(root)
	}

	if vars.DOCS_INTERNAL != "" && !filepath.IsAbs(vars.DOCS_INTERNAL) {
		vars.DOCS_INTERNAL = filepath.Join(root, vars.DOCS_INTERNAL)
	}

	if vars.DOCS_EXTERNAL != "" && !filepath.IsAbs(vars.DOCS_EXTERNAL) {
		vars.DOCS_EXTERNAL = filepath.Join(root, vars.DOCS_EXTERNAL)
	}

	vars.SKILL_PATH = findSkillPath(root)

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

func resolveDocsPathFromMD(root string) string {
	projectMD := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMD)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "docs_internal:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), "\"' ")
			}
		}
	}
	return ""
}

func resolveDocsExternalPathFromMD(root string) string {
	projectMD := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMD)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "docs_external:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), "\"' ")
			}
		}
	}
	return ""
}

func resolveDefaultFlowFromMD(root string) (string, error) {
	projectMD := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMD)
	if err != nil {
		return "", fmt.Errorf("no default flow configured")
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "default_flow:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				val := strings.Trim(strings.TrimSpace(parts[1]), "\"' ")
				if val != "" {
					return val, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no default flow configured")
}
