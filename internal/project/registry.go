package project

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type ProjectRegistry struct {
	Projects []ProjectEntry `yaml:"projects"`
}

type ProjectEntry struct {
	Name        string    `yaml:"name"`
	Path        string    `yaml:"path"`
	Team        string    `yaml:"team,omitempty"`
	ActiveFlow  string    `yaml:"active_flow,omitempty"`
	LastNode    string    `yaml:"last_node,omitempty"`
	LastTask    string    `yaml:"last_task,omitempty"`
	LastActive  time.Time `yaml:"last_active,omitempty"`
}

func DefaultRegistryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".flow", "projects.yaml")
}

func LoadRegistry(path string) (*ProjectRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProjectRegistry{}, nil
		}
		return nil, fmt.Errorf("read registry: %w", err)
	}
	var reg ProjectRegistry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}
	return &reg, nil
}

func SaveRegistry(path string, reg *ProjectRegistry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}
	data, err := yaml.Marshal(reg)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (r *ProjectRegistry) Add(entry ProjectEntry) error {
	for _, p := range r.Projects {
		if p.Name == entry.Name {
			return fmt.Errorf("project '%s' already registered at '%s'", entry.Name, p.Path)
		}
	}
	r.Projects = append(r.Projects, entry)
	return nil
}

func (r *ProjectRegistry) Remove(name string) error {
	for i, p := range r.Projects {
		if p.Name == name {
			r.Projects = append(r.Projects[:i], r.Projects[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("project '%s' not found", name)
}

func (r *ProjectRegistry) Find(name string) *ProjectEntry {
	for i := range r.Projects {
		if r.Projects[i].Name == name {
			return &r.Projects[i]
		}
	}
	return nil
}

func (r *ProjectRegistry) Update(name string, updateFn func(*ProjectEntry)) {
	for i := range r.Projects {
		if r.Projects[i].Name == name {
			updateFn(&r.Projects[i])
			return
		}
	}
}

func EntryFromPath(projectPath string) (*ProjectEntry, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	teamDir := filepath.Join(absPath, ".team")
	if _, err := os.Stat(teamDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no .team/ directory at '%s'. Run 'flow init --v3' first", absPath)
	}

	entry := &ProjectEntry{
		Name: filepath.Base(absPath),
		Path: absPath,
	}

	versionData, err := os.ReadFile(filepath.Join(teamDir, "version"))
	if err == nil {
		_ = string(versionData)
	}

	return entry, nil
}
