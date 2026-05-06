package migratecmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "v1 → v2 迁移",
	Long: `将项目从 v1 (task-pool.md) 迁移到 v2 (beads-native)。

会自动:
  1. 检测 v1 任务
  2. 创建 beads 数据库
  3. 迁移任务到 beads
  4. 更新相关文件引用`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("开始 v1 → v2 迁移...")

		// TODO: 实现迁移逻辑
		// 1. 读取 task-pool.md
		// 2. 调用 bd init
		// 3. 逐条创建 beads issue
		// 4. 创建 v2/ 目录结构

		fmt.Println("⚠️  迁移功能开发中...")
		fmt.Println("请手动:")
		fmt.Println("  1. 运行: bd init")
		fmt.Println("  2. 复制 v1 任务到 beads")
		fmt.Println("  3. 更新项目配置")

		return nil
	},
}

func init() {
	migrateCmd.Flags().Bool("dry-run", false, "干跑，不实际执行")
}
