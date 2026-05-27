---
name: team-flow-v3-create
version: 3.2
description: |
  Create and refine v3 flow JSON definitions.
  Produces valid flow JSON that passes `flow proc validate` and is executable by the engine.
---

# team-flow-v3-create - Flow Creation Skill

## Purpose

Generate or modify v3 flow JSON files. The output must be a valid flow JSON that:
1. Passes `flow proc validate` with zero errors
2. Can be executed by `flow proc run` to produce structured AI instructions

## Critical Rules

### Rule 1: Node IDs are Random Shortcodes

Node IDs MUST be opaque random shortcodes. No semantic meaning, no prefixes, no patterns.

**Format**: `^[a-z][a-z0-9]{3,4}$` (4-5 lowercase alphanumeric chars, starting with a letter)

**Valid examples**: `tri3`, `fa01`, `a7x2`, `k9m`, `bp4w`
**Invalid examples**: `phase-triage`, `gate-entry`, `gc01`, `df-tri3`, `feature-start`

**Why**: Semantic IDs cause duplication across flows and distort intent. Meaning belongs in `name` and `description` fields, not in the ID.

**Generation algorithm**:
1. Generate random 4-5 char string matching the pattern
2. Check uniqueness within the flow (IDs only need to be unique within a single flow)
3. If collision, regenerate

### Rule 2: Edge IDs Follow Node References

**Format**: `^e-{from_node_id}-{to_node_id}$`

Example: `e-tri3-ent7`, `e-fa01-fd02`, `e-qua9-ver4`

### Rule 3: Always Validate Before Saving

Run `flow proc validate {path}` after any change. Zero errors required. Warnings should be resolved when possible.

### Rule 4: Don't Touch v2

Only work with `v3/flows/` and `team/v3/`. Never modify v1 or v2 files.

## File Locations

| Item | Path | 说明 |
|------|------|------|
| Preset flows | `v3/flows/{name}.json` | 框架预设，随工具安装，只读 |
| Project flows | `.team/flows/{name}.json` | 项目自定义，用户创建的放这里 |
| Flow registry | `.team/project.md` → `## Flows` | 注册表，记录所有可用团队 |
| Flow schema | `v3/schema/flow-schema.json` | 验证用 |
| Reference flow | `v3/flows/dev-flow.json` | 标准参考 |

**Resolution order**: `v3/flows/` → `.team/flows/` (v3 preset first, project overrides)

**放置规则**：
- 框架预设团队（dev-flow, feature-flow 等）→ `v3/flows/` — 随 `flow init` / `flow migrate v3` 安装
- 用户创建的团队 → `.team/flows/` — 项目专属，不污染框架预设
- 同名时 `.team/flows/` 优先，可以覆盖预设行为

## CLI Commands (flow proc)

```bash
flow proc list                                    # List available flows
flow proc show {flow-name}                        # Show flow structure
flow proc validate {path-or-name}                 # Validate (MUST pass with zero errors)
flow proc run [--flow {name}] [node-id] [--task {task-id}] [--format json|text]
flow proc create {name} --template {template-name}
```

## Flow JSON Structure

### Top-Level Fields

```json
{
  "version": "v3",
  "metadata": { "name": "my-flow", "description": "...", "tags": [...] },
  "config": { "task_type": "feature", "auto_dispatch": true, "parallel_limit": 3, "timeout_minutes": 480 },
  "variables": { "docs_path": "", "TEAM_PATH": "", "task_id": "", "SKILL_PATH": "", "PROJECT_PATH": "" },
  "components": { "roles": [...], "rules": [...], "tools": [...], "skills": [...], "gates": [...] },
  "nodes": [...],
  "edges": [...]
}
```

### Required Fields

- `version`: always `"v3"`
- `metadata.name`: lowercase kebab-case identifier (`^[a-z][a-z0-9-]*$`)
- `nodes`: at least 1 node
- `edges`: connections between nodes
- At least one node must be `type: "terminal"`

### Variable Placeholders

