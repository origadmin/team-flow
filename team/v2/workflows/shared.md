# Shared Workflow — _team v2

> **Version**: v2.0 | **Applies to**: All roles
> **Core principle**: beads is the single source of truth

## Layer 0: Universal Rules

These rules apply to ALL roles in ALL sessions.

### L0.1 Source of Truth

```
beads (.beads/) = Task state
task-pool-export.md = Human-readable view (read-only)
Project files = Implementation artifacts
```

**Never**:
- ❌ Edit task-pool-export.md to change task state
- ❌ Assume file state matches beads state (always query beads)
- ❌ Skip Dolt sync when working with multiple agents

### L0.2 Status Gates

Before any state transition, verify prerequisites:

| From | To | Gate |
|------|----|----|
| open | in_progress | Check no blockers: `bd ready <id>` |
| in_progress | closed | Verify tests pass, notes complete |
| open | blocked | Document blocking issue in notes |
| blocked | open | Confirm blocker resolved |

```bash
# Gate check before starting work
bd ready <id> && bd update <id> --claim
```

### L0.3 Session Protocol

#### Start of Session

```bash
# 1. Check beads status
bd stats

# 2. Pull latest changes (multi-agent)
bd dolt pull

# 3. Check your assigned issues
bd list --assignee $USER --status in_progress

# 4. Check ready issues
bd ready --json
```

#### During Session

```bash
# Before starting work
bd update <id> --claim

# Regular progress updates
bd update <id> --notes "COMPLETED: X IN PROGRESS: Y"

# After completion
bd close <id> --reason "Fixed: [details]"
```

#### End of Session

```bash
# 1. Export current state
bd list --status open --format table > .team/task-pool-export.md

# 2. Commit changes
bd dolt commit -m "Session: $(date +%Y%m%d_%H%M)"

# 3. Push to remote
bd dolt push
```

## Layer 1: Role Handoff Protocol

### L1.1 Standard Handoff Format

When handing off to another role:

```markdown
## Handoff: <issue-id> → <target-role>

**From**: <your-role>
**To**: <target-role>
**Timestamp**: YYYY-MM-DD HH:MM

**Summary**: <1-2 sentences of what was done>

**Deliverables**:
- [ ] <artifact-1> at <path>
- [ ] <artifact-2> at <path>

**Next Actions**:
1. <action-1>
2. <action-2>

**Blockers**: <none | list>

**Context**: <files modified, decisions made>
```

### L1.2 Handoff via beads

```bash
# Update issue with handoff notes
bd update <id> \
  --notes "## Handoff → <target-role>

Summary: [details]
Deliverables: [list]
Next: [actions]" \
  --assignee "<target-role>" \
  --add-label phase:<next-phase>

# Example: Dev → QA
bd update cms-xxx \
  --notes "## Handoff → qa

Summary: Implemented X feature
Deliverables: src/api/x.go, tests/x_test.go
Next: Verify X works with real data" \
  --assignee "qa-engineer" \
  --add-label phase:verify \
  --remove-label phase:implement
```

### L1.3 Handoff Validation

Receiving role must validate:

```bash
# 1. Check issue state
bd show <id>

# 2. Verify deliverables exist
ls -la <artifact-path>

# 3. Verify no blockers
bd ready <id>

# 4. Acknowledge handoff
bd update <id> --notes "Handoff acknowledged. Starting work."
```

## Layer 2: Phase Transitions

### L2.1 Phase Gate Requirements

| Phase | Entry Requirements | Exit Criteria |
|-------|-------------------|---------------|
| ready | Issue created | Assigned to role |
| analyze | Assigned | Root cause documented |
| design | Analysis complete | Spec written |
| implement | Design approved | Tests pass |
| verify | Implementation done | QA checklist complete |
| review | QA approved | User confirmation |

### L2.2 Phase Transition Command

```bash
# Standard phase transition
bd update <id> \
  --add-label phase:<new-phase> \
  --remove-label phase:<old-phase> \
  --notes "## Phase: <old> → <new>

Reason: <why transitioning>
Deliverables: <what was completed>
Next: <what needs to happen>"
```

### L2.3 Blocked Phase Handling

If blocked during a phase:

```bash
# Mark as blocked
bd update <id> \
  --status blocked \
  --notes "## BLOCKED in <phase>

Blocker: <issue-id or description>
Impact: <what can't proceed>
Workaround: <if any, or 'none'>"

# Link blocking dependency
bd dep add <id> <blocker-id> --type blocks
```

