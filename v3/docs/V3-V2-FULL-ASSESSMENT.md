# v3 vs v2 完整覆盖性评估 (修正版)

**评估时间**: 2026-05-20 03:40 GMT+8 (修正版)
**前版错误**: 把 v3 当成"全自动执行器"评估，要求引擎直接调 beads/跑测试/检查文件。这是错的。

## 修正后的评估模型

**v3 设计哲学: 引擎输出结构化指引 → AI 按规则执行**

- `flow task` CLI → AI 按 SKILL 规则调用，不需要引擎自动调
- `gate_conditions` → 引擎告诉 AI 要检查什么，AI 执行检查
- 架构边界 → 作为规则/约束，AI 遵守
- prompt/标准/技能 → 引擎输出路径和摘要，AI 按需加载

所以评估维度是：**v3 的 flow JSON + 引擎输出 + SKILL 规则是否覆盖了 v2 所有指导信息？** 而不是引擎是否程序化执行了一切。

---

## 一、已完整覆盖

| # | v2 功能 | v3 实现 | 说明 |
|---|--------|---------|------|
| 1 | 流程路径 (Feature/Bug/Change/Analysis/Hotfix) | 5个独立 flow + dev-flow | ✅ |
| 2 | 角色分配 (Triage→TechLead→Dev→QA) | flow nodes role + registry | ✅ |
| 3 | 三层门禁 (Entry/Phase/Completion) | gate 节点 + conditions + on_pass/on_fail | ✅ 比v2更强 |
| 4 | 状态追踪 (beads phase labels) | on_enter + prompt_directives 指引 AI 调 flow task | ✅ AI按规则调用 |
| 5 | Dispatch Guard | rule: dispatch-guard(hard) + prompt_directives[0] | ✅ 双重保证 |
| 6 | 交付物路径 | DocSpec(name, path, format, required) | ✅ |
| 7 | 路径解析 | {TEAM_PATH} 等占位符 + SKILL.md 定义如何解析 | ✅ AI按SKILL规则解析 |
| 8 | 并发控制 (batch ≤3) | batch-flow.json | ✅ |
| 9 | Bug R-iteration | bugfix-flow branch-r-iteration + gate | ✅ |
| 10 | Status Line | buildStatusLine() 引擎自动生成 | ✅ 比v2更强 |
| 11 | 角色切换 | next_options + role + edge 条件 | ✅ |
| 12 | 回归防护 | rule: regression-guard(hard) | ✅ |
| 13 | 硬约束 (禁删/禁BOM/禁跳toolchain) | rule: hard-constraints(hard) | ✅ |
| 14 | Output Guard 5步 | gate-completion 条件 + rule: output-guard | ✅ |
| 15 | 会话交接 | terminal-success on_exit | ✅ |
| 16 | 资产包规格 | DocSpec 路径模板 | ✅ |
| 17 | 批量任务 | batch-flow.json | ✅ v3新增 |
| 18 | 发布流程 | release-flow.json | ✅ v3新增 |
| 19 | Skill 引用 | registry skills[] + trigger/description/path | ✅ |
| 20 | Prompt 加载 | prompt_source + standards_source + prompts[] | ✅ |
| 21 | prompt_directives | RoleDefinition.prompt_directives[] | ✅ v3新增 |
| 22 | 完成门禁 deliverables | gate-completion: deliverables_complete | ✅ |
| 23 | 五层架构边界 | hard-constraints rule 中可声明 | ✅ 作为规则AI遵守 |
| 24 | beads 状态门禁 | prompt_directives 指引 AI 用 flow task ready/claim | ✅ AI按规则执行 |
| 25 | beads 存储位置 | hard-constraints rule 或 prompt_directives 声明 | ✅ 作为规则 |
| 26 | QG-1~4 质量门 | 可注册为 rules (当前未注册，但框架支持) | 🟡 框架就绪 |
| 27 | Session 协议 | SKILL.md 定义启动/退出行为 | 🟡 框架就绪 |
| 28 | Anti-Patterns | 可注册为 rules | 🟡 框架就绪 |
| 29 | 成果物写入规则 | DocSpec path + hard-constraints | ✅ |
| 30 | Handoff Protocol | terminal on_exit + SKILL.md 定义格式 | 🟡 格式待定义 |