| Variable | Description |
|----------|-------------|
| `{docs_path}` | Project documentation root |
| `{TEAM_PATH}` | Path to `.team` directory |
| `{task_id}` | Current task identifier |
| `{SKILL_PATH}` | Path to team-flow skill directory |
| `{PROJECT_PATH}` | Project root directory |

### Node Types (Quick Reference)

| Type | Purpose | Key Config |
|------|---------|------------|
| `phase` | Execution stage | `role`, `deliverables`, `docs` |
| `gate` | Conditional checkpoint | `conditions`, `on_pass`, `on_fail` |
| `branch` | Conditional routing | `conditions[].when`, `goto` |
| `parallel` | Concurrent execution | `branches[]`, `strategy` |
| `subflow` | Nested flow invocation | `flow_ref`, `input_mapping` |
| `loop` | Retry/iteration | `max_attempts`, `exit_condition` |
| `manual` | Human approval | `approvers[]`, `approval_strategy` |
| `event` | External trigger | `trigger`, `timeout_minutes` |
| `terminal` | End state | `status`, `message` |

> **Full node type examples**: See `references/node-types.md`

### Component References (Quick Reference)

| Component | Format | Key Fields |
|-----------|--------|------------|
| ComponentRef | `{ "ref": "...", "source": "..." }` | Source: `builtin/custom/marketplace/trae/file/framework` |
| ToolRef | `{ "ref": "...", "commands": [...] }` | Scope: `codebase/docs/config/all` |
| Constraint | `{ "type": "...", ... }` | Types: `file_write/tool_restriction/timeout/max_sub_agents` |

> **Full component examples + definition registry**: See `references/components.md`

## Verification Chain Architecture

> **Core principle**: AI cannot self-verify. Every role must verify upstream deliverables and produce a verification report.

### Three Verification Rules

When creating a flow, you MUST include these three rules in `components.rules`:

| Rule ID | Instruction | Enforcement | Description |
|---------|-------------|-------------|-------------|
| `v1` | 验证上游交付物再开始工作 | hard | Before starting work, verify upstream deliverables. If broken, report to principal. |
| `v2` | 完成工作后产出验证报告 | hard | After completing work, produce a verification report (what checked, result, evidence). |
| `v3` | 验证失败报告给主理人 | hard | On verification failure, report to principal with details. Do NOT silently retry. |

### Verification Report Doc

Every phase node that produces deliverables MUST also produce a `VERIFY_REPORT` doc:

```json
{
  "name": "VERIFY_REPORT",
  "format": "markdown",
  "path": "{DOCS_INTERNAL}/{task_id}/VERIFY_{node_id}.md",
  "description": "Self-check verification report",
  "required": true,
  "content_rules": [
    "Must contain: what was checked, result (pass/fail), evidence",
    "Must verify: upstream deliverables are complete and consistent",
    "Must verify: own deliverables pass all hard rules"
  ]
}
```

### Edge Pattern: Verified / !Verified

Every phase node MUST have two outgoing edges:

```json
{"from": "{node_id}", "to": "{next_node}", "condition": "verified"},
{"from": "{node_id}", "to": "{principal_node}", "condition": "!verified"}
```

- `verified`: Self-check passed, proceed to next node
- `!verified`: Self-check or upstream verification failed, return to principal

### Principal Handles Failures

The principal node (marked `principal: true`) is the only node that handles verification failures:

- On `!verified` edge → route to principal
- Principal decides: retry same node, adjust strategy, or reject
- Loop Guard: same task to same node ≤ 3 times, then STOP

### Gate Nodes vs Verification Chain

| Concept | Purpose | When to Use |
|---------|---------|-------------|
| Gate node (`type: "gate"`) | Automatic condition check | Pure technical validation (tests pass, lint pass) |
| Verification chain (v1/v2/v3 rules) | AI-driven quality assurance | Every phase node that produces deliverables |
| QA terminal node | Independent final verification | One per flow, before terminal |

