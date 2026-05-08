package migrate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/logger"
	"github.com/spf13/cobra"
)

var (
	migrateDryRun bool
	migrateForce  bool
	log           *logger.Logger
)

type TaskPoolEntry struct {
	ID       string
	Type     string
	Title    string
	Priority string
	Status   string
	Phase    string
	Assignee string
	Docs     string
}

var Cmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate from v1 (task-pool) to v2 (beads-native)",
	Long: `Migrate project from v1 (task-pool.md) to v2 (beads-native + code-review-graph).

Steps:
  1. Parse task-pool.md entries
  2. Initialize beads database (bd init)
  3. Create beads issues from task-pool entries
  4. Build code-review-graph
  5. Update project configuration`,
	RunE: runMigrate,
}

func init() {
	Cmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Dry run, do not execute changes")
	Cmd.Flags().BoolVar(&migrateForce, "force", false, "Force migration even if .beads exists")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	log := logger.GetLogger()
	log.Debug("Starting migration")

	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	log.Debug("Starting migration: " + projectPath)

	versionFile := filepath.Join(projectPath, ".team", "version")
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion := strings.TrimSpace(string(data))
		if currentVersion == "v2" {
			fmt.Println("Already on v2. No migration needed.")
			return nil
		}
		if currentVersion != "v1" {
			fmt.Printf("Unknown version %q. Expected v1.\n", currentVersion)
			fmt.Println("Run: flow init --v1")
			return nil
		}
	} else {
		fmt.Println("⚠ .team/version not found. Assuming v1 project.")
	}

	fmt.Println("=== v1 → v2 Migration ===")
	fmt.Printf("Project: %s\n\n", projectPath)

	taskPoolPath := findTaskPool(projectPath)
	if taskPoolPath == "" {
		return fmt.Errorf("task-pool.md not found. Searched: .team/, .trae/skills/team-flow/, .cursor/skills/team-flow/, .team-flow/skill/")
	}
	fmt.Printf("Found task-pool: %s\n", taskPoolPath)

	entries, err := parseTaskPool(taskPoolPath)
	if err != nil {
		return fmt.Errorf("parse task-pool: %w", err)
	}
	fmt.Printf("Parsed %d task entries\n\n", len(entries))

	if migrateDryRun {
		fmt.Println("=== DRY RUN ===")
		for _, e := range entries {
			fmt.Printf("  %s [%s] %s (P%s, %s)\n", e.ID, e.Type, e.Title, e.Priority, e.Status)
		}
		fmt.Printf("\nWould create %d beads issues\n", len(entries))
		return nil
	}

	bdCmd := bd.FindPath()
	if bdCmd == "" {
		fmt.Println("\n⚠ beads (bd CLI) not found.")
		fmt.Println("Installing beads...")
		if err := bd.Install(); err != nil {
			return fmt.Errorf("bd installation failed: %w\nManual install: https://github.com/steveyegge/beads", err)
		}
		bdCmd = bd.FindPath()
		if bdCmd != "" {
			bd.EnsureOnPath()
		}
	}

	if bdCmd == "" {
		return fmt.Errorf("bd not available after installation")
	}

	if err := bd.EnsureOnPath(); err != nil {
		fmt.Printf("  ⚠ %v\n", err)
	}

	beadsDir := filepath.Join(projectPath, ".beads")
	if _, err := os.Stat(beadsDir); os.IsNotExist(err) || migrateForce {
		fmt.Println("Initializing beads database...")
		if _, err := bd.Run("init"); err != nil {
			return fmt.Errorf("bd init: %w", err)
		}
		fmt.Println("  ✓ bd init done")
	} else {
		fmt.Println("  .beads already exists, skipping bd init")
	}

	fmt.Printf("\nMigrating %d tasks to beads...\n", len(entries))
	migrated := 0
	for _, e := range entries {
		issueType := mapType(e.Type)
		priority := mapPriority(e.Priority)

		createArgs := []string{
			"create", e.Title,
			"-t", issueType,
			"-p", fmt.Sprintf("%d", priority),
			"--external-ref", e.ID,
			"--json",
		}

		output, _ := bd.Run(createArgs...)

		jsonOutput, jsonErr := extractJSON(output)
		if jsonErr != nil {
			fmt.Printf("  ✗ %s: failed to extract JSON: %v\n", e.ID, jsonErr)
			fmt.Printf("    Raw output: %s\n", truncateString(output, 200))
			continue
		}

		var results []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonOutput), &results); err != nil {
			var singleResult map[string]interface{}
			if err2 := json.Unmarshal([]byte(jsonOutput), &singleResult); err2 != nil {
				fmt.Printf("  ✗ %s: parse error: %v\n", e.ID, err)
				continue
			}
			results = []map[string]interface{}{singleResult}
		}

		if len(results) > 0 {
			result := results[0]
			if beadsID, ok := result["id"].(string); ok {
				var title string
				if t, ok := result["title"].(string); ok {
					title = t
					fmt.Printf("  ✓ %s → %s (%s)\n", e.ID, beadsID, title)
				} else {
					fmt.Printf("  ✓ %s → %s\n", e.ID, beadsID)
				}

				if e.Status == "Doing" || e.Status == "Review" {
					bd.Run("update", beadsID, "--status", "in_progress")
				}
				if e.Assignee != "" && e.Assignee != "-" {
					bd.Run("update", beadsID, "--assignee", e.Assignee)
				}
				if title != "" {
					log.Debugf("Created issue %s: %s", beadsID, title)
				}
				notes := fmt.Sprintf("MIGRATED FROM v1 task-pool. Original ID: %s, Status: %s, Phase: %s", e.ID, e.Status, e.Phase)
				if e.Docs != "" && e.Docs != "-" {
					notes += fmt.Sprintf(", Docs: %s", e.Docs)
				}
				bd.Run("update", beadsID, "--notes", notes)
			}
		}

		if _, ok := results[0]["id"].(string); ok {
			migrated++
		}
	}

	fmt.Printf("\n✅ Migrated %d/%d tasks\n", migrated, len(entries))

	if migrated == 0 {
		fmt.Println("\n⚠ No tasks migrated successfully. Not updating version.")
		fmt.Println("Please check the errors above and try again.")
		return fmt.Errorf("migration failed: no tasks migrated")
	}

	versionFile = filepath.Join(projectPath, ".team", "version")
	if err := os.WriteFile(versionFile, []byte("v2"), 0644); err != nil {
		fmt.Printf("  ⚠ Failed to update .team/version: %v\n", err)
	} else {
		fmt.Println("  ✓ .team/version updated to v2")
	}

	graphDir := filepath.Join(projectPath, ".code-review-graph")
	if _, err := os.Stat(graphDir); os.IsNotExist(err) {
		fmt.Println("\nBuilding code-review-graph...")
		if err := runCmd("python", "-m", "code_review_graph", "build"); err != nil {
			fmt.Printf("  ⚠ code-review-graph build failed: %v\n", err)
			fmt.Println("  You can build it later: python -m code_review_graph build")
		} else {
			fmt.Println("  ✓ Graph built successfully")
		}
	}

	fmt.Println("\n=== Migration Complete ===")
	fmt.Println("Next steps:")
	fmt.Println("  1. Verify: bd ready --json")
	fmt.Println("  2. Check graph: python -m code_review_graph status")
	fmt.Println("  3. Update AI config to use v2 SKILL.md")
	fmt.Println("  4. Backup and remove old task-pool.md")

	return nil
}

