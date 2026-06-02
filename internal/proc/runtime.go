package proc

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	skillfs "github.com/origadmin/team-flow"
	"github.com/origadmin/team-flow/internal/config"
	"github.com/origadmin/team-flow/internal/flow"
	"gopkg.in/yaml.v3"
)

func LoadFlow(root, flowID string) (*flow.Flow, error) {
	procPath := ResolveProcPath(root, flowID)
	if procPath != "" {
		f, err := flow.ParseFlowFile(procPath)
		if err != nil {
			return nil, fmt.Errorf("parse error in %s: %w", procPath, err)
		}
		return f, nil
	}

	org := discoverOrg(root)
	if org != "" {
		targetFile := "assets/orgs/" + org + "/flows/" + flowID + ".json"
		data, err := skillfs.FS.ReadFile(targetFile)
		if err == nil {
			f, err := flow.ParseFlow(data)
			if err != nil {
				return nil, fmt.Errorf("parse error in embed flow %s: %w", flowID, err)
			}
			return f, nil
		}
	}

	return nil, fmt.Errorf("flow not found: %s", flowID)
}

func ComputeFlowRevision(root, flowID string) string {
	procPath := ResolveProcPath(root, flowID)
	if procPath == "" {
		return ""
	}
	data, err := os.ReadFile(procPath)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}

func CollectNodeIDs(f *flow.Flow) []string {
	ids := make([]string, len(f.Nodes))
	for i, n := range f.Nodes {
		ids[i] = n.ID
	}
	return ids
}

