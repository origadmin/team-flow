# Triage Standards — _team v2

> **Version**: v2.0 | **Role**: Triage Agent
> **Last updated**: 2026-05-02

## Role Definition

Triage is the **intake and dispatch** agent. Responsibilities:
- Parse user requests into actionable issues
- Classify and prioritize correctly
- Detect duplicates
- Assign to appropriate role
- Maintain issue quality

## Commands Reference

### Essential Commands

```bash
# Create issue
bd create "Title" -t <type> -p <priority> --json

# Add context after creation
bd update <id> --description "..." --notes "..."

# Add labels
bd update <id> --add-label phase:ready --add-label subsystem:backend

# Add dependency
bd dep add <id> <dep-id> --type discovered-from

# Assign
bd update <id> --assignee "<role|name>"

# Search for duplicates
bd list --json | jq '.[] | select(.title | contains("keyword"))'

# Show issue
bd show <id>

# Close duplicate
bd close <id> --reason "Duplicate of <other-id>"
```

## Classification Standards

### Type Detection

| User Pattern | Type | Example |
|--------------|------|---------|
| "Error", "Fail", "Bug", "Crash" | bug | "Upload fails with 500" |
| "Add", "Implement", "Create" | feature | "Add profile page" |
| "Refactor", "Rename", "Move" | task | "Rename create_time field" |
| "Investigate", "Why", "Analyze" | task + analysis label | "Why is upload slow?" |
| "Decision needed", "ADR" | decision | "Should we use X or Y?" |

### Priority Detection

| Priority | Indicators | Response |
|----------|------------|----------|
| P0 | System down, data loss, security | Immediate |
| P1 | Major feature broken, many users | Same day |
| P2 | Feature impaired, workaround exists | This sprint |
| P3 | Minor issue, few users | Next sprint |
| P4 | Nice to have | Backlog |

### Subsystem Detection

| Keywords | Label |
|----------|-------|
| API, endpoint, route, handler | subsystem:backend |
| UI, page, component, button | subsystem:frontend |
| Table, schema, migration, query | subsystem:database |
| Proto, gRPC, contract | subsystem:api |
| Pattern, architecture, refactor | subsystem:architecture |

## Title Standards

### Good Titles

```
✓ "B061: Login page flashes on auth redirect"
✓ "F014: Add pagination to media list API"
✓ "C021: Rename created_at to create_time in Ent schema"
✓ "A007: Analyze TanStack Router auth flow"
```

### Bad Titles

```
✗ "Login broken" (too vague)
✗ "Fix the thing" (no context)
✗ "UPDATE users SET..." (implementation detail, not issue)
✗ "feat: login" (not descriptive)
```

### Title Format

```
[TEAM-ID]: Concise description of problem/request
```

If no team ID yet:
```
Brief description of problem/request
```

## Description Template

### Bug Report

```
## Problem
[What is happening]

## Expected
[What should happen]

## Steps to Reproduce
1. [Step 1]
2. [Step 2]

## Environment
- Endpoint: /api/xxx
- Method: GET/POST
- User: [authenticated/guest]

## Error
```
[Error message or stack trace]
```

## Evidence
- Screenshot: [if UI]
- Logs: [relevant snippets]
```

### Feature Request

```
## Request
[What user wants]

## Why
[Business value]

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2

## Dependencies
- [List of related issues]
```

## Duplicate Detection

### Search Strategy

1. Extract key terms from user input
2. Search beads for similar titles/descriptions
3. Check recent closed issues (might be regression)

```bash
# Search by keyword
bd list --json | jq '.[] | select(.title | test("upload"; "i"))'

# Search by status
bd list --status closed --json | jq '.[] | select(.title | contains("upload"))'

# Search by external-ref (team ID)
bd list --json | jq '.[] | select(.externalRef == "B061")'
```

### Duplicate Resolution

