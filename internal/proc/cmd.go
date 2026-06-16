package proc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/eventlog"
	"github.com/origadmin/team-flow/internal/flow"
	"github.com/origadmin/team-flow/internal/idgen"
	"github.com/origadmin/team-flow/internal/task"
	"github.com/origadmin/team-flow/internal/templates"
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
	procInput      string
	procAnalysis   string
	procAnalysisFile string
	procConclusion string
	procConclusionFile string
	procNewSession bool

	// node subcommand
	nodeAddID    string
	nodeAddType  string
	nodeAddName  string
	nodeAddDesc  string
	nodeAddEntry bool
	nodeAddRole      string
	nodeAddRules     string
	nodeAddStatus    string
	nodeAddAfter     string
	nodeAddCondition string
	nodeEditID     string
	nodeEditName   string
	nodeEditDesc   string
	nodeEditType   string
	nodeEditRole   string
	nodeEditRules  string
	nodeEditStatus string
	nodeRemoveID   string

	// edge subcommand
	edgeAddID         string
	edgeAddFrom       string
	edgeAddTo         string
	edgeAddCondition  string
	edgeAddType       string
	edgeAddSubflowRef string
	edgeEditID        string
	edgeEditCondition string
	edgeRemoveID      string

	// rule subcommand
	ruleAddID          string
	ruleAddName        string
	ruleAddInstruction string
	ruleAddDescription string
	ruleAddEnforcement string
	ruleAddType        string
	ruleAddTarget      string
	ruleEditName        string
	ruleEditInstruction string
	ruleEditDescription string
	ruleEditEnforcement string
	ruleEditType        string
	ruleEditTarget      string
	ruleRemoveTarget    string
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
	Long:  `List all available processes from preset (assets/flows/) and project (.team/flows/) directories.`,
	RunE:  runList,
}

var showCmd = &cobra.Command{
	Use:   "show <flow-name> [node-id]",
	Short: "Show process definition details",
	Long:  `Show process definition details including metadata, node count, and edge count.
With node-id, shows detailed information about a specific node within the flow.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runShow,
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

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a process",
	Long:  `Delete a flow from .team/flows/.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

var ruleCmd = &cobra.Command{
	Use:   "rule <rule-id>",
	Short: "Show full rule definition",
	Long:  `Show the full definition of a rule by its ID, including description and enforcement details.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRule,
}

var ruleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new rule with auto-generated ID",
	Long: `Create a new rule with an auto-generated random ID.
Content (name, instruction, description, enforcement, type) can be set via -- flags or filled later using 'rule edit'.
Adds to team.json by default. Use --target flow to add to the flow JSON instead.

Examples:
  flow proc rule add                                        # Empty shell, random ID, added to team
  flow proc rule add --name "My Rule" --enforcement hard    # With initial values
  flow proc rule add --target flow --flow dev-flow          # Add to flow components.rules`,
	Args: cobra.NoArgs,
	RunE: runRuleAdd,
}

var ruleEditCmd = &cobra.Command{
	Use:   "edit <rule-id>",
	Short: "Edit an existing rule's content",
	Long: `Edit a rule's name, instruction, description, enforcement, or type.
Only the specified -- flags are updated; others remain unchanged.

Examples:
  flow proc rule edit r_a1b2c3 --name "My Rule" --instruction "Do X then Y"
  flow proc rule edit r_a1b2c3 --type process_constraint --enforcement hard
  flow proc rule edit r_a1b2c3 --target flow --flow dev-flow`,
	Args: cobra.ExactArgs(1),
	RunE: runRuleEdit,
}

var ruleRemoveCmd = &cobra.Command{
	Use:   "remove <rule-id>",
	Short: "Remove a rule",
	Long: `Remove a rule from team.json or flow components.rules.

Examples:
  flow proc rule remove r_a1b2c3
  flow proc rule remove r_a1b2c3 --target flow --flow dev-flow`,
	Args: cobra.ExactArgs(1),
	RunE: runRuleRemove,
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

var gatePassCmd = &cobra.Command{
	Use:   "pass <node-id>",
	Short: "Mark an AI judgment gate condition as passed",
	Long: `Mark an AI judgment gate condition as passed for a specific gate node.

This is used when the AI has verified a condition that cannot be automatically checked
(e.g., context_sufficient, type_identified, project_context_verified).

After marking a condition as passed, re-run 'flow proc run <node-id>' to re-evaluate the gate.

Examples:
  flow gate pass ent7 --condition context_sufficient
  flow gate pass ent7 --condition context_sufficient --message "Context verified: user provided clear requirements"
  flow gate pass ent7 --all`,
	Args: cobra.ExactArgs(1),
	RunE: runGatePass,
}

var nextCmd = &cobra.Command{
	Use:   "next [node-id]",
	Short: "Advance to the next node in the flow",
	Long: `Advance to the next node in the current flow.

This is the RECOMMENDED way to progress through the flow. It handles:
  - Auto-resuming from last position (no need to specify node-id)
  - Auto-passing AI judgment conditions at gate nodes
  - Auto-advancing past passed gates to the next work node
  - When called without node-id and already at a work node, advances to the next node

Instead of manually running: flow proc run → flow gate pass → flow proc run
Just run: flow proc next

This single command replaces the multi-step gate traversal pattern.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNext,
}

var roundPathCmd = &cobra.Command{
	Use:   "round-path",
	Short: "Output the next round analysis directory path",
	Long: `Output the path where AI should write analysis.md and conclusion.md files
for the next round. The directory is created if it does not exist.

Expected files:
  <output>/analysis.md   — AI writes: Root Cause, Evidence, Solution, Trade-offs, Verification
  <output>/conclusion.md — AI writes: Decision, Next Action, Blockers

Usage:
  flow proc round-path              # Output next round directory path
  flow proc round-path --json       # Output as JSON with analysis_dir field`,
	Args: cobra.NoArgs,
	RunE: runRoundPath,
}

var autoAdvanceCmd = &cobra.Command{
	Use:   "auto-advance [node-id]",
	Short: "Auto-advance by reading analysis/conclusion from round files",
	Long: `Automatically read analysis.md and conclusion.md from the current round directory
and advance to the specified node. This command eliminates the need to manually specify
--analysis-file and --conclusion-file flags.

Usage:
  flow proc auto-advance <node-id>      # Auto-read round files and advance
  flow proc auto-advance                # Show help`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAutoAdvance,
}

var (
	gatePassCondition string
	gatePassMessage   string
	gatePassAll       bool
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage nodes in a flow",
	Long:  `Add, edit, or remove nodes in a flow definition.`,
}

var nodeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a node to a flow",
	Long: `Add a new node to a flow definition. Node ID is auto-generated if not provided.

Examples:
  flow proc node add --type start --name "Session Start"
  flow proc node add --type gate --name "Quality Gate" --flow dev-flow
  flow proc node add --id a1b2c3 --type phase --name "Custom Node"`,
	RunE: runNodeAdd,
}

var nodeEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a node in a flow",
	Long: `Edit an existing node's properties.

Examples:
  flow proc node edit --id sta0 --name "Session Start v2"
  flow proc node edit --id sta0 --desc "Updated description" --flow dev-flow`,
	RunE: runNodeEdit,
}

var nodeRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a node from a flow",
	Long: `Remove a node and all its associated edges from a flow definition.

Examples:
  flow proc node remove --id sta0
  flow proc node remove --id sta0 --flow dev-flow`,
	RunE: runNodeRemove,
}

var edgeCmd = &cobra.Command{
	Use:   "edge",
	Short: "Manage edges in a flow",
	Long:  `Add, edit, or remove edges in a flow definition.`,
}

var edgeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an edge to a flow",
	Long: `Add a new edge between two nodes. Edge ID is auto-generated if not provided.

Examples:
  flow proc edge add --from a1b2c3 --to d4e5f6
  flow proc edge add --from a1b2c3 --to d4e5f6 --condition "status=task_created" --type conditional
  flow proc edge add --id e_a1b2c3 --from a1b2c3 --to d4e5f6 --flow dev-flow`,
	RunE: runEdgeAdd,
}

var edgeEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an edge in a flow",
	Long: `Edit an existing edge's properties.

Examples:
  flow proc edge edit --id e_a1b2c3 --condition "status=task_created OR status=redirect"
  flow proc edge edit --id e_a1b2c3 --flow dev-flow`,
	RunE: runEdgeEdit,
}

var edgeRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an edge from a flow",
	Long: `Remove an edge from a flow definition.

Examples:
  flow proc edge remove --id e_a1b2c3
  flow proc edge remove --id e_a1b2c3 --flow dev-flow`,
	RunE: runEdgeRemove,
}

func init() {
	Cmd.AddCommand(runCmd)
	Cmd.AddCommand(nextCmd)
	Cmd.AddCommand(roundPathCmd)
	Cmd.AddCommand(autoAdvanceCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(validateCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(ruleCmd)
	ruleCmd.AddCommand(ruleAddCmd)
	ruleCmd.AddCommand(ruleEditCmd)
	ruleCmd.AddCommand(ruleRemoveCmd)
	Cmd.AddCommand(gateCmd)
	gateCmd.AddCommand(gatePassCmd)

	nodeCmd.AddCommand(nodeAddCmd)
	nodeCmd.AddCommand(nodeEditCmd)
	nodeCmd.AddCommand(nodeRemoveCmd)
	edgeCmd.AddCommand(edgeAddCmd)
	edgeCmd.AddCommand(edgeEditCmd)
	edgeCmd.AddCommand(edgeRemoveCmd)

	Cmd.AddCommand(nodeCmd)
	Cmd.AddCommand(edgeCmd)

	Cmd.PersistentFlags().StringVar(&procRootDir, "root", "", "Root directory for preset processes (default: current directory)")
	runCmd.Flags().StringVar(&procFormat, "format", "json", "Output format: json or text")
	runCmd.Flags().StringVar(&procTaskID, "task", "", "Task ID for variable substitution")
	runCmd.Flags().StringVar(&procInput, "input", "", "User input for this round (recorded in context.md)")
	runCmd.Flags().StringVar(&procAnalysis, "analysis", "", "AI analysis section (Root Cause/Evidence/Solution/Trade-offs) for this round")
	runCmd.Flags().StringVar(&procAnalysisFile, "analysis-file", "", "Path to file containing detailed AI analysis (read from file, overrides --analysis)")
	runCmd.Flags().StringVar(&procConclusion, "conclusion", "", "AI conclusion section (Decision/Next Action/Blockers) for this round")
	runCmd.Flags().StringVar(&procConclusionFile, "conclusion-file", "", "Path to file containing detailed AI conclusion (read from file, overrides --conclusion)")
	runCmd.Flags().BoolVar(&procRunGate, "gate", true, "Run automated gate checks when encountering a gate node")
	runCmd.Flags().BoolVar(&procNewSession, "new", false, "Create a new session (only on first call of a conversation)")
	ruleCmd.Flags().StringVar(&procFormat, "format", "text", "Output format: json or text")
	gateCmd.Flags().StringVar(&procFormat, "format", "text", "Output format: json or text")
	gatePassCmd.Flags().StringVar(&gatePassCondition, "condition", "", "Condition type to mark as passed (e.g., context_sufficient, type_identified)")
	gatePassCmd.Flags().StringVar(&gatePassMessage, "message", "", "Optional message explaining the AI judgment")
	gatePassCmd.Flags().BoolVar(&gatePassAll, "all", false, "Mark all AI judgment conditions as passed for this gate node")
	nextCmd.Flags().StringVar(&procFormat, "format", "json", "Output format: json or text")
	nextCmd.Flags().StringVar(&procTaskID, "task", "", "Task ID for variable substitution")
	nextCmd.Flags().StringVar(&procInput, "input", "", "User input for this round (recorded in context.md)")
	nextCmd.Flags().StringVar(&procAnalysis, "analysis", "", "AI analysis section (Root Cause/Evidence/Solution/Trade-offs) for this round")
	nextCmd.Flags().StringVar(&procAnalysisFile, "analysis-file", "", "Path to file containing detailed AI analysis (read from file, overrides --analysis)")
	nextCmd.Flags().StringVar(&procConclusion, "conclusion", "", "AI conclusion section (Decision/Next Action/Blockers) for this round")
	nextCmd.Flags().StringVar(&procConclusionFile, "conclusion-file", "", "Path to file containing detailed AI conclusion (read from file, overrides --conclusion)")
	roundPathCmd.Flags().StringVar(&procFormat, "format", "json", "Output format: json or text")
	createCmd.Flags().StringVar(&procTemplate, "template", "", "Template process name from assets/flows/ to base the new process on")

	// node flags
	nodeAddCmd.Flags().StringVar(&nodeAddID, "id", "", "Node ID (auto-generated if not provided)")
	nodeAddCmd.Flags().StringVar(&nodeAddType, "type", "", "Node type: start, phase, gate, branch, terminal, etc.")
	nodeAddCmd.Flags().StringVar(&nodeAddName, "name", "", "Node display name")
	nodeAddCmd.Flags().StringVar(&nodeAddDesc, "desc", "", "Node description")
	nodeAddCmd.Flags().BoolVar(&nodeAddEntry, "entry", false, "Set as entry node")
	nodeAddCmd.Flags().StringVar(&nodeAddRole, "role", "", "Node role ID (sets Components.Roles and Config.role)")
	nodeAddCmd.Flags().StringVar(&nodeAddRules, "rules", "", "Comma-separated rule IDs (sets Components.Rules)")
	nodeAddCmd.Flags().StringVar(&nodeAddStatus, "status", "", "Terminal status: success, failure, rejected")
	nodeAddCmd.Flags().StringVar(&nodeAddAfter, "after", "", "Source node ID — auto-create edge from this node to the new node")
	nodeAddCmd.Flags().StringVar(&nodeAddCondition, "condition", "", "Condition expression for the auto-created edge (use with --after)")
	nodeAddCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name")

	nodeEditCmd.Flags().StringVar(&nodeEditID, "id", "", "Node ID to edit")
	nodeEditCmd.Flags().StringVar(&nodeEditName, "name", "", "New display name")
	nodeEditCmd.Flags().StringVar(&nodeEditDesc, "desc", "", "New description")
	nodeEditCmd.Flags().StringVar(&nodeEditType, "type", "", "New node type")
	nodeEditCmd.Flags().StringVar(&nodeEditRole, "role", "", "New role ID")
	nodeEditCmd.Flags().StringVar(&nodeEditRules, "rules", "", "New comma-separated rule IDs")
	nodeEditCmd.Flags().StringVar(&nodeEditStatus, "status", "", "New terminal status")
	nodeEditCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name")

	nodeRemoveCmd.Flags().StringVar(&nodeRemoveID, "id", "", "Node ID to remove")
	nodeRemoveCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name")

	// edge flags
	edgeAddCmd.Flags().StringVar(&edgeAddID, "id", "", "Edge ID (auto-generated if not provided)")
	edgeAddCmd.Flags().StringVar(&edgeAddFrom, "from", "", "Source node ID")
	edgeAddCmd.Flags().StringVar(&edgeAddTo, "to", "", "Target node ID")
	edgeAddCmd.Flags().StringVar(&edgeAddCondition, "condition", "", "Condition expression for conditional edges")
	edgeAddCmd.Flags().StringVar(&edgeAddType, "type", "", "Edge type: sequential or conditional")
	edgeAddCmd.Flags().StringVar(&edgeAddSubflowRef, "subflow-ref", "", "Subflow reference — set for dispatch edges (replaces --to)")
	edgeAddCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name")

	edgeEditCmd.Flags().StringVar(&edgeEditID, "id", "", "Edge ID to edit")
	edgeEditCmd.Flags().StringVar(&edgeEditCondition, "condition", "", "New condition expression")
	edgeEditCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name")

	edgeRemoveCmd.Flags().StringVar(&edgeRemoveID, "id", "", "Edge ID to remove")
	edgeRemoveCmd.Flags().StringVar(&edgeAddFrom, "from", "", "Source node (alternative to --id)")
	edgeRemoveCmd.Flags().StringVar(&edgeAddTo, "to", "", "Target node (alternative to --id)")
	edgeRemoveCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name")

	// rule flags
	ruleAddCmd.Flags().StringVar(&ruleAddID, "id", "", "Rule ID (auto-generated if not provided)")
	ruleAddCmd.Flags().StringVar(&ruleAddName, "name", "", "Rule display name")
	ruleAddCmd.Flags().StringVar(&ruleAddInstruction, "instruction", "", "Short instruction text")
	ruleAddCmd.Flags().StringVar(&ruleAddDescription, "description", "", "Full description")
	ruleAddCmd.Flags().StringVar(&ruleAddEnforcement, "enforcement", "", "Enforcement level: hard or soft")
	ruleAddCmd.Flags().StringVar(&ruleAddType, "type", "", "Rule type: process_constraint, behavioral_constraint, output_constraint, quality_constraint")
	ruleAddCmd.Flags().StringVar(&ruleAddTarget, "target", "team", "Target: team (team.json) or flow (flow components.rules)")
	ruleAddCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name (required when --target=flow)")

	ruleEditCmd.Flags().StringVar(&ruleEditName, "name", "", "New display name")
	ruleEditCmd.Flags().StringVar(&ruleEditInstruction, "instruction", "", "New instruction text")
	ruleEditCmd.Flags().StringVar(&ruleEditDescription, "description", "", "New full description")
	ruleEditCmd.Flags().StringVar(&ruleEditEnforcement, "enforcement", "", "New enforcement level: hard or soft")
	ruleEditCmd.Flags().StringVar(&ruleEditType, "type", "", "New rule type")
	ruleEditCmd.Flags().StringVar(&ruleEditTarget, "target", "team", "Target: team or flow")
	ruleEditCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name (required when --target=flow)")

	ruleRemoveCmd.Flags().StringVar(&ruleRemoveTarget, "target", "team", "Target: team or flow")
	ruleRemoveCmd.Flags().StringVar(&procFlowName, "flow", "", "Flow name (required when --target=flow)")
}

// getProjectRootDir 获取项目根目录，优先使用 workspace 配置的 active_project
func getProjectRootDir() string {
	if procRootDir != "" {
		return procRootDir
	}
	dir, _ := os.Getwd()
	return config.ResolveProjectRootWithActive(dir)
}

// getRootDir 获取项目根目录（alias for getProjectRootDir）
func getRootDir() string {
	return getProjectRootDir()
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
	projectRoot := getProjectRootDir()
	workspace := getWorkspaceRoot()

	if workspace != "" && projectRoot == workspace {
		if config.IsMonorepoWorkspace(workspace) {
			projects := config.DiscoverProjects(workspace)
			if len(projects) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "⛔ WORKSPACE ROOT DETECTED")
				fmt.Fprintln(cmd.OutOrStdout(), "  Current directory is a monorepo workspace, not a project.")
				fmt.Fprintln(cmd.OutOrStdout(), "  Development MUST happen in a project directory.")
			fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), "  Available projects:")
				for i, p := range projects {
					fmt.Fprintf(cmd.OutOrStdout(), "    %d. %-20s → %s\n", i+1, p.Name, p.RelPath)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "\n  ACTION: cd projects/{project-name}/ && flow proc run")
				return fmt.Errorf("workspace root is not a project directory")
			}
		}
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

	// Read analysis/conclusion from files ONLY - inline flags are NOT allowed
	// Standard workflow: AI writes analysis.md and conclusion.md to round directory,
	// then flow proc run reads from those files via --analysis-file and --conclusion-file
	analysis := ""
	if procAnalysisFile != "" {
		data, err := os.ReadFile(procAnalysisFile)
		if err != nil {
			return fmt.Errorf("read analysis file %s: %w", procAnalysisFile, err)
		}
		analysis = string(data)
	}
	conclusion := ""
	if procConclusionFile != "" {
		data, err := os.ReadFile(procConclusionFile)
		if err != nil {
			return fmt.Errorf("read conclusion file %s: %w", procConclusionFile, err)
		}
		conclusion = string(data)
	}

	// Mandatory validation: analysis and conclusion files are required when advancing
	// Exempt --new (new session creation) and rescue mode (no node-id provided)
	if !procNewSession && nodeID != "" {
		if procAnalysisFile == "" {
			return fmt.Errorf("--analysis-file is required when advancing nodes. Use 'flow proc round-path' to get the directory, write analysis.md and conclusion.md, then use --analysis-file and --conclusion-file")
		}
		if procConclusionFile == "" {
			return fmt.Errorf("--conclusion-file is required when advancing nodes. Use 'flow proc round-path' to get the directory, write analysis.md and conclusion.md, then use --analysis-file and --conclusion-file")
		}
		if analysis == "" {
			return fmt.Errorf("analysis file %s is empty", procAnalysisFile)
		}
		if conclusion == "" {
			return fmt.Errorf("conclusion file %s is empty", procConclusionFile)
		}
	}

	// Restore task ID from session when --task is not explicitly provided
	if procTaskID == "" && !procNewSession {
		if lgr, err := eventlog.NewLogger(projectRoot); err == nil {
			procTaskID = lgr.LastTask()
		}
	}

	req := ProcRunRequest{
		FlowName:    procFlowName,
		NodeID:      nodeID,
		TaskID:      procTaskID,
		ProjectRoot: projectRoot,
		Workspace:   workspace,
		Format:      procFormat,
		RunGate:     procRunGate,
		Input:       procInput,
		Analysis:    analysis,
		Conclusion:  conclusion,
		NewSession:  procNewSession,
	}

	engine := NewProcRunEngine(projectRoot)
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

