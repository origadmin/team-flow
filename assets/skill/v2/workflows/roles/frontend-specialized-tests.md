# Frontend Specialized Test Standards — team-flow v2 (beads-native)

> **Role**: Frontend Dev / QA Engineer
> **Load**: `{TEAM_PATH}/workflows/roles/frontend-specialized-tests.md`
> **Prerequisite**: Read `{TEAM_PATH}/workflows/shared.md` first
> **Tech Stack**: Bun + Rsbuild + Jest + React + Playwright

---

## 1. Component Rendering Tests

**Trigger**: Any component with visual output (UI components, page components, layout components).

**Checklist**:
- [ ] Renders without crash with default props
- [ ] Renders correctly with data props
- [ ] Loading state shows skeleton/spinner
- [ ] Empty state shows "no data" message
- [ ] Error state shows error message
- [ ] All conditional branches render (v-if / ternary / &&)
- [ ] Children render correctly

**Test Pattern**:

```typescript
import { render, screen } from '@testing-library/react';
import { ComponentName } from './ComponentName';

describe('ComponentName', () => {
    it('renders with default props', () => {
        render(<ComponentName />);
        expect(screen.getByText('Expected Text')).toBeInTheDocument();
    });

    it('shows loading state', () => {
        render(<ComponentName isLoading />);
        expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    });

    it('shows empty state', () => {
        render(<ComponentName data={[]} />);
        expect(screen.getByText(/no data/i)).toBeInTheDocument();
    });
});
```

---

## 2. User Interaction Tests

**Trigger**: Any component with user actions (buttons, forms, modals, drag-drop).

**Checklist**:
- [ ] Click handler called with correct arguments
- [ ] Form input updates state
- [ ] Form submit triggers API call + shows loading
- [ ] Cancel/Close resets form state
- [ ] Delete shows confirmation dialog
- [ ] Keyboard shortcuts work (Enter to submit, Esc to close)
- [ ] Disabled state prevents interaction

**Test Pattern**:

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

describe('ComponentName interactions', () => {
    it('calls onSubmit with form data', async () => {
        const onSubmit = jest.fn();
        render(<ComponentName onSubmit={onSubmit} />);

        await userEvent.type(screen.getByLabelText('Name'), 'test');
        await userEvent.click(screen.getByRole('button', { name: /submit/i }));

        expect(onSubmit).toHaveBeenCalledWith({ name: 'test' });
    });
});
```

---

## 3. Hook Logic Tests

**Trigger**: Any custom hook with state, side effects, or complex logic.

**Checklist**:
- [ ] Returns correct initial state
- [ ] State updates correctly on action
- [ ] Side effects triggered at right time
- [ ] Cleanup on unmount (event listeners, subscriptions, timers)
- [ ] Edge cases: null/undefined input, empty array, concurrent calls
- [ ] Error handling: API errors, network failures

**Test Pattern**:

```typescript
import { renderHook, act } from '@testing-library/react';
import { useCustomHook } from './useCustomHook';

describe('useCustomHook', () => {
    it('returns initial state', () => {
        const { result } = renderHook(() => useCustomHook());
        expect(result.current.data).toBeNull();
        expect(result.current.isLoading).toBe(false);
    });

    it('updates state on action', async () => {
        const { result } = renderHook(() => useCustomHook());
        await act(async () => {
            await result.current.fetchData();
        });
        expect(result.current.data).toBeDefined();
    });
});
```

---

## 4. API Integration Tests (MSW)

**Trigger**: Any component or hook that makes API calls.

**Checklist**:
- [ ] Successful API call renders data
- [ ] Loading state shown during API call
- [ ] Error state shown on API failure
- [ ] Retry mechanism works
- [ ] Cache invalidation after mutation
- [ ] Optimistic updates (if applicable)
- [ ] Pagination works correctly
- [ ] CRUD flow: Create → Read → Update → Delete

**MSW Setup Pattern**:

```typescript
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';

