# Config-Run Architecture Specification

## Overview

This document defines the clear separation between **config** (configuration layer) and **run** (execution layer) in the team-flow framework.

## Core Principle

**Config** = Configuration layer (handles placeholders, path resolution, environment variables)
**Run** = Execution layer (uses resolved values, no placeholders in output)

## Configuration Sources

### Primary Source: `.team/project.yaml` (YAML)
- This is the **main** source of truth for project configuration
- Uses standard YAML format for structured data
- **Must not** be committed to Git (contains local machine-specific settings)

### Backup Source: `.team/project.md` (Markdown)
- **Fallback** only - for backward compatibility
- Should be considered deprecated for new configurations
- Generated automatically from YAML for human readability

### Structure of `project.yaml`

```yaml
name: framework
version: v3
active_flow: dev-flow

paths:
    projects_path: projects/
    docs_internal: _docs/framework/
    docs_external: docs/

toolchain:
    backend:
        language: go
        pipeline: go test ./... | go build -o bin/app
    frontend:
        language: typescript
        pipeline: bun run test | bun run build

flows:
    - id: dev-flow
      source: preset

# Flow CLI-specific settings (optional)
flow:
    path: scripts/flow.exe
    backup_path: _back/
    update_disabled: false
    update_interval: 24h
```

## Config Layer Responsibilities

The `flow config` command handles:

| Responsibility | Description |
|---------------|-------------|
| **Placeholder Resolution** | Resolve relative paths → absolute paths |
| **Path Validation** | Verify executable files exist |
| **Default Values** | Provide sensible defaults when config missing |
| **Priority Handling** | YAML first, MD fallback |

### Config Keys

```bash
flow config show                        # Show all config (resolved)
flow config get project_name           # Get project name
flow config get flow_executable         # Get resolved flow path (absolute)
flow config get docs_internal           # Get internal docs path
flow config get project_root           # Get project root
```

## Run Layer Responsibilities

The `flow proc run` command outputs:

| Output | Example |
|--------|---------|
| **PROJECT_ROOT** | `D:\workspace\project\golang\origadmin\framework` |
| **Commands** | `flow project detect` (not `{flow_path} project detect`) |
| **Paths** | All absolute, no placeholders |

### Run Output Rules

✅ **Correct:**
```
PROJECT_ROOT: D:\workspace\project\golang\origadmin\framework
flow project detect
```

❌ **Incorrect:**
```
PROJECT_ROOT: {PROJECT_ROOT}
{flow_path} project detect
```

## Data Flow

```
.team/project.yaml (primary source)
       ↓
flow config (resolve paths, validate)
       ↓
Resolved Configuration (absolute paths)
       ↓
flow proc run (use resolved values)
       ↓
Structured Output (no placeholders)
```

**Note:** `.team/project.md` is only used as a fallback if YAML is not found.

## Implementation Requirements

### Config Loading Priority

1. **First**: Load from `project.yaml`
2. **Fallback**: Load from `project.md` (for compatibility)
3. **Defaults**: Use sensible defaults if neither exists

### ProjectConfig Structure

```go
type ProjectConfig struct {
    Name         string
    Version      string
    ActiveFlow   string
    Paths        ProjectPaths
    Toolchain    ProjectToolchain
    Flows        []ProjectFlow
    Flow         FlowConfig  // CLI-specific settings
}
```

## Usage for AI

**Correct**:
```bash
# Let flow CLI handle all config
flow project detect
flow proc run

# If you need a specific config value
flow config get flow_executable
```

**Incorrect**:
```bash
# Don't read .team/ files directly!
cat .team/project.yaml  # ❌
cat .team/project.md    # ❌
```

## Summary

| | Config | Run |
|--|--------|-----|
| **Source** | `project.yaml` (main), `project.md` (fallback) | Output from config layer |
| **Output** | Resolved absolute paths | Final values, no placeholders |
| **Format** | YAML structured data | Human-readable instructions |
| **Change Frequency** | Occasionally (setup) | Every execution |
