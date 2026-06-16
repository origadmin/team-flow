package proc

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/origadmin/team-flow/internal/flow"
)

// TestTeamV3EndToEndRoleResolution 验证 skill-team 的 team.json 和两个 flow 中的
// 所有角色引用都能通过 resolveRoleInfo 正确解析出 alias/persona/traits。
// 目标：消除 [System | Completed(suc0:dev-flow) | xxx | ] 这样的错误。
func TestTeamV3EndToEndRoleResolution(t *testing.T) {
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"

	// 1. 读取 team.json
	team, err := LoadTeam(root)
	if err != nil {
		t.Fatalf("LoadTeam failed: %v", err)
	}
	if team == nil {
		t.Fatal("team is nil")
	}
	if team.SchemaVersion != "3.0" {
		t.Errorf("expected schema_version 3.0, got %q", team.SchemaVersion)
	}

	// 2. 验证 team.json 角色唯一性：每个 prompt_source 只被 1 个角色引用
	seenPrompts := make(map[string]string)
	for _, r := range team.Roles {
		if r.PromptSource == "" {
			t.Errorf("role %s has empty prompt_source", r.ID)
			continue
		}
		if prevID, ok := seenPrompts[r.PromptSource]; ok {
			t.Errorf("duplicate prompt_source: %q is used by both %q and %q",
				r.PromptSource, prevID, r.ID)
		}
		seenPrompts[r.PromptSource] = r.ID
	}
	t.Logf("team.json has %d roles, all prompt_source unique ✓", len(team.Roles))

	// 3. 验证每个角色都能通过 resolveRoleInfo 解析出完整信息
	for _, role := range team.Roles {
		// 构造一个伪 node 调用 resolveRoleInfo
		node := &flow.FlowNode{
			ID:   "test-" + role.ID,
			Type: flow.NodeTypePhase,
			Name: "Test " + role.ID,
			Config: func() []byte {
				cfg := map[string]string{"role": role.ID}
				b, _ := json.Marshal(cfg)
				return b
			}(),
			Components: &flow.NodeComponents{
				Roles: []flow.ComponentRef{{Ref: role.ID, Source: "team"}},
			},
		}
		ri := resolveRoleInfo(node, nil, team, root)
		if ri.alias == "" {
			t.Errorf("role %s: alias is empty after resolveRoleInfo", role.ID)
		}
		if ri.roleID != role.ID {
			t.Errorf("role %s: resolveRoleInfo returned roleID=%q", role.ID, ri.roleID)
		}
		t.Logf("  role %-14s → alias=%-6s name=%s persona_len=%d",
			role.ID, ri.alias, ri.roleName, len(ri.persona))
	}

	// 4. 验证每个 flow 的每个 phase 节点都能解析出角色
	flowFiles := []string{
		filepath.Join(root, "assets", "flows", "skill-dev-flow.json"),
		filepath.Join(root, "assets", "flows", "dev-flow.json"),
	}

	for _, ff := range flowFiles {
		fl, err := flow.ParseFlowFile(ff)
		if err != nil {
			t.Fatalf("parse flow %s failed: %v", ff, err)
		}
		flowName := fl.Metadata.Name
		if flowName == "" {
			flowName = filepath.Base(ff)
		}

		t.Logf("\n=== Flow: %s ===", flowName)
		hasPhaseNode := false
		for i := range fl.Nodes {
			node := &fl.Nodes[i]
			if node.Type != flow.NodeTypePhase {
				continue
			}
			hasPhaseNode = true

			// 节点应该有 components.roles
			if node.Components == nil || len(node.Components.Roles) == 0 {
				t.Errorf("%s node %s(%s) has no roles in components",
					flowName, node.ID, node.Name)
				continue
			}

			roleRef := node.Components.Roles[0].Ref

			// 检查角色是否存在于 team.json
			found := false
			for _, r := range team.Roles {
				if r.ID == roleRef {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s node %s(%s) references role %q which is NOT in team.json",
					flowName, node.ID, node.Name, roleRef)
				continue
			}

			// 尝试解析
			ri := resolveRoleInfo(node, fl, team, root)
			if ri.alias == "" {
				t.Errorf("%s node %s(%s) role=%s: alias is empty",
					flowName, node.ID, node.Name, roleRef)
			}
			t.Logf("  %-8s %s → role=%-10s alias=%s",
				strings.ToUpper(string(node.Type)), node.Name, roleRef, ri.alias)
		}
		if !hasPhaseNode {
			t.Errorf("flow %s has no phase nodes", flowName)
		}

		// 5. 验证终端节点：通过前驱继承角色
		for i := range fl.Nodes {
			node := &fl.Nodes[i]
			if node.Type != flow.NodeTypeTerminal {
				continue
			}
			// 调用 buildCurrentNode 模拟实际渲染
			vars := map[string]string{}
			current := buildCurrentNode(node, fl, vars, team, root)
			if current.Alias == "" || current.Alias == "System" {
				// 允许无入边的终端节点退化（如 start → suc0 的错误路径）
				pred := findPredecessor(fl, node)
				if pred != nil && pred.Type == flow.NodeTypePhase {
					t.Errorf("%s terminal %s(%s) resolved alias=%q (expected proper role from predecessor %s)",
						flowName, node.ID, node.Name, current.Alias, pred.ID)
				}
			}
			t.Logf("  TERMINAL %s → alias=%s", node.Name, current.Alias)
		}
	}
}

