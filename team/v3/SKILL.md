---
name: team-flow-v3
version: 3.0
description: |
  team-flow v3: Process-Centric AI Collaboration Framework
  Process is central, with rules, tools, and skills unified under flow definitions.
  Creates self-documenting, executable AI workflows that can be verified, visualized, and refined.
---

# team-flow v3 SKILL.md - Entry Point

> **Version**: v3.0 | **Date**: 2026-05-17

## Core Concept

> **Flow First**: Everything starts with a flow definition. The flow specifies:
> 1. What steps to execute (nodes)
> 2. What order to execute them (edges)
> 3. What rules apply in each step (components)
> 4. What deliverables to produce (docs)
> 5. What gates enforce quality (gates)

## Flow Loading

v3 flows are stored in `{PROJECT}/v3/flows/` directory. Each flow is a JSON file following the v3 schema.

When a user asks for a task, v3:
1. Finds matching flow from `v3/flows/` (or creates one if missing)
2. Validates flow with `flow proc validate`
3. Executes the flow using `flow proc run`

## Two Core Skills

| Skill | Purpose |
|-------|---------|
| team-flow-v3-create | Create/modify/refine v3 flow definitions |
| team-flow-v3-exec | Execute a v3 flow, enforcing all rules and gates |

## Directory Structure

```
team/v3/
├── SKILL.md              ← You are here
├── BOUNDARY.md           # Layer architecture for v3
├── skills/
│   ├── team-flow-v3-create/
│   │   └── SKILL.md      # Flow creation/refinement skill
│   └── team-flow-v3-exec/
│       └── SKILL.md      # Flow execution skill
├── prompts/              # Triage and classification prompts
├── workflows/            # Shared v3 workflows
└── references/           # Reference materials
```
