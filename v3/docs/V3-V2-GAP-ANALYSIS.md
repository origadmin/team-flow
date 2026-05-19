# v3 vs v2 复刻能力评估报告

> 评估方法: skills-creator 6-step framework
> 日期: 2026-05-20
> 评估对象: v3 flow JSON + flow3.exe engine vs v2 SKILL.md + prompts + workflows

## 评估结论

**v3 当前可复刻 v2 约 65% 的核心流程，但存在 6 个结构性缺口和 4 个引擎层缺失。**

---

## 一、v2 功能清单 vs v3 覆盖状态

### ✅ 完全复刻 (A类: 引擎结构性吸收)

| # | v2 功能 | v3 实现方式 | 状态 |
|---|---------|------------|------|
| A1 | 流程执行路由 | flow JSON → engine dispatch | ✅ |
| A2 | 角色分配 | node.config.role + components.roles | ✅ |
| A3 | 三层门禁 | gate-entry / gate-quality / gate-completion | ✅ |
| A4 | Dispatch Guard | components.rules dispatch-guard (enforcement: hard) | ✅ |
| A5 | 交付物追踪 | node.docs[].required + gate-completion deliverables_complete | ✅ |
| A6 | 状态追踪 | node.on_enter update_task_phase | ✅ |
| A7 | 路径解析 | flow config paths --json → variables | ✅ |

### ✅ 形式不同但可复刻 (B类: 需适配)

| # | v2 功能 | v3 现状 | 差距 | 严重度 |
|---|---------|---------|------|--------|
| B1 | 11 角色定义 | 8→12 Role enum | ✅ 已扩展到12 | - |
| B2 | Completion Gate | gate-completion 5-6条件 | ✅ feature/bugfix已有 | - |
| B3 | Skill Routing | components.skills ref | ⚠️ 仅feature-flow注册2个，其他flow缺 | P1 |
| B4 | Status Line | buildStatusLine() | ✅ [Role\|Task\|Phase\|Asset] | - |
| B5 | 资产包规格 | node.docs[] 定义 | ⚠️ bugfix缺INDEX.md，feature缺R0_NAVIGATION_MATRIX | P1 |
| B6 | Bugfix R迭代 | branch-r-iteration + max 3 | ✅ 已实现 | - |
| B7 | Regression Guard | components.rules | ✅ 已注册 | - |
| B8 | Hard Constraints | components.rules hard-constraints | ✅ 已注册 | - |
| B9 | Output Guard | components.rules + on_exit check | ⚠️ 仅部分节点有on_exit检查 | P1 |
| B10 | ID命名规则 | CORRECTION-001 | ⚠️ 文档存在但flow JSON未强制校验 | P2 |
| B11 | 会话交接 | terminal on_exit handoff-doc | ⚠️ 仅feature/bugfix有，其他缺 | P1 |

### ❌ 缺失 (C类: v3 新增或 v2 无法映射)

| # | 功能 | v3 状态 | 影响 |
|---|------|---------|------|
| C1 | Batch 任务流程 | ❌ 无 batch-flow.json | v2 批量任务7阶段完全缺失 |
| C2 | Release 发布流程 | ❌ 无 release-flow.json | v2 发布4阶段完全缺失 |
| C3 | 共识机制 | ❌ 无 consensus 组件 | v2 的 .team/consensus.md 无对应 |
| C4 | Checklist 覆盖 | ❌ 无 checklist 组件 | v2 的 .team/checklist.md 无对应 |
| C5 | Components 内容定义 | ⚠️ 已有components骨架 | 内容不够丰富 |

### ❌ 引擎层缺失 (D类: flow3.exe 未实现)

| # | 功能 | v3 状态 | 影响 |
|---|------|---------|------|
| D1 | Rule 内容加载 | engine 输出rules列表但不输出规则正文 | AI无法知道规则具体内容 |
| D2 | Phase 状态同步 | engine 不自动更新 beads labels | Phase 状态依赖AI手动同步 |
| D3 | Deliverable 存在性检查 | gate condition deliverables_complete 未实现文件检查 | 门禁形同虚设 |
| D4 | 变量解析 | variables 占位符（如{iteration}）未在engine中解析 | R-iteration路径无法动态生成 |
| D5 | Branch 条件求值 | branch-r-iteration 的 when 条件无法实际求值 | 分支决策依赖AI判断 |

