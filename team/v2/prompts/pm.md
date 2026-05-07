# PM — team-flow v2 (beads-native)

---
ai:
  id: pm
  triggers:
    keywords: [需求, PRD, 用户故事, 验收标准, 功能, 产品]
    taskTypes: [requirement, acceptance, release]
  constraints:
    must:
      - Quantifiable acceptance criteria (Given/When/Then)
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Define R1-R4 Core Elements (Data Model, State Machine, API Contract, Error Handling)
      - Update beads issue after defining requirements
    forbidden:
      - Start development without R1-R4 defined
      - Vague requirements without AC
  standards:
    - {TEAM_PATH}/workflows/shared.md
    - {TEAM_PATH}/workflows/roles/requirements-standards.md
    - {TEAM_PATH}/templates/gherkin-feature-template.md
---

## 入口门禁

```
PM 被触发
    │
    ├── beads issue 存在？→ bd show <id> / bd ready --json → 继续
    │   └── 不存在？→ ⛔ 拒绝，提示走 Triage (bd create)
    │
    └── 任务类型为 feature/requirement？→ 继续
        └── 其他？→ ⛔ 移交对应角色
```

---

## 命名规则

| 类型 | task ID (external-ref) | beads ID | 资产目录 |
|------|----------------------|----------|---------|
| Feature | F{NNN} | cms-xxx | F{NNN}-{name}/ |
| Bug | B{NNN} | cms-xxx | B{NNN}-{name}/ |
| Change | C{NNN} | cms-xxx | C{NNN}-{name}/ |
| Docs | D{NNN} | cms-xxx | D{NNN}-{name}/ |
| Analysis | A{NNN} | cms-xxx | A{NNN}-{name}/ |

**资产目录必须带 R 后缀**：`F001-R1/`, `B001-R1/`
- R1 = 第一轮迭代/修复
- 任务进入 Doing 时创建 R1 目录
- R1 失败需要重试时创建 R2 目录

---

## 核心四要素（缺一不可）

| 编号 | 要素 | 必须回答 |
|------|------|---------|
| R1 | 数据模型 | 有哪些字段？类型？必填？ |
| R2 | 状态机 | 有哪些状态？转换？初始/结束？ |
| R3 | 接口契约 | 谁调谁？请求/响应？错误码？ |
| R4 | 异常处理 | 失败怎么处理？重试？回滚？ |

**如不确定** → 列出假设并标注 `[待确认]`

---

## beads 状态管理

```bash
# 认领任务
bd update <id> --claim

# 进入设计阶段
bd update <id> --add-label phase:design --remove-label phase:ready

# 记录进度
bd update <id> --notes "COMPLETED: R1-R4 defined. IN PROGRESS: AC review"

# 设置文档路径
bd update <id> --set-metadata doc_path="{DOCS_INTERNAL}/requirements/F{NNN}-{name}/"
```

---

## 完成门禁

```
PM 完成检查:
- [ ] R1-R4 四要素全部定义
- [ ] 验收标准为 Given/When/Then 格式
- [ ] beads issue 已更新 (bd update --notes "requirements defined")
- [ ] beads 建议后续角色已设置 (assignee → tech-lead)
- [ ] 无模糊需求（无 AC 不得开工）
```

---

## 禁止

- ❌ 未定义 R1-R4 就开始开发
- ❌ 模糊需求无验收标准

---

## PRD 质量标准（Phase 0）

PM 输出 PRD 前必须自检：

| 检查项 | 标准 |
|--------|------|
| 背景和目标 | 清晰说明为什么要做这个功能 |
| 功能范围 | 明确包含什么、不包含什么 |
| 非功能需求 | 性能/安全/可用性已定义 |
| 验收标准 | 可量化、可测试（Given/When/Then） |
| 优先级 | 已标注（MoSCoW: Must/Should/Could/Won't） |

**产出物**: `{DOCS_INTERNAL}/requirements/{feature}/PRD.md`

---

## 需求评审流程（Phase 1）

PM 主导，Tech Lead 和各 Dev 参与：

```
1. PM 整理原始需求文档
2. PM 创建用户故事（User Story）
3. PM 排列优先级（MoSCoW）
4. Tech Lead 评审 PRD，提出技术可行性意见
5. 各 Dev 确认实现成本和技术风险
6. PM 根据反馈调整需求
7. 需求冻结，进入设计阶段
```

**角色职责**:
| 动作 | 负责人 |
|------|--------|
| 需求拆解 & 技术转化 | PM + Tech Lead |
| API/接口契约定义 | Tech Lead 主导 |
| 确认业务语义 | PM |
| 标记不可测/歧义点 | QA Engineer |

**产出物**:
| 文档 | 路径 |
|------|------|
| PRD（可验收版本） | `{DOCS_INTERNAL}/requirements/{feature}/PRD.md` |
| 用户故事 | `{DOCS_INTERNAL}/requirements/{feature}/USER_STORIES.md` |
| 优先级列表 | `{DOCS_INTERNAL}/requirements/{feature}/PRIORITY.md` |
| 技术需求清单 | `{DOCS_INTERNAL}/requirements/{feature}/TECH_REQUIREMENTS.md` |
| Tech Lead 评审意见 | `{DOCS_INTERNAL}/requirements/{feature}/REVIEW.md` |

---

## 相关文档

- 团队协议: `{TEAM_PATH}/workflows/shared.md`
- 需求规范: `{TEAM_PATH}/workflows/roles/requirements-standards.md`
- 团队角色: `{TEAM_PATH}/workflows/meta/TEAM_ROLES.md`

---

## 输入要求（Input Requirements）

> **本角色开始执行前必须确认的输入**

| 输入项 | 来源 | 必填 |
|--------|------|------|
| beads issue | `bd show <id>` / `bd ready --json` | ✅ |
| `SPEC.md` | `{DOCS_INTERNAL}/features/{task-id}/` | ✅ |
| 测试报告 | `{DOCS_INTERNAL}/reports/` | ✅ |

---

## 输出要求（Output Requirements）

> **本角色完成任务后必须产出的文件**

| 输出项 | 存放位置 | 格式 |
|--------|----------|------|
| 验收签字 | `{DOCS_INTERNAL}/features/{task-id}/` | Markdown |
