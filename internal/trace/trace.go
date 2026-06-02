// Package trace provides per-session operational tracing for post-mortem
// analysis and issue detection.
//
// Storage: per-session trace.jsonl under .team/sessions/<id>/.
//
// Design: append-only JSONL, same locking as eventlog (per-session lock file).
// Zero perf impact — fire-and-forget writes, errors are silently logged to stderr.
//
// Event types capture the full lifecycle:
//   - flow.node_enter / flow.node_exit — node transitions
//   - flow.edge_traverse — edge traversal with condition result
//   - flow.gate_check — individual gate condition results
//   - ai.tool_call — AI tool invocations (exec, write, read, etc.)
//   - ai.file_write — file creation/modification with checksum
//   - ai.subagent_spawn — sub-agent dispatch
//   - task.created / task.status — task lifecycle
//   - session.analysis — analysis round recorded
//   - state.drift — detected state inconsistency (auto-fixed or flagged)
package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const traceFileName = "trace.jsonl"

// ─── Logger ─────────────────────────────────────────────────────────────────

// Logger writes structured trace events to a per-session trace.jsonl file.
// All writes are append-only, fire-and-forget (errors go to stderr).
type Logger struct {
	sessionsDir string
	mu          sync.Mutex // guards in-flight writes per logger instance
}

// NewLogger creates a trace logger rooted at the given .team directory.
// sessionsDir is typically <project>/.team/sessions.
func NewLogger(teamDir string) *Logger {
	return &Logger{
		sessionsDir: filepath.Join(teamDir, "sessions"),
	}
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

// ─── Public API ─────────────────────────────────────────────────────────────

// NodeEnter records entry into a flow node.
func (l *Logger) NodeEnter(sessionName, taskID, nodeID, nodeName, flowName string) {
	l.append(sessionName, map[string]interface{}{
		"ts":    now(),
		"event": "flow.node_enter",
		"node":  nodeID,
		"name":  nodeName,
		"flow":  flowName,
		"task":  taskID,
	})
}

// NodeExit records exit from a flow node, including deliverables produced.
func (l *Logger) NodeExit(sessionName, taskID, nodeID, nodeName, flowName string, durationMs int64, deliverables, nextNode string) {
	evt := map[string]interface{}{
		"ts":       now(),
		"event":    "flow.node_exit",
		"node":     nodeID,
		"name":     nodeName,
		"flow":     flowName,
		"task":     taskID,
		"duration": durationMs,
	}
	if len(deliverables) > 0 {
		evt["deliverables"] = deliverables
	}
	if nextNode != "" {
		evt["next_node"] = nextNode
	}
	l.append(sessionName, evt)
}

// EdgeTraverse records traversal of an edge, including its type and condition result.
func (l *Logger) EdgeTraverse(sessionName, taskID, edgeID, fromNode, toNode, condExpr, condResult, flowName string) {
	l.append(sessionName, map[string]interface{}{
		"ts":          now(),
		"event":       "flow.edge_traverse",
		"edge":        edgeID,
		"from":        fromNode,
		"to":          toNode,
		"condition":   condExpr,
		"cond_result": condResult,
		"flow":        flowName,
		"task":        taskID,
	})
}

// GateCheck records the result of a single gate condition check.
func (l *Logger) GateCheck(sessionName, taskID, nodeID, nodeName, gateType string, passed bool, message string) {
	l.append(sessionName, map[string]interface{}{
		"ts":      now(),
		"event":   "flow.gate_check",
		"node":    nodeID,
		"name":    nodeName,
		"gate":    gateType,
		"passed":  passed,
		"message": message,
		"task":    taskID,
	})
}

// ToolCall records an AI tool invocation (exec, read, write, browser, etc.).
func (l *Logger) ToolCall(sessionName, tool, command string, exitCode int, durationMs int64) {
	l.append(sessionName, map[string]interface{}{
		"ts":       now(),
		"event":    "ai.tool_call",
		"tool":     tool,
		"command":  command,
		"exit_code": exitCode,
		"duration": durationMs,
	})
}

// FileWrite records a file creation or modification by the AI.
func (l *Logger) FileWrite(sessionName, filePath string, sizeBytes int64, checksum string) {
	l.append(sessionName, map[string]interface{}{
		"ts":        now(),
		"event":     "ai.file_write",
		"path":      filePath,
		"size_bytes": sizeBytes,
		"checksum":  checksum,
	})
}

// SubagentSpawn records a sub-agent dispatch by the AI.
func (l *Logger) SubagentSpawn(sessionName, role, task string) {
	l.append(sessionName, map[string]interface{}{
		"ts":    now(),
		"event": "ai.subagent_spawn",
		"role":  role,
		"task":  task,
	})
}

// TaskCreated records task creation.
func (l *Logger) TaskCreated(sessionName, taskID, taskType, title string) {
	l.append(sessionName, map[string]interface{}{
		"ts":    now(),
		"event": "task.created",
		"task":  taskID,
		"type":  taskType,
		"title": title,
	})
}

// TaskStatus records a task status change.
func (l *Logger) TaskStatus(sessionName, taskID, from, to string) {
	l.append(sessionName, map[string]interface{}{
		"ts":    now(),
		"event": "task.status",
		"task":  taskID,
		"from":  from,
		"to":    to,
	})
}

// Analysis records a session analysis round.
func (l *Logger) Analysis(sessionName string, round int, status, taskType, task, topic string) {
	l.append(sessionName, map[string]interface{}{
		"ts":        now(),
		"event":     "session.analysis",
		"round":     round,
		"status":    status,
		"task_type": taskType,
		"task":      task,
		"topic":     topic,
	})
}

// StateDrift records a detected state inconsistency (auto-fixed or flagged).
func (l *Logger) StateDrift(sessionName, taskID, driftType, detail string, autoFixed bool) {
	l.append(sessionName, map[string]interface{}{
		"ts":         now(),
		"event":      "state.drift",
		"drift_type": driftType,
		"detail":     detail,
		"auto_fixed": autoFixed,
		"task":       taskID,
	})
}

// ─── Internal ───────────────────────────────────────────────────────────────

// append writes a single JSON line to trace.jsonl for the given session.
// Errors are printed to stderr and discarded — tracing is best-effort.
func (l *Logger) append(sessionName string, evt map[string]interface{}) {
	dir := filepath.Join(l.sessionsDir, sessionName)
	path := filepath.Join(dir, traceFileName)

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.Write(data)
	f.Write([]byte("\n"))
}

// ─── Reader ─────────────────────────────────────────────────────────────────

// ReadTraces reads all trace events for a session in chronological order.
func (l *Logger) ReadTraces(sessionName string) ([]map[string]interface{}, error) {
	path := filepath.Join(l.sessionsDir, sessionName, traceFileName)
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
		events = append(events, evt)
	}
	return events, nil
}

