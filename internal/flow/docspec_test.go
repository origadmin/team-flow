package flow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDocSpecPaths(t *testing.T) {
	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "{docs_path}/requirements/{task_id}/SPEC.md"},
		{Name: "AC.md", Format: "markdown", Path: "{TEAM_PATH}/docs/{task_id}/AC.md", Template: "{SKILL_PATH}/templates/ac.md"},
	}

	vars := map[string]string{
		"docs_path":  "/project/docs",
		"TEAM_PATH":  "/project/.team",
		"task_id":    "framework-34p",
		"SKILL_PATH": "/skills/team-flow",
	}

	resolved := ResolveDocSpecPaths(docs, vars)

	if resolved[0].Path != "/project/docs/requirements/framework-34p/SPEC.md" {
		t.Errorf("expected resolved path /project/docs/requirements/framework-34p/SPEC.md, got %s", resolved[0].Path)
	}
	if resolved[1].Path != "/project/.team/docs/framework-34p/AC.md" {
		t.Errorf("expected resolved path /project/.team/docs/framework-34p/AC.md, got %s", resolved[1].Path)
	}
	if resolved[1].Template != "/skills/team-flow/templates/ac.md" {
		t.Errorf("expected resolved template /skills/team-flow/templates/ac.md, got %s", resolved[1].Template)
	}
}

func TestResolveDocSpecPaths_PartialVars(t *testing.T) {
	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "{docs_path}/requirements/{task_id}/SPEC.md"},
	}

	vars := map[string]string{
		"docs_path": "/project/docs",
	}

	resolved := ResolveDocSpecPaths(docs, vars)

	expected := "/project/docs/requirements/{task_id}/SPEC.md"
	if resolved[0].Path != expected {
		t.Errorf("expected partially resolved path %s, got %s", expected, resolved[0].Path)
	}
}

func TestResolveDocSpecPaths_NoVars(t *testing.T) {
	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "/static/path/SPEC.md"},
	}

	resolved := ResolveDocSpecPaths(docs, nil)

	if resolved[0].Path != "/static/path/SPEC.md" {
		t.Errorf("expected unchanged path, got %s", resolved[0].Path)
	}
}

func TestResolveDocSpecPaths_AllVariables(t *testing.T) {
	docs := []DocSpec{
		{Name: "test.md", Format: "markdown", Path: "{docs_path}/{TEAM_PATH}/{task_id}/{SKILL_PATH}/{PROJECT_PATH}/test.md"},
	}

	vars := map[string]string{
		"docs_path":    "docs",
		"TEAM_PATH":    "team",
		"task_id":      "123",
		"SKILL_PATH":   "skills",
		"PROJECT_PATH": "project",
	}

	resolved := ResolveDocSpecPaths(docs, vars)

	expected := "docs/team/123/skills/project/test.md"
	if resolved[0].Path != expected {
		t.Errorf("expected %s, got %s", expected, resolved[0].Path)
	}
}

func TestCheckDeliverables_ExistingAndMissing(t *testing.T) {
	dir := t.TempDir()

	existingFile := filepath.Join(dir, "SPEC.md")
	if err := os.WriteFile(existingFile, []byte("# Spec"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "SPEC.md"},
		{Name: "AC.md", Format: "markdown", Path: "AC.md"},
	}

	existing, missing := CheckDeliverables(docs, dir)

	if len(existing) != 1 {
		t.Errorf("expected 1 existing file, got %d", len(existing))
	}
	if len(missing) != 1 {
		t.Errorf("expected 1 missing file, got %d", len(missing))
	}

	if len(existing) > 0 && !filepath.IsAbs(existing[0]) {
		t.Errorf("expected absolute path, got %s", existing[0])
	}
	if len(missing) > 0 && !filepath.IsAbs(missing[0]) {
		t.Errorf("expected absolute path, got %s", missing[0])
	}
}

func TestCheckDeliverables_AllExisting(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"SPEC.md", "AC.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("# "+name), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "SPEC.md"},
		{Name: "AC.md", Format: "markdown", Path: "AC.md"},
	}

	existing, missing := CheckDeliverables(docs, dir)

	if len(existing) != 2 {
		t.Errorf("expected 2 existing files, got %d", len(existing))
	}
	if len(missing) != 0 {
		t.Errorf("expected 0 missing files, got %d", len(missing))
	}
}

func TestCheckDeliverables_AllMissing(t *testing.T) {
	dir := t.TempDir()

	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "SPEC.md"},
		{Name: "AC.md", Format: "markdown", Path: "AC.md"},
	}

	existing, missing := CheckDeliverables(docs, dir)

	if len(existing) != 0 {
		t.Errorf("expected 0 existing files, got %d", len(existing))
	}
	if len(missing) != 2 {
		t.Errorf("expected 2 missing files, got %d", len(missing))
	}
}

func TestCheckDeliverables_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	absPath := filepath.Join(dir, "SPEC.md")
	if err := os.WriteFile(absPath, []byte("# Spec"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: absPath},
	}

	existing, missing := CheckDeliverables(docs, dir)

	if len(existing) != 1 {
		t.Errorf("expected 1 existing file, got %d", len(existing))
	}
	if len(missing) != 0 {
		t.Errorf("expected 0 missing files, got %d", len(missing))
	}
}

func TestCollectAllDocSpecs(t *testing.T) {
	required := true
	flow := &Flow{
		Version:  "v3",
		Metadata: FlowMetadata{Name: "test"},
		Nodes: []FlowNode{
			{
				ID:   "triage",
				Type: NodeTypePhase,
				Name: "Triage",
				Docs: []DocSpec{
					{Name: "SPEC.md", Format: "markdown", Path: "/docs/SPEC.md", Required: &required},
				},
			},
			{
				ID:   "dev",
				Type: NodeTypePhase,
				Name: "Dev",
				Docs: []DocSpec{
					{Name: "IMPL.md", Format: "markdown", Path: "/docs/IMPL.md", Required: &required},
					{Name: "TEST.md", Format: "markdown", Path: "/docs/TEST.md", Required: &required},
				},
			},
			{ID: "done", Type: NodeTypeTerminal, Name: "Done"},
		},
		Edges: []FlowEdge{},
	}

	docs := CollectAllDocSpecs(flow)
	if len(docs) != 3 {
		t.Errorf("expected 3 doc specs, got %d", len(docs))
	}
}

func TestValidateDocSpecPaths(t *testing.T) {
	required := true
	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "/docs/SPEC.md", Required: &required},
	}

	if err := ValidateDocSpecPaths(docs); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateDocSpecPaths_EmptyName(t *testing.T) {
	docs := []DocSpec{
		{Name: "", Format: "markdown", Path: "/docs/SPEC.md"},
	}

	if err := ValidateDocSpecPaths(docs); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestValidateDocSpecPaths_EmptyPath(t *testing.T) {
	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: ""},
	}

	if err := ValidateDocSpecPaths(docs); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestValidateDocSpecPaths_UnresolvedVariable(t *testing.T) {
	required := true
	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "{docs_path}/SPEC.md", Required: &required},
	}

	if err := ValidateDocSpecPaths(docs); err == nil {
		t.Fatal("expected error for unresolved variable in required doc path")
	}
}

func TestValidateDocSpecPaths_UnresolvedVariableNotRequired(t *testing.T) {
	notRequired := false
	docs := []DocSpec{
		{Name: "SPEC.md", Format: "markdown", Path: "{docs_path}/SPEC.md", Required: &notRequired},
	}

	if err := ValidateDocSpecPaths(docs); err != nil {
		t.Errorf("expected no error for non-required doc with unresolved variable, got %v", err)
	}
}
