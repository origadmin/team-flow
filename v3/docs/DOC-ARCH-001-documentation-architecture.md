# DOC-ARCH-001: v3 Documentation Architecture

> **Status**: Draft  
> **Date**: 2026-05-18  
> **Source**: User requirement — "文档整体混乱,导致后续规则开发制定出现严重割裂, AI完全不知道怎么改,改什么,生成的文档AI完全不清楚作用是什么,在哪里"  
> **Severity**: P0 — Blocks all downstream development

---

## 1. Problem Diagnosis

### 1.1 What AI Cannot Answer Today

When an AI agent (or human) wants to modify the team-flow v3 system, it cannot answer:

| Question | Why It Can't Answer |
|----------|-------------------|
| "修改Triage阶段的规则" → 去哪个文件？ | 规则分散在8个flow的nodes[].components.rules里，有的是ref引用，有的是内联定义（novel-flow） |
| "output-guard协议改了" → 影响哪些flow？ | 没有跨flow的规则索引，只能grep 8个JSON文件 |
| "新增一个role" → 需要改哪些文件？ | role枚举在schema里，引用在各flow的nodes[].config.role里，但没有映射文档 |
| "feature-flow的phase-analyze节点输出什么？" | 只能读feature-flow.json的nodes数组，找id=phase-analyze的节点，看docs字段 |
| "这个规则的完整定义在哪？" | ref=dispatch-guard, source=builtin → "builtin"是什么？在哪里？ |

### 1.2 Root Cause

**v3只有"数据层"（flow JSON），没有"文档层"（说明、索引、映射）。**

```
当前状态:
  flow-schema.json  → 定义了数据结构 ✅
  flows/*.json      → 定义了流程数据 ✅
  SKILL.md          → 入口说明 ✅ (但太简略)
  ---
  缺失:
  ❌ 节点资源索引 (ID → 含义/位置)
  ❌ 规则定义库 (ref → 完整规则内容)
  ❌ 角色定义库 (role → 能力/职责)
  ❌ 变量映射表 (variable → 来源/用途)
  ❌ 跨flow依赖图 (谁引用了谁)
  ❌ 修改影响分析 (改X会影响哪些flow/节点)
```

### 1.3 The ID-Resource Mapping Gap

用户明确指出: "ID获取对应节点资源,这个资源里面才会有phase-triage这种定义"

这意味着：
- **ID是随机黑盒** → 运行时通过ID查找节点
- **节点才是资源** → 节点内部才有语义信息（name, config.role, components, docs等）
- **但当前没有索引** → AI拿到一个ID，不知道去哪个flow找，找到后不知道这个节点做什么

## 2. Architecture: Three-Layer Documentation

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: INDEX (跨flow全局索引)                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │ROLE-     │ │RULE-     │ │NODE-     │ │FLOW-     │       │
│  │INDEX.md  │ │INDEX.md  │ │INDEX.md  │ │INDEX.md  │       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘       │
└─────────────────────────────────────────────────────────────┘
         │              │              │              │
         ▼              ▼              ▼              ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: REGISTRY (定义库, 单一来源)                         │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐              │
│  │ roles/     │ │ rules/     │ │ gates/     │              │
│  │  triage.md │ │  output-   │ │  quality-  │              │
│  │  dev.md    │ │  guard.md  │ │  check.md  │              │
│  │  qa.md     │ │  dispatch- │ │  entry-    │              │
│  │  ...       │ │  guard.md  │ │  gate.md   │              │
│  └────────────┘ └────────────┘ └────────────┘              │
└─────────────────────────────────────────────────────────────┘
         │              │              │
         ▼              ▼              ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: FLOW (流程定义, 引用Layer 2)                        │
│  ┌──────────────────────────────────────────────────┐       │
│  │ flows/dev-flow.json                               │       │
│  │   nodes[].components.rules → ref="output-guard"  │       │
│  │                          → source="registry"     │       │
│  │                          → 指向 rules/output-     │       │
│  │                             guard.md              │       │
│  └──────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

### 2.1 Layer 1: Flow Data (已存在)

位置: `{PROJECT}/v3/flows/*.json`

**职责**: 纯数据，定义流程拓扑和组件引用。

**修改原则**:
- 节点内部不定义规则内容，只引用 `ref`
- `source` 字段标识来源: `"registry"` (Layer 2定义库) / `"builtin"` (引擎内置) / `"file"` (外部文件)
- 语义信息放在 `name`/`description` 字段，不编码进ID

### 2.2 Layer 2: Registry (需创建)

位置: `{PROJECT}/v3/registry/`

**职责**: 每种可复用组件的唯一定义来源 (Single Source of Truth)。

