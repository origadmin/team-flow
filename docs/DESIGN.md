# 设计原则 (Design Principles)

> **版本**: v3.0  
> **状态**: 核心文档  
> **目的**: 定义本项目的设计原则、模式和约定。任何新功能/修改都应遵循这些原则。  
> **最后更新**: 2026-06-09

---

## 1. 核心设计原则

### 1.1 分层架构（Layered Architecture）

遵循严格的分层依赖方向：

```
┌─────────────────────────────────┐
│  Presentation (cmd/)            │  ← 只做命令注册和 IO，无业务逻辑
├─────────────────────────────────┤
│  Engine (internal/proc/)        │  ← 流程引擎：驱动节点推进
├─────────────────────────────────┤
│  Service (internal/task/, etc.) │  ← 领域服务：任务/技能/项目
├─────────────────────────────────┤
│  Data Model (internal/flow/)    │  ← 纯数据：无副作用，可独立测试
├─────────────────────────────────┤
│  Infrastructure (internal/bd/,  │  ← 外部集成：beads, 文件系统
│  internal/config/, etc.)        │
└─────────────────────────────────┘
```

**依赖规则**：上层可以依赖下层，下层**绝不**依赖上层。

**验证方法**：检查 `go.mod` 的导入方向 — `flow` 包不应导入任何其他内部包。

### 1.2 纯函数优先（Pure-First）

- **核心业务逻辑使用纯函数**：输入 → 输出，无副作用，无隐藏依赖
- **副作用隔离在边缘**：文件 IO、外部进程调用只出现在最外层
- **纯函数必须有单元测试**

**示例**（`internal/bd/client.go` 中的 `parseBeadsID`）：
```go
// ❌ 坏: 混合副作用和逻辑
func CreateIssue(...) string {
    cmd := exec.Command(...)
    output, _ := cmd.Output()
    // 在这里做解析...
}

// ✅ 好: 副作用和纯逻辑分离
func RunQuiet(args ...string) (string, error) { ... }  // 有副作用

func parseBeadsID(output string) string { ... }          // 纯函数，可测试
```

### 1.3 防御性设计（Defensive Design）

外部系统的输出永远不可信任：

1. **容忍冗余**：外部 CLI 的 stdout 可能包含警告/空行/多余信息
2. **明确格式检测**：基于特征检测（如 beads ID 包含 `-`）而非假设特定格式
3. **清晰的失败信号**：解析失败返回 `""`（或错误），不返回部分匹配

### 1.4 单一职责（Single Responsibility）

每个包做一件事：

| 包 | 职责 | 不应做 |
|----|------|--------|
| `internal/flow` | Flow 数据模型 + 解析/序列化 | 不执行命令，不访问文件 |
| `internal/bd` | beads CLI 客户端封装 | 不处理任务 CRUD，不生成 flow |
| `internal/task` | 任务 CRUD + 生命周期 | 不直接调用 `exec.Command` |
| `internal/proc` | 流程节点推进引擎 | 不处理任务数据格式 |
| `internal/config` | 配置/路径解析 | 不执行业务逻辑 |

### 1.5 显式优于隐式（Explicit Over Implicit）

- 函数签名应清楚表达其依赖
- 需要外部工具时，在函数命名和注释中明确声明
- 避免全局状态（`var` 级别的可变状态）

---

## 2. 关键设计模式

### 2.1 外部 CLI 集成模式

当需要集成外部工具（如 beads CLI）时：

```
1. 薄封装层: Run(args...) / RunQuiet(args...)
   ├── 只负责启动进程 + 捕获输出
   └── 不解析输出内容
   
2. 纯解析函数: parseXXX(output string) ...
   ├── 只接受字符串输入
   ├── 不访问外部
   └── 完整的单元测试覆盖

3. 高层函数: CreateIssue(title, type, desc)
   ├── 调用 RunQuiet() 获取输出
   ├── 调用 parseBeadsID() 解析
   └── 返回结构化结果
```

**为什么这样设计？**
- 外部 CLI 的 stdout 格式可能随版本变化
- 纯解析函数可以模拟各种边缘情况进行测试
- 当 CLI 输出格式变化时，只需修改解析函数

