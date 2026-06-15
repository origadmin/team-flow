package flow

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
)

type ValidationIssue struct {
	Severity  ValidationSeverity `json:"severity"`
	NodeID    string             `json:"node_id,omitempty"`
	EdgeID    string             `json:"edge_id,omitempty"`
	Message   string             `json:"message"`
	Field     string             `json:"field,omitempty"`
}

type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationIssue `json:"errors,omitempty"`
	Warnings []ValidationIssue `json:"warnings,omitempty"`
}

func ValidateFlow(flow *Flow) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if flow == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "flow is nil",
		})
		return result
	}

	// Add cycle detection
	validateCycles(flow, result)

	validateSyntax(flow, result)
	validateSemantics(flow, result)

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

func ValidateTeam(team *TeamDefinition) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if team == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "team is nil",
		})
		return result
	}

	// v3 schema validation: team.json must use reference-only format
	// (装配清单格式), not inline content. 内联角色/规则内容是旧格式(v2/v1),
	// 在 v3 中是 ERROR, 不是 warning. 角色内容存在 prompts/*.md 中，通过 prompt_source 引用。
	validateV3Schema(team, result)

	if team.ID == "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "team id is required",
			Field:    "id",
		})
	}

	if team.Name == "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "team name is required",
			Field:    "name",
		})
	}

	validateTeamRoles(team, result)
	validateTeamRules(team, result)

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

// validateV3Schema 检测 team.json 是否使用旧格式（角色/规则内联内容）。
// v3 规范：team.json 是装配清单，只管理引用，不管理内容。
// 角色内容必须在 assets/skill/v3/prompts/{role-id}.md 中。
// 规则内容必须在 prompts/rules.md 或 prompts/{rule-id}.md 中。
func validateV3Schema(team *TeamDefinition, result *ValidationResult) {
	// 1. schema_version 必须存在且为 "3.0"
	//    (SchemaVersion 是可选字段 — 旧格式通常没有)
	//    如果有 schema_version 且不是 "3.0" → ERROR
	//    如果没有 schema_version → 检查是否有旧格式内联字段

	hasExplicitSchemaV3 := false
	if team.SchemaVersion != "" {
		if team.SchemaVersion != "3.0" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message: fmt.Sprintf("team.json schema_version %q is invalid; expected \"3.0\"",
					team.SchemaVersion),
				Field: "schema_version",
			})
		} else {
			hasExplicitSchemaV3 = true
		}
	}

	// 2. 检测角色内联字段 (v2/v1 的标志)
	//    以下字段如果在 team.json roles[] 中存在就是旧格式:
	//    name, alias, alias_en, persona, traits, guidance, capabilities
	for i, role := range team.Roles {
		// 收集旧格式字段
		legacyFields := []string{}
		if role.Name != "" {
			legacyFields = append(legacyFields, "name")
		}
		if role.Alias != "" {
			legacyFields = append(legacyFields, "alias")
		}
		if role.AliasEn != "" {
			legacyFields = append(legacyFields, "alias_en")
		}
		if role.Persona != "" {
			legacyFields = append(legacyFields, "persona")
		}
		if len(role.Traits) > 0 {
			legacyFields = append(legacyFields, "traits")
		}
		if role.Guidance != "" {
			legacyFields = append(legacyFields, "guidance")
		}
		if len(role.Capabilities) > 0 {
			legacyFields = append(legacyFields, "capabilities")
		}

		if len(legacyFields) > 0 {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message: fmt.Sprintf(
					"team.json roles[%d] (%s) uses v2/v1 legacy inline format: "+
						"fields [%s] must be in assets/skill/v3/prompts/%s.md "+
						"(team.json is an assembly manifest, not a role content container). "+
						"See docs/ARCHITECTURE.md §6.3",
					i, role.ID, strings.Join(legacyFields, ", "), role.ID),
				Field: "roles",
			})
		}

		// 新格式强制要求 prompt_source
		if hasExplicitSchemaV3 && role.PromptSource == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message: fmt.Sprintf(
					"team.json roles[%d] (%s) is v3 but has no prompt_source field. "+
						"Must reference assets/skill/v3/prompts/%s.md or equivalent path",
					i, role.ID, role.ID),
				Field: "roles[" + role.ID + "].prompt_source",
			})
		}
	}

	// 3. 检测规则内联字段 (v2/v1 的标志)
	for i, rule := range team.Rules {
		legacyFields := []string{}
		if rule.Name != "" {
			legacyFields = append(legacyFields, "name")
		}
		if rule.Instruction != "" {
			legacyFields = append(legacyFields, "instruction")
		}
		if rule.Description != "" {
			legacyFields = append(legacyFields, "description")
		}
		if rule.Enforcement != "" {
			legacyFields = append(legacyFields, "enforcement")
		}

		if len(legacyFields) > 0 {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message: fmt.Sprintf(
					"team.json rules[%d] (%s) uses v2/v1 legacy inline format: "+
						"fields [%s] must be in assets/skill/v3/prompts/rules.md "+
						"(team.json is an assembly manifest, not a rule content container). "+
						"See docs/ARCHITECTURE.md §6.3",
					i, rule.ID, strings.Join(legacyFields, ", ")),
				Field: "rules",
			})
		}

		// 新格式强制要求 source 字段
		if hasExplicitSchemaV3 && rule.Source == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message: fmt.Sprintf(
					"team.json rules[%d] (%s) is v3 but has no source field. "+
						"Must reference assets/skill/v3/prompts/rules.md or equivalent path",
					i, rule.ID),
				Field: "rules[" + rule.ID + "].source",
			})
		}
	}

	// 4. 无 schema_version 但无任何内联内容 → 提示添加 schema_version:"3.0"
	if !hasExplicitSchemaV3 && len(result.Errors) == 0 && (len(team.Roles) > 0 || len(team.Rules) > 0) {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity: SeverityWarning,
			Message: "team.json missing schema_version; add \"schema_version\": \"3.0\" to declare v3 reference-only format compliance",
			Field:   "schema_version",
		})
	}
}

