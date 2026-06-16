# 系统架构 (System Architecture)

> **版本**: v3.0  
> **状态**: 核心文档  
> **目的**: 定义 team-flow 的整体系统架构。修改任何代码前必须理解此文档。  
> **最后更新**: 2026-06-09

---

## 1. 项目定位

team-flow 是一个**基于 Go 的 CLI 工具**，用于驱动 AI 团队协作流程。它：

- 定义可执行的流程（flow JSON）
- 驱动节点推进（`flow proc run`）
- 管理任务生命周期（`flow task create/show/update`）
- 集成外部工具（beads 任务管理系统）

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              team-flow                                  │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐              │
│  │   flow JSON  │──▶│  Proc Engine │──▶│   AI Agent   │              │
│  │  (流程定义)  │   │  (节点推进)   │   │  (执行任务)   │              │
│  └──────────────┘   └──────────────┘   └──────┬───────┘              │
│                                               │                       │
│  ┌──────────────┐   ┌──────────────┐         │                       │
│  │  .team/      │◀──│  task store  │◀────────┘                       │
│  │  (持久化)     │   │  (JSON files) │                                 │
│  └──────────────┘   └──────────────┘                                  │
│                                               │                       │
│  ┌──────────────┐   ┌──────────────┐         │                       │
│  │  beads CLI   │◀──│  bd client   │◀────────┘                       │
│  │  (外部工具)   │   │  (stdout/stderr)│                               │
│  └──────────────┘   └──────────────┘                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 2. 代码结构（Code Organization）

