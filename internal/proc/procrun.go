package proc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/condition"
	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/eventlog"
	"github.com/origadmin/team-flow/internal/flow"
	"github.com/origadmin/team-flow/internal/skill"
	"github.com/origadmin/team-flow/internal/trace"
)

type ProcRunRequest struct {
	FlowName    string
	NodeID      string
	TaskID      string
	ProjectRoot string
	Workspace   string
	Format      string
	RunGate     bool
	Input       string // User input for this round (written to context.md)
	Analysis    string // AI analysis section (Root Cause/Evidence/Solution/Trade-offs)
	Conclusion  string // AI conclusion section (Decision/Next Action/Blockers)
	NewSession  bool   // Force create new session (--new flag)
}

// StatusLineData holds the raw fields for status line formatting.
// The template defines how these fields are combined into a display string.
type StatusLineData struct {
	Alias    string `json:"alias"`
	NodeName string `json:"node_name"`
	NodeID   string `json:"node_id"`
	Flow     string `json:"flow"`
	Ref      string `json:"ref"`
	Phase    string `json:"phase"`
}

type ProcRunResult struct {
	Flow             FlowMeta          `json:"flow"`
	Current          CurrentNode       `json:"current"`
	NextOptions      []NextOption      `json:"next_options"`
	StatusLine       string            `json:"status_line"`
	StatusLineFields *StatusLineData   `json:"status_line_fields,omitempty"`
	StatusLineFmt    string            `json:"status_line_format,omitempty"`
	Task             *TaskInfo         `json:"task,omitempty"`
	TeamIntro        *TeamIntroData    `json:"team_intro,omitempty"`
	ProjectRoot      string            `json:"project_root,omitempty"`
	Workspace        string            `json:"workspace,omitempty"`
	ResumedFrom      string            `json:"resumed_from,omitempty"`
	FlowRevision     string            `json:"flow_revision,omitempty"`
	PathValidation   *PathValidation   `json:"path_validation,omitempty"`
	Skills           *SkillContext     `json:"skills,omitempty"`
	GateCheckResults []GateCheckResult `json:"gate_check_results,omitempty"`
	SessionName      string            `json:"-"`
	RescueContext    string            `json:"rescue_context,omitempty"`
	AnalysisSchema   *AnalysisSchema   `json:"analysis_schema,omitempty"`
	CriticalReminders []string         `json:"critical_reminders,omitempty"`
}

// AnalysisSchema tells AI the structured output contract for the current node.
type AnalysisSchema struct {
	Required  bool     `json:"required"`
	Analysis  string   `json:"analysis_prompt"`
	Conclusion string  `json:"conclusion_prompt"`
	ExampleFlags []string `json:"example_flags,omitempty"`
}

type PathValidation struct {
	ProjectRoot    string `json:"project_root"`
	CurrentCwd    string `json:"current_cwd"`
	IsMatch       bool   `json:"is_match"`
	EnforceLevel  string `json:"enforce_level"`
	Warning       string `json:"warning,omitempty"`
	RequiredAction string `json:"required_action,omitempty"`
}

type SkillContext struct {
	ProjectRoot    string                `json:"project_root"`
	TeamID         string                `json:"team_id,omitempty"`
	RoleID         string                `json:"role_id,omitempty"`
	ResolvedSkills []ResolvedSkillOutput `json:"resolved_skills"`
	CacheFile      string                `json:"cache_file,omitempty"`
}

type ResolvedSkillOutput struct {
	ID      string `json:"id"`
	Source  string `json:"source"`
	Trigger string `json:"trigger,omitempty"`
	Path    string `json:"path,omitempty"`
	Enabled bool   `json:"enabled"`
}

type TeamIntroData struct {
	TeamID          string          `json:"team_id"`
	TeamName        string          `json:"team_name"`
	TeamNameZh      string          `json:"team_name_zh,omitempty"`
	TeamDescription string          `json:"team_description,omitempty"`
	FlowName        string          `json:"flow_name"`
	Roles           []TeamIntroRole `json:"roles"`
	Flows           []TeamIntroFlow `json:"flows"`
	HowItWorks      []string        `json:"how_it_works"`
}

type TeamIntroRole struct {
	Alias     string `json:"alias"`
	AliasEn   string `json:"alias_en"`
	RoleName  string `json:"role_name"`
	Principal bool   `json:"principal,omitempty"`
}

type TeamIntroFlow struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

type TaskInfo struct {
	TaskID string `json:"task_id,omitempty"`
	Type   string `json:"type,omitempty"`
	Phase  string `json:"phase,omitempty"`
	Status string `json:"status,omitempty"`
	URL    string `json:"url,omitempty"`
}

type FlowMeta struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Domain     string `json:"domain"`
	Type       string `json:"type"`
	ParentFlow string `json:"parent_flow,omitempty"`
}

type ParallelBranchOutput struct {
	NodeID   string `json:"node_id"`
	RoleName string `json:"role_name,omitempty"`
	Alias    string `json:"alias,omitempty"`
	Name     string `json:"name,omitempty"`
}

type CurrentNode struct {
	NodeID           string                 `json:"node_id"`
	NodeType         string                 `json:"node_type"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Role             string                 `json:"role"`
	RoleName         string                 `json:"role_name,omitempty"`
	Alias            string                 `json:"alias,omitempty"`
	AliasEn          string                 `json:"alias_en,omitempty"`
	Principal        bool                   `json:"principal,omitempty"`
	Persona          string                 `json:"persona,omitempty"`
	Traits           []string               `json:"traits,omitempty"`
	Guidance         string                 `json:"guidance,omitempty"`
	PromptSource     string                 `json:"prompt_source,omitempty"`
	StandardsSource  string                 `json:"standards_source,omitempty"`
	PromptDirectives []string               `json:"prompt_directives,omitempty"`
	Rules            []RuleOutput           `json:"rules"`
	Tools            []ToolOutput           `json:"tools"`
	Skills           []SkillOutput          `json:"skills"`
	Prompts          []PromptOutput         `json:"prompts"`
	Docs             []DocOutput            `json:"docs"`
	OnEnter          []OnEnterAction        `json:"on_enter"`
	GateConditions   []GateCondOutput       `json:"gate_conditions"`
	ParallelBranches []ParallelBranchOutput `json:"parallel_branches,omitempty"`
	ParallelStrategy string                 `json:"parallel_strategy,omitempty"`
	MergeStrategy    string                 `json:"merge_strategy,omitempty"`
	SubflowRef       string                 `json:"subflow_ref,omitempty"`
	SubflowName      string                 `json:"subflow_name,omitempty"`
	IsTerminal       bool                   `json:"is_terminal"`
	TerminalStatus   string                 `json:"terminal_status,omitempty"`
	TerminalMessage  string                 `json:"terminal_message,omitempty"`
}

type RuleOutput struct {
	Ref         string `json:"ref"`
	Source      string `json:"source,omitempty"`
	Name        string `json:"name,omitempty"`
	Instruction string `json:"instruction,omitempty"`
	Enforcement string `json:"enforcement,omitempty"`
	RuleRef     string `json:"rule_ref,omitempty"`
}

type PromptOutput struct {
	Ref    string `json:"ref"`
	Source string `json:"source,omitempty"`
	Path   string `json:"path,omitempty"`
}

type ToolOutput struct {
	Ref      string   `json:"ref"`
	Source   string   `json:"source,omitempty"`
	Commands []string `json:"commands,omitempty"`
}

type SkillOutput struct {
	Ref         string `json:"ref"`
	Source      string `json:"source,omitempty"`
	Trigger     string `json:"trigger,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
}

