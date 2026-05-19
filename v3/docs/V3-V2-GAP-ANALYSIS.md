# V3 vs V2 Gap Analysis

**Date:** 2026-05-20  
**Status:** P0 completed, P1 previously completed, P2 remaining

## Summary

v3 now covers ~85% of v2 core workflows. Remaining gaps are primarily in P2 (nice-to-have) items and engine-level implementations that need runtime validation.

## P0 Items — COMPLETED ✅

| ID | Item | Status | Details |
|----|------|--------|---------|
| P0-1 | Rule content output | ✅ Done | Engine outputs rule name+description+enforcement via extractRules() |
| P0-2 | Prompt loading path | ✅ Done | RoleDefinition has prompt_source/standards_source; engine exposes in output; node components have prompts refs |
| P0-3 | Batch flow | ✅ Done | batch-flow.json: 7-phase flow with concurrency control (≤3 parallel), failure retry (3x), dependency queue, conflict detection |
| P0-4 | Release flow | ✅ Done | release-flow.json: readiness→verify→accept→deploy with 4 gates, rollback plan, MILESTONES.md update |

## P1 Items — COMPLETED ✅ (previous commit 783b90d)

| ID | Item | Status |
|----|------|--------|
| B5 | Asset spec with R{N} paths | ✅ |
| B6 | R-iteration (bugfix) | ✅ |
| B7 | Regression Guard rules | ✅ |
| B8 | Hard Constraints rules | ✅ |
| B9 | Output Guard integration | ✅ |
| B11 | Handoff docs | ✅ |

## P2 Items — REMAINING

| ID | Item | Priority | Notes |
|----|------|----------|-------|
| B3 | Skill routing / skills registration | Low | v2 has 6 sub-skills with trigger patterns; v3 has skill refs but no auto-trigger |
| B10 | ID naming convention enforcement | Low | Engine does not validate node IDs against `^[a-z][a-z0-9]{3,4}$` pattern |
| B11 | Session handoff mechanism | Low | v2 has sub-agent communication pattern; v3 has handoff-doc on terminal exit only |
| C3 | Consensus mechanism | Low | v3 has no built-in consensus; currently managed via .team/consensus.md in v2 |
| C4 | Checklist coverage | Low | v3 has no checklist integration; currently managed via .team/checklist.md in v2 |
| C8 | v3-create SKILL.md split | Medium | 557 lines exceeds 500-line limit; node type reference should move to references/node-types.md |

## Engine-Level Gaps

| Gap | Description | Impact |
|-----|-------------|--------|
| Prompt loading | Engine outputs paths but does not read/prompt content | Medium — runtime needs to load prompt file content |
| Variable substitution | {TEAM_PATH}, {iteration} etc. in prompt paths not resolved at runtime | Medium — vars in node config resolved, but role def paths not yet |
| Gate evaluation | Gate conditions are declarative but not auto-evaluated by engine | Low — currently informational for AI to enforce |
| Parallel execution | Parallel node type defined but engine runs sequentially | Low — AI can simulate via sub-agents |

## Flow Coverage Matrix

| v2 Workflow | v3 Flow | Coverage |
|-------------|---------|----------|
| Feature flow | feature-flow.json | 95% |
| Bugfix flow | bugfix-flow.json | 95% (with R-iteration) |
| Change flow | change-flow.json | 90% |
| Analysis flow | analysis-flow.json | 85% |
| Hotfix flow | hotfix-flow.json | 90% |
| Dev flow | dev-flow.json | 85% (18KB→multi-subflow) |
| Batch task management | batch-flow.json | 80% (new in v3) |
| Release management | release-flow.json | 85% (new in v3) |
| Game design | game-design-flow.json | N/A (domain-specific) |
| Novel writing | novel-flow.json | N/A (domain-specific) |

## Key Architectural Advantages of v3 over v2

1. **Structural enforcement**: Rules that v2 relies on AI self-discipline are enforced by flow structure (gates, completion checks, dispatch guard)
2. **Component registry**: Centralized definitions eliminate duplication across flows
3. **Prompt loading**: Explicit prompt paths enable runtime prompt injection per role
4. **R-iteration**: First-class iteration tracking with subdirectory isolation
5. **Batch + Release**: New flow types not available in v2
