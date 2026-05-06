# 框架规则修复总结报告 — 2026-04-30

**时间**: 2026-04-30 09:26 ~ 20:00
**范围**: 框架层 `_team/` 规则文件 + 项目层 `_docs/` 迁移

---

## 一、全局概览

本次会话从**测试文档结构规范**出发，逐步发现 Bug 修复流程和框架规则的系统性断裂，最终完成 5 项 A 层修复。

### 发现→修复链路

```
测试文档结构 → Bug 修复总结报告缺失
    → 编号冲突 → 经验沉淀链路断裂
        → R 后缀语义混乱（Feature/Bug/Change 全带 R）
            → 路径体系两套标准
                → lessons 在框架层（只读）无法写入
                    → A 层 5 项修复
```

---

## 二、A 层修复明细

### A1: Feature/Change 去掉 R 后缀 ✅

**问题**: Feature 和 Change 目录被强制要求带 R 后缀（如 `F001-R1/`），但 R 的语义仅在 Bug 修复轮次中有意义。Feature 的设计迭代用版本体系跟踪，Change 的调整用新任务跟踪。

**修改**:

| 文件 | 变更 |
|------|------|
| `shared.md` | Feature 路径 `{feature-name}-R{N}/` → `{TASK_ID}-{feature-name}/` |
| `shared.md` | Change 路径 `{change-id}-R{N}/` → `{change-id}/` |
| `shared.md` | 新增命名规则汇总表 |
| `dev.md` prompt | 命名规则表 + R 后缀规则 + 资产包路径全部更新 |
| `triage.md` prompt | "所有目录必须带R" → 分类型规则 |

**规则**:
| 任务类型 | 目录格式 | R 后缀 | 调整跟踪 |
|---------|---------|--------|---------|
| Feature | `{TASK_ID}-{name}/` | ❌ | 版本体系 `versions/` |
| Bug | `{bug-id}-R{N}/` | ✅ | R 递增 |
| Change | `{change-id}/` | ❌ | 新 Change 任务 |

---

### A2: Feature 目录命名格式 `{TASK_ID}-{name}` ✅

**问题**: Feature 目录缺少 TASK_ID 前缀，实际项目中有 18/24 个 requirements 子目录缺前缀。

**修改**:

| 文件 | 变更 |
|------|------|
| `shared.md` | Feature 资产包新增"目录命名规则（强制）"段落 |
| `dev.md` prompt | 示例统一为 `F014-unified-pagination/` |

**补充规则**: 已有目录缺前缀的，下次操作该 Feature 时补齐。

---

### A3: Bug R 递增强制触发机制 ✅

**问题**: Bug 修复验证未通过时，AI 在 R1 目录内反复覆盖修改，因为无检查点强制开 R2。

**修改**:

| 文件 | 变更 |
|------|------|
| `shared.md` | R 迭代流程增加"⛔ 强制触发 R 递增"分支 + 触发规则表 |
| `shared.md` | task-pool 阶段列格式化为 `Phase {N} (R{M})` |
| `bugfix-standards.md` | 新增"八、⛔ R 递增强制触发规则"整节 |
| `bugfix-standards.md` | 后续章节重新编号（八→九→...→十六） |
| `task-pool-template.md` | 关联文档列和阶段列定义更新 |

**强制触发条件**:

| 触发条件 | 判断者 | 必须动作 |
|---------|-------|---------|
| Phase 3 验证"未通过" | 执行角色 | 创建 R{N+1} + 阶段重置 + 更新 task-pool |
| 用户反馈"还没修好" | 执行角色 | 同上 |
| AI 自测发现修复方向错误 | 执行角色 | 同上 |

**禁止**: 在当前 R 内覆盖修改、跳过 RCA 直接重试、不更新 task-pool。

---

### A4: bugfix 路径统一到 `reports/bugs/` ✅

**问题**: bugfix-standards.md 使用 `{docs_internal}/bugs/BUG-{id}-SUMMARY.md`，shared.md 使用 `{docs_internal}/reports/bugs/{bug-id}-R{N}/`，两套路径体系。

**修改**:

| 文件 | 变更 |
|------|------|
| `bugfix-standards.md` | 输出成果清单路径统一为 `{docs_internal}/reports/bugs/{bug-id}-R{N}/` |

**统一后的 Bug 产出物路径**:
```
{docs_internal}/reports/bugs/{bug-id}-R{N}/
├── RCA.md       ← 根因分析
├── TEST_CASE.md ← 复现验证
├── SUMMARY.md   ← 修复总结报告
└── SCOPE.md     ← 变更报告
```

