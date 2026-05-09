package task

import (
	"fmt"
	"os"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/spf13/cobra"
)

var commandAliases = map[string]string{
	"append": "note",
	"export": "export",
}

var Cmd = &cobra.Command{
	Use:   "task",
	Short: "Unified task management command",
	Long: `Unified AI-facing task command that routes based on .team/version.

Stage routing:
  v1 → task-pool.md (document-based)
  v2 → beads (.beads/ via bd CLI)
  v3 → configurable backend

AI always uses "flow task" regardless of backend.
Auto-timestamps: created_at and updated_at are managed automatically.

Command aliases (flow task → bd):
  append → note  (append conversation record to task)
  Other commands pass through directly to bd.`,
	DisableFlagParsing: true,
	RunE:               runTask,
}

func runTask(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	if args[0] == "--help" || args[0] == "-h" {
		return cmd.Help()
	}

	if !bd.IsAvailable() {
		fmt.Fprintln(os.Stderr, "Error: bd CLI not found. Run 'flow init' to install.")
		fmt.Fprintln(os.Stderr, "Or install manually: https://github.com/steveyegge/beads")
		os.Exit(1)
	}

	if alias, ok := commandAliases[args[0]]; ok {
		args[0] = alias
	}

	output, err := bd.Run(args...)
	if err != nil {
		fmt.Print(output)
		return err
	}

	fmt.Print(output)
	return nil
}
