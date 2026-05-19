package proc

import "strings"

func SubstituteVarsInPath(path string, vars map[string]string) string {
	result := path
	for key, val := range vars {
		if val != "" {
			placeholder := "{" + key + "}"
			result = strings.ReplaceAll(result, placeholder, val)
		}
	}
	return result
}
