# Templates Index

> 模板文件索引，供 AI 按需加载使用。

---

## 文档模板

| 模板 | 用途 | 触发条件 |
|------|------|---------|
| `prd-template.md` | PRD 需求文档 | 新功能/需求 |
| `gherkin-feature-template.md` | BDD 验收标准 | 开发前验收定义 |
| `architecture-template.md` | 架构设计文档 | 架构变更 |
| `api-issue-template.md` | API 问题追踪 | API 相关问题 |
| `ui-design-template.md` | UI 设计文档 | UI 变更 |
| `scope-template.md` | 项目范围定义 | 项目启动 |
| `backlog-template.md` | 产品待办列表 | 需求管理 |
| `user-story-template.md` | 用户故事 | 敏捷需求 |
| `test-report-template.md` | 测试报告 | 测试完成后 |
| `closed-loop-verification-template.md` | 闭环验证 | 修复验证 |

## 测试模板

| 模板 | 用途 | 触发条件 |
|------|------|---------|
| `feature-test-template.md` | Feature 测试覆盖声明（后端） | Feature 开发（后端） |
| `bug-test-template.md` | Bug 复现测试用例（后端） | Bug 修复（后端） |
| `frontend-feature-test-template.md` | Feature 测试覆盖声明（前端） | Feature 开发（前端） |
| `frontend-bug-test-template.md` | Bug 复现测试用例（前端） | Bug 修复（前端） |

---

## 使用说明

1. **加载**: 根据变更类型，选择对应模板
2. **填充**: 将模板内容复制到目标位置，替换占位符 `{xxx}`
3. **验证**: 确保所有占位符已替换，文档路径正确

**模板路径规则**:
- 共通模板（`{TEAM_PATH}/templates/`）→ 通用场景
- 项目模板（`projects/{project}/templates/`）→ 项目特定场景