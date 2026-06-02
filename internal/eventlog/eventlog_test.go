package eventlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSession_WritesContextMD(t *testing.T) {
	tmpDir := t.TempDir()

	lgr, err := NewLogger(tmpDir)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	name, err := lgr.CreateSession(3, "测试主题", "测试输入")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if name == "" {
		t.Fatal("expected non-empty session name")
	}

	ctxPath := filepath.Join(lgr.SessionsDir(), name, "context.md")
	data, err := os.ReadFile(ctxPath)
	if err != nil {
		t.Fatalf("context.md not found: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Session: "+name) {
		t.Errorf("context.md missing session name, got:\n%s", content)
	}
	if !strings.Contains(content, "**Round**: R3") {
		t.Errorf("context.md missing round, got:\n%s", content)
	}
	if !strings.Contains(content, "**Topic**: 测试主题") {
		t.Errorf("context.md missing topic, got:\n%s", content)
	}
	if !strings.Contains(content, "**Status**: active") {
		t.Errorf("context.md missing status, got:\n%s", content)
	}
}

func TestWriteContext_UpdatesContextMD(t *testing.T) {
	tmpDir := t.TempDir()

	lgr, err := NewLogger(tmpDir)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	name, err := lgr.CreateSession(1, "orig", "test")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	newContent := "Updated context for session " + name
	if err := lgr.WriteContext(name, newContent); err != nil {
		t.Fatalf("WriteContext failed: %v", err)
	}

	got, err := lgr.ReadContext(name)
	if err != nil {
		t.Fatalf("ReadContext failed: %v", err)
	}
	if got != newContent {
		t.Errorf("ReadContext: got %q, want %q", got, newContent)
	}
}

func TestReadContext_MissingSession(t *testing.T) {
	tmpDir := t.TempDir()

	lgr, err := NewLogger(tmpDir)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	got, err := lgr.ReadContext("nonexistent")
	if err != nil {
		t.Fatalf("ReadContext unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("ReadContext for missing session: got %q, want empty", got)
	}
}

func TestCreateSession_EventsMDL(t *testing.T) {
	tmpDir := t.TempDir()

	lgr, err := NewLogger(tmpDir)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	name, err := lgr.CreateSession(5, "event测试", "hello world")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	_ = name // session name is captured but events go to global events.mdl

	eventsPath := filepath.Join(tmpDir, ".team", "state", "events.mdl")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("events.mdl not found: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("events.mdl is empty")
	}
	if !strings.Contains(string(data), "session.start") {
		t.Error("events.mdl missing session.start event")
	}
}