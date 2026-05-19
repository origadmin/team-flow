---
name: team-flow-v3-create
version: 3.1
description: |
  Create and refine v3 flow JSON definitions.
  Produces valid flow JSON that passes `flow proc validate` and is executable by the engine.
---

# team-flow-v3-create - Flow Creation Skill

## Purpose

Generate or modify v3 flow JSON files. The output must be a valid flow JSON that:
1. Passes `flow proc validate` with zero errors
2. Can be executed by `flow proc run` to produce structured AI instructions

## Critical Rules

### Rule 1: Node IDs are Random Shortcodes

Node IDs MUST be opaque random shortcodes. No semantic meaning, no prefixes, no patterns.

**Format**: `^[a-z][a-z0-9]{3,4}$` (4-5 lowercase alphanumeric chars, starting with a letter)

**Valid examples**: `tri3`, `fa01`, `a7x2`, `k9m`, `bp4w`
**Invalid examples**: `phase-triage`, `gate-entry`, `gc01`, `df-tri3`, `feature-start`

**Why**: Semantic IDs cause duplication across flows and distort intent. Meaning belongs in `name` and `description` fields, not in the ID.

**Generation algorithm**:
1. Generate random 4-5 char string matching the pattern
2. Check uniqueness within the flow (IDs only need to be unique within a single flow)
3. If collision, regenerate

### Rule 2: Edge IDs Follow Node References

**Format**: `^e-{from_node_id}-{to_node_id}$`

Example: `e-tri3-ent7`, `e-fa01-fd02`, `e-qua9-ver4`

### Rule 3: Always Validate Before Saving

Run `flow proc validate {path}` after any change. Zero errors required. Warnings should be resolved when possible.

### Rule 4: Don't Touch v2

Only work with `v3/flows/` and `team/v3/`. Never modify v1 or v2 files.

## File Locations

| Item | Path |
|------|------|
| Preset flows | `v3/flows/{name}.json` |
| Project flows | `.team/flows/{name}.json` |
| Flow schema | `v3/schema/flow-schema.json` |
| Reference flow | `v3/flows/dev-flow.json` (canonical example with correct IDs) |

**Resolution order**: `.team/flows/` → `v3/flows/` (project overrides preset)

## CLI Commands (flow proc)

```bash
# List available flows
flow proc list

# Show flow structure
flow proc show {flow-name}

# Validate a flow file (MUST pass with zero errors)
flow proc validate {path-or-name}

# Run a node (get AI instruction)
flow proc run [--flow {name}] [node-id] [--task {task-id}] [--format json|text]

# Create new flow from template
flow proc create {name} --template {template-name}
```

## Flow JSON Structure

### Top-Level Fields

```json
{
  "version": "v3",
  "metadata": { "name": "my-flow", "description": "...", "tags": [...] },
  "config": { "task_type": "feature", "auto_dispatch": true, "parallel_limit": 3, "timeout_minutes": 480 },
  "variables": { "docs_path": "", "TEAM_PATH": "", "task_id": "", "SKILL_PATH": "", "PROJECT_PATH": "" },
  "components": { "roles": [...], "rules": [...], "tools": [...], "skills": [...], "gates": [...] },
  "nodes": [...],
  "edges": [...]
}
```

### Required Fields

- `version`: always `"v3"`
- `metadata.name`: lowercase kebab-case identifier (`^[a-z][a-z0-9-]*$`)
- `nodes`: at least 1 node
- `edges`: connections between nodes
- At least one node must be `type: "terminal"`

### Variable Placeholders

Paths in `docs[].path` and `components[].path` support these variables:

| Variable | Description |
|----------|-------------|
| `{docs_path}` | Project documentation root |
| `{TEAM_PATH}` | Path to `.team` directory |
| `{task_id}` | Current task identifier |
| `{SKILL_PATH}` | Path to team-flow skill directory |
| `{PROJECT_PATH}` | Project root directory |

## Node Types Reference

### phase (execution stage)

