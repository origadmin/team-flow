# team-flow v3 - Boundary Definition

## Architecture Layers

> **Process First**: The flow definition is the single source of truth for AI behavior.

```
┌─────────────────────────────────────────────────────────┐
│                   Editor (Visualization)                │
│     - Load flow from v3/flows/                          │
│     - Visualize nodes/edges/components                  │
│     - Allow human refinement                            │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│              team-flow-v3-create Skill                  │
│     - Generate flows from user input                    │
│     - Validate flows with flow proc validate            │
│     - Store in v3/flows/                                │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                 Flow Engine (Go)                        │
│     - flow proc validate - Verify flow correctness      │
│     - flow proc run     - Execute the flow              │
│     - Enforce gates, constraints, rules                │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│              team-flow-v3-exec Skill                    │
│     - Follows flow steps strictly                       │
│     - Uses flow-specified roles/rules/tools            │
│     - Produces expected deliverables                    │
└─────────────────────────────────────────────────────────┘
```

## Hard Boundaries

1. **v3 ≠ v2**: v3 is independent, don't modify anything under `team/v2/`
2. **Flow is King**: All AI behavior must be specified in a flow, no ad-hoc work
3. **Verification is Mandatory**: A flow must pass `flow proc validate` before execution
4. **Gates are Enforced**: If a gate fails, flow stops - no bypass
