package proc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFrontmatterMinimalRepro(t *testing.T) {
	root := `d:\workspace\project\golang\origadmin\framework\projects\team-flow`
	path := filepath.Join(root, "assets/skill/v3/prompts/triage.md")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	fm, err := extractFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(fm, "\n")
	t.Logf("Total lines: %d", len(lines))

	// Try parsing progressively to find the exact break point
	for end := 1; end <= len(lines); end++ {
		partial := strings.Join(lines[:end], "\n")
		var m map[string]interface{}
		err := yaml.Unmarshal([]byte(partial), &m)
		if err != nil {
			t.Logf("FAIL at %d lines (%d chars): %v", end, len(partial), err)
			// Show the last few lines
			start := max(0, end-3)
			t.Logf("Last 3 lines:")
			for i := start; i < end; i++ {
				t.Logf("  L%02d: |%s|", i+1, lines[i])
			}
			break
		}
	}

	// Try with just the must section, no forbidden
	testMust := `ai:
  constraints:
    must:
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Triage 分析后创建正式 Task (F001/B001/C001...)，等待用户确认后再分发
      - MUST wait for user confirmation before executing (dispatch/phase transition)
      - MUST dispatch to sub-agent via Task tool after confirmation
      - Self-check before any action: "Is this Triage duty or sub-agent duty?"
      - Triage is the bridge between user and AI - translate human input to AI-understandable format
      - Triage should NOT handle files directly - coordinate roles to do file work`
	var m2 map[string]interface{}
	err = yaml.Unmarshal([]byte(testMust), &m2)
	t.Logf("Must-only YAML: err=%v", err)

	testMustForbidden := testMust + `
    forbidden:
      - test item`
	err = yaml.Unmarshal([]byte(testMustForbidden), &m2)
	t.Logf("Must+forbidden YAML: err=%v", err)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}