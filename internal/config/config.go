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
	Project       string `json:"project"`
	TEAM_PATH     string `json:"TEAM_PATH"`
	DOCS_INTERNAL string `json:"DOCS_INTERNAL"`
	DOCS_EXTERNAL string `json:"DOCS_EXTERNAL"`
	BACKUP_PATH   string `json:"BACKUP_PATH"`
	BEADS_DB      string `json:"BEADS_DB"`
	SKILL_PATH    string `json:"SKILL_PATH"`
	FLOW_DIR      string `json:"FLOW_DIR"`
	ACTIVE_FLOW   string `json:"ACTIVE_FLOW"`
}

// FindWorkspaceRoot 查找 workspace 根目录
// workspace 根是包含 projects_path 配置指向的目录的根
func FindWorkspaceRoot(startDir string) string {
	searchDir := startDir
	visited := make(map[string]bool)
	
	for {
		absDir, err := filepath.Abs(searchDir)
		if err != nil {
			break
		}
		
		if visited[absDir] {
			break
		}
		visited[absDir] = true
		
		// 检查是否包含 .team 目录
		teamDir := filepath.Join(searchDir, ".team")
		if _, err := os.Stat(teamDir); err == nil {
			// 尝试获取 projects_path 配置
			cfg, err := LoadProjectConfig(searchDir)
			if err == nil && cfg.Paths.ProjectsPath != "" {
				projectsPath := cfg.Paths.ProjectsPath
				if filepath.IsAbs(projectsPath) {
					// 绝对路径，直接使用
					if _, err := os.Stat(projectsPath); err == nil {
						return searchDir
					}
				} else {
					// 相对路径
					absProjects := filepath.Join(searchDir, projectsPath)
					if _, err := os.Stat(absProjects); err == nil {
						return searchDir
					}
				}
			}
			// 如果没有配置，使用默认的 projects/
			defaultProjects := filepath.Join(searchDir, "projects")
			if _, err := os.Stat(defaultProjects); err == nil {
				return searchDir
			}
		}
		
		parentDir := filepath.Dir(searchDir)
		if parentDir == searchDir {
			break
		}
		searchDir = parentDir
	}
	
	return ""
}

// IsWorkspaceRoot 判断给定路径是否是 workspace 根目录
// - 有 .team/project.yaml
// - 有 projects 字段 或 存在 projects/ 目录
// - 不应该有 team.json 或 flows/ 目录（但检测时不强制）
func IsWorkspaceRoot(path string) bool {
	if path == "" {
		return false
	}

	// 检查是否有 .team/project.yaml
	cfg, err := LoadProjectConfig(path)
	if err != nil {
		return false
	}

	// 如果有 projects 字段，直接判断为 workspace
	if len(cfg.Projects) > 0 {
		return true
	}

	// 检查是否有 projects/ 目录
	projectsDir := filepath.Join(path, "projects")
	if _, err := os.Stat(projectsDir); err == nil {
		return true
	}

	return false
}

// IsProjectRoot 判断给定路径是否是 project 根目录
// - 有 .team/project.yaml
// - 不是 workspace 根
func IsProjectRoot(path string) bool {
	if path == "" {
		return false
	}
	// 首先检查是否有 .team/project.yaml
	_, err := LoadProjectConfig(path)
	if err != nil {
		return false
	}
	// 是 workspace 根就不是 project 根
	if IsWorkspaceRoot(path) {
		return false
	}
	return true
}

// 保留旧函数兼容
func IsMonorepoWorkspace(root string) bool {
	return IsWorkspaceRoot(root)
}

// GetProjectsDir 获取 projects 目录路径
func GetProjectsDir(workspaceRoot string) string {
	cfg, err := LoadProjectConfig(workspaceRoot)
	if err != nil || cfg.Paths.ProjectsPath == "" {
		return filepath.Join(workspaceRoot, "projects")
	}
	
	projectsPath := cfg.Paths.ProjectsPath
	if filepath.IsAbs(projectsPath) {
		return projectsPath
	}
	return filepath.Join(workspaceRoot, projectsPath)
}

