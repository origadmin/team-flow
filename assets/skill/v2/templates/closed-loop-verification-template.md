# 闭环验证报告模板 — v2 beads-native

---

## beads 关联

| 字段 | 内容 |
|------|------|
| Issue ID | `<beads-id>` |
| Milestone | {M-ID} |

---

## 基本信息

- Milestone: {M-ID}
- 版本: v{version}
- 验证日期: {date}
- 验证人: QA Engineer

---

## 验证范围

| 任务 ID | 设计文档 | 实现情况 | 闭环 |
|---------|---------|---------|------|
| F001 | SPEC.md | ... | ✅/❌ |

---

## 测试补充

| 补充测试用例 | 测试结果 |
|-------------|---------|
| ... | 通过/失败 |

---

## 遗留问题

| 问题 | 严重程度 | 处理方式 |
|------|---------|---------|
| ... | Major/Minor | ... |

---

## 最终结论

- [ ] 功能闭环: 是/否
- [ ] 质量闭环: 是/否
- [ ] QE 签字: {name}
- [ ] PM 放行: {name}

---

## beads 状态更新

```bash
flow task update <id> --add-label phase:closed --notes "Closed-loop verification passed"
```
