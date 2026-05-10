# Triage: Post-Verification & Review (loaded on demand)

> **Loading rule**: 锟?Bug 杩斿洖鍚庨獙璇併€丷eview 纭銆佸垎鍙戝瓙 Agent銆佺鐞嗙姸锟?Session 鏃跺姞杞芥鏂囦欢
> **Source**: 锟?`triage.md` 鎷嗗垎锛寁17.0

---

## Bug Post-Verification (Triage Executes)

After bugfix-expert returns, Triage must verify:

### Part 1: File Existence Check

| # | Verification Item | Expected Path | If Missing |
|---|--------|---------|--------|
| 1 | RCA.md | {DOCS_INTERNAL}/reports/bugs/B{NNN}/R{N}/RCA.md | Mark as "RCA missing", require completion |
| 2 | TEST_CASE.md | {DOCS_INTERNAL}/reports/bugs/B{NNN}/R{N}/TEST_CASE.md | Mark as "TEST_CASE missing", require completion |
| 3 | Reproduction test code | tests/bugs/B{NNN}-*/ or web/tests/bugs/B{NNN}-*/ | Mark as "reproduction test missing", require completion |

### Part 2: Test Execution Verification (IMPORTANT 锟?must run actual tests)

鈿狅笍 Triage MUST execute the following commands to verify the fix actually works:

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
```

**鈿狅笍 Frontend Bug 棰濆蹇呴』: UI 杩愯鏃堕獙锟?*:
```bash
# Step 5: Start dev server and verify UI
bun run dev
# 鐒跺悗蹇呴』鎵ц:
# - 鎵撳紑 Bug 娑夊強鐨勯〉闈紙瀵艰埅锟?Bug URL锟?# - 妫€鏌ラ〉闈㈠唴瀹癸紙鏂囨湰/鏁版嵁/缁勪欢姝ｇ‘鏄剧ず锟?# - 鎵ц Bug 娑夊強鐨勪氦浜掞紙鐐瑰嚮/杈撳叆/鎻愪氦锟?# - 閲嶇幇 Bug 鍘熷瑙﹀彂姝ラ锛岀‘锟?Bug 鐜拌薄娑堝け
# - 鎴浘淇濆瓨锟?{DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/
```

| # | Verification Item | Check Method | If Failed |
|---|--------|---------|--------|
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
|---|--------|---------|--------|
| 10 | RCA.md contains "Data Flow Tracing" section | Read RCA.md, search for "Data Flow Tracing" | Mark as "missing data flow tracing" |
| 11 | Data flow tracing includes breakpoint analysis | Read RCA.md, search for "breakpoint" | Mark as "incomplete data flow tracing" |
| 12 | TEST_CASE.md includes real-scenario verification | Read TEST_CASE.md, search for "real scenario" | Mark as "missing real-scenario verification" |

### Part 4: R Iteration Quality Check (must execute for R2+)

| # | Verification Item | Check Method | If Missing |
|---|--------|---------|--------|
| 13 | RCA.md contains previous round failure analysis | Read RCA.md, search for "R{N-1}" or "previous round" | Mark as "missing R iteration analysis" |
| 14 | R iteration >= R4, has user been asked to confirm? | Check beads notes | Mark as "R4+ not paused" |

### Verification Result

- All present + All tests pass -> Update beads: `flow task update <id> --add-label phase:review --remove-label phase:verify`, assignee=QA
- Any missing or test failure -> Update beads: `flow task update <id> --notes "MISSING: {specific items}, need completion"`, send back to bugfix-expert
- Report verification results to user (include test output)

**Forbidden**: Creating RCA.md / TEST_CASE.md. These are created by bugfix-expert. Triage only verifies, never creates.

---

## R-Iteration Handling

### R 杩唬瑙勫垯

```
Bug 淇锟?QA 楠岃瘉
    锟?    鈹溾攢 楠岃瘉閫氳繃 锟?杩涘叆 Review 纭娴佺▼
    锟?    鈹斺攢 楠岃瘉鏈€氳繃
        锟?    璇㈤棶鐢ㄦ埛: "Bug 鏈慨濂斤紝闇€锟?R{N+1} 杩唬锟?
        锟?    鈹屸攢 锟?锟?鎵撳洖 bugfix-expert锛岃Е锟?R{N+1}
    锟?  鈹斺攢 鍒涘缓鏂扮洰锟?B{NNN}/R{N+1}/锛堢姝㈣锟?R{N}锟?    锟?  鈹斺攢 flow task update <id> --add-label phase:implement --remove-label phase:verify
    锟?    鈹斺攢 锟?锟?鍒涘缓锟?Bug 鎴栧叧锟?```

### R 杩唬璐ㄩ噺闂ㄧ

| 杩唬 | 棰濆瑕佹眰 |
|------|---------|
| R1 | 鏍囧噯 Bug 淇娴佺▼ |
| R2 | RCA 蹇呴』鍖呭惈 R1 澶辫触鍒嗘瀽 |
| R3 | RCA 蹇呴』鍖呭惈 R1+R2 澶辫触鍒嗘瀽 + 鏇夸唬鏂规璇勪及 |
| R4+ | 鈿狅笍 蹇呴』鏆傚仠锛岃闂敤鎴锋槸鍚︾户缁凯锟?|

**R4+ 鏆傚仠瑙勫垯**锟?- 杈惧埌 R4 鏃讹紝Triage 蹇呴』閫氱煡鐢ㄦ埛
- 鐢ㄦ埛纭鍚庢墠鍏佽缁х画 R4
- 寤鸿锛歊4+ 鏃惰€冭檻閲嶆柊璇勪及 Bug 鏍瑰洜鎴栧崌绾у锟?
---

## Review Confirmation & Archiving

### 瑙﹀彂鏃舵満

QA 瀹屾垚楠岃瘉鍚庯紝蹇呴』鎵ц Review 纭娴佺▼锟?
```
QA 楠岃瘉閫氳繃
    锟?Triage 鎵ц Review 纭
    锟?鈹屸攢 Bug 绫诲瀷 锟?妫€锟?LESSON-NEEDED
