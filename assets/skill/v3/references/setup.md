# First-Time Setup

v3 有三种初次使用场景：

## Scenario A: 全新项目（CLI 初始化）

```bash
flow init --v3 [--team {team-id}] [--flow {flow-name}]
```

`flow init --v3` 会：
1. 创建 `.team/` 目录和配置文件
2. 展示预设团队菜单（5个预设团队），用户选择最接近的
3. 不合适选"0: 创建新团队"，AI 用 create skill 从零设计
4. 也可以 `--team dev-team` 直接指定团队
5. 或 `--flow dev-flow` 通过 flow 名找到对应团队
6. 安装选中团队的所有 flows 到 `.team/flows/`
7. 初始化 beads 数据库
8. 在 project.yaml 中设置 `active_flow`

## Scenario B: AI 会话中发现无流程绑定

当 AI 执行 Session Startup Protocol 时，发现 `active_flow` 为空或指向不存在的流程：

```
Step 1: 运行 flow proc list
  → 查看所有已注册的团队

Step 2: 判断可用团队
  → 如果有预设团队 → 向用户展示选项，引导选择
  → 如果没有团队 → 直接进入创建流程

Step 3: 用户选择
  → 用户选了某个团队 → 更新 project.yaml 的 active_flow → 重新执行 flow proc run
  → 用户要创建新团队 → 加载 team-flow-v3-create skill → 创建完成后注册并设置 active_flow
  → 用户想先看看 → flow proc show {name} 展示详情

Step 4: 设置 active_flow
  → 编辑 .team/project.yaml，设置 active_flow: {chosen-flow}
  → 重新执行 flow proc run，进入正常执行流程
```

**引导话术示例**：
> 这个项目还没有绑定团队流程。我找到了以下可用团队：
> 1. dev-flow — 软件开发全流程
> 2. novel-flow — 小说创作流程
> 3. ...
>
> 选择一个最接近你需求的，或者告诉我你的具体场景，我帮你创建一个专属团队。

## Scenario C: v2 项目升级到 v3

```bash
flow migrate v3 [--flow {name}]
```

`flow migrate v3` 会：
1. 备份 v2 配置
2. 安装 v3 技能到 IDE skill 目录
3. 安装预设流程到 `.team/flows/` 并注册到 project.yaml
4. 更新 IDE bridge 文件指向 v3
5. 更新 `.team/version` 为 v3
6. 设置 `active_flow`

## 团队放置规则

| 来源 | 放置路径 | 说明 |
|------|----------|------|
| 框架预设（团队模板） | `assets/orgs/{team-id}/flows/` | 随工具嵌入，只读 |
| 项目安装 | `.team/flows/` | `flow init --v3` 安装到这里 |
| 项目自定义 | `.team/flows/` | 用户创建的放这里 |
| 解析优先级 | `.team/flows/` | 项目目录优先 |

## 预设团队模板

| 团队 | ID | Flows | 适用场景 |
|------|----|-------|----------|
| 软件开发团队 | `dev-team` | 8 (dev/feature/bugfix/hotfix/change/release/analysis/batch) | 通用软件开发 |
| 内容创作团队 | `content-team` | 3 (content-distribution/novel/promo-video) | 文章、视频、小说 |
| 游戏设计团队 | `game-team` | 1 (game-design) | 游戏设计 |
| 交易分析团队 | `trading-team` | 1 (trading) | 交易分析 |
| 技能开发团队 | `skill-team` | 2 (skill-dev/test-simple) | AI 技能开发 |

## GitHub Issue Sync

Sync GitHub Issues to local tasks before triage analysis:

```bash
flow task sync --source github --repo owner/repo   # Sync open issues
flow task sync --source github --label bug          # Sync only bugs (auto-detect repo)
flow task sync --source github --dry-run            # Preview without writing
flow task sync --source github --overwrite          # Re-sync existing tasks
```

Issue → Task mapping: `[GH#42]` prefix, auto-classify type/priority from labels, `--external-ref` for dedup.
See `flow task sync --help` for full options.
