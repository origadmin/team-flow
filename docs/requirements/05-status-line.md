# Status Line 规则

> **版本**: v1.3 | **日期**: 2026-05-11 | **状态**: 待确认

## 1. 格式

```
[Role: {role} | TaskPool: {beads-id}#{cr-index} | Phase: {phase} | Asset: {project-name}]
```

## 2. 字段定义

| 字段 | 值 | 说明 |
|------|-----|------|
| Role | Triage / TechLead / Dev / QA / PM / DevOps / Analysis / UIDesigner | 子 Agent 执行时变化，完成后回到 Triage |
| TaskPool | `{beads-id}#{cr-index}` | beads ID + 变更轮次索引 |
| Phase | ready / analyze / design / implement / verify / review / -(N/A) | 当前阶段 |
| Asset | 目标项目名 | 当前操作的项目 |

## 3. beads ID 格式

beads 创建任务时自动生成 ID：

| ID | 结构 | 说明 |
|----|------|------|
| `framework-l9v` | `{prefix}-{random}` | 独立任务或 epic |
| `framework-l9v.1` | `{parent-id}.{N}` | 父任务的第 N 个子任务（beads 自动编号） |
| `framework-9gy` | `{prefix}-{random}` | 独立任务（无父子关系） |

**`.N` 是 beads 自动生成的子任务序号**，不是 cr-index。

## 4. cr-index 管理

### cr-index 是什么

cr-index（Change Request index）跟踪同一任务的变更轮次。每次有意义的操作，cr-index +1。

### 三种管理方案对比

| 方案 | 机制 | 优点 | 缺点 |
|------|------|------|------|
| A: beads labels | `flow task update --add-label cr-index:N` | 数据库管理，可靠 | 每次操作需额外命令，`flow` 可能不支持 label 覆盖 |
| B: 文档内部 Index | 任务文档中维护 `cr-index` 字段 | 简单，不需要额外命令，可追溯 | 依赖 AI 遵守流程 |
| C: `flow` 自动管理 | 修改 `flow` 源码自动递增 | 最可靠 | 需要改源码，工作量大 |

### 推荐方案：B（文档内部 Index）

**理由**：
1. 每个任务本来就有对应的文档目录（如 `_docs/orig-cms-ee/F032/`）
2. 文档里加一个 Index 字段，AI 每次操作时更新
3. 不需要额外命令，只需要 AI 遵守流程
4. 文档本身就是可追溯的——比 beads label 更直观
5. 不依赖 `flow` 是否支持 label 覆盖

### 文档内部 Index 实现

每个任务的文档目录中有一个 `INDEX.md`（或任务主文档中包含 Index 部分）：

```markdown
# Task: F032 Notification System Design

## Index

| # | Role | Action | Phase | Timestamp |
|---|------|--------|-------|-----------|
| 1 | Triage | Created task | ready | 2026-05-11 10:00 |
| 2 | Triage | Dispatched to Tech Lead | design | 2026-05-11 10:05 |
| 3 | TechLead | Completed design | design | 2026-05-11 11:30 |
| 4 | Triage | Reviewed, dispatched to Dev | implement | 2026-05-11 11:35 |
| 5 | Dev | Implemented feature | implement | 2026-05-11 14:00 |
```

**当前 cr-index = 表格最后一行的 # 值**。

### cr-index 递增规则

| 事件 | cr-index 变化 | AI 操作 |
|------|-------------|---------|
| Triage 创建任务 | 初始化为 1 | 在文档中添加 Index 表，写入第 1 行 |
| Triage 分发到子 Agent | +1 | 在 Index 表追加新行 |
| 子 Agent 返回结果 | +1 | 在 Index 表追加新行 |
| Triage 汇报用户 | +1 | 在 Index 表追加新行 |
| 用户反馈，Triage 路由 | +1 | 在 Index 表追加新行 |
| 纯对话（无任务操作） | 不变 | 不更新 Index 表 |

### cr-index 读取

AI 在输出 Status Line 前，读取任务文档的 Index 表，取最后一行的 # 值：

```
1. 读取 {DOCS_INTERNAL}/{task-id}/INDEX.md
2. 找到 Index 表最后一行
3. 取 # 值作为 cr-index
4. 输出 Status Line: [Role: ... | TaskPool: {beads-id}#{cr-index} | ...]
```

**禁止**：AI 不读取文档就自己编 cr-index。

### 迁移路径

| 阶段 | 方案 | 说明 |
|------|------|------|
| 前期（当前） | B: 文档内部 Index | 简单可靠，不依赖 flow 扩展 |
| 后期 | C: flow 自动管理 | 可靠性最高，需要改 flow 源码 |

迁移时只需将 INDEX.md 中的数据导入 flow，AI 读取来源从文档改为 `flow task show`。

### 方案 B 的优势

| 对比项 | beads labels | 文档内部 Index |
|--------|-------------|---------------|
| 可追溯 | 只有数字 | 有完整操作历史 |
| 可读性 | 需要命令查看 | 直接打开文档看 |
| 依赖 | 依赖 `flow` 支持 label | 只依赖文件读写 |
| 维护成本 | 每次操作一条命令 | 每次操作追加一行文字 |
| 扩展性 | 只能存数字 | 可以存 Role、Action、Phase 等 |

## 5. TaskPool 值验证

| TaskPool 值 | 含义 | 状态 |
|-------------|------|------|
| `framework-l9v.2#3` | beads ID + 文档管理的 cr-index（正确） | ✅ |
| `framework-l9v.2` | 缺少 cr-index（不完整） | ⚠️ |
| `framework` | 目录名（错误） | ❌ |
| `-(N/A)` | 未创建 Task（严重违规） | ❌ |
| `framework-l9v.2#1`（AI 自编） | cr-index 未从文档读取（错误） | ❌ |

## 6. TaskPool 与 Asset 的关系（按模式区分）

### standalone 模式

TaskPool 前缀**必须**匹配 Asset。

```
[Role: Triage | TaskPool: orig-cms-ee-9gy#2 | Phase: analyze | Asset: orig-cms-ee]
                         ↑ 前缀匹配 Asset ✅
```

### workspace 模式

TaskPool 前缀是 workspace 名，Asset 是目标项目名。**可以不同**。

```
[Role: Triage | TaskPool: framework-l9v.2#3 | Phase: analyze | Asset: orig-cms-ee]
                         ↑ workspace 名       ↑ 目标项目 ✅
```

| 模式 | TaskPool | Asset | 一致性规则 |
|------|----------|-------|-----------|
| standalone | `{project}-{id}#{cr}` | `{project}` | **前缀必须匹配** |
| workspace | `{workspace}-{id}#{cr}` | `{target-project}` | **可以不同** |

## 7. Phase 与 TaskPool 同步

Phase 是 `analyze`/`design`/`implement` 等时，TaskPool 必须显示 beads ID + cr-index，禁止显示 N/A。

## 8. Task-first 规则

Session 启动时，Triage 必须先创建/查找任务。每次对话都有 TaskPool 值，N/A 是被禁止的。
