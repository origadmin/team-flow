package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Flow struct {
	Nodes []Node `json:"nodes"`
}

type Node struct {
	ID           string   `json:"id"`
	Config       Config   `json:"config"`
	Docs         []Doc    `json:"docs"`
	PromptSource string   `json:"prompt_source"`
}

type Config struct {
	PromptSource string `json:"prompt_source"`
	Docs         []Doc  `json:"docs"`
}

type Doc struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: flow-lint <path-to-json>")
		os.Exit(1)
	}

	jsonPath := os.Args[1]
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var flow Flow
	if err := json.Unmarshal(data, &flow); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	projectRoot, _ := filepath.Abs(".")
	errors := 0

	checkPath := func(nodeID, label, path string) {
		if path == "" {
			return
		}
		// Replace variables for linting
		cleanPath := strings.ReplaceAll(path, "{TEAM_PATH}", ".team")
		cleanPath = strings.ReplaceAll(cleanPath, "{DOCS_INTERNAL}", "_docs")
		cleanPath = strings.ReplaceAll(cleanPath, "{DOCS_EXTERNAL}", "docs")
		
		// If it's a template or variable path, we might not be able to check easily
		if strings.Contains(cleanPath, "{") {
			return
		}

		fullPath := filepath.Join(projectRoot, "projects/team-flow", cleanPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Printf("FAIL [%s]: %s path not found: %s\n", nodeID, label, cleanPath)
			errors++
		}
	}

	for _, node := range flow.Nodes {
		checkPath(node.ID, "prompt_source", node.PromptSource)
		checkPath(node.ID, "config.prompt_source", node.Config.PromptSource)
		for _, doc := range node.Docs {
			checkPath(node.ID, fmt.Sprintf("doc[%s]", doc.Name), doc.Path)
		}
		for _, doc := range node.Config.Docs {
			checkPath(node.ID, fmt.Sprintf("config.doc[%s]", doc.Name), doc.Path)
		}
	}

	if errors > 0 {
		fmt.Printf("\nLint failed with %d error(s).\n", errors)
		os.Exit(1)
	}
	fmt.Println("✅ All static assets are reachable.")
}
