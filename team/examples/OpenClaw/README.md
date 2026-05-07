# OpenClaw 配置示例

> 基于 Trae 工作流规范适配 OpenClaw 环境
> 版本: v2.0 | 目标: 多 Agent 协作执行 CMS 项目任务

---

## 目录结构

```
examples/OpenClaw/
├── CONFIG.md    ← 环境配置（toolchain 对照、子 Agent 分发规范）
└── README.md    ← 本文件（快速开始、核心约束）
```

> 各角色 prompt 在 `{TEAM_PATH}/prompts/` 目录下，OpenClaw 运行时通过 `sessions_spawn` 加载。

---

## 快速开始

### 1. 初始化工作区

启动 OpenClaw 后，Agent 会自动加载 workspace context：
- `SOUL.md` → Agent 人设
- `AGENTS.md` → 工作区规则（含层边界）
- `MEMORY.md` → 长期记忆
- `memory/YYYY-MM-DD.md` → 今日记忆

### 2. 分发任务

所有用户输入经 Triage 分类后，通过 `sessions_spawn` 启动对应子 Agent：

```bash
# 示例：用户提交 Bug
# → Triage 分析 → sessions_spawn(bugfix, task: "修复 B019-R1")
```

### 3. 查看任务池

```
{PROJECT_PATH}/.team/task-pool.md
```

---

## 与 Trae 的主要差异

| 方面 | Trae | OpenClaw |
|------|------|---------|
| 子 Agent 工具 | `Task` | `sessions_spawn`（run/session 双模式）|
| 项目配置 | 内联在 prompt 中 | workspace context 已自动加载 |
| 历史记忆 | 无内置 | `MEMORY.md` + `memory/` |
| 定时任务 | 无 | `cron` |
| 消息推送 | 无 | `message` |
| 设备控制 | 无 | `nodes`（手机通知、拍照等）|
| 心跳检查 | 无 | `HEARTBEAT.md` |

---

## 工具链对照（必读）

每个 Agent prompt 中的工具适配对照：

| Trae | OpenClaw | 用途 |
|------|----------|------|
| `Task` 工具 | `sessions_spawn` | 启动子 Agent |
| `shell` / `Bash` | `exec` | 执行 Shell 命令 |
| 无 | `memory_search` / `memory_get` | 查询历史上下文 |
| 无 | `lcm_grep` / `lcm_expand` | 对话历史检索 |
| 无 | `cron` | 设置定时任务 |
| 无 | `message` | 跨渠道消息发送 |

---

## 核心约束（所有 Agent 必须遵守）

### 层边界
- `{TEAM_PATH}/` = 多项目框架层，**只读**，禁止写入
- `{PROJECT_PATH}/.team/` = 当前项目层，可读写

### ID 命名
- `F{NNN}` = Feature / `B{NNN}` = Bug / `C{NNN}` = Chore / `D{NNN}` = Docs / `A{NNN}` = Analysis
- Bug 修复用 R 后缀：`B019-R1`（第一轮尝试）→ `B019-R2`（第二轮）
- task-pool 用基础 ID，资产目录用 R 后缀

### 禁止项（所有 Agent）
- ❌ 中文注释
- ❌ 跳过 RCA / SCOPE / 验收标准
- ❌ 为同一 Bug 创建新 ID（B019 修复失败只能 B019-R2，不能 B026）
- ❌ 直接写 `{PROJECT_PATH}/_docs/`（正确路径是 `{DOCS_INTERNAL}/`）

---

### 状态检查

```bash
# 验证 task-pool 位置
Get-Content "{PROJECT_PATH}/.team/task-pool.md" -Head 5
```

---

## 相关文档

- Trae 配置示例: `examples/Trae/`（包含 11 个角色 agent prompt）
- 框架层规则: `{TEAM_PATH}/BOUNDARY.md`
- Bug 修复规范: `{TEAM_PATH}/workflows/shared.md`
- 项目配置: `{PROJECT_PATH}/.team/project.md`