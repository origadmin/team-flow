# MILESTONES 同步规则

> **来源**: 从 `{TEAM_PATH}/workflows/shared.md` 提取
> **用途**: Triage 根据 beads task 状态自动同步 MILESTONES

---

📌 MILESTONES 是甲方需求清单，Triage 根据 beads task 状态自动同步

| beads Task 事件 | MILESTONES 操作 |
|-----------------|------------------|
| 创建 Feature 任务 | 对应 Milestone 新增任务卡片 |
| 创建阻断性 Bug | 对应 Milestone 新增 Bug 卡片 |
| 任务 → in_progress | 状态更新为 🔄 进行中 |
| 任务 → closed (Review) | 状态更新为 ⏳ 待确认 |
| 任务 → Archived | 状态更新为 ✅ 已完成 |
