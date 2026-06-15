package proc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	skillfs "github.com/origadmin/team-flow"
	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/eventlog"
	"github.com/origadmin/team-flow/internal/flow"
)

// gateCheckTimeout is the maximum duration for any external command run during gate checks.
// Commands exceeding this are terminated and reported as FAIL with timeout message.
const gateCheckTimeout = 5 * time.Minute

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
	TaskID               string
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
	flow.GateCondChecklistComplete:    func(gc *GateChecker, cond GateCondOutput) GateCheckResult { return gc.checkChecklistComplete(cond) },
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
	} else if gc.ExplicitlyTargeted && !cond.Required {
		// AI explicitly targeted this node and this is not a required condition.
		// For non-required conditions, we trust AI judgment.
		// Required conditions MUST still be verified automatically.
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

	output, err := runCommandWithTimeout(name, args, gc.ProjectRoot)
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: %s\n%s", command, truncateOutput(output, 500))
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

	output, err := runCommandWithTimeout(name, args, gc.ProjectRoot)
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: %s\n%s", cmd, truncateOutput(output, 500))
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

	output, err := runCommandWithTimeout(name, args, gc.ProjectRoot)
	if err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("FAIL: %s\n%s", cmd, truncateOutput(output, 500))
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
	result.Skipped = true
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
	return RunGateCheckWithFlow(projectRoot, conditions, team, docsInternalResolved, "", false, "", "")
}

func RunGateCheckWithFlow(projectRoot string, conditions []GateCondOutput, team *flow.TeamDefinition, docsInternalResolved string, flowName string, explicitlyTargeted bool, sessionName string, taskID string) []GateCheckResult {
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
		TaskID:               taskID,
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

// runCommandWithTimeout executes an external command with a deadline.
// Returns (output, nil) on success, or (captured output, error) on failure/timeout.
func runCommandWithTimeout(name string, args []string, dir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gateCheckTimeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, name, args...)
	execCmd.Dir = dir
	execCmd.Env = append(os.Environ(), "GOWORK=off")

	output, err := execCmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("command timed out after %v: %s %s", gateCheckTimeout, name, strings.Join(args, " "))
	}
	return string(output), err
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

// ChecklistEvidence represents a single checklist item with evidence requirements.
type ChecklistEvidence struct {
	Desc         string
	EvidenceType string // command, file_exists, pinchtab, ai_judgment
	Cmd          string // for command type
	Path         string // for file_exists type
	URL          string // for pinchtab type: page URL
	Steps        string // for pinchtab type: verification steps (comma-separated)
	Auth         string // for pinchtab type: auth method (login, api-key, none)
	Domain       string // optional domain filter
	Module       string // optional module filter
	Required     bool
}

// EvidenceResult represents the outcome of evaluating a single checklist item.
type EvidenceResult struct {
	Desc     string
	Evidence string
	Passed   bool
	Message  string
}

func (gc *GateChecker) checkChecklistComplete(cond GateCondOutput) GateCheckResult {
	if gc.ExplicitlyTargeted {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   true,
			Message:  fmt.Sprintf("PASS: AI confirmed checklist via explicit targeting: %s", cond.Check),
			Required: cond.Required,
			Auto:     false,
		}
	}

	checklistPath := filepath.Join(gc.ProjectRoot, ".team", "checklist.md")
	var content string
	if _, err := os.Stat(checklistPath); err == nil {
		data, err := os.ReadFile(checklistPath)
		if err != nil {
			return GateCheckResult{
				Type:     cond.Type,
				Passed:   false,
				Required: cond.Required,
				Auto:     true,
				Message:  fmt.Sprintf("FAIL: cannot read project checklist: %v", err),
			}
		}
		content = string(data)
	} else {
		templatePath := "assets/skill/v3/templates/checklist-template.md"
		data, err := skillfs.FS.ReadFile(templatePath)
		if err != nil {
			return GateCheckResult{
				Type:     cond.Type,
				Passed:   false,
				Required: cond.Required,
				Auto:     true,
				Message:  fmt.Sprintf("FAIL: no project checklist and cannot read template: %v", err),
			}
		}
		content = string(data)
	}

	items := ParseChecklistEvidence(content)
	if len(items) == 0 {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   true,
			Required: cond.Required,
			Auto:     true,
			Message:  "PASS: no checklist items found",
		}
	}

	passedCount, failedCount := 0, 0
	var lines []string
	lines = append(lines, fmt.Sprintf("Checklist: %d items", len(items)))

	for _, item := range items {
		result := gc.evaluateEvidence(item)
		if result.Passed {
			passedCount++
			lines = append(lines, fmt.Sprintf("  ✅ %s  [%s]", result.Desc, result.Evidence))
		} else {
			failedCount++
			lines = append(lines, fmt.Sprintf("  ❌ %s  [%s] — %s", result.Desc, result.Evidence, result.Message))
		}
	}

	lines = append(lines, fmt.Sprintf("\n  Passed: %d  Failed: %d", passedCount, failedCount))

	if failedCount > 0 {
		return GateCheckResult{
			Type:     cond.Type,
			Passed:   false,
			Required: cond.Required,
			Auto:     true,
			Message:  strings.Join(lines, "\n"),
		}
	}
	return GateCheckResult{
		Type:     cond.Type,
		Passed:   true,
		Required: cond.Required,
		Auto:     true,
		Message:  strings.Join(lines, "\n"),
	}
}

