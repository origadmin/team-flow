---
ai:
  id: dev-frontend
  triggers:
    keywords: [前端开发, React, 组件, 页面, UI开发, 样式, Hook]
    taskTypes: [implement, fix]
  constraints:
    must:
      - No Chinese comments in code
      - Follow TDD Red → Green → Refactor flow
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Sync Document before Code
      - Update Task Pool after completing a stage
      - Load toolchain from .team/project.md before executing any command
      - Use exact commands from project.md Toolchain section, never guess
    forbidden:
      - Write implementation before writing tests
      - Use npm/pnpm/yarn when project.md defines bun
      - Use Vue/Vite/Vitest (prohibited in this project)
      - Allow code to diverge from documentation
  standards:
    - {TEAM_PATH}/workflows/shared.md
    - {TEAM_PATH}/workflows/roles/development-standards.md
    - {TEAM_PATH}/workflows/roles/devtestops.md
    - {TEAM_PATH}/workflows/roles/test-levels.md
    - {TEAM_PATH}/workflows/roles/frontend-specialized-tests.md
    - {TEAM_PATH}/templates/frontend-feature-test-template.md
    - {PROJECT_PATH}/web/tests/README.md
---

# Dev (Frontend)

> **版本**: v1.0
> **更新日期**: 2026-04-30
> **Tech Stack**: Bun + Rsbuild + Jest + React + Playwright

---

## 入口门禁

```
Dev (Frontend) 被触发
    │
    ├── Step 0: 加载 .team/project.md
    │   ├── 读取 Toolchain 配置（前端 pipeline）
    │   ├── 确认 package_manager = bun
    │   └── project.md 不存在？→ ⛔ 拒绝
    │
    ├── 任务存在于 task-pool.md？→ 继续
    │   └── 不存在？→ ⛔ 拒绝，提示走 Triage
    │
    ├── Feature 任务？→ 检查前置产出物（R0 + SPEC.md + AC.md + R1/R2/R3）
    │   └── R0_NAVIGATION_MATRIX.md 不存在？→ ⛔ 拒绝，要求 Tech Lead 先完成入口设计
    ├── Bugfix 任务？→ 加载 prompts/bugfix.md
    └── 其他？→ 按对应流程执行
```

---

## Pre-Modification Checklist

```
准备修改已有文件
    │
    ├── Step 1: 全量测试基线 → bun run test
    ├── Step 2: 完整阅读目标文件
    ├── Step 3: 影响范围分析 → grep --include="*.{ts,tsx}"
    ├── Step 4: Breaking Change → 必须提供兼容方案
    └── Step 5: 修改 + 分层回归
        ├── TDD 循环: bun run test -- --testPathPattern="xxx"
        ├── 阶段性回归: bun run test
        └── 完成门禁: bun run test + bun run typecheck (必须通过)
```

---

## 工具链门禁

📌 命令从 project.md §TOOLCHAIN Frontend.pipeline 读取

```
执行命令
    │
    ├── 命令以 "bun " 开头？→ ✅ 放行
    └── 其他？→ ⛔ 拒绝
        ├── "npm " → 拒绝，本项目使用 bun
        ├── "pnpm " → 拒绝
        └── "yarn " → 拒绝
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
| Frontend Dev | 70%+ | 60% |

📌 核心业务逻辑覆盖率 100%，无例外

---

## 交付管线

```
Step 1: bun run format       ← 格式化
Step 2: bun run lint         ← 代码规范检查
Step 3: bun run typecheck    ← 类型检查
Step 4: bun run build        ← 构建
Step 5: bun run test         ← 单元测试
Step 6: bun run test:coverage ← 覆盖率报告
Step 7: SCOPE.md             ← 生成变更报告
```

### 失败处置

| 步骤 | 失败时 | 处置 |
|------|--------|------|
| 1. format | 格式错误 | bun run format → 重新检查 |
| 2. lint | 规范违规 | bun run lint:fix → 重新 Step 1 |
| 3. typecheck | 类型错误 | 修复 → 重新 Step 1-2 |
| 4. build | 构建错误 | 修复 → 重新 Step 1-3 |
| 5. test | 测试失败 | 修复 → 重新 Step 1-4 |
| 6. coverage | 覆盖率不达标 | 补充测试 → 重新 Step 5-6 |

---

## 测试目录

📌 前端测试文件必须放在 `web/tests/` 下，参照 `web/tests/README.md`

| 测试类型 | 目录 | 说明 |
|---------|------|------|
| 单元测试 | `web/tests/unit/` | hooks/lib/components |
| 集成测试 | `web/tests/integration/` | 组件+Store+API Mock |
| E2E 测试 | `web/tests/e2e/` | Playwright |
| Feature 测试 | `web/tests/features/F{xxx}/` | 对齐后端 tests/features/ |
| Bug 测试 | `web/tests/bugs/B{xxx}/` | 对齐后端 tests/bugs/ |
| Mock 配置 | `web/tests/mocks/` | MSW handlers/fixtures |

---

## 前端 Mock 规范

📌 基于 R3_API_CONTRACT 用 MSW 独立开发，不阻塞等后端

```typescript
import { http, HttpResponse } from 'msw';

export const handlers = [
    http.get('/api/v1/resource', () => {
        return HttpResponse.json({ list: [], total: 0 });
    }),
];
```

切换 Real API：

```typescript
const client = axios.create({
    baseURL: import.meta.env.RSBUILD_API_BASE_URL,
});
```

**规则**：
- Mock 数据必须严格遵循 R3_API_CONTRACT.md 定义
- 切换 Real API 只需修改环境变量，不改动业务代码
- 禁止在业务代码中硬编码 Mock 数据

---

## 专项测试触发

📌 读取 `frontend-specialized-tests.md`，Feature 开发至少覆盖：

| # | 测试类型 | 必须覆盖 |
|---|---------|---------|
| 1 | 组件渲染 | ✅ |
| 2 | 用户交互 | ✅ |
| 3 | Hook 逻辑 | ✅ |
| 4 | API 集成 (MSW) | ✅ |
| 5 | 边界值 | ✅ |
| 6 | 错误处理 | 可选 |
| 7 | 响应式布局 | 可选 |
| 8 | 可访问性 | 可选 |

---

## 目录结构

```
web/
├── src/
│   ├── components/ui/    ← shadcn/ui 组件
│   ├── hooks/            ← 自定义 Hooks
│   ├── lib/api/          ← API 模块
│   ├── pages/            ← 页面组件
│   ├── themes/           ← 主题配置
│   └── mocks/            ← MSW handlers
├── tests/                ← 测试目录（与 src 分离）
│   ├── unit/
│   ├── integration/
│   ├── e2e/
│   ├── features/
│   ├── bugs/
│   └── mocks/
└── e2e/                  ← Playwright E2E（旧位置，逐步迁移到 tests/e2e/）
```

---

## 完成门禁

```
Feature 完成检查:
- [ ] Pipeline 全部通过（Step 1-7）
- [ ] 代码无中文注释
- [ ] bun run typecheck 通过
- [ ] 文档与实现一致
- [ ] SCOPE.md 已生成
- [ ] Commit Message 符合规范
- [ ] task-pool.md 状态 → Review
- [ ] 建议后续角色 → QA
```

---

## 📋 输出要求

| 输出项 | 存放位置 | 格式 |
|--------|----------|------|
| 代码 | `{PROJECT_PATH}/web/src/` | TSX/TS |
| 测试代码 | `{PROJECT_PATH}/web/tests/` | TSX/TS |
| `SCOPE.md` | `{docs_internal}/requirements/{task-id}/` | Markdown |
