# 功能开发输出规范 — team-flow v2 (beads-native)

> **角色**: Backend / Frontend Dev
> **加载文件**: `{TEAM_PATH}/workflows/roles/development-standards.md`
> **前置**: 先读 `{TEAM_PATH}/workflows/shared.md` 了解通用原则，再读 `{TEAM_PATH}/templates/gherkin-feature-template.md` 了解验收层 BDD AC 编写规范

---

## 0. 输出质量控制（强制）

> **强制要求**: 每次输出前必须检查是否重复，避免重复内容。

### 0.1 重复检查清单

| 检查项 | 处理方式 |
|--------|----------|
| 章节重复 | 删除重复章节，保留一个 |
| 配置重复 | 引用环境变量，不重复配置 |
| 规范重复 | 引用规范文件，不重复定义 |
| 占位符 {} | 替换为实际值或删除 |

### 0.2 禁止行为

- ❌ 手动复制模板内容到项目文件
- ❌ 在项目文件中保留模板占位符 {xxx}
- ❌ 重复配置已在环境变量中定义的配置
- ❌ 在多个位置定义相同的规范

### 0.3 角色→规范文件映射（强制）

> 根据 project.md 中的团队角色类型，加载对应的规范文件。

| 团队角色 | 加载规范文件 | 职责 |
|---------|------------|------|
| backend-dev | `{TEAM_PATH}/workflows/roles/development-standards.md` + `{TEAM_PATH}/templates/gherkin-feature-template.md` + `{TEAM_PATH}/templates/feature-test-template.md` | 功能开发、TDD、验收层 AC |
| frontend-dev | `{TEAM_PATH}/workflows/roles/development-standards.md` + `{TEAM_PATH}/templates/gherkin-feature-template.md` + `{TEAM_PATH}/templates/frontend-feature-test-template.md` | 前端开发、TDD、验收层 AC |
| 任意 dev（Bug 修复） | `{TEAM_PATH}/workflows/roles/bugfix-standards.md` + `{TEAM_PATH}/templates/bug-test-template.md` | Bug 修复 |
| frontend-dev（Bug 修复） | `{TEAM_PATH}/workflows/roles/bugfix-standards.md` + `{TEAM_PATH}/templates/frontend-bug-test-template.md` | 前端 Bug 修复 |
| qa-engineer | `{TEAM_PATH}/workflows/roles/test-standards.md` + `{TEAM_PATH}/templates/feature-test-template.md` + `{TEAM_PATH}/templates/frontend-feature-test-template.md` + `{TEAM_PATH}/templates/bug-test-template.md` + `{TEAM_PATH}/templates/frontend-bug-test-template.md` | 测试验证 |
| pm | `{TEAM_PATH}/workflows/roles/requirements-standards.md` | 需求管理 |
| devops | `{TEAM_PATH}/workflows/roles/devops-standards.md` | 部署运维 |

**角色来源**：`project.md` 中的 `teamRoles` 字段（由初始化流程创建）。

### 0.4 文档路径（见 shared.md 通用原则）

> **⚠️ 创建任何文档前，必须先按 shared.md 的「文档路径强制规则」判断存放位置！**
>
> 核心规则：内部文档 → `{docs_internal}` | 对外文档 → `{docs_external}` | 不确定 → 默认内部
>
> **必须输出**：
> ```
> 📂 文档路径检查:
>    - 文档类型: [内部/对外]
>    - 存放位置: [{docs_internal}/{docs_external}]
>    - 路径: xxx/xxx.md
> ```

---

## 1. 触发条件

- 用户说：**"添加 xxx 功能"**、**"实现 xxx"**、**"开发 xxx"**、**"新增 xxx"**
- 例如：添加用户管理功能、实现订单查询、开发支付模块

---

## 1.5 Change Impact Analysis（变更影响分析 — 强制）

> **强制要求**: 修改任何已有代码文件前，必须执行变更影响分析。未执行分析直接修改 → ⛔ 禁止。

