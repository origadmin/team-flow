package proc

import (
	"encoding/json"

	"github.com/origadmin/team-flow/internal/flow"
)

func ExtractGateConditions(node *flow.FlowNode) []GateCondOutput {
	// v3 format: conditions at node level (FlatNode)
	if len(node.Conditions) > 0 {
		conditions := make([]GateCondOutput, 0, len(node.Conditions))
		for _, c := range node.Conditions {
			required := true
			if c.Required != nil {
				required = *c.Required
			}
			conditions = append(conditions, GateCondOutput{
				Type:         string(c.Type),
				Threshold:    c.Threshold,
				Required:     required,
				Check:        c.Check,
				Expected:     c.Expected,
				Deliverables: c.Deliverables,
				NodeID:       node.ID,
				NodeName:     node.Name,
			})
		}
		return conditions
	}

	// v2 format: conditions inside config (GateNodeConfig)
	if node.Config == nil {
		return nil
	}
	var gateCfg flow.GateNodeConfig
	if err := json.Unmarshal(node.Config, &gateCfg); err != nil {
		return nil
	}
	if len(gateCfg.Conditions) == 0 {
		return nil
	}
	conditions := make([]GateCondOutput, 0, len(gateCfg.Conditions))
	for _, c := range gateCfg.Conditions {
		required := true
		if c.Required != nil {
			required = *c.Required
		}
		conditions = append(conditions, GateCondOutput{
			Type:         string(c.Type),
			Threshold:    c.Threshold,
			Required:     required,
			Check:        c.Check,
			Expected:     c.Expected,
			Deliverables: c.Deliverables,
			NodeID:       node.ID,
			NodeName:     node.Name,
		})
	}
	return conditions
}

func SubstituteGateConditions(conditions []GateCondOutput, vars map[string]string) []GateCondOutput {
	result := make([]GateCondOutput, len(conditions))
	for i, c := range conditions {
		result[i] = c
		result[i].Check = substituteVars(c.Check, vars)
		result[i].Expected = substituteVars(c.Expected, vars)
	}
	return result
}

func BuildNextOptionsFromGateFallback(node *flow.FlowNode, edges []flow.FlowEdge, nodeMap map[string]*flow.FlowNode) []NextOption {
	var gateCfg flow.GateNodeConfig
	if node.Config != nil {
		_ = json.Unmarshal(node.Config, &gateCfg)
	}

	options := make([]NextOption, 0, len(edges))

	for i, e := range edges {
		role := ""
		name := ""
		if target, ok := nodeMap[e.To]; ok {
			role = extractRole(target)
			name = target.Name
		}

		var condition *string
		isDefault := false

		if gateCfg.OnPass != "" && e.To == gateCfg.OnPass {
			passedExpr := "gate.passed"
			condition = &passedExpr
			isDefault = true
		} else if gateCfg.OnFail != "" && e.To == gateCfg.OnFail {
			failExpr := "!gate.passed"
			condition = &failExpr
		}

		if i == 0 && !isDefault {
			isDefault = true
		}

		options = append(options, NextOption{
			NodeID:    e.To,
			Name:      name,
			Role:      role,
			Condition: condition,
			IsDefault: isDefault,
		})
	}

	return options
}

func IsPositiveCondition(expr string) bool {
	return expr == "gate.passed" || expr == "true" ||
		(len(expr) > 0 && expr[0] != '!')
}
