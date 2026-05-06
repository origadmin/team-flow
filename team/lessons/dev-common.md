# Dev Common Lessons

> 所有 Dev 角色（Backend/Frontend/Android/iOS）都必须遵循的经验规则
> 
> 来源：Error Reports 结构化审阅 → Lesson Analyst 翻译
> 加载时机：Dev 角色被触发时预读

---

## Rule 1: Toolchain Gate — 包管理器强制检查

**Problem**: AI 直接使用 npm/pnpm/yarn，未读取 project.md Toolchain 配置，导致 lock 文件冲突和依赖不一致。

**Rule**:
1. 执行任何包管理命令前，必须先读取 `.team/project.md` 的 `Toolchain` 部分
2. 使用 `Toolchain.package_manager` 定义的命令（bun/npm/pnpm/yarn）
3. **禁止**使用与配置不符的包管理器
4. 如果 project.md 不存在 → ⛔ 拒绝执行任务

**Correct**:
- 先读取 project.md → 提取 Toolchain.package_manager → 使用该命令
- bun 项目: `bun add`, `bun install`, `bun run`
- npm 项目: `npm install`, `npm run`

**Wrong Example**:
```bash
# ❌ 错误：未读取配置直接使用 npm
npm install lodash

# ❌ 错误：使用 pnpm 但项目配置是 bun
pnpm add lodash
```

**Correct Example**:
```yaml
# project.md
Toolchain:
  package_manager: bun
  install: "bun add"
  install_dev: "bun add -d"
  run: "bun run"
```

```bash
# ✅ 正确：先读取配置，再执行命令
bun add lodash
```

**Applies When**: 任何需要执行包管理命令的场景（install/add/run/build）

**Source**: Error Report E00001-F001-001

---

## Rule 2: Zero Chinese Comments — 精确范围

**Problem**: AI 有时在文档或测试文件中也禁止中文，或 conversely 在代码文件中保留中文注释。

**Rule**:
1. **禁止中文注释**的文件类型：`.go`, `.ts`, `.tsx`, `.js`, `.jsx`, `.java`, `.kt`, `.swift`
2. **允许中文**的文件类型：`.md`, `.txt`, `.yaml`, `.yml`, `.json`（配置/文档）
3. 测试文件（`*_test.go`, `*.test.ts`）属于代码文件，同样禁止中文注释

**Correct**:
```go
// GetUser retrieves user by ID
// Returns error if user not found
func GetUser(id string) (*User, error) {
    // implementation
}
```

**Wrong Example**:
```go
// 获取用户信息
// 如果找不到返回错误
func GetUser(id string) (*User, error) {
    // 实现逻辑
}
```

**Applies When**: 编写代码文件时

**Source**: Error Report E00005-F003-001

---

## Rule 3: Commit Message — Type 必须小写

**Problem**: AI 使用大写 Commit Type（如 `Feat:`, `Fix:`），不符合规范。

**Rule**:
1. Commit Type 必须**小写**：`feat:`, `fix:`, `docs:`, `style:`, `refactor:`, `test:`, `chore:`
2. 禁止大写：`Feat:`, `FIX:`, `DOCS:` 都是错误的

**Correct**:
```
feat: add user authentication
fix: resolve null pointer in GetUser
docs: update API documentation
```

**Wrong Example**:
```
Feat: add user authentication      # ❌ 大写 F
FIX: resolve null pointer          # ❌ 大写 FIX
DOCS: update API                   # ❌ 大写 DOCS
```

**Applies When**: 执行 git commit 时

**Source**: Error Report E00008-F004-001

---

## Rule 4: Regression Guard — 分层测试（先局部后全量）

**Problem**: AI 修改代码后只跑当前任务的测试，不跑全量已有测试，导致已有功能被静默破坏。但每次修改都跑全量测试又太慢。

**Rule**:
1. **开发中（TDD 循环）**: 只跑当前模块测试 `go test ./internal/features/xxx/...`，快速反馈
2. **阶段性回归**: 连续修改同一模块时，每 N 次局部测试后跑一次全量（建议 N=3）
3. **完成门禁**: 修改完成后，必须跑全量测试 `go test ./...`，确认无回归
4. **零回归容忍**: 全量测试中任何之前通过的测试失败 → ⛔ 禁止继续，必须修复或回滚

