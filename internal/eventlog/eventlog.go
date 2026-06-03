// Package eventlog provides a per-session event log for tracking conversations.
//
// Storage: single events.mdl file under .team/state/ for all sessions,
// with per-session context.md and state files under .team/sessions/.
//
//	.team/
//	  state/
//	    events.mdl                <- single event log (JSONL inside .mdl, one event per line)
//	  sessions/
//	    2026-05-29-2230-a1b2c3/    <- conversation directory
//	      context.md                <- structured summary (AI recovery)
//	      active_flow               <- persisted flow context
//	    2026-05-29-2245-d4e5f6/
//	      context.md
//	      active_flow
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

## Latest Analysis

_Awaiting first analysis..._

## Round History

_No rounds yet_
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
func (l *Logger) UpdateContextSnapshot(sessionName, taskID, flowName, nodeName, phase, statusLine, userInput string) error {
	// Read existing context.md to preserve Round/Topic/Started/Round History
	existing, _ := l.ReadContext(sessionName)

	var b strings.Builder

	// Preserve header section (everything before "## Current State" or "## Latest Analysis")
	// If no existing content, create a minimal header
	if existing != "" {
		// Find the first "## Current State" or "## Latest Analysis" marker
		cutIdx := -1
		for _, marker := range []string{"## Current State", "## Latest Analysis"} {
			idx := strings.Index(existing, marker)
			if idx >= 0 && (cutIdx < 0 || idx < cutIdx) {
				cutIdx = idx
			}
		}
		if cutIdx > 0 {
			b.WriteString(existing[:cutIdx])
		} else {
			// No markers found — keep the header (everything before last _Last updated_)
			lastUpdatedIdx := strings.LastIndex(existing, "_Last updated:")
			if lastUpdatedIdx > 0 {
				// Walk back to start of line
				lineStart := strings.LastIndex(existing[:lastUpdatedIdx], "\n")
				if lineStart > 0 {
					b.WriteString(existing[:lineStart+1])
				} else {
					b.WriteString(existing)
				}
			} else {
				b.WriteString(existing)
			}
		}
	} else {
		// No existing content — write a fresh header
		b.WriteString(fmt.Sprintf("# Session: %s\n\n", sessionName))
		b.WriteString("- **Status**: active\n")
		if flowName != "" {
			b.WriteString(fmt.Sprintf("- **Flow**: %s\n", flowName))
		}
		b.WriteString("\n")
	}

	// Write current state section
	b.WriteString("## Current State\n\n")
	if flowName != "" {
		b.WriteString(fmt.Sprintf("- **Flow**: %s\n", flowName))
	}
	if nodeName != "" {
		b.WriteString(fmt.Sprintf("- **Current Node**: %s\n", nodeName))
	}
	if phase != "" {
		b.WriteString(fmt.Sprintf("- **Phase**: %s\n", phase))
	}
	if taskID != "" {
		b.WriteString(fmt.Sprintf("- **Task**: %s\n", taskID))
	}
	if statusLine != "" {
		b.WriteString(fmt.Sprintf("- **StatusLine**: `%s`\n", statusLine))
	}
	if userInput != "" {
		// Truncate long input for context.md readability
		displayInput := userInput
		if len([]rune(displayInput)) > 200 {
			displayInput = string([]rune(displayInput)[:200]) + "..."
		}
		b.WriteString(fmt.Sprintf("- **User Input**: %s\n", displayInput))
	} else if existing != "" {
		// Preserve existing User Input if no new input provided (auto-advance case)
		if idx := strings.Index(existing, "- **User Input**:"); idx >= 0 {
			endIdx := strings.Index(existing[idx:], "\n")
			if endIdx >= 0 {
				b.WriteString(existing[idx : idx+endIdx+1])
			} else {
				b.WriteString(existing[idx:])
			}
		}
	}
	b.WriteString(fmt.Sprintf("\n_Last updated: %s_\n", now()))

	return l.WriteContext(sessionName, b.String())
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