---

## 二、按 skills-creator 6-step 评估

### Step 1: 理解 (Understanding)

**v3 是否理解 v2 的完整工作流？**

- ✅ 核心流程理解: Feature/Bugfix/Change/Analysis/Hotfix 都有对应flow
- ❌ Batch 流程理解缺失: v2 有完整的7阶段批量任务管理，v3 无对应
- ❌ Release 流程理解缺失: v2 有4阶段发布管理，v3 无对应
- ⚠️ 角色prompt理解缺失: v2 有14个详细prompt文件，v3 没有 prompt 加载机制

**评分: 6/10** — 核心理解到位，但批量/发布两个完整流程被遗漏

### Step 2: 规划 (Planning)

**v3 的可复用内容规划是否充分？**

- ✅ components 复用: roles/rules/tools/skills 可跨flow复用
- ✅ 变量机制: docs_path/TEAM_PATH/task_id 支持路径参数化
- ❌ 缺少 references 机制: v2 有 references/ 目录（commands.md, path-resolution.md, agent-mapping.md），v3 无对应
- ❌ 缺少 templates 机制: v2 有20+模板文件，v3 无模板系统
- ❌ 缺少 workflow 共享: v2 有 shared.md/shared-protocol.md/shared-checklist.md，v3 把这些打散到各flow

**评分: 5/10** — 结构化规划不足，共享机制缺失

### Step 3: 初始化 (Initialization)

**v3 的初始化流程是否完整？**

- ✅ flow3.exe proc create — 可创建flow
- ✅ flow3.exe proc validate — 可验证flow
- ✅ flow3.exe proc run — 可运行flow节点
- ❌ 无 flow3.exe proc init — 缺少项目初始化命令
- ❌ 无 .team/ 自动配置 — v2 的 flow init 创建项目配置，v3 无对应

**评分: 6/10**

### Step 4: 编辑 (Editing)

**v3 的 skill 文件是否足够指导正确开发？**

- v3-create SKILL.md: 557行 — 内容丰富但超500行限制(C8)
- v3-exec SKILL.md: 280行 — 基本够用
- ❌ 缺少 v3-design SKILL.md — 无设计流程指导
- ❌ 缺少 v3-review SKILL.md — 无评审流程指导
- ❌ 缺少 v3-git SKILL.md — 无git操作指导

对比 v2 有6个专门skill: design, build, git, check-impl, review, evolve

**评分: 4/10** — skill 数量严重不足

### Step 5: 打包 (Packaging)

**v3 的分发和部署是否可行？**

- ✅ flow JSON 是自包含的
- ✅ flow3.exe 是独立二进制
- ❌ 无 package_skill.py 对等工具
- ❌ 无版本兼容性检查

**评分: 5/10**

### Step 6: 迭代 (Iteration)

**v3 是否支持自身迭代改进？**

- ⚠️ version-evolve-flow.json 已设计但未实现
- ⚠️ 自举路径（v2→v3→v4）理论可行但缺少元流程
- ❌ 无跨版本兼容机制

**评分: 3/10**

---

## 三、结构性缺口详细分析

### 缺口 1: Batch 流程 (C1)

v2 有完整的批量任务流程（7阶段），支持：
- 自动识别多需求输入
- 并发控制（≤3同时，依赖排序）
- 失败重试（3次）
- 子任务结果收集

v3 需要: batch-flow.json + parallel node type + 并发控制逻辑

### 缺口 2: Release 流程 (C2)

v2 有4阶段发布流程：
- R-Phase 0: 就绪检查（Triage）
- R-Phase 1: 集成验证（QA + DevOps）
- R-Phase 2: 验收放行（PM）
- R-Phase 3: 上线部署（DevOps）

v3 需要: release-flow.json + DevOps/PM 角色激活

