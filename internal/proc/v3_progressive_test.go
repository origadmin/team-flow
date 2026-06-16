package proc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFindYAMLError(t *testing.T) {
	root := `d:\workspace\project\golang\origadmin\framework\projects\team-flow`

	files := []string{"pm.md", "qa.md", "devops.md", "analysis.md"}
	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			path := filepath.Join(root, "assets/skill/v3/prompts", f)
			data, err := os.ReadFile(path)
			if err != nil { t.Fatal(err) }

			content := string(data)
			fm, err := extractFrontmatter(content)
			if err != nil { t.Fatal(err) }

			lines := strings.Split(fm, "\n")

			// Try progressive parsing
			for end := 1; end <= len(lines); end++ {
				partial := strings.Join(lines[:end], "\n")
				// Also try with \r\n → \n normalization
				partial = strings.ReplaceAll(partial, "\r", "")
				var m map[string]interface{}
				err := yaml.Unmarshal([]byte(partial), &m)
				if err != nil {
					t.Logf("FAIL at line %d/%d: %v", end, len(lines), err)
					start := max(0, end-4)
					for j := start; j < min(len(lines), end+1); j++ {
						t.Logf("  L%02d: |%s|", j+1, lines[j])
					}
					break
				}
			}
		})
	}
}