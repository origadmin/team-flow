# Layer Boundary — _team v2

> **Purpose**: Define what goes where, preventing cross-layer contamination

## Layer Definitions

### Layer 0: Framework Core (Read-Only)

```
_team/v2/
├── SKILL.md              ← Entry point
├── prompts/*.md          ← Role execution rules (framework-standard)
├── workflows/*.md        ← Shared workflows (framework-standard)
└── docs/*.md             ← Reference documentation
```

**Rules**:
- AI reads from here
- AI **NEVER** writes to here
- Projects inherit via `project/.team/SKILL.md` pointing here
- To modify: update framework, then propagate to all projects

### Layer 1: Project Config (Project-Scoped)

```
projects/orig-cms/.team/
├── SKILL.md              ← Points to framework + project overrides
├── project.md            ← Project metadata (paths, team, status)
├── ai-context.md         ← Current session focus, recent decisions
└── roles/*.md            ← Project-specific role configs (optional)
```

**Rules**:
- AI reads/writes here for project context
- `project.md` defines project paths
- `ai-context.md` tracks session state
- Modifications here affect only this project

### Layer 2: Project Data (Project-Scoped, Read-Only for AI)

```
projects/orig-cms/_docs/orig-cms/
├── requirements/{ID}/       ← 需求文档（SPEC.md + AC.md + R1-R5）
│   ├── F014-unified-pagination/
│   │   ├── SPEC.md          ← 功能规格
│   │   ├── AC.md            ← 验收标准
│   │   ├── R1_DATA_MODEL.md ← 数据模型
│   │   ├── R2_STATE_MACHINE.md ← 状态机
│   │   └── R3_API_CONTRACT.md  ← API 契约
│   └── ...
├── reports/
│   ├── bugs/{ID}/            ← Bug 报告（按 R 子目录隔离）
│   │   ├── B001/
│   │   │   ├── INDEX.md      ← ⚠️ 索引文件（AI 必须先读此文件）
│   │   │   ├── R1/           ← 首次修复
│   │   │   │   ├── RCA.md
│   │   │   │   ├── TEST_CASE.md
│   │   │   │   └── SCOPE.md
│   │   │   ├── R2/           ← 第二次修复（reopen 后）
│   │   │   │   ├── RCA.md
│   │   │   │   ├── TEST_CASE.md
│   │   │   │   └── SCOPE.md
│   │   │   └── R3/           ← 当前迭代（INDEX.md 指向此处）
│   │   │       ├── RCA.md
│   │   │       ├── TEST_CASE.md
│   │   │       └── SCOPE.md
│   │   └── ...
│   ├── analysis/            ← 分析报告
│   │   ├── A008-quality-check-enhancement.md
│   │   └── ...
│   └── changes/{ID}-R{n}/   ← 变更报告
│       ├── C005-R1/
│       │   └── SCOPE.md
│       └── ...
├── analysis/                ← 技术分析文档
│   ├── watch-page/
│   ├── upload-flow/
│   └── ...
├── architecture/            ← 架构设计文档
│   ├── channel/
│   ├── api/
│   └── design/
├── conventions/             ← 项目约定
│   ├── common.md
│   ├── dev-backend.md
│   └── dev-frontend.md
├── design/                  ← 设计规范
│   ├── tokens.md
│   ├── components.md
│   └── layouts.md
├── lessons/                 ← 经验教训
│   └── dev-common.md
├── reference/               ← 参考资料
│   └── ...
└── meetings/                ← 会议纪要
```

**Rules**:
- AI reads from here for context
- AI writes here **ONLY** when explicitly tasked to create documentation
- Human reviews all writes here

### beads → 文档映射规则

| beads 类型 | _team ID 格式 | 文档目录 | Phase → 产出物 |
|-----------|--------------|---------|---------------|
| feature | F{xxx} | `requirements/F{xxx}-{name}/` | phase:analyze → SPEC.md + AC.md |
| | | | phase:design → R1_DATA_MODEL + R2_STATE_MACHINE + R3_API_CONTRACT |
| | | | phase:implement → 代码 + 测试 |
| | | | phase:verify → TEST_COVERAGE.md |
| | | | phase:review → SCOPE.md |
| bug | B{xxx} | `reports/bugs/B{xxx}/` | phase:analyze → R{n}/RCA.md |
| | | | phase:implement → 修复代码 + 复现测试 |
| | | | phase:verify → R{n}/TEST_CASE.md |
| | | | phase:review → R{n}/SCOPE.md |
| | | | ⚠️ R{n} 由 beads EventReopened 事件自动追踪，首次=R1 |
| | | | ⚠️ 每个 R 独立子目录，INDEX.md 标记当前迭代 |
| | | | ⚠️ AI 只读 INDEX.md + 当前 R 目录，禁止读历史 R |
| task (change) | C{xxx} | `reports/changes/C{xxx}-R{n}/` | phase:analyze → SPEC.md |
| | | | phase:review → SCOPE.md |
| task (analysis) | A{xxx} | `reports/analysis/A{xxx}-{name}.md` | phase:analyze → 分析文档 |
| | | | phase:review → 结论摘要 |

**ID 查找规则**:

### Deliverable Write Separation (v1 Lesson)

> **v1 教训**: AI 把根因分析、设计决策、测试结果等详情内容全部写入 task-pool.md，导致文件从 75 行膨胀到 1822 行，且大量 Bug 缺少独立 RCA.md。v2 中同样的风险会转移到 beads notes。

**原则**: 状态追踪是轻量表，详情内容必须写入独立成果物文件。