鈹斺攢 Feature/Change 锟?鍙€夋锟?    锟?杈撳嚭纭鎶ュ憡 锟?鐢ㄦ埛纭
    锟?鈹屸攢 纭 锟?鍏抽棴 task
鈹斺攢 鏈夐棶锟?锟?鎵撳洖鎴栧垱寤烘柊 Bug
```

### 鍏抽棴鍓嶆锟?
```bash
# Bug 绫诲瀷蹇呴』妫€锟?flow task list --label lesson:needed --json | jq '.[] | select(.externalRef == "B001")'

# 濡傛灉锟?LESSON-NEEDED 鏍囩
flow task show <id> --json | jq '.labels'
```

```markdown
## 鈿狅笍 鍏抽棴鍓嶆锟?
**Task**: B001 (Bug)
**LESSON-NEEDED**: 鈿狅笍 锟?1 涓緟澶勭悊

| 鏉ユ簮 | 鎻忚堪 | 鐘讹拷?|
|------|------|------|
| Bugfix | 鏈鐞嗙┖鎸囬拡寮傚父 | 寰呭锟?|

**璇烽€夋嫨**:
[A] 鍏堝锟?LESSON-NEEDED
[B] 璺宠繃锛岀◢鍚庡锟?```

### Review 纭娴佺▼

```bash
# 鏌ユ壘鎵€锟?phase:review 锟?issue
flow task list --label phase:review --json
```

```markdown
## Review 寰呯‘锟?
| ID | 绫诲瀷 | 鎻忚堪 | 浜や粯锟?| 鏉ユ簮 | LESSON |
|----|------|------|--------|------|--------|
| F001 | Feature | 鐢ㄦ埛鐧诲綍 | SPEC.md, AC.md | Tech Lead | 锟?|
| B001 | Bug | 鐧诲綍瓒呮椂 | RCA.md, TEST_CASE.md | Bugfix | 鈿狅笍 1 |

