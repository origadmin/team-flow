package version

import "fmt"

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func Info() string {
	return fmt.Sprintf("flow version %s (build: %s, commit: %s)", Version, BuildTime, GitCommit)
}
