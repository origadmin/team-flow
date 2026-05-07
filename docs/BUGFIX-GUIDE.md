<!-- AUTO-GENERATED from _team/prompts/bugfix.md — DO NOT EDIT MANUALLY -->
<!-- To update: modify _team/prompts/bugfix.md, then run DOMGEN regeneration -->

# Bug 修复指南 (Bugfix Guide)

> **版本**: v5.2 | **生成日期**: 2026-04-24

---

## 概述

Bug 修复遵循严格的阶段门禁流程：先分析根因，再修复实现，最后验证。核心原则：**不跳过 RCA 直接修复，不"先修再说，文档后面补"**。

---

## 入口流程

Bugfix 角色被触发时：

1. 确认任务存在于 task-pool.md（不存在则拒绝）
2. 确认任务类型为 bugfix（不是则移交对应角色）
3. 确认状态为 Todo/Doing（Review/Archived 则拒绝）

---

## 阶段流程

### Phase 0: Bug 接收（Triage）

Triage 创建 task-pool.md 条目，分发给 Dev。

### Phase 1: 根因分析（Dev）

**为什么必须先做 RCA？** 直接修复容易治标不治本——修了症状但根因仍在，同类 Bug 会反复出现。RCA 强制找到代码层面的具体漏洞，从根源解决问题。

必须产出 **RCA.md**，内容包含：

| 项目 | 说明 |
|------|------|
| Bug 描述 | 一句话 |
| 影响范围 | 模块/功能/用户 |
| 根本原因 | 代码层面具体漏洞 |
| 根因类型 | 代码Bug/设计缺失/流程缺失/环境/集成/边界条件 |
| 修复方案 | 代码变更草案 |
| 预防措施 | 如何避免再次发生 |

### Phase 2: 修复实现（Dev）

遵循 TDD 循环：

1. **[红]** 编写复现测试 → 失败
2. **[绿]** 最小修复 → 通过
3. **[重构]** 优化 → 保持通过

必须产出 **TEST_CASE.md**，内容包含：

| 项目 | 说明 |
|------|------|
| 复现步骤 | 1. 2. 3. |
| 预期结果 | 修复后应达到的状态 |
| 验证结果 | ✅ 通过 / ❌ 失败 |

### Phase 3: 验证（QA）

QA 基于测试报告验证，任务状态 → Review。

---

## 完成门禁

Bugfix 完成检查：

- RCA.md 存在且包含：现象/根因/影响/预防
- TEST_CASE.md 存在且包含：复现步骤/预期/验证结果
- 代码修复已提交
- go vet ./... 通过
- go test ./... 通过
- 回归测试通过
- task-pool.md 状态 → Review
- 建议后续角色 → QA
- 用户确认前不得归档

---

## 资产包路径

```
{docs_internal}/reports/bugs/{bug-id}-R{N}/
├── RCA.md              ← Dev 创建（Phase 1）
├── TEST_CASE.md        ← Dev 创建（Phase 2）
└── SCOPE.md            ← Dev 创建（完成时）
```

**R 后缀规则**：目录必须带 R 后缀（如 `B001-R1/`），标识修复迭代轮次。禁止创建无 R 后缀的资产目录。

---

## 禁止事项

- **跳过 RCA 直接修复** — 不分析根因就修复，容易治标不治本，同类 Bug 反复出现
- **跳过 TEST_CASE.md** — 没有复现验证的修复，无法确认 Bug 确实被修复
- **"先修再说，文档后面补"** — 经验表明"后面补"永远不会发生
- **未通过完成门禁就更新状态为 Review** — 门禁是架构约束，不是建议

<!-- Last generated: 2026-04-24 18:55 | Source: _team/prompts/bugfix.md | Hash: 63305F8A -->