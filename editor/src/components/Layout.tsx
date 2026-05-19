import { useState, useEffect, useRef } from 'react';
import { useNodesState, useEdgesState, useReactFlow } from '@xyflow/react';
import type { Node, Edge } from '@xyflow/react';

import { LibraryPanel } from './LibraryPanel';
import { Canvas, layoutFlowNodes, flowEdgeToReactEdge, flowNodeToReactNode } from './Canvas';
import { PropertiesPanel } from './PropertiesPanel';
import { useFlowStore } from '../stores/flow-store';
import { t } from '../i18n';
import type { FlowNode } from '../types/flow';

export function Layout() {
  const { currentFlow, nodes: flowNodes, edges: flowEdges, selectedNodeId, addNode, dirty, setDirty, validate, validationResult, undo, redo, importFlowFromJson } = useFlowStore();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { fitView, zoomIn, zoomOut } = useReactFlow();
  const [libraryVisible, setLibraryVisible] = useState(true);
  const [showRunDialog, setShowRunDialog] = useState(false);
  const dropPositionRef = useRef<{ x: number; y: number } | null>(null);

  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  const prevFlowIdRef = useRef<string | null>(null);
  const prevNodeCountRef = useRef(0);
  const prevEdgeCountRef = useRef(0);

  const i18n = t();

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
        const dropPos = dropPositionRef.current;
        dropPositionRef.current = null;
        const newReactNodes = newFlowNodes.map((fn, i) => {
          const rn = flowNodeToReactNode(fn, fn.id === selectedNodeId);
          rn.position = dropPos
            ? { x: dropPos.x + i * 200, y: dropPos.y }
            : { x: 100 + i * 200, y: 100 };
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
        data: {
          ...n.data,
          selected: n.id === selectedNodeId,
        },
      }))
    );
  }, [selectedNodeId, setNodes]);

  const handleSave = () => {
    const flow = currentFlow;
    if (!flow) return;
    const exportFlow = { ...flow, nodes, edges };
    const json = JSON.stringify(exportFlow, null, 2);
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${flow.metadata.name}.json`;
    a.click();
    URL.revokeObjectURL(url);
    setDirty(false);
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.ctrlKey || e.metaKey) {
        switch (e.key) {
          case 's':
            e.preventDefault();
            handleSave();
            break;
          case 'z':
            e.preventDefault();
            if (e.shiftKey) {
              redo();
            } else {
              undo();
            }
            break;
          case 'd':
            e.preventDefault();
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
  }, [currentFlow, nodes, edges, zoomIn, zoomOut, fitView]);

  const handleAutoLayout = () => {
    const layoutedNodes = layoutFlowNodes(flowNodes, flowEdges, selectedNodeId);
    setNodes(layoutedNodes);
    setTimeout(() => fitView({ padding: 0.2, duration: 300 }), 50);
  };

  const handleDropNode = (node: FlowNode, position: { x: number; y: number }) => {
    dropPositionRef.current = position;
    addNode(node);
  };

  const handleValidate = () => {
    validate();
  };

  const handleRun = () => {
    setShowRunDialog(true);
  };

  const handleImport = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (ev) => {
      const json = ev.target?.result as string;
      if (json) {
        const success = importFlowFromJson(json);
        if (success) {
          setTimeout(() => fitView({ padding: 0.2, duration: 300 }), 100);
        }
      }
    };
    reader.readAsText(file);
    e.target.value = '';
  };

  const flowName = currentFlow?.metadata.name ?? i18n.app.title;
  const nodeCount = flowNodes.length;
  const edgeCount = flowEdges.length;

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
    <div
      style={{
        width: '100vw',
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        background: '#fff',
      }}
    >
      <header
        style={{
          height: 44,
          borderBottom: '1px solid #E5E7EB',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 16px',
          background: '#fff',
          flexShrink: 0,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontWeight: 600, fontSize: 14 }}>{i18n.app.title}</span>
          <span style={{ color: '#6B7280', fontSize: 13 }}>—</span>
          <span style={{ fontSize: 13 }}>{flowName}</span>
          {dirty && <span style={{ color: '#F59E0B', fontSize: 11 }}>●</span>}
        </div>
        <div style={{ display: 'flex', gap: 6 }}>
          <button onClick={handleAutoLayout} style={{ ...btnBase, border: '1px solid #D1D5DB', background: '#fff' }}>
            {i18n.header.autoLayout}
          </button>
          <button onClick={handleImport} style={{ ...btnBase, border: '1px solid #D1D5DB', background: '#fff' }}>
            导入
          </button>
          <button onClick={handleValidate} style={{ ...btnBase, border: '1px solid #D1D5DB', background: '#fff' }}>
            {i18n.header.validate}
          </button>
          <button onClick={handleRun} style={{ ...btnBase, border: '1px solid #D1D5DB', background: '#fff' }}>
            {i18n.header.run}
          </button>
          <button onClick={handleSave} style={{ ...btnBase, border: '1px solid #3B82F6', background: '#3B82F6', color: '#fff' }}>
            {i18n.header.save}
          </button>
        </div>
        <input
          ref={fileInputRef}
          type="file"
          accept=".json"
          style={{ display: 'none' }}
          onChange={handleFileChange}
        />
      </header>

      <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        <LibraryPanel
          visible={libraryVisible}
          onToggle={() => setLibraryVisible((v) => !v)}
        />
        <Canvas
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onDropNode={handleDropNode}
        />
        <PropertiesPanel />
      </div>

      {validationResult && (
        <div
          style={{
            position: 'fixed',
            bottom: 40,
            right: 340,
            background: validationResult.valid ? '#F0FDF4' : '#FEF2F2',
            border: `1px solid ${validationResult.valid ? '#86EFAC' : '#FCA5A5'}`,
            borderRadius: 6,
            padding: '8px 12px',
            fontSize: 12,
            maxWidth: 300,
            zIndex: 100,
            boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
          }}
        >
          <div style={{ fontWeight: 600, marginBottom: 4 }}>
            {validationResult.valid ? i18n.validate.pass : i18n.validate.fail}
          </div>
          {validationResult.errors.map((e, i) => (
            <div key={i} style={{ color: '#DC2626' }}>• {e}</div>
          ))}
          {validationResult.warnings.map((w, i) => (
            <div key={i} style={{ color: '#D97706' }}>• {w}</div>
          ))}
        </div>
      )}

      {showRunDialog && (
        <div
          style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.3)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 }}
          onClick={() => setShowRunDialog(false)}
        >
          <div
            style={{ background: '#fff', borderRadius: 8, padding: 24, width: 400, boxShadow: '0 4px 24px rgba(0,0,0,0.15)' }}
            onClick={(e) => e.stopPropagation()}
          >
            <div style={{ fontWeight: 600, fontSize: 16, marginBottom: 16 }}>{i18n.run.title}</div>
            <div style={{ padding: '16px 0', color: '#6B7280', fontSize: 13, textAlign: 'center' }}>
              {i18n.run.notAvailable}
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 16 }}>
              <button
                onClick={() => setShowRunDialog(false)}
                style={{ padding: '6px 16px', borderRadius: 4, border: '1px solid #D1D5DB', background: '#fff', cursor: 'pointer', fontSize: 12 }}
              >
                {i18n.run.cancel}
              </button>
            </div>
          </div>
        </div>
      )}

      <footer
        style={{
          height: 28,
          borderTop: '1px solid #E5E7EB',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 16px',
          background: '#F9FAFB',
          fontSize: 11,
          color: '#6B7280',
          flexShrink: 0,
        }}
      >
        <span>
          {validationResult && !validationResult.valid && (
            <span style={{ color: '#DC2626', marginRight: 8 }}>
              ❌ {validationResult.errors.length} errors
            </span>
          )}
          {i18n.app.ready} | {i18n.footer.nodes}: {nodeCount} | {i18n.footer.edges}: {edgeCount}
        </span>
        <span>{i18n.app.version}</span>
      </footer>
    </div>
  );
}
