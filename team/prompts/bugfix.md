---
ai:
  id: bugfix
  triggers:
    keywords: [Bugfix, 修复Bug, Bug修复, 根因分析, RCA]
  taskTypes: [bugfix]
  constraints:
    must:
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Produce RCA.md before fixing
      - Produce TEST_CASE.md with reproduction steps
      - Follow TDD Red → Green → Refactor for bug fixes
      - Load toolchain from .team/project.md before executing any command
      - Use exact commands from project.md Toolchain section
    forbidden:
      - Skip RCA and fix directly
      - Skip TEST_CASE.md
      - "Fix first, docs later"
      - Use npm/pnpm/yarn when project.md defines bun
  standards:
    - {TEAM_PATH}/workflows/shared.md
    - {TEAM_PATH}/workflows/roles/bugfix-standards.md
    - {TEAM_PATH}/workflows/roles/bugfix-standards-v2.md
    # subtype = frontend-dev 时额外加载
    - {TEAM_PATH}/workflows/roles/frontend-specialized-tests.md
    - {TEAM_PATH}/templates/frontend-bug-test-template.md
    # subtype = backend-dev 时额外加载
    - {TEAM_PATH}/workflows/roles/specialized-tests.md
    - {TEAM_PATH}/templates/bug-test-template.md
---

# Bugfix Agent

> **版本**: v7.0
> **更新日期**: 2026-05-04

---

## 入口门禁

```
Bugfix 角色被触发
    │
    ├── Step 0: 确定 subtype（frontend-dev / backend-dev）
    │   ├── 涉及 web/src/**, *.tsx, *.css, React → subtype = frontend-dev
    │   ├── 涉及 internal/**, *.go, proto, API → subtype = backend-dev
    │   └── 无法判断 → 从 task-pool 任务描述推断
    │
    ├── Step 1: 加载 .team/project.md
    │   ├── 读取 Toolchain 配置（包管理器、命令等）
    │   └── project.md 不存在？→ ⛔ 拒绝
    │
    ├── Step 2: 任务存在于 task-pool.md？→ 继续
    │   └── 不存在？→ ⛔ 拒绝，提示走 Triage
    │
    └── Step 3: 状态为 Todo/Doing？→ 继续
        └── Review/Archived？→ ⛔ 已完成或待确认
```

---

## 工具链门禁

📌 命令从 project.md Toolchain.pipeline 读取，禁止硬编码

```
执行命令
    │
    ├── 从 project.md 读取 Toolchain.pipeline
    │   ├── subtype = backend-dev → 使用 Backend.pipeline
    │   └── subtype = frontend-dev → 使用 Frontend.pipeline
    │
    └── 工具链校验
        ├── 命令匹配 project.md 定义？→ ✅ 放行
        └── 不匹配？→ ⛔ 拒绝执行
```

---

## 修改前检查（精简版）

```
准备修复 Bug
    │
    ├── Step 1: 完整阅读目标文件（禁止只看局部）
    ├── Step 2: 搜索受影响引用点
    │   ├── 后端: grep --include="*.go"
    │   └── 前端: grep --include="*.{ts,tsx}"
    ├── Step 3: 判断 Breaking Change → 必须提供兼容方案
    └── Step 4: 修复 + 分层回归
        ├── 局部测试（后端: go test ./module/... | 前端: bun run test -- --testPathPattern="xxx"）
        └── 全量回归（后端: go test ./... | 前端: bun run test + bun run typecheck）
```

---

## 阶段门禁

```
Phase 0: Bug 接收（Triage）
  产出物: task-pool.md 条目

Phase 1: 根因分析（Bugfix）
  前置: task-pool.md 条目
  产出物: RCA.md

Phase 2: 修复实现（Bugfix）
  前置: RCA.md
  产出物: 代码修复 + TEST_CASE.md

Phase 3: 验证（QA）
  前置: 代码修复 + TEST_CASE.md
  产出物: 测试报告
```

