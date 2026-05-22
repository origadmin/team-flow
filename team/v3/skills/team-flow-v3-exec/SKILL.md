---
name: team-flow-v3-exec
version: 3.2
description: |
  Execute v3 flows by running `flow proc run` and following the engine's structured output.
  Translates engine output into concrete AI actions: role adoption, rule enforcement,
  work dispatch, deliverable production, gate evaluation, and flow progression.
---

# team-flow-v3-exec — Flow Execution Skill

## Purpose

Execute a validated v3 flow. The engine (`flow proc run`) produces structured instructions at each node — this skill tells you **exactly how to translate those instructions into AI actions**.

## Core Principle

**The engine is the source of truth.** You do not interpret the flow JSON directly. You run `flow proc run` and follow the structured output. This skill is the bridge between engine output and AI behavior.

---

## Status Line (MANDATORY — Every Response)

Every AI response MUST start with a status line. Data comes from `flow proc run` output — never hardcode.

**Format**: `[Role: {alias} | Flow: {flow-name} | Node: {node-id} | Phase: {phase}]`

| Field | Source | Example |
|-------|--------|---------|
| Role | `current.alias` (or `current.role` if no alias) | 齐活林 |
| Flow | `flow.name` | dev-flow |
| Node | `current.node_id` | fa01 |
| Phase | `current.on_enter` → action `update_task_phase` → `phase` field | analyze |

**When beads task is active** (engine returns `task.beads_id`):

**Format**: `[Role: {alias} | Flow: {flow-name}#{beads-id} | Node: {node-id} | Phase: {phase}]`

**Examples**:
```
[Role: 齐活林 | Flow: dev-flow | Node: fa01 | Phase: analyze]
[Role: 齐活林 | Flow: dev-flow#team-flow-6x9 | Node: fa01 | Phase: analyze]
[Role: 寇豆码 | Flow: dev-flow#team-flow-6x9 | Node: fd03 | Phase: implement]
```

**⛔ Forbidden**: Empty status line, hardcoded values, or skipping status line.

---

## Session Startup Protocol

When a new session starts, execute in order:

```
Step 1: Read CONSENSUS.md
  → {TEAM_PATH}/../CONSENSUS.md (v3 共识文件)
  → Contains all confirmed decisions. Do NOT re-ask anything already decided.

Step 2: Read project configuration
  → Read .team/project.md
  → Extract: default_flow name, docs_path, TOOLCHAIN

Step 3: Start the flow
  → flow proc run
  → Engine reads default_flow from project config automatically
  → Receive first node instruction

Step 4: Adopt the principal role
  → The first node's role with principal: true is your identity
  → Read persona, traits, guidance from engine output
  → Announce yourself to the user

Step 4.5: Team Introduction (first session only)
  → When flow proc run outputs the TEAM_INTRO block (🏠 Welcome to...), use it to introduce the team
  → Present the team in a natural, conversational way:
     "你好！我是 {alias}，你的{role_name}。欢迎加入我们的团队！

      我们团队有这些成员：
      - ⭐ {alias}({alias_en}) — {role_name}：[brief description based on persona]
      - 🔧 {alias}({alias_en}) — {role_name}：[brief description based on persona]
      ...

      工作流程：
      1. 你告诉我你需要什么
      2. 我分类并分发给最合适的专家
      3. 专家完成工作并产出交付物
      4. 我验证后向你汇报

      可用流程：{list flows from TEAM_INTRO}
      当前使用：{default_flow} ⭐

      有什么我可以帮你的？"
  → This introduction helps users understand the team's capabilities and workflow
  → Only show on first flow proc run (no nodeID specified); subsequent runs skip this
  → Wait for user input
```

**If no flow is bound** (default_flow is empty or points to non-existent flow):

