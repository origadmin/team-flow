import { useEffect } from 'react';
import { useFlowStore } from '../stores/flow-store';
import type { Flow } from '../types/flow';

export function useFlowLoader(flow: Flow) {
  const loadFlow = useFlowStore((s) => s.loadFlow);
  const createFlow = useFlowStore((s) => s.createFlow);
  const flows = useFlowStore((s) => s.flows);
  const currentFlow = useFlowStore((s) => s.currentFlow);

  useEffect(() => {
    if (flows.length === 0) {
      createFlow(flow.metadata.name, flow.metadata.domain ?? '', flow);
    }
    if (!currentFlow) {
      loadFlow(flow);
    }
  }, [flow, loadFlow, createFlow, flows.length, currentFlow]);
}
