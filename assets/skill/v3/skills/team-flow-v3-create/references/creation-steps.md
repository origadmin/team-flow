# Creation Steps Reference (Steps 1-10)

> Referenced by: team-flow-v3-create SKILL.md
> Load this file after Step 0 (Requirement Discovery) is complete and user has confirmed team structure.

## Step 1: Load Team Definition

Take the confirmed team structure from Step 0 and formalize it into structured data:

1. **Team Name & Domain**: From Step 0 confirmation
2. **Members → Roles**: Each confirmed member becomes a role with persona, traits, guidance, and capabilities
3. **Collaboration Pattern**: How do members work together? What are the phases/stages?
4. **Output Goal**: What does this team produce?
5. **Quality Standards**: What are the checkpoints/gates in the workflow?

**If the user provided a reference team definition** (e.g., an existing team.md showing what others have done), use it as a quality benchmark — your output should reach the same level of completeness and professionalism. Do NOT copy it; use it to calibrate your own design.

## Step 2: Check Existing Flows

```bash
flow proc list
```

If a similar flow already exists, inform the user rather than creating a duplicate:
```bash
flow proc show {existing-flow}
```

## Step 3: Generate Flow Structure

Design the node graph based on the scenario:
1. List all phases with their roles (from Step 1)
2. Identify gate/decision points between phases
3. Determine edge connections and conditions
4. Add terminal nodes (success + failure paths)
5. Generate random shortcode IDs for all nodes

## Step 4: Generate Role Definitions

For each team member from the scenario, create a role definition:

```json
{
  "id": "r4k2",
  "name": "Product Manager",
  "description": "Defines product requirements, prioritizes features, and ensures alignment with business goals",
  "capabilities": ["requirements-analysis", "prioritization", "stakeholder-management"]
}
```

**Role ID format**: Random shortcode (same as node IDs). NOT semantic.

## Step 5: Generate Rule Definitions (CRITICAL)

Rules are derived from the team's domain and collaboration pattern. The team definition already implies what rules are needed — the AI's job is to make them explicit and enforceable.

**Rule sources (in priority order)**:
1. **Team definition itself**: If the team description mentions quality standards, compliance, or processes, encode those as rules first
2. **Domain conventions**: Each domain has known best practices (see domain table below)
3. **Universal rules**: Every flow needs Output Guard, Phase Transition, No Hallucination, Loop Guard

### Rule Generation Framework

Every flow MUST have these **universal rules** (adapt instruction to domain):

| Rule | ID Pattern | Instruction Template |
|------|-----------|---------------------|
| Output Guard | `{prefix}1` | "每个阶段必须产出所有 required 文档后才能进入下一阶段" |
| Phase Transition Guard | `{prefix}2` | "严格按流程顺序推进，禁止跳过阶段；gate 必须诚实评估" |
| No Hallucination Guard | `{prefix}3` | "只引用已定义的{domain}元素，禁止编造未在文档中确立的设定" |
| Loop Guard | `{prefix}4` | "迭代最多3次；可恢复错误重试1次；逻辑错误2次换思路；3次STOP请求帮助" |

Then generate **domain-specific rules** based on the domain:

| Domain | Typical Rules |
|--------|--------------|
| Software Dev | Development Standards, Test Standards, Regression Guard, Hard Constraints, Review Standards, Architecture Standards, Requirements Standards |
| Writing/Content | Narrative Standards, Character Consistency, World Consistency, Editing Standards, Outline Adherence |
| Trading/Finance | Data Accuracy, Risk Management, Compliance Check, Signal Validation, Position Sizing |
| Research | Source Verification, Citation Standards, Methodology Rigor, Peer Review, Bias Detection |
| Game Design | Game Design Standards, Mechanic Consistency, Narrative Consistency, Balance Standards |
| Data Analysis | Data Quality, Statistical Rigor, Visualization Standards, Insight Validation |

Plus **Quality Gates (QG1-QG4)** for ALL flows:

| Rule | ID | Instruction |
|------|----|-------------|
| QG-1: Think Before Action | `qg1` | "行动前验证假设：我在基于什么假设？错了最坏情况？应先澄清什么？" |
| QG-2: Simplicity First | `qg2` | "能用更简单的方式解决吗？在加未要求的功能吗？过度设计了吗？" |
| QG-3: Precise Changes | `qg3` | "只在任务范围内操作；不优化或重构相邻内容；清理自己产生的孤立产物" |
| QG-4: Goal-Driven Execution | `qg4` | "执行前定义成功标准；每步有具体验证标准；不用模糊的'完成'描述" |

