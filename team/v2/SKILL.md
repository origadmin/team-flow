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
> **Core change**: Frontmatter standardization + path resolution + flow task unification

## Status Line (MANDATORY — Highest Priority)

Every AI response MUST start with: `[Role: {role} | TaskPool: {beads-id}#{cr-index} | Phase: {phase} | Asset: {PROJECT-basename}]`

| Field | Values | Description |
|-------|--------|-------------|
| Role | Triage / TechLead / Dev / QA / PM / DevOps / Analysis / UIDesigner | Changes when sub-agent executes, returns to Triage on completion |
| TaskPool | {beads-id}#{cr-index} (e.g., `team-flow-6x9.15#5`) | Current task + conversation index |
| Phase | ready / analyze / design / implement / verify / review / -(N/A) | Current phase |
| Asset | {PROJECT} basename from `flow config paths --json` | Current project |

**Asset 说明**:
- Asset = basename of {PROJECT} from `flow config paths --json`
- Example: working on team-flow → Asset: team-flow
- Example: no project context → Asset: -(N/A)

**TaskPool 说明**:
- {beads-id}: task ID from `flow task create`
- {cr-index}: current conversation record index, stored in task's cr-index label
- ⛔ Every conversation must have a TaskPool value, N/A is forbidden
- Append conversation: `flow task append {id} --speaker {role} --content "..."` → returns new index
- When Triage receives input, must immediately create/find task → Status Line shows beads-id#1

**⛔ Task-first**: Session 启动时，Triage 必须先创建/查找任务。每次对话都有 TaskPool 值。

**Phase 与 TaskPool 同步**: Phase 是 `analyze`/`design`/`implement` 等时，TaskPool 必须显示 `{beads-id}#{cr-index}`，禁止显示 N/A。

## Role Switching

```
[Triage] 分析 → 分发 → [Sub-agent: TechLead/Dev/QA/...] 执行 → [Triage] 汇总结果
```

Example: `[Role: Triage | TaskPool: team-flow-6x9.15#1 | Phase: analyze | Asset: team-flow]` → `[Role: Dev | TaskPool: team-flow-6x9.15#3 | Phase: implement | Asset: team-flow]` → `[Role: Triage | TaskPool: team-flow-6x9.15#5 | Phase: review | Asset: team-flow]`

## Overview

team-flow v2 is the beads-native evolution of team-flow. `flow task` CLI is the unified AI-facing command, installed and verified during `flow init`. It routes based on `.team/version`: v1→docs, v2→beads. Triage uses flow task exclusively for all task creation, tracking, and status updates. The `task-pool-export.md` file is a human-readable export (read-only).

## Task Command Availability

`flow task` is always available after `flow init` (which installs, verifies, and adds to PATH). It internally routes to the correct backend based on `.team/version` (v1→docs, v2→beads). If flow task not found on PATH:

```bash
# Verify flow task installation
flow doctor

# If flow task found but not on PATH, re-run init to fix PATH
flow init --force

# Or manually discover flow task path
where.exe flow task 2>$null; Get-ChildItem "$env:LOCALAPPDATA\Programs\flow\flow.exe" -ErrorAction SilentlyContinue
```

**flow task is ALWAYS installed** — `flow init` guarantees it. If shell can't find flow task, use the full path reported by flow doctor.

## Path Resolution

AI must resolve all paths at startup via `flow config paths --json`. This is the **single source of truth** for path variables — AI must never self-resolve relative paths.

**Startup sequence**:
```bash
flow config paths --json
```

**Path variables** (from `flow config paths --json` output):

| Variable | Description | Required |
|----------|-------------|----------|
| `{WORKSPACE}` | Multi-project workspace root | yes |
| `{PROJECT}` | Current project root directory | yes |
| `{DOCS_INTERNAL}` | Internal docs (team-only, not public) | no |
| `{DOCS_EXTERNAL}` | External docs (public, open-source documentation) | no |
| `{BEADS_DB}` | beads database directory | yes |
| `{TEAM_PATH}` | Skill installation path | yes |
| `{TMP_DIR}` | AI temporary files directory | no |

**Path anchor rules** (in `.team/config.yaml` or `.team/project.md`):

| Prefix | Anchor | Example | Resolves to |
|--------|--------|---------|-------------|
| `_` | workspace root | `_docs/orig-cms/` | `{WORKSPACE}/_docs/orig-cms/` |
| other | project root | `docs/` | `{PROJECT}/docs/` |
| not configured | — | — | path not available |
| not configured | project root | — | `{PROJECT}/.team/tmp/` (default) |

