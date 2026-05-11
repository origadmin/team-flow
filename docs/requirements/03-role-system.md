# 角色体系

> **版本**: v1.2 | **日期**: 2026-05-11 | **状态**: 待确认

## 1. 角色总览

| 角色 | 职责 | 分发类型 |
|------|------|---------|
| **Triage** | 接收需求 → 分类 → 分发 → 追踪 → 归档 | 主 Agent（不分发） |
| **Tech Lead** | 需求分析 + 技术设计 | `tech-lead-architect` |
| **Dev** | 代码实现 | `developer-engineer` |
| **Bugfix** | Bug 根因分析 + 修复 | `bugfix-expert` |
| **QA** | 验证 + 测试 | `qa-engineer` |
| **Analysis** | 技术分析 | `analysis-expert` |
| **DevOps** | 部署 + CI/CD | `devops-engineer` |
| **UI Designer** | UI 设计 | `ui-designer` |
| **PM** | 验收放行 | `pm-documenter` |

## 2. 技能系统（Skill Profile + Skill Loading）

> **核心思想**：角色必须掌握对应技能才能操作。技能 = 档案（声明会什么）+ 技能文件（实际怎么做）。分发时自动加载对应技能，避免技能冲突和工具混乱。

### 2.1 技能档案（Skill Profile）

技能档案从 `.team/project.md` 的 Tech Stack + Toolchain 动态构建：

```yaml
### Tech Stack (MANDATORY)
- **Language**: go
- **Backend Framework**: gin
- **ORM**: ent
- **Frontend Runtime**: bun
- **Frontend Framework**: react + typescript

### Toolchain (MANDATORY)
- **Package Manager**: bun
- **Build**: bun run build
- **Test**: bun run test
```

### 2.2 技能文件（Skill Files）

每个技术栈对应一个技能文件，包含该技术的**具体规则、约定、命令用法**：

| 技术栈 | 技能文件 | 内容 |
|--------|---------|------|
| go | `skills/go-best-practices/SKILL.md` | Go 编码规范、错误处理、并发模式 |
| gin | `skills/gin/SKILL.md` | Gin 路由、中间件、请求处理约定 |
| ent | `skills/ent/SKILL.md` | Ent schema 定义、迁移、查询约定 |
| bun | `skills/bun/SKILL.md` | bun install/run/add 用法、与 npm 的区别 |
| react | `skills/react/SKILL.md` | React 组件约定、hooks 使用、状态管理 |
| typescript | `skills/typescript/SKILL.md` | TypeScript 类型约定、配置 |

**技能文件是可加载的知识**——不是"禁止什么"，而是"怎么做才对"。

### 2.3 技能映射表

在 `.team/project.md` 中声明技术栈到技能文件的映射：

```yaml
### Skill Mapping (auto-load on dispatch)
- **go**: go-best-practices
- **gin**: gin
- **ent**: ent
- **bun**: bun
- **react**: react
- **typescript**: typescript-advanced-types
```

### 2.4 技能自动加载流程

```
Triage 分发任务
    |
    +-- 1. 确定目标项目（如 orig-cms-ee）
    +-- 2. 确定子类型（如 backend-dev）
    +-- 3. 读取子项目的 .team/project.md
    +-- 4. 提取 Tech Stack → 构建技能档案
    +-- 5. 查找 Skill Mapping → 确定要加载的技能文件
    +-- 6. 加载技能文件内容
    +-- 7. 注入子 Agent prompt：
    |      - 技能档案（声明范围）
    |      - 技能文件内容（具体怎么做）
    +-- 8. 子 Agent 在技能范围内、按技能文件规则操作
```

### 2.5 技能边界规则

| 规则 | 说明 |
|------|------|
| 技能内操作 | 角色只能使用技能档案中列出的技术栈和工具链 |
| 技能外停止 | 遇到技能档案外的技术/工具 → 停止，报告给 Triage |
| 禁止替换 | 不能用技能外的替代品（如定义了 bun 就不能用 npm） |
| 禁止引入 | 不能引入技能外的库/框架（如定义了 ent 就不能引入 gorm） |
| 按技能文件执行 | 在技能范围内，必须按技能文件的规则和约定操作 |

