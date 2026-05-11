# 核心需求

> **版本**: v1.1 | **日期**: 2026-05-11 | **状态**: 待确认

## 1. 框架定位

team-flow 是 AI 协作框架，不是业务代码。它的职责是：

1. **协调**：接收用户需求 → 分类 → 分发给正确的角色
2. **追踪**：任务状态在 beads 中管理，可追溯
3. **门禁**：确保流程不被跳过，质量有保障
4. **隔离**：框架层只读，项目层可写，不交叉污染

## 2. 核心原则

| # | 原则 | 说明 |
|---|------|------|
| 1 | 配置驱动，不猜测 | AI 读取配置决定行为，禁止启发式检测 |
| 2 | 单一信息源 | beads 是任务状态的唯一真相源 |
| 3 | 框架 ≠ 项目 | 框架层只读，项目问题在项目层解决（见下方模式差异） |
| 4 | Triage 不执行 | Triage 只协调，不写代码/不调试/不做设计 |
| 5 | 每步确认 | 修改前必须确认理解，不确定就停 |
| 6 | 成果物分离 | 详情写入独立文件，不塞进 beads notes |

### 原则 3 的模式差异

"框架 ≠ 项目"在不同管理模式下边界不同：

| 模式 | 框架层 | 项目层 | 边界 |
|------|--------|--------|------|
| standalone | `{TEAM_PATH}/` 只读 | `{PROJECT}/` 可写 | 严格分离，框架和项目在不同目录 |
| workspace | `{TEAM_PATH}/` 只读 | `{WORKSPACE}/` 可写（含多个子项目） | 框架在 workspace 内，但仍是只读；项目操作在子项目目录内 |

**关键**：无论哪种模式，`{TEAM_PATH}/` 始终只读。区别在于项目操作的范围——standalone 在单一项目目录，workspace 在 workspace 内的子项目目录。

## 3. `.team/project.md` 的角色：从"信息"到"规则"

### 问题

当前 `.team/project.md` 是信息性的——AI 读了但不一定遵守。比如：
- 定义了 `bun`，AI 执行到一半换成 `npm`
- 定义了 `go + gin + ent`，AI 开发到一半换成 `gorm`
- 定义了工具链，AI 用默认值覆盖

### 方案

`.team/project.md` 应该是**强制性规则配置**，AI 必须遵守。内容从"项目描述"升级为"项目规则"：

```yaml
## Project Rules

### Management
- **Mode**: workspace

### Tech Stack (MANDATORY — AI must use these, no substitution)
- **Language**: go
- **Backend Framework**: gin
- **ORM**: ent
- **Frontend Runtime**: bun
- **Frontend Framework**: react + typescript

### Toolchain (MANDATORY — AI must use these exact commands)
- **Package Manager**: bun          # → bun install, bun run, bun add
- **Build**: bun run build
- **Test**: bun run test
- **Lint**: bun run lint
- **TypeCheck**: bun run typecheck

### Registered Projects (workspace mode)
| Project | Path | Status | Description |
|---------|------|--------|-------------|
| framework | ./ | active | Workspace host |
| orig-cms-ee | ./projects/orig-cms-ee/ | active | CMS project |
```

### 强制执行机制

| 规则类型 | 执行方式 |
|---------|---------|
| Tech Stack | AI 选择库/框架时，必须从 Tech Stack 列表中选择，禁止替换 |
| Toolchain | AI 执行命令时，必须使用 Toolchain 中定义的命令，禁止替换 |
| Management Mode | AI 确定项目上下文时，必须读取 Mode 配置，禁止猜测 |

**Tech Stack 违规示例**：
- 定义 `ORM: ent` → AI 使用 `gorm` → ❌ 违规
- 定义 `Package Manager: bun` → AI 使用 `npm install` → ❌ 违规
- 定义 `Backend Framework: gin` → AI 引入 `echo` → ❌ 违规

**Toolchain 违规示例**：
- 定义 `Package Manager: bun` → AI 执行 `npm run build` → ❌ 违规
- 定义 `Test: bun run test` → AI 执行 `npm test` → ❌ 违规

## 4. 流程控制（替代"不应该做"列表）

> **核心思想**：不用负面约束（"不要做 X"），用正面流程控制（"只有满足条件 Y 时才能做 X"）。

### 门禁式流程控制

| 操作 | 前置条件 | 不满足时 |
|------|---------|---------|
| 创建 beads 任务 | Triage 已完成分类 | 禁止创建，先完成分类 |
| 分发到子 Agent | 分类报告已输出 | 禁止分发，先输出报告 |
| 标记任务 Review | 完成门禁检查全部通过 | 禁止标记，补齐缺失项 |
| 归档任务 | 用户已确认 | 禁止归档，等待确认 |
| 使用工具链命令 | 已读取 project.md Toolchain | 禁止执行，先读取配置 |
| 选择技术栈 | 已读取 project.md Tech Stack | 禁止选择，先读取配置 |
| 写入框架层 | 用户明确要求修改框架 | 禁止写入，框架层只读 |
| 删除文件 | 用户明确要求删除 | 禁止删除 |
| git commit | 用户明确要求提交 | 禁止提交 |

### 替代原来的"不应该做"和"可以做"

原来的"不应该做"列表 → 改为上表的门禁式控制
原来的"可以做"列表 → 改为上表的"前置条件"列

**优势**：
- 不再给 AI 一个"不该做的事"清单（有时反而诱导错误）
- 每个操作都有明确的前置条件，不满足就禁止执行
- AI 不需要记住"不该做什么"，只需要检查"能不能做"

## 5. 工具链规则

> **核心**：AI 必须使用 project.md 中定义的工具链和技术栈，禁止替换。

### 规则

| 规则 | 说明 |
|------|------|
| 读取优先 | 执行任何构建/测试/安装命令前，必须先读取 project.md Toolchain |
| 严格匹配 | 定义了 `bun` 就用 `bun`，定义了 `npm` 就用 `npm` |
| 禁止替换 | 不允许"bun 不可用则回退 npm"——如果不可用应该报错 |
| 禁止混用 | 同一项目中不能混用 bun 和 npm |
| Tech Stack 同理 | 定义了 `ent` 就用 `ent`，不能中途换成 `gorm` |

### PRE-FLIGHT 检查

AI 执行任何命令前必须输出：

```
PRE-FLIGHT: pkg=bun, orm=ent, framework=gin
```

与 project.md 定义不一致 → 禁止执行。
