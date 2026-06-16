# team-flow v3 核心架构与设计

> 版本: v3 | 更新时间: 2026-06-08 | 维护: team-flow

---

## 1. 概述

team-flow v3 是一个**流程驱动的 AI 协作框架**（Process-Centric AI Collaboration Framework）。它通过预定义的 JSON 流程模板，将软件开发任务拆分为有序的节点链，每个节点绑定角色、规则、技能和门控条件，确保 AI 按照规范化的流程执行任务。

### 核心设计理念

| 原则 | 说明 |
|------|------|
| **流程即约束** | JSON 流程模板定义全部执行路径，运行时不可偏离 |
| **角色分离** | Triage(需求整理) → TechLead(技术评估) → Dev(实现) → QA(验证) |
| **门控驱动** | Gate 节点在关键节点验证条件，不通过则回退 |
| **条件路由** | Edge 上的条件表达式决定多分支路径选择 |
| **会话持久化** | 每次交互记录到 eventlog，支持上下文恢复 |
| **嵌入模板** | 流程模板嵌入二进制(assets/)，项目本地 `.team/flows/` 为运行时覆盖 |

---

## 2. 包结构

```
projects/team-flow/
├── cmd/flow/main.go              # CLI 入口 (flow task/proc/session...)
├── internal/
│   ├── flow/                     # 核心类型 & 流程解析
│   │   ├── types.go              # Flow, FlowNode, FlowEdge, Task 等结构体
│   │   ├── parser.go             # JSON → Flow 解析器
│   │   └── parser_test.go        # 条件规范化测试
│   ├── proc/                     # 流程执行引擎
│   │   ├── procrun.go            # ProcRun 核心: 节点解析/推进/条件路由
│   │   ├── gate_checker.go       # Gate 门控检查系统
│   │   ├── gate_handler.go       # Gate 条件提取
│   │   ├── runtime.go            # 运行时环境初始化
│   │   └── cmd.go                # proc CLI 命令定义
│   ├── condition/                # 条件表达式求值
│   │   └── condition.go          # Eval(key=value, OR, NOT, AND)
│   ├── task/                     # 任务管理
│   │   └── task.go               # 创建/更新/关闭/查询 (4位hex ID)
│   ├── eventlog/                 # 会话事件日志
│   │   └── eventlog.go           # 事件写入/上下文快照/会话恢复
│   ├── config/                   # 配置解析
│   ├── team/                     # 团队/角色定义
│   ├── idgen/                    # ID 生成 (RandHex)
│   └── trace/                    # 追踪日志
├── assets/
│   ├── orgs/dev-team/            # 团队模板
│   │   ├── flows/dev-flow.json   # 嵌入式开发流程模板
│   │   └── team.yaml             # 团队角色定义
│   └── schema/flow-schema.json   # JSON Schema 验证
└── docs/                         # 文档
    ├── v1/ v2/ v3/               # 按版本归档
    └── requirements/             # 需求文档
```

---

## 3. 核心数据结构

### 3.1 Flow（流程定义）

```go
type Flow struct {
    Version    string              // "v3"
    Metadata   FlowMetadata        // name, description, tags, domain, type
    Config     *FlowConfig         // task_type, auto_dispatch, parallel_limit
    Extends    string              // 继承的父流程
    Nodes      []FlowNode          // 节点列表
    Edges      []FlowEdge          // 边列表
    Variables  map[string]interface{} // 模板变量
}
```

### 3.2 FlowNode（流程节点）

```go
type FlowNode struct {
    ID          string          // 唯一标识 (sta0, tri3, ent7, fa01...)
    Type        NodeType        // start|phase|gate|branch|parallel|terminal|...
    Name        string          // 显示名称
    Description string          // 节点描述
    Entry       bool            // 是否为入口节点
    Config      json.RawMessage // 节点配置 (PhaseConfig|GateNodeConfig|...)
    Components  *NodeComponents // 角色/规则/工具/技能/提示词/Prompt
    Docs        []DocSpec       // 产出文档定义
    OnEnter     []Action        // 进入节点时执行的动作
    OnExit      []Action        // 退出节点时执行的动作
    OnError     *ErrorHandler   // 错误处理策略
}
```

