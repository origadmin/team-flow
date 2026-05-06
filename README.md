# team-flow - AI Team Collaboration Framework CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/origadmin/team-flow)](https://goreportcard.com/report/github.com/origadmin/team-flow)

`flow` is the CLI for the `_team` AI collaboration framework. It provides initialization, v1→v2 migration, code graph analysis, and workflow visualization.

## Features

- **Project Init** - Quick setup with v1 (task-pool) or v2 (beads + graph) mode
- **v1 → v2 Migration** - Automatic migration from task-pool.md to beads-native
- **Code Graph Analysis** - Semantic search, impact analysis, change detection, flow tracking
- **Task Tracking** - Integration with beads (bd CLI) for persistent task management
- **Flow Editor** - Web UI for visual workflow editing (in development)

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.21+ | https://go.dev/dl/ |
| beads (bd) | 1.0.3+ | `irm https://raw.githubusercontent.com/steveyegge/beads/main/install.ps1 \| iex` |
| code-review-graph | 2.3.2+ | `pip install code-review-graph` |

## Install

```bash
go install github.com/origadmin/team-flow/cmd/flow@latest
```

Or build from source:

```bash
git clone https://github.com/origadmin/team-flow.git
cd team-flow
go build -o flow ./cmd/flow/
```

## Quick Start

### 1. Initialize Project

```bash
cd /path/to/your/project

# v2 mode (default, recommended)
flow init

# v1 mode (legacy)
flow init --v1
```

### 2. Setup v2 Tools

```bash
# Initialize beads database
bd init

# Build code graph
flow graph build
```

### 3. Check Status

```bash
flow status
```

### 4. Migrate from v1

```bash
# Dry run first
flow migrate --dry-run

# Actual migration
flow migrate
```

## Commands

| Command | Description |
|---------|-------------|
| `flow init` | Initialize project |
| `flow init --v1` | Use v1 (task-pool) mode |
| `flow migrate` | Migrate v1 → v2 |
| `flow migrate --dry-run` | Preview migration |
| `flow status` | Show project status |
| `flow graph build` | Build code graph |
| `flow graph status` | Show graph statistics |
| `flow graph search <query>` | Semantic search |
| `flow graph query --pattern <p> --target <t>` | Structured graph query |
| `flow graph impact [files...]` | Analyze impact radius |
| `flow graph flows [files...]` | Trace affected flows |
| `flow graph changes` | Detect and analyze changes |
| `flow graph review [files...]` | Get review context |
| `flow graph serve` | Start MCP server |
| `flow editor` | Start visual editor |
| `flow version` | Show version |

## Graph Analysis Examples

```bash
# Semantic search
flow graph search "tag handler list"

# Find callers of a function
flow graph query --pattern callers_of --target ListTags

# Impact analysis for changed files
flow graph impact internal/features/content/service/tag_handler.go

# Trace affected flows
flow graph flows internal/features/content/service/tag_handler.go

# Detect current changes with risk scoring
flow graph changes
```

## Project Structure

```
team-flow/
├── cmd/flow/              # CLI entry point
├── internal/
│   ├── initcmd/           # flow init
│   ├── migratecmd/        # flow migrate (v1→v2)
│   ├── statuscmd/         # flow status
│   ├── graphcmd/          # flow graph (code-review-graph)
│   ├── editor/            # flow editor (Web UI)
│   └── version/           # Version info
├── team/                  # _team framework files
│   ├── prompts/           # Role execution rules
│   ├── templates/         # Document templates
│   ├── workflows/         # Workflow standards
│   └── v2/                # v2 (beads-native) rules
├── go.mod
└── README.md
```

## v1 vs v2

| Aspect | v1 | v2 |
|--------|----|----|
| Task tracking | task-pool.md (manual) | beads (bd CLI) |
| Code analysis | Manual grep | code-review-graph |
| Impact analysis | Manual search | `flow graph impact` |
| Change detection | git diff | `flow graph changes` |
| Task source of truth | task-pool.md | .beads/ database |
| Multi-agent sync | git conflicts | Dolt git-native |

## License

Apache License 2.0
