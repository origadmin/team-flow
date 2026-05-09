# SCOPE.md — Machine-Readable Change Report
# Encoding: UTF-8
# Parse: Each section is a YAML block, separated by ---

task_id: "team-flow-6x9.28"
beads_id: "team-flow-6x9.28"
task_type: "change"
iteration: "R1"
date: "2026-05-10"
author: "Triage"

files:
  added:
  modified:
    - path: "team/v2/prompts/qa-engineer.md"
      description: "Add Entry Gate: TaskPool beads ID check, remove trigger keywords to prevent direct activation"
    - path: "team/v2/prompts/dev-frontend.md"
      description: "Add Entry Gate: TaskPool beads ID check"
    - path: "team/v2/prompts/dev-backend.md"
      description: "Add Entry Gate: TaskPool beads ID check"
    - path: "team/v2/prompts/bugfix.md"
      description: "Add Entry Gate: TaskPool beads ID check, remove trigger keywords to prevent direct activation"
    - path: "team/v2/prompts/analysis.md"
      description: "Add Entry Gate: TaskPool beads ID check, remove trigger keywords to prevent direct activation"
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
  Fixed Triage bypass vulnerability: added TaskPool Entry Gate to all 5 sub-agents, removed direct trigger keywords. Prevents QA/Dev/Bugfix/Analysis from talking to users directly.

---
# Above this line: machine-readable (YAML)
# Below this line: human notes (optional)

## Change Details

### Entry Gate Pattern
Added identical Entry Gate to:
1. qa-engineer.md
2. dev-frontend.md
3. dev-backend.md
4. bugfix.md
5. analysis.md

### Entry Gate Logic
```
Sub-agent activation check:
    |
    +-- Status Line TaskPool has valid beads ID? -> continue
    |   +-- TaskPool = -(N/A)? -> REJECT. You are bypassing Triage dispatch.
    |
    +-- beads issue exists? -> continue
    +-- Phase label correct? -> continue
```

### Keywords Removed
- qa-engineer.md: [测试, QA, 验证, 回归, 质量]
- bugfix.md: [Bugfix, 修复Bug, fix bug, 根因]
- analysis.md: [分析, 调研, evaluation, research]

### Git Reference
- Commit: `778ad0e` (team-flow subrepo)
