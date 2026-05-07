# team-flow Framework v6.0 → Trae Agent Prompt 生成总结

> 生成日期: 2026-04-25 | 版本: v6.1 | 状态: ✅ 已完成

---

## 1. 项目背景

### 1.1 核心问题

在单 Agent 模式下，AI 反复绕过 `team-flow` 框架中已写入的规则，最典型的案例是 **Toolchain Gate**：

- `dev.md` 和 `development-standards.md` 明确要求"从 project.md 读取包管理器"，但 AI 仍在 bun 项目中使用 `npx` 或错误的构建命令
- 根因分析：单 Agent 模式下规则是"软约束"（soft constraints），AI 在压力下会跳过"先读 X 再做 Y"的指令，npx 属于肌肉记忆

### 1.2 解决方案

**Plan A（主方案）：Trae 多 Agent 路由**

将 `team-flow` 的每个角色转换为 Trae IDE 的独立 Agent，**所有规则内联到 Prompt 中**（Trae 无文件系统访问能力，无法在运行时读取外部文件）。

**Plan B（备选）：单 Agent 回退 + Git Hooks / CI**

如 Trae 多 Agent 不可用，回退到单 Agent 模式，通过 git hooks 和 CI 强制执行规则。

### 1.3 核心设计原则

> **所有规则必须 INLINE（内联）** — 不允许"读取 X 文件后执行 Y"的运行时依赖

| 传统 team-flow Prompt | Trae Agent Prompt |
|---|---|
| 引用外部文件："读取 conventions/dev-common.md" | 全文内联：直接嵌入规则内容 |
| PRE-FLIGHT 作为入口检查步骤 | PRE-FLIGHT 内联到每个 Agent 的开头 |
| 依赖文件系统访问 | 零文件系统依赖 |

---

## 2. 生成结果

### 2.1 文件清单

目标目录：`{TEAM_PATH}/examples/Trae/`

| # | 文件名 | 角色 | 大小 | PRE-FLIGHT 内联内容 |
|---|--------|------|------|-------------------|
| — | CONFIG.md | 环境变量配置 | 414 B | — |
| 01 | triage.md | Triage 分诊员 | 4,486 B | §ROLES + §PATHS |
| 02 | dev.md | Dev 开发工程师 | 7,107 B | §TOOLCHAIN + §CONSTRAINTS + conventions |
| 03 | tech-lead.md | Tech Lead 技术负责人 | 7,160 B | §CONSTRAINTS + conventions |
| 04 | qa-engineer.md | QA 测试工程师 | 5,505 B | §TOOLCHAIN + §CONSTRAINTS + conventions |
| 05 | pm.md | PM 产品经理 | 4,229 B | §CONSTRAINTS only |
| 06 | devops.md | DevOps 运维工程师 | 2,828 B | §TOOLCHAIN(Backend) + §CONSTRAINTS |
| 07 | ui-designer.md | UI Designer 设计师 | 4,608 B | §CONSTRAINTS + conventions + design tokens |
| 08 | analysis.md | Analysis 分析师 | 2,772 B | §CONSTRAINTS only |
| 09 | framework-architect.md | Framework Architect 架构师 | 2,992 B | §CONSTRAINTS only |
| 10 | bugfix.md | Bugfix 修复工程师 | 6,265 B | §TOOLCHAIN + §CONSTRAINTS + conventions |

**总计：11 个文件，约 48.9 KB**

### 2.2 PRE-FLIGHT 角色差异矩阵

| 角色 | §TOOLCHAIN | §CONSTRAINTS | Conventions | Design Tokens | §ROLES | §PATHS |
|------|:----------:|:------------:|:-----------:|:-------------:|:------:|:------:|
| Triage | — | — | — | — | ✅ | ✅ |
| Dev | ✅ Full | ✅ | ✅ | — | — | — |
| Tech Lead | — | ✅ | ✅ | — | — | — |
| QA | ✅ Full | ✅ | ✅ | — | — | — |
| PM | — | ✅ | — | — | — | — |
| DevOps | ✅ Backend | ✅ | — | — | — | — |
| UI Designer | — | ✅ | ✅ | ✅ | — | — |
| Analysis | — | ✅ | — | — | — | — |
| Framework Architect | — | ✅ | — | — | — | — |
| Bugfix | ✅ Full | ✅ | ✅ | — | — | — |

