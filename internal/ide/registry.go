package ide

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	skillfs "github.com/origadmin/team-flow"
	"gopkg.in/yaml.v3"
)

type IDEEntry struct {
	Name          string `yaml:"name"`
	Subdir        string `yaml:"subdir"`
	BridgePath    string `yaml:"bridge_path"`
	BridgeFormat  string `yaml:"bridge_format"`
	AlwaysInstall bool   `yaml:"always_install"`
	NpxAgent      string `yaml:"npx_agent"`
	SkillDir      string `yaml:"skill_dir"`
}

type IDERegistry struct {
	IDEs []IDEEntry `yaml:"ides"`
}

type IDEInfo struct {
	Name          string
	Subdir        string
	SkillDir      string
	ConfigDir     string
	BridgePath    string
	BridgeFmt     string
	AlwaysInstall bool
	NpxAgent      string
	Detected      bool
}

var cachedRegistry *IDERegistry

func LoadRegistry(fsys embed.FS) (*IDERegistry, error) {
	if cachedRegistry != nil {
		return cachedRegistry, nil
	}

	data, err := fsys.ReadFile(skillfs.SkillRoot + "/ide-registry.yaml")
	if err != nil {
		return nil, fmt.Errorf("read ide-registry.yaml: %w", err)
	}

	var registry IDERegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("parse ide-registry.yaml: %w", err)
	}

	cachedRegistry = &registry
	return &registry, nil
}

func DetectIDEs(projectPath string, fsys embed.FS) []IDEInfo {
	registry, err := LoadRegistry(fsys)
	if err != nil {
		fmt.Printf("  ⚠ Failed to load IDE registry: %v\n", err)
		fmt.Println("  Using built-in defaults")
		return defaultIDEs(projectPath)
	}

	return detectFromRegistry(projectPath, registry)
}

func DetectIDEsWithParentSearch(projectPath string, fsys embed.FS) []IDEInfo {
	registry, err := LoadRegistry(fsys)
	if err != nil {
		fmt.Printf("  ⚠ Failed to load IDE registry: %v\n", err)
		fmt.Println("  Using built-in defaults")
		return defaultIDEs(projectPath)
	}

	searchPaths := []string{projectPath}
	for dir := filepath.Dir(projectPath); dir != "" && dir != "." && dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		searchPaths = append(searchPaths, dir)
	}

	var ides []IDEInfo
	for _, entry := range registry.IDEs {
		for _, base := range searchPaths {
			configDir := filepath.Join(base, entry.Subdir)
			if _, err := os.Stat(configDir); err == nil {
				ides = append(ides, IDEInfo{
					Name:          entry.Name,
					Subdir:        entry.Subdir,
					SkillDir:      filepath.Join(base, entry.Subdir, "skills"),
					ConfigDir:     configDir,
					BridgePath:    entry.BridgePath,
					BridgeFmt:     entry.BridgeFormat,
					AlwaysInstall: entry.AlwaysInstall,
					NpxAgent:      entry.NpxAgent,
					Detected:      true,
				})
				break
			}
		}
	}

	if len(ides) == 0 {
		return defaultIDEs(projectPath)
	}

	return ides
}

func detectFromRegistry(projectPath string, registry *IDERegistry) []IDEInfo {
	var ides []IDEInfo
	for _, entry := range registry.IDEs {
		configDir := filepath.Join(projectPath, entry.Subdir)
		if _, err := os.Stat(configDir); err == nil {
			ides = append(ides, IDEInfo{
				Name:          entry.Name,
				Subdir:        entry.Subdir,
				SkillDir:      filepath.Join(projectPath, entry.Subdir, "skills"),
				ConfigDir:     configDir,
				BridgePath:    entry.BridgePath,
				BridgeFmt:     entry.BridgeFormat,
				AlwaysInstall: entry.AlwaysInstall,
				NpxAgent:      entry.NpxAgent,
				Detected:      true,
			})
		}
	}

	if len(ides) == 0 {
		return defaultIDEs(projectPath)
	}

	return ides
}

func defaultIDEs(projectPath string) []IDEInfo {
	return []IDEInfo{
		{Name: "Trae", Subdir: ".trae", SkillDir: filepath.Join(projectPath, ".trae", "skills"), ConfigDir: filepath.Join(projectPath, ".trae"), BridgePath: ".trae/rules/team-flow.md", BridgeFmt: "trae", AlwaysInstall: true, NpxAgent: "trae", Detected: true},
		{Name: "Cursor", Subdir: ".cursor", SkillDir: filepath.Join(projectPath, ".cursor", "skills"), ConfigDir: filepath.Join(projectPath, ".cursor"), BridgePath: ".cursor/rules/team-flow.mdc", BridgeFmt: "cursor", AlwaysInstall: false, NpxAgent: "cursor", Detected: true},
		{Name: "Claude", Subdir: ".claude", SkillDir: filepath.Join(projectPath, ".claude", "skills"), ConfigDir: filepath.Join(projectPath, ".claude"), BridgePath: ".claude/rules/team-flow.md", BridgeFmt: "claude", AlwaysInstall: true, NpxAgent: "claude", Detected: true},
		{Name: "OpenClaw", Subdir: ".openclaw", SkillDir: filepath.Join(projectPath, ".openclaw", "skills"), ConfigDir: filepath.Join(projectPath, ".openclaw"), BridgePath: ".openclaw/rules/team-flow.md", BridgeFmt: "openclaw", AlwaysInstall: true, NpxAgent: "openclaw", Detected: true},
		{Name: "Gemini", Subdir: ".gemini", SkillDir: filepath.Join(projectPath, ".gemini", "skills"), ConfigDir: filepath.Join(projectPath, ".gemini"), BridgePath: ".gemini/rules/team-flow.md", BridgeFmt: "gemini", AlwaysInstall: false, NpxAgent: "gemini", Detected: true},
	}
}
