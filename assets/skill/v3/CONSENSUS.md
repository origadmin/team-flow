# team-flow v3 共识文件

> 本文件记录所有已确认的架构决策和设计共识。新会话必须先读取此文件，禁止重复确认已决策事项。
> 最后更新：2026-05-21

---

## 一、架构定位

### 1.1 v3 与 v2 的关系
- v3 和 v2 **完全分离**，绝不交叉污染
- v2 是独立系统（team/v2/），v3 是独立新系统（v3/ + team/v3/）
- v3 的目标是替代 v2，exec SKILL v3.2 已覆盖 v2 全部核心能力
- v2 的 team.md 不是 v3 的输入，是**成果标杆**——展示别人能做到的水平
- 安装 v3 时 v2 技能自动保留在 `v2/` 子目录，可通过 `flow migrate rollback` 回退

### 1.2 v3 核心定位
- v3 是**限制和控制 AI 流程的工具**
- AI 创建流程规则 → 规则可被 team-flow 跟踪执行 → Editor 可查看微调流程 → AI 加载流程完成任务
- 适用场景：小说编写、程序开发、游戏开发、内容分发、宣传片创作等任何团队协作场景

### 1.3 当前状态
- team-flow 项目已切换到 v3（`.team/version` = v3）
- 默认流程：`skill-dev-flow`（SKILL开发全流程）
- 引擎模式：flow-engine（`flow proc run` 驱动执行）

---

## 二、流程定义规范

### 2.1 规则信息架构（方案3：引用+精炼内容）
- `instruction`：精炼中文指令（~30字/规则），AI 直接看到
- `rule_ref`：引用路径，格式 `flow proc rule {id}`
- `description`：完整英文描述（~200字/规则），AI 需要详情时执行 `flow proc rule {id}` 获取
- **三层架构**：instruction（精炼）→ 引用路径（按需）→ description（完整）

### 2.2 规则 ID
- 随机短码（如 d5f, r3k7, c4p1），**不是**语义化名称（如 output-guard）
- name 才是有意义的项

### 2.3 规则类型
- behavioral_constraint, process_constraint, quality_constraint, output_constraint
- enforcement: hard（必须通过）/ soft（尽力而为）

### 2.4 Gate 条件类型
- tests_pass, lint_pass, deliverables_complete, no_regressions, task_exists, type_matches, custom

### 2.5 流程节点类型
- phase, start, gate, branch, parallel, subflow, loop, manual, event, terminal

### 2.6 `--flow` 参数
- 不是必须的，应从项目配置自动获取（AC1: 一个项目=一个流程）

---

## 三、角色定义规范

### 3.1 人设命名（C1 约定）
每个角色必须有三项标识：

| 字段 | 格式 | 示例 | 用途 |
|------|------|------|------|
| `alias` | 中文2-3字名 | 齐活林, 匠思远, 铸灵手 | 展示名，记忆标识 |
| `alias_en` | 英文人名 | Qi, Craft, Forge | 跨语言引用 |
| `name` | 领域自然语言 | 交付总监, 技能架构师, 技能工匠 | 功能标识 |

persona 格式：`"你是{alias}({alias_en})，{name}。..."`

命名规则：
- 中文 alias 2-3字，体现角色本质（不是泛泛职位）
- 英文 alias 是有性格的人名，不是直译（点石=Spark 不是 Stone）
- name 用领域内自然称呼，不用系统术语（交付总监 不是 Triage Agent）

### 3.2 入口角色（C2 约定）
- 每个流程有且仅有一个 `"principal": true` 角色
- principal 是团队的唯一用户接口（AC2 约束）
- principal 名字随领域走（主理人/Triage/总监），不强制统一
- 引擎显示 ⭐ 标记 principal

### 3.3 角色必须字段
- id, name, alias, alias_en, persona, traits, guidance
- principal（仅入口角色设 true）
- description, capabilities 可选

### 3.4 人设质量标准
- 每个角色必须有鲜明的性格和态度，不能退化为泛泛职位描述
- 没有参考时也必须创造同等水平的人设（不能因为没 team.md 对应就退化）
- 人设要体现"最恨什么""强迫症""绝不做什么"等鲜明态度

---

## 四、执行架构约束

### 4.1 AC1: 一个项目 = 一个流程
- 项目绑定唯一流程，存储在 project.md 的 `default_flow` 字段
- `--flow` 参数不需要，引擎自动读取绑定
- 切换流程是项目级决策
- 没有流程？按 First-Time Setup 引导选择或创建

