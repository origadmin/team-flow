# OpenClaw 配置示例 (v3)

> 基于 team-flow v3 流程驱动架构适配 OpenClaw 环境

## 快速开始

1. 读取 `.team/version` → 确认为 v3
2. 加载 `.openclaw/skills/team-flow/SKILL.md` 作为入口
3. 执行 `flow proc run` 启动流程引擎
4. 按 `flow proc run` 输出执行当前节点指令

详细配置见 [CONFIG.md](./CONFIG.md)

## OpenClaw 特有适配

| v3 概念 | OpenClaw 工具 | 说明 |
|---------|--------------|------|
| 子 Agent 执行 | `sessions_spawn` | 启动子 Agent 执行节点任务 |
| 历史上下文 | `memory_search` / `memory_get` | 查询历史记忆 |
| 对话检索 | `lcm_grep` / `lcm_expand` | 对话历史检索 |
| 定时任务 | `cron` | 设置定时任务 |
| 消息推送 | `message` | 跨渠道消息发送 |
