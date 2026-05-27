package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type GateRecord struct {
	NodeID    string    `yaml:"node_id"`
	Passed    bool      `yaml:"passed"`
	CheckedAt time.Time `yaml:"checked_at"`
	Summary   string    `yaml:"summary,omitempty"`
}

type FlowState struct {
	Flow          string        `yaml:"flow"`
	Node          string        `yaml:"node"`
	TaskID        string        `yaml:"task_id,omitempty"`
	Phase         string        `yaml:"phase,omitempty"`
	Role          string        `yaml:"role,omitempty"`
	RoleAlias     string        `yaml:"role_alias,omitempty"`
	Principal     bool          `yaml:"principal,omitempty"`
	UpdatedAt     time.Time     `yaml:"updated_at"`
	Suspended     bool          `yaml:"suspended"`
	SuspendedAt   *time.Time    `yaml:"suspended_at,omitempty"`
	SuspendReason string        `yaml:"suspend_reason,omitempty"`
	GateHistory   []GateRecord  `yaml:"gate_history,omitempty"`
	VisitedNodes  []string      `yaml:"visited_nodes,omitempty"`
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

func RecordGateResult(root string, flowName string, nodeID string, passed bool, summary string) error {
	s, err := LoadFlowState(root)
	if err != nil {
		return err
	}
	if s == nil {
		s = &FlowState{}
	}

	if s.Flow != "" && s.Flow != flowName {
		s.GateHistory = nil
		s.VisitedNodes = nil
	}

	s.Flow = flowName

	for i, g := range s.GateHistory {
		if g.NodeID == nodeID {
			s.GateHistory[i].Passed = passed
			s.GateHistory[i].CheckedAt = time.Now()
			s.GateHistory[i].Summary = summary
			return SaveFlowState(root, s)
		}
	}

	s.GateHistory = append(s.GateHistory, GateRecord{
		NodeID:    nodeID,
		Passed:    passed,
		CheckedAt: time.Now(),
		Summary:   summary,
	})
	return SaveFlowState(root, s)
}

func IsGatePassed(root string, flowName string, nodeID string) bool {
	s, err := LoadFlowState(root)
	if err != nil || s == nil {
		return false
	}
	if s.Flow != flowName {
		return false
	}
	for _, g := range s.GateHistory {
		if g.NodeID == nodeID && g.Passed {
			return true
		}
	}
	return false
}

func RecordVisitedNode(root string, flowName string, nodeID string) error {
	s, err := LoadFlowState(root)
	if err != nil {
		return err
	}
	if s == nil {
		s = &FlowState{}
	}

	if s.Flow != "" && s.Flow != flowName {
		s.GateHistory = nil
		s.VisitedNodes = nil
	}

	s.Flow = flowName

	for _, n := range s.VisitedNodes {
		if n == nodeID {
			return nil
		}
	}

	s.VisitedNodes = append(s.VisitedNodes, nodeID)
	return SaveFlowState(root, s)
}

func HasVisitedNode(root string, flowName string, nodeID string) bool {
	s, err := LoadFlowState(root)
	if err != nil || s == nil {
		return false
	}
	if s.Flow != flowName {
		return false
	}
	for _, n := range s.VisitedNodes {
		if n == nodeID {
			return true
		}
	}
	return false
}

func ResetFlowState(root string) error {
	path := StatePath(root)
	return os.Remove(path)
}
