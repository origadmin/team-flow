# B{xxx}-R{n}: {Bug Title} — Frontend Reproduction Test Case

**Severity**: Critical / Major / Minor
**Fix Version**: v{version}
**Report Date**: {YYYY-MM-DD}
**Reporter**: {reporter}
**Related Report**: `{docs_internal}/reports/bugs/Bxxx-Rx/`
**Test Code Path**: `{PROJECT_PATH}/web/tests/bugs/B{xxx}-{short-name}/`
**Tech Stack**: Bun + Rsbuild + Jest + React + Playwright

---

## 1. Bug Overview

| Field | Content |
|-------|---------|
| Bug ID | B{xxx}-R{n} |
| Title | {One-line description} |
| Affected Area | {Module/feature affected} |
| Reproduction Rate | 100% / Frequent / Occasional / Hard to reproduce |

---

## 2. Reproduction Steps

### 2.1 Prerequisites

- [ ] Logged in as {role}
- [ ] {Resource} exists
- [ ] Browser: Chrome / Firefox / Safari

### 2.2 Reproduction Flow

| Step | Action | Expected Result | Actual Result |
|------|--------|----------------|---------------|
| 1 | {Action1} | {Expected1} | ❌ {Actual1} |
| 2 | {Action2} | {Expected2} | ❌ {Actual2} |
| 3 | {Action3} | {Expected3} | ❌ {Actual3} |

### 2.3 Environment Info

| Field | Value |
|-------|-------|
| OS | Windows / macOS / Linux |
| Browser | Chrome / Firefox / Safari |
| Version | v{version} |
| API Version | /api/v1/... |

---

## 3. Root Cause Analysis (Summary)

> ⚠️ Full RCA must be completed before fix. See RCA.md. This section is a brief summary.

{Root cause description}

---

## 4. Fix Content

| Type | File/Module | Change Description |
|------|-----------|-------------------|
| Code fix | `{file}` | {Fix description} |
| Config change | `{file}` | {Change description} |

---

## 5. Regression Test Cases

> ⚠️ **Must execute after fix** to prevent recurrence

### 5.1 Reproduction Scenario Regression

| Case ID | Description | Expected Result | Coverage |
|---------|------------|----------------|----------|
| B{xxx}-RT-001 | Follow reproduction steps | Bug fixed ✅ | ☐ |
| B{xxx}-RT-002 | Boundary: {scenario} | {Expected} | ☐ |

### 5.2 Related Feature Regression

| Case ID | Description | Expected Result | Coverage |
|---------|------------|----------------|----------|
| B{xxx}-RG-001 | {Related feature 1} | Normal ✅ | ☐ |
| B{xxx}-RG-002 | {Related feature 2} | Normal ✅ | ☐ |

### 5.3 Frontend-Specific Regression

| Check Item | Coverage |
|-----------|----------|
| Original reproduction scenario passes | ☐ |
| Related components not affected | ☐ |
| Responsive layout not broken | ☐ |
| Console errors cleared | ☐ |
| Loading/error states still work | ☐ |

---

## 6. Component Regression Tests

| Component | Test Type | Verification Point | Coverage |
|-----------|----------|-------------------|----------|
| {ComponentName} | Unit test | Renders correctly after fix | ☐ |
| {ComponentName} | Interaction test | User actions work correctly | ☐ |
| {PageName} | Integration test | Page flow works end-to-end | ☐ |

---

## 7. API Regression Tests

| API | Method | Verification Point | Coverage |
|-----|--------|-------------------|----------|
| /api/v1/{resource} | GET | {Verification point} | ☐ |
| /api/v1/{resource}/{id} | PATCH | {Verification point} | ☐ |

---

## 8. MSW Mock Regression

> If the bug involves API response handling, verify MSW mock covers the edge case.

| Scenario | MSW Handler | Coverage |
|----------|------------|----------|
| Normal response | `http.get('/api/v1/...')` → 200 | ☐ |
| Error response | `http.get('/api/v1/...')` → 500 | ☐ |
| Empty response | `http.get('/api/v1/...')` → `{ list: [] }` | ☐ |
| Slow response | `http.get('/api/v1/...')` → delay 5000ms | ☐ |

---

## 9. Verification Record

| Date | Executor | Test Type | Result | Notes |
|------|----------|----------|--------|-------|
| {YYYY-MM-DD} | {dev} | Reproduction regression | ✅ Pass | - |
| {YYYY-MM-DD} | {qa} | Full regression | ✅ Pass | - |

### 9.1 Sign-off

| Role | Name | Date | Sign |
|------|------|------|------|
| Developer | | | |
| QA | | | |

---

## 10. Related Files

| Type | Path |
|------|------|
| Bug report | `{DOCS_INTERNAL}/reports/bugs/Bxxx-Rx/RCA.md` |
| Test code | `{PROJECT_PATH}/web/tests/bugs/B{xxx}-{short-name}/` |
| MSW handlers | `{PROJECT_PATH}/web/tests/mocks/handlers.ts` |
