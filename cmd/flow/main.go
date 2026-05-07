package main

import (
	"fmt"
	"os"

	"github.com/origadmin/team-flow/internal/boot"
	"github.com/origadmin/team-flow/internal/doctor"
	"github.com/origadmin/team-flow/internal/editor"
	"github.com/origadmin/team-flow/internal/export"
	"github.com/origadmin/team-flow/internal/graph"
	"github.com/origadmin/team-flow/internal/migrate"
	"github.com/origadmin/team-flow/internal/status"
	"github.com/origadmin/team-flow/internal/ver"
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
	rootCmd.AddCommand(boot.Cmd)
	rootCmd.AddCommand(doctor.Cmd)
	rootCmd.AddCommand(migrate.Cmd)
	rootCmd.AddCommand(export.Cmd)
	rootCmd.AddCommand(status.Cmd)
	rootCmd.AddCommand(editor.Cmd)
	rootCmd.AddCommand(graph.Cmd)
	rootCmd.AddCommand(ver.Cmd)

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
