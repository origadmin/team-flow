# Directory Structure (v3 Active)

## Embedded Assets (in flow binary)

```
assets/                        ← Unified embed root (skillfs.go)
├── skill/                     ← Skill files (v1/v2/v3 coexist)
│   ├── v1/                    ← v1: task-pool workflow
│   ├── v2/                    ← v2: beads-native workflow
│   └── v3/                    ← v3: flow-driven workflow
│       ├── SKILL.md           ← v3 entry point (HOT)
│       ├── references/        ← On-demand references (COLD)
│       ├── prompts/           ← Role execution rules
│       ├── workflows/         ← Shared workflows + role standards
│       ├── templates/         ← Document templates
│       ├── skills/            ← Specialized domain skills
│       │   ├── team-flow-v3-create/
│       │   └── team-flow-v3-exec/
│       ├── BOUNDARY.md
│       └── CONSENSUS.md
├── orgs/                      ← Organization templates (team + flows)
│   ├── dev-team/
│   │   ├── team.json
│   │   └── flows/
│   ├── content-team/
│   ├── game-team/
│   ├── trading-team/
│   └── skill-team/
├── flows/                     ← Preset flow definitions
└── schema/                    ← JSON Schema for flow validation
```

## Project-Level (installed by flow init --v3)

```
.team/
├── project.yaml           ← Runtime config (primary, ~100 tokens)
├── constraints.md         ← Project constraints (~50 tokens)
├── project.md             ← Legacy fallback (auto-generated)
├── version                ← v3
├── team.json              ← Installed team definition
├── skills.yaml            ← Skill resolution config
├── flows/                 ← Installed flow JSON files
└── docs/                  ← Internal docs (fallback when docs_internal not set)
    ├── lessons/           ← Experience records
    ├── sessions/          ← Session logs (traceability)
    └── conventions/       ← Project conventions
```

## IDE Skill Directory (installed by npx or flow init)

```
.trae/skills/team-flow/
├── SKILL.md               ← Root entry (routes to v1/v2/v3)
├── bin/flow.exe           ← Compiled binary
└── (v1/v2/v3 skill files copied from assets/skill/)
```
