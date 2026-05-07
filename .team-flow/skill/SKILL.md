# _team v2 SKILL.md — Entry Point

> **Version**: v2.0 | **Date**: 2026-05-02
> **Core change**: Triage + Task Management → 100% beads

## Overview

_team v2 is the beads-native evolution of _team. Triage uses `bd` CLI exclusively for all task creation, tracking, and status updates. The `task-pool.md` file becomes a read-only export for human-readable reference and AI context injection.

## Path Variables

```yaml
{TEAM_PATH}:     .trae/skills/team-flow/       # Skill installation path (IDE-relative)
{PROJECT_PATH}:  projects/orig-cms/            # Current project
{BEADS_DB}:      projects/orig-cms/.beads/     # beads data directory
{DOCS_INTERNAL}: framework/_docs/orig-cms/     # Internal project docs
{VERSION}:       .team/version                 # Read for active version (v1 or v2)
```

## Quick Start (First Session)

```bash
# 1. Check beads status
cd {PROJECT_PATH} && bd ready --json

# 2. Read task pool
cd {PROJECT_PATH} && bd list --status open --priority 0,1 --json | ConvertTo-Json -Depth 5

# 3. Read latest AI guidance
type {PROJECT_PATH}/.team/ai-context.md
```

## v2 vs v1 Key Differences

| Aspect | v1 | v2 |
|--------|----|----|
| Triage writes | task-pool.md (manual) | `bd create/update` (beads) |
| Task source of truth | task-pool.md | beads `.beads/` |
| task-pool.md | Active + writable | Read-only export |
| ID format | F/B/C/A-NNN | cms-xxx (beads auto) |
| Status tracking | task-pool.md columns | `bd status` |
| Multi-agent sync | task-pool.md git conflicts | Dolt git-native |
| Export | N/A | `bd export --format table` |

## Directory Structure

```
_team/v2/
├── SKILL.md              ← You are here
├── BOUNDARY.md           # Layer architecture definition
├── prompts/              # Role execution rules (v2, beads-native)
│   ├── triage.md
│   ├── dev.md
│   ├── dev-backend.md
│   ├── dev-frontend.md
│   ├── bugfix.md
│   ├── tech-lead.md
│   ├── qa-engineer.md
│   ├── pm.md
│   ├── analysis.md       # Analysis Expert (v2 补足)
│   ├── devops.md         # DevOps Engineer (v2 补足)
│   ├── ui-designer.md    # UI Designer (v2 补足)
│   └── framework-architect.md  # Framework Architect (v2 补足)
├── workflows/            # Shared workflows + role standards
│   ├── shared.md
│   ├── framework-workflow.md    # Framework dev workflow (v2 补足)
│   ├── pre-flight.md            # Pre-flight checks (v2 补足)
│   ├── roles/
│   │   ├── triage-standards.md
│   │   ├── development-standards.md
│   │   ├── bugfix-standards.md
│   │   ├── devtestops.md
│   │   ├── test-levels.md
│   │   ├── specialized-tests.md
│   │   ├── frontend-specialized-tests.md
│   │   ├── analysis-standards.md      # (v2 补足)
│   │   ├── architecture-standards.md  # (v2 补足)
│   │   ├── devops-standards.md        # (v2 补足)
│   │   ├── ui-standards.md            # (v2 补足)
│   │   ├── review-standards.md        # (v2 补足)
│   │   ├── requirements-standards.md  # (v2 补足)
│   │   ├── test-standards.md          # (v2 补足)
│   │   └── checklist.md               # (v2 补足)
│   └── meta/
│       └── TEAM_ROLES.md              # (v2 补足)
├── templates/            # Document templates
│   ├── README.md
│   ├── feature-test-template.md
│   ├── bug-test-template.md
│   ├── frontend-feature-test-template.md
│   ├── frontend-bug-test-template.md
│   ├── scope-template.md
│   ├── test-report-template.md
│   ├── closed-loop-verification-template.md
│   ├── gherkin-feature-template.md
│   ├── architecture-template.md
│   ├── prd-template.md
│   ├── ui-design-template.md
│   ├── api-issue-template.md
│   ├── user-story-template.md
│   └── bug-index-template.md
├── scripts/              # Automation
│   ├── migrate-tasks.ps1
│   └── export-task-pool.ps1
├── docs/
│   ├── MIGRATION.md
│   └── BEADS_INTEGRATION.md
└── BOUNDARY.md
```

## Critical Rules

### Efficiency Rules

1. **Hot start**: Load `{TEAM_PATH}/prompts/triage.md` for classification or execution.
2. **No redundant checks**: Unless `.team/` is missing or user requests, do not re-run "version check" or "project init" logic.
3. **Heartbeat**: Role response first line: `[Role: {role} | TaskPool: {status} | Phase: {phase} | Asset: {status}]`
   - Role: Triage / Dev / QA / `-` (no role loaded)
   - TaskPool: `bd ready` / `task-pool` / `-(N/A)`
   - Phase: `Phase N` / `blocked` / `-(N/A)`
   - Asset: `ok` / `missing` / `-(N/A)`