---

**鎿嶄綔**: [纭鍏ㄩ儴] / [閫愪釜纭] / [鏈夐棶棰樼殑鎵撳洖]
```

### QA 鍙戠幇 Bug 鐨勫锟?
锟?QA 锟?Review 杩囩▼涓彂鐜版柊 Bug 鏃讹細

```
QA 鍙戠幇锟?Bug
    锟?璇㈤棶鐢ㄦ埛: "鍙戠幇 X 鐜拌薄锛岃繖鏄柊 Bug 杩樻槸 B001 锟?R2锟?
    锟?鈹屸攢 B001 锟?R2 锟?璇㈤棶: "Bug 鏈慨濂斤紝闇€锟?R2 杩唬锟?
锟?  鈹斺攢 锟?锟?鎵撳洖 B001锛岃Е锟?R2
锟?  鈹斺攢 锟?锟?鍒涘缓锟?Bug
锟?鈹斺攢 锟?Bug 锟?鍒涘缓 B{N+1}
```

```markdown
## 鈿狅笍 QA 鍙戠幇 Bug

**鐜拌薄**: {鎻忚堪}
**鍙兘鏉ユ簮**:
- B001 鏈慨锟?锟?R2 杩唬
- 锟?Bug 锟?鍒涘缓 B{N+1}

**璇风‘锟?*:
[A] B001 锟?R2
[B] 锟?Bug
[C] 瑙傚療璁板綍锛屼笉鍒涘缓浠诲姟
```

### Task 瀹屾垚楠岃瘉

鍏抽棴 Task 鍓嶇殑鏈€缁堟鏌ワ細

| # | 妫€鏌ラ」 | 楠岃瘉鏂规硶 |
|---|--------|---------|
| 1 | 鎵€鏈変氦浠樼墿宸茬敓锟?| 妫€鏌ユ枃浠惰矾锟?|
| 2 | beads 鐘舵€佸凡鏇存柊 | `flow task show <id>` |
| 3 | LESSON-NEEDED 宸插鐞嗭紙Bug 绫诲瀷锟?| `flow task list --label lesson:needed` |
| 4 | 鐢ㄦ埛宸茬‘锟?| 纭璁板綍 |
| 5 | 鏃犻仐锟?R 杩唬 | 妫€锟?beads notes |

---

## LESSON-NEEDED 鍚庡彴鎵弿

### 瑙﹀彂鏈哄埗

**涓嶉樆濉炰富娴佺▼**锛屼粎鍦ㄤ互涓嬫椂鏈烘彁绀猴細
- Session 鍚姩鏃讹紙绠€鐭彁绀猴級
- 鐢ㄦ埛涓诲姩璇㈤棶 "鏈夊摢浜涘緟澶勭悊锟?lesson"
- Session 缁撴潫锟?
```bash
# 鎵弿 LESSON-NEEDED 鏍囩
flow task list --label lesson:needed --json
```

### 鎵弿杈撳嚭

```markdown
## 鈿狅笍 寰呭锟?LESSON-NEEDED

| ID | 鏉ユ簮 | 鎻忚堪 | 鏍囪鏃堕棿 |
|----|------|------|----------|
| B001 | Bugfix | 鏈鐞嗙┖鎸囬拡寮傚父 | 2h ago |

[澶勭悊] / [绋嶅悗澶勭悊] / [蹇界暐]
```

### 澶勭悊娴佺▼

```
Triage 澶勭悊 LESSON-NEEDED
    锟?鐢熸垚 lesson 鍐呭
    锟?鍐欏叆 {DOCS_INTERNAL}/lessons/
    锟?flow task update <id> --remove-label lesson:needed --add-label lesson:done
