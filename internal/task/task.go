package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/eventlog"
	"github.com/origadmin/team-flow/internal/idgen"
	taskSync "github.com/origadmin/team-flow/internal/sync"
	"github.com/spf13/cobra"
)

// Task is the in-memory representation of a flow task.
type Task struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Description string                 `json:"description,omitempty"`
	Parent      string                 `json:"parent,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Notes       []string               `json:"notes,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ClosedAt    *time.Time             `json:"closed_at,omitempty"`
	Extras      map[string]interface{} `json:"extras,omitempty"`
}

var Cmd = &cobra.Command{
	Use:   "task",
	Short: "Task management",
	Long: `Task management for v3. Tasks are stored locally under
.team/tasks/ inside the project root and tracked in the project-level
events log.

Subcommands:
  create   Create a task (--title, --type, --parent, --description)
  show     Show task details
  update   Update task (--status, --type, --claim)
  list     List open tasks
  close    Close a task
  label    Manage task labels (add, remove, list)
  children List child tasks of a parent task
  note     Append a note to a task
  sync     Sync external issues (GitHub) to local tasks`,
	DisableFlagParsing: true,
	RunE:               runTask,
}

func init() {
	Cmd.AddCommand(taskSync.Cmd)
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

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectRoot := config.ResolveProjectRootWithActive(cwd)
	if projectRoot == "" {
		projectRoot = cwd
	}

	return runTaskV3(args, projectRoot)
}

func runTaskV3(args []string, projectRoot string) error {
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

// ─── storage helpers ───────────────────────────────────────────────────────

func tasksDir(root string) string {
	return filepath.Join(root, ".team", "tasks")
}

func taskPath(root, id string) string {
	return filepath.Join(tasksDir(root), id+".json")
}

func loadTask(root, id string) (*Task, error) {
	data, err := os.ReadFile(taskPath(root, id))
	if err != nil {
		return nil, err
	}
	var t Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse task %s: %w", id, err)
	}
	return &t, nil
}

func saveTask(root string, t *Task) error {
	dir := tasksDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	t.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(taskPath(root, t.ID), data, 0o644)
}

func allTaskIDs(root string) ([]string, error) {
	dir := tasksDir(root)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			ids = append(ids, strings.TrimSuffix(name, ".json"))
		}
	}
	return ids, nil
}

// ─── subcommands ───────────────────────────────────────────────────────────

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
	if taskType == "" {
		taskType = "task"
	}

	// Generate task ID using idgen (not beads — flow owns its task IDs).
	taskID := idgen.RandHex(2)

	now := time.Now().UTC()
	task := &Task{
		ID:          taskID,
		Title:       title,
		Type:        taskType,
		Status:      "open",
		Description: description,
		Parent:      parent,
		CreatedAt:   now,
		UpdatedAt:   now,
		Labels:      map[string]string{},
		Notes:       []string{},
	}

	if err := saveTask(projectRoot, task); err != nil {
		return fmt.Errorf("save task: %w", err)
	}

	if parent != "" {
		fmt.Printf("✓ Sub-task created: %s (parent: %s)\n", task.ID, parent)
	} else {
		fmt.Printf("✓ Task created: %s\n", task.ID)
	}
	fmt.Printf("  Title: %s\n", title)
	fmt.Printf("  Type: %s\n", taskType)

	if lgr, err := eventlog.NewLogger(projectRoot); err == nil {
		sessionName, _ := lgr.LastSessionName()
		_ = lgr.TaskCreated(sessionName, task.ID, taskType, title)
	}

	fmt.Printf("\n  Next: flow proc run --task %s tri3\n", task.ID)
	return nil
}

