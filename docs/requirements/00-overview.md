# team-flow 需求文档 — 概述

> **版本**: v1.0 | **日期**: 2026-05-11 | **状态**: 待确认

## 项目定义

team-flow 是一个多 Agent AI 协作框架，基于 beads 实现任务管理，提供 Triage→TechLead→Dev→QA 流水线、三层门禁和 11 个角色定义。

## 文档索引

| 文档 | 内容 | 状态 |
|------|------|------|
| [01-core-requirements.md](01-core-requirements.md) | 核心需求：做什么、不做什么 | 待确认 |
| [02-project-management.md](02-project-management.md) | 项目管理模式：workspace vs standalone | 待确认 |
| [03-role-system.md](03-role-system.md) | 角色体系：Triage/Dev/QA 等 | 待确认 |
| [04-task-lifecycle.md](04-task-lifecycle.md) | 任务生命周期：创建→执行→验证→归档 | 待确认 |
| [05-status-line.md](05-status-line.md) | Status Line 规则 | 待确认 |
| [06-inconsistencies.md](06-inconsistencies.md) | 已发现的不一致问题 | 待确认 |
| [07-improvement-plan.md](07-improvement-plan.md) | 改进计划 | 待确认 |

## 当前实现文件

| 文件 | 路径 | 行数 | 说明 |
|------|------|------|------|
| SKILL.md | `.trae/skills/team-flow/SKILL.md` | ~300 | 入口文件 |
| shared.md | `.trae/skills/team-flow/workflows/shared.md` | ~1017 | 共享协议 |
| BOUNDARY.md | `.trae/skills/team-flow/BOUNDARY.md` | ~265 | 层边界 |
| triage.md | `.trae/skills/team-flow/prompts/triage.md` | ~1025 | Triage 提示词 |
| dev.md | `.trae/skills/team-flow/prompts/dev.md` | ~452 | Dev 提示词 |
| bugfix.md | `.trae/skills/team-flow/prompts/bugfix.md` | ~441 | Bugfix 提示词 |
| tech-lead.md | `.trae/skills/team-flow/prompts/tech-lead.md` | ~390 | Tech Lead 提示词 |
| qa-engineer.md | `.trae/skills/team-flow/prompts/qa-engineer.md` | ~300 | QA 提示词 |
| commands.md | `.trae/skills/team-flow/references/commands.md` | ~57 | 命令参考 |
| path-resolution.md | `.trae/skills/team-flow/references/path-resolution.md` | ~102 | 路径解析 |
| agent-mapping.md | `.trae/skills/team-flow/references/agent-mapping.md` | ~56 | Agent 映射 |

## 文档约定

- 每个文档不超过 200 行
- 需求用"必须/应该/可以"三级描述
- 不一致问题用 🔴 严重 / 🟡 中等 / 🟢 轻微 标记
- 所有需求必须可追溯到实现文件