```
v3/registry/
├── roles/                    # Role definitions
│   ├── triage.md            # What Triage does, capabilities, constraints
│   ├── dev.md               # What Dev does
│   ├── qa.md                # What QA does
│   ├── tech-lead.md         # What TechLead does
│   └── ...
├── rules/                    # Rule definitions
│   ├── output-guard.md      # Output Guard protocol (5-step self-check)
│   ├── dispatch-guard.md    # Triage dispatch rules
│   ├── regression-guard.md  # Regression prevention rules
│   ├── development-standards.md
│   ├── architecture-standards.md
│   ├── review-standards.md
│   └── ...
├── gates/                    # Gate definitions
│   ├── entry-gate.md        # Entry gate conditions
│   ├── quality-gate.md      # Quality gate conditions
│   └── ...
└── tools/                    # Tool definitions
    ├── task.md              # task tool (show, update, append)
    ├── search.md            # search tool
    └── ...
```

**每个registry文件的格式**:
```markdown
# {component-id}

> **Type**: role | rule | gate | tool  
> **Source**: registry  
> **Used In**: flow1, flow2, ... (auto-generated from Layer 3 index)

## Definition
{完整定义内容}

## Constraints
{约束条件}

## Examples
{使用示例}
```

### 2.3 Layer 3: Index (需创建)

位置: `{PROJECT}/v3/index/`

**职责**: 跨flow全局索引，回答"X在哪？"和"改X影响什么？"

```
v3/index/
├── ROLE-INDEX.md      # role → 哪些flow/节点使用
├── RULE-INDEX.md      # rule ref → 哪些flow/节点引用
├── NODE-INDEX.md      # node ID → 属于哪个flow + 语义摘要
├── FLOW-INDEX.md      # flow → 包含哪些节点/角色/规则/门控
├── VARIABLE-INDEX.md  # variable → 哪些flow使用 + 来源
└── CHANGELOG.md       # 所有修改的追溯记录
```

**NODE-INDEX.md 示例**:
```markdown
# Node Index

> Auto-generated from flows/*.json. Do not edit manually.

| ID | Flow | Type | Name | Role | Rules | Docs |
|----|------|------|------|------|-------|------|
| tri3 | dev-flow | phase | Task Triage | Triage | dispatch-guard | TRIAGE.md |
| ent7 | dev-flow | gate | Entry Gate | - | - | - |
| fa01 | dev-flow | phase | Requirements Analysis | TechLead | output-guard, requirements-standards | SPEC.md, AC.md |
| ... | ... | ... | ... | ... | ... | ... |
```

**RULE-INDEX.md 示例**:
```markdown
# Rule Index

> Auto-generated from flows/*.json + registry/rules/.

| Rule Ref | Registry Definition | Used In (flow:node) | Enforcement |
|----------|-------------------|--------------------:|-------------|
| output-guard | rules/output-guard.md | dev-flow:fa01, dev-flow:bi01, dev-flow:ai01, feature-flow:phase-analyze | hard |
| dispatch-guard | rules/dispatch-guard.md | dev-flow:tri3, feature-flow:phase-triage, bugfix-flow:phase-triage | hard |
| ... | ... | ... | ... |
```

## 3. The "builtin" Problem

当前flow中大量使用 `"source": "builtin"`，但没有任何文档说明builtin是什么。

### 3.1 Current "builtin" Usage

| ref | source | 实际含义 |
|-----|--------|---------|
| triage | builtin | ??? 引擎内置? v2 prompt? 哪个文件? |
| dispatch-guard | builtin | ??? 同上 |
| task | builtin | ??? |
| search | builtin | ??? |

### 3.2 Resolution

**方案**: `"builtin"` 表示引擎内置组件，但必须在 `registry/` 中有对应的映射文档说明其来源。

```markdown
# registry/roles/triage.md

> **Type**: role  
> **Source**: builtin (flow engine built-in)  
> **Mapped From**: team/v2/prompts/triage.md (v2 prompt) → v3 role: Triage

## Definition
The Triage role is responsible for classifying incoming tasks...

## v2 Migration Note
This role was defined in `team/v2/prompts/triage.md` in v2.
In v3, it's referenced as `builtin:triage` in flow node components.
```

**source字段的精确语义**:

| source值 | 含义 | 定义位置 |
|----------|------|---------|
| `registry` | v3/registry/ 下的定义文件 | `v3/registry/{type}/{ref}.md` |
| `builtin` | 流程引擎内置 | `v3/registry/{type}/{ref}.md` (标注builtin + 来源映射) |
| `file` | 外部文件引用 | flow中 `path` 字段指定 |
| `custom` | 用户自定义 | `v3/registry/{type}/{ref}.md` (标注custom) |

## 4. The "ref" Resolution Problem

当前flow中 `components.rules[].ref` 指向什么，完全不清：