func autoCreateTask(projectRoot string) (string, error) {
	projectName := filepath.Base(projectRoot)
	taskID := task.GenerateLocalTaskID(projectName)

	t := &task.Task{
		ID:          taskID,
		Title:       "Auto-created task for new session",
		Type:        "auto-task",
		Status:      "open",
		Description: "",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Labels:      map[string]string{},
		Notes:       []string{},
	}

	if err := task.SaveTask(projectRoot, t); err != nil {
		return "", err
	}

	if lgr, err := eventlog.NewLogger(projectRoot); err == nil {
		sessionName, _ := lgr.LastSessionName()
		_ = lgr.TaskCreated(sessionName, taskID, "auto-task", "Auto-created task for new session")
	}

	return taskID, nil
}

func runNext(cmd *cobra.Command, args []string) error {
	projectRoot := getProjectRootDir()
	workspace := getWorkspaceRoot()

	if workspace != "" && projectRoot == workspace {
		if config.IsMonorepoWorkspace(workspace) {
			projects := config.DiscoverProjects(workspace)
			if len(projects) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "⛔ WORKSPACE ROOT DETECTED")
				fmt.Fprintln(cmd.OutOrStdout(), "  Development MUST happen in a project directory.")
				fmt.Fprintln(cmd.OutOrStdout(), "\n  Available projects:")
				for i, p := range projects {
					fmt.Fprintf(cmd.OutOrStdout(), "    %d. %-20s → %s\n", i+1, p.Name, p.RelPath)
				}
				return fmt.Errorf("workspace root is not a project directory")
			}
		}
	}

	nodeID := ""
	if len(args) > 0 {
		nodeID = args[0]
	}

	// Read analysis/conclusion from files ONLY - inline flags are NOT allowed
	// Standard workflow: AI writes analysis.md and conclusion.md to round directory,
	// then flow proc run reads from those files via --analysis-file and --conclusion-file
	analysis := ""
	if procAnalysisFile != "" {
		data, err := os.ReadFile(procAnalysisFile)
		if err != nil {
			return fmt.Errorf("read analysis file %s: %w", procAnalysisFile, err)
		}
		analysis = string(data)
	}
	conclusion := ""
	if procConclusionFile != "" {
		data, err := os.ReadFile(procConclusionFile)
		if err != nil {
			return fmt.Errorf("read conclusion file %s: %w", procConclusionFile, err)
		}
		conclusion = string(data)
	}

	// Mandatory validation: analysis and conclusion files are required when advancing
	if procAnalysisFile == "" {
		return fmt.Errorf("--analysis-file is required when advancing nodes. Use 'flow proc round-path' to get the directory, write analysis.md and conclusion.md, then use --analysis-file and --conclusion-file")
	}
	if procConclusionFile == "" {
		return fmt.Errorf("--conclusion-file is required when advancing nodes. Use 'flow proc round-path' to get the directory, write analysis.md and conclusion.md, then use --analysis-file and --conclusion-file")
	}
	if analysis == "" {
		return fmt.Errorf("analysis file %s is empty", procAnalysisFile)
	}
	if conclusion == "" {
		return fmt.Errorf("conclusion file %s is empty", procConclusionFile)
	}

	req := ProcRunRequest{
		FlowName:    procFlowName,
		NodeID:      nodeID,
		TaskID:      procTaskID,
		ProjectRoot: projectRoot,
		Workspace:   workspace,
		Format:      procFormat,
		RunGate:     true,
		Input:       procInput,
		Analysis:    analysis,
		Conclusion:  conclusion,
		NewSession:  procNewSession,
	}

	engine := NewProcRunEngine(projectRoot)
	tracer := engine.Trace
	result, err := engine.Run(context.Background(), req)
	if err != nil {
		return err
	}

	maxAutoAdvance := 20
	for i := 0; i < maxAutoAdvance; i++ {
		if result.Current.NodeType == "gate" {
			passed := true
			if len(result.GateCheckResults) > 0 {
				for _, cond := range result.GateCheckResults {
					if !cond.Passed && cond.Required && !cond.Skipped {
						passed = false
						break
					}
				}
			}
			if passed && len(result.NextOptions) > 0 {
				defaultNextNodeID := ""
				for _, opt := range result.NextOptions {
					if opt.IsDefault {
						defaultNextNodeID = opt.NodeID
						break
					}
				}
				if defaultNextNodeID == "" && len(result.NextOptions) > 0 {
					defaultNextNodeID = result.NextOptions[0].NodeID
				}
				if defaultNextNodeID == "" {
					break
				}
				nextReq := req
				nextReq.NodeID = defaultNextNodeID
				nextResult, nextErr := engine.Run(context.Background(), nextReq)
				if nextErr != nil || nextResult == nil {
					break
				}
				nextResult.ResumedFrom = "gate-auto-advance"
				if tracer != nil && result.SessionName != "" {
					tracer.EdgeTraverse(result.SessionName, req.TaskID, "",
						result.Current.NodeID, nextResult.Current.NodeID,
						"gate_auto_pass", "true", result.Flow.Name)
				}
				result = nextResult
				continue
			}
			// Gate not passed or no next options — stop advancing and report the gate result
			if !passed {
				result.ResumedFrom = "gate-blocked"
			}
			break
		}

		if (result.Current.NodeType == "start" || result.Current.NodeType != "gate") && len(result.NextOptions) > 0 {
			defaultNextNodeID := ""
			for _, opt := range result.NextOptions {
				if opt.IsDefault {
					defaultNextNodeID = opt.NodeID
					break
				}
			}
			if defaultNextNodeID == "" {
				defaultNextNodeID = result.NextOptions[0].NodeID
			}
			if defaultNextNodeID == "" {
				break
			}
			nextReq := req
			nextReq.NodeID = defaultNextNodeID
			nextResult, nextErr := engine.Run(context.Background(), nextReq)
			if nextErr != nil || nextResult == nil {
				result.ResumedFrom = "state"
				break
			}
			nextResult.ResumedFrom = "auto-advance"
			if tracer != nil && result.SessionName != "" {
				tracer.EdgeTraverse(result.SessionName, req.TaskID, "",
					result.Current.NodeID, nextResult.Current.NodeID,
					"default", "auto", result.Flow.Name)
			}
			result = nextResult
			continue
		}

		if result.Current.NodeType != "gate" && result.Current.NodeType != "start" {
			break
		}
	}

	switch procFormat {
	case "text":
		return FormatText(cmd.OutOrStdout(), result)
	default:
		return FormatJSON(cmd.OutOrStdout(), result)
	}
}

