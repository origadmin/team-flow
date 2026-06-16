package validate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/proc"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "validate [task-id] [flow-type]",
	Short: "Validate task completion against checklist",
	Long: `Validate task completion against the completion checklist.

Examples:
  flow validate F001 feature    # Validate Feature task
  flow validate B001 bugfix     # Validate Bugfix task

Reads checklist from .team/checklist.md with evidence markers and verifies all items.`,
	RunE: runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: flow validate <task-id> <flow-type>")
	}
	taskID := args[0]
	flowType := args[1]

	cwd, _ := os.Getwd()
	projectRoot := cwd

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║      team-flow Completion Checker        ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("  Task: %s\n", taskID)
	fmt.Printf("  Type: %s\n\n", flowType)

	// Step 1: verify task exists
	if !taskExists(taskID, projectRoot) {
		fmt.Println("  ──────────────────────────────────────────")
		fmt.Println("  ❌ FAIL: Task not found")
		fmt.Printf("  Task %q does not exist. Create it with:\n", taskID)
		fmt.Println("    flow task create --id " + taskID + " --title <name>")
		return fmt.Errorf("task %q not found", taskID)
	}
	fmt.Println("  ✅ Task exists")

	items, err := loadChecklist(projectRoot)
	if err != nil {
		return fmt.Errorf("load checklist: %w", err)
	}

	passedCount, failedCount := 0, 0
	for _, item := range items {
		result := evaluateItem(item, taskID, projectRoot)
		if result.Passed {
			passedCount++
			fmt.Printf("  ✅ %s  [%s]\n", result.Desc, result.Evidence)
		} else {
			failedCount++
			fmt.Printf("  ❌ %s  [%s] — %s\n", result.Desc, result.Evidence, result.Message)
		}
	}

	fmt.Println("  ──────────────────────────────────────────")
	fmt.Printf("  ✅ Passed: %d\n", passedCount)
	fmt.Printf("  ❌ Failed: %d\n", failedCount)
	fmt.Println("  ──────────────────────────────────────────")

	if failedCount == 0 {
		fmt.Println("\n  ✅ All checks passed! Task is ready for completion.")
		return nil
	}
	fmt.Println("\n  ❌ Some checks failed. Please address the issues above.")
	return fmt.Errorf("validation failed: %d items not completed", failedCount)
}

func loadChecklist(projectRoot string) ([]proc.ChecklistEvidence, error) {
	checklistPath := filepath.Join(projectRoot, ".team", "checklist.md")

	data, err := os.ReadFile(checklistPath)
	if err != nil {
		return nil, fmt.Errorf("read project checklist: %w (create .team/checklist.md or run 'flow init')", err)
	}

	items := proc.ParseChecklistEvidence(string(data))
	if len(items) == 0 {
		return nil, fmt.Errorf("no checklist items found in %s", checklistPath)
	}

	return items, nil
}

func evaluateItem(item proc.ChecklistEvidence, taskID string, projectRoot string) proc.EvidenceResult {
	r := proc.EvidenceResult{Desc: item.Desc, Evidence: item.EvidenceType}

	switch item.EvidenceType {
	case "command":
		if item.Cmd == "" {
			r.Passed = false
			r.Message = "no command specified"
			return r
		}
		// Simple execution — use PowerShell on Windows
		output := runEvidenceCommand(item.Cmd, projectRoot)
		if output.err != nil {
			r.Passed = false
			r.Message = truncate(output.out, 200)
		} else {
			r.Passed = true
			r.Message = "OK"
		}
		return r

	case "file_exists":
		path := strings.ReplaceAll(item.Path, "{task_id}", taskID)
		if !filepath.IsAbs(path) {
			path = filepath.Join(projectRoot, path)
		}
		if _, err := os.Stat(path); err == nil {
			r.Passed = true
			r.Message = path
		} else {
			r.Passed = false
			r.Message = fmt.Sprintf("not found: %s", path)
		}
		return r

	case "ai_judgment":
		r.Passed = false
		r.Message = "requires AI judgment — run explicitly to confirm"
		return r

	case "pinchtab":
		r.Passed = false
		var details []string
		if item.URL != "" {
			details = append(details, fmt.Sprintf("url=%s", item.URL))
		}
		if item.Steps != "" {
			details = append(details, fmt.Sprintf("steps=%s", item.Steps))
		}
		if item.Auth != "" {
			details = append(details, fmt.Sprintf("auth=%s", item.Auth))
		}
		if len(details) > 0 {
			r.Message = fmt.Sprintf("requires PinchTab session [%s] — run explicitly to confirm", strings.Join(details, ", "))
		} else {
			r.Message = "requires PinchTab session — run explicitly to confirm"
		}
		return r

	default:
		// Legacy: check by keyword
		docsRoot := proc.ResolveDocsPath(projectRoot)
		desc := strings.ToLower(item.Desc)
		if strings.Contains(desc, "spec.md") {
			path := filepath.Join(docsRoot, "requirements", fmt.Sprintf("%s-spec.md", taskID))
			if _, err := os.Stat(path); err == nil {
				r.Passed = true
				r.Message = path
				return r
			}
		}
		if strings.Contains(desc, "ac.md") {
			path := filepath.Join(docsRoot, "requirements", fmt.Sprintf("%s-ac.md", taskID))
			if _, err := os.Stat(path); err == nil {
				r.Passed = true
				r.Message = path
				return r
			}
		}
		r.Passed = false
		r.Message = "no evidence type, cannot verify"
		return r
	}
}

type cmdResult struct {
	out string
	err error
}

func runEvidenceCommand(command string, dir string) cmdResult {
	var name string
	var args []string
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return cmdResult{err: fmt.Errorf("empty command")}
	}
	name = parts[0]
	args = parts[1:]

	return runCmd(name, args, dir)
}

func runCmd(name string, args []string, dir string) cmdResult {
	ecmd := exec.Command(name, args...)
	ecmd.Dir = dir
	output, err := ecmd.CombinedOutput()
	return cmdResult{out: string(output), err: err}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// taskExists verifies the task exists in the local .team/tasks/ directory.
func taskExists(taskID string, projectRoot string) bool {
	if taskID == "" {
		return false
	}
	taskPath := filepath.Join(projectRoot, ".team", "tasks", taskID+".json")
	_, err := os.Stat(taskPath)
	return err == nil
}