package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/origadmin/team-flow/internal/boot"
	"github.com/origadmin/team-flow/internal/doctor"
	"github.com/origadmin/team-flow/internal/editor"
	"github.com/origadmin/team-flow/internal/export"
	"github.com/origadmin/team-flow/internal/graph"
	"github.com/origadmin/team-flow/internal/logger"
	"github.com/origadmin/team-flow/internal/migrate"
	"github.com/origadmin/team-flow/internal/status"
	"github.com/origadmin/team-flow/internal/task"
	"github.com/origadmin/team-flow/internal/toolrunner"
	"github.com/origadmin/team-flow/internal/tools"
	"github.com/origadmin/team-flow/internal/ver"
	"github.com/origadmin/team-flow/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verboseFlag bool
	configPath  string
	log         *logger.Logger
)

var rootCmd = &cobra.Command{
	Use:   "flow",
	Short: "team-flow: AI Team Collaboration Framework CLI",
	Long: `flow is the CLI for team-flow AI collaboration framework.

Supports v1 (task-pool) and v2 (beads-native + code-review-graph) modes.
Provides initialization, migration, graph analysis, and diagnostics.

Tools: External tools can be configured and run through 'flow tools'.
Default tools include: beads (task management with Dolt git-native storage).`,
	Version: version.Version,
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("flow version %s (build: %s, commit: %s)\n", version.Version, version.BuildTime, version.GitCommit))
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output (log to .team/logs/)")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file path (default: ~/.flow/config.yaml)")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		log = logger.GetLogger()
		if verboseFlag {
			log.SetLevel(logger.DEBUG)
		}
		log.Debug("Command executed: " + cmd.Name())
		return nil
	}
}

func main() {
	home, _ := os.UserHomeDir()
	defaultConfig := filepath.Join(home, ".flow", "config.yaml")

	if configPath == "" && viper.GetString("config") == "" {
		if _, err := os.Stat(defaultConfig); err == nil {
			viper.SetConfigFile(defaultConfig)
		}
	}

	if err := viper.ReadInConfig(); err == nil {
		log = logger.GetLogger()
		log.Debug("Loaded config: " + viper.ConfigFileUsed())
	}

	rootCmd.AddCommand(tools.Cmd)
	rootCmd.AddCommand(task.Cmd)
	rootCmd.AddCommand(boot.Cmd)
	rootCmd.AddCommand(doctor.Cmd)
	rootCmd.AddCommand(migrate.Cmd)
	rootCmd.AddCommand(export.Cmd)
	rootCmd.AddCommand(status.Cmd)
	rootCmd.AddCommand(editor.Cmd)
	rootCmd.AddCommand(graph.Cmd)
	rootCmd.AddCommand(ver.Cmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initTools() *toolrunner.ToolsManager {
	tm, err := toolrunner.NewToolsManager("")
	if err != nil {
		return nil
	}

	home, _ := os.UserHomeDir()
	defaultTools := filepath.Join(home, ".flow", "tools.yaml")

	if _, err := os.Stat(defaultTools); err == nil {
		tm2, err := toolrunner.NewToolsManager(defaultTools)
		if err == nil {
			for _, t := range tm2.List() {
				tm.Register(t)
			}
		}
	}

	return tm
}
