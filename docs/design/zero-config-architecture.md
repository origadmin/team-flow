# Zero/Minimal-Config Architecture for team-flow v3

## 1. 概述

### 1.1 目标

将 team-flow v3 的配置复杂度从"必须提供 `team.json` + `project.yaml`"降为三级：

| 级别 | 用户提供 | 系统行为 |
|------|---------|---------|
| **Zero Config** | 无 | 自动使用 `dev-flow`，org 推导为 `dev-team` |
| **Minimal Config** | `project.yaml`（仅 `version: v3`，可选 `active_flow`） | 按 `active_flow` 运行，org 按约定推导 |
| **Full Custom** | `team.json`（含 roles/rules/flows/skills） | 保留现有行为，全量覆盖 |

### 1.2 核心变更

1. **移除 `builtinTeamDefaults` map** — 用约定替代硬编码映射
2. **移除 `team.json.id` 作为独立字段** — id 不再承担 org 回退职责
3. **org 从 `active_flow` 按约定推导** — `{name}-flow` → `{name}-team`
4. **版本默认为 `v3`** — 无 `project.yaml` 时自动使用 v3
5. **统一 `project.yaml` 加载** — 消除 `projectYAML`（私有）和 `ProjectConfig`（公开）两套结构

---

## 2. 新 `project.yaml` 结构

### 2.1 最小结构

```yaml
# .team/project.yaml — Minimal Config
version: v3
active_flow: skill-dev-flow  # 可选，默认 dev-flow
```

### 2.2 完整结构（合并现有 ProjectConfig + projectYAML）

```yaml
# .team/project.yaml — Full Config
version: v3
active_flow: dev-flow
name: my-project           # 项目名（可选）

paths:
  projects_path: projects/  # monorepo 子项目路径
  docs_internal: .team/docs
  docs_external: ""
  backup_path: ""

flow:
  path: ""                  # 自定义 flow 路径
  backup_path: ""
  update_disabled: false
  update_interval: 24h

toolchain:
  backend:
    language: go
    pipeline: kratos
  frontend:
    language: typescript
    pipeline: vite

dependencies:               # monorepo 依赖
  - name: shared-lib
    path: ../shared-lib
    type: library

skills:                     # skill 管理
  tags: [go, backend]
  local: []
  disabled: []
  overrides: {}

projects:                   # monorepo 子项目声明
  - name: my-service
    path: services/my-service
    type: backend
```

### 2.3 结构体定义（重构后）

```go
// config/project.go — 合并后的唯一 ProjectConfig

type ProjectConfig struct {
    Version      string             `yaml:"version"`       // 默认 "v3"
    ActiveFlow   string             `yaml:"active_flow"`   // 默认 "dev-flow"
    Name         string             `yaml:"name,omitempty"`
    Paths        ProjectPaths       `yaml:"paths"`
    Toolchain    ProjectToolchain   `yaml:"toolchain"`
    Flow         FlowConfig         `yaml:"flow,omitempty"`
    Dependencies []ProjectDependency `yaml:"dependencies,omitempty"`
    Skills       ProjectSkillConfig `yaml:"skills,omitempty"`
    Projects     []ProjectDecl      `yaml:"projects,omitempty"`
    // 删除: ActiveProject（从未真正使用）
    // 删除: Flows（team.json 已覆盖）
    // 删除: Team（不再需要，org 从 active_flow 推导）
}

// 删除: projectYAML（runtime.go 私有结构体）
```

---

## 3. `active_flow` → `org` 推导

### 3.1 约定规则

```
active_flow:  "{prefix}-flow"      →  org: "{prefix}-team"
active_flow:  "dev-flow"           →  org: "dev-team"
active_flow:  "skill-dev-flow"     →  org: "skill-team"
active_flow:  "game-design-flow"   →  org: "game-team"
active_flow:  "trading-flow"       →  org: "trading-team"
active_flow:  "content-distribution-flow" →  org: "content-team"
```

