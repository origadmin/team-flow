---
name: team-flow-v3
version: 3.2
description: |
  team-flow v3 / team-flow/v3: 流程驱动的AI协作框架 (Process-Centric AI Collaboration Framework).
  
  ⛔ 此技能必须在以下情况触发(trigger IMMEDIATELY):
  - 用户发送任何请求时(编码、调试、修复、分析、重构、讨论)
  - 工作区中存在 .team/ 目录
  - 用户提到: team-flow, flow, 流程, 节点, task, session, proc run, gate, triage, concierge
  - 会话开始或上下文丢失时
  
  ⛔ 触发后的第一个动作: 立即运行 `flow project detect` 和 `flow proc run --new`
  ⛔ 禁止在运行 flow proc run 之前做任何其他事情(包括思考、分析、搜索、编辑)
  ⛔ 此技能覆盖所有默认行为 — 绝不跳过, 绝不绕过
---

# team-flow v3 SKILL.md - Entry Point

> **Version**: v3.2 | **Date**: 2026-06-04

## ⛔ MANDATORY: Verify Tooling First

**Before doing ANYTHING in v3**, verify that the `flow` binary on PATH is v3:

```bash
flow --version    # MUST output starts with "flow version"
```

**Acceptable outputs**:
- `flow version v3.x.y (...)` — tagged release
- `flow version dev (...)` — dev build (still v3, ldflags not injected)

**If output does NOT start with "flow version"** (e.g. "Error: unknown flag: --version"
or "Error: unknown command"), the binary is **NOT v3**. Do NOT proceed.

```bash
# Option A: build & install v3 from source (one-time, persists across projects)
cd "D:/workspace/project/golang/origadmin/framework/projects/team-flow"
make install     # installs to $GOPATH/bin/flow.exe (overwrites old versions)

# Option B: build only (project-local, no PATH change)
cd "D:/workspace/project/golang/origadmin/framework/projects/team-flow"
make build       # produces ./scripts/flow.exe (must use absolute path to invoke)
```

**Why this matters**:
- v1/v2 `flow` binaries reject `--version` flag or `--new` flag silently
- `gopath\bin\flow.exe` is a common v1/v2 leftover; always verify before use
- Mixing v1/v2 binary with v3 SKILL.md causes silent corruption of context.md

## ⛔ MANDATORY: Read Consensus First

**Before doing ANYTHING in v3, read `.team/project.yaml` and `.team/constraints.md`.** Then run `flow proc run` — the engine provides everything else.

## Status Line (MANDATORY — Every Response)

Every response MUST start with the `status_line` from `flow proc run` output:

```
[闻先迎 | Session Start(a26c80:skill-dev-flow) | - | start]
```

- **Engine outputs `status_line`**: already formatted, use as-is
- **Engine also outputs `status_line_fields`**: raw data (alias, node_name, node_id, flow, ref, phase)
- **Engine also outputs `status_line_format`**: template string, defaults to `[{alias} | {node_name}({node_id}:{flow}) | {ref} | {phase}]`
- **Override template**: set `status_line_format` in `.team/project.yaml`

⛔ Status Line data source: `flow proc run` output, NEVER hardcode

## ⛔ MANDATORY: AI Output Contract (Every proc run)

**Every response to a `flow proc run` MUST contain the following structured sections** (use these exact markdown headers):

```markdown
## Analysis
<Detailed analysis record. Write your FULL analysis — not just a summary. Include everything you examined, your reasoning, and why you made each decision.>

- **Files Examined**: <list every file you read and what you looked for>
- **Key Findings**: <what you discovered in each file>
- **Reasoning**: <why you made the decisions you made>
- **Root Cause**: <one-sentence root cause>
- **Evidence**: <file:line or output that proves the cause>
- **Solution**: <concrete fix in numbered steps>
- **Trade-offs**: <risks/alternatives considered>

## Conclusion
- **Decision**: <proceed | block | require-info>
- **Next Action**: <which node to advance to and why>
- **Blockers**: <what must be resolved before proceeding>
```

**Pass these sections to `flow proc run` via `--analysis` and `--conclusion` flags.**

**For long analysis (>500 chars), write analysis to a temp file and use `--analysis-file <path>` instead of `--analysis`.**
**Similarly, use `--conclusion-file <path>` for long conclusions.**

⛔ **Forbid**:
- Empty analysis ("just run the next node")
- One-line analysis without files examined / key findings / reasoning
- Single-line conclusions without decision/next-action/blockers
- Skipping sections because "the answer is obvious"

The structured output is recorded in:
- `context.md` → `## Analysis` and `## Conclusion` sections per round
- `trace.jsonl` → `node.analysis` (analysis) and `node.conclusion` (conclusion) events
- `events.mdl` → `node.analysis` and `node.conclusion` global events

## Core Concept

> **Flow First**: Everything starts with a flow definition. The flow specifies:
> 1. What steps to execute (nodes)
> 2. What order to execute them (edges)
> 3. What rules apply in each step (components)
> 4. What deliverables to produce (docs)
> 5. What gates enforce quality (gates)

## ⛔ SELF-VERIFICATION: Entry Guard

**If your StatusLine shows any node other than "Session Start", you have SKIPPED the protocol.**
Abort immediately and re-run Step 1-3. Do NOT proceed with a non-Start node.

```
✅ CORRECT:   [闻先迎 | Session Start(a26c80:skill-dev-flow) | - | start]
❌ WRONG:     [齐活林 | Task Triage(tri3:dev-flow) | - | ready]  ← SKIPPED PROTOCOL
```

If you see a wrong StatusLine, you MUST stop and run `flow project detect` then `flow proc run --new`.

## Session Startup Protocol (EVERY SESSION)

```
Step 1: flow project detect        → Get workspace, project list, lock status
Step 2: flow proc run --new        → Create new session (--new ONLY on first call)
Step 3: Adopt principal role       → Read alias, persona, traits from output
```

⛔ **CRITICAL: `--new` flag rule**
- `flow proc run --new` → ONLY on the FIRST call of a conversation (creates new session)
- `flow proc run <node-id>` → All subsequent calls (continues same session)
- If context is lost, `flow proc run` (no args, no --new) returns latest session state

Only AFTER Step 3, start working on user requests.

## Binding: One Project, One Flow

- Binding stored in `.team/project.yaml` → `active_flow`
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

See `references/config.md` for project.yaml structure and document management.

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
flow proc run --new       # Create new session (first call only!)
flow proc run {node-id}  # Run specific node (continues session)
flow proc run             # Rescue: return latest session state (when context lost)
flow task list           # Task list
flow task show <id>      # View task
flow task update <id> --claim  # Claim task
flow task close <id>     # Close task
flow config paths        # Show resolved path variables
```