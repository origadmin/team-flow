# 任务生命周期

> **版本**: v1.1 | **日期**: 2026-05-11 | **状态**: 待确认

## 1. 状态流转

```
open → in_progress → closed (Review) → (用户确认) → Archived
```

| 状态 | 含义 | 触发者 |
|------|------|--------|
| open | 待领取 | Triage 创建 (`flow task create`) |
| in_progress | 执行中 | Triage 分发后 |
| closed (Review) | 待确认 | 子 Agent 完成后，Triage 更新 |
| Archived | 已归档 | Triage 在用户确认后执行 |

## 2. Triage 中间人原则

> **核心**：Triage 是人与 AI 之间的桥梁。用户永远不直接和子 Agent 对话，所有沟通都通过 Triage 中转。

### 通信模型

```
用户 ←→ Triage ←→ 子 Agent

❌ 错误：用户 ←→ 子 Agent（跳过 Triage）
❌ 错误：Triage 创建任务后不管了
✅ 正确：Triage 始终是中间人
```

### Triage 在每个阶段的作用

| 阶段 | Triage 的角色 | 子 Agent 的角色 |
|------|-------------|---------------|
| 接收需求 | 接收、理解、补全、分类 | 不参与 |
| 创建任务 | 创建 beads 任务 | 不参与 |
| 分发任务 | 注入完整上下文 + 技能档案，分发 | 接收任务，开始执行 |
| 执行中 | 监控状态，中转用户补充信息 | 执行任务 |
| 执行完成 | 接收结果，检查门禁 | 返回结果给 Triage |
| 汇报 | 向用户汇报结果 | 不直接接触用户 |
| 用户反馈 | 接收反馈，按决策树路由 | 等待 Triage 指令 |
| 归档 | 执行归档 | 不参与 |

### Triage 需求补全

Triage 收到用户需求后，不是原样转发，而是先**补全**：

| 补全项 | 说明 | 示例 |
|--------|------|------|
| 目标项目 | 用户说的项目是哪个 | "给 CMS 加通知" → orig-cms-ee |
| 技术栈 | 目标项目用什么技术 | go + gin + ent |
| 影响范围 | 改动涉及哪些模块 | 只涉及后端 API，不涉及前端 |
| 验收标准 | 怎么算做完了 | 通知 API 返回 200，单元测试通过 |
| 边界条件 | 什么不算在内 | 不包括推送通知，只做接口 |

**补全后注入子 Agent 的上下文**，减少子 Agent 回来问问题的次数。

### 用户反馈路由决策树

当用户对结果提出反馈时，Triage 按决策树路由，不需要"聪明地判断"：

```
用户反馈
│
├── 类型 A: 代码 Bug（实现和需求不一致）
│   条件: 需求明确，但代码没按需求实现
│   路由: 直接给 Dev 修复
│   不改需求，不改设计
│
├── 类型 B: 需求理解偏差（AI 理解错了用户意图）
│   条件: AI 做的东西和用户想的不一样
│   路由: Triage 重新补全需求 → 重新分发
│   不需要重新设计，但需求要修正
│
├── 类型 C: 需求变更（用户改主意了）
│   条件: 用户想要不同的东西
│   路由: 回到 Tech Lead 重新设计
│   需求变了，设计也要变
│
├── 类型 D: 新需求（和当前任务无关）
│   条件: 反馈内容超出当前任务范围
│   路由: 创建新任务
│   当前任务继续，新需求另起
│
└── 类型 E: 验收通过
    条件: 用户确认结果 OK
    路由: Triage 归档
```

**关键**：Triage 不需要"判断"，只需要按条件匹配。条件是客观的，不是主观的。

### 用户补充信息的流程

```
子 Agent 执行中遇到问题
    → 子 Agent 报告给 Triage
    → Triage 判断：能自己补全？→ 补全后转达
                     不能？→ 向用户提问
    → 用户回答（如果需要）
    → Triage 将补全后的信息转达给子 Agent
    → 子 Agent 继续执行
```

**Triage 优先自己补全**，只有确实需要用户决策时才问用户。减少来回次数。

**禁止**：子 Agent 直接向用户提问或直接接收用户指令。

## 3. Phase 跟踪

