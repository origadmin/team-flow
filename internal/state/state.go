package state

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FlowState struct {
	Suspended     bool       `yaml:"suspended"`
	SuspendedAt   *time.Time `yaml:"suspended_at,omitempty"`
	SuspendReason string     `yaml:"suspend_reason,omitempty"`
}

func stateDir(root string) string {
	return filepath.Join(root, ".team", "state")
}

func LoadFlowState(root string) (*FlowState, error) {
	s := &FlowState{}

	suspendPath := filepath.Join(stateDir(root), "_project_suspended")
	data, err := os.ReadFile(suspendPath)
	if err == nil {
		s.Suspended = true
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 {
			s.SuspendReason = lines[0]
			if len(lines) > 1 {
				if t, err := time.Parse(time.RFC3339, strings.TrimSpace(lines[1])); err == nil {
					s.SuspendedAt = &t
				}
			}
		}
	}

	return s, nil
}

func SuspendState(root string, reason string) error {
	dir := stateDir(root)
	os.MkdirAll(dir, 0755)
	now := time.Now()
	suspendPath := filepath.Join(dir, "_project_suspended")
	return os.WriteFile(suspendPath, []byte(reason+"\n"+now.Format(time.RFC3339)+"\n"), 0644)
}

func ResumeState(root string) error {
	suspendPath := filepath.Join(stateDir(root), "_project_suspended")
	return os.Remove(suspendPath)
}
