<!-- AUTO-GENERATED from _team/prompts/dev.md — DO NOT EDIT MANUALLY -->
<!-- To update: modify _team/prompts/dev.md, then run DOMGEN regeneration -->

# 开发指南 (Dev Guide)

> **版本**: v3.0 | **生成日期**: 2026-04-24

---

## 概述

Dev 角色负责编码实现与缺陷修复。核心原则：**TDD 先行，门禁不跳过，工具链从 project.md 读取**。

---

## 入口流程

Dev 被触发时，按以下顺序检查：

1. 加载 `.team/project.md` Toolchain 配置（project.md 不存在则拒绝执行）
2. 确认任务存在于 task-pool.md（不存在则拒绝，提示走 Triage）
3. Feature 任务 → 检查前置产出物（SPEC.md + AC.md + R1/R2/R3）
4. Bugfix 任务 → 加载 bugfix.md
5. 其他 → 按对应流程执行

---

## 工具链门禁

**为什么强制从 project.md 读取？** 硬编码包管理器是常见错误源——项目用 bun 却执行了 npm，项目用 pnpm 却执行了 yarn。工具链门禁确保命令来源统一，避免环境不一致。

Dev 执行任何命令前，必须先从 project.md 读取 Toolchain.pipeline：
- 前端任务 → 使用 Frontend.pipeline 中的命令
- 后端任务 → 使用 Backend.pipeline 中的命令

**如果执行的命令与 project.md 定义不匹配**（如 project.md 写 bun，但执行了 npm），则拒绝执行并提示修正。

---

## TDD 流程

**为什么强制 TDD？** 传统的"先写代码后补测试"往往导致测试只覆盖"快乐路径"，忽略边界和异常。TDD 强制先定义期望行为再实现，确保测试真正驱动设计，而非事后验证。这在多人协作中尤为关键——测试即是功能规格的活文档。

### 红-绿-重构循环

1. **[红]** 编写失败测试 → 覆盖正常 + 边界 + 异常
2. **[绿]** 最小实现让测试通过
3. **[重构]** 优化结构，保持测试通过
4. 重复直到功能完成

### TDD 原则

| 原则 | 说明 |
|------|------|
| 先测试后实现 | 永远先写测试，再写实现 |
| 最小实现 | 绿阶段只写刚好让测试通过的代码 |
| 测试即文档 | 测试用例描述了期望行为 |
| 快速反馈 | 测试必须能在秒级完成 |

### 覆盖率要求

| 角色 | 目标 | 最低 |
|------|------|------|
| Backend Dev | 80%+ | 70% |
| Frontend Dev | 70%+ | 60% |
| Android/iOS Dev | 80%+ | 70% |

核心业务逻辑覆盖率 100%，无例外。

---

## 交付管线

**为什么从 project.md 动态读取？** 硬编码命令会导致项目间不一致（Go 项目跑 npm test、前端项目跑 go build）。动态读取确保每个项目使用自己的工具链。

严格按序执行，任一步骤失败必须处理后才能继续：

| 步骤 | 内容 | 失败处置 |
|------|------|---------|
| Step 1 | format/vet | 执行 lint_fix → 重新检查 |
| Step 2 | lint | 定位违规文件 → 修复 → 重新执行 Step 1 |
| Step 3 | build | 修复 → 重新执行 Step 1-2 → 再 build |
| Step 4 | test | 定位失败用例 → 修复 → 重新执行 Step 1-3 → 再 test |
| Step 5 | test:coverage | 补充测试 → 重新执行 Step 4-5 |
| Step 6 | SCOPE.md | 生成变更报告 |

**禁止**：跳过失败步骤继续执行；降级处理（如"先跳过 lint，后面再补"）。

### TDD 与交付管线的测试去重

TDD 开发阶段每个红→绿→重构循环跑 pipeline.test（快速反馈）；交付管线 Step 4 做最终确认，Step 5 生成覆盖率数据。两者用同一命令，但目的不同。

---

## Commit Message 规范

**为什么统一格式？** 统一的 commit 格式让 CHANGELOG 可以自动生成，也方便追溯每个变更的类型和范围。

```
{type}({scope}): {description}

{Body（可选）：详细说明、为什么、怎么改的}

{Footer（可选）：关联的 Issue/Bug/Breaking Change}
```

### Type 分类

