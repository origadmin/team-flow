package main

import (
    "fmt"
    skillfs "github.com/origadmin/team-flow"
)

func main() {
    data, err := skillfs.FS.ReadFile("assets/orgs/dev-team/flows/dev-flow.json")
    if err != nil {
        fmt.Printf("ERR: %v\n", err)
        return
    }
    content := string(data)
    // Find tri3 section
    idx := 0
    for i := 0; i < 5; i++ {
        // search for tri3
        tri3Pos := indexOf(content, `"id": "tri3"`, idx)
        if tri3Pos < 0 {
            break
        }
        endPos := tri3Pos + 300
        if endPos > len(content) {
            endPos = len(content)
        }
        fmt.Printf("--- tri3 block %d start ---\n", i+1)
        fmt.Println(content[tri3Pos:endPos])
        fmt.Printf("--- tri3 block %d end ---\n\n", i+1)
        idx = endPos
    }
}

func indexOf(s, substr string, start int) int {
    if start >= len(s) {
        return -1
    }
    for i := start; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
    return -1
}
