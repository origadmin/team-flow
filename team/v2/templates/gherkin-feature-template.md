---
version: "2.0"
owner: "[PM / QA Engineer]"
lastUpdated: "2026-05-06"
---

# Gherkin 场景模板 — v2 beads-native

> **位置**: `{DOCS_INTERNAL}/features/{module}.feature`

---

## Feature 文件结构

```gherkin
@tag1 @tag2
Feature: {功能名称}
  "作为 [角色]"
  "我想要 [功能]"
  "以便 [业务价值]"

  Background:
    Given 系统已启动
    And 我已登录为管理员

  Scenario: {场景名称}
    Given {前置状态}
    When {用户行为}
    Then {预期结果}

  Scenario Outline: {数据驱动场景}
    Given <输入值>
    When 我执行操作
    Then 结果应为 <预期输出>

    Examples:
      | 输入值 | 预期输出 |
      | value1 | result1 |
      | value2 | result2 |
```

---

## Scenario 编写规范

| 关键字 | 用途 |
|--------|------|
| Given | 前置条件 |
| When | 用户行为 |
| Then | 预期结果 |
| And/But | 连接步骤 |

---

## 标签规范

| 标签 | 用途 |
|------|------|
| @p0 | 核心场景 |
| @p1 | 重要场景 |
| @smoke | 冒烟测试 |
| @regression | 回归测试 |

---

## beads 关联

```bash
flow tools beads update <id> --set-metadata feature_path="{DOCS_INTERNAL}/features/{module}.feature"
```
