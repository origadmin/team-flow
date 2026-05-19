import { useState } from 'react';
import { t } from '../i18n';
import { useFlowStore } from '../stores/flow-store';
import type { NodeType } from '../types/flow';
import { sampleDevFlow, sampleGameDesignFlow, sampleNovelFlow } from '../data/sample-flow';
import { DOMAIN_OPTIONS } from '../data/builtin-components';
import { NodeIcon } from './nodes/NodeIcon';

const TEMPLATES = [
  { name: 'dev-flow', label: 'Dev Flow (Main)', domain: 'development' },
  { name: 'game-design-flow', label: 'Game Design Flow', domain: 'game-design' },
  { name: 'novel-flow', label: 'Novel Writing Flow', domain: 'writing' },
];

const nodeTypes: { type: NodeType; color: string }[] = [
  { type: 'start', color: '#10B981' },
  { type: 'phase', color: '#3B82F6' },
  { type: 'gate', color: '#F59E0B' },
  { type: 'terminal', color: '#6B7280' },
];

interface LibraryPanelProps {
  visible: boolean;
  onToggle: () => void;
}

export function LibraryPanel({ visible, onToggle }: LibraryPanelProps) {
  const { currentFlow, loadFlow, updateFlowMetadata } = useFlowStore();
  const [showTemplateDialog, setShowTemplateDialog] = useState(false);

  const i18n = t();

  const onDragStart = (event: React.DragEvent, nodeType: NodeType) => {
    event.dataTransfer.setData('application/reactflow', nodeType);
    event.dataTransfer.effectAllowed = 'move';
  };

  const handleLoadTemplate = (templateName: string) => {
    switch (templateName) {
      case 'dev-flow':
        loadFlow(sampleDevFlow);
        break;
      case 'game-design-flow':
        loadFlow(sampleGameDesignFlow);
        break;
      case 'novel-flow':
        loadFlow(sampleNovelFlow);
        break;
    }
    setShowTemplateDialog(false);
  };

  const handleDomainChange = (domain: string) => {
    updateFlowMetadata({ domain });
  };

  if (!visible) {
    return (
      <div
        style={{
          width: 40,
          borderRight: '1px solid #E5E7EB',
          background: '#F9FAFB',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          paddingTop: 8,
        }}
      >
        <button
          onClick={onToggle}
          style={{
            background: 'none',
            border: '1px solid #D1D5DB',
            borderRadius: 4,
            padding: '4px 8px',
            cursor: 'pointer',
            fontSize: 16,
          }}
          title={i18n.library.title}
        >
          ≡
        </button>
      </div>
    );
  }

  return (
    <>
      <div
        style={{
          width: 240,
          borderRight: '1px solid #E5E7EB',
          background: '#F9FAFB',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
          flexShrink: 0,
        }}
      >
        <div style={{ padding: '8px 12px', borderBottom: '1px solid #E5E7EB', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontWeight: 600, fontSize: 13 }}>{i18n.library.title}</span>
          <button
            onClick={onToggle}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              fontSize: 14,
              color: '#6B7280',
            }}
            title={i18n.library.hide}
          >
            ✕
          </button>
        </div>

        <div style={{ padding: '8px 12px', borderBottom: '1px solid #E5E7EB' }}>
          <div style={{ fontSize: 11, color: '#6B7280', fontWeight: 600, textTransform: 'uppercase', marginBottom: 6 }}>
            {i18n.library.flows}
          </div>
          <div style={{ padding: '6px 8px', background: '#fff', border: '1px solid #D1D5DB', borderRadius: 4, fontSize: 12, color: '#111827' }}>
            {currentFlow?.metadata.name ?? '-'}
          </div>
          <button
            onClick={() => setShowTemplateDialog(true)}
            style={{
              marginTop: 6,
              padding: '4px 12px',
              borderRadius: 4,
              border: '1px solid #3B82F6',
              background: '#fff',
              color: '#3B82F6',
              cursor: 'pointer',
              fontSize: 12,
              width: '100%',
              textAlign: 'center',
            }}
          >
            从模板加载
          </button>
        </div>

        <div style={{ padding: '8px 12px', borderBottom: '1px solid #E5E7EB', flex: 1, overflow: 'auto' }}>
          <div style={{ fontWeight: 600, fontSize: 11, color: '#6B7280', textTransform: 'uppercase', marginBottom: 6 }}>
            {i18n.library.nodes}
          </div>
          {nodeTypes.map((item) => {
            const nodeI18n = i18n.nodes[item.type];
            return (
              <div
                key={item.type}
                draggable
                onDragStart={(e) => onDragStart(e, item.type)}
                style={{
                  padding: '6px 8px',
                  marginBottom: 2,
                  borderRadius: 4,
                  cursor: 'grab',
                  fontSize: 12,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  borderLeft: `3px solid ${item.color}`,
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLDivElement).style.background = '#E5E7EB';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLDivElement).style.background = 'transparent';
                }}
              >
                <NodeIcon type={item.type} size={16} />
                <div>
                  <div style={{ fontWeight: 500 }}>{nodeI18n.label}</div>
                  <div style={{ color: '#9CA3AF', fontSize: 10 }}>{nodeI18n.description}</div>
                </div>
              </div>
            );
          })}
        </div>

        <div style={{ padding: '8px 12px' }}>
          <div style={{ fontWeight: 600, fontSize: 11, color: '#6B7280', textTransform: 'uppercase', marginBottom: 6 }}>
            {i18n.library.templates}
          </div>
          <div style={{ color: '#9CA3AF', fontSize: 11, fontStyle: 'italic' }}>
            {i18n.library.comingSoon}
          </div>
        </div>

        <div style={{ padding: '8px 12px', borderTop: '1px solid #E5E7EB' }}>
          <div style={{ fontWeight: 600, fontSize: 11, color: '#6B7280', textTransform: 'uppercase', marginBottom: 6 }}>
            Domain
          </div>
          <select
            value={currentFlow?.metadata.domain ?? 'development'}
            onChange={(e) => handleDomainChange(e.target.value)}
            style={{
              width: '100%',
              padding: '6px 8px',
              borderRadius: 4,
              border: '1px solid #D1D5DB',
              background: '#fff',
              fontSize: 12,
              color: '#111827',
              cursor: 'pointer',
            }}
          >
            {DOMAIN_OPTIONS.map((d) => (
              <option key={d} value={d}>{d}</option>
            ))}
          </select>
        </div>
      </div>

      {showTemplateDialog && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0,0,0,0.3)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
          }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowTemplateDialog(false);
          }}
        >
          <div
            style={{
              background: '#fff',
              borderRadius: 8,
              padding: 24,
              width: 400,
              boxShadow: '0 4px 24px rgba(0,0,0,0.15)',
            }}
          >
            <div style={{ fontWeight: 600, fontSize: 15, marginBottom: 16 }}>
              从模板加载
            </div>

            <div style={{ marginBottom: 16 }}>
              {(() => {
                const domains = [...new Set(TEMPLATES.map((t) => t.domain))];
                return domains.map((domain) => (
                  <div key={domain}>
                    <div style={{ fontWeight: 600, fontSize: 11, color: '#6B7280', textTransform: 'uppercase', marginBottom: 4, marginTop: 8 }}>
                      {domain}
                    </div>
                    {TEMPLATES.filter((t) => t.domain === domain).map((tpl) => (
                      <div
                        key={tpl.name}
                        onClick={() => handleLoadTemplate(tpl.name)}
                        style={{
                          padding: '8px 12px',
                          marginBottom: 4,
                          borderRadius: 4,
                          cursor: 'pointer',
                          fontSize: 12,
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center',
                          border: '1px solid #E5E7EB',
                        }}
                        onMouseEnter={(e) => {
                          (e.currentTarget as HTMLDivElement).style.background = '#DBEAFE';
                        }}
                        onMouseLeave={(e) => {
                          (e.currentTarget as HTMLDivElement).style.background = '#fff';
                        }}
                      >
                        <span style={{ fontWeight: 500 }}>{tpl.label}</span>
                        <span style={{ color: '#9CA3AF', fontSize: 10 }}>{tpl.name}</span>
                      </div>
                    ))}
                  </div>
                ));
              })()}
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
              <button
                onClick={() => setShowTemplateDialog(false)}
                style={{
                  padding: '6px 16px',
                  borderRadius: 4,
                  border: '1px solid #D1D5DB',
                  background: '#fff',
                  color: '#374151',
                  cursor: 'pointer',
                  fontSize: 12,
                }}
              >
                {i18n.library.newFlowDialog.cancel}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
