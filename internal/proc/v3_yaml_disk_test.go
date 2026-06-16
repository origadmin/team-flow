package proc

import (
	"os"
	"path/filepath"
	"testing"
)

// TestYAMLFrontmatterParsing 直接读磁盘文件测试 YAML frontmatter 解析
func TestYAMLFrontmatterParsing(t *testing.T) {
	root := `d:\workspace\project\golang\origadmin\framework\projects\team-flow`

	files := []string{
		"assets/skill/v3/prompts/triage.md",
		"assets/skill/v3/prompts/dev.md",
		"assets/skill/v3/prompts/tech-lead.md",
		"assets/skill/v3/prompts/pm.md",
		"assets/skill/v3/prompts/concierge.md",
		"assets/skill/v3/prompts/qa.md",
		"assets/skill/v3/prompts/devops.md",
		"assets/skill/v3/prompts/analysis.md",
	}

	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			path := filepath.Join(root, f)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read file: %v", err)
			}

			fm, err := extractFrontmatter(string(data))
			if err != nil {
				t.Fatalf("extractFrontmatter: %v", err)
			}
			if fm == "" {
				t.Fatal("frontmatter is empty")
			}

			role, err := parseRoleFrontmatter(data, "test")
			if err != nil {
				t.Fatalf("parseRoleFrontmatter: %v", err)
			}
			if role == nil {
				t.Fatal("parseRoleFrontmatter returned nil")
			}
			t.Logf("OK: id=%s name=%s alias=%s persona_len=%d traits=%v rules=%v",
				role.ID, role.Name, role.Alias, len(role.Persona), role.Traits, role.Rules)
		})
	}
}