type DocOutput struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Format       string   `json:"format"`
	Required     bool     `json:"required"`
	Description  string   `json:"description,omitempty"`
	Template     string   `json:"template,omitempty"`
	ContentRules []string `json:"content_rules,omitempty"`
}

type OnEnterAction struct {
	Action  string `json:"action"`
	Phase   string `json:"phase,omitempty"`
	Command string `json:"command,omitempty"`
}

type GateCondOutput struct {
	Type         string   `json:"type"`
	Threshold    string   `json:"threshold,omitempty"`
	Required     bool     `json:"required"`
	Check        string   `json:"check,omitempty"`
	Expected     string   `json:"expected,omitempty"`
	Deliverables []string `json:"deliverables,omitempty"`
	NodeID       string   `json:"node_id,omitempty"`
	NodeName     string   `json:"node_name,omitempty"`
}

type NextOption struct {
	NodeID    string  `json:"node_id"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	Condition *string `json:"condition"`
	IsDefault bool    `json:"is_default"`
}

type TeamLoader interface {
	LoadTeam(root string) (*flow.TeamDefinition, error)
}

type DefaultTeamLoader struct{}

func (l *DefaultTeamLoader) LoadTeam(root string) (*flow.TeamDefinition, error) {
	return LoadTeam(root)
}

type ProcRunEngine struct {
	FlowResolver   FlowResolver
	NodeResolver   NodeResolver
	VarSubstitutor VarSubstitutor
	TeamLoader     TeamLoader
	Trace          *trace.Logger
}

func NewProcRunEngine(root string) *ProcRunEngine {
	teamDir := filepath.Join(root, ".team")
	return &ProcRunEngine{
		FlowResolver:   &ActiveFlowResolver{Root: root},
		NodeResolver:   &DefaultNodeResolver{},
		VarSubstitutor: &DefaultVarSubstitutor{},
		TeamLoader:     &DefaultTeamLoader{},
		Trace:          trace.NewLogger(teamDir),
	}
}

func (e *ProcRunEngine) Run(ctx context.Context, req ProcRunRequest) (*ProcRunResult, error) {
	// --- Load session state for node progress tracking (v4#24) ---
	lgr, lgrErr := eventlog.NewLogger(req.ProjectRoot)
	var sessionName string
	if lgrErr == nil {
		sessionName = ensureSessionName(lgr, req.NewSession, req.Input)
	}
	if sessionName == "" {
		s, _ := lgr.LastSessionName()
		sessionName = s
		if sessionName == "" {
			sessionName = "unknown"
		}
	}

	fl, err := e.FlowResolver.Resolve(ctx, req.FlowName, req.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve flow: %w", err)
	}

	// Node resolution: use explicit node-id, or start from root.
	// Rescue: when no node-id and not --new, return latest session context.
	resolvedNodeID := req.NodeID
	if resolvedNodeID == "" {
		// Rescue mode: no explicit node, not --new → return latest session state
		if !req.NewSession && sessionName != "" && sessionName != "unknown" {
			sessionCtx, _ := lgr.ReadContext(sessionName)
			if sessionCtx != "" {
				// Use the actual current node from events, not the root/start node
				currentNodeID := lgr.CurrentNode(sessionName)
				if currentNodeID == "" {
					if rootNode, rerr := FindRootNode(fl); rerr == nil && rootNode != nil {
						currentNodeID = rootNode.ID
					}
				}
				node, err := e.NodeResolver.Resolve(ctx, fl, currentNodeID)
				if err == nil && node != nil {
					vars := e.VarSubstitutor.CollectVars(ctx, fl, req, node.ID)
					// Resolve team for proper role alias display
					var rescueTeam *flow.TeamDefinition
					if e.TeamLoader != nil {
						t, tErr := e.TeamLoader.LoadTeam(req.ProjectRoot)
						if tErr == nil && t != nil {
							rescueTeam = t
						}
					}
					result := generateResult(fl, node, vars, rescueTeam, true, req.ProjectRoot, req.Workspace, req.RunGate)
					condCtx := buildConditionContext(fl, node, req, result)
					result.NextOptions = buildNextOptions(fl, node.ID, condCtx)
					result.RescueContext = sessionCtx
					result.SessionName = sessionName
					// Resolve task info for proper StatusLine ref
					if taskInfo := resolveTaskInfo(req.TaskID); taskInfo != nil {
						result.Task = taskInfo
						if taskInfo.TaskID != "" {
							result.StatusLine = buildStatusLineWithRef(fl, node, result.Current, taskInfo.TaskID)
							sld := buildStatusLineData(fl, node, result.Current, taskInfo.TaskID)
							result.StatusLineFields = &sld
						}
					}
					if lgrErr == nil {
						phase := resolvePhaseFromNode(node)
						_ = lgr.UpdateContextSnapshot(sessionName, req.TaskID, fl.Metadata.Name,
							node.Name, phase, result.StatusLine, req.Input, req.Analysis, req.Conclusion)
						if req.Analysis != "" {
							_ = lgr.RecordNodeAnalysis(sessionName, req.TaskID, req.Analysis, fl.Metadata.Name, node.ID)
						}
						if req.Conclusion != "" {
							_ = lgr.RecordNodeConclusion(sessionName, req.TaskID, req.Conclusion, fl.Metadata.Name, node.ID)
						}
					}
					return result, nil
				}
			}
		}
		// Fallback: start from root
		if rootNode, rerr := FindRootNode(fl); rerr == nil && rootNode != nil {
			resolvedNodeID = rootNode.ID
		}
	}

	node, err := e.NodeResolver.Resolve(ctx, fl, resolvedNodeID)
	if err != nil {
		return nil, fmt.Errorf("resolve node: %w", err)
	}

	vars := e.VarSubstitutor.CollectVars(ctx, fl, req, node.ID)

	var team *flow.TeamDefinition
	if e.TeamLoader != nil {
		t, tErr := e.TeamLoader.LoadTeam(req.ProjectRoot)
		if tErr == nil && t != nil {
			team = t
		}
	}
	if team == nil && req.FlowName != "" {
		flowPath := req.FlowName
		if !filepath.IsAbs(flowPath) {
			flowPath = filepath.Join(req.ProjectRoot, flowPath)
		}
		t, tErr := LoadTeamFromFlowPath(flowPath)
		if tErr == nil && t != nil {
			team = t
		}
	}

	result := generateResult(fl, node, vars, team, resolvedNodeID == "", req.ProjectRoot, req.Workspace, req.RunGate)
	result.FlowRevision = ComputeFlowRevision(req.ProjectRoot, fl.Metadata.Name)

	condCtx := buildConditionContext(fl, node, req, result)
	result.NextOptions = buildNextOptions(fl, node.ID, condCtx)
	result.SessionName = sessionName

	// Trace: node entered (including Start — it has a role now)
	if e.Trace != nil && sessionName != "" {
		e.Trace.NodeEnter(sessionName, req.TaskID, node.ID, node.Name, fl.Metadata.Name)
	}

	if lgrErr == nil {
		if node.Type == flow.NodeTypeStart {
			_ = lgr.FlowStarted(sessionName, req.TaskID, fl.Metadata.Name, node.ID)
		} else {
			phase := resolvePhaseFromNode(node)
			_ = lgr.FlowNodeAdvance(sessionName, req.TaskID, node.ID, node.Name, phase, fl.Metadata.Name)
		}
	}

	if req.RunGate && node.Type == flow.NodeTypeGate && len(result.Current.GateConditions) > 0 {
		explicitlyTargeted := req.NodeID != ""
		substitutedConds := SubstituteGateConditions(result.Current.GateConditions, vars)
		docsInternal := config.ResolveInternalDocs(req.ProjectRoot)
		gateResults := RunGateCheckWithFlow(req.ProjectRoot, substitutedConds, team, docsInternal, fl.Metadata.Name, explicitlyTargeted, sessionName, req.TaskID)
		result.GateCheckResults = gateResults

		// Gate analysis file validation: verify AI wrote proper analysis.md + conclusion.md
		// to the latest round directory (方案B: round-path → AI writes → run --analysis-file).
		if lgrErr == nil && sessionName != "" {
			analysisContent, conclusionContent, round, err := lgr.ReadLatestAnalysis(sessionName)
			if err != nil {
				// Missing or unreadable analysis files → gate failure
				gateResults = append(gateResults, GateCheckResult{
					Type:     "analysis_file_validation",
					Required: true,
					Auto:     true,
					Passed:   false,
					Message:  fmt.Sprintf("Gate check: failed to read analysis files for round %d: %v", round, err),
				})
			} else {
				// Content quality checks
				analysisMinLen := 50
				conclusionMinLen := 20

				hasRootCause := strings.Contains(analysisContent, "Root Cause")
				hasEvidence := strings.Contains(analysisContent, "Evidence")
				analysisLenOK := len(analysisContent) >= analysisMinLen
				conclusionLenOK := len(conclusionContent) >= conclusionMinLen

				passed := analysisLenOK && conclusionLenOK && hasRootCause && hasEvidence
				detail := fmt.Sprintf("analysis=%d chars (min %d), conclusion=%d chars (min %d)",
					len(analysisContent), analysisMinLen, len(conclusionContent), conclusionMinLen)
				if !hasRootCause {
					detail += " | MISSING: Root Cause"
				}
				if !hasEvidence {
					detail += " | MISSING: Evidence"
				}

				gateResults = append(gateResults, GateCheckResult{
					Type:     "analysis_file_validation",
					Required: true,
					Auto:     true,
					Passed:   passed,
					Message:  fmt.Sprintf("Round %d: %s", round, detail),
				})
			}
			result.GateCheckResults = gateResults
		}

		// Trace: each gate check result
		if e.Trace != nil && sessionName != "" {
			for _, gc := range gateResults {
				e.Trace.GateCheck(sessionName, req.TaskID, node.ID, node.Name, gc.Type, gc.Passed, gc.Message)
			}
		}

		if req.TaskID != "" {
			passed, _ := GateOverallResult(gateResults)
			appendSessionLog(req.ProjectRoot, req.TaskID, node, gateResults, passed, req.Input, req.Analysis, req.Conclusion)
		}
	}

	if taskInfo := resolveTaskInfo(req.TaskID); taskInfo != nil {
		result.Task = taskInfo
		if taskInfo.TaskID != "" {
			result.StatusLine = buildStatusLineWithRef(fl, node, result.Current, taskInfo.TaskID)
			sld := buildStatusLineData(fl, node, result.Current, taskInfo.TaskID)
			result.StatusLineFields = &sld
		}
	}

	// Write flow.ended event for terminal nodes (always, even without task)
	if node.Type == flow.NodeTypeTerminal {
		status := "success"
		if result.Current.TerminalStatus != "" {
			status = result.Current.TerminalStatus
		}
		if lgrErr == nil {
			_ = lgr.FlowEnded(sessionName, req.TaskID, fl.Metadata.Name, status)
			_ = lgr.MarkSessionCompleted(sessionName)
		}
	}

	// Auto-update context.md with current node/task snapshot
	if lgrErr == nil && sessionName != "" && sessionName != "unknown" {
		nextNodeName := ""
		if len(result.NextOptions) > 0 {
			for _, opt := range result.NextOptions {
				if opt.IsDefault {
					nextNodeName = opt.Name
					break
				}
			}
			if nextNodeName == "" {
				nextNodeName = result.NextOptions[0].Name
			}
		}

		phase := resolvePhaseFromNode(node)
		_ = lgr.UpdateContextSnapshot(sessionName, req.TaskID, fl.Metadata.Name,
			node.Name, phase, result.StatusLine, req.Input, req.Analysis, req.Conclusion)

		// Record AI analysis to events.mdl if provided
		if req.Analysis != "" {
			_ = lgr.RecordNodeAnalysis(sessionName, req.TaskID, req.Analysis, fl.Metadata.Name, node.ID)
		}
		if req.Conclusion != "" {
			_ = lgr.RecordNodeConclusion(sessionName, req.TaskID, req.Conclusion, fl.Metadata.Name, node.ID)
		}
	}

	return result, nil
}

func resolveTaskInfo(taskID string) *TaskInfo {
	root, _ := os.Getwd()

	if taskID != "" {
		if info := resolveTaskInfoFromLocal(root, taskID); info != nil {
			return info
		}
		if bd.IsAvailable() {
			if info := resolveTaskInfoFromBeads(taskID); info != nil {
				return info
			}
		}
		return &TaskInfo{TaskID: taskID, Phase: "task-bound", Status: "open"}
	}

	if bd.IsAvailable() {
		if info := resolveTaskInfoFromBeads(""); info != nil {
			return info
		}
	}
	return resolveTaskInfoFromLocal(root, "")
}

func resolveTaskInfoFromLocal(root, taskID string) *TaskInfo {
	if taskID == "" {
		tasksDir := filepath.Join(root, ".team", "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil || len(entries) == 0 {
			return nil
		}
		var latest os.DirEntry
		var latestMod time.Time
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			fi, err := e.Info()
			if err != nil {
				continue
			}
			if latest == nil || fi.ModTime().After(latestMod) {
				latest = e
				latestMod = fi.ModTime()
			}
		}
		if latest == nil {
			return nil
		}
		taskID = strings.TrimSuffix(latest.Name(), ".json")
	}
	if taskID == "" {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(root, ".team", "tasks", taskID+".json"))
	if err != nil {
		return nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	info := &TaskInfo{}
	if id, ok := raw["id"].(string); ok {
		info.TaskID = id
	}
	if t, ok := raw["type"].(string); ok {
		info.Type = t
	}
	if status, ok := raw["status"].(string); ok {
		info.Status = status
	}
	if info.TaskID == "" && info.Type == "" {
		return nil
	}

	if labels, ok := raw["labels"].(map[string]interface{}); ok {
		if phase, ok := labels["phase"].(string); ok {
			info.Phase = phase
		}
	}
	return info
}

func resolveTaskInfoFromBeads(taskID string) *TaskInfo {
	args := []string{"show", "--current", "--json"}
	if taskID != "" {
		args = []string{"show", taskID, "--json"}
	}
	output, err := bd.RunQuiet(args...)
	if err != nil || output == "" {
		return nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return nil
	}
	info := &TaskInfo{}
	if id, ok := raw["id"].(string); ok {
		info.TaskID = id
	}
	if t, ok := raw["type"].(string); ok {
		info.Type = t
	}
	if status, ok := raw["status"].(string); ok {
		info.Status = status
	}
	if url, ok := raw["url"].(string); ok {
		info.URL = url
	}
	if labels, ok := raw["labels"].([]interface{}); ok {
		for _, l := range labels {
			if s, ok := l.(string); ok && len(s) > 6 && s[:6] == "phase:" {
				info.Phase = s[6:]
				break
			}
		}
	}
	if info.TaskID == "" && info.Phase == "" {
		return nil
	}
	return info
}

func generateResult(fl *flow.Flow, node *flow.FlowNode, vars map[string]string, team *flow.TeamDefinition, isFirstNode bool, projectRoot string, workspace string, runGate bool) *ProcRunResult {
	domain := fl.Metadata.Domain
	if fl.Config != nil {
		if domain == "" {
			domain = fl.Config.Domain
		}
		if domain == "" {
			domain = string(fl.Config.TaskType)
		}
	}

	current := buildCurrentNode(node, fl, vars, team, projectRoot)

	result := &ProcRunResult{
		Flow: FlowMeta{
			Name:       fl.Metadata.Name,
			Version:    fl.Version,
			Domain:     domain,
			Type:       fl.Metadata.Type,
			ParentFlow: fl.Metadata.ParentFlow,
		},
		Current:     current,
		StatusLine:  buildStatusLine(fl, node, current, vars),
		ProjectRoot: projectRoot,
		Workspace:   workspace,
		StatusLineFmt: DefaultStatusLineFmt,
		AnalysisSchema: &AnalysisSchema{
			Required: true,
			Analysis:  "Detailed analysis: Files Examined + Key Findings + Reasoning + Root Cause + Evidence + Solution + Trade-offs",
			Conclusion: "Decision (proceed|block|require-info) + Next Action + Blockers",
			ExampleFlags: []string{"--analysis", "--analysis-file", "--conclusion", "--conclusion-file"},
		},
	}
	sld := buildStatusLineData(fl, node, current, "")
	result.StatusLineFields = &sld

	if runGate && node.Type == flow.NodeTypeGate && len(current.GateConditions) > 0 {
		substitutedConds := SubstituteGateConditions(current.GateConditions, vars)
		docsInternal := config.ResolveInternalDocs(projectRoot)
		result.GateCheckResults = RunGateCheck(projectRoot, substitutedConds, team, docsInternal)
	}

	if isFirstNode {
		result.TeamIntro = buildTeamIntro(fl, team, projectRoot)
	}

	if result.Current.IsTerminal {
		result.NextOptions = []NextOption{}
	}

	result.PathValidation = validatePath(projectRoot)

	if projectRoot != "" {
		skillMgr := skill.NewSkillManager(projectRoot)
		roleID := current.Role
		if skillsFile, err := skillMgr.Resolve(roleID); err == nil {
			skillCtx := &SkillContext{
				ProjectRoot: projectRoot,
				TeamID:      skillsFile.Team,
				RoleID:      skillsFile.Role,
				CacheFile:   skillMgr.CachePath(),
			}
			for _, s := range skillsFile.Skills {
				skillCtx.ResolvedSkills = append(skillCtx.ResolvedSkills, ResolvedSkillOutput{
					ID:      s.ID,
					Source:  s.Source,
					Trigger: s.Trigger,
					Path:    s.Path,
					Enabled: s.Enabled,
				})
			}
			result.Skills = skillCtx
		}
	}

	result.CriticalReminders = buildCriticalReminders(current)

	return result
}

func buildCriticalReminders(current CurrentNode) []string {
	var reminders []string
	if current.PromptSource != "" {
		reminders = append(reminders, fmt.Sprintf("🚨 MANDATORY: You MUST read the persona prompt at '%s' before acting.", current.PromptSource))
	}
	for _, doc := range current.Docs {
		if doc.Required {
			reminders = append(reminders, fmt.Sprintf("🚨 REQUIRED: You MUST read the reference document '%s' at '%s' before acting.", doc.Name, doc.Path))
		}
	}
	return reminders
}

func validatePath(projectRoot string) *PathValidation {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	cwdAbs, _ := filepath.Abs(cwd)
	projectAbs, _ := filepath.Abs(projectRoot)

	isMatch := strings.HasPrefix(cwdAbs, projectAbs)

	warning := ""
	requiredAction := ""
	if !isMatch {
		warning = "AI is not in the project directory"
		requiredAction = fmt.Sprintf("cd %s", projectRoot)
	}

	return &PathValidation{
		ProjectRoot:     projectRoot,
		CurrentCwd:     cwd,
		IsMatch:        isMatch,
		EnforceLevel:   "hard",
		Warning:        warning,
		RequiredAction: requiredAction,
	}
}

func buildCurrentNode(node *flow.FlowNode, fl *flow.Flow, vars map[string]string, team *flow.TeamDefinition, root string) CurrentNode {
	current := CurrentNode{
		NodeID:     node.ID,
		NodeType:   string(node.Type),
		Name:       node.Name,
		Description: node.Description,
		IsTerminal: node.Type == flow.NodeTypeTerminal,
	}

	switch node.Type {
	case flow.NodeTypeStart:
		ri := resolveRoleInfo(node, fl, team, root)
		// Start node is mechanical — no role assigned. Do NOT fall back to principal.
		current.Role = extractRole(node)
		if current.Role == "" && ri.roleID != "" {
			current.Role = ri.roleID
		}
		current.RoleName = ri.roleName
		current.Alias = ri.alias
		current.AliasEn = ri.aliasEn
		current.Principal = ri.principal
		current.Persona = ri.persona
		current.Traits = ri.traits
		if ri.guidance != "" {
			current.Guidance = ri.guidance
		} else {
			current.Guidance = "Record and analyze user intent. Determine if a task should be created or if this is a continuation of an existing conversation. For new requests, create a task via 'flow task create'. For continuations, identify the existing task and proceed."
		}
		current.PromptSource = substituteVars(ri.promptSource, vars)
		current.StandardsSource = substituteVars(ri.standardsSource, vars)
		current.PromptDirectives = ri.directives
		current.Rules = extractRules(node, fl, team, ri.roleRules, root)
		current.Tools = extractTools(node)
		current.Skills = extractSkills(node, fl)
		current.Prompts = extractPrompts(node, vars)
		current.Docs = extractDocs(node, vars)
		current.OnEnter = extractOnEnter(node)
		if len(current.OnEnter) == 0 {
			current.OnEnter = []OnEnterAction{
				{Action: "record_context", Phase: "start"},
				{Action: "analyze_intent", Phase: "start"},
			}
		}

	case flow.NodeTypePhase:
		ri := resolveRoleInfo(node, fl, team, root)
		current.Role = extractRole(node)
		current.RoleName = ri.roleName
		current.Alias = ri.alias
		current.AliasEn = ri.aliasEn
		current.Principal = ri.principal
		current.Persona = ri.persona
		current.Traits = ri.traits
		current.Guidance = ri.guidance
		current.PromptSource = substituteVars(ri.promptSource, vars)
		current.StandardsSource = substituteVars(ri.standardsSource, vars)
		current.PromptDirectives = ri.directives
		current.Rules = extractRules(node, fl, team, ri.roleRules, root)
		current.Tools = extractTools(node)
		current.Skills = extractSkills(node, fl)
		current.Prompts = extractPrompts(node, vars)
		current.Docs = extractDocs(node, vars)
		current.OnEnter = extractOnEnter(node)
		current.GateConditions = nil

	case flow.NodeTypeGate:
		current.Role = ""
		current.GateConditions = ExtractGateConditions(node)
		current.OnEnter = extractOnEnter(node)
		current.Rules = extractRules(node, fl, team, nil, root)

	case flow.NodeTypeSubflow:
		current.GateConditions = nil
		if node.Config != nil {
			var sfCfg flow.SubflowConfig
			if err := json.Unmarshal(node.Config, &sfCfg); err == nil {
				current.SubflowRef = sfCfg.FlowRef
				// Resolve subflow name for display
				if sf, err := LoadFlow(root, sfCfg.FlowRef); err == nil {
					current.SubflowName = sf.Metadata.Name
				}
			}
		}
		current.OnEnter = extractOnEnter(node)
		current.Rules = extractRules(node, fl, team, nil, root)

	case flow.NodeTypeTerminal:
		// Terminal nodes inherit role from the predecessor for status line display.
		// Look backward through edges to find the node that led here.
		var terminalRoleRules []string
		if predecessor := findPredecessor(fl, node); predecessor != nil {
			ri := resolveRoleInfo(predecessor, fl, team, root)
			current.Role = extractRole(predecessor)
			current.RoleName = ri.roleName
			current.Alias = ri.alias
			current.AliasEn = ri.aliasEn
			current.Principal = ri.principal
			terminalRoleRules = ri.roleRules
		}
		current.IsTerminal = true
		current.GateConditions = nil
		current.Rules = extractRules(node, fl, team, terminalRoleRules, root)
		if node.Config != nil {
			var termCfg flow.TerminalConfig
			if err := json.Unmarshal(node.Config, &termCfg); err == nil {
				current.TerminalStatus = string(termCfg.Status)
				current.TerminalMessage = termCfg.Message
			}
		}

	case flow.NodeTypeParallel:
		current.Role = ""
		current.GateConditions = nil
		if node.Config != nil {
			var parCfg flow.ParallelConfig
			if err := json.Unmarshal(node.Config, &parCfg); err == nil {
				current.ParallelStrategy = string(parCfg.Strategy)
				current.MergeStrategy = string(parCfg.MergeStrategy)
				for _, b := range parCfg.Branches {
					pbo := ParallelBranchOutput{NodeID: b.Node}
					for i := range fl.Nodes {
						if fl.Nodes[i].ID == b.Node {
							pbo.Name = fl.Nodes[i].Name
							ri := resolveRoleInfo(&fl.Nodes[i], fl, team, root)
							pbo.RoleName = ri.roleName
							pbo.Alias = ri.alias
							break
						}
					}
					current.ParallelBranches = append(current.ParallelBranches, pbo)
				}
			}
		}

	default:
		ri := resolveRoleInfo(node, fl, team, root)
		current.Role = extractRole(node)
		current.RoleName = ri.roleName
		current.Alias = ri.alias
		current.AliasEn = ri.aliasEn
		current.Principal = ri.principal
		current.Persona = ri.persona
		current.Traits = ri.traits
		current.Guidance = ri.guidance
		current.PromptSource = substituteVars(ri.promptSource, vars)
		current.StandardsSource = substituteVars(ri.standardsSource, vars)
		current.PromptDirectives = ri.directives
		current.Rules = extractRules(node, fl, team, ri.roleRules, root)
		current.Tools = extractTools(node)
		current.Skills = extractSkills(node, fl)
		current.Prompts = extractPrompts(node, vars)
		current.Docs = extractDocs(node, vars)
		current.OnEnter = extractOnEnter(node)
	}

	return current
}

func extractRole(node *flow.FlowNode) string {
	if node.Config == nil {
		return ""
	}
	var phaseCfg flow.PhaseConfig
	if err := json.Unmarshal(node.Config, &phaseCfg); err == nil {
		return phaseCfg.Role
	}
	return ""
}

func buildConditionContext(fl *flow.Flow, node *flow.FlowNode, req ProcRunRequest, result *ProcRunResult) map[string]string {
	ctx := make(map[string]string)

	if req.TaskID != "" {
		ctx["status"] = "task_created"
		if result.Task != nil {
			if result.Task.Type != "" {
				ctx["task_type"] = result.Task.Type
			}
		}
	} else {
		ctx["status"] = "no_task"
	}

	if result.GateCheckResults != nil {
		passed, _ := GateOverallResult(result.GateCheckResults)
		if passed {
			ctx["gate.passed"] = "true"
			ctx["verified"] = "true"
		} else {
			ctx["gate.passed"] = ""
		}
	}

	if result.Task != nil {
		if result.Task.Phase != "" {
			ctx["phase"] = result.Task.Phase
		}
		if result.Task.Status != "" {
			ctx[result.Task.Status] = "true"
		}
	}

	return ctx
}

type roleInfo struct {
	roleID          string
	promptSource    string
	standardsSource string
	directives      []string
	persona         string
	traits          []string
	guidance        string
	alias           string
	aliasEn         string
	roleName        string
	principal       bool
	roleRules       []string
}

func resolveRoleInfo(node *flow.FlowNode, fl *flow.Flow, team *flow.TeamDefinition, root string) roleInfo {
	info := roleInfo{}
	
	// Phase 1: Try node config + component roles (normal path)
	if node.Config != nil {
		var phaseCfg flow.PhaseConfig
		if err := json.Unmarshal(node.Config, &phaseCfg); err == nil {
			if node.Components != nil && len(node.Components.Roles) > 0 {
				roleRef := node.Components.Roles[0].Ref
				return resolveRoleByRef(roleRef, fl, team, root)
			}
		}
	}
	
	// Phase 2: Try flow-level components roles
	if fl.Components != nil {
		for _, r := range fl.Components.Roles {
			if r.Principal != nil && *r.Principal {
				return resolveRoleByRef(r.ID, fl, team, root)
			}
		}
	}
	
	// Phase 3: Fallback to team's principal role
	if team != nil {
		for _, r := range team.Roles {
			if r.Principal != nil && *r.Principal {
				return resolveRoleByRef(r.ID, fl, team, root)
			}
		}
	}
	
	// Phase 4: Zero config fallback - derive org and find principal role
	if team == nil && fl != nil {
		org := deriveOrg(fl.Metadata.Name)
		// Look for principal role in org's team.json (embedded)
		teamDef, err := loadTeamFromEmbed(org)
		if err == nil && teamDef != nil {
			for _, r := range teamDef.Roles {
				if r.Principal != nil && *r.Principal {
					return resolveRoleByRef(r.ID, fl, teamDef, root)
				}
			}
		}
	}
	
	return info
}

// resolveRoleByRef 解析指定角色引用，按优先级从 team → flow → org 查找
func resolveRoleByRef(roleRef string, fl *flow.Flow, team *flow.TeamDefinition, root string) roleInfo {
	// Phase 1: Check team-level roles
	if team != nil {
		for _, r := range team.Roles {
			if r.ID == roleRef {
				// v3 架构: 如果 team.json 角色没有内联内容但有 prompt_source,
				// 从 prompts/*.md 解析角色内容
				if isReferenceOnlyRole(r) && r.PromptSource != "" {
					if resolved, err := loadRoleFromPromptSource(r.ID, r.PromptSource, root); err == nil && resolved != nil {
						// 合并：从 prompt_source 解析出内容后, 保留 team.json 的 id 和 prompt_source
						resolved.ID = r.ID
						resolved.PromptSource = r.PromptSource
						resolved.StandardsSource = r.StandardsSource
						if r.Principal != nil {
							resolved.Principal = r.Principal
						}
						return fillRoleInfo(*resolved)
					}
				}
				// 回退: 旧格式（有内联内容）或 prompt_source 加载失败
				return fillRoleInfo(r)
			}
		}
	}

	// Phase 2: Check flow-level roles
	if fl.Components != nil {
		for _, r := range fl.Components.Roles {
			if r.ID == roleRef {
				// v3 同样处理 flow-level 角色引用
				if isReferenceOnlyRole(r) && r.PromptSource != "" {
					if resolved, err := loadRoleFromPromptSource(r.ID, r.PromptSource, root); err == nil && resolved != nil {
						resolved.ID = r.ID
						resolved.PromptSource = r.PromptSource
						resolved.StandardsSource = r.StandardsSource
						if r.Principal != nil {
							resolved.Principal = r.Principal
						}
						return fillRoleInfo(*resolved)
					}
				}
				return fillRoleInfo(r)
			}
		}
	}

	// Phase 3: Load from org roles (v1/v2 legacy format)
	var org string
	if team != nil {
		org = team.Org
		if org == "" {
			org = team.ID
		}
	} else {
		org = deriveOrg(fl.Metadata.Name)
	}
	
	r, err := loadRole(roleRef, org, root)
	if err == nil && r != nil {
		if isReferenceOnlyRole(*r) && r.PromptSource != "" {
			if resolved, err := loadRoleFromPromptSource(r.ID, r.PromptSource, root); err == nil && resolved != nil {
				resolved.ID = r.ID
				resolved.PromptSource = r.PromptSource
				resolved.StandardsSource = r.StandardsSource
				if r.Principal != nil {
					resolved.Principal = r.Principal
				}
				return fillRoleInfo(*resolved)
			}
		}
		return fillRoleInfo(*r)
	}
	
	return roleInfo{}
}

// isReferenceOnlyRole 判断角色是否为 v3 引用格式（只有 id/prompt_source/principal, 无内联内容）。
// 是 v3 引用格式则返回 true。
func isReferenceOnlyRole(r flow.RoleDefinition) bool {
	return r.Persona == "" &&
		r.Guidance == "" &&
		len(r.Traits) == 0 &&
		len(r.Capabilities) == 0 &&
		r.Name == "" &&
		r.Alias == ""
}

func fillRoleInfo(r flow.RoleDefinition) roleInfo {
	info := roleInfo{
		roleID:          r.ID,
		promptSource:    r.PromptSource,
		standardsSource: r.StandardsSource,
		directives:      r.PromptDirectives,
		persona:         r.Persona,
		traits:          r.Traits,
		guidance:        r.Guidance,
		alias:           r.Alias,
		aliasEn:         r.AliasEn,
		roleName:        r.Name,
		roleRules:       r.Rules,
	}
	if r.Principal != nil && *r.Principal {
		info.principal = true
	}
	return info
}

func extractPrompts(node *flow.FlowNode, vars map[string]string) []PromptOutput {
	if node.Components == nil || len(node.Components.Prompts) == 0 {
		return nil
	}
	prompts := make([]PromptOutput, 0, len(node.Components.Prompts))
	for _, p := range node.Components.Prompts {
		prompts = append(prompts, PromptOutput{
			Ref:    p.Ref,
			Source: string(p.Source),
			Path:   substituteVars(p.Path, vars),
		})
	}
	return prompts
}

func extractRules(node *flow.FlowNode, fl *flow.Flow, team *flow.TeamDefinition, roleRules []string, root string) []RuleOutput {
	ruleDefs := make(map[string]flow.RuleDefinition)
	if team != nil {
		for _, r := range team.Rules {
			ruleDefs[r.ID] = r
		}
	}
	if fl.Components != nil {
		for _, r := range fl.Components.Rules {
			ruleDefs[r.ID] = r
		}
	}
	rules := make([]RuleOutput, 0)
	if node.Components != nil {
		for _, r := range node.Components.Rules {
			out := buildRuleOutput(r.Ref, string(r.Source), ruleDefs, team, root)
			rules = append(rules, out)
		}
	}
	for _, ref := range roleRules {
		found := false
		for _, existing := range rules {
			if existing.Ref == ref {
				found = true
				break
			}
		}
		if found {
			continue
		}
		out := buildRuleOutput(ref, "role", ruleDefs, team, root)
		rules = append(rules, out)
	}
	// 默认规则回退: 任何节点都必须加载 sl1 (StatusLine 最高规则)，
	// 没有其他规则时至少应用 qg4 (output-guard)
	if team != nil {
		// sl1 (StatusLine) 是最高规则，无条件加入
		if _, ok := ruleDefs["sl1"]; ok {
			hasSL1 := false
			for _, r := range rules {
				if r.Ref == "sl1" {
					hasSL1 = true
					break
				}
			}
			if !hasSL1 {
				out := buildRuleOutput("sl1", "builtin", ruleDefs, team, root)
				// sl1 插在最前面，因为是最高优先级
				rules = append([]RuleOutput{out}, rules...)
			}
		}
		// 其他规则为空时，至少补充 qg4 / output-guard；或者对于 start/gate/terminal 节点，必须包含 qg4 / output-guard
		isSpecialNode := node != nil && (node.Type == flow.NodeTypeStart || node.Type == flow.NodeTypeTerminal || node.Type == flow.NodeTypeGate)
		if len(rules) <= 1 || isSpecialNode {
			defaultRuleIDs := []string{"qg4", "output-guard"}
			for _, ruleID := range defaultRuleIDs {
				if _, ok := ruleDefs[ruleID]; ok {
					has := false
					for _, r := range rules {
						if r.Ref == ruleID {
							has = true
							break
						}
					}
					if !has {
						out := buildRuleOutput(ruleID, "builtin", ruleDefs, team, root)
						rules = append(rules, out)
					}
				}
			}
		}
	}
	if len(rules) == 0 {
		return nil
	}
	return rules
}

func buildRuleOutput(ref string, source string, ruleDefs map[string]flow.RuleDefinition, team *flow.TeamDefinition, root string) RuleOutput {
	out := RuleOutput{
		Ref:    ref,
		Source: source,
	}
	if def, ok := ruleDefs[ref]; ok {
		out.Name = def.Name
		out.Instruction = def.Instruction
		out.Enforcement = string(def.Enforcement)
		out.RuleRef = fmt.Sprintf("flow proc rule %s", def.ID)
		return out
	}
	if team != nil {
		org := team.Org
		if org == "" {
			org = team.ID
		}
		if org != "" {
			r, err := loadRule(ref, org, root)
			if err == nil && r != nil {
				out.Name = r.Name
				out.Instruction = r.Instruction
				out.Enforcement = string(r.Enforcement)
				out.RuleRef = fmt.Sprintf("flow proc rule %s", r.ID)
			}
		}
	}
	return out
}

func extractTools(node *flow.FlowNode) []ToolOutput {
	if node.Components == nil {
		return nil
	}
	tools := make([]ToolOutput, 0, len(node.Components.Tools))
	for _, t := range node.Components.Tools {
		tools = append(tools, ToolOutput{
			Ref:      t.Ref,
			Source:   string(t.Source),
			Commands: t.Commands,
		})
	}
	if len(tools) == 0 {
		return nil
	}
	return tools
}

func extractSkills(node *flow.FlowNode, fl *flow.Flow) []SkillOutput {
	if node.Components == nil {
		return nil
	}
	skillDefs := make(map[string]flow.SkillDefinition)
	if fl.Components != nil {
		for _, s := range fl.Components.Skills {
			skillDefs[s.ID] = s
		}
	}
	skills := make([]SkillOutput, 0, len(node.Components.Skills))
	for _, s := range node.Components.Skills {
		out := SkillOutput{
			Ref:    s.Ref,
			Source: string(s.Source),
		}
		if def, ok := skillDefs[s.Ref]; ok {
			out.Trigger = def.Trigger
			out.Path = def.Path
			out.Description = def.Name
			if def.Description != "" {
				out.Description = def.Description
			}
		}
		skills = append(skills, out)
	}
	if len(skills) == 0 {
		return nil
	}
	return skills
}

func extractDocs(node *flow.FlowNode, vars map[string]string) []DocOutput {
	if len(node.Docs) == 0 {
		return nil
	}
	docs := make([]DocOutput, 0, len(node.Docs))
	for _, d := range node.Docs {
		path := substituteVars(d.Path, vars)
		required := false
		if d.Required != nil {
			required = *d.Required
		}
		template := substituteVars(d.Template, vars)
		out := DocOutput{
			Name:        d.Name,
			Path:        path,
			Format:      d.Format,
			Required:    required,
			Description: d.Description,
			Template:    template,
		}
		if len(d.ContentRules) > 0 {
			out.ContentRules = d.ContentRules
		}
		docs = append(docs, out)
	}
	return docs
}

func extractOnEnter(node *flow.FlowNode) []OnEnterAction {
	if len(node.OnEnter) == 0 {
		return nil
	}
	actions := make([]OnEnterAction, 0, len(node.OnEnter))
	for _, a := range node.OnEnter {
		actions = append(actions, OnEnterAction{
			Action:  string(a.Action),
			Phase:   a.Phase,
			Command: a.Command,
		})
	}
	return actions
}

func buildNextOptions(fl *flow.Flow, currentNodeID string, ctx map[string]string) []NextOption {
	var outgoingEdges []flow.FlowEdge
	for _, e := range fl.Edges {
		if e.From == currentNodeID {
			outgoingEdges = append(outgoingEdges, e)
		}
	}

	if len(outgoingEdges) == 0 {
		return nil
	}

	nodeMap := make(map[string]*flow.FlowNode, len(fl.Nodes))
	for i := range fl.Nodes {
		nodeMap[fl.Nodes[i].ID] = &fl.Nodes[i]
	}

	hasConditionalEdges := false
	for _, e := range outgoingEdges {
		if len(e.Conditions) > 0 {
			hasConditionalEdges = true
			break
		}
	}

	if hasConditionalEdges {
		return buildNextOptionsFromEdges(outgoingEdges, nodeMap, ctx)
	}

	currentNode := nodeMap[currentNodeID]
	if currentNode != nil && currentNode.Type == flow.NodeTypeGate {
		return BuildNextOptionsFromGateFallback(currentNode, outgoingEdges, nodeMap)
	}

	return buildNextOptionsFromEdges(outgoingEdges, nodeMap, ctx)
}

func buildNextOptionsFromEdges(edges []flow.FlowEdge, nodeMap map[string]*flow.FlowNode, ctx map[string]string) []NextOption {
	matchingEdges := make([]flow.FlowEdge, 0, len(edges))
	for _, e := range edges {
		if len(e.Conditions) == 0 || ctx == nil {
			matchingEdges = append(matchingEdges, e)
			continue
		}
		allMatch := true
		for _, c := range e.Conditions {
			if !condition.Eval(c.Expression, ctx) {
				allMatch = false
				break
			}
		}
		if allMatch {
			matchingEdges = append(matchingEdges, e)
		}
	}

	if len(matchingEdges) == 0 {
		matchingEdges = edges
	}

	options := make([]NextOption, 0, len(matchingEdges))
	defaultSet := false

	for _, e := range matchingEdges {
		role := ""
		name := ""
		if target, ok := nodeMap[e.To]; ok {
			role = extractRole(target)
			name = target.Name
		}

		var condition *string
		if len(e.Conditions) > 0 {
			condition = &e.Conditions[0].Expression
		}

		isDefault := false
		if !defaultSet {
			if condition == nil {
				isDefault = true
				defaultSet = true
			} else if IsPositiveCondition(*condition) {
				isDefault = true
				defaultSet = true
			}
		}

		options = append(options, NextOption{
			NodeID:    e.To,
			Name:      name,
			Role:      role,
			Condition: condition,
			IsDefault: isDefault,
		})
	}

	if !defaultSet && len(options) > 0 {
		options[0].IsDefault = true
	}

	return options
}

func buildTeamIntro(fl *flow.Flow, team *flow.TeamDefinition, projectRoot string) *TeamIntroData {
	intro := &TeamIntroData{
		FlowName: fl.Metadata.Name,
		HowItWorks: []string{
			"You tell the principal what you need",
			"Principal classifies and dispatches to the right specialist",
			"Specialist does the work and produces deliverables",
			"Principal verifies and reports back to you",
		},
	}

	if team != nil {
		intro.TeamID = team.ID
		intro.TeamName = team.Name
		intro.TeamNameZh = team.NameZh
		intro.TeamDescription = team.Description

		for _, r := range team.Roles {
			intro.Roles = append(intro.Roles, TeamIntroRole{
				Alias:     r.Alias,
				AliasEn:   r.AliasEn,
				RoleName:  r.Name,
				Principal: r.Principal != nil && *r.Principal,
			})
		}

		for _, f := range team.Flows {
			intro.Flows = append(intro.Flows, TeamIntroFlow{
				ID:          f.ID,
				Description: f.Description,
				IsDefault:   f.ID == fl.Metadata.Name,
			})
		}
	}

	if len(intro.Roles) == 0 {
		seen := make(map[string]bool)
		for i := range fl.Nodes {
			n := &fl.Nodes[i]
			ri := resolveRoleInfo(n, fl, team, projectRoot)
			if ri.alias != "" && !seen[ri.alias] {
				seen[ri.alias] = true
				intro.Roles = append(intro.Roles, TeamIntroRole{
					Alias:     ri.alias,
					AliasEn:   ri.aliasEn,
					RoleName:  ri.roleName,
					Principal: ri.principal,
				})
			}
		}
	}

	if len(intro.Flows) == 0 {
		intro.Flows = loadAvailableFlows(fl.Metadata.Name, projectRoot)
	}

	return intro
}

func loadAvailableFlows(activeFlow, projectRoot string) []TeamIntroFlow {
	var flows []TeamIntroFlow
	flowsDir := filepath.Join(projectRoot, ".team", "flows")
	entries, err := os.ReadDir(flowsDir)
	if err != nil {
		flows = append(flows, TeamIntroFlow{ID: activeFlow, IsDefault: true})
		return flows
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		flowID := strings.TrimSuffix(entry.Name(), ".json")
		flows = append(flows, TeamIntroFlow{
			ID:        flowID,
			IsDefault: flowID == activeFlow,
		})
	}
	if len(flows) == 0 {
		flows = append(flows, TeamIntroFlow{ID: activeFlow, IsDefault: true})
	}
	return flows
}

// DefaultStatusLineFmt is the built-in template for status line formatting.
// Variables: {alias}, {node_name}, {node_id}, {flow}, {ref}, {phase}
// Override via .team/project.yaml → status_line_format
const DefaultStatusLineFmt = "[{alias} | {node_name}({node_id}:{flow}) | {ref} | {phase}]"

func buildStatusLine(fl *flow.Flow, node *flow.FlowNode, current CurrentNode, vars map[string]string) string {
	return buildStatusLineWithRef(fl, node, current, "")
}

func buildStatusLineWithRef(fl *flow.Flow, node *flow.FlowNode, current CurrentNode, ref string) string {
	data := buildStatusLineData(fl, node, current, ref)
	return formatStatusLine(data, DefaultStatusLineFmt)
}

// buildStatusLineData extracts raw fields for status line formatting.
func buildStatusLineData(fl *flow.Flow, node *flow.FlowNode, current CurrentNode, ref string) StatusLineData {
	alias := current.Alias
	if alias == "" {
		alias = current.Role
	}
	if alias == "" {
		alias = "System"
	}

	flowPart := fl.Metadata.Name

	if ref == "" {
		ref = "-"
	}

	phase := ""
	for _, a := range current.OnEnter {
		if a.Action == "update_task_phase" && a.Phase != "" {
			phase = a.Phase
		}
	}
	if phase == "" {
		for _, a := range current.OnEnter {
			if a.Phase != "" {
				phase = a.Phase
				break
			}
		}
	}

	return StatusLineData{
		Alias:    alias,
		NodeName: current.Name,
		NodeID:   node.ID,
		Flow:     flowPart,
		Ref:      ref,
		Phase:    phase,
	}
}

// formatStatusLine applies a template to StatusLineData fields.
// Template variables: {alias}, {node_name}, {node_id}, {flow}, {ref}, {phase}
func formatStatusLine(data StatusLineData, tmpl string) string {
	s := tmpl
	s = strings.ReplaceAll(s, "{alias}", data.Alias)
	s = strings.ReplaceAll(s, "{node_name}", data.NodeName)
	s = strings.ReplaceAll(s, "{node_id}", data.NodeID)
	s = strings.ReplaceAll(s, "{flow}", data.Flow)
	s = strings.ReplaceAll(s, "{ref}", data.Ref)
	s = strings.ReplaceAll(s, "{phase}", data.Phase)
	return s
}

func substituteVars(path string, vars map[string]string) string {
	return SubstituteVarsInPath(path, vars)
}

func findPrincipalNode(fl *flow.Flow) string {
	for _, n := range fl.Nodes {
		if n.Type == flow.NodeTypePhase {
			var phaseCfg flow.PhaseConfig
			if n.Config != nil {
				_ = json.Unmarshal(n.Config, &phaseCfg)
			}
			if phaseCfg.Role != "" {
				return n.ID
			}
		}
	}
	if len(fl.Nodes) > 0 {
		return fl.Nodes[0].ID
	}
	return ""
}

type principalRoleInfo struct {
	roleID          string
	alias           string
	aliasEn         string
	roleName        string
	persona         string
	traits          []string
	guidance        string
	promptSource    string
	standardsSource string
	directives      []string
	roleRules       []string
}

func findPrincipalRole(fl *flow.Flow, team *flow.TeamDefinition) roleInfo {
	info := roleInfo{}

	if team != nil {
		for _, r := range team.Roles {
			if r.Principal != nil && *r.Principal {
				info.roleID = r.ID
				info.alias = r.Alias
				info.aliasEn = r.AliasEn
				info.roleName = r.Name
				info.persona = r.Persona
				info.traits = r.Traits
				info.guidance = r.Guidance
				info.promptSource = r.PromptSource
				info.standardsSource = r.StandardsSource
				info.directives = r.PromptDirectives
				info.roleRules = r.Rules
				return info
			}
		}
	}

	if fl.Components != nil {
		for _, r := range fl.Components.Roles {
			if r.Principal != nil && *r.Principal {
				info.roleID = r.ID
				info.alias = r.Alias
				info.aliasEn = r.AliasEn
				info.roleName = r.Name
				info.persona = r.Persona
				info.traits = r.Traits
				info.guidance = r.Guidance
				info.promptSource = r.PromptSource
				info.standardsSource = r.StandardsSource
				info.directives = r.PromptDirectives
				info.roleRules = r.Rules
				return info
			}
		}
	}

	return info
}

func appendSessionLog(projectRoot string, taskID string, node *flow.FlowNode, gateResults []GateCheckResult, passed bool, input, analysis, conclusion string) {
	docsInternal := config.ResolveInternalDocs(projectRoot)
	logDir := filepath.Join(docsInternal, "task", taskID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}

	logPath := filepath.Join(logDir, "SESSION.md")

	var sb strings.Builder

	data, err := os.ReadFile(logPath)
	if err != nil || len(data) == 0 {
		sb.WriteString("# Session Log\n\n")
	} else {
		sb.Write(data)
		if !strings.HasSuffix(string(data), "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	now := time.Now().Format("2006-01-02 15:04")
	status := "PASSED"
	if !passed {
		status = "FAILED"
	}

	sb.WriteString(fmt.Sprintf("## [%s] Gate: %s (%s) — %s\n\n", now, node.Name, node.ID, status))

	if input != "" {
		sb.WriteString(fmt.Sprintf("**User Input**: %s\n\n", input))
	}
	if analysis != "" {
		sb.WriteString(fmt.Sprintf("**AI Analysis**: %s\n\n", analysis))
	}
	if conclusion != "" {
		sb.WriteString(fmt.Sprintf("**Conclusion**: %s\n\n", conclusion))
	}

	for _, gr := range gateResults {
		icon := "✅"
		if !gr.Passed && !gr.Skipped {
			if gr.Required {
				icon = "❌"
			} else {
				icon = "⚠️"
			}
		} else if gr.Skipped {
			icon = "⏭️"
		}
		reqLabel := "required"
		if !gr.Required {
			reqLabel = "advisory"
		}
		sb.WriteString(fmt.Sprintf("- %s **%s** (%s): %s\n", icon, gr.Type, reqLabel, gr.Message))
	}

	sb.WriteString("\n")

	_ = os.WriteFile(logPath, []byte(sb.String()), 0644)
}

func resolvePhaseFromNode(node *flow.FlowNode) string {
	if node.Config == nil {
		return ""
	}
	var phaseCfg flow.PhaseConfig
	if err := json.Unmarshal(node.Config, &phaseCfg); err != nil {
		return ""
	}
	for _, a := range node.OnEnter {
		if a.Action == flow.ActionUpdateTaskPhase && a.Phase != "" {
			return a.Phase
		}
	}
	return ""
}

func ensureSessionName(lgr *eventlog.Logger, newSession bool, input string) string {
	if !newSession {
		name, _ := lgr.LastSessionName()
		if name != "" {
			return name
		}
	}
	name, err := lgr.CreateSession(1, "", input)
	if err != nil {
		return ""
	}
	return name
}
