import { ReactFlowProvider } from '@xyflow/react';
import { Layout } from './components/Layout';
import { useFlowLoader } from './hooks/useFlowLoader';
import { sampleDevFlow } from './data/sample-flow';

function AppContent() {
  useFlowLoader(sampleDevFlow);
  return <Layout />;
}

export default function App() {
  return (
    <ReactFlowProvider>
      <AppContent />
    </ReactFlowProvider>
  );
}