**Correct**:
```bash
# 开发中 — TDD 快速反馈（秒级）
go test ./internal/features/user/...    # 局部 PASS ✅
go test ./internal/features/user/...    # 局部 PASS ✅
go test ./internal/features/user/...    # 局部 PASS ✅

# 阶段性回归 — 每 N 次局部后跑一次（分钟级）
go test ./...                           # 全量 PASS ✅ → 继续

# 完成门禁 — 最终确认
go test ./...                           # 全量 PASS ✅ → 可以提交
```

**Wrong Example**:
```bash
# ❌ 错误：开发中每次都跑全量（太慢，卡死）
go test ./...    # 每改一行就全量，浪费时间

# ❌ 错误：完成后只跑局部测试
go test ./internal/features/user/...    # 漏掉其他模块的回归

# ❌ 错误：发现已有测试失败但继续开发新功能
```

**Regression Failure Protocol**:
```
发现回归失败
    │
    ├── 新代码导致？→ 修复新代码 → 重新全量测试
    ├── 原有 Bug 暴露？→ 记录为 B{NNN}，不混入当前任务
    └── 不确定？→ git stash 新代码 → 全量测试 → 确认基线 → 逐步引入
```

**Applies When**: 修改任何已有代码文件时（新增文件只需完成门禁全量测试）

**Source**: AI 回归破坏频发问题

---

## Rule 5: Read Before Modify — 修改前必读完整上下文

**Problem**: AI 只看局部代码就修改，不理解函数/模块的完整职责和调用关系，导致"改了 A 坏了 B"。

**Rule**:
1. **修改任何已有文件前**，必须先完整阅读该文件，理解：
   - 文件的整体职责和结构
   - 被修改函数的完整逻辑（不只是要改的那几行）
   - 该函数的所有调用方（grep 调用点）
2. **修改公共接口时**（exported function/method/interface），必须：
   - 搜索所有引用点：`grep -r "FunctionName" --include="*.go"`
   - 列出受影响的调用方
   - 确认修改后所有调用方仍然兼容
3. **禁止局部盲改**: 禁止只看 diff 范围内的代码就做修改

**Correct**:
```
收到修改请求
    │
    ├── Step 1: 完整阅读目标文件
    ├── Step 2: grep 目标函数/类型的所有引用点
    ├── Step 3: 理解调用链和依赖关系
    ├── Step 4: 评估修改影响范围
    └── Step 5: 执行修改 + 全量回归测试
```

**Wrong Example**:
```go
// ❌ 错误：只看了这个函数就改签名，不知道有 20 个调用方
func GetUser(ctx context.Context, id string) (*User, error) {
    // AI 直接加了一个参数，破坏了所有调用方
}
```

**Correct Example**:
```go
// ✅ 正确：先搜索引用点
// grep -r "GetUser" --include="*.go" → 发现 20 个调用方
// 评估：修改签名会影响 20 处 → 改为新增函数，保留旧函数兼容
func GetUser(ctx context.Context, id string) (*User, error) {
    // 保持不变
}

func GetUserWithOptions(ctx context.Context, id string, opts ...Option) (*User, error) {
    // 新增函数
}
```

**Applies When**: 修改任何已有代码文件时

**Source**: AI 局部盲改导致连锁破坏问题

---

## Rule 6: Protected Interface Contract — 接口契约不可单方面变更

**Problem**: AI 修改了接口/协议定义，但没有同步更新所有实现方和调用方，导致编译通过但运行时行为不一致。

**Rule**:
1. **接口签名变更** = Breaking Change，必须：
   - 在 R3_API_CONTRACT.md 中记录变更
   - 通知所有实现方和调用方
   - 提供兼容方案（版本化接口或废弃标记）
2. **数据库 Schema 变更** = Breaking Change，必须：
   - 评估数据迁移方案
   - 确认向后兼容
3. **Proto 文件变更** = Breaking Change，必须：
   - 遵循 Proto 兼容性规则（字段编号不可复用、不可删除）
   - 生成新代码后全量编译验证

**Breaking Change 判定**:
| 变更类型 | 是否 Breaking | 处理方式 |
|---------|-------------|---------|
| 新增字段/方法 | ❌ Non-Breaking | 直接添加 |
| 修改字段类型 | ✅ Breaking | 版本化 + 兼容层 |
| 删除字段/方法 | ✅ Breaking | 废弃标记 → 延迟删除 |
| 修改返回值结构 | ✅ Breaking | 版本化接口 |
| 修改错误码 | ✅ Breaking | 保留旧码 + 新增新码 |

