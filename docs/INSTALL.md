<!-- AUTO-GENERATED from _team/examples/ + _team/config/ — DO NOT EDIT MANUALLY -->
<!-- To update: modify source files, then run DOMGEN regeneration -->

# Team Workflow 安装指南

> **版本**: v6.0 | **生成日期**: 2026-04-24

---

## 快速开始

### 步骤 1：复制项目配置

在你的项目根目录创建 `SKILL.md`：

```markdown
加载 D:/workspace/project/golang/origadmin/framework/_team/SKILL.md，按其中步骤执行。

## 项目信息

- 项目名: {your-project}
- 技术栈: {your-stack}
- 工作目录: .
```

### 步骤 2：配置你的 AI 工具

选择你的工具，复制对应示例：

| 工具 | 示例位置 | 复制到 |
|------|----------|--------|
| Trae | `examples/Trae/` | `.trae/rules/SKILL.md` |
| Cursor | `examples/Cursor/` | 参考配置 |
| Gemini | `examples/Gemini/` | 参考配置 |
| OpenClaw | `examples/OpenClaw/` | 参考配置 |
| Claude Code | `examples/ClaudeCode/` | 参考配置 |

### 步骤 3：验证

```
初始化项目，检查 _team 状态
```

---

## 两层架构

**项目 SKILL.md**：项目层入口，存放项目特定信息（名称、路径、技术栈）。每个项目一份。

**框架 SKILL.md**：框架层规则，存放通用协作规则。所有项目共享。

分离后框架可共享、可独立更新，项目可定制。

---

## 示例目录

```
_team/
├── examples/              ← 复制这些示例
│   ├── QUICKSTART.md      ← 快速开始指南
│   ├── Trae/              ← Trae 配置示例
│   │   └── CONFIG.md
│   ├── Cursor/            ← Cursor 配置示例
│   │   └── CONFIG.md
│   ├── Gemini/            ← Gemini CLI 示例
│   │   └── CONFIG.md
│   ├── OpenClaw/          ← OpenClaw 示例
│   │   └── CONFIG.md
│   └── ClaudeCode/        ← Claude Code 示例
│       └── CONFIG.md
└── ...
```

---

## 常见问题

**Q: 项目 SKILL.md 和框架 SKILL.md 什么关系？**
A: 项目 SKILL.md 是入口，它告诉 AI 去哪里加载框架 SKILL.md。框架 SKILL.md 是实际驱动逻辑。

**Q: 为什么需要两层？**
A: 项目层存放项目特定信息（名称、路径、技术栈），框架层存放通用规则。分离后框架可共享，项目可定制。

**Q: 如何更新 _team？**
A: 引用方式：更新框架目录，所有项目自动生效。复制方式：重新复制到各项目。

**Q: 可以修改框架内容吗？**
A: 引用方式不建议修改（影响所有项目）。复制方式可以修改（仅影响当前项目）。

**Q: 多人开发时路径冲突怎么办？**
A: 每人在自己的工具配置中设置自己的环境变量值，`.team/project.md` 引用变量不写死路径。

<!-- Last generated: 2026-04-24 18:55 | Source: _team/examples/ + _team/config/ | Hash: AD611053 -->