# team-flow 安装指南

> 如何在项目中使用 team-flow 多角色协作系统
> **TEAM_VERSION=6.0** | 2026-04-25

---

## 快速开始

### 步骤 1：创建项目 SKILL.md

在你的项目根目录创建 `SKILL.md`：

```markdown
加载 {TEAM_PATH}/SKILL.md，按其中步骤执行。

## 项目信息

- 项目名: your-project
- 技术栈: Go + React
- 工作目录: .
```

### 步骤 2：配置你的 AI 工具

在 AI 工具中设置系统提示词，指向项目 SKILL.md 即可。

支持的 AI 工具：
- Trae / Gemini CLI / OpenClaw / WorkBuddy / QClaw

### 步骤 3：验证

```
初始化项目，检查 team-flow 状态
```

AI 会自动：
1. 加载框架 SKILL.md
2. 创建 `.team/` 目录（含 project.md, task-pool.md）
3. 进入工作状态

---

## 目录结构

```
{TEAM_PATH}/
├── SKILL.md                    ← 入口（AI 加载）
├── INSTALL.md                  ← 本文件
├── README.md                   ← 框架说明（人看的）
│
├── workflows/                  ← 工作流程（AI 执行）
│   ├── shared.md               ← 核心流程
│   ├── framework-workflow.md   ← 框架开发流程
│   ├── meta/TEAM_ROLES.md      ← 团队角色定义
│   └── roles/                  ← 14 个角色规范文件
│
├── prompts/                    ← 11 个角色提示词（YAML frontmatter）
│
├── templates/                  ← 文档模板
│   ├── project.md
│   ├── task-pool.md
│   └── ...
│
└── config/
    └── ...
```

---

## 常见问题

**Q: 项目 SKILL.md 和框架 SKILL.md 什么关系？**
A: 项目 SKILL.md 是入口，它告诉 AI 去哪里加载框架 SKILL.md。框架 SKILL.md 是实际驱动逻辑。

**Q: 为什么需要两层？**
A: 项目层存放项目特定信息（名称、路径、技术栈），框架层存放通用规则。分离后框架可共享，项目可定制。

**Q: 如何更新 team-flow？**
A: 引用方式：更新框架目录，所有项目自动生效。复制方式：重新复制到各项目。

**Q: 可以修改框架内容吗？**
A: 引用方式不建议修改（影响所有项目）。复制方式可以修改（仅影响当前项目）。

**Q: 多人开发时路径冲突怎么办？**
A: 每人在自己的工具配置中设置自己的环境变量值，`.team/project.md` 引用变量不写死路径。