func (gc *GateChecker) evaluateEvidence(item ChecklistEvidence) EvidenceResult {
	switch item.EvidenceType {
	case "command":
		return gc.checkEvidenceCommand(item)
	case "file_exists":
		return gc.checkEvidenceFileExists(item)
	case "pinchtab":
		return gc.checkEvidencePinchTab(item)
	case "ai_judgment":
		return gc.checkEvidenceAIJudgment(item)
	default:
		return gc.checkEvidenceLegacy(item)
	}
}

func (gc *GateChecker) checkEvidenceCommand(item ChecklistEvidence) EvidenceResult {
	r := EvidenceResult{Desc: item.Desc, Evidence: "command"}
	if item.Cmd == "" {
		r.Passed = false
		r.Message = "no command specified"
		return r
	}

	parts := strings.Fields(item.Cmd)
	if len(parts) == 0 {
		r.Passed = false
		r.Message = "empty command"
		return r
	}
	name := parts[0]
	args := parts[1:]

	output, err := runCommandWithTimeout(name, args, gc.ProjectRoot)
	if err != nil {
		r.Passed = false
		r.Message = truncateOutput(strings.TrimSpace(string(output)), 200)
		return r
	}
	r.Passed = true
	r.Message = "OK"
	return r
}

func (gc *GateChecker) checkEvidenceFileExists(item ChecklistEvidence) EvidenceResult {
	r := EvidenceResult{Desc: item.Desc, Evidence: "file_exists"}
	path := item.Path
	if path == "" {
		r.Passed = false
		r.Message = "no path specified"
		return r
	}

	path = strings.ReplaceAll(path, "{task_id}", gc.TaskID)

	if !filepath.IsAbs(path) {
		path = filepath.Join(gc.ProjectRoot, path)
	}

	if _, err := os.Stat(path); err == nil {
		r.Passed = true
		r.Message = path
		return r
	}
	r.Passed = false
	r.Message = fmt.Sprintf("not found: %s", path)
	return r
}

func (gc *GateChecker) checkEvidencePinchTab(item ChecklistEvidence) EvidenceResult {
	r := EvidenceResult{Desc: item.Desc, Evidence: "pinchtab"}
	if gc.ExplicitlyTargeted {
		r.Passed = true
		r.Message = "AI confirmed via explicit targeting"
		return r
	}
	r.Passed = false
	r.Message = "requires PinchTab session — run explicitly to confirm"
	return r
}

func (gc *GateChecker) checkEvidenceAIJudgment(item ChecklistEvidence) EvidenceResult {
	r := EvidenceResult{Desc: item.Desc, Evidence: "ai_judgment"}
	if gc.ExplicitlyTargeted {
		r.Passed = true
		r.Message = "AI confirmed"
		return r
	}
	r.Passed = false
	r.Message = "requires AI judgment — run explicitly to confirm"
	return r
}

