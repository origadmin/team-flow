# team-flow v2 完成门禁清单 - shared-checklist.md

> **更新日期**: 2026-05-18
> **用途**: Layer 3 完成门禁和详细检查清单（含 Output Guard 前置检查）

---

## Feature 完成门禁

```
Output Guard 检查（前置 — 未通过禁止继续）:
- [ ] Output Guard Summary 已输出（全部 ✅）
- [ ] AC Compliance Matrix 无 ❌ 项
- [ ] 编译/类型检查实际执行并展示输出
- [ ] 测试实际执行并展示通过数量
- [ ] Self-Critique 已执行

文档与实现检查:
- [ ] R0_NAVIGATION_MATRIX.md 存在且非空
- [ ] R0 中定义的所有入口已在代码中实现
- [ ] SPEC.md 存在且非空
- [ ] AC.md 存在且非空
- [ ] R1_DATA_MODEL.md 存在且非空
- [ ] R2_STATE_MACHINE.md 存在且非空
- [ ] R3_API_CONTRACT.md 存在且非空
- [ ] 核心代码已实现
- [ ] 单元测试存在且通过
- [ ] Pipeline 通过（Step 1-5）
- [ ] 代码无中文注释
- [ ] SCOPE.md 已生成
- [ ] TEST_COVERAGE.md 存在且非空（使用 feature-test-template.md）
- [ ] 前后端接口路径对照: 前端API调用路径 vs 后端路由注册路径 100% 匹配
- [ ] 前后端参数命名对照: 前端请求参数名 vs 后端期望参数名 100% 匹配
- [ ] 前后端响应结构对照: 前端TypeScript类型 vs 后端Proto/JSON响应 100% 匹配
- [ ] Handler注册完整性: Proto定义的所有API均有对应Handler注册
- [ ] beads task 状态已更新 (flow task update <id> --status closed)
- [ ] {DOCS_INTERNAL}/PROJECT.md 已同步（如有范围变更）
- [ ] 用户确认前不得归档
```

---

## Bugfix 完成门禁

```
Output Guard 检查（前置 — 未通过禁止继续）:
- [ ] Output Guard Summary 已输出（全部 ✅）
- [ ] AC Compliance Matrix 无 ❌ 项
- [ ] 编译/类型检查实际执行并展示输出
- [ ] 测试实际执行并展示通过数量
- [ ] Self-Critique 已执行（含数据流 + 真实场景验证）

⚠️ 测试验证执行（必须展示实际命令输出）:
- [ ] 后端Bug: go build ./... 编译通过（展示输出）
- [ ] 后端Bug: go test ./... 全部通过（展示通过数量）
- [ ] 前端Bug: bun run typecheck 类型检查通过（展示输出）
- [ ] 前端Bug: bun run lint 代码检查通过（展示输出）
- [ ] 前端Bug: bun run test 全部通过（展示通过数量）
- [ ] Bug 复现测试通过（展示测试输出）
- [ ] 全量回归测试通过（展示测试输出）

文档检查:
- [ ] RCA.md 存在且包含：现象、根因、影响、预防
- [ ] TEST_CASE.md 存在且包含：复现步骤、预期结果、验证结果
- [ ] 代码无中文注释
- [ ] SCOPE.md 已生成
- [ ] TEST_CASE.md 存在且包含复现步骤和回归用例（使用 bug-test-template.md）
- [ ] 修复涉及API变更？→ 是则执行前后端接口路径/参数/响应对照检查

[v2] 数据流追踪（涉及API/权限/状态/交互的Bug必须）:
- [ ] RCA.md 包含"数据流追踪"章节
- [ ] 断点已定位到具体环节（不是"可能是xxx"）

[v2] 真实场景验证（涉及API/权限/状态/交互的Bug必须）:
- [ ] TEST_CASE.md 包含真实场景验证（非仅mock测试）
- [ ] 覆盖默认值/零值/边界条件
- [ ] 验证结果为通过（不是"已编写"）

[v2] 运行时验证（涉及API/权限/状态/交互的Bug必须）:
- [ ] 后端Bug需HTTP请求验证通过（非仅UseCase单元测试）
- [ ] 前端Bug需UI运行时验证（非仅组件mock测试）
- [ ] UI_VERIFICATION.md 已生成（使用 ui-verification-template.md）
- [ ] 启动dev server，打开Bug涉及页面
- [ ] 检查页面内容（文本/数据/组件正确显示）
- [ ] 执行交互行为（点击/输入/提交）
- [ ] 重现Bug原始触发步骤，确认Bug现象消失
- [ ] 副作用检查（相邻功能、导航正常）
- [ ] 截图保存到 {DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/screenshots/
- [ ] 截图命名: B{NNN}-R{N}-{3位步骤号}-{动作}-{状态}.png
- [ ] 截图数量 >= Bug类型最低要求
- [ ] 修复后的完整链路已验证（非仅断点环节）

流程检查:
- [ ] beads task 状态已更新 (flow task update <id> --status closed)
- [ ] {DOCS_INTERNAL}/PROJECT.md 已同步（如有影响模块变更）
- [ ] 用户确认前不得归档
```

