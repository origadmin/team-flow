# flow - AI Team Collaboration Framework CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/origadmin/team-flow)](https://goreportcard.com/report/github.com/origadmin/team-flow)
[![Go Reference](https://pkg.go.dev/badge/github.com/origadmin/team-flow.svg)](https://pkg.go.dev/github.com/origadmin/team-flow)

`flow` 是 `_team` AI 团队协作框架的命令行工具，提供初始化、迁移和可视化编辑功能。

## ✨ 功能特性

- 🚀 **项目初始化** - 快速初始化 `_team` 框架到项目中
- 🔄 **v1 → v2 迁移** - 自动迁移 task-pool 到 beads-native
- 📊 **状态查看** - 查看项目当前使用的框架版本
- 🎨 **Flow 可视化编辑器** - Web UI 查看和编辑工作流（开发中）

## 📦 安装

### 方式一：go install（推荐）

```bash
go install github.com/origadmin/team-flow/cmd/flow@latest
```

### 方式二：从源码编译

```bash
git clone https://github.com/origadmin/team-flow.git
cd team-flow
go build -o flow cmd/flow/main.go
mv flow /usr/local/bin/  # 或添加到 PATH
```

## 🚀 快速开始

### 1. 初始化项目

```bash
# 进入你的项目目录
cd /path/to/your/project

# 初始化（默认 v2 - beads-native）
flow init

# 或者使用 v1 (task-pool)
flow init --v1
```

初始化会：
- 创建 `.team/` 目录
- 生成 `.team/project.md` 配置模板
- 复制 `framework/_team/` 框架文件

### 2. 查看状态

```bash
flow status
```

输出示例：
```
项目状态:
  路径: /path/to/project
  版本: v2 (beads-native)
  路径: /path/to/project/framework/_team/v2
  配置: /path/to/project/.team ✓
```

### 3. v1 → v2 迁移

```bash
flow migrate
```

### 4. 启动 Flow 编辑器（开发中）

```bash
flow editor
# 访问 http://localhost:3000
```

## 📋 命令列表

| 命令 | 简写 | 说明 |
|------|------|------|
| `flow init` | `flow i` | 初始化项目 |
| `flow init --v1` | | 使用 v1 版本 |
| `flow init --v2` | | 使用 v2 版本（默认）|
| `flow migrate` | `flow m` | v1 → v2 迁移 |
| `flow status` | `flow s` | 查看项目状态 |
| `flow editor` | `flow e` | 启动 Flow 编辑器 |
| `flow version` | `flow v` | 查看版本信息 |

## 🎨 Flow 可视化编辑器

启动 Web UI 查看和编辑 `_team` 工作流：

```bash
flow editor --port 8080 --host 0.0.0.0
```

功能（开发中）：
- 📊 可视化角色流转图（Triage → Tech Lead → Dev → QA → PM）
- ⚙️ 编辑工作流配置
- 💾 实时保存修改到 `workflows/*.md`

## 📁 项目结构

```
team-flow/
├── cmd/
│   └── flow/              ← CLI 入口
│       └── main.go
├── internal/
│   ├── init/              ← 初始化逻辑
│   ├── migrate/           ← 迁移逻辑
│   ├── editor/            ← Web 编辑器
│   └── version/           ← 版本信息
├── framework/             ← _team 框架文件
│   ├── v1/               ← v1 体系
│   ├── v2/               ← v2 体系（beads-native）
│   ├── prompts/           ← 角色规则
│   ├── workflows/         ← 工作流规范
│   └── templates/         ← 文档模板
├── go.mod
└── README.md
```

## 🔧 开发

### 前置条件

- Go 1.21+
- Node.js 18+（Flow Editor 前端）

### 本地运行

```bash
# 克隆仓库
git clone https://github.com/origadmin/team-flow.git
cd team-flow

# 安装依赖
go mod download

# 运行（开发模式）
go run cmd/flow/main.go init

# 编译
go build -o flow cmd/flow/main.go
```

### 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

Apache License 2.0

## 🔗 相关链接

- `_team` 框架文档：见 `framework/` 目录
- GitHub 仓库：https://github.com/origadmin/team-flow
- 问题反馈：https://github.com/origadmin/team-flow/issues

---

**Built with ❤️ for AI-Native Team Collaboration**
