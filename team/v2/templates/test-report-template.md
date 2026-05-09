# 测试报告模板 — v2 beads-native

---

## beads 关联

| 字段 | 内容 |
|------|------|
| Issue ID | `<beads-id>` |
| 报告日期 | {YYYY-MM-DD} |

---

## 基本信息

| 字段 | 内容 |
|------|------|
| 模块名称 | {Module} |
| 测试时间 | {YYYY-MM-DD HH:MM ~ HH:MM} |
| 测试人员 | {Tester} |

---

## 测试概况

| 类型 | 用例数 | 通过 | 失败 | 通过率 |
|------|--------|------|------|--------|
| 功能测试 | {N} | {N} | {N} | {XX%} |
| API 测试 | {N} | {N} | {N} | {XX%} |
| E2E 测试 | {N} | {N} | {N} | {XX%} |
| **总计** | **{N}** | **{N}** | **{N}** | **{XX%}** |

---

## 测试结果

| 用例 ID | 用例名称 | 优先级 | 结果 | Bug ID |
|---------|----------|--------|------|--------|
| TC-001 | {名称} | P0 | Pass/Fail | `<beads-id>` |

---

## Bug 统计

| 严重程度 | 数量 | 已修复 | 未修复 |
|----------|------|--------|--------|
| Critical | {N} | {N} | {N} |
| Major | {N} | {N} | {N} |
| Minor | {N} | {N} | {N} |

---

## 结论

- [ ] 所有 P0 用例通过
- [ ] 无 Critical Bug

### beads 状态更新

```bash
flow task update <id> --add-label phase:review --notes "Test report: {path}"
```
