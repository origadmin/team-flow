---
applyTo: "**/*.feature"
version: "1.0"
owner: "[PM / QA Engineer]"
lastUpdated: "2026-04-22"
---

# Gherkin 场景模板

> **用途**: 编写 BDD（行为驱动开发）验收标准
>
> **使用者**: PM（编写 Feature 文件）→ Dev（理解 AC，编写 Step Definitions）→ QA（执行 E2E 场景）
>
> **位置约定**: Feature 文件统一放在 `projects/{project}/features/` 目录，按模块命名子目录
>
> **前置规范**: 先读 `{TEAM_PATH}/workflows/roles/development-standards.md` 了解实现层测试设计，本模板聚焦验收层（E2E）AC 编写
>
> **与 development-standards.md 的分工**:
> - `development-standards.md` → **实现层**：状态机、Data Model、API Contract、测试点拆解（Dev 视角）
> - 本模板 → **验收层**：Gherkin Feature / Scenario（PM + QA 视角，BDD AC）

---

## 一、文件命名规范

```
# 格式
{module}-{sub-feature}.feature

# 示例
user-management.feature          # 用户管理（按模块）
user-authentication.feature       # 用户认证
media-upload.feature              # 媒体上传
comment-crud.feature              # 评论增删改查
```

Feature 文件按模块放目录：`features/user/`, `features/media/`, `features/admin/` 等。

---

## 二、Feature 文件基本结构

```gherkin
@tag1 @tag2
Feature: {功能名称}
  "作为 [角色]"
  "我想要 [功能]"
  "以便 [业务价值]"

  # 业务规则层（可选，对场景分组）
  Rule: {规则名称 1}
    # 该规则下所有 Scenario 共享此上下文

    Scenario: {场景 1}
      Given ...
      When ...
      Then ...

    Scenario: {场景 2}
      Given ...
      When ...
      Then ...

  Rule: {规则名称 2}

  # 公共前置（每个 Scenario 前自动执行）
  Background:
    Given 系统已启动
    And 我已登录为管理员

  # 独立场景
  Scenario: {场景名称}
    Given {前置状态}
    When {用户行为}
    Then {预期结果}

  # 数据驱动场景（参数化）
  Scenario Outline: {场景名称}
    Given <输入值>
    When 我执行操作
    Then 结果应该是 <预期>
    Examples:
      | 输入值 | 预期 |
      | 有效值 | 成功 |
      | 空值   | 失败 |
```

---

## 三、Rule 场景分组（业务规则层）

> **Rule** 是 Gherkin R4+ 支持的关键字，用于对相关场景分组，表达业务规则。
> 适用于：同一 Feature 下有多个业务规则，每个规则包含多个场景。

```gherkin
Feature: 用户权限管理

  Rule: 管理员拥有全部权限
    # 所有管理员相关场景共享此上下文
    Scenario: 管理员可以创建用户
      Given 我已登录为管理员
      When 我访问用户管理页面
      Then 我应该看到"新建用户"按钮

    Scenario: 管理员可以删除任意用户
      Given 用户 "张三" 已存在
      And 我已登录为管理员
      When 我删除用户 "张三"
      Then 删除应该成功

  Rule: 普通用户权限受限
    Scenario: 普通用户无法访问管理页面
      Given 我已登录为普通用户
      When 我访问用户管理页面
      Then 我应该看到"无权访问"提示

    Scenario: 普通用户无法删除其他用户
      Given 用户 "张三" 已存在
      And 我已登录为普通用户
      When 我删除用户 "张三"
      Then 删除应该失败
```

---

## 四、Scenario Outline 数据驱动

> **Scenario Outline + Examples** 用于参数化测试：同一场景用多组数据验证。
> 适用于：表单验证、边界值测试、多种输入组合。

```gherkin
  Scenario Outline: 邮箱格式验证
    Given 我在注册页面
    When 我填写邮箱 "<邮箱>"
    And 我提交注册
    Then 应该显示 "<提示信息>"

    @invalid
    Examples:
      | 邮箱          | 提示信息        |
      | 无@符号       | 邮箱格式不正确  |
      | 无域名        | 邮箱格式不正确  |
      | 包含空格      | 邮箱格式不正确  |
      | 全角字符      | 邮箱格式不正确  |

    @valid
    Examples:
      | 邮箱                    | 提示信息  |
      | user@example.com        | 注册成功  |
      | test.user@domain.com    | 注册成功  |
```

---

## 五、DocString 多行文本

