# 三层门禁 + 任务生命周期 + 发布流程

> **TEAM_VERSION=7.0** | **更新日期**: 2026-04-25

📌 三层门禁确保流程不可绕过；任务层管功能块，发布层管上线交付

---

## Layer 1: 入口门禁

**执行时机**: 收到用户输入的第一时间。

```
用户输入
    │
    ├── 意图识别 → 自动分类（Feature/Bug/Change/Analysis/Docs）
    ├── task-pool 任务 ID？→ 进入任务执行流程
    ├── "R:" 前缀 + Milestone ID？→ 进入发布流程
    └── 澄清/状态查询？→ 直接回答
```

📌 所有用户输入自动进入分类流程，无需特殊前缀

**豁免场景**:
| 场景 | 处理方式 |
|------|---------|
| 首次启动，task-pool 不存在 | 创建 task-pool 后进入正常流程 |
| 用户询问任务状态/产出物 | 读取 task-pool，回答问题 |
| 用户确认产出物 | 读取产出物，提供确认选项 |

---

## Layer 2: 阶段门禁

**执行时机**: 角色开始执行前第一件事。

### Feature 任务阶段

```
Phase 0: 任务创建（Triage）
  产出物: task-pool.md 条目

Phase 1: 需求分析 + 技术设计（Tech Lead）
  前置: task-pool.md 条目
  产出物: SPEC.md + AC.md + R1/R2/R3
  文档同步: 创建/更新 _docs/ PROJECT.md

Phase 2: 实现（Dev）
  前置: SPEC.md + AC.md + R1/R2/R3
  产出物: 代码 + 单元测试 + SCOPE.md

Phase 3: 验证（QA）
  前置: 代码 + 单元测试
  产出物: 测试报告
  → 任务状态 → Review（功能块就绪）
```

📌 任务层到 Phase 3 为止，功能块完成不代表可以上线

### Bugfix 任务阶段

```
Phase 0: Bug 接收（Triage）
  产出物: task-pool.md 条目

Phase 1: 根因分析（Dev）
  产出物: RCA.md

Phase 2: 修复实现（Dev）
  前置: RCA.md
  产出物: 代码修复 + TEST_CASE.md + SCOPE.md

Phase 3: 验证（QA）
  前置: 代码修复 + TEST_CASE.md
  产出物: 测试报告
  → 任务状态 → Review
```

### 发布阶段（Release Workflow）

📌 发布层由 Milestone 触发，非单个任务触发

```
R-Phase 0: 就绪检查（Triage）
  条件: Milestone 下所有任务 → Review 或 Archived
  不满足 → 列出未完成任务，拒绝进入发布

R-Phase 1: 集成验证（QA + DevOps）
  产出物: 闭环验证报告
  检查: 集成测试/回归测试/功能闭环/非功能需求/Major+ Bug

R-Phase 2: 验收放行（PM）
  产出物: 验收签字
  检查: 验收标准/业务闭环
  📌 PM 是唯一放行人

R-Phase 3: 上线部署（DevOps）
  产出物: 部署报告
  检查: CI Pipeline/Docker/部署/烟雾测试/监控
```

---

## Layer 3: 完成门禁

**执行时机**: 角色准备更新状态为 Review 时。

### Feature 完成门禁

```
- [ ] R0_NAVIGATION_MATRIX.md 存在且非空
- [ ] R0 中定义的所有入口已在代码中实现
- [ ] SPEC.md 存在且非空
- [ ] AC.md 存在且非空
- [ ] R1_DATA_MODEL.md 存在且非空
- [ ] R2_STATE_MACHINE.md 存在且非空
- [ ] R3_API_CONTRACT.md 存在且非空
- [ ] 核心代码已实现
- [ ] 单元测试存在且通过
- [ ] Pipeline 通过（Step 1-5）
- [ ] 代码无中文注释
- [ ] SCOPE.md 已生成
- [ ] TEST_COVERAGE.md 存在且非空（使用 feature-test-template.md）
- [ ] 前后端接口路径对照: 前端API调用路径 vs 后端路由注册路径 100% 匹配
- [ ] 前后端参数命名对照: 前端请求参数名 vs 后端期望参数名 100% 匹配
- [ ] 前后端响应结构对照: 前端TypeScript类型 vs 后端Proto/JSON响应 100% 匹配
- [ ] Handler注册完整性: Proto定义的所有API均有对应Handler注册
- [ ] task-pool.md 状态已更新
- [ ] _docs/ PROJECT.md 已同步（如有范围变更）
- [ ] 用户确认前不得归档
```