func (gc *GateChecker) checkEvidenceLegacy(item ChecklistEvidence) EvidenceResult {
	r := EvidenceResult{Desc: item.Desc, Evidence: "legacy"}
	docsRoot := ResolveDocsPath(gc.ProjectRoot)
	desc := strings.ToLower(item.Desc)

	if strings.Contains(desc, "spec.md") {
		path := filepath.Join(docsRoot, "requirements", fmt.Sprintf("%s-spec.md", gc.TaskID))
		if _, err := os.Stat(path); err == nil {
			r.Passed = true
			r.Message = path
			return r
		}
	}
	if strings.Contains(desc, "ac.md") {
		path := filepath.Join(docsRoot, "requirements", fmt.Sprintf("%s-ac.md", gc.TaskID))
		if _, err := os.Stat(path); err == nil {
			r.Passed = true
			r.Message = path
			return r
		}
	}
	r.Passed = false
	r.Message = "no evidence type specified, cannot verify"
	return r
}

// ParseChecklistEvidence parses checklist.md content with evidence markers.
// Marker syntax: `[evidence: command]` `[cmd: go test ./...]` `[path: _docs/xxx.md]` `[domain: xxx]` `[module: xxx]`
func ParseChecklistEvidence(content string) []ChecklistEvidence {
	var items []ChecklistEvidence
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ") {
			continue
		}
		if !strings.HasPrefix(line, "- [ ] ") && !strings.HasPrefix(line, "- [x] ") {
			continue
		}

		desc := strings.TrimPrefix(line, "- [ ] ")
		desc = strings.TrimPrefix(desc, "- [x] ")
		desc = strings.TrimSpace(desc)

		item := ChecklistEvidence{Desc: desc, Required: true}

		remain := desc
		for {
			start := strings.Index(remain, "[")
			if start == -1 {
				break
			}
			end := strings.Index(remain[start:], "]")
			if end == -1 {
				break
			}
			marker := remain[start+1 : start+end]
			colon := strings.Index(marker, ":")
			if colon == -1 {
				remain = remain[start+end+1:]
				continue
			}
			key := strings.TrimSpace(marker[:colon])
			val := strings.TrimSpace(marker[colon+1:])

			switch key {
			case "evidence":
				item.EvidenceType = val
			case "cmd":
				item.Cmd = val
			case "path":
				item.Path = val
			case "url":
				item.URL = val
			case "steps":
				item.Steps = val
			case "auth":
				item.Auth = val
			case "domain":
				item.Domain = val
			case "module":
				item.Module = val
			}
			remain = remain[start+end+1:]
		}

		cleanDesc := desc
		for {
			start := strings.Index(cleanDesc, "[evidence:")
			if start == -1 {
				start = strings.Index(cleanDesc, "[cmd:")
			}
			if start == -1 {
				start = strings.Index(cleanDesc, "[path:")
			}
			if start == -1 {
				start = strings.Index(cleanDesc, "[url:")
			}
			if start == -1 {
				start = strings.Index(cleanDesc, "[steps:")
			}
			if start == -1 {
				start = strings.Index(cleanDesc, "[auth:")
			}
			if start == -1 {
				start = strings.Index(cleanDesc, "[domain:")
			}
			if start == -1 {
				start = strings.Index(cleanDesc, "[module:")
			}
			if start == -1 {
				break
			}
			// Remove surrounding backticks if present
			end := strings.Index(cleanDesc[start:], "]")
			if end == -1 {
				break
			}
			end += start + 1 // absolute end position of ']'
			cutStart, cutEnd := start, end
			if cutStart > 0 && cleanDesc[cutStart-1] == '`' {
				cutStart--
			}
			if cutEnd < len(cleanDesc) && cleanDesc[cutEnd] == '`' {
				cutEnd++
			}
			cleanDesc = strings.TrimSpace(cleanDesc[:cutStart] + cleanDesc[cutEnd:])
		}
		// Clean up leftover backtick pairs
		cleanDesc = strings.ReplaceAll(cleanDesc, "`` ``", "")
		cleanDesc = strings.ReplaceAll(cleanDesc, "`` ``", "")
		cleanDesc = strings.ReplaceAll(cleanDesc, "` `", "")
		cleanDesc = strings.ReplaceAll(cleanDesc, "  ", " ")
		cleanDesc = strings.TrimSpace(cleanDesc)
		item.Desc = cleanDesc

		if item.Desc == "" {
			continue
		}
		items = append(items, item)
	}
	return items
}