**Gate nodes should NOT be used for routing** — routing is the principal's action, not a gate's job.

**Do NOT create separate gate nodes for dispatch/routing** — the principal node handles dispatch directly via conditional edges.

## Edge Configuration

### Sequential (default)

```json
{ "id": "e-fa01-fd02", "from": "fa01", "to": "fd02" }
```

### Conditional

```json
{
  "id": "e-ent7-fa01",
  "from": "ent7",
  "to": "fa01",
  "type": "conditional",
  "conditions": [{ "expression": "task_type=feature" }]
}
```

**Common expressions**:
- `gate.passed` / `!gate.passed` — gate result
- `task_type=feature` — task type matching
- `task_type=feature AND gate.passed` — combined

## Engine Runtime Behavior

When `flow proc run` executes, the engine:

1. **LoadFlow** — reads `v3/flows/{name}.json` or `.team/flows/{name}.json`
2. **FindNode** — finds node by ID (first phase/start node if no ID specified)
3. **CollectVars** — assembles variable map from project config + flow metadata
4. **BuildCurrentNode** — produces structured output:
   - For `phase`/`start`: role, rules (enriched from flow-level definitions), tools, skills, docs (with vars substituted), on_enter actions
   - For `gate`: gate conditions only
   - For `terminal`: status + message
5. **BuildNextOptions** — lists reachable next nodes with roles and conditions

The output is a JSON object that AI agents consume directly. **AI agents never read markdown docs — they get structured JSON from the engine.**

## Creation Workflow

### Step 0: Requirement Discovery (MANDATORY — Before Any Design)

**Never jump to design.** When a user requests a team, your first job is to help them clarify what they actually need — through conversation, not forms.

#### 0.1: Parse the User's Request

Extract what's explicitly stated and what's implied:

| Dimension | What to Look For | Example |
|-----------|-----------------|---------|
| **Domain** | What field/industry? | "内容分发" → content distribution |
| **Scope** | What's the boundary? | "全域" → multi-platform, but which ones? |
| **Starting Point** | Where does the user enter the flow? | Already has content? Starting from strategy? |
| **Pain Point** | What problem are they solving? | Can't distribute? Distributed but no effect? |
| **Output Goal** | What counts as "done"? | Published everywhere? Measurable results? |
| **Scale** | How big is the operation? | Solo creator? Team? Enterprise? |

#### 0.2: Identify Gaps and Propose Clarifications

Based on the parsed request, identify what's unclear. Then **propose options with reasoning**, not open-ended questions.

**❌ Wrong (form-filling)**:
> "请告诉我：1. 目标平台有哪些？2. 团队规模？3. 是否需要数据分析？"

**✅ Right (guided discovery)**:
> "全域内容分发——我理解核心是多平台适配和分发。帮你确认几个方向：
> 1. **起点**：你已有内容需要分发，还是从策略规划开始全链路？这决定了是否需要策略师角色。
> 2. **平台范围**：国内为主（微信/抖音/小红书/B站），还是国内+国际（+YouTube/Instagram/Twitter）？国内国际的合规和运营逻辑完全不同，通常需要拆成两个专家。
> 3. **闭环需求**：发完就结束，还是需要数据回收+迭代优化？如果需要闭环，要加分析师和优化师。"

#### 0.3: Propose a Draft Team Structure

After the user confirms, propose a team structure **before writing any JSON**:

```
基于你的确认，团队结构建议：
- 策万端（策略师）：制定分发策略和平台规划
- 晓红（国内平台专家）：微信/抖音/小红书/B站适配
- 格罗（国际平台专家）：YouTube/Instagram/Twitter适配
- 发如风（调度师）：排期和发布执行
- 数洞察（分析师）：跨平台数据归一化分析

关键设计决策：
1. 国内/国际拆成两个角色（运营域差异大）
2. 包含数据闭环（分析→优化→再分发）
3. 不含内容创作（你已有内容，只做分发）

这个方向对吗？有要调整的吗？
```

#### 0.4: Confirm Before Proceeding

