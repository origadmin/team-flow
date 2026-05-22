package flow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseFlow_ValidJSON(t *testing.T) {
	data := []byte(`{
		"version": "v3",
		"metadata": {
			"name": "test-flow",
			"description": "A test flow"
		},
		"nodes": [
			{
				"id": "triage",
				"type": "phase",
				"name": "Triage",
				"config": {
					"role": "Triage",
					"auto_dispatch": true
				}
			},
			{
				"id": "done",
				"type": "terminal",
				"name": "Done",
				"config": {
					"status": "success"
				}
			}
		],
		"edges": [
			{"from": "triage", "to": "done"}
		]
	}`)

	flow, err := ParseFlow(data)
	if err != nil {
		t.Fatalf("ParseFlow returned error: %v", err)
	}
	if flow.Version != "v3" {
		t.Errorf("expected version v3, got %s", flow.Version)
	}
	if flow.Metadata.Name != "test-flow" {
		t.Errorf("expected name test-flow, got %s", flow.Metadata.Name)
	}
	if len(flow.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(flow.Nodes))
	}
	if flow.Nodes[0].ID != "triage" {
		t.Errorf("expected first node id triage, got %s", flow.Nodes[0].ID)
	}
	if flow.Nodes[0].Type != NodeTypePhase {
		t.Errorf("expected first node type phase, got %s", flow.Nodes[0].Type)
	}
	if flow.Nodes[1].Type != NodeTypeTerminal {
		t.Errorf("expected second node type terminal, got %s", flow.Nodes[1].Type)
	}
	if len(flow.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(flow.Edges))
	}
	if flow.Edges[0].From != "triage" || flow.Edges[0].To != "done" {
		t.Errorf("expected edge triage->done, got %s->%s", flow.Edges[0].From, flow.Edges[0].To)
	}
}

