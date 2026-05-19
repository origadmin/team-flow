package flow

import (
	"encoding/json"
	"testing"
)

func validFlow() *Flow {
	return &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "test-flow"},
		Nodes: []FlowNode{
			{ID: "triage", Type: NodeTypePhase, Name: "Triage", Config: mustMarshal(PhaseConfig{Role: "Triage"})},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{
			{From: "triage", To: "done"},
		},
	}
}

func TestValidateFlow_ValidFlow(t *testing.T) {
	result := ValidateFlow(validFlow())
	if !result.Valid {
		t.Errorf("expected valid flow, got errors: %v", result.Errors)
	}
}

func TestValidateFlow_NilFlow(t *testing.T) {
	result := ValidateFlow(nil)
	if result.Valid {
		t.Fatal("expected invalid for nil flow")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error")
	}
}

func TestValidateFlow_MissingVersion(t *testing.T) {
	flow := validFlow()
	flow.Version = ""
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for missing version")
	}
	found := false
	for _, e := range result.Errors {
		if e.Field == "version" {
			found = true
		}
	}
	if !found {
		t.Error("expected error on version field")
	}
}

func TestValidateFlow_MissingMetadataName(t *testing.T) {
	flow := validFlow()
	flow.Metadata.Name = ""
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for missing metadata.name")
	}
	found := false
	for _, e := range result.Errors {
		if e.Field == "metadata.name" {
			found = true
		}
	}
	if !found {
		t.Error("expected error on metadata.name field")
	}
}

func TestValidateFlow_NoTerminalNode(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "no-terminal"},
		Nodes: []FlowNode{
			{ID: "triage", Type: NodeTypePhase, Name: "Triage"},
		},
		Edges: []FlowEdge{},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for no terminal node")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == "flow must have at least one terminal node" {
			found = true
		}
	}
	if !found {
		t.Error("expected error about missing terminal node")
	}
}

func TestValidateFlow_DuplicateNodeID(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "dup-id"},
		Nodes: []FlowNode{
			{ID: "triage", Type: NodeTypePhase, Name: "Triage 1"},
			{ID: "triage", Type: NodeTypePhase, Name: "Triage 2"},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for duplicate node id")
	}
	found := false
	for _, e := range result.Errors {
		if e.NodeID == "triage" && e.Field == "nodes[].id" {
			found = true
		}
	}
	if !found {
		t.Error("expected error about duplicate node id")
	}
}

func TestValidateFlow_InvalidNodeType(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "bad-type"},
		Nodes: []FlowNode{
			{ID: "bad", Type: NodeType("invalid"), Name: "Bad"},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for invalid node type")
	}
	found := false
	for _, e := range result.Errors {
		if e.NodeID == "bad" && e.Field == "nodes[].type" {
			found = true
		}
	}
	if !found {
		t.Error("expected error about invalid node type")
	}
}

func TestValidateFlow_EdgeReferencesNonExistentNode(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "bad-edge"},
		Nodes: []FlowNode{
			{ID: "triage", Type: NodeTypePhase, Name: "Triage"},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{
			{From: "triage", To: "nonexistent"},
		},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for edge referencing non-existent node")
	}
	found := false
	for _, e := range result.Errors {
		if e.Field == "edges[].to" && e.Message == "edge to references non-existent node: nonexistent" {
			found = true
		}
	}
	if !found {
		t.Error("expected error about edge referencing non-existent node")
	}
}

func TestValidateFlow_OrphanNodeWarning(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "orphan"},
		Nodes: []FlowNode{
			{ID: "triage", Type: NodeTypePhase, Name: "Triage"},
			{ID: "orphan", Type: NodeTypePhase, Name: "Orphan"},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{
			{From: "triage", To: "done"},
		},
	}
	result := ValidateFlow(flow)
	if !result.Valid {
		t.Fatalf("expected valid (orphan is a warning, not error), got errors: %v", result.Errors)
	}
	found := false
	for _, w := range result.Warnings {
		if w.NodeID == "orphan" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about orphan node")
	}
}

func TestValidateFlow_PhaseMissingRequiredDocs(t *testing.T) {
	required := true
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "missing-docs"},
		Nodes: []FlowNode{
			{
				ID:   "triage",
				Type: NodeTypePhase,
				Name: "Triage",
				Docs: []DocSpec{
					{Name: "SPEC.md", Format: "markdown", Path: "", Required: &required},
				},
			},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{{From: "triage", To: "done"}},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for phase missing required doc path")
	}
	found := false
	for _, e := range result.Errors {
		if e.NodeID == "triage" && e.Field == "nodes[].docs" {
			found = true
		}
	}
	if !found {
		t.Error("expected error about phase missing required docs")
	}
}

func TestValidateFlow_EmptyNodes(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "empty-nodes"},
		Nodes:    []FlowNode{},
		Edges:    []FlowEdge{},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for empty nodes")
	}
}

func TestValidateFlow_NodeMissingName(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "no-name"},
		Nodes: []FlowNode{
			{ID: "unnamed", Type: NodeTypePhase},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for node missing name")
	}
}

