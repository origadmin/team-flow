package proc

import (
	"context"
	"encoding/json"
	"fmt"

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
}

type FlowMeta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Domain  string `json:"domain"`
}

type CurrentNode struct {
	NodeID          string            `json:"node_id"`
	NodeType        string            `json:"node_type"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Role            string            `json:"role"`
	Rules           []RuleOutput      `json:"rules"`
	Tools           []ToolOutput      `json:"tools"`
	Skills          []SkillOutput     `json:"skills"`
	Docs            []DocOutput       `json:"docs"`
	OnEnter         []OnEnterAction   `json:"on_enter"`
	GateConditions  []GateCondOutput  `json:"gate_conditions"`
	IsTerminal      bool              `json:"is_terminal"`
	TerminalStatus  string            `json:"terminal_status,omitempty"`
	TerminalMessage string            `json:"terminal_message,omitempty"`
}

type RuleOutput struct {
	Ref         string `json:"ref"`
	Source      string `json:"source,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Enforcement string `json:"enforcement,omitempty"`
}

type ToolOutput struct {
	Ref      string   `json:"ref"`
	Source   string   `json:"source,omitempty"`
	Commands []string `json:"commands,omitempty"`
}

type SkillOutput struct {
	Ref    string `json:"ref"`
	Source string `json:"source,omitempty"`
}

type DocOutput struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Format      string `json:"format"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
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

type ProcRunEngine struct {
	FlowResolver   FlowResolver
	NodeResolver   NodeResolver
	VarSubstitutor VarSubstitutor
}

func NewProcRunEngine(root string) *ProcRunEngine {
	return &ProcRunEngine{
		FlowResolver:   &DefaultFlowResolver{Root: root},
		NodeResolver:   &DefaultNodeResolver{},
		VarSubstitutor: &DefaultVarSubstitutor{},
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

	result := generateResult(fl, node, vars)

	return result, nil
}

func generateResult(fl *flow.Flow, node *flow.FlowNode, vars map[string]string) *ProcRunResult {
	domain := ""
	if fl.Config != nil {
		domain = fl.Config.Domain
		if domain == "" {
			domain = string(fl.Config.TaskType)
		}
	}

	result := &ProcRunResult{
		Flow: FlowMeta{
			Name:    fl.Metadata.Name,
			Version: fl.Version,
			Domain:  domain,
		},
		Current: buildCurrentNode(node, fl, vars),
	}

	result.NextOptions = buildNextOptions(fl, node.ID)

	if result.Current.IsTerminal {
		result.NextOptions = []NextOption{}
	}

	return result
}

func buildCurrentNode(node *flow.FlowNode, fl *flow.Flow, vars map[string]string) CurrentNode {
	current := CurrentNode{
		NodeID:      node.ID,
		NodeType:    string(node.Type),
		Name:        node.Name,
		Description: node.Description,
		IsTerminal:  node.Type == flow.NodeTypeTerminal,
	}

	switch node.Type {
	case flow.NodeTypePhase, flow.NodeTypeStart:
		current.Role = extractRole(node)
		current.Rules = extractRules(node, fl)
		current.Tools = extractTools(node)
		current.Skills = extractSkills(node)
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

	default:
		current.Role = extractRole(node)
		current.Rules = extractRules(node, fl)
		current.Tools = extractTools(node)
		current.Skills = extractSkills(node)
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

func extractRules(node *flow.FlowNode, fl *flow.Flow) []RuleOutput {
	if node.Components == nil {
		return nil
	}
	ruleDefs := make(map[string]flow.RuleDefinition)
	if fl.Components != nil {
		for _, r := range fl.Components.Rules {
			ruleDefs[r.ID] = r
		}
	}
	rules := make([]RuleOutput, 0, len(node.Components.Rules))
	for _, r := range node.Components.Rules {
		out := RuleOutput{
			Ref:    r.Ref,
			Source: string(r.Source),
		}
		if def, ok := ruleDefs[r.Ref]; ok {
			out.Name = def.Name
			out.Description = def.Description
			out.Enforcement = string(def.Enforcement)
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

func extractSkills(node *flow.FlowNode) []SkillOutput {
	if node.Components == nil {
		return nil
	}
	skills := make([]SkillOutput, 0, len(node.Components.Skills))
	for _, s := range node.Components.Skills {
		skills = append(skills, SkillOutput{
			Ref:    s.Ref,
			Source: string(s.Source),
		})
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
		docs = append(docs, DocOutput{
			Name:        d.Name,
			Path:        path,
			Format:      d.Format,
			Required:    required,
			Description: d.Description,
		})
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

func substituteVars(path string, vars map[string]string) string {
	return SubstituteVarsInPath(path, vars)
}
