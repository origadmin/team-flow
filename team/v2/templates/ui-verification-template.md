# B{NNN}-R{N}: UI Verification Report

> **Template**: ui-verification-template.md
> **Usage**: Frontend Bug fix MUST produce this document with screenshots as evidence.

## Metadata

| Field | Value |
|-------|-------|
| Bug ID | B{NNN}-R{N} |
| Bug Title | {title} |
| Verification Date | {YYYY-MM-DD} |
| Verifier | {AI role / human} |
| Dev Server URL | http://localhost:{port} |
| Browser | Chrome / Firefox / Safari |

---

## Screenshot Rules

### Save Path

```
{DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/screenshots/
```

### Naming Convention

```
B{NNN}-R{N}-{step_number}-{action}-{state}.png
```

| Part | Rule | Example |
|------|------|---------|
| B{NNN} | Bug ID | B099 |
| R{N} | R iteration | R1 |
| {step_number} | 3-digit step number | 001, 002, 003 |
| {action} | What user does | click-submit, type-email, navigate-dashboard |
| {state} | before / after / result | before, after, result |

**Examples**:
- `B099-R1-001-navigate-login-before.png` — Login page before fix
- `B099-R1-002-click-submit-result.png` — After clicking submit button
- `B099-R1-003-check-dashboard-result.png` — Dashboard page showing fix works

### Screenshot Count

| Bug Type | Minimum Screenshots | Required |
|----------|-------------------|----------|
| UI rendering Bug | 2 | before-fix + after-fix |
| Interaction Bug (click/input/submit) | 3 | before-action + during-action + after-action |
| Navigation/routing Bug | 2 | wrong-page + correct-page |
| Data display Bug | 2 | wrong-data + correct-data |
| Form validation Bug | 3 | empty-form + invalid-input + valid-submit |
| Multi-step flow Bug | N+1 | each-step + final-result (N = number of steps) |

### Screenshot Quality

- Resolution: 1280x720 minimum (desktop), 375x812 (mobile if applicable)
- Format: PNG (not JPEG, not SVG)
- Content: Full page or focused area with enough context
- Must show: URL bar (for navigation verification), element state (for interaction verification)

---

## Verification Steps

### Step 1: Dev Server Startup

| Item | Content |
|------|---------|
| Command | `bun run dev` |
| Expected | Server starts without errors |
| Actual | {result} |
| Screenshot | N/A (terminal output) |

**Terminal Output**:
```
{paste dev server startup output here}
```

---

### Step 2: Navigate to Bug Page

