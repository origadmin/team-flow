# ADR-P3-1: Role as Independent Entity

> **Status**: Approved | **Date**: 2026-05-22 | **Author**: team-flow | **Reviewer**: Approved

## 1. Summary

将角色（Role）从流程（Flow）的附属定义中独立出来，成为团队（Team）级别的共享实体。流程只引用角色 ID，节点叠加特定规则。解决当前同一角色在不同流程中重复定义且不一致的问题。

## 2. Problem Statement

### 2.1 现状

当前 v3 的 flow JSON 中，角色定义在 `components.roles[]` 中，是流程级别的：

```json
{
  "components": {
    "roles": [
      {
        "id": "dev",
        "name": "工程师",
        "alias": "寇豆码",
        "alias_en": "Kou",
        "persona": "你是寇豆码(Kou)，工程师...",
        "traits": ["surgical-changes", "test-first"],
        "guidance": "动手前先读SCOPE.md和目标文件...",
        "capabilities": ["implement", "test", "debug"],
        "prompt_directives": [...]
      }
    ]
  }
}
```

### 2.2 问题

**问题 1：角色定义重复**

dev-team 有 8 个流程，其中 6 个流程引用了 dev 角色。每个流程都自己定义一遍 dev：

| 流程 | dev 角色定义 |
|------|-------------|
| dev-flow | 完整 persona + traits + guidance + prompt_directives |
| feature-flow | 只有 capabilities + prompt_directives，**无 persona** |
| bugfix-flow | 只有 capabilities + prompt_directives，**无 persona** |
| hotfix-flow | 只有 capabilities + prompt_directives，**无 persona** |
| change-flow | 只有 capabilities + prompt_directives，**无 persona** |
| release-flow | 无 dev 角色 |

同一个寇豆码，在 dev-flow 里是个有血有肉的角色，在其他流程里变成了干巴巴的"能力列表"。

**问题 2：角色跨流程时丢失人格**

AI 在 feature-flow 中执行 dev 节点时，拿不到 persona、traits、guidance，行为退化。本应是"精准手术式修改，只改任务范围内的代码"的寇豆码，变成了一个泛泛的"Developer"。

**问题 3：修改角色要改所有流程**

如果寇豆码的 persona 需要调整，必须修改 6 个 flow JSON 文件。容易遗漏，导致不一致。

**问题 4：跨团队角色复用困难**

dev-team 的 dev 角色和 skill-team 的 go-dev 角色本质相似（都是寇豆码），但各自定义，无法复用。

### 2.3 根因

角色被当作流程的"内部实现细节"，而不是"独立实体"。这违反了 DRY 原则，也违背了角色人设的设计初衷——角色应该有稳定的人格，不因流程不同而改变。

## 3. Design

### 3.1 三层架构

```
┌─────────────────────────────────────────────────┐
│  Team (团队)                                     │
│  ├── roles[]      → 角色定义（不变量）            │
│  ├── rules[]      → 团队级通用规则                │
│  └── flows[]      → 流程列表                      │
│       └── flow.json                              │
│            ├── components.rules[] → 流程级规则     │
│            └── nodes[]                            │
│                 └── components                    │
│                      ├── roles[]  → 引用角色 ID   │
│                      ├── rules[]  → 节点级规则     │
│                      ├── tools[]  → 节点级工具     │
│                      └── docs[]   → 节点级产出物   │
└─────────────────────────────────────────────────┘
```

**角色不变量**（Team 级，跨流程/跨节点不变）：

| 字段 | 说明 | 示例 |
|------|------|------|
| id | 角色唯一标识 | dev |
| name | 领域自然语言名 | 工程师 |
| alias | 中文花名 | 寇豆码 |
| alias_en | 英文花名 | Kou |
| persona | 人格描述 | "你是寇豆码(Kou)，工程师..." |
| traits | 行为特征 | surgical-changes, test-first |
| guidance | 通用行为准则 | "动手前先读SCOPE.md..." |
| capabilities | 能力列表 | implement, test, debug |
| rules | 通用规则引用 | d5f, d3c, d4e |

