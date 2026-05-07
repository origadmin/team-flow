# 项目约定示例

> 展示如何在 project.md 中定义项目级规范，让所有 AI 统一遵循

---

## 场景 1: orig-cms 项目

### 项目约定（写入 project.md）

```markdown
## 项目约定 (Project Conventions)

### URL / 路由约定

| 类型 | 前缀 | 示例 |
|------|------|------|
| Admin 管理页面 | `/admin` | `/admin/content`, `/admin/media` |
| API 接口 | `/api/v1` | `/api/v1/articles`, `/api/v1/categories` |
| 上传接口 | `/api/v1/upload` | `/api/v1/upload/image`, `/api/v1/upload/video` |
| 健康检查 | `/health` | `/health`, `/health/ready` |

### API 设计约定

| 约定 | 说明 | 示例 |
|------|------|------|
| 列表分页 | 统一用 cursor 分页 | `?cursor=xxx&limit=20` |
| 时间格式 | ISO 8601 | `2024-01-15T10:30:00Z` |
| 响应包装 | 统一结构 | `{code, message, data}` |
| 错误码 | 按模块分段 | 内容模块: 2000-2999 |

### 数据库约定

| 约定 | 规范 |
|------|------|
| 表名 | 下划线命名，复数 | `articles`, `media_files` |
| 时间字段 | created_at, updated_at, deleted_at |
| 软删除 | 统一用 deleted_at |
| 外键 | 表名_单数_id | `article_id`, `category_id` |

### 代码组织约定

| 层级 | 职责 | 禁止 |
|------|------|------|
| `service` | HTTP/gRPC 接口 | 禁止直接操作 SQL |
| `biz` | 业务逻辑 | 禁止依赖具体存储 |
| `data` | 数据访问 | 禁止调用外部服务 |
```

### AI 执行效果

**用户**: "创建文章管理页面"

**AI 行为**:
1. 加载 project.md → 读取 Project Conventions
2. 发现 Admin 页面必须 `/admin` 开头
3. 创建路由: `/admin/articles`
4. 发现 API 必须 `/api/v1` 开头
5. 创建 API: `/api/v1/articles`
6. 统一响应格式: `{code, message, data}`

**结果**: 所有 AI 创建的页面和 API 自动遵循统一规范

---

## 场景 2: 多项目对比

### 项目 A (orig-cms)

```yaml
Project Conventions:
  URL:
    admin_prefix: "/admin"
    api_prefix: "/api/v1"
  Database:
    naming: "snake_case_plural"  # articles, user_profiles
```

### 项目 B (orig-admin-dashboard)

```yaml
Project Conventions:
  URL:
    admin_prefix: "/dashboard"   # 不同！
    api_prefix: "/api/v2"        # 不同！
  Database:
    naming: "snake_case_plural"  # 相同
```

### AI 行为差异

| 任务 | 项目 A | 项目 B |
|------|--------|--------|
| 创建管理页面 | `/admin/users` | `/dashboard/users` |
| 创建 API | `/api/v1/users` | `/api/v2/users` |

**关键**: AI 根据当前项目的 project.md 自动调整，不会混淆

---

## 快速定义模板

```markdown
## 项目约定 (Project Conventions)

### URL / 路由约定

| 类型 | 前缀 | 示例 |
|------|------|------|
| Admin 页面 | `{prefix}` | `{example}` |
| API 接口 | `{prefix}` | `{example}` |
| 健康检查 | `{prefix}` | `{example}` |

### 命名约定

| 类型 | 规范 | 示例 |
|------|------|------|
| 数据库表 | `{rule}` | `{example}` |
| 结构体 | `{rule}` | `{example}` |
| API 端点 | `{rule}` | `{example}` |

### 模块边界

| 模块 | 职责 | 禁止 |
|------|------|------|
| `{layer}` | `{responsibility}` | `{forbidden}` |
```

---

## 与经验规则的区别

| 对比项 | 项目约定 (Project Conventions) | 经验规则 (Lessons) |
|--------|-------------------------------|-------------------|
| **来源** | 项目本身的架构设计 | 从错误中学习 |
| **目的** | 统一项目风格 | 防止重复犯错 |
| **位置** | `project.md` | `lessons/*.md` |
| **更新时机** | 项目初始化、架构变更 | 每次错误发生后 |
| **示例** | API 前缀、命名规范 | 不要用 npm 代替 bun |

---

## 完整配置检查清单

项目启动时，AI 应该检查：

```
✅ project.md 存在
✅ Project Conventions 章节存在
  ✅ URL 前缀定义
  ✅ 命名规范定义
  ✅ 模块边界定义
✅ Toolchain 配置存在
  ✅ package_manager 定义
  ✅ 常用命令定义
✅ lessons/dev-common.md 存在（可选）
```

---

## 下一步

1. 在 `.team/project.md` 中添加「项目约定」章节
2. 定义你的 URL 前缀、命名规范、模块边界
3. 确保各角色 prompt 的入口门禁读取 project.md
4. 所有 AI 自动遵循统一规范
