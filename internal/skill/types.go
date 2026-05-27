package skill

import "time"

const (
	SkillSourceTrae   = "trae"
	SkillSourceLocal  = "local"
	SkillSourceTeam   = "team"
	SkillSourceGlobal = "global"

	SkillsCacheFile = "skills-resolved.yaml"
	DefaultCacheTTL = 1 * time.Hour

	SkillsDirName = "skills"
	SkillFileName = "SKILL.md"
)

type ResolvedSkill struct {
	ID      string `yaml:"id" json:"id"`
	Source  string `yaml:"source" json:"source"`
	Trigger string `yaml:"trigger,omitempty" json:"trigger,omitempty"`
	Path    string `yaml:"path,omitempty" json:"path,omitempty"`
	Enabled bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

type SkillRequirement struct {
	Tag         string `yaml:"tag" json:"tag"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool   `yaml:"required" json:"required"`
}

type CacheMetadata struct {
	ProjectSkillsMtime string `yaml:"project_skills_mtime,omitempty"`
	TeamConfigMtime    string `yaml:"team_config_mtime,omitempty"`
	ProjectConfigMtime string `yaml:"project_config_mtime,omitempty"`
	TeamFlowRootMtime  string `yaml:"teamflow_root_mtime,omitempty"`
}

type SkillsFile struct {
	Version        int              `yaml:"version"`
	ResolvedAt     string           `yaml:"resolved_at"`
	Team           string           `yaml:"team"`
	Project        string           `yaml:"project,omitempty"`
	Role           string           `yaml:"role,omitempty"`
	SkillTags      []string         `yaml:"skill_tags"`
	CacheMetadata  CacheMetadata    `yaml:"cache_metadata,omitempty"`
	Skills         []ResolvedSkill  `yaml:"skills"`
}

type ResolveOptions struct {
	ForceRefresh bool              `json:"force_refresh"`
	RoleID       string            `json:"role_id"`
	ExtraTags    []string          `json:"extra_tags"`
	Disabled     []string          `json:"disabled"`
	Overrides    map[string]string `json:"overrides"`
}

type SkillContext struct {
	ProjectRoot    string
	TeamID         string
	RoleID         string
	ResolvedSkills []ResolvedSkill
	CacheFile      string
}

var DefaultSkillTagMap = map[string][]ResolvedSkill{
	"go":         {{ID: "go-best-practices", Source: "trae", Trigger: ".go files, go.mod"}},
	"react":      {{ID: "shadcn-ui", Source: "trae", Trigger: "shadcn components"}, {ID: "tanstack-query", Source: "trae", Trigger: "data fetching"}},
	"shadcn":     {{ID: "shadcn-ui", Source: "trae", Trigger: "shadcn components"}, {ID: "tailwind-v4-shadcn", Source: "trae", Trigger: "tailwind + shadcn setup"}},
	"bun":        {{ID: "tailwind-v4-shadcn", Source: "trae", Trigger: "bun + tailwind"}},
	"typescript": {{ID: "typescript-advanced-types", Source: "trae", Trigger: "complex type logic"}},
	"vue":        {{ID: "tanstack-query", Source: "trae", Trigger: "data fetching"}},
	"next":       {{ID: "shadcn-ui", Source: "trae", Trigger: "next + shadcn"}},
	"tailwind":   {{ID: "tailwind-v4-shadcn", Source: "trae", Trigger: "tailwind CSS"}},
}
