# Node Types Reference

> Referenced by: team-flow-v3-create SKILL.md
> Load this file when you need to write node definitions in flow JSON.

## phase (execution stage)

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

## gate (conditional checkpoint)

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

## branch (conditional routing)

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

## parallel (concurrent execution)

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

## subflow (nested flow invocation)

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

## loop (retry/iteration)

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

## manual (human approval)

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

## event (external trigger)

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

## terminal (end state)

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
