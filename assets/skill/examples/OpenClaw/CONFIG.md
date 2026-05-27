# OpenClaw 配置示例 (v3)

## 环境变量

| 变量 | 值 | 说明 |
|------|-----|------|
| PROJECT_PATH | {PROJECT_PATH} | 项目根目录 |
| DOCS_INTERNAL | {DOCS_INTERNAL} | 内部文档目录 |
| DOCS_EXTERNAL | {DOCS_EXTERNAL} | 外部文档目录 |

## 初始化

1. 读取 `.team/version` → 确认为 v3
2. 加载 `.openclaw/skills/team-flow/SKILL.md` 作为入口
3. 执行 `flow proc run` 启动流程引擎
4. 按 `flow proc run` 输出执行当前节点指令

## Bridge 文件

OpenClaw 的 bridge 文件位于 `.openclaw/rules/team-flow.md`，由 `flow init --v3` 自动生成。
Bridge 仅包含 SKILL.md 路径引用。

## OpenClaw 特有功能

| 功能 | 工具 | 用途 |
|------|------|------|
| 子 Agent | `sessions_spawn` | 启动子 Agent 执行节点任务 |
| 历史记忆 | `memory_search` / `memory_get` | 查询历史上下文 |
| 对话检索 | `lcm_grep` / `lcm_expand` | 对话历史检索 |
| 定时任务 | `cron` | 设置定时任务 |
| 消息推送 | `message` | 跨渠道消息发送 |
