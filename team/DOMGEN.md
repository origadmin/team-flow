# DOMGEN - _team 维护入口

📌 DOMGEN 是 _team 的自维护系统:管理文档生成和 _team 规则维护。用户说"加载 DOMGEN 规则"即可进入维护模式

> **版本**: 1.2 | **更新日期**: 2026-04-24

---

## 定位

DOMGEN 与 SKILL.md 职责完全分离:

| | SKILL.md | DOMGEN.md |
|---|----------|------------|
| **定位** | 项目开发流程入口 | _team 维护入口 |
| **管什么** | 门禁 + 工作流 + 角色职责 | 文档生成 + 规则维护 |
| **谁读** | AI 执行项目任务时 | AI 维护 _team 时 |
| **互相引用** | 不引用 DOMGEN | 可以读 SKILL 作为源文件 |

📌 SKILL 不需要知道 DOMGEN 的存在。DOMGEN 可以受 SKILL 影响(SKILL 是文档生成的源文件之一)

---

## 触发条件

| 触发方式 | 说明 |
|---------|------|
| 手动 | 用户说"加载 DOMGEN 规则" / "更新团队文档" / "regenerate docs" / "重新生成文档" |
| 自动 | _team/ 下任何 AI 规则文件被修改时(AI 主动检测) |
| 周期 | Heartbeat 检查时,对比 hash 发现变更 |

---

## 源文件 → 产出映射

| 源文件 | 生成文件 | 生成规则 |
|--------|---------|---------|
| `SKILL.md` | `docs/README.md` | 提取入口门禁 + 角色表 + 两层工作流概述 + 版本历史 |
| `workflows/shared.md` | `docs/WORKFLOW.md` | 提取两层工作流详细描述 + 门禁逻辑 + 资产包规格 + 发布流程 |
| `prompts/dev.md` | `docs/DEV-GUIDE.md` | 提取 TDD + 交付管线 + Commit/分支规范 + 异常处理 + Mock + 目录结构 |
| `prompts/tech-lead.md` | `docs/DESIGN-GUIDE.md` | 提取设计流程 + ADR 模板 + 后端/前端架构规范 |
| `prompts/triage.md` | `docs/TRIAGE-GUIDE.md` | 提取分类流程 + MILESTONES 同步规则 |
| `prompts/bugfix.md` | `docs/BUGFIX-GUIDE.md` | 提取 Bug 修复流程 + RCA 模板 |

---

## 生成格式

### 📌 背景展开

AI 规则文件中 `📌` 标记的行为(≤50字),展开为 1-2 段说明(100-200字):

```
规则文件:
📌 TDD 确保测试先行,避免"先写代码后补测试"的自欺行为

生成文件:
**为什么强制 TDD?** 传统的"先写代码后补测试"往往导致测试只覆盖"快乐路径",
忽略边界和异常。TDD 强制先定义期望行为再实现,确保测试真正驱动设计,
而非事后验证。这在多人协作中尤为关键--测试即是功能规格的活文档。
```

### 判断树 → 流程描述

```
规则文件:
```
用户输入
    │
    ├── "T:" 前缀?→ Triage
    └── 都不满足?→ ⛔ 拒绝
```

生成文件:
当用户提交输入时,系统首先检查是否以 "T:" 前缀开头。如果是,
进入 Triage 分类流程;如果输入的是已有的任务 ID,则加载对应角色执行任务;
如果都不匹配,系统会拒绝处理并提示正确用法。
```

### 模板 → 填充示例

```
规则文件:
task_id: "{TASK_ID}"

生成文件:
task_id: "F001"  ← 实际任务 ID,如 Feature 001
```

### 禁止清单 → "为什么禁止"说明

```
规则文件:
- ❌ 不写测试就写实现

生成文件:
- ❌ **不写测试就写实现** - 跳过测试直接实现会导致代码缺乏回归保护,
  后续修改容易引入隐蔽 bug,且无法验证功能是否符合设计规格。
```

---

## 产出位置

```
_team/                           ← AI 规则文件(AI 读写)
_team/docs/                      ← 人读文档入口(AI 生成 / AI 维护)
  ├── README.md                  ← [派生] 从 SKILL.md 生成
  ├── WORKFLOW.md                ← [派生] 从 shared.md 生成
  ├── DEV-GUIDE.md               ← [派生] 从 dev.md 生成
  ├── DESIGN-GUIDE.md            ← [派生] 从 tech-lead.md 生成
  ├── TRIAGE-GUIDE.md            ← [派生] 从 triage.md 生成
  ├── BUGFIX-GUIDE.md            ← [派生] 从 bugfix.md 生成
  ├── INSTALL.md                 ← [派生] 从 examples/ + config/ 生成
  │
  ├── design/                    ← 架构设计决策
  │   ├── ARCHITECTURE.md        ← [派生] 从 tech-lead.md ADR 规则 + framework-architect.md 汇总
  │   └── DESIGN-v6.md           ← [快照] v6 设计方案,版本升级后冻结
  │
  └── migration/                 ← 版本迁移记录
      ├── MIGRATION_v4.md        ← [快照] v3.7→v4.0 升级记录,已冻结
      ├── MIGRATION_v5.md        ← [快照] v5.0 升级记录,已冻结
      └── REFACTOR_SUMMARY.md    ← [快照] v5.0 重构总结,已冻结
```