const handlers = [
    http.get('/api/v1/resource', () => {
        return HttpResponse.json({ list: [], total: 0 });
    }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

---

## 5. Route Guard Tests

**Trigger**: Any route with authentication or authorization requirements.

**Checklist**:
- [ ] Unauthenticated user redirected to login
- [ ] Authenticated user can access protected route
- [ ] Non-admin user redirected from admin route
- [ ] Already-logged-in user redirected from login page
- [ ] Redirect param preserved after login
- [ ] Route params passed correctly to page component

**Test Pattern**:

```typescript
import { render, screen } from '@testing-library/react';
import { createMemoryRouter, RouterProvider } from '@tanstack/react-router';

describe('Route guards', () => {
    it('redirects to login when not authenticated', async () => {
        const router = createMemoryRouter({
            routeTree,
            initialEntries: ['/admin'],
        });
        render(<RouterProvider router={router} />);
        await waitFor(() => {
            expect(router.state.location.pathname).toBe('/auth/signin');
        });
    });
});
```

---

## 6. Responsive Layout Tests

**Trigger**: Components with responsive behavior (sidebar collapse, mobile layout, grid changes).

**Checklist**:
- [ ] Desktop layout (>= 1200px): full sidebar, full grid
- [ ] Tablet layout (768-1199px): collapsed sidebar, adjusted grid
- [ ] Mobile layout (< 768px): hidden sidebar, stacked layout
- [ ] Breakpoint transitions smooth
- [ ] Content not clipped or overflowed at any breakpoint

**Test Pattern**:

```typescript
import { render } from '@testing-library/react';

describe('ResponsiveLayout', () => {
    it('shows collapsed sidebar on tablet', () => {
        Object.defineProperty(window, 'innerWidth', { value: 900 });
        render(<Layout />);
        expect(screen.getByTestId('sidebar')).toHaveClass('collapsed');
    });
});
```

---

## 7. Accessibility Tests

**Trigger**: All interactive components (buttons, forms, modals, navigation).

**Checklist**:
- [ ] All interactive elements have accessible names
- [ ] ARIA roles correct (button, dialog, navigation, etc.)
- [ ] Keyboard navigation works (Tab, Enter, Escape, Arrow keys)
- [ ] Focus management correct (modal trap, return focus on close)
- [ ] Screen reader announcements for dynamic content
- [ ] Color contrast meets WCAG AA (4.5:1 for text)
- [ ] Form labels associated with inputs
- [ ] Error messages linked to form fields

**Test Pattern**:

```typescript
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

describe('Accessibility', () => {
    it('modal traps focus and returns on close', async () => {
        render(<ComponentWithModal />);
        await userEvent.click(screen.getByRole('button', { name: /open/i }));
        const modal = screen.getByRole('dialog');
        expect(modal).toBeInTheDocument();
        await userEvent.keyboard('{Escape}');
        expect(modal).not.toBeInTheDocument();
    });
});
```

---

## 8. State Management Tests

**Trigger**: Components using AuthContext, TanStack Query, or complex local state.

**Checklist**:
- [ ] AuthContext provides correct user state
- [ ] Login updates auth state and persists token
- [ ] Logout clears auth state and removes token
- [ ] TanStack Query cache invalidation after mutation
- [ ] Optimistic updates roll back on error
- [ ] Stale data refetched on window focus (if configured)

---

## 9. i18n Tests

**Trigger**: Components with translatable text.

**Checklist**:
- [ ] Default language renders correctly
- [ ] Language switch updates all text
- [ ] Missing translations show key (not blank)
- [ ] Date/number formatting locale-aware

---

## 10. Theme Tests

**Trigger**: Components using theme tokens or ThemeProvider.

**Checklist**:
- [ ] Default theme applies correct CSS variables
- [ ] Theme switch updates CSS variables
- [ ] Theme preference persisted in localStorage
- [ ] All color tokens have sufficient contrast in both themes

---

## 11. Visual Regression Tests (Playwright)

**Trigger**: Any component with visual output that has design specifications or Figma references.

**Checklist**:
- [ ] Screenshot comparison against baseline
- [ ] No unexpected visual changes after code modification
- [ ] Responsive breakpoints captured
- [ ] Dark/light theme screenshots captured
- [ ] Animation states captured (if applicable)

**Test Pattern**:

```typescript
import { test, expect } from '@playwright/test';

test('component visual regression', async ({ page }) => {
    await page.goto('/components/button');
    await expect(page.locator('[data-testid="button-primary"]')).toHaveScreenshot('button-primary.png');
});

test('responsive layout visual regression', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/dashboard');
    await expect(page).toHaveScreenshot('dashboard-tablet.png');
});
```

---

## 12. E2E Flow Tests (Playwright)

**Trigger**: Any multi-page user flow (login → dashboard → action → result).

**Checklist**:
- [ ] Critical user journey completes end-to-end
- [ ] Cross-page state preserved correctly
- [ ] Error recovery flows work (retry, fallback)
- [ ] Concurrent user scenarios (if applicable)

**Test Pattern**:

```typescript
import { test, expect } from '@playwright/test';

test('user login and create resource flow', async ({ page }) => {
    await page.goto('/auth/signin');
    await page.fill('[name="email"]', 'test@example.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('/dashboard');

    await page.click('[data-testid="create-resource"]');
    await page.fill('[name="resource-name"]', 'Test Resource');
    await page.click('button[type="submit"]');

    await expect(page.locator('.toast')).toContainText('Created successfully');
});
```

---

## 13. Playwright MCP Integration for Automated UI Verification

> 自动测试工具配置见 `{TEAM_PATH}/prompts/qa-engineer.md` §Automated Test Tools

**Trigger**: Any component with visual output that requires automated UI verification after implementation or bug fix.

**Rule**: After completing frontend implementation, use Playwright MCP to verify key user flows before marking task as phase:verify.

---

## Trigger Judgment Output

```markdown
### Frontend Specialized Test Judgment

| Test Type | Triggered | Reason | Status |
|-----------|-----------|--------|--------|
| Component rendering | ✅/❌ | {hit/skip reason} | 🟢/⚪ |
| User interaction | ✅/❌ | ... | 🟢/⚪ |
| Hook logic | ✅/❌ | ... | 🟢/⚪ |
| API integration (MSW) | ✅/❌ | ... | 🟢/⚪ |
| Route guard | ✅/❌ | ... | 🟢/⚪ |
| Responsive layout | ✅/❌ | ... | 🟢/⚪ |
| Accessibility | ✅/❌ | ... | 🟢/⚪ |
| State management | ✅/❌ | ... | 🟢/⚪ |
| i18n | ✅/❌ | ... | 🟢/⚪ |
| Theme | ✅/❌ | ... | 🟢/⚪ |
| Visual regression (Playwright) | ✅/❌ | ... | 🟢/⚪ |
| E2E flow (Playwright) | ✅/❌ | ... | 🟢/⚪ |
| Playwright MCP verification | ✅/❌ | ... | 🟢/⚪ |
```

---

## beads Status Management

> beads 状态管理命令见 `{TEAM_PATH}/workflows/shared.md` §beads Status Management