func ValidateFlowWithTeam(fl *Flow, team *TeamDefinition) *ValidationResult {
	result := ValidateFlow(fl)

	if team == nil || fl == nil {
		return result
	}

	teamRoles := make(map[string]bool)
	for _, r := range team.Roles {
		teamRoles[r.ID] = true
	}

	teamRules := make(map[string]bool)
	for _, r := range team.Rules {
		teamRules[r.ID] = true
	}
	for _, role := range team.Roles {
		for _, ruleRef := range role.Rules {
			teamRules[ruleRef] = true
		}
	}

	for _, node := range fl.Nodes {
		if node.Components == nil {
			continue
		}
		for _, roleRef := range node.Components.Roles {
			if roleRef.Source == SourceTeam && !teamRoles[roleRef.Ref] {
				result.Warnings = append(result.Warnings, ValidationIssue{
					Severity: SeverityWarning,
					NodeID:   node.ID,
					Message:  fmt.Sprintf("node %s references team role not found in team.json: %s", node.ID, roleRef.Ref),
					Field:    "nodes[].components.roles",
				})
			}
		}
	}

	var filtered []ValidationIssue
	for _, w := range result.Warnings {
		if isTeamDefinedWarning(w, teamRoles, teamRules) {
			continue
		}
		filtered = append(filtered, w)
	}
	result.Warnings = filtered

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

func isTeamDefinedWarning(w ValidationIssue, teamRoles, teamRules map[string]bool) bool {
	if w.Field != "nodes[].components.rules" && w.Field != "nodes[].components.roles" {
		return false
	}
	if !strings.Contains(w.Message, "references undefined") {
		return false
	}
	for roleID := range teamRoles {
		if strings.HasSuffix(w.Message, ": "+roleID) {
			return true
		}
	}
	for ruleID := range teamRules {
		if strings.HasSuffix(w.Message, ": "+ruleID) {
			return true
		}
	}
	return false
}

func validateTeamRoles(team *TeamDefinition, result *ValidationResult) {
	if len(team.Roles) == 0 {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity: SeverityWarning,
			Message:  "team has no roles defined",
			Field:    "roles",
		})
		return
	}

	roleIDs := make(map[string]bool)
	principalCount := 0
	for _, role := range team.Roles {
		if role.ID == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  "role definition missing id",
				Field:    "roles",
			})
			continue
		}

		if roleIDs[role.ID] {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("duplicate role id: %s", role.ID),
				Field:    "roles",
			})
		}
		roleIDs[role.ID] = true

		if role.Principal != nil && *role.Principal {
			principalCount++
		}

		if len(role.Persona) < 50 {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("role %s persona too short (minimum 50 chars, got %d)", role.ID, len(role.Persona)),
				Field:    "roles",
			})
		}

		if len(role.Traits) < 2 {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("role %s needs at least 2 traits (got %d)", role.ID, len(role.Traits)),
				Field:    "roles",
			})
		}

		if len(role.Guidance) < 30 {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("role %s guidance too short (minimum 30 chars, got %d)", role.ID, len(role.Guidance)),
				Field:    "roles",
			})
		}

		if len(role.Capabilities) < 1 {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("role %s needs at least 1 capability", role.ID),
				Field:    "roles",
			})
		}
	}

	if principalCount == 0 {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Severity: SeverityWarning,
			Message:  "team has no principal role defined",
			Field:    "roles",
		})
	}

	if principalCount > 1 {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("team has %d principal roles (maximum 1)", principalCount),
			Field:    "roles",
		})
	}
}

