# _team 框架流程审核报告 — 第四轮
## 核心目标：消除歧义，确保 AI 严格遵照执行

**审核时间**: 2026-04-20 19:21
**审核范围**: SKILL.md + shared.md + 全部 prompts + 全部 workflows/roles + team-input-matcher.md
**前置审核**: memory/2026-04-20-P1-flow-audit-report.md（P1-P2 修复）+ memory/flow-audit-v3.md（F1-F14）

---

## 一、正确执行路径（无歧义版）

```
用户输入
    │
    ├─ SKILL.md 步骤 4 ──→ shared.md「步骤 0: Triage 检查」
    │                           │
    │                           ├─ 触发 ──→ prompts/triage.md + triage-standards.md
    │                           │              │
    │                           │              ├─ Step1 分析（5维度含复杂度）
    │                           │              ├─ Step2 分类报告（3选项确认）
    │                           │              ├─ Step3 写入（禁止跳过确认）
    │                           │              └─ 确认后 → 写 .team/task-pool.md
    │                           │
    │                           └─ 不触发 ──→ shared.md 步骤1「意图分类」
    │                                              │
    │                                       shared.md 步骤2「关键词匹配」
    │                                              │
    │                              加载 prompts/{role}.md
    │                            + workflows/roles/{role}.md
    │                                              │
    │                              按角色规范执行任务
    │                                              │
    └───────────────────────────────────→ 更新 .team/task-pool.md
```

---

## 二、核心文件一致性核查（发现新问题）

### C1: shared.md 角色表（10角色）vs SKILL.md 步骤4意图映射 vs team-input-matcher.md（5意图）

| 问题 | shared.md 角色表 | SKILL.md 步骤4 | team-input-matcher.md |
|------|---------------|--------------|---------------------|
| 数量 | 10个角色 | 8个意图 | 5个优先级 |
| 架构师映射 | 架构师→architecture-standards.md | 架构分析→framework-architect.md | ❌无架构分析 |
| prompts文件 | ❌未列出 | ✅有 | ❌无 |

**风险**: AI 从 shared.md 角色表加载了 `architecture-standards.md`，但没有加载 `framework-architect.md`。后者含详细角色定义、约束、行为模式，缺失会导致角色执行不完整。

**修复**: shared.md 角色表补充 prompts 路径：

```markdown
| **架构师** | `{TEAM_PATH}/prompts/framework-architect.md` + `{TEAM_PATH}/workflows/roles/architecture-standards.md` | 架构设计... |
```

同时在 team-input-matcher.md 补充「架构分析」和「需求分析」意图（当前只有5个，shared.md 有10类）。

---

### C2: team-input-matcher.md 与 shared.md 关键词不一致

| team-input-matcher.md（5类） | shared.md 关键词匹配（12+行） | 缺失 |
|----------------------------|--------------------------|------|
| 错误/Bug/报错... | ✅ 一致 | — |
| 不一致/差异... | ✅ 一致 | — |
| 添加/新建/实现... | ✅ 一致 | — |
| 测试/验证... | ✅ 一致 | — |
| 架构/重构... | ✅ 存在 | — |
| — | 需求/PRD/用户故事 | ❌ 缺失 |
| — | 部署/运维/Docker... | ❌ 缺失 |
| — | Review/审查... | ❌ 缺失 |

**风险**: 如果 AI 使用 team-input-matcher.md 而非 shared.md 作为匹配来源，会遗漏需求分析、部署运维、代码审查等场景。

**修复**: 统一 team-input-matcher.md 与 shared.md 的关键词列表，或明确标注「team-input-matcher.md 仅作快速参考，详细规则见 shared.md」。

---

### C3: SKILL.md 步骤 4 的 prompts 映射缺少「代码审查」和「架构分析」两个 prompt 文件路径

SKILL.md 步骤4意图→prompts 映射表：

