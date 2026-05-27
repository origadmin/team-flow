package proc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/flow"
	"github.com/origadmin/team-flow/internal/skill"
	"github.com/origadmin/team-flow/internal/state"
)

type ProcRunRequest struct {
	FlowName    string
	NodeID      string
	TaskID      string
	ProjectRoot string
	TeamRoot    string
	Workspace   string
	Format      string
	RunGate     bool
}

type ProcRunResult struct {
	Flow            FlowMeta          `json:"flow"`
	Current         CurrentNode       `json:"current"`
	NextOptions     []NextOption      `json:"next_options"`
	StatusLine      string            `json:"status_line"`
	Task            *TaskInfo         `json:"task,omitempty"`
	TeamIntro       *TeamIntroData    `json:"team_intro,omitempty"`
	ProjectRoot     string            `json:"project_root,omitempty"`
	TeamRoot        string            `json:"team_root,omitempty"`
	Workspace       string            `json:"workspace,omitempty"`
	PathValidation  *PathValidation   `json:"path_validation,omitempty"`
	Skills          *SkillContext      `json:"skills,omitempty"`
	GateCheckResults []GateCheckResult `json:"gate_check_results,omitempty"`
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
	TeamID          string              `json:"team_id"`
	TeamName        string              `json:"team_name"`
	TeamNameZh      string              `json:"team_name_zh,omitempty"`
	TeamDescription string              `json:"team_description,omitempty"`
	FlowName        string              `json:"flow_name"`
	Roles           []TeamIntroRole     `json:"roles"`
	Flows           []TeamIntroFlow     `json:"flows"`
	HowItWorks      []string            `json:"how_it_works"`
}

type TeamIntroRole struct {
	Alias    string `json:"alias"`
	AliasEn  string `json:"alias_en"`
	RoleName string `json:"role_name"`
	Principal bool  `json:"principal,omitempty"`
}

type TeamIntroFlow struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

type TaskInfo struct {
	TaskID string `json:"task_id,omitempty"`
	Phase  string `json:"phase,omitempty"`
	Status string `json:"status,omitempty"`
	URL    string `json:"url,omitempty"`
}

type FlowMeta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Domain  string `json:"domain"`
}

type ParallelBranchOutput struct {
	NodeID   string `json:"node_id"`
	RoleName string `json:"role_name,omitempty"`
	Alias    string `json:"alias,omitempty"`
	Name     string `json:"name,omitempty"`
}

type CurrentNode struct {
	NodeID            string                  `json:"node_id"`
	NodeType          string                  `json:"node_type"`
	Name              string                  `json:"name"`
	Description       string                  `json:"description"`
	Role              string                  `json:"role"`
	RoleName          string                  `json:"role_name,omitempty"`
	Alias             string                  `json:"alias,omitempty"`
	AliasEn           string                  `json:"alias_en,omitempty"`
	Principal         bool                    `json:"principal,omitempty"`
	Persona           string                  `json:"persona,omitempty"`
	Traits            []string                `json:"traits,omitempty"`
	Guidance          string                  `json:"guidance,omitempty"`
	PromptSource      string                  `json:"prompt_source,omitempty"`
	StandardsSource   string                  `json:"standards_source,omitempty"`
	PromptDirectives  []string                `json:"prompt_directives,omitempty"`
	Rules             []RuleOutput            `json:"rules"`
	Tools             []ToolOutput            `json:"tools"`
	Skills            []SkillOutput           `json:"skills"`
	Prompts           []PromptOutput          `json:"prompts"`
	Docs              []DocOutput             `json:"docs"`
	OnEnter           []OnEnterAction         `json:"on_enter"`
	GateConditions    []GateCondOutput        `json:"gate_conditions"`
	ParallelBranches  []ParallelBranchOutput  `json:"parallel_branches,omitempty"`
	ParallelStrategy  string                  `json:"parallel_strategy,omitempty"`
	MergeStrategy     string                  `json:"merge_strategy,omitempty"`
	SubflowRef        string                  `json:"subflow_ref,omitempty"`
	IsTerminal        bool                    `json:"is_terminal"`
	TerminalStatus    string                  `json:"terminal_status,omitempty"`
	TerminalMessage   string                  `json:"terminal_message,omitempty"`
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
	Action string `json:"action"`
	Phase  string `json:"phase,omitempty"`
}