func runAutoAdvance(cmd *cobra.Command, args []string) error {
	projectRoot := getProjectRootDir()
	workspace := getWorkspaceRoot()

	if workspace != "" && projectRoot == workspace {
		if config.IsMonorepoWorkspace(workspace) {
			projects := config.DiscoverProjects(workspace)
			if len(projects) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "⛔ WORKSPACE ROOT DETECTED")
				fmt.Fprintln(cmd.OutOrStdout(), "  Current directory is a monorepo workspace, not a project.")
				return fmt.Errorf("workspace root is not a project directory")
			}
		}
	}

	nodeID := ""
	if len(args) > 0 {
		nodeID = args[0]
	}
	if nodeID == "" {
		return fmt.Errorf("node-id is required. Usage: flow proc auto-advance <node-id>")
	}

	lgr, err := eventlog.NewLogger(projectRoot)
	if err != nil {
		return fmt.Errorf("eventlog init: %w", err)
	}

	sessionName, err := lgr.LastSessionName()
	if err != nil {
		return fmt.Errorf("no active session: %w", err)
	}

	dir, round, err := lgr.NextRoundDir(sessionName)
	if err != nil {
		return fmt.Errorf("get round dir: %w", err)
	}

	analysisPath := filepath.Join(dir, "analysis.md")
	conclusionPath := filepath.Join(dir, "conclusion.md")

	if _, err := os.Stat(analysisPath); os.IsNotExist(err) {
		return fmt.Errorf("analysis file not found: %s. Run 'flow proc round-path' to get the directory, write analysis.md and conclusion.md, then try again", analysisPath)
	}
	if _, err := os.Stat(conclusionPath); os.IsNotExist(err) {
		return fmt.Errorf("conclusion file not found: %s. Run 'flow proc round-path' to get the directory, write analysis.md and conclusion.md, then try again", conclusionPath)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "  Reading round %d files:\n", round)
	fmt.Fprintf(cmd.ErrOrStderr(), "    analysis: %s\n", analysisPath)
	fmt.Fprintf(cmd.ErrOrStderr(), "    conclusion: %s\n", conclusionPath)

	analysis, err := os.ReadFile(analysisPath)
	if err != nil {
		return fmt.Errorf("read analysis file: %w", err)
	}
	conclusion, err := os.ReadFile(conclusionPath)
	if err != nil {
		return fmt.Errorf("read conclusion file: %w", err)
	}

	req := ProcRunRequest{
		FlowName:    procFlowName,
		NodeID:      nodeID,
		TaskID:      procTaskID,
		ProjectRoot: projectRoot,
		Workspace:   workspace,
		Format:      procFormat,
		RunGate:     procRunGate,
		Input:       procInput,
		Analysis:    string(analysis),
		Conclusion:  string(conclusion),
		NewSession:  false,
	}

	engine := NewProcRunEngine(projectRoot)
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

// runRoundPath outputs the path where AI should write per-round analysis files.
func runRoundPath(cmd *cobra.Command, args []string) error {
	projectRoot := getProjectRootDir()
	if projectRoot == "" {
		return fmt.Errorf("not in a project directory")
	}

	lgr, err := eventlog.NewLogger(projectRoot)
	if err != nil {
		return fmt.Errorf("eventlog init: %w", err)
	}

	sessionName, err := lgr.LastSessionName()
	if err != nil {
		return fmt.Errorf("no active session: %w", err)
	}

	dir, round, err := lgr.NextRoundDir(sessionName)
	if err != nil {
		return fmt.Errorf("create round dir: %w", err)
	}

	analysisPath := filepath.Join(dir, "analysis.md")
	conclusionPath := filepath.Join(dir, "conclusion.md")

	if procFormat == "json" {
		data, err := json.MarshalIndent(struct {
			AnalysisDir      string `json:"analysis_dir"`
			AnalysisFile     string `json:"analysis_file"`
			ConclusionFile   string `json:"conclusion_file"`
			Round            int    `json:"round"`
			Session          string `json:"session"`
		}{AnalysisDir: dir, AnalysisFile: analysisPath, ConclusionFile: conclusionPath, Round: round, Session: sessionName}, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal round-path: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Round %d Analysis Directory:\n", round)
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", dir)
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "AI must write these files:")
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", analysisPath)
	fmt.Fprintf(cmd.OutOrStdout(), "    → Content: Root Cause, Evidence, Solution, Trade-offs, Verification\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", conclusionPath)
	fmt.Fprintf(cmd.OutOrStdout(), "    → Content: Decision, Next Action, Blockers\n")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "After writing files, advance flow with:")
	fmt.Fprintf(cmd.OutOrStdout(), "  flow proc run <node-id> --analysis-file %s --conclusion-file %s\n", analysisPath, conclusionPath)
	return nil
}

type ProcEntry struct {
	Name        string
	ID          string
	Path        string
	Source      string
	Description string
	IsActive    bool
	Registered  bool
}

func listProcs(root string) []ProcEntry {
	var entries []ProcEntry

	regMap := readFlowRegistry(root)
	activeFlow := readActiveFlow(root)

	// Read metadata.id from each flow JSON
	getFlowID := func(flowPath string) string {
		f, err := flow.ParseFlowFile(flowPath)
		if err != nil {
			return ""
		}
		return f.Metadata.ID
	}

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
				ID:          getFlowID(f),
				Path:        f,
				Source:      "preset",
				Description: desc,
				IsActive:    name == activeFlow,
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
				ID:          getFlowID(f),
				Path:        f,
				Source:      "project",
				Description: desc,
				IsActive:    name == activeFlow,
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

func readActiveFlow(root string) string {
	projectMd := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMd)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "active_flow:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "active_flow:"))
		}
	}
	return ""
}

