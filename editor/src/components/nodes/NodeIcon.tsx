import type { NodeType } from '../../types/flow';

interface NodeIconProps {
  type: NodeType;
  size?: number;
}

const iconDefs: Record<NodeType, { d: string; color: string }> = {
  start: {
    d: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 14.5v-9l6 4.5-6 4.5z',
    color: '#10B981',
  },
  phase: {
    d: 'M3 3h18v18H3V3zm2 2v14h14V5H5zm3 3h8v2H8V8zm0 4h8v2H8v-2z',
    color: '#3B82F6',
  },
  gate: {
    d: 'M12 2L2 7v10l10 5 10-5V7L12 2zm0 2.2L20 8.2v7.6L12 19.8 4 15.8V8.2L12 4.2zM11 9h2v4h-2V9zm0 6h2v2h-2v-2z',
    color: '#F59E0B',
  },
  branch: {
    d: 'M14 4h8v8h-2V7.4l-5.3 5.3-1.4-1.4L18.6 6H14V4zM4 6h7v2H6v8h5v2H4V6zm7 5h2v2h-2v-2z',
    color: '#10B981',
  },
  parallel: {
    d: 'M3 3h8v18H3V3zm2 2v14h4V5H5zm9-2h8v18h-8V3zm2 2v14h4V5h-4z',
    color: '#8B5CF6',
  },
  subflow: {
    d: 'M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14zm-8-2h2v-4h4v-2h-4V7h-2v4H7v2h4v4z',
    color: '#06B6D4',
  },
  loop: {
    d: 'M12 4V1L8 5l4 4V6c3.31 0 6 2.69 6 6s-2.69 6-6 6-6-2.69-6-6H4c0 4.42 3.58 8 8 8s8-3.58 8-8-3.58-8-8-8z',
    color: '#F97316',
  },
  manual: {
    d: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z',
    color: '#EC4899',
  },
  event: {
    d: 'M7 2v11h3v9l7-12h-4l4-8H7z',
    color: '#EAB308',
  },
  terminal: {
    d: 'M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z',
    color: '#6B7280',
  },
};

export function NodeIcon({ type, size = 16 }: NodeIconProps) {
  const icon = iconDefs[type];
  if (!icon) return null;

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill={icon.color}
      style={{ flexShrink: 0 }}
    >
      <path d={icon.d} />
    </svg>
  );
}

export function getNodeColor(type: NodeType): string {
  return iconDefs[type]?.color ?? '#9CA3AF';
}

export function getNodeIcon(type: NodeType): { d: string; color: string } {
  return iconDefs[type];
}
