# Architecture Decisions — team-flow v3 Redesign

### D1: v1/v2/v3 并行共存（2026-05-27）

- **Decision**: v1、v2、v3 三个版本同步并行，各自有不同的能力和流程，不存在"v1 升级到 v2"或"v2 升级到 v3"的线性关系
- **Reason**: 三个版本面向不同使用场景，用户按需选择
- **Alternatives**: 线性升级（v1→v2→v3），废弃旧版本
- **Impact**: 所有代码和文档必须同时支持三个版本

| 版本 | 能力范围 | 流程管理 | 任务管理 |
|------|---------|---------|---------|
| v1 | 纯文档管理开发流程 | Markdown 文件驱动 | 无 |
| v2 | 使用 flow task 管理开发流程 | Markdown + flow task | beads (bd CLI) |
| v3 | 使用 flow 全功能，管理各种流程和团队 | JSON 流程定义 + 引擎驱动 | 配置驱动（beads/file/custom） |

### D2: 配置驱动工具路由（2026-05-27）

- **Decision**: 后端不使用硬编码 Go 实现，而是使用配置驱动的方式。新工具/技能的增加只需更新配置，不需要修改代码
- **Reason**: 随着技能增加，使用的工具会持续增加，不可能一直修改代码
- **Alternatives**: 硬编码 Backend 接口（BeadsBackend/FileBackend/GitBackend 等）
- **Impact**: internal/tools/ 实现通用 ToolRouter，工具注册通过 tools.yaml 配置文件

### D3: Flow 统一入口架构（2026-05-27）

- **Decision**: Flow 的职责是统一入口 → 内部处理（可选）→ 转发 → 内部处理（可选）
- **Reason**: Flow 作为唯一入口，对 AI 屏蔽内部实现细节，所有操作通过 flow 命令完成
- **Alternatives**: AI 直接调用底层工具（bd, git 等）
- **Impact**: Flow 命令 = 入口 + 路由 + 转发，不是业务逻辑实现者

### D4: 统一工具路径（2026-05-27）

- **Decision**: tools.yaml 中定义统一工具路径，AI 通过 `flow config paths` 和 `flow tools list` 发现可用工具
- **Reason**: AI 无法直接找到外部命令路径，需要 flow 提供工具发现机制
- **Alternatives**: AI 自行搜索 PATH，硬编码工具路径
- **Impact**: tools.yaml 包含 paths 段，flow 提供 `flow tools list` 和 `flow tools run` 命令

### D5: 单词目录命名（2026-05-27）

- **Decision**: internal/ 下的包名使用单词命名，不使用双单词缩写
- **Reason**: 当前复杂度不需要双单词，单词更简洁清晰
- **Alternatives**: toolcfg, configcmd 等双单词缩写
- **Impact**: toolcfg → tools, configcmd → 合并到 cmd/config.go, 其他双单词包同理

### D6: 两目录归一（2026-05-27）

- **Decision**: 只保留 skill/ 和 orgs/ 两个顶层目录，删除 teams/ 和 v3/teams/
- **Reason**: 当前 team/（技能）、teams/（v2模板）、v3/（v3模板）三个目录职责交叉、内容重复
- **Alternatives**: 保留三目录，增加版本子目录
- **Impact**: skill/ 包含 v1/v2/v3 三版技能文件，orgs/ 包含组织模板（v3 only）

### D7: 组织模板绑定团队+流程（2026-05-27）

- **Decision**: 流程必须绑定团队，没有团队无法执行。团队+流程 = 组织模板（Organization Template）
- **Reason**: 没有团队 = 公司没有人 = 无法工作
- **Alternatives**: 团队和流程松散关联，可独立选择
- **Impact**: org.json = 团队+流程的绑定容器，选流程 = 选组织

### D8: 主流程+子流程分层（2026-05-27）

- **Decision**: 模板层支持多个主流程，项目层只能绑定一个主流程。子流程不可脱离主流程独立使用
- **Reason**: feature-flow、bugfix-flow 无法脱离 dev-flow 独立使用
- **Alternatives**: 所有流程平级
- **Impact**: org.json 中 main_flows[] + sub_flows[]，flow init 只展示主流程

### D9: 工具化编辑（2026-05-27）

- **Decision**: 所有内容编辑通过 flow CLI 命令，禁止 AI 手动编辑 JSON/YAML 配置文件
- **Reason**: AI 手动编辑可能导致角色不存在、技能名称不标准等问题
- **Alternatives**: 允许 AI 直接编辑配置文件
- **Impact**: 新增 flow org * 命令族，CLI 内置验证链

### D10: 冷热分层追踪（2026-05-27）

- **Decision**: 每步对话必须记录，但使用冷热分层存储省 TOKEN
- **Reason**: 不能把所有历史加载到 prompt，也不能丢失讨论内容
- **Alternatives**: 全部加载（浪费 TOKEN），不记录（丢失内容）
- **Impact**: HOT（~300 tokens 在 prompt）+ COLD（磁盘按需查询）
