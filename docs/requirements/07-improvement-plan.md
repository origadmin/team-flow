# 改进计划

> **版本**: v1.2 | **日期**: 2026-05-11 | **状态**: 待确认

## 改进原则

1. **先确认需求，再动手修改** — 不再边理解边改
2. **规则只定义一次** — 消除重复定义导致的不一致
3. **配置驱动** — 用配置替代猜测
4. **最小修改** — 只改必须改的，不做额外"优化"
5. **门禁式控制** — 每个操作有前置条件，不满足就禁止执行

## 已确认决策

### D1: Bug 目录结构 — `B001/R1/`

```
B001/               ← 任务目录
  INDEX.md          ← 该任务的 cr-index 跟踪表
  R1/               ← 第一轮
    RCA.md
    FIX.md
  R2/               ← 第二轮（如果需要）
    RCA.md
```

- 每个 task 管自己的 INDEX.md
- 全局讨论 index 放 `.team/` 根目录
- 内部管理不污染正常文档
- 如果配置了内部文档路径（如 `_docs/`），INDEX 放内部文档路径

### D2: 版本号统一

当前 3 个版本号不一致：
- SKILL.md frontmatter: `version: 2.3`
- SKILL.md body: `**Version**: v2.4`
- shared.md: `TEAM_VERSION=8.1`

统一方案：只保留 frontmatter 一个版本号，body 和 shared.md 引用 frontmatter。

### D3: 重复规则处理

shared.md 中的重复规则不直接删除，改为引用 SKILL.md：

```markdown
> 项目管理模式定义见 SKILL.md §Project Context
```

评估：AI 读 shared.md 时知道去 SKILL.md 找，不会漏掉规则。

### D4: 技能文件与 SKILL 技能统一

team-flow 的技能文件应兼容 Skill tool 的格式：

| 加载方式 | 场景 | 说明 |
|----------|------|------|
| Triage 分发注入 | 子 Agent 执行任务时 | 内联方式，技能内容注入 prompt |
| Skill tool 调用 | AI 主动加载时 | 调用方式，通过 Skill tool invoke |

两种方式加载的是同一份技能文件，内容一致。技能文件格式应兼容 Skill tool 的 skill 定义格式。

### D5: 流程成果存放位置

优先级：
1. **内部文档路径**（如 `_docs/team-flow/`）— 最佳
2. **`.team/`** — 没有内部文档时的降级方案

当前 framework 的内部文档路径是 `_docs/`，team-flow 流程成果放 `_docs/team-flow/`。

## 改进项

### P1: 修复 bd CLI 残留（I-01）

**范围**: triage.md、BOUNDARY.md、shared.md
**操作**: 将所有 `bd` CLI 引用替换为 `flow task` CLI
**验证**: 全文搜索 `bd CLI`、`bd --help`、`Use bd`，结果应为 0

### P2: 统一项目上下文规则定义位置（I-04, I-13）

**问题**: 项目管理模式在 4 个文件中重复定义
**方案**（D3）:
- SKILL.md 定义核心规则（简版）
- 其他文件引用 SKILL.md，不重复定义
- 具体操作流程在 triage.md 中定义

### P3: 统一 Bug 目录结构（I-12）

**方案**（D1）: 统一为 `B001/R1/` 结构，每个任务目录含 INDEX.md

### P4: 修复 triage.md 标题和内容不一致（I-02, I-03）

**操作**: 
- 第 56 行: "Detect directory type" → "Resolve project context"
- 第 76 行: "⛔ STOP" → "Help user configure, then proceed"

### P5: 统一版本号（I-05, I-11）

**方案**（D2）: 只保留 frontmatter 版本号，body 和 shared.md 引用 frontmatter

### P6: 验证 flow task 命令可用性（I-15）

**操作**: 检查所有文档中引用的 `flow task` 命令是否都在 commands.md 中有定义

### P7: 实现 cr-index 文档内部 INDEX.md 管理（I-06）

**需求来源**: 05-status-line.md 方案 B
**操作**:
- 每个任务目录下创建 INDEX.md（如 `_docs/team-flow/B001/INDEX.md`）
- 全局 INDEX.md 放 `.team/INDEX.md`（或内部文档路径）
- 表格格式: `| # | Role | Action | Phase | Timestamp |`
- 当前 cr-index = 表格最后一行的 # 值
- AI 每次更新 Status Line 时同步更新 INDEX.md
**后期迁移**: 方案 C — flow 自动管理 cr-index

### P8: 实现技能系统（I-07）

**需求来源**: 03-role-system.md
**操作**:
- 创建 Skill Profile 机制：从 project.md Tech Stack 动态构建技能档案
- 创建 Skill Files：每个技术栈对应一个技能文件（如 `skills/go-ent.md`、`skills/react-bun.md`）
- 技能文件格式兼容 Skill tool（D4），支持两种加载方式
- 创建 Skill Mapping：project.md 中声明技术栈→技能文件的映射
- 实现自动加载：Triage 分发时自动加载对应技能文件注入子 Agent prompt
- 实现 Skill Learning：技能文件缺失时触发学习模式（研究→生成草稿→用户确认→保存复用）
**技能文件目录**: `.trae/skills/team-flow/skills/`（与现有 SKILL 同级，兼容 Skill tool 加载）

### P9: 实现 Triage 中间人原则（I-08）

**需求来源**: 04-task-lifecycle.md
**操作**:
- 修改 SKILL.md 和 triage.md，明确 `用户 ←→ Triage ←→ 子 Agent` 通信模式
- Triage 在每个阶段保持参与（接收→补全→分类→分发→监控→汇报→路由→归档）
- 禁止用户直接和子 Agent 对话

### P10: 升级 project.md 为强制性规则配置（I-09）

**需求来源**: 01-core-requirements.md
**操作**:
- project.md 新增 MANDATORY 标记的 Tech Stack 和 Toolchain 字段
- 实现 PRE-FLIGHT 检查：AI 执行命令前必须输出确认使用的工具链
- 修改 SKILL.md 中 project.md 的读取逻辑：从"了解信息"改为"加载规则"

### P11: 实现门禁式流程控制（I-10）

**需求来源**: 01-core-requirements.md
**操作**:
- 将 SKILL.md 和 BOUNDARY.md 中的"不应该做"列表替换为门禁表
- 门禁表格式: `操作 | 前置条件 | 不满足时`
- 每个关键操作（创建任务、分发、执行命令、修改文件）都有明确前置条件

### P12: 实现 Triage 需求补全 + 反馈路由决策树（I-14）

**需求来源**: 04-task-lifecycle.md
**操作**:
- triage.md 新增需求补全步骤：目标项目、技术栈、影响范围、验收标准、边界条件
- triage.md 新增反馈路由决策树：5 种类型（A:代码Bug / B:需求理解偏差 / C:需求变更 / D:新需求 / E:验收通过）
- 每种类型有明确的路由目标（Dev / Triage / 新建Task / 归档）

## 执行顺序

```
Phase 1: 基础修复（P1, P4, P5, P6）
  └── 修复明显的错误和不一致

Phase 2: 结构统一（P2, P3）
  └── 统一规则定义位置和目录结构

Phase 3: 核心机制（P7, P8, P9, P10, P11, P12）
  └── 实现需求文档中定义的新机制
      ├── P7: cr-index INDEX.md 管理
      ├── P8: 技能系统
      ├── P9: Triage 中间人原则
      ├── P10: project.md 强制规则
      ├── P11: 门禁式流程控制
      └── P12: 需求补全 + 反馈路由

Phase 4: 验证
  └── 用 skill-creator 评估改进后的 team-flow
```