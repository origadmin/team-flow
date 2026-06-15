// Package eventlog provides a per-session event log for tracking conversations.
//
// Storage: single events.mdl file under .team/state/ for all sessions,
// with per-session context.md and state files under .team/sessions/.
//
//	.team/
//	  state/
//	    events.mdl                <- single event log (JSONL, one event per line)
//	  sessions/
//	    2026-05-29-2230-a1b2c3/    <- conversation directory
//	      context.md                <- structured summary (AI recovery)
//	      trace.jsonl               <- per-session operational tracing
//
// Directory naming: YYYY-MM-DD-HHMM-<6 random hex chars>.
// Sorting by name = sorting by creation time.
//
// Concurrency: per-session lock file (mkdir atomic). Different sessions can
// be written concurrently without contention. Same-session writes are
// serialized.
//
// Context recovery: read the latest session's context.md — no scanning, no
// event type filtering needed.
package eventlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/origadmin/team-flow/internal/idgen"
	"github.com/origadmin/team-flow/internal/trace"
)

const (
	sessionsDirName = "sessions"
	stateDirName    = "state"
	eventsFileName  = "events.mdl"
)

// Event type constants
const (
	EventSessionStart    = "session.start"
	EventSessionAnalysis = "session.analysis"
	EventTaskCreated     = "task.created"
	EventTaskStatus      = "task.status"
	EventFlowStarted     = "flow.started"
	EventFlowNode        = "flow.node"
	EventFlowEnded       = "flow.ended"
	EventError           = "error"
)

const (
	StatusClarifying  = "clarifying"
	StatusTaskCreated = "task_created"
	StatusNoTask      = "no_task"
	StatusRedirect    = "redirect"
)

// ─── event types ─────────────────────────────────────────────────────────────

type SessionStart struct {
	Event   string `json:"event"`
	TS      string `json:"ts"`
	Round   int    `json:"round"`
	Topic   string `json:"topic,omitempty"`
	Input   string `json:"input"`
	Session string `json:"session,omitempty"`
}

type SessionAnalysis struct {
	Event    string `json:"event"`
	TS       string `json:"ts"`
	Round    int    `json:"round"`
	Status   string `json:"status"`
	TaskType string `json:"task_type,omitempty"`
	Task     string `json:"task,omitempty"`
	Topic    string `json:"topic,omitempty"`
	Session  string `json:"session,omitempty"`
}