⚠️ If `docs_internal` or `docs_external` not in output → not available, AI must not use those paths.
⚠️ AI must never self-resolve relative paths. Always use flow-resolved absolute paths.
⚠️ AI must write temporary files to `{TMP_DIR}` only (default: .team/tmp/)
⚠️ Never create temp files in project root or workspace root
⚠️ Clean up {TMP_DIR} when session ends

**Legacy variables** (deprecated, use flow config paths instead):
- `{DOCS_PATH}` → replaced by `{DOCS_INTERNAL}`
- `{PROJECT_PATH}` → replaced by `{PROJECT}`

**Current variables** (new, not legacy):
- `{TMP_DIR}` → AI temporary files directory (default: `{PROJECT}/.team/tmp/`)

## Task Command Routing

`flow task` is the unified AI-facing command. It routes based on `.team/version`:

| Stage | .team/version | `flow task create` routes to | `flow task list` reads from |
|-------|---------------|------------------------------|----------------------------|
| v1 | 1 | task-pool.md (docs) | task-pool.md |
| v2 | 2 | beads (.beads/) | beads (.beads/) |
| v3 | 3 | configurable (beads/git/other) | configurable |

**Auto-timestamps**: `flow task create` auto-sets `created_at`. `flow task update` auto-sets `updated_at`. AI never needs to manually write timestamps — the tool handles it.

**Conversation records**: `flow task append {id} --speaker {role} --content "..."` auto-increments cr-index and sets timestamp.

## Human-Readable Export

beads is AI-managed. Humans need readable exports.

- Source of truth: beads `.beads/` (via `flow task` commands)
- Human-readable export: `flow export > .team/task-pool-export.md`
- task-pool-export.md is **read-only** — never edit it to change task state

## Quick Start (First Session)

```bash
# 1. Check beads status
cd {PROJECT} && flow task ready --json

# 2. Read task pool
cd {PROJECT} && flow task list --status open --priority 0,1 --json | ConvertTo-Json -Depth 5

# 3. Read latest AI guidance
type {PROJECT}/.team/ai-context.md
```

## v2 vs v1 Key Differences

| Aspect | v1 | v2 |
|--------|----|----|
| Triage writes | task-pool.md (manual) | flow task create/update (beads) |
| Task source of truth | task-pool.md | beads `.beads/` |
| task-pool.md | Active + writable | Read-only export |
| ID format | F/B/C/A-NNN | `<beads-id>` (beads auto) |
| Status tracking | task-pool.md columns | flow task status |
| Multi-agent sync | task-pool.md git conflicts | Dolt git-native |
| Export | N/A | flow task export --format table |

## Directory Structure

```
{TEAM_PATH}/
├── SKILL.md              ← You are here
├── BOUNDARY.md           # Layer architecture definition
├── prompts/              # Role execution rules (11 roles)
├── workflows/            # Shared workflows + role standards
│   ├── shared.md         # Three-layer gates + lifecycle + handoff
│   ├── framework-workflow.md
│   ├── pre-flight.md
│   └── roles/            # Per-role standards (15 files)
├── templates/            # Document templates (14 files)
├── scripts/              # Automation
└── docs/                 # MIGRATION.md, BEADS_INTEGRATION.md
```

## Critical Rules

### Efficiency Rules

1. **Hot start**: Load `{TEAM_PATH}/prompts/triage.md` for classification or execution.
2. **No redundant checks**: Unless `.team/` is missing or user requests, do not re-run "version check" or "project init" logic.
3. **Heartbeat**: Role response first line: `[Role: {role} | TaskPool: {beads-id}#{cr-index} | Phase: {phase} | Asset: {PROJECT-basename}]`
   - Role: Triage / Dev / QA / `-` (no role loaded)
   - TaskPool: {beads-id}#{cr-index} (e.g., team-flow-6x9.15#5)
   - Phase: ready / analyze / design / implement / verify / review / -(N/A)
   - Asset: {PROJECT} basename from `flow config paths --json` (e.g., team-flow) or -(N/A)
4. **No-role guard**: When `Role: -`, only clarification answers allowed, no project modifications.

### Triage Loading Rules

- Core: `{TEAM_PATH}/prompts/triage.md` (always loaded)
- Clarify: `{TEAM_PATH}/prompts/triage-clarify.md` (load when classifying requirements)
- Verify: `{TEAM_PATH}/prompts/triage-verify.md` (load when bug returns, review needed, dispatching sub-agent, or managing session)

### Output Format Rules

When output exceeds 100 lines or covers 3+ independent topics:
- Use `<details><summary>标题</summary>` for each section
- Keep a 1-line summary outside the fold
- Active task status and next steps should NOT be folded

### Dispatch Guard (same priority as Regression Guard)

> **AI bypassing dispatch is the most common behavioral deviation. Triage must dispatch, never execute directly.**

