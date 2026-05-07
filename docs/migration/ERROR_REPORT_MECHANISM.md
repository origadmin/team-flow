# Error Report & Lesson Learned Mechanism

> **Version**: v2.0 | **Date**: 2026-04-24 | **Status**: Agreed

---

## 1. Background

_team rules define the standard workflow (Triage → Tech Lead → Dev → QA → Release).
But several gaps were identified during multi-AI collaborative development:

| Gap | Description |
|-----|-------------|
| No test stratification | Every test runs full suite, should have incremental vs full |
| No error report mechanism | Errors happen but no structured recording |
| No lesson-learned flow | AI repeats same errors, no rule supplement process |
| No change scope control | Multi-AI may cause out-of-scope changes |
| No error propagation handling | A's change breaks B, unclear who fixes it |
| No Git workflow for multi-AI | Manual commit, no structured process |

This document records the agreed design for: **Error Report Mechanism** and **Lesson-Learned Flow**.

---

## 2. Error Report Mechanism

### 2.1 Storage Path

```
{DOCS_INTERNAL}/reports/errors/       ← Flat directory, no sub-directories
├── INDEX.md                           ← Index file (tracks progress)
├── E00001-F001-001.md                 ← Error report files
├── E00002-F001-002.md
├── E00003-B003-R1-001.md
└── E00004-INT-M001-001.md
```

- `{DOCS_INTERNAL}` is resolved via environment variable (project-specific)
- Example for orig-cms: `_docs/orig-cms/reports/errors/`
- **Flat directory**: No sub-directories by type (feature/bugfix/integration)
- Classification is done via file metadata and INDEX.md

### 2.2 File Naming Convention

```
E{NNNNN}-{TASK_ID}-{SEQ}.md

E       = Error report prefix
NNNNN   = Global sequential number (5 digits, zero-padded)
TASK_ID = Associated task ID (e.g., F001, B003-R1, INT-M001)
SEQ     = Sequence within this task (3 digits, zero-padded)
```

Examples:
- `E00001-F001-001.md` — 1st error report, linked to F001, 1st error for F001
- `E00003-B003-R1-001.md` — 3rd error report, linked to B003-R1

### 2.3 INDEX.md Format

```markdown
# Error Report Index

## Progress
- Latest: E00004
- Processed up to: E00002
- Pending: E00003 ~ E00004

## Records

| # | File | TaskID | Type | Phase | Role | Date | Status |
|---|------|--------|------|-------|------|------|--------|
| E00001 | E00001-F001-001.md | F001 | Feature | Phase 2 | Dev-Backend | 2026-04-24 | Processed |
| E00002 | E00002-F001-002.md | F001 | Feature | Phase 3 | QA | 2026-04-24 | Processed |
| E00003 | E00003-B003-R1-001.md | B003-R1 | Bugfix | Phase 2 | Dev-Backend | 2026-04-24 | Pending |
| E00004 | E00004-INT-M001-001.md | M001 | Integration | R-Phase 1 | QA | 2026-04-24 | Pending |
```

- **Lesson Analyst** reads INDEX.md to find pending entries
- Only processes pending entries (saves token)
- After processing, updates status to "Processed"

### 2.4 Error Report Template

```markdown
# Error Report: {ID}

## Basic Info
- TaskID: {TASK_ID}
- TaskType: Feature / Bugfix / Integration
- Phase: Phase 2 / Phase 3 / R-Phase 1
- Role: {role} (e.g., Dev-Backend, Dev-Frontend, QA)
- Date: {YYYY-MM-DD}

## Error Details
- ErrorType: {type} (test_failure / build_failure / scope_drift / cross_module / api_mismatch)
- Symptom: {one-line description}
- RootCause: {analysis}
- Resolution: {how fixed}
- TimeCost: {XXmin}

## Impact Scope
- ChangedModule: {module}
- AffectedModule: {module} (or "None" if no out-of-scope impact)
- CrossModule: Yes / No
- CrossRole: Yes / No (e.g., frontend-backend API mismatch)
```

### 2.5 Trigger Mechanism

Error reports are triggered when:

| Trigger | Who Records | Example |
|---------|-------------|---------|
| Pipeline step failure | Dev | Step 4 test fails |
| Out-of-scope change detected | Dev | Need to modify file outside expected_scope |
| Integration test failure | QA | R-Phase 1 integration fails |
| Cross-module impact | QA or Dev | A's change breaks B's tests |

---

## 3. Classification System

### 3.1 Error Categories

| Category | Code | Definition | Judgment Criteria |
|----------|------|------------|-------------------|
| API | `api` | API-related issues | Interface definition, request/response format, field type, version compatibility |
| Database | `db` | Database-related issues | Field definition, index, migration, PK/FK, query performance |
| Code | `code` | Code implementation issues | Business logic, boundary conditions, concurrency, error handling |
| Scope | `scope` | Change scope issues | Modified files outside expected scope, missed required changes |
| Test | `test` | Testing-related issues | Test case coverage, test data, test environment |
| Architecture | `arch` | Architecture design issues | Module division, dependency direction, circular reference, extensibility |

