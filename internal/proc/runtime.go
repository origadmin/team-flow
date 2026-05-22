package proc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/flow"
)

func LoadFlow(root, flowID string) (*flow.Flow, error) {
	procPath := resolveProcPath(root, flowID)
	if procPath == "" {
		return nil, fmt.Errorf("flow not found: %s", flowID)
	}

	f, err := flow.ParseFlowFile(procPath)
	if err != nil {
		return nil, fmt.Errorf("parse error in %s: %w", procPath, err)
	}

	return f, nil
}

func FindNode(f *flow.Flow, nodeID string) (*flow.FlowNode, error) {
	if len(f.Nodes) == 0 {
		return nil, fmt.Errorf("flow has no nodes")
	}

	if nodeID == "" {
		return FindRootNode(f)
	}

	for i := range f.Nodes {
		if f.Nodes[i].ID == nodeID {
			return &f.Nodes[i], nil
		}
	}

	available := make([]string, 0, len(f.Nodes))
	for _, n := range f.Nodes {
		available = append(available, n.ID)
	}

	return nil, fmt.Errorf("node not found: %s. Available: %s", nodeID, strings.Join(available, ", "))
}

func FindRootNode(f *flow.Flow) (*flow.FlowNode, error) {
	for i := range f.Nodes {
		if f.Nodes[i].Type == flow.NodeTypePhase || f.Nodes[i].Type == flow.NodeTypeStart {
			return &f.Nodes[i], nil
		}
	}

	if len(f.Nodes) > 0 {
		return &f.Nodes[0], nil
	}

	return nil, fmt.Errorf("flow has no nodes")
}

func GetOutgoingEdges(f *flow.Flow, nodeID string) []flow.FlowEdge {
	var edges []flow.FlowEdge
	for _, e := range f.Edges {
		if e.From == nodeID {
			edges = append(edges, e)
		}
	}
	return edges
}

func GetNodeRole(node *flow.FlowNode) string {
	if node.Config == nil {
		return ""
	}
	var phaseCfg flow.PhaseConfig
	if err := json.Unmarshal(node.Config, &phaseCfg); err == nil {
		return phaseCfg.Role
	}
	return ""
}

func ResolveDefaultFlowName(root string) (string, error) {
	projectMD := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMD)
	if err != nil {
		return "", fmt.Errorf("no default flow configured. Use --flow flag")
	}

	name := extractFlowNameFromProjectMD(data)
	if name == "" {
		return "", fmt.Errorf("no default flow configured. Use --flow flag")
	}

	return name, nil
}

func ResolveDocsPath(root string) string {
	projectMD := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMD)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "docs_internal:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, "\"' ")
				return val
			}
		}
	}
	return ""
}

func ResolveDocsExternalPath(root string) string {
	projectMD := filepath.Join(root, ".team", "project.md")
	data, err := os.ReadFile(projectMD)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "docs_external:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, "\"' ")
				return val
			}
		}
	}
	return ""
}
