import React, { useEffect, useState } from 'react';
import { reactQueryKeys } from '@/clients/reactQueryConfig';
import { fetchLayer, fetchStateGraph } from '@/clients/layers/client';
import { useQuery } from '@tanstack/react-query';
import ReactFlowView from './ReactFlowView';
import { buildReactFlow, type ReactFlowGraph } from '@/utils/stateGraph';

export interface LayerStateGraphProps {
  variant?: 'light' | 'dark';
  namespace: string;
  name: string;
}

const LayerStateGraph: React.FC<LayerStateGraphProps> = ({
  variant = 'light',
  namespace,
  name
}) => {
  const layerQuery = useQuery({
    queryKey: reactQueryKeys.layer(namespace, name),
    queryFn: () => fetchLayer(namespace, name)
  });

  const stateGraphQuery = useQuery({
    queryKey: ['stateGraph', namespace, name],
    queryFn: () => fetchStateGraph(namespace, name),
    enabled: !!layerQuery.data
  });

  const [rf, setRf] = useState<ReactFlowGraph>({ nodes: [], edges: [] });

  useEffect(() => {
    let cancelled = false;
    if (!stateGraphQuery.data) {
      setRf({ nodes: [], edges: [] });
      return;
    }
    buildReactFlow(stateGraphQuery.data).then((res) => {
      if (cancelled) return;
      setRf(res);
    });
    return () => {
      cancelled = true;
    };
  }, [stateGraphQuery.data]);

  if (layerQuery.isLoading) {
    return (
      <div className="flex items-center justify-center h-full p-4">
        <div className="text-slate-500">Loading layer...</div>
      </div>
    );
  }

  if (layerQuery.isError) {
    return (
      <div className="flex items-center justify-center h-full p-4">
        <div className="text-red-500">
          Error loading layer: {(layerQuery.error as Error).message}
        </div>
      </div>
    );
  }

  if (stateGraphQuery.isLoading) {
    return (
      <div className="flex items-center justify-center h-full p-4">
        <div className="text-slate-500">Loading state graph...</div>
      </div>
    );
  }

  if (stateGraphQuery.isError) {
    return (
      <div className="flex items-center justify-center h-full p-4">
        <div className="text-slate-500">
          No state graph available for this layer
        </div>
      </div>
    );
  }

  if (!stateGraphQuery.data || rf.nodes.length === 0) {
    return (
      <div className="flex items-center justify-center h-full p-4">
        <div className="text-slate-500">
          No state graph data available for this layer
        </div>
      </div>
    );
  }

  return (
    <div className="h-full w-full">
      <ReactFlowView rf={rf} variant={variant} />
    </div>
  );
};

export default LayerStateGraph;
