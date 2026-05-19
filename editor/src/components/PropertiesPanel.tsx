import { useState } from 'react';
import { t } from '../i18n';
import { useFlowStore } from '../stores/flow-store';
import type { EdgeCondition, EdgeType, GateConfig, PhaseConfig, StartConfig, TerminalConfig, TerminalStatus, ComponentRef, ToolRef, DocSpec } from '../types/flow';
import { BUILTIN_ROLES, BUILTIN_RULES, BUILTIN_TOOLS, BUILTIN_SKILLS, DOMAIN_OPTIONS, getRolesByDomain, getRulesByDomain, getSkillsByDomain } from '../data/builtin-components';
import { ProcRunPreview } from './ProcRunPreview';

const inputStyle: React.CSSProperties = {
  width: '100%',
  padding: '4px 8px',
  border: '1px solid #D1D5DB',
  borderRadius: 4,
  fontSize: 12,
  boxSizing: 'border-box',
};

const textareaStyle: React.CSSProperties = {
  ...inputStyle,
  minHeight: 60,
  resize: 'vertical',
};

const labelStyle: React.CSSProperties = {
  fontSize: 11,
  color: '#6B7280',
  marginBottom: 2,
  display: 'block',
};

const sectionStyle: React.CSSProperties = {
  fontSize: 11,
  color: '#6B7280',
  fontWeight: 600,
  textTransform: 'uppercase',
  marginTop: 12,
  marginBottom: 6,
  borderTop: '1px solid #E5E7EB',
  paddingTop: 8,
};

const addBtnStyle: React.CSSProperties = {
  color: '#2563EB',
  cursor: 'pointer',
  fontSize: 12,
  background: 'none',
  border: 'none',
  padding: 0,
};

const removeBtnStyle: React.CSSProperties = {
  color: '#DC2626',
  cursor: 'pointer',
  fontSize: 12,
  background: 'none',
  border: 'none',
  padding: 0,
};

const cardStyle: React.CSSProperties = {
  marginBottom: 6,
  padding: '4px 6px',
  background: '#fff',
  border: '1px solid #E5E7EB',
  borderRadius: 4,
};

const fieldStyle: React.CSSProperties = {
  marginBottom: 8,
};

const selectStyle: React.CSSProperties = {
  ...inputStyle,
  background: '#fff',
};

function ComponentSelector({ items, onAdd }: { items: { ref: string; source: string; label: string; commands?: string[] }[]; onAdd: (item: ComponentRef | ToolRef) => void }) {
  const [selected, setSelected] = useState('');
  const available = items;
  return (
    <div style={{ display: 'flex', gap: 4, marginBottom: 4 }}>
      <select style={selectStyle} value={selected} onChange={(e) => setSelected(e.target.value)}>
        <option value="">-- 选择 --</option>
        {available.map((item) => (
          <option key={item.ref} value={item.ref}>{item.label}</option>
        ))}
      </select>
      <button
        style={addBtnStyle}
        disabled={!selected}
        onClick={() => {
          const item = available.find((i) => i.ref === selected);
          if (!item) return;
          if (item.commands) {
            onAdd({ ref: item.ref, source: item.source as any, commands: item.commands });
          } else {
            onAdd({ ref: item.ref, source: item.source as any });
          }
          setSelected('');
        }}
      >
        +
      </button>
    </div>
  );
}

function StartEditor() {
  return (
    <div style={{ color: '#9CA3AF', fontSize: 12, textAlign: 'center', marginTop: 40 }}>
      流程入口，无需配置
    </div>
  );
}

