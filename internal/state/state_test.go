package state

import (
	"testing"
)

func TestLoadFlowState_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := LoadFlowState(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Error("expected non-nil state (empty state is valid)")
	}
	if s.Suspended {
		t.Error("expected suspended=false for nonexistent state")
	}
}

func TestSuspendAndResumeState(t *testing.T) {
	tmpDir := t.TempDir()

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