type DiscoveredProject struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	RelPath string `json:"rel_path"`
	Type    string `json:"type"`
	Source  string `json:"source"`
	HasTeam bool   `json:"has_team"`
}

// ResolveProjectRoot 从 cwd 向上查找项目根目录
// 优先找 .team 目录（项目有团队配置），其次找项目 marker（go.mod 等）
func ResolveProjectRoot(cwd string) string {
	dir := cwd

	for {
		// 检查 .team 目录
		if _, err := os.Stat(filepath.Join(dir, ".team")); err == nil {
			return dir
		}

		// 检查项目 marker
		projectMarkers := []string{"go.mod", "package.json", "Cargo.toml", "pyproject.toml", "pom.xml", "build.gradle"}
		for _, m := range projectMarkers {
			if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
				return dir
			}
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return cwd
}

func DiscoverProjects(workspaceRoot string) []DiscoveredProject {
	seen := make(map[string]bool)
	var projects []DiscoveredProject

	cfg, cfgErr := LoadProjectConfig(workspaceRoot)
	if cfgErr == nil && len(cfg.Projects) > 0 {
		for _, p := range cfg.Projects {
			absPath := p.Path
			if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(workspaceRoot, absPath)
			}
			absPath = filepath.Clean(absPath)
			if seen[absPath] {
				continue
			}
			seen[absPath] = true
			relPath := p.Path
			hasTeam := false
			if _, err := os.Stat(filepath.Join(absPath, ".team")); err == nil {
				hasTeam = true
			}
			projects = append(projects, DiscoveredProject{
				Name:    p.Name,
				Path:    absPath,
				RelPath: relPath,
				Type:    p.Type,
				Source:  "config",
				HasTeam: hasTeam,
			})
		}
	}

	goWorkPath := filepath.Join(workspaceRoot, "go.work")
	if data, err := os.ReadFile(goWorkPath); err == nil {
		inUseBlock := false
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "go ") || strings.HasPrefix(trimmed, "//") {
				continue
			}
			if strings.HasPrefix(trimmed, "use") {
				inUseBlock = true
				trimmed = strings.TrimPrefix(trimmed, "use")
				trimmed = strings.TrimSpace(trimmed)
				trimmed = strings.TrimPrefix(trimmed, "(")
				trimmed = strings.TrimSpace(trimmed)
			}
			if trimmed == ")" {
				inUseBlock = false
				continue
			}
			if !inUseBlock && !strings.HasPrefix(trimmed, "./") {
				continue
			}
			usePath := strings.TrimPrefix(trimmed, "./")
			usePath = strings.TrimSuffix(usePath, "/")
			if usePath == "" || usePath == "." {
				continue
			}
			if !strings.HasPrefix(usePath, "projects/") && strings.Contains(usePath, "/") {
				continue
			}
			absPath := filepath.Join(workspaceRoot, usePath)
			absPath = filepath.Clean(absPath)
			if seen[absPath] {
				continue
			}
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				continue
			}
			seen[absPath] = true
			hasTeam := false
			if _, err := os.Stat(filepath.Join(absPath, ".team")); err == nil {
				hasTeam = true
			}
			projType := detectProjectType(absPath)
			projects = append(projects, DiscoveredProject{
				Name:    filepath.Base(absPath),
				Path:    absPath,
				RelPath: usePath,
				Type:    projType,
				Source:  "go.work",
				HasTeam: hasTeam,
			})
		}
	}

	projectsDir := GetProjectsDir(workspaceRoot)
	if entries, err := os.ReadDir(projectsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			absPath := filepath.Join(projectsDir, entry.Name())
			absPath = filepath.Clean(absPath)
			if seen[absPath] {
				continue
			}
			seen[absPath] = true
			hasTeam := false
			if _, err := os.Stat(filepath.Join(absPath, ".team")); err == nil {
				hasTeam = true
			}
			hasMarker := false
			markers := []string{"go.mod", "package.json", "Cargo.toml", "pyproject.toml"}
			for _, m := range markers {
				if _, err := os.Stat(filepath.Join(absPath, m)); err == nil {
					hasMarker = true
					break
				}
			}
			if !hasMarker && !hasTeam {
				continue
			}
			relPath, _ := filepath.Rel(workspaceRoot, absPath)
			projType := detectProjectType(absPath)
			projects = append(projects, DiscoveredProject{
				Name:    entry.Name(),
				Path:    absPath,
				RelPath: relPath,
				Type:    projType,
				Source:  "scan",
				HasTeam: hasTeam,
			})
		}
	}

	return projects
}

