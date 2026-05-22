import { useState, useEffect, useRef } from 'react';
import { useNodesState, useEdgesState, useReactFlow } from '@xyflow/react';
import type { Node, Edge } from '@xyflow/react';

import { TeamTree } from './TeamTree';
import { Canvas, layoutFlowNodes, flowEdgeToReactEdge, flowNodeToReactNode } from './Canvas';
import { PropertiesPanel } from './PropertiesPanel';
import { useEditorStore } from '../stores/editor-store';
import type { FlowNode } from '../types/flow';

export function Layout() {
  const {
    currentFlow, nodes: flowNodes, edges: flowEdges,
    selectedNodeId, addNode, dirty, setDirty,
    validationResult, teamValidationResult,
    mode, setMode, saveFlow, validateCurrentFlow, validateTeam,
    undo, redo, currentFlowSource,
  } = useEditorStore();
  const { fitView, zoomIn, zoomOut } = useReactFlow();
  const [treeVisible, setTreeVisible] = useState(true);

  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  const prevFlowIdRef = useRef<string | null>(null);
  const prevNodeCountRef = useRef(0);
  const prevEdgeCountRef = useRef(0);

  useEffect(() => {
    const flowId = currentFlow?.metadata.name ?? null;
    if (flowId !== prevFlowIdRef.current) {
      prevFlowIdRef.current = flowId;
      prevNodeCountRef.current = flowNodes.length;
      prevEdgeCountRef.current = flowEdges.length;
      if (flowNodes.length === 0 && flowEdges.length === 0) {
        setNodes([]);
        setEdges([]);
        return;
      }
      const layoutedNodes = layoutFlowNodes(flowNodes, flowEdges, selectedNodeId);
      setNodes(layoutedNodes);
      const reactEdges = flowEdges.map(flowEdgeToReactEdge);
      setEdges(reactEdges);
      setTimeout(() => fitView({ padding: 0.2, duration: 300 }), 100);
      return;
    }

    const nodeDiff = flowNodes.length - prevNodeCountRef.current;
    const edgeDiff = flowEdges.length - prevEdgeCountRef.current;
    prevNodeCountRef.current = flowNodes.length;
    prevEdgeCountRef.current = flowEdges.length;

    if (nodeDiff > 0) {
      const existingIds = new Set(nodes.map((n) => n.id));
      const newFlowNodes = flowNodes.filter((fn) => !existingIds.has(fn.id));
      if (newFlowNodes.length > 0) {
        const newReactNodes = newFlowNodes.map((fn, i) => {
          const rn = flowNodeToReactNode(fn, fn.id === selectedNodeId);
          rn.position = { x: 100 + i * 200, y: 100 };
          return rn;
        });
        setNodes((nds) => [...nds, ...newReactNodes]);
      }
    }

    if (nodeDiff < 0) {
      const flowNodeIds = new Set(flowNodes.map((n) => n.id));
      setNodes((nds) => nds.filter((n) => flowNodeIds.has(n.id)));
    }

    if (edgeDiff > 0) {
      const existingEdgeIds = new Set(edges.map((e) => e.id));
      const newFlowEdges = flowEdges.filter((fe) => !existingEdgeIds.has(fe.id ?? `edge-${fe.from}-${fe.to}`));
      if (newFlowEdges.length > 0) {
        const newReactEdges = newFlowEdges.map(flowEdgeToReactEdge);
        setEdges((eds) => [...eds, ...newReactEdges]);
      }
    }

    if (edgeDiff < 0) {
      const flowEdgeIds = new Set(flowEdges.map((e) => e.id ?? `edge-${e.from}-${e.to}`));
      setEdges((eds) => eds.filter((e) => flowEdgeIds.has(e.id)));
    }
  }, [currentFlow?.metadata.name, flowNodes, flowEdges, setNodes, setEdges]);

  useEffect(() => {
    setNodes((nds) =>
      nds.map((n) => ({
        ...n,
        selected: n.id === selectedNodeId,
        data: { ...n.data, selected: n.id === selectedNodeId },
      }))
    );
  }, [selectedNodeId, setNodes]);

  const handleAutoLayout = () => {
    const layoutedNodes = layoutFlowNodes(flowNodes, flowEdges, selectedNodeId);
    setNodes(layoutedNodes);
    setTimeout(() => fitView({ padding: 0.2, duration: 300 }), 50);
  };

  const handleDropNode = (node: FlowNode, position: { x: number; y: number }) => {
    addNode(node);
  };

  const handleValidate = () => {
    validateCurrentFlow();
  };

  const handleTeamValidate = () => {
    validateTeam();
  };

  const handleSave = async () => {
    await saveFlow();
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.ctrlKey || e.metaKey) {
        switch (e.key) {
          case 's':
            e.preventDefault();
            if (mode === 'edit') handleSave();
            break;
          case 'z':
            e.preventDefault();
            if (e.shiftKey) redo(); else undo();
            break;
          case '=':
          case '+':
            e.preventDefault();
            zoomIn();
            break;
          case '-':
            e.preventDefault();
            zoomOut();
            break;
          case '0':
            e.preventDefault();
            fitView({ padding: 0.2, duration: 300 });
            break;
        }
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [mode, zoomIn, zoomOut, fitView]);

  const flowName = currentFlow?.metadata.name ?? 'Flow Editor';
  const nodeCount = flowNodes.length;
  const edgeCount = flowEdges.length;
  const isPreset = currentFlowSource === 'preset';

  const btnBase: React.CSSProperties = {
    padding: '4px 12px',
    borderRadius: 4,
    cursor: 'pointer',
    fontSize: 12,
    display: 'flex',
    alignItems: 'center',
    gap: 4,
  };

  return (
    <div style={{ width: '100vw', height: '100vh', display: 'flex', flexDirection: 'column', fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif', background: '#fff' }}>
      <header style={{ height: 44, borderBottom: '1px solid #E5E7EB', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 16px', background: '#fff', flexShrink: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontWeight: 600, fontSize: 14 }}>Flow Editor</span>
          <span style={{ color: '#6B7280', fontSize: 13 }}>—</span>
          <span style={{ fontSize: 13 }}>{flowName}</span>
          {isPreset && <span style={{ fontSize: 10, color: '#8B5CF6', background: '#EDE9FE', padding: '1px 6px', borderRadius: 3 }}>PRESET</span>}
          {dirty && <span style={{ color: '#F59E0B', fontSize: 11 }}>●</span>}
        </div>
        <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
          <button
            onClick={() => setMode(mode === 'browse' ? 'edit' : 'browse')}
            style={{
              ...btnBase,
              border: mode === 'edit' ? '1px solid #F59E0B' : '1px solid #10B981',
              background: mode === 'edit' ? '#FEF3C7' : '#ECFDF5',
              color: mode === 'edit' ? '#92400E' : '#065F46',
              fontWeight: 600,
            }}
          >
            {mode === 'browse' ? '📖 Browse' : '✏️ Edit'}
          </button>
          <button onClick={handleAutoLayout} style={{ ...btnBase, border: '1px solid #D1D5DB', background: '#fff' }}>
            Auto Layout
          </button>
          <button onClick={handleValidate} style={{ ...btnBase, border: '1px solid #D1D5DB', background: '#fff' }}>
            Validate
          </button>
          <button onClick={handleTeamValidate} style={{ ...btnBase, border: '1px solid #D1D5DB', background: '#fff' }}>
            Team Check
          </button>
          {mode === 'edit' && !isPreset && (
            <button onClick={handleSave} style={{ ...btnBase, border: '1px solid #3B82F6', background: '#3B82F6', color: '#fff' }}>
              Save
            </button>
          )}
        </div>
      </header>

      <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        {treeVisible ? (
          <TeamTree />
        ) : (
          <div style={{ width: 40, borderRight: '1px solid #E5E7EB', background: '#F9FAFB', display: 'flex', flexDirection: 'column', alignItems: 'center', paddingTop: 8 }}>
            <button onClick={() => setTreeVisible(true)} style={{ background: 'none', border: '1px solid #D1D5DB', borderRadius: 4, padding: '4px 8px', cursor: 'pointer', fontSize: 16 }}>
              ≡
            </button>
          </div>
        )}
        <Canvas
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onDropNode={mode === 'edit' ? handleDropNode : undefined}
          readOnly={mode === 'browse'}
        />
        <PropertiesPanel />
      </div>

      {validationResult && (
        <div style={{ position: 'fixed', bottom: 40, right: 340, background: validationResult.valid ? '#F0FDF4' : '#FEF2F2', border: `1px solid ${validationResult.valid ? '#86EFAC' : '#FCA5A5'}`, borderRadius: 6, padding: '8px 12px', fontSize: 12, maxWidth: 300, zIndex: 100, boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <div style={{ fontWeight: 600, marginBottom: 4 }}>
            {validationResult.valid ? '✅ Validation Passed' : '❌ Validation Failed'}
          </div>
          {validationResult.errors.map((e, i) => (
            <div key={i} style={{ color: '#DC2626' }}>• [{e.field}] {e.msg}</div>
          ))}
          {validationResult.warnings.map((w, i) => (
            <div key={i} style={{ color: '#D97706' }}>• [{w.field}] {w.msg}</div>
          ))}
        </div>
      )}

      {teamValidationResult && (
        <div style={{ position: 'fixed', bottom: 40, left: 260, background: teamValidationResult.valid ? '#F0FDF4' : '#FEF2F2', border: `1px solid ${teamValidationResult.valid ? '#86EFAC' : '#FCA5A5'}`, borderRadius: 6, padding: '8px 12px', fontSize: 12, maxWidth: 300, zIndex: 100, boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
          <div style={{ fontWeight: 600, marginBottom: 4 }}>
            {teamValidationResult.valid ? '✅ Team Check Passed' : '❌ Team Check Failed'}
          </div>
          {teamValidationResult.errors.map((e, i) => (
            <div key={i} style={{ color: '#DC2626' }}>• [{e.field}] {e.msg}</div>
          ))}
          {teamValidationResult.warnings.map((w, i) => (
            <div key={i} style={{ color: '#D97706' }}>• [{w.field}] {w.msg}</div>
          ))}
        </div>
      )}

      <footer style={{ height: 28, borderTop: '1px solid #E5E7EB', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 16px', background: '#F9FAFB', fontSize: 11, color: '#6B7280', flexShrink: 0 }}>
        <span>
          {validationResult && !validationResult.valid && (
            <span style={{ color: '#DC2626', marginRight: 8 }}>❌ {validationResult.errors.length} errors</span>
          )}
          {validationResult && validationResult.warnings.length > 0 && (
            <span style={{ color: '#D97706', marginRight: 8 }}>⚠️ {validationResult.warnings.length} warnings</span>
          )}
          {mode === 'browse' ? '📖 Browse' : '✏️ Edit'} | Nodes: {nodeCount} | Edges: {edgeCount}
        </span>
        <span>v3 Editor</span>
      </footer>
    </div>
  );
}
