# 已发现的不一致问题

> **版本**: v1.1 | **日期**: 2026-05-11 | **状态**: 待确认

## 🔴 严重问题

### I-01: bd CLI 引用残留

**位置**: 多个文件中仍有 `bd` CLI 引用，应统一为 `flow task` CLI。

| 文件 | 行号 | 内容 |
|------|------|------|
| triage.md | 31 | `**Core tool**: \`bd\` CLI (beads)` |
| triage.md | 964 | `Use \`bd\` CLI for all task operations` |
| BOUNDARY.md | 174 | `AI interacts via \`bd\` CLI only` |
| BOUNDARY.md | 220 | `Use \`bd\` CLI for all task operations` |
| shared.md | 949 | `Edit task-pool-export.md to change task state (use \`bd\` commands)` |

**影响**: AI 可能使用 `bd` 命令而非 `flow task` 命令，导致操作失败。

### I-02: triage.md 第 56 行标题与内容不一致

**位置**: triage.md 第 56 行

**当前**: `2. **Detect directory type** (see Project Context Detection below)`
**应为**: `2. **Resolve project context** (see Project Context Resolution below)`

**影响**: AI 可能按旧的"检测"逻辑而非"配置驱动"逻辑执行。

### I-03: triage.md 第 76 行仍写"⛔ STOP"

**位置**: triage.md 第 76 行

**当前**: `|   +-- Mode not set → ⛔ STOP. Ask user to configure before proceeding.`
**应为**: `|   +-- Mode not set → Help user configure, then proceed.`

**影响**: AI 可能在 Mode 未配置时停止操作，而不是帮助用户配置。

### I-04: BOUNDARY.md 缺少 Management Mode 概念

**位置**: BOUNDARY.md 全文

**问题**: BOUNDARY.md 仍然使用旧的 Layer 定义，没有 Management Mode（workspace/standalone）概念。与 SKILL.md 和 triage.md 的项目上下文规则冲突。

**影响**: AI 读取 BOUNDARY.md 时可能按旧逻辑操作，与其他文件的新逻辑矛盾。

### I-05: SKILL.md 版本号不一致

**位置**: SKILL.md frontmatter vs body

- frontmatter: `version: 2.3`
- body: `**Version**: v2.4`

**影响**: 版本号混乱，难以追踪变更。

### I-06: cr-index 是 AI 自编的，beads 数据库无此字段

**位置**: Status Line 中的 `#{cr-index}` 部分

**问题**: 
- beads 数据库（`.beads/issues.jsonl`）中没有 cr-index 字段
- `#1`、`#2` 等序号是 AI 自己在对话中累计的，不可靠
- 跨会话时 cr-index 会丢失或重置

**需求文档方案**: 05-status-line.md 已定义方案 B — 文档内部 INDEX.md 表格管理 cr-index

**影响**: Status Line 中的 cr-index 不可靠，跨会话无法追踪。

### I-07: 技能系统缺失

**位置**: 整个 team-flow 实现

**问题**:
- 没有 Skill Profile（技能档案）机制
- 没有 Skill Files（技能文件）机制
- 没有 Skill Learning（技能学习）机制
- AI 不知道当前项目用什么技术栈，经常用错工具（如 npm 替代 bun）

**需求文档方案**: 03-role-system.md 已定义完整技能系统

**影响**: AI 频繁使用错误工具链，破坏项目一致性。

### I-08: Triage 中间人原则未在实现文件中体现

**位置**: SKILL.md、triage.md、BOUNDARY.md

**问题**:
- 当前流程允许用户直接和子 Agent 对话
- Triage 创建任务后"不管了"
- 没有反馈路由机制

**需求文档方案**: 04-task-lifecycle.md 已定义 Triage 中间人原则 + 反馈路由决策树

**影响**: 用户反馈无法正确路由，子 Agent 可能收到不完整的需求。

### I-09: project.md 仍为信息性文档，未升级为强制性规则

**位置**: `.team/project.md` 及所有引用它的文件

**问题**:
- project.md 当前是"信息性"的（告诉 AI 项目是什么）
- 应该是"强制性"的（Tech Stack MANDATORY + Toolchain MANDATORY）
- AI 读取后不一定会遵守

**需求文档方案**: 01-core-requirements.md 已定义 project.md 强制规则 + PRE-FLIGHT 检查

**影响**: AI 可能忽略 project.md 中的配置，使用错误的技术栈和工具链。

### I-10: 门禁式流程控制在实现文件中未体现

**位置**: SKILL.md、BOUNDARY.md

**问题**:
- 当前使用"不应该做"列表（负面约束），AI 经常无视
- 应该使用门禁式流程控制（正面约束：操作有前置条件）

**需求文档方案**: 01-core-requirements.md 已定义门禁式流程控制表

**影响**: AI 可能跳过必要的前置检查，直接执行操作。

## 🟡 中等问题

### I-11: shared.md 版本号与 SKILL.md 不一致

- shared.md: `TEAM_VERSION=8.1`
- SKILL.md: `v2.4`

两套版本号体系，容易混淆。

### I-12: BOUNDARY.md Bug 目录结构与 shared.md 不一致

- BOUNDARY.md: `B001/R1/`（R 是子目录）
- shared.md: `B001-R1/`（R 是目录名后缀）

两种目录结构冲突。

### I-13: 多处文件重复定义相同规则

项目管理模式在 SKILL.md、triage.md、shared.md、path-resolution.md 四个文件中都有定义。如果修改一处忘记同步其他，就会产生不一致。

**建议**: 规则只在一个文件中定义，其他文件引用。

### I-14: Triage 需求补全 + 反馈路由决策树未实现

**位置**: triage.md

**问题**:
- Triage 收到需求后没有补全步骤（目标项目、技术栈、影响范围、验收标准、边界条件）
- 用户反馈没有路由决策（A:代码Bug / B:需求理解偏差 / C:需求变更 / D:新需求 / E:验收通过）

**需求文档方案**: 04-task-lifecycle.md 已定义需求补全表格 + 5 类型反馈路由决策树

**影响**: 用户反馈可能被错误处理，需求不完整时直接分发给子 Agent。

## 🟢 轻微问题

### I-15: triage.md 中 `flow task stats --json` 命令可能不存在

**位置**: triage.md 第 57 行

`flow task stats --json` 在 commands.md 中没有定义。

### I-16: roles/ 目录下文件只有 1 行

**位置**: `.trae/skills/team-flow/workflows/roles/` 下 15 个文件

这些文件都只有 1 行，可能是空文件或占位文件。