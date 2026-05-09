# team-flow SKILL.md — Entry Point

> **Version**: 4.3 | **Date**: 2026-05-08
> **Core change**: TRIAGE-INBOX 统一入口 + 拆分机制

## Status Line (MANDATORY — Highest Priority)

Every AI response MUST start with:

```
[Role: {role} | TaskPool: {id} | Phase: {phase} | Asset: {project-name}]
```

| Field | Values | Description |
|-------|--------|-------------|
| Role | Triage / TechLead / Dev / QA / PM / DevOps / Analysis / UIDesigner | **Role Switching**: Triage dispatches sub-agent → Role changes to sub-agent's role. Sub-agent completes → Role returns to Triage. |

**禁止**: Phase 是 `analyze`/`design`/`implement` 等时，TaskPool 还显示 `INBOX`。详见 `{TEAM_PATH}/{VERSION}/prompts/triage.md`。

## Role Switching (角色切换)

**Role 随子 Agent 执行而变化**：

```
[Triage] 分析 → 分发
    ↓
启动子 Agent → [Role: Tech Lead] / [Role: Dev] / [Role: Bugfix] / [Role: QA]
    ↓
子 Agent 执行 → Status Line 实时反映当前角色
    ↓
子 Agent 完成 → [Role: Triage] 汇总结果
```
| TaskPool | `INBOX` 或 task ID (e.g., `F014`, `B001`) | Current task ID |
| Phase | ready / analyze / design / implement / verify / review / -(N/A) | Current phase |
| Asset | project name (e.g., `orig-cms`) or -(N/A) | Current project |

**Example**:
```
[Role: Triage | TaskPool: abc-123 (INBOX) | Phase: -(N/A) | Asset: team-flow]
→ Session started, inbox active. All inputs go through INBOX.

[Role: Triage | TaskPool: abc-456 (F014) | Phase: ready | Asset: orig-cms]
→ Task split from inbox, awaiting classification confirmation.

[Role: Dev | TaskPool: abc-789 (F014) | Phase: implement | Asset: orig-cms]
→ Working on F014
```

**⛔ TRIAGE-INBOX**: Session 启动时，确保 task-pool.md 中有 `INBOX` 条目作为统一入口。所有输入在 Inbox 中分析，**分析完成后必须拆分**。

**Phase 与 TaskPool 同步规则**：

| 当前 Phase | TaskPool 来源 | 说明 |
|-----------|--------------|------|
| `-(N/A)` 或 `triaging` | `INBOX` | Triage 在分析 |
| `ready` 及以后 | `F014` / `B001` 等 | 已拆分到正式 task |

**禁止**: Phase 是 `analyze`/`design`/`implement` 等时，TaskPool 还显示 `INBOX`。详见 `{TEAM_PATH}/{VERSION}/prompts/triage.md`。

## Version Detection (FIRST ACTION)

Read `.team/version` to determine active version. This is the FIRST thing to do on every session.

```
If .team/version = "v1" → Load v1 rules (task-pool workflow)
If .team/version = "v2" → Load v2 rules (beads-native workflow)
If .team/version missing → STOP, ask user to run `flow init`
```

## Version Management

Three operations only:

1. **`flow init --v1`** or **`flow init --v2`** → Install specified version rules
2. **`flow migrate`** → Upgrade v1 → v2 (converts task-pool data to beads, updates .team/version)
3. **`flow ver`** → Show current version + auto-detect low version (v1 suggests `flow migrate`)

```
init --v2 (default)  →  .team/version = "v2"  →  beads workflow
init --v1            →  .team/version = "v1"  →  task-pool workflow
migrate              →  v1 → v2 upgrade        →  .team/version = "v2"
ver                  →  show status            →  v1 shows upgrade hint
```

No version switching. No dual installation. One version per project.

## Migration (v1 → v2)

```bash
flow migrate          # Migrate task-pool.md data to beads
```