### 3.2 推导函数

```go
// DeriveOrgFromFlow 从 flow 名称推导 org 名称
// 规则: 去掉 "-flow" 后缀，加上 "-team" 后缀
// 示例: "dev-flow" → "dev-team", "skill-dev-flow" → "skill-team"
func DeriveOrgFromFlow(flowName string) string {
    if strings.HasSuffix(flowName, "-flow") {
        return flowName[:len(flowName)-len("-flow")] + "-team"
    }
    // 兼容: 无后缀的 flow 名, org = flow + "-team"
    return flowName + "-team"
}
```

### 3.3 Org 来源优先级

```
1. team.json 显式 org 字段               → 最高优先级（用户显式指定）
2. 从 active_flow 按约定推导             → 默认行为
3. "dev-team"                           → 硬编码回退（零配置场景）
```

### 3.4 org 在代码中的使用点

org 用于以下场景，均使用推导结果：

| 使用点 | 函数 | 说明 |
|--------|------|------|
| Flow 加载 | `LoadFlow` → embedded fallback | `assets/orgs/{org}/flows/{flowID}.json` |
| Role 加载 | `loadRole` → embedded fallback | `assets/orgs/{org}/roles/{roleID}.json` |
| Role 合并 | `loadTemplateRoles` | `assets/orgs/{org}/team.json` |
| Rule 加载 | `loadRule` | `assets/orgs/{org}/rules/{ruleID}.json` |
| Rule 展开 | `buildRuleOutput` | 从 org 的 embed 加载缺失 rule |

---

## 4. 配置加载链（重构后）

### 4.1 总体流程

```
┌──────────────────────────────────────────────────────────┐
│                  flow proc run                           │
└───────────────────────┬──────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────┐
│ Step 1: LoadProjectConfig(root)                          │
│   读取 .team/project.yaml                                │
│   ┌─────────────────────────────────────────────────┐    │
│   │ 不存在? → 返回默认配置:                          │    │
│   │   Version: "v3"                                 │    │
│   │   ActiveFlow: "dev-flow"   ← ZERO CONFIG        │    │
│   └─────────────────────────────────────────────────┘    │
│   ┌─────────────────────────────────────────────────┐    │
│   │ 存在但无 active_flow?                             │    │
│   │   → ActiveFlow = "dev-flow" ← MINIMAL CONFIG    │    │
│   └─────────────────────────────────────────────────┘    │
└───────────────────────┬──────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────┐
│ Step 2: DeriveOrgFromFlow(active_flow)                   │
│   "dev-flow" → "dev-team"                               │
│   (team.json.org 可覆盖此推导结果)                        │
└───────────────────────┬──────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────┐
│ Step 3: LoadTeam(root) — 可选                            │
│   读取 .team/team.json (FULL CUSTOM)                     │
│   存在 → 加载 roles/rules/flows/skills                   │
│   不存在 → 返回 nil（零配置场景）                          │
│                                                          │
│   如果 team.json 有 org 字段 → 覆盖 Step 2 的推导结果     │
│   如果 team.json 有 active_flow → 覆盖 project.yaml       │
└───────────────────────┬──────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────┐
│ Step 4: LoadFlow(root, active_flow)                      │
│   1. team.json flows[*].file → 显式文件路径               │
│   2. ResolveProcPath → 磁盘查找                          │
│   3. embedded: assets/orgs/{org}/flows/{flow}.json       │
│   4. embedded: assets/orgs/dev-team/flows/{flow}.json    │
│      (保留 dev-team 作为全局硬回退)                       │
└───────────────────────┬──────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────┐
│ Step 5: LoadTemplateRoles(team) — 当 team 只有 org 时     │
│   读取 assets/orgs/{org}/team.json                       │
│   将内嵌模板的 roles/rules/flows/skills 合并到 team       │
└──────────────────────────────────────────────────────────┘
```