📌 `docs/` 全部由 AI 维护。人只读,不写。AI 执行任务时不主动读取 docs/

### docs/ 文档分类

📌 每份文档都必须有明确的维护者,禁止出现无人负责的"手动维护"

| 类型 | 标记 | 维护者 | 生命周期 | 同步方式 |
|------|------|--------|---------|----------|
| **派生文档** | `[派生]` | DOMGEN | 规则变 → 文档变 | hash 校验 + 自动重生成 |
| **快照文档** | `[快照]` | AI(版本升级时) | 写入后冻结,下版本产生新文件 | 不跟踪 hash,永不覆盖 |

### 派生文档(DOMGEN 管理)

源规则变了 → 自动重新生成 → 覆盖旧版本。

| 文档 | 源 | 生成规则 |
|------|-----|---------|
| `docs/README.md` | SKILL.md | 提取入口门禁 + 角色表 + 两层工作流概述 + 版本历史 |
| `docs/WORKFLOW.md` | workflows/shared.md | 提取两层工作流详细描述 + 门禁逻辑 + 资产包规格 + 发布流程 |
| `docs/DEV-GUIDE.md` | prompts/dev.md | 提取 TDD + 交付管线 + Commit/分支规范 + 异常处理 + Mock + 目录结构 |
| `docs/DESIGN-GUIDE.md` | prompts/tech-lead.md | 提取设计流程 + ADR 模板 + 后端/前端架构规范 |
| `docs/TRIAGE-GUIDE.md` | prompts/triage.md | 提取分类流程 + MILESTONES 同步规则 |
| `docs/BUGFIX-GUIDE.md` | prompts/bugfix.md | 提取 Bug 修复流程 + RCA 模板 |
| `docs/INSTALL.md` | examples/ + config/ | 提取安装步骤 + 配置示例 + 快速开始 |
| `docs/design/ARCHITECTURE.md` | prompts/tech-lead.md + prompts/framework-architect.md | 汇总架构决策(ADR 格式),规则变则重生成 |

### 快照文档(版本里程碑)

📌 快照文档是版本升级时的决策记录,写入后冻结,不因后续规则变更而修改

#### 快照文档子分类

| 子类 | 目录 | 内容 | 产生时机 |
|------|------|------|----------|
| **设计快照** | `docs/design/` | 版本设计方案(决策记录) | 新版本设计完成后 |
| **迁移快照** | `docs/migration/` | 版本迁移操作记录 | 版本升级完成后 |

#### 设计快照

| 文档 | 产生时机 | 冻结时间 |
|------|---------|----------|
| `docs/design/DESIGN-v6.md` | v6 设计阶段 | v6 规则全部落地后冻结 |

📌 设计快照只记录"为什么这样设计"的决策,不记录执行指令。执行完的合并清单和实施步骤应删除

#### 迁移快照

| 文档 | 产生时机 | 冻结时间 |
|------|---------|----------|
| `docs/migration/MIGRATION_v4.md` | v3.7→v4.0 迁移完成 | 迁移完成即刻冻结 |
| `docs/migration/MIGRATION_v5.md` | v5.0 迁移完成 | 迁移完成即刻冻结 |
| `docs/migration/REFACTOR_SUMMARY.md` | v5.0 重构完成 | 重构完成即刻冻结 |

📌 迁移快照的维护规则:版本升级时由 AI 一次性生成,记录操作步骤和结果,写入后冻结

- 产生方式:用户手动触发 AI 生成("生成 v7 迁移文档")
- 内容:升级操作步骤 + 受影响文件 + 验证结果
- 生命周期:写入后冻结,新版本产生新文件,旧文件永不修改
- DOMGEN 不触碰迁移快照(不生成、不覆盖、不校验 hash)

**快照文档规则**:
- ❌ 禁止 DOMGEN 覆盖快照文档
- ❌ 禁止因规则变更修改快照文档内容
- ✅ 新版本升级时,AI 生成新的 MIGRATION/DESIGN 文档
- ✅ 快照文档只有头部标注(无 hash),标明冻结日期

### 快照文档头部标注

```markdown
<!-- SNAPSHOT - Frozen as of {YYYY-MM-DD} - DO NOT REGENERATE -->
<!-- This document records decisions made during {version} upgrade -->
<!-- For current architecture, see docs/design/ARCHITECTURE.md -->
```

---

## 同步保障

📌 自动标注 + hash 校验,确保人读文档与规则文件不漂移