### 3.3 FlowEdge（流程边）

```go
type FlowEdge struct {
    ID         string     // 边标识
    From       string     // 源节点ID
    To         string     // 目标节点ID
    Type       EdgeType   // sequential|conditional|subflow
    Conditions []string   // 条件表达式 ["task_type=feature", "gate.passed"]
}
```

**重要**: `Conditions` 是 `[]string` 类型，直接存储条件表达式字符串（如 `"task_type=feature"`），而非对象数组。这是与 JSON 模板格式一致的简化设计。

### 3.4 Task（任务定义）

```go
type Task struct {
    ID          string    // 4位随机十六进制 (如 "bbca")
    Title       string    // 任务标题
    Type        string    // feature|bug|hotfix|analysis|change|docs|release
    Status      string    // open|in_progress|blocked|closed
    Description string
    Parent      string    // 父任务ID
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 3.5 节点类型常量

| 类型 | 常量 | 说明 |
|------|------|------|
| `start` | NodeTypeStart | 会话起点，自动记录上下文 |
| `phase` | NodeTypePhase | 执行阶段，绑定角色和产出 |
| `gate` | NodeTypeGate | 门控节点，验证条件通过才推进 |
| `branch` | NodeTypeBranch | 分支节点 |
| `parallel` | NodeTypeParallel | 并行执行 |
| `subflow` | NodeTypeSubflow | 子流程 |
| `loop` | NodeTypeLoop | 循环 |
| `terminal` | NodeTypeTerminal | 终态节点 (success/failed/rejected) |

---

## 4. 流程执行生命周期

### 4.1 ProcRun 引擎

`ProcRunEngine.Run()` 是核心执行入口，每次调用完成一个节点的处理：

```
flow proc run [--task <id>] [<node-id>] [--new]
```

**执行流程**:

```
1. 加载 Session (eventlog)
   ├── --new → 创建新会话
   ├── Rescue模式 → 返回最新会话状态
   └── 继续已有会话

2. 解析 Flow (FlowResolver)
   ├── 优先: .team/flows/<name>.json (本地覆盖)
   └── 回退: assets/orgs/<team>/flows/<name>.json (嵌入模板)

3. 解析节点 (NodeResolver)
   ├── 显式 node-id → 直接定位
   └── 无 node-id → 从入口节点开始

4. 收集变量 (VarSubstitutor)
   └── 模板变量替换: {DOCS_INTERNAL}, {task_id}, {TEAM_PATH}...

5. 生成结果 (generateResult)
   ├── 构建 CurrentNode (角色/规则/工具/技能/文档)
   ├── 构建条件上下文 (task_type, gate.passed, status...)
   └── 计算 NextOptions (条件边过滤)

6. Gate 检查 (仅 gate 节点)
   ├── 提取 GateConditions → 替换变量
   ├── 执行 GateCheck (task_exists, type_matches, tests_pass...)
   └── 结果写入 GateCheckResults

7. 记录事件 (eventlog)
   ├── flow.started / flow.node / flow.ended
   ├── 更新 context.md (上下文快照)
   └── 记录 Analysis/Conclusion (AI 分析)
```

### 4.2 请求/响应结构

```go
type ProcRunRequest struct {
    FlowName    string   // 流程名 (dev-flow)
    NodeID      string   // 目标节点ID (可选)
    TaskID      string   // 任务ID
    ProjectRoot string   // 项目根目录
    RunGate     bool     // 是否执行门控检查
    NewSession  bool     // --new 标志
    Input       string   // 用户输入
    Analysis    string   // AI 分析
    Conclusion  string   // AI 结论
}

