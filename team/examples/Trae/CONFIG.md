# Trae 配置示例

## ⛔ ENTRY GATE — Mandatory Path Check (MUST DO BEFORE ANY FILE WRITE)

> **AI 必须在写任何文件前先执行此检查，否则路径错误 100% 复发**

```
在写入文件前，你必须：
1. 识别文件类型（task-pool / backlog / SPEC / AC / RCA / 代码 / 测试 / ...）
2. 对照路径表验证：

   | 文件类型 | 正确路径 | 错误路径 |
   |----------|----------|----------|
   | task-pool.md | {PROJECT_PATH}/.team/task-pool.md | _team/task-pool.md |
   | backlog.md | {PROJECT_PATH}/.team/backlog.md | _team/backlog.md |
   | project.md | {PROJECT_PATH}/.team/project.md | _team/project.md |
   | SPEC/AC | {docs_internal}/{project}/requirements/... | _team/... |
   | RCA | {docs_internal}/{project}/reports/errors/... | _team/... |
   | 代码 | {PROJECT_PATH}/... | _team/... |

3. 如果路径不存在于列表中 → 停止，询问用户

```

> **这是强制检查点。不要跳过。路径错误 = 项目污染。**

---

## 环境变量
TEAM_PATH: D:/workspace/project/golang/origadmin/framework/_team
PROJECT_PATH: D:/workspace/project/golang/origadmin/framework/projects/orig-cms
docs_internal: D:/workspace/project/golang/origadmin/framework/_docs/orig-cms/
docs_external: D:/workspace/project/golang/origadmin/framework/projects/orig-cms/docs/

## 初始化
加载 {TEAM_PATH}/SKILL.md，按其中步骤执行