| Item | Content |
|------|---------|
| URL | {full URL, e.g. http://localhost:3000/admin/users} |
| Navigation Path | {how to get there, e.g. Sidebar > User Management} |
| Expected | Page renders without white screen / console errors |
| Actual | {result} |
| Screenshot | `B{NNN}-R{N}-001-navigate-{page}-result.png` |

**Page Content Check**:

| Check Item | Expected | Actual | Pass |
|-----------|----------|--------|------|
| Page title | {expected title} | {actual} | ✅/❌ |
| Key elements visible | {list elements} | {actual} | ✅/❌ |
| No console errors | 0 errors | {count} | ✅/❌ |
| No white screen | Page fully rendered | {actual} | ✅/❌ |

---

### Step 3: Verify Bug Area Content

| Item | Content |
|------|---------|
| Bug Area | {which component/section on the page} |
| Expected Content | {what should be displayed} |
| Actual Content | {what is actually displayed} |
| Screenshot | `B{NNN}-R{N}-002-check-{area}-result.png` |

**Content Detail Check**:

| Check Item | Expected | Actual | Pass |
|-----------|----------|--------|------|
| Text content | {expected text} | {actual} | ✅/❌ |
| Data values | {expected values} | {actual} | ✅/❌ |
| Component state | {expected state} | {actual} | ✅/❌ |
| Styling | {expected appearance} | {actual} | ✅/❌ |

---

### Step 4: Verify Interaction Behavior

| Item | Content |
|------|---------|
| Interaction Type | click / type / submit / drag / hover / scroll |
| Target Element | {CSS selector or data-testid} |
| Expected Behavior | {what should happen} |
| Actual Behavior | {what actually happens} |
| Screenshot (before) | `B{NNN}-R{N}-003-{action}-{target}-before.png` |
| Screenshot (after) | `B{NNN}-R{N}-004-{action}-{target}-after.png` |

**Interaction Steps**:

| Step | Action | Target | Expected Result | Actual Result | Pass |
|------|--------|--------|----------------|---------------|------|
| 1 | {click/type/...} | {element} | {expected} | {actual} | ✅/❌ |
| 2 | {click/type/...} | {element} | {expected} | {actual} | ✅/❌ |
| 3 | {click/type/...} | {element} | {expected} | {actual} | ✅/❌ |

---

### Step 5: Bug Symptom Verification

| Item | Content |
|------|---------|
| Original Bug Symptom | {what was broken} |
| Reproduction Steps | {steps that triggered the bug} |
| Expected After Fix | {bug symptom should NOT appear} |
| Actual After Fix | {bug symptom gone or still present} |
| Screenshot | `B{NNN}-R{N}-005-verify-fix-result.png` |

**Reproduction Attempt**:

| Step | Action | Expected (Fixed) | Actual | Pass |
|------|--------|------------------|--------|------|
| 1 | {original trigger step 1} | {correct behavior} | {actual} | ✅/❌ |
| 2 | {original trigger step 2} | {correct behavior} | {actual} | ✅/❌ |
| 3 | {original trigger step 3} | {correct behavior} | {actual} | ✅/❌ |

**Bug Status**: ✅ Fixed / ❌ Still Present / ⚠️ Partially Fixed

---

### Step 6: Side Effect Check

| Item | Content |
|------|---------|
| Adjacent Features | {features near the bug area} |
| Navigation | {other pages still work} |
| Screenshot | `B{NNN}-R{N}-006-side-effect-result.png` |

| Check Item | Expected | Actual | Pass |
|-----------|----------|--------|------|
| Adjacent feature 1 | Works normally | {actual} | ✅/❌ |
| Adjacent feature 2 | Works normally | {actual} | ✅/❌ |
| Navigation intact | All links work | {actual} | ✅/❌ |

---

## Screenshot Index

| # | Filename | Step | Description | Pass |
|---|----------|------|-------------|------|
| 1 | `B{NNN}-R{N}-001-navigate-{page}-result.png` | Step 2 | Page navigation result | ✅/❌ |
| 2 | `B{NNN}-R{N}-002-check-{area}-result.png` | Step 3 | Bug area content | ✅/❌ |
| 3 | `B{NNN}-R{N}-003-{action}-{target}-before.png` | Step 4 | Before interaction | ✅/❌ |
| 4 | `B{NNN}-R{N}-004-{action}-{target}-after.png` | Step 4 | After interaction | ✅/❌ |
| 5 | `B{NNN}-R{N}-005-verify-fix-result.png` | Step 5 | Bug symptom verification | ✅/❌ |
| 6 | `B{NNN}-R{N}-006-side-effect-result.png` | Step 6 | Side effect check | ✅/❌ |

**Total Screenshots**: {N} (minimum: {minimum based on bug type})

---

## Verification Result

| Category | Result | Details |
|----------|--------|---------|
| Dev server | ✅/❌ | {details} |
| Page rendering | ✅/❌ | {details} |
| Content correctness | ✅/❌ | {details} |
| Interaction behavior | ✅/❌ | {details} |
| Bug symptom gone | ✅/❌ | {details} |
| No side effects | ✅/❌ | {details} |
| Screenshots complete | ✅/❌ | {N} screenshots saved |

**Overall**: ✅ PASS / ❌ FAIL

---

## beads Status Update

```bash
flow tools beads update <id> --notes "UI_VERIFICATION: B{NNN}-R{N} {PASS/FAIL}. Screenshots: {N}. Bug symptom: {gone/present}"
```
