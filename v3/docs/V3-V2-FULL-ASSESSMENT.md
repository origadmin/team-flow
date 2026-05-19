# v3 vs v2 完整覆盖性评估 (skills-creator 方法论)

**评估时间**: 2026-05-20 03:25 GMT+8
**评估基准**: v2 SKILL.md v2.4 + shared.md v8.0 + shared-protocol.md + shared-checklist.md + BOUNDARY.md + 14 prompts + 15 role standards + 6 skills + 20 templates
**v3 基准**: 11 flow JSONs + flow-schema.json + procrun.go + types.go + v3-create/exec SKILL.md

---

## 一、评估方法

按 skills-creator 6步评估流程中的 Step 5（验证）维度：
1. **功能完整性** — v2 每个功能点是否在 v3 有对应实现？
2. **等价性** — 对应实现是否达到 v2 同等效果？
3. **结构性保证** — v3 是否把 v2 的"AI自觉遵守"升级为"引擎强制"？

---

## 二、逐项覆盖分析

### A. 已完整覆盖（引擎结构性保证 > v2 的文档约束）

| # | v2 功能 | v3 实现 | 覆盖等级 |
|---|--------|---------|---------|
| A1 | 流程执行路径（Feature/Bug/Change/Analysis/Hotfix） | 5个独立 flow JSON + dev-flow | ✅ 完整 |
| A2 | 角色分配（Triage→TechLead→Dev→QA pipeline） | flow nodes 的 role 字段 + component registry | ✅ 完整 |
| A3 | 三层门禁（Entry/Phase/Completion） | gate 节点 + gate_conditions + on_pass/on_fail 路由 | ✅ 完整，结构性更强 |
| A4 | 状态追踪（beads labels: phase:xxx） | on_enter actions + TaskInfo(beads_id,phase,status,url) | 🟡 结构就绪，调用逻辑未集成 |
| A5 | Dispatch Guard（Triage 不直接执行） | rule: dispatch-guard(hard) + prompt_directives[0] | ✅ 双重保证 |
| A6 | 交付物路径规则 | DocSpec(name, path, format, required, description) | ✅ 完整 |
| A7 | 路径解析（flow config paths --json） | {TEAM_PATH}/{task_id} 等变量占位符 | 🟡 声明式，引擎不运行时解析 |
| A8 | 并发控制（batch ≤3 parallel） | batch-flow.json concurrency rules | ✅ 完整 |
| A9 | Bug R-iteration 追踪 | bugfix-flow branch-r-iteration 节点 + gate 重试 | ✅ 完整 |
| A10 | Status Line 强制输出 | buildStatusLine() 引擎自动生成 | ✅ 比v2更强（v2靠AI自觉） |
| A11 | 角色切换（Triage→sub-agent→Triage） | next_options + role 字段 + edge 条件路由 | ✅ 完整 |
| A12 | 回归防护规则 | rule: regression-guard(hard) + prompt_directives | ✅ 双重保证 |
| A13 | 硬约束（禁删文件/禁BOM/禁跳toolchain gate） | rule: hard-constraints(hard) | ✅ 完整 |
| A14 | Output Guard 5步协议 | gate-completion 条件 + rule: output-guard | ✅ 完整 |
| A15 | 会话交接文档 | terminal-success on_exit: create-handoff-doc | ✅ 完整 |
| A16 | 资产包规格 | DocSpec 路径模板 + format + required | ✅ 完整 |
| A17 | 批量任务流程 | batch-flow.json (独立flow) | ✅ v3新增 |
| A18 | 发布流程 | release-flow.json (独立flow) | ✅ v3新增 |
| A19 | Skill 引用 | component registry skills[] + trigger/description/path | ✅ 完整 |
| A20 | Prompt 加载路径 | prompt_source + standards_source + prompts[] | ✅ 完整 |
| A21 | prompt_directives 内联 | RoleDefinition.prompt_directives[] | ✅ v3新增，v2无 |
| A22 | 完成门禁 deliverables 检查 | gate-completion: deliverables_complete | ✅ 完整 |

### B. 部分覆盖（有框架但缺细节或执行逻辑）

| # | v2 功能 | v3 现状 | 缺口 | 优先级 |
|---|--------|---------|------|--------|
| B1 | **beads 集成**（flow task CLI 全套命令） | TaskInfo 结构体已定义，resolveTaskInfo() 存在 | 引擎 Run 方法未实现 on_enter/on_exit 自动调用 flow task CLI | 🔴 P0 |
| B2 | **完成门禁实际检查**（40+ checklist 项） | gate_conditions 声明式（deliverables_complete, tests_pass） | 引擎不执行实际文件检查/测试运行；checklist 内容未结构化 | 🔴 P0 |
| B3 | **Skill 自动路由**（6 sub-skill trigger patterns） | skills[] 有 trigger/description | 仅 manual trigger，无 phase:implement→auto-load 模式 | 🟡 P2 |
| B4 | **ID 命名规则执行**（^[a-z][a-z0-9]{3,4}$） | CORRECTION-001 文档定义规则 | flow proc validate 不检查节点 ID 格式 | 🟡 P2 |
| B5 | **Prompt 文件内容加载** | prompt_source 路径 + prompt_directives 摘要 | 引擎不读文件内容，AI 需自行按路径加载 | 🟡 P1 |
| B6 | **变量占位符运行时解析** | {TEAM_PATH}/{task_id} 在 JSON 中 | 引擎运行时不解析角色定义中的路径变量 | 🟡 P1 |

### C. 未覆盖（v2 有，v3 完全缺失）

