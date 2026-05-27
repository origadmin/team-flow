package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/config"
)

type ProjectContext struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Version   string `json:"version"`
	Workspace string `json:"workspace"`
	TeamRoot  string `json:"team_root,omitempty"`
	Type      string `json:"type,omitempty"`
	Source    string `json:"source,omitempty"`
	HasTeam   bool   `json:"has_team,omitempty"`
	IsLocked  bool   `json:"is_locked,omitempty"`
}

type ProjectDetectionResult struct {
	CurrentProject    *ProjectContext          `json:"current,omitempty"`
	AvailableProjects []ProjectContext         `json:"available_projects"`
	Workspace         string                   `json:"workspace"`
	TeamRoot          string                   `json:"team_root,omitempty"`
	NeedsConfirmation bool                     `json:"needs_confirmation"`
	Status            string                   `json:"status"`
	SelectionGuidance string                   `json:"selection_guidance,omitempty"`
}

func DetectProject() *ProjectDetectionResult {
	cwd, err := os.Getwd()
	if err != nil {
		return &ProjectDetectionResult{Status: "error"}
	}

	workspace := config.FindWorkspaceRoot(cwd)
	if workspace == "" {
		workspace = filepath.Dir(cwd)
	}

	teamRoot := config.FindTeamRoot(cwd)
	projectRoot := config.ResolveProjectRoot(cwd)

	result := &ProjectDetectionResult{
		Workspace: workspace,
		TeamRoot:  teamRoot,
	}

	discovered := config.DiscoverProjects(workspace)
	for _, dp := range discovered {
		pc := ProjectContext{
			Name:      dp.Name,
			Path:      dp.Path,
			Version:   getProjectVersion(dp.Path),
			Workspace: workspace,
			TeamRoot:  teamRoot,
			Type:      dp.Type,
			Source:    dp.Source,
			HasTeam:   dp.HasTeam,
		}
		result.AvailableProjects = append(result.AvailableProjects, pc)
	}

	if projectRoot != "" {
		isLocked := projectRoot != teamRoot
		result.CurrentProject = &ProjectContext{
			Name:      getProjectName(projectRoot),
			Path:      projectRoot,
			Version:   getProjectVersion(projectRoot),
			Workspace: workspace,
			TeamRoot:  teamRoot,
			IsLocked:  isLocked,
		}
	}

	if projectRoot != "" && projectRoot != teamRoot {
		result.Status = "ready"
		result.SelectionGuidance = "Project locked. Run 'flow proc run' to continue."
	} else if len(result.AvailableProjects) == 1 {
		result.CurrentProject = &result.AvailableProjects[0]
		result.CurrentProject.IsLocked = true
		result.Status = "ready"
		result.SelectionGuidance = "Single project detected. Run 'flow proc run' to continue."
	} else if len(result.AvailableProjects) > 1 {
		result.NeedsConfirmation = true
		result.Status = "project_not_selected"
		result.SelectionGuidance = "Multiple projects detected. cd to the project directory, then run 'flow proc run'."
	} else if teamRoot != "" {
		result.Status = "ready"
		result.SelectionGuidance = "Single workspace. Run 'flow proc run' to continue."
	} else {
		result.Status = "no_workspace"
		result.SelectionGuidance = "No workspace found. Run 'flow init --v3' to initialize."
	}

	return result
}

func ListProjects() []ProjectContext {
	cwd, err := os.Getwd()
	if err != nil {
		return []ProjectContext{}
	}

	workspace := config.FindWorkspaceRoot(cwd)
	if workspace == "" {
		workspace = filepath.Dir(cwd)
	}

	teamRoot := config.FindTeamRoot(cwd)
	discovered := config.DiscoverProjects(workspace)
	var projects []ProjectContext

	for _, dp := range discovered {
		projects = append(projects, ProjectContext{
			Name:      dp.Name,
			Path:      dp.Path,
			Version:   getProjectVersion(dp.Path),
			Workspace: workspace,
			TeamRoot:  teamRoot,
			Type:      dp.Type,
			Source:    dp.Source,
			HasTeam:   dp.HasTeam,
		})
	}

	return projects
}