### 影响分析流程

```
识别到需要修改已有文件
    │
    ├── Step 1: 读取目标文件 → 理解完整上下文
    │
    ├── Step 2: 识别变更类型
    │   ├── 新增代码（不修改已有逻辑）→ 低风险
    │   ├── 修改已有逻辑 → 中风险
    │   └── 修改公共接口/签名 → 高风险（Breaking Change）
    │
    ├── Step 3: 影响范围扫描
    │   ├── 搜索引用点: grep -r "TargetSymbol" --include="*.go"
    │   ├── 搜索依赖方: 谁依赖了这个包/模块？
    │   └── 搜索测试覆盖: 哪些测试覆盖了这段代码？
    │
    ├── Step 4: 风险评级与策略
    │   ├── 低风险 → 直接修改 + 全量回归测试
    │   ├── 中风险 → 输出影响报告 → 修改 + 全量回归测试
    │   └── 高风险 → 输出影响报告 + 兼容方案 → 确认后修改 + 全量回归测试
    │
    └── Step 5: 分层回归验证
        ├── TDD 循环: go test ./current/module/... (局部，秒级)
        ├── 阶段性回归: 每 N 次局部后 go test ./... (全量，分钟级)
        ├── 完成门禁: go test ./... (全量，必须通过)
        └── 任何回归 → ⛔ 停止，修复或回滚
```

### 影响报告格式

```markdown
🔍 Change Impact Report:
   - 目标文件: {file_path}
   - 变更类型: {新增/修改逻辑/修改接口}
   - 风险等级: {低/中/高}
   - 影响文件数: {N}
   - 受影响模块: {module1, module2, ...}
   - Breaking Change: {是/否}
   - 兼容方案: {如有 Breaking Change}
   - 现有测试覆盖: {有/无} → 无测试覆盖时需先补充测试
```

### 禁止

- ❌ 未读完整文件就修改
- ❌ 未搜索引用点就改公共接口
- ❌ 完成后未跑全量测试就标记完成
- ❌ 开发中每次都跑全量测试（太慢，应先局部后全量）
- ❌ 发现回归失败后继续开发
- ❌ 对无测试覆盖的代码做修改而不先补充测试

---

## 2. 功能开发的迭代特性

> 功能开发不是一次性任务，而是一个迭代过程。需要先识别开发意图，再决定输出内容。

```
需求定义 → 参考分析 ← ← ← ← ← ←
    ↓                                 ↑
设计规划 ← ← ← ← ← ← ← ← ← ← ↑
    ↓                                 ↑ 设计调整
迭代开发 ← ← ← ← ← ← ← ← ← ← ↑
    ↓                                 ↑ 新参考需求
测试验证 ← ← ← ← ← ← ← ← ← ← ↑
    ↓                                 ↑ 设计变更
上线交付
```

### 2.1 常见情况处理

| 情况 | 处理方式 |
|------|---------|
| 需要参考 | 触发参考分析任务，记录新的参考需求 |
| 设计调整 | 记录设计变更，更新版本 |
| 发现 Bug | 记录 Bug，继续开发或暂停修复 |
| 涉及较大架构调整 | 评估变更级别，决定新版本号 |

---

## 3. 测试前置设计规范（强制）

> **强制要求**: 功能开发前必须完成测试前置设计。缺少任一项视为设计不完整，不得进入开发阶段。

### 3.1 状态机设计模板