```json
{
  "id": "fa01",
  "type": "phase",
  "name": "Requirements Analysis",
  "description": "Analyze requirements and produce specification",
  "config": {
    "role": "TechLead",
    "auto_dispatch": true,
    "timeout_minutes": 120,
    "deliverables": ["SPEC.md", "AC.md"]
  },
  "components": {
    "roles": [{ "ref": "tech-lead", "source": "builtin" }],
    "rules": [
      { "ref": "requirements-standards", "source": "builtin" },
      { "ref": "output-guard", "source": "framework" }
    ],
    "tools": [
      { "ref": "task", "commands": ["show", "update", "append"] },
      { "ref": "search" }
    ],
    "skills": [{ "ref": "api-design", "source": "builtin" }],
    "constraints": [
      { "type": "file_write", "allowed_paths": ["{docs_path}/{task_id}/"] }
    ]
  },
  "docs": [
    {
      "name": "SPEC.md",
      "format": "markdown",
      "path": "{docs_path}/{task_id}/SPEC.md",
      "required": true,
      "description": "Feature specification: requirements, scope, constraints"
    }
  ],
  "on_enter": [{ "action": "update_task_phase", "phase": "analyze" }]
}
```

### gate (conditional checkpoint)

```json
{
  "id": "ent7",
  "type": "gate",
  "name": "Entry Gate",
  "description": "Validate task has sufficient context",
  "config": {
    "conditions": [
      { "type": "task_exists", "required": true, "check": "Verify a task exists" },
      { "type": "type_matches", "expected": "feature" }
    ],
    "on_pass": "fa01",
    "on_fail": "rej8",
    "auto_retry": { "max_attempts": 2, "delay_minutes": 5 }
  }
}
```

**Gate condition types**: `task_exists`, `type_matches`, `deliverables_complete`, `tests_pass`, `lint_pass`, `no_regressions`, `custom`

**Note**: `on_pass`/`on_fail` reference node IDs. When using random shortcodes, these must match actual node IDs in the flow.

### branch (conditional routing)

```json
{
  "id": "br2k",
  "type": "branch",
  "name": "Task Type Router",
  "description": "Route by task type",
  "config": {
    "conditions": [
      { "when": "task.type == 'feature'", "goto": "fa01" },
      { "when": "task.type == 'bug'", "goto": "bi01" },
      { "default": "goto", "target": "ai01" }
    ]
  }
}
```

### parallel (concurrent execution)

```json
{
  "id": "pm3w",
  "type": "parallel",
  "name": "Parallel Verification",
  "description": "Run verifications concurrently",
  "config": {
    "max_concurrency": 2,
    "strategy": "all_success",
    "branches": [
      { "node": "vr1a", "weight": 1 },
      { "node": "vr2b", "weight": 1 }
    ],
    "merge_strategy": "wait_all"
  }
}
```

### subflow (nested flow invocation)

```json
{
  "id": "sf4n",
  "type": "subflow",
  "name": "Run Bugfix Flow",
  "description": "Invoke bugfix sub-flow",
  "config": {
    "flow_ref": "builtin:bugfix-flow",
    "input_mapping": { "task_id": "{task_id}" },
    "output_mapping": { "fix_result": "impl_result" }
  }
}
```

**flow_ref format**: `namespace:name` (e.g. `builtin:bugfix-flow`, `custom:my-flow`)

### loop (retry/iteration)

```json
{
  "id": "lp5m",
  "type": "loop",
  "name": "Retry Gate",
  "description": "Retry until quality gate passes",
  "config": {
    "max_attempts": 3,
    "delay_minutes": 5,
    "backoff": "exponential",
    "exit_condition": "gate.passed == true"
  }
}
```

### manual (human approval)

```json
{
  "id": "ap6j",
  "type": "manual",
  "name": "Design Approval",
  "description": "Require TechLead approval before implementation",
  "config": {
    "approvers": [{ "role": "TechLead", "required": true }],
    "approval_strategy": "all",
    "timeout_minutes": 60,
    "on_timeout": "reject"
  }
}
```

### event (external trigger)

```json
{
  "id": "ev7h",
  "type": "event",
  "name": "Wait for CI",
  "description": "Wait for CI pipeline completion",
  "config": {
    "trigger": { "type": "webhook", "pattern": "/ci/callback" },
    "timeout_minutes": 30,
    "on_timeout": "fail"
  }
}
```