---

## 二、真正的缺口（v2 有指导信息，v3 引擎输出中缺失）

| # | 缺口 | 说明 | 类型 | 优先级 |
|---|------|------|------|--------|
| 1 | **shared-checklist.md 40+ 检查项** 未结构化到 v3 | v2 Feature 17项/Bugfix 25项/R-Phase 12项详细检查清单，v3 gate_conditions 只有 6 个大项 | 内容缺口 | P1 |
| 2 | **14 角色 prompt 文件内容** | v3 有 prompt_source 路径和 prompt_directives 摘要，但完整 prompt 内容不在 v3 目录 | 迁移工作 | P1 |
| 3 | **15 角色标准文件内容** | v3 有 standards_source 路径，文件不在 v3 目录 | 迁移工作 | P1 |
| 4 | **6 子技能完整内容** | v3 skills[] 有 ref/trigger/description/path，但 SKILL.md 内容不在 v3 | 迁移工作 | P1 |
| 5 | **20 模板文件** | DocSpec.template 字段存在但未填充路径值 | 内容填充 | P1 |
| 6 | **QG-1~4 未注册为 rules** | 框架支持但未在 flow JSON 中声明 | 注册工作 | P2 |
| 7 | **Handoff 格式未定义** | on_exit 触发但输出格式未规范 | 规范定义 | P2 |
| 8 | **ID 命名格式验证** | 规则在 CORRECTION-001 但 flow validate 不检查 | 引擎增强 | P2 |
| 9 | **game-design-flow/novel-flow 缺 components** | 这两个 flow 无 roles/rules/skills 注册 | 内容填充 | P2 |

---

## 三、修正后评估

| 维度 | v2 | v3 当前 | 说明 |
|------|-----|---------|------|
| 流程覆盖 | 5 | 5 | 全部流程有对应 flow JSON |
| 门禁体系 | 3(靠AI自觉) | 4(引擎+规则双重) | v3更强 |
| 角色体系 | 5 | 3.5 | 12角色+路径+摘要，但内容文件需迁移 |
| 任务追踪 | 5 | 4 | prompt_directives 指引 AI 调用 flow task，结构足够 |
| 架构边界 | 5 | 3.5 | 可作为规则声明，但不如 v2 五层矩阵详细 |
| Skill 体系 | 5 | 3 | ref+元数据完整，内容需迁移 |
| 模板体系 | 5 | 2 | 字段在，值未填 |
| 结构性保证 | 2 | 4.5 | v3 核心优势 |

**综合覆盖率: ~90%（结构+规则）/ ~75%（含内容迁移）**

结构层面 v3 已基本覆盖 v2 全部功能。剩余工作主要是**内容迁移**（prompt/标准/技能/模板）和**细节填充**（checklist 细化、QG 注册、template 值），不是设计缺口。

---

## 四、优先级排序

### P1 — 内容迁移（工作量，非设计缺口）

1. checklist 细化 — 将 shared-checklist.md 40+ 项结构化到 gate_conditions 或 rules
2. 14 prompt + 15 standards 文件适配 v3 路径
3. 6 子技能内容迁移
4. 20 模板路径填充到 DocSpec.template
5. QG-1~4 注册为 rules

### P2 — 细节优化

6. Handoff 输出格式定义
7. ID 命名格式 validate 检查
8. game-design-flow/novel-flow components 补充
9. 五层架构边界规则细化（从 hard-constraints 拆分为独立 rule）

---

## 五、关键判断

**v3 能替代 v2 吗？**

**结构层面：基本可以。** v3 的 flow JSON + 引擎输出 + SKILL 规则覆盖了 v2 全部核心流程指导。beads 调用、门禁检查、架构边界都是 AI 按规则执行，不需要引擎直接调。

**内容层面：还差迁移工作。** prompt/标准/技能/模板的内容需要从 v2 适配到 v3 路径结构。这是纯工作量，不是设计缺陷。

**v3 比 v2 多了什么：**
- prompt_directives（核心指令内联，v2 无）
- 引擎自动 StatusLine（v2 靠 AI 自觉）
- batch-flow / release-flow（v2 无独立 flow）
- component registry（消除跨 flow 重复定义）