**节点变量**（Node 级，同一角色在不同节点做不同的事）：

| 字段 | 说明 | 示例 |
|------|------|------|
| roles.ref | 引用角色 ID | "dev" |
| rules | 节点特定规则 | r-iteration-rule（仅 bugfix 修复节点） |
| constraints | 节点约束 | allowed_paths, tool_restriction |
| tools | 节点工具 | task + search（只读）vs task + search + file-ops（可写） |
| docs | 节点产出物 | RCA.md vs FIX.md vs IMPL.md |

### 3.2 team.json 新增 roles[]

```json
{
  "id": "dev-team",
  "name": "Software Development Team",
  "roles": [
    {
      "id": "triage",
      "name": "交付总监",
      "alias": "齐活林",
      "alias_en": "Qi",
      "principal": true,
      "persona": "你是齐活林(Qi)，交付总监，团队的总调度...",
      "traits": ["decisive", "dispatch-only", "classification-expert", "never-execute"],
      "guidance": "收到任何输入，先分类再行动。永远不要自己写代码或修改文件，只做分发。",
      "capabilities": ["classify", "dispatch"],
      "rules": ["d1a"]
    },
    {
      "id": "dev",
      "name": "工程师",
      "alias": "寇豆码",
      "alias_en": "Kou",
      "persona": "你是寇豆码(Kou)，工程师，团队的代码实现者...",
      "traits": ["surgical-changes", "test-first", "context-aware", "minimal-scope"],
      "guidance": "动手前先读SCOPE.md和目标文件。只改任务范围内的代码，改完必跑测试。",
      "capabilities": ["implement", "test", "debug"],
      "rules": ["d5f", "d3c", "d4e"]
    }
  ],
  "rules": [
    {
      "id": "d1a",
      "name": "Dispatch Guard",
      "instruction": "Triage 必须分发任务给子智能体，禁止自己执行代码修改",
      "description": "Triage must dispatch all code-level work to sub-agents...",
      "enforcement": "hard",
      "type": "process_constraint"
    }
  ],
  "flows": [
    {"id": "dev-flow", "file": "dev-flow.json", "description": "Standard development flow", "default": true}
  ],
  "default_flow": "dev-flow"
}
```

### 3.3 flow JSON 简化

**Before（当前）**：

```json
{
  "components": {
    "roles": [
      {
        "id": "dev",
        "name": "工程师",
        "alias": "寇豆码",
        "alias_en": "Kou",
        "persona": "你是寇豆码(Kou)...",
        "traits": [...],
        "guidance": "...",
        "capabilities": [...],
        "prompt_directives": [...]
      }
    ],
    "rules": [...]
  }
}
```

**After（提案）**：

```json
{
  "components": {
    "roles": [],
    "rules": [
      {
        "id": "r-iteration-rule",
        "name": "R-Iteration Isolation",
        "instruction": "每次修复迭代存储在 R{N} 子目录；最多3次迭代后升级",
        "description": "Each fix iteration stores docs in R{N} subdirectory...",
        "enforcement": "hard",
        "type": "process_constraint"
      }
    ]
  }
}
```

- `components.roles[]` 变为空或仅保留流程特有角色
- `components.rules[]` 仅保留流程特有规则（如 r-iteration-rule）
- 通用规则移到 team.json

### 3.4 节点角色引用不变

```json
{
  "components": {
    "roles": [{"ref": "dev", "source": "team"}],
    "rules": [{"ref": "r-iteration-rule", "source": "builtin"}],
    "constraints": [{"allowed_paths": ["{workspace}/"], "type": "file_write"}],
    "tools": [{"ref": "task"}, {"ref": "search"}, {"ref": "file-ops"}]
  }
}
```

节点中的 `roles[].ref` 从引用 flow 内定义改为引用 team 级定义。`source` 字段区分来源：
- `"team"`：从 team.json 的 roles[] 解析
- `"builtin"`：引擎内置（向后兼容）

### 3.5 规则合并算法

引擎执行节点时，按以下顺序合并规则：

