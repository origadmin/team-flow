# 经验学习系统示例：从错误到规则

> 演示如何将"AI 反复犯的错误"提取为"不可违反的公式"

---

## 场景：AI 反复用错包管理器

### 错误现象

**用户指令**: "安装依赖"

**AI 错误做法** (反复发生):
```bash
npm install xxx      # ❌ 错误！项目用 bun，不是 npm
# 或
pnpm add xxx         # ❌ 错误！项目用 bun，不是 pnpm
```

**正确做法**:
```bash
bun add xxx          # ✅ 正确！从 project.md 读取 Toolchain
```

### 错误根因分析

| 层级 | 问题 |
|------|------|
| 直接原因 | AI 使用默认/熟悉的命令，未检查项目配置 |
| 系统原因 | 没有强制机制阻止 AI 使用错误命令 |
| 知识原因 | AI "知道"要用 project.md，但执行时"忘记" |

---

## 提取公式

### Step 1: 记录 Error Report

```markdown
# Error Report: E00001-F001-001.md

## Basic Info
- TaskID: F001
- TaskType: Feature
- Phase: Phase 2
- Role: Dev-Backend
- Date: 2026-04-25

## Error Details
- ErrorType: build_failure
- Symptom: 使用 npm install 而不是 bun add，导致 lock 文件冲突
- RootCause: AI 未读取 project.md Toolchain 配置，直接使用默认 npm
- Resolution: 删除错误 lock 文件，重新用 bun add 安装
- TimeCost: 15min

## Impact Scope
- ChangedModule: package.json
- AffectedModule: bun.lockb
- CrossModule: No
- CrossRole: No
```

### Step 2: PM/Tech Lead 结构化审阅

```markdown
---
review_id: R-2026-04-25-001
milestone: M001
date: 2026-04-25
reviewer: Tech Lead
summary: "3次包管理器错误，必须形成强制规则"
---

## Error Filter Decision

| ErrorID | Decision | Reason |
|---------|----------|--------|
| E00001 | CODIFY | 重复发生，token 浪费严重 |
| E00005 | CODIFY | 同样问题，不同 AI |
| E00012 | CODIFY | 同样问题，前端项目 |

## Codify Details

### E00001 → Toolchain Gate Missing

- **Roles**: dev-backend, dev-frontend
- **PrimaryCategory**: code
- **SecondaryCategory**: (none)
- **RuleDirection**: "执行任何命令前必须加载 project.md Toolchain，使用其中定义的 package_manager，禁止直接使用 npm/pnpm/yarn"
- **SourceReport**: E00001-F001-001.md
```

### Step 3: Lesson Analyst 生成 AI Rule

**写入**: `{DOCS_INTERNAL}/lessons/dev-common.md`

```markdown
### Rule 1: Toolchain Gate — 包管理器强制检查

**Problem**: AI 直接使用 npm/pnpm/yarn，未读取 project.md Toolchain 配置，导致 lock 文件冲突和依赖不一致。

**Rule**:
1. 执行任何包管理命令前，必须先读取 `.team/project.md` 的 `Toolchain` 部分
2. 使用 `Toolchain.package_manager` 定义的命令（bun/npm/pnpm/yarn）
3. **禁止**使用与配置不符的包管理器
4. 如果 project.md 不存在 → ⛔ 拒绝执行任务

**Correct**:
- 先读取 project.md → 提取 Toolchain.package_manager → 使用该命令
- bun 项目: `bun add`, `bun install`, `bun run`
- npm 项目: `npm install`, `npm run`

**Wrong Example**:
```bash
# ❌ 错误：未读取配置直接使用 npm
npm install lodash

# ❌ 错误：使用 pnpm 但项目配置是 bun
pnpm add lodash
```

**Correct Example**:
```yaml
# project.md
Toolchain:
  package_manager: bun
  install: "bun add"
  install_dev: "bun add -d"
  run: "bun run"
```

```bash
# ✅ 正确：先读取配置，再执行命令
# AI 内部逻辑：
# 1. Read project.md → Toolchain.package_manager = "bun"
# 2. Use "bun add" instead of "npm install"
bun add lodash
```

**Applies When**: 任何需要执行包管理命令的场景（install/add/run/build）

**Source**: [E00001](../reports/errors/E00001-F001-001.md)
```

---

## Step 4: 在 Prompt 中预加载规则

在 `prompts/dev.md` 中添加：

```yaml
standards:
  - {TEAM_PATH}/workflows/shared.md
  - {DOCS_INTERNAL}/lessons/dev-common.md    # ← 新增：加载经验规则
```

并在入口门禁中强调：

```markdown
## 入口门禁

```
Dev 被触发
    │
    ├── Step 0: 加载 lessons/dev-common.md (如果存在)
    │   └── 包含 Toolchain Gate 规则
    │
    ├── Step 1: 加载 .team/project.md Toolchain 配置
    │   └── project.md 不存在？→ ⛔ 拒绝，提示创建
    │
    ...
```
```

---

## 效果验证

| 指标 | 之前 | 之后 |
|------|------|------|
| 包管理器错误次数 | 3次/周 | 0次 |
| 修复时间 | 15min/次 | 0 |
| Token 浪费 | 高（重试） | 低（预读规则） |

---

## 通用提取公式

任何"AI 反复犯的错误"都可以按这个流程提取为规则：

```
错误发生
    ↓
记录 Error Report（轻量，每次）
    ↓
PM/Tech Lead 审阅（里程碑时）
    ↓
决定是否 CODIFY（值得形成规则？）
    ↓
Lesson Analyst 翻译为 AI Rule
    ↓
写入 lessons/{role}-{type}.md
    ↓
在 prompts/{role}.md 中预加载
    ↓
AI 下次执行前自动读取规则
    ↓
错误不再发生
```

---

## 关键原则

1. **预读 > 重试**: 出错后再读规则是浪费，执行前预读更高效
2. **强制 > 建议**: "应该"没用，必须变成门禁检查
3. **具体 > 抽象**: 给出 exact command，不要模糊描述
4. **分层 > 集中**: 按角色+类型拆分，避免文件膨胀
