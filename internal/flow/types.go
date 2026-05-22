package flow

import "encoding/json"

type NodeType string

const (
	NodeTypePhase    NodeType = "phase"
	NodeTypeStart    NodeType = "start"
	NodeTypeGate     NodeType = "gate"
	NodeTypeBranch   NodeType = "branch"
	NodeTypeParallel NodeType = "parallel"
	NodeTypeSubflow  NodeType = "subflow"
	NodeTypeLoop     NodeType = "loop"
	NodeTypeManual   NodeType = "manual"
	NodeTypeEvent    NodeType = "event"
	NodeTypeTerminal NodeType = "terminal"
)

type EdgeType string

const (
	EdgeTypeSequential  EdgeType = "sequential"
	EdgeTypeConditional EdgeType = "conditional"
)

type TaskType string

const (
	TaskTypeFeature  TaskType = "feature"
	TaskTypeBug      TaskType = "bug"
	TaskTypeChange   TaskType = "change"
	TaskTypeAnalysis TaskType = "analysis"
	TaskTypeDocs     TaskType = "docs"
	TaskTypeHotfix   TaskType = "hotfix"
	TaskTypeRelease  TaskType = "release"
	TaskTypeBatch    TaskType = "batch"
)

type ComponentSource string

const (
	SourceBuiltin    ComponentSource = "builtin"
	SourceCustom     ComponentSource = "custom"
	SourceMarketplace ComponentSource = "marketplace"
	SourceTrae       ComponentSource = "trae"
	SourceFile       ComponentSource = "file"
	SourceFramework  ComponentSource = "framework"
	SourceTeam       ComponentSource = "team"
)

type ConstraintType string

const (
	ConstraintFileWrite       ConstraintType = "file_write"
	ConstraintToolRestriction ConstraintType = "tool_restriction"
	ConstraintTimeout         ConstraintType = "timeout"
	ConstraintMaxSubAgents    ConstraintType = "max_sub_agents"
	ConstraintCustom          ConstraintType = "custom"
)

type GateConfigType string

const (
	GateTypeDeliverableCheck GateConfigType = "deliverable_check"
	GateTypeTestsPass        GateConfigType = "tests_pass"
	GateTypeLintPass         GateConfigType = "lint_pass"
	GateTypeNoRegressions    GateConfigType = "no_regressions"
	GateTypeQualityCheck     GateConfigType = "quality_check"
	GateTypeCustom           GateConfigType = "custom"
)

type GateConditionType string

const (
	GateCondTestsPass           GateConditionType = "tests_pass"
	GateCondLintPass            GateConditionType = "lint_pass"
	GateCondDeliverablesComplete GateConditionType = "deliverables_complete"
	GateCondNoRegressions       GateConditionType = "no_regressions"
	GateCondTaskExists          GateConditionType = "task_exists"
	GateCondTypeMatches         GateConditionType = "type_matches"
	GateCondCustom              GateConditionType = "custom"
)

type ActionType string

const (
	ActionUpdateTaskPhase ActionType = "update_task_phase"
	ActionNotify          ActionType = "notify"
	ActionLog             ActionType = "log"
	ActionSetVariable     ActionType = "set_variable"
	ActionInvokeTool      ActionType = "invoke_tool"
	ActionCustom          ActionType = "custom"
)

type ErrorStrategy string

const (
	ErrorStrategyRetry    ErrorStrategy = "retry"
	ErrorStrategySkip     ErrorStrategy = "skip"
	ErrorStrategyFail     ErrorStrategy = "fail"
	ErrorStrategyFallback ErrorStrategy = "fallback"
)

type TerminalStatus string

const (
	TerminalSuccess   TerminalStatus = "success"
	TerminalFailed    TerminalStatus = "failed"
	TerminalRejected  TerminalStatus = "rejected"
	TerminalCancelled TerminalStatus = "cancelled"
	TerminalTimeout   TerminalStatus = "timeout"
)

type ParallelStrategy string

const (
	ParallelAllSuccess     ParallelStrategy = "all_success"
	ParallelAnySuccess     ParallelStrategy = "any_success"
	ParallelFirstSuccess   ParallelStrategy = "first_success"
	ParallelMajoritySuccess ParallelStrategy = "majority_success"
)

type MergeStrategy string

const (
	MergeWaitAll     MergeStrategy = "wait_all"
	MergeWaitAny     MergeStrategy = "wait_any"
	MergeWaitMajority MergeStrategy = "wait_majority"
)

type BackoffStrategy string

const (
	BackoffNone        BackoffStrategy = "none"
	BackoffLinear      BackoffStrategy = "linear"
	BackoffExponential BackoffStrategy = "exponential"
)

