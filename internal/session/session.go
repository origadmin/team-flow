package session

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/eventlog"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "session",
	Short: "Manage conversation sessions (one dir per conversation)",
	Long: `Each conversation gets its own directory under .team/sessions/:

  .team/sessions/2026-05-29-2230-a1b2c3/
    events.jsonl  → event timeline (this session only)
    context.md    → structured summary for AI recovery

Subcommands:
  start     Start a new session (creates directory, writes session.start)
  analysis  Record an analysis round (writes session.analysis)
  last      Show last session context for AI recovery
  list      List recent sessions
  history   Show events from a specific session`,
}

func init() {
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(analysisCmd)
	Cmd.AddCommand(lastCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(historyCmd)
}

// ─── start ───────────────────────────────────────────────────────────────────

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new conversation session",
	Long: `Create a new session directory and record the user's input.

  flow session start --input "Fix login 500 error"

Creates .team/sessions/YYYY-MM-DD-HHMM-XXXXXX/ and writes:
  events.jsonl → session.start
  context.md  → initial empty context`,
	RunE: runStart,
}

var startInput string
var startTopic string

func init() {
	startCmd.Flags().StringVar(&startInput, "input", "", "User's original input (required)")
	startCmd.Flags().StringVar(&startTopic, "topic", "", "Topic identifier (optional)")
	_ = startCmd.MarkFlagRequired("input")
}

func runStart(cmd *cobra.Command, args []string) error {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return err
	}

	lgr, err := eventlog.NewLogger(projectRoot)
	if err != nil {
		return err
	}

	round := lgr.LastSessionRound() + 1

	sessionName, err := lgr.CreateSession(round, startTopic, startInput)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	fmt.Printf("✓ Session %s started (round %d)\n", sessionName, round)
	fmt.Printf("  Directory: %s\n", filepath.Join(lgr.SessionsDir(), sessionName))
	return nil
}

// ─── analysis ────────────────────────────────────────────────────────────────

var analysisCmd = &cobra.Command{
	Use:   "analysis",
	Short: "Record an analysis round (call MULTIPLE times per session)",
	Long: `Write a session.analysis event to the current session.

Call this each time you finish an analysis round:
  Round 1: status=clarifying
  Round 2: status=clarifying
  ...
  Round N: status=task_created/redirect/no_task`,
	RunE: runAnalysis,
}

var analysisSession string
var analysisRound int
var analysisStatus string
var analysisTaskType string
var analysisTask string
var analysisTopic string

func init() {
	analysisCmd.Flags().StringVar(&analysisSession, "session", "", "Session directory name (omit = auto-detect latest)")
	analysisCmd.Flags().IntVar(&analysisRound, "round", 0, "Round number (omit = auto)")
	analysisCmd.Flags().StringVar(&analysisStatus, "status", "", "Analysis status: clarifying, task_created, no_task, redirect (required)")
	analysisCmd.Flags().StringVar(&analysisTaskType, "task-type", "", "Task type")
	analysisCmd.Flags().StringVar(&analysisTask, "task", "", "Created task ID")
	analysisCmd.Flags().StringVar(&analysisTopic, "topic", "", "Topic identifier")
	_ = analysisCmd.MarkFlagRequired("status")
}

func runAnalysis(cmd *cobra.Command, args []string) error {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return err
	}

	lgr, err := eventlog.NewLogger(projectRoot)
	if err != nil {
		return err
	}

	if analysisSession == "" {
		analysisSession, err = lgr.LastSessionName()
		if err != nil || analysisSession == "" {
			return fmt.Errorf("no active session found (run 'flow session start' first)")
		}
	}

	if analysisRound <= 0 {
		// Try to find the round from the session's events
		events, _ := lgr.ReadEvents(analysisSession)
		for _, evt := range events {
			if evt["event"] == eventlog.EventSessionStart {
				if r, ok := evt["round"].(float64); ok {
					analysisRound = int(r)
					break
				}
			}
		}
	}
	if analysisRound <= 0 {
		analysisRound = lgr.LastSessionRound()
	}
	if analysisRound <= 0 {
		return fmt.Errorf("--round is required (could not auto-detect)")
	}

	if err := lgr.RecordAnalysis(analysisSession, analysisRound, analysisStatus, analysisTaskType, analysisTask, analysisTopic); err != nil {
		return fmt.Errorf("write analysis: %w", err)
	}

	fmt.Printf("✓ Analysis recorded (session=%s, round=%d, status=%s", analysisSession, analysisRound, analysisStatus)
	if analysisTask != "" {
		fmt.Printf(", task=%s", analysisTask)
	}
	fmt.Println(")")
	return nil
}

