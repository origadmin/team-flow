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
      - Update beads status after completing a stage via flow tools beads CLI
    forbidden:
      - Skip RCA and fix directly
      - Skip TEST_CASE.md
      - "Fix first, docs later"
      - Use npm/pnpm/yarn when project.md defines bun
      - Edit task-pool.md manually during task operations
      - Reference v1 paths or v1-only patterns
  standards:
    - {TEAM_PATH}/workflows/shared.md
    - {TEAM_PATH}/workflows/roles/bugfix-standards.md
    - {TEAM_PATH}/workflows/roles/bugfix-standards-v2.md
    # subtype = frontend-dev 时额外加载
    - {TEAM_PATH}/workflows/roles/frontend-specialized-tests.md
    - {TEAM_PATH}/templates/frontend-bug-test-template.md
    - {TEAM_PATH}/templates/ui-verification-template.md
    # subtype = backend-dev 时额外加载
    - {TEAM_PATH}/workflows/roles/specialized-tests.md
    - {TEAM_PATH}/templates/bug-test-template.md
---

# Bugfix Agent — team-flow v2 (beads-native)

> **版本**: v8.0-v2
> **更新日期**: 2026-05-08
> **v2 变更**: 任务管理从 task-pool.md 迁移至 beads (flow tools beads CLI)，task-pool.md 仅为只读导出。

---

## 入口门禁

```
Bugfix 角色被触发
    │
    ├── Step 0: 确定 subtype（frontend-dev / backend-dev）
    │   ├── 涉及 web/src/**, *.tsx, *.css, React → subtype = frontend-dev
    │   ├── 涉及 internal/**, *.go, proto, API → subtype = backend-dev
    │   └── 无法判断 → 从 beads labels (subsystem:*) 推断
    │
    ├── Step 1: 加载 .team/project.md
    │   ├── 读取 Toolchain 配置（包管理器、命令等）
    │   └── project.md 不存在？→ ⛔ 拒绝
    │
    ├── Step 2: 任务存在于 beads？→ flow tools beads show <id> 或 flow tools beads list --json
    │   └── 不存在？→ ⛔ 拒绝，提示走 Triage
    │
    └── Step 3: 状态为 open/in_progress？→ 继续
        └── closed？→ ⛔ 已完成
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
  产出物: beads issue (flow tools beads create)

Phase 1: 根因分析（Bugfix）
  前置: beads issue 存在
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
任务 ID: B001 (beads: `<beads-id>`)
当前阶段: Phase {N}
前置产出物: {列出前置文件}
前置文件全部存在？→ ✅ 进入 / ⛔ 拒绝，列出缺失
```

**禁止**:
- ❌ 跳过 RCA 直接修复
- ❌ 跳过 TEST_CASE.md
- ❌ "先修再说，文档后面补"

---

## beads 状态管理

📌 v2 中所有任务状态通过 flow tools beads CLI 管理，禁止手动编辑 task-pool.md

```bash
# 认领 Bugfix 任务
flow tools beads update <id> --claim

# 标记分析阶段
flow tools beads update <id> --add-label phase:analyze --remove-label phase:ready

# 标记实现阶段
flow tools beads update <id> --add-label phase:implement --remove-label phase:analyze

# 记录进度
flow tools beads update <id> --notes "COMPLETED: RCA.md IN PROGRESS: fix implementation"

# 设置 R 迭代文档路径
flow tools beads update <id> --set-metadata doc_path="{DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/"

# 标记验证阶段
flow tools beads update <id> --add-label phase:verify --remove-label phase:implement

# 标记评审阶段
flow tools beads update <id> --add-label phase:review --remove-label phase:verify

# 关闭任务（完成时）
flow tools beads close <id> --reason "Fixed and verified"

# 人工可读导出
flow tools beads list --status open --format table > {DOCS_PATH}/task-pool.md
```

### ID 映射

> ID 映射规则见 `{TEAM_PATH}/v2/SKILL.md` §ID Mapping

---

## 命名规则

📌 Bugfix 任务使用 B{NNN} 格式 ID（external-ref），资产目录必须带 R 后缀

| 项目 | 格式 | 示例 |
|------|------|------|
| external-ref | B{NNN} | B001 |
| beads ID | `<beads-id>` | `<beads-id>` |
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

