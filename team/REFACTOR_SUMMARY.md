# _team 重构总结文档

> 记录当前发现的问题、决策和待办事项
> 版本：v6.0 | 日期：2026-04-25

---

## 一、v6.0 架构（当前）

### 目录结构

```
_team/
├── SKILL.md                    ← 入口（驱动 TEAM，极简）
├── INSTALL.md                  ← 安装指南（人看的）
├── README.md                   ← 框架说明（人看的）
├── REFACTOR_SUMMARY.md         ← 本文件
│
├── workflows/                  ← 工作流程（AI 执行）
│   ├── shared.md               ← 核心流程（v6.0，三层门禁+任务生命周期）
│   ├── framework-workflow.md   ← 框架开发流程（v1.0）
│   ├── meta/
│   │   └── TEAM_ROLES.md       ← 团队角色定义（从 standards.md 拆分）
│   └── roles/                  ← 14 个角色规范文件
│       ├── analysis-standards.md      → analysis.md, tech-lead.md
│       ├── architecture-standards.md  → tech-lead.md
│       ├── bugfix-standards.md        → dev.md, bugfix.md
│       ├── checklist.md               → qa-engineer.md
│       ├── development-standards.md   → dev.md, framework-architect.md
│       ├── devops-standards.md        → devops.md
│       ├── devtestops.md              → dev.md, qa-engineer.md
│       ├── requirements-standards.md  → pm.md
│       ├── review-standards.md        → tech-lead.md
│       ├── specialized-tests.md       → qa-engineer.md
│       ├── test-levels.md             → dev.md, qa-engineer.md
│       ├── test-standards.md          → qa-engineer.md
│       ├── triage-standards.md        → triage.md
│       └── ui-standards.md            → ui-designer.md
│
├── prompts/                    ← 角色提示词（YAML frontmatter）
│   ├── analysis.md             ← 分析师
│   ├── bugfix.md               ← Bug 修复（非 YAML）
│   ├── dev.md                  ← 开发工程师
│   ├── devops.md               ← 运维工程师
│   ├── framework-architect.md  ← 框架架构师
│   ├── pm.md                   ← 产品经理
│   ├── qa-engineer.md          ← 测试工程师
│   ├── team-input-matcher.md   ← 输入路由
│   ├── tech-lead.md            ← 技术总监
│   ├── triage.md               ← 分发工具（非 YAML）
│   └── ui-designer.md          ← UI 设计师
│
├── templates/                  ← 文档模板（AI 复制用）
│   ├── project.md
│   ├── task-pool.md
│   ├── gherkin-feature-template.md
│   ├── ui-design-template.md
│   └── ...
│
└── config/
    └── ...
```

### 引用链（Prompt → Standards）

| Prompt | 加载的 Standards |
|--------|-----------------|
| analysis.md | shared + analysis-standards |
| bugfix.md | shared + bugfix-standards（非 YAML，body 引用） |
| dev.md | shared + development + bugfix + devtestops + test-levels + specialized-tests |
| devops.md | shared + devops-standards |
| framework-architect.md | shared + development-standards + framework-workflow |
| pm.md | shared + requirements-standards + gherkin-template |
| qa-engineer.md | shared + test + devtestops + test-levels + specialized-tests + checklist + gherkin-template |
| team-input-matcher.md | 无（路由角色） |
| tech-lead.md | shared + analysis + architecture + review-standards |
| triage.md | shared + triage-standards（非 YAML，body 引用） |
| ui-designer.md | shared + ui-standards + ui-design-template |

---

## 二、历史重构记录

### v6.0（2026-04-25）— 切分与引用修复

**删除文件**：
| 文件 | 原因 | 内容去向 |
|------|------|---------|
| output-standards.md | 纯索引，功能被 YAML standards 替代 | — |
| shared-implementation.md | 内容重复 | Zero Chinese Comments → dev.md |
| dev-workflow.md | 17KB 跨阶段流程 | Phase 0/1 → pm.md/tech-lead.md；Phase 4 → qa-engineer.md |
| standards.md | 14KB 大杂烩 | 团队角色 → meta/TEAM_ROLES.md |

**新增文件**：
- `workflows/meta/TEAM_ROLES.md` — 团队角色定义

**修复引用链**：
- dev.md → +devtestops, +test-levels, +specialized-tests
- qa-engineer.md → +devtestops, +test-levels, +specialized-tests, +checklist
- devops.md → +devops-standards
- framework-architect.md → +framework-workflow
- tech-lead.md → +architecture-standards, +review-standards

**增强 prompts**：
- pm.md + PRD 质量标准 + 需求评审流程
- tech-lead.md + 需求评审职责
- qa-engineer.md + 测试执行流程 + 提测准入

### v5.0（2026-04-24）— 三层门禁 + 任务生命周期

- shared.md 重写为 v6.0：三层门禁 + 资产包规格 + 任务生命周期 + 发布流程
- triage.md v5.1：纠正检测 + 分类流程
- bugfix.md v5.0：阶段门禁 + 错误报告

### v3.0（2026-04-17）— 初始重构

- 删除 AGENTS.md
- 精简 SKILL.md
- 两层架构（框架层 + 项目层）
- 文档占位符 `{docs_internal}`

---

## 三、已知问题

### 1. 非 YAML 格式 Prompt

| 文件 | 问题 |
|------|------|
| bugfix.md | 无 YAML frontmatter，AI 工具无法自动加载 standards |
| triage.md | 无 YAML frontmatter，body 中引用 triage-standards.md |

### 2. shared.md 膨胀（20KB）

应只保留门禁 + 工作流 + 任务生命周期。以下内容可考虑拆出：
- SCOPE.md 模板 → templates/
- task-pool/backlog/issues 模板 → templates/
- AI-EXEC-ANCHOR 规范 → 独立文件
- MILESTONES 同步 → 可独立
- 质量门槛 → qa-engineer.md 或独立
- API 问题矩阵 → dev.md 或独立
- 文件空间定义 → 可独立

### 3. 未解决问题

- Git 工作流（AI 修改后人确认手动 commit，无具体流程）
- 变更范围外变更处理流程
- Bugfix 错误报告存放位置
- 流程 AI 读优化

---

## 四、备注

- 本文档随重构进展更新
- 发现新问题在此文档追加
- 重大决策在此文档记录