// ─── last ───────────────────────────────────────────────────────────────────

var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "Show last session context for AI recovery",
	Long: `Read the most recent session's context.md and events for AI to recover context.

Use --for-ai to get structured markdown output for system prompt injection.`,
	RunE: runLast,
}

var lastForAI bool

func init() {
	lastCmd.Flags().BoolVar(&lastForAI, "for-ai", false, "Output markdown for AI system prompt injection")
}

func runLast(cmd *cobra.Command, args []string) error {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return err
	}

	lgr, err := eventlog.NewLogger(projectRoot)
	if err != nil {
		return err
	}

	name, err := lgr.LastSessionName()
	if err != nil || name == "" {
		if lastForAI {
			fmt.Println("[No previous session]")
		} else {
			fmt.Println("No previous session found.")
		}
		return nil
	}

	if lastForAI {
		printLastForAI(lgr, name)
	} else {
		printLastHuman(lgr, name)
	}

	return nil
}

func printLastForAI(lgr *eventlog.Logger, sessionName string) {
	fmt.Printf("## Context Recovery: %s\n\n", sessionName)

	// Read context.md first (fast path)
	ctx, err := lgr.ReadContext(sessionName)
	if err == nil && ctx != "" {
		fmt.Println(ctx)
		fmt.Println()
	}

	// Fallback: parse events.jsonl if context.md is empty/short
	if ctx == "" || len(ctx) <= 200 {
		events, _ := lgr.ReadEvents(sessionName)
		groups := eventlog.SessionRoundGroups(events)

		var rounds []int
		for r := range groups {
			rounds = append(rounds, r)
		}
		sort.Ints(rounds)

		for _, r := range rounds {
			evts := groups[r]
			var input, lastStatus, lastTask, lastTaskType string
			for _, evt := range evts {
				switch evt["event"] {
				case eventlog.EventSessionStart:
					if s, ok := evt["input"].(string); ok {
						input = s
					}
				case eventlog.EventSessionAnalysis:
					if s, ok := evt["status"].(string); ok {
						lastStatus = s
					}
					if t, ok := evt["task"].(string); ok {
						lastTask = t
					}
					if tt, ok := evt["task_type"].(string); ok {
						lastTaskType = tt
					}
				}
			}
			fmt.Printf("### Round %d\n", r)
			if input != "" {
				fmt.Printf("- Input: %s\n", input)
			}
			fmt.Printf("- Status: %s\n", lastStatus)
			if lastTask != "" {
				fmt.Printf("- Task: %s (type=%s)\n", lastTask, lastTaskType)
			}
			fmt.Println()
		}
	}

	activeTasks, _ := lgr.ActiveTasks()
	if len(activeTasks) > 0 {
		fmt.Println("### Active Tasks Across Sessions")
		for _, t := range activeTasks {
			fmt.Printf("- %s\n", t)
		}
		fmt.Println()
	}
}

func printLastHuman(lgr *eventlog.Logger, sessionName string) {
	fmt.Printf("=== Last Session: %s ===\n\n", sessionName)

	ctx, err := lgr.ReadContext(sessionName)
	if err == nil && ctx != "" {
		fmt.Println(ctx)
		return
	}

	events, _ := lgr.ReadEvents(sessionName)
	groups := eventlog.SessionRoundGroups(events)

	var lastStart, lastAnalysis map[string]interface{}
	var lastRound int
	for r, evts := range groups {
		if r > lastRound {
			lastRound = r
		}
		for _, evt := range evts {
			switch evt["event"] {
			case eventlog.EventSessionStart:
				if r == lastRound {
					lastStart = evt
				}
			case eventlog.EventSessionAnalysis:
				if r == lastRound {
					lastAnalysis = evt
				}
			}
		}
	}

	if lastStart != nil {
		input, _ := lastStart["input"].(string)
		if len(input) > 200 {
			input = input[:200] + "..."
		}
		fmt.Printf("Input:  %s\n", input)
	}
	if lastAnalysis != nil {
		status, _ := lastAnalysis["status"].(string)
		taskType, _ := lastAnalysis["task_type"].(string)
		task, _ := lastAnalysis["task"].(string)
		fmt.Printf("Status: %s (type=%s, task=%s)\n", status, taskType, task)
	}
}