> **DocString（"""）** 用于传递多行文本、JSON、HTML 等复杂内容。

```gherkin
  Scenario: 提交包含多行内容的评论
    Given 我已登录为普通用户
    When 我提交评论:
      """
      第一行评论内容
      第二行评论内容

      段落间隔后的内容
      """
    Then 评论应该被保存
    And 评论内容应包含 "第一行评论内容"
```

---

## 六、步骤词汇表

### 6.1 Given（前置状态）

```gherkin
Given 系统已启动
Given 我已登录为 <角色>
Given 我在 <页面> 页面
Given 用户 <用户名> 已存在
Given 系统中已存在数据:
  | 字段 | 值 |
Given 我已进入 <模块> 页面
Given 服务器返回正常响应
```

### 6.2 When（用户行为 / 触发）

```gherkin
When 我点击 "<按钮>" 按钮
When 我点击 "<链接>" 链接
When 我填写表单:
  | 字段 | 值 |
When 我选择 "<选项>" 选项
When 我提交 "<操作>"
When 我上传文件 "<文件名>"
When 我执行 "<操作>" 操作
When 我等待 <秒> 秒
When API 被调用 "<endpoint>"
When 定时任务被触发
```

### 6.3 Then（预期结果）

```gherkin
Then 我应该看到 "<文本>" 提示
Then 我应该看到 "<文本>" 错误提示
Then 我应该被重定向到 <页面>
Then <列表> 应该包含 "<内容>"
Then <列表> 不应该包含 "<内容>"
Then <字段> 应该等于 <值>
Then API 返回状态码 <200>
Then 请求应该成功
Then 响应应该包含:
  | 字段 | 值 |
```

### 6.4 And / But（连续步骤）

```gherkin
And 我填写密码 "xxx"
And 我选择"记住我"
And 我点击"登录"
But 我不应该被允许访问
But 用户不应该被创建
```

### 6.5 Error 错误处理

```gherkin
Then 我应该看到 "<错误信息>" 错误提示
Then 错误提示应该显示 "<详情>"
Then 系统应该记录错误日志
Then 请求应该返回 400
Then 字段 "<字段>" 应该报错 "<错误>"
```

### 6.6 状态变更

```gherkin
Given 订单状态为 "<待支付>"
When 我完成支付
Then 订单状态应该变为 "<已支付>"
And 支付时间应该被记录
And 库存应该被扣减
```

### 6.7 异步 / 等待

```gherkin
When 我触发异步任务
Then 等待任务完成
And 任务状态应该为 "完成"
And 通知应该被发送
```

### 6.8 翻页 / 列表

```gherkin
When 我翻到第 <N> 页
Then 当前页应该显示第 <N> 页
And 每页应该显示 <数量> 条记录
And 总页数应该为 <N>
```

---

## 七、常用场景分类模板

### 7.1 CRUD 场景

```gherkin
Feature: {模块} CRUD

  Scenario: 创建{实体}
    Given 我已登录为管理员
    When 我创建 {实体}:
      | 字段 | 值 |
    Then {实体} 应该被创建
    And {实体} 应该出现在列表中

  Scenario: 查询{实体}
    Given {实体} 已存在
    When 我查询 {实体}
    Then 我应该看到 {实体} 详情

  Scenario: 更新{实体}
    Given {实体} 已存在
    When 我更新 {实体} 信息:
      | 字段 | 新值 |
    Then {实体} 应该被更新

  Scenario: 删除{实体}
    Given {实体} 已存在
    When 我删除 {实体}
    Then {实体} 应该被删除
    And {实体} 不应该出现在列表中
```

### 7.2 权限场景

```gherkin
Feature: {模块} 权限控制

  Scenario: {角色} 可以访问 {资源}
    Given 我已登录为 <角色>
    When 我访问 <资源>
    Then 访问应该成功

  Scenario: {角色} 无权访问 {资源}
    Given 我已登录为 <角色>
    When 我访问 <资源>
    Then 我应该看到"无权访问"提示
    And 访问应该被拒绝
```

### 7.3 搜索场景

```gherkin
Feature: {模块} 搜索

  Scenario: 搜索存在结果
    Given 存在符合条件的数据
    When 我搜索 "<关键词>"
    Then 搜索结果应该包含数据

  Scenario: 搜索无结果
    When 我搜索 "<不存在的关键词>"
    Then 我应该看到"无搜索结果"提示

  Scenario: 搜索结果分页
    Given 存在超过 10 条符合条件的数据
    When 我搜索 "<关键词>"
    Then 第一页应该显示 10 条
    And 总数应该显示 "<总条数>"
```

---

## 八、标签体系

### 8.1 标签说明

| 标签 | 用途 | 执行频率 |
|------|------|---------|
| `@smoke` | 冒烟测试，核心流程 | 每次 CI 必须执行 |
| `@critical` | 关键路径，P0 | 每次部署必须执行 |
| `@regression` | 回归测试，P1 | 每次发版执行 |
| `@e2e` | 端到端测试 | 定期执行 |
| `@slow` | 执行时间 > 30s | 可选执行 |
| `@wip` | 开发中，暂不执行 | CI 跳过 |
| `@api` | API 层测试 | 单元测试级别 |
| `@unit` | 单元测试级别 | 快速执行 |

### 8.2 Feature 级标签

```gherkin
@smoke @user-management
Feature: 用户管理
  # 该 Feature 下所有 Scenario 继承 @smoke 和 @user-management
```

### 8.3 Scenario 级标签覆盖

```gherkin
@smoke
Feature: 用户认证

  @critical @smoke
  Scenario: 登录成功
    # 覆盖 Feature 级标签，critical > smoke

  @slow
  Scenario: 记住登录状态
    # 独立标签，不继承 @smoke
```

---

## 九、反模式（避免的错误）

```gherkin
# ❌ 错误：步骤包含实现细节
Scenario: 错误示范
  When 我调用 POST /api/users with JSON body
  Then 数据库中应该有记录

# ✅ 正确：步骤表达行为，不暴露实现
Scenario: 正确示范
  When 我创建用户 "张三"
  Then 用户 "张三" 应该被创建
```

```gherkin
# ❌ 错误：多个断言写在同一个 Then
Scenario: 错误示范
  Then 系统应该返回 200
  And 数据库应该更新
  And 邮件应该被发送
  And 通知应该显示

# ✅ 正确：一个场景，一个核心验证点；多验证点拆为多个 Scenario
Scenario: 正确示范 - 创建用户成功
  When 我创建用户 "张三"
  Then 用户应该被创建

Scenario: 正确示范 - 创建用户触发通知
  When 我创建用户 "张三"
  Then 管理员应该收到通知
```

```gherkin
# ❌ 错误：Given 步骤中做操作（Given 不是 When）
Scenario: 错误示范
  Given 我创建了用户 "张三"
  When 我登录为 "张三"
  Then 登录应该成功

# ✅ 正确：Given 表达状态，When 表达行为
Scenario: 正确示范
  Given 用户 "张三" 已存在
  When 我登录为 "张三"
  Then 登录应该成功
```

---

## 十、Gherkin → 测试分层映射

| 层级 | 测试对象 | 工具 | 对应本模板 |
|------|---------|------|-----------|
| **E2E** | 完整用户流程 | Cucumber / Playwright BDD | **Gherkin Feature 文件**（本模板） |
| **Integration** | 模块间交互 | Go test, API mock | development-standards.md |
| **Unit** | 单个函数/方法 | Go test | development-standards.md |

**工作流中的位置**:

```
PM  编写 Feature (.feature) + Gherkin AC   →  附加到 TaskPool
Dev 参考 Gherkin AC 理解需求                  →  编写 Step Definitions
QA  执行 Gherkin Feature 文件                  →  输出测试报告
```

> Dev 负责实现 Step Definitions（将 Gherkin 步骤绑定到实际代码），参考 `development-standards.md` 中的测试点拆解模板设计测试用例。

---

## 十一、完整示例

```gherkin
@user-management @smoke @e2e
Feature: 用户管理
  作为系统管理员
  我想要管理用户账号
  以便控制系统访问权限

  Rule: 用户名唯一性约束
    Scenario: 用户名重复时拒绝创建
      Given 用户 "张三" 已存在
      When 我创建用户:
        | 用户名 | 张三 |
        | 邮箱   | new@example.com |
      Then 我应该看到"用户名已存在"错误提示
      And 用户不应该被创建

    Scenario: 不同用户可以使用相同邮箱
      Given 用户 "张三" 已存在
      When 我创建用户:
        | 用户名 | 李四 |
        | 邮箱   | zhangsan@example.com |
      Then 用户 "李四" 应该被创建

  Background:
    Given 我已登录为管理员
    And 我在用户管理页面

  @critical
  Scenario: 成功创建新用户
    When 我点击"新建用户"按钮
    And 我填写用户信息:
      | 用户名 | 王五 |
      | 邮箱   | wangwu@example.com |
      | 角色   | 普通用户 |
    And 我点击"保存"按钮
    Then 我应该看到"用户创建成功"提示
    And 用户列表应该包含 "王五"

  Scenario Outline: 邮箱格式验证
    Given 我在注册页面
    When 我填写邮箱 "<邮箱>"
    And 我点击"注册"按钮
    Then 我应该看到 "<提示>"

    @invalid
    Examples:
      | 邮箱       | 提示           |
      | no-at-sign | 邮箱格式不正确 |
      | no-domain  | 邮箱格式不正确 |

    @valid
    Examples:
      | 邮箱              | 提示      |
      | test@example.com  | 注册成功  |
```

---

> **版本**: v1.0 | **更新**: 2026-04-22 | **维护者**: Team Framework
