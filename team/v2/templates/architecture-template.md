# 架构设计模板 — v2 beads-native

> **位置**: `{DOCS_INTERNAL}/design/{feature}/ARCHITECTURE.md`

---

## beads 关联

| 字段 | 内容 |
|------|------|
| Issue ID | cms-xxx |
| Feature | {feature-name} |

---

## 1. 概述

### 1.1 背景
{为什么需要这个架构设计}

### 1.2 目标
{架构设计要达成什么目标}

### 1.3 范围
{包含什么、不包含什么}

---

## 2. 架构视图

### 2.1 系统上下文
```
┌─────────────────────────────────────────────┐
│              External Systems               │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐      │
│  │  User   │  │  Admin  │  │   API   │      │
│  └────┬────┘  └────┬────┘  └────┬────┘      │
└───────┼────────────┼────────────┼───────────┘
        │            │            │
        ▼            ▼            ▼
┌─────────────────────────────────────────────┐
│              Our System                     │
└─────────────────────────────────────────────┘
```

### 2.2 模块划分
| 模块 | 职责 | 依赖 |
|------|------|------|
| {module} | {responsibility} | {deps} |

---

## 3. 技术选型

| 组件 | 选型 | 理由 |
|------|------|------|
| {component} | {choice} | {reason} |

---

## 4. 接口契约

```
POST /api/v1/{resource}
Request: { ... }
Response: { ... }
```

---

## 5. 非功能需求

| 需求 | 目标 | 策略 |
|------|------|------|
| 性能 | {target} | {strategy} |
| 安全 | {target} | {strategy} |

---

## beads 状态更新

```bash
bd update <id> --set-metadata doc_path="{DOCS_INTERNAL}/design/{feature}/ARCHITECTURE.md"
```