### Bugfix 完成门禁

```
⛔ 测试验证执行（HARD GATE — 必须展示实际命令输出）:
- [ ] 后端Bug: go build ./... 编译通过（展示输出）
- [ ] 后端Bug: go test ./... 全部通过（展示通过数量）
- [ ] 前端Bug: bun run typecheck 类型检查通过（展示输出）
- [ ] 前端Bug: bun run lint 代码检查通过（展示输出）
- [ ] 前端Bug: bun run test 全部通过（展示通过数量）
- [ ] Bug 复现测试通过（展示测试输出）
- [ ] 全量回归测试通过（展示测试输出）

文档检查:
- [ ] RCA.md 存在且包含：现象、根因、影响、预防
- [ ] TEST_CASE.md 存在且包含：复现步骤、预期结果、验证结果
- [ ] 代码无中文注释
- [ ] SCOPE.md 已生成
- [ ] TEST_CASE.md 存在且包含复现步骤和回归用例（使用 bug-test-template.md）
- [ ] 修复涉及API变更？→ 是则执行前后端接口路径/参数/响应对照检查

[v2] 数据流追踪（涉及API/权限/状态/交互的Bug必须）:
- [ ] RCA.md 包含"数据流追踪"章节
- [ ] 断点已定位到具体环节（不是"可能是xxx"）

[v2] 真实场景验证（涉及API/权限/状态/交互的Bug必须）:
- [ ] TEST_CASE.md 包含真实场景验证（非仅mock测试）
- [ ] 覆盖默认值/零值/边界条件
- [ ] 验证结果为通过（不是"已编写"）

[v2] 运行时验证（涉及API/权限/状态/交互的Bug必须）:
- [ ] 后端Bug需HTTP请求验证通过（非仅UseCase单元测试）
- [ ] 前端Bug需UI运行时验证（非仅组件mock测试）:
  - [ ] UI_VERIFICATION.md 已生成（使用 ui-verification-template.md）
  - [ ] 启动dev server，打开Bug涉及页面
  - [ ] 检查页面内容（文本/数据/组件正确显示）
  - [ ] 执行交互行为（点击/输入/提交）
  - [ ] 重现Bug原始触发步骤，确认Bug现象消失
  - [ ] 副作用检查（相邻功能、导航正常）
  - [ ] 截图保存到 {docs_internal}/reports/bugs/B{NNN}-R{N}/screenshots/
  - [ ] 截图命名: B{NNN}-R{N}-{3位步骤号}-{动作}-{状态}.png
  - [ ] 截图数量 >= Bug类型最低要求
- [ ] 修复后的完整链路已验证（非仅断点环节）

流程检查:
- [ ] task-pool.md 状态已更新
- [ ] _docs/ PROJECT.md 已同步（如有影响模块变更）
- [ ] 用户确认前不得归档
```

> **v2 增强说明**: 标记 `[v2]` 的检查项来自 `bugfix-standards-v2.md`，针对"简单Bug多轮修复失败"问题新增。纯样式/纯配置类Bug可跳过数据流追踪，但须在RCA.md中说明跳过原因。

### R-Phase 完成门禁

**R-Phase 1（闭环验证）**:
- [ ] 集成/回归测试通过
- [ ] 功能对照设计文档 100% 闭环
- [ ] 非功能需求达标
- [ ] 无 Major+ Bug
- [ ] 闭环验证报告已输出
- [ ] QA Engineer 签字

**R-Phase 2（验收放行）**:
- [ ] 验收标准 100% 满足
- [ ] 业务闭环确认
- [ ] PM 签字放行

**R-Phase 3（上线部署）**:
- [ ] CI Pipeline 全部 Job 通过
- [ ] Docker 镜像构建成功
- [ ] 部署到生产环境成功
- [ ] 烟雾测试通过
- [ ] 监控指标正常
- [ ] CHANGELOG 已更新
- [ ] 版本号已更新（SemVer）

---

## 资产包规格

### Feature 资产包

