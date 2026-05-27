---
name: team-flow-v3
version: 3.2
description: |
  team-flow v3: Process-Centric AI Collaboration Framework.
  当用户提到以下任何关键词时必须触发此技能：team-flow, team-flow/v3, 团队流程, 流程管理,
  软件开发流程, 项目管理流程, v3流程, dev-flow, bugfix-flow, feature-flow, hotfix-flow,
  flow proc run, flow init, 齐活林, 匠思远, 铸灵手, 严过关, 合无间, 寇豆码, 溯源, 高见远.
  当 .team/version = v3 时，此技能是所有工作的强制入口。
  做任何工作之前必须先运行 flow proc run 进入流程驱动执行协议。
---

# team-flow — Entry Point

> **Version**: v3.2 | **Date**: 2026-05-27

## Session Startup Protocol (EVERY SESSION)

```
Step 1: flow project detect    → Lock the project
Step 2: flow proc run          → Get work instructions (role, rules, tools, docs)
Step 3: Adopt principal role   → Start working
```

⛔ Only AFTER Step 3, start working on user requests.

## First-Time Setup

If `flow proc run` fails because no flow is bound:

```bash
# Option A: Interactive setup (choose from 5 preset teams)
flow init --v3

# Option B: Specify a team directly
flow init --v3 --team dev-team

# Option C: Specify a flow directly (auto-finds the team)
flow init --v3 --flow dev-flow
```

Then re-run `flow proc run`.

## Available Teams

| Team | ID | Flows | Best For |
|------|----|-------|----------|
| Software Dev | `dev-team` | 8 (dev/feature/bugfix/hotfix/change/release/analysis/batch) | Software development |
| Content | `content-team` | 3 (content-distribution/novel/promo-video) | Articles, videos, novels |
| Game Design | `game-team` | 1 (game-design) | Game design |
| Trading | `trading-team` | 1 (trading) | Trading analysis |
| Skill Dev | `skill-team` | 2 (skill-dev/test-simple) | AI skill development |

## Version Routing

| `.team/version` | Skill File | Capability |
|-----------------|-----------|-----------|
| `v3` | `assets/skill/v3/SKILL.md` | Full flow engine |
| `v2` | `assets/skill/v2/SKILL.md` | flow task management |
| `v1` | `assets/skill/v1/SKILL.md` | Document-only workflow |

## Key Commands

```bash
flow project detect      # Detect and lock project
flow proc run            # Start flow engine
flow proc run {node-id}  # Run specific node
flow init --v3           # Initialize v3 project
flow migrate v3          # Migrate from v2 to v3
flow task list           # List tasks
flow task show <id>      # View task
flow task update <id> --claim  # Claim task
flow task close <id>     # Close task
flow config paths        # Show resolved paths
```

## Installation

```bash
# Via npm (recommended for AI tool users)
npx skills add @origadmin/team-flow

# Via Go
go install github.com/origadmin/team-flow/cmd/flow@latest
```

## Compatible AI Tools

Claude Code, Cursor, Trae, Windsurf, GitHub Copilot, OpenCode, OpenClaw, Cline, Gemini CLI
