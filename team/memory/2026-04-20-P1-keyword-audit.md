# _team 框架关键词审核报告

**审核时间**: 2026-04-20 21:44
**审核范围**: shared.md 步骤2 + 全部 10 个 prompts/*.md
**核心问题**: 存在 **两套独立关键词系统**，且各 prompt 文件的关键词未参与输入匹配

---

## 一、关键词完整分布表

### 1. shared.md（输入匹配用，AI 实际执行匹配时读取）

#### 步骤 0: Triage 触发条件（5 条）

| 条件 | 说明 | 复杂度 |
|------|------|--------|
| 多事项 | 一段话包含多个独立任务 | 高 |
| 模糊意图 | 无法明确归类 | 高 |
| 跨多角色 | 涉及 ≥2 个不同角色 | 高 |
| 过长输入 | 一次说了一段话（>3 句） | 中高 |
| 笼统抱怨 | 没指明问题类型 | 高 |

#### 步骤 1: 意图分类（10 类，不含关键词）

| 意图类别 | 复杂度 | 特征 | 默认角色 |
|---------|--------|------|----------|
| 问题排查 | 中 | 错误码、报错、异常、Bug | bugfix |
| 功能开发 | 中 | 新功能、接口、实现、添加 | development |
| 参考分析 | 高 | 对比、差异、参考、借鉴、调研 | analysis |
| 流程分析 | 高 | 分析流程、业务流程、制定流程 | analysis |
| 阶段推进 | 中 | 按这个方案来、开始实现、通过 | analysis→dev |
| 架构分析 | 高 | 技术选型、ADR、架构设计 | analysis |
| 需求分析 | 中高 | 需求、PRD、功能描述 | requirements |
| 测试验证 | 中 | 测试、验证、QA、回归 | test |
| 部署运维 | 中 | 部署、运维、Docker、K8s、集群 | devops |
| 代码审查 | 高 | Review、审查、看代码 | tech-lead |

#### 步骤 2: 关键词匹配（12 行）

| 意图 | 关键词 |
|------|--------|
| Bug fix (行1) | 错误/Bug/报错/panic/404/500/**崩溃**/**闪退**/异常/**不一致**/**差异**/**偏差**/**对齐**/**样式不匹配**/**高度**/**宽度** |
| Bug fix (行2) | 有问题/出了点问题/好像不对 |
| Development (行3) | 开发/实现/添加/新建/创建/接口/后端/前端 |
| Development (行4) | TDD/测试驱动 |
| 流程分析 | 分析流程/业务流程/制定流程/设计流程/测试流程/完整流程/流程测试/mock/覆盖全部 |
| 参考分析 | 对比/差异/参考/借鉴/调研/看看别人的/分析xxx的 |
| 阶段推进 | 按这个方案来/开始实现/通过/可以/审核通过/继续/按方案执行/开始BDD/开始TDD |
| 架构分析 | 架构/技术选型/ADR/方案/设计 |
| 需求分析 | 需求/PRD/功能/产品/业务 |
| 测试验证 | 测试/验证/QA/回归 |
| 部署运维 | 部署/运维/Docker/K8s/集群 |
| 代码审查 | Review/审查/看代码 |

---

### 2. prompts/*.md 各文件关键词

#### prompts/dev.md

```yaml
# 通用
开发/实现/功能/修复/Bug/TDD
# 后端
后端/API/gRPC/微服务/Go/业务逻辑
# 前端
前端/React/组件/UI/页面/前端开发
# 移动端
Android/Kotlin/移动端/App/iOS/Swift/原生开发
```

#### prompts/analysis.md

```yaml
# 分析
分析/调研/对比/差异/参考/借鉴
# 流程
流程分析/业务流程/制定流程/设计流程
# 业务
业务分析/需求分析/市场分析
```

#### prompts/bugfix.md

```yaml
Bug/bug/报错/问题/修复/错误/异常/崩溃/闪退
404/500/502/503/panic/crash
```

#### prompts/pm.md

```yaml
需求/PRD/用户故事/验收标准/功能/产品
```

#### prompts/qa-engineer.md

```yaml
测试/QA/Bug/验证/E2E/场景/Gherkin
```

#### prompts/devops.md

```yaml
部署/CI/CD/Docker/K8s/运维/监控
```

#### prompts/tech-lead.md

```yaml
架构/技术方案/代码审查/ADR/技术选型/设计
```

#### prompts/framework-architect.md

```yaml
框架/架构/架构师/模块设计/技术决策
```

#### prompts/triage.md（无 YAML keywords 区，正文定义触发时机）

```yaml
场景: 用户一次说多件事/意图模糊/跨多个角色/用户抱怨但没指明问题类型/用户给了一大段话
```

#### team-input-matcher.md

```yaml
# Step 2 关键词（仅 5 类，缺 7 类）
错误/Bug/报错/panic/崩溃/闪退/异常
不一致/差异/偏差/对齐/样式不匹配
添加/新建/实现/开发/创建
测试/验证/QA/回归
架构/重构/设计模式  ← 不同于 shared.md 的"架构/技术选型/ADR/方案/设计"
```

---

## 二、两套关键词系统的问题

### K1: 两套关键词系统独立存在（核心结构问题）

| 系统 | 定义位置 | 用途 | 实际参与输入匹配？ |
|------|---------|------|-----------------|
| **系统 A** | shared.md 步骤 2（12 行） | AI 输入匹配时使用 | ✅ 是 |
| **系统 B** | 各 prompts/*.md 的 `keywords:` 区 | AI 角色行为触发/参考 | ❌ 否（几乎不参与匹配） |

**证据**: SKILL.md 步骤 4 写的是「按 shared.md 步骤 2 关键词匹配」，而不是「按各 prompt 的 keywords 匹配」。

**问题**: prompts 文件中的 `keywords:` 区相当于装饰性注释——定义了但从不被使用。
这导致：

1. prompts/analysis.md 的 keywords 含「市场分析」，但 shared.md 无此场景
2. prompts/qa-engineer.md 的 keywords 含「Bug/Gherkin」，但 shared.md 测试验证意图的关键词只有「测试/验证/QA/回归」
3. prompts/dev.md 的 keywords 含「Android/Swift」，但 shared.md 没有任何移动开发意图
4. prompts/tech-lead.md 的 keywords 含「ADR」，但 shared.md 的架构分析意图里 ADR 是词而不是触发词

---

### K2: "差异" 在两处意图中重复（匹配歧义）

| shared.md 位置 | 意图 | 问题 |
|---------------|------|------|
| Bug fix 行1 | 关键词含「差异」 | 样式不匹配、偏差 |
| 参考分析 行 | 关键词含「差异」 | 对比分析场景 |

**歧义场景**: 用户说「帮我分析下 A 和 B 的差异」

- shared.md 匹配：先看 bugfix 行1（优先级取决于表格从上到下），命中「差异」→ bugfix ❌
- 正确理解：参考分析/对比分析 → analysis ✅

**修复**: 从 bugfix 行1 删除「差异」，因为「不一致/差异/偏差/对齐」在 bugfix 行1 的语义不清：

```
# 当前（歧义）
不一致/差异/偏差/对齐/样式不匹配/高度/宽度/不一致

# 修复后（明确是行为偏差）
不一致/偏差/对齐/样式不匹配/高度不对/宽度不对
```

---

### K3: "Bug" 在两处意图中重复（关键词过度覆盖）

| shared.md 位置 | 意图 | 问题 |
|---------------|------|------|
| Bug fix 行1 | 关键词含「Bug」 | ✅ 合理 |
| 参考分析 行 | 关键词含「参考/借鉴/调研」 | 无 Bug |
| prompts/qa-engineer.md | keywords 含「Bug」 | QA 也处理 Bug |
| prompts/dev.md | keywords 含「Bug」 | Dev 也处理 Bug |

**歧义场景**: 用户说「有个 Bug，帮我看看怎么参考 XX 项目」

shared.md 匹配 bugfix → 忽略后半句 → 遗漏参考分析需求

**建议**: 关键词不唯一属于某个意图，AI 匹配后应检查是否还有未匹配的剩余意图。

---

### K4: team-input-matcher.md 的 5 类 vs shared.md 的 10+ 类

| team-input-matcher.md（5 类） | shared.md 有无 |
|------------------------------|--------------|
| 问题排查 | ✅ 有（2 行 bugfix） |
| 功能开发 | ✅ 有（2 行 development） |
| 测试验证 | ✅ 有 |
| 架构设计 | ✅ 有（架构分析） |
| 需求分析 | ❌ 缺失 |
| — | ✅ 流程分析（缺失） |
| — | ✅ 阶段推进（缺失） |
| — | ✅ 部署运维（缺失） |
| — | ✅ 代码审查（缺失） |

**问题**: team-input-matcher.md 只能覆盖 5/10 场景。如果 AI 使用此文件作为匹配来源，5 个场景直接被遗漏。

**当前状态**: team-input-matcher.md 与 shared.md 是两个不同版本的匹配器。

---

### K5: prompts/analysis.md 的「流程分析」与 shared.md 的「流程分析」含义不同

| 文件 | 「流程分析」的含义 |
|------|-----------------|
| prompts/analysis.md | 分析流程 → 找出问题 → 制定改进方案 |
| shared.md 步骤2「流程分析」 | 分析流程/业务流程/制定流程/设计流程/测试流程... |

**歧义**: prompts/analysis.md 在「分析类型」正文里定义了三种分析：参考、流程整理（样式不对/有bug）、实现调整。但 shared.md 的「流程分析」意图只有第一种（参考分析类），而后两种分别走 bugfix 和 development。

AI 如果只读 prompts/analysis.md 而不参考 shared.md，会误解 shared.md 的「流程分析」意图。

---

### K6: prompts/qa-engineer.md 的「Bug/Gherkin」关键词在 shared.md 测试验证意图中缺失

| prompts/qa-engineer.md keywords | shared.md 测试验证意图 |
|--------------------------------|---------------------|
| 测试/QA/**Bug**/验证/E2E/场景/**Gherkin** | 测试/验证/QA/回归 |

**问题**: prompts/qa-engineer.md 的 keywords 含「Bug」，但 shared.md 的 bugfix 意图已覆盖「Bug」，测试验证意图不包含「Bug」。「Gherkin」是 QA 核心关键词，但在 shared.md 任何意图中都不存在。

**影响**: 用户说「用 Gherkin 写个场景」→ shared.md 无匹配 → 兜底 → development → 错过 QA 角色

---

### K7: prompts/dev.md 的 keywords 与 prompts/bugfix.md 的 keywords 有重叠但不完全一致

| prompts/bugfix.md | prompts/dev.md keywords | 差异 |
|-------------------|----------------------|------|
| Bug/bug/报错 | Bug/TDD | dev 缺报错 |
| 崩溃/闪退 | 缺失 | dev 缺移动端异常 |
| 404/500/502/503/panic/crash | 缺失 | dev 缺错误码 |
| — | 开发/实现/添加/新建 | dev 独有线 |
| — | React/组件/Android/Swift | dev 独有平台 |

**问题**: prompts/dev.md 的 constraints 写了「加载 bugfix-standards.md」，但 dev.md 自己的 keywords 并不包含 bugfix.md 的全部关键词。

---

### K8: prompts/framework-architect.md 与 prompts/tech-lead.md 关键词高度重叠

| prompts/tech-lead.md | prompts/framework-architect.md | 重叠度 |
|---------------------|-------------------------------|--------|
| 架构/技术方案/代码审查/ADR/技术选型/设计 | 框架/架构/架构师/模块设计/技术决策 | 高 |

两者都有「架构/技术选型/技术决策」，区别在于：
- tech-lead.md 额外有「代码审查/ADR」
- framework-architect.md 额外有「框架/架构师/模块设计」

但 shared.md 的意图映射中：
- tech-lead → 代码审查（review-standards.md）
- 架构分析 → framework-architect.md（architecture-standards.md）

**歧义**: prompts/analysis.md 的 aliases 包含 `tech-lead`，shared.md 步骤1「代码审查」映射到 tech-lead，但 shared.md 步骤2「架构分析」映射到 framework-architect.md。

如果用户说「帮我架构一下代码并审查」，AI 应该触发 framework-architect 还是 tech-lead？

---

### K9: shared.md「阶段推进」关键词无对应 prompts keywords 区

| shared.md 阶段推进关键词 |
|----------------------|
| 按这个方案来/开始实现/通过/可以/审核通过/继续/按方案执行/开始BDD/开始TDD |

**问题**: 无任何 prompts/*.md 文件在自己的 keywords 区定义这些词。

但 SKILL.md 步骤4「意图→prompts 映射」对「阶段推进」映射到 `prompts/dev.md`。
所以如果用户说「开始实现」，AI 会匹配 development → prompts/dev.md。

dev.md 的 keywords 含「开发/实现/添加」，这可以覆盖「开始实现」。但 dev.md 的 keywords 里没有 BDD/TDD（dev.md 有「TDD」，但那是通用关键词区，不是从 shared.md 复制的）。

**建议**: 在 prompts/dev.md 的 keywords 区补充：

```yaml
# 阶段推进（来自 shared.md 步骤2）
- 开始实现
- 审核通过
- 按方案执行
- 开始BDD
- 开始TDD
```

---

## 三、关键词覆盖矩阵

### 各场景在各文件的关键词定义状态

| 场景 | shared.md 步骤2 | prompts keywords | team-input-matcher | 状态 |
|------|---------------|----------------|-------------------|------|
| Bug fix | ✅ 有（行1+2） | ✅ bugfix.md | ✅ 有 | ✅ 完整 |
| 功能开发 | ✅ 有（行3+4） | ✅ dev.md | ✅ 有 | ✅ 完整 |
| 参考分析 | ✅ 有 | ⚠️ analysis.md（覆盖） | ❌ 无 | ⚠️ 缺失「调研」 |
| 流程分析 | ✅ 有 | ⚠️ analysis.md（含义不同）| ❌ 无 | ⚠️ 不一致 |
| 阶段推进 | ✅ 有 | ❌ 无 | ❌ 无 | ❌ 缺失 |
| 架构分析 | ✅ 有 | ⚠️ tech-lead.md + framework-architect.md（重叠）| ⚠️ 有（但缺ADR）| ⚠️ 重叠 |
| 需求分析 | ✅ 有 | ✅ pm.md | ❌ 无 | ✅ 完整 |
| 测试验证 | ✅ 有 | ⚠️ qa-engineer.md（缺 Bug/Gherkin）| ✅ 有 | ⚠️ 不完整 |
| 部署运维 | ✅ 有 | ✅ devops.md | ❌ 无 | ✅ 完整 |
| 代码审查 | ✅ 有 | ⚠️ tech-lead.md | ❌ 无 | ⚠️ 重叠 |
| Triage | ✅ 步骤0 | ✅ triage.md | ✅ 有（步骤0）| ✅ 完整 |
| 移动端开发 | ❌ 无 | ✅ dev.md | ❌ 无 | ⚠️ shared.md 缺失 |

---

## 四、修复优先级

### 🔴 K0: 两套关键词系统合一（最高优先级）

**问题**: prompts/*.md 的 keywords 区和 shared.md 步骤2 的关键词是独立的两套。

**决策**: 选定一套作为「权威关键词定义」，另一套作为「补充说明」。

**建议方案**: 以 shared.md 步骤2 为准（因为 AI 实际执行匹配时读取它），在 prompts 文件的 keywords 区加注来源：

```yaml
# prompts/analysis.md
triggers:
  keywords:
    # shared.md 步骤2「参考分析」: 对比/差异/参考/借鉴/调研/看看别人的/分析xxx的
    # shared.md 步骤2「流程分析」: 分析流程/业务流程/制定流程/设计流程...
    分析/调研/对比/参考/借鉴/看看别人的/分析xxx的/制定流程/业务流程
```

所有 prompts 文件的 keywords 区统一标注来源，减少维护负担。

### 🔴 K2: 修复「差异」歧义

从 shared.md bugfix 行1 删除「差异」：

```
# 修复前
不一致/差异/偏差/对齐/样式不匹配/高度/宽度/不一致

# 修复后
不一致/偏差/对齐/样式不匹配/高度不对/宽度不对
```

### 🟡 K4: 统一 team-input-matcher.md 与 shared.md

方案 A（推荐）：在 team-input-matcher.md 开头加警告：

```markdown
> ⚠️ **注意**: 本文件是简化版匹配器，仅覆盖 5 个高频场景。
> 完整匹配规则（包括流程分析/阶段推进/需求分析/部署运维/代码审查）见 shared.md 步骤2。
```

方案 B：补充缺失的 5 个场景到 team-input-matcher.md。

### 🟡 K6: 在 shared.md 测试验证意图补充「Gherkin」

```markdown
测试验证 | 测试/验证/QA/回归/Gherkin
```

### 🟠 K8: 分离 tech-lead 与 framework-architect 关键词

在 prompts/tech-lead.md 和 prompts/framework-architect.md 的 keywords 区加互斥标注：

```yaml
# prompts/tech-lead.md — 仅用于代码审查
# prompts/framework-architect.md — 仅用于架构分析
```

shared.md 的代码审查意图映射到 tech-lead.md，架构分析意图映射到 framework-architect.md，两者职责边界已清晰。只需在两个 prompt 文件的 keywords 区各加一行说明即可。

---

## 五、修复清单

| 优先级 | # | 修复内容 | 涉及文件 |
|--------|---|---------|---------|
| 🔴 K0 | 1 | prompts keywords 区统一标注来源 shared.md 步骤2 | 全部 prompts/*.md |
| 🔴 K2 | 2 | bugfix 行1 删除「差异」 | shared.md |
| 🔴 K0 | 3 | team-input-matcher.md 加「非完整版」警告 | team-input-matcher.md |
| 🟡 K4 | 4 | 补充「需求分析/部署运维/代码审查」到 team-input-matcher.md | team-input-matcher.md |
| 🟡 K6 | 5 | shared.md 测试验证意图加「Gherkin」 | shared.md |
| 🟡 K8 | 6 | prompts/tech-lead.md 和 framework-architect.md keywords 加互斥说明 | tech-lead.md, framework-architect.md |
| 🟡 K9 | 7 | prompts/dev.md keywords 补充「开始实现/审核通过/按方案执行/开始BDD/开始TDD」 | dev.md |
| 🟠 K5 | 8 | prompts/analysis.md 加注「流程分析」含义区别 | analysis.md |

---

## 六、版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1 | 2026-04-20 | 初始审核：K1-K9 共9个问题，K0-K9共8个修复项 |