```
{docs_internal}/requirements/{TASK_ID}-{feature-name}/
├── R0_NAVIGATION_MATRIX.md ← 导航与入口矩阵（必须，最先创建）
├── SPEC.md              ← 业务背景
├── AC.md               ← 验收标准
├── R1_DATA_MODEL.md    ← 数据模型
├── R2_STATE_MACHINE.md ← 状态机
├── R3_API_CONTRACT.md  ← 接口契约
└── SCOPE.md            ← Dev 创建，变更报告
```

📌 **Feature 目录命名规则（强制）**：
- 格式：`{TASK_ID}-{kebab-case-name}/`，如 `F014-unified-pagination/`
- TASK_ID 前缀必须与 task-pool 条目 ID 一致
- Feature 目录**不带 R 后缀**（R 后缀仅用于 Bug 修复轮次）
- Feature 的设计迭代通过 `delivery/{feature}/versions/` 版本体系跟踪
- 已有目录缺 TASK_ID 前缀的，需在下次操作该 Feature 时补齐

### Bugfix 资产包

```
{docs_internal}/reports/bugs/{bug-id}-R{N}/
├── RCA.md              ← 根因分析（每轮 R 独立）
├── TEST_CASE.md        ← 复现验证
├── SUMMARY.md          ← Bug 修复总结报告（使用 bugfix-standards.md 第七节格式）
└── SCOPE.md            ← Dev 创建，变更报告
```

📌 **Bug 目录 R 后缀规则（强制）**：
- R 后缀仅用于 Bug 修复轮次跟踪：R1→R2→R3...
- 每轮修复有独立目录，历史目录保留不删除
- R 目录由执行角色（Bugfix/Dev）在修复时创建，Triage 不预创建

### Change 资产包

```
{docs_internal}/reports/changes/{change-id}/
├── SCOPE.md            ← Dev 创建，变更报告（必须）
└── （其他参考文档按需）
```

📌 **Change 报告强制要求**：Dev 完成 Change 后必须创建 `SCOPE.md`，包含变更清单、设计决策、影响范围、QA 验证要点。报告路径写入 task-pool.md 的关联文档列。

📌 **reports/changes/ 目录必须存在**（由 Triage 在 Change 任务创建时预先创建），Dev 完成时写入 SCOPE.md。

📌 Change 目录**不带 R 后缀**（R 后缀仅用于 Bug 修复轮次）。如 Change 需要调整，创建新 Change 任务。

📌 资产目录命名规则汇总：
| 任务类型 | 目录格式 | R 后缀 | 调整跟踪方式 |
|---------|---------|--------|------------|
| Feature | `{TASK_ID}-{name}/` | ❌ 不带 | 版本体系 `delivery/{feature}/versions/` |
| Bug | `{bug-id}-R{N}/` | ✅ 必带 | R 递增（R1→R2→R3） |
| Change | `{change-id}/` | ❌ 不带 | 新 Change 任务 |

---

## ⛔ R0 导航与入口矩阵（Feature 强制）

> **核心原则**: 功能没有入口等于不存在。设计必须从"用户如何触达"开始，而非从"数据如何存储"开始。

### 为什么 R0 是强制的

历史教训：多个功能实现后遗漏入口，导致功能"存在但不可达"：
- 频道创建: 实现了 CreateChannelDialog 但没有入口按钮
- 订阅Tab: 实现了 SubscriptionsTabContent 但没有导航入口
- "我的"菜单: 实现了 navigation.ts 配置但 Header 用户菜单没有对应项

**根因**: 设计从数据/API出发，不从用户旅程出发。入口是最后补的，不是最先设计的。

### 入口类型定义

| 入口类型 | 形式 | 交互方式 | 适用场景 | 示例 |
|---------|------|---------|---------|------|
| **导航项** | Sidebar 菜单项 / Header 菜单项 | 点击跳转页面 | 功能有独立页面 | Sidebar "我的频道" → /me/channels |
| **操作按钮** | Button / IconButton | 点击触发动作 | 功能是当前页面的操作 | "创建频道" 按钮 → 打开对话框 |
| **上下文入口** | 卡片内按钮 / 行内链接 | 点击触发动作 | 功能与当前内容相关 | 频道卡片 "查看频道" / "设置" |
| **空状态引导** | CTA 按钮 + 说明文字 | 点击触发动作 | 用户首次使用或无数据时 | "您还没有频道" + "创建频道" 按钮 |
| **流程引导** | 拦截页 / 对话框 | 阻断或引导 | 用户操作依赖前置条件 | 上传页检测无频道 → 引导创建 |
| **URL 直访** | 直接输入 URL | 浏览器地址栏 | 高级用户 / 分享链接 | /@username /me/channels |
| **快捷入口** | Header 图标 / FAB | 点击触发动作 | 高频操作 | Header "+" 上传按钮 |

