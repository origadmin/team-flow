package flow

import (
	"encoding/json"
	"testing"
)

func TestSerializeFlow_Basic(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "test-flow", Description: "A test"},
		Nodes: []FlowNode{
			{ID: "start", Type: NodeTypePhase, Name: "Start", Config: mustMarshal(PhaseConfig{Role: "Triage"})},
			{ID: "end", Type: NodeTypeTerminal, Name: "End", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{
			{From: "start", To: "end"},
		},
	}

	data, err := SerializeFlow(flow)
	if err != nil {
		t.Fatalf("SerializeFlow error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("serialized output is not valid JSON: %v", err)
	}
	if parsed["version"] != "v3" {
		t.Errorf("expected version v3, got %v", parsed["version"])
	}
}

func TestSerializeFlow_Nil(t *testing.T) {
	_, err := SerializeFlow(nil)
	if err == nil {
		t.Fatal("expected error for nil flow")
	}
}

func TestRoundTrip(t *testing.T) {
	original := []byte(`{
		"version": "v3",
		"metadata": {
			"name": "round-trip-test",
			"description": "Round trip test",
			"author": "test",
			"tags": ["test", "round-trip"]
		},
		"config": {
			"task_type": "feature",
			"auto_dispatch": true,
			"parallel_limit": 3
		},
		"nodes": [
			{
				"id": "triage",
				"type": "phase",
				"name": "Triage",
				"description": "Triage phase",
				"config": {"role": "Triage", "auto_dispatch": true},
				"docs": [
					{
						"name": "SPEC.md",
						"format": "markdown",
						"path": "/docs/SPEC.md",
						"required": true
					}
				],
				"on_enter": [{"action": "update_task_phase", "phase": "triage"}],
				"on_exit": [{"action": "log"}]
			},
			{
				"id": "done",
				"type": "terminal",
				"name": "Done",
				"config": {"status": "success"}
			}
		],
		"edges": [
			{"from": "triage", "to": "done", "type": "sequential"}
		],
		"variables": {
			"docs_path": "/project/docs"
		}
	}`)

	flow1, err := ParseFlow(original)
	if err != nil {
		t.Fatalf("first parse error: %v", err)
	}

	serialized, err := SerializeFlow(flow1)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	flow2, err := ParseFlow(serialized)
	if err != nil {
		t.Fatalf("second parse error: %v", err)
	}

	if !FlowsEqual(flow1, flow2) {
		t.Error("round-trip failed: flows are not equal")
		t.Logf("flow1: %s", mustJSON(flow1))
		t.Logf("flow2: %s", mustJSON(flow2))
	}
}

func TestRoundTrip_ComplexFlow(t *testing.T) {
	autoDispatch := true
	parallelLimit := 3
	required := true

	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "complex", Description: "Complex flow", Author: "test", Tags: []string{"complex"}},
		Config: &FlowConfig{
			TaskType:       TaskTypeFeature,
			AutoDispatch:   &autoDispatch,
			ParallelLimit:  &parallelLimit,
		},
		Nodes: []FlowNode{
			{
				ID:   "triage",
				Type: NodeTypePhase,
				Name: "Triage",
				Config: mustMarshal(PhaseConfig{
					Role:         "Triage",
					AutoDispatch: &autoDispatch,
					Deliverables: []string{"SPEC.md"},
				}),
				Docs: []DocSpec{
					{Name: "SPEC.md", Format: "markdown", Path: "/docs/SPEC.md", Required: &required},
				},
				Gates: []GateConfig{
					{Type: GateTypeDeliverableCheck, Required: []string{"SPEC.md"}},
				},
				OnEnter: []Action{
					{Action: ActionUpdateTaskPhase, Phase: "triage"},
				},
				OnExit: []Action{
					{Action: ActionLog},
				},
				OnError: &ErrorHandler{
					Strategy: ErrorStrategyRetry,
					MaxRetries: intPtr(2),
				},
			},
			{
				ID:   "done",
				Type: NodeTypeTerminal,
				Name: "Done",
				Config: mustMarshal(TerminalConfig{Status: TerminalSuccess}),
			},
		},
		Edges: []FlowEdge{
			{ID: "e1", From: "triage", To: "done", Type: EdgeTypeSequential},
		},
		Variables: map[string]interface{}{
			"docs_path": "/project/docs",
		},
	}

	data, err := SerializeFlow(flow)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	reparsed, err := ParseFlow(data)
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	if !FlowsEqual(flow, reparsed) {
		t.Error("round-trip failed for complex flow")
	}
}

func TestRoundTrip_AllNodeConfigs(t *testing.T) {
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "all-configs"},
		Nodes: []FlowNode{
			{ID: "phase1", Type: NodeTypePhase, Name: "Phase", Config: mustMarshal(PhaseConfig{Role: "Dev"})},
			{ID: "gate1", Type: NodeTypeGate, Name: "Gate", Config: mustMarshal(GateNodeConfig{Conditions: []GateCondition{{Type: GateCondTestsPass}}})},
			{ID: "branch1", Type: NodeTypeBranch, Name: "Branch", Config: mustMarshal(BranchConfig{Conditions: []BranchCondition{{When: "true", Goto: "phase1"}}})},
			{ID: "parallel1", Type: NodeTypeParallel, Name: "Parallel", Config: mustMarshal(ParallelConfig{Branches: []ParallelBranch{{Node: "phase1"}}})},
			{ID: "subflow1", Type: NodeTypeSubflow, Name: "Subflow", Config: mustMarshal(SubflowConfig{FlowRef: "builtin:bugfix"})},
			{ID: "loop1", Type: NodeTypeLoop, Name: "Loop", Config: mustMarshal(LoopConfig{MaxAttempts: intPtr(3)})},
			{ID: "manual1", Type: NodeTypeManual, Name: "Manual", Config: mustMarshal(ManualConfig{Approvers: []Approver{{Role: "TechLead"}}})},
			{ID: "event1", Type: NodeTypeEvent, Name: "Event", Config: mustMarshal(EventConfig{Trigger: EventTriggerConfig{Type: TriggerWebhook}})},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done", Config: mustMarshal(TerminalConfig{Status: TerminalSuccess})},
		},
		Edges: []FlowEdge{{From: "phase1", To: "done"}},
	}

	data, err := SerializeFlow(flow)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	reparsed, err := ParseFlow(data)
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	if !FlowsEqual(flow, reparsed) {
		t.Error("round-trip failed for all node configs")
	}
}

func TestFlowsEqual(t *testing.T) {
	a := &Flow{Version: "v3", Metadata: FlowMetadata{Name: "test"}}
	b := &Flow{Version: "v3", Metadata: FlowMetadata{Name: "test"}}
	if !FlowsEqual(a, b) {
		t.Error("expected equal flows")
	}

	c := &Flow{Version: "v3", Metadata: FlowMetadata{Name: "different"}}
	if FlowsEqual(a, c) {
		t.Error("expected different flows")
	}

	if FlowsEqual(nil, a) {
		t.Error("expected nil != non-nil")
	}
	if !FlowsEqual((*Flow)(nil), (*Flow)(nil)) {
		t.Error("expected nil == nil")
	}
}

func intPtr(v int) *int {
	return &v
}

func mustJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
