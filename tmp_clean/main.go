package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := os.Args[1]
	fmt.Printf("Scanning: %s\n", root)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "-flow.json") {
			return nil
		}
		fmt.Printf("Processing: %s\n", path)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			fmt.Printf("  PARSE ERR: %v\n", err)
			return nil
		}
		comps, ok := m["components"].(map[string]interface{})
		if !ok {
			fmt.Printf("  No components\n")
			return nil
		}
		roles, ok := comps["roles"].([]interface{})
		if !ok {
			fmt.Printf("  No roles\n")
			return nil
		}
		fmt.Printf("  Found %d roles\n", len(roles))
		changed := false
		for _, r := range roles {
			role, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			if _, has := role["prompt_source"]; has {
				delete(role, "prompt_source")
				changed = true
				fmt.Printf("  Removed prompt_source from %s\n", role["id"])
			}
			if _, has := role["standards_source"]; has {
				delete(role, "standards_source")
				changed = true
				fmt.Printf("  Removed standards_source from %s\n", role["id"])
			}
		}
		if changed {
			out, _ := json.MarshalIndent(m, "", "  ")
			os.WriteFile(path, append(out, '\n'), 0644)
			fmt.Printf("Cleaned: %s\n", filepath.Base(path))
		} else {
			fmt.Printf("  No changes needed\n")
		}
		return nil
	})
}
