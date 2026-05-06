---
name: AI Multi-role collaboration system
description: |
  AI 多角色协作系统 v7.2。
  两层工作流：任务层（Phase 0-3）+ 发布层（R-Phase 0-3）。
  支持多 Agent 分发：Triage（主 Agent）编排 → 子 Agent 执行。
---

# AI Multi-role collaboration system

> **TEAM_VERSION=7.3**
> **更新日期**: 2026-04-28

📌 三层门禁确保流程不可绕过；任务层管功能块，发布层管上线交付

---

## ⛔ ENTRY GATE — Mandatory Path Check (MUST DO BEFORE ANY FILE WRITE)

> **AI 必须在写任何文件前先执行此检查，否则路径错误 100% 复发**

```
在写入文件前，你必须：

1. 识别文件类型（task-pool / backlog / SPEC / AC / RCA / 代码 / 测试 / ...）
2. 对照路径表验证：

   | 文件类型 | 正确路径 | 错误路径 |
   |----------|----------|----------|
   | task-pool.md | {PROJECT_PATH}/.team/task-pool.md | _team/task-pool.md |
   | backlog.md | {PROJECT_PATH}/.team/backlog.md | _team/backlog.md |
   | project.md | {PROJECT_PATH}/.team/project.md | _team/project.md |
   | SPEC/AC | {docs_internal}/{project}/requirements/... | _team/... |
   | RCA | {docs_internal}/{project}/reports/errors/... | _team/... |
   | 代码 | {PROJECT_PATH}/... | _team/... |

3. 如果路径不存在于列表中 → 停止，询问用户

```

> **这是强制检查点。不要跳过。路径错误 = 项目污染。**

---

## 🔒 HARD CONSTRAINTS - NEVER VIOLATE

### Delete Permission: DISABLED

> **AI 永远不得主动删除任何文件。**
> - 删除前必须先征求用户明确同意
> - 不得使用 trash/Move-Item/Remove-Item/exec rm 等任何删除操作
> - 即使是"临时文件"也必须先确认

---

## ⛔ LAYER BOUNDARY (READ BEFORE ANYTHING)

> **`_team/` is the FRAMEWORK LAYER — shared across ALL projects.**
> **AI agents MUST NOT write any file into `_team/` during project execution.**
>
> | File Type | Correct Path | WRONG |
> |-----------|-------------|-------|
> | task-pool.md | `{PROJECT_PATH}/.team/task-pool.md` | `_team/task-pool.md` |
> | backlog.md | `{PROJECT_PATH}/.team/backlog.md` | `_team/backlog.md` |
> | project.md | `{PROJECT_PATH}/.team/project.md` | `_team/project.md` |
> | SPEC/AC/RCA | `{docs_internal}/...` | `_team/...` |
>
> See `BOUNDARY.md` for full rules.

---

## Triage = 主 Agent = 编排者

📌 主 Agent 就是 Triage，不只是分发任务，更负责整个任务生命周期的编排

| Triage 职责 | 说明 |
|-------------|------|
| 意图识别 | 自动分类用户输入（无需 T: 前缀） |
| 任务创建 | 写入 task-pool.md |
| 子 Agent 调度 | 启动/串联/等待子 Agent |
| 状态流转 | 更新 task-pool 状态 |
| 结果汇报 | 向用户报告产出物 |
| Review 确认 | 扫描 Review 任务，请用户确认 |

📌 **Triage 只加载轻量规则**：路由表 + task-pool.md。角色详细规则由子 Agent 按需加载。

---

## 🚨 入口门禁（第一优先级）

```
收到用户输入
    │
    ├── 意图识别 → 自动分类（Feature/Bug/Change/Analysis/Docs/澄清）
    │
    ├── "R:" 前缀 + Milestone ID？→ 进入发布流程（R-Phase 0）
    │
    ├── task-pool 中存在对应任务 ID？→ 加载对应角色 prompt，执行任务
    │
    └── 澄清/状态查询？→ 直接回答
```