```
projects/team-flow/
├── cmd/flow/
│   ├── main.go              # CLI 入口，注册所有子命令
│   └── debug.go             # 调试工具
│
├── internal/                # 核心库（importable 包）
│   ├── flow/                # Flow 数据模型 + 解析 + 序列化 + 验证
│   │   ├── types.go         # 核心类型: NodeType, TaskType, Flow, FlowNode 等
│   │   ├── parser.go        # Flow JSON 解析
│   │   ├── serializer.go    # Flow JSON 序列化
│   │   ├── validator.go     # Flow 结构验证
│   │   └── docspec.go       # 文档规范定义
│   │
│   ├── proc/                # 流程引擎（核心）
│   │   ├── cmd.go           # flow proc 子命令注册 (run/next/gate/...)
│   │   ├── procrun.go       # ProcRunEngine: 节点执行核心
│   │   ├── resolver.go      # Flow/节点解析
│   │   ├── formatter.go     # 输出格式化（JSON/text）
│   │   ├── gate_checker.go  # 门控条件检查
│   │   ├── gate_handler.go  # 门控通过/失败处理
│   │   ├── substitutor.go   # 变量替换（{{task_id}}等）
│   │   └── startup_check.go # Session Start 检查
│   │
│   ├── task/                # 任务管理
│   │   ├── task.go          # flow task 子命令 + CRUD 逻辑
│   │   └── trace.go         # 任务追踪
│   │
│   ├── bd/                  # beads CLI 客户端
│   │   ├── client.go        # bd 命令封装 + CreateIssue + RunQuiet
│   │   └── client_test.go   # 单元测试
│   │
│   ├── config/              # 配置系统
│   │   ├── config.go        # 项目配置 + 路径解析
│   │   └── project.go       # 项目级配置
│   │
│   ├── session/             # 会话管理
│   │   └── session.go
│   │
│   ├── eventlog/            # 事件日志（流程执行记录）
│   │   └── eventlog.go
│   │
│   ├── skill/               # 技能系统
│   │   ├── manager.go       # 技能加载/管理
│   │   ├── resolver.go      # 技能引用解析
│   │   ├── cache.go         # 技能缓存
│   │   └── cmd.go           # flow skill 子命令
│   │
│   ├── project/             # 项目管理
│   │   ├── manager.go       # 项目发现/切换
│   │   └── cmd.go           # flow project 子命令
│   │
│   ├── editor/              # 流程编辑器 API
│   │   ├── api.go
│   │   └── editor.go
│   │
│   ├── migrate/             # v2 → v3 迁移
│   │   └── migrate.go
│   │
│   ├── validate/            # flow 验证
│   │   └── validate.go
│   │
│   ├── doctor/              # 诊断工具
│   │   └── doctor.go
│   │
│   ├── status/              # 状态检查
│   │   └── status.go
│   │
│   ├── export/              # 数据导出
│   │   └── export.go
│   │
│   ├── graph/               # 图分析
│   │   └── graph.go
│   │
│   ├── version/             # 版本信息
│   │   └── version.go
│   │
│   ├── update/              # 更新检查
│   │   └── update.go
│   │
│   ├── logger/              # 日志
│   │   └── logger.go
│   │
│   ├── idgen/               # ID 生成
│   │   └── idgen.go
│   │
│   ├── boot/                # 启动初始化
│   │   ├── init.go
│   │   └── session.go
│   │
│   └── trace/               # 执行追踪
│       └── trace.go
│
├── assets/flows/            # 预置 flow 模板（由 flow 使用和参考）
│   ├── dev-flow.json        # 通用开发流程
│   └── skill-dev-flow.json  # SKILL 开发流程
│
├── assets/skill/v3/         # 技能模板和资源
│   ├── workflows/           # 工作流规则
│   │   ├── shared.md
│   │   ├── framework-workflow.md
│   │   └── roles/*.md       # 各角色标准
│   ├── templates/           # 文档模板
│   ├── prompts/             # 角色提示词
│   ├── BOUNDARY.md          # 边界定义
│   └── SKILL.md             # 技能入口
│
├── v3/docs/                 # 架构说明和决策（历史记录）
│   ├── DOC-ARCH-001-documentation-architecture.md
│   ├── DECISIONS.md
│   ├── IMPLEMENTATION-STATUS.md
│   ├── ADR-*.md             # 架构决策记录（历史）
│   └── CORRECTION-*.md      # 修正说明
│
├── docs/                    # ⭐ 核心文档（修改代码前必读）
│   ├── ARCHITECTURE.md      # 本文档
│   ├── DESIGN.md            # 设计原则
│   ├── CONSENSUS.md         # 项目共识
│   └── STANDARDS.md         # 编码和测试标准
│
├── .team/                   # 项目级运行时数据（不提交到代码库的动态数据除外）
│   ├── project.md           # 项目配置（active_flow 等）
│   ├── team.json            # 团队定义（角色/规则）
│   ├── tasks/               # 任务文件（{task-id}.json）
│   └── flows/               # 项目级自定义 flow
│
├── go.mod                   # Go module: github.com/origadmin/team-flow
├── Makefile                 # 构建脚本
└── README.md                # 项目入口（链接到核心文档）
```

---

## 3. 核心数据流（Core Data Flow）

### 3.1 命令执行路径：`flow task create --title "..."`

```
1. main.go: rootCmd → task.Cmd
2. task.go: runTask() → taskCreateV3()
3. taskCreateV3():
   ├── 解析参数 (--title, --type, --description)
   ├── 调用 bd.CreateIssue(title, type, description)  获取 beads 格式 ID
   ├── 创建 Task 对象 { ID, Title, Type, Status: "open", CreatedAt }
   ├── saveTask() → 写入 .team/tasks/{task-id}.json
   └── 写事件日志 eventlog.TaskCreated(...)
```

### 3.2 流程执行路径：`flow proc run`