// ─── list ────────────────────────────────────────────────────────────────────

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent sessions",
	RunE:  runList,
}

var listLimit int

func init() {
	listCmd.Flags().IntVar(&listLimit, "limit", 10, "Max sessions to show")
}

func runList(cmd *cobra.Command, args []string) error {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return err
	}

	lgr, err := eventlog.NewLogger(projectRoot)
	if err != nil {
		return err
	}

	names, err := lgr.ListSessions()
	if err != nil || len(names) == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	fmt.Println("=== Sessions ===")
	for i, name := range names {
		if i >= listLimit {
			break
		}
		// Read context for topic
		ctx, _ := lgr.ReadContext(name)
		topic := ""
		if ctx != "" {
			// Extract first line (topic)
			lines := splitLines(ctx, 2)
			if len(lines) > 0 {
				topic = lines[0]
				if len(topic) > 60 {
					topic = topic[:60] + "..."
				}
			}
		} else {
			// Fallback: read session.start input
			events, _ := lgr.ReadEvents(name)
			for _, evt := range events {
				if evt["event"] == eventlog.EventSessionStart {
					if input, ok := evt["input"].(string); ok {
						if len(input) > 60 {
							input = input[:60] + "..."
						}
						topic = input
					}
					break
				}
			}
		}
		if topic == "" {
			topic = "(no input recorded)"
		}
		fmt.Printf("%s  %s\n", name, topic)
	}

	return nil
}

// ─── history ─────────────────────────────────────────────────────────────────

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show events from a specific session",
	Long: `Show events from a session directory.

  flow session history                     # latest session
  flow session history --session NAME      # specific session`,
	RunE: runHistory,
}

var historySession string

func init() {
	historyCmd.Flags().StringVar(&historySession, "session", "", "Session directory name (omit = latest)")
}

func runHistory(cmd *cobra.Command, args []string) error {
	projectRoot, err := resolveProjectRoot()
	if err != nil {
		return err
	}

	lgr, err := eventlog.NewLogger(projectRoot)
	if err != nil {
		return err
	}

	if historySession == "" {
		historySession, err = lgr.LastSessionName()
		if err != nil || historySession == "" {
			fmt.Println("No sessions found.")
			return nil
		}
	}

	events, err := lgr.ReadEvents(historySession)
	if err != nil || len(events) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	fmt.Printf("=== Events for %s ===\n\n", historySession)
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
	fmt.Fprintln(w, "TIME\tTYPE\tDETAILS")

	for _, evt := range events {
		ts, _ := evt["ts"].(string)
		eventType, _ := evt["event"].(string)

		var details string
		switch eventType {
		case eventlog.EventSessionStart:
			round, _ := evt["round"].(float64)
			input, _ := evt["input"].(string)
			if len(input) > 60 {
				input = input[:60] + "..."
			}
			details = fmt.Sprintf("[R%d] %s", int(round), input)
		case eventlog.EventSessionAnalysis:
			status, _ := evt["status"].(string)
			task, _ := evt["task"].(string)
			details = fmt.Sprintf("%s task=%s", status, task)
		case eventlog.EventTaskCreated:
			task, _ := evt["task"].(string)
			title, _ := evt["title"].(string)
			details = fmt.Sprintf("%s: %s", task, title)
		case eventlog.EventTaskStatus:
			task, _ := evt["task"].(string)
			from, _ := evt["from"].(string)
			to, _ := evt["to"].(string)
			details = fmt.Sprintf("%s: %s → %s", task, from, to)
		case eventlog.EventFlowNode:
			node, _ := evt["node"].(string)
			details = node
		case eventlog.EventFlowEnded:
			status, _ := evt["status"].(string)
			details = fmt.Sprintf("status=%s", status)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", ts, eventType, details)
	}
	w.Flush()

	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func splitLines(s string, n int) []string {
	lines := make([]string, 0, n)
	for i, line := range strings.Split(s, "\n") {
		if i >= n {
			break
		}
		lines = append(lines, line)
	}
	return lines
}

func resolveProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root := config.ResolveProjectRoot(cwd)
	if root == "" {
		root = cwd
	}
	return root, nil
}

func EnsureEventLogCLI(dir string) {
	_ = os.MkdirAll(filepath.Join(dir, ".team", "sessions"), 0755)
}