### Rule Definition Format

Every rule MUST have all three content fields:

```json
{
  "id": "d5f",
  "name": "Development Standards",
  "instruction": "禁止中文注释和提交信息；先写测试再实现；代码必须可编译",
  "description": "Development standards: (1) No Chinese comments in code — all comments must be in English. (2) No Chinese commit messages. (3) Follow TDD: write tests before implementation when possible. (4) Code must compile and pass lint before marking phase complete. (5) Use the project's established libraries and patterns — check neighboring files before introducing new dependencies. (6) Run self-test (build + test + fmt) before completion.",
  "type": "quality_constraint",
  "enforcement": "hard"
}
```

- `instruction`: 1-2 sentence Chinese directive for AI (shown in `flow proc run`)
- `description`: Full English specification (shown in `flow proc rule {id}`)
- `id`: Random shortcode, NOT semantic

### Rule Assignment to Nodes

Assign rules to each phase node based on the node's purpose:
- **Every phase**: Output Guard + No Hallucination Guard + relevant QG
- **Creation phases**: Domain-specific quality rules (e.g., Narrative Standards for writing)
- **Review/Edit phases**: Editing/Review standards
- **Gate nodes**: No rules (gates are checkpoints, not execution)
- **Terminal nodes**: No rules

## Step 6: Write the Flow JSON

**放置路径规则**：
- 框架预设团队 → `v3/flows/{name}.json`（随工具安装，只读）
- 用户创建的团队 → `.team/flows/{name}.json`（项目专属）

**默认**：用户创建的团队放 `.team/flows/`。

Follow the structure and rules above. Key checklist:
- [ ] All node IDs are random shortcodes (`^[a-z][a-z0-9]{3,4}$`)
- [ ] All edge IDs follow `e-{from}-{to}` pattern
- [ ] `gate.config.on_pass`/`on_fail` reference valid node IDs
- [ ] At least one terminal node exists
- [ ] All edges reference existing nodes
- [ ] Variable placeholders use correct names (`{DOCS_INTERNAL}`, `{task_id}`, etc.)
- [ ] All rule refs in nodes have corresponding definitions in flow-level `components.rules`
- [ ] All role refs in nodes have corresponding definitions in flow-level `components.roles`
- [ ] Every rule has `instruction` + `description` + `enforcement`
- [ ] Every role has `description` + `capabilities`
- [ ] Universal rules (Output Guard, Phase Transition, No Hallucination, Loop Guard) are present
- [ ] QG1-QG4 are present

## Step 7: Validate

```bash
flow proc validate .team/flows/{name}.json
```

Fix all errors. Address warnings (especially undefined rule/role refs — these mean the AI won't see rule content).

## Step 7.5: Register Flow

After validation passes, register the flow in the project registry (`.team/project.md` → `## Flows`):

```markdown
## Flows
- {name} (.team/flows/{name}.json) - {brief description}
```

Or append via command:
```bash
echo "- {name} (.team/flows/{name}.json) - {brief description}" >> .team/project.md
```

## Step 8: Test Run

```bash
flow proc run --flow {name}
```

Verify the engine produces valid structured output for the first node. Check that rules show instruction + `→ flow proc rule {id}`.

## Step 9: Walk Through All Nodes

```bash
flow proc run --flow {name} {node-id}
```

For each node, verify:
- Rules are appropriate for the phase
- Role is correctly assigned
- Docs/deliverables are defined
- Next options are correct
- Gate conditions are sensible

## Step 10: User Review

Present the flow to the user for review. The user can:
- Use the Editor to visualize and tweak the flow
- Request changes to rules, roles, or structure
- Add domain-specific rules we missed

After user feedback, iterate from Step 6.

## Common Patterns

### Linear Flow (simple pipeline)

```
Phase A → Phase B → Phase C → Terminal
```

### Branching Flow (by task type)

```
Triage → Gate → [Feature path | Bug path | Hotfix path] → Quality Gate → Verify → Terminal
```

Use `gate` node with conditional edges, or `branch` node.

### Parallel Verification

```
Implementation → Parallel[Lint, Unit Tests, Integration] → Review → Terminal
```

### Retry Loop

```
Implementation → Quality Gate → (fail) → Loop back to Implementation
```

Use gate `on_fail` pointing back, or a `loop` node.

### Subflow Delegation

```
Main Flow → Subflow[bugfix-flow] → Continue
```

Use `subflow` node with `flow_ref: "builtin:bugfix-flow"`.