### 入口可达性规则

**规则 1: 最少 2 个入口原则**
每个功能必须至少有 2 个不同类型的入口，防止单一入口失效。

| 功能级别 | 最少入口数 | 必须包含 |
|---------|-----------|---------|
| 核心功能（创建/发布/管理） | 3 | 导航项 + 操作按钮 + 空状态引导 |
| 次要功能（设置/编辑/查看） | 2 | 导航项 + 上下文入口 |
| 辅助功能（分享/导出/通知） | 1 | 上下文入口 |

**规则 2: 入口层级覆盖**
功能入口必须覆盖以下层级中的至少 2 个：

| 层级 | 位置 | 说明 |
|------|------|------|
| L1 全局 | Header / Sidebar | 始终可见，不依赖上下文 |
| L2 区域 | 页面内固定位置 | 在特定页面可见 |
| L3 上下文 | 卡片/列表项内 | 与具体内容关联 |
| L4 引导 | 空状态/流程拦截 | 条件触发 |

**规则 3: 新用户首次可达**
新注册用户必须能在 3 次点击内从首页到达任何核心功能。

### R0 文档模板

```markdown
# R0 导航与入口矩阵: {功能名称}

## 1. 功能入口表

| 功能 | 入口类型 | 位置 | 交互 | 优先级 | 状态 |
|------|---------|------|------|--------|------|
| 创建频道 | 导航项 | Sidebar "你"区域 | 点击 → /me/channels | P0 | ☐ |
| 创建频道 | 操作按钮 | /me/channels 页面右上角 | 点击 → 打开对话框 | P0 | ☐ |
| 创建频道 | 操作按钮 | Header 用户菜单 | 点击 → /me/channels | P0 | ☐ |
| 创建频道 | 空状态引导 | /me/channels 无频道时 | 点击 → 打开对话框 | P0 | ☐ |
| 创建频道 | 流程引导 | /me/upload 无频道时 | 点击 → 打开对话框 | P1 | ☐ |
| 创建频道 | 空状态引导 | /@username 无频道用户资料页 | 点击 → 打开对话框 | P1 | ☐ |

## 2. 用户旅程图

```
首页 → Sidebar "我的频道" → /me/channels → 点击 "创建频道" → CreateChannelDialog → 创建成功 → /@handle
首页 → Header 用户菜单 → "我的频道" → /me/channels → ...
上传页 → 检测无频道 → 引导创建 → CreateChannelDialog → 创建成功 → 返回上传
/@username → 无频道用户资料页 → "创建频道" CTA → CreateChannelDialog → 创建成功 → /@handle
```

## 3. 入口可达性检查

- [ ] 核心功能 ≥ 3 个入口
- [ ] 入口覆盖 ≥ 2 个层级（L1/L2/L3/L4）
- [ ] 新用户 3 次点击内可达
- [ ] 每个入口类型已明确定义（导航项/按钮/引导等）

## 4. 入口与组件映射

| 入口 | 触发组件 | 目标组件/页面 |
|------|---------|-------------|
| Sidebar "我的频道" | NavItem | /me/channels → MyChannels.tsx |
| Header "我的频道" | MenuItem | /me/channels → MyChannels.tsx |
| "创建频道" 按钮 | Button | CreateChannelDialog.tsx |
| 上传引导 | 拦截页 | CreateChannelDialog.tsx |
```

### 设计流程变更

```
旧流程:  R1 数据 → R2 状态 → R3 API → R4 组件 → R5 计划 → (入口最后补)
新流程:  R0 入口 → R1 数据 → R2 状态 → R3 API → R4 组件 → R5 计划
```

R0 必须在 R1 之前完成，因为：
1. 入口决定用户旅程，用户旅程决定页面结构，页面结构决定组件设计
2. 先设计入口可以避免"功能实现但无法触达"的问题
3. 入口矩阵是 AC 验收的核心依据