type ProcRunResult struct {
    Flow             FlowMeta           // 流程元信息
    Current          CurrentNode        // 当前节点详情
    NextOptions      []NextOption       // 可选的下一步节点
    StatusLine       string             // 状态行
    Task             *TaskInfo          // 关联任务
    GateCheckResults []GateCheckResult  // 门控检查结果
    PathValidation   *PathValidation    // 路径验证
    AnalysisSchema   *AnalysisSchema    // 分析输出契约
}
```

---

## 5. 条件系统

### 5.1 条件表达式语法

`condition.Eval()` 支持以下表达式：

| 语法 | 示例 | 说明 |
|------|------|------|
| `key=value` | `task_type=feature` | 精确匹配上下文变量 |
| `!expr` | `!gate.passed` | 逻辑非 |
| `expr OR expr` | `status=task_created OR status=redirect` | 逻辑或 |
| 裸 key | `gate.passed` | key 存在且非空即为真 |

**注意**: 当前不支持 `AND` 运算符。如需复合条件，可在 `conditions` 数组中并列多条，所有条件都满足才匹配。

### 5.2 条件上下文

条件求值时可用的上下文变量：

| 变量 | 来源 | 说明 |
|------|------|------|
| `task_type` | `result.Task.Type` | 任务类型 (feature/bug/hotfix/...) |
| `status` | 会话状态 | task_created/no_task/redirect |
| `gate.passed` | Gate 检查结果 | 门控是否通过 |

### 5.3 边条件匹配

`buildNextOptionsFromEdges()` 遍历所有 outgoing 边：

1. `Conditions` 为空 → 无条件边，直接加入 matchingEdges
2. `Conditions` 非空 → 所有条件都满足才加入 matchingEdges
3. 第一条正条件边设为 `is_default=true`

```
ent7 → fa01  (conditions: ["task_type=feature"])    → 匹配 feature 任务
ent7 → bi01  (conditions: ["task_type=bug"])         → 匹配 bug 任务
ent7 → hi01  (conditions: ["task_type=hotfix"])      → 匹配 hotfix 任务
ent7 → rej8  (conditions: ["!gate.passed"])          → gate 未通过时回退
```

---

## 6. Gate 门控系统

### 6.1 Gate 检查类型

| 类型 | 常量 | 检查逻辑 |
|------|------|----------|
| `task_exists` | GateCondTaskExists | 检查 `.team/tasks/<id>.json` 是否存在 |
| `type_matches` | GateCondTypeMatches | 检查任务类型是否匹配预期 |
| `tests_pass` | GateCondTestsPass | 运行测试命令 |
| `lint_pass` | GateCondLintPass | 运行 lint 命令 |
| `deliverables_complete` | GateCondDeliverablesComplete | 检查产出文档是否存在 |
| `file_exists` | 自定义 | 检查指定文件是否存在 |
| `content_check` | 自定义 | 检查文件内容是否包含关键词 |

### 6.2 Gate 检查流程

```
GateChecker.CheckCondition(cond)
  ├── task_exists    → checkTaskExists()
  ├── type_matches   → checkTypeMatches()
  ├── tests_pass     → checkTestsPass()
  ├── lint_pass      → checkLintPass()
  └── custom         → checkCustomGate()
```

### 6.3 Gate 检查结果

```go
type GateCheckResult struct {
    Type     string  // 检查类型
    Passed   bool    // 是否通过
    Message  string  // 结果描述
    Required bool   // 是否必须通过
    Auto     bool   // 是否自动检查
}
```

---

## 7. 任务管理

### 7.1 任务生命周期

```
flow task create → 创建任务 (生成4位随机hex ID)
flow task update → 更新状态/类型
flow task show   → 查看任务详情
flow task close  → 关闭任务
```

### 7.2 存储格式

任务存储在 `.team/tasks/<id>.json`：

```json
{
  "id": "bbca",
  "title": "Feature test task",
  "type": "feature",
  "status": "open",
  "created_at": "2026-06-08T..."
}
```

### 7.3 ID 生成

使用 `crypto/rand` 生成 2 字节随机数 → 4 位十六进制字符串：

```go
func RandHex(n int) string {
    b := make([]byte, n)
    rand.Read(b)
    return hex.EncodeToString(b)
}
// RandHex(2) → "bbca", "3f2a", "9e01"...
```

---

## 8. 会话管理 (eventlog)

### 8.1 目录结构

```
.team/
├── sessions/
│   └── 2026-06-08-1430-a1b2c3/   # 会话目录 (YYYY-MM-DD-HHMM-6hex)
│       ├── context.md             # 上下文快照 (AI 恢复用)
│       └── trace.jsonl            # 操作追踪日志
├── state/
│   └── events.mdl                 # 全局事件日志 (JSONL)
└── tasks/
    └── <id>.json                  # 任务文件
