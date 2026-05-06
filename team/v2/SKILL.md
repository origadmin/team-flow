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
├── prompts/              # Role execution rules (v2, beads-native)
│   ├── triage.md
│   ├── dev.md
│   ├── dev-backend.md    # Backend Dev 专属规则（从 v1 迁移）
│   ├── dev-frontend.md   # Frontend Dev 专属规则（从 v1 迁移）
│   ├── bugfix.md         # Bugfix 专属规则（从 v1 迁移）
│   ├── tech-lead.md
│   ├── qa-engineer.md
│   └── pm.md
├── workflows/            # Shared workflows + role standards
│   ├── shared.md
│   └── roles/
│       ├── triage-standards.md
│       ├── development-standards.md
│       ├── bugfix-standards.md
│       ├── devtestops.md
│       ├── test-levels.md
│       ├── specialized-tests.md          # 后端 8 类专项
│       └── frontend-specialized-tests.md  # 前端 10 类专项
├── templates/            # Document templates（从 v1 迁移 + beads 适配）
│   ├── README.md
│   ├── feature-test-template.md           # 后端 Feature 测试覆盖
│   ├── bug-test-template.md               # 后端 Bug 复现测试
│   ├── frontend-feature-test-template.md  # 前端 Feature 测试覆盖
│   ├── frontend-bug-test-template.md      # 前端 Bug 复现测试
│   ├── scope-template.md                  # 变更报告
│   ├── test-report-template.md            # 测试报告
│   ├── closed-loop-verification-template.md
│   ├── gherkin-feature-template.md
│   ├── architecture-template.md
│   ├── prd-template.md
│   ├── ui-design-template.md
│   ├── api-issue-template.md
│   └── user-story-template.md
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

## Entry Point for AI

When loaded by the AI runtime, read this file to understand v2 structure, then load:

1. `prompts/triage.md` — if your role is triage
2. `workflows/shared.md` — always (layer 1/2 gates)
3. `workflows/roles/<role>-standards.md` — role-specific standards

For beads commands reference: see `docs/BEADS_INTEGRATION.md`