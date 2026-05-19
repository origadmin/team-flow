package flow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var supportedVariables = []string{
	"docs_path",
	"TEAM_PATH",
	"task_id",
	"SKILL_PATH",
	"PROJECT_PATH",
}

func ResolveDocSpecPaths(docs []DocSpec, vars map[string]string) []DocSpec {
	result := make([]DocSpec, len(docs))
	for i, doc := range docs {
		result[i] = doc
		result[i].Path = resolveTemplatePath(doc.Path, vars)
		if doc.Template != "" {
			result[i].Template = resolveTemplatePath(doc.Template, vars)
		}
	}
	return result
}

func resolveTemplatePath(path string, vars map[string]string) string {
	s := path
	for _, key := range supportedVariables {
		if val, ok := vars[key]; ok {
			placeholder := "{" + key + "}"
			s = strings.ReplaceAll(s, placeholder, val)
		}
	}
	return s
}

func CheckDeliverables(docs []DocSpec, basePath string) (existing []string, missing []string) {
	for _, doc := range docs {
		fullPath := doc.Path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(basePath, fullPath)
		}

		if _, err := os.Stat(fullPath); err == nil {
			existing = append(existing, fullPath)
		} else {
			missing = append(missing, fullPath)
		}
	}
	return existing, missing
}

func CollectAllDocSpecs(flow *Flow) []DocSpec {
	var docs []DocSpec
	for _, node := range flow.Nodes {
		docs = append(docs, node.Docs...)
	}
	return docs
}

func ValidateDocSpecPaths(docs []DocSpec) error {
	for _, doc := range docs {
		if doc.Name == "" {
			return fmt.Errorf("docspec name is required")
		}
		if doc.Path == "" {
			return fmt.Errorf("docspec path is required for doc %s", doc.Name)
		}
		if doc.Required != nil && *doc.Required {
			for _, key := range supportedVariables {
				placeholder := "{" + key + "}"
				if strings.Contains(doc.Path, placeholder) {
					return fmt.Errorf("required doc %s has unresolved variable %s in path", doc.Name, placeholder)
				}
			}
		}
	}
	return nil
}
