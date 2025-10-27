import React, { useEffect, useMemo, useState } from 'react';
import { reactQueryKeys } from '@/clients/reactQueryConfig';
import { fetchLayer, fetchStateGraph } from '@/clients/layers/client';
import { useQuery } from '@tanstack/react-query';
import ReactFlowView from './ReactFlowView';
import { buildReactFlow, type ReactFlowGraph } from '@/utils/stateGraph';
import { StateGraph, StateGraphNode } from '@/clients/layers/types';
import type { ParsedTerraformPlan } from '@/utils/terraformPlan';
import { augmentStateGraphWithPlan } from '@/utils/terraformPlan';

export interface LayerStateGraphProps {
  variant?: 'light' | 'dark';
  namespace: string;
  name: string;
  onNodeClick?: (n: StateGraphNode) => void;
  plan?: ParsedTerraformPlan | null;
  planLoading?: boolean;
}

const LayerStateGraph: React.FC<LayerStateGraphProps> = ({
  variant = 'light',
  namespace,
  name,
  onNodeClick,
  plan,
  planLoading = false
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
  const [graph, setGraph] = useState<StateGraph | null>(null);

  const augmentedGraph = useMemo<StateGraph>(() => {
    return augmentStateGraphWithPlan(stateGraphQuery.data, plan);
  }, [stateGraphQuery.data, plan]);

  useEffect(() => {
    let cancelled = false;
    setGraph(augmentedGraph);
    buildReactFlow(augmentedGraph).then((res) => {
      if (cancelled) return;
      if (!plan) {
        setRf(res);
        return;
      }
      const nodesWithPlan = res.nodes.map((node) => {
        const exact = plan.byAddr.get(node.id);
        const base = plan.byBase.get(node.id);
        const change = exact?.action ?? base?.action ?? null;
        const future =
          exact?.after ??
          (base?.instances && Object.keys(base.instances).length > 0
            ? Object.fromEntries(
                Object.entries(base.instances).map(([idx, value]) => [
                  idx,
                  value.after
                ])
              )
            : base?.single?.after);
        return {
          ...node,
          data: {
            ...node.data,
            change,
            future
          }
        };
      });
      setRf({ nodes: nodesWithPlan, edges: res.edges });
    });
    return () => {
      cancelled = true;
    };
  }, [augmentedGraph, plan]);

  const hasGraphData = (graph?.nodes?.length ?? 0) > 0;

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

  if (!hasGraphData && (stateGraphQuery.isLoading || planLoading)) {
    return (
      <div className="flex items-center justify-center h-full p-4">
        <div className="text-slate-500">Loading state graph...</div>
      </div>
    );
  }

  if (
    !hasGraphData &&
    stateGraphQuery.isError &&
    !planLoading &&
    (!plan || plan.byBase.size === 0)
  ) {
    return (
      <div className="flex items-center justify-center h-full p-4">
        <div className="text-slate-500">
          No state graph available for this layer
        </div>
      </div>
    );
  }

  if (!hasGraphData || rf.nodes.length === 0) {
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
      <ReactFlowView
        rf={rf}
        variant={variant}
        onNodeClick={(id) => {
          if (!onNodeClick || !graph?.nodes) return;
          const node = graph.nodes.find((n) => n.id === id);
          if (node) {
            onNodeClick(node);
          }
        }}
      />
    </div>
  );
};

export default LayerStateGraph;