```
1. main.go: rootCmd → proc.Cmd → runCmd
2. proc/cmd.go: runRun():
   ├── 解析参数 (--task, --analysis, --conclusion, --new)
   ├── 创建 ProcRunRequest
   └── 调用 ProcRunEngine.Run(ctx, request)

3. proc/procrun.go: ProcRunEngine.Run():
   ├── 检查/创建 session（eventlog + .team/version）
   ├── 从 project.md 读取 active_flow → 解析 flow JSON
   ├── 查找目标节点（指定 node-id 或 root node）
   ├── 收集节点角色/规则/工具/文档
   ├── 应用变量替换（{{task_id}} 等）
   ├── 构建 ProcRunResult（包含当前节点信息 + 下一步选项）
   └── 格式化输出（JSON 或 text）

4. AI Agent 读取 ProcRunResult → 执行节点任务 → 调用 flow proc run next
```

### 3.3 Beads 集成：`bd.CreateIssue()`

```
Go 代码 (bd/client.go)
    │
    │  exec.Command("bd", "create", "--title", title, "--type", type, "--silent")
    │  注意: 使用 cmd.Output() 只捕获 stdout
    │
    ▼
beads CLI (外部进程)
    │
    │  stderr: 警告信息（"no beads configuration found" 等）
    │  stdout: team-flow-xxx  ← 真正的 ID
    │
    ▼
Go 代码: parseBeadsID(output)
    │
    │  解析规则：
    │  - 跳过以 "warning" 开头的行（防御性）
    │  - 查找包含 "-" 的非空行（beads ID 格式: project-xxx）
    │  - 返回第一个匹配项
    │
    ▼
返回: "team-flow-xxx" (或错误)
```

**设计要点**：
- `cmd.Output()` — 只捕获 stdout，警告信息在 stderr 被丢弃
- `parseBeadsID()` — 防御性解析，容忍输出中的冗余行
- ID 格式：`project-xxx`（由 beads CLI 生成，包含连字符）

---

## 4. 模块依赖关系（Module Dependencies）

```
main (cmd/flow)
  │
  ├── proc (流程引擎)
  │   ├── flow (数据模型)
  │   ├── config (项目配置)
  │   ├── eventlog (事件日志)
  │   ├── skill (技能系统)
  │   └── session (会话)
  │
  ├── task (任务管理)
  │   ├── bd (beads 客户端)
  │   ├── config (项目配置)
  │   └── eventlog (事件日志)
  │
  ├── bd (beads 客户端)
  │   └── [外部进程: bd CLI]
  │
  ├── skill (技能系统)
  │   ├── cache (缓存)
  │   └── resolver (引用解析)
  │
  ├── config (项目/工作区配置)
  │
  ├── project (项目发现/管理)
  │   └── config
  │
  ├── editor (流程编辑器 API)
  │   └── flow
  │
  ├── validate (flow 验证)
  │   └── flow
  │
  ├── migrate (v2 → v3)
  │   └── config
  │
  ├── version (版本信息)
  │
  └── [其他模块: doctor, status, export, graph, update, tools]
```

### 依赖原则

1. **`flow` 包是纯数据层** — 只定义类型和解析逻辑，不依赖任何其他内部包
2. **`proc` 包是核心执行层** — 依赖 `flow`、`config`、`eventlog`、`session`、`skill`
3. **`bd` 包是独立的客户端层** — 只依赖标准库 `os/exec`，可独立测试
4. **`task` 包依赖 `bd`** — 任务 ID 生成委托给 beads
5. **`main` 是薄的组装层** — 只负责注册命令和启动引擎

---

## 5. flow JSON 结构（Flow Definition）

每个 flow 是一个 JSON 文件，结构如下：