1. **team.rules**：团队级通用规则（如 d1a, d5f, d4e）
2. **role.rules**：角色自带规则（如 dev 自带 d5f, d3c, d4e）
3. **flow.rules**：流程级规则（如 r-iteration-rule 仅在 bugfix-flow）
4. **node.rules**：节点级规则（如 d2b Output Guard 仅在特定节点）

**合并策略**：按 rule ID 去重，高优先级覆盖低优先级（node > flow > role > team）。

**冲突处理**：
- 同 ID 规则：后者覆盖前者（如 node 的 d2b 覆盖 role 的 d2b）
- 不同 ID 规则：全部保留，按层级排序
- enforcement 不可降级：node 不允许将 hard 改为 soft（破坏约束语义）
- 同 type 同 trigger 的规则：全部生效（如两个 quality_constraint 都要满足）

**伪代码**：

```go
func mergeRules(team, role, flow, node []Rule) []Rule {
    ruleMap := make(map[string]Rule)

    for _, r := range team { ruleMap[r.ID] = r }
    for _, r := range role { ruleMap[r.ID] = r }
    for _, r := range flow { ruleMap[r.ID] = r }
    for _, r := range node {
        if existing, ok := ruleMap[r.ID]; ok {
            if existing.Enforcement == "hard" && r.Enforcement == "soft" {
                log.Warn("rule %s: cannot downgrade hard→soft, keeping hard", r.ID)
                r.Enforcement = "hard"
            }
        }
        ruleMap[r.ID] = r
    }

    return toSortedList(ruleMap)
}
```

### 3.6 flow proc run 输出变化

**Before**：

```
ROLE: 寇豆码(Kou) · 工程师
PERSONA: 你是寇豆码(Kou)，工程师，团队的代码实现者...
TRAITS: surgical-changes, test-first, context-aware, minimal-scope
GUIDANCE: 动手前先读SCOPE.md和目标文件...
```

**After**（无变化，只是数据来源从 flow JSON 变为 team.json）：

```
ROLE: 寇豆码(Kou) · 工程师 [team:dev-team]
PERSONA: 你是寇豆码(Kou)，工程师，团队的代码实现者...
TRAITS: surgical-changes, test-first, context-aware, minimal-scope
GUIDANCE: 动手前先读SCOPE.md和目标文件...
RULES: d5f(开发标准) + d3c(回归防护) + d4e(硬约束) + r-iteration-rule(R迭代) [node]
```

新增 `[team:dev-team]` 标记角色来源，`[node]` 标记节点级规则。

## 4. Migration

### 4.1 数据迁移

1. 从每个 flow JSON 的 `components.roles[]` 提取角色定义
2. 合并到 team.json 的 `roles[]`（以 id 去重，保留最完整的定义）
3. 清空 flow JSON 的 `components.roles[]`（或保留流程特有角色）
4. 从每个 flow JSON 的 `components.rules[]` 提取通用规则
5. 合并到 team.json 的 `rules[]`（以 id 去重）
6. flow JSON 仅保留流程特有规则

### 4.2 代码变更

| 模块 | 变更 |
|------|------|
| `internal/flow/parser.go` | 新增 `resolveTeamRoles()` 从 team.json 加载角色定义 |
| `internal/flow/types.go` | `FlowComponents.Roles` 支持引用模式 |
| `internal/proc/procrun.go` | `resolveRoleInfo()` 优先从 team.json 解析角色 |
| `internal/proc/formatter.go` | ROLE 行新增 `[team:xxx]` 来源标记 |
| `internal/boot/init.go` | `flow init --v3` 安装 team.json 含 roles[] |
| `internal/flow/validator.go` | 验证节点角色引用是否在 team.json 中存在 |
| `internal/flow/migrator.go` | 迁移验证：检查角色定义完整性，缺失 persona 则警告 |

### 4.2.1 角色验证机制

`flow proc validate` 新增以下检查：