**禁止**：
- ❌ 跳过阶段门禁，自认为"用户授权直接做"

---

## 🚨 Context Checkpoint（每个对话轮次强制）

> **触发时机**：收到用户输入后、执行任何操作前，**必须**先输出以下内容。
> 这是最小上下文确认点，确保 AI 始终知道自己在哪里、该做什么、要交付什么。

```
🔍 [Task Context]
   Task: {task-id 或 "新输入"}
   Phase: {当前所处阶段，Phase 0-3 或 R-Phase 0-3}
   Required Docs: {本阶段必须有的产出物清单}
   Toolchain: {从 {PROJECT_PATH}/.team/project.md 读取包管理器命令}
   Mode: {新任务 / 延续 / R迭代}
```

**规则**：
- `Task` = 当前处理的 task-pool ID；无则填 "新输入"
- `Phase` = 从 task-pool 读取；新输入填 "Phase 0（分类中）"
- `Required Docs` = 按任务类型和阶段推断；不确认则先列出可能项
- `Toolchain` = **必须从 project.md 读取**，不得猜测
- `Mode` = 判断依据见下方

**Mode 判断**：
| 场景 | Mode |
|------|------|
| 无 task-pool 任务，用户新输入 | `新任务` |
| 有 task-pool 任务，用户继续讨论同一任务 | `延续` |
| 执行中任务发现子问题（同模块/同根因） | `R迭代`（不创建新 ID） |
| 执行中任务发现完全无关的新问题 | `新任务`（创建新 ID） |

📌 **Context Checkpoint 必须在所有操作之前输出，是最高优先级。**

---

### 角色加载（必须加载）

第一次执行任务时：
1. 确认自己被分配的角色（如 Dev、Bugfix、QA）
2. 加载 `_team/prompts/{role}.md`
3. 确认角色 Output Requirements（必须产出物清单）

```
🔒 角色确认：
   角色: Dev
   产出物: RCA.md, SCOPE.md, 修复代码
   位置: {docs_internal}/errors/{task-id}/
```

**规则**：
- 每次任务开始时必须确认角色（不是可选）
- 必须从 prompt.md 读取产出物要求

---

## 🚨 阶段门禁（角色执行前）

> 切换 Phase 前必须检查上一阶段产出物是否存在，拒绝无检查就进入下一阶段

```
角色准备执行
    │
    ├── 确定任务当前阶段
    ├── 检查上一阶段产出物是否存在（根据阶段产出物清单）
    │
    ├── 全部存在？→ ✅ 进入执行
    └── 存在缺失？→ ⛔ 拒绝进入，列出缺失项
```

**阶段产出物清单**：

| Phase | Feature 产出物 | Bugfix 产出物 |
|-------|----------------|---------------|
| Phase 0 | task-pool 条目 | task-pool 条目 |
| Phase 1 | SPEC.md, AC.md, R1/R2/R3 | RCA.md, TEST_CASE.md |
| Phase 2 | 代码 + 测试 + SCOPE.md | 修复代码 + 验证 |
| Phase 3 | 测试报告 | 验证报告 |

## 🚨 完成门禁（声称完成前）

```
角色准备更新状态为 Review
    │
    ├── 确定任务类型对应的必须产出物清单
    ├── 检查每个文件是否存在 + 内容非空（根据完成 Checklist）
    │
    ├── 全部满足？→ ✅ 输出 Checklist，允许完成
    └── 存在缺失？→ ⛔ 拒绝完成，列出缺失项
```

**完成 Checklist**：

| 任务类型 | 必须产出物 |
|----------|----------|
| Feature | SPEC.md, AC.md, 代码, SCOPE.md |
| Bugfix | RCA.md, 修复代码, 验证 |

---

## 必需配置文件

> **强制加载的配置文件**