```markdown
# 状态机设计: {功能名称}

---

## 一、状态枚举定义

| 状态值 | 显示名称 | 说明 | 是否终态 |
|--------|---------|------|---------|
| pending | 等待中 | 初始状态 | N |
| processing | 处理中 | 正在处理 | N |
| completed | 已完成 | 成功结束 | Y |
| failed | 失败 | 可重试 | N |
| cancelled | 已取消 | 用户主动取消 | Y |

## 二、状态流转图

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

## 三、状态流转规则

| 当前状态 | 事件 | 目标状态 | 前置条件 | 副作用 |
|---------|------|---------|---------|--------|
| pending | submit | processing | 参数校验通过 | 发送通知 |
| processing | complete | completed | 业务逻辑完成 | 更新统计 |
| processing | fail | failed | 发生错误 | 记录错误日志 |
| failed | retry | processing | 重试次数 < 3 | - |
| pending | cancel | cancelled | 用户主动取消 | 释放资源 |

## 四、状态字段定义

| 字段名 | 类型 | 说明 | 约束 |
|--------|------|------|------|
| status | enum | 当前状态 | 必须从枚举值中选择 |
| status_changed_at | timestamp | 状态变更时间 | 自动更新 |
| error_code | string | 错误码（失败时） | 仅 failed 时填写 |

## 五、可观测性要求

| 要求 | 说明 |
|------|------|
| 日志 | 每次状态变更必须记录日志 |
| 事件 | 关键状态变更（completed/failed）必须发布事件 |
| 监控 | processing > 10min 需要告警 |
```

### 3.2 数据模型定义模板

```markdown
# 数据模型定义: {功能名称}

---

## 一、数据表设计

### {table_name} 表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uuid | PK | 主键 |
| status | enum | NOT NULL | 状态 |
| created_at | timestamp | NOT NULL | 创建时间 |
| updated_at | timestamp | NOT NULL | 更新时间 |

---

## 二、字段按状态定义

### 2.1 状态: pending（进入条件: 创建记录时）

| 字段 | 必填 | 值/说明 |
|------|------|--------|
| id | Y | UUID 自动生成 |
| status | Y | pending 固定值 |
| created_at | Y | 当前时间 |
| sha256 | N | null（此时未计算） |
| completed_at | N | null（此时未完成） |

### 2.2 状态: processing（进入条件: 开始处理时）

| 字段 | 必填 | 值/说明 |
|------|------|--------|
| status | Y | processing |
| sha256 | Y | 必须计算 |
| started_at | Y | 当前时间 |
| progress | Y | 0-100 |

### 2.3 状态: completed（进入条件: 处理成功时）

| 字段 | 必填 | 值/说明 |
|------|------|--------|
| status | Y | completed |
| progress | Y | 100 |
| completed_at | Y | 当前时间 |

### 2.4 状态: failed（进入条件: 处理异常时）

| 字段 | 必填 | 值/说明 |
|------|------|--------|
| status | Y | failed |
| error_code | Y | 必填 |
| error_message | Y | 必填 |
| retry_count | Y | 0-N |

---

## 三、字段变更约束

| 约束类型 | 说明 | 示例 |
|---------|------|------|
| 不可变更 | 某些字段一旦写入不可修改 | id, created_at |
| 状态锁定 | 某些字段只能在特定状态下修改 | sha256 仅 processing 时可写 |
| 级联更新 | 状态变更时必须同时更新 | updated_at 必须随 status 更新 |

## 四、数据校验规则

| 字段 | 校验规则 | 错误信息 |
|------|---------|--------|
| file_size | > 0 | 文件大小必须大于 0 |
| file_size | ≤ 10GB | 文件大小不能超过 10GB |
| sha256 | 长度为 64 | SHA256 格式错误 |
```

### 3.3 接口契约定义模板

```markdown
# 接口契约定义: {功能名称}

---

## 一、接口概览

| 属性 | 值 |
|------|-----|
| HTTP 方法 | GET/POST/PUT/DELETE |
| 路径 | /api/v1/{resource} |
| 认证要求 | Bearer Token |
| 超时 | 5s |

## 二、请求规范

### 请求头

| 头部 | 必填 | 说明 |
|------|------|------|
| Authorization | Y | Bearer {token} |
| Content-Type | Y | application/json |

### 请求参数（Path / Query / Body）

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | Y | 名称，最大 64 字符 |

## 三、响应规范

### 成功响应 (200 / 201)

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 12345,
    "status": "pending",
    "created_at": "2026-04-16T12:00:00Z"
  },
  "request_id": "uuid"
}
```

