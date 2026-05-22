package proc

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/flow"
)

type FlowResolver interface {
	Resolve(ctx context.Context, name string, projectRoot string) (*flow.Flow, error)
}

type NodeResolver interface {
	Resolve(ctx context.Context, f *flow.Flow, nodeID string) (*flow.FlowNode, error)
}

type VarSubstitutor interface {
	CollectVars(ctx context.Context, f *flow.Flow, req ProcRunRequest, resolvedNodeID string) map[string]string
}

type DefaultFlowResolver struct {
	Root string
}

func (r *DefaultFlowResolver) Resolve(ctx context.Context, name string, projectRoot string) (*flow.Flow, error) {
	root := projectRoot
	if root == "" {
		root = r.Root
	}

	flowName := name
	if flowName == "" {
		var err error
		flowName, err = ResolveDefaultFlowName(root)
		if err != nil {
			return nil, err
		}
	}

	return LoadFlow(root, flowName)
}

type DefaultNodeResolver struct{}

func (r *DefaultNodeResolver) Resolve(ctx context.Context, f *flow.Flow, nodeID string) (*flow.FlowNode, error) {
	return FindNode(f, nodeID)
}

type DefaultVarSubstitutor struct{}

func (s *DefaultVarSubstitutor) CollectVars(ctx context.Context, f *flow.Flow, req ProcRunRequest, resolvedNodeID string) map[string]string {
	vars := make(map[string]string)

	if req.ProjectRoot != "" {
		vars["TEAM_PATH"] = filepath.Join(req.ProjectRoot, ".team")
		vars["workspace"] = req.ProjectRoot
		vars["SKILL_PATH"] = filepath.Join(req.ProjectRoot, ".trae", "skills", "team-flow")
	}

	docsPath := ResolveDocsPath(req.ProjectRoot)
	if docsPath != "" {
		vars["DOCS_INTERNAL"] = docsPath
	}

	docsExternalPath := ResolveDocsExternalPath(req.ProjectRoot)
	if docsExternalPath != "" {
		vars["DOCS_EXTERNAL"] = docsExternalPath
	}

	if req.TaskID != "" {
		vars["task_id"] = req.TaskID
	}

	if f != nil {
		domain := ""
		if f.Config != nil {
			domain = f.Config.Domain
			if domain == "" {
				domain = string(f.Config.TaskType)
			}
		}
		vars["domain"] = domain
		vars["flow_id"] = f.Metadata.Name
	}

	if resolvedNodeID != "" {
		vars["node_id"] = resolvedNodeID
	}

	return vars
}

func extractFlowNameFromProjectMD(data []byte) string {
	lines := strings.Split(string(data), "\n")
	inFlowSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			inFlowSection = false
			continue
		}
		if strings.Contains(trimmed, "flow:") || strings.Contains(trimmed, "default_flow:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				val := strings.Trim(strings.TrimSpace(parts[1]), "\"' ")
				if val != "" {
					return val
				}
			}
			inFlowSection = true
		}
		if inFlowSection && strings.HasPrefix(trimmed, "- ") {
			val := strings.Trim(strings.TrimPrefix(trimmed, "- "), "\"' ")
			if val != "" {
				return val
			}
		}
	}
	return ""
}
