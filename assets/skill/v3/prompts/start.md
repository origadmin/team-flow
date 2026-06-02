# Start Node Instructions

## What This Node Does

You are at the beginning of the development flow. Your job is to:

1. Understand what the user wants
2. Recover context from previous conversations (if any)
3. Clarify ambiguities through discussion
4. Determine if a beads task should be created
5. Dispatch to Task Triage (tri3) when ready

## Event Logging Rules

The event log (`.team/state/events.mdl`) tracks sessions. Use these CLI commands:

### On Entry: Record user input
```bash
flow session start --input "<user's original message>" [--topic "<topic>"]
```
Call this EXACTLY ONCE at the start of this Start node run.

### On Each Analysis Round: Record your analysis
```bash
flow session analysis [--round <n>] --status <status> [--task-type <type>] [--task <id>] [--topic "<topic>"]
```
Call this ONCE PER ANALYSIS ROUND, each time you finish analyzing.

Status values:
- `clarifying`   → Need more info from the user. Ask a question, then wait for response.
                   This is the NORMAL state during multi-round discussion.
- `task_created` → Decision made, task has been created (via `flow task create`).
- `no_task`      → No development task needed (chat, question, etc).
- `redirect`     → User wants to resume an existing task.

The round number auto-detects if `--round` is omitted.

### Recover Previous Context
```bash
flow session last --for-ai
```
Call this FIRST to load previous session context and active tasks.

## Decision Flow

```
┌─────────────────────────────────────────────────┐
│ 1. Recover context: flow session last --for-ai  │
│ 2. Record start:   flow session start --input   │
└───────────────┬─────────────────────────────────┘
                ▼
┌───────────────────────────────────────┐
│ Analyze user input + context          │
│ Determine what the user wants         │
└───────────────┬───────────────────────┘
                ▼
         ┌──────┴──────┐
         │ Enough info? │
         └──────┬───────┘
          No    │    Yes
           ▼    │     ▼
    ┌──────────┐│┌──────────────────────┐
    │ Ask      │││ Determine action:    │
    │ clarifying│││ - Need task? → create│
    │ question │││   flow task create    │
    │          │││ - No task? → no_task  │
    │ Record:  │││ - Resume? → redirect  │
    │ analysis │││                       │
    │ status=  │││ Record: analysis      │
    │clrifying │││ status=task_created/  │
    │          │││        no_task/redir  │
    │ WAIT for │││                       │
    │ user     │││ → Advance to tri3     │
    │ response ││└──────────────────────┘
    │ then     ││
    │ LOOP     ││
    └──────────┘│
                │
                └── When status = task_created / no_task / redirect
                    Your job is done. The flow engine advances to tri3.
```

## Important Rules

1. **Always recover context first** — call `flow session last --for-ai` before anything else.
2. **session.start is written ONCE** — only on entry. Don't call it again in the same session.
3. **session.analysis is written MULTIPLE TIMES** — once per analysis round. Status evolves.
4. **task creation = final round** — after creating a task, record `analysis --status task_created --task <ID>`.
5. **Don't read the whole events.mdl** — use `flow session last` which filters by event type.
6. **Clarify before creating** — never create a task until you fully understand what needs to be done.

## Task Type Determination

After analysis, assign one of:
- `feature`    — New functionality
- `bug`        — Defect fix
- `hotfix`     — Urgent production fix
- `analysis`   — Investigation / research
- `change`     — Refactoring / maintenance

Use `flow task create --title "<summary>" --type <type> --description "<details>"` to create the task.
Then record: `flow session analysis --status task_created --task-type <type> --task <ID>`
