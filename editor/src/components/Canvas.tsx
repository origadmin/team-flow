import { useCallback, useRef } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useReactFlow,
  type OnNodesChange,
  type OnEdgesChange,
  type OnConnect,
  type Connection,
  type Node,
  type Edge,
  type ReactFlowInstance,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { PhaseNode } from './nodes/PhaseNode';
import { GateNode } from './nodes/GateNode';
import { TerminalNode } from './nodes/TerminalNode';
import { StartNode } from './nodes/StartNode';
import { ConditionalEdge } from './edges/ConditionalEdge';
import { useFlowStore } from '../stores/flow-store';
import type { FlowNode, FlowEdge, StartConfig, PhaseConfig, GateConfig, TerminalConfig, StartNodeData, PhaseNodeData, GateNodeData, TerminalNodeData, ConditionalEdgeData, NodeType } from '../types/flow';
import { getNodeColor } from './nodes/NodeIcon';
import { generateNodeId } from '../utils/id';

const nodeTypes = {
  start: StartNode,
  phase: PhaseNode,
  gate: GateNode,
  terminal: TerminalNode,
};

const edgeTypes = {
  conditional: ConditionalEdge,
};

interface CanvasProps {
  nodes: Node[];
  edges: Edge[];
  onNodesChange: OnNodesChange;
  onEdgesChange: OnEdgesChange;
  onDropNode?: (node: FlowNode, position: { x: number; y: number }) => void;
}

