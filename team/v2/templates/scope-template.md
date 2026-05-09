# SCOPE.md 模板

> Machine-Readable Change Report

```yaml
# SCOPE.md — Machine-Readable Change Report
# Encoding: UTF-8
# Parse: Each section is a YAML block, separated by ---

task_id: "{TASK_ID}"
beads_id: "<beads-id>"
task_type: "feature|bugfix|change|analysis"
iteration: "R{N}"
date: "{YYYY-MM-DD}"
author: "{role}"

files:
  added:
    - path: "{relative/path/to/file}"
      description: "{one-line purpose}"
  modified:
    - path: "{relative/path/to/file}"
      description: "{what changed}"
  deleted:
    - path: "{relative/path/to/file}"
      description: "{why removed}"

test_changes:
  new_tests:
    - path: "{test_file_path}"
      cases:
        - "{TestName}: {what it verifies}"
  modified_tests:
    - path: "{test_file_path}"
      cases:
        - "{TestName}: {what changed}"

pipeline:
  fmt: "pass|fail|skip"
  lint: "pass|fail|skip"
  build: "pass|fail|skip"
  test: "pass|fail|skip"
  test_coverage: "{XX}%"

chinese_check: "pass|fail"
chinese_violations: 0

summary: |
  {One paragraph: what was done, why, impact scope}

---
# Above this line: machine-readable (YAML)
# Below this line: human notes (optional)
```

**规则**:
- `---` 分隔线上方为 YAML，工具可直接 parse
- `beads_id` 关联 beads issue（v2 新增）
- `files` 列出所有变更文件（added/modified/deleted）
- `pipeline` 记录管线各步骤结果，任一为 fail 则不允许完成
- `chinese_check` + `chinese_violations` 记录中文检查结果
- `---` 下方可写人读备注，工具读取时忽略

## beads 状态更新

生成 SCOPE.md 后执行：

```bash
flow task update <issue-id> --notes "SCOPE: {summary}"
flow task update <issue-id> --add-label phase:review --remove-label phase:verify
```
