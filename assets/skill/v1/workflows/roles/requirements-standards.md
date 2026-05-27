# 需求定义输出规范

> **角色**: PM / Analyst
> **加载文件**: `{TEAM_PATH}/workflows/roles/requirements-standards.md`
> **前置**: 先读 `{TEAM_PATH}/workflows/shared.md` 了解通用原则

---

## 1. 触发条件

- 用户说：**"需求 xxx"**、**"PRD xxx"**、**"功能需求 xxx"**、**"开发 xxx 功能"**
- 任何未在需求文档中记录的功能点，都视为需求未完成

---

## 2. 需求完整性检查（强制）

> **开发前置条件**：以下每一项必须在需求文档中明确，不得在开发过程中临时决定。

### 核心四要素（缺一不可）

| 编号 | 要素 | 问题 | 如不确定则 |
|------|------|------|-----------|
| R1 | **数据模型** | 有哪些字段？类型是什么？哪些必填？ | 列出假设并标注 `[待确认]` |
| R2 | **状态机** | 有哪些状态？哪些转换？初始/结束？ | 设计枚举，不允许"进行中"模糊词 |
| R3 | **接口契约** | 谁调谁？请求/响应格式？错误码？ | 先定义契约再写代码 |
| R4 | **异常处理** | 失败怎么处理？重试？回滚？超时？ | 列出所有异常分支 |

### 扩展要素（按需必填）

| 编号 | 要素 | 问题 | 如不确定则 |
|------|------|------|-----------|
| R5 | **依赖关系** | 依赖哪些外部服务/模块？ | 列出并标注状态 |
| R6 | **边界条件** | 空值/最大值/并发？ | 列出测试点 |
| R7 | **权限控制** | 谁能操作什么？ | 列出角色权限矩阵 |
| R8 | **验收标准** | 怎么算完成？ | 用 BDD Given/When/Then |
| R9 | **UI/UX 规格** | 页面长什么样？组件用哪些？交互行为？ | 引用 ui-standards.md + UI Designer 产出 Feature UI Design |

### AI 开发铁律

- 缺少 R1-R4 任一项 → **必须补充后才能开始开发**
- 涉及 UI 的功能缺少 R9 → **必须等 UI Design 文档完成后才能开始前端开发**
- 不确定某项 → 标注 `[待确认]` + 向用户提问，不得自行假设
- 开发过程中发现新要素 → **先更新需求文档，再改代码**

---

## 3. 标准需求文档格式

```markdown
# 需求文档: {功能名称}

**版本**: v{version}
**更新时间**: {YYYY-MM-DD HH:mm}
**状态**: [起草/评审中/已确认/开发中/已完成]

---

## 一、业务背景

### 1.1 业务背景
[为什么需要这个功能]

### 1.2 问题陈述
[当前存在的问题或痛点]

### 1.3 目标
[本次需求要达成的目标（可量化）]
```

---

## 二、核心开发四要素

> **强制章节**。每一项都必须明确填写，不允许跳过或模糊描述。

### 2.1 数据模型定义

#### 实体: {EntityName}

| 字段名 | 类型 | 必填 | 说明 | 约束 |
|--------|------|------|------|------|
| id | uint64 | Y | 主键 | 自增 |
| name | string | Y | 名称 | 最大 64 字符 |
| status | enum | Y | 状态 | 见状态机定义 |
| created_at | timestamp | Y | 创建时间 | - |
| [field] | [type] | Y/N | [说明] | [约束] |

#### 关系图
```
{EntityA} 1──N {EntityB}
{EntityA} 1──1 {EntityC}
```

---

### 2.2 状态机定义

#### 状态枚举
```go
const (
    StatusPending    = "pending"     // 待处理
    StatusProcessing = "processing"  // 处理中
    StatusSuccess    = "success"    // 成功
    StatusFailed     = "failed"     // 失败
    StatusCancelled  = "cancelled"   // 已取消
)
```

#### 状态流转图
```
  ┌─────────┐
  │ pending │
  └────┬────┘
       │ submit
       ▼
  ┌───────────┐
  │processing │
  └─────┬─────┘
        │ complete/fail
    ┌───┴───┐
    ▼       ▼
 completed  failed
```

#### 状态流转规则