func TestParseFlow_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json}`)
	_, err := ParseFlow(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseFlow_EmptyJSON(t *testing.T) {
	data := []byte(`{}`)
	flow, err := ParseFlow(data)
	if err != nil {
		t.Fatalf("ParseFlow returned error: %v", err)
	}
	if flow.Version != "" {
		t.Errorf("expected empty version, got %s", flow.Version)
	}
	if flow.Metadata.Name != "" {
		t.Errorf("expected empty name, got %s", flow.Metadata.Name)
	}
}

func TestParseFlowWithVars(t *testing.T) {
	data := []byte(`{
		"version": "v3",
		"metadata": {"name": "test-flow"},
		"nodes": [
			{
				"id": "triage",
				"type": "phase",
				"name": "Triage",
				"docs": [
					{
						"name": "SPEC.md",
						"format": "markdown",
						"path": "{DOCS_INTERNAL}/requirements/{task_id}/SPEC.md",
						"required": true
					}
				]
			},
			{
				"id": "done",
				"type": "terminal",
				"name": "Done",
				"config": {"status": "success"}
			}
		],
		"edges": [{"from": "triage", "to": "done"}]
	}`)

	vars := map[string]string{
		"DOCS_INTERNAL": "/project/docs",
		"task_id":       "framework-34p",
	}

	flow, err := ParseFlowWithVars(data, vars)
	if err != nil {
		t.Fatalf("ParseFlowWithVars returned error: %v", err)
	}

	expectedPath := "/project/docs/requirements/framework-34p/SPEC.md"
	if len(flow.Nodes[0].Docs) == 0 {
		t.Fatal("expected docs on first node")
	}
	if flow.Nodes[0].Docs[0].Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, flow.Nodes[0].Docs[0].Path)
	}
}

func TestParseFlow_AllNodeTypes(t *testing.T) {
	data := []byte(`{
		"version": "v3",
		"metadata": {"name": "all-types"},
		"nodes": [
			{"id": "phase1", "type": "phase", "name": "Phase", "config": {"role": "Dev"}},
			{"id": "gate1", "type": "gate", "name": "Gate", "config": {"conditions": [{"type": "tests_pass"}]}},
			{"id": "branch1", "type": "branch", "name": "Branch", "config": {"conditions": [{"when": "true", "goto": "phase1"}]}},
			{"id": "parallel1", "type": "parallel", "name": "Parallel", "config": {"branches": [{"node": "phase1"}]}},
			{"id": "subflow1", "type": "subflow", "name": "Subflow", "config": {"flow_ref": "builtin:bugfix-flow"}},
			{"id": "loop1", "type": "loop", "name": "Loop", "config": {"max_attempts": 3}},
			{"id": "manual1", "type": "manual", "name": "Manual", "config": {"approvers": [{"role": "TechLead"}]}},
			{"id": "event1", "type": "event", "name": "Event", "config": {"trigger": {"type": "webhook"}}},
			{"id": "done", "type": "terminal", "name": "Done", "config": {"status": "success"}}
		],
		"edges": [{"from": "phase1", "to": "done"}]
	}`)

	flow, err := ParseFlow(data)
	if err != nil {
		t.Fatalf("ParseFlow returned error: %v", err)
	}

	expectedTypes := []NodeType{
		NodeTypePhase, NodeTypeGate, NodeTypeBranch, NodeTypeParallel,
		NodeTypeSubflow, NodeTypeLoop, NodeTypeManual, NodeTypeEvent, NodeTypeTerminal,
	}
	for i, expected := range expectedTypes {
		if flow.Nodes[i].Type != expected {
			t.Errorf("node %d: expected type %s, got %s", i, expected, flow.Nodes[i].Type)
		}
	}
}

func TestParseFlowFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-flow.json")

	flowData := Flow{
		Version: "v3",
		Metadata: FlowMetadata{Name: "file-test"},
		Nodes: []FlowNode{
			{ID: "start", Type: NodeTypePhase, Name: "Start"},
			{ID: "end", Type: NodeTypeTerminal, Name: "End", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{{From: "start", To: "end"}},
	}

	data, _ := json.MarshalIndent(flowData, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	flow, err := ParseFlowFile(path)
	if err != nil {
		t.Fatalf("ParseFlowFile returned error: %v", err)
	}
	if flow.Metadata.Name != "file-test" {
		t.Errorf("expected name file-test, got %s", flow.Metadata.Name)
	}
}

func TestParseFlowFile_NotFound(t *testing.T) {
	_, err := ParseFlowFile("/nonexistent/path/flow.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestApplyOverrides(t *testing.T) {
	base := &Flow{
		Version: "v3",
		Metadata: FlowMetadata{Name: "base-flow"},
		Nodes: []FlowNode{
			{ID: "triage", Type: NodeTypePhase, Name: "Triage", Description: "Original"},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{{From: "triage", To: "done"}},
	}

	overrides := FlowOverrides{
		Nodes: []FlowNodeOverride{
			{ID: "triage", Name: "Custom Triage", Description: "Overridden"},
		},
	}

	result := ApplyOverrides(base, overrides)
	if result.Nodes[0].Name != "Custom Triage" {
		t.Errorf("expected overridden name 'Custom Triage', got %s", result.Nodes[0].Name)
	}
	if result.Nodes[0].Description != "Overridden" {
		t.Errorf("expected overridden description, got %s", result.Nodes[0].Description)
	}
	if result.Nodes[0].ID != "triage" {
		t.Errorf("ID should not change, got %s", result.Nodes[0].ID)
	}
	if result.Nodes[1].Name != "Done" {
		t.Errorf("unrelated node should not change, got %s", result.Nodes[1].Name)
	}
}

func TestApplyOverrides_Config(t *testing.T) {
	autoDispatch := true
	base := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "base"},
		Nodes:    []FlowNode{{ID: "start", Type: NodeTypePhase, Name: "Start"}},
		Edges:    []FlowEdge{},
		Config:   &FlowConfig{TaskType: TaskTypeFeature, AutoDispatch: &autoDispatch},
	}

	newConfig := &FlowConfig{TaskType: TaskTypeBug}
	result := ApplyOverrides(base, FlowOverrides{Config: newConfig})
	if result.Config.TaskType != TaskTypeBug {
		t.Errorf("expected overridden task type bug, got %s", result.Config.TaskType)
	}
}

func TestApplyOverrides_Edges(t *testing.T) {
	base := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "base"},
		Nodes: []FlowNode{
			{ID: "a", Type: NodeTypePhase, Name: "A"},
			{ID: "b", Type: NodeTypeTerminal, Name: "B", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{{From: "a", To: "b"}},
	}

	newEdges := []FlowEdge{{From: "a", To: "b", Type: EdgeTypeConditional}}
	result := ApplyOverrides(base, FlowOverrides{Edges: newEdges})
	if len(result.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(result.Edges))
	}
	if result.Edges[0].Type != EdgeTypeConditional {
		t.Errorf("expected conditional edge type, got %s", result.Edges[0].Type)
	}
}

func TestApplyOverrides_Variables(t *testing.T) {
	base := &Flow{
		Version:    "v3",
		Metadata:   FlowMetadata{Name: "base"},
		Nodes:      []FlowNode{{ID: "start", Type: NodeTypePhase, Name: "Start"}},
		Edges:      []FlowEdge{},
		Variables:  map[string]interface{}{"DOCS_INTERNAL": "/old"},
	}

	overrideVars := map[string]interface{}{"DOCS_INTERNAL": "/new", "extra": "value"}
	result := ApplyOverrides(base, FlowOverrides{Variables: overrideVars})
	if result.Variables["DOCS_INTERNAL"] != "/new" {
		t.Errorf("expected overridden DOCS_INTERNAL /new, got %v", result.Variables["DOCS_INTERNAL"])
	}
	if result.Variables["extra"] != "value" {
		t.Errorf("expected extra variable value, got %v", result.Variables["extra"])
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