#### Rule 1: Must dispatch after classification
- After intent recognition, **must use Task tool to start sub-agent**
- Never do it yourself after classification (read code, modify code, debug, write design docs)
- Only exception: clarification questions can be answered directly

#### Rule 2: Self-check before action
- Before any action, **must ask yourself**:
  ```
  Is this Triage's responsibility or sub-agent's?
  Triage: intent recognition, task creation, start sub-agent, update status, report results
  Sub-agent: read code, modify code, debug, design, analyze, deploy
  ```
- If sub-agent responsibility -> **stop, use Task tool to dispatch**

#### Rule 3: No "just doing it quickly"
- Even simple tasks **must be dispatched to sub-agent**
- "Just doing it" is overreach, not efficiency
- Simple tasks -> sub-agent executes faster (has dedicated prompts and tools)

#### Rule 4: Classification report + dispatch is atomic
- After classification report, **immediately start sub-agent**, do not wait for user confirmation
- Classification report and Task call must be in the same response

### Regression Guard (highest priority)

> See `{TEAM_PATH}/workflows/shared.md` §质量门（强制） and `{TEAM_PATH}/workflows/roles/bugfix-standards.md` for full details.

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

### v2 Task Lifecycle

```
flow task create "Title" -t bug|feature|task -p 0-4 --json
    ↓
flow task update <id> --claim   (status → in_progress)
    ↓
flow task update <id> --notes "COMPLETED: ... IN PROGRESS: ..."
    ↓
flow task close <id> --reason "Done" --json
```

### ID Mapping

- task ID (F001, B061, etc.) lives in the `external-ref` field of the beads issue
- Use flow task list --json | ConvertFrom-Json | Where-Object { $_.externalRef -match 'F001' } to find by task ID
- Export: flow task list --json includes externalRef for cross-reference

### DO NOT

- ❌ Edit task-pool.md manually during task operations
- ❌ Create tasks in task-pool.md that aren't also in beads
- ❌ Skip flow task dolt push after significant status changes
- ❌ Write to `{TEAM_PATH}/v1/` (framework-original, read-only reference)
- ❌ Dump deliverable content into beads notes — see `{TEAM_PATH}/workflows/shared.md` §文件空间定义
- ❌ Modify `{TEAM_PATH}/` rules to solve project-specific problems — use `.team/project.md §CONSTRAINTS` and `_docs/.../lessons/` instead

### Core Principle: Framework ≠ Project

> See `{TEAM_PATH}/workflows/shared.md` §文件空间定义 for full details.

```
{TEAM_PATH}/        = AI framework rules (cross-project, read-only) → Only modify when developing the skill itself
.team/project.md = Project constraints (Triage reads and follows)
_docs/.../lessons/ = Project lessons (Dev/Bugfix reads)
```

**Project problems → solve at project layer, never patch framework**

## Three-Layer Gates

### Layer 1: Entry Gate (Before Starting)

```
Role triggered
    │
    ├── Issue exists in beads? → flow task ready --json → continue
    │   └── Not found? → ⚠️ Reject, suggest Triage create via flow task create
    │
    ├── Issue type matches role? → continue
    │   └── Mismatch? → ⚠️ Hand off to correct role
    │
    └── Required docs loaded? → continue
        └── Missing? → ⚠️ Load before proceeding
```

### Layer 2: Phase Gate (Between Phases)

```
Phase transition
    │
    ├── Current phase deliverables complete? → continue
    │   └── Incomplete? → ⚠️ Complete before transitioning
    │
    ├── Tests passing? → continue
    │   └── Failing? → ⚠️ Fix before proceeding
    │
    └── beads issue updated? → continue
        └── Not updated? → ⚠️ flow task update before proceeding
```

### Layer 3: Completion Gate (Before Closing)

```
Task complete
    │
    ├── All deliverables produced? → continue
    │   └── Missing? → ⚠️ Produce before closing
    │
    ├── All tests passing? → continue
    │   └── Failing? → ⚠️ Fix before closing
    │
    ├── No regressions? → continue
    │   └── Regressions found? → ⚠️ Fix or document known issues
    │
    └── beads issue closable? → flow task close
        └── Not ready? → flow task update --notes with remaining items
```

### Completion Checklists

> See `{TEAM_PATH}/workflows/shared.md` §Layer 3: 完成门禁 for full checklists.

**Feature** (19 items): Code + tests + no regressions + API compatible + docs + SCOPE.md + beads updated
**Bugfix** (18 items): Root cause + data flow traced + fix targets cause + tests + real scenario verified + RCA.md + SCOPE.md + R-iteration ≤ 4

## Context Checkpoint

