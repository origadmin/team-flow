---
name: team-beads
description: |
  team-flow framework task tracking using beads. Use when managing tasks,
  checking ready work, tracking progress, or session handoff.
version: "1.0.0"
tags: [task-management, beads, issue-tracking, multi-session]
---

# team-flow Beads Integration

## Trigger

当用户提到：任务、Task、task-pool、待办、Todo、进展、状态、Ready、Doing、Review、交接

## Prerequisites

```bash
bd --version  # Requires 1.0.3+
```

## Path Variables

```markdown
{PROJECT_PATH} = {PROJECT_PATH}/
{BEADS_DB}     = {PROJECT_PATH}/.beads/
```

## Session Protocol

### 启动时

```bash
cd {PROJECT_PATH}
bd ready --json
bd stats
```

报告格式：
> " beads 状态：X 个就绪任务，Y 个进行中，Z 个阻塞"

### 工作中

```bash
bd update <id> --claim --json
bd update <id> --notes "COMPLETED: xxx. IN PROGRESS: yyy. NEXT: zzz."
```

发现新工作时：

```bash
bd create "新任务标题" -t bug|feature|task -p 0-4 --json
bd dep add <新id> <当前id> --type discovered-from
```

### 结束时

```bash
bd update <id> --notes "COMPLETED: X. IN PROGRESS: Y. NEXT: Z. KEY DECISION: ..."
bd dolt push
```

## Issue Type Mapping

| task 类型 | Beads 类型 | 说明 |
|------------|-----------|------|
| bugfix | bug | 缺陷修复 |
| feature | feature | 新功能 |
| change | task | 架构变更 |
| analysis | task | 分析任务 |

## Priority Mapping

| task | Beads | 说明 |
|-------|-------|------|
| P0 | 0 | 阻塞主流程 |
| P1 | 1 | 重要但不紧急 |
| P2 | 2 | 常规任务 |
| P3 | 3 | 优化项 |

## Status Mapping

| task 状态 | Beads 状态 | CLI |
|------------|-----------|-----|
| Todo | open | default |
| Doing | in_progress | `bd update <id> --status in_progress` |
| Review | open + notes | 等待确认 |
| Archived | closed | `bd close <id> --reason "..."` |

## Notes Format (Session Handoff)

```
COMPLETED: 已完成的具体工作
IN PROGRESS: 当前进行中的部分
NEXT: 下一步具体行动
BLOCKER: 阻塞项（如有）
KEY DECISION: 关键决策和理由
```

## Parallel Operation

- task-pool.md 保留归档
- beads 管理活跃任务状态和依赖

## Key Commands

```bash
bd ready --json
bd blocked --json
bd create "Title" -t bug -p 0 -d "Desc" --json
bd show <id> --json
bd update <id> --status in_progress --json
bd close <id> --reason "Done" --json
bd dep add <child> <parent> --type blocks
bd dep tree <id>
bd dolt push
```

## Migration Mapping

`{PROJECT_PATH}/.beads/task-pool-mapping.json`