```
Step 1: Run flow proc list
  → Check which flows are available and registered

Step 2: Present options to user
  → If preset flows exist:
      "这个项目还没有绑定团队流程。我找到了以下可用团队：
       1. dev-flow — 软件开发全流程
       2. novel-flow — 小说创作流程
       ...
       选择一个最接近你需求的，或者告诉我你的具体场景，我帮你创建专属团队。"
  → If no flows at all:
      "这个项目还没有任何团队流程。告诉我你的需求场景，我帮你创建一个专属团队。"

Step 3: Handle user's choice
  → User selects existing flow:
      1. Edit .team/project.md → set default_flow: {chosen-flow}
      2. Re-run flow proc run → enter normal execution
  → User wants to create new team:
      1. Load team-flow-v3-create skill (via Skill tool)
      2. Follow create workflow (Step 0-10)
      3. After creation: register in project.md → set default_flow → re-run flow proc run
  → User wants to browse:
      1. Run flow proc show {name} for details
      2. Return to Step 2 after browsing

Step 4: Resume normal execution
  → After default_flow is set, continue from Session Startup Protocol Step 3
```

**⛔ Never proceed without a flow binding.** AC1 requires one project = one flow.

---

## Session Close Protocol

When the session ends (user says goodbye, or terminal node reached):

```
Step 1: Verify current node deliverables
  → Check all required docs exist and are non-empty
  → If deliverables incomplete → warn user before closing

Step 2: Git operations (if code was changed)
  → git add -A
  → git commit -m "session: {flow-name} {node-id} {date}"
  → git push
  → Verify push succeeded

Step 3: Update task status (if beads task active)
  → flow task update {beads-id} --notes "Session closed at node {node-id}"
  → If terminal node → flow task close {beads-id}

Step 4: Report to user
  → Summary of what was accomplished
  → List of deliverables produced
  → Current position in the flow (node ID + name)
  → Next steps for next session
```

---

## Execution Loop

The core execution loop runs continuously until a terminal node is reached:

```
1. flow proc run [{node-id}]
   → Receive structured instruction for the current node

2. Execute the node (see Node Execution Protocol below)
   → Adopt role, follow rules, produce deliverables

3. Choose next node from next_options
   → If only one option → proceed automatically
   → If conditional → evaluate condition and choose
   → If gate → determine pass/fail and route accordingly

4. flow proc run {next-node-id}
   → Repeat from step 2

5. Terminal reached → Execute Session Close Protocol
```

---

## Node Execution Protocol

This is the core of the exec skill — what AI does at each node.

### Phase Nodes (the most common type)

When `current.node_type == "phase"`:

```
Step 1: Execute on_enter actions
  → For each action in current.on_enter:
    - "update_task_phase" → Update task phase to the specified value
    - Other actions → Execute as specified

Step 2: Adopt the role
  → Set your identity: "I am {alias}({alias_en}), {role_name}"
  → Internalize persona: Read current.persona — this is your character
  → Internalize traits: Read current.traits — these are your behavioral tendencies
  → Follow guidance: Read current.guidance — this is your decision framework
  → If principal: true → You are the team's user interface

Step 3: Load rules
  → For each rule in current.rules:
    - Read rule.instruction — this is the immediate constraint (~30 chars, Chinese)
    - If you need detail → Run `flow proc rule {rule.ref}` to get full description
    - If rule.enforcement == "hard" → MUST follow, no exceptions
    - If rule.enforcement == "soft" → Best effort, but try hard

Step 4: Determine execution mode
  → IF principal == true:
      → Principal Execution Mode (see below)
  → ELSE:
      → Worker Execution Mode (see below)

Step 5: Produce deliverables
  → For each doc in current.docs:
    - Create the file at doc.path
    - Follow doc.description for content requirements
    - If doc.template is specified → Use it as the structure
    - If doc.content_rules is specified → Follow each rule
    - If doc.required == true → File MUST exist and be non-empty after execution

Step 6: Verify completion
  → All hard-enforcement rules followed?
  → All required docs produced at correct paths?
  → Only specified tools used?
  → Ready to move to next node
```

### Principal Execution Mode

When `current.principal == true`, you are the team's user interface. Your job is to **receive, classify, and dispatch** — NOT to execute directly.

```
P1: Receive user input
  → Listen to what the user says
  → If user sends new input mid-execution → Stop current work, process new input

P2: Classify the input
  → What type of work is this? (new task, clarification, feedback, interruption)
  → What does the flow say about handling this? (check next_options)
  → Which sub-role should handle this? (check next_options[].role)

P3: Dispatch to sub-role
  → Run: flow proc run {next-node-id}
  → Get the sub-role's node instruction
  → Use Task tool to dispatch (see Dispatch Protocol below)

P4: Wait for sub-agent completion
  → Sub-agent returns with deliverables list
  → Verify deliverables exist (see Deliverable Verification below)

P5: Decide next step
  → If deliverables verified → Move to next node
  → If deliverables missing → Re-dispatch or ask user
  → If sub-agent reported error → Decide: retry, reassign, or escalate to user

P6: Report to user
  → Summarize what was accomplished
  → Show deliverables produced
  → Explain what happens next
```

