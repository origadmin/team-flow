# P0 流程重构提案：Triage 升级为唯一入口，废弃关键词匹配

**版本**: v1.1
**日期**: 2026-04-20
**作者**: AI + User 讨论
**状态**: ✅ 已执行
**优先级**: P0（框架级架构变更）

---

## 1. 问题描述

### 1.1 关键词匹配的两套系统问题

当前框架存在两套独立定义的关键词系统：

| 系统 | 定义位置 | 用途 | 参与匹配？ |
|------|---------|------|----------|
| 系统 A | `shared.md` 步骤2（12行表格） | AI 执行输入匹配时使用 | ✅ 是 |
| 系统 B | 各 `prompts/*.md` 的 `keywords:` 区 | 角色行为参考 | ❌ 否 |

**证据**: SKILL.md 步骤4 写的是「按 `shared.md` 步骤2 关键词匹配」，而不是「按各 prompt 的 keywords 匹配」。

**后果**: prompts 文件的 `keywords:` 区是装饰性注释——定义了但从不被匹配流程使用。

### 1.2 关键词歧义问题

| 歧义词 | 所属意图 | 冲突 |
|--------|---------|------|
| 「差异」 | bugfix（行1）| 「差异」也属于参考分析意图 |
| 「Bug」 | bugfix | 「Bug」也出现在 qa-engineer.md / dev.md |
| 「架构」 | 架构分析 | 同时出现在 tech-lead.md + framework-architect.md |

**典型错误场景**:
```
用户：「帮我分析下 A 和 B 的差异」
预期：参考分析 → analysis
实际：命中 bugfix 行1「差异」→ bugfix ❌
```

### 1.3 team-input-matcher.md 与 shared.md 不同步

`team-input-matcher.md` 只覆盖 5 类意图，但 `shared.md` 有 10+ 类。

| team-input-matcher.md（5类）| shared.md 有无 |
|--------------------------|--------------|
| 问题排查 | ✅ 有 |
| 功能开发 | ✅ 有 |
| 测试验证 | ✅ 有 |
| 架构设计 | ✅ 有 |
| 需求分析 | ❌ 缺失 |
| — | ❌ 流程分析（缺失）|
| — | ❌ 阶段推进（缺失）|
| — | ❌ 部署运维（缺失）|
| — | ❌ 代码审查（缺失）|

### 1.4 Triage 触发条件过于保守

当前 `shared.md` 步骤0 Triage 仅在以下情况触发：
- 多事项 / 模糊意图 / 跨多角色 / 过长输入 / 笼统抱怨

**问题**: 单意图但关键词模糊的输入（如「有个问题」「帮我看看」「感觉不对」）无法被正确分类，只能走兜底逻辑（默认 development），导致路由错误。

---

## 2. 建议方案

### 2.1 核心改动：Triage 作为唯一入口

```
用户输入 → Triage（100%入口，无例外）
         → AI 分析 + 语义分类
         → 输出分类报告（任务类型 / 角色 / 优先级 / 文档路径）
         → 用户确认 [A 正确 / B 修改 / C 重分]
         → 加载 prompts + workflow → 执行
```

**对比**:
| 方面 | 关键词方案 | Triage 统一入口 |
|------|---------|--------------|
| 匹配准确性 | 依赖关键词覆盖度，容易漏/冲突 | AI 语义理解，更准确 |
| 维护成本 | 两套系统，修改要同步 | 一套分类模板 |
| 用户控制 | 隐式路由 | 显式确认 |
| SKILL.md 复杂度 | 6步（含关键词匹配）| 2-3步 |
| 新场景覆盖 | 要补充关键词 | 自动理解 |
| 单意图误路由 | 兜底逻辑脆弱 | 始终先分析再执行 |

### 2.2 分类体系（替代关键词表）

不再用关键词表，改用 **任务类型枚举**：

| 任务类型 | 负责角色 | 对应 workflow |
|---------|---------|-------------|
| `feature` | Dev | development-standards.md |
| `bugfix` | Bugfix | bugfix-standards.md |
| `analysis` | Analysis | analysis-standards.md |
| `requirement` | PM | requirements-standards.md |
| `test` | QA | test-standards.md |
| `deploy` | DevOps | devops-standards.md |
| `review` | Tech Lead | review-standards.md |
| `architecture` | Framework Architect | architecture-standards.md |
| `mixed` | Triage | 多角色协作，拆分任务池 |

### 2.3 分类报告模板

Triage 执行后输出如下报告，用户确认后才进入执行阶段：

```markdown
# 🔍 分类报告

**输入摘要**: {用户输入的简要复述}

**任务类型**: {feature / bugfix / analysis / requirement / test / deploy / review / architecture / mixed}
**负责角色**: {Dev / Bugfix / Analysis / PM / QA / DevOps / Tech Lead / Framework Architect}
**复杂度**: {高 / 中 / 低}
**优先级**: {P0 / P1 / P2 / P3}

**建议文档路径**: 
- 如果是 feature: `{docs_internal}/requirements/{feature}/SPEC.md`
- 如果是 bugfix: `{docs_internal}/bugs/BUG-{序号}.md`
- 如果是 analysis: `{docs_internal}/analysis/{topic}/REPORT.md`
- 如果是 requirement: `{docs_internal}/requirements/{feature}/PRD.md`
- 如果是 test: `{docs_internal}/reports/{feature}/TEST.md`

**匹配规范**:
- workflow: `{TEAM_PATH}/workflows/roles/{xxx}-standards.md`
- prompts: `{TEAM_PATH}/prompts/{xxx}.md`

---

**⚠️ 请确认以上分类是否正确？**

[A] ✅ 正确，开始执行
[B] 🔄 修改（指出哪里需要调整）
[C] ❌ 全部重分
```

