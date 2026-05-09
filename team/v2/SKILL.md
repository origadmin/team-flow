---
name: team-flow
version: 2.3
description: |
  Multi-agent AI collaboration framework with beads-based task management.
  Provides Triage→TechLead→Dev→QA pipeline, three-layer gates, and 11 role definitions.
  Task management via flow task CLI (routes to beads v2 / docs v1).
  Path resolution via flow config paths --json.
  Compatible: Go backend + React/TypeScript frontend projects.
tools:
  - name: task
    required: true
    commands: [create, update, close, list, show, ready, append, dep, dolt, export]
config:
  paths_source: "flow config paths --json"
  config_file: ".team/config.yaml"
  project_file: ".team/project.md"
evolution:
  v2: "beads migration, task management centralization"
  v3: "flow as center, tools/plugins/rules unified management"
---

# team-flow v2 SKILL.md — Entry Point

> **Version**: v2.3 | **Date**: 2026-05-09

## Status Line (MANDATORY — Highest Priority)

Every AI response MUST start with: `[Role: {role} | TaskPool: {beads-id}#{cr-index} | Phase: {phase} | Asset: {PROJECT-basename}]`

| Field | Values | Description |
|-------|--------|-------------|
| Role | Triage / TechLead / Dev / QA / PM / DevOps / Analysis / UIDesigner | Changes when sub-agent executes, returns to Triage on completion |
| TaskPool | {beads-id}#{cr-index} (e.g., `team-flow-6x9.15#5`) | Current task + conversation index |
| Phase | ready / analyze / design / implement / verify / review / -(N/A) | Current phase |
| Asset | {PROJECT} basename from `flow config paths --json` | Current project |

**⛔ Task-first**: Session 启动时，Triage 必须先创建/查找任务。每次对话都有 TaskPool 值，N/A is forbidden。

**Phase 与 TaskPool 同步**: Phase 是 `analyze`/`design`/`implement` 等时，TaskPool 必须显示 `{beads-id}#{cr-index}`，禁止显示 N/A。

## Role Switching

```
[Triage] 分析 → 分发 → [Sub-agent: TechLead/Dev/QA/...] 执行 → [Triage] 汇总结果
```

## Overview

team-flow v2 is the beads-native evolution. `flow task` CLI is the unified AI-facing command, installed and verified during `flow init`. It routes based on `.team/version`: v1→docs, v2→beads. Triage uses flow task exclusively for all task creation, tracking, and status updates.

## Path Resolution

Resolve all paths at startup via `flow config paths --json`. See `references/path-resolution.md` for full details.

**Key variables**: `{WORKSPACE}`, `{PROJECT}`, `{BEADS_DB}`, `{TEAM_PATH}`, `{TMP_DIR}`

⚠️ AI must never self-resolve relative paths. Always use flow-resolved absolute paths.

## Commands

`flow task` is the unified command. See `references/commands.md` for full reference.

**Essential commands**:

| Operation | Command |
|-----------|---------|
| Create task | `flow task create "Title" -t {bug\|feature\|task\|epic} -p {0-4} --parent {id}` |
| Claim task | `flow task update {id} --claim` |
| Close task | `flow task close {id} --reason "..."` |
| Show task | `flow task show {id}` |
| Find work | `flow task ready --json` |
| Append record | `flow task append {id} --speaker {role} --content "..."` |
| Resolve paths | `flow config paths --json` |

## Critical Rules

### Efficiency Rules

1. **Hot start**: Load `{TEAM_PATH}/prompts/triage.md` for classification or execution.
2. **No redundant checks**: Unless `.team/` is missing or user requests, do not re-run "version check" or "project init" logic.
3. **Heartbeat**: Role response first line: `[Role: {role} | TaskPool: {beads-id}#{cr-index} | Phase: {phase} | Asset: {PROJECT-basename}]`
4. **No-role guard**: When `Role: -`, only clarification answers allowed, no project modifications.

### Triage Loading Rules

- Core: `{TEAM_PATH}/prompts/triage.md` (always loaded)
- Clarify: `{TEAM_PATH}/prompts/triage-clarify.md` (load when classifying requirements)
- Verify: `{TEAM_PATH}/prompts/triage-verify.md` (load when bug returns, review needed, dispatching sub-agent, or managing session)

### Dispatch Guard

> **AI bypassing dispatch is the most common behavioral deviation. Triage must dispatch, never execute directly.**

