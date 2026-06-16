package proc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLineEndingsCauseYAMLError(t *testing.T) {
	root := `d:\workspace\project\golang\origadmin\framework\projects\team-flow`
	path := filepath.Join(root, "assets/skill/v3/prompts/triage.md")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	// Check line endings
	crlf := strings.Count(content, "\r\n")
	lf := strings.Count(content, "\n") - crlf
	t.Logf("Line endings: \\r\\n=%d, \\n=%d, total \\n=%d", crlf, lf, strings.Count(content, "\n"))

	// Extract frontmatter
	fmOrig, err := extractFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}

	// Try original
	var m1 map[string]interface{}
	err1 := yaml.Unmarshal([]byte(fmOrig), &m1)
	t.Logf("Original frontmatter parse: err=%v", err1)

	// Try with \r\n → \n
	fmFixed := strings.ReplaceAll(fmOrig, "\r\n", "\n")
	var m2 map[string]interface{}
	err2 := yaml.Unmarshal([]byte(fmFixed), &m2)
	t.Logf("Normalized frontmatter parse: err=%v", err2)

	if err1 != nil && err2 == nil {
		t.Log("CONFIRMED: \\r\\n line endings cause YAML parse failure!")
	}
}