4. **No-role guard**: When `Role: -`, only clarification answers allowed, no project modifications.

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

> **AI breaking existing functionality is the most frequent issue. These rules have priority over all development instructions.**

#### Rule 1: Must read before modifying
- Before modifying any existing code file, **must read the complete file first**
- Never modify based on partial view or diff range only

#### Rule 2: Must search before changing exported symbols
- Before changing exported function/method/interface/type, **must search all reference points**
- Command: `grep -r "SymbolName" --include="*.go"` or equivalent
- Changing exported symbols without searching references -> forbidden

#### Rule 3: Layered testing - local first, then full
- **During TDD cycle**: only run current module tests
  - Backend: `go test ./internal/features/xxx/...`
  - Frontend: `bun run test -- --testPathPattern="xxx"`
- **After modification (regression verification)**: run full tests
  - Backend: `go test ./...`
  - Frontend: `bun run test`
- **Zero regression tolerance**: any previously passing test fails -> stop, fix or rollback
- **High-frequency modification**: after every N local tests, run full suite (recommended N=3)

```
Development flow:
  TDD red->green->refactor: local tests (seconds)
       ↓ repeat N times
  Periodic regression:     full tests (minutes)
       ↓
  Completion gate:         full tests (must pass)
       ↓
  Frontend extra:         bun run typecheck (must pass)
```

#### Rule 4: Breaking changes must be compatible
- Changing interface signatures, deleting methods, modifying return value structures = Breaking Change
- Breaking changes must provide compatibility solution (new function / versioned interface / deprecation marker)
- Never change interface without updating all callers

#### Rule 5: Add tests before modifying untested code
- Before modifying existing code without test coverage, **add tests first**
- Only modify after tests pass

#### Rule 6: TanStack Router parent routes with child routes must use Outlet
- When a route has child routes, parent **must render `<Outlet/>`**, otherwise child route URL matches but page doesn't switch
- Correct: `xxx/route.tsx` -> `<Outlet/>` + `xxx/index.tsx` -> list component
- Wrong: `xxx.tsx` -> render component directly (child routes cannot display)
- Always check for child routes when creating new routes

#### Rule 7: Phenomenon first - Bug investigation must confirm phenomenon first
- When user reports visual/interaction issues, **must check rendering code to confirm phenomenon first**, never skip to data layer
- Investigation order: **Rendering layer -> API layer -> Business layer -> Data layer** (top-down)
- Never assume the problem is in the data layer without confirming the phenomenon

#### Rule 8: UI code self-check - must check for duplicate rendering
- After writing UI components, **must check if same data is rendered multiple times**
- **Compilation passing != logic correct**, UI code must be manually reviewed
- Check method: `grep "formatDate\|formatDuration\|t('"` target file, confirm each data rendered once

#### Rule 9: Data flow tracing - Bug fixes must trace runtime data flow
- For bugs involving API/permissions/state/interaction, **must trace complete data flow from source to sink**, never "guess where the problem is and fix there"
- Trace steps: define start/end -> list each step -> verify each step to find breakpoint -> fix breakpoint -> verify complete chain
- **Compilation passing != fix complete**, mock test passing != functionality available
- Detailed spec: `{TEAM_PATH}/workflows/roles/bugfix-standards.md`

#### Rule 10: Real scenario verification - mock test passing != functionality available
- For bugs involving API/permissions/state/interaction, **must perform real scenario verification** (HTTP request / page-level), not just mock tests
- Backend bugs: at minimum use httptest to send real HTTP requests, not just test UseCase
- Frontend bugs: at minimum verify page renders normally + core interactions work, not just test components
- R-iteration > 4 must force pause, ask user to confirm fix direction

### v2 Task Lifecycle

```
bd create "Title" -t bug|feature|task -p 0-4 --json
    ↓
bd update <id> --claim   (status → in_progress)
    ↓
bd update <id> --notes "COMPLETED: ... IN PROGRESS: ..."
    ↓
bd close <id> --reason "Done" --json
```

### ID Mapping

- _team ID (F001, B061, etc.) lives in the `external-ref` field of the beads issue
- Use `bd list --json | ConvertFrom-Json | Where-Object { $_.externalRef -match 'F001' }` to find by _team ID
- Export: `bd list --json` includes `externalRef` for cross-reference

### DO NOT

- ❌ Edit task-pool.md manually during task operations
- ❌ Create tasks in task-pool.md that aren't also in beads
- ❌ Skip `bd dolt push` after significant status changes
- ❌ Write to `_team/v1/` (framework-original, read-only reference)
- ❌ Dump deliverable content into beads notes — write to independent files (RCA.md, SPEC.md, etc.)
- ❌ Modify `_team/` rules to solve project-specific problems — use `.team/project.md §CONSTRAINTS` and `_docs/.../lessons/` instead

### Core Principle: Framework ≠ Project

> **v1 教训**: AI 在修复项目问题时，把项目规则写入了 `_team/`（框架层），等于篡改了 AI 的大脑来适应项目问题。

