<!-- AUTO-GENERATED from _team/workflows/shared.md — DO NOT EDIT MANUALLY -->
<!-- To update: modify _team/workflows/shared.md, then run DOMGEN regeneration -->

# 工作流程详解

> **版本**: v6.0 | **生成日期**: 2026-04-24

---

## 概述

Team Workflow 的核心流程由三层门禁和两层工作流构成。三层门禁确保流程不可绕过；两层工作流将任务层和发布层分离，让功能块完成和上线交付成为两个独立流程。

---

## 三层门禁

### 入口门禁（Layer 1）

所有用户输入必须经过入口门禁检查：

1. **"T:" 前缀** → 进入 Triage 分类流程
2. **task-pool 中已有任务 ID** → 加载对应角色 prompt 执行任务
3. **"R:" 前缀 + Milestone ID** → 进入发布流程（R-Phase 0）
4. **都不满足** → 拒绝处理，提示正确用法

**豁免场景**：首次启动 task-pool 不存在时先创建；用户询问任务状态或确认产出物时直接回答。

### 阶段门禁（Layer 2）

角色开始执行前，必须检查上一阶段产出物是否存在。全部存在才放行，有缺失则拒绝进入并列出缺失项。

### 完成门禁（Layer 3）

角色准备更新状态为 Review 时，必须检查所有必须产出物是否存在且内容非空。全部满足才允许完成，否则拒绝并列出缺失项。

---

## 任务层工作流

**为什么任务层到 Phase 3 为止？** 功能块完成只代表该任务就绪，不代表 Milestone 完成，更不代表可以上线部署。发布是独立的流程，由 Milestone 整体触发。

### Feature 任务

| 阶段 | 角色 | 产出物 |
|------|------|--------|
| Phase 0: 创建 | Triage | task-pool.md 条目 |
| Phase 1: 设计 | Tech Lead | SPEC.md + AC.md + R1/R2/R3 |
| Phase 2: 实现 | Dev | 代码 + 单元测试 |
| Phase 3: 验证 | QA | 测试报告 → 状态 Review |

Phase 1 完成后需同步 _docs/ PROJECT.md（如有范围/模块变更）。

### Bugfix 任务

| 阶段 | 角色 | 产出物 |
|------|------|--------|
| Phase 0: 创建 | Triage | task-pool.md 条目 |
| Phase 1: 根因分析 | Dev | RCA.md |
| Phase 2: 修复实现 | Dev | 代码修复 + TEST_CASE.md |
| Phase 3: 验证 | QA | 测试报告 → 状态 Review |

---

## 发布层工作流

**为什么发布是独立层？** 发布流程涉及多个任务的集成验证，由 Milestone 触发而非单个任务触发。所有任务就绪后才进入发布流程。

| 阶段 | 角色 | 检查内容 |
|------|------|---------|
| R-Phase 0: 就绪检查 | Triage | Milestone 下所有任务 → Review 或 Archived？ |
| R-Phase 1: 集成验证 | QA + DevOps | 集成测试 + 回归 + 闭环验证 + 非功能需求 |
| R-Phase 2: 验收放行 | PM | 验收标准 100% 满足 + PM 签字 |
| R-Phase 3: 上线部署 | DevOps | CI 通过 + 部署成功 + 烟雾测试 + 监控正常 |

### 闭环验证

**什么是闭环验证？** 闭环验证确保每个功能的实现与设计文档完全对应，避免"做了但没做对"的问题。QA 逐条对照设计文档，确认功能 100% 闭环。

闭环验证报告包含：任务 ID、设计文档、实现情况、是否闭环，以及补充测试用例和遗留问题。

### PM 唯一放行

**为什么 PM 是唯一放行人？** Tech Lead 关注技术正确性，QA 关注质量达标，但只有 PM 代表业务方确认"这就是我们要的功能"。放行权归属 PM，确保业务验证不可跳过。

---

## 完成门禁详细清单

### Feature 完成门禁

- SPEC.md 存在且非空
- AC.md 存在且非空
- R1_DATA_MODEL.md 存在且非空
- R2_STATE_MACHINE.md 存在且非空
- R3_API_CONTRACT.md 存在且非空
- 核心代码已实现
- 单元测试存在且通过
- Pipeline 通过（Step 1-5）
- 代码无中文注释
- SCOPE.md 已生成
- task-pool.md 状态已更新
- _docs/ PROJECT.md 已同步
- 用户确认前不得归档

### Bugfix 完成门禁

- RCA.md 包含：现象、根因、影响、预防
- TEST_CASE.md 包含：复现步骤、预期结果、验证结果
- Pipeline 通过
- 代码无中文注释
- SCOPE.md 已生成
- task-pool.md 状态已更新
- 用户确认前不得归档

### R-Phase 1 完成门禁（闭环验证）

- 集成测试全部通过
- 回归测试全部通过
- 功能对照设计文档 100% 闭环
- 非功能需求达标（性能/安全）
- 无 Major+ Bug 未关闭
- 闭环验证报告已输出
- QA Engineer 签字

### R-Phase 2 完成门禁（验收放行）

- 验收标准 100% 满足
- 业务闭环确认
- PM 签字放行

### R-Phase 3 完成门禁（上线部署）

- CI Pipeline 全部 Job 通过
- Docker 镜像构建成功
- 部署到生产环境成功
- 烟雾测试通过
- 监控指标正常
- CHANGELOG 已更新
- 版本号已更新（SemVer）

---

## API 问题处理矩阵

**为什么需要这个矩阵？** API 问题容易引发前后端推诿——"接口返回不对"、"前端调错了"。矩阵明确每种场景的责任人，避免扯皮。

