import { create } from 'zustand';
import type { Flow, FlowNode, FlowEdge, FlowMetadata, FlowConfig, EdgeCondition } from '../types/flow';

interface ValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

interface FlowStore {
  currentFlow: Flow | null;
  flows: Flow[];
  nodes: FlowNode[];
  edges: FlowEdge[];
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  dirty: boolean;
  validationResult: ValidationResult | null;
  undoStack: [FlowNode[], FlowEdge[]][];
  redoStack: [FlowNode[], FlowEdge[]][];

  loadFlow: (flow: Flow) => void;
  createFlow: (name: string, domain: string, template?: Flow | null) => void;
  selectNode: (id: string | null) => void;
  selectEdge: (id: string | null) => void;
  addNode: (node: FlowNode) => void;
  updateNode: (id: string, data: Partial<Omit<FlowNode, 'config'>> & { config?: unknown }) => void;
  removeNode: (id: string) => void;
  addEdge: (edge: FlowEdge) => void;
  updateEdge: (id: string, data: Partial<FlowEdge>) => void;
  removeEdge: (id: string) => void;
  updateFlowMetadata: (data: Partial<FlowMetadata>) => void;
  updateFlowConfig: (data: Partial<FlowConfig>) => void;
  updateFlowVariables: (variables: Record<string, string>) => void;
  validate: () => ValidationResult;
  setDirty: (dirty: boolean) => void;
  undo: () => void;
  redo: () => void;
  exportFlowAsJson: () => void;
  importFlowFromJson: (json: string) => boolean;
}

function pushUndo(state: FlowStore): void {
  const snapshot: [FlowNode[], FlowEdge[]] = [[...state.nodes], [...state.edges]];
  state.undoStack = [...state.undoStack, snapshot];
  if (state.undoStack.length > 50) state.undoStack = state.undoStack.slice(-50);
  state.redoStack = [];
}

