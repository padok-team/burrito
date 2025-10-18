import React, { useContext, useEffect, useState } from 'react';


import { ThemeContext } from '@/contexts/ThemeContext';
import LayerStateGraph from '@/components/tools/LayerStateGraph';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { reactQueryKeys } from '@/clients/reactQueryConfig';
import { fetchLayer, syncLayer } from '@/clients/layers/client';
import LayerStatus from '@/components/status/LayerStatus';
import Button from '@/components/core/Button';
import type { Layer, StateGraphNode } from '@/clients/layers/types';
import SlidingPane from '@/modals/SlidingPane';

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

  return (
    <div className="flex flex-col flex-1 h-screen min-w-0">
      <SlidingPane
        isOpen={showResourcePane}
        onClose={() => setShowResourcePane(false)}
        variant={theme}
      >
        <div className="p-6">
          <h2 className="text-2xl font-semibold mb-2">{selectedResourceData?.name}</h2>
          <h3 className="text-sm mb-2 uppercase">{selectedResourceData?.type}</h3>
          <p className="text-sm mb-2">Provider: <span className="font-medium">{selectedResourceData?.provider}</span></p>
          <p className="text-sm mb-2">Instance count: <span className="font-medium">{selectedResourceData?.instances_count}</span></p>
          { selectedResourceData?.module !== undefined && <p className="text-sm mb-2">Module: <span className="font-medium">{selectedResourceData?.module || '(root)'}</span></p>}
          <p className="text-sm mb-2">Address: <span className="font-medium">{selectedResourceData?.addr}</span></p>
          <h3 className="text-lg font-semibold mt-4 mb-2">Instance{ (selectedResourceData?.instances_count ?? 0) > 1 ? 's' : '' } details</h3>
          <ul className="list-disc list-inside">
            {selectedResourceData?.instances?.map((inst) => (
              <li key={inst.addr} className="mb-2">
                <p className="text-sm">Address: <span className="font-medium">{inst.addr}</span></p>
                { inst.created_at && <p className="text-sm">Created at: <span className="font-medium">{new Date(inst.created_at).toLocaleString()}</span></p> }
                { inst.dependencies && inst.dependencies.length > 0 && (
                  <p className="text-sm">Dependencies: <span className="font-medium">{inst.dependencies.join(', ')}</span></p>
                ) }
                { inst.attributes && (
                  <details className="mt-1">
                    <summary className="cursor-pointer text-sm text-primary-500">View attributes</summary>
                    <pre className="bg-nuances-white p-2 rounded mt-1 overflow-auto text-xs text-nuances-black">
                      {JSON.stringify(inst.attributes, null,

                        2)}
                    </pre>
                  </details>
                ) }
              </li>
            )) }
          </ul>
          <h3 className="text-lg font-semibold mt-4 mb-2">Raw data</h3>
          <pre className="bg-nuances-white p-4 rounded-lg overflow-auto text-sm text-nuances-black">
            {JSON.stringify(selectedResourceData, null, 2)}
          </pre>
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
