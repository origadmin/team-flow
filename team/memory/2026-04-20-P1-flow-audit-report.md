# _team 框架审核报告 — 第二轮 + 深度流程审核

**审核时间**: 2026-04-20
**审核范围**: SKILL.md、shared.md、workflows/ 全部文件、prompts/ 全部文件
**修复轮次**: 第一轮（P0-P2）+ 第二轮（结构矛盾 S1-S9）

---

## 一、修复清单

### 第一轮修复（路径 + P0 回归）

| # | 文件 | 问题 | 修复方式 |
|---|------|------|---------|
| P0-1 | prompts 9个文件 | `{TEAM_PATH}/{TEAM_PATH}/` 双路径（PowerShell替换回归bug） | 批量 `\{TEAM_PATH\}/\{TEAM_PATH\}/ → \{TEAM_PATH\}/` |
| P1-2 | `prompts/dev.md:296` | `_team/task-pool.md` 路径错误 | 改为 `.team/task-pool.md` |
| P1-3 | `QUICKSTART.md` / `CONFIG.md` | `Project.md` 大小写不一致 | 统一 `Project.md` |
| P2-6 | `team-input-matcher.md` | 缺少 Triage Step 0 | 新增 Step 0: Triage 检查 |
| P2-7 | `output-standards.md` | 角色表缺 4 个角色（triage/devops/architect/review） | 补全 10 个角色 |
| P2-8 | `development-standards.md` | 自测表标题「自测要求」应为「自测流程」 | 标题修正 |
| P2-9 | `shared.md` | Phase 3 TDD「绿」阶段描述缺失 | 补完红→绿→重构正文（后核实为误报，实际内容在dev-workflow.md） |
| P2-10 | `development-standards.md` | 0.4 节复制 shared.md 路径规则 | 改为引用 shared.md |
| P1-5 | `workflows/roles/*.md` 7个文件 | 头部「加载文件」用相对路径 | 改为 `{TEAM_PATH}/workflows/roles/xxx` |

### 第二轮修复（核心流程矛盾 S1-S9）

| # | 文件 | 问题 | 修复方式 |
|---|------|------|---------|
| **S1** | `shared.md` | 初始化步骤 5 `{role}` 未定义（初始化时角色未知） | 改为「跳过！角色加载在输入匹配后执行」 |
| **S2** | `shared.md` | 初始化含步骤 5-6（角色加载不属于初始化阶段） | 明确标注跳过，指向「输入匹配规则」 |
| **S3** | `shared.md` | 步骤 0 Triage 触发表混入「高复杂度任务」行（既非触发条件也非触发关键词） | 删除该行 |
| **S4** | `shared.md` | Triage 执行描述「将任务写入 task-pool」与 `prompts/triage.md`「用户确认后才写」矛盾 | 改为明确 3 步流程 + 「禁止跳过用户确认直接写」警告 |
| **S5** | `development-standards.md` | 残留「匹配流程」3 行 + 2 处「智能体」文本 | 删除段落，替换「智能体」为「角色」 |
| **S6** | `development-standards.md` | 0.3 节「智能体匹配」表用词歧义（引用不存在的多智能体系统） | 改为「角色→规范文件映射」表，引用 project.md |
| **S7** | `development-standards.md` | 0.4 节完整复制 shared.md 路径规则（重复风险） | 改为引用 shared.md |
| **S8** | `development-standards.md` | 自测表标题「自测要求（强制）」应为「自测流程（强制）」 | 标题修正 |
| **S9** | `development-standards.md` | PowerShell `Set-Content -NoNewline` 导致文件损坏（0字节）→ git 恢复 | 改用 `git restore` + edit 工具精准修改 |

### 第三轮修复（SKILL.md + shared.md 委托说明）

| # | 文件 | 问题 | 修复方式 |
|---|------|------|---------|
| P1-3 | `SKILL.md` | 步骤 3 delegation 说明「按 shared.md 初始化执行」但只有 4 步，shared.md 有 6 步，AI 不知具体执行哪些 | 改为明确 4 步（步骤 1-4），角色加载移到步骤 4 |
| P1-4 | `SKILL.md` | 步骤 4 delegation 不含 prompts 加载（只有规范文件） | 补全 6 子步骤（步骤 0-5：Triage检查→意图分类→关键词匹配→兜底处理→加载prompts→加载规范） |

---

## 二、最终正确执行路径（无歧义版）

