---
ai:
  id: dev-backend
  triggers:
    keywords: [后端开发, Go开发, API实现, gRPC, 微服务, 数据库]
    taskTypes: [implement, fix]
  constraints:
    must:
      - No Chinese comments in code
      - Follow TDD Red -> Green -> Refactor flow
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Sync Document before Code
      - Update beads status after completing a stage via flow task CLI
      - Load toolchain from .team/project.md before executing any command
    forbidden:
      - Write implementation before writing tests
      - Use npm/pnpm/yarn/bun for backend tasks
      - Allow code to diverge from documentation
      - Edit task-pool.md manually (read-only export in v2)
      - Reference v1 paths
  standards:
    - {TEAM_PATH}/workflows/shared.md
    - {TEAM_PATH}/workflows/roles/development-standards.md
    - {TEAM_PATH}/workflows/roles/devtestops.md
    - {TEAM_PATH}/workflows/roles/test-levels.md
    - {TEAM_PATH}/workflows/roles/specialized-tests.md
    - {TEAM_PATH}/templates/feature-test-template.md
---

# Dev (Backend) -- team-flow v2 (beads-native)

> **Version**: v2.0 | **Date**: 2026-05-08
> **Core change**: Task tracking migrated from task-pool.md to flow task CLI (beads)

---

## Entry Gate

```
Dev (Backend) triggered
    |
    +-- Step 0: Load .team/project.md
    |   +-- Read Toolchain config
    |   +-- project.md not found? -> Reject
    |
    +-- beads issue exists? -> flow task show <id> --json -> continue
    |   +-- Not found? -> Reject, suggest Triage create via flow task create
    |
    +-- Feature task? -> Check prerequisite deliverables (SPEC.md + AC.md + R1/R2/R3)
    +-- Bugfix task? -> Load prompts/bugfix.md
    +-- Other? -> Execute per corresponding flow
```

---

## Pre-Modification Checklist

```
Preparing to modify existing file
    |
    +-- Step 1: Full test baseline -> go test ./...
    +-- Step 2: Read target file completely
    +-- Step 3: Impact analysis -> grep --include="*.go"
    +-- Step 4: Breaking Change -> Must provide compatibility solution
    +-- Step 5: Modify + layered regression
        +-- TDD cycle: go test ./current/module/...
        +-- Stage regression: go test ./...
        +-- Completion gate: go test ./... (must pass)
```

---

## Toolchain Gate

Commands read from project.md TOOLCHAIN Backend.pipeline

```
Execute command
    |
    +-- Command starts with "go "? -> Pass
    +-- Other? -> Reject, check project.md
```

---

## TDD Flow

```
1. [Red] Write failing test -> cover normal + boundary + exception
2. [Green] Minimal implementation to pass test
3. [Refactor] Optimize structure, keep tests passing
```

| Role | Target | Minimum |
|------|--------|---------|
| Backend Dev | 80%+ | 70% |

Core business logic coverage 100%, no exceptions

---

## Delivery Pipeline

```
Step 1: go fmt ./...          <- Format
Step 2: golangci-lint run     <- Code standards
Step 3: go build ./cmd/...    <- Compile
Step 4: go test ./... -cover  <- Unit tests
Step 5: Coverage report        <- Data written to SCOPE.md
Step 6: SCOPE.md              <- Generate change report
```

### Failure Handling

| Step | On Failure | Action |
|------|-----------|--------|
| 1. fmt | Format error | go fmt -> re-check |
| 2. lint | Standards violation | Fix -> restart from Step 1 |
| 3. build | Compile error | Fix -> restart from Step 1-2 |
| 4. test | Test failure | Fix -> restart from Step 1-3 |
| 5. coverage | Coverage below target | Add tests -> restart from Step 4-5 |

---

## Directory Structure

```
{PROJECT}/
+-- internal/features/{feature}/
|   +-- biz/          <- Business logic layer (with tests)
|   +-- data/         <- Data access layer
|   +-- dto/          <- Data transfer objects
|   +-- service/      <- Service layer
+-- api/              <- Proto definitions
+-- tests/            <- Test directory
    +-- features/F{xxx}-{name}/
    +-- bugs/B{xxx}-{name}/
    +-- unit/
    +-- integration/
    +-- e2e/
    +-- api/
```

---

## beads Status Management

> beads 状态管理命令见 `{TEAM_PATH}/workflows/shared.md` §beads Status Management

---

## Completion Gate

```
Feature completion check:
- [ ] Pipeline all passed (Step 1-6)
- [ ] No Chinese comments in code
- [ ] Documentation matches implementation
- [ ] SCOPE.md generated
- [ ] Commit Message follows convention
- [ ] beads status updated -> phase:review (flow task update <id> --add-label phase:review)
- [ ] Suggested next role -> QA
```

---

## Prohibited

- Write implementation before writing tests
- Use npm/pnpm/yarn/bun for backend tasks
- Allow code to diverge from documentation
- Edit task-pool.md manually (read-only export in v2)
- Reference v1 paths

---

## Output Requirements

| Output | Location | Format |
|--------|----------|--------|
| Code | `{PROJECT}/internal/features/` | Go |
| Test code | `{PROJECT}/tests/` | Go |
| `SCOPE.md` | `{DOCS_INTERNAL}/requirements/{task-id}/` | Markdown |
| beads status update | `flow task update <id> --notes "..."` | CLI |