```javascript
{
  "version": "v3",
  "metadata": {
    "id": "skill-dev-flow",          // flow 唯一标识
    "name": "Skill Development Flow",// 显示名
    "description": "...",            // 说明
    "tags": ["development", "skill"] // 标签
  },
  "config": {
    "task_type": "feature",          // 默认任务类型
    "active_flow": true              // （可在 project.md 中引用）
  },
  "nodes": [
    {
      "id": "a26c80",                // 节点 ID（随机短码，非语义化）
      "type": "start",               // start | phase | gate | branch | terminal
      "name": "Session Start",       // 人类可读名称
      "description": "...",
      "components": {                // 此节点关联的组件
        "roles": [...],              // 角色引用
        "rules": [...],              // 规则引用
        "tools": [...],              // 工具引用
        "skills": [...],             // 技能引用
        "prompts": [...],            // 提示词引用
        "docs": [...]                // 文档引用
      },
      "on_enter": [...],             // 进入时动作
      "on_exit": [...]               // 退出时动作
    }
  ],
  "edges": [
    {
      "id": "e-001",
      "from": "a26c80",              // 源节点 ID
      "to": "c3d4e5",                // 目标节点 ID
      "type": "sequential",          // sequential | conditional
      "conditions": [...]            // 条件边的条件定义
    }
  ]
}
```

**关键设计**：节点 ID 是随机短码（如 `a26c80`），**不是**语义化名称。语义在 `name` 字段中。这保持了 flow 拓扑和节点含义的分离，便于重构和替换节点实现。

---

## 6. 持久化（Persistence）

### 6.1 存储位置

| 数据类型 | 位置 | 格式 |
|---------|------|------|
| 任务 | `.team/tasks/{task-id}.json` | JSON |
| 会话/事件 | `.team/sessions/` 或 `.team/eventlog.jsonl` | JSON Lines |
| 项目配置 | `.team/project.md` | Markdown + YAML frontmatter |
| 团队配置 | `.team/team.json` | JSON |
| Flow 定义 | `assets/flows/*.json` 或 `.team/flows/*.json` | JSON |

### 6.2 任务文件格式（`.team/tasks/{task-id}.json`）

```json
{
  "id": "team-flow-abc",
  "title": "任务标题",
  "type": "bug",
  "status": "open",
  "description": "详细描述",
  "parent": "team-flow-parent",
  "labels": {},
  "notes": [],
  "created_at": "2026-06-09T10:00:00Z",
  "updated_at": "2026-06-09T10:00:00Z",
  "closed_at": null,
  "extras": {}
}
```

### 6.3 Team JSON 格式（`.team/team.json` 或 `assets/orgs/{org}/team.json`）

**架构原则：team.json 只管理引用，不管理内容。**

team.json 是团队的**装配清单**，不是角色/规则内容的容器。所有角色内容（persona、traits、guidance、constraints 等）和规则内容（instruction、description 等）都必须放在 `assets/skill/v3/prompts/*.md` 文件中，由 team.json 通过 ID 引用。

```
                                      ┌──────────────────────┐
                                      │   team.json (装配)    │
                                      │  {id, name, version,  │
                                      │   roles: [{id, prompt_│
                                      │   source, principal}],│
                                      │   rules: [{id}],      │
                                      │   flows: [{id, type}] │
                                      └─────────┬────────────┘
                                                │引用
                        ┌───────────────────────┼───────────────────────┐
                        │                       │                       │
             ┌──────────▼─────────┐   ┌────────▼──────────┐   ┌──────▼────────┐
             │ prompts/dev.md    │   │ prompts/tech-lead │   │ assets/flows/ │
             │ (persona, traits, │   │ .md              │   │ dev-flow.json  │
             │  guidance, rules, │   │ (persona, traits, │   │ (节点定义，    │
             │  standards)       │   │ guidance, etc.)   │   │ 引用角色ID)    │
             └───────────────────┘   └───────────────────┘   └───────────────┘
```

#### 6.3.1 为什么 team.json 不能管理内容

1. **关注点分离**：装配和内容是两个不同维度的关注点。team.json 负责"谁在什么流程中扮演什么角色"，prompts 负责"这个角色是什么"。
2. **避免重复和漂移**：如果每个 org 的 team.json 都复制一份角色定义，角色迭代时必然出现"这个 org 用的是旧版本的 dev 角色"的问题。
3. **格式演进困难**：修改 RoleDefinition 字段需要同时更新几十个 team.json 文件，容易遗漏。
4. **测试与验证困难**：角色定义嵌入 JSON 后，难以独立测试和验证。
5. **违反单一职责**：team.json 同时负责装配和内容，导致"改一处碰全局"的脆弱性。