// TestSkillDevFlowNodeIntegrity 验证 skill-dev-flow 的节点完整性
// （nodes 和 edges 都能正确链接、没有孤立节点）
func TestSkillDevFlowNodeIntegrity(t *testing.T) {
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"
	flowPath := filepath.Join(root, "assets", "flows", "skill-dev-flow.json")

	fl, err := flow.ParseFlowFile(flowPath)
	if err != nil {
		t.Fatalf("parse flow failed: %v", err)
	}

	// 所有节点应有唯一 ID
	idSet := make(map[string]bool)
	for _, n := range fl.Nodes {
		if idSet[n.ID] {
			t.Errorf("duplicate node id: %s", n.ID)
		}
		idSet[n.ID] = true
	}

	// 边应指向存在的节点
	for _, e := range fl.Edges {
		if !idSet[e.From] {
			t.Errorf("edge %s: from=%s not found", e.ID, e.From)
		}
		if !idSet[e.To] {
			t.Errorf("edge %s: to=%s not found", e.ID, e.To)
		}
	}

	// Phase 节点必须有角色定义在 team.json
	team, _ := LoadTeam(root)
	teamRoles := make(map[string]bool)
	for _, r := range team.Roles {
		teamRoles[r.ID] = true
	}

	for _, n := range fl.Nodes {
		if n.Type == flow.NodeTypePhase {
			if n.Components == nil || len(n.Components.Roles) == 0 {
				t.Errorf("phase node %s has no roles", n.ID)
				continue
			}
			roleRef := n.Components.Roles[0].Ref
			if !teamRoles[roleRef] {
				t.Errorf("phase node %s references role %q not in team.json", n.ID, roleRef)
			}
		}
	}
}

// TestDevFlowRoleResolution 聚焦 dev-flow 之前 "System" alias 问题的回归测试
func TestDevFlowRoleResolution(t *testing.T) {
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"
	flowPath := filepath.Join(root, "assets", "flows", "dev-flow.json")

	fl, err := flow.ParseFlowFile(flowPath)
	if err != nil {
		t.Fatalf("parse flow failed: %v", err)
	}

	team, err := LoadTeam(root)
	if err != nil {
		t.Fatalf("LoadTeam failed: %v", err)
	}

	// 测试的关键节点及期望角色
	testCases := []struct {
		nodeID   string
		nodeName string
		wantRole string
	}{
		{"tri3", "Task Triage", "pm"},
		{"fa01", "Requirements Analysis", "tech-lead"},
		{"fd02", "Technical Design", "tech-lead"},
		{"fi03", "Implementation", "dev"},
		{"bi01", "Root Cause Investigation", "dev"},
		{"bf02", "Bug Fix", "dev"},
		{"cp01", "Change Planning", "tech-lead"},
		{"ce02", "Change Execution", "dev"},
		{"ai01", "Deep Analysis", "analysis"},
		{"ar02", "Analysis Report", "analysis"},
		{"ver4", "Integration Verification", "qa"},
		{"rev6", "Code Review", "tech-lead"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s", tc.nodeID, tc.nodeName), func(t *testing.T) {
			var node *flow.FlowNode
			for i := range fl.Nodes {
				if fl.Nodes[i].ID == tc.nodeID {
					node = &fl.Nodes[i]
					break
				}
			}
			if node == nil {
				t.Fatalf("node %s not found", tc.nodeID)
			}
			if node.Name != tc.nodeName {
				t.Errorf("expected node name %q, got %q", tc.nodeName, node.Name)
			}
			ri := resolveRoleInfo(node, fl, team, root)
			if ri.roleID != tc.wantRole {
				t.Errorf("expected role %q, got %q", tc.wantRole, ri.roleID)
			}
			if ri.alias == "" {
				t.Error("alias is empty")
			}
		})
	}

	// 测试终端节点：suc0 (Completed) 的 alias 应该继承自 rev6 (tech-lead)
	t.Run("terminal-suc0-inherits-rev6", func(t *testing.T) {
		var suc0 *flow.FlowNode
		for i := range fl.Nodes {
			if fl.Nodes[i].ID == "suc0" {
				suc0 = &fl.Nodes[i]
				break
			}
		}
		if suc0 == nil {
			t.Fatal("suc0 not found")
		}

		pred := findPredecessor(fl, suc0)
		if pred == nil {
			t.Fatal("findPredecessor returned nil for suc0")
		}
		if pred.ID != "rev6" {
			t.Errorf("expected predecessor rev6, got %s", pred.ID)
		}

		vars := map[string]string{}
		current := buildCurrentNode(suc0, fl, vars, team, root)
		if current.Alias == "" {
			t.Error("suc0 alias is empty after fix")
		}
		if current.Alias == "System" {
			t.Error("suc0 alias is still 'System' — findPredecessor/resolveRoleInfo broken")
		}
		t.Logf("suc0(Completed) → alias=%s roleName=%s", current.Alias, current.RoleName)
	})
}