```
_team/           = AI 框架规则（跨项目共享，只读）
                   → 只有开发 AI SKILL 本身时才能修改

.team/project.md = 项目约束（Triage 读取并遵守）
_docs/.../lessons/ = 项目经验教训（Dev/Bugfix 读取）
```

**项目问题 → 在项目层解决，绝不在框架层打补丁**

## Three-Layer Gates

### Layer 1: Entry Gate (Before Starting)

```
Role triggered
    │
    ├── Issue exists in beads? → bd ready --json → continue
    │   └── Not found? → ⛔ Reject, suggest Triage create via bd create
    │
    ├── Issue type matches role? → continue
    │   └── Mismatch? → ⛔ Hand off to correct role
    │
    └── Required docs loaded? → continue
        └── Missing? → ⛔ Load before proceeding
```

### Layer 2: Phase Gate (Between Phases)

```
Phase transition
    │
    ├── Current phase deliverables complete? → continue
    │   └── Incomplete? → ⛔ Complete before transitioning
    │
    ├── Tests passing? → continue
    │   └── Failing? → ⛔ Fix before proceeding
    │
    └── beads issue updated? → continue
        └── Not updated? → ⛔ bd update before proceeding
```

### Layer 3: Completion Gate (Before Closing)

```
Task complete
    │
    ├── All deliverables produced? → continue
    │   └── Missing? → ⛔ Produce before closing
    │
    ├── All tests passing? → continue
    │   └── Failing? → ⛔ Fix before closing
    │
    ├── No regressions? → continue
    │   └── Regressions found? → ⛔ Fix or document known issues
    │
    └── beads issue closable? → bd close
        └── Not ready? → bd update --notes with remaining items
```

### Feature Completion Checklist (19 items)

```
- [ ] Code implemented per SPEC.md requirements
- [ ] Unit tests written and passing (TDD red→green)
- [ ] Integration tests passing (if applicable)
- [ ] No lint/typecheck errors
- [ ] No regressions in existing tests
- [ ] API contract unchanged or backward-compatible
- [ ] Error handling covers edge cases
- [ ] Logging/observability added (if applicable)
- [ ] Configuration documented (if new config introduced)
- [ ] Migration script provided (if DB schema changed)
- [ ] Frontend: No duplicate rendering (铁律8)
- [ ] Frontend: Responsive rules verified
- [ ] Frontend: Accessibility checked
- [ ] Documentation updated (if applicable)
- [ ] SCOPE.md written
- [ ] beads issue updated with deliverables
- [ ] Code reviewed (self-review or peer)
- [ ] Pre-modification checklist completed (铁律1-2)
- [ ] Impact radius verified (flow graph impact or Grep)
```

### Bugfix Completion Checklist (18 items)

```
- [ ] Root cause identified (not just symptom)
- [ ] Data flow traced from source to sink (铁律9)
- [ ] Fix targets root cause, not symptom
- [ ] Unit test reproduces the bug (TDD red→green)
- [ ] Integration test verifies end-to-end (if applicable)
- [ ] No regressions in existing tests
- [ ] No new lint/typecheck errors
- [ ] API contract unchanged or backward-compatible
- [ ] Error message is actionable
- [ ] Frontend: Visual issue confirmed before fix (铁律7)
- [ ] Frontend: No duplicate rendering after fix (铁律8)
- [ ] Frontend: MSW mock updated (if API changed)
- [ ] Real scenario verified (not just mock) (铁律10)
- [ ] RCA.md written (for P0/P1 bugs)
- [ ] SCOPE.md written
- [ ] beads issue updated with root cause and fix
- [ ] R-iteration count ≤ 4 (if >4, pause for user confirmation)
- [ ] Impact radius verified (flow graph impact or Grep)
```

## Context Checkpoint

> **Every conversation turn must output the following context block.**

```
[Task Context]
  Task: {beads issue ID + title}
  Phase: {ready|analyze|design|implement|verify|review}
  Required Docs: {list files loaded}
  Toolchain: {go|bun|python} {version}
  Mode: {v1-compat|v2-native}
[/Task Context]
```

This ensures AI maintains context across turns and prevents task drift.

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

| _team Status | beads Status | Meaning |
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
- **NEVER commit unless user asks** — Explicit confirmation required
- **NEVER use PowerShell Set-Content / Out-File** — these add UTF-8 BOM, corrupting source files
  - Use Write tool (built-in, guarantees UTF-8 without BOM)
  - Use SearchReplace tool (built-in, precise replacement)
  - If must use command line: `[System.IO.File]::WriteAllText('path', $content, [System.Text.UTF8Encoding]::new($false))`

## Sub-agent Prompt Template

```
You are {role_name}, executing task {task_id}: {task_description}

Rules: {TEAM_PATH}/prompts/{role}.md
Shared protocol: {TEAM_PATH}/workflows/shared.md
Task tracking: bd CLI (v2) or task-pool.md (v1)
Docs: {DOCS_INTERNAL}

After completion:
1. bd update <id> --notes "COMPLETED: ... IN PROGRESS: ..." (v2)
   or update task-pool.md status (v1)
2. Set suggested next role
3. Report deliverables list
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