func printProcs(w io.Writer, entries []ProcEntry) {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No processes found.")
		return
	}

	fmt.Fprintf(w, "%-25s %-8s %-10s %-6s %-10s %s\n", "NAME", "ID", "SOURCE", "REG", "CURRENT", "DESCRIPTION")
	fmt.Fprintf(w, "%-25s %-8s %-10s %-6s %-10s %s\n", strings.Repeat("-", 25), strings.Repeat("-", 8), strings.Repeat("-", 10), strings.Repeat("-", 6), strings.Repeat("-", 10), strings.Repeat("-", 30))
	for _, e := range entries {
		regMark := "  "
		if e.Registered {
			regMark = "✓ "
		}
		activeMark := ""
		if e.IsActive {
			activeMark = "⭐"
		}
		desc := e.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		fmt.Fprintf(w, "%-25s %-8s %-10s %-6s %-10s %s\n", e.Name, e.ID, e.Source, regMark, activeMark, desc)
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
	if f.Metadata.ID != "" {
		fmt.Fprintf(w, "ID:          %s\n", f.Metadata.ID)
	}
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
			fmt.Fprintf(w, "  %-20s %-12s %s\n", n.Name, n.Type, n.ID)
		}
	}

	if len(f.Edges) > 0 {
		fmt.Fprintln(w, "\nEdges:")
		for _, e := range f.Edges {
			label := e.From + " -> " + e.To
			if e.SubflowRef != "" {
				label = e.From + " -> subflow:" + e.SubflowRef
			}
			if e.Type != "" {
				label += " (" + string(e.Type) + ")"
			}
			fmt.Fprintf(w, "  %s\n", label)
		}
	}
}

func showNode(w io.Writer, n flow.FlowNode) {
	fmt.Fprintf(w, "ID:          %s\n", n.ID)
	fmt.Fprintf(w, "Type:        %s\n", n.Type)
	fmt.Fprintf(w, "Name:        %s\n", n.Name)
	if n.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", n.Description)
	}
	if n.Entry {
		fmt.Fprintf(w, "Entry:       true\n")
	}
	if n.Status != "" {
		fmt.Fprintf(w, "Status:      %s\n", n.Status)
	}

	if n.Components != nil {
		fmt.Fprintln(w)
		if len(n.Components.Roles) > 0 {
			fmt.Fprintln(w, "Roles:")
			for _, r := range n.Components.Roles {
				fmt.Fprintf(w, "  %s", r.Ref)
				if r.Source != "" {
					fmt.Fprintf(w, " (%s)", r.Source)
				}
				fmt.Fprintln(w)
			}
		}
		if len(n.Components.Rules) > 0 {
			fmt.Fprintln(w, "Rules:")
			for _, r := range n.Components.Rules {
				fmt.Fprintf(w, "  %s", r.Ref)
				if r.Source != "" {
					fmt.Fprintf(w, " (%s)", r.Source)
				}
				fmt.Fprintln(w)
			}
		}
		if len(n.Components.Tools) > 0 {
			fmt.Fprintln(w, "Tools:")
			for _, t := range n.Components.Tools {
				fmt.Fprintf(w, "  %s", t.Ref)
				if t.Source != "" {
					fmt.Fprintf(w, " (%s)", t.Source)
				}
				fmt.Fprintln(w)
			}
		}
		if len(n.Components.Skills) > 0 {
			fmt.Fprintln(w, "Skills:")
			for _, s := range n.Components.Skills {
				fmt.Fprintf(w, "  %s", s.Ref)
				if s.Source != "" {
					fmt.Fprintf(w, " (%s)", s.Source)
				}
				fmt.Fprintln(w)
			}
		}
	}

	if len(n.Docs) > 0 {
		fmt.Fprintln(w, "\nDocs:")
		for _, d := range n.Docs {
			required := "optional"
			if d.Required != nil && *d.Required {
				required = "required"
			}
			fmt.Fprintf(w, "  %s (%s, %s)\n", d.Name, d.Format, required)
		}
	}

	if len(n.Gates) > 0 {
		fmt.Fprintln(w, "\nGates:")
		for _, g := range n.Gates {
			fmt.Fprintf(w, "  type=%s", g.Type)
			if g.OnFail != "" {
				fmt.Fprintf(w, " on_fail=%s", g.OnFail)
			}
			fmt.Fprintln(w)
		}
	}

	if len(n.OnEnter) > 0 {
		fmt.Fprintln(w, "\nOn Enter:")
		for _, a := range n.OnEnter {
			fmt.Fprintf(w, "  action=%s", a.Action)
			if a.Phase != "" {
				fmt.Fprintf(w, " phase=%s", a.Phase)
			}
			fmt.Fprintln(w)
		}
	}

	if len(n.OnExit) > 0 {
		fmt.Fprintln(w, "\nOn Exit:")
		for _, a := range n.OnExit {
			fmt.Fprintf(w, "  action=%s", a.Action)
			if a.Phase != "" {
				fmt.Fprintf(w, " phase=%s", a.Phase)
			}
			fmt.Fprintln(w)
		}
	}
}

func runShow(cmd *cobra.Command, args []string) error {
	root := getRootDir()

	procPath := ResolveProcPath(root, args[0])
	if procPath == "" {
		activeFlow := ResolveActiveFlow(root)
		if activeFlow == "" {
			return fmt.Errorf("process not found: %s", args[0])
		}
		procPath = ResolveProcPath(root, activeFlow)
		if procPath == "" {
			return fmt.Errorf("process not found: %s", args[0])
		}
		f, err := flow.ParseFlowFile(procPath)
		if err != nil {
			return fmt.Errorf("parse process: %w", err)
		}
		nodeID := args[0]
		w := cmd.OutOrStdout()
		for _, n := range f.Nodes {
			if n.ID == nodeID {
				showNode(w, n)
				return nil
			}
		}
		return fmt.Errorf("node not found: %s in flow %s", nodeID, activeFlow)
	}

	f, err := flow.ParseFlowFile(procPath)
	if err != nil {
		return fmt.Errorf("parse process: %w", err)
	}

	w := cmd.OutOrStdout()

	if len(args) == 2 {
		nodeID := args[1]
		for _, n := range f.Nodes {
			if n.ID == nodeID {
				showNode(w, n)
				return nil
			}
		}
		return fmt.Errorf("node not found: %s in flow %s", nodeID, args[0])
	}

	showProc(w, f)
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
		templatePath := ResolveProcPath(root, tmpl)
		if templatePath == "" {
			return "", fmt.Errorf("template not found: %s", tmpl)
		}

		tmplFlow, err := flow.ParseFlowFile(templatePath)
		if err != nil {
			return "", fmt.Errorf("parse template: %w", err)
		}

		// Generate new metadata.id (even for template)
		tmplFlow.Metadata.ID = idgen.RandHex(2)

		if err := flow.SerializeFlowFile(tmplFlow, targetPath); err != nil {
			return "", fmt.Errorf("write process: %w", err)
		}

		return targetPath, nil
	}

	startID := idgen.RandHex(3)
	doneID := idgen.RandHex(3)

	newFlow := &flow.Flow{
		Version: "v3",
		Metadata: flow.FlowMetadata{
			ID:   idgen.RandHex(2),
			Name: name,
		},
		Nodes: []flow.FlowNode{
			{ID: startID, Type: flow.NodeTypePhase, Name: "Start"},
			{ID: doneID, Type: flow.NodeTypeTerminal, Name: "Done"},
		},
		Edges: []flow.FlowEdge{
			{From: startID, To: doneID},
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

	path, err := createProc(root, name, procTemplate)
	if err != nil {
		return err
	}

	// Read back to show metadata.id
	f, parseErr := flow.ParseFlowFile(path)
	flowID := ""
	if parseErr == nil {
		flowID = f.Metadata.ID
	}

	if procTemplate != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created process '%s' (id: %s) from template: %s\n", name, flowID, path)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Created process '%s' (id: %s): %s\n", name, flowID, path)
	}
	return nil
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	root := getRootDir()

	projectDir := filepath.Join(root, ".team", "flows")
	targetPath := filepath.Join(projectDir, name+".json")

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("process not found: %s", targetPath)
	}

	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("delete process: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted process: %s\n", name)
	return nil
}

