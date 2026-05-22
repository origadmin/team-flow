const API_BASE = '/api';

export interface TeamMeta {
  id: string;
  name: string;
  name_zh: string;
  description: string;
  flow_count: number;
  default_flow: string;
}

export interface TeamFlowMeta {
  id: string;
  file: string;
  description: string;
  default: boolean;
}

export interface ProjectConfig {
  name: string;
  default_flow: string;
  paths: {
    docs_internal?: string;
    docs_external?: string;
  };
  toolchain?: Record<string, unknown>;
}

export interface ProjectFlowMeta {
  id: string;
  default: boolean;
}

export interface ValidationError {
  node_id?: string;
  field: string;
  msg: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationError[];
}

export interface TeamValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationError[];
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, init);
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as { error?: string }).error ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export async function fetchTeams(): Promise<TeamMeta[]> {
  return request<TeamMeta[]>('/teams');
}

export async function fetchTeamFlows(teamId: string): Promise<TeamFlowMeta[]> {
  return request<TeamFlowMeta[]>(`/teams/${teamId}/flows`);
}

export async function fetchTeamFlowJSON(teamId: string, flowId: string): Promise<Record<string, unknown>> {
  const res = await fetch(`${API_BASE}/teams/${teamId}/flows/${flowId}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export async function fetchProjectConfig(): Promise<ProjectConfig> {
  return request<ProjectConfig>('/project');
}

export async function fetchProjectFlows(): Promise<ProjectFlowMeta[]> {
  return request<ProjectFlowMeta[]>('/project/flows');
}

export async function fetchProjectFlowJSON(flowName: string): Promise<Record<string, unknown>> {
  const res = await fetch(`${API_BASE}/project/flows/${flowName}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export async function saveProjectFlowJSON(flowName: string, data: unknown): Promise<{ ok: boolean; path: string }> {
  return request<{ ok: boolean; path: string }>(`/project/flows/${flowName}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
}

export async function validateFlow(flowName: string, data: unknown): Promise<ValidationResult> {
  return request<ValidationResult>(`/project/flows/${flowName}/validate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
}

export async function validateTeam(): Promise<TeamValidationResult> {
  return request<TeamValidationResult>('/project/validate-team', {
    method: 'POST',
  });
}
