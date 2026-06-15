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
	"github.com/origadmin/team-flow/internal/templates"
	"gopkg.in/yaml.v3"
)

func init() {
	// Register the templates package as the source for embed://templates/<id>.json
	// paths. This lets the flow parser transparently serve inherited sub_flow
	// templates without any team having to provide a JSON file.
	flow.RegisterTemplateLoader(func(id string) ([]byte, error) {
		return templates.Get(id)
	})
}

func LoadFlow(root, flowID string) (*flow.Flow, error) {
	// 优先读取 team.json 中的 flows[*].file 配置
	if team, _ := LoadTeam(root); team != nil {
		for _, f := range team.Flows {
			if f.ID == flowID && f.File != "" {
				// team.json 中 file 可以是绝对路径或相对路径（相对于 root）
				procPath := f.File
				if !filepath.IsAbs(procPath) {
					procPath = filepath.Join(root, f.File)
				}
				if _, err := os.Stat(procPath); err == nil {
					if flow, err := flow.ParseFlowFile(procPath); err == nil {
						return flow, nil
					}
				}
				// 如果相对磁盘路径不存在，尝试从嵌入式 FS 读取
				if data, err := skillfs.FS.ReadFile(f.File); err == nil {
					if flow, err := flow.ParseFlow(data); err == nil {
						return flow, nil
					}
				}
			}
		}
	}

	// 回退到路径解析
	procPath := ResolveProcPath(root, flowID)
	if procPath != "" {
		f, err := flow.ParseFlowFile(procPath)
		if err != nil {
			return nil, fmt.Errorf("parse error in %s: %w", procPath, err)
		}
		return f, nil
	}

	org := discoverOrg(root)
	if org == "" {
		// Zero config: derive org from flow name convention.
		org = deriveOrg(flowID)
	}

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

// findPredecessor returns the first predecessor node that has a role assignment.
// When multiple edges point to a terminal node (e.g. sta0→suc0 conditional and
// rev6→suc0 unconditional), we prefer the predecessor that actually has role info
// for the status line display.
func findPredecessor(f *flow.Flow, node *flow.FlowNode) *flow.FlowNode {
	if f == nil || node == nil {
		return nil
	}
	var fallback *flow.FlowNode
	for _, e := range f.Edges {
		if e.To == node.ID {
			for i := range f.Nodes {
				if f.Nodes[i].ID == e.From {
					// Prefer predecessor with role assignment
					if GetNodeRole(&f.Nodes[i]) != "" {
						return &f.Nodes[i]
					}
					if fallback == nil {
						fallback = &f.Nodes[i]
					}
				}
			}
		}
	}
	return fallback
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

// defaultFlow is the absolute tool-level default when no project configuration
// provides an active_flow. Zero-config projects run dev-flow.
const defaultFlow = "dev-flow"

// deriveOrg derives the org name from a flow name by convention:
//
//	"dev-flow"       → "dev-team"
//	"skill-dev-flow" → "skill-team"
//	"unknown"        → "dev-team" (safe fallback)
//
// An org is a template namespace for looking up embedded roles, rules, and flows.
// It is derived, not configured — users set active_flow, and the org follows.
func deriveOrg(flowName string) string {
	if after, ok := strings.CutSuffix(flowName, "-flow"); ok {
		return after + "-team"
	}
	return "dev-team"
}

// ResolveActiveFlow returns the flow that the project should run.
// Lookup order (high → low priority):
//  1. project.yaml active_flow  (user explicit, project-level)
//  2. defaultFlow               (tool-level absolute default)
func ResolveActiveFlow(root string) string {
	cfg, err := loadProjectConfig(root)
	if err == nil && cfg.ActiveFlow != "" {
		return cfg.ActiveFlow
	}
	return defaultFlow
}

func ResolveDocsPath(root string) string {
	cfg, err := loadProjectConfig(root)
	if err == nil && cfg.Paths.DocsInternal != "" {
		workspace := config.FindWorkspaceRoot(root)
		if workspace == "" {
			workspace = root
		}
		return config.ResolvePathWithAnchor(root, workspace, cfg.Paths.DocsInternal)
	}
	return ""
}

func ResolveDocsExternalPath(root string) string {
	cfg, err := loadProjectConfig(root)
	if err == nil && cfg.Paths.DocsExternal != "" {
		workspace := config.FindWorkspaceRoot(root)
		if workspace == "" {
			workspace = root
		}
		return config.ResolvePathWithAnchor(root, workspace, cfg.Paths.DocsExternal)
	}
	return ""
}

func ResolveInternalDocs(root string) string {
	docsPath := ResolveDocsPath(root)
	if docsPath != "" {
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
	ActiveFlow string `yaml:"active_flow"`
	Team       string `yaml:"team"`
	Paths      struct {
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
	if len(team.Roles) == 0 {
		// No explicit roles: use team.ID as org to load embedded template roles.
		// This is the zero/minimal-config path — team.json has only {"id":"skill-team"}
		// and all roles/rules/flows are loaded from the embedded org template.
		if team.Org == "" {
			team.Org = team.ID
		}
		if team.Org != "" {
			if err := loadTemplateRoles(&team); err != nil {
				return nil, err
			}
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
	if team.ActiveFlow == "" {
		team.ActiveFlow = template.ActiveFlow
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

// loadTeamFromEmbed 从嵌入的 org 模板加载团队定义
func loadTeamFromEmbed(org string) (*flow.TeamDefinition, error) {
	path := "assets/orgs/" + org + "/team.json"
	data, err := skillfs.FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("org template not found: %s: %w", org, err)
	}
	var team flow.TeamDefinition
	if err := json.Unmarshal(data, &team); err != nil {
		return nil, fmt.Errorf("invalid org template: %w", err)
	}
	return &team, nil
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
	if team.Org != "" {
		return team.Org
	}
	// Fallback: team.json only sets "id" (e.g. dev-team), which also
	// identifies the org for locating embedded flow templates.
	return team.ID
}

func loadRole(roleID string, org string, root string) (*flow.RoleDefinition, error) {
	// 1. Local .team/roles/<id>.json
	localPath := filepath.Join(root, ".team", "roles", roleID+".json")
	if data, err := os.ReadFile(localPath); err == nil {
		var role flow.RoleDefinition
		if json.Unmarshal(data, &role) == nil && role.ID == roleID {
			return &role, nil
		}
	}

	// 2. Org-specific embed: assets/orgs/<org>/roles/<id>.json
	if org != "" {
		embedPath := "assets/orgs/" + org + "/roles/" + roleID + ".json"
		if data, err := skillfs.FS.ReadFile(embedPath); err == nil {
			var role flow.RoleDefinition
			if json.Unmarshal(data, &role) == nil && role.ID == roleID {
				return &role, nil
			}
		}
	}

	// 3. Global dev-team fallback: assets/orgs/dev-team/roles/<id>.json
	//    (same strategy as LoadFlow — when a team uses dev-team flows,
	//     roles defined in dev-team org should be resolvable)
	if org != "dev-team" {
		fallbackPath := "assets/orgs/dev-team/roles/" + roleID + ".json"
		if data, err := skillfs.FS.ReadFile(fallbackPath); err == nil {
			var role flow.RoleDefinition
			if json.Unmarshal(data, &role) == nil && role.ID == roleID {
				return &role, nil
			}
		}
	}

	return nil, fmt.Errorf("role not found: %s", roleID)
}

// loadRoleFromPromptSource 从 prompts/*.md 文件解析角色元数据。
// v3 架构: team.json 是装配清单，角色内容存在 prompts/*.md 的 YAML frontmatter 中。
// 本函数解析 frontmatter 提取: persona, traits, guidance, alias, alias_en, name,
// prompt_directives, rules (规则 ID 列表), capabilities。
func loadRoleFromPromptSource(roleID, promptSource, root string) (*flow.RoleDefinition, error) {
	if promptSource == "" {
		return nil, nil
	}

	// 尝试多个路径: 1) project-local (.team/{prompt_source})
	//               2) embedded (assets/skill/v3/prompts/{name}.md)
	//               3) absolute / relative path from project root
	var data []byte
	var readErr error

	// 路径 1: embedded assets (最常见, 如 "assets/skill/v3/prompts/dev.md")
	data, readErr = skillfs.FS.ReadFile(promptSource)

	// 路径 2: project-local (如 ".team/prompts/dev.md")
	if readErr != nil {
		localPath := filepath.Join(root, promptSource)
		data, readErr = os.ReadFile(localPath)
	}

	// 路径 3: 作为文件名, 在 assets/skill/v3/prompts/ 下查找
	if readErr != nil {
		embedName := "assets/skill/v3/prompts/" + promptSource
		if !strings.HasSuffix(embedName, ".md") {
			embedName += ".md"
		}
		data, readErr = skillfs.FS.ReadFile(embedName)
	}

	if readErr != nil {
		return nil, fmt.Errorf("prompt_source not found: %s", promptSource)
	}

	role, err := parseRoleFrontmatter(data, roleID)
	if err != nil {
		return nil, fmt.Errorf("parse prompt_source %s: %w", promptSource, err)
	}
	return role, nil
}

// parseRoleFrontmatter 从 Markdown 文件的 YAML frontmatter 中提取角色元数据。
// 返回部分填充的 RoleDefinition（id, persona, traits, guidance, alias, alias_en,
// name, prompt_directives, capabilities, rules）。
func parseRoleFrontmatter(data []byte, roleID string) (*flow.RoleDefinition, error) {
	content := string(data)
	frontmatter, err := extractFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("extract frontmatter: %w", err)
	}
	if frontmatter == "" {
		return nil, nil
	}

	// 解析 YAML frontmatter
	var fm struct {
		AI struct {
			ID       string   `yaml:"id"`
			Aliases  []string `yaml:"aliases"`
			Persona  string   `yaml:"persona"`
			Alias    string   `yaml:"alias"`
			AliasEn  string   `yaml:"alias_en"`
			Name     string   `yaml:"name"`
			Traits   []string `yaml:"traits"`
			Guidance string   `yaml:"guidance"`
			Triggers struct {
				Keywords []string `yaml:"keywords"`
				TaskTypes []string `yaml:"task_types"`
			} `yaml:"triggers"`
			PromptDirectives []string `yaml:"prompt_directives"`
			Capabilities     []string `yaml:"capabilities"`
			Rules            []string `yaml:"rules"`
			Standards        []string `yaml:"standards"`
		} `yaml:"ai"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	id := roleID
	if fm.AI.ID != "" {
		id = fm.AI.ID
	}

	role := &flow.RoleDefinition{
		ID:               id,
		Name:             fm.AI.Name,
		Alias:            fm.AI.Alias,
		AliasEn:          fm.AI.AliasEn,
		Persona:          fm.AI.Persona,
		Traits:           fm.AI.Traits,
		Guidance:         fm.AI.Guidance,
		PromptDirectives: fm.AI.PromptDirectives,
		Capabilities:     fm.AI.Capabilities,
		Rules:            fm.AI.Rules,
		PromptSource:     "", // caller sets this
	}
	return role, nil
}

// extractFrontmatter 从 Markdown 中提取 YAML frontmatter 内容（不包含 --- 分隔符）。
func extractFrontmatter(content string) (string, error) {
	const sep = "---"
	// 跳过开头空白
	start := 0
	for start < len(content) && (content[start] == '\r' || content[start] == '\n' || content[start] == '\t' || content[start] == ' ') {
		start++
	}
	if !strings.HasPrefix(content[start:], sep) {
		return "", nil
	}
	afterOpen := start + len(sep)
	// 必须紧跟换行
	if afterOpen >= len(content) || (content[afterOpen] != '\n' && content[afterOpen] != '\r') {
		return "", nil
	}
	// 跳过换行符
	idx := afterOpen
	for idx < len(content) && (content[idx] == '\r' || content[idx] == '\n') {
		idx++
	}
	// 查找结束分隔符
	closePos := strings.Index(content[idx:], sep)
	if closePos < 0 {
		return "", fmt.Errorf("unterminated frontmatter: missing closing ---")
	}
	body := content[idx : idx+closePos]
	return strings.TrimSpace(body), nil
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

// ResolveActiveFlowName returns the active flow name for a project root.
func ResolveActiveFlowName(root string) string {
	return ResolveActiveFlow(root)
}
