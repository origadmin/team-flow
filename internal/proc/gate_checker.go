package proc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/eventlog"
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
	ProjectRoot          string
	Toolchain            ToolchainConfig
	TeamCheckers         map[string]flow.TeamGateChecker
	DocsInternalResolved string
	FlowName             string
	ExplicitlyTargeted   bool
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
	flow.GateCondHasActiveTasks:       func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkHasActiveTasks(cond) },
	flow.GateCondHasSessionHistory:    func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkHasSessionHistory(cond) },
}

func RegisterGateChecker(condType flow.GateConditionType, fn GateCheckerFunc) {
	gateCheckerRegistry[condType] = fn
}

func (gc *GateChecker) CheckCondition(cond GateCondOutput) GateCheckResult {
	condType := flow.GateConditionType(cond.Type)

	var result GateCheckResult

	if fn, ok := gateCheckerRegistry[condType]; ok {
		result = fn(gc, cond)
	} else if checker, ok := gc.TeamCheckers[cond.Type]; ok {
		result = gc.checkTeamRegistered(cond, checker)
	} else if gc.ExplicitlyTargeted {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   true,
			Message:  fmt.Sprintf("AI confirmed (explicitly targeted): %s", cond.Check),
			Required: cond.Required,
			Auto:     false,
		}
	} else if condType == flow.GateCondCustom {
		nodeRef := cond.NodeName
		if nodeRef == "" {
			nodeRef = cond.NodeID
		}
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Message:  fmt.Sprintf("custom check requires AI judgment: %s (explicitly run 'flow proc run %s' to confirm)", cond.Check, nodeRef),
			Required: cond.Required,
			Auto:     false,
		}
	} else {
		nodeRef := cond.NodeName
		if nodeRef == "" {
			nodeRef = cond.NodeID
		}
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Message:  fmt.Sprintf("requires AI judgment: %s (explicitly run 'flow proc run %s' to confirm)", cond.Check, nodeRef),
			Required: cond.Required,
			Auto:     false,
		}
	}

	// Auto-confirm SKIP results when explicitly targeted (AI has verified via direct targeting)
	if gc.ExplicitlyTargeted && result.Skipped {
		result.Passed = true
		result.Skipped = false
		result.Auto = false
		if result.Message != "" {
			result.Message = fmt.Sprintf("PASS (AI confirmed via explicit targeting): %s", result.Message)
		} else {
			result.Message = fmt.Sprintf("PASS: AI confirmed via explicit targeting for %s", cond.Type)
		}
	}

	return result
}

