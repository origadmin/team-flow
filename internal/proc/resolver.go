package proc

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/origadmin/team-flow/internal/eventlog"
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

type ActiveFlowResolver struct {
	Root string
}

func (r *ActiveFlowResolver) Resolve(ctx context.Context, name string, projectRoot string) (*flow.Flow, error) {
	root := projectRoot
	if root == "" {
		root = r.Root
	}

	flowName := name
	// Strip builtin: namespace prefix (v1#1: builtin namespace resolution)
	flowName = strings.TrimPrefix(flowName, "builtin:")
	if flowName == "" {
		flowName = ResolveActiveFlow(root)
	}

	fl, err := LoadFlow(root, flowName)
	if err != nil {
		return nil, err
	}

	// v3.2 architecture enforcement: sub_flows cannot be invoked directly
	if fl.Metadata.Type == "sub" {
		return nil, fmt.Errorf("flow %q is a sub_flow (parent=%q) and cannot be invoked directly — invoke via main flow dispatch", flowName, fl.Metadata.ParentFlow)
	}

	return fl, nil
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
	} else if req.ProjectRoot != "" {
		lgr, lgrErr := eventlog.NewLogger(req.ProjectRoot)
		if lgrErr == nil {
			activeTasks, _ := lgr.ActiveTasks()
			if len(activeTasks) == 1 {
				vars["task_id"] = activeTasks[0]
			}
		}
	}

	if f != nil {
		domain := ""
		taskType := ""
		if f.Config != nil {
			domain = f.Config.Domain
			if domain == "" {
				domain = string(f.Config.TaskType)
			}
			taskType = string(f.Config.TaskType)
		}
		vars["domain"] = domain
		vars["flow_id"] = f.Metadata.Name
		vars["DOC_CATEGORY"] = MapDocCategory(taskType)
	}

	if resolvedNodeID != "" {
		vars["node_id"] = resolvedNodeID
	}

	return vars
}

func MapDocCategory(taskType string) string {
	switch taskType {
	case string(flow.TaskTypeFeature), string(flow.TaskTypeDocs):
		return "requirements"
	case string(flow.TaskTypeBug), string(flow.TaskTypeHotfix):
		return "reports/bugs"
	case string(flow.TaskTypeChange):
		return "reports/changes"
	case string(flow.TaskTypeAnalysis):
		return "analysis"
	case string(flow.TaskTypeRelease):
		return "releases"
	case string(flow.TaskTypeBatch):
		return "batch"
	default:
		return "requirements"
	}
}
