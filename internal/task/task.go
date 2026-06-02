package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/eventlog"
	taskSync "github.com/origadmin/team-flow/internal/sync"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "task",
	Short: "Task management (beads-driven)",
	Long: `Task management for v3. Routes to beads internally.

AI always uses "flow task" — beads is the internal implementation.

Subcommands:
  create   Create a task (--title, --type, --parent, --description)
  show     Show task details
  update   Update task (--status, --claim)
  list     List tasks
  close    Close a task
  label    Manage task labels (add, remove, list)
  children List child tasks of a parent
  note     Append a note to a task
  sync     Sync external issues (GitHub) to local tasks`,
	DisableFlagParsing: true,
	RunE:               runTask,
}

func init() {
	Cmd.AddCommand(taskSync.Cmd)
}

func ensureBeads() error {
	if bd.IsAvailable() {
		return nil
	}
	fmt.Println("⚠ beads (bd CLI) not found, installing...")
	if err := bd.Install(); err != nil {
		return fmt.Errorf("beads installation failed: %w\n  Manual install: https://github.com/steveyegge/beads", err)
	}
	if err := bd.EnsureOnPath(); err != nil {
		fmt.Printf("  ⚠ Installed but not on PATH. Restart terminal or add %s to PATH.\n", filepath.Dir(bd.FindPath()))
	}
	fmt.Println("  ✓ beads installed")
	return nil
}

func runTask(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	if args[0] == "--help" || args[0] == "-h" {
		return cmd.Help()
	}
	if args[0] == "sync" {
		subCmd, _, err := cmd.Find(args)
		if err == nil && subCmd != cmd {
			subCmd.SetArgs(args[1:])
			return subCmd.Execute()
		}
	}

	if err := ensureBeads(); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectRoot := config.ResolveProjectRoot(cwd)
	if projectRoot == "" {
		projectRoot = cwd
	}

	return runTaskV3(cmd, args, projectRoot)
}

func runTaskV3(cmd *cobra.Command, args []string, projectRoot string) error {
	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "create", "ready":
		return taskCreateV3(projectRoot, subArgs)
	case "show", "get":
		return taskShowV3(projectRoot, subArgs)
	case "update":
		return taskUpdateV3(projectRoot, subArgs)
	case "append", "note":
		return taskNoteV3(projectRoot, subArgs)
	case "list":
		return taskListV3(projectRoot, subArgs)
	case "close", "done":
		return taskCloseV3(projectRoot, subArgs)
	case "label":
		return taskLabelV3(projectRoot, subArgs)
	case "children":
		return taskChildrenV3(projectRoot, subArgs)
	default:
		return fmt.Errorf("unknown task subcommand: %s (available: create, show, update, note, list, close, label, children)", subCmd)
	}
}

func taskCreateV3(projectRoot string, args []string) error {
	var title, taskType, description, parent string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				title = args[i+1]
				i++
			}
		case "--type":
			if i+1 < len(args) {
				taskType = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			}
		case "--parent":
			if i+1 < len(args) {
				parent = args[i+1]
				i++
			}
		}
	}

	if title == "" {
		return fmt.Errorf("--title is required")
	}

	bdArgs := []string{"create", "--silent", "--title", title}
	if taskType != "" {
		bdArgs = append(bdArgs, "--type", taskType)
	} else {
		bdArgs = append(bdArgs, "--type", "task")
	}
	if description != "" {
		bdArgs = append(bdArgs, "--description", description)
	}
	if parent != "" {
		bdArgs = append(bdArgs, "--parent", parent)
	}

	output, err := bd.RunQuiet(bdArgs...)
	if err != nil {
		return fmt.Errorf("create task: %w\n%s", err, output)
	}

	taskID := strings.TrimSpace(output)
	if parent != "" {
		fmt.Printf("✓ Sub-task created: %s (parent: %s)\n", taskID, parent)
	} else {
		fmt.Printf("✓ Task created: %s\n", taskID)
	}
	fmt.Printf("  Title: %s\n", title)
	if taskType != "" {
		fmt.Printf("  Type: %s\n", taskType)
	}

	// Write task.created event to events.mdl
	if lgr, err := eventlog.NewLogger(projectRoot); err == nil {
		sessionName, _ := lgr.LastSessionName()
		_ = lgr.TaskCreated(sessionName, taskID, taskType, title)
	}

	fmt.Printf("\n  Next: flow proc run --task %s tri3\n", taskID)
	return nil
}