```

### 8.2 事件类型

| 事件 | 说明 |
|------|------|
| `session.start` | 会话开始 |
| `session.analysis` | 状态分析 |
| `task.created` | 任务创建 |
| `task.status` | 任务状态变更 |
| `flow.started` | 流程开始 |
| `flow.node` | 节点进入 |
| `flow.ended` | 流程结束 |
| `error` | 错误事件 |

### 8.3 上下文恢复

当 AI 丢失上下文时，读取最新 session 的 `context.md` 即可恢复：

```
flow proc run          # 无参数 = Rescue模式
→ 返回最新会话状态 + StatusLine + 当前节点
```

---

## 9. 角色系统

### 9.1 角色定义

角色通过 `assets/orgs/<team>/team.yaml` 定义：

```yaml
roles:
  triage:
    name: 交付总监
    alias: 齐活林
    persona: 你是齐活林(Qi)，交付总监，团队的总调度...
    traits: [decisive, dispatch-only, classification-expert, never-execute]
    guidance: 收到任何输入，先分类再行动...

  tech-lead:
    name: 技术主管
    alias: 闻先迎
    ...

  dev:
    name: 开发工程师
    ...
```

### 9.2 角色映射 (dev-flow)

| 节点 | 角色 | 职责 |
|------|------|------|
| sta0 | Triage | 会话启动，任务创建 |
| tri3 | Triage | 需求整理，结构化分析 |
| assess | TechLead | 技术评估，判断任务类型 |
| fa01 | TechLead | 需求分析，产出 SPEC/AC |
| fd02 | TechLead | 技术设计，数据模型/API |
| fi03 | Dev | 代码实现 |
| bi01 | Dev | Bug 根因分析 |
| hi01 | Dev | 紧急修复 |
| qua9 | (Gate) | 质量门控 (测试/lint) |
| ver4 | QA | 集成验证 |
| rev6 | TechLead | 代码审查 |

---

## 10. 流程模板 (dev-flow.json)

### 10.1 完整流程拓扑

```
                              ┌─────────────────────────────────────┐
                              │           dev-flow 主流程            │
                              └─────────────────────────────────────┘

sta0 ──[status=task_created]──→ tri3 ──→ req-gate ──→ assess ──→ ent7
  │                              │          │                        │
  └──[status=no_task]──→ suc0    │    ┌─────┘(fail→tri3)            │
                                 │    │                              │
                          ┌──────┘    │         ┌────────────────────┤
                          ▼           │         ▼                    ▼
                    [TRIAGE.md]       │    [task_type=feature]  [task_type=bug]
                          │           │         │                    │
                          ▼           │         ▼                    ▼
                       完成           │       fa01                  bi01
                                      │         │                    │
                                      │         ▼                    ▼
                                      │       fd02                  bf02
                                      │         │                    │
                                      │         ▼                    │
                                      │       fi03 ──────────┐      │
                                      │         │             │      │
                                      │         │    ┌────────┘      │
                                      │         │    │               │
                                      ▼         ▼    ▼               ▼
                                    [task_type=hotfix]→ hi01 ──→ qua9
                                    [task_type=analysis]→ ai01 → ar02
                                    [task_type=change] → cp01 → ce02
                                    [!gate.passed]     → rej8
                                                          │
                              ┌────────────────────────────┤
                              │                            │
                              ▼                            ▼
                        [gate.passed]              [!gate.passed]
                              │                            │
                              ▼                ┌───────────┴──────────┐
                            ver4               │ 回退到对应实现节点     │
                              │                │ feature→fi03          │
                              ▼                │ bug→bf02              │
                            rev6               │ change→ce02           │
                              │                │ hotfix→tri3           │
                              ▼                └───────────────────────┘
                            suc0