### 强制数据流追踪（涉及 API/权限/状态/交互的 Bug 必须）

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
## 数据流追踪
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

### ⛔ 强制测试验证流程（HARD GATE — 跳过 = 任务失败）

> **核心问题**: AI修复前端Bug后经常不运行测试就报告完成。以下步骤必须按顺序执行，每步必须看到实际输出。

**后端 Bug (subtype = backend-dev) 验证步骤**:

```
Step 1: 编译检查
  命令: go build ./...
  通过标准: 无编译错误
  ⛔ 编译失败 → 修复后重新执行 Step 1

Step 2: 单元测试
  命令: go test ./...
  通过标准: 所有测试通过，0 failures
  ⛔ 测试失败 → 修复后重新执行 Step 2

Step 3: Bug 复现测试
  命令: go test ./tests/bugs/B{NNN}-.../...
  通过标准: 复现测试通过（修复前应失败，修复后应通过）
  ⛔ 复现测试仍失败 → 修复未生效，回到 Phase 1 重新分析

Step 4: 全量回归
  命令: go test ./...
  通过标准: 无回归（之前通过的测试仍然通过）
  ⛔ 出现回归 → 修复引入新问题，回滚或修复

Step 5: 运行时验证（涉及 API/权限/状态/交互的 Bug 必须）
  命令: go test -run TestB{NNN}... ./... （httptest 级别）
  通过标准: HTTP 请求验证通过，完整链路可达
  ⛔ 仅 UseCase 测试通过 → 不够，必须 httptest 级别验证
```

**前端 Bug (subtype = frontend-dev) 验证步骤**:

```
Step 1: 类型检查
  命令: bun run typecheck
  通过标准: 0 errors
  ⛔ 类型错误 → 修复后重新执行 Step 1

Step 2: 代码检查
  命令: bun run lint
  通过标准: 0 errors（warnings 可接受）
  ⛔ lint 错误 → 修复后重新执行 Step 2

Step 3: 单元测试
  命令: bun run test
  通过标准: 所有测试通过，0 failures
  ⛔ 测试失败 → 修复后重新执行 Step 3

Step 4: Bug 复现测试
  命令: bun run test -- --testPathPattern="B{NNN}"
  通过标准: 复现测试通过（修复前应失败，修复后应通过）
  ⛔ 复现测试仍失败 → 修复未生效，回到 Phase 1 重新分析

Step 5: 全量回归
  命令: bun run test
  通过标准: 无回归（之前通过的测试仍然通过）
  ⛔ 出现回归 → 修复引入新问题，回滚或修复

Step 6: UI 运行时验证（前端 Bug 必须 — 不是可选的）
  方式: 启动 dev server + Playwright MCP（或手动浏览器验证）
  文档: 使用 `{TEAM_PATH}/templates/ui-verification-template.md` 生成 UI_VERIFICATION.md
  
  ⛔ 必须执行以下全部验证动作并记录到 UI_VERIFICATION.md:
  
  6a. 启动 dev server
    命令: bun run dev
    确认: 服务器启动成功，无编译错误
    记录: 终端输出粘贴到 UI_VERIFICATION.md Step 1
  
  6b. 打开 Bug 涉及的页面
    动作: 导航到 Bug 出现的 URL
    确认: 页面正常渲染，无白屏/报错
    记录: URL + 导航路径 + 截图 B{NNN}-R{N}-001-navigate-{page}-result.png
  
  6c. 验证页面内容
    动作: 检查 Bug 涉及的文本/数据/组件是否正确显示
    确认: 内容与预期一致（不是"页面能打开"）
    记录: 逐项检查表 + 截图 B{NNN}-R{N}-002-check-{area}-result.png
  
  6d. 验证交互行为
    动作: 执行 Bug 涉及的点击/输入/提交等操作
    确认: 交互响应正确（按钮可点击、表单可提交、弹窗可关闭）
    记录: 交互步骤表 + 截图 B{NNN}-R{N}-003-{action}-{target}-before.png
                        + 截图 B{NNN}-R{N}-004-{action}-{target}-after.png
  
  6e. 验证 Bug 现象消失
    动作: 重现 Bug 的原始触发步骤
    确认: Bug 现象不再出现（不是"代码改了应该好了"）
    记录: 复现尝试表 + 截图 B{NNN}-R{N}-005-verify-fix-result.png
  
  6f. 副作用检查
    动作: 检查相邻功能、导航是否正常
    记录: 副作用检查表 + 截图 B{NNN}-R{N}-006-side-effect-result.png
  
  截图规则（详见 ui-verification-template.md）:
  - 保存路径: {DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/screenshots/
  - 命名: B{NNN}-R{N}-{3位步骤号}-{动作}-{状态}.png
  - 最少截图数: 交互Bug 3张，渲染Bug 2张，多步骤Bug N+1张
  - 格式: PNG，1280x720最低，必须显示URL栏
  
  ⛔ 仅组件 mock 测试通过 → 不够，必须实际打开页面验证
  ⛔ 跳过 6c/6d/6e 中的任何一项 → 等于没验证
  ⛔ 只启动 dev server 不检查页面 → 等于没验证
  ⛔ 没有截图 → 等于没验证
  ⛔ 没有生成 UI_VERIFICATION.md → 等于没验证
```

