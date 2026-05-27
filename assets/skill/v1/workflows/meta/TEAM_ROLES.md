# 团队角色定义

> 所有角色共享的团队成员定义
> 路径: `{TEAM_PATH}/workflows/meta/TEAM_ROLES.md`

---

## 角色映射表

| 通用模型 | 我们的角色 | 核心职责 | 谁说了算 |
|---------|-----------|---------|---------|
| **Req** | PM | 定义问题、验收标准、做决策、负最终责任 | **PM** - 最终放行人 |
| **Arch** | Tech Lead | 技术方案、架构决策、质量属性保障 | Tech Lead - 技术推荐 |
| **Dev** | Backend Dev / Frontend Dev / Android Dev / iOS Dev | 编码实现、TDD、缺陷修复 | Dev - 代码怎么写 |
| **QE** | QA Engineer | Code Review、测试盲区、闭环验证 | **QA** - 一票否决上线 |

> **一句话总结**: PM 决定做什么，Tech Lead 决定怎么做，Dev 负责写出来，QA 保证没写错。

---

## 应用开发团队（6人基础 + 可扩展）

| 角色 | ID | 核心职责 | 汇报对象 |
|------|-----|---------|----------|
| Tech Lead | tech-lead | 架构设计、技术选型、代码审查 | - |
| Backend Dev | backend-dev | Go/go-kratos 微服务实现 | Tech Lead |
| Frontend Dev | frontend-dev | 跨框架 UI 开发（具体框架见 Project.md 的 {技术栈}） | Tech Lead |
| **Android Dev** | android-dev | Android 原生开发（Kotlin） | Tech Lead |
| **iOS Dev** | ios-dev | iOS 原生开发（Swift） | Tech Lead |
| QA Engineer | qa-engineer | 测试策略、质量把关、闭环验证 | Tech Lead |
| DevOps | devops | CI/CD、容器化、监控 | Tech Lead |
| PM | pm | 需求管理、PRD、进度、最终放行 | - |

> **扩展说明**: Dev 是通用角色，可根据需要增加 Android Dev、iOS Dev、Flutter Dev 等。职责统一为「编码实现 + TDD + 缺陷修复」。

---

## 框架开发团队（4人）

| 角色 | ID | 核心职责 | 负责模块 |
|------|-----|---------|---------|
| Framework Architect | framework-architect | 架构标准、技术决策、ADR 撰写 | 全局 |
| Toolkits Dev | toolkits-dev | 基础工具组件（TDD 实现） | toolkits/ |
| Runtime Dev | runtime-dev | DI 容器、启动流程（TDD 实现） | runtime/ |
| Contrib Dev | contrib-dev | 服务发现、消息队列（TDD 实现） | contrib/ |

---

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-04-25 | 从 standards.md 拆分，提取团队角色定义 |
