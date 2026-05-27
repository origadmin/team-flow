package skillfs

import "embed"

//go:embed all:assets
var FS embed.FS

const (
	SkillRoot  = "assets/skill"
	OrgsRoot   = "assets/orgs"
	FlowsRoot  = "assets/flows"
	SchemaRoot = "assets/schema"
)
