# Beads Task Tracking Integration

> **Version**: 1.0.0
> **Updated**: 2026-05-01
> **Status**: Active (parallel with task-pool.md)

## Overview

使用 [beads](https://github.com/gastownhall/beads) (CLI: `bd`) 替代 task-pool.md 管理活跃任务。
Dolt-backed 持久化存储，支持图依赖、多会话上下文恢复、原子 claim。

## Quick Reference

```bash
bd ready              # 查看就绪任务（无 blocker）
bd create "Title" -t bug -p 0 --json   # 创建 P0 bug
bd update <id> --claim --json          # 原子领取任务
bd show <id> --json                    # 查看详情
bd close <id> --reason "Done" --json   # 关闭任务
bd dep add <child> <parent> --type blocks  # 添加依赖
```

## Session Protocol

### 启动时

```bash
cd {PROJECT_PATH}
bd ready --json
bd stats
```

### 工作中

```bash
bd update <id> --claim --json
bd update <id> --notes "COMPLETED: X. IN PROGRESS: Y. NEXT: Z."
```

### 结束时

```bash
bd update <id> --notes "COMPLETED: X. IN PROGRESS: Y. NEXT: Z. KEY DECISION: ..."
bd dolt push
```

## Type Mapping

| _team type | beads type |
|------------|-----------|
| bugfix | bug |
| feature | feature |
| change | task |
| analysis | task |

## Priority Mapping

| _team | beads |
|-------|-------|
| P0 | 0 |
| P1 | 1 |
| P2 | 2 |
| P3 | 3 |

## Dependency Types

| _team concept | beads dep type |
|---------------|---------------|
| A blocks B | `blocks` |
| related | `related` |
| subtask | `parent-child` |
| discovered during work | `discovered-from` |

## Notes Format (Session Handoff)

```
COMPLETED: 已完成的具体工作
IN PROGRESS: 当前进行中的部分
NEXT: 下一步具体行动
BLOCKER: 阻塞项（如有）
KEY DECISION: 关键决策和理由
```

## Parallel Operation

**过渡期**: task-pool.md 保留归档，beads 管理活跃状态
**最终**: task-pool.md 改为静态归档，beads 为唯一任务系统

## Migration Mapping

45 个任务已迁移，映射表：`{PROJECT_PATH}/.beads/task-pool-mapping.json`

| Old ID | Beads ID |
|--------|----------|
| A001 | cms-0vn |
| F014 | cms-zjn |
| B086 | cms-4rc |
| ... | 见 mapping.json |