func validateTeamRules(team *TeamDefinition, result *ValidationResult) {
	ruleIDs := make(map[string]bool)
	for _, rule := range team.Rules {
		if rule.ID == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  "rule definition missing id",
				Field:    "rules",
			})
			continue
		}

		if ruleIDs[rule.ID] {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("duplicate rule id: %s", rule.ID),
				Field:    "rules",
			})
		}
		ruleIDs[rule.ID] = true

		if rule.Name == "" {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("rule %s missing name", rule.ID),
				Field:    "rules",
			})
		}

		if rule.Instruction == "" {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("rule %s has no instruction — AI cannot follow a rule without content", rule.ID),
				Field:    "rules",
			})
		}
	}

	for _, role := range team.Roles {
		for _, ruleRef := range role.Rules {
			if !ruleIDs[ruleRef] {
				result.Warnings = append(result.Warnings, ValidationIssue{
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("role %s references undefined rule: %s", role.ID, ruleRef),
					Field:    "roles[].rules",
				})
			}
		}
	}
}

func validateSyntax(flow *Flow, result *ValidationResult) {
	if flow.Version == "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "version is required",
			Field:    "version",
		})
	}

	if flow.Metadata.Name == "" {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "metadata.name is required",
			Field:    "metadata.name",
		})
	}

	if len(flow.Nodes) == 0 {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "nodes must have at least one element",
			Field:    "nodes",
		})
	}

	nodeIDs := make(map[string]bool, len(flow.Nodes))
	for _, node := range flow.Nodes {
		if node.ID == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				NodeID:   node.ID,
				Message:  "node id is required",
				Field:    "nodes[].id",
			})
			continue
		}

		if nodeIDs[node.ID] {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				NodeID:   node.ID,
				Message:  fmt.Sprintf("duplicate node id: %s", node.ID),
				Field:    "nodes[].id",
			})
		}
		nodeIDs[node.ID] = true

		if !isValidNodeType(node.Type) {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				NodeID:   node.ID,
				Message:  fmt.Sprintf("invalid node type: %s", node.Type),
				Field:    "nodes[].type",
			})
		}

		if node.Name == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				NodeID:   node.ID,
				Message:  "node name is required",
				Field:    "nodes[].name",
			})
		}
	}

	edgeIDs := make(map[string]bool)
	for _, edge := range flow.Edges {
		if edge.ID != "" {
			if edgeIDs[edge.ID] {
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					EdgeID:   edge.ID,
					Message:  fmt.Sprintf("duplicate edge id: %s", edge.ID),
					Field:    "edges[].id",
				})
			}
			edgeIDs[edge.ID] = true
		}

		if edge.From == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				EdgeID:   edge.ID,
				Message:  "edge from is required",
				Field:    "edges[].from",
			})
		}

		if edge.To == "" && edge.SubflowRef == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				EdgeID:   edge.ID,
				Message:  "edge to or subflow_ref is required",
				Field:    "edges[].to",
			})
		}
	}
}

