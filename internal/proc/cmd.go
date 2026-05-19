package proc

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/flow"
	"github.com/spf13/cobra"
)

var (
	procRootDir  string
	procTemplate string
	procFlowName string
	procFormat   string
	procTaskID   string
)

var Cmd = &cobra.Command{
	Use:   "proc",
	Short: "Process lifecycle management",
	Long:  `Manage process lifecycle: run, list, show, validate, and create processes.`,
}

var runCmd = &cobra.Command{
	Use:   "run [node-id]",
	Short: "Run a process node and return instruction",
	Long: `Run a process node and return structured instruction for AI execution.
Without node-id, returns the root node instruction.
With node-id, returns the specified node's instruction.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRun,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available processes",
	Long:  `List all available processes from preset (v3/flows/) and project (.team/flows/) directories.`,
	RunE:  runList,
}

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show process definition details",
	Long:  `Show process definition details including metadata, node count, and edge count.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a process DSL file",
	Long:  `Validate a process DSL file and report any errors or warnings.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runValidate,
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new process",
	Long:  `Create a new process from a template or blank.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	Cmd.AddCommand(runCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(validateCmd)
	Cmd.AddCommand(createCmd)

	Cmd.PersistentFlags().StringVar(&procRootDir, "root", "", "Root directory for preset processes (default: current directory)")
	runCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name (default: project's configured flow from .team/project.md)")
	runCmd.Flags().StringVar(&procFormat, "format", "json", "Output format: json or text")
	runCmd.Flags().StringVar(&procTaskID, "task", "", "Task ID for variable substitution")
	createCmd.Flags().StringVar(&procTemplate, "template", "", "Template process name from v3/flows/ to base the new process on")
}

func getRootDir() string {
	if procRootDir != "" {
		return procRootDir
	}
	dir, _ := os.Getwd()
	return dir
}

func runRun(cmd *cobra.Command, args []string) error {
	root := getRootDir()

	nodeID := ""
	if len(args) > 0 {
		nodeID = args[0]
	}

	req := ProcRunRequest{
		FlowName:    procFlowName,
		NodeID:      nodeID,
		TaskID:      procTaskID,
		ProjectRoot: root,
		Format:      procFormat,
	}

	engine := NewProcRunEngine(root)
	result, err := engine.Run(context.Background(), req)
	if err != nil {
		return err
	}

	switch procFormat {
	case "text":
		return FormatText(cmd.OutOrStdout(), result)
	default:
		return FormatJSON(cmd.OutOrStdout(), result)
	}
}

type ProcEntry struct {
	Name   string
	Path   string
	Source string
}

func listProcs(root string) []ProcEntry {
	var entries []ProcEntry

	presetDir := filepath.Join(root, "v3", "flows")
	if files, err := filepath.Glob(filepath.Join(presetDir, "*.json")); err == nil {
		for _, f := range files {
			name := strings.TrimSuffix(filepath.Base(f), ".json")
			entries = append(entries, ProcEntry{
				Name:   name,
				Path:   f,
				Source: "preset",
			})
		}
	}

	projectDir := filepath.Join(root, ".team", "flows")
	if files, err := filepath.Glob(filepath.Join(projectDir, "*.json")); err == nil {
		for _, f := range files {
			name := strings.TrimSuffix(filepath.Base(f), ".json")
			entries = append(entries, ProcEntry{
				Name:   name,
				Path:   f,
				Source: "project",
			})
		}
	}

	return entries
}

func printProcs(w io.Writer, entries []ProcEntry) {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No processes found.")
		return
	}

	fmt.Fprintf(w, "%-30s %-10s %s\n", "NAME", "SOURCE", "PATH")
	fmt.Fprintf(w, "%-30s %-10s %s\n", strings.Repeat("-", 30), strings.Repeat("-", 10), strings.Repeat("-", 40))
	for _, e := range entries {
		fmt.Fprintf(w, "%-30s %-10s %s\n", e.Name, e.Source, e.Path)
	}
	fmt.Fprintf(w, "\nTotal: %d process(es)\n", len(entries))
}

func runList(cmd *cobra.Command, args []string) error {
	root := getRootDir()
	entries := listProcs(root)
	printProcs(cmd.OutOrStdout(), entries)
	return nil
}

func showProc(w io.Writer, f *flow.Flow) {
	fmt.Fprintf(w, "Name:        %s\n", f.Metadata.Name)
	if f.Metadata.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", f.Metadata.Description)
	}
	if f.Metadata.Author != "" {
		fmt.Fprintf(w, "Author:      %s\n", f.Metadata.Author)
	}
	fmt.Fprintf(w, "Version:     %s\n", f.Version)
	fmt.Fprintf(w, "Nodes:       %d\n", len(f.Nodes))
	fmt.Fprintf(w, "Edges:       %d\n", len(f.Edges))

	if len(f.Metadata.Tags) > 0 {
		fmt.Fprintf(w, "Tags:        %s\n", strings.Join(f.Metadata.Tags, ", "))
	}

	if f.Config != nil && f.Config.TaskType != "" {
		fmt.Fprintf(w, "Task Type:   %s\n", f.Config.TaskType)
	}

	if len(f.Nodes) > 0 {
		fmt.Fprintln(w, "\nNodes:")
		for _, n := range f.Nodes {
			fmt.Fprintf(w, "  %-20s %-12s %s\n", n.ID, n.Type, n.Name)
		}
	}

	if len(f.Edges) > 0 {
		fmt.Fprintln(w, "\nEdges:")
		for _, e := range f.Edges {
			label := e.From + " -> " + e.To
			if e.Type != "" {
				label += " (" + string(e.Type) + ")"
			}
			fmt.Fprintf(w, "  %s\n", label)
		}
	}
}

