# PRD 模板 — v2 beads-native

> **位置**: `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/PRD.md`

---

## beads 关联

| 字段 | 内容 |
|------|------|
| Issue ID | `<beads-id>` |
| Feature ID | F{NNN} |

---

## 1. 背景

### 1.1 问题陈述
{当前存在什么问题}

### 1.2 目标用户
{谁会使用这个功能}

### 1.3 业务价值
{为什么需要做这个功能}

---

## 2. 功能范围

### 2.1 功能列表
| 功能 | 优先级 | 说明 |
|------|--------|------|
| {feature} | Must/Should | {description} |

### 2.2 不包含
{明确不做什么}

---

## 3. 验收标准

### 3.1 用户故事
```
作为 {角色}
我想要 {功能}
以便 {价值}
```

### 3.2 AC（Given/When/Then）
```gherkin
Scenario: {场景名称}
  Given {前置条件}
  When {用户操作}
  Then {预期结果}
```

---

## 4. 非功能需求

| 需求 | 目标 |
|------|------|
| 性能 | {target} |
| 安全 | {target} |

---

## 5. 里程碑

| 阶段 | 内容 | 时间 |
|------|------|------|
| 设计 | {deliverable} | {date} |
| 开发 | {deliverable} | {date} |
| 测试 | {deliverable} | {date} |

---

## beads 状态更新

```bash
flow tools beads update <id> --set-metadata doc_path="{DOCS_INTERNAL}/requirements/F{NNN}-{name}/PRD.md"
```
