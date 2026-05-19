import { useState } from 'react';
import { useFlowStore } from '../stores/flow-store';
import type { FlowNode, FlowEdge, PhaseConfig, GateConfig, TerminalConfig, ComponentRef, ToolRef, DocSpec, GateCondition } from '../types/flow';

interface SessionResetRule {
  trigger: string;
  action: string;
  role: string;
  rules: string[];
}

interface ProcRunResult {
  flow: { name: string; version: string; domain: string };
  current: {
    node_id: string;
    node_type: string;
    name: string;
    description: string;
    role: string;
    rules: ComponentRef[];
    tools: ToolRef[];
    skills: (ComponentRef & { resolved_path?: string })[];
    docs: DocSpec[];
    on_enter: { action: string; [k: string]: unknown }[];
    gate_conditions: GateCondition[] | null;
    is_terminal: boolean;
    terminal_status?: string;
    terminal_message?: string;
    constraints: { type: string; [k: string]: unknown }[];
    status_line: string;
    session_reset: SessionResetRule | null;
  };
  next_options: {
    node_id: string;
    role: string;
    condition: string | null;
    is_default: boolean;
  }[];
}

function generateProcRunResult(
  node: FlowNode,
  edges: FlowEdge[],
  nodes: FlowNode[],
  flowName: string,
  flowDomain: string,
): ProcRunResult {
  const outgoingEdges = edges.filter((e) => e.from === node.id);
  const nodeMap = new Map(nodes.map((n) => [n.id, n]));

  let role = '';
  let gate_conditions: GateCondition[] | null = null;
  let is_terminal = false;
  let terminal_status: string | undefined;
  let terminal_message: string | undefined;
  const constraints: { type: string; [k: string]: unknown }[] = [];

  switch (node.type) {
    case 'phase': {
      const config = (node.config ?? {}) as PhaseConfig;
      role = config.role ?? '';
      break;
    }
    case 'gate': {
      const config = (node.config ?? {}) as GateConfig;
      gate_conditions = config.conditions ?? null;
      break;
    }
    case 'terminal': {
      const config = (node.config ?? {}) as TerminalConfig;
      is_terminal = true;
      terminal_status = config.status;
      terminal_message = config.message;
      break;
    }
  }

  if (node.components?.constraints) {
    for (const c of node.components.constraints) {
      constraints.push({ ...c, type: c.type });
    }
  }

  const next_options = outgoingEdges.map((edge, idx) => {
    const targetNode = nodeMap.get(edge.to);
    let targetRole = '';
    if (targetNode?.type === 'phase') {
      const tc = (targetNode.config ?? {}) as PhaseConfig;
      targetRole = tc.role ?? '';
    }
    const condition = edge.conditions?.[0]?.expression ?? null;
    const isDefault = idx === 0 && (edge.type !== 'conditional' || (condition !== null && !condition.startsWith('!')));
    return { node_id: edge.to, role: targetRole, condition, is_default: isDefault };
  });

  const SKILL_PATH = '{SKILL_PATH}';
  const TEAM_PATH = '{TEAM_PATH}';
  const skills = (node.components?.skills ?? []).map((s) => {
    let resolved_path = '';
    switch (s.source) {
      case 'builtin': resolved_path = `${SKILL_PATH}/skills/${s.ref}/SKILL.md`; break;
      case 'custom': resolved_path = `${TEAM_PATH}/components/skills/${s.ref}/SKILL.md`; break;
      case 'framework': resolved_path = `${SKILL_PATH}/framework/${s.ref}/SKILL.md`; break;
      case 'marketplace': resolved_path = `${SKILL_PATH}/marketplace/${s.ref}/SKILL.md`; break;
      case 'trae': resolved_path = `.trae/skills/${s.ref}/SKILL.md`; break;
      case 'file': resolved_path = s.path ?? ''; break;
    }
    return { ...s, resolved_path };
  });

  const phase = node.on_enter?.[0] && 'phase' in node.on_enter[0]
    ? String(node.on_enter[0].phase)
    : node.type;
  const statusLine = `[Role: ${role || (node.type === 'gate' ? 'Gate' : node.type === 'terminal' ? 'Terminal' : '')} | TaskPool: {task_id}#1 | Phase: ${phase} | Asset: {project-name}]`;

  let session_reset: SessionResetRule | null = null;
  if (node.type === 'terminal') {
    session_reset = {
      trigger: 'new_user_input',
      action: 'flow proc run',
      role: 'Triage',
      rules: [
        'Check flow task ready for in_progress tasks',
        'If no in_progress tasks, execute flow proc run to return to Triage',
        'Never skip Triage — all new input must be classified first',
      ],
    };
  }

  return {
    flow: { name: flowName, version: 'v3', domain: flowDomain },
    current: {
      node_id: node.id,
      node_type: node.type,
      name: node.name,
      description: node.description ?? '',
      role,
      rules: node.components?.rules ?? [],
      tools: node.components?.tools ?? [],
      skills,
      docs: node.docs ?? [],
      on_enter: node.on_enter ?? [],
      gate_conditions,
      is_terminal,
      terminal_status,
      terminal_message,
      constraints,
      status_line: statusLine,
      session_reset,
    },
    next_options,
  };
}

