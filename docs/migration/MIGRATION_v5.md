<!-- SNAPSHOT — Frozen as of 2026-04-24 — DO NOT REGENERATE -->
<!-- This document records decisions made during v5.0 upgrade -->
<!-- For current architecture, see docs/design/ARCHITECTURE.md -->

# Team Workflow 框架升级文档

> **升级日期**: 2026-04-23~24

---

## v5.5 (2026-04-24) — AI 文件空间优化

### 变更原因

`_team/` 下大量文件含说明性内容，AI 读全部文件时浪费 token + 上下文污染。

### 删除的文件（从未被 AI prompt 引用）

| 文件 | 大小 | 原因 |
|------|------|------|
| `workflows/standards.md` | 13.8KB | 人读说明文档，无 AI 引用 |
| `workflows/dev-workflow.md` | 16.8KB | 人读流程说明，无 AI 引用 |
| `workflows/framework-workflow.md` | 8.8KB | 人读流程说明，无 AI 引用 |
| `workflows/shared-implementation.md` | 6.7KB | 人读规范说明，无 AI 引用 |
| `workflows/output-standards.md` | 3.2KB | 明确标注废弃，0 AI 引用 |
| `prompts/team-input-matcher.md` | 1.9KB | v3.3 已废弃 |
| `MIGRATION_v4.md` | 9.6KB | v4 历史文件 |
| `REFACTOR_SUMMARY.md` | 8.7KB | v5 重构历史文件 |

**共删除**: ~69.5KB 说明性文件

### 新增的文件

| 文件 | 大小 | 用途 |
|------|------|------|
| `.team/issues.md` | ~0.5KB | 框架问题独立追踪（从 task-pool.md 分离） |

### 更新文件模板

| 文件 | 变更 |
|------|------|
| `.team/task-pool.md` | v6.0：精简列数（12→10），分离框架问题到 issues.md，删除空表 |
| `.team/backlog.md` | v6.0：增加状态/Blocker 列，任务详情移到引用文档 |
| `workflows/shared.md` | v5.5：新增 AI 运营文件更新规则 + _docs AI-EXEC-ANCHOR 规范 |

### AI 读取规则

```
读取 _team/ 文件时
    │
    ├── prompts/*.md         → 全部读取（角色执行规则）
    ├── workflows/shared.md  → 全部读取（共享门禁规则）
    ├── workflows/roles/*.md → 各自角色对应文件
    ├── SKILL.md             → 全部读取（框架入口）
    ├── README.md             → 人读，可跳过
    ├── ARCHITECTURE.md       → 人读，可跳过
    └── MIGRATION_*.md        → 历史文件，可跳过
```

---

## v5.4 (2026-04-24) — 文件空间定义

- `.team/` = AI 运营文件（task-pool, backlog, project）
- `_docs/` = 共享知识库（AI 写 → 人读 → 人反馈 → AI 执行基准）
- backlog.md 从 _docs 迁入 .team/
- 文档同步绑阶段门禁

---

## v5.0 (2026-04-23) — 甲方-乙方分离

核心变更见 `MIGRATION_v5.md` 完整内容。

---

## AI 规则文件编写原则

AI 规则文件只保留：
1. **判断树** — 收到输入后怎么做
2. **执行模板** — 输出什么格式
3. **禁止清单** — 明确不能做什么
4. **触发条件** — 什么时候做什么

删除所有：背景介绍、架构图、对比表、"为什么"段落。