---

## 3. 改动范围

| 文件 | 当前状态 | 改动 | 优先级 |
|------|---------|------|--------|
| `SKILL.md` 步骤4 | 有意图分类+关键词匹配 | 删除步骤1-3，简化为「输入→Triage→确认→加载→执行」 | 🔴 P0 |
| `shared.md` | 步骤0-3（含关键词表）| 删除步骤0-3，改写为「Triage 统一入口说明」 | 🔴 P0 |
| `prompts/triage.md` | 仅处理多意图/模糊 | 升级为唯一入口，补充分类报告模板 | 🔴 P0 |
| `prompts/team-input-matcher.md` | 独立匹配器（5类）| 废弃，合并到 triage.md 或删除 | 🟠 P1 |
| `prompts/dev.md` `keywords:` | 装饰性 | 保留（作为角色描述的一部分，不参与匹配）| 🟡 P2 |
| `prompts/analysis.md` `keywords:` | 装饰性 | 保留（同上）| 🟡 P2 |
| `prompts/bugfix.md` `keywords:` | 装饰性 | 保留（同上）| 🟡 P2 |
| `...` 其余 prompts/*.md | 装饰性 | 保留（同上）| 🟡 P2 |

**注意**: `prompts/*.md` 的 `keywords:` 区虽然不参与匹配，但保留作为角色定义文档的一部分（描述角色职责边界），而非删除。

---

## 4. 分类报告规范（提案）

### 4.1 文件命名规范

```
{YYYY-MM-DD}-P{priority}-{category}-{title}.md
```

| 字段 | 说明 | 示例 |
|------|------|------|
| `YYYY-MM-DD` | 审核/变更日期 | 2026-04-20 |
| `P{priority}` | P0=框架级 / P1=规范级 / P2=文档级 | P0 |
| `category` | flow-audit / kw-audit / design / decision / refactor | design |
| `title` | 简短标题，用英文 | triage-unified-entry |

### 4.2 存放位置

| 类型 | 路径 |
|------|------|
| 流程/架构变更报告 | `{TEAM_PATH}/memory/` |
| 持续改进建议（未执行） | `{TEAM_PATH}/memory/proposals/` |
| 已执行变更 | `{TEAM_PATH}/memory/` + 对应文件同步更新 version 字段 |
| 对外公开的变更记录 | `{docs_internal}/_docs/_team/CHANGELOG.md` |

### 4.3 报告文件模板

```markdown
# {TITLE}

**版本**: v1.0
**日期**: {YYYY-MM-DD}
**作者**: {AI / User}
**状态**: [提议中 / 已批准 / 已执行 / 已回滚]
**优先级**: P0/P1/P2

## 问题描述
{当前框架的问题}

## 建议方案
{改进建议}

## 改动范围
| 文件 | 改动 | 状态 |
|------|------|------|
| ... | ... | [待改/已改] |

## 影响评估
{对现有流程的影响}

## 讨论记录
- {YYYY-MM-DD HH:mm}: {内容}

## 决策结果
{用户最终决定 + 理由}

## 执行记录
- {YYYY-MM-DD}: 执行了 {改动内容}
```

---

## 5. 影响评估

### 5.1 正面影响
- ✅ 消除关键词歧义和两套系统问题
- ✅ 用户主导路由，降低 AI 错误路由风险
- ✅ 新场景无需维护关键词表
- ✅ SKILL.md 大幅简化

### 5.2 潜在风险
- ⚠️ 每个输入都多一步「分类确认」，轻微增加交互成本
- ⚠️ Triage AI 需要足够强的分类能力（依赖模型能力）
- ⚠️ 用户如果不习惯确认流程，可能觉得繁琐

### 5.3 缓解措施
- 高频简单任务（如「帮我加个接口」「修个Bug」）可以默认「A」快速确认
- 可选：增加 `::fast` 标签跳过确认，直接执行

---

## 6. 决策

**等待用户确认后执行**

用户选项：
- **[A]** ✅ 同意，按此方案执行
- **[B]** 🔄 部分修改（指出需要调整的地方）
- **[C]** ❌ 不同意，说明原因

---

## 7. 执行记录

- **2026-04-20 22:27**: 用户确认，开始执行
- **2026-04-20 22:30**: 执行完成，共修改 5 个文件

| 文件 | 改动 | 状态 |
|------|------|------|
| `SKILL.md` | v3.2 → v3.3，删除步骤1-3关键词匹配，简化为4步Triage流程 | ✅ 已改 |
| `workflows/shared.md` | v2.1 → v3.0，删除步骤0-3（意图分类+关键词匹配+兜底），新增Triage统一入口章节+任务类型枚举 | ✅ 已改 |
| `prompts/triage.md` | v1.0 → v2.0，从「仅处理模糊输入」升级为「唯一入口」，新增语义分析流程、分类决策指南、分类报告格式 | ✅ 已改 |
| `workflows/roles/triage-standards.md` | v1.0 → v2.0，废弃关键词匹配，新增语义分析流程，更新分类报告格式 | ✅ 已改 |
| `prompts/team-input-matcher.md` | v1.0 → 废弃，保留原始内容作为历史参考 | ✅ 已改 |