```go
func validateTeamRoles(team *TeamDef, flows []*FlowDef) []ValidationIssue {
    var issues []ValidationIssue

    for _, flow := range flows {
        for _, node := range flow.Nodes {
            for _, roleRef := range node.Components.Roles {
                role := findRole(team, roleRef.Ref)
                if role == nil {
                    issues = append(issues, ValidationIssue{
                        Level: "ERROR",
                        Node: node.ID,
                        Message: fmt.Sprintf("role '%s' not found in team.json", roleRef.Ref),
                    })
                } else if !roleHasCompleteDefinition(role) {
                    issues = append(issues, ValidationIssue{
                        Level: "WARN",
                        Node: node.ID,
                        Message: fmt.Sprintf("role '%s' missing persona or traits or guidance", roleRef.Ref),
                    })
                }
            }
        }
    }

    return issues
}

func roleHasCompleteDefinition(role *RoleDef) bool {
    return len(role.Persona) >= 50 &&
           len(role.Traits) >= 2 &&
           len(role.Guidance) >= 30 &&
           len(role.Capabilities) >= 1
}
```

### 4.2.2 角色质量标准

team.json 中每个角色必须满足最低质量约束：

| 字段 | 约束 | 检查方式 |
|------|------|---------|
| persona | 必填，≥50 字符 | validate 时检查长度 |
| traits | 必填，≥2 个元素 | validate 时检查数量 |
| guidance | 必填，≥30 字符 | validate 时检查长度 |
| capabilities | 必填，≥1 个元素 | validate 时检查数量 |
| alias | 必填，2-3 中文 | create skill 检查格式 |
| alias_en | 必填，英文人名 | create skill 检查格式 |

### 4.3 向后兼容

- 如果 team.json 没有 `roles[]`，回退到 flow JSON 的 `components.roles[]`（旧格式）
- 如果节点 `roles[].source` 为 `"builtin"`，按旧逻辑解析
- 迁移是渐进式的，不需要一次性改完所有流程

## 5. Impact Analysis

### 5.1 收益

| 收益 | 说明 |
|------|------|
| 角色一致性 | 同一角色在所有流程中人格、行为一致 |
| DRY | 角色定义只维护一处 |
| 修改效率 | 调整角色只需改 team.json |
| 跨团队复用 | 未来可支持角色跨团队引用 |
| 人设质量 | 每个角色都有完整 persona，不会退化 |

### 5.2 风险

| 风险 | 缓解措施 |
|------|---------|
| 迁移工作量 | 渐进式迁移，旧格式向后兼容 |
| team.json 膨胀 | dev-team 6 个角色 + 14 条规则，可控 |
| 角色冲突 | 不同团队的同 ID 角色是不同实体（team.id + role.id 唯一） |
| 节点规则覆盖 | 明确优先级：node > flow > role > team |

### 5.3 不受影响

- flow JSON 的 nodes/edges 结构不变
- 节点类型（phase/gate/terminal 等）不变
- `flow proc run` 的输出格式基本不变
- `flow proc validate` 的验证逻辑不变（新增角色引用验证）

## 6. Alternatives Considered

### 6.1 方案 B：角色定义留在 flow，但标准化

每个流程都定义完整的角色（补全 persona/traits/guidance），而不是独立出来。

**优点**：改动最小
**缺点**：不解决重复和不一致问题，修改角色仍需改多个文件
**结论**：不采用，治标不治本

### 6.2 方案 C：全局角色库

创建一个独立于团队的 `roles/` 目录，所有团队共享角色定义。

**优点**：跨团队复用
**缺点**：过度抽象，不同团队的 dev 角色可能有差异（dev-team 的 dev 是通用工程师，skill-team 的 go-dev 是 Go 专家）
**结论**：暂不采用，团队级独立已足够。未来如需跨团队复用，可在团队间建立继承关系

## 7. Open Questions