### ⛔ Bug R-迭代跟踪规则（强制）

> Bug 修复几乎不可能一次成功（特别是 AI），必须通过 R1→R2→R3... 持续跟踪。

**核心原则：一个 Bug 一个 ID，用 R 后缀跟踪修复轮次。禁止为同一 Bug 的不同修复尝试分配新 B-ID。**

| 场景 | 正确做法 | ❌ 错误做法 |
|------|---------|------------|
| B019 第一轮修复 | 创建 `B019-R1/` 目录 | — |
| B019 修复未通过，需重试 | 创建 `B019-R2/`，关联文档改为 `B019-R2/` | 新建 B026 |
| B019 修复导致新子问题 | 子问题记入 B019 当前 R 的 SCOPE.md | 新建 B022 |
| B019 R1 验证中发现新现象 | 补充到 B019 的 TEST_CASE.md | 新建 B023 |

**R 迭代流程**：

```
B019-R1/ (Phase 1: RCA → Phase 2: Fix → Phase 3: Verify)
  ├── 验证通过 → Done → task-pool 状态 Review → 用户确认 → Archived
  └── 验证失败 → ⛔ 强制触发 R 递增（见下方规则）
       └── B019-R2/ (回到 Phase 1，新 RCA 分析失败原因)
            ├── R2 验证通过 → Done
            └── R2 验证失败 → ⛔ 强制触发 R 递增 → B019-R3/ ...
```

**⛔ R 递增强制触发规则（新增）**：

> **强制要求**：当 Bug 修复验证未通过时，执行角色**必须**创建 R{N+1} 目录并重置阶段，**禁止在当前 R 目录内覆盖修改**。

| 触发条件 | 判断者 | 必须动作 |
|---------|-------|--------|
| Phase 3 验证结果为“未通过” | 执行角色（Bugfix/Dev） | 1. 创建 `{bug-id}-R{N+1}/` 目录 2. task-pool 阶段重置为 Phase 1 3. 关联文档列更新为 R{N+1} 目录 |
| 用户反馈“还没修好” | 执行角色 | 同上 |
| AI 自测发现修复方向错误 | 执行角色 | 同上 |

**禁止**：
- ❌ 验证未通过时继续在当前 R 目录内修改（必须开新 R）
- ❌ 跳过 RCA 直接在新 R 目录中重试（每轮 R 必须重新分析失败原因）
- ❌ 未更新 task-pool 关联文档列就创建新 R 目录

**task-pool 中的体现**：

- task-pool 条目 ID 永远是基础 ID（`B019`），不随 R 变化
- `关联文档` 列指向当前最新的 R 目录（`B019-R2/`）
- `阶段` 列格式：`Phase {N} (R{M})`，如 `Phase 1 (R2)` — 结构化可解析
- 状态反映最新 R 的阶段（R2 在 Phase 1 → 状态就是 Doing）
- 所有历史 R 目录保留不删除

---

## 文件空间定义

> ⛔ **CRITICAL: Two-layer architecture. `{TEAM_PATH}/` ≠ `.team/`.**
> - `{TEAM_PATH}/` = framework layer (READ-ONLY, shared across all projects)
> - `.team/` = project layer (READ-WRITE, per-project operational files)
>
> **AI must NEVER write to `{TEAM_PATH}/`. All operational files go to `{PROJECT_PATH}/.team/`.**

### Framework Layer (READ-ONLY for AI)

```
framework/
  {TEAM_PATH}/        ← Framework rules (READ-ONLY — do NOT modify during execution)
    ├── SKILL.md       ← entry point
    ├── BOUNDARY.md    ← layer boundary rules
    ├── prompts/       ← role execution rules
    ├── workflows/     ← shared workflows
    └── templates/     ← document templates

  _docs/{project}/    ← Project internal docs (AI writes here)
```

### Project Layer (READ-WRITE for AI)

```
{PROJECT_PATH}/       ← Project root
  .team/              ← AI operational files (READ-WRITE)
    ├── project.md    ← Toolchain/paths/constraints
    ├── task-pool.md  ← Task status (Triage maintains)
    ├── backlog.md    ← Deferred tasks
    └── issues.md     ← Issue tracking

{DOCS_INTERNAL}/      ← Project docs (AI writes here)
  ├── PROJECT.md      ← Project overview
  ├── requirements/   ← Requirements docs
  ├── design/         ← Design docs
  ├── reports/        ← Change/Bug reports
  ├── lessons/        ← Experience lessons (AI writes here, NOT {TEAM_PATH}/lessons/)
  └── test/           ← Test reports
```

