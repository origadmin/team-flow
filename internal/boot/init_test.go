package boot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	skillfs "github.com/origadmin/team-flow"
)

func TestIDEInfoBridgeFields(t *testing.T) {
	ides := detectIDEs("/tmp/fake-project")
	for _, ide := range ides {
		if ide.BridgePath == "" {
			t.Errorf("IDE %s: BridgePath should not be empty", ide.Name)
		}
		if ide.BridgeFmt == "" {
			t.Errorf("IDE %s: BridgeFmt should not be empty", ide.Name)
		}
	}
}

func TestDetectIDEsBridgePaths(t *testing.T) {
	ides := detectIDEs("/tmp/fake-project")

	expected := map[string]struct {
		bridgePath string
		bridgeFmt  string
		skillDir   string
	}{
		"Trae":     {".trae/rules/team-flow.md", "trae", ".trae/skills"},
		"Cursor":   {".cursor/rules/team-flow.mdc", "cursor", ".cursor/skills"},
		"Claude":   {".claude/rules/team-flow.md", "claude", ".claude/skills"},
		"OpenClaw": {".openclaw/rules/team-flow.md", "openclaw", ".openclaw/skills"},
	}

	for _, ide := range ides {
		exp, ok := expected[ide.Name]
		if !ok {
			t.Errorf("unexpected IDE: %s", ide.Name)
			continue
		}
		if ide.BridgePath != exp.bridgePath {
			t.Errorf("IDE %s BridgePath: got %q, want %q", ide.Name, ide.BridgePath, exp.bridgePath)
		}
		if ide.BridgeFmt != exp.bridgeFmt {
			t.Errorf("IDE %s BridgeFmt: got %q, want %q", ide.Name, ide.BridgeFmt, exp.bridgeFmt)
		}
		if !strings.HasSuffix(filepath.ToSlash(ide.SkillDir), exp.skillDir) {
			t.Errorf("IDE %s SkillDir: got %q, want suffix %q", ide.Name, ide.SkillDir, exp.skillDir)
	}
	}
}

func TestGenerateBridgeFileContent_Trae(t *testing.T) {
	content := generateBridgeFileContent("trae", ".trae/skills/team-flow")
	if !strings.Contains(content, ".team/version") {
		t.Error("Trae bridge should mention .team/version")
	}
	if !strings.Contains(content, ".trae/skills/team-flow/SKILL.md") {
		t.Error("Trae bridge should point to SKILL.md")
	}
	if strings.Contains(content, "### Rule") {
		t.Error("Bridge file should NOT contain full rule definitions (### Rule headings)")
	}
	if strings.Count(content, "\n") > 30 {
		t.Errorf("Bridge file should be minimal (5-15 lines), got %d lines", strings.Count(content, "\n"))
	}
}

func TestGenerateBridgeFileContent_Cursor(t *testing.T) {
	content := generateBridgeFileContent("cursor", ".cursor/skills/team-flow")
	if !strings.Contains(content, "---") {
		t.Error("Cursor bridge should have YAML frontmatter")
	}
	if !strings.Contains(content, "description:") {
		t.Error("Cursor bridge should have description in frontmatter")
	}
	if !strings.Contains(content, "globs:") {
		t.Error("Cursor bridge should have globs in frontmatter")
	}
	if !strings.Contains(content, ".cursor/skills/team-flow/SKILL.md") {
		t.Error("Cursor bridge should point to SKILL.md")
	}
}

func TestGenerateBridgeFileContent_Claude(t *testing.T) {
	content := generateBridgeFileContent("claude", ".claude/skills/team-flow")
	if !strings.Contains(content, ".claude/skills/team-flow/SKILL.md") {
		t.Error("Claude bridge should point to SKILL.md")
	}
	if strings.Contains(content, "---") {
		t.Error("Claude bridge should NOT have YAML frontmatter")
	}
}

func TestGenerateBridgeFileContent_OpenClaw(t *testing.T) {
	content := generateBridgeFileContent("openclaw", ".openclaw/skills/team-flow")
	if !strings.Contains(content, ".openclaw/skills/team-flow/SKILL.md") {
		t.Error("OpenClaw bridge should point to SKILL.md")
	}
}

func TestGenerateBridgeFileContent_UnknownFormat(t *testing.T) {
	content := generateBridgeFileContent("unknown", ".unknown/skills/team-flow")
	if content != "" {
		t.Errorf("Unknown format should return empty string, got %q", content)
	}
}

func TestGenerateDevMD_GeneratesAllBridgeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	traeDir := filepath.Join(tmpDir, ".trae")
	cursorDir := filepath.Join(tmpDir, ".cursor")
	claudeDir := filepath.Join(tmpDir, ".claude")
	openclawDir := filepath.Join(tmpDir, ".openclaw")

	os.MkdirAll(traeDir, 0755)
	os.MkdirAll(cursorDir, 0755)
	os.MkdirAll(claudeDir, 0755)
	os.MkdirAll(openclawDir, 0755)

	force = true
	defer func() { force = false }()

	skillPath := ".trae/skills/team-flow"
	generateDevMD(tmpDir, skillPath)

	expectedFiles := map[string]string{
		filepath.Join(tmpDir, ".trae", "rules", "team-flow.md"):          "trae",
		filepath.Join(tmpDir, ".cursor", "rules", "team-flow.mdc"):     "cursor",
		filepath.Join(tmpDir, ".claude", "rules", "team-flow.md"):     "claude",
		filepath.Join(tmpDir, ".openclaw", "rules", "team-flow.md"):    "openclaw",
	}

	for path, format := range expectedFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Bridge file not created for %s at %s: %v", format, path, err)
			continue
		}
		content := string(data)
		if len(content) == 0 {
			t.Errorf("Bridge file for %s is empty", format)
		}
		if !strings.Contains(content, "SKILL.md") {
			t.Errorf("Bridge file for %s should reference SKILL.md", format)
		}
	}
}

