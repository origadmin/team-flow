# BOUNDARY.md - Framework Layer Boundary Rules

> **This file MUST be read before any other team-flow/ file.**
> All AI agents (Trae, Gemini CLI, OpenClaw, Cursor, etc.) MUST comply.

---

## LAYER ARCHITECTURE

`
framework/                    <-- MULTI-PROJECT MANAGEMENT PLATFORM
  {TEAM_PATH}/                    <-- FRAMEWORK LAYER (shared rules, read-only for AI)
    SKILL.md                  <-- framework entry point
    prompts/                  <-- role execution rules (shared)
    workflows/                <-- shared workflows
    templates/                <-- document templates
    config/                   <-- framework config
    examples/                 <-- tool configuration examples
    BOUNDARY.md               <-- THIS FILE

  _docs/                      <-- FRAMEWORK LAYER (cross-project docs)
    {project-name}/           <-- per-project internal docs

  projects/
    {project-name}/           <-- PROJECT LAYER (per-project workspace)
      .team/                  <-- PROJECT OPERATIONAL FILES (AI reads/writes here)
        project.md            <-- project config (paths, toolchain, constraints)
        task-pool.md          <-- task status (Triage maintains)
        backlog.md            <-- deferred tasks
      docs/                   <-- project external docs
      src/                    <-- project source code
`

## HARD RULES

### 1. {TEAM_PATH}/ is READ-ONLY for AI execution

| Operation | {TEAM_PATH}/ (Framework) | .team/ (Project) |
|-----------|:------------------:|:----------------:|
| Read rules/prompts/workflows | YES | N/A |
| Write task-pool.md | **NEVER** | YES |
| Write backlog.md | **NEVER** | YES |
| Write project.md | **NEVER** | YES |
| Write any new file | **NEVER** | YES (if project operational) |
| Modify existing files | **NEVER** | YES |

### 2. File placement rules

| File | Correct Location | WRONG Location |
|------|-----------------|----------------|
| task-pool.md | {PROJECT_PATH}/.team/task-pool.md | {TEAM_PATH}/task-pool.md |
| backlog.md | {PROJECT_PATH}/.team/backlog.md | {TEAM_PATH}/backlog.md |
| project.md | {PROJECT_PATH}/.team/project.md | {TEAM_PATH}/project.md |
| SPEC.md | {docs_internal}/requirements/... | {TEAM_PATH}/... |
| AC.md | {docs_internal}/requirements/... | {TEAM_PATH}/... |
| RCA.md | {docs_internal}/reports/bugs/... | {TEAM_PATH}/... |

### 3. Path variable usage

When you see {TEAM_PATH}, it means:
- **Framework rules location**: {TEAM_PATH} (auto-detected at runtime)
- You READ from here: prompts, workflows, templates, shared.md
- You NEVER WRITE here

When you see {PROJECT_PATH}, it means:
- **Project workspace**: {PROJECT_PATH} (current working directory)
- You READ project config from: {PROJECT_PATH}/.team/project.md
- You WRITE operational files to: {PROJECT_PATH}/.team/

When you see {docs_internal}, it means:
- **Project internal docs**: {DOCS_INTERNAL} (auto-detected)
- You WRITE project docs here (SPEC.md, AC.md, RCA.md, etc.)

### 4. Self-check before every write

Before writing ANY file, ask yourself:

`
Am I about to write to {TEAM_PATH}/ directory?
  ├── YES -> STOP! Re-check the path. Operational files go to {PROJECT_PATH}/.team/
  └── NO  -> OK, proceed
`

---

*Version: 1.0 | Created: 2026-04-27*
*Purpose: Prevent AI agents from writing project operational files to the framework layer.*