import { useState, useEffect, useCallback } from 'react';
import {
  fetchTeams,
  fetchTeamFlows,
  fetchProjectFlows,
  type TeamMeta,
  type TeamFlowMeta,
  type ProjectFlowMeta,
} from '../api/editor-api';
import { useEditorStore } from '../stores/editor-store';

export function TeamTree() {
  const [teams, setTeams] = useState<TeamMeta[]>([]);
  const [teamFlows, setTeamFlows] = useState<Map<string, TeamFlowMeta[]>>(new Map());
  const [projectFlows, setProjectFlows] = useState<ProjectFlowMeta[]>([]);
  const [expandedTeams, setExpandedTeams] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const { currentFlowId, currentFlowSource, loadProjectFlow, loadPresetFlow } = useEditorStore();

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      const [t, pf] = await Promise.all([fetchTeams(), fetchProjectFlows()]);
      setTeams(t);
      setProjectFlows(pf);

      const flowMap = new Map<string, TeamFlowMeta[]>();
      await Promise.all(
        t.map(async (team) => {
          try {
            const flows = await fetchTeamFlows(team.id);
            flowMap.set(team.id, flows);
          } catch {
            flowMap.set(team.id, []);
          }
        }),
      );
      setTeamFlows(flowMap);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const toggleTeam = (teamId: string) => {
    setExpandedTeams((prev) => {
      const next = new Set(prev);
      if (next.has(teamId)) next.delete(teamId);
      else next.add(teamId);
      return next;
    });
  };

  const handleProjectFlowClick = (flowId: string) => {
    loadProjectFlow(flowId);
  };

  const handlePresetFlowClick = (teamId: string, flowId: string) => {
    loadPresetFlow(teamId, flowId);
  };

  const sectionHeader: React.CSSProperties = {
    fontSize: 11,
    fontWeight: 600,
    color: '#6B7280',
    textTransform: 'uppercase' as const,
    letterSpacing: '0.05em',
    padding: '8px 12px 4px',
  };

  const flowItem = (id: string, isDefault: boolean, source: 'project' | 'preset', teamId?: string): React.CSSProperties => ({
    padding: '5px 12px 5px 20px',
    fontSize: 12,
    cursor: 'pointer',
    display: 'flex',
    alignItems: 'center',
    gap: 6,
    background: currentFlowId === id && currentFlowSource === source ? '#DBEAFE' : 'transparent',
    borderLeft: currentFlowId === id && currentFlowSource === source ? '3px solid #3B82F6' : '3px solid transparent',
  });

  const teamHeader = (teamId: string): React.CSSProperties => ({
    padding: '6px 12px',
    fontSize: 12,
    cursor: 'pointer',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    fontWeight: 500,
    background: expandedTeams.has(teamId) ? '#F3F4F6' : 'transparent',
  });

  if (loading) {
    return (
      <div style={{ width: 240, borderRight: '1px solid #E5E7EB', background: '#F9FAFB', padding: 12, fontSize: 12, color: '#6B7280' }}>
        Loading...
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ width: 240, borderRight: '1px solid #E5E7EB', background: '#F9FAFB', padding: 12 }}>
        <div style={{ fontSize: 12, color: '#DC2626', marginBottom: 8 }}>Error: {error}</div>
        <button onClick={loadData} style={{ fontSize: 11, padding: '4px 8px', border: '1px solid #D1D5DB', borderRadius: 4, background: '#fff', cursor: 'pointer' }}>
          Retry
        </button>
      </div>
    );
  }

  return (
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
      <div style={{ padding: '8px 12px', borderBottom: '1px solid #E5E7EB', fontWeight: 600, fontSize: 13 }}>
        Team Explorer
      </div>

      <div style={{ flex: 1, overflow: 'auto' }}>
        <div style={sectionHeader}>Project Flows</div>
        {projectFlows.length === 0 ? (
          <div style={{ padding: '4px 20px', fontSize: 11, color: '#9CA3AF', fontStyle: 'italic' }}>
            No installed flows
          </div>
        ) : (
          projectFlows.map((f) => (
            <div
              key={f.id}
              style={flowItem(f.id, f.default, 'project')}
              onClick={() => handleProjectFlowClick(f.id)}
              onMouseEnter={(e) => { if (currentFlowId !== f.id) (e.currentTarget as HTMLDivElement).style.background = '#F3F4F6'; }}
              onMouseLeave={(e) => { if (currentFlowId !== f.id) (e.currentTarget as HTMLDivElement).style.background = 'transparent'; }}
            >
              {f.default && <span style={{ color: '#F59E0B' }}>⭐</span>}
              <span>{f.id}</span>
            </div>
          ))
        )}

        <div style={{ ...sectionHeader, marginTop: 8 }}>Preset Teams</div>
        {teams.map((team) => {
          const flows = teamFlows.get(team.id) ?? [];
          const isExpanded = expandedTeams.has(team.id);
          return (
            <div key={team.id}>
              <div
                style={teamHeader(team.id)}
                onClick={() => toggleTeam(team.id)}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  <span style={{ fontSize: 10, color: '#9CA3AF' }}>{isExpanded ? '▼' : '▶'}</span>
                  <span>{team.name_zh || team.name}</span>
                </div>
                <span style={{ fontSize: 10, color: '#9CA3AF' }}>{team.flow_count}</span>
              </div>
              {isExpanded && flows.map((f) => (
                <div
                  key={f.id}
                  style={flowItem(f.id, f.default, 'preset', team.id)}
                  onClick={() => handlePresetFlowClick(team.id, f.id)}
                  onMouseEnter={(e) => { if (!(currentFlowId === f.id && currentFlowSource === 'preset')) (e.currentTarget as HTMLDivElement).style.background = '#F3F4F6'; }}
                  onMouseLeave={(e) => { if (!(currentFlowId === f.id && currentFlowSource === 'preset')) (e.currentTarget as HTMLDivElement).style.background = 'transparent'; }}
                >
                  {f.default && <span style={{ color: '#F59E0B' }}>⭐</span>}
                  <span>{f.id}</span>
                </div>
              ))}
            </div>
          );
        })}
      </div>
    </div>
  );
}