### 2.6 技能学习（Skill Learning）

> **核心思想**：技能文件不存在时，不是简单报错停止，而是像人一样"先学再干"。学过的技能保存下来，下次复用。

#### 学习触发条件

当 Triage 分发任务时，发现 Skill Mapping 中某个技术栈没有对应的技能文件：

```
Tech Stack: go + gin + ent
Skill Mapping: go → go-best-practices ✅ | gin → gin ✅ | ent → ?? ❌
                                                    ent 技能文件不存在
```

#### 学习流程

```
1. Triage 发现技能文件缺失
2. 暂停分发，进入"学习模式"
3. AI 研究该技术：
   a. 搜索官方文档和最佳实践
   b. 分析项目中已有的使用代码（如果有的话）
   c. 生成技能文件草稿
4. 将草稿提交用户审阅
5. 用户确认后，保存为 .trae/skills/{skill}/SKILL.md
6. 继续分发，加载新学的技能
```

#### 学习模式详细步骤

| 步骤 | 操作 | 产出 |
|------|------|------|
| 研究 | 搜索官方文档、GitHub README、最佳实践 | 知识收集 |
| 分析 | 读取项目中已有的该技术相关代码 | 项目约定发现 |
| 生成 | 按技能文件模板生成 SKILL.md 草稿 | 技能文件草稿 |
| 审阅 | 用户确认草稿内容是否正确 | 确认/修改 |
| 保存 | 保存到 `.trae/skills/{skill}/SKILL.md` | 可复用技能 |
| 加载 | 分发时自动加载该技能 | 技能可用 |

#### 技能文件模板

生成的技能文件应包含以下结构：

```markdown
---
name: {technology-name}
description: {when to use, what it provides}
---

# {Technology Name} Skill

## Core Rules
- {该技术的核心规则}

## Commands
- {常用命令及用法}

## Conventions
- {项目中的使用约定}

## Anti-Patterns
- {常见错误和避免方法}

## Integration Points
- {与其他技术栈的集成方式}
```

#### 学习策略

| 场景 | 策略 |
|------|------|
| 项目中已有该技术的代码 | 优先从现有代码学习约定，再补充官方文档 |
| 项目中没有该技术的代码 | 从官方文档和最佳实践学习 |
| 技术非常新/冷门 | 搜索可用资源，生成基础技能，标注"待实践验证" |
| 用户赶时间 | 先用基础技能上岗，后续补充完善 |

#### 技能复用

学过的技能保存在 `.trae/skills/` 下，所有项目共享：

```
.trae/skills/
├── go-best-practices/SKILL.md    ← 学过一次，所有 go 项目复用
├── gin/SKILL.md                   ← 学过一次，所有 gin 项目复用
├── ent/SKILL.md                   ← 学过一次，所有 ent 项目复用
├── bun/SKILL.md                   ← 学过一次，所有 bun 项目复用
└── team-flow/SKILL.md             ← 框架本身的技能
```

**新项目不需要重新学习**——如果新项目也用 `go + gin + ent`，直接复用已有技能文件。

### 2.7 技能冲突避免

| 场景 | 冲突 | 解决方案 |
|------|------|---------|
| 后端 Dev 加载 go + gin + ent | 无冲突，三者互补 | 全部加载 |
| 前端 Dev 加载 bun + react + ts | 无冲突，三者互补 | 全部加载 |
| 后端任务涉及前端文件 | 技能不匹配 | Dev 路由到 frontend-dev 子类型，加载前端技能 |
| 技能文件不存在 | 缺少知识 | 触发技能学习，生成后继续 |
| 多个技能文件有矛盾 | 规则冲突 | 以 Tech Stack 声明为准，技能文件只提供"怎么做" |

