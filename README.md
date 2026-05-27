# team-flow — Process-Centric AI Collaboration Framework

[![Go Report Card](https://goreportcard.com/badge/github.com/origadmin/team-flow)](https://goreportcard.com/report/github.com/origadmin/team-flow)
[![npm version](https://img.shields.io/npm/v/@origadmin/team-flow.svg)](https://www.npmjs.com/package/@origadmin/team-flow)

**team-flow** is a process-driven AI collaboration framework that transforms chaotic AI interactions into structured, verifiable workflows. Define your team, define your process, and let AI execute with quality gates and verification chains.

> **v3 is the current version.** Process is central — rules, tools, and skills are unified under flow definitions.

## Why team-flow?

| Problem | team-flow Solution |
|---------|-------------------|
| AI agents work without structure | **Flow-driven execution** — every task follows a defined process |
| AI says "done" but quality is unknown | **Verification chain** — downstream roles verify upstream deliverables |
| Multiple AI agents talk to the user | **Principal role** — single user interface, all others execute silently |
| No quality enforcement | **Gate nodes** — automated quality checks block progress until passed |
| Prompt engineering is ad-hoc | **Structured rules** — instruction (30 chars) + reference + full description |
| Work is lost between sessions | **Session persistence** — every session produces traceable logs |
| Can't visualize what AI is doing | **Flow Editor** — web UI for visual workflow editing and monitoring |

## Key Features

### Flow-Driven Architecture
Everything starts with a flow definition. A flow specifies:
1. **What steps to execute** (nodes)
2. **What order** (edges with conditions)
3. **What rules apply** per step (components)
4. **What deliverables to produce** (docs with templates)
5. **What gates enforce quality** (automated checks)

### Verification Chain
AI cannot self-verify. Every role does two things: execute work + verify upstream deliverables.

```
Role A (execute + self-check) → Role B (verify A + execute + self-check) → ... → QA (independent final check) → Done
```

### Principal Role System
Every flow has exactly one **principal** role — the team's public face:
- Only the principal communicates with the user
- User input always routes to the principal
- Sub-roles report results to the principal
- Interruptions return to the principal for re-assessment

### 5 Preset Teams, 15+ Flows

| Team | Flows | Best For |
|------|-------|----------|
| **Dev Team** | 8 flows (dev/feature/bugfix/hotfix/change/release/analysis/batch) | Software development |
| **Content Team** | 3 flows (content-distribution/novel/promo-video) | Articles, videos, novels |
| **Game Team** | 1 flow (game-design) | Game design |
| **Trading Team** | 1 flow (trading) | Trading analysis |
| **Skill Team** | 2 flows (skill-dev/test-simple) | AI skill development |

### Compatible with 10+ AI Tools

Claude Code, Cursor, Trae, Windsurf, GitHub Copilot, OpenCode, OpenClaw, Cline, Gemini CLI, WorkBuddy, QClaw

## Installation

### Via Go (Recommended for developers)

```bash
go install github.com/origadmin/team-flow/cmd/flow@latest
```

### Via npm (Recommended for AI tool users)

```bash
npx skills add @origadmin/team-flow
```

This uses embedded skill files — no network required after install.

### Build from Source

```bash
git clone https://github.com/origadmin/team-flow.git
cd team-flow
go build -o flow ./cmd/flow/
```

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Build and run the CLI |
| Node.js | 18+ | npm installation method |

## Quick Start

### 1. Initialize a New Project (v3)

```bash
cd /path/to/your/project

# Interactive setup — choose from 5 preset teams
flow init --v3

# Or specify a team directly
flow init --v3 --team dev-team

# Or specify a flow directly
flow init --v3 --flow dev-flow
```

`flow init --v3` will:
1. Create `.team/` directory and configuration files
2. Show preset team menu (5 teams) — pick the closest match
3. Install selected team's flows to `.team/flows/`
4. Set `default_flow` in project configuration

### 2. Start Working

```bash
# Enter the flow engine — get your work instructions
flow proc run

# The engine tells you: your role, persona, rules, tools, deliverables
```

### 3. Check Project Status

```bash
flow project detect    # Detect project and lock status
flow status            # Show project status
flow config paths      # Show resolved path variables
```

### 4. Migrate from v2

```bash
# Migrate to v3 (7-step automatic migration)
flow migrate v3

# Rollback to v2 if needed (6-step automatic rollback)
flow migrate rollback
```

## Commands Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `flow init --v3` | Initialize project with v3 flow system |
| `flow init --v3 --team <id>` | Initialize with specific team |
| `flow proc run` | Start flow engine — get work instructions |
| `flow proc run <node-id>` | Run a specific flow node |
| `flow proc list` | List all registered flows |
| `flow proc validate` | Validate flow definitions |
| `flow proc rule <id>` | Show full rule description |
| `flow project detect` | Detect and lock project |
| `flow status` | Show project status |
| `flow version` | Show version info |

### Configuration

| Command | Description |
|---------|-------------|
| `flow config paths` | Show resolved path variables |
| `flow config paths --json` | Show paths in JSON format |
| `flow config get <key>` | Get specific config value |

### Task Management

| Command | Description |
|---------|-------------|
| `flow task list` | List all tasks |
| `flow task show <id>` | View task details |
| `flow task update <id> --claim` | Claim a task |
| `flow task close <id>` | Close a task |
| `flow task sync --source github --repo owner/repo` | Sync GitHub Issues |

### Migration

| Command | Description |
|---------|-------------|
| `flow migrate v3` | Migrate from v2 to v3 |
| `flow migrate rollback` | Rollback from v3 to v2 |

### Visualization

| Command | Description |
|---------|-------------|
| `flow editor` | Start visual flow editor (Web UI) |

### Code Graph Analysis

| Command | Description |
|---------|-------------|
| `flow graph build` | Build code graph |
| `flow graph search <query>` | Semantic search |
| `flow graph impact [files...]` | Analyze impact radius |
| `flow graph changes` | Detect and analyze changes |

## Project Structure

```
team-flow/
├── cmd/flow/                  # CLI entry point
├── internal/
│   ├── boot/                  # flow init (v1/v2/v3)
│   ├── proc/                  # flow proc run (v3 engine)
│   ├── flow/                  # Flow parser, validator, serializer
│   ├── config/                # flow config paths
│   ├── project/               # flow project detect
│   ├── migrate/               # flow migrate (v2→v3, rollback)
│   ├── task/                  # flow task management
│   ├── skill/                 # Skill management
│   ├── editor/                # flow editor (Web UI)
│   ├── graph/                 # flow graph (code analysis)
│   ├── doctor/                # flow doctor (diagnostics)
│   └── version/               # Version info
├── v3/
│   ├── flows/                 # 15+ v3 flow definitions (JSON)
│   ├── schema/                # Flow JSON Schema
│   ├── teams/                 # 5 preset team templates
│   └── docs/                  # Architecture decisions & specs
├── team/
│   ├── v3/                    # v3 skill files (SKILL.md, prompts, templates)
│   ├── v2/                    # v2 fallback (preserved during migration)
│   └── v1/                    # v1 legacy (preserved during migration)
├── editor/                    # Flow visual editor (React + TypeScript)
├── skillfs.go                 # Embedded FS (team/ + v3/flows + v3/schema)
├── go.mod
├── package.json               # npm package (@origadmin/team-flow)
└── SKILL.md                   # v3 entry point for AI tools
```

## v2 vs v3

| Aspect | v2 | v3 |
|--------|----|----|
| Architecture | Prompt-driven (markdown files) | **Flow-driven** (JSON process definitions) |
| Execution | AI reads prompts, ad-hoc execution | **Engine-driven** — `flow proc run` provides structured instructions |
| Quality | Manual review | **Verification chain** — downstream verifies upstream |
| User interface | Any role can talk to user | **Principal role** — single user interface |
| Gates | None | **Automated quality gates** block progress |
| Rules | Scattered in markdown | **Three-layer** — instruction + reference + full description |
| Teams | Single generic team | **5 preset teams** with domain-specific flows |
| Configuration | project.md (monolithic) | **project.yaml** (structured, ~100 tokens) |
| Session tracking | None | **Session logs** for traceability |
| Migration | v1→v2 manual | **v2→v3 automatic** (7 steps) with rollback |
| Visual editor | None | **Flow Editor** — web UI for workflow visualization |
| Compatible AI tools | Limited | **10+ tools** — Claude Code, Cursor, Trae, etc. |

## Config-Run Architecture

team-flow separates **configuration** from **execution**:

| Layer | Responsibility | Output |
|-------|---------------|--------|
| **Config** | Resolve paths, validate settings | Resolved absolute paths |
| **Run** | Execute flows, use resolved values | Structured output (no placeholders) |

```bash
# Config layer — resolve and validate
flow config paths          # Show all resolved paths
flow config get docs_internal  # Get specific path

# Run layer — execute with resolved values
flow proc run              # Get work instructions (all paths resolved)
```

## Session Startup Protocol

Every AI session follows this protocol:

```
Step 1: flow project detect    → Lock the project
Step 2: flow proc run          → Get work instructions (role, rules, tools, docs)
Step 3: Adopt principal role   → Start working
```

## License

MIT License