通过 beads labels 跟踪细粒度阶段：

```
phase:ready → phase:analyze → phase:design → phase:implement → phase:verify → phase:review
```

## 4. Feature 流程

```
Phase 0: Triage 创建任务 (phase:ready)
         Triage 分类 → 创建 beads 任务 → 分发给 Tech Lead

Phase 1: Tech Lead 需求分析 + 技术设计 (phase:design)
         产出: R0 + SPEC.md + AC.md + R1 + R2 + R3
         → Tech Lead 返回结果给 Triage
         → Triage 检查门禁 → 汇报用户
         → Triage 分发给 Dev

Phase 2: Dev 实现 (phase:implement)
         产出: 代码 + 单元测试 + SCOPE.md
         → Dev 返回结果给 Triage
         → Triage 检查门禁 → 汇报用户
         → Triage 分发给 QA

Phase 3: QA 验证 (phase:verify)
         产出: 测试报告
         → QA 返回结果给 Triage
         → Triage 检查门禁 → 汇报用户

→ Triage 标记 Review → 用户确认 → Triage 归档
```

## 5. Bug 流程

```
Phase 0: Triage 创建任务 (phase:ready)
         Triage 分类 → 创建 beads 任务 → 分发给 Bugfix

Phase 1: Bugfix 根因分析 (phase:analyze)
         产出: RCA.md
         → Bugfix 返回结果给 Triage
         → Triage 检查门禁 → 汇报用户

Phase 2: Bugfix 修复实现 (phase:implement)
         产出: 代码修复 + TEST_CASE.md + SCOPE.md
         → Bugfix 返回结果给 Triage
         → Triage 检查门禁 → 汇报用户
         → Triage 分发给 QA

Phase 3: QA 验证 (phase:verify)
         产出: 测试报告
         → QA 返回结果给 Triage
         → Triage 检查门禁 → 汇报用户

→ Triage 标记 Review → 用户确认 → Triage 归档
```

### R 迭代规则

- 一个 Bug 一个 ID，用 R 后缀跟踪修复轮次
- 验证未通过 → Triage 创建 R{N+1} 目录，禁止覆盖当前 R
- 每轮 R 必须重新分析失败原因（不能跳过 RCA）
- beads task ID 永远是基础 ID，不随 R 变化
- `doc_path` metadata 指向当前最新的 R 目录

## 6. 三层门禁

### Layer 1: 入口门禁（Triage 收到用户输入时）

| 输入类型 | Triage 的处理 |
|---------|-------------|
| 新需求 | 分类 → 创建 beads 任务 → 分发 |
| 已有任务的补充信息 | 转达给对应子 Agent |
| 已有任务的状态查询 | 从 beads 读取状态 → 回答用户 |
| 澄清/闲聊 | 直接回答 |

**关键**：所有输入都先经过 Triage，Triage 决定如何处理。不存在"直接加载子角色"。

### Layer 2: 阶段门禁（子 Agent 完成阶段后，Triage 检查）

| 检查项 | 说明 |
|--------|------|
| 产出物完成 | 该阶段的产出物是否已生成 |
| 测试通过 | 代码变更是否通过测试 |
| beads 任务已更新 | 状态和 Phase 是否已更新 |

**不通过**：Triage 将问题反馈给子 Agent 修正，不直接标记完成。

### Layer 3: 完成门禁（Triage 标记 Review 前）

| 检查项 | 说明 |
|--------|------|
| 所有产出物已生成 | 全流程产出物齐全 |
| 所有测试通过 | 无失败测试 |
| 无回归 | 未引入新问题 |
| beads 任务可关闭 | 状态正确 |

**不通过**：Triage 将任务退回对应子 Agent，不提交 Review。

## 7. ID 命名规则

| ID 格式 | 类型 | R 后缀 |
|---------|------|--------|
| F{NNN} | Feature | 不带 |
| B{NNN} | Bug | beads 中不带，资产目录带 R |
| C{NNN} | Change | 不带 |
| D{NNN} | Docs | 不带 |
| A{NNN} | Analysis | 不带 |

**强制规则**：
- TYPE 只允许 F/B/C/D/A 五种
- 同一 beads 数据库中全局唯一
- 新任务 ID = 当前同类型最大 ID + 1