### 4.2 AC2: 主理人是唯一用户接口
- 只有 principal 角色和用户交流
- 其他角色只执行任务，不直接和用户对话
- 用户中断后重新输入 → 自动路由到 principal
- 子角色完成 → 报告给 principal → principal 决定下一步
- 子角色出错 → 报告给 principal → principal 决定重试/重分配/升级

路由规则：
| 事件 | 路由到 |
|------|--------|
| 用户发新输入 | principal 节点 |
| 用户中断执行 | principal 节点 |
| 用户提问 | principal |
| 子角色完成任务 | principal (via task update) |
| 用户反馈 | principal |
| 子角色执行出错 | principal |

### 4.3 AC3: 验证链架构

**核心原则**：AI 不能自检——角色完成工作后说"没问题"不可信。

**验证链模型**：每个角色做两件事——执行工作 + 验证上游交付物。

```
角色A(执行+自检) → 角色B(验证A+执行+自检) → ... → QA(独立终检) → 完成
```

**Gate 节点规则**：
1. Gate 不做路由——路由是主理人的动作，不是 Gate 的职责
2. Gate 只做质量检查——验证前置条件是否满足
3. 分流节点（如 ent2）应合并到主理人节点——主理人直接输出条件分支
4. 每个流程只需要一个 QA 终检节点——轻量级信任链 + 重量级独立验证

**验证失败处理**：
- 任何角色验证上游失败 → 报告给主理人 → 主理人决定下一步
- 角色之间不直接通信，一切通过主理人
- 失败边统一回到 principal 节点（不是回到上游节点）

**循环保护**：
- 同一任务回到同一节点 ≤ 3次
- 第2次：主理人调整策略
- 第3次：必须STOP，向用户报告

**产出物报告**：
- 角色完成工作后产出：交付物 + 验证报告
- 验证报告包含：做了什么、结果如何、证据
- 下游角色拿到验证报告，验证报告可信度，不重新执行动作
- 自动检查（go build, flow proc validate）的结果写入报告，不可伪造

---

## 五、Create 流程规范

### 5.1 Step 0: 需求发现（强制，在任何设计之前）
- 0.1 解析用户请求（Domain/Scope/起点/痛点/目标/规模）
- 0.2 识别缺口 + 提议澄清（**Propose, don't ask** — 给选项+理由，不是空白表单）
- 0.3 提议团队结构草案（写 JSON 之前先展示）
- 0.4 确认关卡（用户确认后才能进入 Step 1）
- 0.5 迭代补充（可随时重入，添加/删除/调整角色）

### 5.2 疏导原则
1. 提议而非询问 — 给用户方案去反应
2. 解释理由 — "拆两个角色因为运营域差异大"
3. 揭示权衡 — "不需要闭环可以省掉两个角色"
4. 尊重用户领域 — 用户比 AI 更懂业务
5. 一轮搞定 — 目标1-2轮确认

### 5.3 流程步骤细化（C3 约定）
- 用户说的每个步骤映射为独立节点，不合并
- 用户指定的工具/技术（如 HyperFrames）记入描述
- 多个用户步骤可能映射同一角色但不同阶段

### 5.4 领域标签（C4 约定）
- metadata.tags 用 kebab-case
- 包含团队类型和产出类型

### 5.5 团队放置规则
| 来源 | 放置路径 | 说明 |
|------|----------|------|
| 框架预设 | `v3/flows/` | 随工具安装，v3 优先查找 |
| 项目自定义 | `.team/flows/` | 用户创建的放这里 |
| 解析优先级 | `v3/flows/` > `.team/flows/` | v3 是当前版本，优先加载 |

### 5.6 团队注册
- project.md 的 `## Flows` 部分记录所有可用团队
- `flow proc list` 显示注册状态（REG 列）
- `default_flow` 标记默认团队（⭐）

---

## 六、规则修复原则（工具先行）

> **核心原则**：工具创建 → 具体流程规则，一切的开始都在于工具。

当需要补充或修复某条具体规则时，必须按以下顺序思考：

1. **工具层面**：这条规则的执行/验证需要什么工具支持？
   - 如果工具不支持 → 先完善工具，再添加规则
   - 如果工具已支持 → 检查工具输出是否足够 AI 判断规则是否满足
   - 例：规则"flow proc validate 必须零错误"→ 工具 `flow proc validate` 已存在且输出明确

2. **数据层面**：工具输出是否包含足够的信息让 AI 遵守规则？
   - `flow proc validate` 输出 errors/warnings → AI 可以判断是否零错误 ✅
   - 如果工具输出模糊 → 先改进工具输出，再依赖规则