### 2.8 技能系统如何解决实际问题

| 问题 | 旧方式 | 新方式 |
|------|--------|--------|
| bun 项目用了 npm | 规则"禁止 npm"→ AI 无视 | 技能档案只有 bun → npm 不在技能中 → 自动停止 |
| ent 项目用了 gorm | 规则"禁止 gorm"→ AI 无视 | 技能档案只有 ent → gorm 不在技能中 → 自动停止 |
| 不知道 bun 怎么用 | AI 凭记忆，可能用错 | 加载 bun SKILL.md → 按技能文件执行 |
| gin 项目用了 echo | AI 凭记忆选框架 | 加载 gin SKILL.md → 按 gin 约定开发 |

## 3. Triage 职责边界

**Triage 是协调者，不是执行者。**

### 门禁式控制

| 操作 | 前置条件 | 不满足时 |
|------|---------|---------|
| 写代码/读代码/调试 | 已分发到子 Agent | 禁止执行，分发出去 |
| 创建 SPEC/AC/RCA | 已分发到子 Agent | 禁止执行，分发出去 |
| 做架构决策 | 已分发到 Tech Lead | 禁止执行，分发出去 |
| "顺手做了" | — | 禁止，即使最简单也必须分发 |

## 4. 分发映射

| 任务类型 | 分发到 | 子 Agent 类型 |
|---------|--------|-------------|
| Feature | Tech Lead → Dev | `tech-lead-architect` → `developer-engineer` |
| Bug | Dev (Bugfix) | `bugfix-expert` |
| Change | Tech Lead | `tech-lead-architect` |
| Docs | Tech Lead | `tech-lead-architect` |
| Analysis | Analysis | `analysis-expert` |
| UI | UI Designer | `ui-designer` |
| DevOps | DevOps | `devops-engineer` |
| PM | PM | `pm-documenter` |

## 5. 分发时注入技能

Triage 分发任务时，注入两部分内容：

### 5.1 技能档案（声明范围）

```
## Skill Profile (MANDATORY — operate only within this profile)
- Language: go
- Backend Framework: gin
- ORM: ent
- Package Manager: bun
- Build: bun run build
- Test: bun run test

## Skill Boundary
- You can ONLY use technologies listed above
- If you encounter something outside your skill profile → STOP and report back
```

### 5.2 技能文件内容（具体怎么做）

```
## Skill: go-best-practices
[go-best-practices/SKILL.md 的内容]

## Skill: gin
[gin/SKILL.md 的内容]

## Skill: ent
[ent/SKILL.md 的内容]
```

### 5.3 注入流程

```
1. Triage 确定目标项目（如 orig-cms-ee）
2. 确定子类型（如 backend-dev）
3. 读取子项目的 .team/project.md → 获取 Tech Stack + Toolchain + Skill Mapping
4. 构建技能档案
5. 按 Skill Mapping 加载技能文件
6. 将技能档案 + 技能文件内容注入子 Agent prompt
7. 子 Agent 在技能范围内、按技能文件规则操作
```

## 6. 角色切换规则

```
[Triage] 分析 → 分发（附带技能档案 + 技能文件） → [子 Agent] 执行 → [Triage] 汇总结果
```

- 子 Agent 完成后返回结果给 Triage
- Triage 更新 beads 状态
- Triage 报告结果给用户
- 角色切换时 Status Line 的 Role 字段随之变化

## 7. Dev 子类型路由

Dev 接到任务后，根据涉及文件判断子类型：

| 条件 | 子类型 | 加载技能 | 技能文件 |
|------|--------|---------|---------|
| 涉及 `web/src/**`、`*.tsx`、`*.css`、React | frontend-dev | 前端 Tech Stack | bun + react + typescript SKILL.md |
| 涉及 `internal/**`、`*.go`、`proto`、API | backend-dev | 后端 Tech Stack | go + gin + ent SKILL.md |

子类型确定后，Dev 只加载对应技能文件，不加载无关技能。
