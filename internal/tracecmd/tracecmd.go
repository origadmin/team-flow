// Package tracecmd provides the flow trace CLI commands for post-mortem analysis.
//
// Commands:
//   flow trace show <session-name>  — Display full trace timeline
//   flow trace issues <session-name> — Detect issues from trace data
package tracecmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/spf13/cobra"
)

var (
	traceSession string
	traceFormat  string
)

// Cmd is the root trace command registered under flow trace.
var Cmd = &cobra.Command{
	Use:   "trace",
	Short: "Post-mortem trace analysis",
	Long:  `Read and analyze per-session trace.jsonl files for debugging and issue detection.`,
}

var showCmd = &cobra.Command{
	Use:   "show [session-name]",
	Short: "Show trace timeline for a session",
	Long: `Display the full trace timeline for a session.

Without session-name, uses the most recent session.
Use --format json for machine-readable output.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTraceShow,
}

var issuesCmd = &cobra.Command{
	Use:   "issues [session-name]",
	Short: "Detect issues from session trace data",
	Long: `Scan a session's trace.jsonl for anomalies (state drift, stuck nodes, loops).

Without session-name, uses the most recent session.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTraceIssues,
}

func init() {
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(issuesCmd)
	showCmd.Flags().StringVar(&traceFormat, "format", "text", "Output format: text or json")
	issuesCmd.Flags().StringVar(&traceFormat, "format", "text", "Output format: text or json")
}

func runTraceShow(cmd *cobra.Command, args []string) error {
	sessionName := ""
	if len(args) > 0 {
		sessionName = args[0]
	}

	projectRoot := findProjectRoot()
	if projectRoot == "" {
		return fmt.Errorf("project not found")
	}

	tracePath, err := resolveTracePath(projectRoot, sessionName)
	if err != nil {
		return err
	}

	events, err := readTrace(tracePath)
	if err != nil {
		return fmt.Errorf("read trace: %w", err)
	}

	if len(events) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trace events found.")
		return nil
	}

	if traceFormat == "json" {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(events)
	}

	// Text output: timeline
	fmt.Fprintf(cmd.OutOrStdout(), "Trace timeline (%d events)\n", len(events))
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("─", 80))
	for _, e := range events {
		formatTraceEvent(cmd, e)
	}
	return nil
}

func runTraceIssues(cmd *cobra.Command, args []string) error {
	sessionName := ""
	if len(args) > 0 {
		sessionName = args[0]
	}

	projectRoot := findProjectRoot()
	if projectRoot == "" {
		return fmt.Errorf("project not found")
	}

	tracePath, err := resolveTracePath(projectRoot, sessionName)
	if err != nil {
		return err
	}

	events, err := readTrace(tracePath)
	if err != nil {
		return fmt.Errorf("read trace: %w", err)
	}

	issues := detectIssues(events)

	if traceFormat == "json" {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(issues)
	}

	if len(issues) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "✓ No issues detected.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Issues detected: %d\n", len(issues))
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("─", 80))
	for _, issue := range issues {
		fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", issue.Severity, issue.Message)
		if issue.Detail != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Detail: %s\n", issue.Detail)
		}
		if issue.Suggestion != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Fix: %s\n", issue.Suggestion)
		}
	}
	return nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func findProjectRoot() string {
	dir, _ := os.Getwd()
	return config.ResolveProjectRoot(dir)
}

func resolveTracePath(projectRoot, sessionName string) (string, error) {
	teamDir := filepath.Join(projectRoot, ".team")
	traceDir := filepath.Join(teamDir, "sessions")

	if sessionName == "" {
		// Find most recent session with a trace file
		entries, err := os.ReadDir(traceDir)
		if err != nil {
			return "", fmt.Errorf("no sessions directory: %w", err)
		}
		var newest string
		var newestTime int64
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			traceFile := filepath.Join(traceDir, e.Name(), "trace.jsonl")
			info, err := os.Stat(traceFile)
			if err != nil {
				continue
			}
			if info.ModTime().Unix() > newestTime {
				newestTime = info.ModTime().Unix()
				newest = e.Name()
			}
		}
		if newest == "" {
			return "", fmt.Errorf("no session trace files found")
		}
		sessionName = newest
	}

	return filepath.Join(traceDir, sessionName, "trace.jsonl"), nil
}

