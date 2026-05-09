# Flow Task 命令参考

## 统一命令接口

`flow task` 是 AI 统一使用的命令接口。根据 `.team/version` 自动路由：

| Stage | .team/version | 后端 |
|-------|---------------|------|
| v1 | 1 | task-pool.md (文档) |
| v2 | 2 | beads (.beads/) |
| v3 | 3 | 可配置 (beads/git/other) |

AI 始终使用 `flow task`，无需关心底层实现。

## 任务生命周期命令

```bash
flow task create "Title" -t {bug|feature|task|epic} -p {0-4} --parent {id} --labels "label1,label2" --silent
flow task update {id} --claim|--notes "..."|--add-label key:value|--remove-label key:value
flow task close {id} --reason "..."
flow task list [--status open|closed|all] [--json] [--format table] [--created-after YYYY-MM-DD] [--updated-after YYYY-MM-DD]
flow task show {id}
flow task ready [--json]
flow task append {id} --speaker {role} --content "..."
```

| 参数 | 说明 |
|------|------|
| `-t` | 类型：bug / feature / task / epic |
| `-p` | 优先级：0（紧急）→ 4（低） |
| `--parent` | 父任务 ID，用于子任务关联 |
| `--labels` | 逗号分隔标签，如 `phase:implement,team:backend` |
| `--silent` | 静默创建，不触发通知 |
| `--claim` | 认领任务，原子操作 |
| `--notes` | 添加备注 |
| `--add-label` | 添加标签 |
| `--remove-label` | 移除标签 |
| `--reason` | 关闭原因 |
| `--status` | 筛选状态：open / closed / all |
| `--json` | JSON 格式输出 |
| `--format` | 输出格式：table |
| `--created-after` | 筛选创建日期之后（YYYY-MM-DD 或 RFC3339） |
| `--created-before` | 筛选创建日期之前 |
| `--updated-after` | 筛选更新日期之后 |
| `--updated-before` | 筛选更新日期之前 |
| `--closed-after` | 筛选关闭日期之后 |
| `--closed-before` | 筛选关闭日期之前 |
| `--sort` | 排序字段：priority / created / updated / closed / status |
| `--speaker` | 对话记录发言者角色 |
| `--content` | 对话记录内容 |

## 自动时间戳

`flow task` 自动管理时间戳，AI 无需手动写入：

| 操作 | 自动添加 |
|------|----------|
| `flow task create` | `created_at` |
| `flow task update` | `updated_at` |
| `flow task append` | `timestamp` + `cr-index` 自增 |

## 依赖管理命令

```bash
flow task dep add {id} {depends-on|blocks} {target-id}
flow task dep remove {id} {depends-on|blocks} {target-id}
flow task dep list {id}
```

| 关系类型 | 说明 |
|----------|------|
| `depends-on` | 当前任务依赖目标任务 |
| `blocks` | 当前任务阻塞目标任务 |

## 数据命令

```bash
flow task dolt push
flow task dolt pull
flow export > .team/task-pool-export.md
```

| 命令 | 说明 |
|------|------|
| `dolt push` | 推送数据到远程仓库 |
| `dolt pull` | 拉取远程数据 |
| `flow export` | 导出任务池为 Markdown 文件 |

## 路径解析命令

```bash
flow config paths --json
flow config paths --name {variable}
flow config paths --validate
```

| 参数 | 说明 |
|------|------|
| `--json` | 以 JSON 格式输出所有路径变量 |
| `--name` | 查询指定路径变量值 |
| `--validate` | 验证所有路径是否存在且可访问 |

## 快速参考

| 操作 | 命令 |
|------|------|
| 创建任务 | `flow task create "Title" -t {type} -p {0-4}` |
| 认领任务 | `flow task update {id} --claim` |
| 更新阶段 | `flow task update {id} --add-label phase:xxx` |
| 关闭任务 | `flow task close {id} --reason "..."` |
| 查看任务 | `flow task show {id}` |
| 查找工作 | `flow task ready --json` |
| 列出任务 | `flow task list --status open --format table` |
| 按日期筛选 | `flow task list --created-after 2026-05-09 --json` |
| 按更新筛选 | `flow task list --updated-after 2026-05-09 --sort updated` |
| 追加对话 | `flow task append {id} --speaker {role} --content "..."` |
| 添加依赖 | `flow task dep add {id} depends-on {target-id}` |
| 推送数据 | `flow task dolt push` |
| 导出任务 | `flow export > .team/task-pool-export.md` |
| 解析路径 | `flow config paths --json` |
