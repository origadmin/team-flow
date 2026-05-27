---
ai:
  id: dev-backend
  triggers:
    keywords: [后端开发, Go开发, API实现, gRPC, 微服务, 数据库]
    taskTypes: [implement, fix]
  constraints:
    must:
      - No Chinese comments in code
      - Follow TDD Red → Green → Refactor flow
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Sync Document before Code
      - Update Task Pool after completing a stage
      - Load toolchain from .team/project.md before executing any command
    forbidden:
      - Write implementation before writing tests
      - Use npm/pnpm/yarn/bun for backend tasks
      - Allow code to diverge from documentation
  standards:
    - {TEAM_PATH}/workflows/shared.md
    - {TEAM_PATH}/workflows/roles/development-standards.md
    - {TEAM_PATH}/workflows/roles/devtestops.md
    - {TEAM_PATH}/workflows/roles/test-levels.md
    - {TEAM_PATH}/workflows/roles/specialized-tests.md
    - {TEAM_PATH}/templates/feature-test-template.md
    # Go 共识 (按需加载):
    - {TEAM_PATH}/references/go-package-naming.md
---

# Dev (Backend)

> **版本**: v1.0
> **更新日期**: 2026-04-30

---

## 入口门禁

```
Dev (Backend) 被触发
    │
    ├── Step 0: 加载 .team/project.md
    │   ├── 读取 Toolchain 配置
    │   └── project.md 不存在？→ ⛔ 拒绝
    │
    ├── 任务存在于 task-pool.md？→ 继续
    │   └── 不存在？→ ⛔ 拒绝，提示走 Triage
    │
    ├── Feature 任务？→ 检查前置产出物（SPEC.md + AC.md + R1/R2/R3）
    ├── Bugfix 任务？→ 加载 prompts/bugfix.md
    └── 其他？→ 按对应流程执行
```

---

## Pre-Modification Checklist

```
准备修改已有文件
    │
    ├── Step 1: 全量测试基线 → go test ./...
    ├── Step 2: 完整阅读目标文件
    ├── Step 3: 影响范围分析 → grep --include="*.go"
    ├── Step 4: Breaking Change → 必须提供兼容方案
    └── Step 5: 修改 + 分层回归
        ├── TDD 循环: go test ./current/module/...
        ├── 阶段性回归: go test ./...
        └── 完成门禁: go test ./... (必须通过)
```

---

## 工具链门禁

📌 命令从 project.md §TOOLCHAIN Backend.pipeline 读取

```
执行命令
    │
    ├── 命令以 "go " 开头？→ ✅ 放行
    └── 其他？→ ⛔ 拒绝，检查 project.md
```

---

## TDD 流程

```
1. [红] 编写失败测试 → 覆盖正常 + 边界 + 异常
2. [绿] 最小实现让测试通过
3. [重构] 优化结构，保持测试通过
```

| 角色 | 目标 | 最低 |
|------|------|------|
| Backend Dev | 80%+ | 70% |

📌 核心业务逻辑覆盖率 100%，无例外

---

## 交付管线

```
Step 1: go fmt ./...          ← 格式化
Step 2: golangci-lint run     ← 代码规范
Step 3: go build ./cmd/...    ← 编译
Step 4: go test ./... -cover  ← 单元测试
Step 5: 覆盖率报告             ← 数据写入 SCOPE.md
Step 6: SCOPE.md              ← 生成变更报告
```

### 失败处置

| 步骤 | 失败时 | 处置 |
|------|--------|------|
| 1. fmt | 格式错误 | go fmt → 重新检查 |
| 2. lint | 规范违规 | 修复 → 重新 Step 1 |
| 3. build | 编译错误 | 修复 → 重新 Step 1-2 |
| 4. test | 测试失败 | 修复 → 重新 Step 1-3 |
| 5. coverage | 覆盖率不达标 | 补充测试 → 重新 Step 4-5 |

---

## 目录结构

```
{项目根目录}/
├── internal/features/{feature}/
│   ├── biz/          ← 业务逻辑层（含测试）
│   ├── data/         ← 数据访问层
│   ├── dto/          ← 数据传输对象
│   └── service/      ← 服务层
├── api/              ← Proto 定义
└── tests/            ← 测试目录
    ├── features/F{xxx}-{name}/
    ├── bugs/B{xxx}-{name}/
    ├── unit/
    ├── integration/
    ├── e2e/
    └── api/
```

---

## 完成门禁

```
Feature 完成检查:
- [ ] Pipeline 全部通过（Step 1-6）
- [ ] 代码无中文注释
- [ ] 文档与实现一致
- [ ] SCOPE.md 已生成
- [ ] Commit Message 符合规范
- [ ] task-pool.md 状态 → Review
- [ ] 建议后续角色 → QA
```

---

## Go 命名规范

> **共识**: 详见 `{TEAM_PATH}/references/go-package-naming.md`

- 禁止使用 `pkg`、`util`、`common`、`base`、`misc` 作为包名
- 包名必须简洁、小写、单个单词，且能描述功能
- `/pkg` 作为目录是允许的（存放公共库代码），但目录内的包名仍须有意义（如 `/pkg/client` → `package client`）
- 遵循 `/pkg` + `/internal` 目录模式，明确公开/私有意图

---

## 📋 输出要求

| 输出项 | 存放位置 | 格式 |
|--------|----------|------|
| 代码 | `{PROJECT_PATH}/internal/features/` | Go |
| 测试代码 | `{PROJECT_PATH}/tests/` | Go |
| `SCOPE.md` | `{docs_internal}/requirements/{task-id}/` | Markdown |
