---
name: team-flow-v3
version: 3.1
description: |
  team-flow v3: Process-Centric AI Collaboration Framework
  Process is central, with rules, tools, and skills unified under flow definitions.
  Creates self-documenting, executable AI workflows that can be verified, visualized, and refined.
---

# team-flow v3 SKILL.md - Entry Point

> **Version**: v3.1 | **Date**: 2026-05-21

## ⛔ MANDATORY: Read Consensus First

**Before doing ANYTHING in v3, read [CONSENSUS.md](./CONSENSUS.md).** It contains all confirmed architectural decisions. Do NOT re-ask or re-confirm anything already decided there. Violating this wastes the user's time.

## Status Line (MANDATORY — Every Response)

Every response MUST start with:

```
[Role: {alias} | Flow: {flow-name}#{beads-id} | Node: {node-id} | Phase: {phase}]
```

- Role: alias from `flow proc run` output (e.g., 齐活林, 匠思远, 铸灵手)
- Flow: flow name + beads task ID (e.g., skill-dev-flow#team-flow-6x9)
- Node: current node ID (e.g., tri1, fa03, fd04)
- Phase: current execution phase (analyze/design/implement/verify/review)

⛔ Status Line data source: `flow proc run` output, NEVER hardcode

## Core Concept

> **Flow First**: Everything starts with a flow definition. The flow specifies:
> 1. What steps to execute (nodes)
> 2. What order to execute them (edges)
> 3. What rules apply in each step (components)
> 4. What deliverables to produce (docs)
> 5. What gates enforce quality (gates)

## First-Time Setup

v3 有三种初次使用场景，每种都有对应的入口：

### Scenario A: 全新项目（CLI 初始化）

```bash
flow init --v3 [--team {team-id}] [--flow {flow-name}]
```

`flow init --v3` 会：
1. 创建 `.team/` 目录和配置文件
2. 展示预设团队菜单（5个预设团队），用户选择最接近的
3. 不合适选"0: 创建新团队"，AI 用 create skill 从零设计
4. 也可以 `--team dev-team` 直接指定团队
5. 或 `--flow dev-flow` 通过 flow 名找到对应团队
6. 安装选中团队的所有 flows 到 `.team/flows/`
7. 初始化 beads 数据库
8. 在 project.md 中设置 `default_flow`

### Scenario B: AI 会话中发现无流程绑定

当 AI 执行 Session Startup Protocol 时，发现 `default_flow` 为空或指向不存在的流程：

```
Step 1: 运行 flow proc list
  → 查看所有已注册的团队

Step 2: 判断可用团队
  → 如果有预设团队 → 向用户展示选项，引导选择
  → 如果没有团队 → 直接进入创建流程

Step 3: 用户选择
  → 用户选了某个团队 → 更新 project.md 的 default_flow → 重新执行 flow proc run
  → 用户要创建新团队 → 加载 team-flow-v3-create skill → 创建完成后注册并设置 default_flow
  → 用户想先看看 → flow proc show {name} 展示详情

Step 4: 设置 default_flow
  → 编辑 .team/project.md，设置 default_flow: {chosen-flow}
  → 重新执行 flow proc run，进入正常执行流程
```

**引导话术示例**：
> 这个项目还没有绑定团队流程。我找到了以下可用团队：
> 1. dev-flow — 软件开发全流程
> 2. novel-flow — 小说创作流程
> 3. ...
> 
> 选择一个最接近你需求的，或者告诉我你的具体场景，我帮你创建一个专属团队。

### Scenario C: v2 项目升级到 v3

```bash
flow migrate v3 [--flow {name}]
```

`flow migrate v3` 会：
1. 备份 v2 配置
2. 安装 v3 技能到 IDE skill 目录
3. 安装预设流程到 `.team/flows/` 并注册到 project.md
4. 更新 IDE bridge 文件指向 v3
5. 更新 `.team/version` 为 v3
6. 设置 `default_flow`

### 团队放置规则

| 来源 | 放置路径 | 说明 |
|------|----------|------|
| 框架预设（团队模板） | `teams/{team-id}/flows/` | 随工具嵌入，只读 |
| 项目安装 | `.team/flows/` | `flow init --v3` 安装到这里 |
| 项目自定义 | `.team/flows/` | 用户创建的放这里 |
| 解析优先级 | `.team/flows/` | 项目目录优先 |

### 预设团队模板

| 团队 | ID | Flows | 适用场景 |
|------|----|-------|----------|
| 软件开发团队 | `dev-team` | 8 (dev/feature/bugfix/hotfix/change/release/analysis/batch) | 通用软件开发 |
| 内容创作团队 | `content-team` | 3 (content-distribution/novel/promo-video) | 文章、视频、小说 |
| 游戏设计团队 | `game-team` | 1 (game-design) | 游戏设计 |
| 交易分析团队 | `trading-team` | 1 (trading) | 交易分析 |
| 技能开发团队 | `skill-team` | 2 (skill-dev/test-simple) | AI 技能开发 |

## Binding: One Project, One Flow

A project is bound to exactly one team flow. This is non-negotiable.

- The binding is stored in project configuration (`default_flow` in `.team/project.md`)
- All tasks within the project execute under the same flow
- To change the flow, the user must explicitly switch (project-level decision)
- If no flow is bound, follow the First-Time Setup (Scenario B) above

## Principal: The Sole User Interface

Every flow has exactly one **principal** role (marked `"principal": true`). This role is the team's public face:

- **Only the principal communicates with the user** — all other roles are execution-only
- **User input always routes to the principal** — regardless of which node is currently active
- **The principal dispatches and receives results** — other roles never speak directly to the user
- **Interruptions return to the principal** — if the user sends new input mid-execution, the principal re-assesses

This means:
1. When you receive user input → find the principal node → route there
2. When a sub-role finishes → report to principal → principal decides next step
3. When user interrupts → stop → route to principal → principal re-dispatches

## Flow Loading

v3 flows are installed to `.team/flows/` by `flow init --v3 --team {team-id}`. Each flow is a JSON file following the v3 schema.

When a user asks for a task, v3:
1. Finds matching flow from `.team/flows/` (or creates one if missing)
2. Validates flow with `flow proc validate`
3. Executes the flow using `flow proc run`

## Two Core Skills

| Skill | Purpose |
|-------|---------|
| team-flow-v3-create | Create/modify/refine v3 flow definitions |
| team-flow-v3-exec | Execute a v3 flow, enforcing all rules and gates |

## GitHub Issue Sync

Sync GitHub Issues to local tasks before triage analysis:

```bash
flow task sync --source github --repo owner/repo   # Sync open issues
flow task sync --source github --label bug          # Sync only bugs (auto-detect repo)
flow task sync --source github --dry-run            # Preview without writing
flow task sync --source github --overwrite          # Re-sync existing tasks
```

Issue → Task mapping: `[GH#42]` prefix, auto-classify type/priority from labels, `--external-ref` for dedup.
See `flow task sync --help` for full options.

## Directory Structure (v3 Active)

```
.trae/skills/team-flow/
├── SKILL.md              ← You are here (v3 entry point)
├── CONSENSUS.md          # v3 共识文件
├── BOUNDARY.md           # v3 架构边界
├── skills/
│   ├── team-flow-v3-create/   # Flow creation/refinement
│   │   ├── SKILL.md
│   │   └── references/        # Detailed reference docs
│   └── team-flow-v3-exec/     # Flow execution
│       └── SKILL.md
└── v2/                        # v2 fallback (not loaded by v3)

# Team templates (embedded, not installed to project)
teams/
├── dev-team/
│   ├── team.json              # 团队元数据 (id, name, flows, default_flow)
│   └── flows/                 # 8 个软件开发流程
├── content-team/
│   ├── team.json
│   └── flows/                 # 3 个内容创作流程
├── game-team/
│   ├── team.json
│   └── flows/
├── trading-team/
│   ├── team.json
│   └── flows/
└── skill-team/
    ├── team.json
    └── flows/
```

## ⚠️ v2 Legacy Directories (DO NOT LOAD)

The following directories under this skill directory are **v2 legacy** and should NOT be loaded by v3:

| Directory | Status | Action |
|-----------|--------|--------|
| `prompts/` | v2 legacy | Complete copy exists in `v2/prompts/` |
| `workflows/` | v2 legacy | Complete copy exists in `v2/workflows/` |
| `templates/` | v2 legacy | Complete copy exists in `v2/templates/` |
| `references/` | v2 legacy | Complete copy exists in `v2/references/` |
| `scripts/` | v2 legacy | Complete copy exists in `v2/scripts/` |
| `skills/team-flow-build/` | v2 legacy | Copy exists in `v2/skills/` |
| `skills/team-flow-check-impl/` | v2 legacy | Copy exists in `v2/skills/` |
| `skills/team-flow-design/` | v2 legacy | Copy exists in `v2/skills/` |
| `skills/team-flow-evolve/` | v2 legacy | Copy exists in `v2/skills/` |
| `skills/team-flow-git/` | v2 legacy | Copy exists in `v2/skills/` |
| `skills/team-flow-review/` | v2 legacy | Copy exists in `v2/skills/` |

**Note**: Team templates have been moved to the `teams/` directory (project root level), separate from project rules.

**v2 Fallback**: To revert to v2, point SKILL.md to `v2/SKILL.md` instead.