**Applies When**: 修改接口、Proto、数据库 Schema 时

**Source**: AI 单方面改接口导致集成崩溃问题

---

## Rule 7: Sub-Agent PRE-FLIGHT Enforcement — 子 Agent 必须执行工具链门禁

**Problem**: bugfix-expert 子 Agent 在执行 B068 任务时，直接使用 npm/npx 命令而非项目配置的 bun。尽管 project.md §TOOLCHAIN、pre-flight.md、dev-common.md Rule 1 都明确规定了工具链门禁，但子 Agent 完全忽略了这些规则。根因：子 Agent prompt 未注入 PRE-FLIGHT 规则和 Toolchain 配置，AI 凭训练数据肌肉记忆执行 npm/npx。

**Rule**:
1. **Triage 分发任务时**，必须在子 Agent prompt 中注入以下信息：
   - `package_manager: bun`（从 project.md §TOOLCHAIN 读取）
   - 明确禁止的命令：`npm`, `npx`, `pnpm`, `yarn`
   - 正确替代：`bun install`, `bun run`, `bunx`
2. **子 Agent 执行任何 shell 命令前**，必须先输出 PRE-FLIGHT 确认：
   ```
   ✅ PRE-FLIGHT: pkg=bun | constraints=已读 | conventions=已读/无
   ```
3. **Triage 在分发 prompt 模板中**，必须包含 Toolchain 门禁段落，不可省略

**Correct**:
```
Triage 分发 prompt 示例:

---
## 工具链门禁（HARD GATE — 违反 = 任务失败）

本项目前端使用 bun，以下命令 ⛔ 禁止使用：
- npm install / npm run → 用 bun install / bun run
- npx xxx → 用 bunx xxx
- pnpm add → 用 bun add
- yarn dev → 用 bun run dev

执行任何命令前必须输出: ✅ PRE-FLIGHT: pkg=bun
---
```

**Wrong Example**:
```
# ❌ 错误：Triage 分发 prompt 中未提及工具链
你是 bugfix-expert，执行任务 B068: 修复 useTheme 问题
（无任何工具链信息 → 子 Agent 凭肌肉记忆用 npm）
```

**Applies When**: Triage 分发任何涉及前端命令执行的子 Agent 任务时

**Source**: B068 任务中 bugfix-expert 使用 npm/npx 而非 bun

---

## Rule 8: API Contract Alignment — 前后端接口契约对照

**Problem**: AI修改后端API或前端API调用时，未同步更新对端，导致接口404、字段不匹配、参数映射错误。典型案例：B062(扁平vs嵌套结构)、B063(路径复数vs单数)、B067(create_time vs created_at)、B069(status vs state)。

**Rule**:
1. 修改后端路由/Handler时，必须同步检查前端API调用路径
2. 修改后端响应结构时，必须同步检查前端TypeScript类型定义
3. 修改后端请求参数名时，必须同步检查前端API调用参数
4. 新增后端API时，必须确认前端有对应调用且路径/参数/响应一致
5. 使用 `grep -r "apiPath" web/src/` 搜索前端引用点

**Correct**:
```
修改后端路由 POST /api/v1/media/:id/like
    │
    ├── Step 1: grep -r "/media/" web/src/lib/api/ → 找到前端调用
    ├── Step 2: 对比前端路径 vs 后端路径 → 确认一致
    ├── Step 3: 对比前端参数 vs 后端期望参数 → 确认一致
    ├── Step 4: 对比前端TypeScript类型 vs 后端响应 → 确认一致
    └── Step 5: 修改完成 → 更新门禁检查项
```

**Wrong Example**:
```
# ❌ 错误：修改后端路由 /like → /likes 但未检查前端
# ❌ 错误：后端参数名 state 但前端传 status
# ❌ 错误：后端返回 create_time 但前端期望 created_at
```

**Applies When**: 修改任何后端Handler/路由/Proto/响应类型，或修改任何前端API调用

**Source**: B062/B063/B066/B067/B069 五个Bug的共同根因分析(A008)

---

## Rule 9: Field Naming Convention — 字段命名规约（Ent→Proto→前端全链路统一）

