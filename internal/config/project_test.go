package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProjectConfig_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	teamDir := filepath.Join(tmpDir, ".team")
	os.MkdirAll(teamDir, 0755)

	yamlContent := `name: test-project
version: v3
default_flow: dev-flow

paths:
  docs_internal: _docs/test-project/
  docs_external: docs/

toolchain:
  backend:
    language: go
    pipeline: go test ./... | go build -o bin/app
  frontend:
    language: typescript
    pipeline: bun run test | bun run build
`
	yamlPath := filepath.Join(teamDir, "project.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	cfg, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadProjectConfig failed: %v", err)
	}
	if cfg.Name != "test-project" {
		t.Errorf("Name = %q, want %q", cfg.Name, "test-project")
	}
	if cfg.DefaultFlow != "dev-flow" {
		t.Errorf("DefaultFlow = %q, want %q", cfg.DefaultFlow, "dev-flow")
	}
	if cfg.Paths.DocsInternal != "_docs/test-project/" {
		t.Errorf("DocsInternal = %q, want %q", cfg.Paths.DocsInternal, "_docs/test-project/")
	}
	if cfg.Paths.DocsExternal != "docs/" {
		t.Errorf("DocsExternal = %q, want %q", cfg.Paths.DocsExternal, "docs/")
	}
	if cfg.Toolchain.Backend.Language != "go" {
		t.Errorf("Backend.Language = %q, want %q", cfg.Toolchain.Backend.Language, "go")
	}
	if cfg.Toolchain.Frontend.Language != "typescript" {
		t.Errorf("Frontend.Language = %q, want %q", cfg.Toolchain.Frontend.Language, "typescript")
	}
}

func TestLoadProjectConfig_MDFallback(t *testing.T) {
	tmpDir := t.TempDir()
	teamDir := filepath.Join(tmpDir, ".team")
	os.MkdirAll(teamDir, 0755)

	mdContent := `## Project Configuration

- **Project Name**: legacy-project

## Paths

docs_internal: _docs/legacy/
docs_external: docs/
default_flow: bugfix-flow
`
	mdPath := filepath.Join(teamDir, "project.md")
	os.WriteFile(mdPath, []byte(mdContent), 0644)

	cfg, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadProjectConfig failed: %v", err)
	}
	if cfg.DefaultFlow != "bugfix-flow" {
		t.Errorf("DefaultFlow = %q, want %q", cfg.DefaultFlow, "bugfix-flow")
	}
	if cfg.Paths.DocsInternal != "_docs/legacy/" {
		t.Errorf("DocsInternal = %q, want %q", cfg.Paths.DocsInternal, "_docs/legacy/")
	}
}

func TestLoadProjectConfig_YAMLPriority(t *testing.T) {
	tmpDir := t.TempDir()
	teamDir := filepath.Join(tmpDir, ".team")
	os.MkdirAll(teamDir, 0755)

	yamlContent := `name: yaml-project
version: v3
default_flow: yaml-flow
`
	os.WriteFile(filepath.Join(teamDir, "project.yaml"), []byte(yamlContent), 0644)

	mdContent := `default_flow: md-flow
`
	os.WriteFile(filepath.Join(teamDir, "project.md"), []byte(mdContent), 0644)

	cfg, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadProjectConfig failed: %v", err)
	}
	if cfg.DefaultFlow != "yaml-flow" {
		t.Errorf("YAML should take priority, got %q", cfg.DefaultFlow)
	}
}

func TestSaveProjectConfig(t *testing.T) {
	tmpDir := t.TempDir()
	teamDir := filepath.Join(tmpDir, ".team")
	os.MkdirAll(teamDir, 0755)

	cfg := &ProjectConfig{
		Name:        "save-test",
		Version:     "v3",
		DefaultFlow: "dev-flow",
		Paths: ProjectPaths{
			DocsInternal: "_docs/save-test/",
			DocsExternal: "docs/",
		},
		Flows: []ProjectFlow{
			{ID: "dev-flow", Source: "preset"},
		},
	}

	if err := SaveProjectConfig(tmpDir, cfg); err != nil {
		t.Fatalf("SaveProjectConfig failed: %v", err)
	}

	loaded, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadProjectConfig after save failed: %v", err)
	}
	if loaded.Name != "save-test" {
		t.Errorf("Name = %q, want %q", loaded.Name, "save-test")
	}
	if loaded.DefaultFlow != "dev-flow" {
		t.Errorf("DefaultFlow = %q, want %q", loaded.DefaultFlow, "dev-flow")
	}
	if len(loaded.Flows) != 1 || loaded.Flows[0].ID != "dev-flow" {
		t.Errorf("Flows not preserved correctly")
	}
}

func TestResolveInternalDocs_Configured(t *testing.T) {
	tmpDir := t.TempDir()
	teamDir := filepath.Join(tmpDir, ".team")
	os.MkdirAll(teamDir, 0755)

	yamlContent := `name: test
version: v3
default_flow: dev-flow
paths:
  docs_internal: _docs/test/
`
	os.WriteFile(filepath.Join(teamDir, "project.yaml"), []byte(yamlContent), 0644)

	result := ResolveInternalDocs(tmpDir)
	expected := filepath.Join(tmpDir, "_docs", "test")
	if result != expected {
		t.Errorf("ResolveInternalDocs = %q, want %q", result, expected)
	}
}

func TestResolveInternalDocs_Fallback(t *testing.T) {
	tmpDir := t.TempDir()
	teamDir := filepath.Join(tmpDir, ".team")
	os.MkdirAll(teamDir, 0755)

	yamlContent := `name: test
version: v3
default_flow: dev-flow
`
	os.WriteFile(filepath.Join(teamDir, "project.yaml"), []byte(yamlContent), 0644)

	result := ResolveInternalDocs(tmpDir)
	expected := filepath.Join(tmpDir, ".team", "docs")
	if result != expected {
		t.Errorf("ResolveInternalDocs fallback = %q, want %q", result, expected)
	}
}