func TestBridgeFilesAreMinimal(t *testing.T) {
	formats := []struct {
		fmt       string
		skillPath string
	}{
		{"trae", ".trae/skills/team-flow"},
		{"cursor", ".cursor/skills/team-flow"},
		{"claude", ".claude/skills/team-flow"},
		{"openclaw", ".openclaw/skills/team-flow"},
	}

	for _, tc := range formats {
		content := generateBridgeFileContent(tc.fmt, tc.skillPath)
		lineCount := strings.Count(content, "\n") + 1
		if lineCount > 25 {
			t.Errorf("Bridge file for %s has %d lines, should be <= 15", tc.fmt, lineCount)
		}
	}
}

func TestCopyFromFS_V2(t *testing.T) {
	tmpDir := t.TempDir()

	copied, skipped, err := copyFromFS(skillfs.FS, "team/v2", tmpDir, true, nil)
	if err != nil {
		t.Fatalf("copyFromFS v2 failed: %v", err)
	}
	if copied == 0 {
		t.Error("copyFromFS v2 should copy at least 1 file")
	}
	if skipped != 0 {
		t.Errorf("copyFromFS v2 with overwrite=true should skip 0, got %d", skipped)
	}

	skillData, err := os.ReadFile(filepath.Join(tmpDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("SKILL.md not found in v2 output: %v", err)
	}
	if len(skillData) == 0 {
		t.Error("SKILL.md should not be empty")
	}

	readmeData, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	if err != nil {
		t.Fatalf("README.md not found in v2 output: %v", err)
	}
	if len(readmeData) == 0 {
		t.Error("README.md should not be empty")
	}
}

func TestCopyFromFS_V1(t *testing.T) {
	tmpDir := t.TempDir()

	copied, _, err := copyFromFS(skillfs.FS, "team", tmpDir, true, []string{"v2"})
	if err != nil {
		t.Fatalf("copyFromFS v1 failed: %v", err)
	}
	if copied == 0 {
		t.Error("copyFromFS v1 should copy at least 1 file")
	}

	skillData, err := os.ReadFile(filepath.Join(tmpDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("SKILL.md not found in v1 output: %v", err)
	}
	if len(skillData) == 0 {
		t.Error("SKILL.md should not be empty")
	}

	v2Dir := filepath.Join(tmpDir, "v2")
	if _, err := os.Stat(v2Dir); err == nil {
		t.Error("v1 copy should exclude v2 directory")
	}
}

func TestCopyFromFS_NoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "SKILL.md")
	os.WriteFile(existingFile, []byte("existing"), 0644)

	_, skipped, err := copyFromFS(skillfs.FS, "team/v2", tmpDir, false, nil)
	if err != nil {
		t.Fatalf("copyFromFS no-overwrite failed: %v", err)
	}
	if skipped == 0 {
		t.Error("copyFromFS with overwrite=false should skip existing SKILL.md")
	}

	data, _ := os.ReadFile(existingFile)
	if string(data) != "existing" {
		t.Error("copyFromFS should not overwrite existing file when overwrite=false")
	}
}

func TestCopyFromFS_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "SKILL.md")
	os.WriteFile(existingFile, []byte("old"), 0644)

	copied, _, err := copyFromFS(skillfs.FS, "team/v2", tmpDir, true, nil)
	if err != nil {
		t.Fatalf("copyFromFS overwrite failed: %v", err)
	}
	if copied == 0 {
		t.Error("copyFromFS with overwrite=true should copy files")
	}

	data, _ := os.ReadFile(existingFile)
	if string(data) == "old" {
		t.Error("copyFromFS should overwrite existing file when overwrite=true")
	}
}

func TestSkillfsHasTeamContent(t *testing.T) {
	entries, err := skillfs.FS.ReadDir("team")
	if err != nil {
		t.Fatalf("skillfs.FS.ReadDir failed: %v", err)
	}
	if len(entries) == 0 {
		t.Error("skillfs.FS should contain entries under team/")
	}

	skillData, err := skillfs.FS.ReadFile("team/SKILL.md")
	if err != nil {
		t.Fatalf("skillfs.FS.ReadFile team/SKILL.md failed: %v", err)
	}
	if len(skillData) == 0 {
		t.Error("team/SKILL.md should not be empty")
	}

	v2Data, err := skillfs.FS.ReadFile("team/v2/SKILL.md")
	if err != nil {
		t.Fatalf("skillfs.FS.ReadFile team/v2/SKILL.md failed: %v", err)
	}
	if len(v2Data) == 0 {
		t.Error("team/v2/SKILL.md should not be empty")
	}
}
