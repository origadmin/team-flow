# team-flow Anti-Patterns & Standards

## ❌ DON'T (Anti-Patterns)
- Skip beads pull/push in multi-agent setups.
- Leave tasks in `phase:ready` unassigned.
- Close tasks without recording outcome.
- Create duplicate tasks (check with `flow task list`).
- Dump deliverable content into beads notes (use independent files).
- Use Chinese comments in code.
- Skip RCA / SCOPE / Acceptance Criteria.
- Write to `framework/{TEAM_PATH}/` (read-only).
- Write to `{PROJECT}/_docs/` (Use `framework/_docs/{project}/`).
- Edit `task-pool-export.md` manually (use CLI).

## beads Notes Protocol
Only record progress summary and handover info:
- ✅ `flow task update <id> --notes "COMPLETED: RCA.md written, fix applied"`
- ❌ `flow task update <id> --notes "Root cause: handler.go wraps response..."`

## Deliverable Locations
| Content | Correct Path | Forbidden |
|---------|--------------|-----------|
| RCA / Test Case | `{DOCS_INTERNAL}/reports/bugs/B{NNN}/R{n}/` | beads notes |
| Scope Report | `{DOCS_INTERNAL}/reports/changes/C{NNN}/` | beads notes |
| SPEC / AC | `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/` | beads notes |
| Design (R1-R5) | `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/` | beads notes |
