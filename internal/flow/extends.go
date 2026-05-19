package flow

import (
	"fmt"
)

func MergeFlow(base, child *Flow) (*Flow, error) {
	if base == nil {
		return child, nil
	}
	if child == nil {
		return base, nil
	}

	result := *base

	if child.Metadata.Name != "" {
		result.Metadata.Name = child.Metadata.Name
	}
	if child.Metadata.Description != "" {
		result.Metadata.Description = child.Metadata.Description
	}
	if child.Version != "" {
		result.Version = child.Version
	}

	if child.Config != nil {
		if result.Config == nil {
			result.Config = child.Config
		} else {
			mergeConfig(result.Config, child.Config)
		}
	}

	nodeMap := make(map[string]FlowNode, len(base.Nodes))
	for _, n := range base.Nodes {
		nodeMap[n.ID] = n
	}
	for _, n := range child.Nodes {
		nodeMap[n.ID] = n
	}
	result.Nodes = make([]FlowNode, 0, len(nodeMap))
	for _, n := range base.Nodes {
		if merged, ok := nodeMap[n.ID]; ok {
			result.Nodes = append(result.Nodes, merged)
			delete(nodeMap, n.ID)
		}
	}
	for _, n := range child.Nodes {
		if _, exists := nodeMap[n.ID]; exists {
			result.Nodes = append(result.Nodes, n)
		}
	}

	result.Edges = append(result.Edges, child.Edges...)

	return &result, nil
}

func mergeConfig(base, child *FlowConfig) {
	if child.TaskType != "" {
		base.TaskType = child.TaskType
	}
	if child.Domain != "" {
		base.Domain = child.Domain
	}
	if child.Extends != "" {
		base.Extends = child.Extends
	}
	if child.AutoDispatch != nil {
		base.AutoDispatch = child.AutoDispatch
	}
	if child.ParallelLimit != nil {
		base.ParallelLimit = child.ParallelLimit
	}
	if child.TimeoutMinutes != nil {
		base.TimeoutMinutes = child.TimeoutMinutes
	}
}

func ResolveExtends(loader func(string) (*Flow, error), current *Flow) (*Flow, error) {
	if current.Config == nil || current.Config.Extends == "" {
		return current, nil
	}
	base, err := loader(current.Config.Extends)
	if err != nil {
		return nil, fmt.Errorf("resolve extends %s: %w", current.Config.Extends, err)
	}
	base, err = ResolveExtends(loader, base)
	if err != nil {
		return nil, err
	}
	return MergeFlow(base, current)
}
