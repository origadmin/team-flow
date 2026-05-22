# AI 团队协作框架 — 快速开始 (v3)

> 框架版本：v3.1

## 支持的平台

| 平台 | 配置文件 | Bridge 格式 |
|------|---------|------------|
| Trae | `examples/Trae/CONFIG.md` | `.trae/rules/team-flow.md` |
| Cursor | `examples/Cursor/CONFIG.md` | `.cursor/rules/team-flow.mdc` (YAML frontmatter) |
| Claude Code | `examples/ClaudeCode/CONFIG.md` | `.claude/rules/team-flow.md` |
| OpenClaw | `examples/OpenClaw/CONFIG.md` | `.openclaw/rules/team-flow.md` |
| Gemini CLI | `examples/Gemini/CONFIG.md` | 无内置 bridge |

## 初始化流程

```bash
# 全新项目
flow init --v3 [--flow {name}]

# v2 项目升级
flow migrate v3 [--flow {name}]
```

初始化后：
1. 读取 `.team/version` → 确认为 v3
2. 加载 `{SKILL_PATH}/SKILL.md` 作为入口
3. 执行 `flow proc run` 启动流程引擎
4. 按 `flow proc run` 输出执行当前节点指令

## 核心文件

| 文件 | 用途 |
|------|------|
| `SKILL.md` | v3 入口 + 核心规则 |
| `CONSENSUS.md` | 共识文件（架构决策） |
| `skills/team-flow-v3-create/SKILL.md` | 流程创建技能 |
| `skills/team-flow-v3-exec/SKILL.md` | 流程执行技能 |
| `.team/project.md` | 项目配置（含 default_flow） |
| `.team/version` | 版本标识（v3） |

## v2 → v3 变化

| v2 | v3 | 说明 |
|----|----|------|
| `TEAM_PATH` | 已移除 | v3 不区分框架层/项目层 |
| `workflows/shared.md` | `flow proc run` | 流程引擎替代静态工作流 |
| `task-pool.md` | `flow task` (beads) | 任务管理由 beads 驱动 |
| `prompts/*.md` | Flow JSON 中的 roles | 角色定义内嵌在流程中 |
| `TEAM_VERSION=7.0` | `.team/version=v3` | 版本标识简化 |
