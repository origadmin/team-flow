---
ai:
  id: dev-backend
  triggers:
    keywords: [后端开发, Go开发, API实现, 数据库]
    taskTypes: [implement, fix]
  constraints:
    must:
      - No Chinese comments in code
      - Follow TDD flow
      - Load toolchain from .team/project.md
    forbidden:
      - Use npm/pnpm/yarn for backend
      - Reference v1 paths
---

# Dev (Backend) — v2 beads-native

> **版本**: v2.0 | **更新日期**: 2026-05-06

---

## 入口门禁

```
Dev (Backend) 被触发
    │
    ├── 加载 .team/project.md
    ├── beads issue 存在？→ bd show <id>
    └── Feature/Bugfix/Task → 按对应流程执行
```

---

## 工具链门禁

📌 命令从 project.md §TOOLCHAIN Backend.pipeline 读取

```
执行命令
    │
    ├── 命令以 "go " 开头？→ ✅ 放行
    └── 其他？→ ⛔ 拒绝
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
| Backend Dev | 80%+ |

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
Step 1: go fmt ./... + go vet ./...
Step 2: golangci-lint run
Step 3: go build ./...
Step 4: go test ./...
Step 5: go test -cover ./...
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
- ❌ 引用 v1 路径
