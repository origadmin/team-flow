# Migration Guide: _team v1 → v2

> **Version**: v2.0 | **Date**: 2026-05-02
> **Scope**: Triage workflow, Task pool management, Multi-agent coordination

## Executive Summary

_team v2 replaces task-pool.md as the source of truth with beads (bd CLI), eliminating:
- Git merge conflicts from concurrent edits
- Manual sync overhead between agents
- Inconsistent task state tracking

## Key Changes at a Glance

| Aspect | v1 | v2 |
|--------|----|----|
| **Task source** | `task-pool.md` file | beads `.beads/` database |
| **ID format** | F/B/C/A-NNN (manual) | cms-xxx (auto-generated) |
| **Status sync** | Edit file + git commit | `bd update` (Dolt handles sync) |
| **Multi-agent** | Git merge conflicts | Dolt git-native collaboration |
| **Triage writes** | Parse + modify markdown | `bd create/update/close` |
| **Context injection** | Read task-pool.md | `bd list --json` or export |

## Migration Steps

### Phase 1: Parallel Run (Recommended)

Run v1 and v2 in parallel for 1-2 weeks before switching fully.

#### 1.1 Initialize beads in project

```bash
cd projects/orig-cms
bd init --prefix cms
```

#### 1.2 Migrate existing tasks

Use the migration script:

```powershell
# From framework root
.\_team\v2\scripts\migrate-tasks.ps1 -ProjectPath "projects/orig-cms" -DryRun

# Review output, then run for real
.\_team\v2\scripts\migrate-tasks.ps1 -ProjectPath "projects/orig-cms"
```

Migration script will:
- Read `.team/task-pool.md`
- Create beads issue for each row
- Store _team ID in `external_ref` field
- Add phase/subsystem labels
- Generate mapping file `.beads/task-pool-mapping.json`

#### 1.3 Validate migration

```bash
# Check counts match
bd stats
# Should show same count as task-pool.md rows

# Spot check specific issues
bd list --json | jq '.[] | select(.externalRef == "F014")'
```

#### 1.4 Update CLAUDE.md

Point to v2 SKILL.md:

```markdown
<!-- In projects/orig-cms/CLAUDE.md -->
Load _team rules: {TEAM_PATH}/v2/SKILL.md
```

### Phase 2: Full Cutover

#### 2.1 Stop writing to task-pool.md

Update all agent prompts to use `bd` commands instead of file edits.

#### 2.2 Archive v1

```bash
# Keep v1 for reference (read-only)
mv _team _team_v1_archive
ln -s _team_v1_archive _team  # Symlink for compatibility
```

#### 2.3 Update all SKILL.md references

```bash
# In project .team/SKILL.md
Load framework: {TEAM_PATH}/v2/SKILL.md
```

#### 2.4 Establish export routine

```powershell
# Schedule daily export for human review
# In Windows Task Scheduler or cron
bd list --status open --format table > .team/task-pool-export.md
```

## Field Mapping Reference

| task-pool.md Column | beads Field | Example |
|---------------------|-------------|---------|
| Task ID | `external_ref` | F014 → `--external-ref F014` |
| Task | `title` | "Implement X" → `bd create "Implement X"` |
| Type | `issue_type` | Feature → `-t feature` |
| Dependencies | `bd dep add` | A001,B062 → `bd dep add $ID $A001_ID $B062_ID` |
| Milestone | `--milestone` or parent epic | M1 → `--milestone M1` |
| Owner | `assignee` | alice → `--assignee alice` |
| Priority | `priority` | P0 → `-p 0` |
| Status | `status` | Doing → `--status in_progress` |
| Phase | `phase:*` label | todo → `--add-label phase:ready` |
| Sub-column | `subsystem:*` label | Backend → `--add-label subsystem:backend` |
| Doc | `doc_path` metadata | → `--set-metadata doc_path=_docs/...` |

## Workflow Changes

### Before (v1): Triage Updates task-pool.md

```
User input → Triage prompt → Parse task-pool.md → Modify markdown → Write file → Git commit
```

Pain points:
- Complex parsing logic
- Race conditions
- Git merge conflicts
- Manual sync between agents

### After (v2): Triage Uses bd CLI

```
User input → Triage prompt → bd create/update → Dolt auto-sync
```

Benefits:
- Single source of truth
- Native concurrency (Dolt)
- Audit trail
- Simple commands

## Common Migration Issues

### Issue: "bd: command not found"

Solution: Install beads CLI

```bash
npm install -g beads
# or
bun install -g beads
```

### Issue: ".beads/ not found"

Solution: Initialize beads in project

```bash
cd projects/orig-cms
bd init --prefix cms
```

### Issue: "Dolt repository not initialized"

Solution:

```bash
cd .beads
dolt init
dolt remote add origin <remote-url>
```

### Issue: "Task ID not found in mapping"

Solution: Run migration script with `--force` to recreate mapping

```powershell
.\_team\v2\scripts\migrate-tasks.ps1 -ProjectPath "projects/orig-cms" -Force
```

### Issue: "Phase labels not showing"

Solution: Ensure labels are added after creation

```bash
bd update cms-xxx --add-label phase:implement
```

## Rollback Plan

If v2 causes issues, rollback to v1:

1. Stop all v2 agents
2. Restore task-pool.md from git history
3. Update CLAUDE.md to point to `_team_v1_archive/SKILL.md`
4. Investigate issue in v2, fix, re-deploy

## Checklist

- [ ] beads installed (`bd --version`)
- [ ] beads initialized in project (`bd init`)
- [ ] Tasks migrated (`bd stats` shows correct count)
- [ ] Mapping file generated (`.beads/task-pool-mapping.json`)
- [ ] CLAUDE.md updated to v2
- [ ] Export routine established
- [ ] Team trained on `bd` commands
- [ ] v1 archived (not deleted)

## Training Resources

- beads CLI reference: `bd --help`, `bd <command> --help`
- _team v2 docs: `{TEAM_PATH}/v2/docs/`
- Integration guide: `BEADS_INTEGRATION.md`
