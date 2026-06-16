# 编码和测试标准 (Standards)

> **版本**: v3.0  
> **状态**: 核心文档  
> **目的**: 定义编码规范、测试要求和代码质量标准。提交 PR 前必须阅读。  
> **最后更新**: 2026-06-09

---

## 1. Go 编码规范

### 1.1 包结构

- 每个包一个目录
- 包名简短、小写、无下划线（如 `bd`, `task`, `proc`）
- 包的职责单一（一个包做一件事）
- `internal/` 下的包仅供本项目使用（Go 编译器保证）

### 1.2 命名

| 类型 | 规则 | 示例 |
|------|------|------|
| 包名 | 小写、短、含义清晰 | `flow`, `bd`, `task` |
| 导出函数 | PascalCase，动词开头，表达意图 | `CreateIssue`, `RunQuiet` |
| 内部函数 | camelCase，动词开头 | `parseBeadsID`, `saveTask` |
| 变量名 | camelCase，缩写保持大写（ID, URL） | `taskID`, `userURL` |
| 常量 | PascalCase（Go 风格，不使用 SCREAMING_SNAKE_CASE） | `MaxTaskIDLength` |
| 文件名 | snake_case，简短 | `client.go`, `client_test.go` |

### 1.3 错误处理

```go
// ✅ 好：具体的错误信息
output, err := cmd.Output()
if err != nil {
    return "", fmt.Errorf("run bd create: %w", err)
}

// ✅ 好：哨兵错误（可选）
var ErrInvalidTaskID = errors.New("invalid task ID format")

// ❌ 坏：吞掉错误
output, _ := cmd.Output()  // 忽略错误
```

### 1.4 注释

- 每个导出的包、函数、类型必须有 doc 注释
- doc 注释以名称开头（Go 规范）
- 非导出的复杂逻辑也应有行内注释
- 注释应解释"为什么"，而不是"做什么"（代码已经告诉你做什么）

```go
// CreateIssue creates a new issue via the beads CLI and returns the generated ID.
// The ID format is "project-xxx" as specified by beads.
// Returns an error if the beads CLI fails or returns unparseable output.
func CreateIssue(title, issueType, description string) (string, error) {
    // ...
}
```

### 1.5 函数签名

- 尽量返回 `(result, error)` 元组
- 接受上下文时，`context.Context` 作为第一个参数
- 函数参数不超过 4-5 个，过多时考虑 struct 传递

```go
// ✅ 好：清晰的返回类型
func CreateIssue(title, issueType, description string) (string, error)

// ❌ 坏：太多参数，用 struct
func CreateIssue(title, issueType, desc, assignee, priority, labels, project string) (string, error)
```

### 1.6 纯函数

- 逻辑核心应抽取为纯函数
- 纯函数不依赖外部状态，不需要 mock
- 纯函数必须有单元测试

```go
// 纯函数：只依赖输入，可测试
func parseBeadsID(output string) string { ... }

// 有副作用：仅在边缘使用
func RunQuiet(args ...string) (string, error) { ... }
```

---

## 2. 测试规范

### 2.1 测试文件

- 测试文件放在被测文件同一目录
- 命名为 `<source>_test.go`
- 使用 `package <same>` 或 `package <same>_test`（根据是否需要访问内部）

### 2.2 测试命名

- `Test<Function>` — 函数级别（如 `TestParseBeadsID`）
- `Test<Function>_<Scenario>` — 场景级别（如 `TestParseBeadsID_WithWarning`）
- `Test<Function>_Failure` — 失败场景

### 2.3 表驱动测试

对多场景的函数，使用 Go 风格的表驱动测试：

```go
func TestParseBeadsID(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"standard", "team-flow-abc\n", "team-flow-abc"},
        {"with-warning", "warning: something\nteam-flow-abc\n", "team-flow-abc"},
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := parseBeadsID(tt.input)
            if got != tt.expected {
                t.Errorf("got %q, want %q", got, tt.expected)
            }
        })
    }
}
```

### 2.4 测试覆盖要求

| 代码类型 | 单元测试要求 |
|---------|------------|
| 纯函数 | **必须** 有单元测试 |
| 业务逻辑 | **必须** 有测试 |
| 外部集成 | 至少有集成测试（模拟外部工具） |
| CLI 入口 | 集成测试或端到端测试 |

### 2.5 失败测试

每个模块至少包含一些**负向测试用例**（输入为空、格式错误、边界情况等）。

### 2.6 运行测试

```bash
# 全部
go test ./...

# 指定模块（带详细输出）
go test ./internal/bd/ -v

# 仅运行特定测试
go test ./internal/bd/ -v -run TestParseBeadsID

# 覆盖率
go test ./... -cover
```

---

## 3. 代码质量

### 3.1 检查

提交前至少：

1. `go build ./...` — 确保编译通过
2. `go test ./...` — 确保测试通过
3. `go vet ./...` — 静态检查（常见错误）

### 3.2 常见问题清单

- 未处理的 error
- 未使用的 import
- 命名不一致
- 过长的函数（> 50 行可能需要拆分）
- 混合副作用和纯逻辑的函数
- 隐式依赖全局状态

### 3.3 代码风格

- 使用 `go fmt` 格式化（或 IDE 自动格式化）
- 不要在一行写多个语句
- 合理使用空行分隔逻辑块
- 优先使用短变量名（但不要过度缩减到不可读）

---

## 4. 模块和文件组织

### 4.1 文件组织（按职责分文件）

| 典型文件名 | 职责 |
|-----------|------|
| `<module>.go` | 主逻辑（如 `client.go`） |
| `<module>_test.go` | 单元测试 |
| `types.go` | 类型定义 |
| `cmd.go` | CLI 命令注册 |
| `utils.go` | 通用工具（不要过度使用） |

### 4.2 不要创建"万能"文件

- 不要写 1000 行的 `main.go`
- 不要把所有逻辑放一个文件
- 按功能拆文件，每个文件 < 500 行（软限制）

---

## 5. Git / 提交规范

### 5.1 提交消息

```
<type>(<scope>): <subject>

<body>

<footer>
```

Type: `feat` (新功能), `fix` (bug 修复), `docs` (文档), `test` (测试), `refactor` (重构)

### 5.2 示例

```
fix(bd): use cmd.Output() instead of CombinedOutput() to avoid warning noise

- bd create writes warnings to stderr, ID to stdout
- CombinedOutput() mixed both, causing parse failures
- Switched to Output() which captures only stdout
- Added parseBeadsID() pure function with unit tests

Fixes: task ID format issue
```

---

## 6. 相关文档

- [ARCHITECTURE.md](ARCHITECTURE.md) — 系统架构总览
- [DESIGN.md](DESIGN.md) — 设计原则
- [CONSENSUS.md](CONSENSUS.md) — 项目共识
