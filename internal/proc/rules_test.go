package proc

import (
	"testing"

	"github.com/origadmin/team-flow/internal/flow"
)

// TestRuleLoadingChain 验证规则加载完整链路：
// 1) 节点 components.rules -> flow 级 rules -> team 级 rules
// 2) 角色 roleRules（从 prompt YAML 的 ai.rules 解析）
// 3) 终端/gate/start 节点的规则回退
func TestRuleLoadingChain(t *testing.T) {
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"

	team, err := LoadTeam(root)
	if err != nil || team == nil {
		t.Fatalf("LoadTeam failed: err=%v team=%v", err, team)
	}

	fl, err := LoadFlow(root, "skill-dev-flow")
	if err != nil {
		t.Fatalf("LoadFlow(skill-dev-flow): %v", err)
	}

	// 1. 验证 team.json 中注册的规则
	t.Logf("Team rules registered: %d", len(team.Rules))
	for _, r := range team.Rules {
		t.Logf("  team rule: id=%s source=%s", r.ID, r.Source)
	}
	if len(team.Rules) == 0 {
		t.Error("team.json should register rules")
	}

	// 2. 验证各关键节点的 rules 提取
	nodeIDs := []string{"a26c80", "c3d4e5", "2ef533", "7837a8"}
	for _, nid := range nodeIDs {
		var node *flow.FlowNode
		for i := range fl.Nodes {
			if fl.Nodes[i].ID == nid {
				node = &fl.Nodes[i]
				break
			}
		}
		if node == nil {
			t.Errorf("node %s not found in skill-dev-flow", nid)
			continue
		}

		ri := resolveRoleInfo(node, fl, team, root)
		rules := extractRules(node, fl, team, ri.roleRules, root)

		t.Logf("node=%s(%s) nodeComponents=%v roleRules=%v extractedRules=%d",
			nid, node.Type,
			node.Components != nil && len(node.Components.Rules) > 0,
			len(ri.roleRules), len(rules))
		for _, r := range rules {
			t.Logf("  extracted rule: ref=%s source=%s name=%s", r.Ref, r.Source, r.Name)
		}

		// 检查: start/gate/terminal 节点至少应该有 qg4
		if node.Type == flow.NodeTypeStart || node.Type == flow.NodeTypeTerminal || node.Type == flow.NodeTypeGate {
			hasQG4 := false
			for _, r := range rules {
				if r.Ref == "qg4" {
					hasQG4 = true
					break
				}
			}
			if !hasQG4 {
				t.Errorf("  ❌ node %s(%s) should have qg4 default rule", nid, node.Type)
			}
		}
	}

	// 3. 验证 PM 角色的 roleRules
	//    从 pm.md 的 frontmatter 解析 ai.rules → 这些 rules 在 c3d4e5 节点应该生效
	for _, r := range team.Roles {
		if r.ID == "pm" {
			if resolved, err := loadRoleFromPromptSource(r.ID, r.PromptSource, root); err == nil && resolved != nil {
				t.Logf("pm roleRules from frontmatter: %v", resolved.Rules)
				// 检查这些 rules 是否在 team.json 中注册
				for _, ruleID := range resolved.Rules {
					found := false
					for _, tr := range team.Rules {
						if tr.ID == ruleID {
							found = true
							break
						}
					}
					if !found {
						t.Logf("  ⚠ pm role rule '%s' NOT registered in team.json rules", ruleID)
					}
				}
			}
			break
		}
	}
}