### 4.2 关键函数变更

#### `ResolveActiveFlow` — 简化

```go
// 之前:
func ResolveActiveFlow(root string) (string, error) {
    cfg, err := loadProjectConfig(root)      // → 私有 projectYAML
    if err == nil && cfg.ActiveFlow != "" {
        return cfg.ActiveFlow, nil
    }
    teamID := detectTeamID(root)             // → 读 team.json.id
    if def, ok := builtinTeamDefaults[teamID]; ok {  // → 硬编码 map
        return def, nil
    }
    return "", fmt.Errorf("no active flow...")
}

// 之后:
func ResolveActiveFlow(root string) (string, error) {
    cfg, _ := config.LoadProjectConfig(root)     // → 公开 ProjectConfig
    if cfg != nil && cfg.ActiveFlow != "" {
        return cfg.ActiveFlow, nil
    }
    // team.json 的 active_flow（Full Custom 场景）
    if team, _ := LoadTeam(root); team != nil && team.ActiveFlow != "" {
        return team.ActiveFlow, nil
    }
    // 零配置回退
    return "dev-flow", nil   // 不再返回 error
}
```

#### `DeriveOrg` — 新增统一入口

```go
// DeriveOrg 返回项目使用的 org 名称
// 优先级: team.json.org > active_flow 推导 > "dev-team" 回退
func DeriveOrg(root, activeFlow string) string {
    if team, _ := LoadTeam(root); team != nil && team.Org != "" {
        return team.Org
    }
    if activeFlow != "" {
        return DeriveOrgFromFlow(activeFlow)
    }
    return "dev-team"
}
```

#### `LoadTeam` — 移除 id 強制檢查

```go
// 之前:
func LoadTeam(root string) (*flow.TeamDefinition, error) {
    // ...
    if json.Unmarshal(data, &team) != nil || team.ID == "" {
        return nil, nil  // ← id 为空则丢弃
    }
    // ...
}

// 之后:
func LoadTeam(root string) (*flow.TeamDefinition, error) {
    // ...
    if err := json.Unmarshal(data, &team); err != nil {
        return nil, nil
    }
    // id 可以为空（由其他字段驱动行为）
    // org 为空时从 active_flow 推导
    // 只要有 roles/rules/flows 中任意一项就视为有效 team
    if team.hasContent() {
        return &team, nil
    }
    return nil, nil
}

func (t *TeamDefinition) hasContent() bool {
    return len(t.Roles) > 0 || len(t.Rules) > 0 || len(t.Flows) > 0 ||
           t.ActiveFlow != "" || t.Org != ""
}
```

#### `discoverOrg` — 重写为使用 DeriveOrg

```go
// 之前:
func discoverOrg(root string) string {
    // 读 team.json → team.Org → 回退 team.ID
}

// 之后:
func discoverOrg(root string) string {
    activeFlow, _ := ResolveActiveFlow(root)
    return DeriveOrg(root, activeFlow)
}
```

#### `LoadTemplateRoles` — org 来源变更

```go
// 之前: team.Org 来自 team.json 的显式字段
// 之后: team.Org 可以由调用方预先设置（通过 DeriveOrg）
// 如果 team.Org 为空但 team 有内容，由调用方在调用前填入推导的 org

// 在 LoadTeam 中:
func LoadTeam(root string) (*flow.TeamDefinition, error) {
    // ... 读取 team.json
    if team.Org == "" && team.hasContent() {
        // 推导 org
        activeFlow, _ := ResolveActiveFlow(root)
        team.Org = DeriveOrgFromFlow(activeFlow)
    }
    if len(team.Roles) == 0 && team.Org != "" {
        loadTemplateRoles(&team)
    }
    return &team, nil
}
```

---

## 5. 移除的内容

### 5.1 `builtinTeamDefaults` map

**位置**: `runtime.go:210-216`