| 当前状态 | 事件 | 目标状态 | 前置条件 | 副作用 |
|---------|------|---------|---------|--------|
| pending | submit | processing | 参数校验通过 | 发送通知 |
| processing | complete | success | 业务逻辑完成 | 更新统计 |
| processing | fail | failed | 发生错误 | 记录错误日志 |
| failed | retry | processing | 重试次数 < 3 | - |
| pending | cancel | cancelled | 用户主动取消 | 释放资源 |

#### 异常状态处理

| 异常状态 | 触发条件 | 处理方式 |
|---------|---------|---------|
| 超时 | processing 持续 > 30min | 自动标记 failed，发送告警 |

---

### 2.3 接口契约定义

#### REST API

##### 创建 {Entity}

**Endpoint**: `POST /api/v1/{entity}`
**权限**: 角色要求（见权限矩阵）
**超时**: 5s

**Request**:
```json
{
  "name": "string (必填, 1-64字符)",
  "description": "string (可选, 最大256字符)"
}
```

**Response** (201 Created):
```json
{
  "id": 12345,
  "name": "xxx",
  "status": "pending",
  "created_at": "2026-04-16T12:00:00Z"
}
```

**Error Responses**:

| Code | 条件 | Body |
|------|------|------|
| 400 | 参数校验失败 | `{"error": "INVALID_PARAM", "message": "name is required"}` |
| 401 | 未认证 | `{"error": "UNAUTHORIZED"}` |
| 403 | 无权限 | `{"error": "FORBIDDEN"}` |
| 409 | 资源冲突 | `{"error": "CONFLICT"}` |

##### 查询 {Entity} 状态

**Endpoint**: `GET /api/v1/{entity}/{id}`
**权限**: 所有者或管理员
**响应时间**: < 100ms P95

**Response** (200 OK):
```json
{
  "id": 12345,
  "name": "xxx",
  "status": "processing",
  "progress": 45,
  "created_at": "2026-04-16T12:00:00Z"
}
```

#### 内部 gRPC（如果适用）

**Service**: `{Entity}Service`
**Proto**:
```protobuf
rpc Create{Entity}(Create{Entity}Request) returns (Create{Entity}Response);

message Create{Entity}Request {
    string name = 1 [(validate.rules).string.min_len = 1];
}
message Create{Entity}Response {
    uint64 id = 1;
    string status = 2;
}
```

---

### 2.4 异常处理设计

#### 异常分类与处理策略

| 异常类型 | 示例 | 处理策略 | 用户感知 |
|---------|------|---------|---------|
| 参数错误 | 字段缺失/格式错误 | 立即返回 400 | 提示具体错误 |
| 业务错误 | 余额不足 | 事务回滚，返回 422 | 友好提示 |
| 外部依赖错误 | 第三方超时 | 重试 N 次后失败 | 稍后重试 |
| 系统错误 | 数据库连接失败 | 记录日志，告警 | 服务异常稍后 |

#### 异常响应格式（统一）
```json
{
  "error": "ERROR_CODE",
  "message": "用户可见的错误描述",
  "details": { "field": "name", "reason": "exceeds max length" },
  "request_id": "uuid用于追踪"
}
```

#### 重试策略
```yaml
retry:
  max_attempts: 3
  backoff: exponential
  initial_interval: 100ms
  max_interval: 5s
```

---

## 三、扩展开发要素

### 3.1 边界条件定义

| 边界场景 | 输入 | 预期行为 |
|---------|------|---------|
| 空名称 | `name: ""` | 返回 400，提示 name 必填 |
| 超长名称 | `name: "xxx...(>64)"` | 返回 400，提示最大长度 |
| 并发创建 | 同一 name 同时请求 | 只有一个成功，另一个返回 409 |
| 外部服务超时 | 依赖服务 > 5s 无响应 | 超时失败，触发重试 |

### 3.2 权限控制矩阵

| 操作 | 管理员 | 普通用户 | 访客 |
|------|--------|---------|------|
| 创建 | Y | N | N |
| 查询自己的 | Y | Y | N |
| 查询他人的 | Y | N | N |
| 更新 | Y | 仅自己的 | N |
| 删除 | Y | 仅自己的 | N |

### 3.3 依赖关系

#### 前置依赖（必须先完成）