func FindNode(f *flow.Flow, nodeID string) (*flow.FlowNode, error) {
	if len(f.Nodes) == 0 {
		return nil, fmt.Errorf("flow has no nodes")
	}

	if nodeID == "" {
		return FindRootNode(f)
	}

	// Try exact ID match first
	for i := range f.Nodes {
		if f.Nodes[i].ID == nodeID {
			return &f.Nodes[i], nil
		}
	}

	// Fallback: try name match (case-insensitive)
	for i := range f.Nodes {
		if strings.EqualFold(f.Nodes[i].Name, nodeID) {
			return &f.Nodes[i], nil
		}
	}

	available := make([]string, 0, len(f.Nodes))
	for _, n := range f.Nodes {
		available = append(available, n.ID+"("+n.Name+")")
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

func ResolveActiveFlow(root string) (string, error) {
	cfg, err := loadProjectConfig(root)
	if err == nil {
		if cfg.ActiveFlow != "" {
			return cfg.ActiveFlow, nil
		}
		if cfg.DefaultFlow != "" {
			return cfg.DefaultFlow, nil
		}
	}

	return "", fmt.Errorf("no active flow configured. Use --flow flag")
}

func ResolveDocsPath(root string) string {
	cfg, err := loadProjectConfig(root)
	if err == nil && cfg.Paths.DocsInternal != "" {
		return cfg.Paths.DocsInternal
	}
	return ""
}

func ResolveDocsExternalPath(root string) string {
	cfg, err := loadProjectConfig(root)
	if err == nil && cfg.Paths.DocsExternal != "" {
		return cfg.Paths.DocsExternal
	}
	return ""
}

func ResolveInternalDocs(root string) string {
	docsPath := ResolveDocsPath(root)
	if docsPath != "" {
		workspace := config.FindWorkspaceRoot(root)
		if workspace == "" {
			workspace = root
		}
		return config.ResolvePathWithAnchor(root, workspace, docsPath)
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
	ActiveFlow  string `yaml:"active_flow"`
	DefaultFlow string `yaml:"default_flow"`
	Paths       struct {
		DocsInternal string `yaml:"docs_internal"`
		DocsExternal string `yaml:"docs_external"`
	} `yaml:"paths"`
}

func LoadTeam(root string) (*flow.TeamDefinition, error) {
	teamPath := filepath.Join(root, ".team", "team.json")
	data, err := os.ReadFile(teamPath)
	if err != nil {
		return nil, nil
	}

	var team flow.TeamDefinition
	if json.Unmarshal(data, &team) != nil || team.ID == "" {
		return nil, nil
	}

	if team.Org != "" && len(team.Roles) == 0 {
		if err := loadTemplateRoles(&team); err != nil {
			return nil, err
		}
	}

	return &team, nil
}

func loadTemplateRoles(team *flow.TeamDefinition) error {
	path := "assets/orgs/" + team.Org + "/team.json"
	data, err := skillfs.FS.ReadFile(path)
	if err != nil {
		return fmt.Errorf("org template not found: %s: %w", team.Org, err)
	}
	var template flow.TeamDefinition
	if err := json.Unmarshal(data, &template); err != nil {
		return fmt.Errorf("invalid org template: %w", err)
	}
	team.Roles = template.Roles
	team.Rules = template.Rules
	if team.Name == "" {
		team.Name = template.Name
	}
	if team.DefaultFlow == "" {
		team.DefaultFlow = template.DefaultFlow
	}
	team.Flows = template.Flows
	team.SkillTags = template.SkillTags
	team.SkillRequirements = template.SkillRequirements
	if team.Toolchain == nil {
		team.Toolchain = template.Toolchain
	}
	if team.GateCheckers == nil {
		team.GateCheckers = template.GateCheckers
	}
	return nil
}

func discoverOrg(root string) string {
	data, err := os.ReadFile(filepath.Join(root, ".team", "team.json"))
	if err != nil {
		return ""
	}
	var team flow.TeamDefinition
	if json.Unmarshal(data, &team) != nil {
		return ""
	}
	return team.Org
}

func loadRole(roleID string, org string, root string) (*flow.RoleDefinition, error) {
	localPath := filepath.Join(root, ".team", "roles", roleID+".json")
	if data, err := os.ReadFile(localPath); err == nil {
		var role flow.RoleDefinition
		if json.Unmarshal(data, &role) == nil && role.ID == roleID {
			return &role, nil
		}
	}

	if org != "" {
		embedPath := "assets/orgs/" + org + "/roles/" + roleID + ".json"
		if data, err := skillfs.FS.ReadFile(embedPath); err == nil {
			var role flow.RoleDefinition
			if json.Unmarshal(data, &role) == nil && role.ID == roleID {
				return &role, nil
			}
		}
	}

	return nil, fmt.Errorf("role not found: %s", roleID)
}

func loadRule(ruleID string, org string, root string) (*flow.RuleDefinition, error) {
	localPath := filepath.Join(root, ".team", "rules", ruleID+".json")
	if data, err := os.ReadFile(localPath); err == nil {
		var rule flow.RuleDefinition
		if json.Unmarshal(data, &rule) == nil && rule.ID == ruleID {
			return &rule, nil
		}
	}

	if org != "" {
		embedPath := "assets/orgs/" + org + "/rules/" + ruleID + ".json"
		if data, err := skillfs.FS.ReadFile(embedPath); err == nil {
			var rule flow.RuleDefinition
			if json.Unmarshal(data, &rule) == nil && rule.ID == ruleID {
				return &rule, nil
			}
		}
	}

	return nil, fmt.Errorf("rule not found: %s", ruleID)
}

func LoadTeamFromFlowPath(flowPath string) (*flow.TeamDefinition, error) {
	dir := filepath.Dir(flowPath)
	parentDir := filepath.Dir(dir)

	for _, candidate := range []string{
		filepath.Join(dir, "team.json"),
		filepath.Join(parentDir, "team.json"),
	} {
		data, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}
		var team flow.TeamDefinition
		if json.Unmarshal(data, &team) == nil && team.ID != "" {
			return &team, nil
		}
	}

	return nil, nil
}

func ResolveDefaultFlowName(root string) (string, error) {
	return ResolveActiveFlow(root)
}