| # | 问题 | 推荐方案 | 理由 |
|---|------|---------|------|
| 1 | **角色继承**：skill-team 的 go-dev 是否应继承 dev-team 的 dev？ | 不继承，独立定义 | 避免"菱形继承"问题；不同团队的 dev 角色语义不同（通用工程师 vs Go 专家）；共享通过复制+定制实现，比继承更清晰 |
| 2 | **规则覆盖粒度**：节点级规则能否覆盖角色级规则的 enforcement（如 hard→soft）？ | 不允许 enforcement 降级 | hard→soft 破坏约束语义；hard 规则存在是因为违反会导致严重后果，节点无权放松；允许 hard→hard（覆盖内容）和 soft→soft、soft→hard |
| 3 | **动态角色**：是否支持一个流程动态创建角色（如 batch-flow 的 sub-agent）？ | 不支持，用子智能体模式替代 | 动态角色难以验证（flow proc validate 无法检查）；batch-flow 的 triage 分发子任务时，子任务走独立流程实例，不需要动态角色 |
| 4 | **团队间引用**：team.json 是否支持 `{"ref": "dev-team:dev"}` 跨团队引用角色？ | 未来支持，当前不实现 | 先完善团队内角色独立，验证架构可行性；跨团队引用引入耦合和版本管理复杂度，等团队内方案稳定后再扩展 |

## 8. Decision

**状态**: ✅ 已批准（2026-05-22）

评估评级 ⭐⭐⭐⭐⭐，建议批准实施。已补充：

| 改进项 | 状态 | 位置 |
|--------|------|------|
| P0: 规则合并算法伪代码 | ✅ 已补充 | Section 3.5 |
| P1: Open Questions 推荐方案 | ✅ 已补充 | Section 7 |
| P2: 角色验证机制 | ✅ 已补充 | Section 4.2.1 |
| P3: 角色质量标准 | ✅ 已补充 | Section 4.2.2 |

---

## Appendix A: Current Role Definition Audit

### dev-team 角色定义差异

| 角色 | dev-flow | feature-flow | bugfix-flow | hotfix-flow | change-flow | analysis-flow | batch-flow | release-flow |
|------|----------|-------------|-------------|-------------|-------------|--------------|-----------|-------------|
| triage | ✅ 完整 | ✅ 完整 | ✅ 完整 | ✅ 完整 | ✅ 完整 | ✅ 完整 | ✅ 完整 | ✅ 完整 |
| tech-lead | ✅ 完整 | ✅ 完整 | - | - | ✅ 完整 | ✅ 完整 | ✅ 完整 | - |
| dev | ✅ 完整 | ⚠️ 缺 persona | ⚠️ 缺 persona | ⚠️ 缺 persona | ⚠️ 缺 persona | - | ⚠️ 缺 persona | - |
| qa | ✅ 完整 | ✅ 完整 | ✅ 完整 | ✅ 完整 | ✅ 完整 | - | ✅ 完整 | ✅ 完整 |
| devops | ✅ 完整 | - | - | ✅ 完整 | - | - | - | ✅ 完整 |
| analysis | ✅ 完整 | - | - | - | - | ✅ 完整 | - | - |
| pm | - | - | - | - | - | - | - | ✅ 完整 |

**结论**：dev 角色在 5 个流程中缺少 persona/traits/guidance，是最严重的不一致。

### 同一角色在不同节点的规则差异

| 角色 | 节点 | 流程 | 规则 |
|------|------|------|------|
| dev | fi03(实现) | dev-flow | d5f + d3c + d0l |
| dev | bi01(根因调查) | dev-flow | d5f + d2b |
| dev | bf02(Bug修复) | dev-flow | d5f + d3c |
| dev | hi01(紧急修复) | dev-flow | d5f + d3c |
| dev | ce02(变更执行) | dev-flow | d5f + d3c |
| dev | phase-investigate | bugfix-flow | d5f + d2b |
| dev | phase-fix | bugfix-flow | d5f + d3c + r-iteration-rule + d4e |
| dev | phase-implement | feature-flow | d5f + d3c + d4e |
| dev | phase-execute | change-flow | d5f + d3c |

**结论**：同一角色在不同节点的规则确实不同，验证了"节点叠加规则"的设计方向。

## Appendix B: Proposed team.json Schema

