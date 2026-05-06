# _team 规则体系：四大流程处理总结

> **Version**: v1.0 | **Date**: 2026-04-29 | **Author**: QClaw

---

## 概述

_team 规则体系围绕四个核心维度构建，确保 AI 多角色协作的可控性和质量：

| 维度 | 核心原则 | 一句话 |
|------|----------|--------|
| 并发 | 隔离 | 并发靠隔离 |
| 记忆 | 强制读取 | 记忆靠强制读取 |
| 纠错 | 门禁拦截 | 纠错靠门禁拦截 |
| 复盘 | 结构化沉淀 | 复盘靠结构化沉淀 |

---

## 1️⃣ 并发（多角色/多Agent协作）

### 1.1 角色隔离

每个角色只做自己的事，禁止越权：

| 角色 | 允许 | 禁止 |
|------|------|------|
| Triage | 分类、分发、编排、状态流转 | 写代码、做架构设计、创建资产包 |
| Tech Lead | 技术设计、创建资产包 | 维护 MILESTONES、写实现代码 |
| Dev | 编码实现、Bug 修复 | 跳过 Tech Lead 直接创建资产包 |
| QA | 测试验证 | 放行上线 |
| PM | 验收放行 | 技术决策 |
| DevOps | CI/CD、部署 | 代码审查、验收签字 |

**文件**: SKILL.md, triage.md, shared.md

### 1.2 分发仲裁

- **Triage = 唯一编排者**，所有任务经 Triage 分类后才分发给子 Agent
- 分类+分发是**原子操作**，不允许"顺手做了"
- 分发自检：每次操作前必须判断"这是 Triage 职责还是子 Agent 职责"

**文件**: triage.md 分发自检

### 1.3 Agent 映射

每种任务类型绑定固定 subagent_type，不可随意分配：

| 任务类型 | 分发给 | subagent_type |
|---------|--------|--------------|
| Feature | Tech Lead | tech-lead-architect → developer-engineer |
| Bug | Dev | bugfix-expert |
| Change | Tech Lead | tech-lead-architect |
| Analysis | Analysis Expert | analysis-expert |
| UI | UI Designer | ui-designer |
| DevOps | DevOps | devops-engineer |

**文件**: SKILL.md Agent 映射表

### 1.4 边界防护

两层架构，严格读写分离：

| 层 | 路径 | AI 权限 |
|----|------|---------|
| 框架层 | `_team/` | **只读** — 禁止写入 |
| 项目层 | `.team/` | **读写** — 运营文件在这里 |

**文件**: BOUNDARY.md

### 1.5 R 迭代规则

同一 Bug 不同修复尝试不创建新 ID：

| 场景 | 正确做法 | 错误做法 |
|------|---------|---------|
| B019 第一轮修复 | B019-R1/ | — |
| 修复未通过，需重试 | B019-R2/ | 新建 B026 |
| 修复导致新子问题 | 合并到 B019 的 R 迭代 | 新建 B022 |

**文件**: shared.md Bug R-迭代跟踪规则

### 1.6 接口契约保护

Breaking Change 必须通知所有实现方/调用方：

| 变更类型 | 是否 Breaking | 处理方式 |
|---------|-------------|---------|
| 新增字段/方法 | ❌ | 直接添加 |
| 修改字段类型 | ✅ | 版本化 + 兼容层 |
| 删除字段/方法 | ✅ | 废弃标记 → 延迟删除 |

**文件**: lessons/dev-common.md Rule 6

---

## 2️⃣ 记忆（防止AI遗忘上下文/规则）

### 2.1 Context Checkpoint

每个对话轮次开始必须输出以下内容（最高优先级）：

```
🔍 [Task Context]
   Task: {task-id 或 "新输入"}
   Phase: {当前所处阶段}
   Required Docs: {本阶段必须有的产出物清单}
   Toolchain: {从 project.md 读取包管理器命令}
   Mode: {新任务 / 延续 / R迭代}
```

Mode 判断逻辑：

| 场景 | Mode |
|------|------|
| 无 task-pool 任务，用户新输入 | 新任务 |
| 有进行中任务，用户继续讨论同一任务 | 延续 |
| 执行中任务发现子问题（同模块/同根因） | R迭代 |
| 执行中任务发现完全无关的新问题 | 新任务（创建新 ID） |

**文件**: SKILL.md 🚨 Context Checkpoint

### 2.2 角色确认 + Input/Output Requirements

任务开始时必须确认角色和产出物清单：

```
🔒 角色确认：
   角色: Dev
   产出物: SCOPE.md, 代码实现
   位置: {docs_internal}/features/{task-id}/
```

10 个角色的 Input/Output Requirements 已全部添加到 prompts/{role}.md：

