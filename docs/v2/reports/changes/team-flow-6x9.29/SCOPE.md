# SCOPE.md — Machine-Readable Change Report
# Encoding: UTF-8
# Parse: Each section is a YAML block, separated by ---

task_id: "team-flow-6x9.29"
beads_id: "team-flow-6x9.29"
task_type: "change"
iteration: "R1"
date: "2026-05-10"
author: "Triage"

files:
  added:
  modified:
    - path: "internal/boot/init.go"
      description: "Add isInteractive() function, auto-skip npx in non-interactive mode, add 2-minute timeout for npx command"
  deleted:

test_changes:
  new_tests:
  modified_tests:

pipeline:
  fmt: "pass"
  lint: "pass"
  build: "pass"
  test: "pass"
  test_coverage: "skip"

chinese_check: "pass"
chinese_violations: 0

summary: |
  Fixed flow init hanging in non-interactive environments: added isInteractive() detection, auto-use embedded files, added 2-minute timeout for npx. No CI/CD hangs anymore.

---
# Above this line: machine-readable (YAML)
# Below this line: human notes (optional)

## Change Details

### Problem
`flow init` could hang indefinitely in:
1. Non-interactive environments (CI/CD, pipes, scripts)
2. When npx network is slow or unavailable
3. When --yes flag is not provided

### Solution
1. **isInteractive() function**: detects if stdin is a tty
2. **Auto-skip npx**: in non-interactive mode, directly use embedded files
3. **2-minute timeout**: for npx command in interactive mode, auto-fallback after timeout

### Code Changes
```go
// Added function
func isInteractive() bool {
  fileInfo, _ := os.Stdin.Stat()
  return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Added context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

// Updated installation flow
```

### Git Reference
- Commit: `e2be886` (team-flow subrepo)