### terminal (end state)

```json
{
  "id": "suc0",
  "type": "terminal",
  "name": "Completed",
  "description": "Task completed successfully",
  "config": { "status": "success", "message": "Task completed successfully" }
}
```

**Terminal statuses**: `success`, `failed`, `rejected`, `cancelled`, `timeout`

## Component References

### ComponentRef (roles, rules, skills)

```json
{ "ref": "output-guard", "source": "framework" }
```

**Source enum**: `builtin`, `custom`, `marketplace`, `trae`, `file`, `framework`

### ToolRef

```json
{ "ref": "task", "source": "builtin", "commands": ["show", "update", "append"], "scope": "codebase" }
```

**Scope enum**: `codebase`, `docs`, `config`, `all`

### Constraints

```json
{ "type": "file_write", "allowed_paths": ["{docs_path}/{task_id}/"] }
{ "type": "tool_restriction", "forbidden": ["shell-exec"] }
{ "type": "timeout", "minutes": 60 }
{ "type": "max_sub_agents", "limit": 3 }
```

## Flow-Level Components (Definition Registry)

The `components` field at flow level serves as a definition registry. When the engine processes a node's `components.rules[]`, it looks up each `ref` in the flow-level `components.rules[]` to enrich the output with `name`, `description`, and `enforcement`.

**If a rule is referenced in a node but not defined at flow level, the engine still outputs it but with empty name/description — the validator will warn about this.**

```json
{
  "components": {
    "rules": [
      {
        "id": "output-guard",
        "name": "Output Guard Protocol",
        "description": "5-step self-assessment before completion gate",
        "source": "framework",
        "type": "process_constraint",
        "enforcement": "hard"
      },
      {
        "id": "regression-guard",
        "name": "Regression Prevention",
        "description": "Prevent regressions in existing functionality",
        "source": "builtin",
        "type": "quality_constraint",
        "enforcement": "hard"
      }
    ],
    "roles": [
      {
        "id": "triage",
        "name": "Triage Agent",
        "capabilities": ["classify", "dispatch", "assess"]
      }
    ]
  }
}
```

**Rule types**: `behavioral_constraint`, `output_constraint`, `process_constraint`, `quality_constraint`
**Enforcement levels**: `hard` (must pass), `soft` (best effort)

## In-Node Gates vs Gate Nodes

There are two ways to define gates:

1. **Gate node** (`type: "gate"`): A dedicated checkpoint node with `on_pass`/`on_fail` routing. Use for decision points in the flow graph.

2. **In-node gates** (`gates[]` on any node): Pre-completion checks. The node must pass these before it can transition to the next node. Use for deliverable validation within a phase.

```json
{
  "id": "fa01",
  "type": "phase",
  "name": "Requirements Analysis",
  "gates": [
    {
      "type": "deliverable_check",
      "required": ["SPEC.md", "AC.md"],
      "on_fail": "fa01"
    }
  ]
}
```

**Gate config types**: `deliverable_check`, `tests_pass`, `lint_pass`, `no_regressions`, `quality_check`, `custom`

## Edge Configuration

### Sequential (default)

```json
{ "id": "e-fa01-fd02", "from": "fa01", "to": "fd02" }
```

### Conditional

```json
{
  "id": "e-ent7-fa01",
  "from": "ent7",
  "to": "fa01",
  "type": "conditional",
  "conditions": [{ "expression": "task_type=feature" }]
}
```

**Common expressions**:
- `gate.passed` / `!gate.passed` — gate result
- `task_type=feature` — task type matching
- `task_type=feature AND gate.passed` — combined

## Engine Runtime Behavior

When `flow proc run` executes, the engine:

1. **LoadFlow** — reads `v3/flows/{name}.json` or `.team/flows/{name}.json`
2. **FindNode** — finds node by ID (first phase/start node if no ID specified)
3. **CollectVars** — assembles variable map from project config + flow metadata
4. **BuildCurrentNode** — produces structured output:
   - For `phase`/`start`: role, rules (enriched from flow-level definitions), tools, skills, docs (with vars substituted), on_enter actions
   - For `gate`: gate conditions only
   - For `terminal`: status + message
