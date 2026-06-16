// Package bd is a stub that provides no-op implementations for legacy beads operations.
// All beads functionality has been replaced by local task management in .team/tasks/.
// This package exists only for backward compatibility with migration code.
package bd

import "fmt"

// IsAvailable always returns false — beads is no longer available.
func IsAvailable() bool { return false }

// FindPath returns empty string — beads executable no longer exists.
func FindPath() string { return "" }

// Run is a no-op — beads has been replaced by local task management.
func Run(args ...string) (string, error) {
	return "", fmt.Errorf("beads has been removed; use local task management: flow task")
}

// RunQuiet is a no-op — beads has been replaced by local task management.
func RunQuiet(args ...string) (string, error) {
	return "", fmt.Errorf("beads has been removed; use local task management: flow task")
}

// Install is a no-op — beads is no longer installable.
func Install() error {
	return fmt.Errorf("beads has been removed; use local task management: flow task")
}

// EnsureOnPath is a no-op — beads is no longer needed on PATH.
func EnsureOnPath() error {
	return fmt.Errorf("beads has been removed; use local task management: flow task")
}