```
| 需求分析 | prompts/pm.md |
| 测试验证 | prompts/qa-engineer.md |
| 部署运维 | prompts/devops.md |
| ← 缺少「代码审查 → prompts/tech-lead.md」|
| ← 缺少「架构分析 → prompts/framework-architect.md」|
```

**风险**: AI 匹配到「代码审查」或「架构分析」时不知道该加载哪个 prompts 文件。

**修复**: 补充：

```markdown
| 代码审查 | prompts/tech-lead.md | workflows/roles/review-standards.md |
| 架构分析 | prompts/framework-architect.md | workflows/roles/architecture-standards.md |
```

---

## 三、路径一致性核查

### P1: prompts/analysis.md 与 prompts/tech-lead.md 职责重叠

| 文件 | AI触发id | shared.md映射到 |
|------|---------|---------------|
| `prompts/analysis.md` | `analysis`（aliases: tech-lead, reference-analyst） | Tech Lead → analysis-standards.md |
| `prompts/tech-lead.md` | `tech-lead` | 代码审查 → review-standards.md |

**问题**: `analysis.md` 的 aliases 包含 `tech-lead`，意味着「分析任务」和「代码审查」都可以用 `tech-lead` 这个 id。
但 `tech-lead.md` 是单独的 prompts 文件。两者的触发关键词分别是：
- `analysis.md`: 分析/调研/对比/流程分析/参考分析
- `tech-lead.md`: Review/审查/看代码

**风险**: 如果用户说「分析代码并审查」，AI 可能混淆应该加载哪个 prompts。

**建议**: 明确 separation：在 prompts 头部加 `exclusive: true` 字段，或在 shared.md 中加「当同时匹配多个意图时，优先级为 Triage > bugfix > analysis > dev > test > review」。

---

### P2: prompts/bugfix.md 与 workflows/roles/bugfix-standards.md 触发关键词覆盖差异

| prompts/bugfix.md keywords | shared.md 步骤2 关键词 |
|--------------------------|---------------------|
| Bug/bug/报错/问题/修复/错误/异常/崩溃/闪退 | 错误/Bug/报错/panic/404/500/崩溃/闪退/异常 |
| — | 不一致/差异/偏差/对齐/样式不匹配/高度/宽度 |

**问题**: shared.md 步骤2 额外包含了「不一致/差异/偏差/对齐」等场景，这些在 `prompts/bugfix.md` 的 keywords 中没有覆盖。

**风险**: AI 匹配到 shared.md 的 bugfix 意图，但加载 `bugfix.md` 后，发现用户描述「样式不对齐」不在 keywords 中，可能误判这不是 bugfix 场景。

**修复**: 在 `prompts/bugfix.md` 的 keywords 中补充：

```yaml
keywords:
  # 行为偏差
  - 不一致
  - 差异
  - 偏差
  - 对齐
  - 样式不匹配
  - 高度不对
  - 宽度不对
```

---

### P3: `{docs_internal}` 占位符替换时机的歧义

| 位置 | 占位符出现位置 | 替换时机 |
|------|-------------|---------|
| SKILL.md 步骤2 | "读取 `{PROJECT_PATH}/.team/project.md`" | ✅ 明确 |
| shared.md | "路径值由工具配置（环境变量）注入" | ⚠️ 歧义 |
| 各规范文件 | `{docs_internal}/requirements/...` | ⚠️ AI 可能不替换 |

**歧义点**: shared.md 说「路径值由工具配置注入」，但：
1. 工具配置在 `.trae/rules/SKILL.md` 中（示例文件）
2. AI 在不同工具中可能无法访问该配置
3. 各规范文件（如 development-standards.md）用 `{docs_internal}` 但没有说明 AI 如何获取实际值

**风险**: AI 可能保留 `{docs_internal}` 字面量，而不是替换为实际路径。

**修复**: 在 SKILL.md 步骤 2「加载/完善项目基础信息」之后，补充：