func TestValidateFlow_SubflowInvalidFlowRef(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "bad-subflow"},
		Nodes: []FlowNode{
			{ID: "sub", Type: NodeTypeSubflow, Name: "Sub", Config: mustMarshal(SubflowConfig{FlowRef: "invalid-no-colon"})},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for subflow with invalid flow_ref")
	}
	found := false
	for _, e := range result.Errors {
		if e.NodeID == "sub" && e.Field == "nodes[].config.flow_ref" {
			found = true
		}
	}
	if !found {
		t.Error("expected error about invalid subflow flow_ref")
	}
}

func TestValidateFlow_SubflowValidFlowRef(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "good-subflow"},
		Nodes: []FlowNode{
			{ID: "sub", Type: NodeTypeSubflow, Name: "Sub", Config: mustMarshal(SubflowConfig{FlowRef: "builtin:bugfix-flow"})},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{{From: "sub", To: "done"}},
	}
	result := ValidateFlow(flow)
	if !result.Valid {
		t.Errorf("expected valid flow, got errors: %v", result.Errors)
	}
}

func TestValidateFlow_EdgeFromNonExistent(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "bad-from"},
		Nodes: []FlowNode{
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{
			{From: "ghost", To: "done"},
		},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for edge from non-existent node")
	}
}

func TestValidateFlow_DuplicateEdgeID(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "dup-edge"},
		Nodes: []FlowNode{
			{ID: "a", Type: NodeTypePhase, Name: "A"},
			{ID: "b", Type: NodeTypePhase, Name: "B"},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{
			{ID: "e1", From: "a", To: "b"},
			{ID: "e1", From: "b", To: "done"},
		},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for duplicate edge id")
	}
}

func TestUnmarshalNodeConfig(t *testing.T) {
	raw := json.RawMessage(`{"role":"Dev","auto_dispatch":true}`)
	var cfg PhaseConfig
	err := unmarshalNodeConfig(raw, &cfg)
	if err != nil {
		t.Fatalf("unmarshalNodeConfig error: %v", err)
	}
	if cfg.Role != "Dev" {
		t.Errorf("expected role Dev, got %s", cfg.Role)
	}
	if cfg.AutoDispatch == nil || !*cfg.AutoDispatch {
		t.Error("expected auto_dispatch true")
	}
}

func TestUnmarshalNodeConfig_Empty(t *testing.T) {
	var cfg PhaseConfig
	err := unmarshalNodeConfig(nil, &cfg)
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestValidateComponents_RuleWithoutDescription(t *testing.T) {
	flow := validFlow()
	flow.Components = &ComponentRegistry{
		Rules: []RuleDefinition{
			{ID: "output-guard", Name: "Output Guard"},
		},
	}
	result := ValidateFlow(flow)
	if !result.Valid {
		t.Fatalf("expected valid (no description is warning, not error), got errors: %v", result.Errors)
	}
	found := false
	for _, w := range result.Warnings {
		if w.Message == "rule output-guard has no description — AI cannot follow a rule without content" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about rule without description")
	}
}

func TestValidateComponents_RuleWithDescription(t *testing.T) {
	flow := validFlow()
	flow.Components = &ComponentRegistry{
		Rules: []RuleDefinition{
			{ID: "output-guard", Name: "Output Guard", Description: "Every phase must produce deliverables"},
		},
	}
	result := ValidateFlow(flow)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	for _, w := range result.Warnings {
		if w.Field == "components.rules" && contains(w.Message, "no description") {
			t.Errorf("unexpected warning about no description: %s", w.Message)
		}
	}
}

func TestValidateComponents_UndefinedRuleRef(t *testing.T) {
	flow := validFlow()
	flow.Components = &ComponentRegistry{
		Rules: []RuleDefinition{
			{ID: "output-guard", Name: "Output Guard", Description: "Produce deliverables"},
		},
	}
	flow.Nodes[0].Components = &NodeComponents{
		Rules: []ComponentRef{
			{Ref: "undefined-rule", Source: SourceBuiltin},
		},
	}
	result := ValidateFlow(flow)
	if !result.Valid {
		t.Fatalf("expected valid (undefined rule ref is warning), got errors: %v", result.Errors)
	}
	found := false
	for _, w := range result.Warnings {
		if contains(w.Message, "references undefined rule: undefined-rule") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about undefined rule reference")
	}
}

func TestValidateComponents_RuleMissingID(t *testing.T) {
	flow := validFlow()
	flow.Components = &ComponentRegistry{
		Rules: []RuleDefinition{
			{Name: "No ID Rule"},
		},
	}
	result := ValidateFlow(flow)
	if result.Valid {
		t.Fatal("expected invalid for rule missing id")
	}
	found := false
	for _, e := range result.Errors {
		if e.Message == "rule definition missing id" {
			found = true
		}
	}
	if !found {
		t.Error("expected error about rule missing id")
	}
}

func TestValidateComponents_NilComponents(t *testing.T) {
	flow := validFlow()
	flow.Components = nil
	result := ValidateFlow(flow)
	if !result.Valid {
		t.Fatalf("expected valid with nil components, got errors: %v", result.Errors)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
