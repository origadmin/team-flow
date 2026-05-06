# _team v2 SKILL.md — Entry Point

> **Version**: v2.0 | **Date**: 2026-05-02
> **Core change**: Triage + Task Management → 100% beads

## Overview

_team v2 is the beads-native evolution of _team. Triage uses `bd` CLI exclusively for all task creation, tracking, and status updates. The `task-pool.md` file becomes a read-only export for human-readable reference and AI context injection.

## Path Variables

```yaml
{TEAM_PATH}:     framework/_team/v2/          # This directory (framework-relative)
{PROJECT_PATH}:  projects/orig-cms/           # Current project
{BEADS_DB}:      projects/orig-cms/.beads/   # beads data directory
{DOCS_INTERNAL}:framework/_docs/orig-cms/  # Internal project docs
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

## Pre-Modification Checklist (铁律1-2)

Before modifying any existing code file:

1. **Read the complete file first** — Never modify based on partial view
2. **Search all references** — `Grep` for exported symbols before changing interfaces
3. **Run local tests** — Only current module tests during TDD cycle
4. **Check impact radius** — `flow graph impact [files]` or recursive `Grep`
5. **Verify no breaking changes** — If interface changed, provide compatibility layer

## Bug Post-Verification (铁律7-10)

After bugfix-expert returns, Triage must verify:

1. **Phenomenon confirmed first** — Check rendering code before analyzing data layer
2. **Data flow traced** — From source to sink, not just symptom
3. **Real scenario tested** — HTTP request or page-level, not just mock
4. **No duplicate rendering** — Check UI for same data rendered twice
5. **R-iteration ≤ 4** — If >4 rounds, pause for user direction
6. **Mock test ≠ functional** — Mock passing doesn't mean feature works

## HARD CONSTRAINTS

- **Delete Permission: DISABLED** — Never delete files unless explicitly asked
- **Framework ≠ Project** — Project problems solved at project layer, never patch framework
- **No secrets in code** — Never expose or log secrets/keys
- **NEVER commit unless user asks** — Explicit confirmation required

## Entry Point for AI

When loaded by the AI runtime, read this file to understand v2 structure, then load:

1. `prompts/triage.md` — if your role is triage
2. `workflows/shared.md` — always (layer 1/2 gates)
3. `workflows/roles/<role>-standards.md` — role-specific standards

For beads commands reference: see `docs/BEADS_INTEGRATION.md`