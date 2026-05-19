package flow

import (
	"encoding/json"
	"fmt"
	"os"
)

func SerializeFlow(flow *Flow) ([]byte, error) {
	if flow == nil {
		return nil, fmt.Errorf("serialize flow: flow is nil")
	}
	data, err := json.MarshalIndent(flow, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serialize flow: %w", err)
	}
	return data, nil
}

func SerializeFlowFile(flow *Flow, path string) error {
	data, err := SerializeFlow(flow)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write flow file %s: %w", path, err)
	}
	return nil
}

func RoundTrip(data []byte) (*Flow, error) {
	flow, err := ParseFlow(data)
	if err != nil {
		return nil, fmt.Errorf("round-trip parse: %w", err)
	}
	serialized, err := SerializeFlow(flow)
	if err != nil {
		return nil, fmt.Errorf("round-trip serialize: %w", err)
	}
	roundTripped, err := ParseFlow(serialized)
	if err != nil {
		return nil, fmt.Errorf("round-trip re-parse: %w", err)
	}
	return roundTripped, nil
}

func FlowsEqual(a, b *Flow) bool {
	if a == nil || b == nil {
		return a == b
	}
	aData, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bData, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aData) == string(bData)
}
