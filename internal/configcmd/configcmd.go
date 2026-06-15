package configcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/version"
	"github.com/spf13/cobra"
)

var (
	configFormat  string
	configKey     string
	configValue   string
	configProject string
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long: `View and manage team-flow configuration.

This command provides unified access to all configuration.
All placeholders and environment variables are resolved before output.

The primary configuration source is .team/project.yaml (YAML).
.team/project.md is only a fallback/backup for compatibility.

Usage:
  flow config show              # Show all configuration (resolved)
  flow config get <key>         # Get specific configuration value (resolved)`,
	RunE: runConfig,
}

func init() {
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(setCmd)
	Cmd.Flags().StringVar(&configFormat, "format", "text", "Output format: text or json")
	Cmd.PersistentFlags().StringVar(&configProject, "project", "", "Project path (default: current directory)")
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all configuration",
	Long:  `Display all team-flow configuration (all values resolved, no placeholders).`,
	Args:  cobra.NoArgs,
	RunE:  runShow,
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get specific configuration value (resolved)",
	Long:  `Get a specific configuration value with all placeholders resolved.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value in project.yaml",
	Long: `Set a configuration value and write it to project.yaml.

Supported keys:
  active_flow    — Set the active flow name (must match a registered flow)

Examples:
  flow config set active_flow dev-flow
  flow config set active_flow feature-flow`,
	Args: cobra.ExactArgs(2),
	RunE: runSet,
}

func runConfig(cmd *cobra.Command, args []string) error {
	return runShow(cmd, args)
}

func runShow(cmd *cobra.Command, args []string) error {
	projectRoot := getProjectRoot()
	cfg, err := loadConfig(projectRoot)
	if err != nil {
		cfg = &Config{}
	}

	// Add flow CLI info
	cfg.FLOW_CLI_Version = version.Version
	cfg.FLOW_CLI_Path = getFlowExePath()
	cfg.FLOW_CLI_BuildTime = version.BuildTime
	cfg.FLOW_CLI_GitCommit = version.GitCommit

	if configFormat == "json" {
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(data))
	} else {
		printConfig(cfg, projectRoot)
	}
	return nil
}

func runGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	projectRoot := getProjectRoot()
	cfg, err := loadConfig(projectRoot)
	if err != nil {
		cfg = &Config{}
	}

	value := cfg.GetResolved(key, projectRoot)

	if configFormat == "json" {
		fmt.Printf("%q\n", value)
	} else {
		fmt.Println(value)
	}
	return nil
}

type Config struct {
	ProjectName         string `json:"project_name,omitempty"`
	TeamVersion         string `json:"team_version,omitempty"`
	ActiveFlow          string `json:"active_flow,omitempty"`
	FlowPath            string `json:"flow_path,omitempty"`
	FlowPathResolved    string `json:"flow_path_resolved,omitempty"`
	BackupPath          string `json:"backup_path,omitempty"`
	DocsInternal        string `json:"docs_internal,omitempty"`
	DocsExternal        string `json:"docs_external,omitempty"`
	UpdateDisabled      bool   `json:"update_check_disabled,omitempty"`
	UpdateInterval      string `json:"update_check_interval,omitempty"`
	ProjectsPath        string `json:"projects_path,omitempty"`
	FLOW_CLI_Version    string `json:"flow_cli_version,omitempty"`
	FLOW_CLI_Path       string `json:"flow_cli_path,omitempty"`
	FLOW_CLI_BuildTime  string `json:"flow_cli_build_time,omitempty"`
	FLOW_CLI_GitCommit  string `json:"flow_cli_git_commit,omitempty"`
}

// GetResolved returns the resolved value for a key (all placeholders resolved)
func (c *Config) GetResolved(key string, projectRoot string) string {
	switch key {
	case "project_name":
		return c.ProjectName
	case "team_version":
		return c.TeamVersion
	case "active_flow":
		return c.ActiveFlow
	case "flow_path":
		// Return resolved flow path (relative → absolute)
		return c.resolveFlowPath(projectRoot)
	case "flow_path_raw":
		// Return raw value from config
		return c.FlowPath
	case "backup_path":
		return c.resolvePath(c.BackupPath, projectRoot)
	case "docs_internal":
		return c.resolvePath(c.DocsInternal, projectRoot)
	case "docs_external":
		return c.resolvePath(c.DocsExternal, projectRoot)
	case "projects_path":
		return c.resolvePath(c.ProjectsPath, projectRoot)
	case "update_disabled":
		return fmt.Sprintf("%v", c.UpdateDisabled)
	case "update_interval":
		return c.UpdateInterval
	case "project_root":
		return projectRoot
	case "workspace":
		return config.FindWorkspaceRoot(projectRoot)
	case "flow_executable":
		// Return the actual executable path that can be used
		return c.resolveFlowPath(projectRoot)
	default:
		return ""
	}
}

// resolveFlowPath resolves flow_path from relative to absolute
func (c *Config) resolveFlowPath(projectRoot string) string {
	flowPath := c.FlowPath
	if flowPath == "" {
		// Default: scripts/flow.exe relative to project root
		flowExePath := filepath.Join(projectRoot, "scripts", "flow.exe")
		if _, err := os.Stat(flowExePath); err == nil {
			return flowExePath
		}
		// Fallback 1: check PATH
		if exePath, err := exec.LookPath("flow"); err == nil {
			return exePath
		}
		// Fallback 2: return the constructed default
		return flowExePath
	}

	// If already absolute, return as-is
	if filepath.IsAbs(flowPath) {
		return flowPath
	}

	// If relative, make it absolute relative to project root
	absPath := filepath.Join(projectRoot, flowPath)

	// Check if file exists
	if _, err := os.Stat(absPath); err == nil {
		return absPath
	}

	// Try finding in PATH
	if exePath, err := exec.LookPath(flowPath); err == nil {
		return exePath
	}

	// Return the constructed path even if not found
	return absPath
}

// resolvePath resolves a path (handles relative paths)
func (c *Config) resolvePath(path, projectRoot string) string {
	if path == "" {
		return ""
	}

	if filepath.IsAbs(path) {
		return path
	}

	workspace := config.FindWorkspaceRoot(projectRoot)
	if workspace == "" {
		workspace = projectRoot
	}
	return config.ResolvePathWithAnchor(projectRoot, workspace, path)
}

func loadConfig(projectRoot string) (*Config, error) {
	cfg := &Config{}

	// 1. First try: Load from project.yaml (primary source)
	projectCfg, err := config.LoadProjectConfig(projectRoot)
	if err == nil {
		// Successfully loaded from YAML
		cfg.ProjectName = projectCfg.Name
		cfg.ActiveFlow = projectCfg.ActiveFlow
		cfg.DocsInternal = projectCfg.Paths.DocsInternal
		cfg.DocsExternal = projectCfg.Paths.DocsExternal
		cfg.ProjectsPath = projectCfg.Paths.ProjectsPath
		// Load flow-specific config from YAML
		cfg.FlowPath = projectCfg.Flow.Path
		cfg.BackupPath = projectCfg.Paths.BackupPath
		cfg.UpdateDisabled = projectCfg.Flow.UpdateDisabled
		cfg.UpdateInterval = projectCfg.Flow.UpdateInterval
	}

	// 2. Load team version from .team/version
	versionPath := filepath.Join(projectRoot, ".team", "version")
	if data, err := os.ReadFile(versionPath); err == nil {
		cfg.TeamVersion = strings.TrimSpace(string(data))
	}

	// 3. For compatibility, also check project.md for flow-specific fields
	// This ensures backward compatibility if some users still use project.md
	if cfg.FlowPath == "" {
		loadFlowConfigFromMD(projectRoot, cfg)
	}

	return cfg, nil
}

// loadFlowConfigFromMD loads flow-specific config from project.md (fallback)
func loadFlowConfigFromMD(projectRoot string, cfg *Config) {
	mdPath := filepath.Join(projectRoot, ".team", "project.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		return
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "flow_path:") {
			cfg.FlowPath = strings.TrimSpace(strings.TrimPrefix(trimmed, "flow_path:"))
		}
		if strings.HasPrefix(trimmed, "backup_path:") {
			cfg.BackupPath = strings.TrimSpace(strings.TrimPrefix(trimmed, "backup_path:"))
		}
		if strings.Contains(trimmed, "update_check_disabled: true") {
			cfg.UpdateDisabled = true
		}
		if strings.HasPrefix(trimmed, "update_check_interval:") {
			cfg.UpdateInterval = strings.TrimSpace(strings.TrimPrefix(trimmed, "update_check_interval:"))
		}
	}
}

func printConfig(cfg *Config, projectRoot string) {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         flow Configuration              ║")
	fmt.Println("╚══════════════════════════════════════════╝")

	fmt.Println("\n[Project]")
	fmt.Printf("  Name:          %s\n", cfg.ProjectName)
	fmt.Printf("  Root:          %s\n", projectRoot)
	fmt.Printf("  Team Version:  %s\n", cfg.TeamVersion)
	fmt.Printf("  Active Flow:   %s\n", cfg.ActiveFlow)
	if cfg.ProjectsPath != "" {
		fmt.Printf("  Projects Path: %s\n", cfg.resolvePath(cfg.ProjectsPath, projectRoot))
	}

	fmt.Println("\n[Paths] (All resolved to absolute paths)")
	fmt.Printf("  Flow Executable: %s\n", cfg.resolveFlowPath(projectRoot))
	if cfg.DocsInternal != "" {
		fmt.Printf("  Docs Internal:   %s\n", cfg.resolvePath(cfg.DocsInternal, projectRoot))
	}
	if cfg.DocsExternal != "" {
		fmt.Printf("  Docs External:   %s\n", cfg.resolvePath(cfg.DocsExternal, projectRoot))
	}
	if cfg.BackupPath != "" {
		fmt.Printf("  Backup Path:     %s\n", cfg.resolvePath(cfg.BackupPath, projectRoot))
	}

	fmt.Println("\n[Update Settings]")
	fmt.Printf("  Check Disabled: %v\n", cfg.UpdateDisabled)
	if cfg.UpdateInterval != "" {
		fmt.Printf("  Check Interval: %s\n", cfg.UpdateInterval)
	}

	fmt.Println("\n[Flow CLI]")
	fmt.Printf("  Version:    %s\n", cfg.FLOW_CLI_Version)
	fmt.Printf("  Binary:     %s\n", cfg.FLOW_CLI_Path)
	fmt.Printf("  Build:      %s\n", cfg.FLOW_CLI_BuildTime)

	fmt.Println("\n[Usage for AI]")
	fmt.Println("  Get resolved flow executable: flow config get flow_executable")
	fmt.Println("  Get resolved paths:           flow config get backup_path")
	fmt.Println("  Get project root:             flow config get project_root")
	fmt.Println("  Get all (JSON):               flow config show --format json")
}

func getProjectRoot() string {
	if configProject != "" {
		return configProject
	}
	dir, _ := os.Getwd()
	return config.ResolveProjectRoot(dir)
}

func getFlowExePath() string {
	exePath, _ := os.Executable()
	return exePath
}

// allowedConfigKeys lists the keys that can be set via flow config set
var allowedConfigKeys = map[string]string{
	"active_flow":    "ActiveFlow",
	"active_project": "ActiveProject",
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	var configRoot string

	if key == "active_project" {
		cwd, _ := os.Getwd()
		configRoot = config.FindWorkspaceRoot(cwd)
		if configRoot == "" {
			configRoot = cwd
		}
	} else {
		configRoot = getProjectRoot()
		if configRoot == "" {
			return fmt.Errorf("project not found")
		}
	}

	yamlField, ok := allowedConfigKeys[key]
	if !ok {
		var keys []string
		for k := range allowedConfigKeys {
			keys = append(keys, k)
		}
		return fmt.Errorf("unknown config key: %s (allowed: %s)", key, strings.Join(keys, ", "))
	}

	// Load existing config
	cfg, err := config.LoadProjectConfig(configRoot)
	if err != nil {
		cfg = &config.ProjectConfig{}
	}

	// Validate value based on key
	switch key {
	case "active_flow":
		if err := validateFlowName(configRoot, value); err != nil {
			return err
		}
		cfg.ActiveFlow = value
		// (no alias — the canonical key is "active_flow")
	case "active_project":
		cfg.ActiveProject = value
	}

	// Suppress unused var warning
	_ = yamlField

	// Save config
	if err := config.SaveProjectConfig(configRoot, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ %s = %s\n", key, value)
	return nil
}

// validateFlowName checks that the flow name exists in preset or project flows
func validateFlowName(root, name string) error {
	candidates := []string{
		filepath.Join(root, "v3", "flows", name+".json"),
		filepath.Join(root, ".team", "flows", name+".json"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return nil
		}
	}
	return fmt.Errorf("flow not found: %s (must be a registered flow in v3/flows/ or .team/flows/)", name)
}