function generateAiPrompt(result: ProcRunResult): string {
  const c = result.current;
  const lines: string[] = [];

  lines.push(c.status_line);
  lines.push('');
  lines.push(`你正在执行阶段 "${c.name}" (${c.node_id})`);
  lines.push('');

  if (c.role) {
    lines.push(`## 你的角色`);
    lines.push(c.role);
    lines.push('');
  }

  if (c.rules.length > 0) {
    lines.push(`## 你必须遵守的规则`);
    for (const r of c.rules) {
      lines.push(`- ${r.ref} (${r.source})`);
    }
    lines.push('');
  }

  if (c.tools.length > 0) {
    lines.push(`## 你可以使用的工具`);
    for (const t of c.tools) {
      const cmds = t.commands?.length ? ` [${t.commands.join(', ')}]` : '';
      lines.push(`- ${t.ref}${cmds} (${t.source})`);
    }
    lines.push('');
  }

  if (c.skills.length > 0) {
    lines.push(`## 你应该使用的技能`);
    for (const s of c.skills) {
      const path = s.resolved_path ? ` → load from ${s.resolved_path}` : '';
      lines.push(`- ${s.ref}${path}`);
    }
    lines.push('');
  }

  if (c.constraints.length > 0) {
    lines.push(`## 约束`);
    for (const ct of c.constraints) {
      if (ct.type === 'file_write') {
        const paths = (ct as any).allowed_paths?.join(', ') ?? '';
        lines.push(`- 文件写入限制: 仅允许写入 ${paths}`);
      } else if (ct.type === 'tool_restriction') {
        const forbidden = (ct as any).forbidden?.join(', ') ?? '';
        lines.push(`- 工具限制: 禁止使用 ${forbidden}`);
      } else if (ct.type === 'timeout') {
        lines.push(`- 超时: ${(ct as any).minutes} 分钟`);
      }
    }
    lines.push('');
  }

  if (c.docs.length > 0) {
    lines.push(`## 你必须产出的交付物`);
    for (const d of c.docs) {
      const req = d.required ? '[必须]' : '[可选]';
      lines.push(`- ${req} ${d.name} → ${d.path}`);
      if (d.description) {
        lines.push(`  ${d.description}`);
      }
    }
    lines.push('');
  }

  if (c.gate_conditions) {
    lines.push(`## 关卡条件（你需要评估）`);
    for (const gc of c.gate_conditions) {
      const req = gc.required !== false ? '[必须]' : '[建议]';
      const th = gc.threshold ? ` 阈值: ${gc.threshold}` : '';
      lines.push(`- ${req} ${gc.type}${th}`);
    }
    lines.push('');
    lines.push(`### AI 行为规则`);
    lines.push(`1. 评估所有条件`);
    lines.push(`2. 不确定时优先选择 is_default 分支`);
    lines.push(`3. 选择非默认分支时必须说明理由`);
    lines.push(`4. 完全无法判断时暂停并询问用户`);
    lines.push('');
  }

  if (c.is_terminal) {
    lines.push(`## 流程结束`);
    lines.push(`状态: ${c.terminal_status}`);
    if (c.terminal_message) lines.push(`消息: ${c.terminal_message}`);
    lines.push('');
    if (c.session_reset) {
      lines.push(`### 会话重置规则`);
      lines.push(`触发: ${c.session_reset.trigger}`);
      lines.push(`动作: ${c.session_reset.action}`);
      lines.push(`角色: ${c.session_reset.role}`);
      lines.push(`规则:`);
      c.session_reset.rules.forEach((r, i) => {
        lines.push(`${i + 1}. ${r}`);
      });
    }
  } else if (result.next_options.length > 0) {
    lines.push(`## 完成后`);
    for (const opt of result.next_options) {
      const cond = opt.condition ? ` [${opt.condition}]` : '';
      const def = opt.is_default ? ' [默认]' : '';
      const r = opt.role ? ` (${opt.role})` : '';
      lines.push(`→ ${opt.node_id}${r}${cond}${def}`);
    }
    lines.push('');
    lines.push(`下一步: flow proc run ${result.next_options[0]?.node_id ?? ''}`);
  }

  return lines.join('\n');
}

