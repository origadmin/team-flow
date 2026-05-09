# 用户故事模板 — v2 beads-native

> **位置**: `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/USER_STORY.md`

---

## beads 关联

| 字段 | 内容 |
|------|------|
| Issue ID | `<beads-id>` |
| Feature ID | F{NNN} |

---

## 用户故事

```
作为 {角色}
我想要 {功能}
以便 {价值}
```

---

## 验收标准

```gherkin
Scenario: {场景名称}
  Given {前置条件}
  When {用户操作}
  Then {预期结果}
```

---

## 优先级

- [ ] Must（必须有）
- [ ] Should（应该有）
- [ ] Could（可以有）
- [ ] Won't（不做）

---

## 估算

| 指标 | 值 |
|------|-----|
| 故事点 | {SP} |
| 工期 | {days} |

---

## beads 状态更新

```bash
flow task update <id> --notes "User story defined: {story-summary}"
```