```
用户输入
    ↓
[SKILL.md 步骤 4] → shared.md「输入匹配规则·步骤 0 Triage 检查」
    ↓
├─ 触发 Triage
│   ├─ 读取 prompts/triage.md + workflows/roles/triage-standards.md
│   ├─ triage-standards.md 按 shared.md 步骤 0 触发条件执行
│   ├─ Step 1: 分析输入 → Step 2: 输出分类报告（3 选项请用户确认）
│   ├─ Step 3: 用户确认 [A] → 写 .team/task-pool.md
│   └─ 重新匹配剩余任务（回到 Triage 检查）
│
└─ 不触发 Triage → shared.md「步骤 1 意图分类」
                      ↓
                    shared.md「步骤 2 关键词匹配」→ 加载
                    `{TEAM_PATH}/prompts/{role}.md` +
                    `{TEAM_PATH}/workflows/roles/{role}.md`
                      ↓
                    按角色规范执行任务
                      ↓
                    更新 .team/task-pool.md（进行中 → 已完成）
```

---

## 三、已消除的核心矛盾

| 矛盾 | 来源 | 解决方案 |
|------|------|---------|
| 「直接写 task-pool」 vs 「用户确认后才写」 | shared.md vs prompts/triage.md | 统一为：用户确认后才写。三方（shared.md / triage-standards.md / prompts/triage.md）完全一致 |
| `{role}` 未定义 | shared.md 初始化步骤 5 | 初始化阶段只做环境设置（.team/ + task-pool + version），角色加载在输入匹配后 |
| 「高复杂度任务」触发 Triage | shared.md 步骤 0 表 | 删除该行，复杂度评估在 triage-standards.md 独立处理（不影响 Triage 触发判断） |
| 「workflows/xxx」相对路径 | 上一轮 PowerShell 替换回归 | 全部改为 `{TEAM_PATH}/workflows/...` |
| 8 个角色规范文件未在 shared.md 列出 | shared.md 角色表只有 5 个角色 | 补全 10 个角色（含 prompts 路径） |
| 「加载 shared.md 后按自己流程」 vs 「委托 shared.md」 | triage-standards.md 含完整 flow | 改为：引用 shared.md 触发条件 + 执行本文件详细流程（避免维护两套触发条件） |
| 多个 prompts 文件 `{TEAM_PATH}/{TEAM_PATH}/` 双路径 | PowerShell `-replace` 二次叠加 | 批量修单：`{TEAM_PATH}/{TEAM_PATH}/ → {TEAM_PATH}/` |

---

## 四、最终静态检查（12/12 通过）

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | 全框架 `{TEAM_PATH}/{TEAM_PATH}/` 双路径 | PASS ✅ |
| 2 | prompts 目录残留 workflows/ 相对路径 | PASS ✅ |
| 3 | prompts/analysis.md 双路径 | PASS ✅ |
| 4 | SKILL.md 重复「步骤 4」段落 | 1 个 ✅ |
| 5 | shared.md 步骤 5 跳过标记 | OK ✅ |
| 6 | shared.md 步骤 6 跳过标记 | OK ✅ |
| 7 | shared.md Triage「用户确认后」描述 | OK ✅ |
| 8 | development-standards.md「智能体」残留 | 0 处 ✅ |
| 9 | development-standards.md 自测表标题 | 「自测流程」✅ |
| 10 | output-standards.md 角色数 | 10 个 ✅ |
| 11 | team-input-matcher.md Step 0 | OK ✅ |
| 12 | SKILL.md 文件大小 | 5645 bytes ✅ |

---

## 五、已知限制（不影响执行，但值得改进）

| # | 描述 | 影响 |
|---|------|------|
| L1 | `team-config.json` 中仍用相对路径 `workflows/shared.md`（JSON 配置，路径相对于 _team 根目录，AI 不会直接读取此文件） | 低 |
| L2 | `QUICKSTART.md` / `REFACTOR_SUMMARY.md` 中含相对路径（人类可读文档，非 AI 执行文件） | 低 |
| L3 | `shared.md` 初始化步骤 4 说「写入任务池」，但步骤 5-6 跳过。若 AI 看到步骤 4 直接执行，可能跳过 shared.md 的「输入匹配规则」章节。需要配合 SKILL.md 步骤 4 明确加载 shared.md「输入匹配规则」而非整个文件 | 低（SKILL.md 已明确步骤 4 加载输入匹配规则） |
| L4 | `development-standards.md` 自测表（section 三）的命令示例只有 Backend Dev，Frontend Dev 命令在 dev-workflow.md Phase 3 中 | 低（实际命令按 SKILL.md 步骤 5 检测） |

---

## 六、文件路径规范（AI 执行标准）

```
✅ 正确：{TEAM_PATH}/workflows/shared.md
❌ 错误：workflows/shared.md
❌ 错误：{TEAM_PATH}/{TEAM_PATH}/workflows/shared.md（双路径）
❌ 错误：_team/workflows/shared.md（绝对路径硬编码）
```
