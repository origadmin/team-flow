# team-flow v3 Workflow Milestone

> **Created**: 2026-05-08
> **Updated**: 2026-05-08
> **Status**: In Discussion

## 概述

v3 是 team-flow 的重大升级，核心变化是**从静态规则文本转向可视化流程配置**。

---

## 核心设计决策

### 1. 分层架构

| 文件 | 位置 | 谁改 | 职责 |
|------|------|------|------|
| SKILL.md | skill 包 | ❌ skill 开发者 | AI 入口 + 核心规则 |
| workflows/* | skill 包 | ❌ skill 开发者 | 默认模板（参考用） |
| rules/* | .team/ | ✅ 项目所有者 | 项目定制规则 |
| workflow.json | .team/ | ✅ 项目所有者 | 流程配置 |

### 2. 规则管理

```
init 时：
  embed/workflows/rules/* → .team/rules/*（全量复制）

运行时：
  AI 读取 .team/rules/*（项目定制）
  若不存在 → 警告 + 建议 'flow repair'
```

### 3. 规则分类模板

```
.team/rules/
├── templates/              # 模板分类
│   ├── dispatch-guard.md   # 原始模板（来自 embed）
│   ├── regression-guard.md
│   └── ...
└── custom/                # 用户自定义
    ├── my-custom-rule.md
    └── ...

修复逻辑：
  用户删除 dispatch-guard.md → 从 templates/ 恢复
  用户重命名为 my-custom-rule.md → 不恢复（视为有意修改）
```

### 4. workflow.json 结构

```json
{
  "version": "v3",
  "template": "default",
  "rules": [
    "dispatch-guard",
    "regression-guard"
  ],
  "nodes": [
    {
      "id": "triage",
      "type": "phase",
      "role": "Triage",
      "standards": ["shared.md"],
      "gates": ["entry"]
    },
    {
      "id": "analyze",
      "type": "phase",
      "role": "TechLead",
      "standards": ["shared.md", "architecture-standards.md"]
    },
    {
      "id": "gate-review",
      "type": "gate",
      "conditions": ["tests_pass", "no_regressions"],
      "on_fail": "defect-new",
      "on_pass": "phase-deploy"
    }
  ],
  "edges": [
    {"from": "triage", "to": "gate-entry"},
    {"from": "gate-entry", "to": "analyze"},
    {"from": "analyze", "to": "gate-review"}
  ]
}
```

### 5. 节点类型

| 节点类型 | 说明 | v2 对应 |
|---------|------|--------|
| phase | 阶段节点 | Phase Gate |
| gate | 条件判断 | Three-Layer Gates |
| branch | 分支处理 | 无 |
| defect | 缺陷跟踪 | Bugfix 流程 |
| retry | 自动重试 | 无 |
| subflow | 子流程 | 无 |

### 6. Skill 包目录结构（符合 agentskills.io 标准）

```
team-flow/
├── SKILL.md              # skill 入口（npx skills 标准）
├── workflows/
│   ├── rules/           # 默认规则模板
│   │   ├── dispatch-guard.md
│   │   ├── regression-guard.md
│   │   └── ...
│   ├── templates/       # 流程模板
│   │   ├── default.json
│   │   ├── bugfix.json
│   │   └── ...
│   └── ...
└── ...

安装方式：
  npx skills add origadmin/team-flow
  或
  flow install
```

### 7. 版本策略

```
默认版本：v3

flow init              # 默认创建 v3
flow init --v1         # 创建 v1（兼容旧项目）
flow init --v2         # 创建 v2（兼容旧项目）

flow migrate           # 升级到 v3
flow migrate --to-v1   # 降级到 v1（不推荐）
flow migrate --to-v2   # 升级到 v2
```

### 8. 迁移路径

| 源版本 | 目标 | 操作 |
|--------|------|------|
| 无 | v3 | `flow init` |
| v1 | v3 | `flow migrate` (v1→v2→v3 或直接) |
| v2 | v3 | `flow migrate` |

---

## 待讨论问题

### 已解决

- [x] SKILL.md 定位：skill 开发者专用
- [x] rules/* 位置：`.team/rules/`
- [x] 规则存储：init 时从 embed 全量复制到 .team/
- [x] 丢失处理：`flow repair` 显式修复
- [x] 规则分类：templates/（模板）+ custom/（自定义）
- [x] 默认版本：v3
- [x] 迁移支持：v1→v3, v2→v3
- [x] skill 包结构：符合 agentskills.io 标准

### 待讨论

- [ ] flow repair 详细设计
- [ ] flow init --v3 详细设计
- [ ] 流程图格式：JSON vs YAML
- [ ] flow edit 编辑器方案

---

## CLI 命令设计

```bash
# 初始化
flow init                 # 默认 v3
flow init --v1           # v1
flow init --v2           # v2
flow init --v3 --template default  # v3 + 指定模板

# 迁移
flow migrate             # 升级到 v3
flow migrate --to-v2    # 升级到 v2

# 编辑
flow edit                # 编辑 workflow.json
flow edit rules <name>  # 编辑特定规则

# 修复
flow repair             # 列出缺失文件
flow repair --rules     # 恢复缺失的规则文件
flow repair --apply     # 确认后恢复

# 流程模板
flow template list      # 列出可用模板
flow template use xxx   # 使用指定模板
```

---

## 文件结构

### Skill 包（team-flow/）

```
team-flow/                  # 符合 agentskills.io 标准
├── SKILL.md              # AI 入口
├── workflows/
│   ├── rules/           # 默认规则模板
│   │   ├── dispatch-guard.md
│   │   ├── regression-guard.md
│   │   └── ...
│   ├── templates/       # 流程模板
│   │   ├── default.json
│   │   ├── bugfix.json
│   │   └── ...
│   └── prompts/         # v2/v3 角色定义
│       ├── triage.md
│       ├── dev.md
│       └── ...
```

### CLI（cmd/flow/）

```
cmd/flow/
└── main.go

internal/
├── init/                  # flow init
├── migrate/              # flow migrate
├── repair/              # flow repair
├── edit/                 # flow edit
├── template/            # flow template
└── workflow/            # v3 流程引擎
```

### 项目（.team/）

```
项目/.team/
├── version               # v1 / v2 / v3
├── workflow.json         # 流程配置（v3）
└── rules/               # 项目规则（从 embed 复制）
    ├── templates/        # 模板规则（可恢复）
    │   ├── dispatch-guard.md
    │   └── ...
    └── custom/          # 自定义规则（不恢复）
        └── my-rule.md
```

---

## 开放问题

1. **workflow.json 格式**：JSON（机器友好）还是 YAML（人类友好）？
2. **flow edit 编辑器**：终端交互式、TUI、还是外部编辑器？
3. **节点配置项**：每个节点的完整配置字段有哪些？
4. **AI 如何读取 workflow.json**：需要专门的解析器吗？

---

## 参考

- [Agent Skills 规范](https://agentskills.io/)
- [vercel-labs/skills](https://github.com/vercel-labs/skills)
- v2 SKILL.md: `team/SKILL.md`
- v2 BOUNDARY.md: `team/v2/BOUNDARY.md`
- team-flow CLI: `cmd/flow/`
