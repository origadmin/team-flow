# PRE-FLIGHT — 执行任何操作前必须完成

> **TEAM_VERSION=7.1** | **更新日期**: 2026-04-27
> **定位**: 所有角色的强制前置步骤，嵌入每个 prompt 的工作流 Step 0

---

## ⛔ 为什么必须有 PRE-FLIGHT

AI 的训练数据充斥着 npm/npx 等默认命令。当任务压力下，AI 会凭肌肉记忆直接执行，跳过"先读配置文件"的步骤。

**PRE-FLIGHT 不是建议，是硬门禁。跳过 = 任务失败。**

---

## Step 0: PRE-FLIGHT 检查清单

**每个角色被触发后，执行任何操作前，必须按序完成以下步骤：**

### 0.1 读取项目配置

```
读取 .team/project.md
    │
    ├── §TOOLCHAIN → 提取包管理器和 pipeline 命令
    ├── §CONSTRAINTS → 提取项目约束
    ├── §PATHS → 提取文档路径
    └── §GIT → 提取 Git 权限
```

**产出**: 在本次会话上下文中记录以下变量：

```yaml
# 从 project.md 读取后记录
TOOLCHAIN:
  frontend:
    package_manager: {从 §TOOLCHAIN.Frontend 读取}
    pipeline: {从 §TOOLCHAIN.Frontend.pipeline 读取}
  backend:
    pipeline: {从 §TOOLCHAIN.Backend.pipeline 读取}
CONSTRAINTS: {从 §CONSTRAINTS 读取}
PATHS:
  docs_external: {从 §PATHS 读取}
  docs_internal: {从 §PATHS 读取}
```

### 0.2 工具链校验（Toolchain Gate）

**每次执行 shell 命令前，必须通过此校验：**

```
准备执行命令
    │
    ├── 命令属于前端任务？
    │   ├── 命令开头匹配 TOOLCHAIN.frontend.package_manager？→ ✅ 放行
    │   └── 不匹配？
    │       ├── 发现 npm/pnpm/yarn → ⚠️ STOP
    │       │   提示: "本项目前端使用 {package_manager}，禁止 {detected}"
    │       │   修正: 使用 {package_manager} 替代
    │       └── 其他未知命令 → 检查 TOOLCHAIN.frontend.pipeline 是否有对应条目
    │
    └── 命令属于后端任务？
        ├── 命令开头匹配 TOOLCHAIN.backend.pipeline？→ ✅ 放行
        └── 不匹配 → ⚠️ STOP，修正命令
```

**具体拦截规则**（以 bun 项目为例）：

| 输入 | 判定 | 正确替代 |
|------|------|---------|
| `npm install` | ⚠️ | `bun install` |
| `npm run build` | ⚠️ | `bun run build` |
| `npx vite build` | ⚠️ | `bun run build`（本项目使用 Rsbuild，禁止 Vite） |
| `pnpm add lodash` | ⚠️ | `bun add lodash` |
| `yarn dev` | ⚠️ | `bun run dev` |
| `bun run build` | ✅ | — |
| `go build ./cmd/...` | ✅ | — |

### ⚠️ PowerShell 文本替换危险警告

> **Windows 平台专属** | 影响所有文本文件写入

PowerShell 管道写入文件时，会随机破坏文件编码（UTF-8 BOM 乱加 / CRLF 错位 / 内容截断），在 Windows 上是系统性缺陷。

**危险操作（禁止）：**
- `Set-Content` / `Out-File` / `Add-Content`（任何带 `-Encoding utf8` 的写法）
- `| ForEach-Object { $_ -replace ... }`
- `$content -replace 'xxx' | Set-Content`
- `Get-Content | % { $_ -replace ... } | Set-Content`

**安全替代：**
- Python: `{python} scripts/write_file.py --path file.txt --content "$content"`
- Node.js: `node -e "require('fs').writeFileSync(...)"`
- PowerShell 读取（安全） + Python/Node 写入（安全）

**读取文件是安全的**，只有写入才危险。

### 0.3 项目约定校验

```
读取 {docs_internal}/conventions/common.md（如存在）
读取 {docs_internal}/conventions/dev-common.md（如存在）
读取 {docs_internal}/conventions/dev-{subtype}.md（如存在）
    │
    └── 记录约定要点到上下文
        ├── URL 前缀约定
        ├── 命名规范
        ├── 模块边界
        └── 错误码范围
```

### 0.4 PRE-FLIGHT 完成标志

**PRE-FLIGHT 完成后，必须输出以下确认（简短，不浪费 token）：**

```
✅ PRE-FLIGHT: pkg={package_manager} | constraints={已读} | conventions={已读/无}
```

**禁止在 PRE-FLIGHT 完成前执行任何 shell 命令或代码修改。**

### 0.5 ⏸️ 暂停点 — 暴露假设

> 执行前，快速确认以下三点（不必写出来，心里有数即可）：
- **假设**：我正在基于哪些未确认的假设行动？
- **风险**：如果这个假设错了，最坏的结果是什么？
- **澄清**：有没有我应该先问清楚的点？

> 如有任何不确定 → ⚠️ 停止，先提问，不要凭记忆继续。

---

## 各角色 PRE-FLIGHT 差异

| 角色 | 需要读 §TOOLCHAIN | 需要读 §CONSTRAINTS | 需要读 conventions | 额外步骤 |
|------|:-:|:-:|:-:|------|
| Dev | ✅ | ✅ | ✅ | — |
| QA | ✅ | ✅ | ✅ | — |
| Tech Lead | ❌ | ✅ | ✅ | — |
| PM | ❌ | ✅ | ❌ | — |
| Triage | ❌ | ❌ | ❌ | 读取 §ROLES + §PATHS |
| DevOps | ✅(Backend) | ✅ | ❌ | — |
| UI Designer | ❌ | ✅ | ✅ | 读取 design/tokens.md |
| Analysis | ❌ | ✅ | ❌ | — |
| Framework Architect | ❌ | ✅ | ❌ | — |

> ❌ = 该角色不需要读该节，跳过以节省 token

---

## 维护职责

**PRE-FLIGHT 本身由 Framework Architect 维护。当新增角色或 project.md 结构变更时，需更新本文件的角色差异表。**

---

