package proc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionState persists execution progress for a flow session.
// Stored as state.json in the session directory.
type SessionState struct {
	CurrentNodeID     string            `json:"current_node_id"`
	CurrentNode       string            `json:"current_node"` // human-readable name
	VisitedNodes      []string          `json:"visited_nodes"` // node IDs visited (ordered, most recent last)
	FlowName          string            `json:"flow_name"`
	TaskID            string            `json:"task_id"`
	LastNodeAt        string            `json:"last_node_at"` // ISO-8601 timestamp
	GateConfirmations map[string]bool   `json:"gate_confirmations,omitempty"` // gate node ID → confirmed
}

// IsGateConfirmed checks if a gate node has been confirmed by AI judgment.
func (s *SessionState) IsGateConfirmed(gateNodeID string) bool {
	if s == nil || s.GateConfirmations == nil {
		return false
	}
	return s.GateConfirmations[gateNodeID]
}

// ConfirmGate marks a gate node as confirmed by AI judgment.
func (s *SessionState) ConfirmGate(gateNodeID string) {
	if s.GateConfirmations == nil {
		s.GateConfirmations = make(map[string]bool)
	}
	s.GateConfirmations[gateNodeID] = true
}

// LoadSessionState reads state.json from the session directory.
// Returns nil if the file does not exist (first run).
func LoadSessionState(sessionsDir, sessionName string) (*SessionState, error) {
	path := filepath.Join(sessionsDir, sessionName, "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read session state: %w", err)
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse session state: %w", err)
	}
	return &state, nil
}

// SaveSessionState writes state.json to the session directory.
func SaveSessionState(sessionsDir, sessionName string, state *SessionState) error {
	if state == nil {
		return nil
	}
	state.LastNodeAt = time.Now().UTC().Format(time.RFC3339)

	path := filepath.Join(sessionsDir, sessionName, "state.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session state: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write session state: %w", err)
	}
	return nil
}