**设计逻辑：**
- 需要写代码的角色（Dev/QA/Bugfix）→ 全量 Toolchain + 约束 + 规范
- 需要技术决策的角色（Tech Lead）→ 约束 + 规范（无需 Toolchain）
- 轻量角色（PM/Analysis/Framework Architect）→ 仅约束
- 特殊角色（UI Designer）→ 约束 + 规范 + 设计 Token
- 运维角色（DevOps）→ 后端 Toolchain + 约束
- 分诊角色（Triage）→ 仅角色定义 + 路径

---

## 3. 内联内容详情

### 3.1 §TOOLCHAIN（工具链门禁）

**Full（Dev/QA/Bugfix）：**

| 类别 | 工具 | 版本 |
|------|------|------|
| Frontend Runtime | bun | 禁止 npm/pnpm/yarn |
| Frontend Build | Rsbuild | — |
| Frontend Framework | React 18 + TypeScript | — |
| Backend Runtime | Go | 1.21+ |
| Backend Framework | Kratos + Gin | — |
| ORM | Ent | — |
| Database | PostgreSQL / MySQL | — |
| Cache | Redis | — |
| API Protocol | proto3 | — |

**Backend Only（DevOps）：** 仅 Go/buf/Ent 部分

**门禁规则：** 拦截 `npm`/`pnpm`/`yarn` → 强制使用 `bun`；拦截 `npx` → 使用 `bunx`

### 3.2 §CONSTRAINTS（硬约束）

- 项目标识：{project-name}
- API 前缀：`/api/v1`（前后端统一）
- 前端路由前缀：`/admin`
- 环境变量前缀：`ORIGCMS_`
- 日志标签：`origcms.{module}`
- 错误码范围：1000-1299 系统通用，2000-2499 按模块分配
- 文档格式：Markdown
- 代码注释：英文

### 3.3 Conventions（项目规范）

**common.md：** 项目标识、环境变量、日志、错误码

**dev-common.md：**
- Git 分支：`feat/`/`fix/`/`hotfix/`
- Commit 类型前缀
- 测试覆盖率：80%+
- 核心流程 E2E 测试
- 1-reviewer 合并规则

**dev-frontend.md：**
- `/admin` 路由前缀（强制）
- `RSBUILD_API_BASE_URL=/api/v1`
- 状态管理：Zustand
- 组件命名：PascalCase
- API 文件：按模块组织

**dev-backend.md：**
- `/api/v1` 前缀（强制）
- Kratos 分层：API → Service → Biz → Data
- 表命名：UUID + 时间戳
- 错误码：2000-2499 按模块
- 分页：Cursor-based

### 3.4 Design Tokens（设计 Token，仅 UI Designer）

| Token 类别 | 关键值 |
|-----------|--------|
| Brand Color | #7C3AED |
| Sidebar Width | 240px |
| Category Colors | 各内容类型专属颜色 |
| Status Colors | 成功/警告/错误/信息 |
| Player Tokens | 播放器相关 UI Token |

---

## 4. 格式规范

### 4.1 统一文件结构

```markdown
# {Role Name} — {中文名}

> OrigCMS 项目专用 | 角色入口：{职责简述}
> 版本: v6.1 | 自动生成于 2026-04-25

---

## ⛔ PRE-FLIGHT（执行任何操作前必须完成）
{内联的 PRE-FLIGHT 检查步骤}

## 🔒 入口门禁
{角色特定的准入条件}

## 📋 工作流
{角色主工作流，含决策树和表格}

## 📤 输出物
{角色交付物清单}

## ✅ 完成门禁
{完成前必须通过的检查}

## ❌ 禁止事项
{角色明确的红线}
```

### 4.2 格式约定

| 符号 | 含义 |
|------|------|
| ⛔ | 阻断/门禁（必须通过才能继续） |
| ✅ | 通过/确认 |
| ❌ | 禁止/不允许 |
| 📌 | 注意事项 |

---

## 5. team-flow Framework v6.0 四点计划进度