#### 6.3.2 v3-only 格式（唯一合法格式）

```json
{
  "id": "dev-team",
  "name": "Software Development Team",
  "version": "v3",
  "schema_version": "3.0",
  "author": "team-flow",
  "flows": [
    {"id": "dev-flow", "type": "main", "default": true},
    {"id": "feature-flow", "type": "sub", "parent": "dev-flow"}
  ],
  "roles": [
    {
      "id": "concierge",
      "prompt_source": "assets/skill/v3/prompts/concierge.md",
      "principal": false
    },
    {
      "id": "triage",
      "prompt_source": "assets/skill/v3/prompts/triage.md",
      "principal": true
    },
    {
      "id": "tech-lead",
      "prompt_source": "assets/skill/v3/prompts/tech-lead.md",
      "principal": false
    },
    {
      "id": "dev",
      "prompt_source": "assets/skill/v3/prompts/dev.md",
      "principal": false
    },
    {
      "id": "qa-engineer",
      "prompt_source": "assets/skill/v3/prompts/qa-engineer.md",
      "principal": false
    },
    {
      "id": "devops",
      "prompt_source": "assets/skill/v3/prompts/devops.md",
      "principal": false
    },
    {
      "id": "analysis",
      "prompt_source": "assets/skill/v3/prompts/analysis.md",
      "principal": false
    },
    {
      "id": "pm",
      "prompt_source": "assets/skill/v3/prompts/pm.md",
      "principal": false
    }
  ],
  "rules": [
    {"id": "d1a", "source": "assets/skill/v3/prompts/rules.md"},
    {"id": "d2b", "source": "assets/skill/v3/prompts/rules.md"},
    {"id": "d3c", "source": "assets/skill/v3/prompts/rules.md"},
    {"id": "no-broken-deploy", "source": "assets/skill/v3/prompts/rules.md"}
  ]
}
```

**字段约束**：

| 字段 | 旧格式 (非法) | 新格式 (v3) |
|------|--------------|-------------|
| `roles[].id` | ✓ 必需 | ✓ 必需 |
| `roles[].name` | ✓ 有 | ✗ **禁止** (在 prompt_source 的 md 文件中) |
| `roles[].alias` | ✓ 有 | ✗ **禁止** (在 prompt_source 的 md 文件中) |
| `roles[].alias_en` | ✓ 有 | ✗ **禁止** (在 prompt_source 的 md 文件中) |
| `roles[].persona` | ✓ 有 | ✗ **禁止** (在 prompt_source 的 md 文件中) |
| `roles[].traits` | ✓ 有 | ✗ **禁止** (在 prompt_source 的 md 文件中) |
| `roles[].guidance` | ✓ 有 | ✗ **禁止** (在 prompt_source 的 md 文件中) |
| `roles[].prompt_source` | ✓ 可选 | ✓ **必需** |
| `roles[].rules` | ✓ 有 (内联 ID 列表) | ✗ **禁止** (在 prompt_source 的 md 文件中) |
| `rules[].id` | ✓ 必需 | ✓ 必需 |
| `rules[].instruction` | ✓ 有 | ✗ **禁止** (在 source 的 md 文件中) |
| `rules[].description` | ✓ 有 | ✗ **禁止** (在 source 的 md 文件中) |
| `rules[].enforcement` | ✓ 有 | ✗ **禁止** (在 source 的 md 文件中) |
| `rules[].source` | - | ✓ **必需** |
| `rules[].type` | ✓ 有 | ✗ **禁止** (在 source 的 md 文件中) |
| `schema_version` | - | ✓ **必需** (必须是 "3.0") |