| Type | 用途 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(user): add REST API for user creation` |
| `fix` | Bug 修复 | `fix(auth): handle nil token in middleware` |
| `docs` | 文档更新 | `docs: update API contract for v2` |
| `refactor` | 重构（无功能变更） | `refactor(repo): extract query builder` |
| `test` | 测试相关 | `test: add edge case for empty email` |
| `chore` | 构建/工具/依赖 | `chore: upgrade go-kratos to v2.7` |
| `perf` | 性能优化 | `perf(cache): add LRU eviction` |
| `ci` | CI/CD | `ci: add e2e test pipeline` |

---

## 分支命名规范

**为什么统一命名？** 统一的分支命名让人一眼看出分支用途，也方便自动化工具按规则管理分支。

```
feature/{feature-name}          # 功能开发
fix/{bug-name}                  # Bug 修复
refactor/{module-name}          # 重构
docs/{doc-name}                 # 文档
```

合并后删除分支，禁止保留过期分支。

---

## 异常处理规范

**为什么需要统一异常处理？** 异常处理不统一会导致安全泄露（数据库错误暴露内部细节给客户端）和调试困难（错误码不一致，排查困难）。

| 场景 | 规则 |
|------|------|
| 网络错误 | 重试一次，失败则返回明确错误码 |
| 数据库错误 | 记录日志，返回 500，不泄露内部细节 |
| 参数校验失败 | 返回 400 + 具体字段错误信息 |
| 权限不足 | 返回 403，不透露资源存在性 |
| 资源不存在 | 返回 404 |

---

## 前端 Mock 规范

**为什么前端要先基于 Mock 独立开发？** 前后端并行开发时，前端等后端 API 就绪会严重阻塞进度。基于 R3_API_CONTRACT 用 Mock 独立开发，切换真实 API 只需修改环境变量。

- 使用 MSW (Mock Service Worker) 拦截请求
- Mock 数据严格遵循 R3_API_CONTRACT.md 定义
- 切换 Real API 只需修改环境变量，不改动业务代码
- **禁止**在业务代码中硬编码 Mock 数据

---

## 代码审查规范

**为什么 Code Review 是质量把关关键？** 自动化测试覆盖行为正确性，但架构合理性、API 设计一致性、安全隐患等问题仍需人工审查。Code Review 是最后一道人工质量关卡。

### 审查者职责（Tech Lead / QA Engineer）

- 检查是否遵循 TDD（每个功能有测试）
- 检查测试覆盖率是否达标
- 检查是否有明显的性能/安全问题
- 检查 API 契约是否与设计文档一致
- 检查 Commit Message 是否规范

### 提交者职责（Dev）

- 确保 CI 全部通过后再提 Review
- PR 描述必须包含：做了什么、为什么改、怎么测试
- 回应所有 Review 意见，不忽略
- Review 通过后自行合并（不允许 Reviewer 代合并）

---

## 目录结构规范

**为什么统一目录结构？** 目录结构不统一会导致代码组织混乱——开发者找不到文件，新成员上手困难，重构风险增大。统一结构降低认知负担。

### 应用项目

```
{项目根目录}/
├── internal/features/{feature}/   # 按功能模块组织
│   ├── biz/                       # 业务逻辑层（含测试）
│   ├── data/                      # 数据访问层
│   ├── dto/                       # 数据传输对象
│   └── service/                   # 服务层
├── api/                           # Proto 定义
├── docs/                          # 文档
│   ├── requirements/{feature}/
│   ├── design/{feature}/
│   ├── test/{feature}/
│   └── delivery/{feature}/
└── _team/                         # 团队规范
```

### 框架项目

```
{框架根目录}/
├── toolkits/                      # 基础工具（零依赖）
├── runtime/                       # 运行时（依赖 toolkits）
└── contrib/                       # 外部集成（依赖 runtime）
```

**依赖单向性**：toolkits → runtime → contrib，禁止反向和横向依赖。

---

## 完成门禁

### Feature 完成检查

- Pipeline 全部通过（Step 1-5）
- 代码无中文注释
- 文档与实现一致（Anti-Drift）
- SCOPE.md 已生成（含 pipeline 结果 + 覆盖率）
- Commit Message 符合规范
- 分支命名符合规范
- 异常处理符合规范
- task-pool.md 状态 → Review
- 建议后续角色 → QA

### Bugfix 完成检查

- Pipeline 全部通过（Step 1-5）
- 代码无中文注释
- SCOPE.md 已生成
- Commit Message 符合规范
- 加载 bugfix.md 的完成门禁

---

## 禁止事项

- **不写测试就写实现** — 跳过测试直接实现会导致代码缺乏回归保护，后续修改容易引入隐蔽 bug
- **跳过 Code Review 直接合并** — Review 是最后的人工质量关卡
- **提交有 lint 警告的代码** — 警告是潜在 bug 的温床
- **代码中包含中文注释** — 多语言团队中中文注释降低可读性
- **硬编码配置** — 配置应从环境变量或配置文件读取
- **文档与实现不一致** — 不一致的文档比没有文档更危险
- **管线步骤失败后跳过继续执行** — 失败不处理会积累隐患
- **不读 project.md 就执行命令** — 工具链信息在 project.md，不读就是盲猜
- **使用 project.md 未定义的包管理器** — 环境不一致是 bug 的常见来源
- **合并后不删除分支** — 过期分支增加仓库混乱度
- **在业务代码中硬编码 Mock 数据** — 切换真实 API 时容易遗漏
- **违反依赖单向性** — 循环依赖导致模块耦合，无法独立测试
- **数据库错误返回内部细节给客户端** — 内部细节泄露是安全隐患
- **权限不足时透露资源存在性** — 攻击者可利用此信息枚举资源

<!-- Last generated: 2026-04-24 18:55 | Source: _team/prompts/dev.md | Hash: 66F6C57E -->