func getProjectName(projectPath string) string {
	projectMD := filepath.Join(projectPath, ".team", "project.md")
	data, err := os.ReadFile(projectMD)
	if err != nil {
		return filepath.Base(projectPath)
	}

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "name:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return filepath.Base(projectPath)
}

func getProjectVersion(projectPath string) string {
	versionFile := filepath.Join(projectPath, ".team", "version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func PrintDetectionResult(result *ProjectDetectionResult) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")

	statusIcon := "✓"
	switch result.Status {
	case "ready":
		statusIcon = "✓"
	case "project_not_selected":
		statusIcon = "⚠"
	case "no_workspace":
		statusIcon = "✗"
	default:
		statusIcon = "?"
	}

	fmt.Fprintf(os.Stdout, "║  %s PROJECT DETECT%-47s║\n", statusIcon, "")
	fmt.Fprintf(os.Stdout, "║%-70s║\n", "")
	fmt.Fprintf(os.Stdout, "║  WORKSPACE: %-54s║\n", result.Workspace)

	if result.TeamRoot != "" && result.TeamRoot != result.Workspace {
		fmt.Fprintf(os.Stdout, "║  TEAM_ROOT: %-54s║\n", result.TeamRoot)
	}

	if result.CurrentProject != nil {
		lockLabel := ""
		if result.CurrentProject.IsLocked {
			lockLabel = " (locked)"
		}
		fmt.Fprintf(os.Stdout, "║  PROJECT: %s%s\n", padLine(result.CurrentProject.Name+lockLabel, 55), "║")
		fmt.Fprintf(os.Stdout, "║    → %-60s║\n", result.CurrentProject.Path)
		if result.CurrentProject.TeamRoot != "" && result.CurrentProject.TeamRoot != result.CurrentProject.Path {
			fmt.Fprintf(os.Stdout, "║    team_config: %-48s║\n", result.CurrentProject.TeamRoot)
		}
	}

	if len(result.AvailableProjects) > 0 {
		fmt.Fprintf(os.Stdout, "║%-70s║\n", "")
		fmt.Fprintf(os.Stdout, "║  AVAILABLE PROJECTS:%-48s║\n", "")
		for i, p := range result.AvailableProjects {
			marker := ""
			if result.CurrentProject != nil && p.Path == result.CurrentProject.Path {
				marker = " ← current"
			}
			typeLabel := ""
			if p.Type != "" {
				typeLabel = " (" + p.Type + ")"
			}
			teamLabel := ""
			if p.HasTeam {
				teamLabel = " ✅"
			}
			relPath := p.Path
			if result.Workspace != "" {
				if rel, err := filepath.Rel(result.Workspace, p.Path); err == nil {
					relPath = rel
				}
			}
			fmt.Fprintf(os.Stdout, "║    [%d] %s%s%s%s\n", i+1, padLine(p.Name, 20), typeLabel, teamLabel, marker)
			fmt.Fprintf(os.Stdout, "║        → %-56s║\n", relPath)
		}
	}

	fmt.Fprintf(os.Stdout, "║%-70s║\n", "")
	fmt.Fprintf(os.Stdout, "║  STATUS: %-56s║\n", result.Status)
	if result.SelectionGuidance != "" {
		lines := wrapGuidance(result.SelectionGuidance, 56)
		for _, line := range lines {
			fmt.Fprintf(os.Stdout, "║  %s\n", padLine(line, 67))
		}
	}

	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func wrapGuidance(s string, width int) []string {
	if len(s) <= width {
		return []string{s}
	}
	var lines []string
	for len(s) > width {
		idx := width
		for i := width; i > 0; i-- {
			if s[i] == ' ' {
				idx = i + 1
				break
			}
		}
		lines = append(lines, s[:idx])
		s = s[idx:]
	}
	if s != "" {
		lines = append(lines, s)
	}
	return lines
}

func padLine(s string, width int) string {
	if len(s) > width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func PrintProjectList(projects []ProjectContext, currentPath string) {
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	for i, p := range projects {
		marker := ""
		if p.Path == currentPath {
			marker = " ← current"
		}

		versionLabel := ""
		if p.Version != "" {
			versionLabel = fmt.Sprintf(" (v%s)", p.Version)
		}

		fmt.Printf("[%d] %s%s\n", i+1, p.Name, versionLabel)
		fmt.Printf("    → %s%s\n\n", p.Path, marker)
	}
}
