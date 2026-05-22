package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type FlowState struct {
	Flow          string     `yaml:"flow"`
	Node          string     `yaml:"node"`
	TaskID        string     `yaml:"task_id,omitempty"`
	Phase         string     `yaml:"phase,omitempty"`
	Role          string     `yaml:"role,omitempty"`
	RoleAlias     string     `yaml:"role_alias,omitempty"`
	Principal     bool       `yaml:"principal,omitempty"`
	BeadsID       string     `yaml:"beads_id,omitempty"`
	UpdatedAt     time.Time  `yaml:"updated_at"`
	Suspended     bool       `yaml:"suspended"`
	SuspendedAt   *time.Time `yaml:"suspended_at,omitempty"`
	SuspendReason string     `yaml:"suspend_reason,omitempty"`
}

func StatePath(root string) string {
	return filepath.Join(root, ".team", "state.yaml")
}

func LoadFlowState(root string) (*FlowState, error) {
	path := StatePath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}
	var s FlowState
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}
	return &s, nil
}

func SaveFlowState(root string, s *FlowState) error {
	s.UpdatedAt = time.Now()
	path := StatePath(root)
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	teamDir := filepath.Join(root, ".team")
	if err := os.MkdirAll(teamDir, 0755); err != nil {
		return fmt.Errorf("create .team dir: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func DeleteFlowState(root string) error {
	path := StatePath(root)
	return os.Remove(path)
}

func SuspendState(root string, reason string) error {
	s, err := LoadFlowState(root)
	if err != nil {
		return err
	}
	if s == nil {
		s = &FlowState{}
	}
	s.Suspended = true
	now := time.Now()
	s.SuspendedAt = &now
	s.SuspendReason = reason
	return SaveFlowState(root, s)
}

func ResumeState(root string) error {
	s, err := LoadFlowState(root)
	if err != nil {
		return err
	}
	if s == nil {
		return nil
	}
	s.Suspended = false
	s.SuspendedAt = nil
	s.SuspendReason = ""
	return SaveFlowState(root, s)
}