3. **规则层面**：规则的 instruction + description 是否足够清晰？
   - instruction：AI 每次看到的简短指令（~30字中文）
   - description：AI 用 `flow proc rule {id}` 获取的详细说明
   - 两者缺一不可：只有 instruction 太简略，只有 description 每次加载太贵

4. **验证层面**：如何验证 AI 确实遵守了这条规则？
   - 有工具输出作为证据（如 validate 结果、test 输出）→ 可验证
   - 纯行为约束（如"不要自己写代码"）→ 需要下游角色验证

**反模式**：
- ❌ 添加规则但工具不支持 → AI 无法执行，规则形同虚设
- ❌ 规则只写 instruction 不写 description → AI 不知道具体怎么做
- ❌ 规则依赖 AI 自律（如"不要做X"）而没有验证机制 → AI 可能违反

---

## 七、初次使用与迁移

### 7.1 三种初次使用场景

| 场景 | 入口 | 说明 |
|------|------|------|
| 全新项目 | `flow init --v3 [--flow {name}]` | 交互式选择预设团队或创建新团队 |
| AI 会话中发现无流程 | AI 自动检测 | `flow proc list` → 引导选择/创建 |
| v2 升级 | `flow migrate v3 [--flow {name}]` | 7步自动迁移 |

### 7.2 版本回退
- `flow migrate rollback` — 从 v3 回退到 v2（6步自动回退）
- 前提：`.team/project.md.v2.bak` 存在 + v2 回退技能目录存在
- 回退操作：恢复 project.md → 更新 version → 恢复 v2 SKILL.md → 更新 bridge 文件
- **v2 回退目录策略**：v2 完整技能保留在 `v2/` 子目录，不被 v3 引用，不消耗 token
- 根级 v2 遗留目录（prompts/, workflows/, templates/ 等）标记为 v2 legacy，待清理

### 7.3 安装时版本保留
- 安装 v3 时，自动把 v2 完整技能复制到 `v2/` 子目录
- 安装 v2 时，自动把 v1 完整技能复制到 `v1/` 子目录
- 当前版本文件放根目录（引用路径不用改），旧版本安静躺在子目录

### 6.4 IDE 检测
- `detectIDEs()` 从项目路径向上搜索父目录，直到找到 `.trae/`、`.claude/` 等
- 解决了项目路径下没有 IDE 配置目录但父级有的问题

---

## 八、路径变量体系

### 8.1 v3 路径变量

| 变量 | 来源 | 用途 |
|------|------|------|
| `{TEAM_PATH}` | CollectVars() | .team/ 目录 |
| `{DOCS_INTERNAL}` | project.md `docs_internal` | 内部文档路径（不对外暴露） |
| `{DOCS_EXTERNAL}` | project.md `docs_external` | 外部文档路径（开源可读取） |
| `{workspace}` | CollectVars() | 项目根 |
| `{SKILL_PATH}` | CollectVars() | .trae/skills/team-flow/ |
| `{task_id}` | --task 参数 | 任务 ID |
| `{domain}` | flow config | 流程域 |
| `{flow_id}` | flow metadata | 流程名 |
| `{node_id}` | 当前节点 | 节点 ID |

### 7.2 安全边界
- `{DOCS_INTERNAL}` 和 `{DOCS_EXTERNAL}` 是不同的安全域
- 内部文档不允许流出给外部，外部文档是开源可读取的
- 所有 flow JSON 中使用 `{DOCS_INTERNAL}` 而不是 `{docs_path}`

---

## 九、待解决问题

（无）

---

## 十、已创建的流程

| 流程 | 文件 | 角色数 | 规则数 | 节点数 | 状态 |
|------|------|--------|--------|--------|------|
| dev-flow | v3/flows/dev-flow.json | 6 | 14 | 17 | ✅ 已验证 |
| novel-flow | v3/flows/novel-flow.json | 2 | 12 | 9 | ✅ 已验证 |
| content-distribution-flow | v3/flows/content-distribution-flow.json | 5 | 12 | 9 | ✅ 已验证 |
| promo-video-flow | v3/flows/promo-video-flow.json | 6 | 12 | 9 | ✅ 已验证 |
| trading-flow | v3/flows/trading-flow.json | 13 | 13 | 10 | ✅ 已验证 |
| game-design-flow | v3/flows/game-design-flow.json | 4 | 11 | - | ✅ 已验证 |
| skill-dev-flow | v3/flows/skill-dev-flow.json | 6 | 15 | 9 | ✅ 已验证（验证链架构） |
| feature-flow | v3/flows/feature-flow.json | - | - | - | ✅ 已验证 |
| bugfix-flow | v3/flows/bugfix-flow.json | - | - | - | ✅ 已验证 |
| hotfix-flow | v3/flows/hotfix-flow.json | - | - | - | ✅ 已验证 |
| analysis-flow | v3/flows/analysis-flow.json | - | - | - | ✅ 已验证 |
| change-flow | v3/flows/change-flow.json | - | - | - | ✅ 已验证 |
| release-flow | v3/flows/release-flow.json | - | - | - | ✅ 已验证 |
| batch-flow | v3/flows/batch-flow.json | - | - | - | ✅ 已验证 |

