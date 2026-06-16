package proc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCurlyBracesCauseYAMLError(t *testing.T) {
	root := `d:\workspace\project\golang\origadmin\framework\projects\team-flow`
	path := filepath.Join(root, "assets/skill/v3/prompts/pm.md")
	data, _ := os.ReadFile(path)
	fm, _ := extractFrontmatter(string(data))

	// Test: remove {TEAM_PATH} from standards section
	fmFixed := strings.ReplaceAll(fm, "{TEAM_PATH}", "TEAM_PATH")
	var m map[string]interface{}
	err := yaml.Unmarshal([]byte(fmFixed), &m)
	t.Logf("After removing {TEAM_PATH}: err=%v", err)

	// Test: put the standards items in quotes
	fm2 := strings.ReplaceAll(fm, `{TEAM_PATH}/workflows/shared.md`, `"{TEAM_PATH}/workflows/shared.md"`)
	fm2 = strings.ReplaceAll(fm2, `{TEAM_PATH}/workflows/shared-protocol.md`, `"{TEAM_PATH}/workflows/shared-protocol.md"`)
	fm2 = strings.ReplaceAll(fm2, `{TEAM_PATH}/workflows/roles/requirements-standards.md`, `"{TEAM_PATH}/workflows/roles/requirements-standards.md"`)
	fm2 = strings.ReplaceAll(fm2, `{TEAM_PATH}/workflows/roles/devops-standards.md`, `"{TEAM_PATH}/workflows/roles/devops-standards.md"`)
	fm2 = strings.ReplaceAll(fm2, `{TEAM_PATH}/workflows/roles/analysis-standards.md`, `"{TEAM_PATH}/workflows/roles/analysis-standards.md"`)
	fm2 = strings.ReplaceAll(fm2, `{TEAM_PATH}/templates/gherkin-feature-template.md`, `"{TEAM_PATH}/templates/gherkin-feature-template.md"`)
	err2 := yaml.Unmarshal([]byte(fm2), &m)
	t.Logf("After quoting {TEAM_PATH} paths: err=%v", err2)

	// Test concierge for comparison (has no {TEAM_PATH} and works)
	path2 := filepath.Join(root, "assets/skill/v3/prompts/concierge.md")
	data2, _ := os.ReadFile(path2)
	fm3, _ := extractFrontmatter(string(data2))
	err3 := yaml.Unmarshal([]byte(fm3), &m)
	t.Logf("concierge.md (no {TEAM_PATH}): err=%v", err3)
}