#### 6.3.3 旧格式检测策略

加载 team.json 时执行严格格式检测：

1. **schema_version 检测**：必须存在且为 "3.0"
2. **角色内联字段检测**：检测 `roles[]` 中的 `name`, `alias`, `alias_en`, `persona`, `traits`, `guidance`, `capabilities` 字段 — 任何一个存在就报错
3. **规则内联字段检测**：检测 `rules[]` 中的 `instruction`, `description`, `enforcement`, `type` 字段 — 任何一个存在就报错
4. **prompt_source 必需性检测**：每个 `roles[]` 必须有 `prompt_source` 字段
5. **source 必需性检测**：每个 `rules[]` 必须有 `source` 字段

检测失败时必须返回清晰的错误，指向正确的文档位置：

```
ERROR: team.json 使用了旧格式（角色/规则内联内容）
  → 受影响字段: roles[2].persona, roles[2].traits, rules[0].instruction
  → 修复步骤:
     1. 将内联的 persona/traits/guidance 等内容移入 assets/skill/v3/prompts/{role-id}.md
     2. 从 team.json 的 roles[] 中删除这些字段
     3. 为每个 roles[] 添加 "prompt_source": "assets/skill/v3/prompts/{role-id}.md"
     4. 为每个 rules[] 添加 "source": "assets/skill/v3/prompts/rules.md"
     5. 添加 "schema_version": "3.0" 到根节点
  → 参考文档: docs/ARCHITECTURE.md §6.3
```

---

## 7. 上下文检测（Context Detection）

`main.go:checkContext()` 实现了 workspace/project 两层检测：

- **Workspace Root**: 包含 `projects/` 目录结构的目录，是存放多个项目的地方
- **Project Root**: 包含 `.team/` 目录的目录，是实际开发的项目
- **规则**: `flow proc`/`flow task`/`flow skill` 等团队操作**必须在 project 下运行**，禁止在 workspace root 运行

检测流程：
```
cwd → IsWorkspaceRoot()? 
    ├─ YES + 需要 project 上下文 → ⛔ WORKSPACE ROOT DETECTED
    └─ NO → 正常执行（假设是 project 或无上下文）
```

---

## 8. 外部依赖（External Dependencies）

### 8.1 Go 依赖

| 库 | 用途 |
|----|------|
| `github.com/spf13/cobra` | CLI 命令框架 |
| `github.com/spf13/viper` | 配置管理 |

### 8.2 外部工具

| 工具 | 用途 | 必需 |
|------|------|------|
| `bd` CLI (beads) | 任务 ID 生成 + 任务管理 | 推荐，用于规范 ID 格式 |

---

## 9. 构建和测试（Build & Test）

```bash
# 构建
cd projects/team-flow
go build -o flow.exe ./cmd/flow/

# 运行测试
go test ./...                  # 所有测试
go test ./internal/bd/ -v      # beads 客户端测试
go test ./internal/flow/ -v    # flow 模型测试
go test ./internal/proc/ -v    # 流程引擎测试
go test ./internal/task/ -v    # 任务管理测试

# 验证端到端
./flow.exe task create --title "Test task" --type bug
```

---

## 10. 文档导航（Documentation Navigation）

| 你想知道什么 → 读哪个文档 |
|-------------------------|
| **如何设计新的 flow/节点？** → [DESIGN.md](DESIGN.md) |
| **项目有哪些共识和决策？** → [CONSENSUS.md](CONSENSUS.md) |
| **编码和测试标准是什么？** → [STANDARDS.md](STANDARDS.md) |
| **架构决策记录在哪里？** → 查看 `v3/docs/ADR-*.md` |
| **flow JSON 的字段有哪些？** → `internal/flow/types.go` |
| **如何修复 bug？** → 遵循 dev-flow.json 的 bi01 → bf02 → qua9 路径 |
