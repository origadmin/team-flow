package flow

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

var defaultVariables = map[string]string{
	"DOCS_INTERNAL": "",
	"DOCS_EXTERNAL": "",
	"TEAM_PATH":     "",
	"task_id":       "",
	"SKILL_PATH":    "",
	"PROJECT_PATH":  "",
}

type flowNodeRaw struct {
	ID          string          `json:"id"`
	Type        NodeType        `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Entry       bool            `json:"entry,omitempty"`
	Status      string          `json:"status,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Components  *NodeComponents `json:"components,omitempty"`
	Docs        []DocSpec       `json:"docs,omitempty"`
	Gates       []GateConfig    `json:"gates,omitempty"`
	Conditions  []GateCondition `json:"conditions,omitempty"`
	OnEnter     []Action        `json:"on_enter,omitempty"`
	OnExit      []Action        `json:"on_exit,omitempty"`
	OnError     *ErrorHandler   `json:"on_error,omitempty"`
	Role        string          `json:"role,omitempty"`
	Rules       []string        `json:"rules,omitempty"`
	Tools       []ToolRef       `json:"tools,omitempty"`
	Skills      []ComponentRef  `json:"skills,omitempty"`
}

type flowRaw struct {
	Version    string                 `json:"version"`
	Metadata   FlowMetadata           `json:"metadata"`
	Config     *FlowConfig            `json:"config,omitempty"`
	Extends    string                 `json:"extends,omitempty"`
	Overrides  []FlowNodeOverride     `json:"overrides,omitempty"`
	Components *ComponentRegistry     `json:"components,omitempty"`
	Nodes      []flowNodeRaw          `json:"nodes"`
	Edges      []FlowEdge             `json:"edges"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
}

func ParseFlow(data []byte) (*Flow, error) {
	var raw flowRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse flow: %w", err)
	}

	nodes := make([]FlowNode, len(raw.Nodes))
	for i, rn := range raw.Nodes {
		nodes[i] = normalizeNode(rn)
	}

	edges := make([]FlowEdge, len(raw.Edges))
	for i := range raw.Edges {
		raw.Edges[i].Normalize()
		edges[i] = raw.Edges[i]
	}

	return &Flow{
		Version:    raw.Version,
		Metadata:   raw.Metadata,
		Config:     raw.Config,
		Extends:    raw.Extends,
		Overrides:  raw.Overrides,
		Components: raw.Components,
		Nodes:      nodes,
		Edges:      edges,
		Variables:  raw.Variables,
	}, nil
}

func normalizeNode(rn flowNodeRaw) FlowNode {
	node := FlowNode{
		ID:          rn.ID,
		Type:        rn.Type,
		Name:        rn.Name,
		Description: rn.Description,
		Entry:       rn.Entry,
		Status:      rn.Status,
		Docs:        rn.Docs,
		Gates:       rn.Gates,
		Conditions:  rn.Conditions,
		OnEnter:     rn.OnEnter,
		OnExit:      rn.OnExit,
		OnError:     rn.OnError,
	}

	if len(rn.Config) > 0 {
		node.Config = rn.Config
	}

	if rn.Components != nil {
		node.Components = rn.Components
		return node
	}

	if rn.Role != "" && len(rn.Config) == 0 {
		phaseCfg := PhaseConfig{Role: rn.Role}
		if cfgData, err := json.Marshal(phaseCfg); err == nil {
			node.Config = cfgData
		}
	}

	comps := &NodeComponents{}
	hasComps := false

	if rn.Role != "" {
		comps.Roles = []ComponentRef{{Ref: rn.Role, Source: SourceBuiltin}}
		hasComps = true
	}

	if len(rn.Rules) > 0 {
		comps.Rules = make([]ComponentRef, len(rn.Rules))
		for i, r := range rn.Rules {
			comps.Rules[i] = ComponentRef{Ref: r, Source: SourceBuiltin}
		}
		hasComps = true
	}

	if len(rn.Tools) > 0 {
		comps.Tools = rn.Tools
		hasComps = true
	}

	if len(rn.Skills) > 0 {
		comps.Skills = rn.Skills
		hasComps = true
	}

	if hasComps {
		node.Components = comps
	}

	return node
}

func ParseFlowWithVars(data []byte, vars map[string]string) (*Flow, error) {
	resolved := resolveVariablesInJSON(data, vars)
	return ParseFlow(resolved)
}

func ParseFlowFile(path string) (*Flow, error) {
	data, err := readFlowBytes(path)
	if err != nil {
		return nil, err
	}
	return ParseFlow(data)
}

// ParseFlowFileWithVars reads the flow file (disk or template) and applies vars.
func ParseFlowFileWithVars(path string, vars map[string]string) (*Flow, error) {
	data, err := readFlowBytes(path)
	if err != nil {
		return nil, err
	}
	resolved := resolveVariablesInJSON(data, vars)
	return ParseFlow(resolved)
}

func readFlowBytes(path string) ([]byte, error) {
	if strings.HasPrefix(path, "embed://templates/") {
		id := strings.TrimPrefix(path, "embed://templates/")
		id = strings.TrimSuffix(id, ".json")
		loader := getTemplateLoader()
		if loader == nil {
			return nil, fmt.Errorf("template %q requested but no template loader registered; path=%s", id, path)
		}
		return loader(id)
	}
	return os.ReadFile(path)
}

// TemplateLoader is the contract between the flow parser and the
// embedded template store. The flow package itself does not depend on
// the templates package; instead, the proc package registers a loader
// at init() time.
type TemplateLoader func(id string) ([]byte, error)

var (
	globalTemplateLoader   TemplateLoader
	globalTemplateLoaderMu sync.RWMutex
)

// RegisterTemplateLoader sets the function used to resolve template ids
// when the flow parser encounters an embed://templates/<id>.json path.
// Pass nil to clear the loader (mainly for tests).
func RegisterTemplateLoader(loader TemplateLoader) {
	globalTemplateLoaderMu.Lock()
	defer globalTemplateLoaderMu.Unlock()
	globalTemplateLoader = loader
}

func getTemplateLoader() TemplateLoader {
	globalTemplateLoaderMu.RLock()
	defer globalTemplateLoaderMu.RUnlock()
	return globalTemplateLoader
}

func ApplyOverrides(base *Flow, overrides FlowOverrides) *Flow {
	result := *base

	if overrides.Config != nil {
		result.Config = overrides.Config
	}

	if len(overrides.Nodes) > 0 {
		overrideMap := make(map[string]FlowNodeOverride, len(overrides.Nodes))
		for _, o := range overrides.Nodes {
			overrideMap[o.ID] = o
		}
		nodes := make([]FlowNode, len(base.Nodes))
		copy(nodes, base.Nodes)
		for i := range nodes {
			if o, ok := overrideMap[nodes[i].ID]; ok {
				nodes[i] = applyNodeOverride(nodes[i], o)
			}
		}
		result.Nodes = nodes
	}

	if len(overrides.Edges) > 0 {
		result.Edges = overrides.Edges
	}

	if len(overrides.Variables) > 0 {
		if result.Variables == nil {
			result.Variables = make(map[string]interface{})
		}
		for k, v := range overrides.Variables {
			result.Variables[k] = v
		}
	}

	if overrides.Components != nil {
		result.Components = overrides.Components
	}

	return &result
}

func applyNodeOverride(node FlowNode, o FlowNodeOverride) FlowNode {
	if o.Name != "" {
		node.Name = o.Name
	}
	if o.Description != "" {
		node.Description = o.Description
	}
	if len(o.Config) > 0 {
		node.Config = o.Config
	}
	if o.Components != nil {
		node.Components = o.Components
	}
	if len(o.Docs) > 0 {
		node.Docs = o.Docs
	}
	if len(o.Gates) > 0 {
		node.Gates = o.Gates
	}
	if len(o.OnEnter) > 0 {
		node.OnEnter = o.OnEnter
	}
	if len(o.OnExit) > 0 {
		node.OnExit = o.OnExit
	}
	if o.OnError != nil {
		node.OnError = o.OnError
	}
	return node
}

func resolveVariablesInJSON(data []byte, vars map[string]string) []byte {
	s := string(data)
	for key, val := range vars {
		placeholder := "{" + key + "}"
		s = strings.ReplaceAll(s, placeholder, val)
	}
	return []byte(s)
}

func MergeVariables(base, override map[string]interface{}) map[string]interface{} {
	if base == nil && override == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}
