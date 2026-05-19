---
name: team-flow-v3-exec
version: 3.1
description: |
  Execute v3 flows by running `flow proc run` and following the engine's structured output.
  Strictly enforces all rules, gates, and deliverables.
---

# team-flow-v3-exec - Flow Execution Skill

## Purpose

Execute a validated v3 flow using the engine CLI. The engine produces structured JSON instructions at each node — follow them exactly.

## Core Principle

**The engine is the source of truth.** You do not interpret the flow JSON directly. You run `flow proc run` and follow the structured output.

## Execution Loop

```
1. flow proc run --flow {name} --task {task-id}
   → Receive structured JSON instruction for the current node

2. Read the instruction:
   - role: adopt this role
   - rules: follow these constraints
   - tools: use only these tools
   - docs: produce these deliverables
   - on_enter: execute these actions first

3. Execute the node:
   - Follow all rules (hard enforcement = must pass, soft = best effort)
   - Use only specified tools
   - Produce all required docs
   - Pass all gate conditions before moving on

4. Choose next node from next_options:
   - If only one option → proceed
   - If conditional → evaluate condition and choose
   - If gate → determine pass/fail and route accordingly

5. flow proc run --flow {name} {next-node-id} --task {task-id}
   → Repeat until terminal node reached

6. Terminal reached → Flow complete
```

## CLI Commands

```bash
# Start or resume flow execution
flow proc run --flow {flow-name} [--task {task-id}]

# Run a specific node
flow proc run --flow {flow-name} {node-id} [--task {task-id}]

# Check available flows
flow proc list

# Inspect a flow's structure
flow proc show {flow-name}

# Validate before execution
flow proc validate {flow-name}
```

## Engine Output Structure

When you run `flow proc run`, the engine returns:

```json
{
  "flow": {
    "name": "feature-flow",
    "version": "v3",
    "domain": "feature"
  },
  "current": {
    "node_id": "fa01",
    "node_type": "phase",
    "name": "Requirements Analysis",
    "description": "Analyze requirements and produce specification",
    "role": "TechLead",
    "rules": [
      {
        "ref": "output-guard",
        "source": "framework",
        "name": "Output Guard Protocol",
        "description": "5-step self-assessment before completion",
        "enforcement": "hard"
      }
    ],
    "tools": [
      { "ref": "task", "source": "builtin", "commands": ["show", "update", "append"] },
      { "ref": "search", "source": "builtin" }
    ],
    "skills": [
      { "ref": "api-design", "source": "builtin" }
    ],
    "docs": [
      {
        "name": "SPEC.md",
        "path": "/path/to/docs/TASK-001/SPEC.md",
        "format": "markdown",
        "required": true,
        "description": "Feature specification"
      }
    ],
    "on_enter": [
      { "action": "update_task_phase", "phase": "analyze" }
    ],
    "gate_conditions": null,
    "is_terminal": false
  },
  "next_options": [
    { "node_id": "fd02", "role": "TechLead", "condition": null, "is_default": true }
  ]
}
```

## Node Type Handling

### Phase Nodes

1. Execute `on_enter` actions (e.g., update task phase)
2. Adopt the assigned `role`
3. Load and follow all `rules` — each rule has name, description, and enforcement level
4. Use only the specified `tools` (with restricted commands if specified)
5. Produce all `docs` listed — each has name, path, required flag, and description
6. Respect `constraints` (file_write paths, tool restrictions, timeouts)
7. If `gates[]` are defined on this node, pass them before completing
8. Report completion and move to next node

### Gate Nodes

1. Evaluate `gate_conditions` — each has type, threshold, required flag
2. Determine pass/fail:
   - `tests_pass`: Are all tests passing? Check threshold (e.g., "100%")
   - `lint_pass`: Is lint clean? Check threshold (e.g., "0 errors")
   - `deliverables_complete`: Are all required files present and non-empty?
   - `no_regressions`: Do existing features still work?
   - `task_exists`: Does the task exist?
   - `type_matches`: Does task type match expected value?
3. Choose from `next_options`:
   - Condition `gate.passed` → pass path
   - Condition `!gate.passed` → fail path
4. If `auto_retry` is configured, you may re-attempt before routing to fail path

### Terminal Nodes

- `is_terminal: true` indicates flow completion
- `terminal_status`: `success` | `failed` | `rejected` | `cancelled` | `timeout`
- Report the status and stop execution

### Branch/Parallel/Subflow/Loop/Manual/Event

- **Branch**: Evaluate conditions in order, first match wins. `default: goto` is the fallback.
- **Parallel**: Execute all branches concurrently. Wait per `merge_strategy`.
- **Subflow**: Invoke another flow via `flow_ref`. Pass inputs, receive outputs.
- **Loop**: Repeat with `max_attempts`. Check `exit_condition` each iteration.
- **Manual**: Wait for human approval per `approvers` and `approval_strategy`.
- **Event**: Wait for external trigger. Handle `on_timeout` if specified.

## Strict Enforcement Rules

### Rule 1: No Ad-Hoc Work

Everything must be defined in the flow. If the engine output doesn't specify a tool, don't use it. If it doesn't list a doc, don't produce it.

### Rule 2: All Rules Are Mandatory (hard enforcement)

Rules with `enforcement: "hard"` MUST be followed. If a rule says "run output guard before completion", you run it. No exceptions.

### Rule 3: All Required Docs Must Be Produced

Every doc with `required: true` must exist at the specified path after node execution. Content must be meaningful — empty files don't count.

### Rule 4: Gates Block Progress

If a gate condition is not met, you MUST NOT proceed to the next node. Either:
- Fix the issue and re-evaluate
- Route to the fail path if one exists
- Report failure

### Rule 5: Tool Restrictions Are Absolute

- Only use tools listed in the engine output
- If `commands` are specified, only use those commands
- If `scope` is specified, stay within that scope
- If `constraints` include `tool_restriction`, forbidden tools are absolutely forbidden

### Rule 6: File Write Constraints

If `file_write` constraint with `allowed_paths` is specified, only write to paths matching those prefixes. Variable placeholders in paths are already resolved by the engine.

## Progress Reporting

After each node execution, report:

1. **Node completed**: Which node (name + ID)
2. **Docs produced**: List of files created
3. **Gate results**: Pass/fail for each condition
4. **Next step**: Which node you're moving to and why
5. **Issues**: Anything unexpected or blocking

## Error Handling

If a node has `on_error` configuration:
- `retry`: Re-attempt execution up to `max_retries` times
- `skip`: Move to next node (only if non-critical)
- `fail`: Terminate flow at the failed terminal
- `fallback`: Route to `fallback_node`

If no `on_error` is configured, the default is `fail`.

## Execution Checklist (Per Node)

Before moving to the next node, verify:

- [ ] All `on_enter` actions executed
- [ ] Role adopted correctly
- [ ] All hard-enforcement rules followed
- [ ] All required docs produced at correct paths with meaningful content
- [ ] All gate conditions evaluated and passed
- [ ] Only specified tools used
- [ ] File write constraints respected
- [ ] Next node selected correctly from `next_options`
