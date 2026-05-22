# 测试覆盖规范更新 — 任务记录

**时间**: 2026-04-30 18:08
**主题**: 区分共通和项目部分，更新测试覆盖规范

---

## 核心共识

### 问题
- 原有测试规范缺少「按功能/Bug 组织测试」的指导
- 没有规定每个变更必须有哪些测试覆盖
- Feature 和 Bug 修复的测试要求混在一起

### 共识
1. **测试按类型分层**：unit/integration/api/e2e（共通）
2. **测试按变更组织**：features/Fxxx/、bugs/Bxxx/（项目）
3. **每个变更必须有配套文档**：
   - Feature → TEST_COVERAGE.md + TEST_CASES.md
   - Bug → TEST_CASE.md

---

## 新增/更新的文件

### 共通部分（`_team/`）

| 文件 | 操作 | 说明 |
|------|------|------|
| `_team/templates/feature-test-template.md` | 新增 | Feature 测试覆盖声明模板 |
| `_team/templates/bug-test-template.md` | 新增 | Bug 复现测试用例模板 |
| `_team/templates/README.md` | 新增 | 模板索引 |
| `_team/workflows/roles/development-standards.md` | 更新 | 角色加载映射 + 引用新模板 |
| `_team/workflows/roles/bugfix-standards.md` | 更新 | 成果清单 + 新增模板引用 |
| `_team/workflows/roles/specialized-tests.md` | 更新 | API 测试增加 CRUD + 列表回显检查项 |

### 项目部分（`projects/orig-cms/tests/`）

| 文件 | 操作 | 说明 |
|------|------|------|
| `projects/orig-cms/tests/README.md` | 重写 | 测试目录结构规范 |

---

## 新增的测试模板结构

```
_team/templates/
├── feature-test-template.md     # Feature 测试覆盖声明
└── bug-test-template.md         # Bug 复现测试用例

projects/orig-cms/tests/
├── features/                    # 按功能组织的测试
│   └── Fxxx-{name}/
│       ├── TEST_COVERAGE.md     # 必须：规定测什么
│       ├── TEST_CASES.md        # 必须：具体测试用例
│       └── integration_*.go     # 功能级集成测试
│
└── bugs/                        # 按 Bug 组织的测试
    └── Bxxx-{name}/
        ├── TEST_CASE.md         # 必须：复现测试用例
        └── regression_*.go      # 回归测试
```

---

## 必须覆盖的测试类别（Feature）

| # | 测试类别 | 说明 |
|---|---------|------|
| 1 | API CRUD 流程测试 | Create→Read→Update→Delete + 列表回显验证 |
| 2 | 数据库直写测试 | DB 写入→API 读取验证 |
| 3 | 字段状态测试 | 多状态字段显示逻辑 |
| 4 | 联合条件测试 | 复合条件→目标字段（如 A∧B∧C → D）|
| 5 | 边界值测试 | null/空/超长/特殊字符 |
| 6 | 权限测试（可选） | 有权限控制时 |
| 7 | 错误处理测试（可选） | 4xx/5xx 场景时 |

---

## 后续行动

1. [ ] 在已有 Feature 目录中补充 TEST_COVERAGE.md（如 F014）
2. [ ] 按规范创建第一个 Bug 测试目录（如 B001）
3. [ ] Review 所有现有测试文件，确认符合规范
4. [ ] 将此规范应用到实际开发流程中