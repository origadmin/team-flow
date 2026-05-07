# Frontend Specialized Test Standards — v2 beads-native

> **前端 10 类专项测试**

---

## 1. Component Rendering Tests

**Trigger**: Any component with visual output.

**Checklist**:
- [ ] Renders without crash
- [ ] Loading state shows skeleton
- [ ] Empty state shows "no data"
- [ ] Error state shows error.

---

## 2. User Interaction Tests

**Trigger**: Any component with user actions.

**Checklist**:
- [ ] Click handler called correctly
- [ ] Form input updates state
- [ ] Form submit triggers API call
- [ ] Delete shows confirmation.

---

## 3. Hook Logic Tests

**Trigger**: Any custom hook.

**Checklist**:
- [ ] Returns correct initial state
- [ ] State updates correctly
- [ ] Cleanup on unmount.

---

## 4. API Integration Tests (MSW)

**Trigger**: Any component making API calls.

**Checklist**:
- [ ] Successful API call renders data
- [ ] Loading state shown
- [ ] Error state shown
- [ ] CRUD flow works.

---

## 5. Route Guard Tests

**Trigger**: Any route with auth/authorization.

**Checklist**:
- [ ] Unauthenticated → redirect
- [ ] Authenticated can access
- [ ] Redirect param preserved.

---

## 6. Responsive Layout Tests

**Trigger**: Components with responsive behavior.

**Checklist**:
- [ ] Desktop layout (>= 1200px)
- [ ] Tablet layout (768-1199px)
- [ ] Mobile layout (< 768px).

---

## 7. Accessibility Tests

**Trigger**: Any interactive component.

**Checklist**:
- [ ] ARIA roles correct
- [ ] Keyboard navigation works.

---

## 8. State Management Tests

**Trigger**: Components using global state.

**Checklist**:
- [ ] State updates propagate
- [ ] Cache invalidation works.

---

## 9. Internationalization Tests

**Trigger**: Components with i18n.

**Checklist**:
- [ ] Language switching works
- [ ] Missing translations handled.

---

## 10. Theme Tests

**Trigger**: Components with theming.

**Checklist**:
- [ ] Theme switching works
- [ ] CSS variables applied.