```

### 10.2 关键节点说明

| 节点ID | 名称 | 类型 | 关键条件 |
|--------|------|------|----------|
| `sta0` | Session Start | start | 入口，创建会话 |
| `tri3` | Task Triage | phase | status=task_created |
| `req-gate` | Requirements Gate | gate | 检查 TRIAGE_REQUIREMENTS.md 存在 |
| `assess` | Technical Assessment | phase | Leader 判断 task_type |
| `ent7` | Entry Gate | gate | task_exists → 类型路由分支 |
| `fa01` | Requirements Analysis | phase | task_type=feature |
| `fd02` | Technical Design | phase | feature 线 |
| `fi03` | Implementation | phase | feature 线 |
| `bi01` | Root Cause Investigation | phase | task_type=bug |
| `bf02` | Bug Fix | phase | bug 线 |
| `hi01` | Emergency Fix | phase | task_type=hotfix |
| `ai01` | Deep Analysis | phase | task_type=analysis |
| `cp01` | Change Planning | phase | task_type=change |
| `qua9` | Quality Gate | gate | 测试/lint/回归 |
| `ver4` | Integration Verification | phase | gate.passed |
| `rev6` | Code Review | phase | 最终审查 |
| `suc0` | Completed | terminal | 成功终态 |
| `rej8` | Rejected | terminal | 入口拒绝 |
| `fai3` | Failed | terminal | 失败终态 |

### 10.3 模板变量

流程模板支持以下变量替换：

| 变量 | 说明 | 示例值 |
|------|------|--------|
| `{DOCS_INTERNAL}` | 内部文档目录 | `docs/internal` |
| `{task_id}` | 当前任务ID | `bbca` |
| `{TEAM_PATH}` | 团队模板路径 | `assets/orgs/dev-team` |
| `{workspace}` | 工作区根目录 | `D:\workspace\...` |
| `{user_input}` | 用户输入 | 原始用户消息 |
| `{task_type}` | 任务类型 | `feature` |

---

## 11. .team/ 目录布局

```
.team/
├── project.yaml              # 项目配置 (active_flow, team, docs_internal...)
├── project.md                # 项目配置 (Markdown 格式，已弃用)
├── flows/                    # 运行时流程覆盖 (git-ignored, 本地自动生成)
│   └── dev-flow.json         # 从嵌入模板加载的本地副本
├── tasks/                    # 任务存储
│   └── <id>.json             # 单个任务文件
├── sessions/                 # 会话目录
│   └── <timestamp>-<hex>/    # 单次会话
│       ├── context.md        # 上下文快照
│       └── trace.jsonl       # 追踪日志
├── state/
│   └── events.mdl            # 全局事件日志 (JSONL)
├── skills-resolved.yaml      # 技能解析缓存
└── .gitignore                # 防止 .team/ 误提交
```

**重要规则**:
- `.team/` 目录**禁止提交到 Git**
- `.team/flows/` 为自动生成，**禁止手动编辑**
- 修改流程应通过 `flow flow edit` 命令或直接修改嵌入模板

---

## 12. 关键设计决策

### 12.1 Conditions 类型: `[]string` vs `[]EdgeCondition`

**决策**: 使用 `[]string` 直接存储条件表达式，而非 `[]EdgeCondition{Expression: string}` 对象数组。

**原因**: 嵌入模板 JSON 使用 `"conditions": ["task_type=feature"]` 格式，`[]string` 与 JSON 格式完全一致，避免 JSON 解析失败导致条件路由静默失效。

**注意**: parser_test.go 中测试 JSON 也使用字符串数组格式。

### 12.2 Task ID: 4位随机十六进制

**决策**: 使用 `idgen.RandHex(2)` 生成 4 字符十六进制 ID。

**原因**: 短 ID 便于人工识别和输入，同时保持足够大的碰撞空间 (65536 种组合)。

### 12.3 嵌入模板优先

**决策**: 流程模板嵌入 `assets/orgs/<team>/flows/` 目录，运行时自动加载到 `.team/flows/`。

**原因**: 
- 确保每个项目有正确的流程模板
- 嵌入模板是版本控制的唯一真相来源
- `.team/flows/` 为运行时副本，git-ignored

### 12.4 Gate 条件 `task_exists` 不依赖 `bd` CLI

**决策**: `checkTaskExists` 优先检查 `.team/tasks/<id>.json` 本地文件，而非调用 `bd` CLI。

**原因**: 本地文件是任务管理的唯一存储，`bd` 是已弃用的外部工具。

### 12.5 `task_type` 上下文变量

**决策**: `buildConditionContext` 中 `task_type` 从 `result.Task.Type` 读取，而非 `result.Task.Status`。

**原因**: `Type` 字段存储 `feature/bug/hotfix` 等任务类型，`Status` 存储 `open/closed` 等状态。`ent7` 的条件路由依赖 `task_type=feature` 等表达式，必须读取正确的字段。

---

## 13. 命令参考

### 流程命令

```bash
# 启动流程
flow proc run --new                      # 新会话
flow proc run --task <id> <node-id>      # 指定节点
flow proc run                            # Rescue模式

