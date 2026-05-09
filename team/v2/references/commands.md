# Command Reference

`flow task` is the unified AI-facing command. It routes based on `.team/version`.

## Routing

| Stage | .team/version | `flow task create` routes to | `flow task list` reads from |
|-------|---------------|------------------------------|----------------------------|
| v1 | 1 | task-pool.md (docs) | task-pool.md |
| v2 | 2 | beads (.beads/) | beads (.beads/) |
| v3 | 3 | configurable (beads/git/other) | configurable |

## Command List

| Operation | Command |
|-----------|---------|
| Create task | `flow task create "Title" -t {bug\|feature\|task\|epic} -p {0-4} --parent {id}` |
| Claim task | `flow task update {id} --claim` |
| Add notes | `flow task update {id} --notes "..."` |
| Add label | `flow task update {id} --add-label phase:implement` |
| Close task | `flow task close {id} --reason "..."` |
| Show task | `flow task show {id}` |
| Find work | `flow task ready --json` |
| List tasks | `flow task list --status open` |
| Filter by date | `flow task list --created-after YYYY-MM-DD --json` |
| Filter by update | `flow task list --updated-after YYYY-MM-DD --sort updated` |
| Append record | `flow task append {id} --speaker {role} --content "..."` |
| Add dependency | `flow task dep add {id} depends-on {target-id}` |
| Push data | `flow task dolt push` |
| Export tasks | `flow export > .team/task-pool-export.md` |
| Resolve paths | `flow config paths --json` |

## Auto-timestamps

- `flow task create` auto-sets `created_at`
- `flow task update` auto-sets `updated_at`
- AI never needs to manually write timestamps

## Conversation Records

`flow task append {id} --speaker {role} --content "..."` auto-increments cr-index and sets timestamp.

## Task Availability

`flow task` is always available after `flow init`. If not found:

```bash
flow doctor
flow init --force
```

## v2 Task Lifecycle

```
flow task create "Title" -t bug|feature|task -p 0-4 --json
    ↓
flow task update <id> --claim   (status → in_progress)
    ↓
flow task update <id> --notes "COMPLETED: ... IN PROGRESS: ..."
    ↓
flow task close <id> --reason "Done" --json
```

## ID Mapping

- task ID (F001, B061, etc.) lives in the `external-ref` field of the beads issue
- Use `flow task list --json | ConvertFrom-Json | Where-Object { $_.externalRef -match 'F001' }` to find by task ID