---

## MILESTONES 同步

📌 MILESTONES 是甲方需求清单，Triage 根据 Task Pool 状态自动同步

| Task Pool 事件 | MILESTONES 操作 |
|---------------|------------------|
| 创建 Feature 任务 | 对应 Milestone 新增任务卡片 |
| 创建阻断性 Bug | 对应 Milestone 新增 Bug 卡片 |
| 任务 → Doing | 状态更新为 🔄 进行中 |
| 任务 → Review | 状态更新为 ⏳ 待确认 |
| 任务 → Archived | 状态更新为 ✅ 已完成 |

---

## ID 命名

**格式**: `{TYPE}{SEQUENCE}[-{ITERATION}]`

| ID | 类型 | R 后缀 |
|----|------|--------|
| F001 | Feature | ❌ 不带 |
| F001-R1 | ~~Feature 第一轮迭代~~ 已废弃 | — |
| B001 | Bugfix | ❌ task-pool 中不带 |
| B001-R1 | Bugfix 第一轮修复 | ✅ 资产目录带 |
| B001-R2 | Bugfix 修复未通过，第二轮修复 | ✅ 资产目录带 |
| C001 | Change | ❌ 不带 |
| D001 | Documentation | ❌ 不带 |
| A001 | Analysis | ❌ 不带 |

### ⛔ 强制规则

1. **TYPE 只允许 F/B/C/D/A 五种**，禁止自创前缀（FE-/BE-/BF-/TASK-/UI- 等）
2. **SEQUENCE 从 001 递增**，禁止跳号、禁止复用已归档 ID
3. **同一 task-pool 中全局唯一**，不分子类型（前端 Bug 也用 B 前缀，描述中注明即可）
4. **R 后缀仅在 Bug 资产目录使用**，task-pool 条目永远用基础 ID
5. **新任务 ID = 当前同类型最大 ID + 1**
6. **Bug 修复迭代用 R 后缀**（R1→R2→R3...），禁止为同一 Bug 的不同修复尝试分配新 B-ID
7. **Feature 和 Change 不带 R 后缀**，设计迭代用版本体系跟踪，调整用新任务跟踪
8. **子问题必须合并到主 Bug**（新增 B-ID = 根因 B 的 R 迭代）：
   - 发现 Bug A 的子问题 B/C/D → 不新建 B-ID，而是作为 A 的 R2/R3/R4 处理
   - 关联文档列写入 `B{A}-R{N}/`
   - 旧分散 B-ID 标记为 Archived，注明去向

📌 **一个 Bug 一个 ID，R 跟踪修复轮次。禁止为同一 Bug 的不同修复尝试分配新 B-ID。**

📌 **子问题禁止新建 B-ID** — 发现子问题时，判断根因是否已有 B-ID：
- 已有根因 B-ID → 作为该 B-ID 的 R 迭代，不分配新 ID
- 无根因 B-ID → 新建 B-ID，从 R1 开始

📌 R 后缀加在资产目录上，不加在 task-pool 条目上

```
task-pool.md:  B001                 （永远是基础 ID，不随 R 变化）
关联文档列:    `B001-R2/`           （指向当前最新的 R 迭代目录）
资产目录:      B001-R1/ B001-R2/    （所有轮次都保留，不删除）
```

---

## 任务生命周期

```
Todo → Doing → Review → (用户确认) → Archived
```

- **Todo**: Triage 创建
- **Doing**: 执行角色认领
- **Review**: 产出物完成，等待用户确认
- **Archived**: Triage 执行归档

**规则**：
- 状态流转必须在 task-pool.md 中记录
- Review 状态必须等待用户确认
- 未通过完成门禁 → 禁止更新为 Review
- 归档必须由 Triage 执行

---

## 相关模板

| 模板 | 路径 |
|------|------|
| 闭环验证报告 | `{TEAM_PATH}/templates/closed-loop-verification-template.md` |
| SCOPE.md | `{TEAM_PATH}/templates/scope-template.md` |
| task-pool.md | `{TEAM_PATH}/templates/task-pool-template.md` |
| backlog.md | `{TEAM_PATH}/templates/backlog-template.md` |