export function Canvas({ nodes, edges, onNodesChange, onEdgesChange, onDropNode }: CanvasProps) {
  const { selectNode, selectEdge, addEdge } = useFlowStore();
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const { screenToFlowPosition } = useReactFlow();

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      selectNode(node.id);
    },
    [selectNode],
  );

  const onEdgeClick = useCallback(
    (_: React.MouseEvent, edge: Edge) => {
      selectEdge(edge.id);
    },
    [selectEdge],
  );

  const onPaneClick = useCallback(() => {
    selectNode(null);
    selectEdge(null);
  }, [selectNode, selectEdge]);

  const onConnect = useCallback(
    (connection: Connection) => {
      if (connection.source && connection.target) {
        const newEdge: FlowEdge = {
          id: `edge-${connection.source}-${connection.target}`,
          from: connection.source,
          to: connection.target,
          type: 'sequential',
        };
        addEdge(newEdge);
      }
    },
    [addEdge],
  );

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      const nodeType = event.dataTransfer.getData('application/reactflow') as NodeType;
      if (!nodeType || !onDropNode) return;

      const position = screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      const existingIds = useFlowStore.getState().nodes.map((n) => n.id);
      const id = generateNodeId(existingIds);
      const newNode: FlowNode = {
        id,
        type: nodeType,
        name: `${nodeType.charAt(0).toUpperCase() + nodeType.slice(1)} Node`,
        description: `New ${nodeType} node`,
        config: getDefaultConfig(nodeType),
      };

      onDropNode(newNode, position);
    },
    [screenToFlowPosition, onDropNode],
  );

  return (
    <div ref={reactFlowWrapper} style={{ flex: 1, height: '100%' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onEdgeClick={onEdgeClick}
        onPaneClick={onPaneClick}
        onConnect={onConnect}
        onDragOver={onDragOver}
        onDrop={onDrop}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        defaultEdgeOptions={{ type: 'conditional' }}
        minZoom={0.3}
        maxZoom={2}
        connectionLineStyle={{ stroke: '#3B82F6', strokeWidth: 2 }}
        snapToGrid
        snapGrid={[16, 16]}
        deleteKeyCode="Delete"
      >
        <Background color="#E5E7EB" gap={20} size={1} />
        <Controls position="bottom-right" />
        <MiniMap
          nodeColor={(node) => getNodeColor(node.type as NodeType)}
          maskColor="rgba(0,0,0,0.1)"
          style={{ background: '#F9FAFB' }}
        />
      </ReactFlow>
    </div>
  );
}

function getDefaultConfig(nodeType: NodeType): FlowNode['config'] {
  switch (nodeType) {
    case 'start':
      return {};
    case 'phase':
      return { role: '' } as PhaseConfig;
    case 'gate':
      return { conditions: [] } as GateConfig;
    case 'terminal':
      return { status: 'success' as const } as TerminalConfig;
    default:
      return { role: '' } as PhaseConfig;
  }
}

export function flowNodeToReactNode(fn: FlowNode, selected: boolean): Node {
  const base = {
    id: fn.id,
    position: { x: 0, y: 0 },
    selected,
  };

  switch (fn.type) {
    case 'start':
      return {
        ...base,
        type: 'start',
        data: {
          flowNode: fn,
          config: (fn.config ?? {}) as StartConfig,
          selected,
        } satisfies StartNodeData,
      };
    case 'phase':
      return {
        ...base,
        type: 'phase',
        data: {
          flowNode: fn,
          config: (fn.config ?? { role: '' }) as PhaseConfig,
          selected,
        } satisfies PhaseNodeData,
      };
    case 'gate':
      return {
        ...base,
        type: 'gate',
        data: {
          flowNode: fn,
          config: (fn.config ?? { conditions: [] }) as GateConfig,
          selected,
        } satisfies GateNodeData,
      };
    case 'terminal':
      return {
        ...base,
        type: 'terminal',
        data: {
          flowNode: fn,
          config: (fn.config ?? { status: 'success' as const }) as TerminalConfig,
          selected,
        } satisfies TerminalNodeData,
      };
    default:
      return {
        ...base,
        type: 'phase',
        data: {
          flowNode: fn,
          config: { role: '' } as PhaseConfig,
          selected,
        } satisfies PhaseNodeData,
      };
  }
}

export function flowEdgeToReactEdge(fe: FlowEdge): Edge {
  return {
    id: fe.id ?? `edge-${fe.from}-${fe.to}`,
    source: fe.from,
    target: fe.to,
    type: 'conditional',
    data: {
      edgeType: fe.type ?? 'sequential',
      conditions: fe.conditions,
    } satisfies ConditionalEdgeData,
  };
}

const NODE_H_GAP = 300;
const NODE_V_GAP = 180;

export function layoutFlowNodes(flowNodes: FlowNode[], flowEdges: FlowEdge[], selectedNodeId: string | null): Node[] {
  if (flowNodes.length === 0) return [];

  const nodeIds = new Set(flowNodes.map((n) => n.id));
  const validEdges = flowEdges.filter((e) => nodeIds.has(e.from) && nodeIds.has(e.to));

  if (validEdges.length === 0) {
    return flowNodes.map((fn, i) => {
      const node = flowNodeToReactNode(fn, fn.id === selectedNodeId);
      node.position = { x: (i % 5) * NODE_H_GAP, y: Math.floor(i / 5) * NODE_V_GAP };
      return node;
    });
  }

  const adjWithIdx = new Map<string, { to: string; idx: number }[]>();
  for (const n of flowNodes) adjWithIdx.set(n.id, []);
  for (let i = 0; i < validEdges.length; i++) {
    const e = validEdges[i]!;
    adjWithIdx.get(e.from)?.push({ to: e.to, idx: i });
  }

  const W = 0, G = 1, B = 2;
  const color = new Map<string, number>();
  for (const n of flowNodes) color.set(n.id, W);
  const backEdgeIdxSet = new Set<number>();

  function dfs(u: string) {
    color.set(u, G);
    for (const { to, idx } of (adjWithIdx.get(u) ?? [])) {
      if ((color.get(to) ?? W) === G) {
        backEdgeIdxSet.add(idx);
      } else if ((color.get(to) ?? W) === W) {
        dfs(to);
      }
    }
    color.set(u, B);
  }

  const inDeg = new Map<string, number>();
  for (const n of flowNodes) inDeg.set(n.id, 0);
  for (const e of validEdges) inDeg.set(e.to, (inDeg.get(e.to) ?? 0) + 1);
  for (const n of flowNodes) {
    if ((inDeg.get(n.id) ?? 0) === 0 && color.get(n.id) === W) dfs(n.id);
  }
  for (const n of flowNodes) {
    if (color.get(n.id) === W) dfs(n.id);
  }

  const dagEdges = validEdges.filter((_, i) => !backEdgeIdxSet.has(i));

  const dagAdj = new Map<string, string[]>();
  const dagInDeg = new Map<string, number>();
  for (const n of flowNodes) {
    dagAdj.set(n.id, []);
    dagInDeg.set(n.id, 0);
  }
  for (const e of dagEdges) {
    dagAdj.get(e.from)?.push(e.to);
    dagInDeg.set(e.to, (dagInDeg.get(e.to) ?? 0) + 1);
  }

  const levels = new Map<string, number>();
  for (const n of flowNodes) levels.set(n.id, 0);

  const q: string[] = [];
  for (const n of flowNodes) {
    if ((dagInDeg.get(n.id) ?? 0) === 0) q.push(n.id);
  }

  while (q.length > 0) {
    const u = q.shift()!;
    for (const v of (dagAdj.get(u) ?? [])) {
      const nl = (levels.get(u) ?? 0) + 1;
      if (nl > (levels.get(v) ?? 0)) levels.set(v, nl);
      const nd = (dagInDeg.get(v) ?? 1) - 1;
      dagInDeg.set(v, nd);
      if (nd === 0) q.push(v);
    }
  }

  const mainPath = new Set<string>();
  let bestTerminal: string | null = null;
  let bestLevel = -1;
  for (const n of flowNodes) {
    if (n.type === 'terminal') {
      const lv = levels.get(n.id) ?? 0;
      if (lv > bestLevel) { bestLevel = lv; bestTerminal = n.id; }
    }
  }
  if (bestTerminal) {
    const revAdj = new Map<string, string[]>();
    for (const n of flowNodes) revAdj.set(n.id, []);
    for (const e of dagEdges) revAdj.get(e.to)?.push(e.from);
    let cur = bestTerminal;
    while (cur) {
      mainPath.add(cur);
      const preds = revAdj.get(cur) ?? [];
      let best: string | null = null;
      let bestLv = -1;
      for (const p of preds) {
        const lv = levels.get(p) ?? 0;
        if (lv > bestLv) { bestLv = lv; best = p; }
      }
      cur = best ?? '';
    }
  }

  const levelGroups = new Map<number, string[]>();
  for (const n of flowNodes) {
    const lv = levels.get(n.id) ?? 0;
    if (!levelGroups.has(lv)) levelGroups.set(lv, []);
    levelGroups.get(lv)!.push(n.id);
  }

  const maxLevel = Math.max(...levels.values(), 0);
  const positions = new Map<string, { x: number; y: number }>();

  for (let lv = 0; lv <= maxLevel; lv++) {
    const ids = levelGroups.get(lv) ?? [];
    const mainIds = ids.filter((id) => mainPath.has(id));
    const branchIds = ids.filter((id) => !mainPath.has(id));
    for (const id of mainIds) {
      positions.set(id, { x: lv * NODE_H_GAP, y: 0 });
    }
    for (let i = 0; i < branchIds.length; i++) {
      positions.set(branchIds[i]!, {
        x: lv * NODE_H_GAP,
        y: (i + 1) * NODE_V_GAP,
      });
    }
  }

  return flowNodes.map((fn) => {
    const pos = positions.get(fn.id) ?? { x: 0, y: 0 };
    const node = flowNodeToReactNode(fn, fn.id === selectedNodeId);
    node.position = pos;
    return node;
  });
}
