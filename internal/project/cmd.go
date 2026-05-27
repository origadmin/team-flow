package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/state"
	"github.com/spf13/cobra"
)

var (
	projectNameOverride string
)

var Cmd = &cobra.Command{
	Use:   "project",
	Short: "Project registry and switching",
	Long:  `Manage project registry: list, add, remove, switch between projects with flow state persistence.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered projects",
	RunE:  runList,
}

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Register a project in the registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Unregister a project from the registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch to another project (suspend current, resume target)",
	Args:  cobra.ExactArgs(1),
	RunE:  runSwitch,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current project status and suspended projects",
	RunE:  runStatus,
}

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "List project dependencies",
	RunE:  runDeps,
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect current project and available projects",
	RunE:  runDetect,
}

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current project",
	RunE:  runCurrent,
}

func init() {
	addCmd.Flags().StringVar(&projectNameOverride, "name", "", "Override project name (default: directory basename)")
	detectCmd.Flags().Bool("json", false, "Output in JSON format")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(removeCmd)
	Cmd.AddCommand(switchCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(depsCmd)
	Cmd.AddCommand(detectCmd)
	Cmd.AddCommand(currentCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	regPath := DefaultRegistryPath()
	reg, err := LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	if len(reg.Projects) == 0 {
		fmt.Println("No projects registered. Use 'flow project add <path>' to register one.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH\tTEAM\tFLOW\tLAST NODE\tLAST ACTIVE")
	for _, p := range reg.Projects {
		nodeDisplay := ""
		s, _ := state.LoadFlowState(p.Path)
		if s != nil {
			nodeDisplay = s.Node
			if s.Suspended {
				nodeDisplay = "suspended:" + s.Node
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			p.Name, truncateStr(p.Path, 40), p.Team, p.DefaultFlow, nodeDisplay, p.LastActive.Format("2006-01-02 15:04"))
	}
	w.Flush()
	return nil
}

func runAdd(cmd *cobra.Command, args []string) error {
	entry, err := EntryFromPath(args[0])
	if err != nil {
		return err
	}

	if projectNameOverride != "" {
		entry.Name = projectNameOverride
	}

	regPath := DefaultRegistryPath()
	reg, err := LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	if err := reg.Add(*entry); err != nil {
		return err
	}

	if err := SaveRegistry(regPath, reg); err != nil {
		return fmt.Errorf("save registry: %w", err)
	}

	fmt.Printf("✓ Project '%s' registered\n", entry.Name)
	return nil
}

func runRemove(cmd *cobra.Command, args []string) error {
	regPath := DefaultRegistryPath()
	reg, err := LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	if err := reg.Remove(args[0]); err != nil {
		return err
	}

	if err := SaveRegistry(regPath, reg); err != nil {
		return fmt.Errorf("save registry: %w", err)
	}

	fmt.Printf("✓ Project '%s' removed from registry\n", args[0])
	return nil
}

func runSwitch(cmd *cobra.Command, args []string) error {
	targetName := args[0]
	regPath := DefaultRegistryPath()
	reg, err := LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	target := reg.Find(targetName)
	if target == nil {
		return fmt.Errorf("project '%s' not found in registry. Use 'flow project add' first", targetName)
	}

	cwd, _ := os.Getwd()
	var fromName, fromNode string
	currentState, _ := state.LoadFlowState(cwd)
	if currentState != nil {
		fromName = resolveProjectName(cwd, reg)
		fromNode = currentState.Node
		if err := state.SuspendState(cwd, "switch to "+targetName); err != nil {
			fmt.Printf("⚠ Failed to suspend current project: %v\n", err)
		}
	}

	if err := state.ResumeState(target.Path); err != nil {
		fmt.Printf("⚠ Failed to resume target project state: %v\n", err)
	}

	targetState, _ := state.LoadFlowState(target.Path)
	targetNode := "tri3"
	if targetState != nil && targetState.Node != "" {
		targetNode = targetState.Node
	}

	reg.Update(targetName, func(e *ProjectEntry) {
		e.LastNode = targetNode
		e.LastActive = time.Now()
	})
	if fromName != "" {
		reg.Update(fromName, func(e *ProjectEntry) {
			e.LastNode = fromNode
		})
	}
	_ = SaveRegistry(regPath, reg)

	printSwitchOutput(fromName, fromNode, targetName, targetNode, target.Path)
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	regPath := DefaultRegistryPath()
	reg, err := LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	currentName := resolveProjectName(cwd, reg)
	currentState, _ := state.LoadFlowState(cwd)

	if currentState != nil {
		s := "active"
		if currentState.Suspended {
			s = "suspended"
		}
		fmt.Printf("CURRENT: %s (node: %s, flow: %s, status: %s)\n\n", currentName, currentState.Node, currentState.Flow, s)
	} else {
		fmt.Printf("CURRENT: %s (no flow state)\n\n", currentName)
	}

	hasSuspended := false
	for _, p := range reg.Projects {
		s, _ := state.LoadFlowState(p.Path)
		if s != nil && s.Suspended {
			if !hasSuspended {
				fmt.Println("SUSPENDED:")
				hasSuspended = true
			}
			fmt.Printf("  %s  (node: %s, suspended: %s, reason: %s)\n",
				p.Name, s.Node, s.SuspendedAt.Format("2006-01-02 15:04"), s.SuspendReason)
		}
	}

	if !hasSuspended {
		fmt.Println("No suspended projects.")
	}

	return nil
}

func runDeps(cmd *cobra.Command, args []string) error {
	fmt.Println("No dependencies configured. Add 'dependencies' section to .team/project.yaml.")
	return nil
}

func printSwitchOutput(fromName, fromNode, toName, toNode, toPath string) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  Project Switched                                          ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")

	if fromName != "" {
		fmt.Printf("║  FROM: %-10s (suspended at node %-20s)║\n", fromName, fromNode)
	}

	fmt.Printf("║  TO:   %-10s (resuming at node %-22s)║\n", toName, toNode)
	fmt.Println("║                                                            ║")
	fmt.Println("║  RESUME INSTRUCTION:                                       ║")
	fmt.Printf("║    cd %-52s║\n", toPath)
	fmt.Printf("║    flow proc run %-41s║\n", toNode)
	fmt.Println("║                                                            ║")
	fmt.Println("║  TO RETURN:                                                ║")

	if fromName != "" {
		fmt.Printf("║    flow project switch %-36s║\n", fromName)
	}

	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	type switchOutput struct {
		Action            string `json:"action"`
		FromName          string `json:"from_name,omitempty"`
		FromNode          string `json:"from_node,omitempty"`
		ToName            string `json:"to_name"`
		ToNode            string `json:"to_node"`
		ToPath            string `json:"to_path"`
		ResumeInstruction string `json:"resume_instruction"`
		ReturnInstruction string `json:"return_instruction,omitempty"`
	}

	out := switchOutput{
		Action:            "project_switch",
		FromName:          fromName,
		FromNode:          fromNode,
		ToName:            toName,
		ToNode:            toNode,
		ToPath:            toPath,
		ResumeInstruction: fmt.Sprintf("cd %s && flow proc run %s", toPath, toNode),
		ReturnInstruction: fmt.Sprintf("flow project switch %s", fromName),
	}

	jsonData, _ := json.Marshal(out)
	fmt.Printf("\n<!-- JSON:%s -->\n", string(jsonData))
}

func resolveProjectName(cwd string, reg *ProjectRegistry) string {
	for _, p := range reg.Projects {
		if p.Path == cwd {
			return p.Name
		}
	}
	return filepath.Base(cwd)
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "..." + s[len(s)-max+3:]
}

func runDetect(cmd *cobra.Command, args []string) error {
	result := DetectProject()

	detectJSON, _ := cmd.Flags().GetBool("json")
	if detectJSON {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	PrintDetectionResult(result)
	return nil
}

func runCurrent(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}

	projectRoot := config.ResolveProjectRoot(cwd)
	teamRoot := config.FindTeamRoot(cwd)
	if projectRoot == "" && teamRoot == "" {
		fmt.Println("Not in any project directory")
		fmt.Println("\nUse 'flow project list' to see available projects")
		fmt.Println("Use 'flow project add <path>' to register a project")
		return nil
	}

	if projectRoot == "" {
		projectRoot = teamRoot
	}

	name := getProjectName(projectRoot)
	version := getProjectVersion(projectRoot)

	fmt.Printf("Current project: %s\n", name)
	if version != "" {
		fmt.Printf("Version: %s\n", version)
	}
	fmt.Printf("PROJECT_ROOT: %s\n", projectRoot)

	if teamRoot != "" && teamRoot != projectRoot {
		fmt.Printf("TEAM_ROOT:    %s\n", teamRoot)
	}

	workspace := config.FindWorkspaceRoot(cwd)
	if workspace != "" {
		fmt.Printf("Workspace:    %s\n", workspace)
	}

	return nil
}
