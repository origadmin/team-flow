# 快速提取公式模板

> 当 AI 反复犯同一个错误时，用这个模板 2 分钟生成规则

---

## 错误描述（用户填写）

| 项目 | 内容 |
|------|------|
| 错误现象 | AI 反复 _____________ |
| 发生场景 | 当 _____________ 时 |
| 正确做法 | 应该 _____________ |
| 发生次数 | 第 __ 次 |
| 涉及角色 | Dev / QA / Tech Lead / PM |

---

## 一键生成规则（复制到 lessons 文件）

```markdown
### Rule {N}: {简短标题}

**Problem**: {错误现象}

**Rule**: 
1. {强制规则 1}
2. {强制规则 2}
3. {强制规则 3}

**Correct**:
- {正确做法 1}
- {正确做法 2}

**Wrong Example**:
```
❌ {错误示例}
```

**Correct Example**:
```
✅ {正确示例}
```

**Applies When**: {适用场景}

**Source**: Error Report {E000XX}
```

---

## 示例：AI 反复用 npm 而不是 bun

### 填写模板

| 项目 | 内容 |
|------|------|
| 错误现象 | 用 `npm install` 而不是 `bun add` |
| 发生场景 | 安装依赖时 |
| 正确做法 | 先读 project.md，用 Toolchain 定义的命令 |
| 发生次数 | 第 3 次 |
| 涉及角色 | Dev |

### 生成规则

```markdown
### Rule 1: Toolchain Gate — 包管理器强制检查

**Problem**: AI 直接使用 npm/pnpm/yarn，未读取 project.md Toolchain 配置。

**Rule**: 
1. 执行包管理命令前必须先读取 `.team/project.md` 的 `Toolchain` 部分
2. 使用 `Toolchain.package_manager` 定义的命令
3. 禁止直接使用 npm/pnpm/yarn

**Correct**:
- 先读取 project.md → 提取 Toolchain.package_manager → 使用该命令

**Wrong Example**:
```bash
# ❌ 错误：未读取配置直接使用 npm
npm install lodash
```

**Correct Example**:
```bash
# ✅ 正确：先读取配置，再执行命令
bun add lodash
```

**Applies When**: 任何需要执行包管理命令的场景

**Source**: Error Report E00001
```

---

## 快速命令

```bash
# 1. 记录错误报告（轻量）
echo "# Error Report: E000XX-..." > reports/errors/E000XX-XXX-001.md

# 2. 添加到 lessons 文件
cat >> lessons/dev-common.md << 'EOF'
### Rule X: ...
...
EOF

# 3. 确保 prompt 加载 lessons
# 检查 prompts/dev.md 的 standards 部分是否包含 lessons/dev-common.md
```

---

## 关键原则

| 原则 | 说明 |
|------|------|
| **预读 > 重试** | 出错后再读规则是浪费，执行前预读更高效 |
| **强制 > 建议** | "应该"没用，必须变成门禁检查 |
| **具体 > 抽象** | 给出 exact command，不要模糊描述 |
| **分层 > 集中** | 按角色+类型拆分，避免文件膨胀 |

---

## 下一步

1. 把这个规则写入 `lessons/dev-common.md`
2. 确保 `prompts/dev.md` 的 standards 包含 `lessons/dev-common.md`
3. 下次 Dev 角色被触发时，自动预读规则
4. 错误不再发生