### 文件头部标注

```markdown
<!-- AUTO-GENERATED from _team/{source} - DO NOT EDIT MANUALLY -->
<!-- To update: modify _team/{source}, then run DOMGEN regeneration -->
```

### 文件尾部标注

```markdown
<!-- Last generated: {YYYY-MM-DD HH:mm} | Source: _team/{source} | Hash: {first-8-chars-of-sha256} -->
```

### Hash 计算

```
源文件内容 → SHA-256 → 取前 8 位
```

- 生成时记录 hash
- Heartbeat 检查时重新计算 hash
- hash 变化 → 触发重新生成

---

## 📌 背景标记规范

📌 背景标记是 AI 规则文件中的极简注释,让人和 AI 都理解规则背后的原因

### 格式

```markdown
## {规则标题}
📌 {一句话背景说明(≤50字)}

{规则内容...}
```

### 放置位置

- 紧跟在章节标题下方
- 只在"为什么需要这条规则"不显然时添加
- 一条规则最多一个 📌

### 当前标记清单

| 文件 | 📌 标记 | 展开主题 |
|------|---------|---------|
| shared.md | 三层门禁确保流程不可绕过 | 门禁机制的设计意图 |
| shared.md | 发布层由 Milestone 触发,非单个任务触发 | 为什么发布是独立层 |
| shared.md | PM 是唯一放行人 | 为什么 Tech Lead/QA 无权放行 |
| shared.md | 闭环验证确保实现与设计一致 | 闭环验证解决的问题 |
| shared.md | API 问题易引发前后端推诿 | API 问题矩阵的意图 |
| shared.md | 质量门槛是发布层 R-Phase 1 的量化检查标准 | 量化标准的必要性 |
| shared.md | ID 命名统一 | ID 格式统一的原因 |
| shared.md | R 后缀加在资产目录上 | R 后缀规则的缘由 |
| shared.md | 目录必须带 R 后缀 | 防止资产目录混淆 |
| shared.md | AI 执行锚点让 AI 只读必要章节 | 节省 token 的设计 |
| shared.md | 文档同步绑在阶段门禁上 | 文档不漂移的保障 |
| shared.md | MILESTONES 是甲方需求清单 | 所有权分离 |
| shared.md | AI 更新时必须遵循格式规则 | 格式一致性保障 |
| dev.md | Dev 负责编码实现与缺陷修复 | Dev 角色定位 |
| dev.md | 硬编码包管理器是常见错误源 | 工具链门禁的必要性 |
| dev.md | TDD 确保测试先行 | TDD 强制的原因 |
| dev.md | 交付管线从 project.md 动态读取 | 避免硬编码命令 |
| dev.md | 统一 commit 格式 | CHANGELOG 自动生成 |
| dev.md | 分支命名统一 | 分支管理可追溯 |
| dev.md | 异常处理不统一导致安全泄露 | 异常规范必要性 |
| dev.md | 前端应先基于 R3 用 Mock 独立开发 | 前后端解耦 |
| dev.md | Code Review 是质量把关关键 | 审查的价值 |
| dev.md | 目录结构统一 | 代码组织一致性 |
| tech-lead.md | Tech Lead 负责技术设计和架构决策 | TL 角色定位 |
| tech-lead.md | 收到 Feature 必须先创建资产包 | 杜绝空手开工 |
| tech-lead.md | Change 采用保守策略 | 变更影响评估原则 |
| tech-lead.md | ADR 记录重大技术决策 | 防止决策失忆 |

---

## 首次生成步骤

1. 确认 `docs/` 目录存在(已在 _team/ 内)
2. 按源文件→产出映射,逐个生成
3. 每个文件加上头部 + 尾部标注
4. 记录 hash

---

## 禁止

- ❌ 手动编辑 `docs/` 下有 `[派生]` 标注的文件(会被下次生成覆盖)
- ❌ DOMGEN 覆盖有 `[快照]` 标注的文件(快照冻结不可变)
- ❌ 人直接编辑 `docs/` 下的任何文件(全部由 AI 维护)
- ❌ AI 执行任务时主动读取 `docs/` 目录
- ❌ 在 `docs/` 文件中嵌入 AI 执行规则
- ❌ 新增 `docs/` 文件而不定义其类型(派生/快照)和维护者
- ❌ 生成时遗漏 📌 背景展开
- ❌ 生成后不记录 hash

---

## 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| **v1.2** | **2026-04-24** | **SKILL/DOMGEN 职责分离：SKILL 不引用 DOMGEN，DOMGEN 定位为 _team 维护入口。Migration 文件维护规则明确。DESIGN-v6 快照清理执行指令** |
| v1.1 | 2026-04-24 | 修复：取消"手动维护"分类，docs/ 全部由 AI 维护。引入派生/快照二分类 |
| v1.0 | 2026-04-24 | 初始版本 |
