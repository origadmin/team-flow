import { create } from 'zustand';
import type { Flow, FlowNode, FlowEdge, FlowMetadata, FlowConfig, EdgeCondition } from '../types/flow';
import {
  fetchProjectFlowJSON,
  fetchTeamFlowJSON,
  validateFlow as apiValidateFlow,
  validateTeam as apiValidateTeam,
  saveProjectFlowJSON,
  type ValidationResult,
  type TeamValidationResult,
} from '../api/editor-api';

type EditorMode = 'browse' | 'edit';

interface EditorStore {
  currentFlow: Flow | null;
  currentFlowId: string | null;
  currentFlowSource: 'project' | 'preset' | null;
  nodes: FlowNode[];
  edges: FlowEdge[];
  selectedNodeId: string | null;
  selectedEdgeId: string | null;
  dirty: boolean;
  mode: EditorMode;
  validationResult: ValidationResult | null;
  teamValidationResult: TeamValidationResult | null;
  loading: boolean;
  loadError: string | null;
  undoStack: [FlowNode[], FlowEdge[]][];
  redoStack: [FlowNode[], FlowEdge[]][];

  loadProjectFlow: (flowId: string) => Promise<void>;
  loadPresetFlow: (teamId: string, flowId: string) => Promise<void>;
  loadFlow: (flow: Flow) => void;
  setMode: (mode: EditorMode) => void;
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
  validateCurrentFlow: () => Promise<ValidationResult | null>;
  validateTeam: () => Promise<TeamValidationResult | null>;
  saveFlow: () => Promise<void>;
  setDirty: (dirty: boolean) => void;
  undo: () => void;
  redo: () => void;
}

function pushUndo(state: EditorStore): void {
  const snapshot: [FlowNode[], FlowEdge[]] = [[...state.nodes], [...state.edges]];
  state.undoStack = [...state.undoStack, snapshot];
  if (state.undoStack.length > 50) state.undoStack = state.undoStack.slice(-50);
  state.redoStack = [];
}

function convertApiFlowToFlow(raw: Record<string, unknown>): Flow {
  const metadata = (raw.metadata ?? {}) as Record<string, unknown>;
  const nodes = (raw.nodes ?? []) as FlowNode[];
  const edges = (raw.edges ?? []) as FlowEdge[];

  const flowEdges = edges.map((e: FlowEdge | { from: string; to: string; label?: string; condition?: string }) => {
    if ('from' in e && 'to' in e && !('id' in e)) {
      const rawEdge = e as { from: string; to: string; label?: string; condition?: string };
      return {
        id: `edge-${rawEdge.from}-${rawEdge.to}`,
        from: rawEdge.from,
        to: rawEdge.to,
        type: 'sequential' as const,
        conditions: rawEdge.label ? [{ expression: rawEdge.label }] : undefined,
      } as FlowEdge;
    }
    return e as FlowEdge;
  });

  const flowNodes = nodes.map((n: FlowNode | { id: string; type?: string; role?: string; label?: string; persona?: string; traits?: string[]; guidance?: string }) => {
    if ('role' in n && !('name' in n)) {
      const rawNode = n as { id: string; type?: string; role?: string; label?: string; persona?: string; traits?: string[]; guidance?: string };
      return {
        id: rawNode.id,
        type: (rawNode.type ?? 'phase') as FlowNode['type'],
        name: rawNode.label ?? rawNode.id,
        description: rawNode.persona ?? '',
        config: { role: rawNode.role ?? '' },
      } as FlowNode;
    }
    return n as FlowNode;
  });

  return {
    version: 'v3',
    metadata: {
      name: (metadata.name ?? 'unnamed') as string,
      description: metadata.description as string | undefined,
      domain: metadata.domain as string | undefined,
    },
    config: raw.config as FlowConfig | undefined,
    components: raw.components as Record<string, unknown> | undefined,
    nodes: flowNodes,
    edges: flowEdges,
  };
}

export const useEditorStore = create<EditorStore>((set, get) => ({
  currentFlow: null,
  currentFlowId: null,
  currentFlowSource: null,
  nodes: [],
  edges: [],
  selectedNodeId: null,
  selectedEdgeId: null,
  dirty: false,
  mode: 'browse',
  validationResult: null,
  teamValidationResult: null,
  loading: false,
  loadError: null,
  undoStack: [],
  redoStack: [],

  loadProjectFlow: async (flowId: string) => {
    set({ loading: true, loadError: null });
    try {
      const raw = await fetchProjectFlowJSON(flowId);
      const flow = convertApiFlowToFlow(raw);
      set((state) => {
        pushUndo(state);
        return {
          currentFlow: flow,
          currentFlowId: flowId,
          currentFlowSource: 'project',
          nodes: flow.nodes,
          edges: flow.edges,
          selectedNodeId: null,
          selectedEdgeId: null,
          dirty: false,
          validationResult: null,
          mode: 'browse',
          loading: false,
          undoStack: state.undoStack,
          redoStack: state.redoStack,
        };
      });
      get().validateCurrentFlow();
    } catch (e) {
      set({ loading: false, loadError: e instanceof Error ? e.message : 'Failed to load flow' });
    }
  },

  loadPresetFlow: async (teamId: string, flowId: string) => {
    set({ loading: true, loadError: null });
    try {
      const raw = await fetchTeamFlowJSON(teamId, flowId);
      const flow = convertApiFlowToFlow(raw);
      set((state) => {
        pushUndo(state);
        return {
          currentFlow: flow,
          currentFlowId: flowId,
          currentFlowSource: 'preset',
          nodes: flow.nodes,
          edges: flow.edges,
          selectedNodeId: null,
          selectedEdgeId: null,
          dirty: false,
          validationResult: null,
          mode: 'browse',
          loading: false,
          undoStack: state.undoStack,
          redoStack: state.redoStack,
        };
      });
      get().validateCurrentFlow();
    } catch (e) {
      set({ loading: false, loadError: e instanceof Error ? e.message : 'Failed to load flow' });
    }
  },

  loadFlow: (flow: Flow) =>
    set((state) => {
      pushUndo(state);
      return {
        currentFlow: flow,
        currentFlowId: flow.metadata.name,
        currentFlowSource: 'project',
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

  setMode: (mode) => set({ mode }),

  selectNode: (id) => set({ selectedNodeId: id, selectedEdgeId: null }),
  selectEdge: (id) => set({ selectedEdgeId: id, selectedNodeId: null }),

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

  validateCurrentFlow: async () => {
    const { currentFlowId, currentFlow, currentFlowSource } = get();
    if (!currentFlowId || !currentFlow || currentFlowSource !== 'project') return null;
    try {
      const result = await apiValidateFlow(currentFlowId, currentFlow);
      set({ validationResult: result });
      return result;
    } catch {
      return null;
    }
  },

  validateTeam: async () => {
    try {
      const result = await apiValidateTeam();
      set({ teamValidationResult: result });
      return result;
    } catch {
      return null;
    }
  },

  saveFlow: async () => {
    const { currentFlowId, currentFlow, currentFlowSource, mode } = get();
    if (!currentFlowId || !currentFlow || currentFlowSource !== 'project') return;
    if (mode !== 'edit') return;

    const result = await get().validateCurrentFlow();
    if (result && !result.valid) return;

    try {
      await saveProjectFlowJSON(currentFlowId, currentFlow);
      set({ dirty: false });
    } catch (e) {
      console.error('Save failed:', e);
    }
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
}));
