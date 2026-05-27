package proc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/flow"
	"github.com/origadmin/team-flow/internal/updater"
	"github.com/origadmin/team-flow/internal/version"
	"github.com/spf13/cobra"
)

var (
	procRootDir  string
	procTemplate string
	procFlowName string
	procFormat   string
	procTaskID   string
	procRunGate  bool
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

var ruleCmd = &cobra.Command{
	Use:   "rule <rule-id>",
	Short: "Show full rule definition",
	Long:  `Show the full definition of a rule by its ID, including description and enforcement details.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRule,
}

var gateCmd = &cobra.Command{
	Use:   "gate [node-id]",
	Short: "Run gate checks for a node",
	Long: `Run automated gate checks for a specified node.
Without node-id, checks the first gate node found in the flow.
With node-id, checks the specified node's gate conditions.

Automated checks (tests_pass, lint_pass, deliverables_complete, no_regressions)
are executed by the engine. Custom checks require AI judgment and are reported as pending.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGate,
}

func init() {
	Cmd.AddCommand(runCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(validateCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(ruleCmd)
	Cmd.AddCommand(gateCmd)

	Cmd.PersistentFlags().StringVar(&procRootDir, "root", "", "Root directory for preset processes (default: current directory)")
	runCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name (default: project's configured flow from .team/project.md)")
	runCmd.Flags().StringVar(&procFormat, "format", "json", "Output format: json or text")
	runCmd.Flags().StringVar(&procTaskID, "task", "", "Task ID for variable substitution")
	runCmd.Flags().BoolVar(&procRunGate, "gate", true, "Run automated gate checks when encountering a gate node")
	ruleCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name to look up the rule definition")
	ruleCmd.Flags().StringVar(&procFormat, "format", "text", "Output format: json or text")
	gateCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name (default: project's configured flow from .team/project.md)")
	gateCmd.Flags().StringVar(&procFormat, "format", "text", "Output format: json or text")
	createCmd.Flags().StringVar(&procTemplate, "template", "", "Template process name from v3/flows/ to base the new process on")
}

func getRootDir() string {
	if procRootDir != "" {
		return procRootDir
	}
	dir, _ := os.Getwd()
	return config.FindTeamRoot(dir)
}

func getProjectRootDir() string {
	dir, _ := os.Getwd()
	return config.ResolveProjectRoot(dir)
}

// getWorkspaceRoot 获取 workspace 根目录
func getWorkspaceRoot() string {
	dir, _ := os.Getwd()
	workspace := config.FindWorkspaceRoot(dir)
	if workspace == "" {
		workspace = dir
	}
	return workspace
}

func runRun(cmd *cobra.Command, args []string) error {
	teamRoot := getRootDir()
	projectRoot := getProjectRootDir()
	workspace := getWorkspaceRoot()

	if projectRoot == "" {
		projectRoot = teamRoot
	}

	// Check for updates (with cache, high performance)
	if procFormat == "text" { // Only show for text format
		updateCheck, _ := updater.CheckForUpdateWithCache(version.Version, projectRoot, false)
		if updateCheck != nil && updateCheck.HasUpdate {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ⬆ flow CLI update available: %s → %s\n", updateCheck.CurrentVersion, updateCheck.LatestVersion)
			fmt.Fprintf(cmd.ErrOrStderr(), "  Run 'flow update' to update\n\n")
		}
	}

	nodeID := ""
	if len(args) > 0 {
		nodeID = args[0]
	}

	req := ProcRunRequest{
		FlowName:    procFlowName,
		NodeID:      nodeID,
		TaskID:      procTaskID,
		ProjectRoot: projectRoot,
		TeamRoot:    teamRoot,
		Workspace:   workspace,
		Format:      procFormat,
		RunGate:     procRunGate,
	}

	engine := NewProcRunEngine(teamRoot)
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
	Name        string
	Path        string
	Source      string
	Description string
	IsDefault   bool
	Registered  bool
}

func listProcs(root string) []ProcEntry {
	var entries []ProcEntry

	regMap := readFlowRegistry(root)
	defaultFlow := readDefaultFlow(root)

	presetDir := filepath.Join(root, "v3", "flows")
	if files, err := filepath.Glob(filepath.Join(presetDir, "*.json")); err == nil {
		for _, f := range files {
			name := strings.TrimSuffix(filepath.Base(f), ".json")
			desc := ""
			if reg, ok := regMap[name]; ok {
				desc = reg
			}
			entries = append(entries, ProcEntry{
				Name:        name,
				Path:        f,
				Source:      "preset",
				Description: desc,
				IsDefault:   name == defaultFlow,
				Registered:  ok(regMap, name),
			})
		}
	}

	projectDir := filepath.Join(root, ".team", "flows")
	if files, err := filepath.Glob(filepath.Join(projectDir, "*.json")); err == nil {
		for _, f := range files {
			name := strings.TrimSuffix(filepath.Base(f), ".json")
			desc := ""
			if reg, ok := regMap[name]; ok {
				desc = reg
			}
			entries = append(entries, ProcEntry{
				Name:        name,
				Path:        f,
				Source:      "project",
				Description: desc,
				IsDefault:   name == defaultFlow,
				Registered:  ok(regMap, name),
			})
		}
	}

	return entries
}

func ok(m map[string]string, key string) bool {
	_, exists := m[key]
	return exists
}

func readFlowRegistry(root string) map[string]string {
	result := make(map[string]string)
	projectMd := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMd)
	if err != nil {
		return result
	}

	inFlowsSection := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Flows" {
			inFlowsSection = true
			continue
		}
		if inFlowsSection && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if inFlowsSection && strings.HasPrefix(trimmed, "|") {
			fields := strings.Split(trimmed, "|")
			if len(fields) >= 5 {
				name := strings.TrimSpace(fields[1])
				desc := strings.TrimSpace(fields[3])
				if name != "" && name != "Flow" && !strings.HasPrefix(name, "-") {
					result[name] = desc
				}
			}
		}
	}

	return result
}

func readDefaultFlow(root string) string {
	projectMd := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMd)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "default_flow:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "default_flow:"))
		}
	}
	return ""
}

