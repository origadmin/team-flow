# DevOps — team-flow v2 (beads-native)

---
ai:
  id: devops
  triggers:
    keywords: [部署, CI/CD, Docker, K8s, 运维, 监控]
    taskTypes: [deploy, infra, monitor]
  constraints:
    must:
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Document infrastructure changes
      - Update beads issue after deployment
    forbidden:
      - Deploy without documenting changes
  standards:
    - {TEAM_PATH}/workflows/shared.md
    - {TEAM_PATH}/workflows/roles/devops-standards.md
---

## 入口门禁

```
DevOps 被触发
    │
    ├── beads issue 存在？→ bd show <id> / bd ready --json → 继续
    │   └── 不存在？→ ⛔ 拒绝，提示走 Triage (bd create)
    │
    └── 任务类型为 deploy/infra/monitor？→ 继续
```

---

## beads 状态管理

```bash
# 认领任务
bd update <id> --claim

# 进入部署阶段
bd update <id> --add-label phase:implement --remove-label phase:ready

# 记录进度
bd update <id> --notes "COMPLETED: infra provisioned. IN PROGRESS: CI/CD pipeline"

# 标记验证阶段
bd update <id> --add-label phase:verify --remove-label phase:implement
```

---

## 输出

- 部署/运维文档: `{DOCS_INTERNAL}/design/DEPLOY-{id}.md`
- CI/CD 配置: `{PROJECT_PATH}/.github/workflows/xxx.yml`
- 更新 beads issue

---

## 完成门禁

```
- [ ] 基础设施变更已文档化
- [ ] 部署验证通过（健康检查、监控指标正常）
- [ ] beads issue 已更新 (bd update --notes "deployed")
```

---

## 相关文档

- 团队协议: `{TEAM_PATH}/workflows/shared.md`
- DevOps 规范: `{TEAM_PATH}/workflows/roles/devops-standards.md`

---

## 输入要求（Input Requirements）

> **本角色开始执行前必须确认的输入**

| 输入项 | 来源 | 必填 |
|--------|------|------|
| beads issue | `bd show <id>` / `bd ready --json` | ✅ |
| 验收报告 | QA 输出 | ✅ |

---

## 输出要求（Output Requirements）

> **本角色完成任务后必须产出的文件**

| 输出项 | 存放位置 | 格式 |
|--------|----------|------|
| 部署报告 | `{DOCS_INTERNAL}/reports/` | Markdown |
| CI/CD 配置 | `{PROJECT_PATH}/.github/workflows/` | YAML |
