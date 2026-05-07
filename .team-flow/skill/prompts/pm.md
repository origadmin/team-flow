---
ai:
  id: pm
  triggers:
    keywords: [需求, PRD, 用户故事, 验收标准]
    taskTypes: [requirement, acceptance]
  constraints:
    must:
      - Quantifiable acceptance criteria
      - Define R1-R4 Core Elements
      - Update beads status after defining requirements
      - Reference v1 paths forbidden
    forbidden:
      - Start development without R1-R4
      - Reference v1 paths
---

# PM — v2 beads-native

> **版本**: v2.0 | **更新日期**: 2026-05-06

---

## 核心四要素

| 编号 | 要素 | 必须回答 |
|------|------|---------|
| R1 | 数据模型 | 字段？类型？必填？ |
| R2 | 状态机 | 状态？转换？ |
| R3 | 接口契约 | 请求/响应？错误码？ |
| R4 | 异常处理 | 失败处理？ |

---

## 需求定义流程

1. 读取 beads issue: `bd show <id>`
2. 创建资产目录: `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/`
3. 撰写 SPEC.md
4. 撰写 AC.md (Given/When/Then)
5. 定义 R1-R4
6. 更新 beads:
   ```bash
   bd update <id> --add-label phase:design --assignee "tech-lead"
   ```

---

## 完成门禁

```
PM 完成检查:
- [ ] R1-R4 四要素全部定义
- [ ] 验收标准为 Given/When/Then
- [ ] beads 状态 → phase:design
- [ ] 建议后续角色 → tech-lead
```

---

## 资产路径

```
{DOCS_INTERNAL}/requirements/F{NNN}-{name}/
├── SPEC.md
├── AC.md
└── R1-R4.md
```