### 3.2 Classification Priority

A single error may involve multiple categories:

1. **Primary category** = Most direct cause of the error
2. **Secondary category** = Contributing or related factors
3. **Each category gets its own rule** (if the rule differs)

Example:
- Error: API field name mismatch causing frontend to fail
- Primary: `api` (API field definition)
- Secondary: `code` (frontend didn't validate field existence)
- Result: Rule in `dev-backend-api.md` + Rule in `dev-frontend-code.md`

### 3.3 Unclassifiable Errors

```
If unable to classify → Tag as [NEEDS_CLASSIFICATION] in review result
→ PM / Tech Lead must supplement category definition
→ AI must NOT guess or force a category
→ Error stays in "Pending" status until classified
```

---

## 4. Lesson-Learned Flow

### 4.1 Two-Layer Mechanism

```
Layer 1 (Lightweight, every time):  Error Report → {DOCS_INTERNAL}/reports/errors/
Layer 2 (Periodic, milestone-level): Review → Structured Summary → AI Rules
```

### 4.2 Summary Flow

```
Milestone ends / Manual trigger
    ↓
PM / Tech Lead reviews error reports
    ↓
PM / Tech Lead produces STRUCTURED review result (not prose)
    ↓    - Which errors to codify (filter)
    ↓    - Which roles affected (categorize)
    ↓    - Which category (tag)
    ↓    - Rule direction (guide)
    ↓
Lesson Analyst reads structured review result
    ↓    - Can reference original error reports for detail
    ↓
Lesson Analyst produces AI rules by role+type
    ↓
Written to {DOCS_INTERNAL}/lessons/{role}-{type}.md
    ↓
Updates INDEX.md: mark processed entries
```

**Key distinction**: PM/Tech Lead does the **judgment** (filter + categorize + guide).
Lesson Analyst does the **translation** (format into AI-consumable rules).
Lesson Analyst must NOT make independent judgment on what to codify.

### 4.3 PM/Tech Lead Review Output Format

Must be structured (YAML frontmatter + structured Markdown), NOT prose:

```markdown
---
review_id: R-2026-04-24-001
milestone: M001
date: 2026-04-24
reviewer: Tech Lead
summary: "5 error reports in this milestone, 3 worth codifying"
---

## Error Filter Decision

| ErrorID | Decision | Reason |
|---------|----------|--------|
| E00001 | CODIFY | API version incompatibility, repeated pattern |
| E00002 | SKIP | One-off input error |
| E00003 | CODIFY | Field naming inconsistency, affects multiple modules |
| E00004 | CODIFY | Missing DB index causing performance issue |
| E00005 | SKIP | Test data issue, already fixed |

## Codify Details

### E00001 → API Version Incompatibility

- **Roles**: dev-backend, dev-frontend
- **PrimaryCategory**: api
- **SecondaryCategory**: (none)
- **RuleDirection**: "API upgrades must maintain backward compatibility, or explicitly mark breaking changes"
- **SourceReport**: E00001-F001-001.md

### E00003 → Field Naming Inconsistency

- **Roles**: dev-backend, dev-frontend, qa
- **PrimaryCategory**: api
- **SecondaryCategory**: db
- **RuleDirection**: "Unified field naming convention (snake_case), cross-module integration must verify field names"
- **SourceReport**: E00002-F001-002.md

### E00004 → Missing Database Index

- **Roles**: dev-backend
- **PrimaryCategory**: db
- **SecondaryCategory**: (none)
- **RuleDirection**: "Query fields must check index existence, tables >100k rows must have index or query optimization"
- **SourceReport**: E00003-B003-R1-001.md

## Review Notes

- Add "Version Compatibility" section to dev-backend-api.md
- Add "Naming Convention" section to common.md
- Add "Interface test must verify field names" rule to qa-test.md
```

**Required fields** (must not be empty):

| Field | Description |
|-------|-------------|
| ErrorID | Which error report |
| Decision | CODIFY / SKIP / NEEDS_CLASSIFICATION |
| Roles | Affected roles (at least one) |
| PrimaryCategory | api / db / code / scope / test / arch |
| SecondaryCategory | Optional, same options as Primary |
| RuleDirection | One-sentence guidance for rule direction |
| SourceReport | Link to original error report |

---

## 5. Lesson Analyst Role

### 5.1 Definition

- **Trigger**: Milestone end (auto) + Manual
- **Input**: PM/Tech Lead structured review result (Section 4.3 format)
- **Output**: AI rules written to `{DOCS_INTERNAL}/lessons/{role}-{type}.md`
- **Constraint**: Translation only, NO independent judgment on what to codify
- **Constraint**: If source review has `[NEEDS_CLASSIFICATION]`, stop and report back

### 5.2 Lesson Analyst Execution Flow

```
Step 1: Read PM/Tech Lead review result
    ↓
Step 2: For each CODIFY entry:
    a. Read original error report if more detail is needed
    b. Determine target file(s) based on Roles + Category
    c. Format rule using template (Section 5.3)
    ↓
Step 3: Write to corresponding lessons file(s)
    - File does not exist → Create with file header
    - File exists → Append to matching section (by ## heading)
    - Section does not exist → Create new section
    ↓
Step 4: Update INDEX.md (mark entries as Processed)
    ↓
Step 5: Output summary:

    Processed {N} lessons:
    - E00001 → dev-backend-api.md (Rule 5)
    - E00003 → common.md (Rule 3), qa-test.md (Rule 2)
    - E00004 → dev-backend-db.md (Rule 4)
```

### 5.3 Lesson Rule Template

Each rule written to lessons file must follow this format:

```markdown
### Rule {N}: {Short Title}

**Problem**: {One-line description of the error pattern}

**Rule**: {Rule body}

**Correct**:
- {Correct approach}

**Wrong Example**:
```go
// ❌ Wrong
func Bad() {
    // ...
}
```

**Correct Example**:
```go
// ✅ Correct
func Good() {
    // ...
}
```

**Applies When**: {When this rule must be followed}

**Source**: [{ErrorID}](../reports/errors/{ErrorID}.md)
```

---

## 6. Lessons Storage & Loading

### 6.1 Two-Level Common Structure

```
{DOCS_INTERNAL}/lessons/
├── common.md                    ← Global common (ALL roles read)
├── dev-common.md                ← Dev role common (Backend + Frontend Dev read)
├── dev-backend-common.md        ← Backend Dev common (API + DB + Code read)
├── dev-backend-api.md           ← Backend Dev API rules
├── dev-backend-db.md            ← Backend Dev DB rules
├── dev-backend-code.md          ← Backend Dev Code rules
├── dev-frontend-common.md       ← Frontend Dev common
├── dev-frontend-api.md          ← Frontend Dev API rules
├── dev-frontend-code.md         ← Frontend Dev Code rules
├── qa-common.md                ← QA role common
├── qa-test.md                  ← QA Test rules
├── tech-lead-common.md          ← Tech Lead role common
└── ...
```

### 6.2 Loading Order

```
Role starts task
    ↓
Load common.md (if exists, ALWAYS read)
    ↓
Load {role-group}-common.md (if exists, e.g., dev-common.md)
    ↓
Load {role}-{subtype}-common.md (if exists, e.g., dev-backend-common.md)
    ↓
Load {role}-{type}.md (by current task's type, may load multiple)
    ↓
None exist → Skip, zero token cost
```

### 6.3 Loading Examples

**Backend Dev doing API + DB task**:
1. `common.md`
2. `dev-common.md`
3. `dev-backend-common.md`
4. `dev-backend-api.md`
5. `dev-backend-db.md`
- Does NOT read: `dev-frontend-*`, `qa-*`, `tech-lead-*`

**QA doing integration test**:
1. `common.md`
2. `qa-common.md`
3. `qa-test.md`
- Does NOT read: `dev-*`, `tech-lead-*`

### 6.4 Rule Loading in Prompts

In `prompts/dev.md`, add lessons reference index:

```
## Lessons Reference

Before executing tasks, load lessons in this order:
1. {DOCS_INTERNAL}/lessons/common.md (if exists)
2. {DOCS_INTERNAL}/lessons/dev-common.md (if exists)
3. {DOCS_INTERNAL}/lessons/dev-{subtype}-common.md (if exists)
4. {DOCS_INTERNAL}/lessons/dev-{subtype}-{type}.md (if exists, by task type)

If file does not exist → skip (no recorded errors).
Pre-reading is mandatory when file exists (avoids retry token cost).
```

### 6.5 Role-Specific Loading Schedule

| Role | Load Before | Files (in order) |
|------|------------|------------------|
| Dev (Backend) | Phase 2 start | common → dev-common → dev-backend-common → dev-backend-{type} |
| Dev (Frontend) | Phase 2 start | common → dev-common → dev-frontend-common → dev-frontend-{type} |
| QA | Phase 3 start | common → qa-common → qa-test |
| Tech Lead | Phase 1 start | common → tech-lead-common |

---

## 7. DOMGEN Boundary

- DOMGEN is **_team internal** mechanism (AI rules → human-readable docs)
- DOMGEN has **nothing to do with** project-level documents
- Project design docs (SPEC.md, AC.md, R1/R2/R3) are created by roles (Tech Lead)
- Whether project-level human docs should use DOMGEN → **Needs separate review** (not in scope here)

---

## 8. Open Items (Not Yet Agreed)

| # | Topic | Status |
|---|-------|--------|
| 1 | Test stratification (incremental vs full) | Design not yet finalized |
| 2 | Change scope control (expected_scope in task-pool) | Design not yet finalized |
| 3 | Error propagation handling (A breaks B) | Design not yet finalized |
| 4 | Git workflow for multi-AI | Design not yet finalized |
| 5 | Lesson Analyst prompt definition | To be created in prompts/ |