| 角色 | 输入 | 输出 |
|------|------|------|
| triage | 用户请求, task-pool | task-pool 更新, 分类报告 |
| tech-lead | 任务池条目, 分类报告 | SPEC, AC, R1-R3 |
| dev | SPEC, AC, R1-R3 | SCOPE.md, 代码 |
| bugfix | 任务池条目, Bug 描述 | RCA, 修复代码, SCOPE |
| qa-engineer | 任务池条目, SCOPE, 代码 | 测试报告, 验证报告 |
| pm | 任务池条目, SPEC, 测试报告 | 验收签字 |
| devops | 任务池条目, 验收报告 | 部署报告, CI/CD 配置 |
| ui-designer | 任务池条目, 需求描述 | 设计稿, UI Spec |
| framework-architect | 任务池条目, 架构需求 | 架构设计, 技术规范 |
| analysis | 任务池条目, 分析目标 | 分析报告 |

**文件**: prompts/{role}.md ×10

### 2.3 Toolchain 门禁

执行包管理命令前必须从 project.md 读取 Toolchain 配置：

1. 读取 `.team/project.md` 的 `Toolchain` 部分
2. 使用 `Toolchain.package_manager` 定义的命令
3. 禁止使用与配置不符的包管理器

**文件**: lessons/dev-common.md Rule 1

### 2.4 修改前必读

修改任何已有文件前必须：

1. 完整阅读目标文件
2. grep 目标函数/类型的所有引用点
3. 理解调用链和依赖关系
4. 评估修改影响范围

**文件**: lessons/dev-common.md Rule 5

### 2.5 Lessons 加载

按层级顺序加载经验规则，文件不存在则跳过（零 token 消耗）：

```
common.md → {role-group}-common.md → {role}-{subtype}-common.md → {role}-{type}.md
```

示例（Backend Dev 做 API + DB 任务）：
1. common.md
2. dev-common.md
3. dev-backend-common.md
4. dev-backend-api.md
5. dev-backend-db.md

**文件**: ERROR_REPORT_MECHANISM.md §6

---

## 3️⃣ 纠错（错误发现与修复）

### 3.1 三层门禁

| 门禁层 | 执行时机 | 检查内容 |
|--------|---------|---------|
| Layer 1: 入口门禁 | 收到用户输入 | 意图识别、任务分类 |
| Layer 2: 阶段门禁 | 角色开始执行前 | 上一阶段产出物是否存在 |
| Layer 3: 完成门禁 | 声称完成前 | 所有必须产出物存在+非空 |

**Feature 阶段产出物清单**：

| Phase | 产出物 |
|-------|--------|
| Phase 0 | task-pool 条目 |
| Phase 1 | SPEC.md, AC.md, R1/R2/R3 |
| Phase 2 | 代码 + 测试 + SCOPE.md |
| Phase 3 | 测试报告 |

**Bugfix 阶段产出物清单**：

| Phase | 产出物 |
|-------|--------|
| Phase 0 | task-pool 条目 |
| Phase 1 | RCA.md |
| Phase 2 | 修复代码 + TEST_CASE.md + SCOPE.md |
| Phase 3 | 验证报告 |

**文件**: shared.md Layer 1-3, SKILL.md 阶段门禁/完成门禁

### 3.2 质量门（QG）

| 质量门 | 执行时机 | 核心问题 |
|--------|---------|---------|
| QG-1: Think Before Coding | 拿到任务后、执行前 | 我在基于哪些未确认的假设行动？ |
| QG-2: Simplicity First | 每次输出前 | 能用更少的代码解决吗？ |
| QG-3: Surgical Changes | 修改代码时 | 只改任务范围内必要的代码 |
| QG-4: Goal-Driven Execution | 任务开始前 | 成功标准是什么？ |

**文件**: shared.md 质量门

### 3.3 回归守卫

分层测试策略，避免全量测试拖慢开发、又避免局部测试漏回归：

| 阶段 | 测试范围 | 频率 |
|------|---------|------|
| 开发中（TDD 循环） | 当前模块测试 | 每次修改 |
| 阶段性回归 | 全量测试 | 每 N 次局部后（建议 N=3） |
| 完成门禁 | 全量测试 | 必须通过 |

**文件**: lessons/dev-common.md Rule 4

### 3.4 错误报告机制

错误发生时结构化记录：

- 存储路径：`{DOCS_INTERNAL}/reports/errors/`
- 文件命名：`E{NNNNN}-{TASK_ID}-{SEQ}.md`
- 索引文件：INDEX.md 跟踪整理进度
- 触发条件：Pipeline 失败 / 超范围变更 / 集成失败 / 跨模块影响

**错误分类体系**：

| 分类 | 代码 | 定义 |
|------|------|------|
| API | api | 接口定义、请求响应格式、版本兼容 |
| 数据库 | db | 字段定义、索引、迁移 |
| 代码 | code | 业务逻辑、边界条件、并发 |
| 范围 | scope | 修改文件超出预期范围 |
| 测试 | test | 测试覆盖、测试数据 |
| 架构 | arch | 模块划分、依赖方向 |

**文件**: ERROR_REPORT_MECHANISM.md §2-3

### 3.5 禁止项