function PhaseEditor({ nodeId }: { nodeId: string }) {
  const { nodes, updateNode, currentFlow } = useFlowStore();
  const node = nodes.find((n) => n.id === nodeId)!;
  const config = (node.config ?? {}) as PhaseConfig;
  const domain = currentFlow?.metadata.domain ?? '';
  const filteredRoles = getRolesByDomain(domain);
  const filteredRules = getRulesByDomain(domain);
  const filteredSkills = getSkillsByDomain(domain);
  const p = t().properties;

  function setConfig(next: PhaseConfig) {
    updateNode(nodeId, { config: next });
  }

  return (
    <>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.name}</label>
        <input style={inputStyle} value={node.name} onChange={(e) => updateNode(nodeId, { name: e.target.value })} />
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>{p.description}</label>
        <textarea style={textareaStyle} value={node.description ?? ''} onChange={(e) => updateNode(nodeId, { description: e.target.value })} placeholder="这个阶段是做什么的？比如：分析需求、设计架构、代码实现等" />
      </div>

      <div style={{ ...sectionStyle, borderTop: 'none', paddingTop: 0, marginTop: 4 }}>
        📦 基础配置
      </div>

      <div style={fieldStyle}>
        <label style={{ ...labelStyle, display: 'inline-flex', alignItems: 'center', gap: 6 }}>
          <input type="checkbox" checked={config.auto_dispatch ?? false} onChange={(e) => setConfig({ ...config, auto_dispatch: e.target.checked })} />
          {p.autoDispatch}
        </label>
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>{p.timeout}</label>
        <input type="number" style={inputStyle} value={config.timeout_minutes ?? ''} onChange={(e) => setConfig({ ...config, timeout_minutes: Number(e.target.value) || undefined })} placeholder="默认无超时" />
      </div>

      <div style={sectionStyle}>📝 交付物（AI需要产出的内容）</div>
      {(node.docs ?? []).map((doc, i) => (
        <div key={i} style={cardStyle}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
            <span style={{ fontSize: 11, fontWeight: 500 }}>#{i + 1}</span>
            <button style={removeBtnStyle} onClick={() => { const docs = [...(node.docs ?? [])]; docs.splice(i, 1); updateNode(nodeId, { docs }); }}>×</button>
          </div>
          <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="文件名，如：SPEC.md" value={doc.name} onChange={(e) => { const docs = [...(node.docs ?? [])]; docs[i] = { ...docs[i]!, name: e.target.value }; updateNode(nodeId, { docs }); }} />
          <select style={{ ...selectStyle, marginBottom: 2 }} value={doc.format} onChange={(e) => { const docs = [...(node.docs ?? [])]; docs[i] = { ...docs[i]!, format: e.target.value }; updateNode(nodeId, { docs }); }}>
            <option value="markdown">markdown</option>
            <option value="yaml">yaml</option>
            <option value="json">json</option>
            <option value="go">go</option>
            <option value="typescript">typescript</option>
            <option value="python">python</option>
            <option value="text">text</option>
          </select>
          <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="保存路径，如：docs/{task_id}/SPEC.md" value={doc.path} onChange={(e) => { const docs = [...(node.docs ?? [])]; docs[i] = { ...docs[i]!, path: e.target.value }; updateNode(nodeId, { docs }); }} />
          <textarea style={{ ...textareaStyle, marginBottom: 2, minHeight: 60 }} placeholder="告诉AI这个文件要写什么内容..." value={doc.description ?? ''} onChange={(e) => { const docs = [...(node.docs ?? [])]; docs[i] = { ...docs[i]!, description: e.target.value }; updateNode(nodeId, { docs }); }} />
          <label style={{ fontSize: 11, color: '#6B7280', display: 'inline-flex', alignItems: 'center', gap: 4 }}>
            <input type="checkbox" checked={doc.required} onChange={(e) => { const docs = [...(node.docs ?? [])]; docs[i] = { ...docs[i]!, required: e.target.checked }; updateNode(nodeId, { docs }); }} />
            {p.deliverableRequired}
          </label>
        </div>
      ))}
      <button style={addBtnStyle} onClick={() => { const docs = [...(node.docs ?? []), { name: '', format: 'markdown', path: '', template: '', required: false, description: '' }]; updateNode(nodeId, { docs }); }}>{p.add}</button>

      <div style={sectionStyle}>🧩 组件配置</div>
      
      <div style={{ fontSize: 11, color: '#6B7280', marginBottom: 2, marginTop: 4 }}>👤 角色</div>
      {(node.components?.roles ?? []).map((role, i) => (
        <div key={i} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 2 }}>
          <span style={{ fontSize: 11 }}>{role.ref} ({role.source})</span>
          <button style={removeBtnStyle} onClick={() => { const components = { ...node.components }; components.roles = [...(components.roles ?? [])]; components.roles.splice(i, 1); updateNode(nodeId, { components }); }}>×</button>
        </div>
      ))}
      <ComponentSelector items={filteredRoles} onAdd={(item) => { const components = { ...node.components }; components.roles = [...(components.roles ?? []), item as ComponentRef]; updateNode(nodeId, { components }); }} />

      <div style={{ fontSize: 11, color: '#6B7280', marginBottom: 2, marginTop: 4 }}>📋 规则</div>
      {(node.components?.rules ?? []).map((rule, i) => (
        <div key={i} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 2 }}>
          <span style={{ fontSize: 11 }}>{rule.ref} ({rule.source})</span>
          <button style={removeBtnStyle} onClick={() => { const components = { ...node.components }; components.rules = [...(components.rules ?? [])]; components.rules.splice(i, 1); updateNode(nodeId, { components }); }}>×</button>
        </div>
      ))}
      <ComponentSelector items={filteredRules} onAdd={(item) => { const components = { ...node.components }; components.rules = [...(components.rules ?? []), item as ComponentRef]; updateNode(nodeId, { components }); }} />

      <div style={{ fontSize: 11, color: '#6B7280', marginBottom: 2, marginTop: 6 }}>🛠️ 工具</div>
      {(node.components?.tools ?? []).map((tool, i) => (
        <div key={i} style={{ ...cardStyle, marginBottom: 4 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontSize: 11 }}>{tool.ref}</span>
            <button style={removeBtnStyle} onClick={() => { const components = { ...node.components }; components.tools = [...(components.tools ?? [])]; components.tools.splice(i, 1); updateNode(nodeId, { components }); }}>×</button>
          </div>
          <input
            style={{ ...inputStyle, fontSize: 10, marginTop: 2 }}
            placeholder="允许的命令，逗号分隔，留空=全部可用"
            value={(tool as any).commands ? (tool as any).commands.join(', ') : ''}
            onChange={(e) => {
              const components = { ...node.components };
              components.tools = [...(components.tools ?? [])];
              const cmds = e.target.value.split(',').map((c: string) => c.trim()).filter(Boolean);
              components.tools[i] = { ...components.tools[i]!, commands: cmds.length > 0 ? cmds : undefined };
              updateNode(nodeId, { components });
            }}
          />
        </div>
      ))}
      <ComponentSelector items={BUILTIN_TOOLS} onAdd={(item) => { const components = { ...node.components }; components.tools = [...(components.tools ?? []), item as ToolRef]; updateNode(nodeId, { components }); }} />

      <div style={{ fontSize: 11, color: '#6B7280', marginBottom: 2, marginTop: 6 }}>🎯 技能</div>
      {(node.components?.skills ?? []).map((skill, i) => (
        <div key={i} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 2 }}>
          <span style={{ fontSize: 11 }}>{skill.ref} ({skill.source})</span>
          <button style={removeBtnStyle} onClick={() => { const components = { ...node.components }; components.skills = [...(components.skills ?? [])]; components.skills.splice(i, 1); updateNode(nodeId, { components }); }}>×</button>
        </div>
      ))}
      <ComponentSelector items={filteredSkills} onAdd={(item) => { const components = { ...node.components }; components.skills = [...(components.skills ?? []), item as ComponentRef]; updateNode(nodeId, { components }); }} />

      <div style={sectionStyle}>🚦 约束（限制AI的行为）</div>
      {(node.components?.constraints ?? []).map((ct, i) => (
        <div key={i} style={cardStyle}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
            <span style={{ fontSize: 11, fontWeight: 500 }}>{ct.type}</span>
            <button style={removeBtnStyle} onClick={() => { const components = { ...node.components }; components.constraints = [...(components.constraints ?? [])]; components.constraints.splice(i, 1); updateNode(nodeId, { components }); }}>×</button>
          </div>
          <select style={{ ...selectStyle, marginBottom: 2 }} value={ct.type} onChange={(e) => { const components = { ...node.components }; components.constraints = [...(components.constraints ?? [])]; components.constraints[i] = { ...components.constraints[i]!, type: e.target.value as any }; updateNode(nodeId, { components }); }}>
            <option value="file_write">文件写入限制</option>
            <option value="tool_restriction">工具限制</option>
            <option value="timeout">超时</option>
            <option value="max_sub_agents">最大子智能体数</option>
          </select>
          {ct.type === 'file_write' && (
            <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="允许的文件路径，如：{workspace}/" value={(ct.allowed_paths ?? []).join(', ')} onChange={(e) => { const components = { ...node.components }; components.constraints = [...(components.constraints ?? [])]; components.constraints[i] = { ...components.constraints[i]!, allowed_paths: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) }; updateNode(nodeId, { components }); }} />
          )}
          {ct.type === 'tool_restriction' && (
            <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="禁止的工具名称，如：shell-exec" value={(ct.forbidden ?? []).join(', ')} onChange={(e) => { const components = { ...node.components }; components.constraints = [...(components.constraints ?? [])]; components.constraints[i] = { ...components.constraints[i]!, forbidden: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) }; updateNode(nodeId, { components }); }} />
          )}
          {ct.type === 'timeout' && (
            <input type="number" style={{ ...inputStyle, marginBottom: 2 }} placeholder="分钟" value={ct.minutes ?? ''} onChange={(e) => { const components = { ...node.components }; components.constraints = [...(components.constraints ?? [])]; components.constraints[i] = { ...components.constraints[i]!, minutes: Number(e.target.value) || undefined }; updateNode(nodeId, { components }); }} />
          )}
          {ct.type === 'max_sub_agents' && (
            <input type="number" style={{ ...inputStyle, marginBottom: 2 }} placeholder="数量" value={ct.limit ?? ''} onChange={(e) => { const components = { ...node.components }; components.constraints = [...(components.constraints ?? [])]; components.constraints[i] = { ...components.constraints[i]!, limit: Number(e.target.value) || undefined }; updateNode(nodeId, { components }); }} />
          )}
        </div>
      ))}
      <button style={addBtnStyle} onClick={() => { const components = { ...node.components }; components.constraints = [...(components.constraints ?? []), { type: 'file_write', allowed_paths: [] }]; updateNode(nodeId, { components }); }}>+ 添加约束</button>

      <div style={sectionStyle}>▶️ 进入节点时执行的动作</div>
      {(node.on_enter ?? []).map((action, i) => (
        <div key={i} style={cardStyle}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
            <span style={{ fontSize: 11, fontWeight: 500 }}>{action.action}</span>
            <button style={removeBtnStyle} onClick={() => { const on_enter = [...(node.on_enter ?? [])]; on_enter.splice(i, 1); updateNode(nodeId, { on_enter }); }}>×</button>
          </div>
          <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="动作名称，如：update_task_phase" value={action.action} onChange={(e) => { const on_enter = [...(node.on_enter ?? [])]; on_enter[i] = { ...on_enter[i]!, action: e.target.value }; updateNode(nodeId, { on_enter }); }} />
          {Object.entries(action).filter(([k]) => k !== 'action').map(([k, v]) => (
            <div key={k} style={{ display: 'flex', gap: 4, marginBottom: 2 }}>
              <input style={{ ...inputStyle, flex: 1 }} value={k} readOnly />
              <input style={{ ...inputStyle, flex: 2 }} value={String(v)} onChange={(e) => { const on_enter = [...(node.on_enter ?? [])]; const next = { ...on_enter[i]! }; next[k] = e.target.value; on_enter[i] = next; updateNode(nodeId, { on_enter }); }} />
            </div>
          ))}
        </div>
      ))}
      <button style={addBtnStyle} onClick={() => { const on_enter = [...(node.on_enter ?? []), { action: 'update_task_phase', phase: '' }]; updateNode(nodeId, { on_enter }); }}>+ 添加动作</button>

      <div style={sectionStyle}>⏹️ 离开节点时执行的动作</div>
      {(node.on_exit ?? []).map((action, i) => (
        <div key={i} style={cardStyle}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
            <span style={{ fontSize: 11, fontWeight: 500 }}>{action.action}</span>
            <button style={removeBtnStyle} onClick={() => { const on_exit = [...(node.on_exit ?? [])]; on_exit.splice(i, 1); updateNode(nodeId, { on_exit }); }}>×</button>
          </div>
          <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="动作名称" value={action.action} onChange={(e) => { const on_exit = [...(node.on_exit ?? [])]; on_exit[i] = { ...on_exit[i]!, action: e.target.value }; updateNode(nodeId, { on_exit }); }} />
          {Object.entries(action).filter(([k]) => k !== 'action').map(([k, v]) => (
            <div key={k} style={{ display: 'flex', gap: 4, marginBottom: 2 }}>
              <input style={{ ...inputStyle, flex: 1 }} value={k} readOnly />
              <input style={{ ...inputStyle, flex: 2 }} value={String(v)} onChange={(e) => { const on_exit = [...(node.on_exit ?? [])]; const next = { ...on_exit[i]! }; next[k] = e.target.value; on_exit[i] = next; updateNode(nodeId, { on_exit }); }} />
            </div>
          ))}
        </div>
      ))}
      <button style={addBtnStyle} onClick={() => { const on_exit = [...(node.on_exit ?? []), { action: 'update_task_phase', phase: '' }]; updateNode(nodeId, { on_exit }); }}>+ 添加退出动作</button>
    </>
  );
}