1. After intent recognition, **must use Task tool to start sub-agent**
2. Never do it yourself after classification (read code, modify code, debug, write design docs)
3. Even simple tasks **must be dispatched to sub-agent**
4. Classification report and Task call must be in the same response

### Regression Guard (highest priority)

> See `{TEAM_PATH}/workflows/shared.md` §质量门（强制） for full details.

**Core principles** (summary):
1. Must read before modifying — never edit based on partial view
2. Must search before changing exported symbols
3. Layered testing — local first, then full
4. Breaking changes must be compatible
5. Add tests before modifying untested code
6. TanStack Router: parent routes with children must use `<Outlet/>`
7. Bug investigation: confirm phenomenon first (top-down)
8. UI code: check for duplicate rendering
9. Data flow tracing: trace runtime data flow, never guess
10. Real scenario verification: mock test passing ≠ functionality available

### DO NOT

- ❌ Edit task-pool.md manually during task operations
- ❌ Create tasks in task-pool.md that aren't also in beads
- ❌ Skip flow task dolt push after significant status changes
- ❌ Dump deliverable content into beads notes — see `{TEAM_PATH}/workflows/shared.md` §文件空间定义
- ❌ Modify `{TEAM_PATH}/` rules to solve project-specific problems — use `.team/project.md §CONSTRAINTS` instead

### Core Principle: Framework ≠ Project

```
{TEAM_PATH}/        = AI framework rules (cross-project, read-only) → Only modify when developing the skill itself
.team/project.md = Project constraints (Triage reads and follows)
_docs/.../lessons/ = Project lessons (Dev/Bugfix reads)
```

## Three-Layer Gates

### Layer 1: Entry Gate (Before Starting)

Issue exists in beads? → flow task ready → type matches role? → docs loaded?

### Layer 2: Phase Gate (Between Phases)

Deliverables complete? → Tests passing? → beads issue updated?

### Layer 3: Completion Gate (Before Closing)

All deliverables produced? → All tests passing? → No regressions? → beads issue closable?

> See `{TEAM_PATH}/workflows/shared.md` for full gate details and completion checklists.

## Agent Mapping

> See `references/agent-mapping.md` for full agent mapping, required config files per role, and MILESTONES sync.

## HARD CONSTRAINTS

- **Delete Permission: DISABLED** — Never delete files unless explicitly asked
- **Framework ≠ Project** — Project problems solved at project layer, never patch framework
- **No secrets in code** — Never expose or log secrets/keys
- **⚠️ NEVER commit unless user asks** — Explicit confirmation required
- **⛔ NEVER use PowerShell Set-Content / Out-File** — these add UTF-8 BOM, corrupting source files
- **⛔ Toolchain Gate** — Read `.team/project.md` TOOLCHAIN section before ANY build/test/lint command. Use exact commands from project.md, never default to npm. If project.md says `bun`, using `npm`/`pnpm`/`yarn` is a hard violation.

## Sub-agent Prompt Template

```
You are {role_name}, executing task {task_id}: {task_description}
Rules: {TEAM_PATH}/prompts/{role}.md | Shared: {TEAM_PATH}/workflows/shared.md | Tracking: flow task CLI
⛔ v2: Never edit task-pool.md manually. All status updates via flow task CLI.
After completion: flow task update <id> --notes "COMPLETED: ..." --add-label phase:review → Report deliverables → Return to Triage
```

## Directory Structure

```
{TEAM_PATH}/
├── SKILL.md              ← You are here
├── BOUNDARY.md           # Layer architecture definition
├── prompts/              # Role execution rules (11 roles)
├── workflows/            # Shared workflows + role standards
│   ├── shared.md         # Three-layer gates + lifecycle + handoff
│   └── roles/            # Per-role standards (15 files)
├── templates/            # Document templates
├── scripts/              # Automation
├── references/           # Detailed references (loaded on demand)
│   ├── commands.md       # Full command reference + routing
│   ├── path-resolution.md # Path variables and anchor rules
│   └── agent-mapping.md  # Agent mapping + role config + MILESTONES
└── docs/                 # Project docs: {PROJECT}/docs/v1/ and docs/v2/
```

## Entry Point for AI

When loaded by the AI runtime, read this file to understand v2 structure, then load:

1. `prompts/triage.md` — if your role is triage
2. `workflows/shared.md` — always (layer 1/2 gates)
3. `workflows/roles/<role>-standards.md` — role-specific standards
4. `references/<topic>.md` — when you need detailed reference on a topic