### 2.2 随机 ID 与可追踪性

**flow 节点 ID**（如 `a26c80`）：随机短码，便于重建和替换  
**任务 ID**（如 `team-flow-abc`）：beads 格式，项目前缀 + 随机码

**设计要点**：
- Flow 节点 ID 不是语义化的 — 含义在 `name` 字段中
- 任务 ID 必须是 beads 格式（`project-xxx`）— 由 `bd.CreateIssue` 保证
- ID 的稳定性和可追踪性是核心要求

### 2.3 事件驱动的执行追踪

```
flow proc run → 节点进入 → eventlog.NodeEntered(...)
                                     ↓
                              事件日志文件
                                     ↓
                    可回溯: "当时在哪个节点执行了什么"
```

**为什么？**
- AI 辅助开发需要可追踪的执行历史
- 出问题时可以回溯到具体的节点和动作

---

### 2.4 配置文件只管理引用，不管理内容（装配 vs 内容分离）

**设计原则**：配置文件（如 team.json）是装配清单，只负责"谁引用谁"，不负责"被引用的东西是什么"。

**反模式示例（当前 team.json 的问题）**：

```json
// ❌ 错误：角色内容嵌入配置文件
"roles": [
  {
    "id": "dev",
    "name": "工程师",
    "alias": "寇豆码",
    "persona": "你是寇豆码(Kou)，工程师，...",  // 内容
    "traits": ["surgical-changes", "test-first"], // 内容
    "guidance": "动手前先读 SCOPE.md...",       // 内容
    "rules": ["d5f", "d3c"]                       // 内容
  }
]
```

问题：
1. 修改角色 persona/traits 需要编辑配置文件 — 导致"改一处碰全局"
2. 多个 org 使用相同角色时需要复制粘贴 — 内容漂移
3. 格式演进时需要同时更新所有 team.json — 遗漏风险

**正确模式**：

```json
// ✓ 正确：配置文件只管理引用
"roles": [
  {
    "id": "dev",
    "prompt_source": "assets/skill/v3/prompts/dev.md",
    "principal": false
  }
]

// 角色内容在 prompts/dev.md 中（单一源）
// role-id 是唯一标识符
// flow 节点通过 role-id 引用角色
// 运行时从 prompt_source 解析出 persona/traits/guidance
```

**关键要点**：

| 关注点 | 位置 | 职责 |
|--------|------|------|
| 角色装配（谁用谁） | team.json roles[] | id + prompt_source + principal |
| 角色内容（是什么） | prompts/{role-id}.md | persona, traits, guidance, constraints, standards |
| 角色引用（哪节点用） | flow.json nodes[].components.roles | ref: role-id, source: team |
| 规则装配 | team.json rules[] | id + source |
| 规则内容 | prompts/rules.md 或 roles/{role-id}.md | instruction, description, enforcement, type |

**格式检测设计**：

```
team.json 加载流程:
  1. 读取 JSON
  2. 检测 schema_version 字段 (必须存在，必须为 "3.0")
  3. 检测 roles[] 是否包含内联字段 (name, alias, persona, traits...)
     → 任何存在就报错，带有修复指引
  4. 检测 rules[] 是否包含内联字段 (instruction, description...)
     → 任何存在就报错，带有修复指引
  5. 检测每个 roles[] 必须有 prompt_source
  6. 检测每个 rules[] 必须有 source
  7. 校验通过，才进入后续逻辑
```

**为什么这是设计问题，而不是简单的代码问题**：

```
旧格式（v2/v1）：  team.json = 装配 + 内容
  → 修改 persona → 改 team.json → 需要重新生成/校验所有 team.json
  → 多 org 共享角色 → 复制粘贴 → 漂移
  
新格式（v3）：   team.json = 装配, prompts = 内容
  → 修改 persona → 改 prompts/dev.md → 所有引用该 prompt 的 team 自动更新
  → 多 org 共享角色 → 同一 file path → 无漂移
  → 格式演进 → 改解析代码，无需改数据
```

**严格性与可用性**：

- 旧格式不是"warning"，而是"error"。因为允许旧格式共存意味着：
  1. 两个格式同时被维护（人力翻倍）
  2. 新功能只能用新格式，旧功能用旧格式（开发者困惑）
  3. 文档需要说"这里有两种格式，第一种已废弃"

