# Backlog 模板 (v6.0)

```markdown
# Backlog (延期与待讨论任务)

**Owner**: Triage | **最后更新**: {YYYY-MM-DD}

## Deferred
| ID | 任务 | 类型 | 优先级 | 状态 | Blocker | 工时 | 来源 | 关联文档 |
|----|------|------|--------|------|---------|------|------|----------|
| F001 | {名称} | feature | P1 | ⏸ waiting | 等待用户指示 | ~46.5h | 需求确认 | `_docs/...` |

状态枚举: ⏸ waiting | 🚧 blocked | 🔥 escalated | ✅ ready
Blocker: 无 | 等待某人 | 需要决策 | 依赖某任务

## TBD (待讨论)
| ID | 议题 | 类型 | 提出人 | 状态 | 备注 |
|----|------|------|--------|------|------|
| A001 | {议题} | analysis | Tech Lead | 🔍 待讨论 | - |
```

**状态枚举规则**:
- `⏸ waiting`: 暂缓，等待某条件
- `🚧 blocked`: 卡住，有明确阻碍
- `🔥 escalated`: 已升级，需要更高层决策
- `✅ ready`: 已解决，可移回 Task Pool