export const useFlowStore = create<FlowStore>((set, get) => ({
  currentFlow: null,
  flows: [],
  nodes: [],
  edges: [],
  selectedNodeId: null,
  selectedEdgeId: null,
  dirty: false,
  validationResult: null,
  undoStack: [],
  redoStack: [],

  loadFlow: (flow) =>
    set((state) => {
      pushUndo(state);
      return {
        currentFlow: flow,
        nodes: flow.nodes,
        edges: flow.edges,
        selectedNodeId: null,
        selectedEdgeId: null,
        dirty: false,
        validationResult: null,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    }),

  createFlow: (name, domain, template) => {
    const newFlow: Flow = template
      ? {
          ...template,
          version: 'v3',
          metadata: { ...template.metadata, name, domain },
          nodes: template.nodes.map((n) => ({ ...n })),
          edges: template.edges.map((e) => ({ ...e })),
        }
      : {
          version: 'v3',
          metadata: { name, domain },
          config: { task_type: name, auto_dispatch: true },
          nodes: [],
          edges: [],
        };
    set((state) => {
      pushUndo(state);
      return {
        flows: [...state.flows, newFlow],
        currentFlow: newFlow,
        nodes: newFlow.nodes,
        edges: newFlow.edges,
        selectedNodeId: null,
        selectedEdgeId: null,
        dirty: false,
        validationResult: null,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    });
  },

  selectNode: (id) =>
    set({ selectedNodeId: id, selectedEdgeId: null }),

  selectEdge: (id) =>
    set({ selectedEdgeId: id, selectedNodeId: null }),

  addNode: (node) =>
    set((state) => {
      pushUndo(state);
      return { nodes: [...state.nodes, node], dirty: true, undoStack: state.undoStack, redoStack: state.redoStack };
    }),

  updateNode: (id, data) =>
    set((state) => {
      pushUndo(state);
      return {
        nodes: state.nodes.map((n) => (n.id === id ? ({ ...n, ...data } as FlowNode) : n)),
        dirty: true,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    }),

  removeNode: (id) =>
    set((state) => {
      pushUndo(state);
      return {
        nodes: state.nodes.filter((n) => n.id !== id),
        edges: state.edges.filter((e) => e.from !== id && e.to !== id),
        selectedNodeId: state.selectedNodeId === id ? null : state.selectedNodeId,
        dirty: true,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    }),

  addEdge: (edge) =>
    set((state) => {
      pushUndo(state);
      return { edges: [...state.edges, edge], dirty: true, undoStack: state.undoStack, redoStack: state.redoStack };
    }),

  updateEdge: (id, data) =>
    set((state) => {
      pushUndo(state);
      return {
        edges: state.edges.map((e) => (e.id === id ? { ...e, ...data } : e)),
        dirty: true,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    }),

  removeEdge: (id) =>
    set((state) => {
      pushUndo(state);
      return {
        edges: state.edges.filter((e) => e.id !== id),
        selectedEdgeId: state.selectedEdgeId === id ? null : state.selectedEdgeId,
        dirty: true,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    }),

  updateFlowMetadata: (data) =>
    set((state) => {
      pushUndo(state);
      return {
        currentFlow: state.currentFlow
          ? { ...state.currentFlow, metadata: { ...state.currentFlow.metadata, ...data } }
          : null,
        dirty: true,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    }),

  updateFlowConfig: (data) =>
    set((state) => {
      pushUndo(state);
      return {
        currentFlow: state.currentFlow
          ? { ...state.currentFlow, config: { ...state.currentFlow.config, ...data } }
          : null,
        dirty: true,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    }),

  updateFlowVariables: (variables) =>
    set((state) => {
      pushUndo(state);
      return {
        currentFlow: state.currentFlow
          ? { ...state.currentFlow, variables }
          : null,
        dirty: true,
        undoStack: state.undoStack,
        redoStack: state.redoStack,
      };
    }),

  validate: () => {
    const { nodes, edges } = get();
    const errors: string[] = [];
    const warnings: string[] = [];

    const nodeIds = new Set(nodes.map((n) => n.id));
    const inDegree = new Map<string, number>();
    const adjacency = new Map<string, string[]>();

    for (const n of nodes) {
      inDegree.set(n.id, 0);
      adjacency.set(n.id, []);
    }
    for (const e of edges) {
      if (!nodeIds.has(e.from) || !nodeIds.has(e.to)) continue;
      inDegree.set(e.to, (inDegree.get(e.to) ?? 0) + 1);
      adjacency.get(e.from)?.push(e.to);
    }

    const roots = nodes.filter((n) => (inDegree.get(n.id) ?? 0) === 0);
    if (roots.length === 0) {
      errors.push('没有入度为0的根节点');
    }

    const terminals = nodes.filter((n) => n.type === 'terminal');
    if (terminals.length === 0) {
      errors.push('没有终端节点');
    }

    const reachable = new Set<string>();
    const queue = roots.map((r) => r.id);
    for (const id of queue) reachable.add(id);
    while (queue.length > 0) {
      const current = queue.shift()!;
      for (const next of adjacency.get(current) ?? []) {
        if (!reachable.has(next)) {
          reachable.add(next);
          queue.push(next);
        }
      }
    }

    for (const n of nodes) {
      if (!reachable.has(n.id)) {
        warnings.push(`孤立节点: ${n.id}`);
      }
    }

    for (const t of terminals) {
      if (!reachable.has(t.id)) {
        errors.push(`不可达的终端节点: ${t.id}`);
      }
    }

    const visited = new Set<string>();
    const path = new Set<string>();
    let hasCycle = false;
    function dfs(id: string) {
      if (hasCycle) return;
      visited.add(id);
      path.add(id);
      for (const next of adjacency.get(id) ?? []) {
        if (path.has(next)) {
          hasCycle = true;
          return;
        }
        if (!visited.has(next)) dfs(next);
      }
      path.delete(id);
    }
    for (const r of roots) dfs(r.id);
    if (hasCycle) {
      warnings.push('检测到循环依赖');
    }

    for (const n of nodes) {
      if (!n.name) {
        errors.push(`节点 ${n.id} 缺少名称`);
      }
      if (n.type === 'gate') {
        const gc = n.config as { conditions?: unknown[] } | undefined;
        if (!gc?.conditions || gc.conditions.length === 0) {
          warnings.push(`关卡 ${n.id} 没有条件`);
        }
      }
    }

    const result = { valid: errors.length === 0, errors, warnings };
    set({ validationResult: result });
    return result;
  },

  setDirty: (dirty) => set({ dirty }),

  undo: () => {
    const state = get();
    if (state.undoStack.length === 0) return;
    const prev = state.undoStack[state.undoStack.length - 1]!;
    const current: [FlowNode[], FlowEdge[]] = [state.nodes, state.edges];
    set({
      nodes: prev[0],
      edges: prev[1],
      undoStack: state.undoStack.slice(0, -1),
      redoStack: [...state.redoStack, current],
      dirty: true,
    });
  },

  redo: () => {
    const state = get();
    if (state.redoStack.length === 0) return;
    const next = state.redoStack[state.redoStack.length - 1]!;
    const current: [FlowNode[], FlowEdge[]] = [state.nodes, state.edges];
    set({
      nodes: next[0],
      edges: next[1],
      undoStack: [...state.undoStack, current],
      redoStack: state.redoStack.slice(0, -1),
      dirty: true,
    });
  },

  exportFlowAsJson: () => {
    const { currentFlow, nodes, edges } = get();
    if (!currentFlow) return;
    const flow: Flow = {
      ...currentFlow,
      nodes,
      edges,
    };
    const json = JSON.stringify(flow, null, 2);
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${currentFlow.metadata.name}.json`;
    a.click();
    URL.revokeObjectURL(url);
  },

  importFlowFromJson: (json) => {
    try {
      const flow = JSON.parse(json) as Flow;
      if (!flow.version || !flow.metadata || !flow.nodes || !flow.edges) {
        console.error('Invalid flow format');
        return false;
      }
      set((state) => {
        pushUndo(state);
        return {
          currentFlow: flow,
          nodes: flow.nodes,
          edges: flow.edges,
          selectedNodeId: null,
          selectedEdgeId: null,
          dirty: false,
          validationResult: null,
          undoStack: state.undoStack,
          redoStack: state.redoStack,
        };
      });
      return true;
    } catch (e) {
      console.error('Failed to parse JSON:', e);
      return false;
    }
  },
}));