// ReadRecentTraces reads the last N trace events for a session.
func (l *Logger) ReadRecentTraces(sessionName string, n int) ([]map[string]interface{}, error) {
	all, err := l.ReadTraces(sessionName)
	if err != nil {
		return nil, err
	}
	if len(all) <= n {
		return all, nil
	}
	return all[len(all)-n:], nil
}

// NodePath extracts the chronological node visit path from traces.
// Returns pairs of [nodeID, flowName] in order.
func (l *Logger) NodePath(sessionName string) ([][2]string, error) {
	traces, err := l.ReadTraces(sessionName)
	if err != nil {
		return nil, err
	}

	var path [][2]string
	for _, t := range traces {
		if t["event"] == "flow.node_enter" {
			node, _ := t["node"].(string)
			flow, _ := t["flow"].(string)
			path = append(path, [2]string{node, flow})
		}
	}
	return path, nil
}

// CountNodeVisits returns how many times each node was entered in a session.
func (l *Logger) CountNodeVisits(sessionName string) (map[string]int, error) {
	traces, err := l.ReadTraces(sessionName)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, t := range traces {
		if t["event"] == "flow.node_enter" {
			node, _ := t["node"].(string)
			counts[node]++
		}
	}
	return counts, nil
}

// Summary returns a human-readable one-line summary of a session's traces.
func (l *Logger) Summary(sessionName string) (string, error) {
	traces, err := l.ReadTraces(sessionName)
	if err != nil {
		return "", err
	}
	if len(traces) == 0 {
		return "no trace events", nil
	}

	nodes := 0
	gatesPassed := 0
	gatesFailed := 0
	toolCalls := 0
	firstTS := ""
	lastTS := ""

	for _, t := range traces {
		switch t["event"] {
		case "flow.node_enter":
			nodes++
		case "flow.gate_check":
			if passed, _ := t["passed"].(bool); passed {
				gatesPassed++
			} else {
				gatesFailed++
			}
		case "ai.tool_call":
			toolCalls++
		}
		if firstTS == "" {
			firstTS, _ = t["ts"].(string)
		}
		lastTS, _ = t["ts"].(string)
	}

	return fmt.Sprintf("%d nodes, %d/%d gates passed, %d tool calls [%s → %s]",
		nodes, gatesPassed, gatesPassed+gatesFailed, toolCalls,
		shortTS(firstTS), shortTS(lastTS)), nil
}

func shortTS(ts string) string {
	if len(ts) >= 16 {
		return ts[11:16] // HH:MM only
	}
	return ts
}