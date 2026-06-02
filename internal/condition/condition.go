package condition

import "strings"

func Eval(expr string, ctx map[string]string) bool {
	if expr == "" || expr == "null" {
		return true // no condition = always pass
	}
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true
	}

	// Handle OR
	if strings.Contains(expr, " OR ") {
		parts := strings.Split(expr, " OR ")
		for _, p := range parts {
			if Eval(strings.TrimSpace(p), ctx) {
				return true
			}
		}
		return false
	}

	// Handle NOT
	if strings.HasPrefix(expr, "!") {
		return !Eval(expr[1:], ctx)
	}

	// Handle key=value
	if idx := strings.Index(expr, "="); idx >= 0 {
		key := strings.TrimSpace(expr[:idx])
		val := strings.TrimSpace(expr[idx+1:])
		got, ok := ctx[key]
		if !ok {
			return false
		}
		return got == val
	}

	// Simple boolean: key exists and is non-empty
	got, ok := ctx[expr]
	return ok && got != "" && got != "false"
}