func detectProjectType(dir string) string {
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		return "node"
	}
	if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		return "rust"
	}
	if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		return "python"
	}
	return ""
}

// ListProjects 列出 workspace 下所有可用的 projects
func ListProjects(workspaceRoot string) []string {
	var projects []string
	
	projectsDir := GetProjectsDir(workspaceRoot)
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return projects
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		teamDir := filepath.Join(projectsDir, entry.Name(), ".team")
		if _, err := os.Stat(teamDir); err != nil {
			continue
		}
		
		// 验证是否是有效的 project
		versionFile := filepath.Join(teamDir, "version")
		flowsDir := filepath.Join(teamDir, "flows")
		projectMd := filepath.Join(teamDir, "project.md")
		
		if _, errV := os.Stat(versionFile); errV == nil {
			projects = append(projects, entry.Name())
			continue
		}
		if _, errF := os.Stat(flowsDir); errF == nil {
			projects = append(projects, entry.Name())
			continue
		}
		if _, errP := os.Stat(projectMd); errP == nil {
			projects = append(projects, entry.Name())
		}
	}
	
	return projects
}

// ResolvePaths 解析所有路径变量，区分 workspace 和 project
func ResolvePaths(workspaceDir string, projectRoot string) PathVars {
	vars := PathVars{}

	configRoot := projectRoot

	versionFile := filepath.Join(configRoot, ".team", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		vars.Version = strings.TrimSpace(string(data))
	}

	vars.Workspace = workspaceDir
	vars.Project = projectRoot
	vars.TEAM_PATH = filepath.Join(configRoot, ".team")
	vars.BEADS_DB = filepath.Join(projectRoot, ".beads")
	vars.FLOW_DIR = filepath.Join(configRoot, ".team", "flows")

	cfg, cfgErr := LoadProjectConfig(configRoot)
	if cfgErr == nil {
		vars.DOCS_INTERNAL = resolvePathWithAnchor(configRoot, workspaceDir, cfg.Paths.DocsInternal)
		vars.DOCS_EXTERNAL = resolvePathWithAnchor(configRoot, workspaceDir, cfg.Paths.DocsExternal)
		vars.BACKUP_PATH = resolvePathWithAnchor(configRoot, workspaceDir, cfg.Paths.BackupPath)
		vars.ACTIVE_FLOW = cfg.ActiveFlow
	} else {
		vars.DOCS_INTERNAL = resolvePathWithAnchor(configRoot, workspaceDir, resolveDocsPathFromMD(configRoot))
		vars.DOCS_EXTERNAL = resolvePathWithAnchor(configRoot, workspaceDir, resolveDocsExternalPathFromMD(configRoot))
		vars.ACTIVE_FLOW, _ = resolveActiveFlowFromMD(configRoot)
	}

	vars.SKILL_PATH = findSkillPath(projectRoot)

	return vars
}

