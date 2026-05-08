package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/origadmin/team-flow/internal/bd"
)

var (
	exportFormat string
	exportOutput string
)

var Cmd = &cobra.Command{
	Use:   "export",
	Short: "Export beads data to human-readable markdown",
	Long: `Export beads task data to human-readable markdown format.

Reads docs_path from .team/project.md for output location.
Falls back to .team/docs/ if not configured.

Usage:
  flow export                  Export to configured docs_path
  flow export --format json    Export as JSON
  flow export --output PATH    Override output path`,
	RunE: runExport,
}

func init() {
	Cmd.Flags().StringVar(&exportFormat, "format", "markdown", "Output format: markdown, json")
	Cmd.Flags().StringVar(&exportOutput, "output", "", "Override output path")
}

func runExport(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	versionFile := filepath.Join(projectPath, ".team", "version")
	if data, err := os.ReadFile(versionFile); err != nil || strings.TrimSpace(string(data)) != "v2" {
		return fmt.Errorf("export requires v2 project (beads). Run: flow migrate")
	}

	docsPath := exportOutput
	if docsPath == "" {
		docsPath = resolveDocsPath(projectPath)
	}

	if err := os.MkdirAll(docsPath, 0755); err != nil {
		return fmt.Errorf("create docs path: %w", err)
	}

	if !bd.IsAvailable() {
		return fmt.Errorf("bd CLI not found. Install beads first")
	}

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        team-flow Export (v2 → Human)     ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("  Project:  %s\n", projectPath)
	fmt.Printf("  Docs:     %s\n", docsPath)
	fmt.Printf("  Format:   %s\n\n", exportFormat)

	output, err := bd.RunQuiet("list", "--json")
	if err != nil {
		return fmt.Errorf("bd list failed: %w\nOutput: %s", err, string(output))
	}

	var issues []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &issues); err != nil {
		return fmt.Errorf("parse bd list output: %w", err)
	}

	fmt.Printf("  Found %d issues\n\n", len(issues))

	switch exportFormat {
	case "json":
		return exportJSON(docsPath, issues)
	default:
		return exportMarkdown(projectPath, docsPath, issues)
	}
}

func resolveDocsPath(projectPath string) string {
	projectMd := filepath.Join(projectPath, ".team", "project.md")
	if data, err := os.ReadFile(projectMd); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "docs_path:") || strings.HasPrefix(line, "docs_path ") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					path := strings.TrimSpace(parts[1])
					path = strings.Trim(path, "\"'")
					if path != "" {
						if filepath.IsAbs(path) {
							return path
						}
						return filepath.Join(projectPath, path)
					}
				}
			}
		}
	}

	projectName := filepath.Base(projectPath)
	candidates := []string{
		filepath.Join(projectPath, "_docs", projectName),
		filepath.Join(projectPath, ".team", "docs"),
	}
	for _, c := range candidates {
		return c
	}
	return candidates[0]
}

func exportJSON(docsPath string, issues []map[string]interface{}) error {
	outFile := filepath.Join(docsPath, "tasks.json")
	data, _ := json.MarshalIndent(issues, "", "  ")
	if err := os.WriteFile(outFile, data, 0644); err != nil {
		return fmt.Errorf("write tasks.json: %w", err)
	}
	fmt.Printf("  ✓ Exported %d issues to %s\n", len(issues), outFile)
	return nil
}

func exportMarkdown(projectPath, docsPath string, issues []map[string]interface{}) error {
	var sb strings.Builder

	sb.WriteString("# Task Pool Export\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Project**: %s\n", filepath.Base(projectPath)))
	sb.WriteString(fmt.Sprintf("**Total issues**: %d\n\n", len(issues)))

	openIssues := filterByStatus(issues, "open")
	inProgress := filterByStatus(issues, "in_progress")
	closed := filterByStatus(issues, "closed")

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Status | Count |\n|--------|-------|\n| Open | %d |\n| In Progress | %d |\n| Closed | %d |\n\n", len(openIssues), len(inProgress), len(closed)))

	if len(openIssues) > 0 {
		sb.WriteString("## Open Tasks\n\n")
		sb.WriteString("| ID | Title | Type | Priority | Assignee | External Ref |\n")
		sb.WriteString("|----|-------|------|----------|----------|-------------|\n")
		for _, issue := range openIssues {
			sb.WriteString(formatIssueRow(issue))
		}
		sb.WriteString("\n")
	}

	if len(inProgress) > 0 {
		sb.WriteString("## In Progress\n\n")
		sb.WriteString("| ID | Title | Type | Priority | Assignee | External Ref |\n")
		sb.WriteString("|----|-------|------|----------|----------|-------------|\n")
		for _, issue := range inProgress {
			sb.WriteString(formatIssueRow(issue))
		}
		sb.WriteString("\n")
	}

	if len(closed) > 0 {
		sb.WriteString("## Recently Closed\n\n")
		sb.WriteString("| ID | Title | Type | Priority | Assignee | External Ref |\n")
		sb.WriteString("|----|-------|------|----------|----------|-------------|\n")
		for _, issue := range closed {
			sb.WriteString(formatIssueRow(issue))
		}
		sb.WriteString("\n")
	}

	outFile := filepath.Join(docsPath, "task-pool.md")
	if err := os.WriteFile(outFile, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write task-pool.md: %w", err)
	}

	fmt.Printf("  ✓ Exported %d issues to %s\n", len(issues), outFile)
	return nil
}

func filterByStatus(issues []map[string]interface{}, status string) []map[string]interface{} {
	var result []map[string]interface{}
	for _, issue := range issues {
		if s, ok := issue["status"].(string); ok && s == status {
			result = append(result, issue)
		}
	}
	return result
}

func formatIssueRow(issue map[string]interface{}) string {
	id := strVal(issue, "id")
	title := strVal(issue, "title")
	issueType := strVal(issue, "type")
	priority := fmt.Sprintf("%v", issue["priority"])
	assignee := strVal(issue, "assignee")
	extRef := strVal(issue, "external_ref")
	if assignee == "" {
		assignee = "-"
	}
	if extRef == "" {
		extRef = "-"
	}
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n", id, title, issueType, priority, assignee, extRef)
}

func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
