package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRegistry_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	reg, err := LoadRegistry(filepath.Join(tmpDir, "projects.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Projects) != 0 {
		t.Error("expected empty registry")
	}
}

func TestSaveAndLoadRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "projects.yaml")

	reg := &ProjectRegistry{
		Projects: []ProjectEntry{
			{Name: "framework", Path: "/path/to/framework", Team: "dev-team"},
			{Name: "team-flow", Path: "/path/to/team-flow", Team: "dev-team"},
		},
	}

	if err := SaveRegistry(path, reg); err != nil {
		t.Fatalf("SaveRegistry failed: %v", err)
	}

	loaded, err := LoadRegistry(path)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}

	if len(loaded.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(loaded.Projects))
	}
	if loaded.Projects[0].Name != "framework" {
		t.Errorf("name mismatch: got %s", loaded.Projects[0].Name)
	}
}

func TestRegistry_Add(t *testing.T) {
	reg := &ProjectRegistry{}

	if err := reg.Add(ProjectEntry{Name: "test", Path: "/test"}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if len(reg.Projects) != 1 {
		t.Error("expected 1 project")
	}

	if err := reg.Add(ProjectEntry{Name: "test", Path: "/test2"}); err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestRegistry_Remove(t *testing.T) {
	reg := &ProjectRegistry{
		Projects: []ProjectEntry{
			{Name: "a", Path: "/a"},
			{Name: "b", Path: "/b"},
		},
	}

	if err := reg.Remove("a"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if len(reg.Projects) != 1 {
		t.Error("expected 1 project after remove")
	}
	if reg.Projects[0].Name != "b" {
		t.Error("expected remaining project to be 'b'")
	}

	if err := reg.Remove("nonexistent"); err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestRegistry_Find(t *testing.T) {
	reg := &ProjectRegistry{
		Projects: []ProjectEntry{
			{Name: "framework", Path: "/fw"},
			{Name: "team-flow", Path: "/tf"},
		},
	}

	found := reg.Find("framework")
	if found == nil || found.Path != "/fw" {
		t.Error("expected to find framework")
	}

	notFound := reg.Find("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent")
	}
}

func TestRegistry_Update(t *testing.T) {
	reg := &ProjectRegistry{
		Projects: []ProjectEntry{
			{Name: "framework", Path: "/fw", LastNode: "tri3"},
		},
	}

	reg.Update("framework", func(e *ProjectEntry) {
		e.LastNode = "suc0"
	})

	if reg.Projects[0].LastNode != "suc0" {
		t.Errorf("expected last_node=suc0, got %s", reg.Projects[0].LastNode)
	}
}

func TestEntryFromPath(t *testing.T) {
	tmpDir := t.TempDir()
	teamDir := filepath.Join(tmpDir, ".team")
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		t.Fatal(err)
	}

	entry, err := EntryFromPath(tmpDir)
	if err != nil {
		t.Fatalf("EntryFromPath failed: %v", err)
	}
	if entry.Name != filepath.Base(tmpDir) {
		t.Errorf("name mismatch: got %s, want %s", entry.Name, filepath.Base(tmpDir))
	}
}

func TestEntryFromPath_NoTeamDir(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := EntryFromPath(tmpDir)
	if err == nil {
		t.Error("expected error for missing .team/ directory")
	}
}
