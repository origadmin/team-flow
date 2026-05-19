# team-flow v2 共享工作流 - shared.md

> **TEAM_VERSION=8.0** | **更新日期**: 2026-05-10
> **引用**: 基础协议见 `shared-protocol.md`，详细检查清单见 `shared-checklist.md`

📌 三层门禁确保流程不可绕过；任务层管功能块，发布层管上线交付

---

## Layer 1: 入口门禁

**执行时机**: 收到用户输入的第一时间。

```
用户输入
    │
    ├── 单个需求 → 自动分类（Feature/Bug/Change/Analysis/Docs）
    ├── 多个需求（2+） → 批量任务流程
    ├── beads task ID？→ 进入任务执行流程
    ├── "R:" 前缀 + Milestone ID？→ 进入发布流程
    └── 澄清/状态查询？→ 直接回答
```

📌 所有用户输入自动进入分类流程，无需特殊前缀

**批量任务识别**:
- 用户一次性发送多个需求（用换行/序号/分号分隔）
- Triage 自动识别并触发批量任务流程

---

## Layer 2: 阶段门禁

**执行时机**: 角色开始执行前第一件事。

### Feature 任务阶段

```
Phase 0: 任务创建（Triage）
Phase 1: 需求分析 + 技术设计（Tech Lead）
Phase 2: 实现（Dev）
Phase 3: 验证（QA）
```

### Bugfix 任务阶段

```
Phase 0: Bug 接收（Triage）
Phase 1: 根因分析（Dev）
Phase 2: 修复实现（Dev）
Phase 3: 验证（QA）
```

### Batch 任务阶段（多任务场景）

📌 当用户一次性发送 2+ 个需求时，触发 Batch 任务流程

**核心原则**：Batch 是 Triage 内部管理机制，Agent 不知道 Batch 存在

```
Phase 0: Batch 创建（Triage）
Phase 1: 范围分析（Triage）
Phase 2: 用户确认（Triage）
Phase 3: 批量分发（Triage）
Phase 4: Agent 执行（多 Agent）
Phase 5: 成果验证（Triage）
Phase 6: 结果收集（Triage）
Phase 7: 状态更新（Triage）
```

### 并发控制规则

| 场景 | 规则 |
|------|------|
| 子任务 ≤ 3 个 | 同时分发 |
| 子任务 > 3 个 | 每批最多 3 个，上一批完成后分发下一批 |
| 有冲突任务 | 必须队列执行 |
| 依赖任务 | 等依赖完成后才分发 |
| 失败重试 | 单任务失败重试 3 次，仍失败标记 FAILED |

### 发布阶段

```
R-Phase 0: 就绪检查（Triage）
R-Phase 1: 集成验证（QA + DevOps）
R-Phase 2: 验收放行（PM）
R-Phase 3: 上线部署（DevOps）
```

---

## Layer 3: 完成门禁

详见 `shared-checklist.md`

---

## 资产包规格

### Feature 资产包

```
{DOCS_INTERNAL}/requirements/{TASK_ID}-{feature-name}/
├── R0_NAVIGATION_MATRIX.md ← 导航与入口矩阵（必须）
├── SPEC.md, AC.md
├── R1_DATA_MODEL.md, R2_STATE_MACHINE.md, R3_API_CONTRACT.md
└── SCOPE.md            ← Dev 创建
```

### Bugfix 资产包

```
{DOCS_INTERNAL}/reports/bugs/{bug-id}-R{N}/
├── RCA.md              ← 根因分析
├── TEST_CASE.md        ← 复现验证
└── SCOPE.md            ← 变更报告
```

### Change 资产包

```
{DOCS_INTERNAL}/reports/changes/{change-id}/
└── SCOPE.md            ← 执行角色创建（必须）
```

### Batch 资产包

```
{DOCS_INTERNAL}/batch/{batch-id}/
├── PLAN.md              ← 执行计划（Triage 创建）
└── {F001,B001...}/     ← 子任务产出物
```

📌 资产目录命名规则：
| 类型 | 格式 | R 后缀 |
|------|------|--------|
| Feature | `{ID}-{name}/` | ❌ |
| Bug | `{bug-id}-R{N}/` | ✅ |
| Change | `{change-id}/` | ❌ |
| Batch | `batch/{batch-id}/` | ❌ |

---

## 会话结束交接文档（强制）

**规则**：每一个会话结束时，Triage 必须创建交接文档。

```
{DOCS_INTERNAL}/handoff-{YYYY}-{MM}-{DD}.md
```

**强制内容**：
- 当前项目状态
- 本会话已完成的工作（含 git 引用）
- 识别但未修复的问题
- 待完成的任务
- 关键文件路径
- 继续工作提示词

---

## 文件空间定义

### Framework Layer (READ-ONLY)

```
framework/
  {TEAM_PATH}/
    ├── SKILL.md       ← 入口
    ├── BOUNDARY.md    ← 层级规则
    ├── prompts/       ← 角色执行规则
    ├── workflows/     ← 共享工作流
    └── templates/     ← 文档模板
```

### Project Layer (READ-WRITE)

```
{PROJECT}/
  .team/              ← AI 操作文件
    ├── project.md
    └── task-pool-export.md
  .beads/            ← beads 数据库（用 flow task 命令管理）
{DOCS_INTERNAL}/    ← 项目文档
```

---

## ID 命名规则

| ID | 类型 |
|----|------|
| F{NNN} | Feature |
| B{NNN} | Bugfix（资产目录带 R 后缀）|
| C{NNN} | Change |
| D{NNN} | Documentation |
| A{NNN} | Analysis |
| BATCH-{YYYYMMDD}-{N} | Batch |

📌 TYPE 只允许 F/B/C/D/A/BATCH 六种，SEQUENCE 从 001 递增

---

## Role Handoff Protocol

### Standard Handoff Format

```markdown
## Handoff: <task-id> → <target-role>

**From**: <your-role>
**To**: <target-role>
**Timestamp**: YYYY-MM-DD HH:MM

**Summary**: <1-2 sentences>

**Deliverables**:
- [ ] <artifact-1> at <path>

**Next Actions**:
1. <action-1>

**Blockers**: <none | list>
```

### Handoff via beads

```bash
flow task update <id> \
  --notes "## Handoff → <target-role>\n\nSummary: [details]" \
  --assignee "<target-role>" \
  --add-label phase:<next-phase>
```

---

## 相关规范

| 规范 | 位置 |
|------|------|
| 基础协议 | `{TEAM_PATH}/workflows/shared-protocol.md` |
| 完成门禁清单 | `{TEAM_PATH}/workflows/shared-checklist.md` |
| Triage 规范 | `{TEAM_PATH}/workflows/roles/triage-standards.md` |
| 开发规范 | `{TEAM_PATH}/workflows/roles/development-standards.md` |
| 测试规范 | `{TEAM_PATH}/workflows/roles/test-standards.md` |

---

## 相关模板

| 模板 | 路径 |
|------|------|
| 闭环验证报告 | `{TEAM_PATH}/templates/closed-loop-verification-template.md` |
| SCOPE.md | `{TEAM_PATH}/templates/scope-template.md` |
| 导航矩阵 | `{TEAM_PATH}/templates/r0-navigation-matrix-template.md` |

---
