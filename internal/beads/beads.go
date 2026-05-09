package beads

import (
	"fmt"
	"os"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/spf13/cobra"
)

var (
	beadsDir string
)

var Cmd = &cobra.Command{
	Use:   "beads",
	Short: "Beads task management (wrapper for bd CLI)",
	Long: `Beads task management commands. Delegates to bd CLI internally.

Examples:
  flow beads init                    # Initialize beads database
  flow beads create "Task title"     # Create a new task
  flow beads list                    # List all tasks
  flow beads show <id>               # Show task details
  flow beads update <id> --title "..." # Update task
  flow beads close <id>             # Close a task

For full bd CLI documentation, see: https://github.com/steveyegge/beads`,
	RunE: runBeads,
}

func init() {
	Cmd.PersistentFlags().StringVar(&beadsDir, "dir", "", "Beads database directory (default: .beads in current directory)")
}

func runBeads(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if err := cmd.Help(); err != nil {
			return err
		}
		return nil
	}

	if !bd.IsAvailable() {
		fmt.Println("⚠ beads (bd CLI) not found.")
		fmt.Println("  Attempting to install...")

		if err := bd.InstallQuiet(); err != nil {
			fmt.Printf("  Install failed: %v\n", err)
			fmt.Println("  Please install manually: https://github.com/steveyegge/beads")
			os.Exit(1)
		}

		fmt.Println("  beads installed successfully!")
	}

	subcommand := args[0]
	subArgs := args[1:]

	if subcommand == "help" && len(subArgs) == 0 {
		return cmd.Help()
	}

	if subcommand == "--help" || subcommand == "-h" {
		subArgs = []string{"--help"}
		subcommand = subArgs[0]
		subArgs = subArgs[1:]
	}

	output, err := bd.RunQuiet(append([]string{subcommand}, subArgs...)...)
	if err != nil {
		fmt.Print(output)
		return fmt.Errorf("beads command failed")
	}

	fmt.Print(output)
	return nil
}