func resolvePaths(workspaceDir string) PathVars {
	vars := PathVars{}

	projectRoot := ResolveProjectRoot(workspaceDir)
	configRoot := projectRoot

	versionFile := filepath.Join(configRoot, ".team", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		vars.Version = strings.TrimSpace(string(data))
	}

	vars.Workspace = workspaceDir
	vars.Project = projectRoot
	vars.TEAM_PATH = filepath.Join(configRoot, ".team")
	vars.BEADS_DB = filepath.Join(projectRoot, ".beads")
	vars.FLOW_DIR = filepath.Join(configRoot, ".team", "flows")

	cfg, cfgErr := LoadProjectConfig(configRoot)
	if cfgErr == nil {
		vars.DOCS_INTERNAL = resolvePathWithAnchor(configRoot, workspaceDir, cfg.Paths.DocsInternal)
		vars.DOCS_EXTERNAL = resolvePathWithAnchor(configRoot, workspaceDir, cfg.Paths.DocsExternal)
		vars.BACKUP_PATH = resolvePathWithAnchor(configRoot, workspaceDir, cfg.Paths.BackupPath)
		vars.ACTIVE_FLOW = cfg.ActiveFlow
	} else {
		vars.DOCS_INTERNAL = resolvePathWithAnchor(configRoot, workspaceDir, resolveDocsPathFromMD(configRoot))
		vars.DOCS_EXTERNAL = resolvePathWithAnchor(configRoot, workspaceDir, resolveDocsExternalPathFromMD(configRoot))
		vars.ACTIVE_FLOW, _ = resolveActiveFlowFromMD(configRoot)
	}

	vars.SKILL_PATH = findSkillPath(projectRoot)

	return vars
}

func ResolvePathWithAnchor(projectRoot, workspaceDir, path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	if strings.HasPrefix(path, "[@]/") || strings.HasPrefix(path, "[@]\\") {
		relPath := path[3:]
		if relPath == "" {
			return workspaceDir
		}
		return filepath.Join(workspaceDir, relPath)
	}
	if strings.HasPrefix(path, "[#]/") || strings.HasPrefix(path, "[#]\\") {
		relPath := path[3:]
		if relPath == "" {
			return projectRoot
		}
		return filepath.Join(projectRoot, relPath)
	}
	if strings.HasPrefix(path, "_/") || strings.HasPrefix(path, "_\\") {
		relPath := path[2:]
		if relPath == "" {
			return workspaceDir
		}
		return filepath.Join(workspaceDir, relPath)
	}
	return filepath.Join(projectRoot, path)
}

func resolvePathWithAnchor(projectRoot, workspaceDir, path string) string {
	return ResolvePathWithAnchor(projectRoot, workspaceDir, path)
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
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	workspace := FindWorkspaceRoot(cwd)
	projectRoot := ResolveProjectRoot(cwd)
	
	if workspace == "" {
		workspace = cwd
	}
	
	vars := ResolvePaths(workspace, projectRoot)

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
		{"project", func() string {
			if vars.Project == "" {
				return "(none - specify with --project)"
			}
			return vars.Project
		}()},
		{"TEAM_PATH", func() string {
			if vars.Project == "" {
				return "(none)"
			}
			return vars.TEAM_PATH
		}()},
		{"DOCS_INTERNAL", func() string {
			if vars.Project == "" {
				return "(none)"
			}
			return vars.DOCS_INTERNAL
		}()},
		{"DOCS_EXTERNAL", func() string {
			if vars.Project == "" {
				return "(none)"
			}
			return vars.DOCS_EXTERNAL
		}()},
		{"BACKUP_PATH", func() string {
			if vars.Project == "" {
				return "(none)"
			}
			return vars.BACKUP_PATH
		}()},
		{"BEADS_DB", func() string {
			if vars.Project == "" {
				return "(none)"
			}
			return vars.BEADS_DB
		}()},
		{"SKILL_PATH", vars.SKILL_PATH},
		{"FLOW_DIR", func() string {
			if vars.Project == "" {
				return "(none)"
			}
			return vars.FLOW_DIR
		}()},
		{"ACTIVE_FLOW", vars.ACTIVE_FLOW},
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

func resolveActiveFlowFromMD(root string) (string, error) {
	projectMD := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMD)
	if err != nil {
		return "", fmt.Errorf("no active flow configured")
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "active_flow:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				val := strings.Trim(strings.TrimSpace(parts[1]), "\"' ")
				if val != "" {
					return val, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no active flow configured")
}