**⛔ Principal NEVER**:
- Writes code or modifies files directly (unless the flow explicitly assigns it such a phase)
- Skips dispatch and does the work itself
- Lets sub-roles communicate directly with the user

### Worker Execution Mode

When `current.principal == false` (or not set), you are a worker executing a specific task.

```
W1: Adopt your role identity
  → "I am {alias}({alias_en}), {role_name}"
  → Internalize persona, traits, guidance from engine output

W2: Understand the task
  → Read the task context from your dispatch prompt (see Dispatch Protocol)
  → Read the node description (current.description)
  → Understand what deliverables are expected (current.docs)

W3: Execute the work
  → Follow all rules (current.rules)
  → Use only specified tools (current.tools)
  → Load specified skills if any (current.skills) — invoke via Skill tool
  → Load specified prompts if any (current.prompts) — read the file at prompt.path

W4: Produce deliverables
  → Create each required doc at the specified path
  → Ensure content is meaningful — empty files don't count

W5: Self-verify (Output Guard)
  → Re-read each required doc — is it non-empty and meaningful?
  → Check all hard-enforcement rules were followed
  → Check only specified tools were used

W6: Report completion
  → Return to principal with:
    - List of deliverables produced (name + path)
    - Any issues or warnings encountered
    - Gate evaluation results (if this was a gate node)
```

---

## Dispatch Protocol

When the principal dispatches work to a sub-role, use the Task tool with this template:

### When to dispatch

- Principal receives user input that requires work from another role
- Flow progression reaches a non-principal node
- Sub-role completes and next node is another sub-role

### How to dispatch

```
Task(
  subagent_type="general_purpose_task",
  query="""
You are {alias}({alias_en}), {role_name}.

## Your Identity
- Persona: {current.persona}
- Traits: {current.traits}
- Guidance: {current.guidance}

## Your Task
- Node: {current.node_id} — {current.name}
- Description: {current.description}

## Rules (MUST follow)
{for each rule in current.rules:
  - [{rule.enforcement}] {rule.instruction}
  - Detail: {rule.rule_ref}
}

## Deliverables (MUST produce)
{for each doc in current.docs:
  - {doc.name} [{doc.format}] {if doc.required: [REQUIRED]}
    Path: {doc.path}
    {doc.description}
}

## Tools (ONLY use these)
{for each tool in current.tools:
  - {tool.ref} [{tool.commands}]
}

## Skills (load via Skill tool if needed)
{for each skill in current.skills:
  - {skill.ref}: {skill.description}
}

## Completion Rules
- Produce ALL required deliverables at the specified paths
- Follow ALL hard-enforcement rules
- Use ONLY the specified tools
- After completion, report: deliverables list + any issues
- Status Line: [Role: {alias} | Flow: {flow-name} | Node: {current.node_id} | Phase: {phase}]
"""
)
```

### For search/analysis tasks

If the sub-role's work is purely read-only (no file writes), use:

```
Task(
  subagent_type="search",
  query="..."
)
```

### After sub-agent returns

```
1. Read the sub-agent's completion report
2. Verify deliverables (see Deliverable Verification below)
3. If verified → flow proc run {next-node-id} to advance
4. If not verified → re-dispatch with specific instructions on what's missing
```

---

## Deliverable Verification

After any node execution (by yourself or sub-agent), verify deliverables:

```
For each doc in current.docs where doc.required == true:
  1. Check file exists at doc.path
  2. Check file is non-empty (content > 0 bytes)
  3. Check content is meaningful (not just placeholders or TODOs)
  4. If doc.format == "markdown" → Check has actual sections, not just headers
  5. If doc.content_rules → Check each rule is satisfied

If ANY required doc fails verification:
  → Do NOT proceed to next node
  → Either fix the issue or re-dispatch
  → Report the gap to user if unresolvable
```

---

## Verification Chain Protocol

> **Core principle**: AI cannot self-verify. Every role must verify upstream deliverables and produce a verification report.

