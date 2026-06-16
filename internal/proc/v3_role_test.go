package proc

import (
	"encoding/json"
	"testing"

	"github.com/origadmin/team-flow/internal/flow"
)

// TestV3RoleResolution 验证 team.json v3 引用格式的角色解析
func TestV3RoleResolution(t *testing.T) {
	// 模拟 v3 team.json：角色只有 id + prompt_source，没有内联内容
	teamJSON := `{
		"id": "test-team",
		"schema_version": "3.0",
		"roles": [
			{
				"id": "triage",
				"prompt_source": "assets/skill/v3/prompts/triage.md",
				"principal": true
			},
			{
				"id": "dev",
				"prompt_source": "assets/skill/v3/prompts/dev.md",
				"principal": false
			}
		],
		"rules": [
			{"id": "d1a", "source": "builtin"}
		]
	}`

	var team flow.TeamDefinition
	if err := json.Unmarshal([]byte(teamJSON), &team); err != nil {
		t.Fatalf("failed to parse team JSON: %v", err)
	}

	// 验证 team.json 是 v3 引用格式（没有内联内容）
	for _, role := range team.Roles {
		if role.Persona != "" {
			t.Errorf("v3 role %s should not have inline persona", role.ID)
		}
		if role.Alias != "" {
			t.Errorf("v3 role %s should not have inline alias", role.ID)
		}
		if role.Name != "" {
			t.Errorf("v3 role %s should not have inline name", role.ID)
		}
		if len(role.Traits) > 0 {
			t.Errorf("v3 role %s should not have inline traits", role.ID)
		}
		if role.PromptSource == "" {
			t.Errorf("v3 role %s must have prompt_source", role.ID)
		}
	}

	// 验证 schema_version
	if team.SchemaVersion != "3.0" {
		t.Errorf("expected schema_version 3.0, got %s", team.SchemaVersion)
	}

	// 验证 isReferenceOnlyRole 检测
	refRole := flow.RoleDefinition{
		ID:           "triage",
		PromptSource: "assets/skill/v3/prompts/triage.md",
		Principal:    nil,
	}
	if !isReferenceOnlyRole(refRole) {
		t.Error("reference-only role should be detected as such")
	}

	legacyRole := flow.RoleDefinition{
		ID:      "triage",
		Persona: "内联persona",
		Alias:   "齐活林",
	}
	if isReferenceOnlyRole(legacyRole) {
		t.Error("legacy inline role should NOT be detected as reference-only")
	}

	// 验证 loadRoleFromPromptSource 能从 frontmatter 正确加载
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"
	resolved, err := loadRoleFromPromptSource("triage", "assets/skill/v3/prompts/triage.md", root)
	if err != nil {
		t.Fatalf("loadRoleFromPromptSource failed: %v", err)
	}
	if resolved == nil {
		t.Fatal("loadRoleFromPromptSource returned nil role")
	}
	if resolved.Alias == "" {
		t.Error("resolved role should have alias from frontmatter")
	}
	if resolved.Persona == "" {
		t.Error("resolved role should have persona from frontmatter")
	}
	if len(resolved.Traits) == 0 {
		t.Error("resolved role should have traits from frontmatter")
	}
	if len(resolved.Rules) == 0 {
		t.Error("resolved role should have rules from frontmatter")
	}
	t.Logf("resolved triage: alias=%s, name=%s, persona_len=%d, traits=%v, rules=%v",
		resolved.Alias, resolved.Name, len(resolved.Persona), resolved.Traits, resolved.Rules)

	// 验证 dev 角色
	resolvedDev, err := loadRoleFromPromptSource("dev", "assets/skill/v3/prompts/dev.md", root)
	if err != nil {
		t.Fatalf("loadRoleFromPromptSource dev failed: %v", err)
	}
	if resolvedDev == nil {
		t.Fatal("loadRoleFromPromptSource dev returned nil")
	}
	if resolvedDev.Alias == "" {
		t.Error("resolved dev role should have alias")
	}
	t.Logf("resolved dev: alias=%s, name=%s", resolvedDev.Alias, resolvedDev.Name)
}

// TestV3ResolveRoleInfo 验证 resolveRoleInfo 在 v3 引用格式下的完整流程
func TestV3ResolveRoleInfo(t *testing.T) {
	// 创建一个模拟的 flow node，引用 triage 角色
	nodeConfig := json.RawMessage(`{"role":"Triage"}`)
	node := &flow.FlowNode{
		ID:     "tri3",
		Type:   flow.NodeTypePhase,
		Name:   "Task Triage",
		Config: nodeConfig,
		Components: &flow.NodeComponents{
			Roles: []flow.ComponentRef{{Ref: "triage", Source: "team"}},
		},
	}

	teamJSON := `{
		"id": "test-team",
		"schema_version": "3.0",
		"roles": [
			{
				"id": "triage",
				"prompt_source": "assets/skill/v3/prompts/triage.md",
				"principal": true
			}
		]
	}`

	var team flow.TeamDefinition
	if err := json.Unmarshal([]byte(teamJSON), &team); err != nil {
		t.Fatalf("failed to parse team JSON: %v", err)
	}

	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"
	ri := resolveRoleInfo(node, nil, &team, root)

	if ri.alias == "" {
		t.Error("resolveRoleInfo did not resolve alias from prompt_source")
	}
	if ri.persona == "" {
		t.Error("resolveRoleInfo did not resolve persona from prompt_source")
	}
	if len(ri.traits) == 0 {
		t.Error("resolveRoleInfo did not resolve traits from prompt_source")
	}
	if ri.roleID != "triage" {
		t.Errorf("expected roleID=triage, got %s", ri.roleID)
	}
	if !ri.principal {
		t.Error("principal should be true for triage")
	}

	t.Logf("resolveRoleInfo result: alias=%s, roleName=%s, principal=%v, roleID=%s, traits=%v",
		ri.alias, ri.roleName, ri.principal, ri.roleID, ri.traits)
}