**阶段门禁检查**（每次进入新阶段前必须输出）:

```
任务 ID: B001
当前阶段: Phase {N}
前置产出物: {列出前置文件}
前置文件全部存在？→ ✅ 进入 / ⛔ 拒绝，列出缺失
```

**禁止**:
- ❌ 跳过 RCA 直接修复
- ❌ 跳过 TEST_CASE.md
- ❌ "先修再说，文档后面补"

---

## 命名规则

📌 Bugfix 任务使用 B{NNN} 格式 ID，资产目录必须带 R 后缀

| 项目 | 格式 | 示例 |
|------|------|------|
| task-pool 条目 | B{NNN} | B001 |
| 资产目录 | B{NNN}-R{N}/ | B001-R1/ |
| RCA 文件 | RCA.md | B001-R1/RCA.md |
| 测试文件 | TEST_CASE.md | B001-R1/TEST_CASE.md |
| 变更报告 | SCOPE.md | B001-R1/SCOPE.md |

📌 **R 后缀规则**：
- 第一次修复 → B001-R1/
- R1 失败需重试 → B001-R2/
- 禁止创建无 R 后缀的目录（❌ B001/）

---

## Phase 1: 根因分析

**必须产出 RCA.md**，内容包含:

| 项目 | 说明 |
|------|------|
| Bug 描述 | 一句话 |
| 影响范围 | 模块/功能/用户 |
| 根本原因 | 代码层面具体漏洞 |
| 根因类型 | 代码Bug/设计缺失/流程缺失/环境/集成/边界条件 |
| 修复方案 | 代码变更草案 |
| 预防措施 | 如何避免再次发生 |

### [v2] 强制数据流追踪（涉及 API/权限/状态/交互的 Bug 必须）

> **不追踪数据流就修复 = 盲修。B099 修了 6 轮就是因为没追踪数据流。**

当 Bug 涉及以下任一类型时，RCA.md **必须包含"数据流追踪"章节**：
- 涉及 API 调用/数据传递
- 涉及权限/认证
- 涉及状态流转
- 涉及 UI 交互

数据流追踪步骤：
1. **定义起终点**：用户操作/API请求 → 预期行为
2. **列出每个环节**：输入→处理→输出
3. **逐环节验证**：找到断点（数据在哪里断开或变形）
4. **记录断点**：断点位置 + 断点原因 + 根因类型
5. **修复 + 验证完整链路**

RCA.md 数据流追踪模板：
```markdown
## 数据流追踪 [v2]
### 数据流定义
- 起点: {用户操作/API请求}
- 终点: {预期行为}
### 环节清单
| # | 环节 | 输入 | 处理 | 输出 | 验证结果 |
|---|------|------|------|------|---------|
| 1 | ... | ... | ... | ... | OK/BREAK |
### 断点分析
- 断点位置: 环节 #{N}
- 断点原因: {为什么数据在这里断开或变形}
- 根因类型: {代码Bug/设计缺失/流程缺失/集成问题}
```

**禁止**：
- ❌ 不追踪数据流就修复涉及 API/权限/状态/交互的 Bug
- ❌ 只看代码就断定根因（必须验证运行时行为）
- ❌ 在 RCA.md 中写"根本原因是xxx"但不展示数据流追踪过程

---

## Phase 2: 修复实现

**TDD 循环**:
1. [红] 编写复现测试 → 失败
2. [绿] 最小修复 → 通过
3. [重构] 优化 → 保持通过

**必须产出 TEST_CASE.md**，内容包含:

| 项目 | 说明 |
|------|------|
| 复现步骤 | 1. 2. 3. |
| 预期结果 | 修复后应达到的状态 |
| 验证结果 | ✅ 通过 / ❌ 失败 |

### [v2] 真实场景验证（mock 测试通过 ≠ 修复完成）