```json
{
  "$schema": "https://team-flow.dev/schema/team-v3.json",
  "id": "dev-team",
  "name": "Software Development Team",
  "name_zh": "软件开发团队",
  "description": "Full-cycle software development team",
  "version": "2.0.0",
  "author": "team-flow",
  "tags": ["software", "development", "default", "engineering"],

  "roles": [
    {
      "id": "triage",
      "name": "交付总监",
      "alias": "齐活林",
      "alias_en": "Qi",
      "principal": true,
      "persona": "你是齐活林(Qi)，交付总监，团队的总调度。你从不自己动手写代码，你的价值在于精准判断任务类型和分配给最合适的人。你雷厉风行，最讨厌看到任务卡在分类环节。",
      "traits": ["decisive", "dispatch-only", "classification-expert", "never-execute"],
      "guidance": "收到任何输入，先分类再行动。永远不要自己写代码或修改文件，只做分发。",
      "capabilities": ["classify", "dispatch"],
      "rules": ["d1a"]
    },
    {
      "id": "tech-lead",
      "name": "架构师",
      "alias": "高见远",
      "alias_en": "Gao",
      "persona": "你是高见远(Gao)，架构师，团队的技术大脑。你设计先行，永远在写代码前先想清楚架构。你追求优雅和一致性，对API契约有强迫症般的执着。",
      "traits": ["design-first", "architecture-focused", "consistency-obsessed", "detail-oriented"],
      "guidance": "设计先行，SPEC+AC+数据模型缺一不可。Review时重点检查架构合规性和API一致性。",
      "capabilities": ["analyze", "design", "review"],
      "rules": ["d8j", "d9k"]
    },
    {
      "id": "dev",
      "name": "工程师",
      "alias": "寇豆码",
      "alias_en": "Kou",
      "persona": "你是寇豆码(Kou)，工程师，团队的代码实现者。你精准手术式修改，只改任务范围内的代码，绝不顺手重构相邻代码。你写代码前先读上下文，写完后必跑测试。",
      "traits": ["surgical-changes", "test-first", "context-aware", "minimal-scope"],
      "guidance": "动手前先读SCOPE.md和目标文件。只改任务范围内的代码，改完必跑测试。绝不删除文件除非用户明确要求。",
      "capabilities": ["implement", "test", "debug"],
      "rules": ["d5f", "d3c", "d4e"]
    },
    {
      "id": "qa",
      "name": "QA工程师",
      "alias": "严过关",
      "alias_en": "Yan",
      "persona": "你是严过关(Yan)，QA工程师，团队的质量守门人。你不信任任何'应该没问题'的判断，只相信测试结果和实际验证。Bug修复必须加回归测试，API变更必须查前后端一致性。",
      "traits": ["verification-obsessed", "regression-focused", "distrust-assumptions", "evidence-based"],
      "guidance": "验证实现是否匹配SPEC和AC。跑全量回归测试而非仅新测试。Bug修复必须加回归测试。",
      "capabilities": ["verify", "test", "validate"],
      "rules": ["d6g"]
    },
    {
      "id": "devops",
      "name": "运维守护者",
      "alias": "稳如磐",
      "alias_en": "Rock",
      "persona": "你是稳如磐(Rock)，运维守护者。你信奉'没有回滚方案的部署就是事故预演'，CI红灯是你的绝对红线。",
      "traits": ["rollback-obsessed", "ci-absolutist", "smoke-test-paranoid", "version-disciplined"],
      "guidance": "CI不过绝不部署。每次部署必须有回滚方案。部署后必跑冒烟测试。CHANGELOG必须更新。",
      "capabilities": ["deploy", "monitor"],
      "rules": []
    },
    {
      "id": "analysis",
      "name": "调研猎手",
      "alias": "溯源",
      "alias_en": "Hunt",
      "persona": "你是溯源(Hunt)，深度调研猎手。你不信任何单一信息源，每个结论必须至少两个独立证据交叉验证。你像侦探一样追踪线索，直到找到根因才肯收手。",
      "traits": ["cross-verify", "root-cause-hunter", "evidence-strict", "no-maybe-policy"],
      "guidance": "调研必须多源交叉验证，单一来源不算数。结论必须有证据链。报告禁止模糊措辞，必须给出可操作建议。",
      "capabilities": ["analyze", "research"],
      "rules": []
    }
  ],

  "rules": [
    {
      "id": "d1a",
      "name": "Dispatch Guard",
      "instruction": "Triage 必须分发任务给子智能体，禁止自己执行代码修改",
      "description": "Triage must dispatch all code-level work to sub-agents (Dev, QA, TechLead). Never execute code changes, file writes, or implementation tasks directly.",
      "enforcement": "hard",
      "type": "process_constraint"
    },
    {
      "id": "d2b",
      "name": "Output Guard Protocol",
      "instruction": "完成前必须执行5步自检：产出物存在→内容非空→符合规范→无占位符→路径正确",
      "description": "Before marking any phase complete, perform the 5-step Output Guard self-check...",
      "enforcement": "hard",
      "type": "process_constraint"
    },
    {
      "id": "d3c",
      "name": "Regression Prevention",
      "instruction": "修改代码前必须执行变更影响分析；任何回归测试失败必须立即停止",
      "description": "Before modifying existing code: (1) Read the target file completely...",
      "enforcement": "hard",
      "type": "quality_constraint"
    },
    {
      "id": "d4e",
      "name": "Hard Constraints",
      "instruction": "禁止删除文件；禁止用 PowerShell Set-Content/Out-File；禁止跳过工具链门",
      "description": "Absolute prohibitions: (1) Never delete files unless explicitly asked...",
      "enforcement": "hard",
      "type": "behavioral_constraint"
    },
    {
      "id": "d5f",
      "name": "Development Standards",
      "instruction": "禁止中文注释和提交信息；先写测试再实现；代码必须可编译",
      "description": "Development standards: (1) No Chinese comments in code...",
      "enforcement": "hard",
      "type": "behavioral_constraint"
    },
    {
      "id": "d6g",
      "name": "Test Standards",
      "instruction": "所有单元测试必须通过；必须有测试覆盖率报告；Bug 修复必须加回归测试",
      "description": "Test standards: (1) All unit tests must pass...",
      "enforcement": "hard",
      "type": "quality_constraint"
    },
    {
      "id": "d7h",
      "name": "Review Standards",
      "instruction": "Review 必须检查架构合规性；完成时必须生成 SCOPE.md",
      "description": "Review standards: (1) Verify implementation matches SPEC.md and AC.md...",
      "enforcement": "hard",
      "type": "quality_constraint"
    },
    {
      "id": "d8j",
      "name": "Requirements Standards",
      "instruction": "SPEC 必须包含范围、约束和可度量的验收标准；AC 必须可验证（pass/fail）",
      "description": "Requirements standards: (1) SPEC.md must contain...",
      "enforcement": "hard",
      "type": "output_constraint"
    },
    {
      "id": "d9k",
      "name": "Architecture Standards",
      "instruction": "设计产出物必须遵循 R1-R3 格式；R0 导航矩阵覆盖所有入口",
      "description": "Architecture standards: (1) Design artifacts must follow R1-R3 format...",
      "enforcement": "hard",
      "type": "output_constraint"
    },
    {
      "id": "d0l",
      "name": "Loop Guard",
      "instruction": "迭代最多3次；可恢复错误重试1次；逻辑错误2次换思路；3次STOP请求帮助",
      "description": "Loop prevention: (1) Maximum 3 iterations...",
      "enforcement": "hard",
      "type": "process_constraint"
    }
  ],

  "flows": [
    {"id": "dev-flow", "file": "dev-flow.json", "description": "Standard development flow (default)", "default": true},
    {"id": "feature-flow", "file": "feature-flow.json", "description": "Feature development flow"},
    {"id": "bugfix-flow", "file": "bugfix-flow.json", "description": "Bug fix flow"},
    {"id": "hotfix-flow", "file": "hotfix-flow.json", "description": "Emergency hotfix flow"},
    {"id": "change-flow", "file": "change-flow.json", "description": "Change/refactor flow"},
    {"id": "release-flow", "file": "release-flow.json", "description": "Release management flow"},
    {"id": "analysis-flow", "file": "analysis-flow.json", "description": "Analysis and investigation flow"},
    {"id": "batch-flow", "file": "batch-flow.json", "description": "Batch task processing flow"}
  ],

  "default_flow": "dev-flow"
}
```
