package skillfs

import "embed"

//go:embed all:team
//go:embed all:teams
//go:embed all:v3/flows
//go:embed all:v3/schema
var FS embed.FS