```json
// dev-flow node fa01
"rules": [
  { "ref": "output-guard", "source": "framework" },  // framework? 哪个文件?
  { "ref": "requirements-standards", "source": "builtin" }  // builtin? 哪个文件?
]
```

### 4.2 Resolution Rules

**ref解析优先级**:
1. `source=registry` → 查找 `v3/registry/rules/{ref}.md`
2. `source=builtin` → 查找 `v3/registry/rules/{ref}.md`（标注builtin映射）
3. `source=framework` → 同 `source=registry`（framework是registry的旧名，统一改为registry）
4. `source=file` → 查找 `path` 字段指定的文件路径

**关键修改**: 所有 `"source": "framework"` 和 `"source": "builtin"` 统一规范，并在registry中建立映射。

## 5. Change Traceability

### 5.1 Every Change Must Have

1. **需求来源** (用户说了什么 / 什么问题驱动)
2. **影响分析** (改了X会影响哪些flow/节点/规则)
3. **修改记录** (哪个文件, 改了什么, 什么时候改的)
4. **验证结果** (修改后是否通过validate)

### 5.2 CHANGELOG Format

```markdown
# CHANGELOG

## 2026-05-18

### CORRECTION-001: Node ID Naming Convention
- **Source**: User correction "ID是随机的,黑盒才是我们想要的"
- **Impact**: All 8 flows (ID rename), schema (pattern constraint)
- **Status**: Spec approved, execution pending
- **Files**:
  - v3/docs/CORRECTION-001-node-id-norm.md (created)
  - v3/flows/*.json (pending rename)
  - v3/schema/flow-schema.json (pending pattern update)
```

## 6. Implementation Priority

| Step | What | Why |
|------|------|-----|
| 1 | Create `v3/registry/` structure | 给所有ref提供解析目标 |
| 2 | Create `v3/index/` structure | 给AI提供查找入口 |
| 3 | Normalize `source` field in all flows | 统一ref解析规则 |
| 4 | Execute CORRECTION-001 (ID rename) | 在index建立前先改ID |
| 5 | Generate NODE-INDEX.md | ID改名后生成索引 |
| 6 | Generate RULE-INDEX.md | 规则注册后生成索引 |
| 7 | Generate ROLE-INDEX.md | 角色定义后生成索引 |
| 8 | Update SKILL.md with doc architecture | 让AI知道文档体系 |

## 7. Directory Structure (After Implementation)

```
team-flow/
├── v3/
│   ├── schema/
│   │   └── flow-schema.json           # JSON Schema (Layer 1)
│   ├── flows/
│   │   ├── dev-flow.json              # Flow definitions (Layer 1)
│   │   ├── feature-flow.json
│   │   └── ...
│   ├── registry/                       # Component definitions (Layer 2) ← NEW
│   │   ├── roles/
│   │   │   ├── triage.md
│   │   │   ├── dev.md
│   │   │   ├── qa.md
│   │   │   └── ...
│   │   ├── rules/
│   │   │   ├── output-guard.md
│   │   │   ├── dispatch-guard.md
│   │   │   └── ...
│   │   ├── gates/
│   │   │   ├── entry-gate.md
│   │   │   └── ...
│   │   └── tools/
│   │       ├── task.md
│   │       └── ...
│   ├── index/                          # Cross-flow indexes (Layer 3) ← NEW
│   │   ├── NODE-INDEX.md
│   │   ├── RULE-INDEX.md
│   │   ├── ROLE-INDEX.md
│   │   ├── FLOW-INDEX.md
│   │   ├── VARIABLE-INDEX.md
│   │   └── CHANGELOG.md
│   └── docs/                           # Corrections & specs ← EXISTS
│       └── CORRECTION-001-node-id-norm.md
├── team/
│   └── v3/
│       ├── SKILL.md
│       ├── BOUNDARY.md
│       └── skills/
│           ├── team-flow-v3-create/SKILL.md
│           └── team-flow-v3-exec/SKILL.md
└── ...
```

## 8. AI Workflow After This Architecture

**Before (当前)**:
```
AI: "修改Triage规则"
  → grep 8个JSON文件找Triage相关内容
  → 找到ref=dispatch-guard, source=builtin
  → builtin是什么??? 不知道
  → 随机修改或放弃
```

**After (架构完成后)**:
```
AI: "修改Triage规则"
  → 查 ROLE-INDEX.md → Triage在dev-flow:tri3, feature-flow:phase-triage, ...
  → 查 RULE-INDEX.md → Triage关联dispatch-guard规则
  → 打开 registry/rules/dispatch-guard.md → 完整规则定义
  → 修改规则 → 影响分析 → 所有引用该规则的flow/节点
  → 更新CHANGELOG.md
```
