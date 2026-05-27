package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/spf13/cobra"
)

var (
	traceDecision string
	traceReason   string
	traceAnomaly  string
	traceModify   string
	traceRole     string
)

var traceCmd = &cobra.Command{
	Use:   "trace [task_id] --decision <text>",
	Short: "Append a decision trace entry to TRACE.md",
	Long: `Append a decision trace entry to the task's TRACE.md file.

This is the MANDATORY mechanism for recording decisions, anomalies,
and modifications during a flow. Every decision point MUST produce
a trace entry — missing trace = blocked flow.

Examples:
  flow task trace orig-cms-ee-7dv --decision "B" --reason "need international payments"
  flow task trace orig-cms-ee-7dv --anomaly "B has circular dependency" --modify "B→B'" --reason "interface decoupling"
  flow task trace orig-cms-ee-7dv --decision "B''" --modify "B'→B''" --reason "simplify crypto module"`,
	Args: cobra.ExactArgs(1),
	RunE: runTrace,
}

func init() {
	traceCmd.Flags().StringVar(&traceDecision, "decision", "", "Decision made (e.g., 'B', 'refactor', 'approve')")
	traceCmd.Flags().StringVar(&traceReason, "reason", "", "Reason for the decision")
	traceCmd.Flags().StringVar(&traceAnomaly, "anomaly", "", "Anomaly or problem discovered")
	traceCmd.Flags().StringVar(&traceModify, "modify", "", "Decision modification path (e.g., 'B→B→B')")
	traceCmd.Flags().StringVar(&traceRole, "role", "", "Role that made the decision (e.g., triage, design, dev)")
	Cmd.AddCommand(traceCmd)
}

func runTrace(cmd *cobra.Command, args []string) error {
	taskID := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	projectRoot := config.ResolveProjectRoot(cwd)
	if projectRoot == "" {
		projectRoot = cwd
	}

	docsInternal := config.ResolveInternalDocs(projectRoot)
	if docsInternal == "" {
		docsInternal = filepath.Join(projectRoot, "_docs")
	}

	traceDir := filepath.Join(docsInternal, "task", taskID)
	tracePath := filepath.Join(traceDir, "TRACE.md")

	if err := os.MkdirAll(traceDir, 0755); err != nil {
		return fmt.Errorf("create trace directory: %w", err)
	}

	entry := buildTraceEntry(taskID)

	f, err := os.OpenFile(tracePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open trace file: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat trace file: %w", err)
	}

	if stat.Size() == 0 {
		header := fmt.Sprintf("# Decision Trace\n\n> task_id: %s\n> created: %s\n\n---\n\n",
			taskID, time.Now().Format("2006-01-02"))
		if _, err := f.WriteString(header); err != nil {
			return fmt.Errorf("write trace header: %w", err)
		}
	}

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("write trace entry: %w", err)
	}

	if err := updateIndex(traceDir, taskID, docsInternal); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ INDEX.md update failed: %v\n", err)
	}

	fmt.Printf("✓ Trace entry appended to %s\n", tracePath)
	return nil
}

func buildTraceEntry(taskID string) string {
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04")
	role := traceRole
	if role == "" {
		role = "unknown"
	}

	counter := getTraceCount(taskID) + 1

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## T%d | %s | %s\n\n", counter, timestamp, role))

	if traceAnomaly != "" {
		sb.WriteString(fmt.Sprintf("**异常**: %s\n", traceAnomaly))
	}

	if traceAnomaly != "" && traceModify != "" {
		sb.WriteString(fmt.Sprintf("**修改**: %s\n", traceModify))
	}

	if traceDecision != "" {
		if traceAnomaly != "" {
			sb.WriteString(fmt.Sprintf("**决策**: %s\n", traceDecision))
		} else {
			sb.WriteString(fmt.Sprintf("**决策**: %s\n", traceDecision))
		}
	}

	if traceReason != "" {
		sb.WriteString(fmt.Sprintf("**理由**: %s\n", traceReason))
	}

	sb.WriteString("\n---\n\n")
	return sb.String()
}

func getTraceCount(taskID string) int {
	cwd, _ := os.Getwd()
	projectRoot := config.ResolveProjectRoot(cwd)
	if projectRoot == "" {
		projectRoot = cwd
	}

	docsInternal := config.ResolveInternalDocs(projectRoot)
	if docsInternal == "" {
		docsInternal = filepath.Join(projectRoot, "_docs")
	}

	tracePath := filepath.Join(docsInternal, "task", taskID, "TRACE.md")
	data, err := os.ReadFile(tracePath)
	if err != nil {
		return 0
	}

	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "## T") {
			count++
		}
	}
	return count
}

func updateIndex(traceDir string, taskID string, docsInternal string) error {
	indexPath := filepath.Join(traceDir, "INDEX.md")

	traceCount := getTraceCount(taskID)

	var docLinks []string
	filepath.Walk(docsInternal, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if info.Name() == "task" {
			return filepath.SkipDir
		}
		if info.Name() == taskID {
			if files, err := os.ReadDir(path); err == nil {
				for _, f := range files {
					if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
						relPath, _ := filepath.Rel(docsInternal, filepath.Join(path, f.Name()))
						docLinks = append(docLinks, fmt.Sprintf("- %s → %s", f.Name(), relPath))
					}
				}
			}
			return filepath.SkipDir
		}
		return nil
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Task Index: %s\n\n", taskID))
	sb.WriteString("## 流程状态\n\n")
	sb.WriteString(fmt.Sprintf("- task_id: %s\n", taskID))
	sb.WriteString(fmt.Sprintf("- trace_entries: %d\n", traceCount))
	sb.WriteString(fmt.Sprintf("- updated: %s\n\n", time.Now().Format("2006-01-02 15:04")))

	if len(docLinks) > 0 {
		sb.WriteString("## 文档索引\n\n")
		for _, link := range docLinks {
			sb.WriteString(link + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 决策追踪\n\n")
	sb.WriteString(fmt.Sprintf("- TRACE → task/%s/TRACE.md (%d entries)\n", taskID, traceCount))

	return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}
