# Feature UI 设计模板

> **用途**: UI Designer 为具体 Feature 编写页面设计文档
>
> **前置**: 先读 `{TEAM_PATH}/workflows/roles/ui-standards.md`（Design System）
>
> **输出路径**: `{docs_internal}/requirements/{feature}/ui-design/DESIGN.md`
>
> **配套文件**:
> - `SNAPSHOT.html` — HTML 快照（用于 AI + QA 结构级验证）
> - `COMPONENT-PROPOSAL.md` — 新组件提案（如有）

---

## 一、文档元信息

| 字段 | 内容 |
|------|------|
| 功能名称 | {feature-name} |
| 页面名称 | {page-name} |
| 版本 | v{version} |
| 更新时间 | {YYYY-MM-DD} |
| 状态 | [起草/评审中/已确认] |
| 设计依据 | `{docs_internal}/requirements/{feature}/SPEC.md` |
| 参考标准 | `{TEAM_PATH}/workflows/roles/ui-standards.md` |
| 设计者 | UI Designer |
| 验证者 | PM / Tech Lead |

---

## 二、功能上下文

> **必须引用**: PM 的 `SPEC.md` 中相关功能描述，确保 UI 设计服务于功能需求。

```markdown
功能模块: {模块名称}
功能描述: {简述功能目标}
目标用户: {用户角色}
使用场景: {使用场景}
```

---

## 三、页面结构

### 3.1 布局概览

描述页面的整体结构，引用 ui-standards.md 中的布局规范：

```
┌──────────────────────────────────────────────────────┐
│  Header: {高度 60px，内容描述}                        │
├────────────┬─────────────────────────────────────────┤
│            │                                         │
│  Sidebar:  │  Main Content:                         │
│  宽度 200px │  最大宽度 1200px，居中                  │
│  内容描述   │                                         │
│            │                                         │
│            │                                         │
└────────────┴─────────────────────────────────────────┘
```

### 3.2 页面类型

- [ ] 列表页
- [ ] 详情页
- [ ] 表单页
- [ ] 弹窗
- [ ] 抽屉（Drawer）
- [ ] 混合页（列表 + 详情同页）

---

## 四、组件使用清单

> **强制引用**: 所有组件必须使用 ui-standards.md 中定义的组件 ID，禁止自创样式。

### 4.1 页面头部

| 区域/组件 | 组件 ID | 说明 |
|-----------|---------|------|
| 标题文字 | `--font-size-xl`, `--font-weight-bold` | 引用 Token |
| 右侧操作区 | `btn-primary` + `btn-secondary` | — |

### 4.2 搜索/筛选区

| 组件 | 组件 ID | 说明 |
|------|---------|------|
| 搜索输入框 | `input-search` | — |
| 筛选 Select | `select-default` | — |
| 重置按钮 | `btn-text` | — |
| 搜索按钮 | `btn-primary` | — |

### 4.3 数据展示区

| 组件 | 组件 ID | 说明 |
|------|---------|------|
| 表格 | `table-default` | — |
| 卡片列表 | `card-default` | 可选 |
| 空状态 | `table-empty` + `message-info` | — |

### 4.4 操作区

| 操作 | 组件 ID | 说明 |
|------|---------|------|
| 新建/新增 | `btn-primary` | 右上角 |
| 编辑 | `btn-text` | 行内 |
| 删除 | `btn-text` + `modal-confirm` | 行内 + 二次确认 |

### 4.5 分页

| 组件 | 规格 |
|------|------|
| 分页组件 | 底部居右，每页条数: 10/20/50/100，显示总数 |

---

## 五、页面内所有状态

### 5.1 状态清单

| 状态 | 触发条件 | 视觉表现 |
|------|---------|---------|
| 默认 | 页面正常加载完成 | 正常显示所有内容 |
| 加载中 | 数据请求中 | `loading-mask` 或 `loading-skeleton` |
| 空状态 | 无数据 | `table-empty` 组件，居中提示 |
| 部分数据 | 数据 < 1 页 | 正常显示，不显示分页 |
| 操作成功 | CRUD 操作完成 | `message-success` 提示，3s 消失 |
| 操作失败 | CRUD 操作失败 | `message-error` 提示，不消失直到用户关闭 |

### 5.2 错误状态