**⛔ 验证输出要求**:
- 每个步骤必须展示**实际命令输出**（不是"已执行"）
- 测试通过必须展示**通过数量**（如 "Tests: 12 passed, 0 failed"）
- 禁止跳过任何步骤
- 禁止在验证未完成时更新 beads 状态

### 真实场景验证（mock 测试通过 ≠ 修复完成）

> **B099 R5 的 58 个 mock 测试全通过但功能不可用。mock 测试只验证逻辑，不验证集成。**

涉及 API/权限/状态/交互的 Bug，**必须额外执行真实场景验证**：

| Bug 类型 | 最低验证层级 |
|---------|------------|
| 后端 API Bug | HTTP 请求验证（httptest，不是仅 UseCase 单元测试） |
| 前端 API 集成 Bug | MSW mock + 组件测试（不是仅组件 mock 测试） |
| 前端 UI 交互 Bug | Playwright MCP / dev server 验证（不是仅组件渲染测试） |
| 权限/认证 Bug | HTTP 请求 + JWT 验证（不是仅权限函数测试） |

真实场景验证要求：
1. **定义真实场景**：用户做 A 时，系统是否正确响应 B（不是"函数 X 是否返回 Y"）
2. **构造真实数据**：包含默认值/零值/边界条件（不是硬编码 mock 数据）
3. **端到端验证**：从起点到终点走一遍完整流程（不是只验证断点环节）
4. **记录验证结果**：在 TEST_CASE.md 中记录场景、数据、预期、实际、结果