func findTaskPool(projectPath string) string {
	candidates := []string{
		filepath.Join(projectPath, ".team", "task-pool.md"),
		filepath.Join(projectPath, ".trae", "skills", "team-flow", "task-pool.md"),
		filepath.Join(projectPath, ".cursor", "skills", "team-flow", "task-pool.md"),
		filepath.Join(projectPath, ".claude", "skills", "team-flow", "task-pool.md"),
		filepath.Join(projectPath, ".team-flow", "skill", "task-pool.md"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func parseTaskPool(path string) ([]TaskPoolEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []TaskPoolEntry
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	tableRegex := regexp.MustCompile(`^\|\s*(F|B|C|A|D)\d{3}\s*\|`)

	for scanner.Scan() {
		line := scanner.Text()
		if !tableRegex.MatchString(line) {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}

		cleanFields := make([]string, len(fields))
		for i, f := range fields {
			cleanFields[i] = strings.TrimSpace(f)
		}

		entry := TaskPoolEntry{
			ID:       cleanFields[1],
			Title:    cleanFields[2],
			Priority: cleanFields[3],
			Status:   cleanFields[4],
			Phase:    cleanFields[5],
			Docs:     cleanFields[6],
		}

		if len(entry.ID) > 0 {
			switch strings.ToUpper(entry.ID[:1]) {
			case "F":
				entry.Type = "feature"
			case "B":
				entry.Type = "bug"
			case "C":
				entry.Type = "task"
			case "A":
				entry.Type = "task"
			case "D":
				entry.Type = "task"
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func mapType(t string) string {
	switch t {
	case "feature":
		return "feature"
	case "bug":
		return "bug"
	default:
		return "task"
	}
}

func mapPriority(p string) int {
	switch strings.TrimPrefix(p, "P") {
	case "0":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	default:
		return 2
	}
}

func ensureBdInstalled() error {
	if !bd.IsAvailable() {
		return fmt.Errorf("bd CLI not found")
	}
	return nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractJSON(output string) (string, error) {
	start := strings.Index(output, "[")
	if start == -1 {
		start = strings.Index(output, "{")
	}
	if start == -1 {
		return "", fmt.Errorf("no JSON found in output")
	}

	end := strings.LastIndex(output, "]")
	if end == -1 || !strings.Contains(output[:end], "[") {
		end = strings.LastIndex(output, "}")
	}

	return strings.TrimSpace(output[start : end+1]), nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
