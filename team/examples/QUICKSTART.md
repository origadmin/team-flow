# AI 团队协作框架 — 快速开始

> 框架版本：TEAM_VERSION=7.0
> 参考格式：`examples/Trae/CONFIG.md`

## 支持的平台

| 平台 | 配置文件 |
|------|---------|
| Trae | `examples/Trae/CONFIG.md` |
| OpenClaw | `examples/OpenClaw/CONFIG.md` |
| Gemini CLI | `examples/Gemini/CONFIG.md` |
| Claude Code | `examples/ClaudeCode/CONFIG.md` |
| Cursor | `examples/Cursor/CONFIG.md` |

## 初始化流程

```
1. 加载 {TEAM_PATH}/SKILL.md（含 Agent 映射表）
2. 加载 .team/project.md（不完整则先补全）
3. 意图识别 → 自动分类（无需 T: 前缀）
4. 分发到子 Agent（Task 工具 + subagent_type）
5. 执行任务，更新 task-pool.md
```

## 核心文件

| 文件 | 用途 |
|------|------|
| `SKILL.md` | 框架入口 + Agent 映射 |
| `workflows/shared.md` | 三层门禁 + 两层工作流 |
| `workflows/roles/*.md` | 各角色行为规范 |
| `prompts/*.md` | 各角色 prompts 定义 |
| `config/team-config.json` | 团队配置 + subagent_type 映射 |
| `templates/project.md` | 项目信息模板 |
