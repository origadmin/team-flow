# 项目约定分层机制示例

> 展示如何在 `_docs/{project}/conventions/` 中按角色分层定义项目规范

---

## 目录结构

```
_docs/orig-cms/conventions/
├── README.md              # 机制说明
├── common.md              # 全局通用（所有角色）
├── dev-common.md          # Dev 通用（前后端都读）
├── dev-backend.md         # 后端专用
├── dev-frontend.md        # 前端专用
├── qa-common.md           # QA 专用（可选）
└── tech-lead-common.md    # Tech Lead 专用（可选）
```

---

## 分层加载机制

### Backend Dev 加载顺序

```
1. conventions/common.md              ← 全局
2. conventions/dev-common.md          ← Dev 通用
3. conventions/dev-backend.md         ← 后端专用
4. lessons/dev-common.md              ← 经验规则（_team 层）
```

### Frontend Dev 加载顺序

```
1. conventions/common.md              ← 全局
2. conventions/dev-common.md          ← Dev 通用
3. conventions/dev-frontend.md        ← 前端专用
4. lessons/dev-common.md              ← 经验规则（_team 层）
```

### 优先级

后加载的覆盖先加载的：

```
后端专用 > Dev 通用 > 全局 > _team 经验规则
```

---

## 内容分层示例

### common.md（全局）

```markdown
# Common Conventions

## 项目标识
- 名称: orig-cms
- 类型: 媒体内容管理

## 通用错误码
- 1000-1099: 系统级
- 1100-1199: 认证授权
```

**所有角色都读** → 知道项目基本信息

---

### dev-common.md（Dev 通用）

```markdown
# Dev Common Conventions

## Git 规范
- 分支: `feat/{id}-{desc}`, `fix/{id}-{desc}`
- Commit: `type: description`

## 测试规范
- 覆盖率 >= 80%
```

**前后端 Dev 都读** → 统一 Git、测试规范

---

### dev-backend.md（后端专用）

```markdown
# Backend Dev Conventions

## URL 前缀（强制）
| 类型 | 前缀 | 示例 |
|------|------|------|
| API | `/api/v1` | `/api/v1/articles` |
| Admin API | `/api/v1/admin` | `/api/v1/admin/users` |

## 数据库规范
- 表名: `snake_case` + 复数
- 字段: `created_at`, `updated_at`

## 模块边界（强制）
- Service 禁止直接访问 DB
- 依赖方向: Service → Biz → Data

## 错误码分配
- 内容模块: 2000-2099
- 媒体模块: 2100-2199
```

**只有后端 Dev 读** → 后端专用规范

---

### dev-frontend.md（前端专用）

```markdown
# Frontend Dev Conventions

## URL 前缀（强制）
| 类型 | 前缀 | 示例 |
|------|------|------|
| Admin 页面 | `/admin` | `/admin/content` |
| 登录页 | `/login` | `/login` |

## 目录结构
```
src/
├── api/           # API 调用
├── components/    # 组件
├── pages/         # 页面
└── stores/        # 状态管理
```

## 命名规范
- 组件: PascalCase
- Hooks: camelCase + use
- 文件: 同名 + .module.css

## API 基础路径
- `RSBUILD_API_BASE_URL = '/api/v1'`
```

**只有前端 Dev 读** → 前端专用规范

---

## 实际效果

### 场景 1: 后端创建 API

**用户**: "创建文章列表 API"

**AI 加载的约定**:
1. common.md → 知道项目叫 orig-cms
2. dev-common.md → 知道 Git 规范、测试规范
3. dev-backend.md → 知道：
   - API 必须 `/api/v1` 开头
   - 表名用下划线复数
   - Service 不能访问 DB
   - 内容模块错误码 2000-2099

**AI 输出**:
```go
// ✅ /api/v1/articles（符合前缀约定）
// ✅ articles 表（符合命名约定）
// ✅ Service → Biz → Data（符合模块边界）
```

---

### 场景 2: 前端创建页面

**用户**: "创建文章管理页面"

**AI 加载的约定**:
1. common.md → 知道项目叫 orig-cms
2. dev-common.md → 知道 Git 规范
3. dev-frontend.md → 知道：
   - Admin 页面必须 `/admin` 开头
   - 组件用 PascalCase
   - API 基础路径 `/api/v1`

**AI 输出**:
```tsx
// ✅ /admin/content/articles（符合前缀约定）
// ✅ ArticleListPage.tsx（符合命名约定）
// ✅ 调用 /api/v1/articles（符合 API 路径）
```

---

## 与 _team 经验规则的协作

| 层级 | 位置 | 内容 | 加载时机 |
|------|------|------|---------|
| **项目约定** | `_docs/*/conventions/` | 项目架构规范 | 角色触发时 |
| **经验规则** | `_team/lessons/` | 从错误中学习 | 角色触发时 |

**协作示例**:

```
Backend Dev 被触发
    ↓
加载项目约定
    ├── conventions/common.md
    ├── conventions/dev-common.md
    └── conventions/dev-backend.md
        └── "API 必须 /api/v1 开头"
    ↓
加载经验规则
    └── _team/lessons/dev-common.md
        └── "不要用 npm 代替 bun"
    ↓
同时遵循两类规则
```

---

## 快速开始

### 1. 创建约定文件

```bash
cd _docs/orig-cms/conventions/

# 编辑后端规范
touch dev-backend.md

# 编辑前端规范  
touch dev-frontend.md
```

### 2. 定义约定内容

参考上面的示例，定义：
- URL 前缀
- 命名规范
- 模块边界
- 错误码分配

### 3. 确保 prompt 加载

检查 `prompts/dev.md`:

```yaml
standards:
  # 项目约定
  - {DOCS_INTERNAL}/conventions/common.md
  - {DOCS_INTERNAL}/conventions/dev-common.md
  - {DOCS_INTERNAL}/conventions/dev-{subtype}.md
  # 经验规则
  - {TEAM_PATH}/lessons/dev-common.md
```

### 4. 完成

所有 AI 自动遵循项目约定

---

## 关键原则

| 原则 | 说明 |
|------|------|
| **存在即遵守** | 文件存在则必须遵循，不存在则跳过 |
| **分层覆盖** | 角色专用 > 角色通用 > 全局 |
| **项目优先** | 项目约定覆盖 _team 经验规则 |
| **预读 > 重试** | 执行前加载，避免出错后修正 |