type ApprovalStrategy string

const (
	ApprovalAll      ApprovalStrategy = "all"
	ApprovalAny      ApprovalStrategy = "any"
	ApprovalMajority ApprovalStrategy = "majority"
)

type EventTriggerType string

const (
	TriggerWebhook    EventTriggerType = "webhook"
	TriggerFileChange EventTriggerType = "file_change"
	TriggerSchedule   EventTriggerType = "schedule"
	TriggerTaskStatus EventTriggerType = "task_status"
	TriggerManual     EventTriggerType = "manual"
	TriggerCustom     EventTriggerType = "custom"
)

type ToolType string

const (
	ToolTypeCLI     ToolType = "cli"
	ToolTypeBuiltin ToolType = "builtin"
	ToolTypeAPI     ToolType = "api"
	ToolTypeMCP     ToolType = "mcp"
)

type RuleType string

const (
	RuleBehavioralConstraint RuleType = "behavioral_constraint"
	RuleOutputConstraint     RuleType = "output_constraint"
	RuleProcessConstraint    RuleType = "process_constraint"
	RuleQualityConstraint   RuleType = "quality_constraint"
)

type Enforcement string

const (
	EnforcementHard Enforcement = "hard"
	EnforcementSoft Enforcement = "soft"
)

type GateDefinitionType string

const (
	GateDefDeliverableCheck GateDefinitionType = "deliverable_check"
	GateDefTestCoverage     GateDefinitionType = "test_coverage"
	GateDefLintCheck        GateDefinitionType = "lint_check"
	GateDefQualityCheck     GateDefinitionType = "quality_check"
	GateDefCustom           GateDefinitionType = "custom"
)

type ToolScope string

const (
	ToolScopeCodebase ToolScope = "codebase"
	ToolScopeDocs     ToolScope = "docs"
	ToolScopeConfig   ToolScope = "config"
	ToolScopeAll      ToolScope = "all"
)

type Flow struct {
	Version    string             `json:"version"`
	Metadata   FlowMetadata       `json:"metadata"`
	Config     *FlowConfig        `json:"config,omitempty"`
	Extends    string             `json:"extends,omitempty"`
	Overrides []FlowNodeOverride  `json:"overrides,omitempty"`
	Components *ComponentRegistry `json:"components,omitempty"`
	Nodes     []FlowNode          `json:"nodes"`
	Edges     []FlowEdge          `json:"edges"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type FlowMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

type FlowConfig struct {
	TaskType       TaskType `json:"task_type,omitempty"`
	Domain         string   `json:"domain,omitempty"`
	Extends        string   `json:"extends,omitempty"`
	AutoDispatch   *bool    `json:"auto_dispatch,omitempty"`
	ParallelLimit  *int     `json:"parallel_limit,omitempty"`
	TimeoutMinutes *int     `json:"timeout_minutes,omitempty"`
}

type FlowNode struct {
	ID          string          `json:"id"`
	Type        NodeType        `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Components  *NodeComponents `json:"components,omitempty"`
	Docs        []DocSpec       `json:"docs,omitempty"`
	Gates       []GateConfig    `json:"gates,omitempty"`
	OnEnter     []Action        `json:"on_enter,omitempty"`
	OnExit      []Action        `json:"on_exit,omitempty"`
	OnError     *ErrorHandler   `json:"on_error,omitempty"`
}

type FlowEdge struct {
	ID         string           `json:"id,omitempty"`
	From       string           `json:"from"`
	To         string           `json:"to"`
	Type       EdgeType         `json:"type,omitempty"`
	Conditions []EdgeCondition  `json:"conditions,omitempty"`
}

type EdgeCondition struct {
	Expression string `json:"expression"`
}

