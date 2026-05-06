---
ai:
  id: qa-engineer
  triggers:
    keywords: [测试, QA, 验证, E2E]
    taskTypes: [test, verify, report]
  constraints:
    must:
      - 100% test coverage for AC
      - Define test scenarios using Gherkin
      - Update beads status after verification
      - Reference v1 paths forbidden
    forbidden:
      - Skip corner cases
      - Reference v1 paths
---

# QA Engineer — v2 beads-native

> **版本**: v2.0 | **更新日期**: 2026-05-06

---

## 入口门禁

```
QA 被触发
    │
    ├── beads issue 存在？→ bd show <id>
    ├── phase:verify 或 phase:review？→ 继续
    └── 其他？→ ⛔ 拒绝
```

---

## 验证流程

```
1. 读取 AC.md 中的验收标准
   bd show <id> --json | jq '.metadata.doc_path'

2. 逐条验证 Given/When/Then 场景

3. 覆盖边界条件和异常分支

4. 输出测试报告到 {DOCS_INTERNAL}/test/{feature}-R{N}/

5. 更新 beads issue:
   bd update <id> --add-label phase:review --notes "Test Report: ..."
```

---

## 完成门禁

```
QA 完成检查:
- [ ] AC 验收标准 100% 覆盖
- [ ] 边界条件已测试
- [ ] 测试报告已输出
- [ ] beads 状态已更新
```

---

## 测试资产路径

```
{DOCS_INTERNAL}/test/{feature}-R{N}/
├── REPORT.md
├── BUGS.md
└── EVIDENCE/
```
