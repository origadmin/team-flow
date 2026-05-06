package main

import (
	"fmt"
	"os"

	"github.com/origadmin/team-flow/internal/doctorcmd"
	"github.com/origadmin/team-flow/internal/editor"
	"github.com/origadmin/team-flow/internal/graphcmd"
	"github.com/origadmin/team-flow/internal/initcmd"
	"github.com/origadmin/team-flow/internal/migratecmd"
	"github.com/origadmin/team-flow/internal/statuscmd"
	"github.com/origadmin/team-flow/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flow",
	Short: "team-flow: AI Team Collaboration Framework CLI",
	Long: `flow is the CLI for team-flow AI collaboration framework.

Supports v1 (task-pool) and v2 (beads-native + code-review-graph) modes.
Provides initialization, migration, graph analysis, and diagnostics.`,
	Version: version.Version,
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("flow version %s (build: %s, commit: %s)\n", version.Version, version.BuildTime, version.GitCommit))
}

func main() {
	rootCmd.AddCommand(initcmd.Cmd)
	rootCmd.AddCommand(doctorcmd.Cmd)
	rootCmd.AddCommand(migratecmd.Cmd)
	rootCmd.AddCommand(statuscmd.Cmd)
	rootCmd.AddCommand(editor.Cmd)
	rootCmd.AddCommand(graphcmd.Cmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Info())
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