---

## 相关规范

| 规范 | 位置 |
|------|------|
| API 问题处理矩阵 | `{TEAM_PATH}/workflows/roles/development-standards.md` |
| 质量门槛 | `{TEAM_PATH}/workflows/roles/test-standards.md` |
| 任务池操作规则 | `{TEAM_PATH}/workflows/roles/triage-standards.md` |
| Bug修复增强规范 v2 | `{TEAM_PATH}/workflows/roles/bugfix-standards-v2.md` |

---

## 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| v6.0 | 2026-04-25 | AI 读优化：精简模板，提取到 templates/ |
| v5.0 | 2026-04-24 | 三层门禁 + 任务生命周期 + 发布流程 |
| **v7.0** | **2026-04-25** | **移除 T: 前缀强制要求，改为意图自动识别** |
| **v7.2** | **2026-04-30** | **A层修复：Feature/Change去掉R后缀；Bug R递增强制触发；lessons迁移到{docs_internal}/；task-pool阶段列结构化** |
| **v7.3** | **2026-05-01** | **R0导航与入口矩阵：Feature设计必须先定义入口再设计数据模型；入口类型定义（导航项/按钮/引导/流程拦截/URL直访/快捷入口）；入口可达性规则（核心功能≥3入口/覆盖≥2层级/新用户3次点击可达）** |
| **v7.4** | **2026-05-04** | **Bugfix完成门禁v2增强：数据流追踪+真实场景验证+运行时验证检查项（来源A014分析/bugfix-standards-v2.md）** |

---

## 质量门（强制）

> **参考**：`examples/ClaudeCode/SKILL.md`（Karpathy Guidelines）
> 所有角色执行时必须通过以下质量门禁

### QG-1: Think Before Coding

**执行时机**：拿到任务后、执行任何操作前（PRE-FLIGHT 暂停点）。

- **假设**：我正在基于哪些未确认的假设行动？
- **风险**：如果这个假设错了，最坏的结果是什么？
- **澄清**：有没有我应该先问清楚的点？

> 不确定就停，不要凭记忆继续。

### QG-2: Simplicity First

**执行时机**：每次输出前自检。

- 能用更少的代码解决吗？
- 我在加未请求的特性吗？
- 有没有过度设计？

### QG-3: Surgical Changes

**执行时机**：修改代码时。

- 只改任务范围内必要的代码
- 不顺手"优化"或重构旁边的代码
- 自己改动产生的孤儿代码自己清理

### QG-4: Goal-Driven Execution

**执行时机**：任务开始前。

每个任务必须先定义成功标准，再执行。统一使用以下格式：

```markdown
## goal_plan

**目标**：{一句话描述最终交付物}

**Plan**：
1. [步骤] → verify: [该步骤完成的具体标准]
2. [步骤] → verify: [该步骤完成的具体标准]

**成功标准**：
- [ ] {标准1}
- [ ] {标准2}

**回滚预案**：如果 {条件} 失败 → {操作}
```

> 每步 verify 必须具体，不能用"完成"模糊描述

### ⛔ 禁止项

- ❌ 中文注释
- ❌ 跳过 RCA / SCOPE / 验收标准
- ❌ PowerShell 管道文本替换（`-replace`、`Set-Content`、`Out-File`、`| ForEach-Object { $_ -replace ... }` 等）—— 会破坏文件编码，详见 `qclaw-text-file` skill
  - ✅ 正确做法：用 Python `scripts/write_file.py` 或 Node.js 替代
  - ✅ 如必须用 shell：只用 `cmd /c` 或 PowerShell `Get-Content -Raw` → 处理 → `bash` 写文件
  - ⚠️ PowerShell 读取文件是安全的（`Get-Content`、`Get-Content -Raw`），写文件是危险的
- ❌ 跳过 PRE-FLIGHT
- ❌ 跳过质量门自检
- ❌ 为同一 Bug 的不同修复尝试分配新 B-ID（用 R 后缀）
- ❌ 写入 `framework/{TEAM_PATH}/`（只读层）
- ❌ 写入 `{PROJECT_PATH}/_docs/`（正确路径：`framework/_docs/{project}/`）