| # | 项目 | 状态 | 说明 |
|---|------|------|------|
| 1 | PRE-FLIGHT 模块 | ✅ 已完成 | `workflows/pre-flight.md` v6.1，定义了每个角色的 PRE-FLIGHT 差异 |
| 2 | shared.md Step 0 更新 | ⏳ 未开始 | 在 shared.md 中加入 PRE-FLIGHT 作为 Step 0 |
| 3 | 经验规则合并到 conventions | ⏳ 未开始 | 将 lessons/ 内容结构化合并到 conventions/ |
| 4 | 原始 Prompt 更新 PRE-FLIGHT 引用 | ⏳ 部分完成 | Trae 生成已完成内联，但 `{TEAM_PATH}/prompts/` 原始文件尚未更新 |

---

## 6. 待解决问题

| # | 问题 | 优先级 | 说明 |
|---|------|--------|------|
| 1 | AI 不遵守已写规则 | 🔴 高 | 根因是单 Agent 软约束，Trae 多 Agent 是解决方案 |
| 2 | 机制落地不完整 | 🟡 中 | 错误报告目录/模板未创建、Lesson Analyst 规范未落地、INDEX.md 未创建 |
| 3 | shared.md 膨胀 | 🟡 中 | 14.5KB，应只保留门禁+工作流+任务生命周期，其余归还到独立文件 |
| 4 | 角色维护职责未补全 | 🟡 中 | 每个 prompt 需加维护职责（谁负责更新哪些文件） |
| 5 | bugfix.md/triage.md 无 YAML frontmatter | 🟡 中 | AI 工具无法自动加载 standards 引用 |
| 6 | Git 工作流未定义 | 🟢 低 | 当前 AI 修改后人确认手动 commit，无具体流程 |
| 7 | INSTALL.md/REFACTOR_SUMMARY.md 过时 | 🟢 低 | 引用已删除文件，版本描述不符 |
| 8 | 变更范围外变更处理流程 | 🟢 低 | 小改动直接改 vs 大改动审批，流程未确认 |
| 9 | Bugfix 错误报告存放位置 | 🟢 低 | 是否也放 reports/errors/ 下未确认 |

---

## 7. 重大事件记录

### 7.1 文件丢失事件（4月24日 15:28）

- **事件**：workflows/roles/ 下 14 个文件从磁盘删除（git 索引 AD 状态）
- **恢复**：通过 `git checkout` 恢复
- **根因**：无法精确定位（LCM 压缩了 exec 细节），可能与 docs/migration/ 操作有关
- **教训**：trash > rm；先确认再删除；仓库零 commit，丢失风险极高

### 7.2 核心认知修正

> **{TEAM_PATH}/ ≠ 项目层级，{TEAM_PATH}/ = AI 执行协议/约束**

- {TEAM_PATH}/ 是"AI 怎么干活怎么不犯错"的规则
- _docs/{project}/ 是"项目的一切决策"
- 两者是不同领域，而非不同层级

---

## 8. 生成过程

| 阶段 | 时间 | 内容 |
|------|------|------|
| 框架整理 v6.0 | 4/24 11:16-15:28 | workflows/ 目录整理、prompts 增强、经验学习系统骨架、项目约定目录 |
| 文件丢失与恢复 | 4/24 15:28 | 14 个 roles 文件删除后恢复 |
| 问题诊断 | 4/24 20:20 | 诊断 AI 绕过规则根因，提出 PRE-FLIGHT + Trae 多 Agent 方案 |
| PRE-FLIGHT 模块 | 4/24 20:34 | 创建 `workflows/pre-flight.md` v6.1 |
| Trae Agent 设计 | 4/24 20:54 | 确认 Trae 格式（name + markdown + MCP tools），内联策略 |
| 01-triage / 02-dev 生成 | 4/24 21:28 | 前两个 Agent 文件生成 |
| 剩余 8 个 Agent 生成 | 4/25 06:02-06:11 | 03-10 全部生成并部署到最终目录 |
| 总结文档 | 4/25 06:31 | 本文档 |

---

*本文档由 QClaw 自动生成，基于 team-flow Framework v6.0 的 Trae Agent Prompt 生成过程记录。*
