import React, { useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { ThemeContext } from '@/contexts/ThemeContext';
import LayerStateGraph from '@/components/tools/LayerStateGraph';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { reactQueryKeys } from '@/clients/reactQueryConfig';
import { fetchLayer, syncLayer } from '@/clients/layers/client';
import { fetchAttempts } from '@/clients/runs/client';
import { fetchPlan } from '@/clients/plans/client';
import LayerStatus from '@/components/status/LayerStatus';
import Button from '@/components/core/Button';
import type { Layer, StateGraph, StateGraphNode } from '@/clients/layers/types';
import SlidingPane from '@/modals/SlidingPane';
import StateGraphInstanceCard from '@/components/cards/StateGraphInstanceCard';
import {
  parseTerraformPlan,
  type AggregatedPlanChange,
  type PlanAction,
  type PlanChange
} from '@/utils/terraformPlan';

const Layer: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const { namespace = '', name = '' } = useParams();
  const navigate = useNavigate();
  const [showResourcePane, setShowResourcePane] = useState(false);
  const [selectedResourceData, setSelectedResourceData] = useState<StateGraphNode | null>(null);
  const [isManualSyncPending, setIsManualSyncPending] = useState<boolean>(false);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);
  const [stateGraph, setStateGraph] = useState<StateGraph | null>(null);

  const layerQuery = useQuery({
    queryKey: reactQueryKeys.layer(namespace, name),
    queryFn: () => fetchLayer(namespace, name),
  });

  const syncSelectedLayer = async (layer: Layer) => {
    const sync = await syncLayer(layer.namespace, layer.name);
    if (sync.status === 200) {
      setIsManualSyncPending(true);
    }
  };
  const layer = layerQuery.data;

  useEffect(() => {
    if (!layer) return;
    const pending = layer.manualSyncStatus === 'pending' || layer.manualSyncStatus === 'annotated';
    setIsManualSyncPending(pending);
  }, [layer]);

  // UI Interactivity: Poll the layer every 10s:
  // - while a manual sync is pending or the layer is running
  // - or if the last run was within the last 2 minutes (to wait for controller to update status)
  useEffect(() => {
    const lastRunTime = layer?.lastRun?.date
      ? new Date(layer.lastRun.date).getTime()
      : 0
    const withinTwoMinutes = Date.now() - lastRunTime < 2 * 60 * 1000
    const shouldPoll =
      isManualSyncPending ||
      !!layer?.isRunning ||
      withinTwoMinutes
    if (!shouldPoll) return;
    const id = setInterval(() => {
      layerQuery.refetch();
    }, 10000);

    return () => clearInterval(id);
  }, [isManualSyncPending, layer, layerQuery]);

  const refresh = () => {
    setIsRefreshing(true);
    layerQuery.refetch().finally(() => {
      setTimeout(() => setIsRefreshing(false), 1000);
    });
  };

  const latestRunId = layer?.lastRun?.id ?? '';

  const attemptsQuery = useQuery({
    queryKey: reactQueryKeys.attempts(namespace, name, latestRunId),
    queryFn: () => fetchAttempts(namespace, name, latestRunId),
    enabled: !!latestRunId
  });

  const latestAttempt = useMemo<number | null>(() => {
    if (!attemptsQuery.data || attemptsQuery.data.count === 0) {
      return null;
    }
    return attemptsQuery.data.count - 1;
  }, [attemptsQuery.data]);

  const planQuery = useQuery({
    queryKey: reactQueryKeys.plan(namespace, name, latestRunId || null, latestAttempt),
    queryFn: () => fetchPlan(namespace, name, latestRunId, latestAttempt!),
    enabled: !!latestRunId && latestAttempt !== null,
    select: (data) => parseTerraformPlan(data)
  });

  const planHighlights = planQuery.data ?? null;

  const gatherPlanChanges = (
    aggregate: AggregatedPlanChange | undefined | null,
    fallback: PlanChange | undefined | null
  ): PlanChange[] => {
    const list: PlanChange[] = [];
    if (aggregate) {
      if (aggregate.single) {
        list.push(aggregate.single);
      }
      for (const change of Object.values(aggregate.instances ?? {})) {
        list.push(change);
      }
    }
    if (list.length === 0 && fallback) {
      list.push(fallback);
    }
    return list;
  };

  const selectedPlanDetails = useMemo(() => {
    if (!selectedResourceData || !planHighlights) {
      return null;
    }
    const baseId = selectedResourceData.id;
    const exact = planHighlights.byAddr.get(baseId);
    const base = planHighlights.byBase.get(baseId);
    const action = (base?.action ?? exact?.action ?? null) as PlanAction | null;
    if (!action) {
      return null;
    }
    const planChanges = gatherPlanChanges(base, exact);
    const futureInstances = planChanges
      .filter((change) => change.after !== undefined && change.after !== null)
      .map((change) => ({
        addr: change.addr,
        // cast to the expected record type or undefined
        attributes: (change.after ?? undefined) as Record<string, unknown> | undefined
      }));
    const planHasOnlyCreates =
      planChanges.length > 0 &&
      planChanges.every((change) => change.before === null || change.before === undefined);
    return {
      action,
      futureInstances,
      planHasOnlyCreates,
      hasPlanChanges: planChanges.length > 0
    };
  }, [planHighlights, selectedResourceData]);

  const currentInstances = useMemo(() => {
    if (!selectedResourceData) {
      return [];
    }
    if (selectedPlanDetails?.action === 'create' && selectedPlanDetails.planHasOnlyCreates) {
      return [];
    }
    return selectedResourceData.instances ?? [];
  }, [selectedPlanDetails, selectedResourceData]);

  const futureInstances = selectedPlanDetails?.futureInstances ?? [];

  const currentInstanceCount = currentInstances.length;
  const futureInstanceCount = selectedPlanDetails
    ? selectedPlanDetails.action === 'delete'
      ? 0
      : futureInstances.length
    : null;
  const showCountArrow =
    futureInstanceCount !== null && futureInstanceCount !== currentInstanceCount;

  const paneTitleClass =
    theme === 'light' ? 'text-nuances-black' : 'text-nuances-50';
  const paneLabelClass = theme === 'light' ? 'text-gray-500' : 'text-nuances-300';
  const paneValueClass = theme === 'light' ? 'text-gray-500' : 'text-nuances-200';
  const paneMutedTextClass =
    theme === 'light' ? 'text-gray-500' : 'text-nuances-200';
  const plannedDeletionClass = twMerge(
    'text-sm rounded-md p-3 border',
    theme === 'light'
      ? 'text-red-600 bg-red-50 border-red-100'
      : 'text-red-200 bg-red-900/30 border-red-800'
  );
  const plannedCreationClass = twMerge(
    'text-sm rounded-md p-3 mt-4 border',
    theme === 'light'
      ? 'text-emerald-700 bg-emerald-50 border-emerald-100'
      : 'text-emerald-200 bg-emerald-900/30 border-emerald-800'
  );

  const resolveDependencyNode = useCallback(
    (dependencyAddr: string): StateGraphNode | null => {
      if (!stateGraph?.nodes?.length) return null;
      const trimmed = dependencyAddr.trim();
      return (
        stateGraph.nodes.find((n) => n.id === trimmed || n.addr === trimmed) ??
        stateGraph.nodes.find((n) =>
          n.instances?.some((inst) => inst.addr === trimmed)
        ) ??
        null
      );
    },
    [stateGraph]
  );

  const handleDependencyClick = (dependencyAddr: string) => {
    const node = resolveDependencyNode(dependencyAddr);
    if (!node) return;
    setSelectedResourceData(node);
    setShowResourcePane(true);
  };

  return (
    <div className="flex flex-col flex-1 h-screen min-w-0">
      <SlidingPane
        isOpen={showResourcePane}
        onClose={() => setShowResourcePane(false)}
        variant={theme}
        width="w-3/7"
      >
        <div>
          <h2 className={twMerge('text-2xl font-bold mb-2', paneTitleClass)}>
            Resource: {selectedResourceData?.name}
          </h2>
          <h3
            className={twMerge(
              'text-sm uppercase font-semibold',
              theme === 'light' ? 'text-primary-600' : 'text-primary-300'
            )}
          >
            {selectedResourceData?.type}
          </h3>
          <div className="grid grid-cols-[min-content_1fr] mt-4 gap-x-8">
            <span className={twMerge('text-sm text-right', paneLabelClass)}>
              Provider:
            </span>
            <span
              className={twMerge('text-sm truncate pr-8', paneValueClass)}
              title={selectedResourceData?.provider}
            >
              {selectedResourceData?.provider}
            </span>
            <span className={twMerge('text-sm text-right', paneLabelClass)}>
              Address:
            </span>
            <span
              className={twMerge('text-sm truncate pr-8', paneValueClass)}
              title={selectedResourceData?.addr}
            >
              {selectedResourceData?.addr}
            </span>
            {selectedResourceData?.module !== undefined && (
              <>
                <span className={twMerge('text-sm text-right', paneLabelClass)}>
                  Module:
                </span>
                <span
                  className={twMerge('text-sm truncate pr-8', paneValueClass)}
                  title={selectedResourceData?.module || '(root)'}
                >
                  {selectedResourceData?.module || '(root)'}
                </span>
              </>
            )}
            <span className={twMerge('text-sm text-right', paneLabelClass)}>
              Count:
            </span>
            <span className={twMerge('text-sm truncate pr-8', paneValueClass)}>
              {currentInstanceCount}
              {showCountArrow ? ` → ${futureInstanceCount}` : ''}
            </span>
          </div>
          {selectedPlanDetails && (
            <div className="mt-4">
              <h3 className="text-lg font-semibold mb-2">Planned change</h3>
              <div className={twMerge('text-sm mb-2', paneMutedTextClass)}>
                <span
                  className={twMerge(
                    'font-medium',
                    theme === 'light' ? 'text-gray-600' : 'text-nuances-100'
                  )}
                >
                  Action:
                </span>{' '}
                {selectedPlanDetails.action}
              </div>
              {selectedPlanDetails.action === 'delete' && (
                <div className={plannedDeletionClass}>
                  All current instances will be destroyed when this plan is applied.
                </div>
              )}
            </div>
          )}
          {currentInstances.length > 0 && (
            <>
              <h3 className="text-lg font-semibold mt-4 mb-2">
                Current instance{currentInstances.length > 1 ? 's' : ''}
              </h3>
              <ul className="list-inside">
                {currentInstances.map((inst) => (
                  <li key={inst.addr} className="mb-2">
                    <StateGraphInstanceCard
                      instance={inst}
                      defaultExpanded={currentInstances.length === 1}
                      variant={theme}
                      tone="current"
                      isDependencyAvailable={(addr) => !!resolveDependencyNode(addr)}
                      onDependencyClick={handleDependencyClick}
                    />
                  </li>
                ))}
              </ul>
            </>
          )}
          {futureInstances.length > 0 && selectedPlanDetails && selectedPlanDetails.action !== 'delete' && (
            <>
              <h3 className="text-lg font-semibold mt-4 mb-2">
                Future instance{futureInstances.length > 1 ? 's' : ''}
              </h3>
              <ul className="list-inside">
                {futureInstances.map((inst) => (
                  <li key={`future-${inst.addr}`} className="mb-2">
                    <StateGraphInstanceCard
                      instance={inst}
                      variant={theme}
                      tone="future"
                      planAction={selectedPlanDetails.action}
                      isDependencyAvailable={(addr) => !!resolveDependencyNode(addr)}
                      onDependencyClick={handleDependencyClick}
                    />
                  </li>
                ))}
              </ul>
            </>
          )}
          {selectedPlanDetails?.action === 'create' &&
            futureInstances.length === 0 &&
            selectedPlanDetails.planHasOnlyCreates && (
              <div className={plannedCreationClass}>
                This resource will be created when the plan is applied.
              </div>
            )}
        </div>
      </SlidingPane>
      <div
        className={`
          p-6
          pb-3
          h-full
          flex
          flex-col
          min-h-0
          ${theme === 'light' ? 'bg-primary-100' : 'bg-nuances-black'}
        `}
      >
        <div className="p-6 pb-3">
          <h1
            className={`
              text-[32px]
              font-extrabold
              leading-[130%]
              ${theme === 'light' ? 'text-nuances-black' : 'text-nuances-50'}
            `}
          >
            Layer - {layer?.namespace}/{layer?.name}
          </h1>
        </div>
        <div className="flex p-6 justify-between gap-4">
          <LayerStatus layer={layer} variant="health" theme={theme} syncPending={isManualSyncPending} />
          <LayerStatus layer={layer} variant="lastOperation" theme={theme} />
          <LayerStatus layer={layer} variant="details" theme={theme} />
          <div className="flex flex-col justify-between gap-2 min-w-32">
            <Button
              theme={theme}
              variant="primary"
              disabled={!layer || isRefreshing}
              onClick={() => refresh()}
            >
              {'Refresh'}
            </Button>
            <Button
              theme={theme}
              variant="secondary"
              onClick={() => syncSelectedLayer(layer!).then(() => layerQuery.refetch())}
              disabled={!layer || isManualSyncPending}
            >
              Run sync
            </Button>
            <Button
              theme={theme}
              variant="secondary"
              onClick={() => navigate(`/logs/${layer!.namespace}/${layer!.name}`)}
            >
              View logs
            </Button>

          </div>
        </div>
        <div className="flex-1 min-h-0 overflow-auto p-6">
          <LayerStateGraph
            namespace={namespace}
            name={name}
            variant={theme === 'light' ? 'light' : 'dark'}
            plan={planHighlights}
            planLoading={planQuery.isLoading || planQuery.isFetching}
            onGraphChange={setStateGraph}
            onNodeClick={(n) => { setShowResourcePane(true)
              setSelectedResourceData(n);
              console.log('Clicked node', n);
            } }
          />
        </div>
      </div>
    </div>
  );
};

export default Layer;
