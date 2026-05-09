# {Fxxx}: {功能名称} — 测试覆盖声明

## beads 关联

| 字段 | 值 |
|------|-----|
| Issue ID | `<beads-id>` |
| task ID | F{xxx} |
| Phase | phase:implement → phase:verify |
| Subsystem | subsystem:backend |

**功能版本**: v{version}
**关联需求**: `{docs_internal}/requirements/Fxxx/`
**测试代码路径**: `{PROJECT}/tests/features/F{xxx}-{short-name}/`

---

## 必须覆盖的测试类别

> ⚠️ 以下类别为**最低要求**，每个 Feature 必须全部覆盖

| # | 测试类别 | 覆盖状态 | 测试文件 | 备注 |
|---|---------|---------|---------|------|
| 1 | **API CRUD 流程测试** | ☐ 未覆盖 / ☑ 已覆盖 | `integration_crud_test.go` | Create→Read→Update→Delete |
| 2 | **数据库直写测试** | ☐ 未覆盖 / ☑ 已覆盖 | `db_write_test.go` | DB 写入→API 读取验证 |
| 3 | **字段状态测试** | ☐ 未覆盖 / ☑ 已覆盖 | `field_state_test.go` | 多状态字段显示逻辑 |
| 4 | **联合条件测试** | ☐ 未覆盖 / ☑ 已覆盖 | `joint_condition_test.go` | 复合条件→目标字段 |
| 5 | **边界值测试** | ☐ 未覆盖 / ☑ 已覆盖 | `boundary_test.go` | null/空/超长/特殊字符 |
| 6 | **权限测试**（可选）| ☐ 未覆盖 / ☑ 不适用 | `permission_test.go` | 有权限控制时填写 |
| 7 | **错误处理测试**（可选）| ☐ 未覆盖 / ☑ 不适用 | `error_test.go` | 4xx/5xx 场景时填写 |

---

## 1. API CRUD 流程测试

**测试内容**: Create → Read → Update → Read验证 → Delete → VerifyGone

**流程检查表**:

| 步骤 | 操作 | API | 验证点 |
|------|------|-----|--------|
| 1 | 创建资源 | POST /api/v1/{resource} | 返回 id，初始状态正确 |
| 2 | 获取单条 | GET /api/v1/{resource}/{id} | 所有字段与创建时一致 |
| 3 | 列表查询 | GET /api/v1/{resource} | 新记录在列表中 ✅ |
| 4 | 更新资源 | PATCH /api/v1/{resource}/{id} | 更新成功 |
| 5 | 列表回显 | GET /api/v1/{resource} | **更新后的值在列表中显示** ✅ |
| 6 | 删除资源 | DELETE /api/v1/{resource}/{id} | 返回 204 |
| 7 | 确认删除 | GET /api/v1/{resource}/{id} | 返回 404 ✅ |

**覆盖状态**: ☐ 未覆盖 | ☑ 已覆盖

---

## 2. 数据库直写测试

**测试内容**: 绕过 API 直接操作数据库，验证数据落库正确性

**检查表**:

| # | 操作 | 验证点 | 覆盖状态 |
|---|------|--------|---------|
| 1 | 直接 INSERT → API 读取 | 数据一致性 | ☐ |
| 2 | API 写入 → DB 查询 | 字段值正确 | ☐ |
| 3 | UPDATE DB → API 感知 | 变更被 API 反映 | ☐ |
| 4 | 并发写入 | 最终状态正确 | ☐ |
| 5 | 事务回滚 | 无残留数据 | ☐ |

---

## 3. 字段状态测试

### 3.1 字段状态矩阵

| 字段名 | 状态值 | 列表显示 | 说明 |
|--------|--------|---------|------|
| status | pending | ❌ 不显示 | 待处理 |
| status | active | ✅ 显示 | 正常 |
| status | disabled | ❌ 不显示 | 已禁用 |

### 3.2 状态转换测试

| 当前状态 | 操作/事件 | 目标状态 | 预期行为 | 覆盖状态 |
|---------|---------|---------|---------|---------|
| pending | 审核通过 | active | 列表从❌变为✅ | ☐ |
| active | 下架 | disabled | 列表从✅变为❌ | ☐ |
| disabled | 重新上架 | active | 列表从❌变为✅ | ☐ |

---

## 4. 联合条件测试

### 4.1 联合条件定义

| 规则ID | 条件描述 | 计算公式 | 目标字段 |
|--------|---------|---------|---------|
| JC-001 | A、B、C 全为 true 时 | `D = A && B && C` | D |
| JC-002 | A 或 B 任一为 true 时 | `E = A \|\| B` | E |

### 4.2 真值表测试

**JC-001: D = A ∧ B ∧ C**

| 用例ID | A | B | C | 预期 D | 实际 D | 结果 |
|--------|---|---|---|--------|--------|------|
| JC-001-1 | 1 | 1 | 1 | 1 | | ☐ |
| JC-001-2 | 1 | 1 | 0 | 0 | | ☐ |
| JC-001-3 | 1 | 0 | 1 | 0 | | ☐ |
| JC-001-4 | 0 | 1 | 1 | 0 | | ☐ |
| JC-001-5 | 0 | 0 | 0 | 0 | | ☐ |

---

## 5. 边界值测试

| 字段类型 | 边界值 | 预期结果 | 覆盖状态 |
|---------|--------|---------|---------|
| string | 空字符串 `""` | | ☐ |
| string | null | | ☐ |
| string | 超长字符 (>1000) | | ☐ |
| string | 特殊字符 (`<>"'&`) | | ☐ |
| int | 0 | | ☐ |
| int | 负数 | | ☐ |
| int | 最大值溢出 | | ☐ |

---

## 6. API 覆盖清单

| API 路径 | 方法 | 测试类别 | 覆盖状态 |
|---------|------|---------|---------|
| /api/v1/{resource} | GET | 列表查询 | ☐ |
| /api/v1/{resource} | POST | 创建 | ☐ |
| /api/v1/{resource}/{id} | GET | 详情 | ☐ |
| /api/v1/{resource}/{id} | PATCH | 更新 | ☐ |
| /api/v1/{resource}/{id} | DELETE | 删除 | ☐ |

---

## 测试执行记录

| 日期 | 执行人 | 覆盖类别数 | 通过 | 失败 | Bug |
|------|--------|---------|------|------|-----|
| {YYYY-MM-DD} | {dev} | 0/7 | 0 | 0 | - |

---

## 关联文件

| 类型 | 路径 |
|------|------|
| 需求文档 | `_docs/.../requirements/Fxxx/SPEC.md` |
| 接口契约 | `_docs/.../requirements/Fxxx/R3_API_CONTRACT.md` |
| 测试代码目录 | `{PROJECT}/tests/features/F{xxx}-{short-name}/` |

---

## beads 状态更新

完成测试后执行：

```bash
flow tools beads update <issue-id> --notes "TEST_COVERAGE: {x}/7 categories covered, {pass}/{fail} results"
flow tools beads update <issue-id> --add-label phase:verify --remove-label phase:implement
```