function GateEditor({ nodeId }: { nodeId: string }) {
  const { nodes, updateNode } = useFlowStore();
  const node = nodes.find((n) => n.id === nodeId)!;
  const config = (node.config ?? {}) as GateConfig;
  const [retryExpanded, setRetryExpanded] = useState(false);
  const p = t().properties;

  function setConfig(next: GateConfig) {
    updateNode(nodeId, { config: next });
  }

  return (
    <>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.name}</label>
        <input style={inputStyle} value={node.name} onChange={(e) => updateNode(nodeId, { name: e.target.value })} />
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>{p.description}</label>
        <textarea style={textareaStyle} value={node.description ?? ''} onChange={(e) => updateNode(nodeId, { description: e.target.value })} placeholder="这个关卡的作用是什么？" />
      </div>

      <div style={sectionStyle}>✅ 关卡条件（AI检查的条件）</div>
      {(config.conditions ?? []).map((cond, i) => (
        <div key={i} style={cardStyle}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
            <span style={{ fontSize: 11, fontWeight: 500 }}>#{i + 1}</span>
            <button style={removeBtnStyle} onClick={() => { const conditions = [...(config.conditions ?? [])]; conditions.splice(i, 1); setConfig({ ...config, conditions }); }}>×</button>
          </div>
          <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="条件类型，如：tests_pass" value={cond.type} onChange={(e) => { const conditions = [...(config.conditions ?? [])]; conditions[i] = { ...conditions[i]!, type: e.target.value }; setConfig({ ...config, conditions }); }} />
          <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="阈值，如：100%" value={cond.threshold ?? ''} onChange={(e) => { const conditions = [...(config.conditions ?? [])]; conditions[i] = { ...conditions[i]!, threshold: e.target.value }; setConfig({ ...config, conditions }); }} />
          <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="期望结果，如：playable" value={cond.expected ?? ''} onChange={(e) => { const conditions = [...(config.conditions ?? [])]; conditions[i] = { ...conditions[i]!, expected: e.target.value }; setConfig({ ...config, conditions }); }} />
          <input style={{ ...inputStyle, marginBottom: 2 }} placeholder="AI如何检查？描述检查方法" value={cond.check ?? ''} onChange={(e) => { const conditions = [...(config.conditions ?? [])]; conditions[i] = { ...conditions[i]!, check: e.target.value }; setConfig({ ...config, conditions }); }} />
          <label style={{ fontSize: 11, color: '#6B7280', display: 'inline-flex', alignItems: 'center', gap: 4 }}>
            <input type="checkbox" checked={cond.required ?? true} onChange={(e) => { const conditions = [...(config.conditions ?? [])]; conditions[i] = { ...conditions[i]!, required: e.target.checked }; setConfig({ ...config, conditions }); }} />
            {p.conditionRequired}
          </label>
        </div>
      ))}
      <button style={addBtnStyle} onClick={() => { const conditions = [...(config.conditions ?? []), { type: '', threshold: '', required: true }]; setConfig({ ...config, conditions }); }}>{p.add}</button>

      <div style={{ ...sectionStyle, cursor: 'pointer', userSelect: 'none' }} onClick={() => setRetryExpanded(!retryExpanded)}>
        {retryExpanded ? '▼' : '▶'} {p.autoRetry}
      </div>
      {retryExpanded && (
        <>
          <div style={fieldStyle}>
            <label style={labelStyle}>{p.maxAttempts}</label>
            <input type="number" style={inputStyle} value={config.auto_retry?.max_attempts ?? ''} onChange={(e) => setConfig({ ...config, auto_retry: { max_attempts: Number(e.target.value) || 0, delay_minutes: config.auto_retry?.delay_minutes ?? 0 } })} />
          </div>
          <div style={fieldStyle}>
            <label style={labelStyle}>{p.delayMinutes}</label>
            <input type="number" style={inputStyle} value={config.auto_retry?.delay_minutes ?? ''} onChange={(e) => setConfig({ ...config, auto_retry: { max_attempts: config.auto_retry?.max_attempts ?? 0, delay_minutes: Number(e.target.value) || 0 } })} />
          </div>
        </>
      )}
    </>
  );
}