备份目录：v3/flows_backup/

---

## 十一、已解决的关键问题

| ID | 问题 | 解决方案 | 日期 |
|----|------|----------|------|
| 核心 | v3 有流程定义没有执行能力 | 重写 exec SKILL v3.2，27个章节覆盖完整执行协议 | 05-20 |
| P0-1 | exec SKILL 缺少节点执行协议 | Node Execution Protocol 6步执行 | 05-20 |
| P0-2 | Principal 分发协议 | Dispatch Protocol + Task 工具分发模板 | 05-20 |
| P0-3 | 会话启动/关闭协议 | Session Startup/Close Protocol | 05-20 |
| P0-4 | 状态行 | `[Role: {alias} \| Flow: {name} \| Node: {id} \| Phase: {phase}]` | 05-20 |
| P1-2 | 工作区边界 | Workspace Boundary 定义 | 05-20 |
| P1-3 | 错误恢复 | Loop Guard: 1次重试 + 2次换思路 + STOP | 05-20 |
| P1-4 | v2→v3 路径变量未替换 | substituteVars() + `//` 清理 + filepath.Clean() | 05-20 |
| P1-1 | 产出物模板缺失 | DocSpec 增加 template + content_rules；8 个 doc 产出物填充；创建 spec/ac 模板文件 | 05-21 |
| P1-5 | flow config paths 缺失 | 新增 `flow config paths` 和 `--json` 命令，输出 9 个路径变量 | 05-21 |
| P1-6 | 初次使用引导 | 三种场景：init --v3 / AI 检测 / migrate v3 | 05-21 |
| P1-7 | flow JSON 扁平格式解析丢失 | normalizeNode() 归一化扁平→结构化 | 05-21 |
| P1-8 | 安装时旧版本被覆盖 | 当前版本放根 + 旧版本放子目录 + rollback 命令 | 05-21 |
| P2-1 | Gate 做路由导致冗余节点 | 删除 ent2，分流合并到 tri1（主理人直接分流） | 05-21 |
| P2-2 | AI 不能自检，验证链缺失 | 验证链架构：v1/v2/v3 规则 + VERIFY_REPORT + verified/!verified 边 | 05-21 |
| P2-3 | 验证失败无标准处理 | 失败→回到主理人→主理人决定下一步+循环保护 | 05-21 |

---

## 十二、代码改动记录

### 核心改动（v3 新增/修改）

