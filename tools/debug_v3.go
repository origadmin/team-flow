// +build ignore

package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func test(name, yamlStr string) {
	var parsed map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlStr), &parsed)
	if err != nil {
		fmt.Printf("FAIL %s: %v\n", name, err)
	} else {
		fmt.Printf("OK   %s\n", name)
	}
}

func main() {
	test("basic must+forbidden", `ai:
  constraints:
    must:
      - item one
    forbidden:
      - bad thing
`)
	
	test("with {TEAM_PATH}", `ai:
  constraints:
    must:
      - Adhere to Protocol in {TEAM_PATH}/shared.md
    forbidden:
      - bad thing
`)
	
	test("full triage context", `ai:
  id: triage
  constraints:
    must:
      - Adhere to the Team Execution Protocol in {TEAM_PATH}/workflows/shared.md
      - Triage analyzes and creates formal Task, waits for user confirmation
    forbidden:
      - bad thing
`)

	test("with colon in value", `ai:
  constraints:
    must:
      - Self-check before any action: "Is this Triage duty or sub-agent duty?"
    forbidden:
      - bad thing
`)

	test("with asterisks quoted", `ai:
  constraints:
    must:
      - test
    forbidden:
      - "\052\052 manual edit \052\052"
`)
	
	test("actual problematic value", `ai:
  constraints:
    must:
      - Self-check before any action: "Is this Triage duty or sub-agent duty?"
      - Triage is the bridge between user and AI - translate human input to AI-understandable format
    forbidden:
      - "**manual edit task-pool.md** (v2: task-pool.md read-only)"
`)
}