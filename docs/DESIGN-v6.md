<!-- SNAPSHOT — Frozen as of 2026-04-24 — DO NOT REGENERATE -->
<!-- This document records decisions made during v6.0 design -->
<!-- For current architecture, see docs/design/ARCHITECTURE.md -->

# _team v6.0 设计方案：两层工作流

> 日期: 2026-04-24 | 状态: 已落地
>
> 本文档记录 v6.0 的核心设计决策。实施细节已执行完毕，合并清单和实施步骤已删除。

---

## 一、核心问题

### 问题 1：Phase 编号混淆任务层与发布层

旧 Phase 0-7 模型把两种不同粒度的流程混在一起：
- **任务层**：一个 Feature/Bug 从创建到"功能块完成"（0→1→2→3）
- **发布层**：一个 Milestone 从"所有功能就绪"到"上线交付"（集成→验证→部署→放行）

Feature 完成只代表功能块就绪，不代表 Milestone 完成，更不代表可以上线部署。
强行把发布流程塞进 Feature Phase 编号，导致 Phase 4-7 无法被单个任务触发。

### 问题 2：人读文档缺失

v5.0 精简 AI 规则文件后，说明性内容被删除。新人/PM 看规则文件无法理解"为什么这样设计"。
需要从 AI 规则文件自动派生人读文档，避免双源漂移。

---

## 二、两层工作流设计

### 任务层（Task Workflow）

Feature/Bug/Change 等单个任务的生命周期。**Phase 到"功能块完成"为止。**

```
Phase 0: 创建（Triage）→ task-pool 条目
Phase 1: 设计（Tech Lead）→ SPEC + AC + R1/R2/R3
Phase 2: 实现（Dev）→ 代码 + 测试 + SCOPE.md
Phase 3: 验证（QA）→ 测试报告 → 任务状态 → Review
```

**任务完成 = 功能块就绪**。后续不归任务管。

### 发布层（Release Workflow）

Milestone 级别的交付流程。**当 Milestone 下所有任务就绪后触发。**

```
R-Phase 0: 就绪检查（Triage）→ 所有任务 Review/Archived？
R-Phase 1: 集成验证（QA + DevOps）→ 集成测试 + 回归 + 环境部署
R-Phase 2: 验收放行（PM）→ 对照需求验收 + PM 签字
R-Phase 3: 上线部署（DevOps）→ 生产部署 + 烟雾测试 + 监控确认
```

**关键差异**：
- 任务层 Phase 由任务 ID 触发
- 发布层 R-Phase 由 Milestone ID 触发
- 两者独立，互不耦合

### 角色映射

| 角色 | 任务层 | 发布层 |
|------|--------|--------|
| Triage | 创建任务 + 同步 MILESTONES | 检查 Milestone 就绪状态 |
| Tech Lead | 设计资产包 | 提供技术验收意见 |
| Dev | 编码实现 | 修复集成问题 |
| QA | 功能测试 | 闭环验证 + 回归测试 |
| PM | — | 验收放行（唯一放行人） |
| DevOps | — | CI/CD + 部署 + 监控 |

---

## 三、设计决策

### 决策 1：两层工作流分离

- **原因**：Phase 0-7 混合了任务粒度和发布粒度，Phase 4-7 无法被单个任务触发
- **方案**：任务层 Phase 0-3 管功能块，发布层 R-Phase 0-3 管上线交付
- **结果**：两个流程独立触发、互不耦合

### 决策 2：极简背景标记

- **原因**：v5.0 精简规则后，新人/PM 无法理解规则背后的意图
- **方案**：AI 规则文件中用 `📌` 标记极简背景（≤50字），DOMGEN 生成人读文档时展开
- **结果**：规则文件保持精简，人读文档提供完整解释

### 决策 3：人读文档自维护（DOMGEN）

- **原因**：人读文档与 AI 规则双源维护，容易漂移
- **方案**：AI 规则文件 = 唯一信源，人读文档从规则文件自动派生
- **结果**：DOMGEN 负责生成和同步，AI 执行任务时不读取人读文档
