# team-flow v2 — Beads-Native Task Management

> **Status**: Ready for testing | **Version**: v2.0 | **Date**: 2026-05-02

## What's New in v2

### Core Change: beads as Source of Truth

**v1**: task-pool.md file (git merge conflicts, manual sync)
**v2**: beads database (Dolt git-native, multi-agent safe)

### Key Benefits

1. **No more merge conflicts** — Dolt handles concurrent edits automatically
2. **Single source of truth** — All task state in `.beads/` database
3. **Audit trail** — Full history of all changes
4. **Rich querying** — Filter, search, aggregate with SQL or CLI
5. **Multi-agent ready** — Designed for concurrent AI sessions

### New Concepts

| v1 Concept | v2 Equivalent |
|------------|---------------|
| Task ID (F001) | `external_ref` field + auto-generated beads ID (`<beads-id>`) |
| Status column | `status` field + `phase:*` labels |
| Sub-column | `subsystem:*` labels |
| task-pool.md | Export from beads (read-only) |
| Git merge | Dolt git-native sync |

## Quick Start

### 1. Initialize beads in your project

```bash
cd projects/your-project
flow task init --prefix <proj>
```

### 2. Load v2 rules

Update your project's CLAUDE.md:

```markdown
Load team-flow rules: {TEAM_PATH}/v2/SKILL.md
```

### 3. Migrate existing tasks (if any)

```powershell
{TEAM_PATH}/v2/scripts/migrate-tasks.ps1 -ProjectPath projects/your-project
```

### 4. Start using beads

```bash
# Create task
flow task create "Fix login bug" -t bug -p 0 --add-label phase:ready --json

# List tasks
flow task list --status open --priority 0,1

# Update task
flow task update <id> --claim --add-label phase:implement

# Close task
flow task close <id> --reason "Fixed in commit abc123"
```

## Directory Structure

```
{TEAM_PATH}/v2/
├── README.md              ← You are here
├── SKILL.md               ← Entry point for AI
├── BOUNDARY.md             ← Layer boundaries
├── prompts/                ← Role execution rules
│   ├── triage.md          ← Triage (beads-native)
│   ├── dev.md             ← Developer
│   ├── tech-lead.md       ← Tech Lead
│   ├── qa-engineer.md    ← QA
│   └── pm.md              ← Product Manager
├── workflows/              ← Shared workflows
│   ├── shared.md          ← Universal rules
│   └── roles/              ← Role-specific standards
│       ├── triage-standards.md
│       ├── development-standards.md
│       └── ...
├── templates/               ← Document templates
│   └── task-pool-template.md
├── scripts/                ← Automation
│   └── export-task-pool.ps1
└── docs/                   ← Reference
    ├── MIGRATION.md        ← v1 → v2 guide
    └── BEADS_INTEGRATION.md ← beads field mapping
```

## Essential Documents

| Document | Purpose | Read When |
|----------|---------|-----------|
| `SKILL.md` | Entry point, structure, quick start | Every session |
| `BOUNDARY.md` | Layer rules, what goes where | Setting up project |
| `docs/MIGRATION.md` | Migrate from v1 | Moving project to v2 |
| `docs/BEADS_INTEGRATION.md` | beads CLI usage, field mapping | All roles |
| `prompts/triage.md` | Triage workflow | Acting as triage |
| `workflows/shared.md` | Universal rules | Every session |

## Workflow Overview

```
User Input
    ↓
Triage → flow task create (phase:ready)
    ↓
Tech Lead → flow task update (phase:analyze → phase:design)
    ↓
Dev → flow task update --claim (phase:implement)
    ↓
QA → flow task update (phase:verify)
    ↓
User Confirmation → flow task close
```

## Phase Labels

Track progress through phases using labels:

```
phase:ready      → Ready for dispatch
phase:analyze    → Under investigation
phase:design     → Writing spec
phase:implement  → Development
phase:verify     → Testing
phase:review     → Waiting for confirmation
```

## Integration with Existing Tools

- **team-flow v1**: Archive as `_team_v1_archive/`, keep for reference # (legacy command)
- **CLAUDE.md**: Point to v2 SKILL.md
- **beads database**: `.beads/` in project root
- **Dolt**: Git-native version control, syncs with beads

## Getting Help

```bash
# beads CLI help
flow task --help
flow task <command> --help

# Check beads status
flow task stats
flow task ready

# Dolt help
flow task dolt --help
```

## Feedback

This is a new framework version. Report issues or suggestions via:
- Framework repo issues
- Team chat channel
- Direct message to framework maintainer