### 缺口 3: Rule 内容加载 (D1)

当前 engine 输出:
```json
"rules": [{"id": "dispatch-guard", "source": "builtin"}]
```

缺少: 规则正文内容。AI 收到 rule id 但不知道规则要求什么。

修复方案: engine 读取 components.rules[].description 注入到输出，或者 AI 根据 rule id 加载对应 prompt/standard 文件。

### 缺口 4: Prompt 加载机制

v2 有14个详细prompt文件（triage.md, dev.md, qa-engineer.md 等），每个300-800行。

v3 没有对应的 prompt 加载机制。当 engine 输出 `role: "Dev"` 时，AI 不知道应该加载哪个 prompt。

修复方案: 在 components.roles[] 中添加 `prompt_ref` 字段，或在 engine 输出中添加 prompt 路径。

### 缺口 5: 共享工作流

v2 的 shared.md (8.0) 包含：
- 三层门禁详细定义
- 资产包规格
- 文件空间定义
- ID命名规则
- Role Handoff Protocol
- 会话结束交接文档

v3 把这些信息分散到各 flow JSON，导致：
- 重复定义（dispatch-guard 在每个flow重复声明）
- 不一致风险（不同flow可能定义不同版本）
- 修改困难（改一个规则需要改所有flow）

修复方案: 提取共享 components 到 shared-components.json，各 flow 用 ref 引用。

### 缺口 6: 模板系统

v2 有20+模板文件，AI 读取模板 → 写入项目目录。

v3 没有模板系统。当 AI 需要创建 SPEC.md 时，没有标准格式参考。

修复方案: 在 v3 skill 中添加 templates/ 目录，或 engine 支持 template 加载。

---

## 四、修复优先级建议

### P0 (必须，阻塞v3替代v2)

1. **Rule 内容输出** (D1) — 没有规则正文，AI无法遵守
2. **Prompt 加载路径** (缺口4) — 没有prompt路径，AI不知道如何执行角色
3. **Batch flow** (C1) — v2核心流程，缺了就不能替代v2

### P1 (重要，影响日常使用)

4. **Release flow** (C2) — 发布管理缺失
5. **共享 Components 提取** (缺口5) — 消除重复和不一致
6. **Deliverable 文件检查** (D3) — 门禁目前是空壳
7. **变量解析** (D4) — R-iteration路径动态生成
8. **补充缺失模板** (B5) — INDEX.md, R0_NAVIGATION_MATRIX.md

### P2 (增强，提升效率)

9. **模板系统** (缺口6) — 标准化文档产出
10. **v3 SKILL.md 拆分** (C8) — 当前557行超限
11. **Branch 条件求值** (D5) — 引擎自动决策
12. **version-evolve-flow** — 自举支持

---

## 五、总结

| 维度 | v2 | v3 | 差距 |
|------|----|----|------|
| 流程覆盖 | Feature/Bugfix/Change/Analysis/Hotfix/Batch/Release | 5/7 缺Batch/Release | ❌ 缺2个完整流程 |
| 角色定义 | 11角色 + 14 prompt 文件 | 12角色 enum + 0 prompt | ❌ 有enum无prompt |
| 规则系统 | SKILL.md + workflows + roles (40+文件) | components.rules (声明式) | ❌ 声明有但内容无 |
| 门禁系统 | 三层 + 详细检查清单 | 三层 + conditions | ⚠️ 结构有但执行弱 |
| 模板系统 | 20+模板文件 | 无 | ❌ 完全缺失 |
| 共享机制 | shared.md + shared-protocol + shared-checklist | 无 | ❌ 完全缺失 |
| 引擎能力 | flow task CLI (beads) | flow3.exe proc run (JSON) | ⚠️ JSON输出但缺实际执行 |
| 自迭代 | 无 | 设计中 | ⚠️ 设计中未实现 |

**核心判断: v3 的结构设计优于v2（声明式、可验证、引擎驱动），但内容层严重不足。v3 目前是"骨架完整、血肉缺失"的状态。要让v3替代v2，P0的3个缺口必须先补齐。**
