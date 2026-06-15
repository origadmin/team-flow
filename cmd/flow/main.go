package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/origadmin/team-flow/internal/boot"
	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/doctor"
	"github.com/origadmin/team-flow/internal/editor"
	"github.com/origadmin/team-flow/internal/export"
	"github.com/origadmin/team-flow/internal/graph"
	"github.com/origadmin/team-flow/internal/logger"
	"github.com/origadmin/team-flow/internal/migrate"
	projectpkg "github.com/origadmin/team-flow/internal/project"
	"github.com/origadmin/team-flow/internal/proc"
	"github.com/origadmin/team-flow/internal/session"
	"github.com/origadmin/team-flow/internal/skill"
	"github.com/origadmin/team-flow/internal/status"
	"github.com/origadmin/team-flow/internal/task"
	"github.com/origadmin/team-flow/internal/toolrunner"
	"github.com/origadmin/team-flow/internal/tools"
	"github.com/origadmin/team-flow/internal/tracecmd"
	"github.com/origadmin/team-flow/internal/update"
	"github.com/origadmin/team-flow/internal/startup"
	"github.com/origadmin/team-flow/internal/validate"
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

Provides initialization, migration, graph analysis, and diagnostics.

Tools: External tools can be configured and run through 'flow tools'.
Default tools include: task management (beads), code review, and flow engine.`,
	Version: version.Version,
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("flow version %s (build: %s, commit: %s)\n", version.Version, version.BuildTime, version.GitCommit))
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file path (default: ~/.flow/config.yaml)")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		log = logger.GetLogger()
		if verboseFlag {
			log.SetLevel(logger.DEBUG)
		}
		log.Debug("Command executed: %s", cmd.Name())

		// 执行上下文检查
		return checkContext(cmd)
	}
}

// 上下文类型
type contextType string

const (
	contextNone contextType = "none"
	contextWorkspace contextType = "workspace"
	contextProject contextType = "project"
)

// 需要在 project 下运行的命令
var projectCommands = map[string]bool{
	"proc": true,
	"skill": true,
	"session": true,
	"task": true,
	"status": true,
	"doctor": true,
	"validate": true,
}

// 上下文检查函数
func checkContext(cmd *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	// 检查命令是否需要在 project 下运行
	needsProject := false
	currentCmd := cmd
	for currentCmd != nil {
		if projectCommands[currentCmd.Name()] {
			needsProject = true
			break
		}
		currentCmd = currentCmd.Parent()
	}

	// 检测当前是什么上下文
	isWorkspace := config.IsWorkspaceRoot(cwd)

	// 首先判断是否在 workspace 根
	if isWorkspace {
		// 如果是需要在 project 下运行的命令，拒绝
		if needsProject {
			fmt.Fprintln(os.Stderr, "⛔ WORKSPACE ROOT DETECTED")
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "团队操作（flow proc、flow skill、flow task 等）不允许在 workspace 根下执行。")
			fmt.Fprintln(os.Stderr)

			// 显示可用项目
			cfg, err := config.LoadProjectConfig(cwd)
			if err == nil {
				fmt.Fprintln(os.Stderr, "请切换到其中一个项目：")
				for _, p := range cfg.Projects {
					fmt.Fprintf(os.Stderr, "  %s → %s\n", p.Name, p.Path)
				}
				fmt.Fprintln(os.Stderr)
			}

			// 如果有 active_project，提示
			cfg, err = config.LoadProjectConfig(cwd)
			if err == nil && cfg.ActiveProject != "" {
				fmt.Fprintf(os.Stderr, "活跃项目: %s\n", cfg.ActiveProject)
				fmt.Fprintf(os.Stderr, "快速切换: cd %s\n", cfg.ActiveProject)
			}

			return fmt.Errorf("workspace root: 请切换到具体项目")
		}
		return nil
	}

	// 在 project 下或无上下文，正常执行
	return nil
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
		log.Debug("Loaded config: %s", viper.ConfigFileUsed())
	}

	rootCmd.AddCommand(tools.Cmd)
	rootCmd.AddCommand(task.Cmd)
	rootCmd.AddCommand(boot.Cmd)
	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(doctor.Cmd)
	rootCmd.AddCommand(migrate.Cmd)
	rootCmd.AddCommand(export.Cmd)
	rootCmd.AddCommand(status.Cmd)
	rootCmd.AddCommand(editor.Cmd)
	rootCmd.AddCommand(graph.Cmd)
	rootCmd.AddCommand(ver.Cmd)
	rootCmd.AddCommand(proc.Cmd)
	rootCmd.AddCommand(projectpkg.Cmd)
	rootCmd.AddCommand(update.Cmd)
	rootCmd.AddCommand(skill.Cmd)
	rootCmd.AddCommand(session.Cmd)
	rootCmd.AddCommand(startup.Cmd)
	rootCmd.AddCommand(tracecmd.Cmd)
	rootCmd.AddCommand(validate.Cmd)

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