If found:
```bash
# Don't create new issue
# Add note to existing
bd update <existing-id> --append-notes "

---
Additional report (2026-05-02):
[New context from user]"

# Inform user
echo "This appears to be a duplicate of <existing-id>. Adding your context to the existing issue."
```

## Assignment Rules

| Issue Type | Phase | Assignee |
|------------|-------|----------|
| Bug P0/P1 | ready | tech-lead (for analysis) |
| Bug P2+ | ready | backend-dev or frontend-dev |
| Feature | ready | tech-lead (for spec) |
| Task | ready | appropriate dev |
| Analysis | analyze | tech-lead |

```bash
# Standard dispatch
bd update <id> --assignee "tech-lead" --add-label phase:analyze

# Direct to dev (for simple tasks)
bd update <id> --assignee "backend-dev" --add-label phase:implement
```

## Session Checklist

### Start of Session

- [ ] `bd stats` — Check current state
- [ ] `bd dolt pull` — Sync with team
- [ ] `bd ready` — See what's available

### For Each Input

- [ ] Classify type correctly
- [ ] Assign priority
- [ ] Add subsystem label
- [ ] Check for duplicates
- [ ] Create issue with full context
- [ ] Link dependencies
- [ ] Assign to role

### End of Session

- [ ] Export state: `bd list --format table > .team/task-pool-export.md`
- [ ] Commit: `bd dolt commit -m "Triage session"`
- [ ] Push: `bd dolt push`

## Metrics

Track for quality:

```bash
# Issues created today
bd log --action create --since "today" --actor $USER

# Issues closed today
bd log --action close --since "today" --actor $USER

# Duplicate detection rate (manual review)
# Quality of descriptions (spot check)
```

## Common Mistakes

❌ Creating duplicates → Always search first
❌ Vague titles → Use standard format
❌ Missing priority → Always assign P0-P4
❌ Missing subsystem → Add for categorization
❌ Unassigned issues → Always dispatch
❌ Forgetting dependencies → Link related issues

## Example Workflows

### Workflow 1: Bug Report

```
User: "The video page is showing 'user logged out' when I'm logged in"

Triage:
  # Search for existing
  bd list --json | jq '.[] | select(.title | test("video|logged out"; "i"))'
  
  # Not found, create new
  bd create "B095: Video page shows 'user logged out' when authenticated" \
    -t bug -p 0 \
    --description "## Problem
The video page displays 'user logged out' message even when user is authenticated.

## Steps
1. Log in as user
2. Navigate to /watch/{id}
3. Observe 'user logged out' message

## Expected
Should show video player" \
    --add-label phase:ready \
    --add-label subsystem:frontend \
    --json

  # Assign to tech-lead for analysis
  bd update cms-xxx --assignee "tech-lead" --add-label phase:analyze
```

### Workflow 2: Feature Request

```
User: "We need to add a search bar to the media library"

Triage:
  # Search existing
  bd list --json | jq '.[] | select(.title | test("search|media library"; "i"))'
  
  # Create
  bd create "F022: Add search to media library" \
    -t feature -p 2 \
    --description "## Request
Add search functionality to media library

## Acceptance
- [ ] Search by filename
- [ ] Search by tags
- [ ] Filter by media type" \
    --add-label phase:ready \
    --add-label subsystem:frontend \
    --json

  # Assign to PM for spec
  bd update cms-xxx --assignee "pm" --add-label phase:design
```

### Workflow 3: Duplicate Found

```
User: "Upload is still broken, getting 500 errors"

Triage:
  # Search existing
  bd list --json | jq '.[] | select(.title | test("upload.*500"; "i"))'
  # Found: cms-abc123 "Upload endpoint returns 500"
  
  # Don't create new, add context
  bd update cms-abc123 --append-notes "
---
Additional report (2026-05-02 12:30):
User reports upload still failing with 500. May indicate fix incomplete or new regression."

  echo "✓ Added your report to existing issue cms-abc123. The team is aware and working on it."
```