- 正确策略：**一次到位，旧格式明确报错**，带有清晰的迁移指引。

---

## 3. 代码约定

### 3.1 包命名

- 小写，不使用下划线或混合大小写
- 简短但不晦涩（`bd` 而非 `beadsClient`，因为这是 beads 的唯一封装）
- 避免与标准库冲突（`internal/config` 而非 `internal/configuration`）

### 3.2 错误处理

- **优先返回 error**：`(result, error)` 而非 panic
- **错误信息要具体**：包含上下文，如 `fmt.Errorf("parse beads ID: %w", err)`
- **日志和返回分开**：可以同时记录日志（log.Debug）并向上返回错误

### 3.3 文件格式

- Flow 定义：JSON（便于机器读写和 diff）
- 任务文件：JSON（便于工具处理）
- 项目配置：Markdown + YAML frontmatter（便于人类编辑）
- 共识文档：Markdown（便于版本控制和协作）

### 3.4 可测试性

- 每个公共函数应考虑："如何为它写单元测试？"
- 如果一个函数需要复杂的测试 setup，通常意味着职责过多
- 外部依赖应通过接口注入（或至少通过包级变量便于 mock）

---

## 4. 修改流程（Change Process）

当你需要修改代码时，遵循以下顺序：

1. **理解当前设计** → 阅读 `docs/ARCHITECTURE.md`（本文档） + 相关模块
2. **检查是否有适用的决策** → 阅读 `docs/CONSENSUS.md`
3. **检查编码标准** → 阅读 `docs/STANDARDS.md`
4. **实现修改** → 遵循设计原则
5. **更新文档** → 如果修改影响了架构/设计，更新对应的文档
6. **添加测试** → 纯函数必须有单元测试；有副作用的代码必须有集成测试
7. **验证** → `go test ./...` + 端到端测试

**⚠️ 重要**：如果你的修改违反了本文档中的原则，说明你需要**先更新文档**，解释为什么需要违反（以及新的原则是什么）。代码修改不会比设计文档变更更先发生。

---

## 5. 反模式（避免的事情）

### ❌ 混合副作用和逻辑

```go
// ❌ 不要在一个函数里既调用外部进程又解析输出
func CreateIssue(...) (string, error) {
    cmd := exec.Command(...)
    output, _ := cmd.Output()
    // 解析逻辑...
    return id, nil
}

// ✅ 分离：一个函数处理副作用，另一个处理纯逻辑
func RunQuiet(args ...string) (string, error) { ... }
func parseBeadsID(output string) string { ... }
func CreateIssue(...) (string, error) {
    output, err := RunQuiet(args...)
    if err != nil { return "", err }
    id := parseBeadsID(output)
    if id == "" { return "", fmt.Errorf("...") }
    return id, nil
}
```

### ❌ 隐式格式假设

```go
// ❌ 假设: "bd 输出永远只有一行"
id := strings.TrimSpace(output)

// ✅ 明确: 查找符合特定格式的行
for _, line := range strings.Split(output, "\n") {
    if looksLikeBeadsID(line) {
        return line, nil
    }
}
```

### ❌ God 包

```go
// ❌ 一个包做所有事情
internal/utils/  # 包含文件处理、字符串操作、ID 生成、配置...

// ✅ 按职责分包
internal/bd/      # beads 客户端
internal/config/  # 配置
internal/idgen/   # ID 生成
```

---

## 6. 设计变更记录

当设计原则需要变更时：

1. 在本节添加记录
2. 说明变更原因
3. 更新受影响的模块
4. 添加测试验证新设计

| 日期 | 变更 | 原因 |
|------|------|------|
| 2026-06-09 | 初始化文档 | 项目缺乏统一设计文档 |

---

## 7. 相关文档

- [ARCHITECTURE.md](ARCHITECTURE.md) — 系统架构总览
- [CONSENSUS.md](CONSENSUS.md) — 项目共识和已做的决策
- [STANDARDS.md](STANDARDS.md) — 编码和测试标准
- `internal/flow/types.go` — Flow 数据模型
