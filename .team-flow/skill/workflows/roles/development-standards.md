# 功能开发输出规范 — v2 beads-native

> **角色**: Backend / Frontend Dev
> **前置**: 先读 `workflows/shared.md`

---

## 角色→规范映射

| 团队角色 | 加载规范文件 |
|---------|------------|
| backend-dev | development-standards.md + feature-test-template.md |
| frontend-dev | development-standards.md + frontend-feature-test-template.md |
| bugfix | bugfix-standards.md + bug-test-template.md |
| qa-engineer | test-standards.md + 专项测试文件 |

---

## 变更影响分析（强制）

```
识别到修改已有文件
    │
    ├── Step 1: 读取目标文件
    ├── Step 2: 识别变更类型
    ├── Step 3: 影响范围扫描
    │   ├── 后端: grep -r "TargetSymbol" --include="*.go"
    │   └── 前端: grep -r "TargetSymbol" --include="*.{ts,tsx}"
    ├── Step 4: 风险评级
    │   ├── 低风险 → 直接修改 + 全量回归
    │   ├── 中风险 → 输出影响报告 → 修改
    │   └── 高风险 → 输出影响报告 + 兼容方案
    └── Step 5: 分层回归验证
```

### 影响报告格式

```markdown
🔍 Change Impact Report:
   - 目标文件: {file_path}
   - 变更类型: {新增/修改逻辑/修改接口}
   - 风险等级: {低/中/高}
   - 影响文件数: {N}
```

---

## TDD 流程（强制）

```
1. [红] 编写失败测试
2. [绿] 最小实现通过
3. [重构] 优化结构
```

| 角色 | 目标覆盖率 |
|------|-----------|
| Backend Dev | 80%+ |
| Frontend Dev | 70%+ |

---

## beads 状态管理

```bash
# 开始开发
bd update <id> --claim --add-label phase:implement

# 完成开发
bd update <id> --add-label phase:verify

# 提交验证
bd update <id> --add-label phase:review
```

---

## 交付管线

**命令从 project.md Toolchain.pipeline 读取**

```
Step 1: format/vet
Step 2: lint
Step 3: build
Step 4: test
Step 5: test:coverage
Step 6: SCOPE.md
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