```go
// 删除:
var builtinTeamDefaults = map[string]string{
    "dev-team":     "dev-flow",
    "skill-team":   "skill-dev-flow",
    "trading-team": "trading-flow",
    "game-team":    "game-design-flow",
    "content-team": "content-distribution-flow",
}
```

**替代方案**: 零配置直接回退 `dev-flow` + `dev-team`。其他团队使用 `project.yaml` 的 `active_flow` 指定，org 自动推导。

### 5.2 `team.json.id` 作为配置驱动的字段

`team.json` 可以保留 `id` 字段用于显示/标识，但不再参与配置逻辑：
- 不再作为 org 回退
- 不再作为 `builtinTeamDefaults` 的 key
- `LoadTeam` 不再因 `id==""` 返回 nil

### 5.3 `org` 作为 `project.yaml` 的独立字段

`project.yaml` 不再有 `team` 字段。org 由 `active_flow` 推导。

### 5.4 私有 `projectYAML` + `loadProjectConfig`

**位置**: `runtime.go:299-319`

合并到公开的 `config.LoadProjectConfig` 中。统一为一个入口。删除 `runtime.go` 中的 `projectYAML` 结构体和 `loadProjectConfig` 函数。

### 5.5 `detectTeamID`

**位置**: `runtime.go:248-265`

不再需要。team id 不再参与配置逻辑。

---

## 6. 完整自定义路径（team.json）共存

### 6.1 team.json 的新角色

`team.json` 成为**纯粹的装配清单**，不再提供基本配置（org/flow 推导）：

```json
{
  "id": "my-custom-team",        // 保留用于显示/标识，不参与逻辑
  "name": "My Team",
  "org": "my-custom-org",        // 可选：显式覆盖 org 推导
  "active_flow": "my-flow",      // 可选：覆盖 project.yaml 的 active_flow
  "roles": [...],                 // 角色定义
  "rules": [...],                 // 规则定义
  "flows": [                      // flow 文件映射
    {"id": "my-flow", "file": "flows/custom.json"}
  ],
  "skills": {...}
}
```

### 6.2 优先级链

```
高
│  team.json.active_flow     → 覆盖 project.yaml.active_flow
│  team.json.org             → 覆盖 active_flow 推导的 org
│  team.json.flows[*].file   → 覆盖 ResolveProcPath 磁盘查找
│  team.json.roles           → 覆盖 assets/orgs/{org}/team.json 模板
│
│  project.yaml.active_flow  → 用户显式指定
│  DeriveOrgFromFlow()       → 从 active_flow 约定推导
│
│  "dev-flow"                → 零配置硬回退
│  "dev-team"                → 零配置硬回退
低
```

---

## 7. 具体代码变更清单

### 7.1 文件: `projects/team-flow/internal/proc/runtime.go`

| 变更 | 说明 |
|------|------|
| **删除** `builtinTeamDefaults` (L210-216) | 不再需要硬编码映射 |
| **删除** `projectYAML` 结构体 (L312-319) | 合并到 config.ProjectConfig |
| **删除** `loadProjectConfig` 函数 (L299-310) | 合并到 config.LoadProjectConfig |
| **删除** `detectTeamID` (L248-265) | 不再需要 |
| **重写** `ResolveActiveFlow` (L227-246) | 零配置回退 "dev-flow"，移除 builtinTeamDefaults 查找 |
| **重写** `LoadTeam` (L321-340) | 移除 `id == ""` 检查，添加 `hasContent()` 逻辑，自动推导 org |
| **新增** `DeriveOrgFromFlow` | 从 flow 名称推导 org |
| **新增** `DeriveOrg` | 统一 org 推导入口（考虑 team.json 覆盖） |
| **重写** `discoverOrg` (L372-387) | 使用 DeriveOrg，不再读取 team.json.id |
| **修改** `loadTemplateRoles` (L342-370) | 无变更（函数签名不变，但调用方保证 team.Org 已填入） |
| **修改** `LoadFlow` (L27-87) | 移除 `discoverOrg` 中的 team.ID 回退依赖，硬回退保持 `dev-team` |