func taskShowV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task_id is required")
	}
	taskID := args[0]
	jsonOut := false
	for i := 1; i < len(args); i++ {
		if args[i] == "--json" {
			jsonOut = true
		}
	}
	task, err := loadTask(projectRoot, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %s (%w)", taskID, err)
	}

	if jsonOut {
		data, _ := json.MarshalIndent(task, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("ID:          %s\n", task.ID)
	fmt.Printf("Title:       %s\n", task.Title)
	fmt.Printf("Type:        %s\n", task.Type)
	fmt.Printf("Status:      %s\n", task.Status)
	if task.Parent != "" {
		fmt.Printf("Parent:      %s\n", task.Parent)
	}
	fmt.Printf("Created:     %s\n", task.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated:     %s\n", task.UpdatedAt.Format(time.RFC3339))
	if task.Description != "" {
		fmt.Printf("Description: %s\n", task.Description)
	}
	if len(task.Labels) > 0 {
		keys := make([]string, 0, len(task.Labels))
		for k := range task.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Println("Labels:")
		for _, k := range keys {
			fmt.Printf("  %s=%s\n", k, task.Labels[k])
		}
	}
	if len(task.Notes) > 0 {
		fmt.Println("Notes:")
		for _, n := range task.Notes {
			fmt.Printf("  - %s\n", n)
		}
	}

	children, err := childrenOf(projectRoot, taskID)
	if err == nil && len(children) > 0 {
		fmt.Printf("\n--- Children of %s (%d) ---\n", taskID, len(children))
		for _, c := range children {
			fmt.Printf("  %s  %s  [%s]\n", c.ID, c.Title, c.Status)
		}
	}
	return nil
}

func taskUpdateV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task_id is required")
	}
	taskID := args[0]
	task, err := loadTask(projectRoot, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %s (%w)", taskID, err)
	}

	prevStatus := task.Status
	changed := false
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--status":
			if i+1 < len(args) {
				task.Status = args[i+1]
				changed = true
				i++
			}
		case "--claim":
			task.Status = "in_progress"
			changed = true
		case "--type":
			if i+1 < len(args) {
				task.Type = args[i+1]
				changed = true
				i++
			}
		case "--title":
			if i+1 < len(args) {
				task.Title = args[i+1]
				changed = true
				i++
			}
		case "--description":
			if i+1 < len(args) {
				task.Description = args[i+1]
				changed = true
				i++
			}
		}
	}

	if !changed {
		fmt.Printf("nothing to update for task %s\n", taskID)
		return nil
	}

	if err := saveTask(projectRoot, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	fmt.Printf("✓ Task updated: %s\n", taskID)

	if prevStatus != task.Status {
		if lgr, err := eventlog.NewLogger(projectRoot); err == nil {
			sessionName, _ := lgr.LastSessionName()
			_ = lgr.TaskStatusChange(sessionName, taskID, prevStatus, task.Status)
		}
	}
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

	task, err := loadTask(projectRoot, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %s (%w)", taskID, err)
	}
	task.Notes = append(task.Notes, fmt.Sprintf("[%s] %s", time.Now().UTC().Format(time.RFC3339), note))
	if err := saveTask(projectRoot, task); err != nil {
		return fmt.Errorf("append note: %w", err)
	}
	fmt.Printf("✓ Note appended to task %s\n", taskID)
	return nil
}