# 查看流程
flow proc run --format json              # JSON 输出
flow proc run --task <id> ent7 --format text  # 文本输出

# Gate 检查
flow proc gate <flow> <node-id>          # 单独检查 gate
```

### 任务命令

```bash
flow task create --title "..." --type feature
flow task show <id> --json
flow task update <id> --type bug
flow task list
flow task close <id>
```

### 会话命令

```bash
flow session start --input "..."
flow session analysis --status task_created
flow session read requirements --for-ai --file TRIAGE_REQUIREMENTS.md
```

---

## 14. 扩展指南

### 添加新的任务类型分支

1. 在 `ent7` 的 edges 中添加新边：

```json
{
  "from": "ent7",
  "to": "new-node-id",
  "type": "conditional",
  "conditions": ["task_type=new_type"]
}
```

2. 添加对应的 phase 节点和后续边链。

### 添加新的 Gate 检查类型

1. 在 `gate_checker.go` 的 `CheckCondition()` 中添加 case
2. 实现对应的 `check*()` 方法
3. 在 `flow-schema.json` 中注册新类型

### 自定义角色

编辑 `assets/orgs/<team>/team.yaml`，添加角色定义后，在流程节点的 `config.role` 中引用。

---

## 附录 A: 条件表达式求值实现

```go
// condition/condition.go
func Eval(expr string, ctx map[string]string) bool {
    // 空表达式 → true
    if expr == "" { return true }
    
    // OR 运算
    if strings.Contains(expr, " OR ") {
        for _, p := range strings.Split(expr, " OR ") {
            if Eval(strings.TrimSpace(p), ctx) { return true }
        }
        return false
    }
    
    // NOT 运算
    if strings.HasPrefix(expr, "!") { return !Eval(expr[1:], ctx) }
    
    // key=value 匹配
    if idx := strings.Index(expr, "="); idx >= 0 {
        key := expr[:idx]
        val := expr[idx+1:]
        return ctx[key] == val
    }
    
    // 裸 key: 存在且非空
    got, ok := ctx[expr]
    return ok && got != "" && got != "false"
}
```

## 附录 B: 关键文件索引

| 文件 | 用途 | 关键函数 |
|------|------|----------|
| `internal/flow/types.go` | 核心类型定义 | Flow, FlowNode, FlowEdge, Task |
| `internal/flow/parser.go` | JSON 解析 | ParseFlow(), ApplyOverrides() |
| `internal/proc/procrun.go` | 流程引擎 | ProcRunEngine.Run(), buildNextOptions() |
| `internal/proc/gate_checker.go` | 门控检查 | GateChecker.CheckCondition(), checkTaskExists() |
| `internal/condition/condition.go` | 条件求值 | Eval() |
| `internal/task/task.go` | 任务管理 | createTask(), showTask(), updateTask() |
| `internal/eventlog/eventlog.go` | 会话日志 | UpdateContextSnapshot(), RecordNodeAnalysis() |
| `internal/idgen/idgen.go` | ID 生成 | RandHex() |
| `assets/orgs/dev-team/flows/dev-flow.json` | 嵌入流程模板 | 全部节点和边定义 |