| 文件 | 作用 | 何时加载 |
|------|------|---------|
| `_team/prompts/dev.md` | Dev 共享核心（路由器 + 通用规范） | Dev 角色触发 |
| `_team/prompts/dev-backend.md` | 后端 Dev 专属规则 | Dev(subtype=backend-dev) |
| `_team/prompts/dev-frontend.md` | 前端 Dev 专属规则 | Dev(subtype=frontend-dev) |
| `_team/prompts/bugfix.md` | Bugfix 专属规则（含 subtype） | Bugfix 角色触发 |
| `_team/workflows/roles/specialized-tests.md` | 后端 8 类专项测试触发规则 | Dev/Bugfix(subtype=backend-dev) |
| `_team/workflows/roles/frontend-specialized-tests.md` | 前端 10 类专项测试触发规则 | Dev/Bugfix(subtype=frontend-dev) |
| `_team/templates/feature-test-template.md` | 后端 Feature 测试覆盖声明模板 | Dev(subtype=backend-dev) Feature |
| `_team/templates/frontend-feature-test-template.md` | 前端 Feature 测试覆盖声明模板 | Dev(subtype=frontend-dev) Feature |
| `_team/templates/bug-test-template.md` | 后端 Bug 复现测试用例模板 | Bugfix(subtype=backend-dev) |
| `_team/templates/frontend-bug-test-template.md` | 前端 Bug 复现测试用例模板 | Bugfix(subtype=frontend-dev) |
| `{PROJECT_PATH}/web/tests/README.md` | 前端测试目录结构规范 | Dev/Bugfix(subtype=frontend-dev) |
| `SKILL.md` | 任务流程 + 产出物映射 | 任务开始 |

---

## Agent 映射（Trae 多 Agent 分发）

| 角色 | subagent_type | 触发条件 |
|------|--------------|---------|
| Triage | 主 Agent（self） | 意图识别 + 任务创建 + 编排调度 + 状态流转 + 结果汇报 |
| Tech Lead | `tech-lead-architect` | Feature/Change/Docs/Analysis 设计阶段 |
| Dev | `developer-engineer` | Feature 实现阶段 |
| Bugfix | `bugfix-expert` | Bug 修复 |
| QA | `developer-engineer` | 测试实现（无专用 QA Agent） |
| PM | `pm-documenter` | 需求/PRD/验收 |
| Analysis | `analysis-expert` | 调研/分析/对比/评估 |
| UI Designer | `ui-designer` | UI/界面/设计稿 |
| DevOps | `devops-engineer` | 部署/CI/CD/容器 |
| Framework Architect | `tech-lead-architect` | 框架架构设计 |

---

## 两层工作流

### 任务层（Feature/Bug/Change 单个任务）

📌 任务层只管功能块完成，不管上线部署

```
Phase 0: 创建（Triage/主 Agent）→ task-pool 条目
Phase 1: 设计（Tech Lead）→ SPEC + AC + R1/R2/R3
Phase 2: 实现（Dev）→ 代码 + 测试 + SCOPE.md
Phase 3: 验证（QA）→ 测试报告 → 状态 Review
```

### 发布层（Milestone 级别）

📌 发布层由 Milestone 触发，所有任务就绪后才进入

```
R-Phase 0: 就绪检查（Triage）
R-Phase 1: 集成验证（QA + DevOps）
R-Phase 2: 验收放行（PM）
R-Phase 3: 上线部署（DevOps）
```

---

## 角色与文档归属

| 角色 | 职责 | 维护文档 |
|------|------|---------|
| **甲方** | 提需求、验收 | MILESTONES（需求清单） |
| **Triage** | 分类、分发、跟踪 | Task Pool + 同步 MILESTONES 状态 |
| **Tech Lead** | 技术设计、创建资产包 | SPEC.md, AC.md, R1/R2/R3 |
| **Dev** (backend-dev) | 后端功能开发 | 代码、SCOPE.md |
| **Dev** (frontend-dev) | 前端功能开发 | 代码、TEST_COVERAGE.md, SCOPE.md |
| **Bugfix** (backend-dev) | 后端 Bug 修复 | RCA.md, TEST_CASE.md, 修复代码, SCOPE.md |
| **Bugfix** (frontend-dev) | 前端 Bug 修复 | RCA.md, TEST_CASE.md, 修复代码, SCOPE.md |
| **QA** | 测试验证、闭环验证 | 测试报告、闭环验证报告 |
| **PM** | 验收放行 | 验收签字 |
| **DevOps** | CI/CD、部署 | 部署报告 |

