package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProjectConfig struct {
	Name        string        `yaml:"name"`
	Version     string        `yaml:"version"`
	DefaultFlow string        `yaml:"default_flow"`
	Paths       ProjectPaths  `yaml:"paths"`
	Toolchain   ProjectToolchain `yaml:"toolchain"`
	Flows       []ProjectFlow `yaml:"flows"`
}

type ProjectPaths struct {
	DocsInternal string `yaml:"docs_internal"`
	DocsExternal string `yaml:"docs_external"`
}

type ProjectToolchain struct {
	Backend  ProjectTool `yaml:"backend"`
	Frontend ProjectTool `yaml:"frontend"`
}

type ProjectTool struct {
	Language       string `yaml:"language"`
	Pipeline       string `yaml:"pipeline"`
	Framework      string `yaml:"framework,omitempty"`
	UILibrary      string `yaml:"ui_library,omitempty"`
	CSSFramework   string `yaml:"css_framework,omitempty"`
	PackageManager string `yaml:"package_manager,omitempty"`
}

type ProjectFlow struct {
	ID     string `yaml:"id"`
	Source string `yaml:"source"`
}

func LoadProjectConfig(root string) (*ProjectConfig, error) {
	yamlPath := filepath.Join(root, ".team", "project.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		mdPath := filepath.Join(root, ".team", "project.md")
		if mdData, mdErr := os.ReadFile(mdPath); mdErr == nil {
			return parseProjectMD(mdData)
		}
		return nil, fmt.Errorf("no project config found: %w", err)
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse project.yaml: %w", err)
	}
	return &cfg, nil
}

func SaveProjectConfig(root string, cfg *ProjectConfig) error {
	yamlPath := filepath.Join(root, ".team", "project.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal project.yaml: %w", err)
	}
	return os.WriteFile(yamlPath, data, 0644)
}

func ResolveInternalDocs(root string) string {
	cfg, err := LoadProjectConfig(root)
	if err != nil {
		return filepath.Join(root, ".team", "docs")
	}
	if cfg.Paths.DocsInternal != "" {
		p := cfg.Paths.DocsInternal
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		return p
	}
	return filepath.Join(root, ".team", "docs")
}

func ResolveExternalDocs(root string) string {
	cfg, err := LoadProjectConfig(root)
	if err != nil {
		return ""
	}
	if cfg.Paths.DocsExternal != "" {
		p := cfg.Paths.DocsExternal
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		return p
	}
	return ""
}

func parseProjectMD(data []byte) (*ProjectConfig, error) {
	cfg := &ProjectConfig{}
	content := string(data)
	lines := splitLines(content)

	for _, line := range lines {
		trimmed := trimSpace(line)
		if contains(trimmed, "docs_internal:") {
			cfg.Paths.DocsInternal = extractValue(trimmed)
		} else if contains(trimmed, "docs_external:") {
			cfg.Paths.DocsExternal = extractValue(trimmed)
		} else if contains(trimmed, "default_flow:") {
			cfg.DefaultFlow = extractValue(trimmed)
		}
	}

	return cfg, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func extractValue(line string) string {
	idx := -1
	for i := 0; i < len(line); i++ {
		if line[i] == ':' {
			idx = i
			break
		}
	}
	if idx < 0 || idx+1 >= len(line) {
		return ""
	}
	val := trimSpace(line[idx+1:])
	if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
		val = val[1 : len(val)-1]
	}
	return val
}
