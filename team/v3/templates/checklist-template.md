# CHECKLIST.md 模板 — v2 beads-native

> **位置**: `.team/checklist.md`

---

## beads 关联

| 字段 | 内容 |
|------|------|
| Issue ID | `<beads-id>` |
| Feature ID | F{NNN} |

---

## 使用说明

此文件定义项目级验收标准覆盖，覆盖 `team/v2/workflows/shared-checklist.md` 中的默认标准。

**读取规则**：
- Dev 在完成实现前必须读取此文件
- QA 在验证前必须读取此文件
- Checklist 条目未通过，禁止报告任务完成

---

## 验收标准覆盖

### 1. 代码质量

- [ ] {所有新增代码必须有单元测试（覆盖率 > 80%）}
- [ ] {所有导出函数必须有英文注释}
- [ ] {禁止使用中文注释}
- [ ] {禁止使用 `fmt.Print*`（使用 `log` 或 `slog`）}

### 2. Git 规范

- [ ] {Commit message 必须使用英文}
- [ ] {Commit message 必须遵循 Conventional Commits 规范}
- [ ] {每个 commit 必须关联 beads issue ID}

### 3. 测试规范

- [ ] {使用 TDD 流程：先写测试，再实现}
- [ ] {测试必须覆盖 happy path 和 error path}
- [ ] {集成测试必须模拟真实场景}

### 4. 文档规范

- [ ] {新增功能必须有设计文档}
- [ ] {API 变更必须更新 API 文档}
- [ ] {用户-facing 功能必须有用户文档}

---

## beads 状态更新

```bash
flow task update <issue-id> --notes "CHECKLIST: {summary}"
```
