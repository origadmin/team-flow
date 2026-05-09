package tools

import (
	"fmt"
	"os"

	"github.com/origadmin/team-flow/internal/toolrunner"
	"github.com/spf13/cobra"
)

var (
	configPath string
	toolsDir   string
)

var Cmd = &cobra.Command{
	Use:   "tools",
	Short: "Manage and run external tools",
	Long: `Manage and run external tools through flow.

Tools are configured via YAML configuration files and can be
registered, listed, and run through this command.

Examples:
  flow tools list                        # List all registered tools
  flow tools show beads                  # Show tool configuration
  flow tools beads init                 # Run tool command
  flow tools beads list --json          # Run with arguments
  flow tools install beads              # Install a tool`,
	RunE: runTools,
}

func init() {
	Cmd.PersistentFlags().StringVar(&configPath, "config", "", "Tools configuration file path")
	Cmd.PersistentFlags().StringVar(&toolsDir, "dir", "", "Tools directory (default: ~/.flow/tools/)")
}

func runTools(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	tm, err := toolrunner.NewToolsManager(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tools config: %v\n", err)
	}

	if args[0] == "list" {
		return runList(tm)
	}

	if args[0] == "show" {
		if len(args) < 2 {
			return fmt.Errorf("tool name required")
		}
		return runShow(tm, args[1])
	}

	if args[0] == "install" {
		if len(args) < 2 {
			return fmt.Errorf("tool name required")
		}
		return runInstall(tm, args[1])
	}

	toolName := args[0]
	toolArgs := args[1:]

	tool, ok := tm.Get(toolName)
	if !ok {
		return fmt.Errorf("tool not found: %s\nRun 'flow tools list' to see available tools", toolName)
	}

	if !tm.IsAvailable(toolName) {
		if tool.Install.Enabled {
			fmt.Printf("⚠ %s not found. Installing...\n", toolName)
			if err := tm.Install(toolName); err != nil {
				return fmt.Errorf("install failed: %w", err)
			}
			fmt.Printf("✓ %s installed successfully\n", toolName)
		} else {
			return fmt.Errorf("%s not found. Please install it first.", toolName)
		}
	}

	subcommand := "help"
	if len(toolArgs) > 0 {
		subcommand = toolArgs[0]
		toolArgs = toolArgs[1:]
	}

	output, err := tm.Run(toolName, subcommand, toolArgs)
	if err != nil {
		fmt.Print(output)
		return fmt.Errorf("tool command failed: %w", err)
	}

	fmt.Print(output)
	return nil
}

func runList(tm *toolrunner.ToolsManager) error {
	tools := tm.List()
	if len(tools) == 0 {
		fmt.Println("No tools registered.")
		fmt.Println("Add tools to ~/.flow/tools.yaml or use 'flow tools register'")
		return nil
	}

	fmt.Println("Registered tools:")
	fmt.Println()
	for _, t := range tools {
		available := "✗"
		if tm.IsAvailable(t.Name) {
			available = "✓"
		}
		fmt.Printf("  %s %s - %s\n", available, t.Name, t.Description)
	}
	return nil
}

func runShow(tm *toolrunner.ToolsManager, name string) error {
	tool, ok := tm.Get(name)
	if !ok {
		return fmt.Errorf("tool not found: %s", name)
	}

	fmt.Printf("Tool: %s\n", tool.Name)
	fmt.Printf("Description: %s\n", tool.Description)
	fmt.Printf("Binary: %s\n", tool.Binary.Path)
	fmt.Printf("Auto-install: %v\n", tool.Install.Enabled)

	if len(tool.Commands) > 0 {
		fmt.Println("\nCommands:")
		for name, cmd := range tool.Commands {
			fmt.Printf("  %s - %s\n", name, cmd.Description)
		}
	}

	return nil
}

func runInstall(tm *toolrunner.ToolsManager, name string) error {
	tool, ok := tm.Get(name)
	if !ok {
		return fmt.Errorf("tool not found: %s", name)
	}

	if !tool.Install.Enabled {
		return fmt.Errorf("auto-install not enabled for %s", name)
	}

	if tm.IsAvailable(name) {
		fmt.Printf("%s is already installed\n", name)
		return nil
	}

	fmt.Printf("Installing %s...\n", name)

	if err := tm.Install(name); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	fmt.Printf("✓ %s installed successfully\n", name)
	return nil
}
