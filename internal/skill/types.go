package skill

type ResolvedSkill struct {
	ID      string `yaml:"id"`
	Source  string `yaml:"source"`
	Trigger string `yaml:"trigger,omitempty"`
}

type SkillsFile struct {
	Version    int             `yaml:"version"`
	ResolvedAt string          `yaml:"resolved_at"`
	Team       string          `yaml:"team"`
	SkillTags  []string        `yaml:"skill_tags"`
	Skills     []ResolvedSkill `yaml:"skills"`
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
	"rsbuild":    {},
	"python":     {},
	"rust":       {},
}
