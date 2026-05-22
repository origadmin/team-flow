package proc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/flow"
	"gopkg.in/yaml.v3"
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
	cfg, err := loadProjectConfig(root)
	if err == nil && cfg.DefaultFlow != "" {
		return cfg.DefaultFlow, nil
	}

	projectMD := filepath.Join(root, ".team", "project.md")
	data, mdErr := os.ReadFile(projectMD)
	if mdErr != nil {
		return "", fmt.Errorf("no default flow configured. Use --flow flag")
	}

	name := extractFlowNameFromProjectMD(data)
	if name == "" {
		return "", fmt.Errorf("no default flow configured. Use --flow flag")
	}

	return name, nil
}

func ResolveDocsPath(root string) string {
	cfg, err := loadProjectConfig(root)
	if err == nil && cfg.Paths.DocsInternal != "" {
		return cfg.Paths.DocsInternal
	}

	projectMD := filepath.Join(root, ".team", "project.md")
	data, mdErr := os.ReadFile(projectMD)
	if mdErr != nil {
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
	cfg, err := loadProjectConfig(root)
	if err == nil && cfg.Paths.DocsExternal != "" {
		return cfg.Paths.DocsExternal
	}

	projectMD := filepath.Join(root, ".team", "project.md")
	data, mdErr := os.ReadFile(projectMD)
	if mdErr != nil {
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

func ResolveInternalDocs(root string) string {
	docsPath := ResolveDocsPath(root)
	if docsPath != "" {
		if !filepath.IsAbs(docsPath) {
			docsPath = filepath.Join(root, docsPath)
		}
		return docsPath
	}
	return filepath.Join(root, ".team", "docs")
}

func loadProjectConfig(root string) (*projectYAML, error) {
	yamlPath := filepath.Join(root, ".team", "project.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}
	var cfg projectYAML
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type projectYAML struct {
	DefaultFlow string `yaml:"default_flow"`
	Paths       struct {
		DocsInternal string `yaml:"docs_internal"`
		DocsExternal string `yaml:"docs_external"`
	} `yaml:"paths"`
}
