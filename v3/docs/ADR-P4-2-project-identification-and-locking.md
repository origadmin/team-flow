# ADR-P4-2: Project Identification and Locking in Monorepo

> **Status**: Approved | **Date**: 2026-05-25 | **Author**: team-flow | **Reviewer**: Pending

## 1. Summary

在 monorepo 场景下，`flow proc run` 输出的 PROJECT_ROOT 指向 workspace 根目录而非用户实际工作的代码项目目录，导致 AI 在错误路径下执行命令。本 ADR 定义项目辨识与锁定机制，确保 AI 始终在正确的项目目录下工作。

## 2. Problem Statement

### 2.1 缺陷

当前 `FindProjectRoot()` 从 cwd 向上搜索 `.team/` 目录，在 monorepo 下总是返回 workspace 根：

```
framework/                    ← 有 .team/ → FindProjectRoot 返回这里
framework/projects/orig-cms-ee/  ← 用户实际工作目录
```

AI 看到 `PROJECT_ROOT: framework/`，认为这是"我的项目"，在 framework/ 下执行 `bd create`、`go test` 等命令，但实际应该在 `projects/orig-cms-ee/` 下执行。

### 2.2 根因

**两个不同的"项目"概念被一个变量承载：**

| 概念 | 含义 | 路径 | 谁需要它 |
|------|------|------|----------|
| Team Project | team-flow 配置所在目录 | `framework/`（有 `.team/`） | 引擎：找 flows、roles、rules |
| Code Project | AI 实际工作的代码目录 | `framework/projects/orig-cms-ee/` | AI：执行 go test、bd create、编辑代码 |

`PROJECT_ROOT` 承载了 Team Project 语义，但 AI 把它当作 Code Project 使用。

### 2.3 时序问题

项目上下文在用户消息中出现（T1），但 `flow proc run` 在会话启动时运行（T0），此时不知道用户要工作在哪个子项目。

### 2.4 之前方案的错误

| 方案 | 为什么不行 |
|------|-----------|
| 加 `--project` 参数 | AI 不知道什么时候传 |
| 加 `PROJECT_CWD` 字段 | AI 看到 PROJECT_ROOT 就认为那是项目，忽略 CWD |
| 让 `FindProjectRoot` 更聪明 | T0 时没有项目上下文，无法推断 |
| 加 `AI_BEHAVIOR` 指令 | 补丁而非解决 |

## 3. Decision

### 3.1 核心原则

| 原则 | 说明 |
|------|------|
| **配置优于猜测** | workspace 配置显式声明项目，不靠扫描猜 |
| **cwd 优于参数** | 通过 cd 切换目录，不通过 --project 参数 |
| **锁定优于自由** | 对话内项目锁定，不自动切换 |
| **显式优于隐式** | 无法确定时询问用户，不静默猜测 |

### 3.2 概念拆分

```
PROJECT_ROOT = AI 工作的代码项目目录（始终指向代码项目）
TEAM_ROOT    = team-flow 配置目录（引擎读 flows/roles/rules 的地方）

单项目场景：PROJECT_ROOT == TEAM_ROOT
Monorepo 场景：PROJECT_ROOT != TEAM_ROOT
```

### 3.3 项目发现算法

```
给定 workspace root，发现所有项目：

优先级 1：workspace 配置显式声明（.team/project.yaml 中的 projects 字段）
优先级 2：go.work 中的 module 条目
优先级 3：projects/ 子目录扫描（检查 go.mod / package.json / .team/）
优先级 4：workspace 根直接子目录扫描（兜底）
```

### 3.4 项目选择算法

```
发现项目列表后，确定当前项目：

优先级 1：cwd 已经在某个项目目录内 → 自动锁定
优先级 2：只有一个项目 → 自动选择
优先级 3：用户消息中提到项目名 → 匹配项目列表
优先级 4：多个项目且无法推断 → 必须询问用户
```

### 3.5 Session Startup Protocol 修正

```
Step 1: flow project detect
         → 检测 workspace 和项目列表
         → 如果 cwd 在子项目内 → 自动锁定
         → 如果 monorepo 且未锁定 → 输出项目列表

Step 2: PROJECT SELECTION（仅未锁定时）
         → AI 从用户意图推断项目，或询问用户
         → cd 到选定项目目录
         → 如果项目没有 .team/，使用 workspace 的 team 配置

Step 3: flow proc run
         → PROJECT_ROOT = cwd（AI 工作目录）
         → TEAM_ROOT = .team/ 所在目录（引擎配置目录）

Step 4: PROJECT LOCK
         → PROJECT_ROOT 写入会话上下文
         → 后续所有操作锁定在此项目
```

### 3.6 workspace 配置格式

```yaml
# .team/project.yaml
projects:
  - name: orig-cms-ee
    path: projects/orig-cms-ee
    type: go
    team: dev-team
    
  - name: origadmin
    path: projects/origadmin
    type: go
    team: dev-team
```

### 3.7 flow project detect 输出

```
场景 1：cwd 在子项目内（自动锁定）
  WORKSPACE: D:\...\framework
  PROJECT: orig-cms-ee (locked) → D:\...\framework\projects\orig-cms-ee
  TEAM_CONFIG: D:\...\framework\.team (dev-team)
  STATUS: ready → run flow proc run

场景 2：cwd 在 workspace 根（需要选择）
  WORKSPACE: D:\...\framework
  AVAILABLE PROJECTS:
    [1] orig-cms-ee → projects/orig-cms-ee/ (go, .team: ✅)
    [2] origadmin   → projects/origadmin/   (go, .team: ✅)
    [3] team-flow   → projects/team-flow/   (go, .team: ✅)
  STATUS: project not selected → ask user or infer from context

场景 3：单项目（无需选择）
  WORKSPACE: /home/user/my-app
  PROJECT: my-app (locked)
  STATUS: ready → run flow proc run
```

### 3.8 flow proc run 输出变更

```
单项目场景（不变）：
  PROJECT_ROOT: /home/user/my-app
  （TEAM_ROOT 不显示，因为和 PROJECT_ROOT 相同）

Monorepo 场景（修复后）：
  PROJECT_ROOT: D:\...\framework\projects\orig-cms-ee
  TEAM_ROOT: D:\...\framework
```

## 4. Cross-Team Compatibility Analysis

| Team | 场景 | 影响 |
|------|------|------|
| dev-team | monorepo | ✅ 修复：PROJECT_ROOT 指向正确的子项目 |
| content-team | 单项目 | ✅ 无影响：PROJECT_ROOT = TEAM_ROOT |
| game-team | 单项目 | ✅ 无影响 |
| trading-team | 单项目 | ✅ 无影响 |
| skill-team | monorepo | ✅ 修复：同 dev-team |

## 5. Consequences

### 5.1 正面

- PROJECT_ROOT 语义明确：始终指"AI 工作目录"
- Monorepo 场景下 AI 不再在错误路径执行命令
- 项目锁定防止对话内项目漂移
- 不需要 AI 传任何参数

### 5.2 风险

- PROJECT_ROOT 语义变更可能导致依赖旧行为的脚本/规则出错
- workspace 配置需要手动维护项目列表