### 7.2 文件: `projects/team-flow/internal/config/project.go`

| 变更 | 说明 |
|------|------|
| **修改** `ProjectConfig` 结构体 | 删除 `Team` 字段，删除 `ActiveProject` 字段，删除 `Flows` 字段 |
| **修改** `LoadProjectConfig` | 当 `project.yaml` 不存在时，返回默认配置而非 error |
| **新增** `DefaultProjectConfig()` | 返回 `{Version: "v3", ActiveFlow: "dev-flow"}` |
| **修改** `parseProjectMD` | 删除项目级 flow 列表解析（不再需要） |

### 7.3 文件: `projects/team-flow/internal/proc/procrun.go`

| 变更 | 说明 |
|------|------|
| **修改** `Run` → team 加载 (L342-358) | `LoadTeam` 现在可能返回推导了 org 的 team |
| **修改** `resolveRoleInfo` → org 回退 (L1008-1011) | 不再使用 `team.ID` 作为 org 回退 |
| **修改** `buildRuleOutput` → org 回退 (L1170-1171) | 确保 `team.Org` 总是已填入 |
| **修改** `buildTeamIntro` → `team.ID` 引用 (L1390) | 保留 `team.ID` 仅用于显示 |

### 7.4 文件: `projects/team-flow/internal/proc/cmd.go`

| 变更 | 说明 |
|------|------|
| **修改** `runRun` (L482-569) | `ResolveActiveFlow` 不再返回 error（总是有值） |
| **修改** `runNext` (L572-733) | 同上 |
| **修改** `runGate` (L1751-1838) | 同上 |
| **修改** `runRule` (L1301-1341) | 同上 |
| **修改** `runShow` (L1083-1131) | 同上 |
| **修改** `readActiveFlow` (L879-892) | 回退逻辑可在 `project.yaml` 缺失时正常工作 |
| **修改** `resolveRuleTargetPath` (L1478-1495) | 同上 |

### 7.5 文件: `projects/team-flow/internal/proc/resolver.go`

| 变更 | 说明 |
|------|------|
| **修改** `ActiveFlowResolver.Resolve` (L38-44) | `ResolveActiveFlow` 不再返回 error |

---

## 8. 迁移路径

### 8.1 现有项目兼容

**场景 A**: 项目有完整 `team.json`（含 `id`, `org`, `roles`, 等）
- ✅ **完全兼容** — `LoadTeam` 加载后，所有字段保留
- `team.Org` 显式值优先于推导值
- 行为与之前完全一致

**场景 B**: 项目有 `project.yaml` + `team.json`（分离配置）
- ✅ **完全兼容** — `project.yaml.active_flow` 和 `team.json.org` 各自独立
- 注意：如果 `project.yaml` 有 `team` 字段，迁移后需要手动移除（不再生效）

**场景 C**: 项目只有 `team.json`（含 `id` 用于 org 回退）
- ⚠️ **行为变化** — `team.json.id` 不再用作 org 回退
- 如果 `team.json` 无 `org` 字段且无 `project.yaml` → org 回退到 `dev-team`
- **修复方式**: 在 `team.json` 中添加 `"org": "original-team-name"` 或创建 `project.yaml` 设置 `active_flow`

### 8.2 迁移检查清单

```bash
# 1. 检查哪些项目依赖 team.json.id 作为 org 回退
grep -r '"id"' .team/team.json | grep -v '"org"'

# 2. 如果 team.json 有 id 但无 org，且期望非 dev-team
#    → 添加 "org": "<expected-org>" 到 team.json

# 3. 检查 project.yaml 中的 team 字段
grep 'team:' .team/project.yaml

# 4. 如果 project.yaml 有 team 字段
#    → 删除 team 字段，改用 active_flow（org 自动推导）
```