type GateCondOutput struct {
	Type         string   `json:"type"`
	Threshold    string   `json:"threshold,omitempty"`
	Required     bool     `json:"required"`
	Check        string   `json:"check,omitempty"`
	Expected     string   `json:"expected,omitempty"`
	Deliverables []string `json:"deliverables,omitempty"`
}

type NextOption struct {
	NodeID    string  `json:"node_id"`
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
}

func NewProcRunEngine(root string) *ProcRunEngine {
	return &ProcRunEngine{
		FlowResolver:   &DefaultFlowResolver{Root: root},
		NodeResolver:   &DefaultNodeResolver{},
		VarSubstitutor: &DefaultVarSubstitutor{},
		TeamLoader:     &DefaultTeamLoader{},
	}
}

func (e *ProcRunEngine) Run(ctx context.Context, req ProcRunRequest) (*ProcRunResult, error) {
	fl, err := e.FlowResolver.Resolve(ctx, req.FlowName, req.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve flow: %w", err)
	}

	node, err := e.NodeResolver.Resolve(ctx, fl, req.NodeID)
	if err != nil {
		return nil, fmt.Errorf("resolve node: %w", err)
	}

	flowName := fl.Metadata.Name

	if req.NodeID != "" && node.Type != flow.NodeTypeGate {
		prereqGates := findPrerequisiteGates(fl, req.NodeID)
		for _, gateID := range prereqGates {
			if !state.IsGatePassed(req.ProjectRoot, flowName, gateID) {
				return nil, fmt.Errorf("GATE BLOCKED: prerequisite gate %s has not been passed in flow '%s' — run 'flow proc run %s' first to pass the gate", gateID, flowName, gateID)
			}
		}
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

	result := generateResult(fl, node, vars, team, req.NodeID == "", req.ProjectRoot, req.TeamRoot, req.Workspace, req.RunGate)

	if req.RunGate && node.Type == flow.NodeTypeGate && len(result.Current.GateConditions) > 0 {
		substitutedConds := SubstituteGateConditions(result.Current.GateConditions, vars)
		docsInternal := config.ResolveInternalDocs(req.ProjectRoot)
		gateResults := RunGateCheck(req.ProjectRoot, substitutedConds, team, docsInternal)
		result.GateCheckResults = gateResults

		passed, summary := GateOverallResult(gateResults)
		_ = state.RecordGateResult(req.ProjectRoot, flowName, node.ID, passed, summary)

		if passed {
			_ = state.RecordVisitedNode(req.ProjectRoot, flowName, node.ID)
		}
	} else if node.Type != flow.NodeTypeGate {
		_ = state.RecordVisitedNode(req.ProjectRoot, flowName, node.ID)
	}

	if taskInfo := resolveTaskInfo(req.TaskID); taskInfo != nil {
		result.Task = taskInfo
		if taskInfo.TaskID != "" {
			result.StatusLine = buildStatusLineWithRef(fl, node, result.Current, taskInfo.TaskID)
		}
	}

	return result, nil
}

// resolveTaskInfo queries the beads backend via flow task CLI for current task state.
// Returns nil if beads is not available or no current task exists.
func resolveTaskInfo(taskID string) *TaskInfo {
	if !bd.IsAvailable() {
		return nil
	}

	args := []string{"show", "--current", "--json"}
	if taskID != "" {
		args = []string{"show", taskID, "--json"}
	}

	output, err := bd.RunQuiet(args...)
	if err != nil || output == "" {
		return nil
	}

	// Parse JSON output from bd CLI
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return nil
	}

	info := &TaskInfo{}
	if id, ok := raw["id"].(string); ok {
		info.TaskID = id
	}
	if status, ok := raw["status"].(string); ok {
		info.Status = status
	}
	if url, ok := raw["url"].(string); ok {
		info.URL = url
	}

	// Extract phase from labels (phase:xxx)
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

func generateResult(fl *flow.Flow, node *flow.FlowNode, vars map[string]string, team *flow.TeamDefinition, isFirstNode bool, projectRoot string, teamRoot string, workspace string, runGate bool) *ProcRunResult {
	domain := ""
	if fl.Config != nil {
		domain = fl.Config.Domain
		if domain == "" {
			domain = string(fl.Config.TaskType)
		}
	}

	current := buildCurrentNode(node, fl, vars, team)

	configRoot := teamRoot
	if configRoot == "" {
		configRoot = projectRoot
	}

	result := &ProcRunResult{
		Flow: FlowMeta{
			Name:    fl.Metadata.Name,
			Version: fl.Version,
			Domain:  domain,
		},
		Current:     current,
		StatusLine:  buildStatusLine(fl, node, current, vars),
		ProjectRoot: projectRoot,
		TeamRoot:    configRoot,
		Workspace:   workspace,
	}

	if runGate && node.Type == flow.NodeTypeGate && len(current.GateConditions) > 0 {
		substitutedConds := SubstituteGateConditions(current.GateConditions, vars)
		docsInternal := config.ResolveInternalDocs(projectRoot)
		result.GateCheckResults = RunGateCheck(projectRoot, substitutedConds, team, docsInternal)
	}

	result.NextOptions = buildNextOptions(fl, node.ID)

	if isFirstNode {
		result.TeamIntro = buildTeamIntro(fl, team, projectRoot)
	}

	if result.Current.IsTerminal {
		result.NextOptions = []NextOption{}
	}

	result.PathValidation = validatePath(projectRoot, teamRoot)

	// Load skills for the current role
	if projectRoot != "" {
		skillMgr := skill.NewSkillManager(projectRoot, "", "")
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

	return result
}

func validatePath(projectRoot string, teamRoot string) *PathValidation {
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
		EnforceLevel:   "warn",
		Warning:        warning,
		RequiredAction: requiredAction,
	}
}

func buildCurrentNode(node *flow.FlowNode, fl *flow.Flow, vars map[string]string, team *flow.TeamDefinition) CurrentNode {
	current := CurrentNode{
		NodeID:      node.ID,
		NodeType:    string(node.Type),
		Name:        node.Name,
		Description: node.Description,
		IsTerminal:  node.Type == flow.NodeTypeTerminal,
	}

	switch node.Type {
	case flow.NodeTypePhase, flow.NodeTypeStart:
		ri := resolveRoleInfo(node, fl, team)
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
		current.Rules = extractRules(node, fl, team, ri.roleRules)
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

	case flow.NodeTypeSubflow:
		current.GateConditions = nil
		if node.Config != nil {
			var sfCfg flow.SubflowConfig
			if err := json.Unmarshal(node.Config, &sfCfg); err == nil {
				current.SubflowRef = sfCfg.FlowRef
			}
		}
		current.OnEnter = extractOnEnter(node)

	case flow.NodeTypeTerminal:
		current.Role = ""
		current.IsTerminal = true
		current.GateConditions = nil
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
						ri := resolveRoleInfo(&fl.Nodes[i], fl, team)
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
		ri := resolveRoleInfo(node, fl, team)
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
		current.Rules = extractRules(node, fl, team, ri.roleRules)
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

type roleInfo struct {
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

func resolveRoleInfo(node *flow.FlowNode, fl *flow.Flow, team *flow.TeamDefinition) roleInfo {
	info := roleInfo{}
	if node.Config == nil {
		return info
	}
	var phaseCfg flow.PhaseConfig
	if err := json.Unmarshal(node.Config, &phaseCfg); err != nil {
		return info
	}
	if node.Components == nil || len(node.Components.Roles) == 0 {
		return info
	}
	roleRef := node.Components.Roles[0].Ref

	if team != nil {
		for _, r := range team.Roles {
			if r.ID == roleRef {
				info.promptSource = r.PromptSource
				info.standardsSource = r.StandardsSource
				info.directives = r.PromptDirectives
				info.persona = r.Persona
				info.traits = r.Traits
				info.guidance = r.Guidance
				info.alias = r.Alias
				info.aliasEn = r.AliasEn
				info.roleName = r.Name
				info.roleRules = r.Rules
				if r.Principal != nil && *r.Principal {
					info.principal = true
				}
				return info
			}
		}
	}

	if fl.Components == nil {
		return info
	}
	for _, r := range fl.Components.Roles {
		if r.ID == roleRef {
			info.promptSource = r.PromptSource
			info.standardsSource = r.StandardsSource
			info.directives = r.PromptDirectives
			info.persona = r.Persona
			info.traits = r.Traits
			info.guidance = r.Guidance
			info.alias = r.Alias
			info.aliasEn = r.AliasEn
			info.roleName = r.Name
			info.roleRules = r.Rules
			if r.Principal != nil && *r.Principal {
				info.principal = true
			}
			break
		}
	}
	return info
}

// extractPrompts returns prompt file references from node components.
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

func extractRules(node *flow.FlowNode, fl *flow.Flow, team *flow.TeamDefinition, roleRules []string) []RuleOutput {
	if node.Components == nil && len(roleRules) == 0 {
		return nil
	}
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
			out := RuleOutput{
				Ref:    r.Ref,
				Source: string(r.Source),
			}
			if def, ok := ruleDefs[r.Ref]; ok {
				out.Name = def.Name
				out.Instruction = def.Instruction
				out.Enforcement = string(def.Enforcement)
				out.RuleRef = fmt.Sprintf("flow proc rule %s", def.ID)
			}
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
		out := RuleOutput{
			Ref:    ref,
			Source: "role",
		}
		if def, ok := ruleDefs[ref]; ok {
			out.Name = def.Name
			out.Instruction = def.Instruction
			out.Enforcement = string(def.Enforcement)
			out.RuleRef = fmt.Sprintf("flow proc rule %s", def.ID)
		}
		rules = append(rules, out)
	}
	if len(rules) == 0 {
		return nil
	}
	return rules
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

// extractSkills resolves skill references from node components and enriches them
// with trigger, path, and description from the flow-level skill registry.
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
			Action: string(a.Action),
			Phase:  a.Phase,
		})
	}
	return actions
}