func validateSemantics(flow *Flow, result *ValidationResult) {
	nodeIDs := make(map[string]bool, len(flow.Nodes))
	for _, node := range flow.Nodes {
		nodeIDs[node.ID] = true
	}

	hasTerminal := false
	for _, node := range flow.Nodes {
		if node.Type == NodeTypeTerminal {
			hasTerminal = true
		}
	}
	if !hasTerminal {
		result.Errors = append(result.Errors, ValidationIssue{
			Severity: SeverityError,
			Message:  "flow must have at least one terminal node",
			Field:    "nodes",
		})
	}

	for _, node := range flow.Nodes {
		if node.Type == NodeTypePhase {
			for _, doc := range node.Docs {
				if doc.Required != nil && *doc.Required {
					if doc.Name == "" || doc.Path == "" {
						result.Errors = append(result.Errors, ValidationIssue{
							Severity: SeverityError,
							NodeID:   node.ID,
							Message:  fmt.Sprintf("phase node %s has required doc with missing name or path", node.ID),
							Field:    "nodes[].docs",
						})
					}
				}
			}
		}
	}

	for _, edge := range flow.Edges {
		if edge.From != "" && !nodeIDs[edge.From] {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				EdgeID:   edge.ID,
				Message:  fmt.Sprintf("edge from references non-existent node: %s", edge.From),
				Field:    "edges[].from",
			})
		}
		// Subflow edges use subflow_ref instead of to; skip to validation
		if edge.SubflowRef != "" {
			continue
		}
		if edge.To != "" && !nodeIDs[edge.To] {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				EdgeID:   edge.ID,
				Message:  fmt.Sprintf("edge to references non-existent node: %s", edge.To),
				Field:    "edges[].to",
			})
		}
	}

	connectedNodes := make(map[string]bool)
	for _, edge := range flow.Edges {
		connectedNodes[edge.From] = true
		connectedNodes[edge.To] = true
	}
	for _, node := range flow.Nodes {
		if node.Type != NodeTypeTerminal && !connectedNodes[node.ID] {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				NodeID:   node.ID,
				Message:  fmt.Sprintf("orphan node detected: %s (not connected by any edge)", node.ID),
				Field:    "nodes",
			})
		}
	}

	for _, node := range flow.Nodes {
		if node.Type == NodeTypeSubflow {
			var cfg SubflowConfig
			if err := unmarshalNodeConfig(node.Config, &cfg); err == nil {
				if cfg.FlowRef != "" && !isValidFlowRef(cfg.FlowRef) {
					result.Errors = append(result.Errors, ValidationIssue{
						Severity: SeverityError,
						NodeID:   node.ID,
						Message:  fmt.Sprintf("subflow flow_ref must match pattern namespace:name: %s", cfg.FlowRef),
						Field:    "nodes[].config.flow_ref",
					})
				}
			}
		}
	}

	validateComponents(flow, result)
}

func isValidNodeType(t NodeType) bool {
	switch t {
	case NodeTypePhase, NodeTypeStart, NodeTypeGate, NodeTypeBranch, NodeTypeParallel,
		NodeTypeSubflow, NodeTypeLoop, NodeTypeManual, NodeTypeEvent, NodeTypeTerminal:
		return true
	}
	return false
}

func isValidFlowRef(ref string) bool {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	return true
}

func unmarshalNodeConfig(raw json.RawMessage, v interface{}) error {
	if len(raw) == 0 {
		return fmt.Errorf("empty config")
	}
	return json.Unmarshal(raw, v)
}

