import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import type { PhaseNodeData } from '../../types/flow';
import { NodeIcon } from './NodeIcon';

function PhaseNodeComponent({ data, selected }: NodeProps) {
  const nodeData = data as unknown as PhaseNodeData;
  const config = nodeData.config;
  const flowNode = nodeData.flowNode;
  const docsCount = flowNode.docs?.length ?? 0;
  const rulesCount = flowNode.components?.rules?.length ?? 0;

  return (
    <div
      className="flow-node phase-node"
      style={{
        borderLeft: `4px solid #3B82F6`,
        background: '#fff',
        borderRadius: 6,
        padding: '8px 12px',
        minWidth: 180,
        boxShadow: selected
          ? '0 0 0 2px #3B82F6, 0 2px 8px rgba(59,130,246,0.3)'
          : '0 1px 4px rgba(0,0,0,0.1)',
        fontSize: 12,
        transition: 'box-shadow 0.15s ease',
      }}
    >
      <Handle type="target" position={Position.Left} style={{ background: '#3B82F6' }} />
      <div style={{ fontWeight: 600, marginBottom: 4, display: 'flex', alignItems: 'center', gap: 6 }}>
        <NodeIcon type="phase" size={16} />
        <span>{flowNode.name}</span>
      </div>
      <div style={{ color: '#9CA3AF', fontSize: 10, fontFamily: 'monospace', marginBottom: 4 }}>id: {flowNode.id}</div>
      {config.role && (
        <div style={{ color: '#6B7280', marginBottom: 4 }}>
          Role: {config.role}
        </div>
      )}
      <div style={{ borderTop: '1px solid #E5E7EB', paddingTop: 6, color: '#6B7280', fontSize: 11 }}>
        {docsCount > 0 && `📄 ${docsCount} doc${docsCount > 1 ? 's' : ''}`}
        {docsCount > 0 && rulesCount > 0 && ' · '}
        {rulesCount > 0 && `📏 ${rulesCount} rule${rulesCount > 1 ? 's' : ''}`}
      </div>
      <Handle type="source" position={Position.Right} style={{ background: '#3B82F6' }} />
    </div>
  );
}

export const PhaseNode = memo(PhaseNodeComponent);