```

---

## Sub-Agent Prompt Templates (涓夊眰浜ゆ帴妯″瀷)

> **Loading rule**: 锟?Triage 鍒嗗彂锟?Agent 鏃跺姞杞芥锟?
### Three-Layer Handoff Model

```
Layer 1: Triage MUST pass (inject into sub-agent prompt)
  鈹溾攢鈹€ Task ID + title + description
  鈹溾攢鈹€ Task type (Feature/Bug/Change)
  鈹溾攢鈹€ Priority
  鈹斺攢鈹€ Key context (user's original words, error messages, etc.)

Layer 2: Sub-agent MUST read on startup (load immediately)
  鈹溾攢鈹€ {TEAM_PATH}/workflows/shared.md (core rules)
  鈹溾攢鈹€ Corresponding role prompt file (e.g., {TEAM_PATH}/prompts/dev.md)
  鈹斺攢鈹€ flow config paths --json (path variables)

Layer 3: Sub-agent reads on demand (load when needed)
  鈹溾攢鈹€ {TEAM_PATH}/workflows/roles/xxx-standards.md
  鈹溾攢鈹€ {TEAM_PATH}/templates/xxx-template.md
  鈹溾攢鈹€ {DOCS_INTERNAL}/reports/... (specific documents)
  鈹斺攢鈹€ {TEAM_PATH}/docs/COMMANDS.md
```

### Handoff Key Rules (浜ゆ帴閾佸緥)

1. **Triage MUST NOT pass shared.md content in the prompt** 锟?娴垂 token锛屽瓙 Agent 鑷璇诲彇 Layer 2
2. **Triage MUST pass user's original words verbatim** 锟?涓嶆憳瑕併€佷笉杞堪锛屼繚鐣欏師濮嬫帾锟?3. **Sub-agent MUST read Layer 2 files before starting work** 锟?鍚姩鍚庣涓€浠朵簨鏄姞锟?Layer 2
4. **Sub-agent MUST NOT read triage.md or triage-clarify.md** 锟?杩欐槸 Triage 鐨勪笂涓嬫枃锛屼笉鏄瓙 Agent 锟?5. **Sub-agent reads Layer 3 files only when the specific task requires it** 锟?鎸夐渶鍔犺浇锛屼笉棰勮

### Feature Task (developer-engineer)

```markdown
You are Dev, executing task {beads_id}: {external-ref}: {title}

## Task Context (from Triage 锟?Layer 1)
- Type: feature
- Priority: {0-4}
- Description: {user's original words}
- Key constraints: {extracted from classification}

## Required Reading (Layer 2 锟?load now)
1. {TEAM_PATH}/workflows/shared.md
2. {TEAM_PATH}/prompts/dev.md
3. Run: flow config paths --json

## On-Demand Reading (Layer 3 锟?load when needed)
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

## Toolchain Gate (IMPORTANT 锟?violation = task failure)

{TOOLCHAIN_GATE}

## Rules
- After completion: flow task update {beads_id} --notes "COMPLETED: {deliverable list}"
- Return to Triage with deliverables list
- 锟?Do NOT re-read triage.md or triage-clarify.md (Triage already processed)
```

### Bug Task (bugfix-expert)

```markdown
You are Bugfix, executing task {beads_id}: {external-ref}: {title}

## Task Context (from Triage 锟?Layer 1)
- Type: bug
- Priority: {0-4}
- Description: {user's original words}
- Key constraints: {extracted from classification}

## Required Reading (Layer 2 锟?load now)
1. {TEAM_PATH}/workflows/shared.md
2. {TEAM_PATH}/prompts/bugfix.md
3. Run: flow config paths --json

## On-Demand Reading (Layer 3 锟?load when needed)
- Standards: {TEAM_PATH}/workflows/roles/bugfix-standards.md
- Templates: {TEAM_PATH}/templates/bug-template.md
- Commands: {TEAM_PATH}/docs/COMMANDS.md
- Frontend: {TEAM_PATH}/workflows/roles/frontend-specialized-tests.md + {TEAM_PATH}/templates/frontend-bug-test-template.md
- Backend: {TEAM_PATH}/workflows/roles/specialized-tests.md + {TEAM_PATH}/templates/bug-test-template.md

## IMPORTANT 锟?violating any rule = task failure, forbidden to report "complete"

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
|---|--------|---------|
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
- 锟?Do NOT re-read triage.md or triage-clarify.md (Triage already processed)
```

### General Task (Change/Analysis/Docs/DevOps/UI/PM)

```markdown
You are {role}, executing task {beads_id}: {external-ref}: {title}

## Task Context (from Triage 锟?Layer 1)
- Type: {change|analysis|docs|devops|ui|pm}
- Priority: {0-4}
- Description: {user's original words}
- Key constraints: {extracted from classification}

## Required Reading (Layer 2 锟?load now)
1. {TEAM_PATH}/workflows/shared.md
2. {TEAM_PATH}/prompts/{role}.md
3. Run: flow config paths --json

## On-Demand Reading (Layer 3 锟?load when needed)
- Standards: {TEAM_PATH}/workflows/roles/{role}-standards.md
- Templates: {TEAM_PATH}/templates/{type}-template.md
- Commands: {TEAM_PATH}/docs/COMMANDS.md

## beads operations
- View task: flow task show <id>
- Update progress: flow task update <id> --notes "PROGRESS: ..."
- Phase transition: flow task update <id> --add-label phase:xxx --remove-label phase:yyy
- Completion: flow task update <id> --add-label phase:review

## Toolchain Gate (IMPORTANT 锟?violation = task failure)

{TOOLCHAIN_GATE}

## Rules
- After completion: flow task update {beads_id} --notes "COMPLETED: {deliverable list}"
- Return to Triage with deliverables list
- 锟?Do NOT re-read triage.md or triage-clarify.md (Triage already processed)
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

1. **"Available commands" != "must execute"** 锟?Clearly mark "use as needed, not all required", avoid AI interpreting as pipeline flow
2. **Forbidden list must be specific** 锟?List each forbidden package manager's common commands with correct alternatives
3. **PRE-FLIGHT is hard gate** 锟?Executing commands without confirmation = task failure

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

### Session 鍚姩鎵弿 (鍚庡彴鎵ц)

**鐩殑**: 蹇€熻幏鍙栫姸鎬侊紝涓嶉樆濉炰富娴佺▼锟?
```bash
# Step 1: 鎵弿 Review 闃舵 issue
flow task list --label phase:review --json

# Step 2: 鎵弿 LESSON-NEEDED 鏍囩
flow task list --label lesson:needed --json
```

**杈撳嚭鏍煎紡** (绠€鐭紝涓嶉樆锟?:

```markdown
## Session 鐘讹拷?
**寰呯‘锟?Review**: 1 锟?(F001)
**寰呭锟?LESSON**: 2 锟?(B001, B002)

[鏌ョ湅璇︽儏] / [缁х画涓绘祦绋媇
```

濡傛灉鐢ㄦ埛璇㈤棶璇︽儏锛屾墠杈撳嚭瀹屾暣鍒楄〃锟?
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
2. CLI command: `flow export` 锟?manual export anytime
3. task-pool-export.md is **read-only** 锟?never edit it to change task state

### Session End

At end of triage session:

```bash
# 1. Create handoff document (MANDATORY, cannot skip)
# Format: {DOCS_INTERNAL}/handoff-YYYY-MM-DD.md
# See shared.md 搂浼氳瘽缁撴潫浜ゆ帴鏂囨。 for full template and requirements

# 2. Export current state
flow task list --status open --format table > {DOCS_INTERNAL}/task-pool-export.md

# 3. Commit Dolt changes (if batch mode)
flow task dolt commit -m "Triage session $(date +%Y%m%d)"

# 4. Push to remote
flow task dolt push
```

**鈿狅笍 MANDATORY**: 
- Handoff document must be created even if no code changes were made
- See `{TEAM_PATH}/workflows/shared.md` 搂浼氳瘽缁撴潫浜ゆ帴鏂囨。 for:
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

