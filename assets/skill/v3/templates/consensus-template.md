# CONSENSUS.md 模板 — v2 beads-native

> **位置**: `.team/consensus.md`

---

## beads 关联

| 字段 | 内容 |
|------|------|
| Issue ID | `<beads-id>` |
| Feature ID | F{NNN} |

---

## 使用说明

此文件记录项目讨论中达成的**共识决策**，防止 AI 反复询问同样问题。

**读取规则**：
- 每个角色在启动时必须读取此文件（见各角色 prompt 的 Activation 部分）
- 新增共识时，立即更新此文件
- 过时的共识标记为 `archived`

---

## 共识列表

| 日期 | 共识内容 | 相关任务 | 状态 |
|------|----------|----------|------|
| {YYYY-MM-DD} | {consensus description} | {TASK_ID} | active/archived |

---

## 共识详情

### {consensus_title}

**日期**: {YYYY-MM-DD}
**相关任务**: {TASK_ID}
**状态**: active

**共识内容**:
{What was decided}

**背景**:
{Why this decision was made}

**影响范围**:
{Which parts are affected}

---

## beads 状态更新

```bash
flow task update <issue-id> --notes "CONSENSUS: {summary}"
```
