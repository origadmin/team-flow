package proc

import (
	"testing"

	"github.com/origadmin/team-flow/internal/flow"
)

// TestDebugSkillDevFlowEntryEdges 验证 skill-dev-flow 的 a26c80 入口节点的 edges 条件解析
func TestDebugSkillDevFlowEntryEdges(t *testing.T) {
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"

	fl, err := LoadFlow(root, "skill-dev-flow")
	if err != nil {
		t.Fatalf("LoadFlow: %v", err)
	}

	var startNode *flow.FlowNode
	for i := range fl.Nodes {
		if fl.Nodes[i].ID == "a26c80" {
			startNode = &fl.Nodes[i]
			break
		}
	}
	if startNode == nil {
		t.Fatal("a26c80 not found")
	}

	t.Logf("a26c80 type=%s entry=%v", startNode.Type, startNode.Entry)

	// Print all edges from a26c80
	var outgoing []flow.FlowEdge
	for _, e := range fl.Edges {
		if e.From == "a26c80" {
			outgoing = append(outgoing, e)
			t.Logf("  edge: to=%s type=%s conditions=%v",
				e.To, e.Type, e.Conditions)
		}
	}
	t.Logf("total outgoing edges: %d", len(outgoing))

	// Test with task_created context (有任务时)
	ctx1 := map[string]string{"status": "task_created"}
	opts1 := buildNextOptions(fl, "a26c80", ctx1)
	t.Logf("ctx=task_created: next_options count=%d", len(opts1))
	for _, opt := range opts1 {
		cond := ""
		if opt.Condition != nil {
			cond = *opt.Condition
		}
		t.Logf("  option: node_id=%s name=%s condition=%s is_default=%v",
			opt.NodeID, opt.Name, cond, opt.IsDefault)
	}

	// Test with no_task context (无任务时)
	ctx2 := map[string]string{"status": "no_task"}
	opts2 := buildNextOptions(fl, "a26c80", ctx2)
	t.Logf("ctx=no_task: next_options count=%d", len(opts2))
	for _, opt := range opts2 {
		cond := ""
		if opt.Condition != nil {
			cond = *opt.Condition
		}
		t.Logf("  option: node_id=%s name=%s condition=%s is_default=%v",
			opt.NodeID, opt.Name, cond, opt.IsDefault)
	}

	// Check that task_created path routes to c3d4e5 (Project Assessment, pm)
	foundProjAssessment := false
	for _, opt := range opts1 {
		if opt.NodeID == "c3d4e5" {
			foundProjAssessment = true
			t.Logf("✓ c3d4e5 (Project Assessment, pm) found in task_created next_options")
		}
	}
	if !foundProjAssessment {
		t.Errorf("✗ c3d4e5 (Project Assessment) NOT found in task_created next_options — flow dead end!")
	}

	// Check that no_task path routes to 7837a8 (Completed)
	foundCompleted := false
	for _, opt := range opts2 {
		if opt.NodeID == "7837a8" {
			foundCompleted = true
			t.Logf("✓ 7837a8 (Completed) found in no_task next_options")
		}
	}
	if !foundCompleted {
		t.Errorf("✗ 7837a8 (Completed) NOT found in no_task next_options")
	}
}
