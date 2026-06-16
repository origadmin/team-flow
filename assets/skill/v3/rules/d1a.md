---
id: d1a
name: Task Classification Rules
enforcement: mandatory
phase: [sta0, tri3]
---

# Task Classification Rules (d1a)

## Required Actions

1. **Understand Intent**: Read user input thoroughly. Ask clarifying questions if ambiguous.
2. **Classify Task Type**: Must categorize into one of: feature, bug, hotfix, analysis, change
3. **Assess Priority**: Assign priority (high/medium/low) based on impact and urgency
4. **Document Decision**: Write rationale for classification and priority

## Deliverables

| Phase | Deliverable | Content |
|-------|------------|---------|
| sta0  | CONTEXT.md | Intent summary, preliminary type, open questions |
| tri3  | TRIAGE.md  | Task type, priority, scope, target_branch |

## Constraints

- **MUST** produce the required doc before advancing to next node
- **MUST** specify a valid task_type (feature|bug|hotfix|analysis|change)
- **MUST** ask user when uncertain — never guess
- **MUST NOT** modify source code during classification

## Gate Conditions

- CONTEXT.md exists and is non-empty (sta0 exit)
- TRIAGE.md exists and contains valid task_type (tri3 exit)
- Task created in .team/tasks/ (tri3 exit)