func taskShowV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task_id is required")
	}
	taskID := args[0]

	output, err := bd.RunQuiet("show", taskID)
	if err != nil {
		return fmt.Errorf("task not found: %s (%w)", taskID, err)
	}
	fmt.Print(output)

	var isEpic bool
	if jsonOutput, jsonErr := bd.RunQuiet("show", taskID, "--json"); jsonErr == nil {
		var raw map[string]interface{}
		if json.Unmarshal([]byte(jsonOutput), &raw) == nil {
			if t, ok := raw["type"].(string); ok && t == "epic" {
				isEpic = true
			}
		}
	}

	if isEpic {
		childrenOutput, childrenErr := bd.RunQuiet("children", taskID)
		if childrenErr == nil && strings.TrimSpace(childrenOutput) != "" {
			fmt.Printf("\n--- Children of %s ---\n", taskID)
			fmt.Print(childrenOutput)
		}
	}

	return nil
}

func taskUpdateV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task_id is required")
	}
	taskID := args[0]

	bdArgs := []string{"update", taskID}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--status":
			if i+1 < len(args) {
				bdArgs = append(bdArgs, "--status", args[i+1])
				i++
			}
		case "--claim":
			bdArgs = append(bdArgs, "--status", "in_progress")
		default:
			bdArgs = append(bdArgs, args[i])
		}
	}

	output, err := bd.RunQuiet(bdArgs...)
	if err != nil {
		return fmt.Errorf("update task: %w\n%s", err, output)
	}
	fmt.Printf("✓ Task updated: %s\n", taskID)
	return nil
}

func taskNoteV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task_id is required")
	}
	taskID := args[0]

	var note string
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--note", "--message":
			if i+1 < len(args) {
				note = args[i+1]
				i++
			}
		}
	}
	if note == "" && len(args) > 1 {
		note = args[len(args)-1]
	}
	if note == "" {
		return fmt.Errorf("--note is required")
	}

	output, err := bd.RunQuiet("note", taskID, "--text", note)
	if err != nil {
		return fmt.Errorf("append note: %w\n%s", err, output)
	}
	fmt.Printf("✓ Note appended to task %s\n", taskID)
	return nil
}

func taskListV3(projectRoot string, args []string) error {
	bdArgs := []string{"list", "--status", "open"}
	if len(args) > 0 {
		bdArgs = append(bdArgs, args...)
	}

	output, err := bd.RunQuiet(bdArgs...)
	if err != nil {
		fmt.Println("No tasks found.")
		return nil
	}
	fmt.Print(output)
	return nil
}

func taskCloseV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task_id is required")
	}
	taskID := args[0]

	var prevStatus string
	var isEpic bool
	if output, err := bd.RunQuiet("show", taskID, "--json"); err == nil {
		var raw map[string]interface{}
		if json.Unmarshal([]byte(output), &raw) == nil {
			if s, ok := raw["status"].(string); ok {
				prevStatus = s
			}
			if t, ok := raw["type"].(string); ok && t == "epic" {
				isEpic = true
			}
		}
	}

	output, err := bd.RunQuiet("close", taskID)
	if err != nil {
		return fmt.Errorf("close task: %w\n%s", err, output)
	}
	fmt.Printf("✓ Task closed: %s\n", taskID)

	if verifyOutput, verifyErr := bd.RunQuiet("show", taskID, "--json"); verifyErr == nil {
		var raw map[string]interface{}
		if json.Unmarshal([]byte(verifyOutput), &raw) == nil {
			if s, ok := raw["status"].(string); ok && s != "closed" {
				fmt.Printf("  ⚠ Warning: task status is '%s', not 'closed'\n", s)
			}
		}
	}

	if isEpic {
		if childrenOutput, childrenErr := bd.RunQuiet("children", taskID, "--json"); childrenErr == nil {
			var children []interface{}
			if json.Unmarshal([]byte(childrenOutput), &children) == nil {
				allClosed := true
				for _, child := range children {
					if childMap, ok := child.(map[string]interface{}); ok {
						if childID, ok := childMap["id"].(string); ok {
							childStatus, err := GetTaskField(childID, "status")
							if err == nil && childStatus != "closed" {
								fmt.Printf("  ⚠ Warning: child task %s is still '%s'\n", childID, childStatus)
								allClosed = false
							}
						}
					}
				}
				if allClosed && len(children) > 0 {
					fmt.Printf("  ✓ All children of epic %s are closed\n", taskID)
				}
			}
		} else {
			if childrenOutput, childrenErr := bd.RunQuiet("children", taskID); childrenErr == nil {
				children := strings.TrimSpace(childrenOutput)
				if children != "" {
					lines := strings.Split(children, "\n")
					allClosed := true
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line == "" || strings.HasPrefix(line, "---") {
							continue
						}
						parts := strings.Fields(line)
						if len(parts) > 0 {
							childID := parts[0]
							childStatus, err := GetTaskField(childID, "status")
							if err == nil && childStatus != "closed" {
								fmt.Printf("  ⚠ Warning: child task %s is still '%s'\n", childID, childStatus)
								allClosed = false
							}
						}
					}
					if allClosed {
						fmt.Printf("  ✓ All children of epic %s are closed\n", taskID)
					}
				}
			}
		}
	}

	if lgr, err := eventlog.NewLogger(projectRoot); err == nil {
		sessionName, _ := lgr.LastSessionName()
		_ = lgr.TaskStatusChange(sessionName, taskID, prevStatus, "closed")
	}

	return nil
}