**Do NOT proceed to Step 1 until the user confirms the team structure.** This is the evaluation checkpoint the user needs.

If the user says "调整一下" or "不对"，go back to 0.2 and refine.
If the user says "可以" or "就这样"，proceed to Step 1.

#### 0.5: Iterative Refinement (Re-enterable)

After the initial draft, the user may request additions, removals, or adjustments at any time — even after the flow JSON is generated. Handle each request by:

1. **Parse the delta**: What specifically changed? (add role, remove role, adjust role, add phase, etc.)
2. **Assess impact**: What does this change affect?
   - Adding a role → Does it need its own phase? Does it join an existing phase? What rules does it need? What docs does it produce?
   - Removing a role → Which phases reference it? What rules become orphaned?
   - Adjusting a role → Does the persona/traits/guidance change? Does it affect rule assignments?
3. **Propose the integration**: Show how the change fits into the existing structure, not just the delta in isolation
4. **Confirm and apply**: Get user confirmation, then update the flow JSON

**Example — User says "加个音乐总监"**:

```
收到，加入音乐总监。整合方案：

角色：韵律（音乐总监）
- 人设：你是韵律，音乐总监。你用声音定义情绪节奏——产品宣传片不是配乐，是音乐叙事。BPM决定观众心跳，音效设计让产品卖点"被听到"而不只是"被看到"。
- 在流程中的位置：视觉导演确认分镜后、制作执行前，音乐总监介入设计配乐方案
- 产出物：MUSIC_BRIEF.md（配乐方案：情绪曲线、BPM规划、音效设计、版权/生成方案）
- 影响的规则：新增"Audio-Visual Sync"规则（画面节奏必须与配乐BPM对齐）

调整后的流程：
创意策划 → 脚本编剧 → Gate → 视觉导演 → 音乐总监(新增) → 制作执行 → Gate → 品控审核 → 完成

这样对吗？
```

**Refinement Principles**:
1. **Show integration, not just addition** — Don't just list the new role; show where it fits in the flow, what it produces, and what it affects
2. **Ripple assessment** — Every role addition may require: new phase node, new edges, new rules, new docs, adjustments to existing phases
3. **Preserve existing structure** — Don't restructure the whole flow for one addition; insert cleanly
4. **Same persona quality** — New roles get the same level of persona/traits/guidance as existing ones

#### Discovery Principles

1. **Propose, don't ask** — Give the user something concrete to react to, not a blank form
2. **Explain your reasoning** — "国内国际拆成两个角色" because "运营域差异大", not just because
3. **Surface trade-offs** — "如果不需要数据闭环，可以省掉分析师和优化师两个角色"
4. **Respect the user's domain** — They know their business better than you. Your job is structure, not domain expertise
5. **One round, not ten** — Aim to get confirmation in 1-2 exchanges, not a long back-and-forth

#### Conventions

##### C1: Persona Naming Convention

Every role MUST have three name identifiers:

| Field | Format | Example | Purpose |
|-------|--------|---------|---------|
| `alias` | Chinese 2-3 character name | 韵律, 点石, 织线 | Display name, memorable identity |
| `alias_en` | English person name | Melody, Spark, Weaver | Cross-language reference |
| `name` | Domain-natural role title | 交付总监, 架构师, 音乐总监 | Functional identification in the team's language |

The `persona` field uses the format: `"你是{alias}({alias_en})，{name}。..."`

**Examples**:
```json
{
  "id": "music-director",
  "name": "Music Director",
  "alias": "韵律",
  "alias_en": "Melody",
  "persona": "你是韵律(Melody)，音乐总监。你用声音定义情绪节奏..."
}
```

**Naming rules**:
- Chinese alias: 2-3 chars, evocative of the role's essence (not generic titles)
- English alias: A real-sounding person name that reflects the role's character
- Avoid literal translations — 点石(Spark) not Stone, 织线(Weaver) not Thread

##### C2: Principal Role (Team Entry Point)

