# Triage: Post-Verification & Review (loaded on demand)

> **Loading rule**: Bug返回后验证、Review确认、分发子Agent、管理状态、Session时加载此文件
> **Source**: 来自triage.md拆分，v17.0

---

## Bug Post-Verification (Triage Executes)

After bugfix-expert returns, Triage must verify:

### Part 1: File Existence Check

| # | Verification Item | Expected Path | If Missing |
|---|-------------------|---------------|------------|
| 1 | RCA.md | {DOCS_INTERNAL}/reports/bugs/B{NNN}/R{N}/RCA.md | Mark as "RCA missing", require completion |
| 2 | TEST_CASE.md | {DOCS_INTERNAL}/reports/bugs/B{NNN}/R{N}/TEST_CASE.md | Mark as "TEST_CASE missing", require completion |
| 3 | Reproduction test code | tests/bugs/B{NNN}-*/ or web/tests/bugs/B{NNN}-*/ | Mark as "reproduction test missing", require completion |

### Part 2: Test Execution Verification (IMPORTANT - must run actual tests)

Triage MUST execute the following commands to verify the fix actually works:

**Backend Bug**:
```bash
# Step 1: Compile check
go build ./...

# Step 2: Run all tests
go test ./...

# Step 3: Run bug-specific regression test
go test ./tests/bugs/B{NNN}-.../...
```

**Frontend Bug**:
```bash
# Step 1: Type check
bun run typecheck

# Step 2: Lint check
bun run lint

# Step 3: Run all tests
bun run test

# Step 4: Run bug-specific regression test
bun run test -- --testPathPattern="B{NNN}"

# Step 5: UI verification (MANDATORY for frontend)
bun run dev
# Then open the bug page, verify content and interactions
```