## Layer 3: Multi-Agent Coordination

### L3.1 Concurrency Control

For multi-agent environments, use Dolt's git-native collaboration:

```bash
# Enable auto-commit for immediate sync
bd config set dolt.auto-commit on

# Or use batch mode for performance
bd config set dolt.auto-commit batch
# Remember to commit manually:
bd dolt commit -m "Batch update"
bd dolt push
```

### L3.2 Conflict Resolution

Dolt handles merge conflicts automatically for most operations. If conflict occurs:

```bash
# Check Dolt status
bd dolt status

# View conflicts
bd dolt conflicts cat <table>

# Resolve using theirs (or ours)
bd dolt checkout --theirs <table>
bd dolt add <table>
bd dolt commit -m "Resolve conflict"
```

### L3.3 Session Isolation

Each agent session should:
1. Claim issues atomically (`--claim`)
2. Push changes frequently
3. Pull before claiming new issues

```bash
# Atomic claim
bd update <id> --claim  # Fails if already claimed by another

# Frequent sync
bd dolt pull
```

## Documentation Requirements

### Doc Storage Paths

| Type | Path | Example |
|------|------|---------|
| Spec/Design | `_docs/orig-cms/requirements/{ID}/` | `_docs/.../F014/` |
| RCA Reports | `_docs/orig-cms/reports/errors/` | `_docs/.../errors/E0001-B061-01.md` |
| Meeting Notes | `_docs/orig-cms/meetings/` | `_docs/.../meetings/2026-05-02.md` |
| Architecture | `_docs/orig-cms/architecture/` | `_docs/.../architecture/auth.md` |

### Doc Metadata in beads

```bash
bd update <id> --set-metadata doc_path="_docs/orig-cms/requirements/F014/"
```

## Quality Gates

### Before Closing Issue

```bash
# Checklist
bd show <id>  # Verify all fields populated
bd children <id>  # Verify sub-tasks closed
bd dep list <id>  # Verify dependencies resolved
git status  # Verify code committed
go test ./...  # Verify tests pass
```

### Close with Summary

```bash
bd close <id> \
  --reason "Completed. Summary:
- Changes: [list of files modified]
- Tests: [test coverage or specific tests]
- Impact: [affected areas]
- Commit: [SHA or PR link]"
```

## Anti-Patterns

❌ **DON'T**:
- Skip Dolt pull/push in multi-agent setups
- Leave issues in `phase:ready` unassigned
- Close issues without recording outcome
- Create duplicate issues (check first!)
- Dump deliverable content (root cause analysis, design decisions, test results) into beads notes — **write to independent deliverable files**
- Add "Task Details" / "Deliverable Tracking" sections to task-pool-export.md
- Modify `_team/` rules to solve project-specific problems — use `.team/project.md §CONSTRAINTS` and `_docs/.../lessons/`

### Deliverable Write Separation (v1 Lesson)

> **v1 教训**: AI 把所有任务详情写入 task-pool.md 而非独立成果物文件，导致 75 行状态表膨胀到 1822 行。v2 中 `bd update --notes` 是同样的风险点。

**beads notes 用途**: 只记录进度摘要和交接信息
```
✅ bd update <id> --notes "COMPLETED: RCA.md written, fix applied, tests passing"
❌ bd update <id> --notes "Root cause: handler.go wraps response with {code,message,data}..."
```

**成果物写入规则**:

| 内容类型 | 正确位置 | 禁止位置 |
|---------|---------|---------|
| 根因分析 | `{DOCS_INTERNAL}/reports/bugs/B{NNN}/R{n}/RCA.md` | beads notes |
| 复现测试 | `{DOCS_INTERNAL}/reports/bugs/B{NNN}/R{n}/TEST_CASE.md` | beads notes |
| 变更报告 | `{DOCS_INTERNAL}/reports/changes/C{NNN}-R{n}/SCOPE.md` | beads notes |
| 需求规格 | `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/SPEC.md` | beads notes |
| 设计决策 | `{DOCS_INTERNAL}/requirements/F{NNN}-{name}/R1-R5.md` | beads notes |

✅ **DO**:
- Use `--json` for scriptable output
- Add phase labels to all issues
- Record progress notes regularly
- Hand off explicitly via assignment
- Export at session end for visibility
