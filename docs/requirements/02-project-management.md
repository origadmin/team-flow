# 项目管理模式

> **版本**: v1.1 | **日期**: 2026-05-11 | **状态**: 待确认

## 1. 两种模式

| 模式 | 说明 | beads 位置 | 典型场景 |
|------|------|-----------|---------|
| **standalone** | 单项目，独立管理 | 项目自己的 `.beads/` | `orig-cms-ee/` 独立运作 |
| **workspace** | 多项目统一管理 | workspace 的 `.beads/`（唯一） | `framework/` 统管 `orig-cms-ee`、`abgen` 等 |

## 2. 模式判定

**配置驱动**：AI 读取 `.team/project.md` 的 `Management Mode` 字段。

```yaml
## Management Mode
- **Mode**: workspace | standalone
```

| Mode 值 | AI 行为 |
|---------|--------|
| `standalone` | 所有操作在当前目录，不切换 |
| `workspace` | 从 Registered Projects 表解析目标项目 |
| 未设置 | 帮助用户配置，然后继续 |

**禁止**：AI 通过检查目录结构来"猜测"项目类型。

## 3. workspace 模式详情

### 目录结构

```
framework/                          ← workspace（统一管理）
├── .team/project.md               ← Mode: workspace + Registered Projects
├── .beads/                         ← 唯一的 beads（所有项目任务都在这里）
├── projects/orig-cms-ee/          ← 子项目
│   └── .team/project.md           ← 子项目规则（Tech Stack + Toolchain）
└── tools/abgen/                   ← 子项目
    └── .team/project.md           ← 子项目规则（Tech Stack + Toolchain）
```

**关键**：workspace 模式下只有**一个** `.beads/`，在 workspace 根目录。子项目**没有**自己的 `.beads/`。

### 两个 `.team/project.md` 的职责分工

| 文件 | 职责 | 内容 |
|------|------|------|
| workspace 的 `.team/project.md` | 管理模式 + 项目注册 | Mode、Registered Projects、workspace 级 Tech Stack |
| 子项目的 `.team/project.md` | 子项目规则 | 子项目自己的 Tech Stack、Toolchain、约束 |

**AI 读取顺序**：
1. 先读 workspace 的 `.team/project.md` → 确定 Mode + 目标项目
2. 确定目标项目后 → 读子项目的 `.team/project.md` → 获取 Tech Stack + Toolchain

### Registered Projects 表

在 workspace 的 `.team/project.md` 中声明：

```markdown
## Registered Projects

| Project | Path | Status | Description |
|---------|------|--------|-------------|
| framework | ./ | active | Workspace host |
| orig-cms-ee | ./projects/orig-cms-ee/ | active | CMS business project |
| abgen | ./tools/abgen/ | active | Code generation tool |
```

AI 只能操作 Registered Projects 表中的项目。

### 项目解析规则

1. 读取 Registered Projects 表（已配置，无需用户交互）
2. 将用户输入与项目名和描述匹配
3. 匹配成功 → 设置 Asset = 目标项目名，读取子项目的 `.team/project.md` 获取规则
4. 无匹配 → 默认为 workspace host（path "./"）
5. 不明确 → 询问用户，不猜测

### workspace 模式下的 `flow task` 执行

| 项目 | `flow task` cwd | beads ID 前缀 | Asset |
|------|----------------|--------------|-------|
| framework（workspace host） | `framework/` | `framework-xxx` | `framework` |
| orig-cms-ee（子项目） | `framework/` | `framework-xxx` | `orig-cms-ee` |
| abgen（子项目） | `framework/` | `framework-xxx` | `abgen` |

**关键**：workspace 模式下，所有 `flow task` 命令都在 workspace 根目录执行，因为 `.beads/` 在那里。beads ID 前缀始终是 workspace 名（`framework-xxx`），但 Asset 反映实际操作的目标项目。

## 4. standalone 模式详情

```
orig-cms-ee/                       ← 独立项目
├── .team/project.md               ← Mode: standalone + Tech Stack + Toolchain
├── .beads/                         ← 项目自己的 beads
└── ...
```

每个项目独立运作，有自己的 `.beads/`，不涉及项目切换。

| 项目 | `flow task` cwd | beads ID 前缀 | Asset |
|------|----------------|--------------|-------|
| orig-cms-ee | `orig-cms-ee/` | `orig-cms-ee-xxx` | `orig-cms-ee` |

TaskPool 前缀必须匹配 Asset。

## 5. 模式切换

一个项目可以从 standalone 切换到 workspace（或反之），但需要：

1. 用户明确决定管理模式
2. 更新 `.team/project.md` 的 Mode 字段
3. 如果从 standalone → workspace：子项目的 `.beads/` 数据需要迁移到 workspace 的 `.beads/`
4. 如果从 workspace → standalone：需要为子项目创建独立的 `.beads/`

**AI 不应该自动切换模式**——这是用户的决定。

## 6. 跨项目隔离规则

| 规则 | 说明 |
|------|------|
| 禁止猜测路径 | 不在 Registered Projects 表中的项目不能操作 |
| 禁止启发式检测 | 不通过检查 `.beads/`、`.trae/` 等判断项目类型 |
| 禁止交叉写入 | workspace 模式下所有任务在同一个 `.beads/`，但 Asset 必须反映正确目标 |
| 配置持久化 | 配置写在 `.team/project.md`，AI 每次启动自动读取 |
| 子项目规则独立 | 每个子项目的 Tech Stack + Toolchain 独立配置，互不影响 |
