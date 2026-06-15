# Project Configuration

## project.yaml structure

```yaml
name: my-project
version: v3
active_flow: dev-flow

paths:
  docs_internal: _docs/my-project/    # Optional, fallback to .team/docs/
  docs_external: docs/

toolchain:
  backend:
    language: go
    pipeline: go test ./... | go build -o bin/app
  frontend:
    language: typescript
    pipeline: bun run test | bun run build
```

## Internal Document Management

AI reads/writes project documents to a managed directory:

| Path | When `docs_internal` is set | When not set (fallback) |
|------|---------------------------|------------------------|
| Lessons | `{docs_internal}/lessons/` | `.team/docs/lessons/` |
| Sessions | `{docs_internal}/sessions/` | `.team/docs/sessions/` |
| Conventions | `{docs_internal}/conventions/` | `.team/docs/conventions/` |

Resolution: `flow config paths` shows the resolved `DOCS_INTERNAL` path.

## Session Persistence

Every conversation must produce a session log for traceability:

```
{internal_docs}/sessions/{date}-{flow}-{node}[-{task}].md
```

Session log contains: Context (flow, node, task), Actions, Decisions, Lessons, Files Changed.

AI writes this at session close, before `git push`.