Every flow MUST designate exactly one role as the **principal** — the team's entry point and public face. This is equivalent to "主理人" in creative teams or "Triage" in dev teams.

```json
{
  "id": "creative-strategist",
  "name": "Creative Strategist",
  "alias": "点石",
  "alias_en": "Spark",
  "principal": true,
  "persona": "你是点石(Spark)，创意策划..."
}
```

**Rules**:
- Exactly one role per flow has `"principal": true`
- The principal role occupies the first phase node in the flow
- The principal's name varies by domain (主理人/Triage/首席/总监) — no forced uniformity
- The engine displays ⭐ next to principal roles in text output
- The principal is the role users interact with first

##### C3: Workflow Step Refinement

Users may specify concrete process steps at any time. These map to phase nodes in the flow.

**When a user says**: "流程是：创意简报、逐镜头分镜、素材生产、HyperFrames剪辑合成、BGM设计与交付"

**Map each step to a phase node**:
1. 创意简报 → Phase: Creative Brief (创意策划)
2. 逐镜头分镜 → Phase: Storyboard (脚本编剧 + 视觉导演)
3. 素材生产 → Phase: Asset Production (制作执行)
4. HyperFrames剪辑合成 → Phase: Editing & Compositing (制作执行)
5. BGM设计与交付 → Phase: Music Design & Delivery (音乐总监)

**Key**: Multiple user steps may map to the same role but different phases. Don't merge steps — each step the user mentions deserves its own node, even if the same role handles it.

**Also**: User-specified tools/techniques (like "HyperFrames") should be noted in the phase description and docs, not ignored as implementation details.

##### C4: Domain Tags

Every flow's `metadata.tags` should include the team's expertise areas, derived from Step 0 discovery and user input:

```json
{
  "metadata": {
    "tags": ["promo-video", "video-production", "product-launch"]
  }
}
```

Tags use lowercase kebab-case. Include both the team type and the output type.

### Steps 1-10: Implementation

After Step 0 is confirmed, follow the detailed implementation steps:

> **Full steps reference**: See `references/creation-steps.md`

| Step | Action | Key Output |
|------|--------|------------|
| 1 | Load Team Definition | Formalized roles, collaboration pattern, quality standards |
| 2 | Check Existing Flows | Avoid duplicates via `flow proc list` |
| 3 | Generate Flow Structure | Node graph with random shortcode IDs |
| 4 | Generate Role Definitions | Role JSON with capabilities |
| 5 | Generate Rule Definitions | Universal + domain-specific + QG rules |
| 6 | Write the Flow JSON | Complete flow file (`.team/flows/` for user-created) |
| 7 | Validate | `flow proc validate` — zero errors |
| 7.5 | Register Flow | Add to `.team/project.md` → `## Flows` |
| 8 | Test Run | `flow proc run --flow {name}` |
| 9 | Walk Through All Nodes | Verify each node's rules, role, docs, next options |
| 10 | User Review | Present for feedback, iterate from Step 6 |

**Step 5 is CRITICAL** — rules make or break flow quality. Every flow needs:
- Universal rules (Output Guard, Phase Transition, No Hallucination, Loop Guard)
- Verification chain rules (v1, v2, v3)
- Domain-specific rules
- Quality Gates (QG1-QG4)

## Anti-Patterns to Avoid

1. **Semantic IDs**: Never use `phase-triage`, `gate-entry`, `impl-node`. Use random shortcodes.
2. **Giant mega-flows**: Don't put all task types in one flow. Use separate flows with subflow nodes.
3. **Inline rule content**: Don't put rule text in node descriptions. Define rules at flow level and reference by `ref`.
4. **Missing terminal nodes**: Every path through the flow must eventually reach a terminal.
5. **Orphan nodes**: Every non-terminal node must be connected by at least one edge.
6. **Undefined refs**: Every `ref` in node components must have a corresponding definition in flow-level `components`.
7. **Gate nodes for routing**: Don't create gate nodes just to dispatch/route — that's the principal's job.
