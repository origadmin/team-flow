# team-flow — AI 多角色协作框架

> OrigAdmin 多项目管理平台 — AI 执行协议
>
> **TEAM_VERSION=6.0** | 2026-04-27

---

## ⛔ FRAMEWORK LAYER BOUNDARY

**`{TEAM_PATH}/` is a MULTI-PROJECT framework. It is NOT a project directory.**

| Rule | Description |
|------|-------------|
| ❌ **NO WRITE** | AI must NEVER create or modify files inside `{TEAM_PATH}/` during project execution |
| ❌ **NO task-pool** | `task-pool.md` belongs at `{PROJECT_PATH}/.team/task-pool.md`, NOT `{TEAM_PATH}/task-pool.md` |
| ❌ **NO operational files** | backlog.md, project.md, issues.md all belong at `{PROJECT_PATH}/.team/` |
| ✅ **READ ONLY** | AI only reads rules/prompts/workflows from `{TEAM_PATH}/` |

**Before writing any file, verify the path does NOT start with `{TEAM_PATH}/`.**
See `BOUNDARY.md` for complete layer architecture.

---

---

## 这是什么？

`team-flow` 是一套让 AI 工具（Trae、Gemini CLI、OpenClaw 等）按照统一规范工作的配置框架。

**解决的问题**：
- AI 每次对话从零开始，不知道项目规范
- 多个 AI 工具各自为政，输出格式不一致
- 没有任务追踪，不知道做了什么、做到哪里

**核心设计**：
- **两层架构**：框架层（通用规则）+ 项目层（项目信息）
- **角色规范**：每个角色有 YAML prompt + standards 文件，按需加载
- **任务池**：任务分配 + 状态追踪
- **文档占位符**：`{docs_internal}` 从项目配置读取实际路径

---

## 快速开始

### 1. 创建项目 SKILL.md

在你的项目根目录创建 `SKILL.md`：

```markdown
# 项目 Skill

## 初始化

加载 {TEAM_PATH}/SKILL.md，按其中步骤执行。

## 项目信息

- **项目名**: {project-name}
- **技术栈**: Go + React
- **工作目录**: .
```

### 2. 启动 AI

对 AI 说：
```
加载 SKILL.md，初始化项目，检查任务池
```

AI 会自动：
1. 加载框架 SKILL.md
2. 创建 `.team/` 目录（含 project.md, task-pool.md）
3. 进入工作状态

---

## 目录结构

```
{TEAM_PATH}/                    ← 框架层（通用规则）
├── SKILL.md                    ← 入口：启动流程、角色路由
├── INSTALL.md                  ← 安装指南
├── README.md                   ← 本文件（人看的）
├── REFACTOR_SUMMARY.md         ← 重构记录
│
├── workflows/                  ← 工作流程（AI 执行）
│   ├── shared.md               ← 核心流程（所有角色必读，v6.0）
│   ├── framework-workflow.md   ← 框架开发流程（v1.0）
│   ├── meta/
│   │   └── TEAM_ROLES.md       ← 团队角色定义
│   └── roles/                  ← 角色分片规范（14个文件）
│       ├── analysis-standards.md
│       ├── architecture-standards.md
│       ├── bugfix-standards.md
│       ├── checklist.md
│       ├── development-standards.md
│       ├── devops-standards.md
│       ├── devtestops.md
│       ├── requirements-standards.md
│       ├── review-standards.md
│       ├── specialized-tests.md
│       ├── test-levels.md
│       ├── test-standards.md
│       ├── triage-standards.md
│       └── ui-standards.md
│
├── prompts/                    ← 角色提示词（YAML frontmatter）
│   ├── analysis.md             ← 分析师
│   ├── bugfix.md               ← Bug 修复
│   ├── dev.md                  ← 开发工程师（Backend/Frontend/Mobile）
│   ├── devops.md               ← 运维工程师
│   ├── framework-architect.md  ← 框架架构师
│   ├── pm.md                   ← 产品经理
│   ├── qa-engineer.md          ← 测试工程师
│   ├── team-input-matcher.md   ← 输入路由
│   ├── tech-lead.md            ← 技术总监
│   ├── triage.md               ← 分发工具
│   └── ui-designer.md          ← UI 设计师
│
├── templates/                  ← 文档模板（AI 复制用）
│   ├── project.md              ← 项目配置模板
│   ├── task-pool.md            ← 任务池模板
│   ├── gherkin-feature-template.md
│   ├── ui-design-template.md
│   └── ...
│
└── config/
    └── ...

你的项目/                       ← 项目层（项目信息）
├── SKILL.md                    ← 你创建的（指向框架）
├── .team/                      ← AI 自动创建
│   ├── version                 ← TEAM_VERSION
│   ├── project.md              ← 项目配置（技术栈、路径）
│   └── task-pool.md            ← 任务池
│
└── _docs/ 或 docs/             ← 项目文档（按 project.md 配置）
```