func taskLabelV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: flow task label <add|remove|list> <task-id> [label...]")
	}
	action := args[0]
	rest := args[1:]

	switch action {
	case "add":
		if len(rest) < 2 {
			return fmt.Errorf("usage: flow task label add <task-id> <label...>")
		}
		taskID := rest[0]
		labels := rest[1:]
		output, err := bd.RunQuiet(append([]string{"label", "add", taskID}, labels...)...)
		if err != nil {
			return fmt.Errorf("add label: %w\n%s", err, output)
		}
		fmt.Printf("✓ Labels added to %s: %s\n", taskID, strings.Join(labels, ", "))

	case "remove":
		if len(rest) < 2 {
			return fmt.Errorf("usage: flow task label remove <task-id> <label...>")
		}
		taskID := rest[0]
		labels := rest[1:]
		output, err := bd.RunQuiet(append([]string{"label", "remove", taskID}, labels...)...)
		if err != nil {
			return fmt.Errorf("remove label: %w\n%s", err, output)
		}
		fmt.Printf("✓ Labels removed from %s: %s\n", taskID, strings.Join(labels, ", "))

	case "list":
		if len(rest) < 1 {
			return fmt.Errorf("usage: flow task label list <task-id>")
		}
		taskID := rest[0]
		output, err := bd.RunQuiet("label", "list", taskID)
		if err != nil {
			return fmt.Errorf("list labels: %w\n%s", err, output)
		}
		fmt.Print(output)

	default:
		return fmt.Errorf("unknown label action: %s (available: add, remove, list)", action)
	}
	return nil
}

func taskChildrenV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("parent task_id is required")
	}
	parentID := args[0]

	bdArgs := []string{"children", parentID}
	if len(args) > 1 {
		bdArgs = append(bdArgs, args[1:]...)
	}

	output, err := bd.RunQuiet(bdArgs...)
	if err != nil {
		return fmt.Errorf("list children: %w\n%s", err, output)
	}
	fmt.Print(output)
	return nil
}

func GetTaskLabels(taskID string) (map[string]string, error) {
	if err := ensureBeads(); err != nil {
		return nil, err
	}

	output, err := bd.RunQuiet("label", "list", taskID, "--json")
	if err != nil {
		return nil, fmt.Errorf("get labels: %w", err)
	}

	var labels []string
	if err := json.Unmarshal([]byte(output), &labels); err != nil {
		labelMap := make(map[string]string)
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				labelMap[parts[0]] = parts[1]
			} else {
				labelMap[line] = ""
			}
		}
		if len(labelMap) > 0 {
			return labelMap, nil
		}
		return nil, fmt.Errorf("parse labels: %w", err)
	}

	labelMap := make(map[string]string)
	for _, l := range labels {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) == 2 {
			labelMap[parts[0]] = parts[1]
		} else {
			labelMap[l] = ""
		}
	}
	return labelMap, nil
}

func SetTaskLabel(taskID, key, value string) error {
	if err := ensureBeads(); err != nil {
		return err
	}

	label := key
	if value != "" {
		label = key + "=" + value
	}

	output, err := bd.RunQuiet("label", "add", taskID, label)
	if err != nil {
		return fmt.Errorf("set label: %w\n%s", err, output)
	}
	return nil
}

func RemoveTaskLabel(taskID, key string) error {
	if err := ensureBeads(); err != nil {
		return err
	}

	output, err := bd.RunQuiet("label", "remove", taskID, key)
	if err != nil {
		return fmt.Errorf("remove label: %w\n%s", err, output)
	}
	return nil
}

func GetTaskField(taskID, field string) (string, error) {
	if err := ensureBeads(); err != nil {
		return "", err
	}

	output, err := bd.RunQuiet("show", taskID, "--json")
	if err != nil {
		return "", fmt.Errorf("show task: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return "", fmt.Errorf("parse task: %w", err)
	}

	if val, ok := raw[field]; ok {
		switch v := val.(type) {
		case string:
			return v, nil
		case fmt.Stringer:
			return v.String(), nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	}
	return "", nil
}
