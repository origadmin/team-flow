# Path Resolution

AI must resolve all paths at startup via `flow config paths --json`. This is the **single source of truth** for path variables — AI must never self-resolve relative paths.

## Startup Sequence

```bash
flow config paths --json
```

## Path Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `{WORKSPACE}` | Multi-project workspace root | yes |
| `{PROJECT}` | Current project root directory | yes |
| `{DOCS_INTERNAL}` | Internal docs (team-only, not public) | no |
| `{DOCS_EXTERNAL}` | External docs (public, open-source documentation) | no |
| `{TASK_DB}` | task database directory | yes |
| `{TEAM_PATH}` | Skill installation path | yes |
| `{TMP_DIR}` | AI temporary files directory | no |

## Path Anchor Rules

| Prefix | Anchor | Example | Resolves to |
|--------|--------|---------|-------------|
| `_` | workspace root | `_docs/orig-cms/` | `{WORKSPACE}/_docs/orig-cms/` |
| other | project root | `docs/` | `{PROJECT}/docs/` |
| not configured | — | — | path not available |
| not configured | project root | — | `{PROJECT}/.team/tmp/` (default) |

## Constraints

- If `docs_internal` or `docs_external` not in output → not available, AI must not use those paths
- AI must never self-resolve relative paths. Always use flow-resolved absolute paths
- AI must write temporary files to `{TMP_DIR}` only (default: .team/tmp/)
- Never create temp files in project root or workspace root
- Clean up {TMP_DIR} when session ends

## Asset Resolution Protocol

Pointers in JSON flow definitions (e.g., `prompt_source`, `ref`, `template`) MUST be resolved to local file paths.

1. **Identify the Reference**: Extract the `ref` or `path` from the engine output.
2. **Resolve the Path**: 
   - If it starts with `assets/skill/v3/` → Resolve relative to `{TEAM_PATH}`.
   - If it's a relative path → Resolve relative to `{PROJECT}`.
3. **MANDATORY Action**: AI MUST use `read_file` on the resolved path BEFORE executing the task for that node. **Implicit knowledge is forbidden.**

Example:
- `prompt_source: "assets/skill/v3/prompts/tech-lead.md"`
- Resolved: `{TEAM_PATH}/prompts/tech-lead.md`
- Action: `read_file(file_path="{TEAM_PATH}/prompts/tech-lead.md")`

- `{DOCS_PATH}` → replaced by `{DOCS_INTERNAL}`
- `{PROJECT_PATH}` → replaced by `{PROJECT}`

## Current Variables

- `{TMP_DIR}` → AI temporary files directory (default: `{PROJECT}/.team/tmp/`)
