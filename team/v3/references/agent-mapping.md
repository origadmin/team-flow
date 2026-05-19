# Agent Mapping & Role Configuration

## Agent Mapping

| Role | subagent_type | Trigger Keywords | Prompt File |
|------|--------------|-----------------|-------------|
| Triage | (main agent) | all | prompts/triage.md |
| Tech Lead | tech-lead-architect | 实现/新增/开发/支持/设计/功能 | prompts/tech-lead.md |
| Dev (Backend) | developer-engineer | 后端/API/数据库/Go | prompts/dev-backend.md |
| Dev (Frontend) | developer-engineer | 前端/React/组件/页面 | prompts/dev-frontend.md |
| Bugfix | bugfix-expert | Bug/报错/崩溃/异常/问题/修复 | prompts/bugfix.md |
| QA | qa-engineer | 测试/验证/质量 | prompts/qa-engineer.md |
| PM | pm-documenter | 需求/PRD/产品/验收 | prompts/pm.md |
| Analysis | analysis-expert | 调研/分析/对比/评估 | prompts/analysis.md |
| DevOps | devops-engineer | 部署/CI/CD/Docker/K8s/运维 | prompts/devops.md |
| UI Designer | ui-designer | UI/界面/设计稿/组件/样式 | prompts/ui-designer.md |
| Framework Architect | tech-lead-architect | 框架/架构师/模块设计 | prompts/framework-architect.md |

## Required Config Files per Role

| Role | Must Load |
|------|-----------|
| Triage | shared.md, triage-standards.md |
| Tech Lead | shared.md, development-standards.md, architecture-standards.md |
| Dev Backend | shared.md, development-standards.md, specialized-tests.md |
| Dev Frontend | shared.md, development-standards.md, frontend-specialized-tests.md |
| Bugfix | shared.md, bugfix-standards.md, bug-test-template.md |
| QA | shared.md, test-levels.md, test-standards.md |
| PM | shared.md, requirements-standards.md, prd-template.md |
| Analysis | shared.md, analysis-standards.md |
| DevOps | shared.md, devops-standards.md |
| UI Designer | shared.md, ui-standards.md, ui-design-template.md |
| Framework Architect | shared.md, architecture-standards.md, framework-workflow.md |

## MILESTONES Sync

| task Status | beads Status | Meaning |
|-------------|-------------|---------|
| Todo | open | Not started |
| Doing | in_progress | In progress |
| Review | in_progress (notes: "awaiting review") | Awaiting confirmation |
| Archived | closed | Completed |
| Blocked Bug | open (labels: ["blocked"]) | Blocked by bug |
| Change Evaluating | open (labels: ["evaluating"]) | Change under evaluation |

## v2 vs v1 Key Differences

| Aspect | v1 | v2 |
|--------|----|----|
| Triage writes | task-pool.md (manual) | flow task create/update (beads) |
| Task source of truth | task-pool.md | beads `.beads/` |
| task-pool.md | Active + writable | Read-only export |
| ID format | F/B/C/A-NNN | `<beads-id>` (beads auto) |
| Status tracking | task-pool.md columns | flow task status |
| Multi-agent sync | task-pool.md git conflicts | Dolt git-native |
| Export | N/A | flow task export --format table |