### 字段语义定义（data）

| 字段名 | 类型 | 说明 | 出现条件 |
|--------|------|------|---------|
| id | uuid | 资源唯一标识 | 始终 |
| status | string | 当前状态 | 始终 |
| progress | int | 进度百分比 | pending/processing 时 |
| error.code | string | 错误码 | failed 时 |

### 错误响应

| Code | 条件 | Body |
|------|------|------|
| 400 | 参数校验失败 | `{"error": "INVALID_PARAM", "message": "..."}` |
| 401 | 未认证 | `{"error": "UNAUTHORIZED"}` |
| 403 | 无权限 | `{"error": "FORBIDDEN"}` |
| 404 | 资源不存在 | `{"error": "NOT_FOUND"}` |
| 409 | 资源冲突 | `{"error": "CONFLICT"}` |
| 500 | 服务器错误 | `{"error": "INTERNAL_ERROR"}` |

### 异常分类与处理策略

| 异常类型 | 示例 | 处理策略 | 用户感知 |
|---------|------|---------|---------|
| 参数错误 | 字段缺失/格式错误 | 立即返回 400 | 提示具体错误 |
| 业务错误 | 余额不足 | 事务回滚，返回 422 | 友好提示 |
| 外部依赖错误 | 第三方超时 | 重试 N 次后失败 | 稍后重试 |
| 系统错误 | 数据库连接失败 | 记录日志，告警 | 服务异常稍后 |

### 重试策略

```yaml
retry:
  max_attempts: 3
  backoff: exponential
  initial_interval: 100ms
  max_interval: 5s
```
```

### 3.4 测试点拆解模板

> **重要**: 本模板用于开发前的测试设计，正式输出时请同时参考 `feature-test-template.md` 生成测试覆盖声明。

```markdown
# 测试点拆解: {功能名称}

---

## 一、测试分层

| 层级 | 主导 | 验证什么 |
|------|------|---------|
| 后端单元测试 | Backend Dev | 业务逻辑正确 |
| 后端集成测试 | Backend Dev | 数据库操作正确 |
| API 层测试 | Backend/QA | 接口契约正确 |
| 前端单元测试 | Frontend Dev | 组件逻辑正确 |
| E2E 测试 | QA Engineer | 业务流程完整 |

---

## 二、按状态拆解测试点

### 2.1 pending

| ID | 描述 | 类型 | 预期结果 | 优先级 |
|----|------|------|---------|--------|
| PENDING-001 | 创建资源 status=pending | 单元测试 | 字段正确写入 | P0 |
| PENDING-002 | file_name 为空时创建 | API 测试 | 返回 400 | P0 |
| PENDING-003 | 未认证创建 | API 测试 | 返回 401 | P0 |

**边界情况**: 空名称、超长名称、快速连续创建

### 2.2 processing

| ID | 描述 | 类型 | 预期结果 | 优先级 |
|----|------|------|---------|--------|
| PROC-001 | pending→processing 流转 | 单元测试 | 状态正确变更 | P0 |
| PROC-002 | sha256 已计算 | 集成测试 | sha256 有值 | P0 |

### 2.3 completed

| ID | 描述 | 类型 | 预期结果 | 优先级 |
|----|------|------|---------|--------|
| COMP-001 | processing→completed 流转 | 单元测试 | 状态正确 | P0 |
| COMP-002 | progress=100 | API 测试 | 进度为 100 | P0 |

### 2.4 failed

| ID | 描述 | 类型 | 预期结果 | 优先级 |
|----|------|------|---------|--------|
| FAIL-001 | 处理异常时状态变为 failed | 单元测试 | 状态正确 | P0 |
| FAIL-002 | retryable=true 时可重试 | API 测试 | 重新发起成功 | P0 |

---

## 三、按接口拆解测试点

### POST /api/v1/{entity}

