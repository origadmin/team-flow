# Gemini CLI 配置示例 (v3)

## 环境变量

| 变量 | 值 | 说明 |
|------|-----|------|
| PROJECT_PATH | {PROJECT_PATH} | 项目根目录 |
| DOCS_INTERNAL | {DOCS_INTERNAL} | 内部文档目录 |
| DOCS_EXTERNAL | {DOCS_EXTERNAL} | 外部文档目录 |

## 初始化

1. 读取 `.team/version` → 确认为 v3
2. 加载 SKILL.md 作为入口（路径取决于 Gemini CLI 的 skill 目录配置）
3. 执行 `flow proc run` 启动流程引擎
4. 按 `flow proc run` 输出执行当前节点指令

## Gemini CLI 注意事项

- Gemini CLI 目前无内置 skill 目录约定，需手动配置 SKILL.md 路径
- 建议将 SKILL.md 内容作为系统提示的一部分加载
- `flow proc run` 输出可直接作为上下文注入
