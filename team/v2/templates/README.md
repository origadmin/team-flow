# v2 Templates

> 模板定义产出物格式规范。beads 管任务状态，模板管产出物内容。

## 模板清单

### P0 — 必须模板（Feature/Bug 开发必用）

| 模板 | 用途 | 触发条件 | Subsystem |
|------|------|---------|-----------|
| `feature-test-template.md` | 后端 Feature 测试覆盖声明 | Feature 开发（后端） | subsystem:backend |
| `bug-test-template.md` | 后端 Bug 复现测试用例 | Bug 修复（后端） | subsystem:backend |
| `bug-index-template.md` | Bug 迭代索引（AI 入口） | Bug 创建/迭代时 | 通用 |
| `frontend-feature-test-template.md` | 前端 Feature 测试覆盖声明 | Feature 开发（前端） | subsystem:frontend |
| `frontend-bug-test-template.md` | 前端 Bug 复现测试用例 | Bug 修复（前端） | subsystem:frontend |
| `scope-template.md` | 变更报告（机器可读 YAML） | 所有任务完成时 | 通用 |

### P1 — 重要模板（QA/验收必用）

| 模板 | 用途 | 触发条件 |
|------|------|---------|
| `test-report-template.md` | 测试报告 | QA 验证 |
| `closed-loop-verification-template.md` | 闭环验证报告 | Bug 关闭 |
| `gherkin-feature-template.md` | BDD 验收标准 | Feature AC 编写 |

### P2 — 辅助模板

| 模板 | 用途 | 触发条件 |
|------|------|---------|
| `architecture-template.md` | 架构设计文档 | Tech Lead 分析 |
| `prd-template.md` | 产品需求文档 | PM 需求编写 |
| `ui-design-template.md` | UI 设计规范 | UI 设计 |
| `api-issue-template.md` | API 问题报告 | 前后端协作 |
| `user-story-template.md` | 用户故事 | 敏捷需求 |

---

## v1 → v2 模板变更

| 变更项 | v1 | v2 |
|--------|----|----|
| 任务状态更新 | "写入 task-pool.md" | `bd update <id> --notes` |
| Phase 转换 | 手动更新状态列 | `bd update <id> --add-label phase:xxx` |
| 任务 ID | F/B/C/A-NNN 为主 ID | cms-xxx 为主 ID，v1 ID 存 external-ref |
| 模板头部 | 无 beads 关联 | 增加 beads 关联表（Issue ID / task ID / Phase / Subsystem） |
| 模板尾部 | 无状态更新命令 | 增加 `bd update` 命令示例 |

---

## 使用方式

子 Agent 在执行任务时，根据 subsystem 和任务类型选择对应模板：

```bash
# 后端 Feature
bd update <id> --notes "Loading template: {TEAM_PATH}/templates/feature-test-template.md"

# 前端 Bug
bd update <id> --notes "Loading template: {TEAM_PATH}/templates/frontend-bug-test-template.md"
```

模板路径在 Triage 分发时注入子 Agent Prompt。