function TerminalEditor({ nodeId }: { nodeId: string }) {
  const { nodes, updateNode } = useFlowStore();
  const node = nodes.find((n) => n.id === nodeId)!;
  const config = (node.config ?? {}) as TerminalConfig;
  const p = t().properties;

  function setConfig(next: TerminalConfig) {
    updateNode(nodeId, { config: next });
  }

  return (
    <>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.name}</label>
        <input style={inputStyle} value={node.name} onChange={(e) => updateNode(nodeId, { name: e.target.value })} />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.description}</label>
        <textarea style={textareaStyle} value={node.description ?? ''} onChange={(e) => updateNode(nodeId, { description: e.target.value })} />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.status}</label>
        <select style={selectStyle} value={config.status ?? 'success'} onChange={(e) => setConfig({ ...config, status: e.target.value as TerminalStatus })}>
          <option value="success">{p.success}</option>
          <option value="failed">{p.failed}</option>
          <option value="rejected">{p.rejected}</option>
          <option value="cancelled">{p.cancelled}</option>
          <option value="timeout">{p.timeoutStatus}</option>
        </select>
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.message}</label>
        <textarea style={textareaStyle} value={config.message ?? ''} onChange={(e) => setConfig({ ...config, message: e.target.value })} />
      </div>
      <div style={sectionStyle}>Session Reset</div>
      <div style={{ background: '#F3F4F6', borderRadius: 4, padding: 8, fontSize: 11, color: '#6B7280', fontFamily: 'monospace', lineHeight: 1.6 }}>
        <div>触发: new_user_input</div>
        <div>动作: flow proc run</div>
        <div>角色: Triage</div>
        <div>规则:</div>
        <div style={{ paddingLeft: 8 }}>1. 检查是否有 in_progress 任务</div>
        <div style={{ paddingLeft: 8 }}>2. 如果没有，重置为 Triage</div>
        <div style={{ paddingLeft: 8 }}>3. 永远不要跳过 Triage</div>
      </div>
    </>
  );
}