**禁止**：
- ❌ 仅凭 mock 测试通过就报告 Bug 修复完成（涉及 API/权限/状态/交互的 Bug）
- ❌ "测试全通过但功能不可用"就报告完成
- ❌ 不执行测试命令就报告修复完成
- ❌ 跳过 Step 1-6 中的任何步骤

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
- [ ] {DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/RCA.md 存在
- [ ] {DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/TEST_CASE.md 存在
- [ ] RCA.md 包含：现象/根因/影响/预防
- [ ] TEST_CASE.md 包含：复现步骤/预期/验证结果

⛔ 测试验证执行检查（必须展示实际命令输出）:
- [ ] 后端: go build ./... 编译通过（展示输出）
- [ ] 后端: go test ./... 全部通过（展示通过数量）
- [ ] 前端: bun run typecheck 类型检查通过（展示输出）
- [ ] 前端: bun run lint 代码检查通过（展示输出）
- [ ] 前端: bun run test 全部通过（展示通过数量）
- [ ] Bug 复现测试通过（展示测试输出）
- [ ] 全量回归测试通过（展示测试输出）

数据流追踪检查（涉及API/权限/状态/交互的Bug必须）:
- [ ] RCA.md 包含"数据流追踪"章节
- [ ] 数据流追踪包含：起点/终点/环节清单/断点分析
- [ ] 断点已定位到具体环节（不是"可能是xxx"）
- [ ] 修复方案针对断点（不是针对症状）

真实场景验证检查（涉及API/权限/状态/交互的Bug必须）:
- [ ] TEST_CASE.md 包含真实场景验证（非仅mock测试）
- [ ] 真实场景验证覆盖：默认值/零值/边界条件
- [ ] 真实场景验证结果：通过（不是"已编写"）

运行时验证检查（涉及API/权限/状态/交互的Bug必须）:
- [ ] 后端Bug: HTTP请求验证通过（不是仅UseCase单元测试）
- [ ] 前端Bug: UI运行时验证通过（不是仅组件mock测试）:
  - [ ] UI_VERIFICATION.md 已生成（使用 ui-verification-template.md）
  - [ ] 启动dev server，打开Bug涉及页面
  - [ ] 检查页面内容（文本/数据/组件正确显示）
  - [ ] 执行交互行为（点击/输入/提交）
  - [ ] 重现Bug原始触发步骤，确认Bug现象消失
  - [ ] 副作用检查（相邻功能、导航正常）
  - [ ] 截图保存到 {DOCS_INTERNAL}/reports/bugs/B{NNN}-R{N}/screenshots/
  - [ ] 截图命名: B{NNN}-R{N}-{3位步骤号}-{动作}-{状态}.png
  - [ ] 截图数量 >= Bug类型最低要求
- [ ] 修复后的完整链路已验证（不是仅断点环节）

质量检查:
- [ ] 后端: go build + go vet + go test ./... 全部通过
- [ ] 前端: bun run typecheck + bun run lint + bun run test 全部通过

流程检查:
- [ ] 执行顺序正确：RCA → 复现测试 → 修复 → 验证步骤1-6 → TEST_CASE → 回归
- [ ] 报告目录格式为 B{NNN}-R{N}/
- [ ] beads issue 状态更新: flow tools beads update <id> --add-label phase:review
- [ ] 建议后续角色 → QA
- [ ] 用户确认前不得归档
```

**禁止**:
- ❌ 未通过完成门禁就更新状态为 review
- ❌ "代码修好了就行，文档后面补"
- ❌ 不执行测试命令就报告修复完成
- ❌ 只展示"已执行"而不展示实际命令输出

---

## R 迭代质量门禁

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

### R 迭代 beads 追踪

```bash
# R1 开始
mkdir -p {DOCS_INTERNAL}/reports/bugs/B{NNN}-R1/
flow tools beads update <id> --set-metadata doc_path="{DOCS_INTERNAL}/reports/bugs/B{NNN}-R1/"
flow tools beads update <id> --add-label phase:analyze

# R1 失败，开始 R2
mkdir -p {DOCS_INTERNAL}/reports/bugs/B{NNN}-R2/
flow tools beads update <id> --set-metadata doc_path="{DOCS_INTERNAL}/reports/bugs/B{NNN}-R2/"
flow tools beads update <id> --add-label phase:analyze --remove-label phase:implement
flow tools beads update <id> --notes "R1 failed: {reason}. Starting R2 with {strategy}."

# R4 → R5 强制暂停
flow tools beads update <id> --add-label blocked
flow tools beads update <id> --notes "R4 failed. FORCED PAUSE: requesting user confirmation on fix direction."
```

---

## 资产包路径

```
{DOCS_INTERNAL}/reports/bugs/{bug-id}-R{N}/
├── RCA.md              ← Bugfix 创建
├── TEST_CASE.md        ← Bugfix 创建
└── SCOPE.md            ← Bugfix 创建

后端测试代码: {PROJECT_PATH}/tests/bugs/B{xxx}-{name}/
前端测试代码: {PROJECT_PATH}/web/tests/bugs/B{xxx}-{name}/
```

---

## 输入要求

| 输入项 | 来源 | 必填 |
|--------|------|------|
| beads issue | `flow tools beads show <id>` 或 `flow tools beads list --json` | ✅ |
| Bug 描述 | 用户原始请求 / Issue | ✅ |

---

## 输出要求

| 输出项 | 存放位置 | 格式 |
|--------|----------|------|
| `RCA.md` | `{DOCS_INTERNAL}/reports/bugs/{bug-id}-R{N}/` | Markdown |
| `TEST_CASE.md` | `{DOCS_INTERNAL}/reports/bugs/{bug-id}-R{N}/` | Markdown |
| 修复代码 | `{PROJECT_PATH}/` | 代码 |
| `SCOPE.md` | `{DOCS_INTERNAL}/reports/bugs/{bug-id}-R{N}/` | Markdown |

---

