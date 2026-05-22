import { ReactFlowProvider } from '@xyflow/react';
import { Layout } from './components/Layout';

export default function App() {
  return (
    <ReactFlowProvider>
      <Layout />
    </ReactFlowProvider>
  );
}
