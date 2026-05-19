import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import type { GateNodeData } from '../../types/flow';
import { NodeIcon } from './NodeIcon';

function GateNodeComponent({ data, selected }: NodeProps) {
  const nodeData = data as unknown as GateNodeData;
  const config = nodeData.config;
  const flowNode = nodeData.flowNode;
  const conditions = config.conditions ?? [];

  return (
    <div
      className="flow-node gate-node"
      style={{
        borderLeft: `4px solid #F59E0B`,
        background: '#fff',
        borderRadius: 6,
        padding: '8px 12px',
        minWidth: 180,
        boxShadow: selected
          ? '0 0 0 2px #F59E0B, 0 2px 8px rgba(245,158,11,0.3)'
          : '0 1px 4px rgba(0,0,0,0.1)',
        fontSize: 12,
        transition: 'box-shadow 0.15s ease',
      }}
    >
      <Handle type="target" position={Position.Left} style={{ background: '#F59E0B' }} />
      <div style={{ fontWeight: 600, marginBottom: 4, display: 'flex', alignItems: 'center', gap: 6 }}>
        <NodeIcon type="gate" size={16} />
        <span>{flowNode.name}</span>
      </div>
      <div style={{ color: '#9CA3AF', fontSize: 10, fontFamily: 'monospace', marginBottom: 4 }}>id: {flowNode.id}</div>
      <div style={{ color: '#6B7280', marginBottom: 6 }}>
        Gate · {conditions.length} condition{conditions.length !== 1 ? 's' : ''}
      </div>
      <div style={{ borderTop: '1px solid #E5E7EB', paddingTop: 6, display: 'flex', flexDirection: 'column', gap: 2 }}>
        {conditions.map((c, i) => (
          <span key={i} style={{ color: '#6B7280', display: 'flex', alignItems: 'center', gap: 4 }}>
            {c.required !== false ? (
              <span style={{ color: '#10B981', fontSize: 10, fontWeight: 600 }}>✅</span>
            ) : (
              <span style={{ color: '#F59E0B', fontSize: 10, fontWeight: 600 }}>⚠</span>
            )}
            <span>{c.type}</span>
          </span>
        ))}
      </div>
      <Handle type="source" position={Position.Right} style={{ background: '#F59E0B' }} />
    </div>
  );
}

export const GateNode = memo(GateNodeComponent);
