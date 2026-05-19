import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';

function StartNodeComponent({ selected }: NodeProps) {
  return (
    <>
      <div
        style={{
          width: 16,
          height: 16,
          borderRadius: '50%',
          background: '#10B981',
          boxShadow: selected
            ? '0 0 0 3px #10B981, 0 0 0 6px rgba(16,185,129,0.2)'
            : '0 1px 3px rgba(0,0,0,0.2)',
          transition: 'box-shadow 0.15s ease',
        }}
      />
      <Handle type="source" position={Position.Right} style={{ background: '#10B981', opacity: 0 }} />
    </>
  );
}

export const StartNode = memo(StartNodeComponent);
