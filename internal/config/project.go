package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProjectDecl struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	Type string `yaml:"type,omitempty"`
	Team string `yaml:"team,omitempty"`
}

type ProjectConfig struct {
	Name         string             `yaml:"name"`
	Version      string             `yaml:"version"`
	ActiveFlow   string             `yaml:"active_flow"`
	ActiveProject string             `yaml:"active_project,omitempty"`
	Paths        ProjectPaths       `yaml:"paths"`
	Toolchain    ProjectToolchain   `yaml:"toolchain"`
	Flows        []ProjectFlow      `yaml:"flows"`
	Dependencies []ProjectDependency `yaml:"dependencies,omitempty"`
	Flow         FlowConfig         `yaml:"flow,omitempty"`
	Skills       ProjectSkillConfig `yaml:"skills,omitempty"`
	Projects     []ProjectDecl      `yaml:"projects,omitempty"`
}

type ProjectSkillConfig struct {
	Tags         []string          `yaml:"tags,omitempty"`
	LocalSkills  []string          `yaml:"local,omitempty"`
	Disabled     []string          `yaml:"disabled,omitempty"`
	Overrides    map[string]string `yaml:"overrides,omitempty"`
}

type FlowConfig struct {
	Path           string `yaml:"path,omitempty"`
	BackupPath     string `yaml:"backup_path,omitempty"`
	UpdateDisabled bool   `yaml:"update_disabled,omitempty"`
	UpdateInterval string `yaml:"update_interval,omitempty"`
}

type ProjectPaths struct {
	ProjectsPath string `yaml:"projects_path"`
	DocsInternal string `yaml:"docs_internal"`
	DocsExternal string `yaml:"docs_external"`
	BackupPath   string `yaml:"backup_path"`
}

type ProjectToolchain struct {
	Backend  ProjectTool `yaml:"backend"`
	Frontend ProjectTool `yaml:"frontend"`
}

type ProjectTool struct {
	Language       string `yaml:"language"`
	Pipeline       string `yaml:"pipeline"`
	Framework      string `yaml:"framework,omitempty"`
	UILibrary      string `yaml:"ui_library,omitempty"`
	CSSFramework   string `yaml:"css_framework,omitempty"`
	PackageManager string `yaml:"package_manager,omitempty"`
}

type ProjectFlow struct {
	ID     string `yaml:"id"`
	Source string `yaml:"source"`
}

type ProjectDependency struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

func LoadProjectConfig(root string) (*ProjectConfig, error) {
	yamlPath := filepath.Join(root, ".team", "project.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		mdPath := filepath.Join(root, ".team", "project.md")
		if mdData, mdErr := os.ReadFile(mdPath); mdErr == nil {
			return parseProjectMD(mdData)
		}
		return nil, fmt.Errorf("no project config found: %w", err)
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse project.yaml: %w", err)
	}
	return &cfg, nil
}

func SaveProjectConfig(root string, cfg *ProjectConfig) error {
	yamlPath := filepath.Join(root, ".team", "project.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal project.yaml: %w", err)
	}
	return os.WriteFile(yamlPath, data, 0644)
}

// ResolveProjectRootWithActive resolves the project root directory,
// preferring the workspace-level active_project config over CWD-based resolution.
func ResolveProjectRootWithActive(cwd string) string {
	workspace := FindWorkspaceRoot(cwd)
	if workspace == "" {
		return ResolveProjectRoot(cwd)
	}
	cfg, err := LoadProjectConfig(workspace)
	if err != nil || cfg.ActiveProject == "" {
		return ResolveProjectRoot(cwd)
	}
	// Resolve active_project path: relative to workspace, or absolute
	activePath := cfg.ActiveProject
	if !filepath.IsAbs(activePath) {
		activePath = filepath.Join(workspace, activePath)
	}
	if _, err := os.Stat(activePath); err != nil {
		return ResolveProjectRoot(cwd)
	}
	return activePath
}

func ResolveInternalDocs(root string) string {
	cfg, err := LoadProjectConfig(root)
	if err != nil {
		return filepath.Join(root, ".team", "docs")
	}
	if cfg.Paths.DocsInternal != "" {
		workspace := FindWorkspaceRoot(root)
		if workspace == "" {
			workspace = root
		}
		return resolvePathWithAnchor(root, workspace, cfg.Paths.DocsInternal)
	}
	return filepath.Join(root, ".team", "docs")
}

func ResolveExternalDocs(root string) string {
	cfg, err := LoadProjectConfig(root)
	if err != nil {
		return ""
	}
	if cfg.Paths.DocsExternal != "" {
		workspace := FindWorkspaceRoot(root)
		if workspace == "" {
			workspace = root
		}
		return resolvePathWithAnchor(root, workspace, cfg.Paths.DocsExternal)
	}
	return ""
}

func parseProjectMD(data []byte) (*ProjectConfig, error) {
	cfg := &ProjectConfig{}
	content := string(data)
	lines := splitLines(content)

	for _, line := range lines {
		trimmed := trimSpace(line)
		if contains(trimmed, "projects_path:") {
			cfg.Paths.ProjectsPath = extractValue(trimmed)
		} else if contains(trimmed, "docs_internal:") {
			cfg.Paths.DocsInternal = extractValue(trimmed)
		} else if contains(trimmed, "docs_external:") {
			cfg.Paths.DocsExternal = extractValue(trimmed)
		} else if contains(trimmed, "active_flow:") {
			if v := extractValue(trimmed); v != "" {
				cfg.ActiveFlow = v
			}
		}
	}

	return cfg, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func extractValue(line string) string {
	idx := -1
	for i := 0; i < len(line); i++ {
		if line[i] == ':' {
			idx = i
			break
		}
	}
	if idx < 0 || idx+1 >= len(line) {
		return ""
	}
	val := trimSpace(line[idx+1:])
	if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
		val = val[1 : len(val)-1]
	}
	return val
}