func buildNextOptions(fl *flow.Flow, currentNodeID string) []NextOption {
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
		return buildNextOptionsFromEdges(outgoingEdges, nodeMap)
	}

	currentNode := nodeMap[currentNodeID]
	if currentNode != nil && currentNode.Type == flow.NodeTypeGate {
		return BuildNextOptionsFromGateFallback(currentNode, outgoingEdges, nodeMap)
	}

	return buildNextOptionsFromEdges(outgoingEdges, nodeMap)
}

func buildNextOptionsFromEdges(edges []flow.FlowEdge, nodeMap map[string]*flow.FlowNode) []NextOption {
	options := make([]NextOption, 0, len(edges))
	defaultSet := false

	for _, e := range edges {
		role := ""
		if target, ok := nodeMap[e.To]; ok {
			role = extractRole(target)
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

// buildStatusLine generates the v2-compatible status line: [Role|TaskPool|Phase|Asset]
// This provides a consistent, machine-readable progress indicator that AI agents
// should include in every response during flow execution.
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
			ri := resolveRoleInfo(n, fl, team)
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

func loadAvailableFlows(defaultFlow, projectRoot string) []TeamIntroFlow {
	var flows []TeamIntroFlow
	flowsDir := filepath.Join(projectRoot, ".team", "flows")
	entries, err := os.ReadDir(flowsDir)
	if err != nil {
		flows = append(flows, TeamIntroFlow{ID: defaultFlow, IsDefault: true})
		return flows
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		flowID := strings.TrimSuffix(entry.Name(), ".json")
		flows = append(flows, TeamIntroFlow{
			ID:        flowID,
			IsDefault: flowID == defaultFlow,
		})
	}
	if len(flows) == 0 {
		flows = append(flows, TeamIntroFlow{ID: defaultFlow, IsDefault: true})
	}
	return flows
}

func buildStatusLine(fl *flow.Flow, node *flow.FlowNode, current CurrentNode, vars map[string]string) string {
	return buildStatusLineWithRef(fl, node, current, "")
}

func buildStatusLineWithRef(fl *flow.Flow, node *flow.FlowNode, current CurrentNode, ref string) string {
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
		phase = current.Name
	}

	return fmt.Sprintf("[%s | %s(%s:%s) | %s | %s]", alias, current.Name, node.ID, flowPart, ref, phase)
}

func substituteVars(path string, vars map[string]string) string {
	return SubstituteVarsInPath(path, vars)
}

func findPrerequisiteGates(fl *flow.Flow, nodeID string) []string {
	nodeMap := make(map[string]bool, len(fl.Nodes))
	for _, n := range fl.Nodes {
		if n.Type == flow.NodeTypeGate {
			nodeMap[n.ID] = true
		}
	}

	var gateIDs []string
	visited := make(map[string]bool)
	var walk func(current string)
	walk = func(current string) {
		if visited[current] {
			return
		}
		visited[current] = true
		for _, e := range fl.Edges {
			if e.To == current {
				if nodeMap[e.From] {
					gateIDs = append(gateIDs, e.From)
				}
				walk(e.From)
			}
		}
	}
	walk(nodeID)
	return gateIDs
}