| ID | 描述 | 预期结果 |
|----|------|---------|
| API-001 | 正常创建 | 返回 id, status=pending |
| API-002 | 缺少必填字段 | 返回 400 |
| API-003 | 未认证 | 返回 401 |

### GET /api/v1/{entity}/{id}

| ID | 描述 | 预期结果 |
|----|------|---------|
| API-101 | 查询存在的资源 | 返回完整信息 |
| API-102 | 查询不存在的资源 | 返回 404 |
| API-103 | pending 时 sha256=null | 字段存在但为 null |
```

---

## 4. 版本管理体系

### 4.1 版本变更级别

| 变更级别 | 判断标准 | 版本命名 |
|---------|---------|---------|
| **Patch** | 仅修复Bug、字段微调、不影响设计 | v{major}.{minor}.{patch+1} |
| **Minor** | 部分模块调整、增加字段、不影响核心架构 | v{major}.{minor+1}.0 |
| **Major** | 核心设计变更、架构调整、方案重构 | v{major+1}.0.0 |

### 4.2 版本血缘记录

```markdown
# 版本元数据: v{version}

## 基本信息

| 字段 | 值 |
|------|-----|
| 版本号 | v{version} |
| 名称 | {name} |
| 状态 | proposed / developing / testing / released / deprecated |
| 创建时间 | {YYYY-MM-DD} |

## 版本血缘

| 字段 | 值 |
|------|-----|
| 父版本 | v{parent} |
| 来源 | 新开发 / 回退自 v{废弃版本} |
| 变更级别 | Minor |
```

### 4.3 版本血缘图（VERSIONS.md）

```markdown
# {功能名称} 版本血缘总览

**当前基线版本**: v{baseline}
**最新版本**: v{latest}

## 版本血缘图

```
v1.0 ──▶ v1.0.1 ──▶ v1.1 (proposed)
  │                    ▲
  │                    │
  └──▶ v2.0 (abandoned)┘
            │  回退
            └──▶ v1.1（继承敏感词模块，丢弃消息队列）
```

## 版本清单

| 版本 | 状态 | 父版本 | 变更级别 | 说明 |
|------|------|-------|---------|------|
| v1.1 | proposed | v1.0 | Minor | 回退自 v2.0 |
| v2.0 | abandoned | v1.0 | Major | 已废弃 |
| v1.0 | released | - | - | 初始版本 |

## 回退记录

| 废弃版本 | 回退目标 | 回退原因 |
|---------|---------|---------|
| v2.0 | v1.1 | 架构变更过大，风险高 |
```

---

## 5. 迭代记录格式

```markdown
# {功能名称} v{version} 迭代记录

---

## 迭代概览

| 迭代 | 时间 | 负责人 | 完成内容 | 状态 |
|------|------|--------|---------|------|
| 迭代 1 | 04-01~04-05 | xxx | 评论审核模块 | Y 完成 |
| 迭代 2 | 04-06~04-10 | xxx | 敏感词过滤 | Y 完成 |

---

## 迭代详情

### 迭代 N: {迭代名称}

**迭代周期**: {start} ~ {end}
**迭代目标**: {本迭代要完成的内容}

#### 上次迭代遗留
- [ ] 遗留项 1

#### 本迭代工作

##### 新增/修改内容

| 类型 | 模块/文件 | 说明 |
|------|---------|------|
| 新增 | `biz/xxx.go` | xxx 功能 |
| 修改 | `api/xxx.proto` | 增加 xxx 字段 |

##### 发现的 Bug

| Bug ID | 描述 | 严重程度 | 状态 |
|--------|------|---------|------|
| BUG-001 | xxx | Major | 已修复 |

##### 遇到的问题

| 问题 | 描述 | 处理方式 |
|------|------|---------|
| 问题 1 | xxx | 参考了新文档，调整了方案 |

##### 设计变更

| 变更ID | 变更内容 | 原因 | 影响范围 |
|--------|---------|------|---------|
| CHANGE-001 | 字段调整 | xxx | 低 |