func printProcs(w io.Writer, entries []ProcEntry) {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No processes found.")
		return
	}

	fmt.Fprintf(w, "%-25s %-10s %-6s %-10s %s\n", "NAME", "SOURCE", "REG", "DEFAULT", "DESCRIPTION")
	fmt.Fprintf(w, "%-25s %-10s %-6s %-10s %s\n", strings.Repeat("-", 25), strings.Repeat("-", 10), strings.Repeat("-", 6), strings.Repeat("-", 10), strings.Repeat("-", 30))
	for _, e := range entries {
		regMark := "  "
		if e.Registered {
			regMark = "✓ "
		}
		defaultMark := ""
		if e.IsDefault {
			defaultMark = "⭐"
		}
		desc := e.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		fmt.Fprintf(w, "%-25s %-10s %-6s %-10s %s\n", e.Name, e.Source, regMark, defaultMark, desc)
	}
	fmt.Fprintf(w, "\nTotal: %d flow(s)", len(entries))

	registered := 0
	for _, e := range entries {
		if e.Registered {
			registered++
		}
	}
	fmt.Fprintf(w, " (%d registered)\n", registered)
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

	root := getRootDir()
	team, _ := LoadTeam(root)

	var result *flow.ValidationResult
	if team != nil {
		result = flow.ValidateFlowWithTeam(f, team)
	} else {
		result = flow.ValidateFlow(f)
	}
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

func runRule(cmd *cobra.Command, args []string) error {
	ruleID := args[0]
	root := getRootDir()

	flowName := procFlowName
	if flowName == "" {
		var err error
		flowName, err = ResolveDefaultFlowName(root)
		if err != nil {
			return fmt.Errorf("--flow is required when no default flow is configured: %w", err)
		}
	}

	procPath := resolveProcPath(root, flowName)
	if procPath == "" {
		return fmt.Errorf("flow not found: %s", flowName)
	}

	f, err := flow.ParseFlowFile(procPath)
	if err != nil {
		return fmt.Errorf("parse flow: %w", err)
	}

	if f.Components != nil {
		for _, r := range f.Components.Rules {
			if r.ID == ruleID {
				return printRule(cmd.OutOrStdout(), r, procFormat)
			}
		}
	}

	team, _ := LoadTeam(root)
	if team == nil {
		team, _ = LoadTeamFromFlowPath(procPath)
	}
	if team != nil {
		for _, r := range team.Rules {
			if r.ID == ruleID {
				return printRule(cmd.OutOrStdout(), r, procFormat)
			}
		}
	}

	return fmt.Errorf("rule %s not found in flow %s or team", ruleID, flowName)
}

