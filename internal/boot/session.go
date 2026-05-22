package boot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/origadmin/team-flow/internal/proc"
)

type SessionLog struct {
	FlowID    string
	NodeID    string
	TaskID    string
	Actions   []string
	Decisions []string
	Lessons   []string
	Files     []string
}

func SaveSessionLog(root string, log *SessionLog) error {
	docsPath := proc.ResolveInternalDocs(root)
	sessionsDir := filepath.Join(docsPath, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return fmt.Errorf("create sessions dir: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s-%s.md", date, log.FlowID, log.NodeID)
	if log.TaskID != "" {
		filename = fmt.Sprintf("%s-%s-%s-%s.md", date, log.FlowID, log.NodeID, log.TaskID)
	}
	filePath := filepath.Join(sessionsDir, filename)

	var content string
	content += fmt.Sprintf("# Session: %s %s %s\n\n", date, log.FlowID, log.NodeID)
	content += "## Context\n\n"
	content += fmt.Sprintf("- Flow: %s\n", log.FlowID)
	content += fmt.Sprintf("- Node: %s\n", log.NodeID)
	if log.TaskID != "" {
		content += fmt.Sprintf("- Task: %s\n", log.TaskID)
	}
	content += fmt.Sprintf("- Date: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	if len(log.Actions) > 0 {
		content += "## Actions\n\n"
		for i, a := range log.Actions {
			content += fmt.Sprintf("%d. %s\n", i+1, a)
		}
		content += "\n"
	}

	if len(log.Decisions) > 0 {
		content += "## Decisions\n\n"
		for _, d := range log.Decisions {
			content += fmt.Sprintf("- %s\n", d)
		}
		content += "\n"
	}

	if len(log.Lessons) > 0 {
		content += "## Lessons\n\n"
		for _, l := range log.Lessons {
			content += fmt.Sprintf("- %s\n", l)
		}
		content += "\n"
	}

	if len(log.Files) > 0 {
		content += "## Files Changed\n\n"
		for _, f := range log.Files {
			content += fmt.Sprintf("- %s\n", f)
		}
		content += "\n"
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write session log: %w", err)
	}

	return nil
}