### Before Starting Work (Rule v1)

```
Step 1: Identify upstream deliverables
  → Read previous node's VERIFY_REPORT if it exists
  → Read previous node's deliverable docs

Step 2: Verify upstream
  → Check completeness: are all required docs present and non-empty?
  → Check consistency: do docs reference each other correctly?
  → Check executability: can you actually use these as input?

Step 3: Decision
  → If upstream is OK → proceed with your work
  → If upstream has issues → report to principal (rule v3)
    → Include: what failed, why, suggested fix
    → Do NOT silently retry or skip
```

### After Completing Work (Rule v2)

```
Step 1: Produce verification report (VERIFY_REPORT doc)
  → What was checked
  → Result (pass/fail) for each check
  → Evidence (command output, file content, etc.)

Step 2: Self-check
  → All required docs exist and are non-empty
  → All hard rules were followed
  → Only specified tools were used

Step 3: Choose next edge
  → verified → proceed to next node
  → !verified → return to principal node
```

### When Verification Fails (Rule v3)

```
Step 1: Report to principal
  → What failed (specific check, specific deliverable)
  → Why it failed (root cause analysis)
  → Suggested fix (actionable recommendation)

Step 2: Wait for principal decision
  → Principal may: retry same node, adjust strategy, route to different node, or reject

Step 3: Loop Guard
  → Same task to same node ≤ 3 times
  → 2nd time: adjust strategy
  → 3rd time: MUST STOP and report to user
```

### QA Node: Independent Final Verification

The QA node (usually the last phase before terminal) does NOT trust self-check reports:

```
Step 1: Independently verify key deliverables
  → Run actual commands (go build, go test, flow proc validate, bun run build)
  → Do NOT rely on VERIFY_REPORTs from other roles

Step 2: Evaluate all layers
  → SKILL layer: SKILL Evaluation Protocol (see below)
  → Code layer: go build + go test + lint
  → Frontend layer: bun run build
  → Integration layer: end-to-end test

Step 3: Decision
  → All pass → gate.passed → terminal node
  → Any fail → !gate.passed → principal node
```

### SKILL Evaluation Protocol (skill-creator methodology)

When the QA node has `skill-creator` in its skills list, apply the skill-creator evaluation methodology adapted for team-flow:

```
Phase 1: SPEC Alignment Check
  → Read the SPEC.md from the architecture phase (fa03)
  → Read the SKILL.md produced by the implementation phase (fd04)
  → Verify: every SPEC requirement has a corresponding SKILL section
  → Verify: no SKILL sections exist without SPEC justification
  → Verify: flow JSON matches SPEC's described workflow

Phase 2: Flow JSON Quality Check
  → Run: flow proc validate {flow-path} — MUST pass with 0 errors
  → Walk through each node: flow proc run --flow {name} {node-id}
  → Verify: every node has a role, rules, and docs
  → Verify: every rule ref has a corresponding definition with name + description
  → Verify: every edge references valid node IDs
  → Verify: verified/!verified edges exist on all phase nodes
  → Verify: verification chain rules (v1/v2/v3) are present

Phase 3: Rule Completeness Check
  → For each rule in components.rules:
    - Has name? (warning if missing)
    - Has instruction? (error if missing — AI won't know what to do)
    - Has description? (warning if missing — AI can't get detail)
    - Has enforcement? (error if missing)
  → Verify: universal rules present (Output Guard, Phase Transition, No Hallucination, Loop Guard)
  → Verify: QG1-QG4 present
  → Verify: verification chain rules (v1/v2/v3) present

Phase 4: Test Prompt Evaluation (skill-creator core method)
  → Create 2-3 realistic test prompts for the SKILL
  → For each test prompt:
    - Run flow proc run with the prompt's scenario
    - Verify the engine produces valid, actionable instructions
    - Verify the AI can follow the instructions without ambiguity
  → Evaluate: are the instructions clear enough for an AI to execute?
  → Evaluate: are there gaps where the AI would be confused?

Phase 5: Doc Quality Check
  → For each doc in each node:
    - Has content_rules? Are they specific enough?
    - Has template? Is it appropriate for the domain?
    - Is the path correct (uses {DOCS_INTERNAL} not {docs_path})?
    - Is the description clear about what the doc should contain?

Phase 6: Produce QA_REPORT
  → Include all evaluation results from Phases 1-5
  → For each check: pass/fail + evidence
  → For failures: specific details + suggested fix
  → Overall assessment: ready for production / needs revision
```

