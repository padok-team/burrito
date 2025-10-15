import React, { useContext, useEffect, useState } from 'react';


import { ThemeContext } from '@/contexts/ThemeContext';
import LayerStateGraph from '@/components/tools/LayerStateGraph';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { reactQueryKeys } from '@/clients/reactQueryConfig';
import { fetchLayer, syncLayer } from '@/clients/layers/client';
import LayerStatus from '@/components/status/LayerStatus';
import Button from '@/components/core/Button';
import type { Layer } from '@/clients/layers/types';

const Layer: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const { namespace = '', name = '' } = useParams();
  const navigate = useNavigate();

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
          />
        </div>
      </div>
    </div>
  );
};

export default Layer;
