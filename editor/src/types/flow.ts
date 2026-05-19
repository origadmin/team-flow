import type { Node, Edge } from '@xyflow/react';

export type NodeType =
  | 'start'
  | 'phase'
  | 'gate'
  | 'branch'
  | 'parallel'
  | 'subflow'
  | 'loop'
  | 'manual'
  | 'event'
  | 'terminal';

export type EdgeType = 'sequential' | 'conditional';

export type TerminalStatus = 'success' | 'failed' | 'rejected' | 'cancelled' | 'timeout';

export type ComponentSource = 'builtin' | 'custom' | 'framework' | 'file' | 'marketplace' | 'trae';

export type ConstraintType = 'file_write' | 'tool_restriction' | 'timeout' | 'max_sub_agents' | 'custom';

export type BackoffStrategy = 'none' | 'linear' | 'exponential';

export type ApprovalStrategy = 'all' | 'any' | 'majority';

export type ParallelStrategy = 'all_success' | 'any_success' | 'first_success' | 'majority_success';

export type MergeStrategy = 'wait_all' | 'wait_any' | 'wait_majority';

export type EventTriggerType = 'webhook' | 'file_change' | 'schedule' | 'task_status' | 'manual' | 'custom';

export type ExecutionStatus = 'pending' | 'running' | 'done' | 'failed' | 'skipped';

export interface FlowMetadata {
  name: string;
  description?: string;
  domain?: string;
  author?: string;
  created_at?: string;
  updated_at?: string;
  tags?: string[];
  task_type?: string;
}

export interface FlowConfig {
  domain?: string;
  task_type?: string;
  auto_dispatch?: boolean;
  parallel_limit?: number;
  timeout_minutes?: number;
  context?: Record<string, string>;
}

export interface ComponentRef {
  ref: string;
  source: ComponentSource;
  version?: string;
  path?: string;
  config?: Record<string, unknown>;
}

export interface ToolRef extends ComponentRef {
  commands?: string[];
  scope?: 'codebase' | 'docs' | 'config' | 'all';
}

export interface Constraint {
  type: ConstraintType;
  allowed_paths?: string[];
  forbidden?: string[];
  allowed?: string[];
  minutes?: number;
  limit?: number;
}

export interface NodeComponents {
  roles?: ComponentRef[];
  rules?: ComponentRef[];
  tools?: ToolRef[];
  skills?: ComponentRef[];
  prompts?: ComponentRef[];
  constraints?: Constraint[];
}

export interface DocSpec {
  name: string;
  format: string;
  path: string;
  template?: string;
  required: boolean;
  description?: string;
}

export interface GateCondition {
  type: string;
  threshold?: string;
  required?: boolean;
  check?: string;
  expected?: string;
}

export interface AutoRetryConfig {
  max_attempts: number;
  delay_minutes: number;
}

export interface PhaseConfig {
  role?: string;
  auto_dispatch?: boolean;
  timeout_minutes?: number;
  deliverables?: string[];
}

export interface GateConfig {
  conditions: GateCondition[];
  on_pass?: string;
  on_fail?: string;
  auto_retry?: AutoRetryConfig;
}

export interface BranchCondition {
  when?: string;
  goto?: string;
  default?: string;
  target?: string;
}

export interface BranchConfig {
  conditions: BranchCondition[];
}

export interface ParallelBranch {
  node: string;
  weight?: number;
}

export interface ParallelConfig {
  max_concurrency?: number;
  strategy?: ParallelStrategy;
  branches: ParallelBranch[];
  merge_strategy?: MergeStrategy;
}

export interface SubflowConfig {
  flow_ref: string;
  input_mapping?: Record<string, string>;
  output_mapping?: Record<string, string>;
}

export interface LoopConfig {
  max_attempts?: number;
  delay_minutes?: number;
  backoff?: BackoffStrategy;
  exit_condition?: string;
}

export interface Approver {
  role: string;
  required: boolean;
}

export interface ManualConfig {
  approvers: Approver[];
  approval_strategy?: ApprovalStrategy;
  timeout_minutes?: number;
  on_timeout?: string;
}

export interface EventTriggerConfig {
  type: EventTriggerType;
  pattern?: string;
  filter?: Record<string, unknown>;
}

export interface EventConfig {
  trigger: EventTriggerConfig;
  timeout_minutes?: number;
  on_timeout?: string;
}

export interface TerminalConfig {
  status: TerminalStatus;
  message?: string;
}

export interface StartConfig {}

export type NodeConfig = StartConfig | PhaseConfig | GateConfig | BranchConfig | ParallelConfig | SubflowConfig | LoopConfig | ManualConfig | EventConfig | TerminalConfig;

export interface Action {
  action: string;
  [key: string]: unknown;
}

export interface ErrorHandler {
  strategy: string;
  retry?: AutoRetryConfig;
  fallback?: string;
}

export interface FlowNode {
  id: string;
  type: NodeType;
  name: string;
  description?: string;
  config?: NodeConfig;
  components?: NodeComponents;
  docs?: DocSpec[];
  gates?: GateConfig[];
  on_enter?: Action[];
  on_exit?: Action[];
  on_error?: ErrorHandler;
}

export interface EdgeCondition {
  expression: string;
}

export interface FlowEdge {
  id?: string;
  from: string;
  to: string;
  type?: EdgeType;
  conditions?: EdgeCondition[];
}

export interface FlowNodeOverride {
  id: string;
  [key: string]: unknown;
}

export interface Flow {
  version: string;
  metadata: FlowMetadata;
  config?: FlowConfig;
  extends?: string;
  overrides?: FlowNodeOverride[];
  components?: Record<string, unknown>;
  nodes: FlowNode[];
  edges: FlowEdge[];
  variables?: Record<string, string>;
}

export interface PhaseNodeData extends Record<string, unknown> {
  flowNode: FlowNode;
  config: PhaseConfig;
  selected: boolean;
  executionStatus?: ExecutionStatus;
}

export interface StartNodeData extends Record<string, unknown> {
  flowNode: FlowNode;
  config: StartConfig;
  selected: boolean;
  executionStatus?: ExecutionStatus;
}

export interface GateNodeData extends Record<string, unknown> {
  flowNode: FlowNode;
  config: GateConfig;
  selected: boolean;
  executionStatus?: ExecutionStatus;
}

export interface TerminalNodeData extends Record<string, unknown> {
  flowNode: FlowNode;
  config: TerminalConfig;
  selected: boolean;
  executionStatus?: ExecutionStatus;
}

export type StartNode = Node<StartNodeData, 'start'>;
export type PhaseNode = Node<PhaseNodeData, 'phase'>;
export type GateNode = Node<GateNodeData, 'gate'>;
export type TerminalNode = Node<TerminalNodeData, 'terminal'>;

export type FlowEditorNode = StartNode | PhaseNode | GateNode | TerminalNode;

export interface ConditionalEdgeData extends Record<string, unknown> {
  edgeType: EdgeType;
  conditions?: EdgeCondition[];
}

export type ConditionalEdge = Edge<ConditionalEdgeData>;

export interface NodePaletteItem {
  type: NodeType;
  icon: string;
  color: string;
  label: string;
  description: string;
}