| # | Verification Item | Check Method | If Failed |
|---|-------------------|--------------|-----------|
| 4 | Backend: go build passes | Execute `go build ./...` | Mark as "compilation failed", send back to bugfix |
| 5 | Backend: go test passes | Execute `go test ./...` | Mark as "tests failing", send back to bugfix |
| 6 | Frontend: typecheck passes | Execute `bun run typecheck` | Mark as "type errors", send back to bugfix |
| 7 | Frontend: lint passes | Execute `bun run lint` | Mark as "lint errors", send back to bugfix |
| 8 | Frontend: tests pass | Execute `bun run test` | Mark as "tests failing", send back to bugfix |
| 9 | Bug-specific test passes | Execute bug regression test | Mark as "fix not effective", send back to bugfix |
| 10 | Frontend: page renders correctly | Open Bug URL in browser | Mark as "page broken", send back to bugfix |
| 11 | Frontend: page content correct | Check text/data/components | Mark as "content wrong", send back to bugfix |
| 12 | Frontend: interaction works | Click/input/submit on Bug area | Mark as "interaction broken", send back to bugfix |
| 13 | Frontend: Bug symptom gone | Reproduce original Bug steps | Mark as "Bug still exists", send back to bugfix |
| 14 | Frontend: screenshot saved | Check {DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/screenshots/ | Mark as "no screenshot evidence" |
| 15 | Frontend: UI_VERIFICATION.md exists | Check {DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/UI_VERIFICATION.md | Mark as "no UI verification report", send back to bugfix |
| 16 | Frontend: screenshots follow naming rule | Check filenames match B{NNN}-R{N}-{NNN}-{action}-{state}.png | Mark as "screenshot naming violation" |
| 17 | Frontend: screenshot count meets minimum | Check screenshot count >= minimum for bug type | Mark as "insufficient screenshots" |

### Part 3: Data Flow Tracing Verification (API/permissions/state/interaction bugs)

| # | Verification Item | Check Method | If Missing |
|---|-------------------|--------------|------------|
| 10 | RCA.md contains "Data Flow Tracing" section | Read RCA.md, search for "Data Flow Tracing" | Mark as "missing data flow tracing" |
| 11 | Data flow tracing includes breakpoint analysis | Read RCA.md, search for "breakpoint" | Mark as "incomplete data flow tracing" |
| 12 | TEST_CASE.md includes real-scenario verification | Read TEST_CASE.md, search for "real scenario" | Mark as "missing real-scenario verification" |

### Part 4: R Iteration Quality Check (must execute for R2+)

| # | Verification Item | Check Method | If Missing |
|---|-------------------|--------------|------------|
| 13 | RCA.md contains previous round failure analysis | Read RCA.md, search for "R{N-1}" or "previous round" | Mark as "missing R iteration analysis" |
| 14 | R iteration >= R4, has user been asked to confirm? | Check beads notes | Mark as "R4+ not paused" |

### Verification Result

- All present + All tests pass -> Update beads: `flow task update <id> --add-label phase:review --remove-label phase:verify`, assignee=QA
- Any missing or test failure -> Update beads: `flow task update <id> --notes "MISSING: {specific items}, need completion"`, send back to bugfix-expert
- Report verification results to user (include test output)

**Forbidden**: Creating RCA.md / TEST_CASE.md. These are created by bugfix-expert. Triage only verifies, never creates.

---

## R Iteration Handling

### R Iteration Rules

```
Bug修复 -> QA验证
    -> 验证通过 -> 进入Review确认流程
    -> 验证未通过
        -> 询问用户: "Bug未修好，是否需要R{N+1}迭代？"
        -> [A] 是 -> 打回bugfix-expert，触发R{N+1}
            -> 创建新目录B{NNN}/R{N+1}/（禁止覆盖R{N}）
            -> flow task update <id> --add-label phase:implement --remove-label phase:verify
        -> [B] 否 -> 创建新Bug或关闭
```

### R Iteration Quality Gate

| Iteration | Extra Requirements |
|-----------|--------------------|
| R1 | Standard Bug修复流程 |
| R2 | RCA MUST contain R1 failure analysis |
| R3 | RCA MUST contain R1+R2 failure analysis + alternative solution evaluation |
| R4+ | MANDATORY pause, ask user if to continue |

**R4+ pause rule**: When reaching R4, Triage MUST notify user and wait for confirmation before continuing.

---

## Review Confirmation & Archiving

### Trigger Timing

After QA completes verification, MUST execute Review confirmation flow:
```
QA验证通过
    -> Triage执行Review确认
    -> [Bug类型] 检查LESSON-NEEDED
    -> [Feature/Change] 可选检查
    -> 输出确认报告 -> 用户确认
    -> [A] 确认 -> 关闭task
    -> [B] 有问题 -> 打回或创建新Bug
```

### Close Task Pre-Check

```bash
# Bug类型必须检查
flow task list --label lesson:needed --json | jq '.[] | select(.externalRef == "B001")'

# 如果有LESSON-NEEDED标签
flow task show <id> --json | jq '.labels'
```

```markdown
## Close Task Pre-Check

**Task**: B001 (Bug)
**LESSON-NEEDED**: Yes, 1 pending

| Source | Description | Status |
|--------|-------------|--------|
| Bugfix | Unhandled null pointer exception | Pending |

**Select option**:
[A] Process LESSON-NEEDED first
[B] Skip, process later
```

### Review Confirmation Flow

```bash
# Find all phase:review issues
flow task list --label phase:review --json
```

```markdown
## Review Pending

| ID | Type | Description | Deliverables | Source | LESSON |
|----|------|-------------|--------------|--------|--------|
| F001 | Feature | User login | SPEC.md, AC.md | Tech Lead | No |
| B001 | Bug | Login timeout | RCA.md, TEST_CASE.md | Bugfix | Yes |

**Actions**: [Confirm all] / [Confirm one by one] / [Reject problematic]
```

### QA Found New Bug Handling

When QA finds a new Bug during Review:

```
QA发现新Bug
    -> 询问用户: "发现X现象，这是新Bug还是B001的R2?"
    -> [A] B001的R2 -> 询问: "Bug未修好，是否需要R2迭代?"
        -> [A] 是 -> 打回B001，触发R2
        -> [B] 否 -> 创建新Bug
    -> [B] 新Bug -> 创建新Bug
```

```markdown
## QA Found New Bug

**Phenomenon**: {description}
**Possible sources**:
- B001未修复好 -> R2迭代
- 新Bug -> 创建B{N+1}

**Select**:
[A] B001的R2
[B] 新Bug
[C] Observe, don't create task
```

### Task Completion Verification

Final check before closing Task:

| # | Check Item | Verification Method |
|---|------------|---------------------|
| 1 | All deliverables generated | Check file paths |
| 2 | beads status updated | `flow task show <id>` |
| 3 | LESSON-NEEDED processed (Bug type only) | `flow task list --label lesson:needed` |
| 4 | User confirmed | Check confirmation record |
| 5 | No missing R iterations | Check beads notes |

---

## LESSON-NEEDED Background Scan

### Trigger Timing

**Does not block main flow**, only prompts at:
- Session start (brief prompt)
- User actively asks "any pending lessons?"
- Session end

```bash
# Scan LESSON-NEEDED labels
flow task list --label lesson:needed --json
```

### Scan Output

```markdown
## Pending LESSON-NEEDED

| ID | Source | Description | Labeled at |
|----|--------|-------------|------------|
| B001 | Bugfix | Unhandled null pointer exception | 2h ago |

[Process] / [Defer later] / [Ignore]
```

### Process Flow

```
Triage processes LESSON-NEEDED
    -> Generate lesson content
    -> Write to {DOCS_INTERNAL}/lessons/
    -> flow task update <id> --remove-label lesson:needed --add-label lesson:done
```

---

## Sub-Agent Prompt Templates (Three-Layer Handoff Model)

> **Loading rule**: Load when Triage dispatches to sub-agent
### Three-Layer Handoff Model

```
Layer 1: Triage MUST pass (inject into sub-agent prompt)
  -> Task ID + title + description
  -> Task type (Feature/Bug/Change)
  -> Priority
  -> Key context (user's original words, error messages, etc.)

Layer 2: Sub-agent MUST read on startup (load immediately)
  -> {TEAM_PATH}/workflows/shared.md (core rules)
  -> Corresponding role prompt file (e.g., {TEAM_PATH}/prompts/dev.md)
  -> flow config paths --json (path variables)

Layer 3: Sub-agent reads on demand (load when needed)
  -> {TEAM_PATH}/workflows/roles/xxx-standards.md
  -> {TEAM_PATH}/templates/xxx-template.md
  -> {DOCS_INTERNAL}/reports/... (specific documents)
  -> {TEAM_PATH}/docs/COMMANDS.md
```

### Handoff Key Rules

1. **Triage MUST NOT pass shared.md content in the prompt** - waste tokens, sub-agent reads Layer 2 themselves
2. **Triage MUST pass user's original words verbatim** - don't summarize, don't paraphrase, keep original wording
3. **Sub-agent MUST read Layer 2 files before starting work** - first thing after startup
4. **Sub-agent MUST NOT read triage.md or triage-clarify.md** - this is Triage context, not sub-agent's
5. **Sub-agent reads Layer 3 files only when the specific task requires it** - load on demand, don't pre-read

### Feature Task (developer-engineer)

```markdown
You are Dev, executing task {beads_id}: {external-ref}: {title}

## Task Context (from Triage - Layer 1)
- Type: feature
- Priority: {0-4}
- Description: {user's original words}
- Key constraints: {extracted from classification}

## Required Reading (Layer 2 - load now)
1. {TEAM_PATH}/workflows/shared.md
2. {TEAM_PATH}/prompts/dev.md
3. Run: flow config paths --json

## On-Demand Reading (Layer 3 - load when needed)
- Standards: {TEAM_PATH}/workflows/roles/dev-standards.md
- Templates: {TEAM_PATH}/templates/feature-template.md
- Commands: {TEAM_PATH}/docs/COMMANDS.md
- Frontend: {TEAM_PATH}/prompts/dev-frontend.md + {TEAM_PATH}/workflows/roles/frontend-specialized-tests.md + {TEAM_PATH}/templates/frontend-feature-test-template.md
- Backend: {TEAM_PATH}/prompts/dev-backend.md + {TEAM_PATH}/workflows/roles/specialized-tests.md + {TEAM_PATH}/templates/feature-test-template.md

## subtype determination (must execute)

Determine subtype based on task description and involved files:
- Involves web/src/**, *.tsx, *.css, React -> subtype = frontend-dev
- Involves internal/**, *.go, proto, API -> subtype = backend-dev

## beads operations
- View task: flow task show <id>
- Update progress: flow task update <id> --notes "PROGRESS: ..."
- Phase transition: flow task update <id> --add-label phase:xxx --remove-label phase:yyy
- Completion: flow task update <id> --add-label phase:review --remove-label phase:verify

## Toolchain Gate (IMPORTANT - violation = task failure)

{TOOLCHAIN_GATE}

## Rules
- After completion: flow task update {beads_id} --notes "COMPLETED: {deliverable list}"
- Return to Triage with deliverables list
- Do NOT re-read triage.md or triage-clarify.md (Triage already processed)
```

### Bug Task (bugfix-expert)

```markdown
You are Bugfix, executing task {beads_id}: {external-ref}: {title}

## Task Context (from Triage - Layer 1)
- Type: bug
- Priority: {0-4}
- Description: {user's original words}
- Key constraints: {extracted from classification}

## Required Reading (Layer 2 - load now)
1. {TEAM_PATH}/workflows/shared.md
2. {TEAM_PATH}/prompts/bugfix.md
3. Run: flow config paths --json

## On-Demand Reading (Layer 3 - load when needed)
- Standards: {TEAM_PATH}/workflows/roles/bugfix-standards.md
- Templates: {TEAM_PATH}/templates/bug-template.md
- Commands: {TEAM_PATH}/docs/COMMANDS.md
- Frontend: {TEAM_PATH}/workflows/roles/frontend-specialized-tests.md + {TEAM_PATH}/templates/frontend-bug-test-template.md
- Backend: {TEAM_PATH}/workflows/roles/specialized-tests.md + {TEAM_PATH}/templates/bug-test-template.md

## IMPORTANT - violating any rule = task failure, forbidden to report "complete"

### Execution order (must strictly follow, forbidden to skip steps)

```
Step 1: Query R iteration count -> flow task show <id> --json | count reopened events + 1
Step 2: Create/update report directory
  - First time: mkdir {DOCS_INTERNAL}/reports/bugs/B{NNN}/R1/
  - Subsequent: mkdir {DOCS_INTERNAL}/reports/bugs/B{NNN}/R{n}/
  - Update INDEX.md (mark current R, historical R marked as failed)
Step 3: Create R{n}/RCA.md -> Must include: Bug description / Impact scope / Root cause / Root cause type / Fix plan / Prevention measures
Step 4: Write Bug reproduction test -> Test must fail (red)
Step 5: Fix Bug -> Reproduction test must pass (green)
Step 6: Create R{n}/TEST_CASE.md -> Must include: Reproduction steps / Expected result / Verification result
Step 7: Full regression test passes
Step 8: flow task update <id> --add-label phase:verify
```

### Gate 1: RCA.md (Step 3 output)
- Path: {DOCS_INTERNAL}/reports/bugs/B{NNN}/R{N}/RCA.md
- RCA.md not created -> Forbidden to start fix code
- Writing RCA after fix code -> Forbidden, must write RCA first then fix

### Gate 2: Bug reproduction test (Step 4 output)
- Backend: {PROJECT}/tests/bugs/B{NNN}-{name}/regression_*.go
- Frontend: {PROJECT}/web/tests/bugs/B{NNN}-{name}/regression_*.test.{ts,tsx}
- No reproduction test -> Forbidden to report "complete"

### Gate 3: TEST_CASE.md (Step 6 output)
- Path: {DOCS_INTERNAL}/reports/bugs/B{NNN}/R{N}/TEST_CASE.md
- TEST_CASE.md not created -> Forbidden to report "complete"

### Gate 4: R iteration rules
- First fix -> B{NNN}/R1/
- R1 verification fails -> B{NNN}/R2/ (forbidden to overwrite R1)
- Forbidden to create directory without R subdirectory

### Gate 5: subtype determination + conditional loading
- Involves web/src/**, *.tsx -> subtype = frontend-dev
- Involves internal/**, *.go -> subtype = backend-dev

When subtype = frontend-dev, load Layer 3 frontend files.
When subtype = backend-dev, load Layer 3 backend files.

### Gate 6: Full regression (Step 7)
- Backend: go test ./... must pass
- Frontend: bun run test + bun run typecheck must pass
- Full test failure -> Forbidden to report "complete"

## Completion blockers (any missing = forbidden to report "complete")

| # | Check Item | Verification Method |
|---|------------|---------------------|
| 1 | INDEX.md exists and current_iteration points to current R | Read file to confirm |
| 2 | R{n}/RCA.md file exists | Read file to confirm |
| 3 | R{n}/TEST_CASE.md file exists | Read file to confirm |
| 4 | Bug reproduction test code exists | Read file to confirm |
| 5 | Reproduction test passes | Execute test to confirm |
| 6 | Full regression test passes | Execute test to confirm |
| 7 | Report directory format B{NNN}/R{N}/ | Check path |

## beads operations
- View task: flow task show <id>
- Update progress: flow task update <id> --notes "PROGRESS: ..."
- Phase transition: flow task update <id> --add-label phase:xxx --remove-label phase:yyy
- Completion: flow task update <id> --add-label phase:review --remove-label phase:verify

## Toolchain Gate

{TOOLCHAIN_GATE}

## Rules
- After completion (must follow this order):
  1. Verify each item in "completion blockers" table
  2. Any missing -> Report "task failed: missing {specific item}", forbidden to mark complete
  3. All pass -> flow task update <id> --add-label phase:review
  4. flow task update {beads_id} --notes "COMPLETED: RCA + TEST_CASE + fix"
  5. Report deliverable list (must include all file paths)
- Return to Triage with deliverables list
- Do NOT re-read triage.md or triage-clarify.md (Triage already processed)
```

### General Task (Change/Analysis/Docs/DevOps/UI/PM)

```markdown
You are {role}, executing task {beads_id}: {external-ref}: {title}

## Task Context (from Triage - Layer 1)
- Type: {change|analysis|docs|devops|ui|pm}
- Priority: {0-4}
- Description: {user's original words}
- Key constraints: {extracted from classification}

## Required Reading (Layer 2 - load now)
1. {TEAM_PATH}/workflows/shared.md
2. {TEAM_PATH}/prompts/{role}.md
3. Run: flow config paths --json

## On-Demand Reading (Layer 3 - load when needed)
- Standards: {TEAM_PATH}/workflows/roles/{role}-standards.md
- Templates: {TEAM_PATH}/templates/{type}-template.md
- Commands: {TEAM_PATH}/docs/COMMANDS.md

## beads operations
- View task: flow task show <id>
- Update progress: flow task update <id> --notes "PROGRESS: ..."
- Phase transition: flow task update <id> --add-label phase:xxx --remove-label phase:yyy
- Completion: flow task update <id> --add-label phase:review

## Toolchain Gate (IMPORTANT - violation = task failure)

{TOOLCHAIN_GATE}

## Rules
- After completion: flow task update {beads_id} --notes "COMPLETED: {deliverable list}"
- Return to Triage with deliverables list
- Do NOT re-read triage.md or triage-clarify.md (Triage already processed)
```

---

## TOOLCHAIN_GATE Dynamic Construction Rules

When Triage dispatches, read `.team/project.md TOOLCHAIN` section and dynamically fill `{TOOLCHAIN_GATE}`:

### Construction Steps

```
1. Read project.md TOOLCHAIN section
   -> frontend.package_manager = {pkg}  (bun/npm/pnpm/yarn)
   -> frontend.pipeline = {available command mapping}

2. Construct forbidden list:
   -> List all package manager commands other than {pkg} as forbidden

3. Construct available command list:
   -> Extract from pipeline, only list command names and usage (does not imply execution order)

4. Generate PRE-FLIGHT confirmation requirement
```

### Template (bun example)

```
This project uses bun for frontend.

Forbidden:
- npm install / npm run / npm test -> Use bun install / bun run / bun test
- npx xxx -> Use bunx xxx
- pnpm add / pnpm run -> Use bun add / bun run
- yarn dev / yarn build -> Use bun run dev / bun run build

Available commands (use as needed, not all required):
- bun run format      Formatting
- bun run lint        Code style check
- bun run lint:fix    Auto-fix style issues
- bun run typecheck   Type check
- bun run build       Build
- bun run test        Unit tests
- bun run test:watch  Watch mode tests
- bun run test:coverage Coverage report
- bun run check       Comprehensive check

Must output before any command: PRE-FLIGHT: pkg=bun
```

### Template (npm example)

```
This project uses npm for frontend.

Forbidden:
- bun install / bun run -> Use npm install / npm run
- pnpm add / pnpm run -> Use npm install / npm run
- yarn dev / yarn build -> Use npm run dev / npm run build
- npx xxx -> Use npx xxx (npm projects allow npx)

Available commands (use as needed, not all required):
- npm run format      Formatting
- npm run lint        Code style check
- npm run build       Build
- npm test            Unit tests

Must output before any command: PRE-FLIGHT: pkg=npm
```

### Key Constraints

1. **"Available commands" != "must execute"** - Clearly mark "use as needed, not all required", avoid AI interpreting as pipeline flow
2. **Forbidden list must be specific** - List each forbidden package manager's common commands with correct alternatives
3. **PRE-FLIGHT is hard gate** - Executing commands without confirmation = task failure

---

## Status Management

### Status Values

| Status | Meaning | Transitions |
|--------|---------|-------------|
| `open` | Ready to pick up | -> in_progress, blocked |
| `in_progress` | Being worked on | -> blocked, closed |
| `blocked` | Waiting on dependency | -> open, in_progress |
| `closed` | Done | -> open (reopen) |

### Phase Tracking (via Labels)

Use labels to track fine-grained phases:

```
phase:ready      -> Just created, ready for dispatch
phase:analyze    -> Under investigation (Tech Lead)
phase:design     -> Writing spec/design doc
phase:implement  -> Development in progress
phase:verify     -> Testing/QA phase
phase:review     -> Waiting for user/business confirmation
```

```bash
# Progress to next phase
flow task update <id> \
  --add-label phase:implement \
  --remove-label phase:design
```

### Status Flow (Triage Responsible)

```
open (phase:ready) -> in_progress (phase:implement) -> in_progress (phase:verify) -> in_progress (phase:review) -> closed
```

- **open -> in_progress**: When Triage launches sub-agent
- **in_progress (verify) -> in_progress (review)**: After sub-agent completes deliverables
- **in_progress (review) -> closed**: After user confirms, Triage archives

```bash
# Launch sub-agent: claim and start
flow task update <id> --claim --add-label phase:implement --remove-label phase:ready

# Sub-agent completes: move to review
flow task update <id> --add-label phase:review --remove-label phase:verify

# User confirms: close
flow task close <id> --reason "Confirmed by user"
```

---

## Beads Storage Layer

### Path Protection (Highest Priority)

> **beads database must be in project directory `.beads/`, forbidden to write to framework layer `{TEAM_PATH}/`.**
>
> | Item | Correct Path (USE THIS) | WRONG Path (NEVER) |
> |------|------------------------|-------------------|
> | beads DB | `{PROJECT}/.beads/` | `{TEAM_PATH}/.beads/` |
> | task-pool-export | `{DOCS_INTERNAL}/task-pool-export.md` | `{TEAM_PATH}/task-pool-export.md` |
>
> `{TEAM_PATH}/` is framework layer, cross-project shared, AI read-only at runtime. Wrong path = cross-project contamination.

### Storage Locations

```
{PROJECT}/.beads/                <- beads database (single source of truth)
{DOCS_INTERNAL}/task-pool-export.md       <- Human-readable export (read-only)
```

### ID Lookup Rules

```bash
# From beads ID find task ID (external-ref)
flow task show <beads-id> --json | jq '.externalRef'
# -> "F014"

# From task ID find beads ID
flow task list --json | jq '.[] | select(.externalRef == "F014") | .id'
# -> "<beads-id>"

# Query Bug R iteration count (reopen count + 1)
flow task show <beads-id> --json | jq '[.events[] | select(.event_type == "reopened")] | length + 1'
# -> 2 (means current is R2)

# From task ID locate doc directory
# F014 -> {DOCS_INTERNAL}/requirements/F014-unified-pagination/
# B001 -> {DOCS_INTERNAL}/reports/bugs/B001/  (directory without R suffix)
# A008 -> {DOCS_INTERNAL}/reports/analysis/A008-quality-check-enhancement.md
```

### Duplicate Detection

Before creating, check for existing issues:

```bash
# Search by keywords
flow task list --json | jq '.[] | select(.title | contains("keyword"))'

# Check by external-ref
flow task list --json | jq '.[] | select(.externalRef == "F014")'

# Search closed issues (might be regression)
flow task list --status closed --json | jq '.[] | select(.title | contains("keyword"))'
```

If duplicate found, add note instead of creating new:

```bash
flow task update <existing-id> --append-notes "Additional report: [new context]"
```

### Dependency Management

```bash
# Link dependencies
flow task dep add <new-id> <dependency-id> --type discovered-from
flow task dep add <new-id> <blocking-id> --type blocks
flow task dep add <new-id> <related-id> --type related-to
```

Dependency types:
- `discovered-from`: This issue was found while investigating another
- `blocks`: This issue must be done before the other can proceed
- `related-to`: Loosely related, informational

---

### Session Start Scan (Background)

**Purpose**: Quick status check, does not block main flow.
```bash
# Step 1: Scan Review phase issues
flow task list --label phase:review --json

# Step 2: Scan LESSON-NEEDED labels
flow task list --label lesson:needed --json
```

**Output format** (brief, no blocking):

```markdown
## Session Status
**Pending Review**: 1 issue (F001)
**Pending LESSON**: 2 issues (B001, B002)

[View details] / [Continue main flow]
```

Only show full list if user asks for details.

### Export for Human Review

At end of session or on demand, export beads state for human readability:

```bash
# Export open issues (table format)
flow task list --status open --format table > {DOCS_INTERNAL}/task-pool-export.md

# Export P0/P1 only
flow task list --priority 0,1 --format table > {DOCS_INTERNAL}/task-pool-urgent.md

# Full JSON export
flow task list --json > {DOCS_INTERNAL}/task-pool-full.json

# Manual export anytime via flow command
flow export
```

**Export rules**:
1. Triage exports task status to `{DOCS_INTERNAL}/task-pool-export.md` after every status change
2. CLI command: `flow export` - manual export anytime
3. task-pool-export.md is **read-only** - never edit it to change task state

### Session End

At end of triage session:

```bash
# 1. Create handoff document (MANDATORY, cannot skip)
# Format: {DOCS_INTERNAL}/handoff-YYYY-MM-DD.md
# See shared.md for full template and requirements

# 2. Export current state
flow task list --status open --format table > {DOCS_INTERNAL}/task-pool-export.md

# 3. Commit Dolt changes (if batch mode)
flow task dolt commit -m "Triage session $(date +%Y%m%d)"

# 4. Push to remote
flow task dolt push
```

**MANDATORY**:
- Handoff document must be created even if no code changes were made
- See `{TEAM_PATH}/workflows/shared.md` for:
  - Exact template format
  - Required content checklist
  - Prohibited items

### Metrics to Track

```bash
# Issues created this session
flow task log --actor $USER --action create --since "2 hours ago"

# Issues closed this session
flow task log --actor $USER --action close --since "2 hours ago"

# Current state
flow task stats
```
