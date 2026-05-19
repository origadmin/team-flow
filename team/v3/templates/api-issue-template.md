# API 问题报告模板 — v2 beads-native

> **位置**: `{DOCS_INTERNAL}/reports/api/API_ISSUE_{id}.md`

---

## 基本信息

| 字段 | 内容 |
|------|------|
| 报告人 | {name} |
| 报告日期 | {YYYY-MM-DD} |
| API 端点 | {endpoint} |
| 问题类型 | {契约不符/响应错误/性能问题} |

---

## 问题描述

### 预期行为
{根据契约应该怎样}

### 实际行为
{实际发生了什么}

### 复现步骤
1. {step 1}
2. {step 2}

---

## 请求/响应示例

### 请求
```http
GET /api/v1/{resource}
Authorization: Bearer {token}
```

### 预期响应
```json
{
  "code": 0,
  "data": { ... }
}
```

### 实际响应
```json
{
  "code": 500,
  "message": "..."
}
```

---

## 责任判定

| 问题类型 | 负责人 | 处理方式 |
|---------|--------|---------|
| 响应格式与契约不符 | 后端 | 后端修改 |
| 前端调用了错误的端点 | 前端 | 前端修改 |
| 业务逻辑计算错误 | 后端 | 后端修改 |

---

## beads 状态更新

```bash
flow task create "API Issue: {endpoint}" -t bug -p 1
```
