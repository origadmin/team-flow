package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSkills(t *testing.T) {
	tags := []string{"go", "react", "shadcn"}
	skills := ResolveSkills(tags)

	if len(skills) == 0 {
		t.Fatal("expected non-empty skills")
	}

	ids := make(map[string]bool)
	for _, s := range skills {
		if ids[s.ID] {
			t.Errorf("duplicate skill ID: %s", s.ID)
		}
		ids[s.ID] = true
	}

	if !ids["go-best-practices"] {
		t.Error("expected go-best-practices from 'go' tag")
	}
	if !ids["shadcn-ui"] {
		t.Error("expected shadcn-ui from 'react' or 'shadcn' tag")
	}
}

func TestResolveSkills_EmptyTags(t *testing.T) {
	skills := ResolveSkills(nil)
	if len(skills) != 0 {
		t.Errorf("expected empty skills for empty tags, got %d", len(skills))
	}
}

func TestResolveSkills_UnknownTag(t *testing.T) {
	skills := ResolveSkills([]string{"unknown-tag"})
	if len(skills) != 0 {
		t.Errorf("expected empty skills for unknown tag, got %d", len(skills))
	}
}

func TestDeduplicateSkills(t *testing.T) {
	skills := []ResolvedSkill{
		{ID: "a", Source: "trae"},
		{ID: "b", Source: "trae"},
		{ID: "a", Source: "trae"},
	}
	deduped := DeduplicateSkills(skills)
	if len(deduped) != 2 {
		t.Errorf("expected 2 deduplicated skills, got %d", len(deduped))
	}
}

func TestNewSkillsFile(t *testing.T) {
	tags := []string{"go", "react"}
	skills := []ResolvedSkill{
		{ID: "go-best-practices", Source: "trae", Trigger: ".go files"},
	}
	sf := NewSkillsFile("dev-team", tags, skills)

	if sf.Version != 1 {
		t.Errorf("expected version 1, got %d", sf.Version)
	}
	if sf.Team != "dev-team" {
		t.Errorf("expected team dev-team, got %s", sf.Team)
	}
	if len(sf.SkillTags) != 2 {
		t.Errorf("expected 2 skill tags, got %d", len(sf.SkillTags))
	}
	if len(sf.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(sf.Skills))
	}
	if sf.ResolvedAt == "" {
		t.Error("expected non-empty resolved_at")
	}
}

func TestSaveAndLoadSkillsFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "skills.yaml")

	tags := []string{"go"}
	skills := []ResolvedSkill{
		{ID: "go-best-practices", Source: "trae", Trigger: ".go files"},
	}
	sf := NewSkillsFile("dev-team", tags, skills)

	if err := SaveSkillsFile(path, sf); err != nil {
		t.Fatalf("SaveSkillsFile failed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("skills.yaml not created: %v", err)
	}

	loaded, err := LoadSkillsFile(path)
	if err != nil {
		t.Fatalf("LoadSkillsFile failed: %v", err)
	}

	if loaded.Version != sf.Version {
		t.Errorf("version mismatch: got %d, want %d", loaded.Version, sf.Version)
	}
	if loaded.Team != sf.Team {
		t.Errorf("team mismatch: got %s, want %s", loaded.Team, sf.Team)
	}
	if len(loaded.Skills) != len(sf.Skills) {
		t.Errorf("skills count mismatch: got %d, want %d", len(loaded.Skills), len(sf.Skills))
	}
	if loaded.Skills[0].ID != sf.Skills[0].ID {
		t.Errorf("skill ID mismatch: got %s, want %s", loaded.Skills[0].ID, sf.Skills[0].ID)
	}
}

func TestLoadSkillsFile_NotFound(t *testing.T) {
	_, err := LoadSkillsFile("/nonexistent/skills.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