- ❌ 中文注释（代码文件）
- ❌ 跳过 RCA / SCOPE / 验收标准
- ❌ PowerShell 管道文本替换（破坏编码）
- ❌ 跳过 PRE-FLIGHT
- ❌ 跳过质量门自检
- ❌ 为同一 Bug 不同修复尝试分配新 B-ID
- ❌ 写入框架层 `_team/`

**文件**: shared.md ⛔ 禁止项

---

## 4️⃣ 复盘（经验沉淀与规则进化）

### 4.1 两层机制

| 层 | 频率 | 产出 |
|----|------|------|
| Layer 1 | 每次错误 | 轻量错误报告 E{NNNNN} |
| Layer 2 | 里程碑结束 / 手动触发 | 结构化审阅 → AI 规则 |

### 4.2 审阅流程

```
Milestone 结束
    ↓
PM / Tech Lead 审阅错误报告
    ↓
输出结构化审阅结果（YAML + 表格，非散文）
    ├── 哪些错误值得沉淀（filter）
    ├── 影响哪些角色（categorize）
    ├── 属于什么分类（tag）
    └── 规则方向（guide）
    ↓
Lesson Analyst 翻译为 AI 规则
    ↓
写入 {DOCS_INTERNAL}/lessons/{role}-{type}.md
    ↓
更新 INDEX.md 标记已处理
```

**关键约束**：Lesson Analyst 只做翻译，不做独立判断。

**文件**: ERROR_REPORT_MECHANISM.md §4-5

### 4.3 规则格式

每条规则必须包含：

```
Rule {N}: {标题}
├── Problem: 一句话描述错误模式
├── Rule: 规则正文
├── Correct: 正确做法
├── Wrong Example: 错误示例
├── Correct Example: 正确示例
├── Applies When: 适用场景
└── Source: 来源错误报告 ID
```

**文件**: ERROR_REPORT_MECHANISM.md §5.3

### 4.4 两级共通结构

经验规则按层级组织，越往下越具体：

```
lessons/
├── common.md                    ← 全局通用（所有角色读）
├── dev-common.md                ← Dev 角色通用（Backend + Frontend 读）
├── dev-backend-common.md        ← Backend Dev 通用
├── dev-backend-api.md           ← Backend Dev API 规则
├── dev-backend-db.md            ← Backend Dev DB 规则
├── dev-frontend-common.md       ← Frontend Dev 通用
├── qa-common.md                 ← QA 角色通用
└── ...
```

加载顺序：common → role-group-common → role-subtype-common → role-type

**文件**: ERROR_REPORT_MECHANISM.md §6

### 4.5 已有经验规则

当前 `lessons/dev-common.md` 已沉淀 6 条规则：

| # | 规则 | 来源 |
|---|------|------|
| Rule 1 | Toolchain Gate — 包管理器强制检查 | E00001-F001-001 |
| Rule 2 | Zero Chinese Comments — 精确范围 | E00005-F003-001 |
| Rule 3 | Commit Message — Type 必须小写 | E00008-F004-001 |
| Rule 4 | Regression Guard — 分层测试 | AI 回归破坏频发 |
| Rule 5 | Read Before Modify — 修改前必读 | AI 局部盲改连锁破坏 |
| Rule 6 | Protected Interface Contract — 接口契约 | AI 单方面改接口 |

---

## 整体流程图

```
用户输入
  │
  ├── [1.并发] Triage 分类分发 → 子 Agent 隔离执行
  │
  ├── [2.记忆] Context Checkpoint 强制确认 + 角色产出物清单
  │
  ├── [3.纠错] 三层门禁拦截 + 质量门自检 + 回归守卫
  │     │
  │     └── 错误发生 → E{NNNNN} 错误报告
  │
  └── [4.复盘] 里程碑结束 → PM 审阅 → Lesson Analyst 翻译 → lessons/*.md
        │
        └── 下次任务加载经验规则 → 避免重复错误
```

---

## 待完善项

| # | 项目 | 状态 |
|---|------|------|
| 1 | 测试分层（增量 vs 整体） | 待设计 |
| 2 | 变更范围控制（task-pool 中 expected_scope） | 待设计 |
| 3 | 错误传播处理（A 改导致 B 失败） | 待设计 |
| 4 | 多 AI 协同 Git 工作流 | 待设计 |
| 5 | Lesson Analyst prompt 定义 | 待创建 |
| 6 | 复盘流程实际运行 | 待验证 |

---

## 相关文件索引

| 文件 | 作用 | 版本 |
|------|------|------|
| SKILL.md | 框架入口、门禁、Agent 映射 | v7.3 |
| shared.md | 三层门禁 + 工作流 + 产出物规格 | v7.1 |
| triage.md | Triage 执行规则 | v7.4 |
| BOUNDARY.md | 层边界规则 | v1.0 |
| ERROR_REPORT_MECHANISM.md | 错误报告与经验沉淀机制 | v2.0 |
| lessons/dev-common.md | Dev 通用经验规则 | 6 条 |
| prompts/{role}.md ×10 | 角色执行规则（含 I/O Requirements） | — |
