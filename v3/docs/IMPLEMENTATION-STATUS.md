# team-flow v3 实施状态

> 最后更新：2026-05-27
> 基于 ARCH-REDESIGN-v3.html 五阶段计划

---

## 已完成

### Phase 0: 目录结构统一 ✅

| 项目 | 状态 | 说明 |
|------|------|------|
| team/ → assets/skill/ | ✅ | git mv, 227 files |
| teams/ → assets/orgs/ | ✅ | git mv, 20 files |
| v3/flows/ → assets/flows/ | ✅ | git mv |
| v3/schema/ → assets/schema/ | ✅ | git mv |
| v3/teams/ 删除 | ✅ | 合并到 assets/orgs/ |
| v3/flows_backup/ 删除 | ✅ | 12个旧文件清理 |
| skillfs.go 统一 embed | ✅ | `//go:embed all:assets` + 4个路径常量 |
| package.json files 更新 | ✅ | `["assets/"]` |
| Go 代码路径引用更新 | ✅ | 6个文件, ~30处替换 |
| SKILL.md HOT/COLD 拆分 | ✅ | 302→107 行 (~65% token 节约) |
| CONSENSUS.md 整合更新 | ✅ | D1-D10 + 路径修正 |

### Gate 修复 ✅

| 项目 | 状态 | 说明 |
|------|------|------|
| Gate bypass 漏洞修复 | ✅ | 前置 gate 检查 |
| Flow 状态隔离 | ✅ | 切换 flow = 重置 gate history |
| 文件后端 gate_checker | ✅ | 无 beads 依赖 |
| --gate 默认值改为 true | ✅ | cmd.go |

---

## 待实施

### Phase 1: 代码重构（internal/ 三层分离）

| 项目 | 状态 | 优先级 | 影响文件 |
|------|------|--------|---------|
| bd/ + beads/ → tools/builtin/ | ⬜ | P0 | 10个文件, 18处调用 |
| tools/ + toolrunner/ + toolchain/ → tools/ | ⬜ | P1 | 3个包合并 |
| ver/ + version/ → core/version/ | ⬜ | P1 | 2个包合并 |
| update/ + updater/ → cmd/ + core/ | ⬜ | P1 | 2个包合并 |
| config/ + configcmd/ → core/config/ + cmd/ | ⬜ | P1 | 2个包合并 |
| proc/ 拆分 | ⬜ | P1 | 8个文件 |
| project.md/yaml 双配置统一 | ⬜ | P0 | 11+8个文件, 87处引用 |
| 模板版本号 2.0.0 → 3.0.0 | ⬜ | P0 | 10个 team.json |
| main.go initTools() 清理 | ⬜ | P1 | 1个文件 |
| main.go 描述更新 | ⬜ | P2 | 1个文件 |

### Phase 2: 配置驱动路由

| 项目 | 状态 | 优先级 | 影响文件 |
|------|------|--------|---------|
| tools.yaml 新格式设计 | ⬜ | P0 | config/default-tools.yaml |
| ToolRouter 实现 | ⬜ | P0 | internal/tools/ |
| flow tools list/run/path 命令 | ⬜ | P0 | 新增 |
| flow task 转发改造 | ⬜ | P0 | internal/task/ |
| flow sync 转发改造 | ⬜ | P1 | internal/sync/ |
| gate_checker 路径配置化 | ⬜ | P1 | internal/proc/gate_checker.go |
| beads CLI 命令隐藏 | ⬜ | P2 | internal/beads/ |

### Phase 3: 两目录内容完善

| 项目 | 状态 | 优先级 |
|------|------|--------|
| team.json → org.json 重命名 | ⬜ | P1 |
| main_flows[] + sub_flows[] 结构 | ⬜ | P1 |
| flow org * 命令族 | ⬜ | P1 |
| 验证链（角色/技能/流程） | ⬜ | P1 |

### Phase 4: 工具化编辑

| 项目 | 状态 | 优先级 |
|------|------|--------|
| flow org create | ⬜ | P1 |
| flow org add-role | ⬜ | P1 |
| flow org add-flow | ⬜ | P1 |
| flow org validate | ⬜ | P1 |
| flow org show | ⬜ | P2 |

### Phase 5: 追踪 + 分发

| 项目 | 状态 | 优先级 |
|------|------|--------|
| Session Log 机制 | ⬜ | P2 |
| Context Chain | ⬜ | P2 |
| 诊断链 D1-D7 | ⬜ | P2 |
| 多技能分发（install/migrate/repair/exec） | ⬜ | P2 |
| 讨论一致性检查 | ⬜ | P2 |

---

## 代码不一致追踪

| ID | 类别 | 优先级 | 影响处数 | 状态 |
|----|------|--------|---------|------|
| I-1 | beads 直接暴露 | P0 | 18+10 | ⬜ |
| I-2 | 硬编码 beads 路径 | P0 | 38 | ⬜ |
| I-3 | project.md/yaml 双配置 | P0 | 66+21 | ⬜ |
| I-4 | 模板版本号 2.0.0 | P0 | 10 | ⬜ |
| I-5 | 6对重复包 | P1 | 30 | ⬜ |
| I-6 | gate_checker 硬编码 | P1 | ~10 | ⬜ |
| I-7 | initTools() 未使用 | P1 | 20行 | ⬜ |
| I-8 | default-tools.yaml 旧格式 | P1 | 全文 | ⬜ |
| I-9 | main.go 描述过时 | P2 | 2行 | ⬜ |
| I-10 | beads 命令暴露 | P2 | 全文 | ⬜ |