| 内容类型 | ✅ 正确位置 | ❌ 禁止位置 |
|---------|-----------|-----------|
| 根因分析 | `{DOCS_INTERNAL}/reports/bugs/B{NNN}/R{n}/RCA.md` | beads notes / task-pool-export.md |
| 复现测试 | `{DOCS_INTERNAL}/reports/bugs/B{NNN}/R{n}/TEST_CASE.md` | beads notes / task-pool-export.md |
| 变更报告 | `{DOCS_INTERNAL}/reports/changes/C{NNN}-R{n}/SCOPE.md` | beads notes / task-pool-export.md |
| 需求规格 | `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/SPEC.md` | beads notes / task-pool-export.md |
| 设计决策 | `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/R1-R5.md` | beads notes / task-pool-export.md |
| 进度摘要 | `bd update <id> --notes "COMPLETED: X IN PROGRESS: Y"` | — (notes 只写进度，不写详情) |

**beads notes 用途**: 只记录进度摘要和交接信息，**禁止**写入根因分析、代码片段、测试结果等详情。

**ID 查找规则**:

```bash
# 从 beads ID 查找 _team ID
bd show <beads-id> --json | jq '.externalRef'
# → "F014"

# 从 _team ID 查找 beads ID
bd list --json | jq '.[] | select(.externalRef == "F014") | .id'
# → "cms-xxx"

# 查询 Bug 的 R 迭代次数（reopen 次数 + 1）
bd show <beads-id> --json | jq '[.events[] | select(.event_type == "reopened")] | length + 1'
# → 2 (表示当前是 R2)

# 从 _team ID 定位文档目录
# F014 → _docs/.../requirements/F014-unified-pagination/
# B001 → _docs/.../reports/bugs/B001/  (目录不带 R 后缀)
# A008 → _docs/.../reports/analysis/A008-quality-check-enhancement.md
```

### Layer 3: Beads Database (Single Source of Truth)

```
projects/orig-cms/.beads/
├── *.db                  ← SQLite/Dolt database
├── issues.jsonl          ← Issue export
└── task-pool-mapping.json ← _team ID ↔ beads ID mapping
```

**Rules**:
- AI interacts via `bd` CLI only
- **NEVER** directly edit `.beads/*.db`
- **NEVER** directly edit `.beads/issues.jsonl`
- Use `bd create/update/close` for all modifications

### Layer 4: Implementation (AI workspace)

```
projects/orig-cms/
├── cmd/                  ← Application code
├── internal/
├── ent/schema/
├── web/                  ← Frontend code
└── proto/                ← API definitions
```

**Rules**:
- AI reads/writes freely during implementation
- Follows code quality rules from qclaw-rules
- Commits to git after changes

## Cross-Layer Access Rules

| From → To | L0 Framework | L1 Project Config | L2 Project Docs | L3 Beads | L4 Implementation |
|-----------|--------------|-------------------|------------------|----------|-------------------|
| **L0** | ✅ Read | ❌ | ❌ | ❌ | ❌ |
| **L1** | ✅ Read | ✅ Read/Write | ✅ Read | ✅ CLI | ✅ Read/Write |
| **L2** | ✅ Read | ✅ Read | ✅ Read | ✅ CLI | ✅ Read |
| **L3** | ❌ | ✅ CLI | ❌ | ✅ CLI | ❌ |
| **L4** | ❌ | ✅ Read | ✅ Read | ❌ | ✅ Read/Write |

## Forbidden Actions

❌ **NEVER**:
- Write to `_team/v2/` (framework layer)
- Edit `.beads/*.db` directly
- Edit `.beads/issues.jsonl` directly
- Create task state outside beads
- Mix project configs across projects
- Dump deliverable content (root cause analysis, design decisions, test results) into beads notes or task-pool-export.md — **always write to independent deliverable files**
- Add "Task Details" / "Deliverable Tracking" / "Current Status" sections to task-pool-export.md
- Modify `_team/` rules to solve project-specific problems — project constraints go to `.team/project.md §CONSTRAINTS` and `_docs/.../lessons/`

## Permitted Actions

✅ **ALWAYS**:
- Use `bd` CLI for all task operations
- Read from L0 for rules
- Write to L1 for project state
- Write to L4 for implementation
- Write to L2 only when documenting
- Commit to git after changes

## Session Context Injection

When starting a session, AI loads:

1. **L0**: `_team/v2/SKILL.md` (rules + structure)
2. **L1**: `.team/project.md` (project paths + team)
3. **L1**: `.team/ai-context.md` (recent focus)
4. **L3**: `bd ready --json` (available tasks)

Example CLAUDE.md:

```markdown
# CLAUDE.md — orig-cms project

Load _team rules: {TEAM_PATH}/v2/SKILL.md
Project config: .team/project.md
Session context: .team/ai-context.md
```

## Migration from v1

If migrating from `_team` (v1):

```bash
# 1. Archive v1 (don't delete)
mv _team _team_v1_archive

# 2. Point project to v2
echo "Load _team rules: {TEAM_PATH}/v2/SKILL.md" > projects/orig-cms/CLAUDE.md

# 3. Migrate tasks to beads
.\_team\v2\scripts\migrate-tasks.ps1 -ProjectPath projects/orig-cms

# 4. Update agent prompts to use bd CLI
```

## Quality Gates

Before any write operation, verify:

```bash
# Am I in the right layer?
pwd

# Is this file in my write scope?
# L0 → Never
# L1 → Yes, if in .team/
# L2 → Yes, if documenting
# L3 → Never (use bd CLI)
# L4 → Yes, if implementing
```