| 问题场景 | 负责人 | 处理方式 |
|---------|--------|---------|
| API 响应格式与契约不符 | 后端 | 后端修改 |
| 前端调用了错误的端点/参数 | 前端 | 前端修改 |
| 数据转换/格式化问题 | 前端 | 前端处理 |
| 业务逻辑计算错误 | 后端 | 后端修改 |
| 接口 500/超时 | 后端 | 后端排查 |
| 接口 404 | 双方确认 | 前端调用错误 or 后端路由缺失 |

**处理流程**：前端发现 API 问题 → 检查 R3_API_CONTRACT.md 是否明确定义 → 契约有定义但行为不符（后端问题，Dev 修复）→ 契约未定义/模糊（需求问题，PM 确认后补充契约 → 后端实现）。

---

## 质量门槛

**为什么需要量化标准？** "质量达标"是主观判断，不同人有不同理解。质量门槛把发布层 R-Phase 1 的检查标准量化，让"是否可以发布"变成可测量的判断。

| 指标 | 目标 | 最低要求 |
|------|------|---------|
| 单元测试覆盖率 | 80%+ | 70% |
| 核心业务逻辑覆盖率 | 100% | 90% |
| API 测试通过率 | 100% | 95% |
| E2E 测试通过率 | 100% | 90% |
| Bug 遗留 (Major+) | 0 | ≤ 3 |
| 代码规范合规率 | 100% | 95% |
| 文档完整率 | 100% | 90% |

---

## 资产包规格

### Feature 资产包

Tech Lead 在 Phase 1 创建：

```
{docs_internal}/requirements/{feature-name}-R{N}/
├── SPEC.md              ← 业务背景 + 索引入口
├── AC.md               ← 验收标准
├── R1_DATA_MODEL.md    ← 数据模型
├── R2_STATE_MACHINE.md ← 状态机
├── R3_API_CONTRACT.md  ← 接口契约
└── ui-design/
    └── DESIGN.md       ← UI 设计（视任务需要）
```

Dev 在 Phase 2 完成后生成 SCOPE.md（变更报告）在同目录。

### Bugfix 资产包

Dev 在 Phase 1-2 创建：

```
{docs_internal}/reports/bugs/{bug-id}-R{N}/
├── RCA.md              ← 根因分析
├── TEST_CASE.md        ← 复现验证
└── SCOPE.md            ← 变更报告
```

### Analysis 资产包

Tech Lead 创建：

```
{docs_internal}/analysis/{name}/
├── INDEX.md            ← 入口索引
├── COMPARISON.md       ← 方案对比
└── ADR-XXX.md          ← 架构决策（如需要）
```

### R 后缀规则

**为什么需要 R 后缀？** 资产目录的 R 后缀（如 `F001-R1/`）标识迭代轮次，防止资产目录混淆。当任务第一轮执行失败需要重新执行时，创建新的 R2 目录，与 R1 区分。

- R 后缀加在**资产目录**上，不加在 task-pool 条目上
- task-pool 中永远是基础 ID（F001），资产目录带 R 后缀（F001-R1）
- 禁止创建无 R 后缀的资产目录

---

## 文件空间

| 路径 | 用途 | 维护者 |
|------|------|--------|
| `.team/project.md` | 工具链/路径/约束 | 所有角色读取 |
| `.team/task-pool.md` | 任务状态 | Triage 读写 |
| `.team/backlog.md` | 里程碑/延期任务 | Triage 读写 |
| `.team/issues.md` | 框架问题追踪 | Triage 读写 |
| `_docs/{project}/` | 共享知识库 | AI 写 → 人读 → 人反馈 |

### 文档同步

**为什么绑在阶段门禁上？** 如果文档同步没有明确触发时机，就会拖延到"后面再补"，结果永远不补。把同步绑在门禁上，确保文档与实现不漂移。

- Phase 1 完成后：Tech Lead 检查 _docs/ PROJECT.md 是否反映当前范围
- Feature/Bugfix 完成门禁：Dev 检查 _docs/ PROJECT.md 是否反映实际变更
- 门禁不放行 → 文档不同步

---

## MILESTONES 同步

**MILESTONES 是甲方需求清单**，不是执行状态。Triage 根据 Task Pool 状态自动同步，但不决定需求内容。

| Task Pool 事件 | MILESTONES 操作 |
|---------------|------------------|
| 创建 Feature 任务 | 对应 Milestone 新增任务卡片 |
| 创建阻断性 Bug | 对应 Milestone 新增 Bug 卡片 |
| 任务 → Doing | 状态 → 🔄 进行中 |
| 任务 → Review | 状态 → ⏳ 待确认 |
| 任务 → Archived | 状态 → ✅ 已完成 |

---

## ID 命名规范

**为什么统一 ID 格式？** 旧格式（如 BUG-P001）与新规则（B001）混用时，AI 无法正确匹配任务。统一格式后，ID 前缀直接映射任务类型。

格式：`{TYPE}{SEQUENCE}[-{ITERATION}]`

| ID | 类型 |
|----|------|
| F001 | Feature |
| F001-R1 | Feature 第一轮迭代 |
| B001 | Bug |
| B001-R1 | Bug 第一次修复尝试 |
| C001 | Change |
| D001 | Documentation |
| A001 | Analysis |

---

## 任务生命周期

```
Todo → Doing → Review → (用户确认) → Archived
```

- Todo: Triage 创建
- Doing: 执行角色认领
- Review: 产出物完成，等待用户确认
- Archived: Triage 执行归档

未通过完成门禁禁止更新为 Review；归档必须由 Triage 执行。

<!-- Last generated: 2026-04-24 18:55 | Source: _team/workflows/shared.md | Hash: B3E2842E -->