**internal/flow/**
- `types.go`: RoleDefinition 新增 Alias, AliasEn, Principal, Persona, Traits, Guidance 字段
- `types.go`: RuleDefinition 新增 Instruction, Description 字段
- `types.go`: GateCondition 新增 Deliverables 字段
- `parser.go`: 新增 `flowNodeRaw`, `flowRaw`, `normalizeNode()` — 扁平格式归一化
- `validator.go`: 新增 validateComponents 验证

**internal/proc/**
- `procrun.go`: CurrentNode 新增角色人设字段；resolveRoleInfo() 替代 extractRoleSources()
- `procrun.go`: buildCurrentNode() 填充角色人设 + substituteVars() 替换路径变量
- `procrun.go`: DocOutput 新增 Template 和 ContentRules 字段；extractDocs() 传递模板和规则
- `procrun.go`: 新增 ParallelBranchOutput 结构体；CurrentNode 新增 ParallelBranches/ParallelStrategy/MergeStrategy 字段
- `procrun.go`: buildCurrentNode() 新增 parallel case：提取分支信息+角色名+别名
- `formatter.go`: ROLE 行显示 "alias(alias_en) · name ⭐"；新增 PERSONA/TRAITS/GUIDANCE 行
- `formatter.go`: DOCS 部分新增 template 和 content_rules 渲染
- `formatter.go`: 新增 PARALLEL BRANCHES 渲染：显示 strategy/merge/分支列表+AI行为规则
- `cmd.go`: 新增 `flow proc rule` 命令；`--flow` 参数改为可选；ProcEntry 新增注册状态
- `cmd.go`: `flow proc list` 显示 NAME/SOURCE/REG/DEFAULT/DESCRIPTION 列
- `cmd.go`: resolveProcPath() 优先查找 `v3/flows/`（v3 是当前版本）
- `cmd.go`: 新增 `flow config paths` 和 `--json` 子命令
- `substitutor.go`: SubstituteVarsInPath() 增加 `//` 清理和 filepath.Clean()
- `resolver.go`: `docs_path` → `DOCS_INTERNAL`；新增 `DOCS_EXTERNAL`
- `runtime.go`: `docs_path:` → `docs_internal:`；新增 ResolveDocsExternalPath()

**internal/boot/**
- `init.go`: 新增 `--v3` 和 `--flow` flags
- `init.go`: v3 Step 3 Flow Setup（selectFlow() 交互菜单）
- `init.go`: installSkillFromFS() 支持 v3 + installFallbackVersion()
- `init.go`: detectIDEs() 改为向上搜索父目录

**internal/migrate/**
- `migrate.go`: `flow migrate v3` 子命令（7步自动迁移）
- `migrate.go`: `flow migrate rollback` 子命令（6步自动回退）
- `migrate.go`: installV3Skills() + installMigrateFallback()
- `migrate.go`: updateBridgeFile() 支持版本参数（v3/v2）
- `migrate.go`: detectIDEs() 改为向上搜索父目录

**internal/ver/**
- `ver.go`: 添加 v3 状态显示 + v2→v3 升级提示

**internal/config/**
- `config.go`: 新增 `flow config paths` 命令，表格/JSON 格式输出 9 个路径变量

**模板策略**
- 每个 SKILL 有专属模板，放在 `team/v3/skills/{skill_name}/templates/` 目录下
- **模板是该 SKILL 自身的一部分**（不是任务产出物）
- skill-dev-flow 在 **fa03 节点**：与用户交互确定文档结构
- skill-dev-flow 在 **fd04 节点**：创建该 SKILL 的专属模板（作为 SKILL_TEMPLATES 产出物）
- 不同领域的 SKILL 有不同的模板（比如小说写作是大纲模板，软件开发是 PRD 模板）

**流程数据**
- `v3/flows/skill-dev-flow.json`: 8 个 doc 产出物添加 content_rules
- `v3/flows/skill-dev-flow.json`: fa03 节点的 SPEC 添加与用户交互确定文档结构的规则
- `v3/flows/skill-dev-flow.json`: fd04 节点新增 SKILL_TEMPLATES 作为该 SKILL 的专属模板产出物
- `v3/flows/skill-dev-flow.json`: qua8 节点使用 skill-creator 技能做 SKILL 质量评估（不是 team-flow-review）
- `v3/flows/skill-dev-flow.json`: qua8 的 QA_REPORT 新增 SKILL 质量评估和 SPEC 对齐检查规则
- `v3/flows/skill-dev-flow.json`: 新增 par1 并行节点（type: parallel），fd05/fd06 从串行改为并行
- `v3/flows/skill-dev-flow.json`: 边调整：fd04→par1→[fd05,fd06]→int7（替代原 fd04→fd05→fd06→int7）
- `v3/flows/skill-dev-flow.json`: 节点数 9→10，边数 17→17（删除 3 条旧边，新增 3 条新边）
- `v3/flows/skill-dev-flow.json`: 15 条规则补全 name 字段（0 warnings）
- `v3/flows/skill-dev-flow.json`: tri1 新增 TASK_CLASSIFICATION doc（含循环计数追踪）
- `v3/flows/skill-dev-flow.json`: v2 规则 description 新增 verified 判定标准（5 条明确标准）
- `v3/flows/skill-dev-flow.json`: 新增 integrator 角色（合无间/Merge/集成大师），int7 从 triage 改为 integrator
- `v3/flows/skill-dev-flow.json`: 角色数 6→7，规则数 15，节点数 10

**评估架构**
- 评估不是独立流程节点，是角色执行动作时调用的技能
- 架构：主理人 → 角色 → 动作（含评估技能）→ 报告回主理人
- qua8（严过关）使用 **skill-creator** 技能做 SKILL 质量评估（专门的技能评估工具）
- 注意：不使用 team-flow-review 做评估（那是 v2 的 review 流程，v3 没有独立 review 节点）

### Go 环境
- PATH: D:\golang\windows\go\bin;D:\workspace\gopath\bin;
