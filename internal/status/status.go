package statuscmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看项目状态",
	Long: `查看当前项目使用的 _team 版本和状态。

显示:
  - v1 或 v2
  - 框架文件路径
  - beads 状态（如果使用 v2）`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return err
		}

		fmt.Println("项目状态:")
		fmt.Printf("  路径: %s\n", projectPath)

		// 检查 v1
		taskPool := filepath.Join(projectPath, "framework", "_team", "task-pool.md")
		if _, err := os.Stat(taskPool); err == nil {
			fmt.Println("  版本: v1 (task-pool)")
			fmt.Printf("  文件: %s\n", taskPool)
		}

		// 检查 v2
		v2Path := filepath.Join(projectPath, "framework", "_team", "v2")
		if _, err := os.Stat(v2Path); err == nil {
			fmt.Println("  版本: v2 (beads-native)")
			fmt.Printf("  路径: %s\n", v2Path)
		}

		// 检查 .team
		teamPath := filepath.Join(projectPath, ".team")
		if _, err := os.Stat(teamPath); err == nil {
			fmt.Printf("  配置: %s ✓\n", teamPath)
		} else {
			fmt.Println("  配置: 未找到 .team/ 目录")
		}

		return nil
	},
}

func init() {
}