---

## Gate Node Protocol

When `current.node_type == "gate"`:

```
Step 1: Evaluate each gate condition
  For each condition in current.gate_conditions:
    - tests_pass → Run the test suite. Pass threshold? (e.g., "100%")
    - lint_pass → Run linter. Error count below threshold? (e.g., "0 errors")
    - deliverables_complete → Check all required docs exist and non-empty
    - no_regressions → Existing features still work?
    - task_exists → Does the beads task exist?
    - type_matches → Does task type match expected value?
    - custom → Evaluate the check expression against expected value

Step 2: Determine overall result
  - If ALL required conditions pass → Gate PASSED
  - If ANY required condition fails → Gate FAILED
  - Advisory conditions are informational only

Step 3: Choose next node from next_options
  - Look for option with condition matching gate result
  - "gate.passed" or positive condition → pass path
  - "!gate.passed" or negative condition → fail path
  - If uncertain → prefer is_default option
  - If completely uncertain → pause and ask user

Step 4: Report gate result
  → Status line with gate result
  → List each condition: PASS/FAIL + evidence
  → Which path you're taking and why
```

---

## Branch/Parallel/Subflow/Loop/Manual/Event Nodes

### Branch Nodes

```
1. Evaluate conditions in order — first match wins
2. If no condition matches → use default goto
3. Run: flow proc run {chosen-node-id}
```

### Parallel Nodes

```
1. Launch multiple Task(general_purpose_task) calls concurrently
2. Each branch gets its own dispatch prompt (see Dispatch Protocol)
3. Wait for all branches to complete
4. Verify deliverables from each branch
5. Merge results per merge_strategy
6. Advance to the merge target node
```

### Subflow Nodes

```
1. The engine provides flow_ref — the name of the sub-flow
2. Run: flow proc run --flow {flow_ref}
3. Execute the sub-flow to completion
4. Return to the parent flow at the next node
```

### Loop Nodes

```
1. Execute the loop body (dispatch to sub-role)
2. Check exit_condition after each iteration
3. If exit_condition met → advance to next node
4. If max_attempts reached → advance to fail path
5. Never loop more than max_attempts times
```

### Manual Nodes

```
1. Pause execution
2. Present the approval request to the user
3. List the approvers and approval_strategy
4. Wait for user to approve/reject
5. Route accordingly
```

### Event Nodes

```
1. Wait for the external trigger
2. If on_timeout is specified and timeout occurs → handle timeout
3. When event arrives → advance to next node
```

---

## Terminal Node Protocol

When `current.is_terminal == true`:

```
Step 1: Read terminal status
  - "success" → Flow completed successfully
  - "failed" → Flow failed
  - "rejected" → Work was rejected
  - "cancelled" → Flow was cancelled
  - "timeout" → Flow timed out

Step 2: Execute Session Close Protocol (see above)

Step 3: Reset for next interaction
  - If user sends new input → Start from Session Startup Protocol
  - New input always starts at the principal node
  - Never skip the principal — all new input must be classified first
```

---

## Architectural Constraints

### AC1: One Project, One Flow

A project is bound to exactly **one** team flow.

- The flow binding is in project configuration (`default_flow` in project.md)
- `--flow` parameter is never needed — the engine reads the binding automatically
- If no flow is bound → user must create one first (via team-flow-v3-create)
- Switching flows is a project-level decision, not per-task

### AC2: Principal Is the Sole User Interface

The **principal** role is the only role that communicates with the user.

**Routing Rules**:

| Event | Route To |
|-------|----------|
| User sends new input | Principal's current node |
| User interrupts mid-execution | Principal's current node |
| User asks a question | Principal |
| Sub-role completes a task | Principal (via task return) |
| User provides feedback | Principal |
| Error in sub-role execution | Principal |

**How to route back to principal**:
1. Stop current execution
2. Run `flow proc run` on the principal's phase node (first phase node, or the node with `principal: true`)
3. Principal receives the user's input and decides next action

---

## Strict Enforcement Rules

### Rule 1: No Ad-Hoc Work

Everything must be defined in the flow. If the engine output doesn't specify a tool, don't use it. If it doesn't list a doc, don't produce it.