| 错误场景 | UI 表现 |
|---------|---------|
| 网络错误 | `message-error`: "网络异常，请稍后重试"，带重试按钮 |
| 权限不足 | `message-warning`: "您无权限访问该内容"，或跳转无权限页 |
| 404 | 全页空状态 + "页面不存在" |
| 表单验证失败 | 输入框 `input-error`，红色错误提示在输入框下方 |

---

## 六、交互行为

### 6.1 页面加载

```
用户进入页面
  → 显示 Header + Sidebar（立即）
  → Main Content 显示 `loading-skeleton`（异步）
  → 数据返回后替换为真实内容（`--transition-fast` 淡入）
```

### 6.2 表格交互

| 交互 | 行为 |
|------|------|
| 行悬停 | 背景 `--color-bg-hover` |
| 行点击 | 高亮行（左边框 3px `--color-primary`），跳转详情或展开 |
| 列排序 | 点击表头排序图标，升序/降序切换 |
| 行选择 | Checkbox 选择，支持全选 |

### 6.3 表单提交

```
用户填写表单
  → 点击 "提交" (`btn-primary`，Loading 态）
  → 按钮显示 `loading-spinner`，禁用点击
  → 请求完成：
      成功 → `message-success` + 跳转/关闭弹窗
      失败 → `message-error` + 按钮恢复
```

### 6.4 删除确认

```
用户点击 "删除"
  → 弹出 `modal-confirm`
    - 标题: "确认删除"
    - 内容: "删除后数据无法恢复，确定要删除吗？"
    - 按钮: "取消" (`btn-secondary`) + "确认删除" (`btn-danger`)
  → 确认后执行删除，后端返回后刷新列表
```

---

## 七、响应式规则

### 7.1 Desktop (≥ 1200px)

完整布局：Sidebar 展开（200px），内容区 3-4 列栅格。

### 7.2 Tablet (768px - 1199px)

Sidebar 折叠为 Icon-only（64px），内容区 2 列栅格，表格水平滚动。

### 7.3 Mobile (< 768px)

Sidebar 隐藏（汉堡菜单触发），内容区单列，表格横向滚动，分页改为 "上一页/下一页"。

---

## 八、HTML 快照

> **用途**: 用于 AI 验证和 QA 结构级检查。提供页面的 HTML 结构摘要，标注关键组件 ID。

```html
<!-- {页面名称} HTML 结构快照 -->
<div class="page-container">
  <!-- Header -->
  <header class="page-header" style="height: 60px;">
    <h1 class="page-title">{页面标题}</h1>
    <div class="header-actions">
      <button class="btn-primary">新建</button>
    </div>
  </header>

  <!-- Main Content -->
  <main class="page-content" style="max-width: 1200px;">
    <!-- 搜索筛选区 -->
    <div class="search-bar">
      <input type="search" class="input-search" placeholder="搜索..." />
      <button class="btn-primary">搜索</button>
    </div>

    <!-- 数据表格 -->
    <table class="table-default table-clickable-row">
      <thead>...</thead>
      <tbody>...</tbody>
    </table>

    <!-- 分页 -->
    <div class="pagination">...</div>
  </main>
</div>
```

> **要求**: HTML 快照不需要完整样式（可省略 CSS），但必须标注组件 ID（class）与 ui-standards.md 中的组件 ID 一一对应。

---

## 九、自定义样式申请

> **如 Feature 需要 ui-standards.md 中不存在的组件或样式，在此申请。**

| 申请编号 | 描述 | 规格 | 申请原因 | 状态 |
|---------|------|------|---------|------|
| — | — | — | — | — |

---

## 十、验证清单

| 检查项 | 要求 | 状态 |
|--------|------|------|
| 功能上下文已引用 PM SPEC.md | — | [ ] |
| 所有组件 ID 来自 ui-standards.md | 无自创组件 | [ ] |
| 颜色使用 Token（无硬编码色值） | — | [ ] |
| 间距使用 `$spacing-{n}` Token | — | [ ] |
| 高度符合规范（40px 输入框 / 48px 表格行） | — | [ ] |
| 所有状态有明确说明（加载/空/错误/成功） | — | [ ] |
| 响应式三端说明完整 | — | [ ] |
| HTML 快照标注组件 ID | — | [ ] |
| PM / Tech Lead 已评审 | 签字确认 | [ ] |