func printRule(w io.Writer, r flow.RuleDefinition, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal rule: %w", err)
		}
		w.Write(data)
		w.Write([]byte("\n"))
		return nil
	default:
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Rule: %s\n", r.Name)
		fmt.Fprintf(w, "ID:   %s\n", r.ID)
		if r.Type != "" {
			fmt.Fprintf(w, "Type: %s\n", r.Type)
		}
		if r.Enforcement != "" {
			fmt.Fprintf(w, "Enforcement: %s\n", strings.ToUpper(string(r.Enforcement)))
		}
		if r.Instruction != "" {
			fmt.Fprintf(w, "\nInstruction:\n  %s\n", r.Instruction)
		}
		if r.Description != "" {
			fmt.Fprintf(w, "\nFull Description:\n  %s\n", r.Description)
		}
		if r.Source != "" {
			fmt.Fprintf(w, "\nSource: %s\n", r.Source)
		}
		fmt.Fprintf(w, "\n")
		return nil
	}
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
		filepath.Join(root, "v3", "flows", id+".json"),
		filepath.Join(root, ".team", "flows", id+".json"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

func runGate(cmd *cobra.Command, args []string) error {
	projectRoot := getRootDir()

	nodeID := ""
	if len(args) > 0 {
		nodeID = args[0]
	}

	flowName := procFlowName
	if flowName == "" {
		var err error
		flowName, err = ResolveDefaultFlowName(projectRoot)
		if err != nil {
			return fmt.Errorf("--flow is required when no default flow is configured: %w", err)
		}
	}

	procPath := resolveProcPath(projectRoot, flowName)
	if procPath == "" {
		return fmt.Errorf("flow not found: %s", flowName)
	}

	fl, err := flow.ParseFlowFile(procPath)
	if err != nil {
		return fmt.Errorf("parse flow: %w", err)
	}

	var targetNode *flow.FlowNode
	if nodeID != "" {
		for i := range fl.Nodes {
			if fl.Nodes[i].ID == nodeID {
				targetNode = &fl.Nodes[i]
				break
			}
		}
		if targetNode == nil {
			return fmt.Errorf("node not found: %s", nodeID)
		}
	} else {
		for i := range fl.Nodes {
			if fl.Nodes[i].Type == flow.NodeTypeGate {
				targetNode = &fl.Nodes[i]
				break
			}
		}
		if targetNode == nil {
			return fmt.Errorf("no gate node found in flow %s", flowName)
		}
	}

	if targetNode.Type != flow.NodeTypeGate {
		return fmt.Errorf("node %s is not a gate node (type: %s)", targetNode.ID, targetNode.Type)
	}

	conditions := ExtractGateConditions(targetNode)
	if len(conditions) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No gate conditions defined for node %s\n", targetNode.ID)
		return nil
	}

	team, _ := LoadTeam(projectRoot)
	docsInternal := config.ResolveInternalDocs(projectRoot)
	results := RunGateCheck(projectRoot, conditions, team, docsInternal)
	passed, summary := GateOverallResult(results)

	switch procFormat {
	case "json":
		data, err := json.MarshalIndent(map[string]interface{}{
			"node_id":  targetNode.ID,
			"passed":   passed,
			"summary":  summary,
			"results":  results,
		}, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal results: %w", err)
		}
		cmd.OutOrStdout().Write(data)
		cmd.OutOrStdout().Write([]byte("\n"))
	default:
		printGateResults(cmd.OutOrStdout(), targetNode.ID, passed, summary, results)
	}

	if !passed {
		return fmt.Errorf("gate check failed: %s", summary)
	}

	return nil
}

func printGateResults(w io.Writer, nodeID string, passed bool, summary string, results []GateCheckResult) {
	icon := "✓"
	if !passed {
		icon = "✗"
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "╔══════════════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(w, "║  %s GATE CHECK: %s\n", icon, padLine(nodeID, 58))
	fmt.Fprintf(w, "║  %s\n", padLine(summary, 68))
	fmt.Fprintf(w, "╠══════════════════════════════════════════════════════════════════════╣\n")

	for _, r := range results {
		statusIcon := "✓"
		if !r.Passed {
			statusIcon = "✗"
		}
		if r.Skipped {
			statusIcon = "⊘"
		}
		autoLabel := "[AUTO]"
		if !r.Auto {
			autoLabel = "[AI]"
		}
		reqLabel := ""
		if r.Required {
			reqLabel = " REQUIRED"
		}
		fmt.Fprintf(w, "║  %s %s %s%s\n", statusIcon, autoLabel, padLine(r.Type+reqLabel, 30), padLine("", 30))
		fmt.Fprintf(w, "║    %s\n", padLine(r.Message, 66))
	}

	fmt.Fprintf(w, "╚══════════════════════════════════════════════════════════════════════╝\n")
	fmt.Fprintf(w, "\n")
}
