package proc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/flow"
)

type GateCheckResult struct {
	Type     string `json:"type"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
	Required bool   `json:"required"`
	Auto     bool   `json:"auto"`
	Skipped  bool   `json:"skipped,omitempty"`
}

type GateChecker struct {
	ProjectRoot         string
	Toolchain           ToolchainConfig
	TeamCheckers        map[string]flow.TeamGateChecker
	DocsInternalResolved string
}

type ToolchainConfig struct {
	TestCommand  string
	LintCommand  string
	BuildCommand string
	DocsInternal string
}

func DefaultToolchainConfig() ToolchainConfig {
	return ToolchainConfig{
		DocsInternal: "_docs",
	}
}

func ToolchainConfigFromTeam(tc *flow.TeamToolchain) ToolchainConfig {
	cfg := DefaultToolchainConfig()
	if tc == nil {
		return cfg
	}
	if tc.TestCommand != "" {
		cfg.TestCommand = tc.TestCommand
	}
	if tc.LintCommand != "" {
		cfg.LintCommand = tc.LintCommand
	}
	if tc.BuildCommand != "" {
		cfg.BuildCommand = tc.BuildCommand
	}
	if tc.DocsInternal != "" {
		cfg.DocsInternal = tc.DocsInternal
	}
	return cfg
}

type GateCheckerFunc func(gc *GateChecker, cond GateCondOutput) GateCheckResult

var gateCheckerRegistry = map[flow.GateConditionType]GateCheckerFunc{
	flow.GateCondTestsPass:            func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkTestsPass(cond) },
	flow.GateCondLintPass:             func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkLintPass(cond) },
	flow.GateCondDeliverablesComplete: func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkDeliverablesComplete(cond) },
	flow.GateCondNoRegressions:        func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkNoRegressions(cond) },
	flow.GateCondTaskExists:           func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkTaskExists(cond) },
	flow.GateCondTypeMatches:          func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkTypeMatches(cond) },
	flow.GateCondTraceUpdated:         func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkTraceUpdated(cond) },
}

func RegisterGateChecker(condType flow.GateConditionType, fn GateCheckerFunc) {
	gateCheckerRegistry[condType] = fn
}

func (gc *GateChecker) CheckCondition(cond GateCondOutput) GateCheckResult {
	condType := flow.GateConditionType(cond.Type)

	if fn, ok := gateCheckerRegistry[condType]; ok {
		return fn(gc, cond)
	}

	if checker, ok := gc.TeamCheckers[cond.Type]; ok {
		return gc.checkTeamRegistered(cond, checker)
	}

	if condType == flow.GateCondCustom {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Message:  fmt.Sprintf("custom check requires AI judgment: %s", cond.Check),
			Required: cond.Required,
			Auto:     false,
		}
	}

	return GateCheckResult{
		Type:     cond.Type,
		Passed:   false,
		Message:  fmt.Sprintf("requires AI judgment: %s", cond.Check),
		Required: cond.Required,
		Auto:     false,
	}
}

func (gc *GateChecker) checkTeamRegistered(cond GateCondOutput, checker flow.TeamGateChecker) GateCheckResult {
	switch checker.Type {
	case "script":
		return gc.checkScriptCommand(cond, checker.Command)
	case "ai_judgment":
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Message:  fmt.Sprintf("requires AI judgment: %s", checker.Description),
			Required: cond.Required,
			Auto:     false,
		}
	default:
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Message:  fmt.Sprintf("unknown checker type %q: %s", checker.Type, checker.Description),
			Required: cond.Required,
			Auto:     false,
		}
	}
}

func (gc *GateChecker) checkScriptCommand(cond GateCondOutput, command string) GateCheckResult {
	result := GateCheckResult{
		Type:     cond.Type,
		Required: cond.Required,
		Auto:     true,
	}

	if command == "" {
		result.Passed = false
		result.Skipped = true
		result.Message = fmt.Sprintf("SKIP: no command configured for %s", cond.Type)
		return result
	}

	parts := strings.Fields(command)
	name := parts[0]
	args := parts[1:]

	execCmd := exec.Command(name, args...)
	execCmd.Dir = gc.ProjectRoot
	execCmd.Env = append(os.Environ(), "GOWORK=off")

	output, err := execCmd.CombinedOutput()
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: %s\n%s", command, truncateOutput(string(output), 500))
	} else {
		result.Passed = true
		result.Message = fmt.Sprintf("PASS: %s", command)
	}

	return result
}

func (gc *GateChecker) checkTestsPass(cond GateCondOutput) GateCheckResult {
	cmd := gc.Toolchain.TestCommand
	if cmd == "" {
		return GateCheckResult{
			Type:     "tests_pass",
			Passed:   false,
			Skipped:  true,
			Message:  "SKIP: no test_command configured in team.json",
			Required: cond.Required,
			Auto:     true,
		}
	}

	result := GateCheckResult{
		Type:     "tests_pass",
		Required: cond.Required,
		Auto:     true,
	}

	parts := strings.Fields(cmd)
	name := parts[0]
	args := parts[1:]

	execCmd := exec.Command(name, args...)
	execCmd.Dir = gc.ProjectRoot
	execCmd.Env = append(os.Environ(), "GOWORK=off")

	output, err := execCmd.CombinedOutput()
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: %s\n%s", cmd, truncateOutput(string(output), 500))
	} else {
		result.Passed = true
		result.Message = fmt.Sprintf("PASS: %s", cmd)
	}

	return result
}

func (gc *GateChecker) checkLintPass(cond GateCondOutput) GateCheckResult {
	cmd := gc.Toolchain.LintCommand
	if cmd == "" {
		return GateCheckResult{
			Type:     "lint_pass",
			Passed:   false,
			Skipped:  true,
			Message:  "SKIP: no lint_command configured in team.json",
			Required: cond.Required,
			Auto:     true,
		}
	}

	result := GateCheckResult{
		Type:     "lint_pass",
		Required: cond.Required,
		Auto:     true,
	}

	parts := strings.Fields(cmd)
	name := parts[0]
	args := parts[1:]

	execCmd := exec.Command(name, args...)
	execCmd.Dir = gc.ProjectRoot
	execCmd.Env = append(os.Environ(), "GOWORK=off")

	output, err := execCmd.CombinedOutput()
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: %s\n%s", cmd, truncateOutput(string(output), 500))
	} else {
		result.Passed = true
		result.Message = fmt.Sprintf("PASS: %s", cmd)
	}

	return result
}

func (gc *GateChecker) checkDeliverablesComplete(cond GateCondOutput) GateCheckResult {
	result := GateCheckResult{
		Type:     "deliverables_complete",
		Required: cond.Required,
		Auto:     true,
	}

	docsDir := gc.resolveDocsDir()

	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: docs directory not found: %s", docsDir)
		return result
	}

	deliverables := cond.resolveDeliverables()
	if len(deliverables) == 0 {
		result.Passed = true
		result.Message = "PASS: no deliverables to check"
		return result
	}

	missing := []string{}
	for _, d := range deliverables {
		found := gc.findDeliverable(docsDir, d)
		if !found {
			missing = append(missing, d)
		}
	}

	if len(missing) > 0 {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: missing deliverables: %s", strings.Join(missing, ", "))
	} else {
		result.Passed = true
		result.Message = fmt.Sprintf("PASS: all %d deliverables found", len(deliverables))
	}

	return result
}

func (gc *GateChecker) findDeliverable(docsDir, name string) bool {
	searchPath := filepath.Join(docsDir, name)
	if _, err := os.Stat(searchPath); err == nil {
		return true
	}

	found := false
	filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return nil
		}
		if !info.IsDir() && info.Name() == name {
			found = true
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), name) {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	return found
}

func (gc *GateChecker) checkNoRegressions(cond GateCondOutput) GateCheckResult {
	result := gc.checkTestsPass(cond)
	result.Type = "no_regressions"
	if result.Skipped {
		result.Message = "SKIP: no test_command configured — AI must verify no regressions"
		result.Auto = false
	}
	return result
}

func (gc *GateChecker) checkTaskExists(cond GateCondOutput) GateCheckResult {
	result := GateCheckResult{
		Type:     "task_exists",
		Required: cond.Required,
		Auto:     true,
	}

	if bd.IsAvailable() {
		args := []string{"show", "--current", "--json"}
		output, err := bd.RunQuiet(args...)
		if err != nil || output == "" {
			result.Passed = false
			result.Message = "FAIL: no active task found (bd unavailable or no current task)"
			return result
		}
		result.Passed = true
		result.Message = "PASS: active task found (beads)"
		return result
	}

	taskID := cond.Check
	if taskID == "" {
		taskID = cond.Expected
	}

	if taskID != "" {
		if gc.checkTaskDirExists(taskID) {
			result.Passed = true
			result.Message = fmt.Sprintf("PASS: task directory found for %s", taskID)
			return result
		}

		if gc.checkTaskInPool(taskID) {
			result.Passed = true
			result.Message = fmt.Sprintf("PASS: task found in task-pool (%s)", taskID)
			return result
		}

		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: task %s not found — run 'flow task ready' or create task directory", taskID)
		return result
	}

	if gc.hasAnyActiveTask() {
		result.Passed = true
		result.Message = "PASS: active task found (file-based)"
		return result
	}

	result.Passed = false
	result.Message = "FAIL: no task found — create a task with 'flow task ready' or provide --task <id>"
	return result
}

func (gc *GateChecker) checkTaskDirExists(taskID string) bool {
	docsDir := gc.resolveDocsDir()
	taskDir := filepath.Join(docsDir, "task", taskID)
	if info, err := os.Stat(taskDir); err == nil && info.IsDir() {
		return true
	}
	taskDirAlt := filepath.Join(docsDir, taskID)
	if info, err := os.Stat(taskDirAlt); err == nil && info.IsDir() {
		indexFile := filepath.Join(taskDirAlt, "INDEX.md")
		if _, err := os.Stat(indexFile); err == nil {
			return true
		}
		specFile := filepath.Join(taskDirAlt, "SPEC.md")
		if _, err := os.Stat(specFile); err == nil {
			return true
		}
	}
	return false
}

func (gc *GateChecker) checkTaskInPool(taskID string) bool {
	poolPath := filepath.Join(gc.ProjectRoot, ".team", "task-pool.md")
	data, err := os.ReadFile(poolPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), taskID)
}

func (gc *GateChecker) hasAnyActiveTask() bool {
	docsDir := gc.resolveDocsDir()
	taskBaseDir := filepath.Join(docsDir, "task")
	if entries, err := os.ReadDir(taskBaseDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				return true
			}
		}
	}

	entries, err := os.ReadDir(docsDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		indexPath := filepath.Join(docsDir, e.Name(), "INDEX.md")
		specPath := filepath.Join(docsDir, e.Name(), "SPEC.md")
		if _, err := os.Stat(indexPath); err == nil {
			return true
		}
		if _, err := os.Stat(specPath); err == nil {
			return true
		}
	}
	return false
}

func (gc *GateChecker) checkTypeMatches(cond GateCondOutput) GateCheckResult {
	result := GateCheckResult{
		Type:     "type_matches",
		Required: cond.Required,
		Auto:     true,
	}

	if cond.Expected != "" && cond.Check != "" {
		if cond.Check == cond.Expected {
			result.Passed = true
			result.Message = fmt.Sprintf("PASS: type %q matches expected %q", cond.Check, cond.Expected)
		} else {
			result.Passed = false
			result.Message = fmt.Sprintf("FAIL: type %q does not match expected %q", cond.Check, cond.Expected)
		}
		return result
	}

	result.Passed = false
	result.Message = "SKIP: type_matches requires both check and expected values — AI must verify"
	result.Auto = false
	return result
}

func (gc *GateChecker) resolveDocsDir() string {
	if gc.DocsInternalResolved != "" {
		if filepath.IsAbs(gc.DocsInternalResolved) {
			return gc.DocsInternalResolved
		}
		return filepath.Join(gc.ProjectRoot, gc.DocsInternalResolved)
	}
	return filepath.Join(gc.ProjectRoot, gc.Toolchain.DocsInternal)
}

func (gc *GateChecker) checkTraceUpdated(cond GateCondOutput) GateCheckResult {
	result := GateCheckResult{
		Type:     "trace_updated",
		Required: true,
		Auto:     true,
	}

	docsDir := gc.resolveDocsDir()
	taskID := cond.Check
	if taskID == "" {
		taskID = cond.Expected
	}
	if taskID == "" {
		result.Passed = false
		result.Message = "FAIL: trace_updated requires task_id (set in check or expected field)"
		return result
	}

	tracePath := filepath.Join(docsDir, "task", taskID, "TRACE.md")

	data, err := os.ReadFile(tracePath)
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: TRACE.md not found at %s — run 'flow task trace %s' to create", tracePath, taskID)
		return result
	}

	content := string(data)
	entryCount := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "## T") {
			entryCount++
		}
	}

	if entryCount == 0 {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: TRACE.md exists but has no entries — run 'flow task trace %s --decision <text>'", taskID)
		return result
	}

	result.Passed = true
	result.Message = fmt.Sprintf("PASS: TRACE.md has %d decision entries", entryCount)
	return result
}

func (c GateCondOutput) resolveDeliverables() []string {
	if len(c.Deliverables) > 0 {
		return c.Deliverables
	}
	if c.Check == "" {
		return nil
	}
	parts := strings.Split(c.Check, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func RunGateCheck(projectRoot string, conditions []GateCondOutput, team *flow.TeamDefinition, docsInternalResolved string) []GateCheckResult {
	var tc ToolchainConfig
	var teamCheckers map[string]flow.TeamGateChecker

	if team != nil {
		tc = ToolchainConfigFromTeam(team.Toolchain)
		teamCheckers = team.GateCheckers
	} else {
		tc = DefaultToolchainConfig()
	}

	if teamCheckers == nil {
		teamCheckers = make(map[string]flow.TeamGateChecker)
	}

	checker := &GateChecker{
		ProjectRoot:          projectRoot,
		Toolchain:            tc,
		TeamCheckers:         teamCheckers,
		DocsInternalResolved: docsInternalResolved,
	}

	results := make([]GateCheckResult, 0, len(conditions))
	for _, cond := range conditions {
		result := checker.CheckCondition(cond)
		results = append(results, result)
	}

	return results
}

func GateOverallResult(results []GateCheckResult) (bool, string) {
	if len(results) == 0 {
		return true, "no conditions to check"
	}

	allPassed := true
	requiredPassed := true
	var failedRequired []string
	var failedAdvisory []string
	var skipped []string

	for _, r := range results {
		if r.Skipped {
			skipped = append(skipped, r.Type)
			continue
		}
		if !r.Passed {
			allPassed = false
			if r.Required {
				requiredPassed = false
				failedRequired = append(failedRequired, r.Type)
			} else {
				failedAdvisory = append(failedAdvisory, r.Type)
			}
		}
	}

	if requiredPassed && allPassed {
		if len(skipped) > 0 {
			return true, fmt.Sprintf("all checked conditions passed (skipped: %s)", strings.Join(skipped, ", "))
		}
		return true, "all conditions passed"
	}
	if requiredPassed && !allPassed {
		return true, fmt.Sprintf("passed (advisory failures: %s)", strings.Join(failedAdvisory, ", "))
	}
	return false, fmt.Sprintf("failed (required failures: %s)", strings.Join(failedRequired, ", "))
}

func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