---

## 角色速查

| 角色 | Prompt | 加载 Standards |
|------|--------|---------------|
| 技术总监 | tech-lead.md | shared + analysis + architecture + review |
| 开发工程师 | dev.md | shared + development + bugfix + devtestops + test-levels + specialized-tests |
| 测试工程师 | qa-engineer.md | shared + test + devtestops + test-levels + specialized-tests + checklist |
| 产品经理 | pm.md | shared + requirements |
| 运维工程师 | devops.md | shared + devops |
| 框架架构师 | framework-architect.md | shared + development + framework-workflow |
| 分析师 | analysis.md | shared + analysis |
| Bug 修复 | bugfix.md | shared + bugfix |
| UI 设计师 | ui-designer.md | shared + ui |
| 分发工具 | triage.md | shared + triage |

---

## 核心规范

### TDD 开发流程（强制）

```
红 → 绿 → 重构
先写测试，再写实现，保持测试通过时重构
```

### 质量门禁

> ⚠️ 必须先执行命令检测（见 SKILL.md 步骤 5），以下为示例命令。

| 平台 | 必须通过 |
|------|---------|
| Go | `go fmt` + `go vet` + `golangci-lint` + `go test -cover -race` + `go build` |
| 前端 | 根据检测结果动态替换：bun/pnpm/npm → lint + 类型检查 + test + build |

### 文档路径占位符

所有 workflow 使用 `{docs_internal}` 表示内部文档路径：

```
{docs_internal}/requirements/{feature}/PRD.md
{docs_internal}/design/{feature}/ARCHITECTURE.md
{docs_internal}/test/{feature}/REPORT.md
{docs_internal}/delivery/{feature}/versions/v{version}/
```

AI 执行时从 `.team/project.md` 读取实际值替换。

---

## 任务池使用

```markdown
# .team/task-pool.md

## 待认领
| ID | 任务 | 类型 | 角色 | 优先级 |
|----|------|------|------|--------|
| TASK-001 | 用户注册 API | implement | dev | P1 |

## 进行中
| ID | 任务 | 负责人 | 进度 | 阻塞 |
|----|------|--------|------|------|

## 已完成
| ID | 任务 | 负责人 | 完成时间 | 产出物 |
|----|------|--------|----------|--------|
```

---

## 工作流程

```
PM 写需求 → Tech Lead 设计 → Dev 实现（TDD）→ QA 测试
    ↓              ↓               ↓               ↓
{docs_internal}/  {docs_internal}/  {docs_internal}/  {docs_internal}/
requirements/     design/          delivery/        test/
```

---

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v6.0 | 2026-04-25 | 切分废弃文件，修复引用链，新增 meta/TEAM_ROLES.md |
| v5.0 | 2026-04-24 | 三层门禁 + 任务生命周期 + 发布流程 |
| v3.0 | 2026-04-17 | 重构：删除 AGENTS.md，精简 SKILL.md，两层架构 |
| v2.0 | 2026-04-16 | 启动流程、角色路由 |
| v1.0 | 2026-04-15 | 初始版本 |

---

## 相关文档

- **安装指南**: `INSTALL.md`
- **重构记录**: `REFACTOR_SUMMARY.md`
- **框架入口**: `SKILL.md`
- **团队角色**: `workflows/meta/TEAM_ROLES.md`