type FlowNodeOverride struct {
	ID          string          `json:"id"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Components  *NodeComponents `json:"components,omitempty"`
	Docs        []DocSpec       `json:"docs,omitempty"`
	Gates       []GateConfig    `json:"gates,omitempty"`
	OnEnter     []Action        `json:"on_enter,omitempty"`
	OnExit      []Action        `json:"on_exit,omitempty"`
	OnError     *ErrorHandler   `json:"on_error,omitempty"`
}

type PhaseConfig struct {
	Role          string   `json:"role,omitempty"`
	AutoDispatch  *bool    `json:"auto_dispatch,omitempty"`
	TimeoutMinutes *int    `json:"timeout_minutes,omitempty"`
	Deliverables  []string `json:"deliverables,omitempty"`
}

type GateNodeConfig struct {
	Conditions []GateCondition `json:"conditions"`
	OnPass     string          `json:"on_pass,omitempty"`
	OnFail     string          `json:"on_fail,omitempty"`
	AutoRetry  *AutoRetryConfig `json:"auto_retry,omitempty"`
}

type GateCondition struct {
	Type         GateConditionType `json:"type"`
	Threshold    string            `json:"threshold,omitempty"`
	Required     *bool             `json:"required,omitempty"`
	Check        string            `json:"check,omitempty"`
	Expected     string            `json:"expected,omitempty"`
	Deliverables []string          `json:"deliverables,omitempty"`
}

type BranchConfig struct {
	Conditions []BranchCondition `json:"conditions"`
}

type BranchCondition struct {
	When    string `json:"when,omitempty"`
	Goto    string `json:"goto,omitempty"`
	Default string `json:"default,omitempty"`
	Target  string `json:"target,omitempty"`
}

type ParallelConfig struct {
	MaxConcurrency *int              `json:"max_concurrency,omitempty"`
	Strategy       ParallelStrategy  `json:"strategy,omitempty"`
	Branches       []ParallelBranch  `json:"branches"`
	MergeStrategy  MergeStrategy     `json:"merge_strategy,omitempty"`
}

type ParallelBranch struct {
	Node   string `json:"node"`
	Weight *int   `json:"weight,omitempty"`
}

type SubflowConfig struct {
	FlowRef       string            `json:"flow_ref"`
	InputMapping  map[string]string `json:"input_mapping,omitempty"`
	OutputMapping map[string]string `json:"output_mapping,omitempty"`
}

type LoopConfig struct {
	MaxAttempts   *int            `json:"max_attempts,omitempty"`
	Delay         *int            `json:"delay,omitempty"`
	DelayMinutes  *int            `json:"delay_minutes,omitempty"`
	Backoff       BackoffStrategy `json:"backoff,omitempty"`
	ExitCondition string          `json:"exit_condition,omitempty"`
}

type ManualConfig struct {
	Approvers         []Approver        `json:"approvers"`
	ApprovalStrategy  ApprovalStrategy  `json:"approval_strategy,omitempty"`
	TimeoutMinutes    *int              `json:"timeout_minutes,omitempty"`
	OnTimeout         string            `json:"on_timeout,omitempty"`
}

type Approver struct {
	Role     string `json:"role"`
	User     string `json:"user,omitempty"`
	Required *bool  `json:"required,omitempty"`
}

type EventConfig struct {
	Trigger       EventTriggerConfig `json:"trigger"`
	TimeoutMinutes *int              `json:"timeout_minutes,omitempty"`
	OnTimeout     string             `json:"on_timeout,omitempty"`
}

type EventTriggerConfig struct {
	Type    EventTriggerType    `json:"type"`
	Pattern string              `json:"pattern,omitempty"`
	Filter  map[string]interface{} `json:"filter,omitempty"`
}

type TerminalConfig struct {
	Status  TerminalStatus `json:"status"`
	Message string         `json:"message,omitempty"`
}

type AutoRetryConfig struct {
	MaxAttempts  *int `json:"max_attempts,omitempty"`
	DelayMinutes *int `json:"delay_minutes,omitempty"`
}

type DocSpec struct {
	Name         string   `json:"name"`
	Format       string   `json:"format"`
	Path         string   `json:"path"`
	Template     string   `json:"template,omitempty"`
	Required     *bool    `json:"required"`
	Description  string   `json:"description,omitempty"`
	ContentRules []string `json:"content_rules,omitempty"`
}

type NodeComponents struct {
	Roles       []ComponentRef `json:"roles,omitempty"`
	Rules       []ComponentRef `json:"rules,omitempty"`
	Tools       []ToolRef      `json:"tools,omitempty"`
	Skills      []ComponentRef `json:"skills,omitempty"`
	Prompts     []ComponentRef `json:"prompts,omitempty"`
	Constraints []Constraint   `json:"constraints,omitempty"`
}

type ComponentRef struct {
	Ref     string                 `json:"ref"`
	Source  ComponentSource        `json:"source"`
	Version string                 `json:"version,omitempty"`
	Path    string                 `json:"path,omitempty"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

type ToolRef struct {
	Ref      string                 `json:"ref"`
	Source   ComponentSource        `json:"source,omitempty"`
	Version  string                 `json:"version,omitempty"`
	Path     string                 `json:"path,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Commands []string               `json:"commands,omitempty"`
	Scope    ToolScope              `json:"scope,omitempty"`
}

type Constraint struct {
	Type          ConstraintType `json:"type"`
	AllowedPaths  []string       `json:"allowed_paths,omitempty"`
	Forbidden     []string       `json:"forbidden,omitempty"`
	Allowed       []string       `json:"allowed,omitempty"`
	Minutes       *int           `json:"minutes,omitempty"`
	Limit         *int           `json:"limit,omitempty"`
}

type GateConfig struct {
	Type     GateConfigType `json:"type"`
	OnFail   string         `json:"on_fail,omitempty"`
	Required []string       `json:"required,omitempty"`
	Threshold string        `json:"threshold,omitempty"`
	Check    string         `json:"check,omitempty"`
}

type Action struct {
	Action  ActionType            `json:"action"`
	Phase   string                `json:"phase,omitempty"`
	Message string                `json:"message,omitempty"`
	Channel string                `json:"channel,omitempty"`
	Name    string                `json:"name,omitempty"`
	Value   interface{}           `json:"value,omitempty"`
	Tool    string                `json:"tool,omitempty"`
	Command string                `json:"command,omitempty"`
	Args    map[string]interface{} `json:"args,omitempty"`
}

type ErrorHandler struct {
	Strategy          ErrorStrategy `json:"strategy,omitempty"`
	MaxRetries        *int          `json:"max_retries,omitempty"`
	RetryDelaySeconds *int          `json:"retry_delay_seconds,omitempty"`
	FallbackNode      string        `json:"fallback_node,omitempty"`
	OnError           []Action      `json:"on_error,omitempty"`
}

type ComponentRegistry struct {
	Roles  []RoleDefinition  `json:"roles,omitempty"`
	Rules  []RuleDefinition  `json:"rules,omitempty"`
	Tools  []ToolDefinition  `json:"tools,omitempty"`
	Skills []SkillDefinition `json:"skills,omitempty"`
	Gates  []GateDefDefinition `json:"gates,omitempty"`
}

type RoleDefinition struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Alias            string   `json:"alias,omitempty"`
	AliasEn          string   `json:"alias_en,omitempty"`
	Principal        *bool    `json:"principal,omitempty"`
	Persona          string   `json:"persona,omitempty"`
	Description      string   `json:"description,omitempty"`
	Traits           []string `json:"traits,omitempty"`
	Guidance         string   `json:"guidance,omitempty"`
	PromptSource     string   `json:"prompt_source,omitempty"`
	StandardsSource  string   `json:"standards_source,omitempty"`
	Capabilities     []string `json:"capabilities,omitempty"`
	PromptDirectives []string `json:"prompt_directives,omitempty"`
	Rules            []string `json:"rules,omitempty"`
}

type RuleDefinition struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Instruction string      `json:"instruction,omitempty"`
	Description string      `json:"description,omitempty"`
	Source      string      `json:"source,omitempty"`
	Type        RuleType    `json:"type,omitempty"`
	Enforcement Enforcement `json:"enforcement,omitempty"`
}

type ToolDefinition struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     ToolType `json:"type,omitempty"`
	Commands []string `json:"commands,omitempty"`
	Scope    interface{} `json:"scope,omitempty"`
}

type SkillDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Source      string `json:"source,omitempty"`
	Trigger     string `json:"trigger,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
}

type GateDefDefinition struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Source string             `json:"source,omitempty"`
	Type   GateDefinitionType `json:"type,omitempty"`
}