function EdgeEditor({ edgeId }: { edgeId: string }) {
  const { edges, updateEdge } = useFlowStore();
  const edge = edges.find((e) => e.id === edgeId)!;
  const p = t().properties;

  return (
    <>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.from}</label>
        <input style={{ ...inputStyle, background: '#F3F4F6' }} value={edge.from} readOnly />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.to}</label>
        <input style={{ ...inputStyle, background: '#F3F4F6' }} value={edge.to} readOnly />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.edgeType}</label>
        <select style={selectStyle} value={edge.type ?? 'sequential'} onChange={(e) => updateEdge(edgeId, { type: e.target.value as EdgeType })}>
          <option value="sequential">{p.sequential}</option>
          <option value="conditional">{p.conditional}</option>
        </select>
      </div>
      {(edge.type ?? 'sequential') === 'conditional' && (
        <>
          <div style={sectionStyle}>{p.conditions}</div>
          {(edge.conditions ?? []).map((cond, i) => (
            <div key={i} style={{ display: 'flex', gap: 4, marginBottom: 4, alignItems: 'center' }}>
              <input style={inputStyle} placeholder="gate.passed / !gate.passed" value={cond.expression} onChange={(e) => { const conditions = [...(edge.conditions ?? [])]; conditions[i] = { expression: e.target.value }; updateEdge(edgeId, { conditions }); }} />
              <button style={removeBtnStyle} onClick={() => { const conditions = [...(edge.conditions ?? [])]; conditions.splice(i, 1); updateEdge(edgeId, { conditions }); }}>×</button>
            </div>
          ))}
          <button style={addBtnStyle} onClick={() => { const conditions = [...(edge.conditions ?? []), { expression: '' }]; updateEdge(edgeId, { conditions }); }}>{p.add}</button>
        </>
      )}
    </>
  );
}

