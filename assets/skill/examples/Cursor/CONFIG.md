# Cursor 配置示例 (v3)

## 环境变量

| 变量 | 值 | 说明 |
|------|-----|------|
| PROJECT_PATH | {PROJECT_PATH} | 项目根目录 |
| DOCS_INTERNAL | {DOCS_INTERNAL} | 内部文档目录 |
| DOCS_EXTERNAL | {DOCS_EXTERNAL} | 外部文档目录 |

## 初始化

1. 读取 `.team/version` → 确认为 v3
2. 加载 `.cursor/skills/team-flow/SKILL.md` 作为入口
3. 执行 `flow proc run` 启动流程引擎
4. 按 `flow proc run` 输出执行当前节点指令

## Bridge 文件

Cursor 的 bridge 文件位于 `.cursor/rules/team-flow.mdc`，由 `flow init --v3` 自动生成。
Bridge 包含 YAML frontmatter（description + globs）和 SKILL.md 路径引用。