5. **BuildNextOptions** — lists reachable next nodes with roles and conditions

The output is a JSON object that AI agents consume directly. **AI agents never read markdown docs — they get structured JSON from the engine.**

## Creation Workflow

### Step 1: Understand the Requirement

Ask the user:
- What kind of workflow? (software dev, content creation, game design, etc.)
- What phases/steps are needed?
- What roles are involved?
- What quality gates are required?
- What deliverables at each step?

### Step 2: Check Existing Flows

```bash
flow proc list
```

If a similar flow exists, use it as a template:
```bash
flow proc create {new-name} --template {existing-flow}
```

### Step 3: Design the Flow

Plan the node graph:
1. List all phases with their roles
2. Identify gate/decision points
3. Determine edge connections and conditions
4. Assign components (rules, tools, skills) to each node
5. Define deliverables (docs) for each phase
6. Generate random shortcode IDs for all nodes

### Step 4: Write the Flow JSON

Follow the structure and rules above. Key checklist:
- [ ] All node IDs are random shortcodes (`^[a-z][a-z0-9]{3,4}$`)
- [ ] All edge IDs follow `e-{from}-{to}` pattern
- [ ] `gate.config.on_pass`/`on_fail` reference valid node IDs
- [ ] `branch.conditions[].goto`/`target` reference valid node IDs
- [ ] `parallel.branches[].node` references valid node IDs
- [ ] At least one terminal node exists
- [ ] All edges reference existing nodes
- [ ] Variable placeholders use correct names (`{docs_path}`, `{task_id}`, etc.)
- [ ] All rule refs in nodes have corresponding definitions in flow-level `components.rules`
- [ ] All role refs in nodes have corresponding definitions in flow-level `components.roles`
- [ ] `config.role` in phase nodes uses a value from the Role enum (case-sensitive)

**Standard roles**: `Triage`, `TechLead`, `Dev`, `QA`, `PM`, `DevOps`, `Analysis`, `UIDesigner`

**Custom roles**: If a role is not in the standard enum, define it in flow-level `components.roles[]` and use it in phase `config.role`. The engine reads role as a string from PhaseConfig.

### Step 5: Validate

```bash
flow proc validate v3/flows/{name}.json
```

Fix all errors. Address warnings (especially undefined rule/role refs).

### Step 6: Test Run

```bash
flow proc run --flow {name}
```

Verify the engine produces valid structured output for the first node.

### Step 7: Iterate

Walk through each node:
```bash
flow proc run --flow {name} {node-id}
```

Ensure each node produces correct instructions with proper rules, tools, docs, and next options.

## Common Patterns

### Linear Flow (simple pipeline)

```
Phase A → Phase B → Phase C → Terminal
```

### Branching Flow (by task type)

```
Triage → Gate → [Feature path | Bug path | Hotfix path] → Quality Gate → Verify → Terminal
```

Use `gate` node with conditional edges, or `branch` node.

### Parallel Verification

```
Implementation → Parallel[Lint, Unit Tests, Integration] → Review → Terminal
```

### Retry Loop

```
Implementation → Quality Gate → (fail) → Loop back to Implementation
```

Use gate `on_fail` pointing back, or a `loop` node.

### Subflow Delegation

```
Main Flow → Subflow[bugfix-flow] → Continue
```

Use `subflow` node with `flow_ref: "builtin:bugfix-flow"`.

## Anti-Patterns to Avoid

1. **Semantic IDs**: Never use `phase-triage`, `gate-entry`, `impl-node`. Use random shortcodes.
2. **Giant mega-flows**: Don't put all task types in one flow. Use separate flows with subflow nodes.
3. **Inline rule content**: Don't put rule text in node descriptions. Define rules at flow level and reference by `ref`.
4. **Missing terminal nodes**: Every path through the flow must eventually reach a terminal.
5. **Orphan nodes**: Every non-terminal node must be connected by at least one edge.
6. **Undefined refs**: Every `ref` in node components must have a corresponding definition in flow-level `components`.