function FlowEditor() {
  const { currentFlow, updateFlowMetadata, updateFlowConfig, updateFlowVariables } = useFlowStore();
  const p = t().properties;

  if (!currentFlow) {
    return <div style={{ color: '#9CA3AF', fontSize: 12, textAlign: 'center', marginTop: 40 }}>{p.selectHint}</div>;
  }

  return (
    <>
      <div style={{ fontWeight: 600, fontSize: 13, marginBottom: 8 }}>{p.flow}</div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.name}</label>
        <input style={inputStyle} value={currentFlow.metadata.name} onChange={(e) => updateFlowMetadata({ name: e.target.value })} />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.description}</label>
        <textarea style={textareaStyle} value={currentFlow.metadata.description ?? ''} onChange={(e) => updateFlowMetadata({ description: e.target.value })} />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.domain}</label>
        <select style={selectStyle} value={currentFlow.metadata.domain ?? ''} onChange={(e) => updateFlowMetadata({ domain: e.target.value })}>
          <option value="">-- 选择 --</option>
          {DOMAIN_OPTIONS.map((d) => (<option key={d} value={d}>{d}</option>))}
        </select>
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>Author</label>
        <input style={inputStyle} value={currentFlow.metadata.author ?? ''} onChange={(e) => updateFlowMetadata({ author: e.target.value })} placeholder="流程作者" />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>Tags</label>
        <input style={inputStyle} value={(currentFlow.metadata.tags ?? []).join(', ')} onChange={(e) => updateFlowMetadata({ tags: e.target.value.split(',').map((t: string) => t.trim()).filter(Boolean) })} placeholder="标签，逗号分隔" />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>Task Type</label>
        <select style={inputStyle} value={currentFlow.metadata.task_type ?? ''} onChange={(e) => updateFlowMetadata({ task_type: e.target.value })}>
          <option value="">无</option>
          <option value="feature">Feature</option>
          <option value="bugfix">Bug Fix</option>
          <option value="hotfix">Hot Fix</option>
          <option value="analysis">Analysis</option>
          <option value="change">Change</option>
        </select>
      </div>
      <div style={fieldStyle}>
        <label style={{ ...labelStyle, display: 'inline-flex', alignItems: 'center', gap: 6 }}>
          <input type="checkbox" checked={currentFlow.config?.auto_dispatch ?? false} onChange={(e) => updateFlowConfig({ auto_dispatch: e.target.checked })} />
          {p.autoDispatch}
        </label>
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.parallelLimit}</label>
        <input type="number" style={inputStyle} value={currentFlow.config?.parallel_limit ?? ''} onChange={(e) => updateFlowConfig({ parallel_limit: Number(e.target.value) || undefined })} />
      </div>
      <div style={fieldStyle}>
        <label style={labelStyle}>{p.flowTimeout}</label>
        <input type="number" style={inputStyle} value={currentFlow.config?.timeout_minutes ?? ''} onChange={(e) => updateFlowConfig({ timeout_minutes: Number(e.target.value) || undefined })} />
      </div>

      <div style={sectionStyle}>{p.variables}</div>
      {Object.entries(currentFlow.variables ?? {}).map(([k, v]) => (
        <div key={k} style={{ display: 'flex', gap: 4, marginBottom: 4, alignItems: 'center' }}>
          <input style={{ ...inputStyle, flex: 1 }} value={k} readOnly />
          <input style={{ ...inputStyle, flex: 2 }} value={v} onChange={(e) => { const variables = { ...(currentFlow.variables ?? {}) }; variables[k] = e.target.value; updateFlowVariables(variables); }} />
          <button style={removeBtnStyle} onClick={() => { const variables = { ...(currentFlow.variables ?? {}) }; delete variables[k]; updateFlowVariables(variables); }}>×</button>
        </div>
      ))}
      <button style={addBtnStyle} onClick={() => { const variables = { ...(currentFlow.variables ?? {}) }; variables['new_var'] = ''; updateFlowVariables(variables); }}>+ 添加变量</button>

      <div style={sectionStyle}>{p.context}</div>
      {Object.entries(currentFlow.config?.context ?? {}).map(([k, v]) => (
        <div key={k} style={{ display: 'flex', gap: 4, marginBottom: 4, alignItems: 'center' }}>
          <input style={{ ...inputStyle, flex: 1 }} value={k} readOnly />
          <input style={{ ...inputStyle, flex: 2 }} value={String(v)} onChange={(e) => { const context = { ...(currentFlow.config?.context ?? {}) }; context[k] = e.target.value; updateFlowConfig({ context }); }} />
          <button style={removeBtnStyle} onClick={() => { const context = { ...(currentFlow.config?.context ?? {}) }; delete context[k]; updateFlowConfig({ context }); }}>×</button>
        </div>
      ))}
      <button style={addBtnStyle} onClick={() => { const context = { ...(currentFlow.config?.context ?? {}) }; context['new_key'] = ''; updateFlowConfig({ context }); }}>+ 添加上下文</button>
    </>
  );
}

