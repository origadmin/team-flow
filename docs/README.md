<!-- AUTO-GENERATED from _team/SKILL.md — DO NOT EDIT MANUALLY -->
<!-- To update: modify _team/SKILL.md, then run DOMGEN regeneration -->

# Team Workflow — AI 多角色协作系统

> **版本**: v6.0 | **生成日期**: 2026-04-24

---

## 这是什么？

Team Workflow 是一个让 AI 按照软件工程规范协作的框架。它定义了：

- **角色分工**：Triage、Tech Lead、Dev、QA、PM、DevOps 各司其职
- **流程约束**：三层门禁确保流程不可绕过
- **两层工作流**：任务层管功能块完成，发布层管上线交付
- **产出规范**：每个任务都有完整的文档资产包

---

## 核心概念

### 甲方-乙方分离

甲方（用户）提需求、维护需求清单、验收确认；乙方（AI 团队）按流程执行、产出文档。MILESTONES 是甲方需求清单，Triage 只同步状态，不决定需求内容。

### 三层门禁

**为什么需要门禁？** 经验表明，没有强制约束时 AI 会"聪明地"绕过流程——"用户急，先做再说"、"这么简单不需要设计"、"文档后面补"。门禁把规则变成架构约束，AI 无法自我豁免，即使用户说"直接做"。

| 门禁 | 作用 | 执行时机 |
|------|------|---------|
| **入口门禁** | 所有需求必须经过 Triage 分类，拒绝直接处理 | 收到用户输入时 |
| **阶段门禁** | 每阶段必须基于上一阶段产出物才能进入 | 角色开始执行前 |
| **完成门禁** | 声称完成前必须检查所有产出物是否存在且非空 | 角色准备更新状态为 Review 时 |

### 两层工作流

**为什么分两层？** 旧模型把任务流程和发布流程混在一起（Phase 0-7），导致发布阶段的 Phase 4-7 无法被单个任务触发——一个 Feature 完成只代表功能块就绪，不代表可以上线。两层分离后，任务层和发布层独立触发、互不耦合。

**任务层**（单个 Feature/Bug/Change）：

```
Phase 0: 创建（Triage）→ task-pool 条目
Phase 1: 设计（Tech Lead）→ SPEC + AC + R1/R2/R3
Phase 2: 实现（Dev）→ 代码 + 测试 + SCOPE.md
Phase 3: 验证（QA）→ 测试报告 → 状态 Review（功能块就绪）
```

**发布层**（Milestone 级别）：

```
R-Phase 0: 就绪检查（Triage）→ 所有任务 Review/Archived？
R-Phase 1: 集成验证（QA + DevOps）→ 闭环验证
R-Phase 2: 验收放行（PM）→ PM 签字
R-Phase 3: 上线部署（DevOps）→ 生产部署 + 监控确认
```

---

## 角色与文档归属

| 角色 | 职责 | 维护文档 |
|------|------|---------|
| **甲方** | 提需求、验收 | MILESTONES（需求清单） |
| **Triage** | 分类、分发、跟踪 | Task Pool + 同步 MILESTONES 状态 |
| **Tech Lead** | 技术设计、创建资产包 | SPEC.md, AC.md, R1/R2/R3 |
| **Dev** | 编码实现、Bug 修复 | 代码、RCA.md, TEST_CASE.md |
| **QA** | 测试验证、闭环验证 | 测试报告、闭环验证报告 |
| **PM** | 验收放行 | 验收签字 |
| **DevOps** | CI/CD、部署 | 部署报告 |

**关键禁止**：
- **Triage 不创建资产包** — Triage 是通用分发工具，不了解具体业务，资产包应由 Tech Lead 创建
- **MILESTONES 归甲方** — 这是需求清单，不是执行状态，乙方不应决定需求
- **PM 是唯一放行人** — Tech Lead 和 QA 无权放行上线，确保业务验证不可跳过
- **Dev 不跳过 Tech Lead** — 不允许 Dev 直接创建资产包，确保设计先行

---

## 任务分类

| 类型 | ID 前缀 | 进 MILESTONES | 分发给 |
|------|---------|-------------|--------|
| Feature | F | ✅ | Tech Lead |
| Bug（阻断发版） | B | ✅ | Dev |
| Bug（普通） | B | ❌ | Dev |
| Change | C | ✅ 保守策略 | Tech Lead |
| Docs | D | ❌ | Tech Lead |
| Analysis | A | ❌ | Tech Lead |

### Change 保守策略

**为什么保守？** 变更可能影响交付，如果默认不更新 MILESTONES，甲方可能忽略重要变更。保守策略默认记录，由 Tech Lead 判断后反馈调整。

### Bug 阻断判断

Triage 收到 Bug 后询问甲方"是否阻断发版？"。阻断则更新 MILESTONES（⚠️ 有Bug），不阻断则只更新 Task Pool。

---

## 快速上手

### 作为用户（甲方）

1. **提需求**：`T: 我需要一个用户登录功能`
2. **确认分类**：Triage 输出分类报告，你确认
3. **等待交付**：乙方按流程执行
4. **验收确认**：Review 状态时确认产出物
5. **发布上线**：`R: M4` 进入发布流程

### 作为 AI（乙方）

1. **读取规则**：加载 `SKILL.md` → `workflows/shared.md` → 对应角色 prompt
2. **执行门禁**：入口 → 阶段 → 完成，顺序检查
3. **产出文档**：按资产包规格创建文件

---

## 版本历史

| 版本 | 日期 | 核心变更 |
|------|------|---------|
| v4.0 | 2026-04-23 | 三层门禁架构 |
| v5.0 | 2026-04-23 | 甲方-乙方分离，Triage 分发工具 |
| v5.5 | 2026-04-24 | task-pool v6.0 模板，backlog v6.0 模板 |
| **v6.0** | **2026-04-24** | **两层工作流分离（任务层 + 发布层），📌 背景标记** |

---

## 相关文档

| 文档 | 内容 | 来源 |
|------|------|------|
| [工作流程](./WORKFLOW.md) | 两层工作流详细描述 + 门禁逻辑 + 资产包规格 | shared.md |
| [开发指南](./DEV-GUIDE.md) | TDD + 交付管线 + Commit/分支规范 | dev.md |
| [设计指南](./DESIGN-GUIDE.md) | 设计流程 + ADR + 前后端架构规范 | tech-lead.md |
| [分类指南](./TRIAGE-GUIDE.md) | 分类流程 + MILESTONES 同步 | triage.md |
| [Bug修复指南](./BUGFIX-GUIDE.md) | Bug 修复流程 + RCA 模板 | bugfix.md |
| [安装指南](./INSTALL.md) | 安装步骤 + 配置示例 | examples/ + config/ |
| [架构设计](./design/ARCHITECTURE.md) | 架构决策（ADR 格式） | tech-lead.md |

<!-- Last generated: 2026-04-24 18:55 | Source: _team/SKILL.md | Hash: 4F95D290 -->