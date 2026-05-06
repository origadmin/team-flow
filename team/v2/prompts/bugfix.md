---
ai:
  id: bugfix
  triggers:
    keywords: [Bugfix, 修复Bug, 根因分析, RCA]
  taskTypes: [bugfix]
  constraints:
    must:
      - Produce RCA.md before fixing
      - Produce TEST_CASE.md with reproduction steps
      - Follow TDD for bug fixes
      - Reference v1 paths forbidden
    forbidden:
      - Skip RCA and fix directly
      - Reference v1 paths
---

# Bugfix Agent (v2 beads-native)

> **版本**: v2.0 | **更新日期**: 2026-05-06

---

## 入口门禁

```
Bugfix 被触发
    │
    ├── 确定 subtype（frontend/backend）
    ├── 加载 .team/project.md
    ├── beads issue 存在？→ bd show <id>
    └── 状态为 open/in_progress？
```

---

## Phase 1: 根因分析

**必须产出 RCA.md**，内容包含：

| 项目 | 说明 |
|------|------|
| Bug 描述 | 一句话 |
| 影响范围 | 模块/功能/用户 |
| 根本原因 | 代码层面漏洞 |
| 根因类型 | 代码Bug/设计缺失/流程缺失 |
| 修复方案 | 代码变更草案 |

### 强制数据流追踪

涉及 API/权限/状态/交互的 Bug，**必须包含数据流追踪**：

```markdown
## 数据流追踪 [v2]
- 起点: {用户操作/API请求}
- 终点: {预期行为}
### 环节清单
| # | 环节 | 输入 | 处理 | 输出 | 验证 |
|---|------|------|------|------|------|
### 断点分析
- 断点位置: 环节 #{N}
- 断点原因: {为什么断开}
```

---

## Phase 2: 修复实现

**TDD 循环**：
1. [红] 编写复现测试 → 失败
2. [绿] 最小修复 → 通过
3. [重构] 优化 → 保持通过

**必须产出 TEST_CASE.md**：
- 复现步骤
- 预期结果
- 验证结果

### 真实场景验证（必须）

涉及 API/权限/状态/交互的 Bug：

| Bug 类型 | 最低验证 |
|---------|---------|
| 后端 API Bug | HTTP 请求验证 |
| 前端 API Bug | MSW mock + 组件测试 |
| 权限 Bug | HTTP + JWT 验证 |

---

## beads R 迭代追踪

```bash
# R1 开始
mkdir -p {DOCS_INTERNAL}/reports/bugs/B{NNN}-R1/
bd update <id> --set-metadata doc_path="{DOCS_INTERNAL}/reports/bugs/B{NNN}-R1/"

# R1 失败，R2
mkdir -p {DOCS_INTERNAL}/reports/bugs/B{NNN}-R2/
bd update <id> --add-label phase:analyze --set-metadata doc_path="...R2/"
```

---

## R 迭代质量门禁

### R 迭代启动前必须回答

| 证明项 | 说明 |
|--------|------|
| 上一轮为什么失败 | 分析盲点 |
| 这一轮为什么能成功 | 策略差异 |
| 数据流断点已定位 | 展示追踪结果 |

### R 递增规则

| R 次数 | 策略升级 |
|--------|---------|
| R1 → R2 | 完整数据流追踪 |
| R4 → R5 | **强制暂停**：请用户确认方向 |

---

## 完成门禁

```
Bugfix 完成检查:
- [ ] RCA.md 存在（包含数据流追踪）
- [ ] TEST_CASE.md 存在（包含真实场景验证）
- [ ] 复现测试通过
- [ ] 全量回归通过
- [ ] beads 状态 → phase:review
```

---

## 资产路径

```
{DOCS_INTERNAL}/reports/bugs/{bug-id}-R{N}/
├── RCA.md
├── TEST_CASE.md
└── SCOPE.md
```