**Problem**: Ent Schema 使用 `created_at`/`updated_at`，但 Proto 和前端使用 `create_time`/`update_time`，导致 abgen 无法自动匹配字段，convpb 转换层全部 miss，时间数据从未传到前端。AI 在修复时只做了映射补丁（convpb 手动映射），未从根源修复（Ent Schema 重命名），导致后续 abgen 重新生成时映射丢失。更严重的是，Ent Schema 重命名后，业务代码中大量 `.CreatedAt`/`.UpdatedAt`/`SetCreatedAt`/`SetUpdatedAt`/`"created_at"` 残留未清理，编译失败。

**Rule**:
1. **字段命名必须在 Ent Schema 层就与 Proto 对齐**，禁止在转换层打补丁
2. **项目标准命名**（F014 已确立）：

| 概念 | ✅ 正确 | ❌ 禁止 |
|------|---------|---------|
| 创建时间 | `create_time` | ~~`created_at`~~ |
| 更新时间 | `update_time` | ~~`updated_at`~~ |
| 创建者 | `create_author` | ~~`created_by`~~ |
| 更新者 | `update_author` | ~~`updated_by`~~ |

3. **Ent Schema 是源头**：字段名在 Schema 定义时就必须使用标准命名，因为：
   - Ent 生成代码的 Go 字段名、JSON tag、数据库列名都源自 Schema
   - abgen 按**同名匹配**生成 convpb 转换代码，Schema 和 Proto 命名一致则自动映射
   - 如果 Schema 用 `created_at` 而 Proto 用 `create_time`，abgen 会标记 `// miss`，需要手动映射，且每次重新生成都会丢失

4. **重命名字段时必须全链路清理**，搜索范围：
   - Ent Schema 定义：`internal/data/entity/schema/*.go`
   - Ent 生成代码：`go generate ./internal/data/entity/` 后自动更新
   - 业务逻辑代码：`.CreatedAt`/`.UpdatedAt` → `.CreateTime`/`.UpdateTime`
   - Ent Update 方法：`SetUpdatedAt()` → `SetUpdateTime()`
   - Ent 查询谓词：`media.CreatedAtGTE()` → `media.CreateTimeGTE()`
   - 排序字段常量：`media.FieldCreatedAt` → `media.FieldCreateTime`
   - 字符串排序参数：`"created_at"` → `"create_time"`
   - 数据库迁移配置：`database.go` 中的列名
   - convpb 转换层：重新生成或手动更新
   - 前端类型定义：`api.ts`/`api.d.ts`
   - 前端 API 调用：`order_by: 'created_at'` → `order_by: 'create_time'`

5. **验证命令**：重命名后必须执行以下搜索确认零残留：
   ```bash
   # 后端：Go 字段引用
   grep -rn "\.CreatedAt\|\.UpdatedAt\|SetCreatedAt\|SetUpdatedAt\|FieldCreatedAt\|FieldUpdatedAt" internal/ --include="*.go"
   # 后端：字符串引用
   grep -rn '"created_at"\|"updated_at"' internal/ --include="*.go"
   # 前端：类型和调用
   grep -rn "created_at\|updated_at" web/src/ --include="*.ts" --include="*.tsx"
   ```

**Correct**:
```
Ent Schema:  field.Time("create_time").Default(time.Now)
Proto:       google.protobuf.Timestamp create_time = 4 [json_name = "create_time"]
前端类型:     create_time?: string
abgen 生成:   CreateTime: from.CreateTime  ← 自动匹配，无需手动映射
```

**Wrong Example**:
```
# ❌ 错误：Ent Schema 用 created_at，Proto 用 create_time
Ent Schema:  field.Time("created_at").Default(time.Now)  ← 命名不一致
Proto:       google.protobuf.Timestamp create_time = 4
abgen 生成:  // miss: CreateTime  ← 无法自动匹配
convpb 手动:  CreateTime: ConvertTimeToTimestamp(from.CreatedAt)  ← 补丁，重新生成会丢失

# ❌ 错误：重命名 Schema 后未清理业务代码
Schema 改为 create_time → go generate → 编译失败
原因：业务代码仍用 .CreatedAt / SetUpdatedAt() / "created_at"

# ❌ 错误：只在 convpb 加映射而不改 Schema 根源
convpb.gen.go: CreateTime: ConvertTimeToTimestamp(from.CreatedAt)  ← 治标不治本
```

**Applies When**: 新建 Ent Schema 字段、重命名 Ent Schema 字段、修改 Proto 字段名、任何涉及 Ent→Proto→前端字段映射的变更

**Source**: C018/C019/C020 用户表 uuid 删除 + 时间字段重命名，暴露的命名不一致根因