##### 新的参考需求

| 参考需求 | 描述 | 状态 |
|---------|------|------|
| REF-001 | xxx | 待分析 |

#### 迭代结果

##### 完成情况
- [x] 完成项 1
- [ ] 延期项（移到下个迭代）

##### 测试覆盖

| 模块 | 单元测试 | 集成测试 | E2E |
|-----|---------|---------|-----|
| 模块 A | Y 85% | Y | Y |

##### 下个迭代计划
- [ ] 完成模块 B 的 E2E 测试
- [ ] 开始模块 C 的开发
```

---

## 6. 设计变更记录格式

```markdown
# 设计变更记录: {功能名称}

---

## 变更清单

### CHANGE-{n}: {变更标题}

**变更日期**: {date}
**变更人**: {name}
**变更阶段**: 需求/设计/开发/测试

#### 变更内容

| 类型 | 原设计 | 新设计 | 变更说明 |
|------|-------|-------|---------|
| 数据字段 | age INT | age BIGINT | 支持更大范围 |
| 接口设计 | POST /api/v1/ | POST /api/v2/ | 版本升级 |

#### 变更原因
[为什么要变更]

#### 影响评估

| 影响范围 | 影响程度 |
|---------|---------|
| 后端代码 | 高 |
| 前端代码 | 中 |
| 数据库 | 高 |

#### 关联迭代
- 迭代 1（第 2 次设计变更）

#### 审批记录

| 角色 | 审批人 | 日期 | 意见 |
|------|--------|------|------|
| Tech Lead | {name} | {date} | 同意 |
```

---

## 7. 回退处理流程

```markdown
# 回退记录: v{废弃版本} → v{目标版本}

**废弃版本**: v{废弃版本}
**目标版本**: v{目标版本}
**回退日期**: {YYYY-MM-DD}

---

## 一、回退原因

| 问题 | 描述 | 影响 |
|------|------|------|
| 架构复杂 | 微服务拆分过多 | 运维成本高 |

## 二、变更评估

### 2.1 变更清单

| 变更ID | 变更描述 | 变更级别 | 是否保留 |
|--------|---------|---------|---------|
| CHANGE-001 | 敏感词模块 | Minor | Y 保留 |
| CHANGE-002 | 消息队列重构 | Major | N 丢弃 |

**最高变更级别**: Minor → 新版本号: v{目标版本}

---

## 三、继承与丢弃

### 继承自 v{废弃版本}

| 变更项 | 说明 | 处理方式 |
|-------|------|---------|
| 敏感词模块 | 保留 | 直接合并到 v{目标版本} |

### 丢弃自 v{废弃版本}

| 变更项 | 丢弃原因 |
|-------|---------|
| 消息队列重构 | 架构变更过大 |

---

## 四、回退执行

- [ ] 创建 v{目标版本} 版本目录
- [ ] 迁移保留的变更
- [ ] 更新 VERSION.md
- [ ] 标记 v{废弃版本} 为 abandoned
```

---

## 8. 版本交付报告

```markdown
# 功能开发交付报告: {功能名称} v{version}

---

## 一、交付概览

| 字段 | 值 |
|------|-----|
| 版本号 | v{version} |
| 状态 | released |
| 父版本 | v{parent} |
| 变更级别 | Minor |

### 开发统计

| 指标 | 数值 |
|------|------|
| 开发周期 | {n} 天 |
| 迭代次数 | {n} |
| 测试覆盖率 | {n}% |

---

## 二、交付内容

### 代码变更

| 文件路径 | 变更说明 | 类型 |
|---------|---------|------|
| `internal/features/{feature}/biz/xxx.go` | 业务逻辑 | 新增 |
| `api/{feature}/v1/xxx.proto` | 新增 RPC 方法 | 新增 |

### API 变更

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/users | 创建用户 |

---

## 三、自测要求（强制）

> **强制要求**: 每次代码修改后必须执行自测，未通过自测不得提交或汇报完成。

