# UI 设计模板 — v2 beads-native

> **位置**: `{DOCS_INTERNAL}/design/{feature}/UI_DESIGN.md`

---

## beads 关联

| 字段 | 内容 |
|------|------|
| Issue ID | cms-xxx |
| Feature | {feature-name} |

---

## 1. 设计概述

### 1.1 设计目标
{这个设计要解决什么问题}

### 1.2 设计原则
{遵循什么设计原则}

---

## 2. 设计稿

### 2.1 页面清单
| 页面 | 状态 | 链接 |
|------|------|------|
| {page} | Draft/Review/Done | {figma-link} |

### 2.2 组件复用
| 组件 | 来源 | 说明 |
|------|------|------|
| {component} | shadcn/ui | {usage} |

---

## 3. 设计规范

### 3.1 颜色
| 用途 | Token | 值 |
|------|-------|-----|
| 主色 | --primary | #xxx |
| 背景 | --background | #xxx |

### 3.2 字体
| 用途 | Token | 值 |
|------|-------|-----|
| 标题 | --font-heading | 24px |
| 正文 | --font-body | 14px |

---

## 4. 交互说明

| 元素 | 触发 | 行为 |
|------|------|------|
| {element} | click | {action} |

---

## beads 状态更新

```bash
bd update <id> --set-metadata doc_path="{DOCS_INTERNAL}/design/{feature}/UI_DESIGN.md"
```
