# {Fxxx}: {Feature Name} — Frontend Test Coverage Declaration

## beads 关联

| 字段 | 值 |
|------|-----|
| Issue ID | `<beads-id>` |
| task ID | F{xxx} |
| Phase | phase:implement → phase:verify |
| Subsystem | subsystem:frontend |

**Feature Version**: v{version}
**Related Requirement**: `{docs_internal}/requirements/Fxxx/`
**Test Code Path**: `{PROJECT_PATH}/web/tests/features/F{xxx}-{short-name}/`
**Tech Stack**: Bun + Rsbuild + Jest + React + Playwright

---

## Required Test Categories

> Minimum requirements for each Feature. All categories must be covered.

| # | Test Category | Coverage Status | Test File | Notes |
|---|--------------|----------------|-----------|-------|
| 1 | **Component rendering** | ☐ / ☑ | `component.test.tsx` | Renders correctly with various props |
| 2 | **User interaction** | ☐ / ☑ | `interaction.test.tsx` | Click/input/submit behavior |
| 3 | **Hook logic** | ☐ / ☑ | `hook.test.ts` | State transitions, return values, edge cases |
| 4 | **API integration (MSW)** | ☐ / ☑ | `integration.test.tsx` | Mock API → component flow |
| 5 | **Edge/boundary values** | ☐ / ☑ | `boundary.test.tsx` | Empty/null/long/special chars |
| 6 | **Error handling** (optional) | ☐ / ☑ | `error.test.tsx` | 4xx/5xx/network error states |
| 7 | **Responsive layout** (optional) | ☐ / ☑ | `responsive.test.tsx` | Breakpoint behavior |
| 8 | **Accessibility** (optional) | ☐ / ☑ | `a11y.test.tsx` | ARIA roles, keyboard navigation |

---

## 1. Component Rendering Tests

| # | Scenario | Verification Point | Coverage |
|---|----------|-------------------|----------|
| 1 | Default props | Renders without crash | ☐ |
| 2 | With data | Displays data correctly | ☐ |
| 3 | Loading state | Shows skeleton/spinner | ☐ |
| 4 | Empty state | Shows empty message | ☐ |
| 5 | Error state | Shows error message | ☐ |

---

## 2. User Interaction Tests

| # | Interaction | Expected Behavior | Coverage |
|---|------------|-------------------|----------|
| 1 | Button click | Handler called with correct args | ☐ |
| 2 | Form input | State updates correctly | ☐ |
| 3 | Form submit | API call triggered, loading shown | ☐ |
| 4 | Cancel/Close | Modal closed, form reset | ☐ |
| 5 | Delete confirm | Confirmation dialog shown | ☐ |

---

## 3. Hook Logic Tests

| # | Hook Function | Input | Expected Output | Coverage |
|---|--------------|-------|----------------|----------|
| 1 | Initial state | Default | Correct initial values | ☐ |
| 2 | State update | Action | Correct new state | ☐ |
| 3 | Edge case | Null/undefined | Graceful handling | ☐ |
| 4 | Cleanup | Unmount | Resources released | ☐ |

---

## 4. API Integration Tests (MSW)

```typescript
import { http, HttpResponse } from 'msw';

export const handlers = [
    http.get('/api/v1/{resource}', () => {
        return HttpResponse.json({ list: [], total: 0 });
    }),
    http.post('/api/v1/{resource}', async ({ request }) => {
        const body = await request.json();
        return HttpResponse.json({ id: 1, ...body }, { status: 201 });
    }),
];
```

**CRUD Flow Checklist**:

| Step | Operation | API | Verification Point |
|------|-----------|-----|-------------------|
| 1 | Create | POST /api/v1/{resource} | Returns id, correct initial state |
| 2 | Read | GET /api/v1/{resource}/{id} | All fields match creation |
| 3 | List | GET /api/v1/{resource} | New record in list |
| 4 | Update | PATCH /api/v1/{resource}/{id} | Update succeeds |
| 5 | List after update | GET /api/v1/{resource} | Updated values in list |
| 6 | Delete | DELETE /api/v1/{resource}/{id} | Returns 204 |
| 7 | Verify gone | GET /api/v1/{resource}/{id} | Returns 404 |

---

## 5. Edge/Boundary Value Tests

| Field Type | Boundary Value | Expected Result | Coverage |
|-----------|---------------|----------------|----------|
| string | Empty `""` | | ☐ |
| string | null | | ☐ |
| string | Long text (>1000 chars) | | ☐ |
| string | Special chars (`<>"'&`) | | ☐ |
| number | 0 | | ☐ |
| number | Negative | | ☐ |
| number | Max overflow | | ☐ |
| array | Empty `[]` | | ☐ |
| array | Large (>100 items) | | ☐ |

---

## 6. API Coverage Checklist

| API Path | Method | Test Category | Coverage |
|---------|--------|--------------|----------|
| /api/v1/{resource} | GET | List query | ☐ |
| /api/v1/{resource} | POST | Create | ☐ |
| /api/v1/{resource}/{id} | GET | Detail | ☐ |
| /api/v1/{resource}/{id} | PATCH | Update | ☐ |
| /api/v1/{resource}/{id} | DELETE | Delete | ☐ |

---

## Test Execution Record

| Date | Executor | Categories Covered | Pass | Fail | Bug |
|------|----------|-------------------|------|------|-----|
| {YYYY-MM-DD} | {dev} | 0/8 | 0 | 0 | - |

---

## Related Files

| Type | Path |
|------|------|
| Requirement | `_docs/.../requirements/Fxxx/SPEC.md` |
| API Contract | `_docs/.../requirements/Fxxx/R3_API_CONTRACT.md` |
| Test code | `{PROJECT_PATH}/web/tests/features/F{xxx}-{short-name}/` |
| MSW handlers | `{PROJECT_PATH}/web/tests/mocks/handlers.ts` |

---

## beads 状态更新

完成测试后执行：

```bash
flow tools beads update <issue-id> --notes "TEST_COVERAGE: {x}/8 categories covered, {pass}/{fail} results"
flow tools beads update <issue-id> --add-label phase:verify --remove-label phase:implement
```