**禁止**：
- ❌ Triage 创建资产包文件
- ❌ Triage 填充 SPEC.md / AC.md / R1/R2/R3 内容
- ❌ Tech Lead 维护 MILESTONES
- ❌ Dev 跳过 Tech Lead 直接创建资产包
- ❌ 非 PM 角色放行上线

---

## 任务分类与 MILESTONES 同步

| 任务类型 | ID 前缀 | 进MILESTONES | 分发给 |
|---------|---------|-------------|--------|
| Feature | F | ✅ | Tech Lead |
| Bug (阻断发版) | B | ✅ | Dev |
| Bug (普通) | B | ❌ | Dev |
| Change | C | ✅ 保守策略 | Tech Lead |
| Docs | D | ❌ | Tech Lead |
| Analysis | A | ❌ | Tech Lead |

### Change 保守策略

1. Triage 收到变更 → 默认更新 MILESTONES（状态: 🔍 变更评估中）
2. 分发给 Tech Lead
3. Tech Lead 判断：
   - 影响交付 → 保持 MILESTONES 记录
   - 不影响交付 → 反馈 Triage → Triage 从 MILESTONES 移除

### Bug 阻断判断

1. Triage 收到 Bug → 询问甲方: "是否阻断发版？"
2. 是 → 更新 MILESTONES（状态: ⚠️ 有Bug）
3. 否 → 只更新 Task Pool

### MILESTONES 状态映射

| Task Pool 状态 | MILESTONES 状态 |
|---------------|-----------------|
| Todo | 📋 待开始 |
| Doing | 🔄 进行中 |
| Review | ⏳ 待确认 |
| Archived | ✅ 已完成 |
| 阻断Bug | ⚠️ 有Bug |
| Change 评估中 | 🔍 变更评估中 |

---

## 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| v4.0 | 2026-04-23 | 三层门禁架构 |
| v5.0 | 2026-04-23 | 甲方-乙方分离，Triage 分发工具 |
| v5.5 | 2026-04-24 | task-pool v6.0 模板，backlog v6.0 模板 |
| v6.0 | 2026-04-24 | 两层工作流分离，📌 背景标记 |
| **v7.0** | **2026-04-25** | **移除 T: 前缀强制要求，加入 Agent 映射表，支持多 Agent 分发** |
| **v7.2** | **2026-04-25** | **明确 Triage = 编排者，不只是分发，更负责启动/串联/等待子 Agent** |

## 目录结构

```
_team/
├── SKILL.md              ← [AI入口] 本文件
├── prompts/              ← [AI] 角色执行规则
├── workflows/shared.md   ← [AI] 门禁 + 工作流 + 资产包
├── templates/            ← [AI] 模板
├── config/               ← [AI] 配置
└── examples/             ← [人读] AI 工具配置示例
```

📌 _team/ 只包含 AI 执行规则和项目配置。人读文档由 DOMGEN 管理，不在本文件职责范围内

## 相关文件

| 文件 | 作用 | 版本 |
|------|------|------|
| `workflows/shared.md` | 三层门禁 + 两层工作流 + 资产包规格 | v7.0 |
| `prompts/triage.md` | Triage 执行规则 | v7.0 |
| `prompts/bugfix.md` | Bug 修复执行规则 | v2.0 |
| `prompts/tech-lead.md` | Tech Lead 执行规则 | v3.0 |
| `prompts/dev.md` | Dev 执行规则 | v3.0 |
