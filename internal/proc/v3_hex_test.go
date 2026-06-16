package proc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontmatterDeepDump(t *testing.T) {
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
	t.Logf("Frontmatter: %d lines", len(lines))
	for i, line := range lines {
		if i >= 18 && i <= 24 {
			// Show bytes as hex for each character
			hexParts := []string{}
			for _, r := range line {
				if r < 128 {
					hexParts = append(hexParts, string(r))
				} else {
					hexParts = append(hexParts, "?")
				}
			}
			t.Logf("L%02d (len=%d): |%s|", i+1, len(line), line)
			t.Logf("      bytes: [%s]", strings.Join(hexParts, " "))
		}
	}
	// Dump lines 22-24 hex
	for i, line := range lines {
		if i >= 21 && i <= 23 {
			hexStr := ""
			for _, b := range []byte(line) {
				hexStr += string("0123456789abcdef"[b>>4])
				hexStr += string("0123456789abcdef"[b&0xf])
				hexStr += " "
			}
			t.Logf("L%02d hex: %s", i+1, hexStr)
		}
	}
}