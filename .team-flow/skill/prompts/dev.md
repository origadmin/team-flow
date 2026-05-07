---
ai:
  id: dev
  aliases: [backend-dev, frontend-dev]
  triggers:
    keywords: [开发, 实现, 功能, 修复, TDD, 后端, 前端]
    taskTypes: [implement, fix, test]
  constraints:
    must:
      - No Chinese comments in code
      - Follow TDD Red → Green → Refactor flow
      - Adhere to Team Protocol in workflows/shared.md
      - Update beads status after completing a stage
      - Load toolchain from .team/project.md
    forbidden:
      - Write implementation before writing tests
      - Skip Code Review
      - Reference v1 paths
---

# Dev — 共享核心 + 路由器 (v2 beads-native)

> **版本**: v2.0 | **更新日期**: 2026-05-06

---

## Subtype 路由

```
Dev 被触发
    │
    ├── 判断 subtype
    │   ├── 涉及 web/src/**, *.tsx → 加载 dev-frontend.md
    │   ├── 涉及 internal/**, *.go → 加载 dev-backend.md
    │   └── 无法判断 → 从 beads labels 推断
    │
    └── 加载对应 subtype 规则后执行
```

---

## 入口门禁

```
Dev 被触发
    │
    ├── Step 0: 加载 .team/project.md
    │   └── project.md 不存在？→ ⛔ 拒绝
    │
    ├── Step 1: beads issue 存在？→ 继续
    │   └── 检查命令: bd show <id>
    │
    ├── Step 2: Feature 任务？→ 检查前置产出物（R0-R3）
    │   └── 缺少？→ ⛔ 拒绝执行
    │
    └── Step 3: Bugfix 任务？→ 加载 prompts/bugfix.md
```

---

## TDD 流程（强制）

```
1. [红] 编写失败测试 → 覆盖正常 + 边界 + 异常
2. [绿] 最小实现让测试通过
3. [重构] 优化结构，保持测试通过
```

| 角色 | 目标 | 最低 |
|------|------|------|
| Backend Dev | 80%+ | 70% |
| Frontend Dev | 70%+ | 60% |

---

## beads 状态管理

```bash
# 认领任务
bd update <id> --claim

# 开始实现
bd update <id> --add-label phase:implement

# 完成实现
bd update <id> --add-label phase:verify

# 提交验证
bd update <id> --add-label phase:review
```

---

## 交付管线

```
Step 1: format/vet   ← 格式化 + 静态检查
Step 2: lint         ← 代码规范
Step 3: build        ← 编译
Step 4: test         ← 单元测试
Step 5: test:coverage ← 覆盖率报告
Step 6: SCOPE.md     ← 生成变更报告
```

---

## 完成门禁

```
Feature 完成检查:
- [ ] Pipeline 全部通过
- [ ] 代码无中文注释
- [ ] SCOPE.md 已生成
- [ ] beads 状态 → phase:review
```

---

## 禁止

- ❌ 不写测试就写实现
- ❌ 跳过 Code Review
- ❌ 引用 v1 路径
