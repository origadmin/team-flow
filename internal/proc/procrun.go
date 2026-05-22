package proc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/bd"
	"github.com/origadmin/team-flow/internal/flow"
)

type ProcRunRequest struct {
	FlowName    string
	NodeID      string
	TaskID      string
	ProjectRoot string
	Format      string
}

type ProcRunResult struct {
	Flow        FlowMeta     `json:"flow"`
	Current     CurrentNode  `json:"current"`
	NextOptions []NextOption `json:"next_options"`
	StatusLine  string       `json:"status_line"`
	Task        *TaskInfo    `json:"task,omitempty"`
	TeamIntro   *TeamIntroData `json:"team_intro,omitempty"`
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
	BeadsID string `json:"beads_id,omitempty"`
	Phase   string `json:"phase,omitempty"`
	Status  string `json:"status,omitempty"`
	URL     string `json:"url,omitempty"`
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
	Type      string `json:"type"`
	Threshold string `json:"threshold,omitempty"`
	Required  bool   `json:"required"`
	Check     string `json:"check,omitempty"`
	Expected  string `json:"expected,omitempty"`
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

	result := generateResult(fl, node, vars, team, req.NodeID == "", req.ProjectRoot)

	// Integrate beads task state if available
	if taskInfo := resolveTaskInfo(req.TaskID); taskInfo != nil {
		result.Task = taskInfo
		// Override StatusLine with beads data when available
		if taskInfo.BeadsID != "" {
			phase := taskInfo.Phase
			if phase == "" {
				for _, a := range result.Current.OnEnter {
					if a.Action == "update_task_phase" && a.Phase != "" {
						phase = a.Phase
						break
					}
				}
			}
			role := result.Current.Role
			if role == "" {
				role = "System"
			}
			domain := string(fl.Config.TaskType)
			if domain == "" {
				domain = fl.Metadata.Name
			}
			asset := ""
			if len(result.Current.Docs) > 0 {
				asset = result.Current.Docs[0].Name
			}
			result.StatusLine = fmt.Sprintf("[%s|%s|%s|%s]", role, taskInfo.BeadsID, phase, asset)
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
		info.BeadsID = id
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

	if info.BeadsID == "" && info.Phase == "" {
		return nil
	}

	return info
}

func generateResult(fl *flow.Flow, node *flow.FlowNode, vars map[string]string, team *flow.TeamDefinition, isFirstNode bool, projectRoot string) *ProcRunResult {
	domain := ""
	if fl.Config != nil {
		domain = fl.Config.Domain
		if domain == "" {
			domain = string(fl.Config.TaskType)
		}
	}

	current := buildCurrentNode(node, fl, vars, team)

	result := &ProcRunResult{
		Flow: FlowMeta{
			Name:    fl.Metadata.Name,
			Version: fl.Version,
			Domain:  domain,
		},
		Current:    current,
		StatusLine: buildStatusLine(fl, node, current, vars),
	}

	result.NextOptions = buildNextOptions(fl, node.ID)

	if isFirstNode {
		result.TeamIntro = buildTeamIntro(fl, team, projectRoot)
	}

	if result.Current.IsTerminal {
		result.NextOptions = []NextOption{}
	}

	return result
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
	role := current.Role
	if role == "" {
		role = "System"
	}

	taskPool := string(fl.Config.TaskType)
	if taskPool == "" {
		taskPool = fl.Metadata.Name
	}

	phase := ""
	for _, a := range current.OnEnter {
		if a.Action == "update_task_phase" && a.Phase != "" {
			phase = a.Phase
		}
	}
	if phase == "" {
		phase = node.Name
	}

	asset := ""
	if len(current.Docs) > 0 {
		asset = current.Docs[0].Name
	}

	return fmt.Sprintf("[%s|%s|%s|%s]", role, taskPool, phase, asset)
}

func substituteVars(path string, vars map[string]string) string {
	return SubstituteVarsInPath(path, vars)
}
