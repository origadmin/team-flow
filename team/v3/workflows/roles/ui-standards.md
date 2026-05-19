# UI 设计规范

> **角色**: UI Designer（维护）/ Frontend Dev（引用）/ QA（验证）
> **加载文件**: `{TEAM_PATH}/workflows/roles/ui-standards.md`
> **前置**: 先读 `{TEAM_PATH}/workflows/shared.md` 了解通用原则
>
> **重要**: 本文档是团队 Design System，所有 UI 设计必须引用本文档。禁止在 Feature Design 中硬编码非本文档定义的样式值。

**版本**: v1.0
**更新日期**: 2026-04-22
**维护者**: UI Designer / Tech Lead

---

## 工作流角色

```
PM 需求 → UI Designer 产出 UI Design → Dev 实现 → QA 截图验证

    ┌─────────────────────────────────────────────┐
    │           ui-standards.md (Design System)    │
    │  维护者: UI Designer / Tech Lead             │
    │  内容: 组件规范、Token 系统、验证规则          │
    └─────────────────────────────────────────────┘
                        ↑ 引用
    Feature UI Design Documents (每功能一份)
    {docs_internal}/requirements/{feature}/ui-design/DESIGN.md
    维护者: UI Designer
    内容: 引用标准 + 页面专属布局/组件/交互
```

---

## 1. 设计 Token 系统

### 1.1 颜色 Token

| Token | 值 | 用途 |
|-------|-----|------|
| `--color-primary` | `#1890FF` | 主要操作、链接、聚焦 |
| `--color-primary-hover` | `#40A9FF` | 主色悬停 |
| `--color-primary-active` | `#096DD9` | 主色按下 |
| `--color-success` | `#52C41A` | 成功状态 |
| `--color-warning` | `#FAAD14` | 警告状态 |
| `--color-error` | `#FF4D4F` | 错误状态、危险操作 |
| `--color-text-primary` | `#262626` | 主文字 |
| `--color-text-secondary` | `#8C8C8C` | 次要文字 |
| `--color-text-disabled` | `#BFBFBF` | 禁用文字 |
| `--color-border` | `#D9D9D9` | 默认边框 |
| `--color-border-focus` | `#1890FF` | 聚焦边框 |
| `--color-bg-base` | `#FFFFFF` | 页面背景 |
| `--color-bg-elevated` | `#FAFAFA` | 卡片/面板背景 |
| `--color-bg-hover` | `#F5F5F5` | 行/项悬停背景 |

### 1.2 字体 Token

| Token | 值 | 用途 |
|-------|-----|------|
| `--font-family` | `"PingFang SC", "Microsoft YaHei", sans-serif` | 字体族 |
| `--font-size-xs` | `12px` | 辅助文字 |
| `--font-size-sm` | `13px` | 次要文字 |
| `--font-size-base` | `14px` | 正文文字 |
| `--font-size-lg` | `16px` | 标题/强调 |
| `--font-size-xl` | `18px` | 页面标题 |
| `--font-size-xxl` | `24px` | 大标题 |
| `--font-weight-normal` | `400` | 正常 |
| `--font-weight-medium` | `500` | medium |
| `--font-weight-bold` | `600` | 粗体 |
| `--line-height-base` | `1.5` | 正文行高 |
| `--line-height-tight` | `1.25` | 紧凑行高 |

### 1.3 间距 Token（8px 基准网格）

| Token | 值 | 用途 |
|-------|-----|------|
| `--spacing-1` | `4px` | 紧凑间距 |
| `--spacing-2` | `8px` | 小间距 |
| `--spacing-3` | `12px` | 标准间距 |
| `--spacing-4` | `16px` | 中等间距 |
| `--spacing-5` | `20px` | 中大间距 |
| `--spacing-6` | `24px` | 大间距 |
| `--spacing-8` | `32px` | 区块间距 |
| `--spacing-10` | `40px` | 大区块间距 |
| `--spacing-12` | `48px` | 页面级间距 |

### 1.4 圆角 Token

| Token | 值 | 用途 |
|-------|-----|------|
| `--radius-sm` | `2px` | 小圆角（Tag） |
| `--radius-base` | `4px` | 基础圆角（按钮/输入框） |
| `--radius-lg` | `8px` | 大圆角（卡片/弹窗） |
| `--radius-full` | `9999px` | 全圆角（圆形按钮/Pill） |

### 1.5 阴影 Token

| Token | 值 | 用途 |
|-------|-----|------|
| `--shadow-sm` | `0 1px 2px rgba(0,0,0,0.06)` | 小阴影 |
| `--shadow-base` | `0 2px 8px rgba(0,0,0,0.10)` | 基础阴影 |
| `--shadow-lg` | `0 4px 16px rgba(0,0,0,0.12)` | 大阴影（弹窗） |

### 1.6 动效 Token