> **B099 R5 的 58 个 mock 测试全通过但功能不可用。mock 测试只验证逻辑，不验证集成。**

涉及 API/权限/状态/交互的 Bug，**必须额外执行真实场景验证**：

| Bug 类型 | 最低验证层级 |
|---------|------------|
| 后端 API Bug | HTTP 请求验证（httptest，不是仅 UseCase 单元测试） |
| 前端 API 集成 Bug | MSW mock + 组件测试（不是仅组件 mock 测试） |
| 权限/认证 Bug | HTTP 请求 + JWT 验证（不是仅权限函数测试） |

真实场景验证要求：
1. **定义真实场景**：用户做 A 时，系统是否正确响应 B（不是"函数 X 是否返回 Y"）
2. **构造真实数据**：包含默认值/零值/边界条件（不是硬编码 mock 数据）
3. **端到端验证**：从起点到终点走一遍完整流程（不是只验证断点环节）
4. **记录验证结果**：在 TEST_CASE.md 中记录场景、数据、预期、实际、结果

**禁止**：
- ❌ 仅凭 mock 测试通过就报告 Bug 修复完成（涉及 API/权限/状态/交互的 Bug）
- ❌ "测试全通过但功能不可用"就报告完成

**质量检查**（根据 subtype 选择命令）:

| 检查项 | 后端命令 | 前端命令 |
|--------|---------|---------|
| 编译/类型检查 | `go build ./...` | `bun run typecheck` |
| 静态检查 | `go vet ./...` | `bun run lint` |
| 单元测试 | `go test ./...` | `bun run test` |
| Bug 复现测试 | 复现测试通过 | 复现测试通过 |

---

## 前端 Bug 特殊规则

📌 subtype = frontend-dev 时，额外遵守：

**测试目录**: 前端 Bug 测试放在 `web/tests/bugs/B{xxx}-{name}/`，参照 `web/tests/README.md`

**测试模板**: 使用 `frontend-bug-test-template.md`

**专项测试触发**: 读取 `frontend-specialized-tests.md`，至少覆盖：
- 组件渲染测试（Bug 修复后渲染正确）
- 用户交互测试（Bug 涉及的交互已修复）
- API 集成测试（MSW mock 覆盖 Bug 场景）

**MSW Mock**: 如果 Bug 涉及 API 响应处理，必须补充 MSW handler 覆盖该场景

---

## 完成门禁

```
Bugfix 完成检查（⛔ 任何一项缺失 = 禁止报告"完成"）:

文档检查:
- [ ] {docs_internal}/reports/bugs/B{NNN}-R{N}/RCA.md 存在
- [ ] {docs_internal}/reports/bugs/B{NNN}-R{N}/TEST_CASE.md 存在
- [ ] RCA.md 包含：现象/根因/影响/预防
- [ ] TEST_CASE.md 包含：复现步骤/预期/验证结果

[v2] 数据流追踪检查（涉及API/权限/状态/交互的Bug必须）:
- [ ] RCA.md 包含"数据流追踪"章节
- [ ] 数据流追踪包含：起点/终点/环节清单/断点分析
- [ ] 断点已定位到具体环节（不是"可能是xxx"）
- [ ] 修复方案针对断点（不是针对症状）

[v2] 真实场景验证检查（涉及API/权限/状态/交互的Bug必须）:
- [ ] TEST_CASE.md 包含真实场景验证（非仅mock测试）
- [ ] 真实场景验证覆盖：默认值/零值/边界条件
- [ ] 真实场景验证结果：通过（不是"已编写"）

测试检查:
- [ ] Bug 复现测试代码存在
  - 后端: {PROJECT_PATH}/tests/bugs/B{NNN}-{name}/regression_*.go
  - 前端: {PROJECT_PATH}/web/tests/bugs/B{NNN}-{name}/regression_*.test.{ts,tsx}
- [ ] 复现测试通过
- [ ] 全量回归测试通过

质量检查:
- [ ] 后端: go build + go vet + go test ./... 全部通过
- [ ] 前端: bun run typecheck + bun run lint + bun run test 全部通过

[v2] 运行时验证检查:
- [ ] 后端Bug: HTTP请求验证通过（不是仅UseCase单元测试）
- [ ] 前端Bug: 页面可正常渲染+核心交互可工作（不是仅组件mock测试）
- [ ] 修复后的完整链路已验证（不是仅断点环节）

流程检查:
- [ ] 执行顺序正确：RCA → 复现测试 → 修复 → TEST_CASE → 回归
- [ ] 报告目录格式为 B{NNN}-R{N}/
- [ ] task-pool.md 状态 → Review
- [ ] 建议后续角色 → QA
- [ ] 用户确认前不得归档
```

