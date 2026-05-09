# beads Integration Guide — team-flow v2

> **Version**: v2.0 | **Date**: 2026-05-02

## Overview

beads (flow task CLI) is the single source of truth for task management in team-flow v2. This document maps beads concepts to team-flow v1 equivalents and defines the new workflow.

## Concept Mapping

| team-flow v1 Concept | beads Equivalent | Notes |
|------------------|------------------|-------|
| task-pool.md | beads `.beads/` database | Source of truth |
| F/B/C/A-NNN ID | `external_ref` field | Stored in beads issue metadata |
| Status column | `status` field | open/in_progress/blocked/closed |
| Phase (todo/doing/done) | `phase:*` label | Use labels for phases |
| Sub-column | N/A → `subsystem:*` label | Use subsystem labels instead |
| Dependency | `flow task dep add` | Native dependency tracking |
| Assignee | `assignee` field | User name or email |
| Priority | `priority` field | 0-4 (P0 highest) |
| Milestone | `--milestone` or parent epic | Link to milestone/epic |

## Phase Labels System

Use labels to track workflow phases. Standard phase labels:

```
phase:ready     # Ready for triage/dispatch
phase:analyze   # Under analysis (Tech Lead)
phase:design    # Design/spec phase
phase:implement # Development in progress
phase:verify    # Testing/QA phase
phase:review    # Waiting for user confirmation
```

Progress through phases by adding/removing labels:

```bash
# Move to implementation
flow task update cms-xxx --add-label phase:implement --remove-label phase:analyze

# Move to review
flow task update cms-xxx --add-label phase:review --remove-label phase:verify
```

## Subsystem Labels (Replaces Sub-columns)

For categorizing work within a phase:

```
subsystem:backend
subsystem:frontend
subsystem:api
subsystem:database
subsystem:architecture
subsystem:testing
```

## Type Mapping

| task Type | beads Type | Priority Default |
|------------|------------|-------------------|
| Bug | bug | P0/P1 |
| Feature | feature | P1/P2 |
| Task | task | P2 |
| Analysis | task + `type:analysis` label | P1 |
| Change | task | P2 |

## ID Cross-Reference

task legacy IDs (F001, B061, etc.) are stored in beads' `external_ref` field:

```bash
# Create with task ID reference
flow task create "Feature: X" -t feature --external-ref "F014" -p 1 --json

# Find by task ID
flow task list --json | jq '.[] | select(.externalRef == "F014")'

# Update mapping
flow task update cms-xxx --external-ref "F014"
```

## Common Workflows

### Triage: New Issue Intake

```bash
# 1. Create issue
ISSUE=$(flow task create "Description" -t bug -p 0 --external-ref "B099" --json)
ID=$(echo $ISSUE | jq -r '.id')

# 2. Add phase label
flow task update $ID --add-label phase:ready

# 3. Add subsystem
flow task update $ID --add-label subsystem:backend

# 4. Link dependencies (if any)
flow task dep add $ID <dependency-id> --type discovered-from
```

### Dev: Claim and Execute

```bash
# 1. Claim issue
flow task update cms-xxx --claim

# 2. Mark implementation start
flow task update cms-xxx --add-label phase:implement --remove-label phase:ready

# 3. Record progress
flow task update cms-xxx --notes "COMPLETED: X IN PROGRESS: Y"

# 4. Mark for verification
flow task update cms-xxx --add-label phase:verify --remove-label phase:implement

# 5. Close when done
flow task close cms-xxx --reason "Fixed and verified"
```

### Tech Lead: Review Analysis

```bash
# 1. List issues needing analysis
flow task list --status open --priority 0,1 -l phase:ready --json

# 2. Claim for analysis
flow task update <beads-id> --claim
flow task update <beads-id> --add-label phase:analyze --remove-label phase:ready

# 3. Document findings
flow task update <beads-id> --design "Root cause: ... Solution: ..."

# 4. Hand off to dev
flow task update <beads-id> --add-label phase:ready --remove-label phase:analyze --assignee "dev-name"
```

## Export to Human-Readable Format

For AI context injection or human review:

```bash
# JSON export
flow task list --json > .team/task-pool-export.json

# Table export (human-readable)
flow task list --format table > .team/task-pool.md

# Filter by status
flow task list --status open --format table > .team/task-pool-open.md

# Filter by priority
flow task list --priority 0 --format table > .team/task-pool-p0.md
```

### Export Script Template

```powershell
# export-task-pool.ps1
$Output = @()
$Issues = flow task list --status open --json | ConvertFrom-Json

foreach ($Issue in $Issues) {
    $Phase = ($Issue.labels | Where-Object { $_ -match '^phase:' }) -replace 'phase:', ''
    $Subsystem = ($Issue.labels | Where-Object { $_ -match '^subsystem:' }) -replace 'subsystem:', ''

    $Output += [PSCustomObject]@{
        ID = $Issue.externalRef
        BeadsID = $Issue.id
        Title = $Issue.title
        Type = $Issue.issue_type
        Priority = "P$($Issue.priority)"
        Phase = $Phase
        Subsystem = $Subsystem
        Status = $Issue.status
    }
}

$Output | Format-Table -AutoSize | Out-File ".team/task-pool.md" -Encoding UTF8
```

## Metadata Schema

Standard metadata fields stored via `--set-metadata`:

```json
{
  "team_id": "F014",           // Original task ID
  "subsystem": "backend",      // Category
  "original_deps": "A001,B062", // Dependencies for export
  "doc_path": "{DOCS_INTERNAL}/requirements/F014/", // Document location
  "milestone": "M1"            // Milestone reference
}
```

```bash
# Set metadata
flow task update cms-xxx --set-metadata team_id=F014
flow task update cms-xxx --set-metadata subsystem=backend
flow task update cms-xxx --set-metadata doc_path="{DOCS_INTERNAL}/requirements/F014/"
```

## Dolt Integration (Git-native)

beads uses Dolt for version control:

```bash
# Check Dolt status
flow task dolt status

# Commit changes (batch mode)
flow task dolt commit -m "Batch update from triage session"

# Push to remote
flow task dolt push
```

For multi-agent environments, use `--dolt-auto-commit on` to auto-commit after each operation.

## Query Examples

```bash
# All P0 bugs in backend
flow task list -t bug --priority 0 -l subsystem:backend

# All issues in implementation phase
flow task list -l phase:implement

# Issues blocked by specific issue
flow task children <beads-id> --type blocks

# Ready to work (no blockers, open status)
flow task ready --json
```

## Configuration

```bash
# Set default actor
flow task config set actor "triage-agent"

# Enable auto-commit
flow task config set dolt.auto-commit on

# Custom status workflow
flow task config set status.custom "pending-review:wip"
```
