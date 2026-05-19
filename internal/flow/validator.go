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

		if edge.To == "" {
			result.Errors = append(result.Errors, ValidationIssue{
				Severity: SeverityError,
				EdgeID:   edge.ID,
				Message:  "edge to is required",
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
