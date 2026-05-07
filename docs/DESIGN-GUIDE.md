<!-- AUTO-GENERATED from _team/prompts/tech-lead.md — DO NOT EDIT MANUALLY -->
<!-- To update: modify _team/prompts/tech-lead.md, then run DOMGEN regeneration -->

# 设计指南 (Design Guide)

> **版本**: v3.0 | **生成日期**: 2026-04-24

---

## 概述

Tech Lead 角色负责技术设计和架构决策，是乙方技术侧的入口。核心原则：**收到任务必须先创建资产包再设计，杜绝空手开工**。

---

## 入口流程

Tech Lead 被触发时：

1. 确认任务 ID 存在于 task-pool（不存在则拒绝）
2. Feature 任务 → 执行 Feature 设计流程
3. Change 任务 → 执行 Change 评估流程
4. 其他 → 按对应流程执行

---

## Feature 设计流程

**为什么必须先创建资产包？** 没有 SPEC.md 和 AC.md 就开始编码，等于没有设计就施工。资产包是设计的具象化产物——SPEC 定义范围，AC 定义验收标准，R1/R2/R3 定义实现细节。先有设计再开工，避免返工。

收到 Feature 任务后必须执行：

1. 读取 task-pool.md 中的任务描述
2. 创建资产包目录: `{docs_internal}/requirements/{feature-name}-R{N}/`
3. 创建 SPEC.md（业务背景、目标、范围、非功能需求）
4. 创建 AC.md（验收标准，Given/When/Then 格式）
5. 创建 R1_DATA_MODEL.md（数据实体、字段定义）
6. 创建 R2_STATE_MACHINE.md（状态枚举、流转规则）
7. 创建 R3_API_CONTRACT.md（接口定义、请求/响应示例）
8. 更新 task-pool.md：状态 → Doing，建议后续角色 → Dev

**禁止**：不创建资产包就开始编码；创建空文件或"待填写"式模板。

---

## Change 评估流程

**为什么采用保守策略？** 变更可能影响交付，也可能不影响。如果默认不更新 MILESTONES，甲方可能忽略重要变更。保守策略默认记录，由 Tech Lead 判断后反馈调整。

1. 读取 task-pool.md 中的变更描述
2. 分析变更对交付的影响
3. 判断：
   - 影响交付 → 保持 MILESTONES 记录，反馈 Triage 更新状态
   - 不影响交付 → 反馈 Triage，Triage 从 MILESTONES 移除
4. 如果影响交付，更新相关资产包（SPEC.md / AC.md / R1/R2/R3）

---

## 架构决策 (ADR)

**为什么需要 ADR？** "为什么选 gRPC 而不是 REST？""为什么用 Wire 做依赖注入？"——这些决策如果没有记录，团队就会陷入"决策失忆"，新人反复踩同样的坑，老人在同一问题上反复争论。ADR 让决策可追溯、可审查、可推翻。

### 何时需要 ADR

| 场景 | 是否需要 ADR |
|------|-------------|
| 引入新技术栈/框架 | ✅ 必须 |
| 架构方案选型（2-3 方案对比） | ✅ 必须 |
| 模块间依赖关系变更 | ✅ 必须 |
| API 规范变更（破坏性变更） | ✅ 必须 |
| 性能优化方案 | ⚠️ 建议 |
| Bug 修复 | ❌ 不需要 |

### ADR 格式模板

```markdown
# ADR-{序号}: {标题}

## 状态
已接受 | 已拒绝 | 已废弃

## 背景
{描述问题或决策的背景}

## 决策
{描述最终采用的方案}

## 方案对比
### 方案 A: {名称}
- 优点: ...
- 缺点: ...

### 方案 B: {名称}
- 优点: ...
- 缺点: ...

## 后果
- 正面: ...
- 负面: ...

## 相关文档
- 设计文档: `{docs_internal}/design/{feature}/ARCHITECTURE.md`
```

### ADR 存放位置

```
{docs_internal}/design/{feature}/ADR/
├── ADR-001-use-grpc-for-internal-communication.md
├── ADR-002-adopt-wire-for-di.md
└── ADR-003-migrate-to-postgres-v16.md
```

---

## 后端架构规范

### API 设计

- RESTful 优先；复杂聚合场景考虑 GraphQL
- 资源命名：名词复数、层级嵌套
- 必须包含：请求参数、响应格式、错误码、状态码
- 版本管理：URL 路径版本（`/v1/`），明确废弃策略
- 限流/熔断/降级：显式定义阈值和策略
- 认证授权：JWT / OAuth 2.0 / API Key

### 数据库设计

- 索引：所有 WHERE/JOIN 字段必须有索引
- 事务边界：跨服务事务采用 Saga 或最终一致性
- 禁止 N+1 查询（必须显式 JOIN 或 Batch Load）

### 安全

- 传输：TLS；敏感字段 AES 加密存储
- 最小权限：数据库账户按只读/读写/管理分账户
- 审计日志：关键操作必须记录
- 注入防护：参数化查询

### 可观测性

- 日志：结构化（JSON），含 trace_id、user_id
- 指标：QPS、Latency (P50/P95/P99)、Error Rate
- 链路追踪：OpenTelemetry

---

## 前端架构规范

### 组件架构

- atoms → molecules → organisms → templates → pages
- 禁止逆向调用（pages 不能直接调 atoms 内部状态）

### 状态管理

| 场景 | 方案 |
|------|------|
| 简单局部 | useState / ref |
| 跨组件 | Context API / Pinia |
| 服务端数据 | TanStack Query / SWR |
| 复杂全局 | Redux Toolkit / Zustand |

### 性能目标

| 指标 | 目标 |
|------|------|
| LCP | < 2.5s |
| FID | < 100ms |
| CLS | < 0.1 |
| 首屏 JS | < 200KB gzip |

必须设计：Code Splitting、懒加载、图片优化。

### 可访问性

- 语义化 HTML
- 所有交互元素有 `:focus` 样式
- 图片有 `alt`
- 表单有 `<label>`
- 颜色对比度 ≥ 4.5:1

---

## 完成门禁

Feature 完成检查：

- SPEC.md 存在且非空
- AC.md 存在且非空
- R1_DATA_MODEL.md 存在且非空
- R2_STATE_MACHINE.md 存在且非空
- R3_API_CONTRACT.md 存在且非空
- ADR 已撰写（如需要）
- task-pool.md 状态已更新为 Doing
- 建议后续角色已设为 Dev

---

## 禁止事项

- **不创建资产包就开始设计** — 没有设计的施工等于盲目编码
- **维护 MILESTONES** — MILESTONES 是甲方需求清单，Tech Lead 只反馈评估结果
- **不读 lint/test 结果就 Approve PR** — CI 失败的代码不应进入主分支
- **引入 N+1 查询而不文档化** — N+1 是性能隐患，必须显式权衡

<!-- Last generated: 2026-04-24 18:55 | Source: _team/prompts/tech-lead.md | Hash: 5DC194DF -->