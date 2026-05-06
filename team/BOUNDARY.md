# BOUNDARY.md - Framework Layer Boundary Rules

> **This file MUST be read before any other _team/ file.**
> All AI agents (Trae, Gemini CLI, OpenClaw, Cursor, etc.) MUST comply.

---

## LAYER ARCHITECTURE

`
framework/                    <-- MULTI-PROJECT MANAGEMENT PLATFORM
  _team/                      <-- FRAMEWORK LAYER (shared rules, read-only for AI)
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

### 1. _team/ is READ-ONLY for AI execution

| Operation | _team/ (Framework) | .team/ (Project) |
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
| task-pool.md | {PROJECT_PATH}/.team/task-pool.md | _team/task-pool.md |
| backlog.md | {PROJECT_PATH}/.team/backlog.md | _team/backlog.md |
| project.md | {PROJECT_PATH}/.team/project.md | _team/project.md |
| SPEC.md | {docs_internal}/requirements/... | _team/... |
| AC.md | {docs_internal}/requirements/... | _team/... |
| RCA.md | {docs_internal}/reports/bugs/... | _team/... |

### 3. Path variable usage

When you see {TEAM_PATH}, it means:
- **Framework rules location**: D:/workspace/project/golang/origadmin/framework/_team
- You READ from here: prompts, workflows, templates, shared.md
- You NEVER WRITE here

When you see {PROJECT_PATH}, it means:
- **Project workspace**: D:/workspace/project/golang/origadmin/framework/projects/orig-cms
- You READ project config from: {PROJECT_PATH}/.team/project.md
- You WRITE operational files to: {PROJECT_PATH}/.team/

When you see {docs_internal}, it means:
- **Project internal docs**: D:/workspace/project/golang/origadmin/framework/_docs/orig-cms/
- You WRITE project docs here (SPEC.md, AC.md, RCA.md, etc.)

### 4. Self-check before every write

Before writing ANY file, ask yourself:

`
Am I about to write to _team/ directory?
  ├── YES -> STOP! Re-check the path. Operational files go to {PROJECT_PATH}/.team/
  └── NO  -> OK, proceed
`

---

*Version: 1.0 | Created: 2026-04-27*
*Purpose: Prevent AI agents from writing project operational files to the framework layer.*