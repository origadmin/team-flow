package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/eventlog"
	"github.com/origadmin/team-flow/internal/flow"
	"github.com/origadmin/team-flow/internal/proc"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Output structured diagnostic information",
	Long: `Output structured diagnostic information about the current flow context,
recent events, and session state.

  flow debug`,
	RunE: runDebug,
}

func init() {
	rootCmd.AddCommand(debugCmd)
}

func runDebug(cmd *cobra.Command, args []string) error {
	projectRoot := resolveDebugProjectRoot()
	if projectRoot == "" {
		return fmt.Errorf("not in a project directory")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)

	// ── Flow Context ──
	fmt.Fprintln(w, "=== FLOW CONTEXT ===")
	fmt.Fprintln(w, "FIELD\tVALUE")

	flowName, _ := proc.ResolveActiveFlow(projectRoot)
	if flowName == "" {
		flowName = "(not configured)"
	}
	fmt.Fprintf(w, "Flow Name\t%s\n", flowName)

	lgr, err := eventlog.NewLogger(projectRoot)
	if err != nil {
		fmt.Fprintf(w, "EventLog\tERROR: %v\n", err)
		w.Flush()
		return nil
	}

	sessionName, _ := lgr.LastSessionName()
	if sessionName == "" {
		sessionName = "(no session)"
	}
	fmt.Fprintf(w, "Session\t%s\n", sessionName)

	var state *proc.SessionState
	var currentNodeName string
	if sessionName != "(no session)" {
		state, _ = proc.LoadSessionState(lgr.SessionsDir(), sessionName)
		if state != nil {
			currentNodeName = state.CurrentNode
			if currentNodeName == "" {
				currentNodeName = state.CurrentNodeID
			}
			fmt.Fprintf(w, "Current Node\t%s (%s)\n", currentNodeName, state.CurrentNodeID)
			fmt.Fprintf(w, "Active Task\t%s\n", state.TaskID)
			fmt.Fprintf(w, "Flow Name (state)\t%s\n", state.FlowName)
			fmt.Fprintf(w, "Last Node At\t%s\n", state.LastNodeAt)
			if len(state.VisitedNodes) > 0 {
				fmt.Fprintf(w, "Visited Nodes\t%d nodes\n", len(state.VisitedNodes))
			}
		} else {
			fmt.Fprintln(w, "Current Node\t(no state.json)")
			fmt.Fprintln(w, "Active Task\t(none)")
		}
	}

	// Resolve flow file to get additional metadata
	if flowName != "(not configured)" {
		procPath := proc.ResolveProcPath(projectRoot, flowName)
		if procPath != "" {
			fl, parseErr := flow.ParseFlowFile(procPath)
			if parseErr == nil {
				fmt.Fprintf(w, "Flow ID\t%s\n", fl.Metadata.ID)
				fmt.Fprintf(w, "Flow Version\t%s\n", fl.Version)
				fmt.Fprintf(w, "Flow Nodes\t%d\n", len(fl.Nodes))
				fmt.Fprintf(w, "Flow Edges\t%d\n", len(fl.Edges))
				if fl.Config != nil && fl.Config.TaskType != "" {
					fmt.Fprintf(w, "Task Type\t%s\n", fl.Config.TaskType)
				}
			}
		}
	}
	w.Flush()

	// ── Session State (detailed) ──
	if state != nil {
		fmt.Fprintln(w, "\n=== SESSION STATE ===")
		fmt.Fprintln(w, "FIELD\tVALUE")
		fmt.Fprintf(w, "TaskID\t%s\n", state.TaskID)
		fmt.Fprintf(w, "CurrentNodeID\t%s\n", state.CurrentNodeID)
		fmt.Fprintf(w, "CurrentNode\t%s\n", state.CurrentNode)
		fmt.Fprintf(w, "FlowName\t%s\n", state.FlowName)
		fmt.Fprintf(w, "LastNodeAt\t%s\n", state.LastNodeAt)
		if len(state.VisitedNodes) > 0 {
			fmt.Fprintf(w, "VisitedNodes\t%s\n", strings.Join(state.VisitedNodes, " → "))
		}
		w.Flush()
	}

	// ── Recent Events ──
	fmt.Fprintln(w, "\n=== RECENT EVENTS (last 10) ===")
	fmt.Fprintln(w, "TS\tEVENT\tDETAILS")

	events, _ := lgr.ReadEvents("")
	start := 0
	if len(events) > 10 {
		start = len(events) - 10
	}
	for _, evt := range events[start:] {
		ts, _ := evt["ts"].(string)
		eventType, _ := evt["event"].(string)
		sess, _ := evt["session"].(string)

		var details string
		switch eventType {
		case eventlog.EventSessionStart:
			round, _ := evt["round"].(float64)
			input, _ := evt["input"].(string)
			if len(input) > 50 {
				input = input[:50] + "..."
			}
			details = fmt.Sprintf("R%d %s [%s]", int(round), input, sess)
		case eventlog.EventSessionAnalysis:
			status, _ := evt["status"].(string)
			task, _ := evt["task"].(string)
			details = fmt.Sprintf("%s task=%s [%s]", status, task, sess)
		case eventlog.EventTaskCreated:
			task, _ := evt["task"].(string)
			title, _ := evt["title"].(string)
			details = fmt.Sprintf("%s: %s [%s]", task, title, sess)
		case eventlog.EventTaskStatus:
			task, _ := evt["task"].(string)
			from, _ := evt["from"].(string)
			to, _ := evt["to"].(string)
			details = fmt.Sprintf("%s: %s→%s [%s]", task, from, to, sess)
		case eventlog.EventFlowNode:
			nodeName, _ := evt["node_name"].(string)
			if nodeName == "" {
				nodeName, _ = evt["node"].(string)
			}
			fl, _ := evt["flow"].(string)
			details = fmt.Sprintf("%s (%s) [%s]", nodeName, fl, sess)
		case eventlog.EventFlowEnded:
			status, _ := evt["status"].(string)
			details = fmt.Sprintf("status=%s [%s]", status, sess)
		default:
			details = fmt.Sprintf("[%s]", sess)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", ts, eventType, details)
	}
	w.Flush()

	// ── Status ──
	fmt.Fprintln(w, "\n=== STATUS ===")
	fmt.Fprintln(w, "FIELD\tVALUE")

	if state != nil && state.TaskID != "" {
		fmt.Fprintf(w, "TaskID\t%s\n", state.TaskID)
	}

	activeTasks, _ := lgr.ActiveTasks()
	if len(activeTasks) > 0 {
		fmt.Fprintf(w, "Active Tasks\t%d\n", len(activeTasks))
		for _, t := range activeTasks {
			fmt.Fprintf(w, "  - %s\tactive\n", t)
		}
	} else {
		fmt.Fprintln(w, "Active Tasks\t(none)")
	}

	// Last gate check info from events
	lastGate := "(none)"
	lastGatePassed := ""
	for i := len(events) - 1; i >= 0; i-- {
		if events[i]["event"] == eventlog.EventFlowNode {
			// Check if there's gate-related info in the trace
			break
		}
	}
	fmt.Fprintf(w, "Last Gate\t%s\n", lastGate)
	if lastGatePassed != "" {
		fmt.Fprintf(w, "Last Gate Passed\t%s\n", lastGatePassed)
	}

	w.Flush()

	return nil
}

func resolveDebugProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	root := config.ResolveProjectRoot(cwd)
	if root == "" {
		root = cwd
	}
	// Verify .team/ exists
	if _, err := os.Stat(filepath.Join(root, ".team")); os.IsNotExist(err) {
		return ""
	}
	return root
}