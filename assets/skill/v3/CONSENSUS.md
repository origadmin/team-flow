# team-flow v3 核心共识

> 本文件记录所有已确认的架构决策和设计共识。新会话必须先读取此文件，禁止重复确认已决策事项。
> 最后更新：2026-05-27

---

## 一、版本体系

### D1: v1/v2/v3 并行共存

三个版本同步并行，各自有不同的能力和流程，不存在线性升级关系。

| 版本 | 能力范围 | 流程管理 | 任务管理 |
|------|---------|---------|---------|
| v1 | 纯文档管理开发流程 | Markdown 文件驱动 | 无 |
| v2 | 使用 flow task 管理开发流程 | Markdown + flow task | beads (bd CLI) |
| v3 | 使用 flow 全功能，管理各种流程和团队 | JSON 流程定义 + 引擎驱动 | 配置驱动（beads/file/custom） |

- v3 和 v2 **完全分离**，绝不交叉污染
- 安装 v3 时 v2 技能自动保留在 `v2/` 子目录，可通过 `flow migrate rollback` 回退
- v2 的 team.md 是**成果标杆**，不是 v3 的输入

---

## 二、架构设计

### D2: 配置驱动工具路由

后端不使用硬编码 Go 实现，而是使用配置驱动的方式。新工具/技能的增加只需更新配置，不需要修改代码。

- internal/tools/ 实现通用 ToolRouter
- 工具注册通过 tools.yaml 配置文件
- 替代方案（已否决）：硬编码 Backend 接口（BeadsBackend/FileBackend/GitBackend 等）

### D3: Flow 统一入口架构

Flow 的职责是：**统一入口 → 内部处理（可选）→ 转发 → 内部处理（可选）**

- Flow 命令 = 入口 + 路由 + 转发，不是业务逻辑实现者
- 对 AI 屏蔽内部实现细节，所有操作通过 flow 命令完成
- beads 是 flow 内部实现，SKILL 无需意识到 beads 概念

### D4: 统一工具路径

tools.yaml 中定义统一工具路径，AI 通过 `flow config paths` 和 `flow tools list` 发现可用工具。

- tools.yaml 包含 paths 段
- flow 提供 `flow tools list` 和 `flow tools run` 命令
- AI 禁止自行搜索 PATH 或硬编码工具路径

### D5: 单词目录命名

internal/ 下的包名使用单词命名，不使用双单词缩写。

- toolcfg → tools
- configcmd → 合并到 cmd/config.go
- 当前复杂度不需要双单词，单词更简洁清晰

### D6: 两目录归一

只保留 skill/ 和 orgs/ 两个顶层资源目录，删除 teams/ 和 v3/teams/。

| 目录 | 内容 | 说明 |
|------|------|------|
| `assets/skill/` | v1/v2/v3 三版技能文件 | 技能文件，按版本子目录 |
| `assets/orgs/` | 组织模板（team.json + flows/） | 团队+流程绑定 |
| `assets/flows/` | 预设流程定义 | 独立流程模板 |
| `assets/schema/` | JSON Schema | 流程验证 |

### D7: 组织模板绑定团队+流程

流程必须绑定团队，没有团队无法执行。团队+流程 = 组织模板（Organization Template）。

- org.json = 团队+流程的绑定容器
- 选流程 = 选组织
- 没有团队 = 公司没有人 = 无法工作

### D8: 主流程+子流程分层

模板层支持多个主流程，项目层只能绑定一个主流程。子流程不可脱离主流程独立使用。

- org.json 中 main_flows[] + sub_flows[]
- feature-flow、bugfix-flow 是 dev-flow 的子流程
- flow init 只展示主流程

### D9: 工具化编辑

所有内容编辑通过 flow CLI 命令，禁止 AI 手动编辑 JSON/YAML 配置文件。

- 新增 flow org * 命令族
- CLI 内置验证链（角色存在性、技能名称标准等）
- 防止 AI 手动编辑导致的配置错误

### D10: 冷热分层追踪

每步对话必须记录，但使用冷热分层存储省 TOKEN。

- HOT（~300 tokens 在 prompt）：当前状态摘要
- COLD（磁盘按需查询）：详细记录
- SKILL.md 也采用 HOT/COLD 分层（骨架 + references/）

---

## 三、Embed 统一路径管理

### 统一 embed 根