func runRule(cmd *cobra.Command, args []string) error {
	ruleID := args[0]
	root := getRootDir()

	flowName := ResolveActiveFlow(root)

	procPath := ResolveProcPath(root, flowName)
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

func runRuleAdd(cmd *cobra.Command, args []string) error {
	root := getRootDir()

	id := ruleAddID
	if id == "" {
		id = generateRuleID()
	}

	targetPath, err := resolveRuleTargetPath(root, ruleAddTarget)
	if err != nil {
		return err
	}

	enforcement := flow.Enforcement(ruleAddEnforcement)
	if ruleAddEnforcement == "" {
		enforcement = ""
	}

	ruleType := flow.RuleType(ruleAddType)
	if ruleAddType == "" {
		ruleType = ""
	}

	rule := flow.RuleDefinition{
		ID:          id,
		Name:        ruleAddName,
		Instruction: ruleAddInstruction,
		Description: ruleAddDescription,
		Enforcement: enforcement,
		Type:        ruleType,
	}

	if err := appendRuleToTarget(targetPath, rule); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Rule '%s' added to %s\n", id, filepath.Base(targetPath))
	return nil
}

func runRuleEdit(cmd *cobra.Command, args []string) error {
	ruleID := args[0]
	root := getRootDir()

	targetPath, err := resolveRuleTargetPath(root, ruleEditTarget)
	if err != nil {
		return err
	}

	updated := flow.RuleDefinition{ID: ruleID}

	// Helper to check if a flag was explicitly set
	flagChanged := func(name string) bool {
		return cmd.Flags().Changed(name)
	}

	if flagChanged("name") {
		updated.Name = ruleEditName
	}
	if flagChanged("instruction") {
		updated.Instruction = ruleEditInstruction
	}
	if flagChanged("description") {
		updated.Description = ruleEditDescription
	}
	if flagChanged("enforcement") {
		updated.Enforcement = flow.Enforcement(ruleEditEnforcement)
	}
	if flagChanged("type") {
		updated.Type = flow.RuleType(ruleEditType)
	}

	if err := editRuleInTarget(targetPath, updated); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Rule '%s' updated in %s\n", ruleID, filepath.Base(targetPath))
	return nil
}

func runRuleRemove(cmd *cobra.Command, args []string) error {
	ruleID := args[0]
	root := getRootDir()

	targetPath, err := resolveRuleTargetPath(root, ruleRemoveTarget)
	if err != nil {
		return err
	}

	if err := removeRuleFromTarget(targetPath, ruleID); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Rule '%s' removed from %s\n", ruleID, filepath.Base(targetPath))
	return nil
}

func generateRuleID() string {
	return fmt.Sprintf("r_%s", idgen.RandHex(3))
}

func resolveRuleTargetPath(root, target string) (string, error) {
	switch target {
	case "team":
		return filepath.Join(root, ".team", "team.json"), nil
	case "flow":
		flowName := procFlowName
		if flowName == "" {
			flowName = ResolveActiveFlow(root)
		}
		return resolveFlowPath(root, flowName)
	default:
		return "", fmt.Errorf("unknown --target: %s (must be 'team' or 'flow')", target)
	}
}

func appendRuleToTarget(path string, rule flow.RuleDefinition) error {
	if filepath.Base(path) == "team.json" || strings.HasSuffix(path, "team.json") {
		team, err := loadTeamFile(path)
		if err != nil {
			return fmt.Errorf("load team: %w", err)
		}
		team.Rules = append(team.Rules, rule)
		return saveTeamFile(path, team)
	}
	return appendRuleToFlow(path, rule)
}

func editRuleInTarget(path string, updated flow.RuleDefinition) error {
	if filepath.Base(path) == "team.json" || strings.HasSuffix(path, "team.json") {
		team, err := loadTeamFile(path)
		if err != nil {
			return fmt.Errorf("load team: %w", err)
		}
		found := false
		for i, r := range team.Rules {
			if r.ID == updated.ID {
				if updated.Name != "" {
					team.Rules[i].Name = updated.Name
				}
				if updated.Instruction != "" {
					team.Rules[i].Instruction = updated.Instruction
				}
				if updated.Description != "" {
					team.Rules[i].Description = updated.Description
				}
				if updated.Enforcement != "" {
					team.Rules[i].Enforcement = updated.Enforcement
				}
				if updated.Type != "" {
					team.Rules[i].Type = updated.Type
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("rule '%s' not found in team", updated.ID)
		}
		return saveTeamFile(path, team)
	}
	return editRuleInFlow(path, updated)
}

func removeRuleFromTarget(path string, ruleID string) error {
	if filepath.Base(path) == "team.json" || strings.HasSuffix(path, "team.json") {
		team, err := loadTeamFile(path)
		if err != nil {
			return fmt.Errorf("load team: %w", err)
		}
		found := false
		for i, r := range team.Rules {
			if r.ID == ruleID {
				team.Rules = append(team.Rules[:i], team.Rules[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("rule '%s' not found in team", ruleID)
		}
		return saveTeamFile(path, team)
	}
	return removeRuleFromFlow(path, ruleID)
}

func loadTeamFile(path string) (*flow.TeamDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read team file %s: %w", path, err)
	}
	var team flow.TeamDefinition
	if err := json.Unmarshal(data, &team); err != nil {
		return nil, fmt.Errorf("parse team: %w", err)
	}
	if team.ID == "" {
		return nil, fmt.Errorf("team.json at %s has no id", path)
	}
	return &team, nil
}

func saveTeamFile(path string, team *flow.TeamDefinition) error {
	data, err := json.MarshalIndent(team, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal team: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write team: %w", err)
	}
	return nil
}

func appendRuleToFlow(path string, rule flow.RuleDefinition) error {
	return loadAndValidate(path, func(f *flow.Flow) {
		if f.Components == nil {
			f.Components = &flow.ComponentRegistry{}
		}
		f.Components.Rules = append(f.Components.Rules, rule)
	})
}

func editRuleInFlow(path string, updated flow.RuleDefinition) error {
	return loadAndValidate(path, func(f *flow.Flow) {
		if f.Components == nil {
			return
		}
		for i, r := range f.Components.Rules {
			if r.ID == updated.ID {
				if updated.Name != "" {
					f.Components.Rules[i].Name = updated.Name
				}
				if updated.Instruction != "" {
					f.Components.Rules[i].Instruction = updated.Instruction
				}
				if updated.Description != "" {
					f.Components.Rules[i].Description = updated.Description
				}
				if updated.Enforcement != "" {
					f.Components.Rules[i].Enforcement = updated.Enforcement
				}
				if updated.Type != "" {
					f.Components.Rules[i].Type = updated.Type
				}
				break
			}
		}
	})
}

func removeRuleFromFlow(path string, ruleID string) error {
	return loadAndValidate(path, func(f *flow.Flow) {
		if f.Components == nil {
			return
		}
		for i, r := range f.Components.Rules {
			if r.ID == ruleID {
				f.Components.Rules = append(f.Components.Rules[:i], f.Components.Rules[i+1:]...)
				break
			}
		}
	})
}

func ResolveProcPath(root, id string) string {
	// 1. Absolute path / explicit .json - bypass
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
	if idx := strings.Index(id, ":"); idx != -1 {
		id = id[idx+1:]
	}

	// 2. Tool-level template (read-only, embedded). Inherited verbatim by any
	//    team that doesn't override it via a "file" key in team.json. A team
	//    with a "file" key is a fork and gets resolved in step 3 instead.
	if templates.Has(id) {
		// Mark this id as a template; downstream code can detect template
		// flows via the side-band after they are parsed. We return a special
		// sentinel path that LoadTemplateFlow recognises.
		return templates.SentinelPath(id)
	}

	// 3. Disk lookup: project-level + team-level flow files
	// v3.2 directory layout: <root>/{v3,.team}/flows/<flow-id>/flow.json
	// v3.2 sub-flow:           <root>/{v3,.team}/flows/<main-id>/subs/<sub-id>/flow.json
	// v3.1 flat layout fallback: <root>/{v3,.team}/flows/<flow-id>.json
	// v3 content library:       <root>/assets/flows/<flow-id>.json
	candidates := []string{
		filepath.Join(root, "v3", "flows", id, "flow.json"),
		filepath.Join(root, ".team", "flows", id, "flow.json"),
		filepath.Join(root, "v3", "flows", id+".json"),
		filepath.Join(root, ".team", "flows", id+".json"),
		filepath.Join(root, "assets", "flows", id+".json"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	// v3.2: search sub-flows under all <main-flow>/subs/<id>/
	// (resolves via parent metadata; this is a best-effort flat scan)
	flowDirs := []string{
		filepath.Join(root, ".team", "flows"),
		filepath.Join(root, "v3", "flows"),
		filepath.Join(root, "assets", "flows"),
	}
	for _, d := range flowDirs {
		if !pathExists(d) {
			continue
		}
		entries, _ := os.ReadDir(d)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			sub := filepath.Join(d, e.Name(), "subs", id, "flow.json")
			if _, err := os.Stat(sub); err == nil {
				return sub
			}
		}
	}

	workspace := config.FindWorkspaceRoot(root)
	if workspace != "" && workspace != root {
		workspaceCandidates := []string{
			filepath.Join(workspace, ".team", "flows", id, "flow.json"),
			filepath.Join(workspace, "v3", "flows", id, "flow.json"),
			filepath.Join(workspace, ".team", "flows", id+".json"),
			filepath.Join(workspace, "v3", "flows", id+".json"),
		}
		for _, c := range workspaceCandidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
		// workspace sub-flow search
		for _, d := range []string{filepath.Join(workspace, ".team", "flows"), filepath.Join(workspace, "v3", "flows")} {
			if !pathExists(d) {
				continue
			}
			entries, _ := os.ReadDir(d)
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				sub := filepath.Join(d, e.Name(), "subs", id, "flow.json")
				if _, err := os.Stat(sub); err == nil {
					return sub
				}
			}
		}
	}

	return ""
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func runGate(cmd *cobra.Command, args []string) error {
	projectRoot := getRootDir()

	nodeID := ""
	if len(args) > 0 {
		nodeID = args[0]
	}

	flowName := ResolveActiveFlow(projectRoot)

	procPath := ResolveProcPath(projectRoot, flowName)
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
	sessionName := ""
	if lgr, err := eventlog.NewLogger(projectRoot); err == nil {
		sessionName, _ = lgr.LastSessionName()
	}
	results := RunGateCheckWithFlow(projectRoot, conditions, team, docsInternal, flowName, true, sessionName, "")
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
		printGateResults(cmd.OutOrStdout(), targetNode.Name, passed, summary, results)
	}

	if !passed {
		return fmt.Errorf("gate check failed: %s", summary)
	}

	return nil
}

func printGateResults(w io.Writer, nodeName string, passed bool, summary string, results []GateCheckResult) {
	icon := "✓"
	if !passed {
		icon = "✗"
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "╔══════════════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(w, "║  %s GATE CHECK: %s\n", icon, padLine(nodeName, 58))
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

func runGatePass(cmd *cobra.Command, args []string) error {
	nodeID := args[0]

	// Gate confirmation is now implicit: when AI explicitly targets a gate node
	// via 'flow proc run <node-id>', the gate is auto-confirmed.
	fmt.Fprintf(cmd.OutOrStdout(), "✓ To confirm gate %s, run: flow proc run %s\n", nodeID, nodeID)
	fmt.Fprintf(cmd.OutOrStdout(), "  (Gate is auto-confirmed when explicitly targeted.)\n")
	return nil
}

// resolveFlowPath resolves a flow name to its file path.
func resolveFlowPath(root, name string) (string, error) {
	if name == "" {
		name = ResolveActiveFlow(root)
	}
	p := ResolveProcPath(root, name)
	if p == "" {
		return "", fmt.Errorf("flow not found: %s", name)
	}
	return p, nil
}

// loadAndValidate loads a flow, applies a mutation, validates, and saves.
// Returns an error without saving if validation fails.
func loadAndValidate(path string, mutate func(f *flow.Flow)) error {
	f, err := flow.ParseFlowFile(path)
	if err != nil {
		return fmt.Errorf("parse flow: %w", err)
	}

	mutate(f)

	result := flow.ValidateFlow(f)
	if !result.Valid {
		lines := []string{"validation failed:"}
		for _, e := range result.Errors {
			lines = append(lines, fmt.Sprintf("  - %s", e.Message))
		}
		return fmt.Errorf("%s", strings.Join(lines, "\n"))
	}

	if err := flow.SerializeFlowFile(f, path); err != nil {
		return fmt.Errorf("save flow: %w", err)
	}

	return nil
}

// runNodeAdd adds a new node to a flow definition.
// When --after is set, auto-creates an edge from the given node to the new node.
func runNodeAdd(cmd *cobra.Command, args []string) error {
	if nodeAddType == "" {
		return fmt.Errorf("--type is required")
	}

	root := getRootDir()
	path, err := resolveFlowPath(root, procFlowName)
	if err != nil {
		return err
	}

	generatedID := ""
	if nodeAddID == "" {
		generatedID, err = generateNodeID(path, flow.NodeType(nodeAddType))
		if err != nil {
			return fmt.Errorf("generate node ID: %w", err)
		}
		nodeAddID = generatedID
	}

	name := nodeAddName
	if name == "" {
		name = nodeAddID
	}

	nt := flow.NodeType(nodeAddType)
	isEntry := nodeAddEntry

	autoEdgeFrom := nodeAddAfter
	autoEdgeCond := nodeAddCondition

	err = loadAndValidate(path, func(f *flow.Flow) {
		// Validate --after refers to an existing node
		if autoEdgeFrom != "" {
			found := false
			for _, n := range f.Nodes {
				if n.ID == autoEdgeFrom {
					found = true
					break
				}
			}
			if !found {
				return // handled outside loadAndValidate
			}
		}

		if isEntry {
			for i := range f.Nodes {
				f.Nodes[i].Entry = false
			}
		}

		node := flow.FlowNode{
			ID:          nodeAddID,
			Type:        nt,
			Name:        name,
			Description: nodeAddDesc,
			Entry:       isEntry,
			Status:      nodeAddStatus,
		}

		if nodeAddRole != "" {
			if node.Components == nil {
				node.Components = &flow.NodeComponents{}
			}
			node.Components.Roles = []flow.ComponentRef{{Ref: nodeAddRole, Source: flow.SourceBuiltin}}
			cfg := flow.PhaseConfig{Role: nodeAddRole}
			if cfgData, err := json.Marshal(cfg); err == nil {
				node.Config = cfgData
			}
		}

		if nodeAddRules != "" {
			if node.Components == nil {
				node.Components = &flow.NodeComponents{}
			}
			ruleIDs := strings.Split(nodeAddRules, ",")
			node.Components.Rules = make([]flow.ComponentRef, len(ruleIDs))
			for i, r := range ruleIDs {
				node.Components.Rules[i] = flow.ComponentRef{Ref: strings.TrimSpace(r), Source: flow.SourceBuiltin}
			}
		}

		f.Nodes = append(f.Nodes, node)

		// Auto-create edge when --after is set
		if autoEdgeFrom != "" {
			edgeID, err := generateEdgeID(path, autoEdgeFrom, nodeAddID)
			if err != nil {
				return
			}
			edgeType := flow.EdgeTypeSequential
			if autoEdgeCond != "" {
				edgeType = flow.EdgeTypeConditional
			}
			edge := flow.FlowEdge{
				ID:   edgeID,
				From: autoEdgeFrom,
				To:   nodeAddID,
				Type: edgeType,
			}
			if autoEdgeCond != "" {
				edge.Conditions = []flow.EdgeCondition{
					{Expression: autoEdgeCond},
				}
			}
			f.Edges = append(f.Edges, edge)
		}
	})
	if err != nil {
		return err
	}

	// Validate --after node exists (done outside loadAndValidate for proper error)
	if autoEdgeFrom != "" {
		f, _ := flow.ParseFlowFile(path)
		found := false
		if f != nil {
			for _, n := range f.Nodes {
				if n.ID == autoEdgeFrom {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("--after node not found: %s", autoEdgeFrom)
		}
	}

	if autoEdgeFrom != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Added node '%s' (type=%s) + edge %s→%s to %s\n", nodeAddID, nodeAddType, autoEdgeFrom, nodeAddID, filepath.Base(path))
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Added node '%s' (type=%s) to %s\n", nodeAddID, nodeAddType, filepath.Base(path))
	}
	return nil
}

func generateNodeID(flowPath string, nodeType flow.NodeType) (string, error) {
	f, err := flow.ParseFlowFile(flowPath)
	if err != nil {
		return "", err
	}

	existingIDs := make(map[string]bool)
	for _, n := range f.Nodes {
		existingIDs[n.ID] = true
	}

	for i := 0; i < 100; i++ {
		candidate := idgen.RandHex(3)
		if !existingIDs[candidate] {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not generate unique ID (exhausted 100 attempts)")
}

// runNodeEdit edits an existing node's properties.
func runNodeEdit(cmd *cobra.Command, args []string) error {
	if nodeEditID == "" {
		return fmt.Errorf("--id is required")
	}
	if nodeEditName == "" && nodeEditDesc == "" && nodeEditType == "" {
		return fmt.Errorf("at least one of --name, --desc, --type is required")
	}

	root := getRootDir()
	path, err := resolveFlowPath(root, procFlowName)
	if err != nil {
		return err
	}

	err = loadAndValidate(path, func(f *flow.Flow) {
		for i := range f.Nodes {
			if f.Nodes[i].ID == nodeEditID {
				if nodeEditName != "" {
					f.Nodes[i].Name = nodeEditName
				}
				if nodeEditDesc != "" {
					f.Nodes[i].Description = nodeEditDesc
				}
				if nodeEditType != "" {
					f.Nodes[i].Type = flow.NodeType(nodeEditType)
				}
				if nodeEditStatus != "" {
					f.Nodes[i].Status = nodeEditStatus
				}
				if nodeEditRole != "" {
					if f.Nodes[i].Components == nil {
						f.Nodes[i].Components = &flow.NodeComponents{}
					}
					f.Nodes[i].Components.Roles = []flow.ComponentRef{{Ref: nodeEditRole, Source: flow.SourceBuiltin}}
					cfg := flow.PhaseConfig{Role: nodeEditRole}
					if cfgData, err := json.Marshal(cfg); err == nil {
						f.Nodes[i].Config = cfgData
					}
				}
				if nodeEditRules != "" {
					if f.Nodes[i].Components == nil {
						f.Nodes[i].Components = &flow.NodeComponents{}
					}
					ruleIDs := strings.Split(nodeEditRules, ",")
					f.Nodes[i].Components.Rules = make([]flow.ComponentRef, len(ruleIDs))
					for j, r := range ruleIDs {
						f.Nodes[i].Components.Rules[j] = flow.ComponentRef{Ref: strings.TrimSpace(r), Source: flow.SourceBuiltin}
					}
				}
				return
			}
		}
	})
	if err != nil {
		return err
	}

	// Check if node exists (need to re-parse since mutation succeeded but may have been a no-op)
	f, _ := flow.ParseFlowFile(path)
	found := false
	if f != nil {
		for _, n := range f.Nodes {
			if n.ID == nodeEditID {
				found = true
				break
			}
		}
	}
	if !found {
		return fmt.Errorf("node not found: %s", nodeEditID)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Edited node '%s' in %s\n", nodeEditID, filepath.Base(path))
	return nil
}

// runNodeRemove removes a node and all its associated edges.
func runNodeRemove(cmd *cobra.Command, args []string) error {
	if nodeRemoveID == "" {
		return fmt.Errorf("--id is required")
	}

	root := getRootDir()
	path, err := resolveFlowPath(root, procFlowName)
	if err != nil {
		return err
	}

	found := false
	err = loadAndValidate(path, func(f *flow.Flow) {
		// Check node exists
		nodeIdx := -1
		for i := range f.Nodes {
			if f.Nodes[i].ID == nodeRemoveID {
				nodeIdx = i
				found = true
				break
			}
		}
		if nodeIdx < 0 {
			return
		}

		// Remove node
		f.Nodes = append(f.Nodes[:nodeIdx], f.Nodes[nodeIdx+1:]...)

		// Remove associated edges
		filtered := f.Edges[:0]
		for _, e := range f.Edges {
			if e.From != nodeRemoveID && e.To != nodeRemoveID {
				filtered = append(filtered, e)
			}
		}
		f.Edges = filtered
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("node not found: %s", nodeRemoveID)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed node '%s' from %s\n", nodeRemoveID, filepath.Base(path))
	return nil
}

// runEdgeAdd adds a new edge to a flow definition.
func runEdgeAdd(cmd *cobra.Command, args []string) error {
	if edgeAddFrom == "" {
		return fmt.Errorf("--from is required")
	}
	if edgeAddTo == "" && edgeAddSubflowRef == "" {
		return fmt.Errorf("--to or --subflow-ref is required")
	}
	if edgeAddTo != "" && edgeAddSubflowRef != "" {
		return fmt.Errorf("--to and --subflow-ref are mutually exclusive")
	}

	root := getRootDir()
	path, err := resolveFlowPath(root, procFlowName)
	if err != nil {
		return err
	}

	edgeID := edgeAddID
	if edgeID == "" {
		edgeID, err = generateEdgeID(path, edgeAddFrom, edgeAddTo)
		if err != nil {
			return fmt.Errorf("generate edge ID: %w", err)
		}
	}

	cond := edgeAddCondition
	et := flow.EdgeType(edgeAddType)
	if edgeAddSubflowRef != "" {
		et = flow.EdgeTypeSubflow
	} else if cond != "" && et == "" {
		et = flow.EdgeTypeConditional
	}

	subflowRef := edgeAddSubflowRef

	// Validate subflow-ref refers to an existing flow
	if subflowRef != "" {
		entries := listProcs(root)
		found := false
		for _, e := range entries {
			f, parseErr := flow.ParseFlowFile(e.Path)
			if parseErr == nil && f.Metadata.ID == subflowRef {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("--subflow-ref %s does not match any registered flow id", subflowRef)
		}
	}

	err = loadAndValidate(path, func(f *flow.Flow) {
		edge := flow.FlowEdge{
			ID:         edgeID,
			From:       edgeAddFrom,
			To:         edgeAddTo,
			Type:       et,
			SubflowRef: subflowRef,
		}
		if cond != "" {
			edge.Conditions = []flow.EdgeCondition{
				{Expression: cond},
			}
		}
		f.Edges = append(f.Edges, edge)
	})
	if err != nil {
		return err
	}

	if subflowRef != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Added subflow dispatch edge '%s' (%s -> subflow:%s) to %s\n", edgeID, edgeAddFrom, subflowRef, filepath.Base(path))
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Added edge '%s' (%s -> %s) to %s\n", edgeID, edgeAddFrom, edgeAddTo, filepath.Base(path))
	}
	return nil
}

func generateEdgeID(flowPath, from, to string) (string, error) {
	f, err := flow.ParseFlowFile(flowPath)
	if err != nil {
		return "", err
	}

	existingIDs := make(map[string]bool)
	for _, e := range f.Edges {
		existingIDs[e.ID] = true
	}

	for i := 0; i < 100; i++ {
		candidate := "e_" + idgen.RandHex(3)
		if !existingIDs[candidate] {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not generate unique edge ID for %s->%s", from, to)
}

// runEdgeEdit edits an existing edge's properties.
func runEdgeEdit(cmd *cobra.Command, args []string) error {
	if edgeEditID == "" {
		return fmt.Errorf("--id is required")
	}

	root := getRootDir()
	path, err := resolveFlowPath(root, procFlowName)
	if err != nil {
		return err
	}

	found := false
	err = loadAndValidate(path, func(f *flow.Flow) {
		for i := range f.Edges {
			if f.Edges[i].ID == edgeEditID {
				found = true
				if edgeEditCondition != "" {
					if f.Edges[i].Type == "" {
						f.Edges[i].Type = flow.EdgeTypeConditional
					}
					f.Edges[i].Conditions = []flow.EdgeCondition{
						{Expression: edgeEditCondition},
					}
				}
				return
			}
		}
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("edge not found: %s", edgeEditID)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Edited edge '%s' in %s\n", edgeEditID, filepath.Base(path))
	return nil
}

// runEdgeRemove removes an edge from a flow definition. Can identify by --id or --from/--to.
func runEdgeRemove(cmd *cobra.Command, args []string) error {
	if edgeRemoveID == "" && (edgeAddFrom == "" || edgeAddTo == "") {
		return fmt.Errorf("--id or (--from and --to) is required")
	}

	root := getRootDir()
	path, err := resolveFlowPath(root, procFlowName)
	if err != nil {
		return err
	}

	found := false
	err = loadAndValidate(path, func(f *flow.Flow) {
		for i, e := range f.Edges {
			match := false
			if edgeRemoveID != "" {
				match = e.ID == edgeRemoveID
			} else {
				match = e.From == edgeAddFrom && e.To == edgeAddTo
			}
			if match {
				found = true
				f.Edges = append(f.Edges[:i], f.Edges[i+1:]...)
				return
			}
		}
	})
	if err != nil {
		return err
	}
	if !found {
		if edgeRemoveID != "" {
			return fmt.Errorf("edge not found: %s", edgeRemoveID)
		}
		return fmt.Errorf("edge not found: %s -> %s", edgeAddFrom, edgeAddTo)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed edge from %s\n", filepath.Base(path))
	return nil
}
