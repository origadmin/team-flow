package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flow",
	Short: "AI Team Collaboration Framework CLI",
	Long: `flow - AI团队协作文框架命令行工具

管理 _team 框架的初始化、迁移和可视化编辑。

示例:
  flow init          # 初始化项目（默认v2）
  flow init --v1     # 使用v1 (task-pool)
  flow migrate       # v1 → v2 迁移
  flow status        # 查看项目状态
  flow editor        # 启动Flow可视化编辑器`,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(editorCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