| 依赖项 | 状态 | 备注 |
|--------|------|------|
| 用户认证模块 | [Y/开发中/待开发] | JWT 验权 |
| 数据库迁移脚本 | [Y/待写] | - |

#### 外部依赖

| 依赖项 | 接口 | SLA | 降级策略 |
|--------|------|------|---------|
| 第三方支付 | REST | 99.9% | 本地重试 |

---

## 四、用户故事与验收标准

### 4.1 用户故事

**故事 1: 创建 {Entity}**
- **作为**: {角色}
- **我想要**: 创建 {Entity}
- **以便**: {业务价值}

### 4.2 验收标准（BDD 格式）

#### AC1: 成功创建 {Entity}
```gherkin
Feature: {Entity} 创建

  Scenario: 管理员成功创建 {Entity}
    Given 用户已登录为管理员
    And   请求参数合法
    When  提交创建请求
    Then  返回 201
    And   状态为 pending
    And   返回 id
```

#### AC2: 参数校验失败
```gherkin
  Scenario: 必填字段缺失
    Given 用户已登录为管理员
    And   name 字段为空
    When  提交创建请求
    Then  返回 400
    And   错误信息包含 "name is required"
```

#### AC3: 权限不足
```gherkin
  Scenario: 普通用户尝试创建
    Given 用户已登录为普通用户
    When  提交创建请求
    Then  返回 403
    And   错误信息包含 "no permission"
```

---

## 五、优先级与里程碑

### 优先级

| 级别 | 需求 | 说明 |
|------|------|------|
| **Must** | [R1-R4 核心] | 必须完成才能上线 |
| **Should** | [扩展功能] | 重要但不影响主流程 |
| **Could** | [优化项] | 锦上添花 |

### 里程碑

| 里程碑 | 日期 | 交付物 | 完成标准 |
|--------|------|--------|---------|
| M1: 需求确认 | {date} | 本文档 | R1-R4 全部确认 |
| M2: 设计完成 | {date} | 架构设计 + 接口定义 | 评审通过 |
| M3: 开发完成 | {date} | 代码 + 单元测试 | 覆盖率 > 70% |
| M4: 测试完成 | {date} | 测试报告 | 所有 AC 通过 |
| M5: 上线 | {date} | 部署文档 | 生产验证通过 |

---

## 六、需求评审检查清单

> **需求评审通过前，不开始开发。**

### 六要素检查

| 检查项 | 要求 | 状态 |
|--------|------|------|
| 数据模型 | 每个字段有类型/必填/约束 | [ ] |
| 状态机 | 状态枚举完整，流转无遗漏，异常状态有定义 | [ ] |
| 接口契约 | 请求/响应格式明确，错误码覆盖所有场景 | [ ] |
| 异常处理 | 每种异常有处理策略，用户感知明确 | [ ] |
| 边界条件 | 空值/最大/并发/超时场景全部列出 | [ ] |
| 验收标准 | 每个核心功能有 BDD 格式的 Given/When/Then | [ ] |

### 开发前检查

- [ ] R1-R4 四要素全部确认（非 `[待确认]`）
- [ ] 依赖模块/服务状态确认
- [ ] 数据库 schema 变更已规划
- [ ] 权限矩阵已确认
- [ ] 评审人签字（PM + Tech Lead）

### 需求变更检查

- [ ] 变更影响范围已评估
- [ ] 相关文档已同步更新
- [ ] 变更原因已记录
- [ ] 评审人重新签字

---

## 7. 输出成果清单

| 成果类型 | 文件路径 |
|---------|---------|
| **需求文档** | `{docs_internal}/requirements/{feature}/SPEC.md` |
| **接口定义** | `{docs_internal}/requirements/{feature}/api.md`（可选） |
| **用户故事** | `{docs_internal}/requirements/{feature}/USER_STORIES.md` |
| **验收标准** | `{docs_internal}/requirements/{feature}/ACCEPTANCE.md` |
| **需求追踪** | `{docs_internal}/requirements/{feature}/TRACKING.md` |

### 目录结构

```
{docs_internal}/requirements/{feature}/
├── SPEC.md              # 需求文档（主文档）
├── USER_STORIES.md      # 用户故事
├── ACCEPTANCE.md        # 验收标准
├── TRACKING.md          # 需求状态追踪表
└── api.md               # 接口详细定义（可选）
```