### Rule 2: Hard Rules Are Mandatory

Rules with `enforcement: "hard"` MUST be followed. No exceptions. If a rule says "run output guard before completion", you run it.

### Rule 3: Required Docs Must Be Produced

Every doc with `required: true` must exist at the specified path after node execution. Content must be meaningful — empty files don't count.

### Rule 4: Gates Block Progress

If a gate condition is not met, you MUST NOT proceed. Either fix the issue, route to the fail path, or report failure.

### Rule 5: Tool Restrictions Are Absolute

- Only use tools listed in the engine output
- If `commands` are specified, only use those commands
- If `scope` is specified, stay within that scope

### Rule 6: File Write Constraints

If `file_write` constraint with `allowed_paths` is specified, only write to paths matching those prefixes.

### Rule 7: Principal Must Dispatch

Principal NEVER executes work directly. Always dispatch via Task tool. "I'll just do it quickly" is overstepping, not efficiency.

---

## Error Recovery (Loop Guard)

| Error Type | Action | Max Retries |
|------------|--------|-------------|
| Recoverable (network/tool temporary failure) | Auto-retry same approach | 1 |
| Logic error (wrong approach, bad assumption) | Change approach, try differently | 2 |
| Persistent failure (3+ attempts all fail) | STOP — report to user, ask for guidance | 0 |

**Never loop infinitely.** After 3 total attempts, stop and escalate.

---

## Workspace Boundary

- **Project paths**: Read and write allowed
- **v3 flow files** (`v3/flows/`): Read only during execution (write only via create skill)
- **v3 skill files** (`team/v3/skills/`): Read only
- **v2 files** (`team/v2/`): Read only — never modify
- **CONSENSUS.md**: Read only during execution (write only via consensus update)
- **External project paths**: Read only for reference

---

## CLI Commands

```bash
flow proc run                           # Start/resume flow (auto-detect from project config)
flow proc run {node-id}                 # Run a specific node
flow proc run --flow {name}             # Run with explicit flow (rare, AC1 prefers auto-detect)
flow proc run --format json             # Get JSON output instead of text
flow proc rule {rule-id}                # Get full rule description
flow proc list                          # List available flows
flow proc show {flow-name}              # Show flow structure
flow proc validate {flow-name}          # Validate flow before execution
```

---

## Progress Reporting

After each node execution, report to the user:

1. **Node completed**: Which node (name + ID)
2. **Deliverables produced**: List of files created
3. **Gate results**: Pass/fail for each condition (if gate node)
4. **Next step**: Which node you're moving to and why
5. **Issues**: Anything unexpected or blocking

Keep reports concise — detailed content goes into deliverable files, not chat.

---

## Execution Checklist (Per Node)

Before moving to the next node, verify:

- [ ] All `on_enter` actions executed
- [ ] Role adopted correctly (persona, traits, guidance internalized)
- [ ] All hard-enforcement rules followed
- [ ] All required docs produced at correct paths with meaningful content
- [ ] All gate conditions evaluated (if gate node)
- [ ] Only specified tools used
- [ ] File write constraints respected
- [ ] Next node selected correctly from `next_options`
- [ ] Status line updated with current node info

---

## v2 Compatibility Mapping

For AI agents familiar with v2, here's how v3 maps to v2 concepts:

| v2 Concept | v3 Equivalent |
|-----------|---------------|
| Triage | Principal role (any role with `principal: true`) |
| `flow task create` | Automatic — flow tracks state via nodes |
| `flow task show --current` | `flow proc run` → `task.beads_id` |
| Task(subagent_type) | Task(general_purpose_task) for write work, Task(search) for read work |
| Skill routing (design/build/review) | `current.skills` — engine tells you which skill to load |
| Prompt files (triage.md, dev.md) | `current.persona` + `current.traits` + `current.guidance` |
| Shared workflows (shared.md) | `current.rules` — engine provides rules per node |
| Three-layer gates | Gate nodes with `gate_conditions` |
| Output Guard | Hard-enforcement rule in `current.rules` |
| Status Line `[Role\|TaskPool\|Phase\|Asset]` | `[Role: {alias}\|Flow: {name}#{beads-id}\|Node: {id}\|Phase: {phase}]` |
| Session End (git push) | Session Close Protocol |