```markdown
### 2.3 路径变量提取（强制执行）

在读取 project.md 后，立即提取以下变量存入内存备用：
- `{docs_internal}` → project.md 的 `docs_internal` 字段
- `{docs_external}` → project.md 的 `docs_external` 字段

在后续所有文档路径中，用实际值替换占位符。
如果 project.md 中字段为空或缺失，询问用户确认路径。
```

---

## 四、Triage 流程一致性

### T1: prompts/triage.md 的「触发时机」与 shared.md「步骤 0」的差异（延续 F2）

已在上轮报告中指出（F2），本轮确认仍未修复：

| 差异行 | prompts/triage.md | shared.md 步骤0 |
|--------|-----------------|----------------|
| 「用户抱怨但没指明问题类型」| ✅ 有 | ❌ 无（只有「笼统抱怨」）|

**注**: 「笼统抱怨」和「用户抱怨但没指明问题类型」语义相近，但措辞不同，AI 可能判断为不同触发条件。

**修复**: 将 shared.md 步骤0「笼统抱怨」改为与 prompts/triage.md 一致的表述：

```markdown
| 笼统抱怨/需要澄清 | 没指明问题类型 | 需要澄清或 Triage 拆解 |
```

---

### T2: prompts/triage.md 的「触发时机」表格第5行「需要澄清」不是触发条件

```markdown
| 需要澄清 | 没指明问题类型 | 需要澄清或 Triage 拆解 |
```

**歧义**: 「需要澄清」是 Triage 的结果，不是触发条件。它描述的是「当用户没指明问题时，Triage 应该反问」，而不是「满足什么条件才触发 Triage」。

**风险**: 如果用户说「我需要澄清一下」，AI 可能误判应该触发 Triage。

**修复**: 将该行改为触发条件：

```markdown
| 模糊意图 | 无法明确归类到单一角色 | Triage 拆解后归类 |
```

删除「需要澄清」行。

---

## 五、执行节点细化（确保 AI 无误解）

### E1: SKILL.md 步骤4「步骤 0-6」的执行顺序歧义

SKILL.md 步骤4原文：

```
步骤 4: 加载步骤 2 匹配的 prompts 文件
步骤 5: 加载步骤 2 匹配的 workflows/roles 规范文件
步骤 6: 按加载的规范开始执行任务
```

**歧义**: 「步骤 0-3」和「步骤 4-6」是并列关系还是包含关系？
当前表述像是「先完成步骤 0-3，再执行步骤 4-6」。

但实际上 Triage 触发时需要递归：Triage → 写任务池 → 重新执行匹配。

**修复**: 明确为决策树：

```markdown
步骤 4: 输入匹配（循环执行）
  │
  ├─ 步骤 0: Triage 检查
  │       ├─ 触发 → 执行 Triage → 用户确认 → 写 task-pool
  │       │         → 重新执行步骤 0（匹配剩余任务）
  │       └─ 不触发 → 继续步骤 1
  │
  ├─ 步骤 1: 意图分类（见 shared.md）
  ├─ 步骤 2: 关键词匹配 → 步骤 4 加载 prompts + 步骤 5 加载规范
  ├─ 步骤 3: 兜底处理
  │
  └─ 步骤 6: 按规范执行
```

---

### E2: shared.md「步骤 3: 兜底处理」的三条规则没有明确优先级

```markdown
1. **默认**: 功能开发 → development
2. **如果包含错误码**: 404/500/502/503/panic/crash → bugfix
3. **如果包含问号**: 提问 → analysis
```

**歧义**: 如果输入同时满足多条（如「404错误？帮我分析下」），AI 按哪个处理？

**修复**: 明确优先级：