type TaskCreated struct {
	Event   string `json:"event"`
	TS      string `json:"ts"`
	Task    string `json:"task"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Session string `json:"session,omitempty"`
}

type TaskStatus struct {
	Event   string `json:"event"`
	TS      string `json:"ts"`
	Task    string `json:"task"`
	From    string `json:"from"`
	To      string `json:"to"`
	Session string `json:"session,omitempty"`
}

type FlowStarted struct {
	Event   string `json:"event"`
	TS      string `json:"ts"`
	Task    string `json:"task,omitempty"`
	Flow    string `json:"flow"`
	Node    string `json:"node"`
	Session string `json:"session,omitempty"`
}

type FlowNode struct {
	Event    string `json:"event"`
	TS       string `json:"ts"`
	Task     string `json:"task,omitempty"`
	Node     string `json:"node"`
	NodeName string `json:"node_name,omitempty"`
	Phase    string `json:"phase,omitempty"`
	Flow     string `json:"flow,omitempty"`
	Session  string `json:"session,omitempty"`
}

type FlowEnded struct {
	Event   string `json:"event"`
	TS      string `json:"ts"`
	Task    string `json:"task,omitempty"`
	Flow    string `json:"flow"`
	Status  string `json:"status"`
	Session string `json:"session,omitempty"`
}

type ErrorEvent struct {
	Event   string `json:"event"`
	TS      string `json:"ts"`
	Task    string `json:"task,omitempty"`
	Error   string `json:"error"`
	Session string `json:"session,omitempty"`
}

// ─── Logger ──────────────────────────────────────────────────────────────────

// Logger manages event logs under .team/state/ and .team/sessions/.
type Logger struct {
	stateDir    string // .team/state/
	sessionsDir string // .team/sessions/
	Trace       *trace.Logger
}

// NewLogger creates a Logger for the given project root.
func NewLogger(projectRoot string) (*Logger, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("eventlog: project root is empty")
	}
	teamDir := filepath.Join(projectRoot, ".team")
	stateDir := filepath.Join(teamDir, stateDirName)
	sessionsDir := filepath.Join(teamDir, sessionsDirName)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("eventlog: create state dir: %w", err)
	}
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return nil, fmt.Errorf("eventlog: create sessions dir: %w", err)
	}
	return &Logger{
		stateDir:    stateDir,
		sessionsDir: sessionsDir,
		Trace:       trace.NewLogger(teamDir),
	}, nil
}

// SessionsDir returns the sessions root directory.
func (l *Logger) SessionsDir() string {
	return l.sessionsDir
}

// ─── session lifecycle ──────────────────────────────────────────────────────

// CreateSession creates a new session directory and writes session.start event
// to the global events.mdl.
// Returns the session directory name (e.g. "2026-05-29-2230-a1b2c3").
// v4#11: input is truncated to 50 chars for the event log; full text is in context.md.
func (l *Logger) CreateSession(round int, topic, input string) (string, error) {
	name := sessionDirName()
	dir := filepath.Join(l.sessionsDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("eventlog: create session dir: %w", err)
	}

	ts := now()

	// v4#11: truncate input to short keyword summary for events.mdl
	shortInput := input
	if len([]rune(shortInput)) > 50 {
		shortInput = string([]rune(shortInput)[:50]) + "..."
	}

	if err := l.writeEvent(SessionStart{
		Event:   EventSessionStart,
		TS:      ts,
		Round:   round,
		Topic:   topic,
		Input:   shortInput,
		Session: name,
	}); err != nil {
		return "", fmt.Errorf("eventlog: write session.start: %w", err)
	}

	ctxPath := filepath.Join(dir, "context.md")
	ctx := fmt.Sprintf(`# Session: %s

- **Round**: R%d
- **Topic**: %s
- **Started**: %s
- **Status**: active

## Current State

_Awaiting first node..._
`, name, round, topic, ts)
	if err := os.WriteFile(ctxPath, []byte(ctx), 0644); err != nil {
		return "", fmt.Errorf("eventlog: write context.md: %w", err)
	}

	return name, nil
}

// RecordAnalysis writes a session.analysis event to the global events.mdl
// and syncs the detailed content to the session's context.md.
func (l *Logger) RecordAnalysis(sessionName string, round int, status, taskType, task, topic string) error {
	if l.Trace != nil {
		l.Trace.Analysis(sessionName, round, status, taskType, task, topic)
	}
	if err := l.writeEvent(SessionAnalysis{
		Event:    EventSessionAnalysis,
		TS:       now(),
		Round:    round,
		Status:   status,
		TaskType: taskType,
		Task:     task,
		Topic:    topic,
		Session:  sessionName,
	}); err != nil {
		return err
	}
	content, err := l.buildAnalysisContext(sessionName, round, status, taskType, task, topic)
	if err != nil {
		return fmt.Errorf("eventlog: build context: %w", err)
	}
	if err := l.WriteContext(sessionName, content); err != nil {
		return fmt.Errorf("eventlog: write context: %w", err)
	}
	return nil
}

// RecordNodeAnalysis records AI analysis for a specific node (simplified API).
func (l *Logger) RecordNodeAnalysis(sessionName, taskID, analysis, flowName, nodeID string) error {
	if l.Trace != nil {
		l.Trace.NodeAnalysis(sessionName, taskID, flowName, nodeID, analysis)
	}
	return l.writeEvent(map[string]interface{}{
		"event":    "node.analysis",
		"ts":       now(),
		"session":  sessionName,
		"task":     taskID,
		"flow":     flowName,
		"node":     nodeID,
		"analysis": analysis,
	})
}

// RecordNodeConclusion records AI conclusion for a specific node.
func (l *Logger) RecordNodeConclusion(sessionName, taskID, conclusion, flowName, nodeID string) error {
	if l.Trace != nil {
		l.Trace.NodeAnalysis(sessionName, taskID, flowName, nodeID, "CONCLUSION: "+conclusion)
	}
	return l.writeEvent(map[string]interface{}{
		"event":      "node.conclusion",
		"ts":         now(),
		"session":    sessionName,
		"task":       taskID,
		"flow":       flowName,
		"node":       nodeID,
		"conclusion": conclusion,
	})
}

// ─── global events (task, flow, error) ──────────────────────────────────────

// TaskCreated writes a task.created event to the global events.mdl.
func (l *Logger) TaskCreated(sessionName, taskID, taskType, title string) error {
	if l.Trace != nil {
		l.Trace.TaskCreated(sessionName, taskID, taskType, title)
	}
	return l.writeEvent(TaskCreated{
		Event:   EventTaskCreated,
		TS:      now(),
		Task:    taskID,
		Type:    taskType,
		Title:   title,
		Session: sessionName,
	})
}

// TaskStatusChange writes a task.status event to the global events.mdl.
func (l *Logger) TaskStatusChange(sessionName, taskID, from, to string) error {
	if l.Trace != nil {
		l.Trace.TaskStatus(sessionName, taskID, from, to)
	}
	return l.writeEvent(TaskStatus{
		Event:   EventTaskStatus,
		TS:      now(),
		Task:    taskID,
		From:    from,
		To:      to,
		Session: sessionName,
	})
}

// FlowStarted writes a flow.started event to the global events.mdl.
func (l *Logger) FlowStarted(sessionName, taskID, flowName, nodeID string) error {
	if l.Trace != nil {
		l.Trace.NodeEnter(sessionName, taskID, nodeID, "Flow Started", flowName)
	}
	return l.writeEvent(FlowStarted{
		Event:   EventFlowStarted,
		TS:      now(),
		Task:    taskID,
		Flow:    flowName,
		Node:    nodeID,
		Session: sessionName,
	})
}

// FlowNodeAdvance writes a flow.node event to the global events.mdl.
func (l *Logger) FlowNodeAdvance(sessionName, taskID, nodeID, nodeName, phase, flowName string) error {
	if l.Trace != nil {
		l.Trace.NodeExit(sessionName, taskID, nodeID, nodeName, flowName, 0, "", "")
	}
	return l.writeEvent(FlowNode{
		Event:    EventFlowNode,
		TS:       now(),
		Task:     taskID,
		Node:     nodeID,
		NodeName: nodeName,
		Phase:    phase,
		Flow:     flowName,
		Session:  sessionName,
	})
}

// FlowEnded writes a flow.ended event to the global events.mdl.
func (l *Logger) FlowEnded(sessionName, taskID, flowName, status string) error {
	return l.writeEvent(FlowEnded{
		Event:   EventFlowEnded,
		TS:      now(),
		Task:    taskID,
		Flow:    flowName,
		Status:  status,
		Session: sessionName,
	})
}

// Error writes an error event to the global events.mdl.
func (l *Logger) Error(sessionName, taskID, errMsg string) error {
	return l.writeEvent(ErrorEvent{
		Event:   EventError,
		TS:      now(),
		Task:    taskID,
		Error:   errMsg,
		Session: sessionName,
	})
}

// MarkSessionCompleted updates the session status in context.md to "completed".
func (l *Logger) MarkSessionCompleted(sessionName string) error {
	existing, err := l.ReadContext(sessionName)
	if err != nil {
		return err
	}
	if existing == "" {
		return nil
	}
	updated := strings.Replace(existing, "- **Status**: active", "- **Status**: completed", 1)
	return l.WriteContext(sessionName, updated)
}

// ─── context.md ──────────────────────────────────────────────────────────────

// WriteContext writes or updates the session's context.md with a structured
// AI-recovery summary.
func (l *Logger) WriteContext(sessionName string, content string) error {
	dir := filepath.Join(l.sessionsDir, sessionName)
	return l.writeLocked(dir, func() error {
		return os.WriteFile(filepath.Join(dir, "context.md"), []byte(content), 0644)
	})
}

// UpdateContextSnapshot writes a lightweight context.md snapshot for v4#23:
// auto-updates on every node advance so AI can recover session context.
// Analysis/conclusion are accumulative: each new round is prepended to the
// existing Current State (newest first), so the full task analysis trail is
// visible at the top level.
func (l *Logger) UpdateContextSnapshot(sessionName, taskID, flowName, nodeName, phase, statusLine, userInput, analysis, conclusion string) error {
	existing, _ := l.ReadContext(sessionName)

	// Extract old Current State before any modification
	oldState := ""
	if existing != "" {
		oldState = extractCurrentState(existing)
	}

	// Increment round counter and capture topic from first user input
	round, topic := extractHeader(existing)
	if topic == "" && userInput != "" {
		topic = userInput
		if len([]rune(topic)) > 80 {
			topic = string([]rune(topic)[:80]) + "..."
		}
	}
	round++

	var history strings.Builder
	newState := buildRound(round, flowName, nodeName, phase, statusLine, userInput, analysis, conclusion)

	// Move old Current State into Round History (full snapshot)
	hasOldContent := oldState != "" && !strings.Contains(oldState, "_Awaiting first node..._")
	if hasOldContent {
		history.WriteString("\n## Round History\n\n")
		history.WriteString(fmt.Sprintf("### %s\n\n", roundTimestamp(existing)))
		history.WriteString(oldState)
		history.WriteString("\n")
	}

	// Accumulate: prepend new round's content before old Current State
	// so the full analysis trail is visible (newest first)
	currentState := newState
	if hasOldContent {
		currentState = newState + "\n\n---\n\n" + oldState
	}

	// Build fresh header (always overwrite placeholders)
	header := fmt.Sprintf(`# Session: %s

- **Round**: R%d
- **Topic**: %s
- **Started**: %s
- **Status**: active
`, sessionName, round, topic, startedTime(existing))

	body := header + "\n## Current State\n\n" + currentState + history.String()
	return l.WriteContext(sessionName, body)
}

// extractHeader parses the existing header for round and topic.
func extractHeader(existing string) (round int, topic string) {
	if existing == "" {
		return 0, ""
	}
	for _, line := range strings.Split(existing, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- **Round**: R"):
			n, err := strconv.Atoi(strings.TrimPrefix(line, "- **Round**: R"))
			if err == nil {
				round = n
			}
		case strings.HasPrefix(line, "- **Topic**: "):
			topic = strings.TrimPrefix(line, "- **Topic**: ")
		}
	}
	return round, topic
}

// extractCurrentState pulls the "## Current State" block out of existing content.
func extractCurrentState(existing string) string {
	const marker = "## Current State"
	idx := strings.Index(existing, marker)
	if idx < 0 {
		return ""
	}
	rest := existing[idx+len(marker):]
	// Trim until next top-level "## " heading or end
	if endIdx := strings.Index(rest, "\n## "); endIdx >= 0 {
		return strings.TrimSpace(rest[:endIdx+1])
	}
	return strings.TrimSpace(rest)
}

func roundTimestamp(existing string) string {
	for _, line := range strings.Split(existing, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "_Last updated:") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "_Last updated:"))
		}
	}
	return time.Now().UTC().Format(time.RFC3339)
}

func startedTime(existing string) string {
	for _, line := range strings.Split(existing, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- **Started**: ") {
			return strings.TrimPrefix(line, "- **Started**: ")
		}
	}
	return time.Now().UTC().Format(time.RFC3339)
}

func buildRound(round int, flowName, nodeName, phase, statusLine, userInput, analysis, conclusion string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("- **Node**: %s", nodeName))
	if phase != "" {
		b.WriteString(fmt.Sprintf(" (%s)", phase))
	}
	b.WriteString("\n")
	if flowName != "" {
		b.WriteString(fmt.Sprintf("- **Flow**: %s\n", flowName))
	}
	if statusLine != "" {
		b.WriteString(fmt.Sprintf("- **StatusLine**: `%s`\n", statusLine))
	}
	if userInput != "" {
		input := userInput
		if len([]rune(input)) > 200 {
			input = string([]rune(input)[:200]) + "..."
		}
		b.WriteString(fmt.Sprintf("- **User Input**: %s\n", input))
	}

	b.WriteString("\n### Analysis\n\n")
	if analysis != "" {
		b.WriteString(analysis)
		b.WriteString("\n")
	} else {
		b.WriteString("_No analysis recorded this round._\n")
	}

	b.WriteString("\n### Conclusion\n\n")
	if conclusion != "" {
		b.WriteString(conclusion)
		b.WriteString("\n")
	} else {
		b.WriteString("_No conclusion recorded this round._\n")
	}

	b.WriteString(fmt.Sprintf("\n_Last updated: %s_", time.Now().UTC().Format(time.RFC3339)))
	return b.String()
}

// ReadContext returns the context.md for the given session, or empty string.
func (l *Logger) ReadContext(sessionName string) (string, error) {
	path := filepath.Join(l.sessionsDir, sessionName, "context.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// buildAnalysisContext constructs a human-readable context.md from session events.
func (l *Logger) buildAnalysisContext(sessionName string, round int, status, taskType, task, topic string) (string, error) {
	events, err := l.ReadEvents(sessionName)
	if err != nil {
		return "", err
	}

	var started string
	for _, evt := range events {
		if evt["event"] == EventSessionStart {
			if s, ok := evt["ts"].(string); ok {
				started = s
			}
			break
		}
	}
	if started == "" {
		started = now()
	}

	var analysisRounds []map[string]interface{}
	for _, evt := range events {
		if evt["event"] == EventSessionAnalysis {
			analysisRounds = append(analysisRounds, evt)
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Session: %s\n\n", sessionName))
	b.WriteString(fmt.Sprintf("- **Round**: R%d\n", round))
	b.WriteString(fmt.Sprintf("- **Topic**: %s\n", topic))
	b.WriteString(fmt.Sprintf("- **Started**: %s\n", started))
	b.WriteString(fmt.Sprintf("- **Status**: %s\n\n", status))

	b.WriteString("## Latest Analysis\n\n")
	b.WriteString(fmt.Sprintf("- **Status**: %s\n", status))
	b.WriteString(fmt.Sprintf("- **Task Type**: %s\n", taskType))
	b.WriteString(fmt.Sprintf("- **Task**: %s\n\n", task))

	b.WriteString("## Round History\n\n")
	for _, r := range analysisRounds {
		rn, _ := r["round"].(float64)
		rs, _ := r["status"].(string)
		rt, _ := r["ts"].(string)
		rtt, _ := r["task_type"].(string)
		rta, _ := r["task"].(string)
		rto, _ := r["topic"].(string)

		b.WriteString(fmt.Sprintf("### R%d — %s — %s\n\n", int(rn), rs, rt))
		b.WriteString(fmt.Sprintf("- **Task**: %s (%s)\n", rta, rtt))
		b.WriteString(fmt.Sprintf("- **Topic**: %s\n\n", rto))
	}

	return b.String(), nil
}

// ─── session listing ─────────────────────────────────────────────────────────

// ListSessions returns session directory names sorted by name descending
// (newest first). Name format YYYY-MM-DD-HHMM-XXXXXX naturally sorts by time.
func (l *Logger) ListSessions() ([]string, error) {
	entries, err := os.ReadDir(l.sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names, nil
}

// LastSessionName returns the name of the most recent session directory.
func (l *Logger) LastSessionName() (string, error) {
	names, err := l.ListSessions()
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", nil
	}
	return names[0], nil
}

// CurrentNode reads events.mdl and returns the node ID of the most recent
// flow.node event for the given session. Returns empty string if no node
// has been advanced yet.
func (l *Logger) CurrentNode(sessionName string) string {
	events, err := l.ReadEvents(sessionName)
	if err != nil || len(events) == 0 {
		return ""
	}
	for i := len(events) - 1; i >= 0; i-- {
		evt := events[i]
		if ev, ok := evt["event"].(string); ok && (ev == EventFlowNode || ev == EventFlowStarted) {
			if n, ok := evt["node"].(string); ok && n != "" {
				return n
			}
		}
	}
	return ""
}

// ReadEvents reads all events from the global events.mdl file.
// Returns all events in chronological order.
func (l *Logger) ReadEvents(sessionName string) ([]map[string]interface{}, error) {
	path := filepath.Join(l.stateDir, eventsFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var events []map[string]interface{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var evt map[string]interface{}
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			continue
		}
		if sessionName != "" {
			if s, _ := evt["session"].(string); s != "" && s != sessionName {
				continue
			}
		}
		events = append(events, evt)
	}
	return events, nil
}

// LasTSessionRound returns the round number from the last session.start event
// in the latest session. Returns 0 if no sessions exist.
func (l *Logger) LastSessionRound() int {
	name, err := l.LastSessionName()
	if err != nil || name == "" {
		return 0
	}
	events, err := l.ReadEvents(name)
	if err != nil {
		return 0
	}
	for _, evt := range events {
		if evt["event"] == EventSessionStart {
			if r, ok := evt["round"].(float64); ok {
				return int(r)
			}
		}
	}
	return 0
}

// LastTask returns the task ID from the most recent task.created event
// across all sessions.
func (l *Logger) LastTask() string {
	events, err := l.ReadEvents("")
	if err != nil {
		return ""
	}
	for i := len(events) - 1; i >= 0; i-- {
		if events[i]["event"] == EventTaskCreated {
			if id, ok := events[i]["task"].(string); ok {
				return id
			}
		}
	}
	return ""
}

// HasActiveTasks checks if any task is still open/in-progress across all sessions.
func (l *Logger) HasActiveTasks() (bool, error) {
	active, err := l.ActiveTasks()
	if err != nil {
		return false, err
	}
	return len(active) > 0, nil
}

// ActiveTasks returns the list of currently active task IDs across all sessions.
func (l *Logger) ActiveTasks() ([]string, error) {
	events, err := l.ReadEvents("")
	if err != nil {
		return nil, err
	}

	tasks := make(map[string]string)
	for _, evt := range events {
		id, _ := evt["task"].(string)
		if id == "" {
			continue
		}
		switch evt["event"] {
		case EventTaskCreated:
			if _, exists := tasks[id]; !exists {
				tasks[id] = "open"
			}
		case EventTaskStatus:
			if to, ok := evt["to"].(string); ok {
				tasks[id] = to
			}
		}
	}

	var active []string
	for id, status := range tasks {
		if status == "open" || status == "in_progress" {
			active = append(active, id)
		}
	}
	return active, nil
}

// SessionRoundGroups groups events by round number. Only useful within
// a single session's events.
func SessionRoundGroups(events []map[string]interface{}) map[int][]map[string]interface{} {
	groups := make(map[int][]map[string]interface{})
	for _, evt := range events {
		if r, ok := evt["round"].(float64); ok {
			groups[int(r)] = append(groups[int(r)], evt)
		}
	}
	return groups
}

// ─── internal helpers ───────────────────────────────────────────────────────

// sessionDirName generates a unique session directory name.
// Format: YYYY-MM-DD-HHMM-<6 random hex chars>
func sessionDirName() string {
	suffix := idgen.RandHex(3)
	return time.Now().UTC().Format("2006-01-02-1504-") + suffix
}

// eventsPath returns the path to the global events.mdl file.
func eventsPath(stateDir string) string {
	return filepath.Join(stateDir, eventsFileName)
}

// writeEvent appends a single JSONL line to the global events.mdl file.
func (l *Logger) writeEvent(payload interface{}) error {
	return l.writeLocked(l.stateDir, func() error {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("eventlog: marshal: %w", err)
		}
		f, err := os.OpenFile(eventsPath(l.stateDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("eventlog: open: %w", err)
		}
		defer f.Close()
		if _, err := f.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("eventlog: write: %w", err)
		}
		return f.Sync()
	})
}

// writeLocked acquires a per-directory lock, runs fn, then releases.
func (l *Logger) writeLocked(dir string, fn func() error) error {
	lockDir := filepath.Join(dir, ".lock")
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := os.Mkdir(lockDir, 0755); err == nil {
			defer os.RemoveAll(lockDir)
			return fn()
		} else if !os.IsExist(err) {
			return fmt.Errorf("eventlog: lock %s: %w", lockDir, err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("eventlog: lock timeout: %s", lockDir)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// ─── round directory (per-round analysis artifacts) ─────────────────────────

// SessionDir returns the full path to a session directory.
func (l *Logger) SessionDir(sessionName string) string {
	return filepath.Join(l.sessionsDir, sessionName)
}

// LatestRoundDir scans the session directory for r{N}/ subdirectories,
// returns the latest one and its round number N.
func (l *Logger) LatestRoundDir(sessionName string) (string, int, error) {
	dir := l.SessionDir(sessionName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", 0, fmt.Errorf("eventlog: read session dir %s: %w", dir, err)
	}
	var maxN int
	var maxDir string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "r") {
			continue
		}
		n, err := strconv.Atoi(name[1:])
		if err != nil {
			continue
		}
		if n > maxN {
			maxN = n
			maxDir = name
		}
	}
	if maxDir == "" {
		return "", 0, fmt.Errorf("eventlog: no round directory found in %s", dir)
	}
	return filepath.Join(dir, maxDir), maxN, nil
}

// NextRoundDir creates the next r{N+1}/ directory under the session dir.
// Returns the full path and the round number.
func (l *Logger) NextRoundDir(sessionName string) (string, int, error) {
	_, prevN, err := l.LatestRoundDir(sessionName)
	if err != nil {
		prevN = 0 // first round: no previous dir found
	}
	nextN := prevN + 1
	dir := filepath.Join(l.SessionDir(sessionName), fmt.Sprintf("r%d", nextN))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, fmt.Errorf("eventlog: create round dir %s: %w", dir, err)
	}
	return dir, nextN, nil
}

// ReadLatestAnalysis reads analysis.md and conclusion.md from the latest round
// directory. Returns analysis content, conclusion content, round number, and error.
func (l *Logger) ReadLatestAnalysis(sessionName string) (analysis, conclusion string, round int, err error) {
	dir, n, err := l.LatestRoundDir(sessionName)
	if err != nil {
		return "", "", 0, err
	}
	analysisBytes, err := os.ReadFile(filepath.Join(dir, "analysis.md"))
	if err != nil {
		return "", "", n, fmt.Errorf("eventlog: read analysis.md from %s: %w", dir, err)
	}
	conclusionBytes, err := os.ReadFile(filepath.Join(dir, "conclusion.md"))
	if err != nil {
		return "", "", n, fmt.Errorf("eventlog: read conclusion.md from %s: %w", dir, err)
	}
	return string(analysisBytes), string(conclusionBytes), n, nil
}