package proc

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/origadmin/team-flow/internal/flow"
)

func TestDebugTri3RoleResolution(t *testing.T) {
	root := "d:\\workspace\\project\\golang\\origadmin\\framework\\projects\\team-flow"

	// 直接解析 dev-flow.json 看 tri3 节点
	procPath := filepath.Join(root, "assets", "flows", "dev-flow.json")
	fl, err := flow.ParseFlowFile(procPath)
	if err != nil {
		t.Fatalf("parse flow: %v", err)
	}

	var tri3 *flow.FlowNode
	for i := range fl.Nodes {
		if fl.Nodes[i].ID == "tri3" {
			tri3 = &fl.Nodes[i]
			break
		}
	}
	if tri3 == nil {
		t.Fatal("tri3 not found")
	}

	t.Logf("tri3 type: %s", tri3.Type)
	t.Logf("tri3 config raw: %s", string(tri3.Config))

	if tri3.Components == nil {
		t.Log("tri3.Components is nil")
	} else {
		t.Logf("tri3.Components.Roles count: %d", len(tri3.Components.Roles))
		for i, r := range tri3.Components.Roles {
			t.Logf("  role[%d]: ref=%s source=%s", i, r.Ref, r.Source)
		}
	}

	var pc flow.PhaseConfig
	if err := json.Unmarshal(tri3.Config, &pc); err != nil {
		t.Logf("PhaseConfig unmarshal err: %v", err)
	} else {
		t.Logf("PhaseConfig.Role=%q", pc.Role)
	}

	// Now via LoadFlow and LoadTeam (like actual tests)
	fl2, err := LoadFlow(root, "dev-flow")
	if err != nil {
		t.Fatalf("LoadFlow: %v", err)
	}

	var tri3V2 *flow.FlowNode
	for i := range fl2.Nodes {
		if fl2.Nodes[i].ID == "tri3" {
			tri3V2 = &fl2.Nodes[i]
			break
		}
	}
	if tri3V2 == nil {
		t.Fatal("tri3 not found via LoadFlow")
	}
	t.Logf("via LoadFlow: tri3 type=%s config=%s", tri3V2.Type, string(tri3V2.Config))
	if tri3V2.Components == nil {
		t.Log("via LoadFlow: tri3.Components is nil")
	} else {
		t.Logf("via LoadFlow: tri3.Components.Roles count=%d", len(tri3V2.Components.Roles))
		for i, r := range tri3V2.Components.Roles {
			t.Logf("  role[%d]: ref=%s source=%s", i, r.Ref, r.Source)
		}
	}

	team, err := LoadTeam(root)
	if err != nil || team == nil {
		t.Fatalf("LoadTeam: err=%v team=%v", err, team)
	}
	t.Logf("team roles count: %d", len(team.Roles))
	for _, r := range team.Roles {
		t.Logf("  team role id=%s prompt=%s", r.ID, r.PromptSource)
	}

	ri := resolveRoleInfo(tri3V2, fl2, team, root)
	t.Logf("resolveRoleInfo(tri3): alias=%q roleID=%q roleName=%q", ri.alias, ri.roleID, ri.roleName)

	// Check if pm is in team roles
	found := false
	for _, r := range team.Roles {
		if r.ID == "pm" {
			found = true
			t.Logf("FOUND pm in team roles: prompt_source=%q, isRefOnly=%v",
				r.PromptSource, isReferenceOnlyRole(r))
		}
	}
	if !found {
		t.Log("pm NOT in team roles!")
	}

	// Try parsePhaseConfig directly
	fmt.Printf("DEBUG tri3.Config=%q\n", string(tri3V2.Config))
}