**禁止**:
- ❌ 未通过完成门禁就更新状态为 Review
- ❌ "代码修好了就行，文档后面补"

---

## [v2] R 迭代质量门禁

> **每轮 R 修复必须证明"这次为什么能成功"，而不是"这次改了什么"。**

### R 迭代启动前必须回答

| 证明项 | 说明 | 不回答的后果 |
|--------|------|------------|
| 上一轮为什么失败 | 分析上一轮修复的盲点 | 禁止开始新 R |
| 这一轮为什么能成功 | 说明修复策略有何不同 | 禁止开始新 R |
| 数据流断点已定位 | 展示数据流追踪结果 | 禁止开始修复 |
| 完整链路已验证 | 端到端验证通过 | 禁止报告完成 |

### R 迭代升级规则

| R 迭代次数 | 必须升级的策略 |
|-----------|-------------|
| R1 → R2 | 标准升级：执行完整数据流追踪 |
| R2 → R3 | 策略升级：扩大追踪范围 |
| R3 → R4 | 方法升级：换一种追踪方法 |
| R4 → R5 | **强制暂停**：输出完整分析报告，请用户确认修复方向 |
| R5+ | **强制重新分析**：假设之前所有分析都有误，从零开始 |

---

## 资产包路径

```
{docs_internal}/reports/bugs/{bug-id}-R{N}/
├── RCA.md              ← Bugfix 创建
├── TEST_CASE.md        ← Bugfix 创建
└── SCOPE.md            ← Bugfix 创建

后端测试代码: {PROJECT_PATH}/tests/bugs/B{xxx}-{name}/
前端测试代码: {PROJECT_PATH}/web/tests/bugs/B{xxx}-{name}/
```

---

## 📋 输入要求

| 输入项 | 来源 | 必填 |
|--------|------|------|
| 任务池条目 | `.team/task-pool.md` | ✅ |
| Bug 描述 | 用户原始请求 / Issue | ✅ |

---

## 📋 输出要求

| 输出项 | 存放位置 | 格式 |
|--------|----------|------|
| `RCA.md` | `{docs_internal}/reports/bugs/{bug-id}-R{N}/` | Markdown |
| `TEST_CASE.md` | `{docs_internal}/reports/bugs/{bug-id}-R{N}/` | Markdown |
| 修复代码 | `{PROJECT_PATH}/` | 代码 |
| `SCOPE.md` | `{docs_internal}/reports/bugs/{bug-id}-R{N}/` | Markdown |

---

## 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| v5.0 | 2026-04-23 | 精简规则，更新阶段定义 |
| v5.2 | 2026-04-24 | 资产目录加 R 后缀，新增 SCOPE.md |
| v6.0 | 2026-04-30 | 加入 subtype 判定 + 工具链门禁 + 修改前检查 + 前端 Bug 规则 + 前后端双命令质量检查 |
| **v7.0** | **2026-05-04** | **v2 增强：强制数据流追踪协议 + 真实场景验证协议 + 增强完成门禁 + R迭代质量门禁（基于B099六轮失败教训）** |
