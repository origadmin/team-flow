package flow

type VariableSource int

const (
	SourceFlowContext VariableSource = iota
	SourceFlowVariables
	SourceProjectConfig
	SourceCLIFlags
	SourceEnvironment
)

type ResolvedVariable struct {
	Name   string
	Value  string
	Source VariableSource
}

func ResolveVariables(sources ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, src := range sources {
		for k, v := range src {
			if v != "" {
				result[k] = v
			}
		}
	}
	return result
}