func (gc *GateChecker) checkTeamRegistered(cond GateCondOutput, checker flow.TeamGateChecker) GateCheckResult {
	switch checker.Type {
	case "script":
		return gc.checkScriptCommand(cond, checker.Command)
	case "ai_judgment":
		if gc.ExplicitlyTargeted {
			return GateCheckResult{
				Type:     cond.Type,
				Passed:   true,
				Message:  fmt.Sprintf("AI confirmed (explicitly targeted): %s", checker.Description),
				Required: cond.Required,
				Auto:     false,
			}
		}
		nodeRef := cond.NodeName
		if nodeRef == "" {
			nodeRef = cond.NodeID
		}
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Message:  fmt.Sprintf("requires AI judgment: %s (explicitly run 'flow proc run %s' to confirm)", checker.Description, nodeRef),
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

	if !bd.IsAvailable() {
		result.Passed = false
		result.Message = "FAIL: beads (bd CLI) not available — run 'flow init' to install"
		return result
	}

	args := []string{"show", "--current", "--json"}
	taskID := cond.Check
	if taskID == "" {
		taskID = cond.Expected
	}
	if taskID != "" {
		args = []string{"show", taskID, "--json"}
	}

	output, err := bd.RunQuiet(args...)
	if err != nil || output == "" {
		if taskID != "" {
			result.Passed = false
			result.Message = fmt.Sprintf("FAIL: task %s not found", taskID)
		} else {
			result.Passed = false
			result.Message = "FAIL: no active task found — run 'flow task create --title <name>'"
		}
		return result
	}

	result.Passed = true
	if taskID != "" {
		result.Message = fmt.Sprintf("PASS: task %s found", taskID)
	} else {
		result.Message = "PASS: active task found"
	}
	return result
}

// typeCompatMap defines which types are compatible for type_matches gate.
// Key is the expected flow type; values are all types that should pass the gate.
// Example: feature-flow gate with expected="feature" also accepts epic and bugfix.
var typeCompatMap = map[string][]string{
	"feature": {"feature", "epic", "bugfix", "hotfix", "change"},
	"bugfix":  {"bugfix", "hotfix", "feature"},
	"hotfix":  {"hotfix", "bugfix"},
	"release": {"release", "feature"},
	"epic":    {"epic", "feature"},
	"change":  {"change", "feature"},
}

// typeMatches checks if actualType is compatible with expectedType.
// Uses typeCompatMap for hierarchy-aware matching, falls back to exact match.
func typeMatches(actualType, expectedType string) bool {
	if strings.EqualFold(actualType, expectedType) {
		return true
	}
	// Check comma-separated expected values
	for _, exp := range strings.Split(expectedType, ",") {
		exp = strings.TrimSpace(exp)
		if strings.EqualFold(actualType, exp) {
			return true
		}
		compat, ok := typeCompatMap[strings.ToLower(exp)]
		if ok {
			for _, c := range compat {
				if strings.EqualFold(actualType, c) {
					return true
				}
			}
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

	// v4#26 fix: if check looks like a task ID (not a literal type value),
	// look up the actual task and compare its type to expected.
	if cond.Check != "" && cond.Expected != "" {
		// Try task lookup first: check if Check is a valid task ID
		taskInfo := resolveTaskInfo(cond.Check)
		if taskInfo != nil && taskInfo.Type != "" {
			if typeMatches(taskInfo.Type, cond.Expected) {
				result.Passed = true
				result.Message = fmt.Sprintf("PASS: task %s type %q is compatible with %q", taskInfo.TaskID, taskInfo.Type, cond.Expected)
			} else {
				result.Passed = false
				result.Message = fmt.Sprintf("FAIL: task %s type %q is not compatible with %q", taskInfo.TaskID, taskInfo.Type, cond.Expected)
			}
			return result
		}

		// Fallback: direct string comparison (for literal type values in check)
		if typeMatches(cond.Check, cond.Expected) {
			result.Passed = true
			result.Message = fmt.Sprintf("PASS: type %q is compatible with %q", cond.Check, cond.Expected)
		} else {
			result.Passed = false
			result.Message = fmt.Sprintf("FAIL: type %q is not compatible with %q", cond.Check, cond.Expected)
		}
		return result
	}

	// v4#26: check is empty but expected is set — try current task
	if cond.Check == "" && cond.Expected != "" {
		taskInfo := resolveTaskInfo("")
		if taskInfo != nil && taskInfo.Type != "" {
			if typeMatches(taskInfo.Type, cond.Expected) {
				result.Passed = true
				result.Message = fmt.Sprintf("PASS: current task %s type %q is compatible with %q", taskInfo.TaskID, taskInfo.Type, cond.Expected)
			} else {
				result.Passed = false
				result.Message = fmt.Sprintf("FAIL: current task %s type %q is not compatible with %q", taskInfo.TaskID, taskInfo.Type, cond.Expected)
			}
			return result
		}
	}

	// Missing check or expected — can't auto-verify
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
	if taskID == "" || taskID == "{task_id}" {
		lgr, lgrErr := eventlog.NewLogger(gc.ProjectRoot)
		if lgrErr == nil {
			activeTasks, _ := lgr.ActiveTasks()
			if len(activeTasks) == 1 {
				taskID = activeTasks[0]
			}
		}
	}
	if taskID == "" || taskID == "{task_id}" {
		result.Passed = false
		result.Message = "FAIL: trace_updated requires task_id (set in check or expected field, or use --task flag)"
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
	return RunGateCheckWithFlow(projectRoot, conditions, team, docsInternalResolved, "", false)
}

func RunGateCheckWithFlow(projectRoot string, conditions []GateCondOutput, team *flow.TeamDefinition, docsInternalResolved string, flowName string, explicitlyTargeted bool) []GateCheckResult {
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
		FlowName:             flowName,
		ExplicitlyTargeted:   explicitlyTargeted,
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

func (gc *GateChecker) checkHasActiveTasks(cond GateCondOutput) GateCheckResult {
	lgr, err := eventlog.NewLogger(gc.ProjectRoot)
	if err != nil {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Required: cond.Required,
			Message:  fmt.Sprintf("cannot access event log: %v", err),
		}
	}

	hasActive, err := lgr.HasActiveTasks()
	if err != nil {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Required: cond.Required,
			Message:  fmt.Sprintf("error reading events: %v", err),
		}
	}

	if hasActive {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   true,
			Required: cond.Required,
			Message:  "active tasks exist",
		}
	}

	return GateCheckResult{
		Type:     cond.Type,
		Passed:   false,
		Required: cond.Required,
		Message:  "no active tasks found",
	}
}

func (gc *GateChecker) checkHasSessionHistory(cond GateCondOutput) GateCheckResult {
	lgr, err := eventlog.NewLogger(gc.ProjectRoot)
	if err != nil {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Required: cond.Required,
			Message:  fmt.Sprintf("cannot access event log: %v", err),
		}
	}

	sessionName, err := lgr.LastSessionName()
	if err != nil {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   true,
			Required: cond.Required,
			Message:  "no sessions found (no history)",
		}
	}

	if sessionName != "" {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   true,
			Required: cond.Required,
			Message:  "session history exists",
		}
	}

	return GateCheckResult{
		Type:     cond.Type,
		Passed:   false,
		Required: cond.Required,
		Message:  "no previous sessions",
	}
}
