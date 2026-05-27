# ADR-P4-1: Engine-Level Extensibility — No Hardcoded Team-Specific Logic

> **Status**: Approved | **Date**: 2026-05-25 | **Author**: team-flow | **Reviewer**: Pending

## 1. Summary

引擎层（internal/flow/, internal/proc/）禁止硬编码任何团队特定逻辑。所有团队差异通过 team.json 配置注入，引擎只提供机制不提供策略。本次 ADR 由 GateChecker 硬编码 `go test` 的缺陷触发。

## 2. Problem Statement

### 2.1 缺陷

GateChecker 实现中硬编码了 Go 工具链：

```go
func DefaultToolchainConfig() ToolchainConfig {
    return ToolchainConfig{
        TestCommand:  "go test ./...",           // ❌ 只有 Go
        LintCommand:  "golangci-lint run ./...",  // ❌ 只有 Go
        DocsInternal: "_docs",                    // ✅ 通用
    }
}
```

GateConditionType 硬编码了软件开发视角的条件类型：

```go
GateCondTestsPass           // ❌ 小说/游戏不跑 go test
GateCondLintPass            // ❌ 小说不 lint，游戏用不同 linter
GateCondNoRegressions       // ❌ 小说无回归概念
```

### 2.2 影响范围

系统当前有 5 个团队类型：

| Team | Domain | tests_pass | lint_pass | 实际需要的 gate 条件 |
|------|--------|------------|-----------|---------------------|
| dev-team | Go 开发 | ✅ go test | ✅ golangci-lint | tests_pass, lint_pass, deliverables_complete |
| content-team | 内容创作 | ❌ 无测试 | ❌ 无 lint | word_count_met, pov_consistent, outline_adherence |
| game-team | 游戏设计 | ⚠️ Unity/Unreal 测试 | ⚠️ 不同 linter | build_passes, performance_budget_met, playtest_feedback |
| trading-team | 交易系统 | ⚠️ 可能 Python | ⚠️ 不同 linter | backtest_passes, risk_limits_met |
| skill-team | 技能开发 | ✅ go test | ✅ golangci-lint | tests_pass, lint_pass |

### 2.3 根因分析

**为什么引入了这个缺陷？**

1. **只看当前使用者**：实现 GateChecker 时只考虑了 dev-team 的需求，没有扫描 v3/teams/ 下所有团队类型
2. **跳过设计审查**：从需求直接跳到实现，没有"这个设计对所有团队是否成立"的验证步骤
3. **没有 ADR**：引擎级变更没有 Architecture Decision Record

**这不是第一次**：ADR-P3-1 之前（角色定义在 flow 里重复）、validator 假阳性（不加载 team.json），都是同一个模式——只看当前使用者，不看系统全貌。

## 3. Decision

### 3.1 核心原则：引擎只提供机制，不提供策略

```
引擎层（机制）           → 接口 + 注册表 + 默认空实现
团队配置层（策略）        → team.json 中的 toolchain + gate_types
流程定义层（组合）        → flow JSON 引用团队定义的 gate 条件
```

### 3.2 具体设计

#### 3.2.1 ToolchainConfig 从 team.json 读取

team.json 新增 `toolchain` 字段：

```json
{
  "id": "dev-team",
  "toolchain": {
    "test_command": "go test ./...",
    "lint_command": "golangci-lint run ./...",
    "build_command": "go build ./...",
    "docs_internal": "_docs"
  }
}
```

```json
{
  "id": "content-team",
  "toolchain": {
    "test_command": "",
    "lint_command": "",
    "build_command": "",
    "docs_internal": "_docs"
  }
}
```

```json
{
  "id": "game-team",
  "toolchain": {
    "test_command": "dotnet test",
    "lint_command": "",
    "build_command": "unity -batchmode -executeMethod Build",
    "docs_internal": "_docs"
  }
}
```

引擎行为：
- `test_command` 为空 → `tests_pass` 条件自动跳过（标记 SKIP，非 FAIL）
- `lint_command` 为空 → `lint_pass` 条件自动跳过
- 从 team.json 读取，不硬编码

#### 3.2.2 GateConditionType 注册表

替换当前的 switch-case 硬编码为注册表模式：

```go
// 引擎内置注册表（可扩展）
var gateCheckerRegistry = map[GateConditionType]GateCheckerFunc{
    GateCondTestsPass:           checkTestsPass,
    GateCondLintPass:            checkLintPass,
    GateCondDeliverablesComplete: checkDeliverablesComplete,
    GateCondNoRegressions:       checkNoRegressions,
    GateCondTaskExists:          checkTaskExists,
    GateCondTypeMatches:         checkTypeMatches,
}

// team.json 可注册自定义 checker
type GateCheckerFunc func(gc *GateChecker, cond GateCondOutput) GateCheckResult

func RegisterGateChecker(condType GateConditionType, fn GateCheckerFunc) {
    gateCheckerRegistry[condType] = fn
}
```

team.json 新增 `gate_checkers` 字段：

```json
{
  "id": "content-team",
  "gate_checkers": {
    "word_count_met": {
      "type": "script",
      "command": "python scripts/check_word_count.py",
      "description": "Verify word count meets chapter target"
    },
    "pov_consistent": {
      "type": "ai_judgment",
      "description": "Verify POV consistency within each scene"
    }
  }
}
```

#### 3.2.3 引擎级变更审查规则

在 AGENTS.md 中新增：

```
⛔ ENGINE-LEVEL CHANGE PROTOCOL

任何影响 internal/flow/ 或 internal/proc/ 的变更，必须：

1. 扫描 v3/teams/ 下所有 team.json
2. 对每个团队类型验证设计是否成立
3. 写 ADR 记录设计决策和跨团队兼容性分析
4. 通过审查后才能开始实现

违反此协议 = 跳过流程 = 产出不可验证
```

## 4. Cross-Team Compatibility Analysis

### 4.1 本次变更对每个团队的影响

| Team | 影响 | 修复前 | 修复后 |
|------|------|--------|--------|
| dev-team | 无负面影响 | go test 硬编码可用 | 从 team.json 读取，行为等价 |
| content-team | gate 自动化从不可用到可用 | tests_pass/lint_pass 报 FAIL | 空命令自动 SKIP，自定义条件走注册表 |
| game-team | gate 自动化从不可用到可用 | tests_pass/lint_pass 报 FAIL | 配置 dotnet test，自定义条件走注册表 |
| trading-team | gate 自动化从不可用到可用 | 同上 | 配置对应命令，自定义条件走注册表 |
| skill-team | 无负面影响 | go test 硬编码可用 | 从 team.json 读取，行为等价 |

### 4.2 向后兼容性

- dev-team 的 team.json 添加 `toolchain` 字段后，行为与硬编码完全等价
- 不添加 `toolchain` 字段的团队，回退到 DefaultToolchainConfig（保持向后兼容）
- 新的 gate 条件类型通过注册表添加，不影响已有流程

## 5. Consequences

### 5.1 正面

- 引擎不再绑定任何特定技术栈
- 新团队类型只需配置 team.json，无需改引擎代码
- 自定义 gate 条件可扩展

### 5.2 负面

- team.json 结构变复杂（新增 toolchain + gate_checkers 字段）
- GateChecker 注册表增加少量运行时开销（可忽略）

### 5.3 风险

- team.json 配置错误可能导致 gate 检查静默跳过 → 需要在 formatter 中明确显示 SKIP 状态