### 8.3 向前兼容

```go
// 代码层面：LoadTeam 如果检测到 team.json 有 id 但无 org，
// 可以打印一条 deprecation warning（不阻塞运行）:
if team.ID != "" && team.Org == "" {
    fmt.Fprintf(os.Stderr, 
        "⚠ team.json has 'id=%s' but no 'org'. "+
        "org will be derived from active_flow. "+
        "Add 'org' field to team.json for explicit control.\n",
        team.ID)
}
```

---

## 9. 行为对比表

| 场景 | 之前 | 之后 |
|------|------|------|
| 无任何配置文件 | ❌ 报错 "no active flow" | ✅ 自动使用 dev-flow + dev-team |
| 仅有 `project.yaml`(version:v3) | ❌ 报错 "no active flow" | ✅ 自动使用 dev-flow + dev-team |
| `project.yaml` + active_flow: skill-dev-flow | ✅ 使用 skill-dev-flow，org 需额外配置 | ✅ 使用 skill-dev-flow，org 自动推导 skill-team |
| team.json 仅 id: "dev-team" | ✅ org="dev-team", flow="dev-flow" | ✅ org 从 active_flow 推导="dev-team"，flow="dev-flow" |
| team.json id:"dev-team" + org:"custom" | ✅ org="custom" | ✅ org="custom"（显式值优先） |
| team.json 有 roles/rules | ✅ 正常加载 | ✅ 正常加载（无需 id 检查） |
| team.json id:"" 但有 roles | ❌ LoadTeam 返回 nil | ✅ 正常加载（hasContent()=true） |

---

## 10. 实现顺序建议

1. **Phase 1**: 在 `config/project.go` 中实现 `DefaultProjectConfig()` 和 `LoadProjectConfig` 零配置回退
2. **Phase 2**: 在 `runtime.go` 中实现 `DeriveOrgFromFlow`，重写 `ResolveActiveFlow`
3. **Phase 3**: 重写 `LoadTeam`，移除 `id` 强制检查
4. **Phase 4**: 删除 `builtinTeamDefaults`、`detectTeamID`、私有 `projectYAML`
5. **Phase 5**: 重写 `discoverOrg`、`LoadFlow` 中的 org 引用
6. **Phase 6**: 更新 `procrun.go` 和 `cmd.go` 中的所有 `ResolveActiveFlow` 调用点
7. **Phase 7**: 测试：零配置项目启动、最小配置项目启动、完整自定义项目启动

---

## 11. 测试用例

### 11.1 零配置测试

```
Given:  .team/ 目录存在，无 project.yaml，无 team.json
When:   flow proc run --new
Then:   active_flow = "dev-flow", org = "dev-team"
        流程从嵌入的 assets/orgs/dev-team/flows/dev-flow.json 加载
```

### 11.2 最小配置测试

```
Given:  .team/project.yaml:
          version: v3
          active_flow: skill-dev-flow
        .team/team.json 不存在
When:   flow proc run --new
Then:   active_flow = "skill-dev-flow", org = "skill-team"
        流程从 assets/orgs/skill-team/flows/skill-dev-flow.json 加载
```

### 11.3 完整自定义测试

```
Given:  .team/project.yaml:
          version: v3
          active_flow: my-flow
        .team/team.json:
          {"org": "my-org", "roles": [...], "rules": [...]}
When:   flow proc run --new
Then:   active_flow = "my-flow", org = "my-org"（team.json 覆盖推导）
        角色从 team.json.roles 加载
```

### 11.4 Org 覆盖测试

```
Given:  .team/project.yaml:
          version: v3
          active_flow: dev-flow
        .team/team.json:
          {"org": "custom-org"}    # 显式 org 覆盖推导
When:   flow proc run --new
Then:   active_flow = "dev-flow", org = "custom-org"
        嵌入资源从 assets/orgs/custom-org/... 查找
```