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
