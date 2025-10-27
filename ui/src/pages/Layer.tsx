import React, { useContext, useEffect, useMemo, useState } from 'react';
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
import type { Layer, StateGraphNode } from '@/clients/layers/types';
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

  return (
    <div className="flex flex-col flex-1 h-screen min-w-0">
      <SlidingPane
        isOpen={showResourcePane}
        onClose={() => setShowResourcePane(false)}
        variant={theme}
      >
        <div>
          <h2 className="text-2xl font-bold mb-2">Ressource: {selectedResourceData?.name}</h2>
          <h3 className="text-sm uppercase font-semibold text-primary-600">{selectedResourceData?.type}</h3>
            <div className="grid grid-cols-[min-content_1fr] mt-4 gap-x-8">
            <span className="text-sm text-gray-500 text-right">Provider:</span>
            <span className="text-sm text-gray-500 truncate pr-8" title={selectedResourceData?.provider}>{selectedResourceData?.provider}</span>
            <span className="text-sm text-gray-500 text-right">Address:</span>
            <span className="text-sm text-gray-500 truncate pr-8" title={selectedResourceData?.addr}>{selectedResourceData?.addr}</span>
            {selectedResourceData?.module !== undefined && (
              <>
              <span className="text-sm text-gray-500 text-right">Module:</span>
              <span className="text-sm text-gray-500 truncate pr-8" title={selectedResourceData?.module || '(root)'}>{selectedResourceData?.module || '(root)'}</span>
              </>
            )}
            <span className="text-sm text-gray-500 text-right">Count:</span>
            <span className="text-sm text-gray-500 truncate pr-8">
              {currentInstanceCount}
              {showCountArrow ? ` → ${futureInstanceCount}` : ''}
            </span>
            </div>
          {selectedPlanDetails && (
            <div className="mt-4">
              <h3 className="text-lg font-semibold mb-2">Planned change</h3>
              <div className="text-sm text-gray-500 mb-2">
                <span className="text-gray-600 font-medium">Action:</span> {selectedPlanDetails.action}
              </div>
              {selectedPlanDetails.action === 'delete' && (
                <div className="text-sm text-red-600 bg-red-50 border border-red-100 rounded-md p-3">
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
                      tone="current"
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
                      tone="future"
                      planAction={selectedPlanDetails.action}
                      badge={
                        <span
                          className={twMerge(
                            'text-[10px] uppercase font-semibold px-2 py-0.5 rounded-full border',
                            selectedPlanDetails.action === 'create' &&
                              'bg-emerald-100 text-emerald-700 border-emerald-300',
                            selectedPlanDetails.action === 'update' &&
                              'bg-amber-100 text-amber-700 border-amber-300',
                            selectedPlanDetails.action === 'replace' &&
                              'bg-violet-100 text-violet-700 border-violet-300'
                          )}
                        >
                          {selectedPlanDetails.action}
                        </span>
                      }
                    />
                  </li>
                ))}
              </ul>
            </>
          )}
          {selectedPlanDetails?.action === 'create' &&
            futureInstances.length === 0 &&
            selectedPlanDetails.planHasOnlyCreates && (
              <div className="text-sm text-emerald-700 bg-emerald-50 border border-emerald-100 rounded-md p-3 mt-4">
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
          <LayerStatus layer={layer} variant="health" syncPending={isManualSyncPending} />
          <LayerStatus layer={layer} variant="lastOperation" />
          <LayerStatus layer={layer} variant="details" />
          <div className="flex flex-col justify-between gap-2 min-w-32">
            <Button
              onClick={() => syncSelectedLayer(layer!).then(() => layerQuery.refetch())}
              disabled={!layer || isManualSyncPending}
            >
              Run sync
            </Button>
            <Button variant='secondary' onClick={() => navigate(`/logs/${layer!.namespace}/${layer!.name}`)}>View logs</Button>
            <Button variant='secondary' disabled={!layer || isRefreshing} onClick={() => refresh()}>{'Refresh'}</Button>
          </div>
        </div>
        <div className="flex-1 min-h-0 overflow-auto p-6">
          <LayerStateGraph
            namespace={namespace}
            name={name}
            variant={theme === 'light' ? 'light' : 'dark'}
            plan={planHighlights}
            planLoading={planQuery.isLoading || planQuery.isFetching}
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