export function PropertiesPanel() {
  const { selectedNodeId, nodes, selectedEdgeId, edges } = useFlowStore();
  const selectedNode = selectedNodeId ? nodes.find((n) => n.id === selectedNodeId) : null;
  const selectedEdge = selectedEdgeId ? edges.find((e) => e.id === selectedEdgeId) : null;
  const p = t().properties;
  const [panelTab, setPanelTab] = useState<'edit' | 'preview'>('edit');

  const showPreview = selectedNode && selectedNode.type !== 'start';

  const tabActive: React.CSSProperties = {
    padding: '6px 12px', fontSize: 11, fontWeight: 600, border: 'none',
    borderBottom: '2px solid #3B82F6', background: 'transparent', color: '#3B82F6', cursor: 'pointer',
  };
  const tabInactive: React.CSSProperties = {
    padding: '6px 12px', fontSize: 11, fontWeight: 500, border: 'none',
    borderBottom: '2px solid transparent', background: 'transparent', color: '#6B7280', cursor: 'pointer',
  };

  return (
    <div style={{ width: 320, borderLeft: '1px solid #E5E7EB', background: '#F9FAFB', display: 'flex', flexDirection: 'column', overflow: 'auto', flexShrink: 0 }}>
      <div style={{ padding: '8px 12px', borderBottom: '1px solid #E5E7EB', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span style={{ fontWeight: 600, fontSize: 13 }}>{p.title}</span>
        {showPreview && (
          <div style={{ display: 'flex' }}>
            <button style={panelTab === 'edit' ? tabActive : tabInactive} onClick={() => setPanelTab('edit')}>编辑</button>
            <button style={panelTab === 'preview' ? tabActive : tabInactive} onClick={() => setPanelTab('preview')}>AI预览</button>
          </div>
        )}
      </div>
      <div style={{ padding: 12, flex: 1 }}>
        {panelTab === 'preview' && showPreview && selectedNode && (
          <ProcRunPreview nodeId={selectedNode.id} />
        )}
        {panelTab === 'edit' && (
          <>
            {selectedNode && selectedNode.type === 'start' && <StartEditor />}
            {selectedNode && selectedNode.type === 'phase' && <PhaseEditor nodeId={selectedNode.id} />}
            {selectedNode && selectedNode.type === 'gate' && <GateEditor nodeId={selectedNode.id} />}
            {selectedNode && selectedNode.type === 'terminal' && <TerminalEditor nodeId={selectedNode.id} />}
            {selectedNode && !['start', 'phase', 'gate', 'terminal'].includes(selectedNode.type) && (
              <div style={{ color: '#9CA3AF', fontSize: 12, textAlign: 'center', marginTop: 40 }}>{p.selectHint}</div>
            )}
            {selectedEdge && !selectedNode && <EdgeEditor edgeId={selectedEdge.id!} />}
            {!selectedNode && !selectedEdge && <FlowEditor />}
          </>
        )}
      </div>
    </div>
  );
}