```go
//go:embed all:assets
var FS embed.FS

const (
    SkillRoot  = "assets/skill"
    OrgsRoot   = "assets/orgs"
    FlowsRoot  = "assets/flows"
    SchemaRoot = "assets/schema"
)
```

- 只需一行 embed，新增资源类型只需在 assets/ 下加子目录
- 所有 Go 代码通过 skillfs 常量访问，禁止硬编码路径
- npx 安装通过 package.json files 字段控制，与 Go embed 独立

---

## 四、流程定义规范

### 4.1 规则信息架构（引用+精炼内容）

- `instruction`：精炼中文指令（~30字/规则），AI 直接看到
- `rule_ref`：引用路径，格式 `flow proc rule {id}`
- `description`：完整英文描述（~200字/规则），AI 需要详情时执行 `flow proc rule {id}` 获取

### 4.2 实体 ID 规范

所有实体 ID 统一使用随机短码（`idgen.RandHex(3)` 生成 6 字符十六进制）：

| 实体类型 | ID 格式 | 示例 | 生成方式 |
|----------|---------|------|----------|
| 规则 (Rule) | 随机短码 | `d5f`, `r3k7` | `idgen.RandHex(3)` |
| 节点 (Node) | 随机短码 | `a26c80` | `idgen.RandHex(3)` |
| 边 (Edge) | `e_` + 随机短码 | `e_c2b046` | `"e_" + idgen.RandHex(3)` |
| 角色 (Role) | 随机短码 | `triage`, `qa` | `idgen.RandHex(3)` |
| 流程 (Flow) | 随机短码 | `dev-flow` | `idgen.RandHex(3)` |

**核心原则**：
- **ID 是 key，不是 name**：用户不应关注 key 叫什么名字，就像数据库主键
- **name 才是有意义的项**：用户通过 `name` 字段识别实体
- **创建 → 返回 ID → 按 ID 编辑**：用 `flow proc node add` 创建节点（自动生成 ID），返回的 ID 用于后续 `flow proc node edit --id <id>`
- **禁止直接修改 JSON**：所有实体变更必须通过 `flow` CLI 完成

### 4.3 规则类型

- behavioral_constraint, process_constraint, quality_constraint, output_constraint
- enforcement: hard（必须通过）/ soft（尽力而为）

### 4.4 Gate 条件类型

- tests_pass, lint_pass, deliverables_complete, no_regressions, task_exists, type_matches, custom

### 4.5 流程节点类型

- phase, start, gate, branch, parallel, subflow, loop, manual, event, terminal

### 4.6 `--flow` 参数

- 不是必须的，应从项目配置自动获取（AC1: 一个项目=一个流程）

---

## 五、角色定义规范

### 5.1 人设命名（C1 约定）

| 字段 | 格式 | 示例 | 用途 |
|------|------|------|------|
| `alias` | 中文2-3字名 | 齐活林, 匠思远, 铸灵手 | 展示名，记忆标识 |
| `alias_en` | 英文人名 | Qi, Craft, Forge | 跨语言引用 |
| `name` | 领域自然语言 | 交付总监, 技能架构师, 技能工匠 | 功能标识 |

persona 格式：`"你是{alias}({alias_en})，{name}。..."`

命名规则：
- 中文 alias 2-3字，体现角色本质（不是泛泛职位）
- 英文 alias 是有性格的人名，不是直译
- name 用领域内自然称呼，不用系统术语

### 5.2 入口角色（C2 约定）

- 每个流程有且仅有一个 `"principal": true` 角色
- principal 是团队的唯一用户接口（AC2 约束）
- 引擎显示 ⭐ 标记 principal

### 5.3 角色必须字段

- id, name, alias, alias_en, persona, traits, guidance
- principal（仅入口角色设 true）

### 5.4 人设质量标准

- 每个角色必须有鲜明的性格和态度，不能退化为泛泛职位描述
- 人设要体现"最恨什么""强迫症""绝不做什么"等鲜明态度

---

## 六、执行架构约束

### 6.1 AC1: 一个项目 = 一个流程

- 项目绑定唯一流程，存储在 project.yaml 的 `active_flow` 字段
- 切换流程是项目级决策
- 没有流程？按 First-Time Setup 引导选择或创建

### 6.2 AC2: 主理人是唯一用户接口

- 只有 principal 角色和用户交流
- 其他角色只执行任务，不直接和用户对话
- 子角色完成 → 报告给 principal → principal 决定下一步

