package proc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndentationDiff(t *testing.T) {
	root := `d:\workspace\project\golang\origadmin\framework\projects\team-flow`

	for _, f := range []string{"concierge.md", "triage.md", "pm.md", "qa.md"} {
		t.Run(f, func(t *testing.T) {
			path := filepath.Join(root, "assets/skill/v3/prompts", f)
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
			// Find the region around must/forbidden
			for i, line := range lines {
				if strings.Contains(line, "forbidden:") {
					start := max(0, i-3)
					t.Logf("=== %s: forbidden at line %d ===", f, i+1)
					for j := start; j <= min(len(lines)-1, i+2); j++ {
						hasTab := strings.Contains(lines[j], "\t")
						spaces := 0
						for _, c := range lines[j] {
							if c == ' ' { spaces++ } else { break }
						}
						t.Logf("  L%02d indent=%d tab=%v |%s|", j+1, spaces, hasTab, lines[j])
					}
					break
				}
			}
		})
	}
}

func min(a, b int) int {
	if a < b { return a }
	return b
}