type FlowOverrides struct {
	Config     *FlowConfig        `json:"config,omitempty"`
	Nodes      []FlowNodeOverride `json:"nodes,omitempty"`
	Edges      []FlowEdge         `json:"edges,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	Components *ComponentRegistry `json:"components,omitempty"`
}

type TeamDefinition struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	NameZh           string            `json:"name_zh,omitempty"`
	Description      string            `json:"description,omitempty"`
	DescriptionZh    string            `json:"description_zh,omitempty"`
	Version          string            `json:"version,omitempty"`
	Author           string            `json:"author,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Roles            []RoleDefinition  `json:"roles,omitempty"`
	Rules            []RuleDefinition  `json:"rules,omitempty"`
	Flows            []TeamFlowRef     `json:"flows,omitempty"`
	DefaultFlow      string            `json:"default_flow,omitempty"`
	SkillTags        []string          `json:"skill_tags,omitempty"`
	SkillRequirements []SkillRequirement `json:"skill_requirements,omitempty"`
}

type SkillRequirement struct {
	Tag         string   `json:"tag"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Skills      []string `json:"skills,omitempty"`
}

type TeamFlowRef struct {
	ID          string `json:"id"`
	File        string `json:"file"`
	Description string `json:"description,omitempty"`
	Default     bool   `json:"default,omitempty"`
}
