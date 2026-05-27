# Component References

> Referenced by: team-flow-v3-create SKILL.md
> Load this file when you need to write component definitions in flow JSON.

## ComponentRef (roles, rules, skills)

```json
{ "ref": "output-guard", "source": "framework" }
```

**Source enum**: `builtin`, `custom`, `marketplace`, `trae`, `file`, `framework`

## ToolRef

```json
{ "ref": "task", "source": "builtin", "commands": ["show", "update", "append"], "scope": "codebase" }
```

**Scope enum**: `codebase`, `docs`, `config`, `all`

## Constraints

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