const jsonStyle: React.CSSProperties = {
  background: '#1E293B',
  color: '#E2E8F0',
  padding: 12,
  borderRadius: 6,
  fontSize: 11,
  fontFamily: 'Consolas, Monaco, "Courier New", monospace',
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-all',
  overflow: 'auto',
  maxHeight: 400,
  lineHeight: 1.5,
};

const promptStyle: React.CSSProperties = {
  background: '#FFFBEB',
  border: '1px solid #FDE68A',
  padding: 12,
  borderRadius: 6,
  fontSize: 11,
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-word',
  overflow: 'auto',
  maxHeight: 400,
  lineHeight: 1.6,
  fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
};

const tabActiveStyle: React.CSSProperties = {
  padding: '6px 12px',
  fontSize: 11,
  fontWeight: 600,
  border: 'none',
  borderBottom: '2px solid #3B82F6',
  background: 'transparent',
  color: '#3B82F6',
  cursor: 'pointer',
};

const tabInactiveStyle: React.CSSProperties = {
  padding: '6px 12px',
  fontSize: 11,
  fontWeight: 500,
  border: 'none',
  borderBottom: '2px solid transparent',
  background: 'transparent',
  color: '#6B7280',
  cursor: 'pointer',
};

export function ProcRunPreview({ nodeId }: { nodeId: string }) {
  const { nodes, edges, currentFlow } = useFlowStore();
  const node = nodes.find((n) => n.id === nodeId);
  const [tab, setTab] = useState<'prompt' | 'json'>('prompt');

  if (!node || !currentFlow) return null;

  const result = generateProcRunResult(
    node,
    edges,
    nodes,
    currentFlow.metadata.name,
    currentFlow.metadata.domain ?? currentFlow.config?.domain ?? '',
  );

  const prompt = generateAiPrompt(result);
  const json = JSON.stringify(result, null, 2);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ display: 'flex', borderBottom: '1px solid #E5E7EB' }}>
        <button style={tab === 'prompt' ? tabActiveStyle : tabInactiveStyle} onClick={() => setTab('prompt')}>
          AI 指令预览
        </button>
        <button style={tab === 'json' ? tabActiveStyle : tabInactiveStyle} onClick={() => setTab('json')}>
          ProcRunResult
        </button>
      </div>

      <div style={{ fontSize: 10, color: '#9CA3AF', marginBottom: 4 }}>
        flow proc run {node.id}
      </div>

      {tab === 'prompt' && (
        <pre style={promptStyle}>{prompt}</pre>
      )}

      {tab === 'json' && (
        <pre style={jsonStyle}>{json}</pre>
      )}
    </div>
  );
}