| 事件 | 路由到 |
|------|--------|
| 用户发新输入 | principal 节点 |
| 用户中断执行 | principal 节点 |
| 子角色完成任务 | principal (via task update) |
| 子角色执行出错 | principal |

### 6.3 AC3: 验证链架构

**核心原则**：AI 不能自检——角色完成工作后说"没问题"不可信。

```
角色A(执行+自检) → 角色B(验证A+执行+自检) → ... → QA(独立终检) → 完成
```

**Gate 节点规则**：
1. Gate 不做路由——路由是主理人的动作
2. Gate 只做质量检查——验证前置条件是否满足
3. 每个流程只需要一个 QA 终检节点

**验证失败处理**：
- 失败边统一回到 principal 节点
- 同一任务回到同一节点 ≤ 3次
- 第3次：必须 STOP，向用户报告

---

## 七、Status Line 格式

```
[{alias} | {node_id}:{node_name}({flow}) | {ref} | {phase}]
```

| 字段 | 来源 | 说明 |
|------|------|------|
| alias | flow proc run → ROLE.alias | 当前角色 |
| node_id:name(flow) | NODE + NODE_NAME + FLOW | 节点在外，流程在括号里 |
| ref | task_id 或 disc-{YYYYMMDD}-{seq} | 无则显示 `-` |
| phase | on_enter/on_exit 等 | 当前阶段 |

⛔ Status Line 数据来源：`flow proc run` 输出，NEVER hardcode

---

## 八、规则修复原则（工具先行）

1. **工具层面**：这条规则的执行/验证需要什么工具支持？工具不支持 → 先完善工具
2. **数据层面**：工具输出是否包含足够信息让 AI 遵守规则？
3. **规则层面**：instruction + description 是否足够清晰？
4. **验证层面**：如何验证 AI 确实遵守了这条规则？

**反模式**：
- ❌ 添加规则但工具不支持 → 规则形同虚设
- ❌ 规则只写 instruction 不写 description → AI 不知道具体怎么做
- ❌ 规则依赖 AI 自律而没有验证机制 → AI 可能违反

---

## 九、路径变量体系

| 变量 | 来源 | 用途 |
|------|------|------|
| `{TEAM_PATH}` | CollectVars() | .team/ 目录 |
| `{DOCS_INTERNAL}` | project.yaml `docs_internal` | 内部文档路径 |
| `{DOCS_EXTERNAL}` | project.yaml `docs_external` | 外部文档路径 |
| `{workspace}` | CollectVars() | 项目根 |
| `{SKILL_PATH}` | CollectVars() | .trae/skills/team-flow/ |
| `{task_id}` | --task 参数 | 任务 ID |

安全边界：`{DOCS_INTERNAL}` 和 `{DOCS_EXTERNAL}` 是不同的安全域，内部文档不允许流出给外部。

---

## 十、已创建的流程

| 流程 | 文件 | 角色数 | 规则数 | 状态 |
|------|------|--------|--------|------|
| dev-flow | assets/flows/dev-flow.json | 6 | 14 | ✅ |
| novel-flow | assets/orgs/content-team/flows/novel-flow.json | 2 | 12 | ✅ |
| content-distribution-flow | assets/orgs/content-team/flows/content-distribution-flow.json | 5 | 12 | ✅ |
| promo-video-flow | assets/orgs/content-team/flows/promo-video-flow.json | 6 | 12 | ✅ |
| trading-flow | assets/orgs/trading-team/flows/trading-flow.json | 13 | 13 | ✅ |
| game-design-flow | assets/orgs/game-team/flows/game-design-flow.json | 4 | 11 | ✅ |
| skill-dev-flow | assets/orgs/skill-team/flows/skill-dev-flow.json | 6 | 15 | ✅ |
| feature-flow | assets/orgs/dev-team/flows/feature-flow.json | - | - | ✅ |
| bugfix-flow | assets/orgs/dev-team/flows/bugfix-flow.json | - | - | ✅ |
| hotfix-flow | assets/orgs/dev-team/flows/hotfix-flow.json | - | - | ✅ |
| analysis-flow | assets/orgs/dev-team/flows/analysis-flow.json | - | - | ✅ |
| change-flow | assets/orgs/dev-team/flows/change-flow.json | - | - | ✅ |
| release-flow | assets/orgs/dev-team/flows/release-flow.json | - | - | ✅ |
| batch-flow | assets/orgs/dev-team/flows/batch-flow.json | - | - | ✅ |
