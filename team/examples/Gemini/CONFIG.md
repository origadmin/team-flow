# Gemini CLI 配置示例

## 环境变量
TEAM_PATH: D:/workspace/project/golang/origadmin/framework/_team
PROJECT_PATH: D:/workspace/project/golang/origadmin/framework/projects/orig-cms
docs_internal: D:/workspace/project/golang/origadmin/framework/_docs/orig-cms/
docs_external: D:/workspace/project/golang/origadmin/framework/projects/orig-cms/docs/

## 初始化规则
根据 {TEAM_PATH}/SKILL.md 中定义的流程，Gemini CLI 在启动时应遵循以下初始化规则：

1. **加载环境变量**: 确保已加载所有必要的环境变量，如 TEAM_PATH, PROJECT_PATH, docs_internal, docs_external。
2. **加载/完善项目基础信息**: 读取 {PROJECT_PATH}/.team/project.md，检查并完善项目名称、技术描述、技术栈、核心模块列表和目录结构。如果缺失，则执行 templates/project.md 中的「已有项目初始化」流程。
3. **执行初始化**: 读取 {TEAM_PATH}/workflows/shared.md，并按其中「首次启动：自动初始化」步骤执行，包括检查 .team/ 目录、对比 TEAM_VERSION、将当前任务写入任务池。
4. **输入匹配**: 每次接收到用户输入时，读取 {TEAM_PATH}/workflows/shared.md 中的「输入匹配规则」，并按规则匹配用户输入到意图类别，然后加载对应的 prompts/ 角色定义和 workflows/roles/ 规范文件，按角色规范执行任务。