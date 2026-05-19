<!-- SNAPSHOT - Frozen as of 2026-04-24 - DO NOT REGENERATE -->
<!-- This document records decisions made during v6.0 refactor -->
<!-- For current architecture, see docs/design/ARCHITECTURE.md -->

# Team Workflow Framework Refactoring Summary

> **Version**: v6.0 | **Date**: 2026-04-24 | **Scope**: All AI rule files + directory reorganization + human-readable doc system

---

## 1. Evolution Overview

| Version | Date | Core Change | Key Decision |
|---------|------|-------------|--------------|
| v3.7 | 2026-04-21 | Multi-role collaboration, rule doc mode | Initial role separation |
| v4.0 | 2026-04-22 | Three-layer gate, process engine | Gates are non-bypassable |
| v5.0 | 2026-04-23 | Client-vendor separation, doc classification | Triage as dispatch tool |
| v5.5 | 2026-04-24 | AI file space optimization, deleted ~69.5KB | Rule files pure rules |
| **v6.0** | **2026-04-24** | **Two-layer workflow + DOMGEN + directory reorg** | **Task/Release separation, auto doc maintenance** |

---

## 2. v6.0 Core Changes

### 2.1 Two-Layer Workflow Separation

**Problem**: v5.0 Phase 0-7 mixed two granularity levels - task level (Feature/Bug creation to completion) and release level (Milestone from ready to deployment). Phase 4-7 could not be triggered by a single task.

**Decision**: Task layer manages feature block completion, release layer manages deployment. Independent triggering, no coupling.

| Layer | Trigger | Phases | Endpoint |
|-------|---------|--------|----------|
| **Task Layer** | Task ID | Phase 0-3 | Feature block ready (Review) |
| **Release Layer** | Milestone ID | R-Phase 0-3 | Deployed to production |

**Impact**:
- SKILL.md entry gate adds R: prefix
- shared.md adds complete release workflow + closed-loop verification + quality gates
- PM role activated (sole release approver)
- DevOps role activated (CI/CD + deploy + monitoring)

### 2.2 DOMGEN - Human-Readable Doc Auto-Maintenance System

**Problem**: After v5.0 simplification, explanatory content was deleted. Newcomers/PMs couldn't understand rule intent. Dual-source maintenance caused drift.

**Decision**: AI rules = single source of truth. Human docs derived automatically via DOMGEN. SKILL/DOMGEN responsibilities fully separated.

**DOMGEN Core Rules**:
- Background markers (<=50 chars) expanded to 100-200 char explanations
- Decision trees converted to flow descriptions
- Forbidden lists supplemented with rationale
- Templates filled with examples
- AUTO-GENERATED header + Hash footer

### 2.3 Background Marker System

**Problem**: v5.0 simplification reduced understandability.

**Decision**: AI rules use brief background markers (<=50 chars), expanded by DOMGEN for human docs. Rules stay concise, human docs provide full context.

Current distribution: shared.md (13), dev.md (10), tech-lead.md (4).

### 2.4 Directory Reorganization

**Before** (v5.5): Human-readable docs scattered in _team root + 14 unused standards files in workflows/roles/

**After** (v6.0): Clean separation - AI rules at _team root, human docs in docs/ (DOMGEN managed), snapshots frozen in docs/design/ and docs/migration/

**Deleted files** (v5.5->v6.0 cumulative): ~63KB of unused human-readable files (standards.md, dev-workflow.md, framework-workflow.md, shared-implementation.md, output-standards.md, 14 roles/* files)

**New files** (v6.0): DOMGEN.md (7.3KB) + 8 derived docs (~50KB total)

**Moved files**: All human docs from root to docs/ subdirectories

---

## 3. Architecture Decision Records

### D1: Triage Does Not Create Asset Packages (v5.0)

Triage only classifies and dispatches. Asset packages created by vendor side (Tech Lead/Dev).

### D2: MILESTONES Belongs to Client (v5.0)

MILESTONES Owner = Client. Triage auto-syncs status from Task Pool.

### D3: Change Conservative Strategy (v5.0)

Default: update MILESTONES with "Change evaluating" status. Vendor assesses impact and feeds back.

### D4: Rule Files Pure Rules (v5.0)

AI rules only contain: decision trees + execution templates + forbidden lists + trigger conditions. Explanatory content moved to human docs.

### D5: Two-Layer Workflow Separation (v6.0)

Task layer Phase 0-3 for feature blocks, release layer R-Phase 0-3 for deployment. Independent triggering, no coupling.

### D6: DOMGEN Auto-Maintenance (v6.0)

AI rules = single source of truth. DOMGEN derives human docs. SKILL/DOMGEN responsibilities fully separated.

### D7: Background Markers (v6.0)

Brief background markers in AI rules, expanded by DOMGEN for human readability.

### D8: Snapshot Freeze Mechanism (v6.0)

Snapshot docs (migration/design) are frozen after creation. DOMGEN never touches snapshots.

### D9: Doc Binary Classification (v6.0)

docs/ only has two types: Derived (DOMGEN managed) and Snapshot (frozen). No "manually maintained" category allowed.

---

## 4. Results Comparison

### 4.1 File Size (v4.0 -> v6.0)

| Type | v4.0 | v5.0 | v6.0 | Change |
|------|------|------|------|--------|
| AI rule files | ~70KB | ~30KB | ~30KB | -57% |
| Human-readable docs | ~50KB (mixed in rules) | ~6KB (README+ARCH only) | ~50KB (docs/ derived) | Independent, no AI context pollution |
| Obsolete files | ~69KB | ~69KB | 0 | -100% |

### 4.2 Maintainability

| Dimension | v4.0 | v5.0 | v6.0 |
|-----------|------|------|------|
| AI rule purity | Rules+explanation mixed | Pure rules | Pure rules + background markers |
| Human docs | Missing | Basic (2 files) | Complete (8 derived + 4 snapshots) |
| Dual-source drift risk | High (manual) | High (manual) | Low (DOMGEN auto-sync) |
| Doc ownership | Unclear | Partially clear | Fully clear (derived/snapshot) |
| Newcomer understandability | Poor | Poor (explanations deleted) | Good (independent human docs) |

---

## 5. DOMGEN Generation Record

### First Full Generation (2026-04-24)

| File | Source | Hash | Size |
|------|--------|------|------|
| docs/README.md | SKILL.md | 4F95D290 | 6055 B |
| docs/WORKFLOW.md | workflows/shared.md | B3E2842E | 10174 B |
| docs/DEV-GUIDE.md | prompts/dev.md | 66F6C57E | 9952 B |
| docs/DESIGN-GUIDE.md | prompts/tech-lead.md | 5DC194DF | 6153 B |
| docs/TRIAGE-GUIDE.md | prompts/triage.md | D118344C | 4884 B |
| docs/BUGFIX-GUIDE.md | prompts/bugfix.md | 63305F8A | 3210 B |
| docs/INSTALL.md | examples/ | AD611053 | 2963 B |
| docs/design/ARCHITECTURE.md | tech-lead.md | 5DC194DF | 5551 B |

---

## 6. Version History

| Date | Version | Change |
|------|---------|--------|
| 2026-04-24 | **v6.0** | Two-layer workflow + DOMGEN + directory reorg + background markers + snapshot freeze |
| 2026-04-23 | v5.0 | Client-vendor separation, doc classification, pure rules |
| 2026-04-22 | v4.0 | Three-layer gate, process engine |
| 2026-04-21 | v3.7 | Multi-role collaboration, rule doc mode |

---

**Report updated**: 2026-04-24 19:07 | **Author**: OpenClaw AI