func readTrace(path string) ([]map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []map[string]interface{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var evt map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &evt); err != nil {
			continue
		}
		events = append(events, evt)
	}
	return events, scanner.Err()
}

func formatTraceEvent(cmd *cobra.Command, evt map[string]interface{}) {
	ts, _ := evt["ts"].(string)
	event, _ := evt["event"].(string)
	if len(ts) > 19 {
		ts = ts[11:19] // HH:MM:SS only
	}

	switch event {
	case "flow.node_enter":
		name, _ := evt["name"].(string)
		if name == "" {
			name, _ = evt["node"].(string)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ▶ ENTER  %-30s\n", ts, name)

	case "flow.node_exit":
		name, _ := evt["name"].(string)
		if name == "" {
			name, _ = evt["node"].(string)
		}
		dur, _ := evt["duration_ms"].(float64)
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ◀ EXIT   %-30s  %.0fms\n", ts, name, dur)

	case "flow.edge_traverse":
		cond, _ := evt["condition"].(string)
		result, _ := evt["result"].(string)
		fmt.Fprintf(cmd.OutOrStdout(), "%s  → EDGE    %s → %s\n", ts, cond, result)

	case "flow.gate_check":
		cond, _ := evt["condition"].(string)
		passed, _ := evt["passed"].(bool)
		status := "✗ FAIL"
		if passed {
			status = "✓ PASS"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ⚑ GATE   %-30s %s\n", ts, cond, status)

	case "ai.tool_call":
		tool, _ := evt["tool"].(string)
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ⚙ TOOL   %s\n", ts, tool)

	case "ai.file_write":
		path, _ := evt["path"].(string)
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ✎ WRITE  %s\n", ts, path)

	case "ai.subagent_spawn":
		role, _ := evt["role"].(string)
		task, _ := evt["task"].(string)
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ⚡ SUB    %s → %s\n", ts, role, task)

	case "task.created":
		task, _ := evt["task"].(string)
		title, _ := evt["title"].(string)
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ✦ TASK   %s: %s\n", ts, task, title)

	case "task.status":
		task, _ := evt["task"].(string)
		from, _ := evt["from"].(string)
		to, _ := evt["to"].(string)
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ⤴ STATUS %s: %s → %s\n", ts, task, from, to)

	case "session.analysis":
		status, _ := evt["status"].(string)
		task, _ := evt["task"].(string)
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ◎ ANALYZE %-20s task=%s\n", ts, status, task)

	case "state.drift":
		driftType, _ := evt["drift_type"].(string)
		detail, _ := evt["detail"].(string)
		autoFixed, _ := evt["auto_fixed"].(bool)
		fix := "FLAGGED"
		if autoFixed {
			fix = "AUTO-FIXED"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ⚠ DRIFT   %s: %s [%s]\n", ts, driftType, detail, fix)

	default:
		fmt.Fprintf(cmd.OutOrStdout(), "%s  ? %s\n", ts, event)
	}
}

// ─── Issue Detection ────────────────────────────────────────────────────────

type Issue struct {
	Severity   string `json:"severity"` // error, warning, info
	Code       string `json:"code"`
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

func detectIssues(events []map[string]interface{}) []Issue {
	var issues []Issue

	// LOOP-001: Repeated node visits without progress
	nodeVisits := make(map[string]int)
	for _, e := range events {
		event, _ := e["event"].(string)
		if event == "flow.node_enter" {
			node, _ := e["node"].(string)
			nodeVisits[node]++
		}
	}
	for node, count := range nodeVisits {
		if count > 10 {
			issues = append(issues, Issue{
				Severity:   "warning",
				Code:       "LOOP-001",
				Message:    fmt.Sprintf("Node %s visited %d times — possible loop", node, count),
				Suggestion: "Check edge definitions for unintended cycles",
			})
		}
	}

	// STUCK-001: Node entered but never exited
	nodesEntered := make(map[string]int)
	nodesExited := make(map[string]int)
	nodesLastTS := make(map[string]string)
	for _, e := range events {
		event, _ := e["event"].(string)
		ts, _ := e["ts"].(string)
		node, _ := e["node"].(string)
		switch event {
		case "flow.node_enter":
			nodesEntered[node]++
			nodesLastTS[node] = ts
		case "flow.node_exit":
			nodesExited[node]++
		}
	}
	for node := range nodesEntered {
		if nodesExited[node] < nodesEntered[node] {
			issues = append(issues, Issue{
				Severity:   "warning",
				Code:       "STUCK-001",
				Message:    fmt.Sprintf("Node %s was entered but never exited (last at %s)", node, nodesLastTS[node]),
				Suggestion: "Verify the session completed or manually advance",
			})
		}
	}

	// DRIFT-001: State drift events
	driftCount := 0
	for _, e := range events {
		event, _ := e["event"].(string)
		if event == "state.drift" {
			driftCount++
		}
	}
	if driftCount > 0 {
		issues = append(issues, Issue{
			Severity: "warning",
			Code:     "DRIFT-001",
			Message:  fmt.Sprintf("%d state drift events detected", driftCount),
		})
	}

	// DRIFT-002: Drift not auto-fixed
	unfixedDrifts := 0
	for _, e := range events {
		event, _ := e["event"].(string)
		if event == "state.drift" {
			autoFixed, _ := e["auto_fixed"].(bool)
			if !autoFixed {
				unfixedDrifts++
			}
		}
	}
	if unfixedDrifts > 0 {
		issues = append(issues, Issue{
			Severity:   "error",
			Code:       "DRIFT-002",
			Message:    fmt.Sprintf("%d state drift events not auto-fixed", unfixedDrifts),
			Suggestion: "Manual intervention required — check session state",
		})
	}

	// GATE-001: All gate checks failed
	gateResults := make(map[string][]bool)
	for _, e := range events {
		event, _ := e["event"].(string)
		if event == "flow.gate_check" {
			cond, _ := e["condition"].(string)
			passed, _ := e["passed"].(bool)
			gateResults[cond] = append(gateResults[cond], passed)
		}
	}
	for cond, results := range gateResults {
		allFailed := true
		for _, r := range results {
			if r {
				allFailed = false
				break
			}
		}
		if allFailed && len(results) > 0 {
			issues = append(issues, Issue{
				Severity:   "error",
				Code:       "GATE-001",
				Message:    fmt.Sprintf("Gate condition '%s' failed in all %d attempts", cond, len(results)),
				Suggestion: "Review condition or provide manual input via 'flow gate pass'",
			})
		}
	}

	// SUBA-001: Subagent spawned but no task created after
	var lastSubagentTS string
	var lastSubagentRole string
	for _, e := range events {
		event, _ := e["event"].(string)
		switch event {
		case "ai.subagent_spawn":
			lastSubagentTS, _ = e["ts"].(string)
			lastSubagentRole, _ = e["role"].(string)
		case "task.created", "task.status", "flow.node_enter":
			lastSubagentTS = ""
		}
	}
	if lastSubagentTS != "" {
		issues = append(issues, Issue{
			Severity: "warning",
			Code:     "SUBA-001",
			Message:  fmt.Sprintf("Subagent '%s' spawned at %s but no follow-up event found", lastSubagentRole, lastSubagentTS),
		})
	}

	// Sort by severity: error > warning > info
	sort.Slice(issues, func(i, j int) bool {
		order := map[string]int{"error": 0, "warning": 1, "info": 2}
		return order[issues[i].Severity] < order[issues[j].Severity]
	})

	return issues
}