```markdown
### 步骤 3: 兜底处理

**优先级（从上到下递减）**：
1. **错误码优先**: 如果包含 404/500/502/503/panic/crash → bugfix
2. **问号其次**: 如果包含问号且无错误码 → analysis
3. **默认**: 其他 → 功能开发 → development

**匹配示例**：
- 「404错误？」（同时满足1+2）→ bugfix（错误码优先）
- 「帮我分析下架构？」（只满足2）→ analysis
- 「添加一个按钮」（只满足3）→ development
```

---

### E3: 初始化「步骤 4」的任务写入与 Triage「禁止跳过确认」的关系

shared.md 初始化步骤4：

```
步骤 4: 将当前任务写入任务池（ID 自动递增）
```

**歧义**: 这与 Triage 的「禁止跳过用户确认直接写 task-pool」是否矛盾？

初始化阶段写入 task-pool（TASK-001）是记录「当前任务」，
而 Triage 阶段是「分析后拆分的新任务需要用户确认」。

**修复**: 在 shared.md 初始化步骤4加注释：

```markdown
步骤 4: 将当前任务写入任务池（ID 自动递增）
         ↑ 注意：这是记录当前任务，用于追踪
         ↑ 与 Triage 阶段不同（Triage 的新任务需要用户确认后才写）
```

---

## 六、自测命令的动态检测（延续 F6）

### S1: dev-workflow.md Phase 3 的命令说明与 SKILL.md 步骤5不完全一致

| 文件 | 命令检测说明 | 状态 |
|------|------------|------|
| SKILL.md 步骤5 | 检测5种环境（bun/pnpm/yarn/npm/go） | ✅ 完整 |
| dev-workflow.md Phase 3 | Go + Frontend（含动态替换说明） | ✅ 完整 |
| development-standards.md section 8 | ⚠️ 只有Go示例表格，Frontend用「动态替换」 | ⚠️ 部分 |
| bugfix-standards.md section 7 | ⚠️ 只有Go示例命令 | ❌ 不完整 |
| shared-implementation.md | ⚠️ 只有Go命令示例 | ❌ 不完整 |

**风险**: AI 在 bugfix 或 shared-implementation 阶段可能忘记动态检测。

**修复**: 在 bugfix-standards.md 和 shared-implementation.md 的自测章节开头统一加：

```markdown
> ⚠️ **命令动态检测（强制）**: 必须先执行 SKILL.md 步骤5的命令检测逻辑，
> 根据检测到的包管理器动态替换以下命令。不得直接使用Go示例命令！
```

---

## 七、优先级排序（修复顺序建议）

### 🔴 P0（立即修复，影响执行正确性）

| # | 问题 | 涉及文件 |
|---|------|---------|
| P0-1 | shared.md 角色表缺 prompts 路径（导致架构师等角色缺少角色定义） | shared.md |
| P0-2 | SKILL.md 步骤4 缺「代码审查」「架构分析」的 prompts 路径 | SKILL.md |
| P0-3 | `{docs_internal}` 替换时机不明（AI可能保留字面量） | SKILL.md（补充步骤2.3）|
| P0-4 | team-input-matcher.md 关键词不全（缺失需求/部署/审查） | team-input-matcher.md |

### 🟡 P1（近期修复，影响执行完整性）

| # | 问题 | 涉及文件 |
|---|------|---------|
| P1-1 | prompts/bugfix.md 缺「不一致/差异」关键词 | prompts/bugfix.md |
| P1-2 | prompts/triage.md「需要澄清」行不是触发条件（应删除） | prompts/triage.md |
| P1-3 | SKILL.md 步骤4 的执行顺序歧义（Triage 递归） | SKILL.md |
| P1-4 | shared.md 步骤3 兜底规则无优先级 | shared.md |

### 🟠 P2（优化，无阻塞性）

| # | 问题 | 涉及文件 |
|---|------|---------|
| P2-1 | development-standards.md / shared-implementation.md 自测命令缺动态检测说明 | development-standards.md, shared-implementation.md |
| P2-2 | prompts/analysis.md 与 prompts/tech-lead.md 职责边界（aliases 重叠） | prompts/analysis.md |
| P2-3 | shared.md「步骤4 任务写入」与 Triage「禁止跳过确认」的关系说明 | shared.md |