func runShow(cmd *cobra.Command, args []string) error {
	id := args[0]
	root := getRootDir()

	procPath := resolveProcPath(root, id)
	if procPath == "" {
		return fmt.Errorf("process not found: %s", id)
	}

	f, err := flow.ParseFlowFile(procPath)
	if err != nil {
		return fmt.Errorf("parse process: %w", err)
	}

	showProc(cmd.OutOrStdout(), f)
	return nil
}

func printValidation(w io.Writer, f *flow.Flow, result *flow.ValidationResult) {
	if result.Valid {
		fmt.Fprintf(w, "✓ Process is valid: %s\n", f.Metadata.Name)
	} else {
		fmt.Fprintf(w, "✗ Process is invalid: %s\n", f.Metadata.Name)
	}

	if len(result.Errors) > 0 {
		fmt.Fprintf(w, "\nErrors (%d):\n", len(result.Errors))
		for _, e := range result.Errors {
			loc := ""
			if e.NodeID != "" {
				loc = fmt.Sprintf(" [node: %s]", e.NodeID)
			} else if e.EdgeID != "" {
				loc = fmt.Sprintf(" [edge: %s]", e.EdgeID)
			}
			field := ""
			if e.Field != "" {
				field = fmt.Sprintf(" (%s)", e.Field)
			}
			fmt.Fprintf(w, "  ✗ %s%s%s\n", e.Message, field, loc)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Fprintf(w, "\nWarnings (%d):\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			loc := ""
			if warn.NodeID != "" {
				loc = fmt.Sprintf(" [node: %s]", warn.NodeID)
			}
			field := ""
			if warn.Field != "" {
				field = fmt.Sprintf(" (%s)", warn.Field)
			}
			fmt.Fprintf(w, "  ⚠ %s%s%s\n", warn.Message, field, loc)
		}
	}
}

func runValidate(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	f, err := flow.ParseFlowFile(filePath)
	if err != nil {
		return fmt.Errorf("parse process: %w", err)
	}

	result := flow.ValidateFlow(f)
	printValidation(cmd.OutOrStdout(), f, result)

	if !result.Valid {
		return ErrValidationFailed
	}

	return nil
}

var ErrValidationFailed = fmt.Errorf("process validation failed")

func createProc(root, name, tmpl string) (string, error) {
	projectDir := filepath.Join(root, ".team", "flows")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", fmt.Errorf("create flows directory: %w", err)
	}

	targetPath := filepath.Join(projectDir, name+".json")
	if _, err := os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("process already exists: %s", targetPath)
	}

	if tmpl != "" {
		templatePath := resolveProcPath(root, tmpl)
		if templatePath == "" {
			return "", fmt.Errorf("template not found: %s", tmpl)
		}

		src, err := os.Open(templatePath)
		if err != nil {
			return "", fmt.Errorf("open template: %w", err)
		}
		defer src.Close()

		dst, err := os.Create(targetPath)
		if err != nil {
			return "", fmt.Errorf("create target: %w", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return "", fmt.Errorf("copy template: %w", err)
		}

		return targetPath, nil
	}

	newFlow := &flow.Flow{
		Version: "v3",
		Metadata: flow.FlowMetadata{
			Name: name,
		},
		Nodes: []flow.FlowNode{
			{ID: "start", Type: flow.NodeTypePhase, Name: "Start"},
			{ID: "done", Type: flow.NodeTypeTerminal, Name: "Done"},
		},
		Edges: []flow.FlowEdge{
			{From: "start", To: "done"},
		},
	}

	if err := flow.SerializeFlowFile(newFlow, targetPath); err != nil {
		return "", fmt.Errorf("write process: %w", err)
	}

	return targetPath, nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	root := getRootDir()

	targetPath, err := createProc(root, name, procTemplate)
	if err != nil {
		return err
	}

	if procTemplate != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created process from template: %s\n", targetPath)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Created blank process: %s\n", targetPath)
	}
	return nil
}

func resolveProcPath(root, id string) string {
	if filepath.IsAbs(id) {
		if _, err := os.Stat(id); err == nil {
			return id
		}
	}

	if filepath.Ext(id) == ".json" {
		if _, err := os.Stat(id); err == nil {
			return id
		}
	}

	candidates := []string{
		filepath.Join(root, ".team", "flows", id+".json"),
		filepath.Join(root, "v3", "flows", id+".json"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}