func validateComponents(flow *Flow, result *ValidationResult) {
	if flow.Components == nil {
		return
	}

	definedRules := make(map[string]bool)
	for _, rule := range flow.Components.Rules {
		if rule.ID == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  "rule definition missing id",
				Field:    "components.rules",
			})
			continue
		}
		definedRules[rule.ID] = true
		if rule.Name == "" {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("rule definition %s missing name", rule.ID),
				Field:    "components.rules",
			})
		}
		if rule.Description == "" {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("rule %s has no description — AI cannot follow a rule without content", rule.ID),
				Field:    "components.rules",
			})
		}
	}

	definedRoles := make(map[string]bool)
	for _, role := range flow.Components.Roles {
		if role.ID == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				Message:  "role definition missing id",
				Field:    "components.roles",
			})
			continue
		}
		definedRoles[role.ID] = true
		if role.Name == "" {
			result.Warnings = append(result.Warnings, ValidationIssue{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("role definition %s missing name", role.ID),
				Field:    "components.roles",
			})
		}
	}

	for _, node := range flow.Nodes {
		if node.Components == nil {
			continue
		}
		for _, ruleRef := range node.Components.Rules {
			if ruleRef.Source == SourceTeam {
				continue
			}
			if !definedRules[ruleRef.Ref] {
				result.Warnings = append(result.Warnings, ValidationIssue{
					Severity: SeverityWarning,
					NodeID:   node.ID,
					Message:  fmt.Sprintf("node %s references undefined rule: %s", node.ID, ruleRef.Ref),
					Field:    "nodes[].components.rules",
				})
			}
		}
		for _, roleRef := range node.Components.Roles {
			if roleRef.Source == SourceTeam {
				continue
			}
			if !definedRoles[roleRef.Ref] {
				result.Warnings = append(result.Warnings, ValidationIssue{
					Severity: SeverityWarning,
					NodeID:   node.ID,
					Message:  fmt.Sprintf("node %s references undefined role: %s", node.ID, roleRef.Ref),
					Field:    "nodes[].components.roles",
				})
			}
		}
	}
}

// validateCycles detects cycles in the flow graph using DFS with 3-color marking.
// Subflow edges are excluded since they point to an external flow, not a loop within this flow.
func validateCycles(flow *Flow, result *ValidationResult) {
	if len(flow.Edges) == 0 {
		return
	}

	// Build adjacency list (only sequential edges — conditional and subflow edges are excluded)
	adj := make(map[string][]string, len(flow.Nodes))
	for _, edge := range flow.Edges {
		if edge.Type == EdgeTypeSubflow {
			continue
		}
		// Conditional edges represent intentional branches, not guaranteed cycles
		if edge.Type == EdgeTypeConditional {
			continue
		}
		adj[edge.From] = append(adj[edge.From], edge.To)
	}

	// 0 = white (unvisited), 1 = gray (in current path), 2 = black (fully processed)
	color := make(map[string]int, len(flow.Nodes))
	var path []string
	var foundCycle bool

	var dfs func(nodeID string)
	dfs = func(nodeID string) {
		if foundCycle {
			return
		}
		color[nodeID] = 1
		path = append(path, nodeID)

		for _, next := range adj[nodeID] {
			c := color[next]
			if c == 1 {
				// Back edge found — build cycle description
				var cycleNodes []string
				inCycle := false
				for _, n := range path {
					if n == next {
						inCycle = true
					}
					if inCycle {
						cycleNodes = append(cycleNodes, n)
					}
				}
				cycleNodes = append(cycleNodes, next)
				result.Errors = append(result.Errors, ValidationIssue{
					Severity: SeverityError,
					Message:  fmt.Sprintf("cycle detected: %s", strings.Join(cycleNodes, " \u2192 ")),
					Field:    "edges",
				})
				foundCycle = true
				return
			}
			if c == 0 {
				dfs(next)
				if foundCycle {
					return
				}
			}
		}

		color[nodeID] = 2
		path = path[:len(path)-1]
	}

	for _, node := range flow.Nodes {
		if color[node.ID] == 0 {
			dfs(node.ID)
			if foundCycle {
				break
			}
		}
	}
}