This converts v1 task-pool entries into beads issues. After migration, `.team/version` is automatically updated to v2.

## Path Variables

```yaml
{TEAM_PATH}:     (auto-detected, see below)
{PROJECT_PATH}:  (current working directory)
{VERSION}:       (read from .team/version)
{DOCS_PATH}:     (read from .team/project.md docs_path, fallback: _docs/{project-name}/ or .team/docs/)
```

Skill path auto-detection order:
1. `~/.agents/skills/team-flow/` (global, npx skills add)
2. `~/.trae-cn/skills/team-flow/` (Trae global)
3. `~/.claude/skills/team-flow/` (Claude global)
4. `.trae/skills/team-flow/` (Trae project-level)
5. `.claude/skills/team-flow/` (Claude project-level)
6. `.agents/skills/team-flow/` (project-level fallback)

## Human-Readable Export (v2)

beads is AI-managed. Humans need readable exports.

**Configuration** (in `.team/project.md`):
```yaml
## Paths
docs_path: _docs/{project-name}/    # User-configurable
```

**Export rules**:
1. Triage exports task status to `{DOCS_PATH}/task-pool.md` after every status change
2. CLI command: `flow export` — manual export anytime
3. Export format: markdown table (same as v1 task-pool.md format)
4. If `docs_path` not configured: fallback to `.team/docs/`

**Resolution order**:
```
1. .team/project.md → docs_path value
2. _docs/{project-name}/           (default convention)
3. .team/docs/                     (minimal fallback)
```

This matches the pattern: `_docs/{project-name}/` for project docs.

## Quick Start (First Session)

```bash
# Check version and status
flow version

# v2: Check beads status
bd ready --json

# v1: Read task pool
cat .team/task-pool.md
```

## Critical Rules

### Efficiency Rules

1. **Hot start**: Load `{TEAM_PATH}/{VERSION}/prompts/triage.md` for classification or execution.
2. **No redundant checks**: Unless `.team/` is missing or user requests, do not re-run "version check" or "project init" logic.
3. **Heartbeat**: Role response first line: `[Role: {role} | Version: {v1|v2} | TaskPool: {status} | Phase: {phase}]`
   - Role: Triage / Dev / QA / `-` (no role loaded)
   - Version: `v1` / `v2`
   - TaskPool: `bd ready` / `task-pool` / `-(N/A)`
   - Phase: `Phase N` / `blocked` / `-(N/A)`
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

#### Rule 10: Real scenario verification - mock test passing != functionality available
- For bugs involving API/permissions/state/interaction, **must perform real scenario verification** (HTTP request / page-level), not just mock tests
- Backend bugs: at minimum use httptest to send real HTTP requests, not just test UseCase
- Frontend bugs: at minimum verify page renders normally + core interactions work, not just test components
- R-iteration > 4 must force pause, ask user to confirm fix direction

## v2 Task Lifecycle

```
bd create "Title" -t bug|feature|task -p 0-4 --json
    ↓
bd update <id> --claim   (status → in_progress)
    ↓
bd update <id> --notes "COMPLETED: ... IN PROGRESS: ..."
    ↓
bd close <id> --reason "Done" --json
```

## v1 Task Lifecycle

```
Triage adds row to .team/task-pool.md with status "Todo"
    ↓
Dev claims task → status "Doing"
    ↓
Dev completes → status "Review"
    ↓
QA/PM verifies → status "Archived"
```

## Agent Mapping

