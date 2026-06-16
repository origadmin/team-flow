// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Same logic as extractFrontmatter
func extractFM(content string) (before, fm, after string, ok bool) {
	if !strings.HasPrefix(content, "---") {
		return "", "", "", false
	}
	// Skip "---"
	idx := 3
	// Skip newlines
	for idx < len(content) && (content[idx] == '\r' || content[idx] == '\n') {
		idx++
	}
	// Find closing ---
	end := strings.Index(content[idx:], "\n---")
	if end < 0 {
		// try without \n prefix
		end = strings.Index(content[idx:], "---")
		if end < 0 {
			return "", "", "", false
		}
	}
	return content[:idx], content[idx : idx+end], content[idx+end:], true
}

func main() {
	root := `d:\workspace\project\golang\origadmin\framework\projects\team-flow\assets\skill\v3\prompts`
	files, _ := filepath.Glob(filepath.Join(root, "*.md"))

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("ERR  %s: %v\n", filepath.Base(f), err)
			continue
		}
		content := string(data)

		before, fm, after, ok := extractFM(content)
		if !ok {
			fmt.Printf("SKIP %s: no frontmatter\n", filepath.Base(f))
			continue
		}

		lines := strings.Split(fm, "\n")
		modified := false
		for i, line := range lines {
			trimmed := strings.TrimLeft(line, " ")
			// Fix: list item value starting with ** (bold in YAML)
			if strings.HasPrefix(trimmed, "- **") && !strings.HasPrefix(trimmed, "- \"**") {
				indent := line[:len(line)-len(trimmed)]
				value := strings.TrimPrefix(trimmed, "- ")
				lines[i] = indent + `- "` + value + `"`
				modified = true
				fmt.Printf("  FIX line %d: ** quoted in %s\n", i+1, filepath.Base(f))
			}
			// Fix: list item value starting with {TEAM_PATH}
			if strings.HasPrefix(trimmed, "- {TEAM_PATH}") && !strings.HasPrefix(trimmed, "- \"{TEAM_PATH}") {
				indent := line[:len(line)-len(trimmed)]
				value := strings.TrimPrefix(trimmed, "- ")
				lines[i] = indent + `- "` + value + `"`
				modified = true
				fmt.Printf("  FIX line %d: {TEAM_PATH} quoted in %s\n", i+1, filepath.Base(f))
			}
			// Fix: partially quoted strings like `"Just do it quickly" (extra text)` 
			if strings.HasPrefix(trimmed, `- "`) && !strings.HasSuffix(strings.TrimRight(line, " "), `"`) {
				// Check if this looks like an unclosed quoted string with trailing text
				rest := strings.TrimPrefix(trimmed, `- "`)
				closeQuote := strings.Index(rest, `"`)
				if closeQuote > 0 && closeQuote < len(rest)-1 {
					// There's text after the closing quote
					indent := line[:len(line)-len(trimmed)]
					lines[i] = indent + `- '` + strings.TrimPrefix(trimmed, "- ") + `'`
					modified = true
					fmt.Printf("  FIX line %d: partial quote fixed in %s\n", i+1, filepath.Base(f))
				}
			}
		}

		if modified {
			newContent := before + strings.Join(lines, "\n") + after
			if err := os.WriteFile(f, []byte(newContent), 0644); err != nil {
				fmt.Printf("ERR  %s: write: %v\n", filepath.Base(f), err)
			} else {
				fmt.Printf("FIXED %s\n", filepath.Base(f))
			}
		} else {
			fmt.Printf("OK    %s\n", filepath.Base(f))
		}
	}
}