---

## 八、静态检查清单

### 路径检查

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | shared.md 角色表含 prompts 路径 | ❌ 缺失（只有规范路径）|
| 2 | SKILL.md 步骤4 意图→prompts 映射完整（10个意图） | ❌ 缺失2个（代码审查/架构分析）|
| 3 | team-input-matcher.md 与 shared.md 关键词一致 | ❌ 不一致（缺3类）|
| 4 | prompts/*.md 含 `{TEAM_PATH}/` 绝对路径 | ✅ 全部正确 |
| 5 | workflows/roles/*.md 含 `{TEAM_PATH}/` 绝对路径 | ✅ 全部正确 |
| 6 | 无 `\{TEAM_PATH\}/\{TEAM_PATH\}/` 双路径残留 | ✅ 无残留 |

### 触发条件检查

| # | 检查项 | 结果 |
|---|--------|------|
| 7 | shared.md 步骤0 Triage 触发条件与 prompts/triage.md 一致 | ❌ 1处差异（笼统抱怨 vs 用户抱怨）|
| 8 | prompts/triage.md 无非触发条件的行 | ❌ 1行（「需要澄清」）|
| 9 | prompts/bugfix.md keywords 覆盖 shared.md bugfix 关键词 | ❌ 缺「不一致/差异」等 |
| 10 | triage-standards.md 的复杂度评估与 prompts/triage.md 一致 | ✅ 一致 |

### 执行流程检查

| # | 检查项 | 结果 |
|---|--------|------|
| 11 | SKILL.md 步骤4 明确 Triage 递归处理 | ❌ 表述为顺序执行 |
| 12 | shared.md 步骤3 兜底规则有明确优先级 | ❌ 无优先级 |
| 13 | 各规范文件自测章节有动态检测说明 | ❌ 部分文件缺失 |
| 14 | prompts/analysis.md 与 prompts/tech-lead.md 职责分离 | ⚠️ aliases 重叠 |

---

## 九、修复后的正确执行路径（目标版）

```
用户输入
    │
    ├─ SKILL.md 步骤 2.3 ──→ 提取 docs_internal / docs_external
    │
    ├─ SKILL.md 步骤 4 ──→ shared.md「步骤 0: Triage 检查」
    │                           │
    │                           ├─ 触发 ──→ prompts/triage.md（复杂度必填）
    │                           │              + triage-standards.md（执行流程）
    │                           │              │
    │                           │              ├─ Step1 分析（5维度+复杂度）
    │                           │              ├─ Step2 分类报告（3选项确认，禁止跳过）
    │                           │              └─ Step3 写入（确认后才写）
    │                           │
    │                           └─ 不触发 ──→ shared.md 步骤1「意图分类」
    │                                              │
    │                                       shared.md 步骤2「关键词匹配」
    │                                       （含 prompts 路径，10类全覆盖）
    │                                              │
    │                                       shared.md 步骤3「兜底处理」
    │                                       （优先级：错误码 > 问号 > 默认）
    │                                              │
    │                              加载 prompts/{role}.md
    │                            + workflows/roles/{role}.md
    │                            + shared-implementation.md（dev通用）
    │                                              │
    │                              按规范执行（动态命令检测）
    │                                              │
    └───────────────────────────────────→ 更新 .team/task-pool.md
```

---

## 十、版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v4 | 2026-04-20 | 本轮审核：新增 C1-C3（一致性）、T1-T2（Triage）、E1-E3（执行节点）、S1（自测命令）|
| v3 | 2026-04-20 | 新增 F1-F14（路径、触发、复杂度、Triage、命令）|
| v2 | 2026-04-20 | 第二轮：S1-S9 结构矛盾修复 |
| v1 | 2026-04-20 | 第一轮：P0-P2 路径修复 |
