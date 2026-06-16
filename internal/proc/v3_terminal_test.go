package proc

import (
	"testing"

	"github.com/origadmin/team-flow/internal/flow"
)

// TestDevFlowTerminalRoleResolution 验证 dev-flow 中终端节点的角色解析是否正确。
// 场景: skill-team 使用 dev-flow，终端节点 suc0 的前驱 rev6 引用 tech-lead 角色。
// 修复前: resolveRoleInfo 找不到 tech-lead → alias 为空 → status_line 显示 "System"
// 修复后: team.json 中添加了 tech-lead → resolveRoleInfo 能正确解析 → status_line 显示正确 alias
func TestDevFlowTerminalRoleResolution(t *testing.T) {
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"

	// 1. 加载 dev-flow
	fl, err := LoadFlow(root, "dev-flow")
	if err != nil {
		t.Fatalf("failed to load dev-flow: %v", err)
	}

	// 2. 加载 skill-team 的 team.json
	team, err := LoadTeam(root)
	if err != nil {
		t.Fatalf("failed to load team: %v", err)
	}
	if team == nil {
		t.Fatal("team is nil")
	}

	// 3. 查找 rev6 (Code Review, 引用 tech-lead)
	var rev6 *flow.FlowNode
	for i := range fl.Nodes {
		if fl.Nodes[i].ID == "rev6" {
			rev6 = &fl.Nodes[i]
			break
		}
	}
	if rev6 == nil {
		t.Fatal("rev6 not found in dev-flow")
	}

	// 4. 验证 rev6 的角色解析
	ri := resolveRoleInfo(rev6, fl, team, root)
	t.Logf("rev6 role resolution: alias=%s, roleName=%s, roleID=%s, persona_len=%d, principal=%v",
		ri.alias, ri.roleName, ri.roleID, len(ri.persona), ri.principal)

	if ri.alias == "" {
		t.Error("rev6 (Code Review → tech-lead) alias should not be empty after fix")
	}
	if ri.roleID != "tech-lead" {
		t.Errorf("expected roleID=tech-lead, got %s", ri.roleID)
	}

	// 5. 查找 suc0 (Completed, terminal node)
	var suc0 *flow.FlowNode
	for i := range fl.Nodes {
		if fl.Nodes[i].ID == "suc0" {
			suc0 = &fl.Nodes[i]
			break
		}
	}
	if suc0 == nil {
		t.Fatal("suc0 not found in dev-flow")
	}

	// 6. 验证 findPredecessor 找到 rev6
	predecessor := findPredecessor(fl, suc0)
	if predecessor == nil {
		t.Fatal("findPredecessor(suc0) should return rev6")
	}
	if predecessor.ID != "rev6" {
		t.Errorf("expected predecessor rev6, got %s", predecessor.ID)
	}

	// 7. 模拟 buildCurrentNode 的终端节点逻辑
	vars := map[string]string{}
	current := buildCurrentNode(suc0, fl, vars, team, root)
	t.Logf("suc0 (Completed) after fix: alias=%s, roleName=%s, role=%s",
		current.Alias, current.RoleName, current.Role)

	if current.Alias == "" {
		t.Error("suc0 alias should not be empty - tech-lead should be resolved from predecessor rev6")
	}
	if current.Alias == "System" {
		t.Error("suc0 alias should NOT be 'System' - tech-lead role should be resolved")
	}

	// 8. 模拟 status line 构建
	sld := buildStatusLineData(fl, suc0, current, "bbca")
	t.Logf("status_line alias: %s", sld.Alias)
	if sld.Alias == "" || sld.Alias == "System" {
		t.Errorf("status_line alias should be a proper role alias, got: %q", sld.Alias)
	}

	// 9. 验证其他 dev-flow 角色也能解析
	testCases := []struct {
		nodeID string
		roleID string
		name   string
	}{
		{"tri3", "pm", "Task Triage"},
		{"fa01", "tech-lead", "Requirements Analysis"},
		{"fd02", "tech-lead", "Technical Design"},
		{"fi03", "dev", "Implementation"},
		{"ai01", "analysis", "Deep Analysis"},
		{"ver4", "qa", "Integration Verification"},
	}

	for _, tc := range testCases {
		var node *flow.FlowNode
		for i := range fl.Nodes {
			if fl.Nodes[i].ID == tc.nodeID {
				node = &fl.Nodes[i]
				break
			}
		}
		if node == nil {
			t.Logf("node %s not found in dev-flow, skipping", tc.nodeID)
			continue
		}

		// 有些节点可能没有 components，跳过
		if node.Components == nil {
			t.Logf("node %s has no components, skipping", tc.nodeID)
			continue
		}

		// DEBUG
		if len(node.Components.Roles) > 0 {
			t.Logf("  DEBUG %s: Components.Roles[0].Ref=%q", tc.nodeID, node.Components.Roles[0].Ref)
		} else {
			t.Logf("  DEBUG %s: Components.Roles is empty", tc.nodeID)
		}
		t.Logf("  DEBUG %s: node.Config=%q", tc.nodeID, string(node.Config))

		ri := resolveRoleInfo(node, fl, team, root)
		if ri.alias == "" {
			t.Errorf("node %s (%s) alias should not be empty", tc.nodeID, tc.name)
		}
		t.Logf("%s (%s): alias=%s, roleName=%s, roleID=%s",
			tc.nodeID, tc.name, ri.alias, ri.roleName, ri.roleID)
	}
}

// TestOldFormatRoleDetection 验证旧格式（内联角色）能被正确检测并报错
func TestOldFormatRoleDetection(t *testing.T) {
	// 旧格式：角色有内联 persona/alias/guidance
	oldFormatRole := flow.RoleDefinition{
		ID:       "triage",
		Persona:  "内联 persona 内容",
		Alias:    "齐活林",
		Guidance: "内联 guidance",
		Traits:   []string{"decisive"},
	}

	if isReferenceOnlyRole(oldFormatRole) {
		t.Error("old format role with inline content should NOT be detected as reference-only")
	}

	// 新格式：只有 id + prompt_source
	newFormatRole := flow.RoleDefinition{
		ID:           "triage",
		PromptSource: "assets/skill/v3/prompts/triage.md",
	}

	if !isReferenceOnlyRole(newFormatRole) {
		t.Error("new format reference-only role should be detected as such")
	}

	// 从实际 team.json 验证所有角色都是引用格式
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"
	team, err := LoadTeam(root)
	if err != nil || team == nil {
		t.Skipf("cannot load team: %v", err)
	}

	for _, role := range team.Roles {
		if !isReferenceOnlyRole(role) {
			t.Errorf("team.json role %s should be reference-only format, but has inline content", role.ID)
		}
		if role.PromptSource == "" {
			t.Errorf("team.json role %s is missing prompt_source", role.ID)
		}
	}
	t.Logf("all %d roles in team.json are v3 reference-only format ✓", len(team.Roles))
}