package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFlowState_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := LoadFlowState(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != nil {
		t.Error("expected nil state for nonexistent file")
	}
}

func TestSaveAndLoadFlowState(t *testing.T) {
	tmpDir := t.TempDir()
	now := time.Now().Truncate(time.Second)

	s := &FlowState{
		Flow:      "dev-flow",
		Node:      "fi03",
		TaskID:    "framework-03w",
		Phase:     "implement",
		Role:      "寇豆码",
		RoleAlias: "Kou",
		UpdatedAt: now,
	}

	if err := SaveFlowState(tmpDir, s); err != nil {
		t.Fatalf("SaveFlowState failed: %v", err)
	}

	path := filepath.Join(tmpDir, ".team", "state.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("state.yaml not created: %v", err)
	}

	loaded, err := LoadFlowState(tmpDir)
	if err != nil {
		t.Fatalf("LoadFlowState failed: %v", err)
	}

	if loaded.Flow != s.Flow {
		t.Errorf("flow mismatch: got %s, want %s", loaded.Flow, s.Flow)
	}
	if loaded.Node != s.Node {
		t.Errorf("node mismatch: got %s, want %s", loaded.Node, s.Node)
	}
	if loaded.TaskID != s.TaskID {
		t.Errorf("task_id mismatch: got %s, want %s", loaded.TaskID, s.TaskID)
	}
}

func TestSuspendAndResumeState(t *testing.T) {
	tmpDir := t.TempDir()

	s := &FlowState{
		Flow:  "dev-flow",
		Node:  "suc0",
		Phase: "complete",
	}
	if err := SaveFlowState(tmpDir, s); err != nil {
		t.Fatalf("SaveFlowState failed: %v", err)
	}

	if err := SuspendState(tmpDir, "switch to team-flow"); err != nil {
		t.Fatalf("SuspendState failed: %v", err)
	}

	loaded, err := LoadFlowState(tmpDir)
	if err != nil {
		t.Fatalf("LoadFlowState failed: %v", err)
	}

	if !loaded.Suspended {
		t.Error("expected suspended=true")
	}
	if loaded.SuspendReason != "switch to team-flow" {
		t.Errorf("suspend_reason mismatch: got %s", loaded.SuspendReason)
	}
	if loaded.SuspendedAt == nil {
		t.Error("expected suspended_at to be set")
	}

	if err := ResumeState(tmpDir); err != nil {
		t.Fatalf("ResumeState failed: %v", err)
	}

	loaded2, err := LoadFlowState(tmpDir)
	if err != nil {
		t.Fatalf("LoadFlowState after resume failed: %v", err)
	}

	if loaded2.Suspended {
		t.Error("expected suspended=false after resume")
	}
	if loaded2.SuspendedAt != nil {
		t.Error("expected suspended_at=nil after resume")
	}
}

func TestDeleteFlowState(t *testing.T) {
	tmpDir := t.TempDir()

	s := &FlowState{Flow: "dev-flow", Node: "tri3"}
	if err := SaveFlowState(tmpDir, s); err != nil {
		t.Fatalf("SaveFlowState failed: %v", err)
	}

	if err := DeleteFlowState(tmpDir); err != nil {
		t.Fatalf("DeleteFlowState failed: %v", err)
	}

	loaded, err := LoadFlowState(tmpDir)
	if err != nil {
		t.Fatalf("LoadFlowState after delete failed: %v", err)
	}
	if loaded != nil {
		t.Error("expected nil state after delete")
	}
}