> See `{TEAM_PATH}/workflows/shared.md` §Session Protocol for session start/end/during templates.

**Every conversation turn must output**:
```
[Task Context] Task: {beads ID + title} | Phase: {phase} | Docs: {loaded} | Toolchain: {go|bun} | Mode: {v2-native} [/Task Context]
```

## Agent Mapping

| Role | subagent_type | Trigger Keywords | Prompt File |
|------|--------------|-----------------|-------------|
| Triage | (main agent) | all | prompts/triage.md |
| Tech Lead | tech-lead-architect | 实现/新增/开发/支持/设计/功能 | prompts/tech-lead.md |
| Dev (Backend) | developer-engineer | 后端/API/数据库/Go | prompts/dev-backend.md |
| Dev (Frontend) | developer-engineer | 前端/React/组件/页面 | prompts/dev-frontend.md |
| Bugfix | bugfix-expert | Bug/报错/崩溃/异常/问题/修复 | prompts/bugfix.md |
| QA | qa-engineer | 测试/验证/质量 | prompts/qa-engineer.md |
| PM | pm-documenter | 需求/PRD/产品/验收 | prompts/pm.md |
| Analysis | analysis-expert | 调研/分析/对比/评估 | prompts/analysis.md |
| DevOps | devops-engineer | 部署/CI/CD/Docker/K8s/运维 | prompts/devops.md |
| UI Designer | ui-designer | UI/界面/设计稿/组件/样式 | prompts/ui-designer.md |
| Framework Architect | tech-lead-architect | 框架/架构师/模块设计 | prompts/framework-architect.md |

## MILESTONES Sync

| task Status | beads Status | Meaning |
|-------------|-------------|---------|
| Todo | open | Not started |
| Doing | in_progress | In progress |
| Review | in_progress (notes: "awaiting review") | Awaiting confirmation |
| Archived | closed | Completed |
| Blocked Bug | open (labels: ["blocked"]) | Blocked by bug |
| Change Evaluating | open (labels: ["evaluating"]) | Change under evaluation |

## Required Config Files per Role

| Role | Must Load |
|------|-----------|
| Triage | shared.md, triage-standards.md |
| Tech Lead | shared.md, development-standards.md, architecture-standards.md |
| Dev Backend | shared.md, development-standards.md, specialized-tests.md |
| Dev Frontend | shared.md, development-standards.md, frontend-specialized-tests.md |
| Bugfix | shared.md, bugfix-standards.md, bug-test-template.md |
| QA | shared.md, test-levels.md, test-standards.md |
| PM | shared.md, requirements-standards.md, prd-template.md |
| Analysis | shared.md, analysis-standards.md |
| DevOps | shared.md, devops-standards.md |
| UI Designer | shared.md, ui-standards.md, ui-design-template.md |
| Framework Architect | shared.md, architecture-standards.md, framework-workflow.md |

## HARD CONSTRAINTS

- **Delete Permission: DISABLED** — Never delete files unless explicitly asked
- **Framework ≠ Project** — Project problems solved at project layer, never patch framework
- **No secrets in code** — Never expose or log secrets/keys
- **⚠️ NEVER commit unless user asks** — Explicit confirmation required
- **⛔ NEVER use PowerShell Set-Content / Out-File** — these add UTF-8 BOM, corrupting source files
  - Use Write tool (built-in, guarantees UTF-8 without BOM)
  - Use SearchReplace tool (built-in, precise replacement)
  - If must use command line: `[System.IO.File]::WriteAllText('path', $content, [System.Text.UTF8Encoding]::new($false))`

## Sub-agent Prompt Template

> See `{TEAM_PATH}/workflows/shared.md` §Role Handoff Protocol for detailed handoff format.

```
You are {role_name}, executing task {task_id}: {task_description}
Rules: {TEAM_PATH}/prompts/{role}.md | Shared: {TEAM_PATH}/workflows/shared.md | Tracking: flow task CLI
⛔ v2: Never edit task-pool.md manually. All status updates via flow task CLI.
After completion: flow task update <id> --notes "COMPLETED: ..." --add-label phase:review → Report deliverables → Return to Triage
```

## Core Protocol

- Shared protocol: `{TEAM_PATH}/workflows/shared.md`
- Triage rules: `{TEAM_PATH}/prompts/triage.md`
- Entry point: `{TEAM_PATH}/SKILL.md`

## Entry Point for AI

When loaded by the AI runtime, read this file to understand v2 structure, then load:

1. `prompts/triage.md` — if your role is triage
2. `workflows/shared.md` — always (layer 1/2 gates)
3. `workflows/roles/<role>-standards.md` — role-specific standards

For beads commands reference: see `docs/BEADS_INTEGRATION.md`