| Token | 值 | 用途 |
|-------|-----|------|
| `--transition-fast` | `150ms ease-in-out` | 快速过渡（悬停/聚焦） |
| `--transition-base` | `250ms ease-in-out` | 标准过渡 |
| `--transition-slow` | `400ms ease-in-out` | 慢过渡（弹窗/展开） |

### 1.7 布局 Token

| Token | 值 | 用途 |
|-------|-----|------|
| `--layout-header-height` | `60px` | 顶部导航高度 |
| `--layout-sidebar-width` | `200px` | 侧边栏宽度 |
| `--layout-sidebar-collapsed` | `64px` | 折叠侧边栏宽度 |
| `--layout-content-max-width` | `1200px` | 内容区最大宽度 |

---

## 2. 组件规范

### 2.1 按钮 Button

| 组件 ID | 类型 | 样式 | 规格 | 使用场景 |
|---------|------|------|------|---------|
| `btn-primary` | 主要操作 | 背景 `--color-primary`，文字白色 | 高度 40px，圆角 `--radius-base` | 页面主操作，每个区域最多 1 个 |
| `btn-secondary` | 次要操作 | 边框 `--color-primary`，背景透明 | 高度 40px，圆角 `--radius-base` | 主操作旁边的辅助操作 |
| `btn-text` | 文字操作 | 无边框/背景，文字 `--color-primary` | 高度 32px | 表格内操作、区块内次要操作 |
| `btn-danger` | 危险操作 | 背景 `--color-error`，文字白色 | 高度 40px，圆角 `--radius-base` | 删除、取消、退出等不可逆操作 |
| `btn-disabled` | 禁用状态 | 所有按钮的禁用态：背景 `#F5F5F5`，文字 `#BFBFBF`，无 hover | — | 禁用状态 |

**交互状态**:

| 状态 | 视觉变化 |
|------|---------|
| 默认 | btn-primary: 背景 `#1890FF` |
| 悬停 Hover | 背景变亮 `#40A9FF`，阴影 `--shadow-sm` |
| 按下 Active | 背景变深 `#096DD9`，无阴影 |
| 禁用 Disabled | 背景 `#F5F5F5`，文字 `#BFBFBF`，cursor: not-allowed |
| 加载 Loading | 文字替换为 Loading Spinner，保持宽度 |

### 2.2 输入框 Input

| 组件 ID | 样式 | 规格 |
|---------|------|------|
| `input-default` | 高度 40px，边框 `--color-border` | 标准输入 |
| `input-focus` | 边框 `--color-border-focus`，外阴影 `0 0 0 2px rgba(24,144,255,0.20)` | 聚焦态 |
| `input-error` | 边框 `--color-error`，下方红色错误提示文字 `--color-error` | 错误态 |
| `input-disabled` | 背景 `#F5F5F5`，文字 `--color-text-disabled` | 禁用态 |
| `input-search` | 左侧 Search Icon，内嵌 `input-default` | 搜索框 |
| `input-textarea` | 最小高度 80px，resize 垂直，字体同 input | 多行输入 |

**高度规范**: 所有输入类组件（Input、Select、DatePicker 等）统一高度 **40px**。

### 2.3 选择器 Select

| 组件 ID | 样式 | 规格 |
|---------|------|------|
| `select-default` | 同 `input-default`，右侧 Chevron Icon | 标准选择 |
| `select-multi` | 同 `input-default`，选中项以 Tag 展示 | 多选 |

### 2.4 表格 Table

| 组件 ID | 规格 |
|---------|------|
| `table-default` | 行高 48px，表头背景 `--color-bg-elevated`，表头文字加粗，斑马纹 `--color-bg-hover` |
| `table-clickable-row` | 行可点击，hover 背景 `--color-bg-hover`，cursor: pointer |
| `table-empty` | 居中图标 + "暂无数据" 文字，文字色 `--color-text-secondary` |

### 2.5 卡片 Card

| 组件 ID | 样式 | 规格 |
|---------|------|------|
| `card-default` | 背景白色，圆角 `--radius-lg`，阴影 `--shadow-base`，内边距 24px | 标准卡片 |
| `card-bordered` | 边框 `--color-border`，无阴影，圆角 `--radius-base` | 低强调卡片 |
| `card-header` | 卡片顶部标题栏，底部分割线，标题区高度 48px | 带标题卡片 |

### 2.6 弹窗 Modal

| 组件 ID | 规格 |
|---------|------|
| `modal-default` | 居中，背景白色，圆角 `--radius-lg`，阴影 `--shadow-lg`，最小宽度 480px，遮罩 rgba(0,0,0,0.45) |
| `modal-footer` | 底部按钮区，右对齐，间距 8px，包含 `btn-primary` + `btn-secondary` 或 `btn-text` |
| `modal-confirm` | 宽度 400px，内容居中，用于二次确认 |
| `modal-drawer` | 从右侧滑入，宽度 480px，用于详情/表单场景 |

**动画**: 弹窗入场/退场使用 `--transition-slow` (400ms ease-in-out)。

### 2.7 消息提示 Message

