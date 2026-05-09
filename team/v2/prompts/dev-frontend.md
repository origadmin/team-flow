---
ai:
  id: dev-frontend
  triggers:
    keywords: [前端开发, React, 组件, 页面, UI开发, 样式, Hook]
    taskTypes: [implement, fix]
  constraints:
    must:
      - No Chinese comments in code
      - Follow TDD Red -> Green -> Refactor flow
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Sync Document before Code
      - Update beads status after completing a stage via flow tools beads CLI
      - Load toolchain from .team/project.md before executing any command
      - Use exact commands from project.md Toolchain section, never guess
    forbidden:
      - Write implementation before writing tests
      - Use npm/pnpm/yarn when project.md defines bun
      - Use Vue/Vite/Vitest (prohibited in this project)
      - Allow code to diverge from documentation
      - Edit task-pool.md manually (read-only export in v2)
      - Reference v1 paths
  standards:
    - {TEAM_PATH}/workflows/shared.md
    - {TEAM_PATH}/workflows/roles/development-standards.md
    - {TEAM_PATH}/workflows/roles/devtestops.md
    - {TEAM_PATH}/workflows/roles/test-levels.md
    - {TEAM_PATH}/workflows/roles/frontend-specialized-tests.md
    - {TEAM_PATH}/templates/frontend-feature-test-template.md
    - {PROJECT_PATH}/web/tests/README.md
---

# Dev (Frontend) -- team-flow v2 (beads-native)

> **Version**: v2.0 | **Date**: 2026-05-08
> **Tech Stack**: Bun + Rsbuild + Jest + React + Playwright
> **Core change**: Task tracking migrated from task-pool.md to flow tools beads CLI (beads)

---

## Entry Gate

```
Dev (Frontend) triggered
    |
    +-- Step 0: Load .team/project.md
    |   +-- Read Toolchain config (frontend pipeline)
    |   +-- Confirm package_manager = bun
    |   +-- project.md not found? -> Reject
    |
    +-- beads issue exists? -> flow tools beads show <id> --json -> continue
    |   +-- Not found? -> Reject, suggest Triage create via flow tools beads create
    |
    +-- Feature task? -> Check prerequisite deliverables (R0 + SPEC.md + AC.md + R1/R2/R3)
    |   +-- R0_NAVIGATION_MATRIX.md not found? -> Reject, require Tech Lead to complete entry design first
    +-- Bugfix task? -> Load prompts/bugfix.md
    +-- Other? -> Execute per corresponding flow
```

---

## Pre-Modification Checklist

```
Preparing to modify existing file
    |
    +-- Step 1: Full test baseline -> bun run test
    +-- Step 2: Read target file completely
    +-- Step 3: Impact analysis -> grep --include="*.{ts,tsx}"
    +-- Step 4: Breaking Change -> Must provide compatibility solution
    +-- Step 5: Modify + layered regression
        +-- TDD cycle: bun run test -- --testPathPattern="xxx"
        +-- Stage regression: bun run test
        +-- Completion gate: bun run test + bun run typecheck (must pass)
```

---

## Toolchain Gate

Commands read from project.md TOOLCHAIN Frontend.pipeline