---

### A5: lessons 迁移到项目层 `{docs_internal}/lessons/` ✅

**问题**: `_team/lessons/` 在框架层（READ-ONLY），AI 无法写入，经验沉淀链路断裂。

**修改**:

| 文件 | 变更 |
|------|------|
| `shared.md` | docs_internal 目录结构新增 `lessons/`，标注"AI writes here, NOT _team/lessons/" |
| `dev.md` prompt | lessons 加载路径 `{TEAM_PATH}/lessons/` → `{DOCS_INTERNAL}/lessons/` |
| `bugfix-standards.md` | 经验沉淀引用"Lesson Analyst" → "Review 角色" |
| `dev-common.md` | 来源描述更新 + 新增路径说明 |

**文件迁移**: `_team/lessons/dev-common.md` → `_docs/orig-cms/lessons/dev-common.md`

---

## 三、版本记录

| 文件 | 版本变更 |
|------|---------|
| shared.md | v7.1 → **v7.2** |

---

## 四、本次会话早期修复（A 层之前）

### Bug 修复总结报告缺失

| # | 修复项 | 文件 |
|---|--------|------|
| 1 | 第七节标题标注为"Bug 修复总结报告" | bugfix-standards.md |
| 2 | 关闭检查增加"总结报告已提供" | bugfix-standards.md |
| 3 | 输出成果清单增加"Bug 修复总结报告"行 | bugfix-standards.md |
| 4 | 编号冲突修复（多个重复"八"→连续中文编号） | bugfix-standards.md |

### 经验沉淀合并到总结报告模板

| # | 修复项 | 文件 |
|---|--------|------|
| 1 | 经验沉淀从独立第十一节合并到第七节（总结报告模板内） | bugfix-standards.md |
| 2 | LESSON-NEEDED/WATCH/NONE 标签体系 | bugfix-standards.md |
| 3 | 旧第十一节删除，编号重排 | bugfix-standards.md |

---

## 五、B 层问题（已识别，待修复）

| # | 问题 | 状态 |
|---|------|------|
| B1 | Review 角色无触发机制（PM/Tech Lead 里程碑审查未定义） | 待修复 |
| B2 | Lesson Analyst 角色文件不存在（建议合并到 review-standards.md） | 待决策 |
| B3 | LESSON-NEEDED 标签无消费者（无流程扫描这些标签） | 待修复 |
| B4 | Error Report 流程在 development-standards.md 中无起点 | 待修复 |
| B5 | 关闭检查未强制"经验规则已更新到 conventions/lessons" | 待修复 |

**待决策**: Lesson Analyst 独立角色 vs 合并到 review-standards.md？

---

## 六、C 层问题（数据修复，低优先级）

| # | 问题 | 状态 |
|---|------|------|
| C1 | 16 个 requirements 子目录缺 TASK_ID 前缀 | 待修复 |
| C2 | C011-AC.md、C011-SPEC.md 散落根目录（应放 `reports/changes/C011/`） | 待修复 |
| C3 | C012/C013 在错误目录 | 待修复 |
| C4 | `my-profile-page-R1` 错误带 R 后缀（Feature 不应带） | 待修复 |

---

## 七、待决策项

| # | 问题 | 选项 |
|---|------|------|
| D1 | Lesson Analyst 角色定位 | a) 独立角色 b) 合并到 review-standards.md |
| D2 | Bug 修复中发现设计缺陷的处理 | a) 临时修复 + Change 任务 b) 升级 Bug 为 Change |
| D3 | 历史目录迁移策略 | a) 批量迁移 b) 按需补齐 |

---

## 修改文件清单

| 文件 | 修改次数 | 关键变更 |
|------|---------|---------|
| `shared.md` | 6 | 资产包路径 + R 规则 + ID 命名 + lessons 路径 + 版本号 |
| `bugfix-standards.md` | 5 | 路径统一 + R 递增规则 + 编号重排 + 经验沉淀合并 |
| `dev.md` prompt | 3 | 命名规则 + R 规则 + lessons 路径 |
| `triage.md` prompt | 1 | 资产目录命名规则 |
| `task-pool-template.md` | 1 | 关联文档列 + 阶段列定义 |
| `_docs/orig-cms/lessons/dev-common.md` | 新建 | 从 _team/lessons/ 迁移 |
