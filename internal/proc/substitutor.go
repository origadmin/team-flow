package proc

import (
	"path/filepath"
	"strings"
)

func SubstituteVarsInPath(path string, vars map[string]string) string {
	result := path
	for key, val := range vars {
		if val != "" {
			placeholder := "{" + key + "}"
			result = strings.ReplaceAll(result, placeholder, val)
		}
	}
	result = strings.ReplaceAll(result, "//", "/")
	if filepath.IsAbs(result) {
		result = filepath.Clean(result)
	}
	return result
}