| 组件 ID | 样式 |
|---------|------|
| `message-success` | 左侧成功图标，文字 `--color-success`，背景浅绿底 `#F6FFED` |
| `message-error` | 左侧错误图标，文字 `--color-error`，背景浅红底 `#FFF2F0` |
| `message-warning` | 左侧警告图标，文字 `--color-warning`，背景浅黄底 `#FFFBE6` |
| `message-info` | 左侧信息图标，文字 `--color-primary`，背景浅蓝底 `#E6F7FF` |

**位置**: 页面顶部居中，距顶部 24px，自动 3s 后消失。

### 2.8 导航 Nav

| 组件 ID | 规格 |
|---------|------|
| `nav-header` | 高度 `--layout-header-height`，背景白色，底边框 1px `--color-border` |
| `nav-sidebar` | 宽度 `--layout-sidebar-width`，背景 `#001529`（深色），文字白色 |
| `nav-item` | 高度 48px，内边距 0 16px，hover 背景 `rgba(255,255,255,0.08)` |
| `nav-item-active` | 左侧边框 3px `--color-primary`，背景同 hover |
| `nav-tab` | Tab 横向排列，高度 44px，下边框聚焦态 |

### 2.9 标签 Tag

| 组件 ID | 样式 |
|---------|------|
| `tag-default` | 背景 `--color-bg-elevated`，文字 `--color-text-secondary`，圆角 `--radius-sm`，padding 2px 8px |
| `tag-status-success` | 背景 `#E6F7FF`，文字 `#1890FF` |
| `tag-status-error` | 背景 `#FFF2F0`，文字 `#FF4D4F` |
| `tag-status-warning` | 背景 `#FFFBE6`，文字 `#FAAD14` |

### 2.10 加载 Loading

| 组件 ID | 规格 |
|---------|------|
| `loading-spinner` | 16px 旋转圆环，颜色 `--color-primary` |
| `loading-mask` | 全页遮罩 rgba(255,255,255,0.8)，居中 `loading-spinner` |
| `loading-skeleton` | 背景渐变动画 `#F5F5F5 → #EBEBEB → #F5F5F5`，时长 1.5s |

---

## 3. 布局规范

### 3.1 页面布局结构

```
┌─────────────────────────────────────────────────────┐
│  Header (60px)                                       │
├────────────┬────────────────────────────────────────┤
│            │                                         │
│  Sidebar   │         Main Content                   │
│  (200px)   │         (max-width: 1200px, 居中)      │
│            │                                         │
│            │                                         │
│            │                                         │
│            │                                         │
└────────────┴────────────────────────────────────────┘
```

### 3.2 栅格系统

| 描述 | 规格 |
|------|------|
| 页面水平内边距 | Desktop: 24px / Tablet: 16px / Mobile: 12px |
| 栅格 | 24 栅格系统，gutter: 24px |
| 常用列数 | 12（半宽）/ 8（三分之一）/ 6（四分之一）/ 4（六分之一） |

### 3.3 响应式断点

| 断点 | 宽度 | 布局变化 |
|------|------|---------|
| Desktop | ≥ 1200px | 完整布局，Sidebar 展开 |
| Tablet | 768px - 1199px | Sidebar 折叠为 Icon-only (64px) |
| Mobile | < 768px | Sidebar 隐藏，改为底部 Tab 导航 |

---

## 4. 新增组件提案流程

如果 Feature Design 需要 ui-standards.md 中不存在的组件，必须先提案：

**提案文件**: `{docs_internal}/requirements/{feature}/ui-design/COMPONENT-PROPOSAL.md`

**提案模板**:

```markdown
# 组件提案: {组件名称}

## 提案编号
`custom-{component-name}-{date}`

## 提案原因
[为什么现有组件无法满足需求]

## 提案规格
| 属性 | 值 |
|------|-----|
| 高度 | xx px |
| 宽度 | xx / 固定宽度 |
| 圆角 | xx px |
| 背景 | Token 或具体色值 |
| 文字 | Token 或具体色值 |
| 交互 | 默认/Hover/Active/Disabled |

## 建议纳入标准组件？
[是/否，原因]
```

**审核流程**: UI Designer → Tech Lead 评审 → 通过后更新 ui-standards.md → Feature Design 改为使用正式组件 ID。

---

## 5. QA 验证规则

QA 在验证 UI 时，对照本文档检查：

| 检查项 | 方法 |
|--------|------|
| 组件 ID 正确 | 对比 DESIGN.md 中引用的组件 ID 与 ui-standards.md |
| 颜色值正确 | 抽取页面实际色值，对比 Token 定义 |
| 间距正确 | 对比关键间距（padding/margin）与 Token 定义 |
| 高度正确 | 测量实际渲染高度，对比规范（40px / 48px / 60px） |
| 组件状态完整 | 测试默认/Hover/Active/Disabled/Loading 所有状态 |

> **注意**: QA 以 DESIGN.md 为验收基准，以 ui-standards.md 为规范依据。两份文档缺一不可。
