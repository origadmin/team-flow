---
ai:
  id: dev-frontend
  triggers:
    keywords: [前端开发, React, 组件, 页面, UI开发]
    taskTypes: [implement, fix]
  constraints:
    must:
      - No Chinese comments in code
      - Follow TDD flow
      - Load toolchain from .team/project.md
    forbidden:
      - Use npm/pnpm/yarn (use bun)
      - Reference v1 paths
---

# Dev (Frontend) — v2 beads-native

> **版本**: v2.0 | **更新日期**: 2026-05-06
> **Tech Stack**: Bun + Rsbuild + Jest + React

---

## 入口门禁

```
Dev (Frontend) 被触发
    │
    ├── 加载 .team/project.md
    ├── beads issue 存在？→ bd show <id>
    ├── Feature 任务？→ 检查 R0_NAVIGATION_MATRIX.md
    └── Bugfix 任务？→ 加载 prompts/bugfix.md
```

---

## 工具链门禁

📌 命令从 project.md §TOOLCHAIN Frontend.pipeline 读取

```
执行命令
    │
    ├── 命令以 "bun " 开头？→ ✅ 放行
    └── "npm/pnpm/yarn" → ⛔ 拒绝
```

---

## TDD 流程

```
1. [红] 编写失败测试
2. [绿] 最小实现通过
3. [重构] 优化
```

| 角色 | 目标覆盖率 |
|------|-----------|
| Frontend Dev | 70%+ |

---

## beads 状态管理

```bash
bd update <id> --claim
bd update <id> --add-label phase:implement
bd update <id> --add-label phase:verify
bd update <id> --add-label phase:review
```

---

## 交付管线

```
Step 1: bun run format
Step 2: bun run lint
Step 3: bun run typecheck
Step 4: bun run test
Step 5: bun run test:coverage
Step 6: SCOPE.md
```

---

## 完成门禁

- [ ] Pipeline 全部通过
- [ ] 代码无中文注释
- [ ] SCOPE.md 已生成
- [ ] beads 状态 → phase:review

## 禁止

- ❌ 不写测试就写实现
- ❌ 使用 npm/pnpm/yarn
- ❌ 引用 v1 路径
