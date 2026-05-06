package initcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	useV1    bool
	useV2    bool
	force    bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化项目 _team 框架",
	Long: `初始化项目，创建 .team/ 目录并复制框架文件。

支持 v1 (task-pool) 和 v2 (beads-native) 两种模式。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 确定版本
		version := "v2"
		if useV1 && !useV2 {
			version = "v1"
		}

		// 获取当前目录
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("获取当前目录失败: %w", err)
		}

		fmt.Printf("初始化项目: %s\n", projectPath)
		fmt.Printf("使用版本: %s\n", version)

		// 创建 .team 目录
		teamDir := filepath.Join(projectPath, ".team")
		if err := os.MkdirAll(teamDir, 0755); err != nil {
			return fmt.Errorf("创建 .team 目录失败: %w", err)
		}

		// 创建 project.md
		projectMd := `## Project Configuration

- **Project Name**: ` + filepath.Base(projectPath) + `
- **Team ID**: {team-id}
- **Create Date**: TBD

## Toolchain

### Backend
pipeline: go test ./... | go build -o bin/app

### Frontend  
pipeline: bun run test | bun run build

## Constraints
- No Chinese comments in code
- TDD required
`
		if err := os.WriteFile(filepath.Join(teamDir, "project.md"), []byte(projectMd), 0644); err != nil {
			return fmt.Errorf("创建 project.md 失败: %w", err)
		}

		// 复制框架文件
		// team/ 目录包含 _team 框架文件（v1/, v2/, prompts/, workflows/, templates/）
		teamSrc := "team"
		teamDst := filepath.Join(projectPath, "framework", "_team")
		
		// 检查 team/ 目录是否存在
		if _, err := os.Stat(teamSrc); os.IsNotExist(err) {
			return fmt.Errorf("team/ 目录不存在，无法复制框架文件")
		}
		
		fmt.Printf("框架文件源: %s\n", teamSrc)
		fmt.Printf("框架文件目标: %s\n", teamDst)
		fmt.Println("⚠️  请手动复制 team/ 目录内容到项目的 framework/_team/")
		fmt.Println("   或运行: Copy-Item -Path team/* -Destination framework/_team/ -Recurse")

		fmt.Println("✅ 初始化完成!")
		fmt.Println("下一步:")
		fmt.Println("  1. 编辑 .team/project.md")
		fmt.Println("  2. 复制 framework/ 目录")
		fmt.Println("  3. 开始使用: AI 读取 framework/_team/v2/SKILL.md")

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&useV1, "v1", false, "使用 v1 (task-pool)")
	initCmd.Flags().BoolVar(&useV2, "v2", true, "使用 v2 (beads-native)")
	initCmd.Flags().BoolVar(&force, "force", false, "强制覆盖已存在的文件")
}
