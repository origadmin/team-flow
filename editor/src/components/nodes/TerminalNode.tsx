import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import type { TerminalNodeData } from '../../types/flow';
import { NodeIcon } from './NodeIcon';

const statusColors: Record<string, string> = {
  success: '#10B981',
  failed: '#EF4444',
  rejected: '#F59E0B',
  cancelled: '#6B7280',
  timeout: '#F97316',
};

function TerminalNodeComponent({ data, selected }: NodeProps) {
  const nodeData = data as unknown as TerminalNodeData;
  const config = nodeData.config;
  const flowNode = nodeData.flowNode;
  const statusColor = statusColors[config.status] ?? '#6B7280';

  return (
    <div
      className="flow-node terminal-node"
      style={{
        borderLeft: `4px solid #6B7280`,
        background: '#fff',
        borderRadius: 6,
        padding: '8px 12px',
        minWidth: 160,
        boxShadow: selected
          ? '0 0 0 2px #6B7280, 0 2px 8px rgba(107,114,128,0.3)'
          : '0 1px 4px rgba(0,0,0,0.1)',
        fontSize: 12,
        transition: 'box-shadow 0.15s ease',
      }}
    >
      <Handle type="target" position={Position.Left} style={{ background: '#6B7280' }} />
      <div style={{ fontWeight: 600, marginBottom: 4, display: 'flex', alignItems: 'center', gap: 6 }}>
        <NodeIcon type="terminal" size={16} />
        <span>{flowNode.name}</span>
      </div>
      <div style={{ color: '#9CA3AF', fontSize: 10, fontFamily: 'monospace', marginBottom: 4 }}>id: {flowNode.id}</div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
        <span style={{ color: '#6B7280' }}>Terminal</span>
        <span
          style={{
            background: statusColor,
            color: '#fff',
            padding: '1px 6px',
            borderRadius: 3,
            fontSize: 10,
            fontWeight: 600,
            textTransform: 'capitalize',
          }}
        >
          {config.status}
        </span>
      </div>
    </div>
  );
}

export const TerminalNode = memo(TerminalNodeComponent);