| Role | subagent_type | Trigger Keywords | v1 Prompt | v2 Prompt |
|------|--------------|-----------------|-----------|-----------|
| Triage | (main agent) | all | v1/prompts/triage.md | v2/prompts/triage.md |
| Tech Lead | tech-lead-architect | implement/add/develop/design/feature | v1/prompts/tech-lead.md | v2/prompts/tech-lead.md |
| Dev (Backend) | developer-engineer | backend/API/database/Go | v1/prompts/dev-backend.md | v2/prompts/dev-backend.md |
| Dev (Frontend) | developer-engineer | frontend/React/component/page | v1/prompts/dev-frontend.md | v2/prompts/dev-frontend.md |
| Bugfix | bugfix-expert | Bug/error/crash/exception/fix | v1/prompts/bugfix.md | v2/prompts/bugfix.md |
| QA | qa-engineer | test/verify/quality | v1/prompts/qa-engineer.md | v2/prompts/qa-engineer.md |
| PM | pm-documenter | requirement/PRD/product/acceptance | v1/prompts/pm.md | v2/prompts/pm.md |
| Analysis | analysis-expert | research/analyze/compare/evaluate | v1/prompts/analysis.md | v2/prompts/analysis.md |
| DevOps | devops-engineer | deploy/CI/CD/Docker/K8s/ops | v1/prompts/devops.md | v2/prompts/devops.md |
| UI Designer | ui-designer | UI/interface/design/component/style | v1/prompts/ui-designer.md | v2/prompts/ui-designer.md |
| Framework Architect | tech-lead-architect | framework/architect/module design | v1/prompts/framework-architect.md | v2/prompts/framework-architect.md |

## Sub-agent Prompt Template

```
You are {role_name}, executing task {task_id}: {task_description}

Version: {VERSION}
Rules: {TEAM_PATH}/{VERSION}/prompts/{role}.md
Shared protocol: {TEAM_PATH}/{VERSION}/workflows/shared.md
Task tracking: bd CLI (v2) or task-pool.md (v1)

After completion:
1. v2: bd update <id> --notes "COMPLETED: ... IN PROGRESS: ..."
   v1: update task-pool.md status
2. Set suggested next role
3. Report deliverables list
```

## Three-Layer Gates

### Layer 1: Entry Gate (Before Starting)

```
Role triggered
    │
    ├── v2: Issue exists in beads? → bd ready --json → continue
    │   v1: Issue exists in task-pool.md? → continue
    │   └── Not found? → ⛔ Reject, suggest Triage create
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
    └── v2: beads issue updated? → continue
        v1: task-pool.md updated? → continue
        └── Not updated? → ⛔ Update before proceeding
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
    └── v2: beads issue closable? → bd close
        v1: task-pool.md status → Archived
        └── Not ready? → Update with remaining items
```

## HARD CONSTRAINTS

- **Delete Permission: DISABLED** — Never delete files unless explicitly asked
- **Framework ≠ Project** — Project problems solved at project layer, never patch framework
- **No secrets in code** — Never expose or log secrets/keys
- **NEVER commit unless user asks** — Explicit confirmation required
- **NEVER use PowerShell Set-Content / Out-File** — these add UTF-8 BOM, corrupting source files
  - Use Write tool (built-in, guarantees UTF-8 without BOM)
  - Use SearchReplace tool (built-in, precise replacement)
  - If must use command line: `[System.IO.File]::WriteAllText('path', $content, [System.Text.UTF8Encoding]::new($false))`

## Core Principle: Skill ≠ Project

> Skill rules (installed at skill path) are cross-project shared and read-only.
> Project problems must be solved at the project level, never by patching skill rules.

```
Skill path/        = AI framework rules (cross-project shared, read-only)
                   → Only modifiable when developing the SKILL itself

.team/project.md   = Project constraints (Triage reads and follows)
.team/version      = Active version (v1 or v2)
.ai/               = AI-generated temporary files (safe to delete)
```

**Project problems → Solve at project layer, NEVER patch framework layer**

## Entry Point for AI

When loaded by the AI runtime:

1. Read `.team/version` → determine v1 or v2
2. Load `{TEAM_PATH}/{VERSION}/prompts/triage.md` — if your role is triage
3. Load `{TEAM_PATH}/{VERSION}/workflows/shared.md` — always (layer 1/2 gates)
4. Load `{TEAM_PATH}/{VERSION}/workflows/roles/<role>-standards.md` — role-specific standards