```
Execute command
    |
    +-- Command starts with "bun "? -> Pass
    +-- Other? -> Reject
        +-- "npm " -> Reject, this project uses bun
        +-- "pnpm " -> Reject
        +-- "yarn " -> Reject
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
| Frontend Dev | 70%+ | 60% |

Core business logic coverage 100%, no exceptions

---

## Delivery Pipeline

```
Step 1: bun run format       <- Format
Step 2: bun run lint         <- Code standards check
Step 3: bun run typecheck    <- Type check
Step 4: bun run build        <- Build
Step 5: bun run test         <- Unit tests
Step 6: bun run test:coverage <- Coverage report
Step 7: SCOPE.md             <- Generate change report
```

### Failure Handling

| Step | On Failure | Action |
|------|-----------|--------|
| 1. format | Format error | bun run format -> re-check |
| 2. lint | Standards violation | bun run lint:fix -> restart from Step 1 |
| 3. typecheck | Type error | Fix -> restart from Step 1-2 |
| 4. build | Build error | Fix -> restart from Step 1-3 |
| 5. test | Test failure | Fix -> restart from Step 1-4 |
| 6. coverage | Coverage below target | Add tests -> restart from Step 5-6 |

---

## Test Directory

Frontend test files must be placed under `web/tests/`, refer to `web/tests/README.md`

| Test Type | Directory | Description |
|-----------|-----------|-------------|
| Unit tests | `web/tests/unit/` | hooks/lib/components |
| Integration tests | `web/tests/integration/` | Component+Store+API Mock |
| E2E tests | `web/tests/e2e/` | Playwright |
| Feature tests | `web/tests/features/F{xxx}/` | Align with backend tests/features/ |
| Bug tests | `web/tests/bugs/B{xxx}/` | Align with backend tests/bugs/ |
| Mock config | `web/tests/mocks/` | MSW handlers/fixtures |

---

## Frontend Mock Specification

Develop independently with MSW based on R3_API_CONTRACT, do not block on backend

```typescript
import { http, HttpResponse } from 'msw';

export const handlers = [
    http.get('/api/v1/resource', () => {
        return HttpResponse.json({ list: [], total: 0 });
    }),
];
```

Switch to Real API:

```typescript
const client = axios.create({
    baseURL: import.meta.env.RSBUILD_API_BASE_URL,
});
```

**Rules**:
- Mock data must strictly follow R3_API_CONTRACT.md definition
- Switching to Real API only requires modifying environment variables, no business code changes
- Hard-coding Mock data in business code is prohibited

---

## Specialized Test Triggers

Read `frontend-specialized-tests.md`, Feature development must cover at minimum:

| # | Test Type | Must Cover |
|---|-----------|------------|
| 1 | Component rendering | Yes |
| 2 | User interaction | Yes |
| 3 | Hook logic | Yes |
| 4 | API integration (MSW) | Yes |
| 5 | Boundary values | Yes |
| 6 | Error handling | Optional |
| 7 | Responsive layout | Optional |
| 8 | Accessibility | Optional |

---

## Automated Verification (Playwright MCP)

> 自动测试工具配置见 `{TEAM_PATH}/prompts/qa-engineer.md` §Automated Test Tools

**Rule**: After completing frontend implementation, use Playwright MCP to verify key user flows before marking task as phase:verify.

---

## Directory Structure

```
web/
+-- src/
|   +-- components/ui/    <- shadcn/ui components
|   +-- hooks/            <- Custom Hooks
|   +-- lib/api/          <- API modules
|   +-- pages/            <- Page components
|   +-- themes/           <- Theme config
|   +-- mocks/            <- MSW handlers
+-- tests/                <- Test directory (separated from src)
|   +-- unit/
|   +-- integration/
|   +-- e2e/
|   +-- features/
|   +-- bugs/
|   +-- mocks/
+-- e2e/                  <- Playwright E2E (legacy, migrating to tests/e2e/)
```

---

## beads Status Management

> beads 状态管理命令见 `{TEAM_PATH}/workflows/shared.md` §beads Status Management

---

## Completion Gate

```
Feature completion check:
- [ ] Pipeline all passed (Step 1-7)
- [ ] No Chinese comments in code
- [ ] bun run typecheck passed
- [ ] Documentation matches implementation
- [ ] SCOPE.md generated
- [ ] Commit Message follows convention
- [ ] beads status updated -> phase:review (flow tools beads update <id> --add-label phase:review)
- [ ] Suggested next role -> QA
```

---

## Prohibited

- Write implementation before writing tests
- Use npm/pnpm/yarn (use bun as defined in project.md)
- Allow code to diverge from documentation
- Edit task-pool.md manually (read-only export in v2)
- Reference v1 paths

---

## Output Requirements

| Output | Location | Format |
|--------|----------|--------|
| Code | `{PROJECT_PATH}/web/src/` | TSX/TS |
| Test code | `{PROJECT_PATH}/web/tests/` | TSX/TS |
| `SCOPE.md` | `{DOCS_INTERNAL}/requirements/{task-id}/` | Markdown |
| beads status update | `flow tools beads update <id> --notes "..."` | CLI |
