import { memo } from 'react';
import { BaseEdge, EdgeLabelRenderer, getBezierPath } from '@xyflow/react';
import type { EdgeProps } from '@xyflow/react';
import type { ConditionalEdgeData } from '../../types/flow';

function buildBackEdgePath(
  sourceX: number,
  sourceY: number,
  targetX: number,
  targetY: number,
): { path: string; labelX: number; labelY: number } {
  const offsetY = 60;
  const midX = (sourceX + targetX) / 2;
  const belowY = Math.max(sourceY, targetY) + offsetY;

  const sx = sourceX;
  const sy = sourceY + 8;
  const tx = targetX;
  const ty = targetY + 8;

  return {
    path: `M ${sx} ${sy} C ${sx} ${belowY}, ${tx} ${belowY}, ${tx} ${ty}`,
    labelX: midX,
    labelY: belowY + 10,
  };
}

function ConditionalEdgeComponent({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
  selected,
}: EdgeProps) {
  const edgeData = data as ConditionalEdgeData | undefined;
  const edgeType = edgeData?.edgeType ?? 'sequential';
  const conditions = edgeData?.conditions;

  const isBackEdge = sourceX > targetX + 50;

  let edgePath: string;
  let labelX: number;
  let labelY: number;

  if (isBackEdge) {
    const back = buildBackEdgePath(sourceX, sourceY, targetX, targetY);
    edgePath = back.path;
    labelX = back.labelX;
    labelY = back.labelY;
  } else {
    [edgePath, labelX, labelY] = getBezierPath({
      sourceX,
      sourceY,
      targetX,
      targetY,
      sourcePosition,
      targetPosition,
    });
  }

  const isConditional = edgeType === 'conditional';
  const isPositive = conditions?.some((c) => !c.expression.startsWith('!'));
  const isNegative = conditions?.some((c) => c.expression.startsWith('!'));

  let strokeColor = '#9CA3AF';
  let strokeDasharray = '';
  if (isBackEdge) {
    strokeColor = '#EF4444';
    strokeDasharray = '6 3';
  } else if (isConditional) {
    strokeDasharray = '5 5';
    if (isPositive && !isNegative) {
      strokeColor = '#10B981';
    } else if (isNegative && !isPositive) {
      strokeColor = '#EF4444';
    }
  }

  const label = isConditional && conditions?.length
    ? conditions.map((c) => c.expression).join(', ')
    : undefined;

  return (
    <>
      <BaseEdge
        id={id}
        path={edgePath}
        style={{
          stroke: strokeColor,
          strokeWidth: selected ? 2 : 1.5,
          strokeDasharray,
        }}
      />
      {label && (
        <EdgeLabelRenderer>
          <div
            style={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              background: '#fff',
              padding: '2px 6px',
              borderRadius: 3,
              fontSize: 10,
              color: strokeColor,
              border: `1px solid ${strokeColor}`,
              pointerEvents: 'all',
              whiteSpace: 'nowrap',
            }}
            className="nodrag nopan"
          >
            {label}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
}

export const ConditionalEdge = memo(ConditionalEdgeComponent);