func taskListV3(projectRoot string, args []string) error {
	statusFilter := "open"
	for _, a := range args {
		if a == "--all" {
			statusFilter = ""
		}
	}
	ids, err := allTaskIDs(projectRoot)
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}
	if len(ids) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	var tasks []Task
	for _, id := range ids {
		if t, err := loadTask(projectRoot, id); err == nil {
			tasks = append(tasks, *t)
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	fmt.Printf("%-8s %-12s %-12s %s\n", "ID", "STATUS", "TYPE", "TITLE")
	fmt.Printf("%-8s %-12s %-12s %s\n", "--------", "------------", "------------", "--------------------")
	for _, t := range tasks {
		if statusFilter != "" && t.Status != statusFilter {
			continue
		}
		fmt.Printf("%-8s %-12s %-12s %s\n", t.ID, t.Status, t.Type, t.Title)
	}
	return nil
}

func taskCloseV3(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task_id is required")
	}
	taskID := args[0]
	task, err := loadTask(projectRoot, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %s (%w)", taskID, err)
	}

	prevStatus := task.Status
	task.Status = "closed"
	now := time.Now().UTC()
	task.ClosedAt = &now
	if err := saveTask(projectRoot, task); err != nil {
		return fmt.Errorf("close task: %w", err)
	}
	fmt.Printf("✓ Task closed: %s\n", taskID)

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
		task, err := loadTask(projectRoot, taskID)
		if err != nil {
			return fmt.Errorf("task not found: %s (%w)", taskID, err)
		}
		if task.Labels == nil {
			task.Labels = map[string]string{}
		}
		for _, kv := range labels {
			if idx := strings.Index(kv, "="); idx >= 0 {
				task.Labels[kv[:idx]] = kv[idx+1:]
			} else {
				task.Labels[kv] = ""
			}
		}
		if err := saveTask(projectRoot, task); err != nil {
			return fmt.Errorf("add label: %w", err)
		}
		fmt.Printf("✓ Labels added to %s: %s\n", taskID, strings.Join(labels, ", "))
	case "remove":
		if len(rest) < 2 {
			return fmt.Errorf("usage: flow task label remove <task-id> <label...>")
		}
		taskID := rest[0]
		labels := rest[1:]
		task, err := loadTask(projectRoot, taskID)
		if err != nil {
			return fmt.Errorf("task not found: %s (%w)", taskID, err)
		}
		for _, k := range labels {
			delete(task.Labels, k)
		}
		if err := saveTask(projectRoot, task); err != nil {
			return fmt.Errorf("remove label: %w", err)
		}
		fmt.Printf("✓ Labels removed from %s: %s\n", taskID, strings.Join(labels, ", "))
	case "list":
		if len(rest) < 1 {
			return fmt.Errorf("usage: flow task label list <task-id>")
		}
		taskID := rest[0]
		task, err := loadTask(projectRoot, taskID)
		if err != nil {
			return fmt.Errorf("task not found: %s (%w)", taskID, err)
		}
		if len(task.Labels) == 0 {
			fmt.Printf("(no labels on %s)\n", taskID)
			return nil
		}
		keys := make([]string, 0, len(task.Labels))
		for k := range task.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s=%s\n", k, task.Labels[k])
		}
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
	children, err := childrenOf(projectRoot, parentID)
	if err != nil {
		return fmt.Errorf("list children: %w", err)
	}
	if len(children) == 0 {
		fmt.Printf("No children of %s\n", parentID)
		return nil
	}
	fmt.Printf("--- Children of %s (%d) ---\n", parentID, len(children))
	for _, c := range children {
		fmt.Printf("  %s  [%s]  %s\n", c.ID, c.Status, c.Title)
	}
	return nil
}

func childrenOf(projectRoot, parentID string) ([]Task, error) {
	ids, err := allTaskIDs(projectRoot)
	if err != nil {
		return nil, err
	}
	var out []Task
	for _, id := range ids {
		if t, err := loadTask(projectRoot, id); err == nil && t.Parent == parentID {
			out = append(out, *t)
		}
	}
	return out, nil
}

// ─── public helpers used by other packages ────────────────────────────────

func GetTaskLabels(root, taskID string) (map[string]string, error) {
	t, err := loadTask(root, taskID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(t.Labels))
	for k, v := range t.Labels {
		out[k] = v
	}
	return out, nil
}

func GetTaskField(root, taskID, field string) (string, error) {
	t, err := loadTask(root, taskID)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(field) {
	case "id":
		return t.ID, nil
	case "title":
		return t.Title, nil
	case "type":
		return t.Type, nil
	case "status":
		return t.Status, nil
	case "parent":
		return t.Parent, nil
	case "description":
		return t.Description, nil
	}
	return "", fmt.Errorf("unknown field: %s", field)
}
