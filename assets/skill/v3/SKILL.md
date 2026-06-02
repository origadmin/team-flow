---
name: team-flow-v3
version: 3.2
description: |
  team-flow v3 / team-flow/v3: 流程驱动的AI协作框架 (Process-Centric AI Collaboration Framework).
  当用户提到以下任何关键词时必须触发此技能：team-flow, team-flow/v3, 团队流程, 流程管理,
  软件开发流程, 项目管理流程, v3流程, flow proc run, flow init.
  当 .team/version = v3 时，此技能是所有工作的强制入口 — 编码、调试、功能开发、Bug修复、
  重构、分析或任何任务。做任何工作之前必须先运行 flow proc run 进入流程驱动执行协议。
  不要直接开始解决问题。当以下情况触发：新会话开始、v3项目中收到任何用户请求、
  或看到 .team/version 包含 v3。此技能覆盖默认的问题解决行为。
---

# team-flow v3 SKILL.md - Entry Point

> **Version**: v3.2 | **Date**: 2026-05-28

## ⛔ MANDATORY: Read Consensus First

**Before doing ANYTHING in v3, read `.team/project.yaml` and `.team/constraints.md`.** Then run `flow proc run` — the engine provides everything else.

## Status Line (MANDATORY — Every Response)

Every response MUST start with:

```
[{alias} | {node_name}({node_id}:{flow}) | {ref} | {phase}]
```

- alias: from `flow proc run` output (dynamic per team)
- node_name(node_id:flow): current node name + node ID + flow name
- ref: task_id or `disc-{YYYYMMDD}-{seq}` or `-`
- Phase: current execution phase (on_enter/on_exit/analyze/design/implement/verify/review)

⛔ Status Line data source: `flow proc run` output, NEVER hardcode

## Core Concept

> **Flow First**: Everything starts with a flow definition. The flow specifies:
> 1. What steps to execute (nodes)
> 2. What order to execute them (edges)
> 3. What rules apply in each step (components)
> 4. What deliverables to produce (docs)
> 5. What gates enforce quality (gates)

## Session Startup Protocol (EVERY SESSION)

```
Step 1: flow project detect → Get workspace, project list, lock status
Step 2: flow proc run       → Get current node's work instructions
Step 3: Adopt principal role → Read alias, persona, traits from output
```

Only AFTER Step 3, start working on user requests.

## Binding: One Project, One Flow

- Binding stored in `.team/project.yaml` → `default_flow`
- If no flow bound → follow First-Time Setup (see `references/setup.md`)
- To change flow → user must explicitly switch (project-level decision)

## Principal: The Sole User Interface

- **Only the principal communicates with the user** — other roles are execution-only
- **User input always routes to the principal** — regardless of current node
- **Sub-role finishes → report to principal → principal decides next step**

## Project Configuration

| File | Purpose | AI reads? | ~Tokens |
|------|---------|-----------|---------|
| `.team/project.yaml` | Runtime config | ✅ Yes | ~100 |
| `.team/constraints.md` | Project constraints | ✅ Yes | ~50 |
| `.team/project.md` | Legacy fallback | ❌ No | — |

See `references/config.md` for project.yaml structure and document management.

## Two Core Skills

| Skill | Purpose |
|-------|---------|
| team-flow-v3-create | Create/modify/refine v3 flow definitions |
| team-flow-v3-exec | Execute a v3 flow, enforcing all rules and gates |

## On-Demand References

| Reference | When to read |
|-----------|-------------|
| `references/setup.md` | First-Time Setup (3 scenarios), team templates, placement rules |
| `references/config.md` | project.yaml structure, document management, session persistence |
| `references/structure.md` | Directory structure (embedded assets + project-level) |
| `references/v2-legacy.md` | v2 legacy directories warning, v2 fallback |

## Key Commands

```bash
flow project detect      # Project detection
flow proc run            # Start flow (GET WORK INSTRUCTIONS!)
flow proc run {node-id}  # Run specific node
flow task list           # Task list
flow task show <id>      # View task
flow task update <id> --claim  # Claim task
flow task close <id>     # Close task
flow config paths        # Show resolved path variables
```