| # | v2 功能 | 严重度 | 说明 | 优先级 |
|---|--------|--------|------|--------|
| C1 | **五层架构边界**（BOUNDARY.md L0-L4） | 🔴 高 | v2 有严格跨层访问矩阵，v3 无任何隔离机制。AI 可能写入框架层 | P0 |
| C2 | **质量门 QG-1~4**（Think Before Coding / Simplicity First / Surgical Changes / Goal-Driven） | 🟡 中 | v2 shared-checklist.md 定义4个质量门，v3 无对应机制 | P1 |
| C3 | **共识机制**（.team/consensus.md 读取/更新） | 🟡 低 | 已在 v2 项目层实现，v3 引擎不感知 | P2 |
| C4 | **Checklist 覆盖**（.team/checklist.md 强制检查） | 🟡 低 | v2 项目层实现，v3 引擎不感知 | P2 |
| C5 | **14 个角色 prompt 文件** | 🔴 高 | v2 有 14 个详细 prompt（triage/tech-lead/dev/dev-backend/dev-frontend/qa-engineer/pm/devops/bugfix/analysis/ui-designer/framework-architect/triage-clarify/triage-verify），v3 仅引用路径但不包含内容 | P1 |
| C6 | **15 个角色标准文件** | 🟡 中 | v2 workflows/roles/ 下 15 个标准文件，v3 prompt_source 指向路径但文件不在 v3 目录 | P1 |
| C7 | **6 个子技能完整内容** | 🟡 中 | v2 有 build/check-impl/design/evolve/git/review 6个完整SKILL.md，v3 仅 skill ref | P1 |
| C8 | **20 个模板文件** | 🟡 中 | v2 templates/ 下 20 个模板，v3 DocSpec.template 字段存在但未填充 | P1 |
| C9 | **Session 协议**（启动检查/退出同步/Dolt pull/push） | 🟡 中 | shared-protocol.md L0.3 的 session 生命周期，v3 无 | P1 |
| C10 | **beads 存储位置规则**（L0.4 项目根 vs 框架层） | 🟡 中 | v2 明确 .beads/ 在项目根不在框架层，v3 无隔离 | P1 |
| C11 | **beads 状态门禁**（L0.2 status transitions） | 🟡 中 | v2 定义 open→in_progress→closed 的前置条件，v3 无 | P1 |
| C12 | **成果物写入规则**（禁止写入 beads notes，独立文件） | 🟡 中 | v2 shared-checklist.md 有明确表格，v3 仅 DocSpec 路径 | P2 |
| C13 | **Anti-Patterns 列表** | 🟡 低 | v2 列出常见错误模式，v3 无 | P2 |
| C14 | **Handoff Protocol**（标准格式 + beads 更新） | 🟡 低 | v2 shared.md 有 Handoff 格式和 beads 命令，v3 terminal 有 on_exit 但格式未定义 | P2 |

---

## 三、汇总评分

| 维度 | v2 | v3 当前 | 差距 |
|------|-----|---------|------|
| 流程覆盖 | 5/5 | 4.5/5 | batch/release v3更好 |
| 门禁执行 | 3/5（靠AI自觉） | 2/5（声明式，无实际检查） | v3 需补执行逻辑 |
| 角色体系 | 5/5（14角色+15标准） | 3/5（12角色，无标准内容） | 需迁移 prompt+标准 |
| 任务追踪 | 5/5（beads 全套） | 1.5/5（结构就绪，调用未集成） | 关键缺口 |
| 架构边界 | 5/5（5层+访问矩阵） | 0/5（无隔离机制） | 关键缺口 |
| Skill 体系 | 5/5（6子技能完整） | 2.5/5（ref+元数据，无内容） | 需迁移 |
| 模板体系 | 5/5（20模板） | 1/5（字段存在，未填充） | 需迁移 |
| 结构性保证 | 2/5（靠文档约束） | 4/5（引擎强制+规则hard） | v3核心优势 |

**综合覆盖率: ~70%（含内容迁移） / ~85%（仅结构框架）**

---

## 四、优先级排序

### P0 — 不补则 v3 无法替代 v2

1. **beads 集成** — 引擎 on_enter/on_exit 自动调用 flow task CLI
2. **五层架构边界** — 至少在规则层声明（constraints 或 rule），防止跨层写入
3. **完成门禁执行逻辑** — gate_conditions 需从声明式升级为可执行检查

### P1 — 影响日常使用质量

4. **角色 prompt + 标准文件迁移** — 将 v2 的 14 prompts + 15 standards 适配到 v3 结构
5. **6 子技能内容迁移** — 至少 build/design/git/check-impl 4个核心
6. **20 模板填充** — DocSpec.template 填入模板路径
7. **Session 协议** — 启动/退出检查点
8. **QG-1~4 质量门** — 可作为 rules 注册
9. **变量占位符运行时解析** — 引擎解析 {TEAM_PATH} 等

### P2 — 锦上添花

10. Skill 自动路由 trigger pattern
11. ID 命名格式验证
12. 共识/Checklist 引擎集成
13. Anti-patterns 注册
14. Handoff 格式标准化

---

## 五、关键判断

**v3 能替代 v2 吗？**

**现在：不能。** 缺3个P0项（beads集成、架构边界、门禁执行），v3 是"流程描述器"而非"流程执行器"。

**补完P0后：基本可以。** 但角色prompt/标准/技能/模板等内容需从v2迁移，这是工作量而非设计缺口。

**v3 的核心优势：** 把 v2 的"AI自觉遵守"变成"引擎强制执行"——dispatch guard、hard rules、gate路由。这是v2永远做不到的。

**建议路径：** P0(1-2周) → P1内容迁移(1-2周) → P2优化(持续)