### 3.1 自测流程

> ⚠️ **必须先执行命令检测（见 SKILL.md 步骤 5），不得直接使用下方命令！**
>
> 检测结果映射：

| 项目类型 | 检测文件 | 后端命令 | 前端命令 |
|---------|---------|---------|---------|
| Go 后端 | `go.mod` | `go build ./...` `go test ./...` `go fmt ./...` | — |
| 前端 | `bun.lockb` / `pnpm-lock.yaml` / `package-lock.json` / `yarn.lock` | — | ⚠️ 根据检测到的包管理器动态替换 |

**自测时用检测到的实际命令替换 `{CMD}`。**

| 步骤 | 检查项 | 命令（示例：检测到 Go 后端） |
| 1 | 编译检查 | `go build ./...` |
| 2 | 单元测试 | `go test ./...` |
| 3 | 格式化验证 | `go fmt ./...` |

### 3.2 自测未通过处理

| 检查项 | 未通过 | 处理方式 |
|--------|-------|----------|
| 编译 | 编译错误 | 修复后重新自测 |
| 测试 | 测试失败 | 修复或补充测试用例后重新自测 |
| 格式化 | 格式问题 | 执行格式化后重新自测 |

---

## 四、测试结果

| 模块 | 单元测试 | 集成测试 | E2E |
|-----|---------|---------|-----|
| 模块 A | Y 85% | Y | Y |
| 模块 B | Y 78% | Y | N 待补充 |

---

## 四、交付检查

- [x] 代码格式化通过
- [x] 静态检查通过
- [x] 单元测试覆盖率 ≥ 80%
- [x] 编译成功
- [x] API 文档已更新
```

---

## 9. 完整目录结构

```
{docs_internal}/delivery/{feature}/
├── VERSIONS.md                 # 版本血缘总览（核心入口）
│
├── versions/                   # 版本目录
│   ├── v1.0/
│   │   ├── VERSION.md          # 版本元数据
│   │   ├── design/
│   │   │   ├── STATE_MACHINE.md
│   │   │   ├── DATA_MODEL.md
│   │   │   ├── API_CONTRACT.md
│   │   │   └── changes/       # 设计变更记录
│   │   ├── iterations/        # 迭代记录
│   │   │   ├── iteration-1.md
│   │   │   └── iteration-2.md
│   │   ├── test/
│   │   │   └── REPORT.md
│   │   └── delivery-report.md
│   │
│   ├── v2.0/                  # 已废弃
│   │   └── VERSION.md          # 标记为 abandoned
│   │
│   └── v1.1/                  # 回退后新版本
│
├── proposals/                  # 提案目录（未立项）
│
└── rollbacks/                 # 回退记录
    └── v2.0-to-v1.1.md
```

---

## 10. 输出成果清单

| 成果类型 | 文件路径 |
|---------|---------|
| 版本总览 | `{docs_internal}/delivery/{feature}/VERSIONS.md` |
| 版本元数据 | `versions/{version}/VERSION.md` |
| 状态机设计 | `versions/{version}/design/STATE_MACHINE.md` |
| 数据模型定义 | `versions/{version}/design/DATA_MODEL.md` |
| 接口契约定义 | `versions/{version}/design/API_CONTRACT.md` |
| 迭代记录 | `versions/{version}/iterations/iteration-{n}.md` |
| 测试报告 | `versions/{version}/test/REPORT.md` |
| 交付报告 | `versions/{version}/delivery-report.md` |
| 回退记录 | `rollbacks/v{from}-{v{to}.md` |

---

## ⛔ 禁止项

- ❌ 中文注释
- ❌ 跳过 RCA / SCOPE / 验收标准
- ❌ 跳过 PRE-FLIGHT
- ❌ [SIMPLICITY CHECK] 输出前自问：能更简单吗？50 行能解决不用 200 行

---

## beads Status Management

> beads 状态管理命令见 `{TEAM_PATH}/workflows/shared.md` §beads Status Management
