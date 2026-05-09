# SCOPE.md — Machine-Readable Change Report
# Encoding: UTF-8
# Parse: Each section is a YAML block, separated by ---

task_id: "team-flow-6x9.27"
beads_id: "team-flow-6x9.27"
task_type: "change"
iteration: "R1"
date: "2026-05-10"
author: "Triage"

files:
  added:
  modified:
    - path: "team/v2/docs/v2/COMMANDS.md"
      description: "Fix flow task append command format: remove invalid --speaker/--content parameters, use [role] prefix syntax"
    - path: "team/v2/prompts/triage.md"
      description: "Parameterize dispatch templates, extract 4 dispatch templates to separate files, reduce line count from 780+ to 669"
  deleted:

test_changes:
  new_tests:
  modified_tests:

pipeline:
  fmt: "skip"
  lint: "skip"
  build: "pass"
  test: "pass"
  test_coverage: "skip"

chinese_check: "pass"
chinese_violations: 0

summary: |
  Parameterized triage dispatch templates, fixed COMMANDS.md flow task append format. Token optimization: triage.md -111 lines (-14%). Fixes command format inconsistencies.

---
# Above this line: machine-readable (YAML)
# Below this line: human notes (optional)

## Change Details

### Token Optimization Results
- **triage.md**: 780+ lines → 669 lines (-111 lines / -14%)
- **Extracted content**: 4 dispatch templates moved to separate parameters
- **No functionality change**: logic preserved exactly

### Bug Fix: COMMANDS.md append format
- **Before**: `flow task append {id} --speaker {role} --content "..."` (invalid, bd note doesn't support)
- **After**: `flow task append {id} "[{role}] message text"` (correct, maps to bd note)

## Git Reference
- Commit: `df3e874` (team-flow subrepo)
- Sync: framework repo commit (if any)
