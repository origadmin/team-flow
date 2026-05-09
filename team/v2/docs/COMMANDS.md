# Flow Tools Beads 命令参考

## 任务生命周期命令

```bash
flow tools beads create "Title" -t {bug|feature|task|epic} -p {0-4} --parent {id} --labels "label1,label2" --silent
flow tools beads update {id} --claim|--notes "..."|--add-label key:value|--remove-label key:value
flow tools beads close {id} --reason "..."
flow tools beads list [--status open|closed|all] [--json] [--format table]
flow tools beads show {id}
flow tools beads ready [--json]
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

## 依赖管理命令

```bash
flow tools beads dep add {id} {depends-on|blocks} {target-id}
flow tools beads dep remove {id} {depends-on|blocks} {target-id}
flow tools beads dep list {id}
```

| 关系类型 | 说明 |
|----------|------|
| `depends-on` | 当前任务依赖目标任务 |
| `blocks` | 当前任务阻塞目标任务 |

## 数据命令

```bash
flow tools beads dolt push
flow tools beads dolt pull
flow export > .team/task-pool-export.md
```

| 命令 | 说明 |
|------|------|
| `dolt push` | 推送 beads 数据到远程仓库 |
| `dolt pull` | 拉取远程 beads 数据 |
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
| 创建任务 | `flow tools beads create "Title" -t {type} -p {0-4}` |
| 认领任务 | `flow tools beads update {id} --claim` |
| 更新阶段 | `flow tools beads update {id} --add-label phase:xxx` |
| 关闭任务 | `flow tools beads close {id} --reason "..."` |
| 查看任务 | `flow tools beads show {id}` |
| 查找工作 | `flow tools beads ready --json` |
| 列出任务 | `flow tools beads list --status open --format table` |
| 添加依赖 | `flow tools beads dep add {id} depends-on {target-id}` |
| 推送数据 | `flow tools beads dolt push` |
| 导出任务 | `flow export > .team/task-pool-export.md` |
| 解析路径 | `flow config paths --json` |
