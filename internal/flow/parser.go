package flow

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var defaultVariables = map[string]string{
	"docs_path":    "",
	"TEAM_PATH":    "",
	"task_id":      "",
	"SKILL_PATH":   "",
	"PROJECT_PATH": "",
}

func ParseFlow(data []byte) (*Flow, error) {
	var flow Flow
	if err := json.Unmarshal(data, &flow); err != nil {
		return nil, fmt.Errorf("parse flow: %w", err)
	}
	return &flow, nil
}

func ParseFlowWithVars(data []byte, vars map[string]string) (*Flow, error) {
	resolved := resolveVariablesInJSON(data, vars)
	var flow Flow
	if err := json.Unmarshal(resolved, &flow); err != nil {
		return nil, fmt.Errorf("parse flow: %w", err)
	}
	return &flow, nil
}

func ParseFlowFile(path string) (*Flow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read flow file %s: %w", path, err)
	}
	return ParseFlow(data)
}

func ParseFlowFileWithVars(path string, vars map[string]string) (*Flow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read flow file %s: %w", path, err)
	}
	return ParseFlowWithVars(data, vars)
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
