import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';

import { fetchLayers } from '@/clients/layers/client';
import { reactQueryKeys } from '@/clients/reactQueryConfig';
import { Layer } from '@/clients/layers/types';

interface StatusBarProps {
  className?: string;
  variant?: 'light' | 'dark';
  layers?: Layer[];
}

type StatusVariant =
  | 'success'
  | 'warning'
  | 'error'
  | 'disabled'
  | 'apply-needed'
  | 'plan-needed'
  | 'running';

interface StatusItem {
  label: string;
  count: number;
  variant: StatusVariant;
  isParenthesis?: boolean;
}

interface LayerCounts {
  total: number;
  ok: number;
  error: number;
  running: number;
  applyNeeded: number;
  planNeeded: number;
}

// Helper functions
const getVariantStyles = (
  variant: StatusVariant,
  theme: 'light' | 'dark'
): string => {
  const variantStyles = {
    success: 'bg-status-success-default text-nuances-black',
    warning: 'bg-status-warning-default text-nuances-black',
    error: 'bg-status-error-default text-nuances-white',
    disabled:
      theme === 'light'
        ? 'bg-nuances-200 text-nuances-400'
        : 'bg-nuances-400 text-nuances-50',
    'apply-needed': 'bg-status-warning-default text-nuances-black',
    'plan-needed': 'bg-status-warning-default text-nuances-black',
    running: 'bg-blue-500 text-nuances-white'
  };

  return `flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium leading-4 ${variantStyles[variant]}`;
};

const formatLabel = (label: string, count: number): string => {
  if (label === 'Errors') {
    return count <= 1 ? 'Error' : 'Errors';
  }
  return label;
};

const computeLayerCounts = (layers: Layer[]): LayerCounts => {
  const counts: LayerCounts = {
    total: layers.length,
    ok: 0,
    error: 0,
    running: 0,
    applyNeeded: 0,
    planNeeded: 0
  };

  layers.forEach((layer) => {
    if (layer.isRunning) {
      counts.running++;
    }

    switch (layer.state) {
      case 'success':
        counts.ok++;
        break;
      case 'error':
        counts.error++;
        break;
      case 'warning':
        if (layer.lastResult?.toLowerCase().includes('apply')) {
          counts.applyNeeded++;
        } else {
          counts.planNeeded++;
        }
        break;
      case 'disabled':
        // Don't count disabled layers in any active status
        break;
    }
  });

  return counts;
};

const createCoreStatuses = (counts: LayerCounts): StatusItem[] => [
  { label: 'Total', count: counts.total, variant: 'disabled' },
  { label: 'OK', count: counts.ok, variant: 'success' },
  {
    label: 'Out of Sync',
    count: counts.applyNeeded + counts.planNeeded,
    variant: 'warning'
  },
  { label: 'Errors', count: counts.error, variant: 'error' }
];

const createAdditionalStatuses = (counts: LayerCounts): StatusItem[] => {
  const additionalStatuses: StatusItem[] = [];
  const hasAdditionalStatuses =
    counts.running > 0 || counts.applyNeeded > 0 || counts.planNeeded > 0;

  if (!hasAdditionalStatuses) {
    return additionalStatuses;
  }

  // Opening parenthesis
  additionalStatuses.push({
    label: '(',
    count: 0,
    variant: 'disabled',
    isParenthesis: true
  });

  // Add status bubbles
  if (counts.running > 0) {
    additionalStatuses.push({
      label: 'Running',
      count: counts.running,
      variant: 'running'
    });
  }

  if (counts.applyNeeded > 0) {
    additionalStatuses.push({
      label: 'Apply Needed',
      count: counts.applyNeeded,
      variant: 'apply-needed'
    });
  }

  if (counts.planNeeded > 0) {
    additionalStatuses.push({
      label: 'Plan Needed',
      count: counts.planNeeded,
      variant: 'plan-needed'
    });
  }

  // Closing parenthesis
  additionalStatuses.push({
    label: ')',
    count: 0,
    variant: 'disabled',
    isParenthesis: true
  });

  return additionalStatuses;
};

const StatusItem: React.FC<{
  item: StatusItem;
  theme: 'light' | 'dark';
}> = ({ item, theme }) => {
  const { label, count, variant, isParenthesis = false } = item;

  // If it's a parenthesis, just show the character without bubble styling
  if (isParenthesis) {
    return (
      <span
        className={`text-xs font-medium ${theme === 'light' ? 'text-nuances-400' : 'text-nuances-200'}`}
      >
        {label}
      </span>
    );
  }

  const styles = getVariantStyles(variant, theme);
  const displayLabel = formatLabel(label, count);

  return (
    <div className={styles}>
      <span>{count.toLocaleString()}</span>
      <span>{displayLabel.toLowerCase()}</span>
    </div>
  );
};

// Loading and Error Components
const LoadingState: React.FC<{ className: string }> = ({ className }) => (
  <div className={`flex items-center gap-2 px-3 py-2 ${className}`}>
    <div className="flex items-center gap-2">
      <div className="w-1 h-1 rounded-full bg-nuances-200 animate-pulse" />
      <div className="w-12 h-3 bg-nuances-200 rounded animate-pulse" />
      <div className="w-6 h-3 bg-nuances-200 rounded animate-pulse" />
    </div>
  </div>
);

const ErrorState: React.FC<{ className: string }> = ({ className }) => (
  <div
    className={`flex items-center gap-2 px-3 py-2 text-status-error-default ${className}`}
  >
    <div className="w-1 h-1 rounded-full bg-status-error-default" />
    <span className="text-xs font-medium">Status error</span>
  </div>
);

const LayersStatusBar: React.FC<StatusBarProps> = ({
  className = '',
  variant = 'light',
  layers
}) => {
  const layersQuery = useQuery({
    queryKey: reactQueryKeys.layers,
    queryFn: fetchLayers,
    refetchInterval: 30000,
    staleTime: 10000,
    gcTime: 60000,
    enabled: !layers
  });

  const layersData = layers || layersQuery.data?.results;

  const statusItems = useMemo(() => {
    if (!layersData) return [];

    const counts = computeLayerCounts(layersData);
    const coreStatuses = createCoreStatuses(counts);
    const additionalStatuses = createAdditionalStatuses(counts);

    return [...coreStatuses, ...additionalStatuses];
  }, [layersData]);

  // Show loading state only if we're fetching and no layers are provided
  if (!layers && layersQuery.isLoading) {
    return <LoadingState className={className} />;
  }

  // Show error state only if we're fetching and no layers are provided
  if (!layers && layersQuery.isError) {
    return <ErrorState className={className} />;
  }

  return (
    <div className={`flex items-center gap-2 px-3 py-2 ${className}`}>
      {statusItems.map((item, index) => (
        <StatusItem
          key={`${item.label}-${index}`}
          item={item}
          theme={variant}
        />
      ))}
      {!layers && layersQuery.isFetching && (
        <div className="flex items-center gap-1 text-nuances-400">
          <div className="w-0.5 h-0.5 rounded-full bg-nuances-400 animate-pulse" />
          <span className="text-xs">•••</span>
        </div>
      )}
    </div>
  );
};

export default LayersStatusBar;