---

## R-Phase 完成门禁

**R-Phase 1（闭环验证）**:
- [ ] 集成/回归测试通过
- [ ] 功能对照设计文档 100% 闭环
- [ ] 非功能需求达标
- [ ] 无 Major+ Bug
- [ ] 闭环验证报告已输出
- [ ] QA Engineer 签字

**R-Phase 2（验收放行）**:
- [ ] 验收标准 100% 满足
- [ ] 业务闭环确认
- [ ] PM 签字放行

**R-Phase 3（上线部署）**:
- [ ] CI Pipeline 全部 Job 通过
- [ ] Docker 镜像构建成功
- [ ] 部署到生产环境成功
- [ ] 烟雾测试通过
- [ ] 监控指标正常
- [ ] CHANGELOG 已更新
- [ ] 版本号已更新（SemVer）

---

## 质量门（强制）

### QG-1: Think Before Coding

**执行时机**：拿到任务后、执行任何操作前（PRE-FLIGHT 暂停点）。

- **假设**：我正在基于哪些未确认的假设行动？
- **风险**：如果这个假设错了，最坏的结果是什么？
- **澄清**：有没有我应该先问清楚的点？

> 不确定就停，不要凭记忆继续。

### QG-2: Simplicity First

**执行时机**：每次输出前自检。

- 能用更少的代码解决吗？
- 我在加未请求的特性吗？
- 有没有过度设计？

### QG-3: Surgical Changes

**执行时机**：修改代码时。

- 只改任务范围内必要的代码
- 不顺手"优化"或重构旁边的代码
- 自己改动产生的孤儿代码自己清理

### QG-4: Goal-Driven Execution

**执行时机**：任务开始前。

每个任务必须先定义成功标准，再执行：

```markdown
## goal_plan

**目标**：{一句话描述最终交付物}

**Plan**：
1. [步骤] → verify: [该步骤完成的具体标准]
2. [步骤] → verify: [该步骤完成的具体标准]

**成功标准**：
- [ ] {标准1}
- [ ] {标准2}

**回滚预案**：如果 {条件} 失败 → {操作}
```

> 每步 verify 必须具体，不能用"完成"模糊描述

---

## Anti-Patterns

❌ **DON'T**:
- Skip Dolt pull/push in multi-agent setups
- Leave tasks in `phase:ready` unassigned
- Close tasks without recording outcome
- Create duplicate tasks (check first with `flow task list`)
- Dump deliverable content into beads notes — **write to independent deliverable files**
- 中文注释
- 跳过 RCA / SCOPE / 验收标准
- 写入 `framework/{TEAM_PATH}/`（只读层）
- 写入 `{PROJECT}/_docs/`（正确路径：`framework/_docs/{project}/`）
- Edit task-pool-export.md to change task state (use `flow task` commands)
- Assume file state matches beads state (always query beads)

### beads notes 用途

只记录进度摘要和交接信息：
```
✅ flow task update <id> --notes "COMPLETED: RCA.md written, fix applied, tests passing"
❌ flow task update <id> --notes "Root cause: handler.go wraps response with {code,message,data}..."
```

### 成果物写入规则

| 内容类型 | 正确位置 | 禁止位置 |
|---------|---------|---------|
| 根因分析 | `{DOCS_INTERNAL}/reports/bugs/B{NNN}/R{n}/RCA.md` | beads notes |
| 复现测试 | `{DOCS_INTERNAL}/reports/bugs/B{NNN}/R{n}/TEST_CASE.md` | beads notes |
| 变更报告 | `{DOCS_INTERNAL}/reports/changes/C{NNN}/SCOPE.md` | beads notes |
| 需求规格 | `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/SPEC.md` | beads notes |
| 设计决策 | `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/R1-R5.md` | beads notes |
